# ADR-016: Browser Extension Token Capture Architecture

## Status
Accepted

## Context

FitBridge needs to authenticate with fitness platforms (TrainingPeaks, MyWhoosh, Zwift) that don't offer public OAuth APIs. Users shouldn't have to manually copy tokens from browser dev tools.

Previous approaches considered:
- Manual token copy (current spike) - works but poor UX
- Playwright automation in cloud - complex, security concerns with stored credentials
- Local desktop agent - requires installation, maintenance

We need a solution that:
1. Works in the user's browser (where they're already logged in)
2. Doesn't require us to store credentials
3. Captures tokens automatically
4. Works with the cloud-hosted Orleans backend

## Decision

**Build a "FitBridge Connector" browser extension that captures auth tokens and sends them to our backend.**

### Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  USER'S BROWSER                                                              │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │ FitBridge Connector Extension                                            ││
│  │                                                                          ││
│  │  Permissions:                                                            ││
│  │  - trainingpeaks.com / tpapi.trainingpeaks.com                          ││
│  │  - mywhoosh.com / service14.mywhoosh.com                                ││
│  │  - zwift.com                                                            ││
│  │  - fitbridge.io (our domain)                                            ││
│  │                                                                          ││
│  │  Capabilities:                                                           ││
│  │  - webRequest: Intercept Authorization headers                          ││
│  │  - cookies: Read auth cookies if needed                                 ││
│  │  - storage: Remember which user to send tokens to                       ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │ FitBridge Web App (fitbridge.io)                                         ││
│  │                                                                          ││
│  │  1. User logs into FitBridge (Entra ID)                                 ││
│  │  2. "Connect TrainingPeaks" button                                       ││
│  │  3. Opens new tab to trainingpeaks.com                                  ││
│  │  4. User logs in (or already logged in)                                 ││
│  │  5. Extension detects auth, sends to backend                            ││
│  │  6. Web app shows "✓ Connected"                                         ││
│  └─────────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────────┘
                                           │
                                           ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│  AZURE                                                                       │
│                                                                              │
│  ┌─────────────────┐  ┌──────────────────────┐  ┌─────────────────────────┐ │
│  │ API Gateway     │  │ Orleans Cluster       │  │ Key Vault              │ │
│  │                 │  │                       │  │                        │ │
│  │ POST /tokens    │─→│ UserGrain             │─→│ Encrypted tokens       │ │
│  │ (from extension)│  │ - Store token         │  │ - Per user+provider    │ │
│  │                 │  │ - Trigger sync        │  │                        │ │
│  └─────────────────┘  └──────────────────────┘  └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Extension Flow

```typescript
// background.ts - Service worker

// Listen for auth headers on provider APIs
chrome.webRequest.onSendHeaders.addListener(
  async (details) => {
    const authHeader = details.requestHeaders?.find(
      h => h.name.toLowerCase() === 'authorization'
    );
    
    if (authHeader?.value?.startsWith('Bearer ')) {
      const token = authHeader.value.replace('Bearer ', '');
      const provider = detectProvider(details.url);
      
      // Get the user's FitBridge session from storage
      const { userId, apiKey } = await chrome.storage.local.get(['userId', 'apiKey']);
      
      if (userId && apiKey) {
        await sendTokenToFitBridge(userId, provider, token, apiKey);
      }
    }
  },
  { urls: [
    '*://tpapi.trainingpeaks.com/*',
    '*://service14.mywhoosh.com/*',
    '*://api.zwift.com/*'
  ]},
  ['requestHeaders']
);

// Link extension to user's FitBridge account
chrome.runtime.onMessageExternal.addListener(
  (message, sender, sendResponse) => {
    if (sender.origin === 'https://fitbridge.io' && message.type === 'LINK_ACCOUNT') {
      chrome.storage.local.set({
        userId: message.userId,
        apiKey: message.apiKey
      });
      sendResponse({ success: true });
    }
  }
);
```

### Web App Integration

```typescript
// FitBridge web app - connect button handler
async function connectProvider(provider: string) {
  // Check if extension is installed
  try {
    const response = await chrome.runtime.sendMessage(
      EXTENSION_ID, 
      { type: 'PING' }
    );
    
    if (response?.installed) {
      // Link extension to this user's account
      await chrome.runtime.sendMessage(EXTENSION_ID, {
        type: 'LINK_ACCOUNT',
        userId: currentUser.id,
        apiKey: await getExtensionApiKey()
      });
      
      // Open provider login in new tab
      window.open(PROVIDER_URLS[provider], '_blank');
      
      // Poll for connection status
      pollForConnection(provider);
    }
  } catch {
    // Extension not installed - show install prompt
    showInstallExtensionModal();
  }
}
```

### Token Capture by Provider

| Provider | Token Location | Capture Method |
|----------|---------------|----------------|
| TrainingPeaks | `Authorization: Bearer xxx` header | webRequest listener |
| MyWhoosh | `accessToken` in request headers | webRequest listener |
| Zwift | `Authorization: Bearer xxx` header | webRequest listener |
| Strava | OAuth (proper flow) | No extension needed |
| Garmin | OAuth (if we get partnership) | No extension needed |
| Intervals.icu | User provides API key | No extension needed |

### Security Considerations

1. **Token transmission**: Extension → Backend uses HTTPS + short-lived API key
2. **Token storage**: Key Vault with per-user encryption
3. **Extension permissions**: Minimal - only specific provider domains
4. **No credentials**: We never see usernames/passwords
5. **User consent**: Clear permission prompt during install
6. **Token scope**: Read-only where possible

## Alternatives Considered

### Option A: Bookmarklet
- **Pros**: No installation, works on any browser
- **Cons**: Manual action required, less reliable token capture
- **Verdict**: Keep as fallback for unsupported browsers

### Option B: Proxy Server
- **Pros**: No extension needed
- **Cons**: User must configure browser, security red flags
- **Verdict**: Rejected - too invasive

### Option C: Desktop App with Embedded Browser
- **Pros**: Full control
- **Cons**: Separate install, platform-specific
- **Verdict**: Rejected - extension is lighter weight

## Consequences

### Positive
- Clean UX: install once, tokens captured automatically
- Secure: no credential storage
- Works with 2FA (user logs in normally)
- Auto-refresh: extension captures new tokens when user logs in again
- Cross-browser: Chrome, Edge, Firefox (same codebase with minor tweaks)

### Negative
- Extension development/maintenance overhead
- Chrome Web Store review process
- Users must trust/install extension
- Firefox/Safari may need separate builds

### Future Enhancements
- **Auto-sync trigger**: Extension notifies backend immediately on new auth
- **Token health monitoring**: Extension pings providers to check token validity
- **Multi-profile support**: Handle multiple browser profiles
