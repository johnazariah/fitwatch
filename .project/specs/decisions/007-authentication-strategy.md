# ADR-007: Authentication Strategy

## Status
Accepted

## Context
We need authentication at two levels:
1. **Application auth**: Users logging into our platform
2. **Provider auth**: Connecting to Garmin, Strava, Intervals.icu, etc.

## Decision

### Application Authentication

**Use Azure Entra ID (formerly Azure AD B2C) for user authentication.**

| Option | Pros | Cons | Decision |
|--------|------|------|----------|
| No auth (single-user) | Simplest | Can't share device | Start here for MVP |
| Azure Entra ID | SSO, MFA, managed | Complexity | Add for multi-user |
| Simple local auth | Self-contained | Security burden | Avoid |

#### MVP: Single-User Mode
```csharp
// No auth required - whoever can access the URL is the user
// Suitable for self-hosted behind VPN/Tailscale
```

#### Post-MVP: Azure Entra ID
```csharp
builder.Services.AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
    .AddMicrosoftIdentityWebApi(builder.Configuration.GetSection("AzureAd"));
```

### Provider OAuth

**Use Azure Functions with stable callback URLs for OAuth flows.**

#### Why Azure Functions for OAuth?

| Challenge | Solution |
|-----------|----------|
| OAuth needs stable callback URL | Functions have stable `*.azurewebsites.net` URLs |
| Each provider needs different flow | One function per provider |
| Token refresh | Function handles silently |
| Token storage | Key Vault via Managed Identity |

#### OAuth Flow

```
1. User clicks "Connect Garmin" in Web UI
2. API generates auth URL with state token
3. User redirected to Garmin login
4. Garmin redirects to Azure Function callback
5. Function exchanges code for tokens
6. Function stores tokens in Key Vault
7. Function redirects user back to Web UI
8. Web UI shows "Connected ✓"
```

#### Provider Callback URLs

```
https://fitsync-functions.azurewebsites.net/api/oauth/garmin/callback
https://fitsync-functions.azurewebsites.net/api/oauth/strava/callback
https://fitsync-functions.azurewebsites.net/api/oauth/intervalsicu/callback
```

#### Token Storage

```csharp
public class ProviderTokenService
{
    private readonly SecretClient _keyVault;
    
    public async Task StoreTokens(string userId, string provider, OAuthTokens tokens)
    {
        var secretName = $"oauth-{userId}-{provider}";
        var secretValue = JsonSerializer.Serialize(new
        {
            AccessToken = tokens.AccessToken,
            RefreshToken = tokens.RefreshToken,
            ExpiresAt = tokens.ExpiresAt
        });
        
        await _keyVault.SetSecretAsync(secretName, secretValue);
    }
    
    public async Task<OAuthTokens?> GetTokens(string userId, string provider)
    {
        var secretName = $"oauth-{userId}-{provider}";
        try
        {
            var secret = await _keyVault.GetSecretAsync(secretName);
            return JsonSerializer.Deserialize<OAuthTokens>(secret.Value.Value);
        }
        catch (RequestFailedException ex) when (ex.Status == 404)
        {
            return null;
        }
    }
}
```

#### Token Refresh

```csharp
public class TokenRefreshMiddleware
{
    public async Task<OAuthTokens> EnsureValidToken(string userId, string provider)
    {
        var tokens = await _tokenService.GetTokens(userId, provider);
        
        if (tokens == null)
            throw new NotConnectedException(provider);
        
        if (tokens.ExpiresAt > DateTime.UtcNow.AddMinutes(5))
            return tokens; // Still valid
        
        // Refresh the token
        var newTokens = await _providerClient.RefreshToken(provider, tokens.RefreshToken);
        await _tokenService.StoreTokens(userId, provider, newTokens);
        
        return newTokens;
    }
}
```

### Provider-Specific Notes

| Provider | OAuth Version | Notes |
|----------|---------------|-------|
| Garmin | OAuth 1.0a | Legacy but works; use library |
| Strava | OAuth 2.0 | Standard; upload scope only |
| Intervals.icu | API Key | No OAuth needed; just store key |
| Wahoo | OAuth 2.0 | Standard |
| TrainingPeaks | OAuth 2.0 | Standard |

## Consequences

### Positive
- Stable OAuth callback URLs without custom domain
- Tokens secured in Key Vault with encryption
- Managed Identity = no secrets in config
- Azure Entra ID handles MFA, account recovery, etc.

### Negative
- Requires Azure subscription
- OAuth 1.0a (Garmin) adds complexity
- Token refresh failures need graceful handling

### Fallback: API Key Entry
For providers without OAuth or difficult OAuth:
```
Enter your Intervals.icu API key:
[________________] 
(Find this at intervals.icu/settings)

[Save]
```

## Related Decisions
- ADR-006: Azure + .NET Aspire Architecture
