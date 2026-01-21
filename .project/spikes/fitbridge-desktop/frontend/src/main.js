import './style.css';
import {GetTokens, ClearTokens} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

const PLATFORMS = {
    mywhoosh: { name: 'MyWhoosh', icon: '🚴' },
    zwift: { name: 'Zwift', icon: '🏔️' },
    igpsport: { name: 'iGPSport', icon: '📡' },
    trainingpeaks: { name: 'TrainingPeaks', icon: '📊' }
};

document.querySelector('#app').innerHTML = `
    <div class="container">
        <h1>🌉 FitBridge</h1>
        <p class="subtitle">Site Connector</p>
        
        <div id="platforms"></div>
        
        <div class="status">
            <div class="status-dot"></div>
            <span>Listening on port 5847</span>
        </div>
        
        <button id="clearBtn" class="btn secondary">Clear All</button>
        
        <p class="help">Open the browser extension and click "Send to FitBridge CLI"</p>
    </div>
`;

function decodeJwt(token) {
    try {
        const parts = token.split('.');
        if (parts.length !== 3) return null;
        const payload = JSON.parse(atob(parts[1]));
        return payload;
    } catch (e) {
        return null;
    }
}

function getExpiryMessage(token) {
    const payload = decodeJwt(token);
    if (!payload || !payload.exp) return 'Connected';
    
    const now = Date.now() / 1000;
    const hoursLeft = (payload.exp - now) / 3600;
    
    if (hoursLeft < 0) return 'Expired - log in again';
    if (hoursLeft < 2) return 'Log in again soon';
    if (hoursLeft < 24) return `Good for ${Math.round(hoursLeft)}h`;
    return `Good for ${Math.round(hoursLeft / 24)}d`;
}

function getStatusClass(token) {
    const payload = decodeJwt(token);
    if (!payload || !payload.exp) return 'connected';
    
    const now = Date.now() / 1000;
    const hoursLeft = (payload.exp - now) / 3600;
    
    if (hoursLeft < 0) return 'expired';
    if (hoursLeft < 2) return 'expiring';
    return 'connected';
}

function render(tokens) {
    const container = document.getElementById('platforms');
    container.innerHTML = '';
    
    for (const [key, info] of Object.entries(PLATFORMS)) {
        const token = tokens[key];
        const div = document.createElement('div');
        
        let statusClass = '';
        let statusText = 'Not connected';
        
        if (token && token.token) {
            statusClass = getStatusClass(token.token);
            statusText = getExpiryMessage(token.token);
        }
        
        div.className = `platform ${statusClass}`;
        div.innerHTML = `
            <span class="icon">${info.icon}</span>
            <div class="info">
                <div class="name">${info.name}</div>
                <div class="status-text">${statusText}</div>
            </div>
        `;
        container.appendChild(div);
    }
}

// Initial load
GetTokens().then(render);

// Listen for updates from the Go backend
EventsOn('tokens-updated', (tokens) => {
    console.log('Tokens updated:', tokens);
    render(tokens);
});

// Clear button
document.getElementById('clearBtn').addEventListener('click', () => {
    ClearTokens().then(() => GetTokens()).then(render);
});
