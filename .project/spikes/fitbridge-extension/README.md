# FitBridge Site Connector

Connects your fitness platform sessions to FitBridge for seamless activity sync.

## Installation (Developer Mode)

1. Open Chrome/Edge and navigate to `chrome://extensions/`
2. Enable **Developer mode** (toggle in top right)
3. Click **Load unpacked**
4. Select the `fitbridge-extension` folder

## Usage

1. Install the extension
2. Log in to your fitness platforms normally:
   - [MyWhoosh](https://mywhoosh.com)
   - [Zwift](https://my.zwift.com)
   - [iGPSport](https://my.igpsport.com)
   - [TrainingPeaks](https://app.trainingpeaks.com)
3. Click the FitBridge extension icon
4. See connected platforms with ✓
5. Click **Copy Tokens to Clipboard** or **Send to FitBridge CLI**

## Token Output Format

When copied, tokens are formatted for FitBridge config:

```json
{
  "mywhooshToken": "eyJ...",
  "zwiftToken": "eyJ...",
  "igpsportToken": "eyJ..."
}
```

## Development

### Reload After Changes

1. Make code changes
2. Go to `chrome://extensions/`
3. Click the refresh icon on the extension card

### View Background Script Logs

1. Go to `chrome://extensions/`
2. Click "Inspect views: service worker"
3. Opens DevTools for background.js

### Test Token Capture

1. Open DevTools Network tab on a fitness site
2. Log in or navigate around
3. Check extension popup for captured token
4. Check background script console for capture logs

## Files

```
fitbridge-extension/
├── manifest.json       # Extension config
├── background.js       # Service worker (token capture)
├── popup/
│   ├── popup.html      # Extension popup UI
│   └── popup.js        # Popup logic
└── icons/              # Extension icons (placeholder)
```

## Notes

- Tokens are stored in `chrome.storage.local` (encrypted by Chrome)
- Tokens are NOT synced across devices
- Zwift tokens expire in ~6 hours
- MyWhoosh/iGPSport tokens last ~7 days
