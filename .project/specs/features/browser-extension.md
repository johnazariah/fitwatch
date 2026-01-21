# FitBridge Browser Extension Spec

## Overview

A Chromium browser extension that captures authentication tokens from fitness platform sessions, enabling FitBridge to sync data on the user's behalf.

**Codename:** Moat Crossing  
**Target Browsers:** Chrome, Edge, Brave (Chromium-based)  
**Status:** Prototype Spec

## Problem Statement

Fitness platforms don't provide OAuth apps for third-party access. Users must manually:
1. Open browser DevTools
2. Find Network tab
3. Locate an API request
4. Copy Authorization header
5. Paste into FitBridge config

This is error-prone and intimidating for non-technical users.

## Solution

A browser extension that:
1. Detects when user logs into supported platforms
2. Captures the authentication token from API requests
3. Securely transmits token to FitBridge (local or cloud)
4. Provides simple "Connect" UI per platform

## Supported Platforms

| Platform | Token Type | Capture Point | Lifetime |
|----------|------------|---------------|----------|
| MyWhoosh | Bearer JWT | `api.mywhoosh.com/*` | ~7 days |
| Zwift | Bearer JWT | `us-or-rly101.zwift.com/*` | ~6 hours |
| iGPSport | Bearer JWT | `prod.igpsport.com/*` | ~7 days |
| TrainingPeaks | Cookie + Bearer | `tpapi.trainingpeaks.com/*` | Session |

## User Flow

### First-Time Setup

```
1. User installs extension from Chrome Web Store
   
2. Extension shows popup: "FitBridge Token Capture"
   - Status: "Not connected to any platforms"
   - Instructions: "Log in to your fitness platforms normally"

3. User navigates to my.zwift.com and logs in

4. Extension detects Zwift API traffic, captures token
   - Badge shows "1" (one platform connected)
   - Popup now shows: "✓ Zwift - Connected"

5. User logs into mywhoosh.com

6. Extension captures MyWhoosh token
   - Badge shows "2"
   - Popup: "✓ Zwift, ✓ MyWhoosh"

7. User clicks "Send to FitBridge"
   - Tokens sent to local CLI or cloud service
```

### Token Refresh Flow

```
1. User's Zwift token expires (6 hours)

2. User opens Zwift in browser (normal usage)

3. Extension auto-captures fresh token

4. Sends to FitBridge automatically (if auto-sync enabled)
```

## Technical Architecture

### Manifest V3 Structure

```
fitbridge-extension/
├── manifest.json
├── background.js          # Service worker
├── popup/
│   ├── popup.html
│   ├── popup.js
│   └── popup.css
├── icons/
│   ├── icon-16.png
│   ├── icon-48.png
│   └── icon-128.png
└── lib/
    └── storage.js
```

### manifest.json

```json
{
  "manifest_version": 3,
  "name": "FitBridge Token Capture",
  "version": "0.1.0",
  "description": "Securely capture fitness platform tokens for FitBridge sync",
  
  "permissions": [
    "storage",
    "webRequest"
  ],
  
  "host_permissions": [
    "https://api.mywhoosh.com/*",
    "https://*.zwift.com/*",
    "https://prod.igpsport.com/*",
    "https://tpapi.trainingpeaks.com/*"
  ],
  
  "background": {
    "service_worker": "background.js"
  },
  
  "action": {
    "default_popup": "popup/popup.html",
    "default_icon": {
      "16": "icons/icon-16.png",
      "48": "icons/icon-48.png",
      "128": "icons/icon-128.png"
    }
  },
  
  "icons": {
    "16": "icons/icon-16.png",
    "48": "icons/icon-48.png",
    "128": "icons/icon-128.png"
  }
}
```

### background.js (Service Worker)

