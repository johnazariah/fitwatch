const PLATFORM_INFO = {
  mywhoosh: { name: 'MyWhoosh', icon: '🚴' },
  zwift: { name: 'Zwift', icon: '🏔️' },
  igpsport: { name: 'iGPSport', icon: '📡' },
  trainingpeaks: { name: 'TrainingPeaks', icon: '📊' }
};

function showToast(message, duration = 2000) {
  const toast = document.getElementById('toast');
  toast.textContent = message;
  toast.classList.add('show');
  setTimeout(() => toast.classList.remove('show'), duration);
}

function decodeJwt(token) {
  try {
    // JWT has 3 parts separated by dots
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    
    // Decode the payload (second part)
    const payload = JSON.parse(atob(parts[1]));
    return payload;
  } catch (e) {
    return null;
  }
}

function getExpiryInfo(token) {
  const payload = decodeJwt(token);
  if (!payload || !payload.exp) {
    return { status: 'unknown', message: 'Connected' };
  }
  
  const now = Date.now() / 1000;
  const exp = payload.exp;
  const hoursLeft = (exp - now) / 3600;
  
  if (hoursLeft < 0) {
    return { status: 'expired', message: 'Please log in again' };
  } else if (hoursLeft < 2) {
    return { status: 'expiring', message: 'Log in again soon' };
  } else if (hoursLeft < 24) {
    return { status: 'ok', message: `Good for ${Math.round(hoursLeft)} hours` };
  } else {
    const daysLeft = Math.round(hoursLeft / 24);
    return { status: 'ok', message: `Good for ${daysLeft} days` };
  }
}

function formatTime(isoString) {
  const date = new Date(isoString);
  const now = new Date();
  const diffMs = now - date;
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  
  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  return date.toLocaleDateString();
}

async function render() {
  const data = await chrome.storage.local.get(['tokens', 'fitbridgeEndpoint']);
  const tokens = data.tokens || {};
  
  // Render platforms
  const container = document.getElementById('platforms');
  container.innerHTML = '';
  
  for (const [key, info] of Object.entries(PLATFORM_INFO)) {
    const token = tokens[key];
    const div = document.createElement('div');
    
    let statusClass = '';
    let statusText = 'Not connected';
    
    if (token) {
      const expiry = getExpiryInfo(token.token);
      statusClass = expiry.status === 'expired' ? 'expired' : 
                    expiry.status === 'expiring' ? 'expiring' : 'connected';
      statusText = expiry.message;
    }
    
    div.className = `platform ${statusClass}`;
    
    div.innerHTML = `
      <span class="icon">${info.icon}</span>
      <div class="info">
        <div class="name">${info.name}</div>
        <div class="status">${statusText}</div>
      </div>
    `;
    container.appendChild(div);
  }
  
  // Enable buttons if we have tokens
  const hasTokens = Object.keys(tokens).length > 0;
  document.getElementById('copyBtn').disabled = !hasTokens;
  document.getElementById('sendBtn').disabled = !hasTokens;
  
  // Load saved endpoint
  const endpoint = data.fitbridgeEndpoint || 'http://localhost:5847/api/tokens';
  document.getElementById('endpoint').value = endpoint;
}

// Copy tokens to clipboard
document.getElementById('copyBtn').addEventListener('click', async () => {
  const data = await chrome.storage.local.get('tokens');
  const tokens = data.tokens || {};
  
  // Format for FitBridge config
  const config = {};
  for (const [key, tokenData] of Object.entries(tokens)) {
    config[`${key}Token`] = tokenData.token;
  }
  
  await navigator.clipboard.writeText(JSON.stringify(config, null, 2));
  showToast('✓ Tokens copied to clipboard');
});

// Send to FitBridge CLI
document.getElementById('sendBtn').addEventListener('click', async () => {
  const data = await chrome.storage.local.get(['tokens', 'fitbridgeEndpoint']);
  const endpoint = data.fitbridgeEndpoint || 'http://localhost:5847/api/tokens';
  
  try {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data.tokens)
    });
    
    if (response.ok) {
      showToast('✓ Tokens sent to FitBridge!');
    } else {
      throw new Error(`HTTP ${response.status}`);
    }
  } catch (e) {
    showToast('✗ Could not reach FitBridge CLI', 3000);
    console.error('Send failed:', e);
  }
});

// Clear all tokens
document.getElementById('clearBtn').addEventListener('click', async () => {
  await chrome.storage.local.remove('tokens');
  await chrome.action.setBadgeText({ text: '' });
  render();
  showToast('Cleared all tokens');
});

// Save endpoint on change
document.getElementById('endpoint').addEventListener('change', async (e) => {
  await chrome.storage.local.set({ fitbridgeEndpoint: e.target.value });
  showToast('Endpoint saved');
});

// Initial render
render();

// Listen for storage changes (tokens captured in background)
chrome.storage.onChanged.addListener((changes, area) => {
  if (area === 'local' && changes.tokens) {
    render();
  }
});
