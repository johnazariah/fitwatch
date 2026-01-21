// Platform detection rules
const PLATFORMS = {
  mywhoosh: {
    name: 'MyWhoosh',
    urlPattern: /services?\.mywhoosh\.com|service26\.mywhoosh\.com/,
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
    urlPattern: /prod\.en\.igpsport\.com/,
    tokenExtractor: (headers) => {
      const auth = headers.find(h => h.name.toLowerCase() === 'authorization');
      return auth?.value?.replace('Bearer ', '');
    }
  },
  trainingpeaks: {
    name: 'TrainingPeaks',
    urlPattern: /trainingpeaks\.com\/api|api\.trainingpeaks\.com|www\.trainingpeaks\.com/,
    tokenExtractor: (headers) => {
      const auth = headers.find(h => h.name.toLowerCase() === 'authorization');
      return auth?.value?.replace('Bearer ', '');
    }
  }
};

// Listen for API requests
chrome.webRequest.onBeforeSendHeaders.addListener(
  (details) => {
    console.log(`[FitBridge] Intercepted: ${details.url}`);
    console.log(`[FitBridge] Headers:`, details.requestHeaders);
    
    for (const [key, platform] of Object.entries(PLATFORMS)) {
      if (platform.urlPattern.test(details.url)) {
        console.log(`[FitBridge] Matched platform: ${key}`);
        const token = platform.tokenExtractor(details.requestHeaders);
        console.log(`[FitBridge] Extracted token: ${token ? token.substring(0, 20) + '...' : 'null'}`);
        if (token && token.length > 20) {
          storeToken(key, token);
          updateBadge();
        }
        break;
      }
    }
  },
  { 
    urls: [
      "https://*.mywhoosh.com/*",
      "https://*.zwift.com/*",
      "https://*.igpsport.com/*",
      "https://*.trainingpeaks.com/*"
    ] 
  },
  ["requestHeaders", "extraHeaders"]
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
    console.log(`[FitBridge] Captured ${platform} token at ${tokens[platform].capturedAt}`);
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

// Log startup
console.log('[FitBridge] Extension loaded, watching for fitness platform tokens...');