```javascript
// Platform detection rules
const PLATFORMS = {
  mywhoosh: {
    name: 'MyWhoosh',
    urlPattern: /api\.mywhoosh\.com/,
    tokenExtractor: (headers) => {
      const auth = headers.find(h => h.name.toLowerCase() === 'authorization');
      return auth?.value?.replace('Bearer ', '');
    }
  },
  zwift: {
    name: 'Zwift',
    urlPattern: /\.zwift\.com.*\/api\//,
    tokenExtractor: (headers) => {
      const auth = headers.find(h => h.name.toLowerCase() === 'authorization');
      return auth?.value?.replace('Bearer ', '');
    }
  },
  igpsport: {
    name: 'iGPSport',
    urlPattern: /prod\.igpsport\.com/,
    tokenExtractor: (headers) => {
      const auth = headers.find(h => h.name.toLowerCase() === 'authorization');
      return auth?.value?.replace('Bearer ', '');
    }
  },
  trainingpeaks: {
    name: 'TrainingPeaks',
    urlPattern: /tpapi\.trainingpeaks\.com/,
    tokenExtractor: (headers) => {
      const auth = headers.find(h => h.name.toLowerCase() === 'authorization');
      return auth?.value?.replace('Bearer ', '');
    }
  }
};

// Listen for API requests
chrome.webRequest.onBeforeSendHeaders.addListener(
  async (details) => {
    for (const [key, platform] of Object.entries(PLATFORMS)) {
      if (platform.urlPattern.test(details.url)) {
        const token = platform.tokenExtractor(details.requestHeaders);
        if (token && token.length > 20) {
          await storeToken(key, token);
          updateBadge();
        }
        break;
      }
    }
    return { requestHeaders: details.requestHeaders };
  },
  { urls: ["<all_urls>"] },
  ["requestHeaders"]
);

async function storeToken(platform, token) {
  const data = await chrome.storage.local.get('tokens') || {};
  const tokens = data.tokens || {};
  
  // Only update if token changed
  if (tokens[platform]?.token !== token) {
    tokens[platform] = {
      token,
      capturedAt: new Date().toISOString(),
      platform: PLATFORMS[platform].name
    };
    await chrome.storage.local.set({ tokens });
    console.log(`Captured ${platform} token`);
  }
}

async function updateBadge() {
  const data = await chrome.storage.local.get('tokens');
  const count = Object.keys(data.tokens || {}).length;
  chrome.action.setBadgeText({ text: count > 0 ? String(count) : '' });
  chrome.action.setBadgeBackgroundColor({ color: '#4CAF50' });
}

// Initialize badge on startup
updateBadge();
```

### popup/popup.html

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body {
      width: 300px;
      padding: 16px;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    }
    h1 {
      font-size: 16px;
      margin: 0 0 16px 0;
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .platform {
      display: flex;
      align-items: center;
      padding: 8px;
      margin: 4px 0;
      border-radius: 4px;
      background: #f5f5f5;
    }
    .platform.connected {
      background: #e8f5e9;
    }
    .platform .status {
      margin-left: auto;
      font-size: 12px;
      color: #666;
    }
    .platform.connected .status {
      color: #4CAF50;
    }
    .actions {
      margin-top: 16px;
      display: flex;
      flex-direction: column;
      gap: 8px;
    }
    button {
      padding: 10px;
      border: none;
      border-radius: 4px;
      cursor: pointer;
      font-size: 14px;
    }
    .primary {
      background: #2196F3;
      color: white;
    }
    .primary:disabled {
      background: #ccc;
    }
    .secondary {
      background: #f0f0f0;
    }
    .help {
      font-size: 12px;
      color: #666;
      margin-top: 16px;
    }
  </style>
</head>
<body>
  <h1>🌉 FitBridge</h1>
  
  <div id="platforms"></div>
  
  <div class="actions">
    <button id="sendBtn" class="primary" disabled>Send to FitBridge</button>
    <button id="clearBtn" class="secondary">Clear All</button>
  </div>
  
  <div class="help">
    Log in to your fitness platforms normally. Tokens are captured automatically.
  </div>
  
  <script src="popup.js"></script>
</body>
</html>
```

### popup/popup.js

```javascript
const PLATFORM_NAMES = {
  mywhoosh: 'MyWhoosh',
  zwift: 'Zwift',
  igpsport: 'iGPSport',
  trainingpeaks: 'TrainingPeaks'
};

async function render() {
  const data = await chrome.storage.local.get(['tokens', 'fitbridgeEndpoint']);
  const tokens = data.tokens || {};
  
  const container = document.getElementById('platforms');
  container.innerHTML = '';
  
  for (const [key, name] of Object.entries(PLATFORM_NAMES)) {
    const token = tokens[key];
    const div = document.createElement('div');
    div.className = `platform ${token ? 'connected' : ''}`;
    
    const capturedTime = token 
      ? new Date(token.capturedAt).toLocaleTimeString() 
      : '';
    
    div.innerHTML = `
      <span>${token ? '✓' : '○'} ${name}</span>
      <span class="status">${token ? capturedTime : 'Not connected'}</span>
    `;
    container.appendChild(div);
  }
  
  // Enable send button if we have any tokens
  document.getElementById('sendBtn').disabled = Object.keys(tokens).length === 0;
}

document.getElementById('sendBtn').addEventListener('click', async () => {
  const data = await chrome.storage.local.get(['tokens', 'fitbridgeEndpoint']);
  const endpoint = data.fitbridgeEndpoint || 'http://localhost:5000/api/tokens';
  
  try {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data.tokens)
    });
    
    if (response.ok) {
      alert('Tokens sent to FitBridge!');
    } else {
      throw new Error(`HTTP ${response.status}`);
    }
  } catch (e) {
    // Fallback: copy to clipboard
    const tokenJson = JSON.stringify(data.tokens, null, 2);
    await navigator.clipboard.writeText(tokenJson);
    alert('Could not reach FitBridge. Tokens copied to clipboard.');
  }
});

document.getElementById('clearBtn').addEventListener('click', async () => {
  await chrome.storage.local.remove('tokens');
  chrome.action.setBadgeText({ text: '' });
  render();
});

// Initial render
render();

// Listen for storage changes
chrome.storage.onChanged.addListener(render);
```

## Token Transmission Options

### Option 1: Local HTTP Server (MVP)

FitBridge CLI runs a local HTTP server:

```
fitbridge serve --port 5000
```

Extension posts tokens to `http://localhost:5000/api/tokens`.

**Pros:** Simple, no cloud needed, works offline  
**Cons:** User must run CLI, port conflicts possible

### Option 2: Native Messaging

Extension uses Chrome native messaging to communicate directly with installed CLI:

```json
// com.fitbridge.cli.json (native messaging host)
{
  "name": "com.fitbridge.cli",
  "description": "FitBridge CLI",
  "path": "C:\\Users\\...\\fitbridge.exe",
  "type": "stdio",
  "allowed_origins": ["chrome-extension://..."]
}
```

**Pros:** Direct CLI integration, no HTTP server  
**Cons:** Complex setup, platform-specific paths

### Option 3: Cloud Relay (Future)

Extension posts to `https://api.fitbridge.io/tokens/{user_id}`:

```
User authenticates with FitBridge → gets user_id
Extension stores user_id
Tokens sent to cloud API
Cloud triggers sync on user's behalf
```

**Pros:** Works without CLI running, enables scheduled sync  
**Cons:** Requires cloud infrastructure, security considerations

## Security Considerations

### Token Storage

- Tokens stored in `chrome.storage.local` (encrypted by Chrome)
- Never synced via `chrome.storage.sync` (too sensitive)
- Cleared on extension uninstall

### Transmission Security

- Local: HTTP to localhost acceptable (never leaves machine)
- Cloud: HTTPS required, user authentication required
- Tokens encrypted in transit

### Minimal Permissions

- Only request `webRequest` for specific fitness domains
- No `tabs`, `history`, or broad permissions
- No content scripts (no page modification)

### Token Visibility

- Extension never displays full tokens
- Popup shows only connection status and timestamp
- Clipboard fallback shows tokens (user-initiated)

## Development Phases

### Phase 1: Local Prototype (2 days)

- [ ] Basic extension structure (manifest, popup, background)
- [ ] Capture tokens from MyWhoosh, Zwift
- [ ] Display connection status in popup
- [ ] Copy tokens to clipboard

### Phase 2: CLI Integration (2 days)

- [ ] Add local HTTP endpoint to FitBridge CLI
- [ ] Extension posts to localhost
- [ ] CLI updates config.json with new tokens
- [ ] Auto-sync trigger option

### Phase 3: Polish (2 days)

- [ ] Add iGPSport, TrainingPeaks capture
- [ ] Token expiry detection/warning
- [ ] Options page for settings
- [ ] Badge shows stale tokens differently

### Phase 4: Cloud Integration (Future)

- [ ] User authentication flow
- [ ] Cloud token relay
- [ ] Push notifications for sync status

## Testing Plan

### Manual Testing

1. Install unpacked extension in Chrome
2. Navigate to each platform, log in
3. Verify token capture (check popup)
4. Click "Send to FitBridge" with CLI running
5. Verify token appears in config.json
6. Run sync, verify it works

### Automated Testing

- Unit tests for token extraction logic
- Mock `chrome.webRequest` events
- Test popup rendering with various states

## Open Questions

1. **Token refresh strategy**: Should extension auto-send on capture, or batch?
2. **Multi-profile support**: Handle multiple browser profiles?
3. **Token validation**: Verify token is valid before storing?
4. **Garmin/Wahoo**: Different auth mechanisms—research needed

## Appendix: Token Formats

### Zwift JWT

```
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```
Payload contains: `sub` (user ID), `exp` (expiry ~6hrs), `scope`

### MyWhoosh JWT

```
eyJhbGciOiJIUzUxMiJ9...
```
Payload contains: `email`, `exp` (~7 days)

### iGPSport JWT

```
eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...
```
Payload contains: `userId`, `exp` (~7 days)

## Related Documents

- [ADR-018: Local Activity Register](../decisions/018-local-activity-register.md)
- [ADR-019: Serverless Architecture](../decisions/019-serverless-multi-tenant-architecture.md)
