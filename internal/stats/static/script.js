const UPDATE_INTERVAL = 5000; // 5 Sekunden

function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatDate(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('de-DE');
}

function showError(message) {
    const summary = document.querySelector('.summary');
    summary.innerHTML = `
        <div class="error-card">
            <h4>Fehler</h4>
            <p>${message}</p>
        </div>`;
}

function updateSummary(stats) {
    const summary = document.querySelector('.summary');
    summary.innerHTML = `
        <div class="stat-card">
            <h4>Gesamtanfragen</h4>
            <p>${stats.total_requests}</p>
        </div>
        <div class="stat-card">
            <h4>Traffic In</h4>
            <p>${formatBytes(stats.total_bytes_in)}</p>
        </div>
        <div class="stat-card">
            <h4>Traffic Out</h4>
            <p>${formatBytes(stats.total_bytes_out)}</p>
        </div>
        <div class="stat-card">
            <h4>Aktive Clients</h4>
            <p>${stats.active_clients}</p>
        </div>
    `;
}

function updateRecentRequests(requests) {
    const tbody = document.querySelector('.recent-requests tbody');
    tbody.innerHTML = requests.map(req => `
        <tr>
            <td>${formatDate(req.timestamp)}</td>
            <td>${req.client_ip}</td>
            <td>${req.method}</td>
            <td>${req.host}</td>
            <td>${req.path}</td>
            <td>${req.status}</td>
            <td>${formatBytes(req.bytes_in + req.bytes_out)}</td>
        </tr>
    `).join('');
}

function updateClientStats(clients) {
    const tbody = document.querySelector('.client-stats tbody');
    tbody.innerHTML = clients.map(client => `
        <tr>
            <td>${client.ip}</td>
            <td>${formatBytes(client.bytes_in)}</td>
            <td>${formatBytes(client.bytes_out)}</td>
            <td>${formatBytes(client.bytes_total)}</td>
            <td>${client.requests}</td>
            <td>${formatDate(client.last_seen)}</td>
        </tr>
    `).join('');
}

async function updateStats() {
    try {
        const response = await fetch('/api/stats');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        
        if (data.error) {
            throw new Error(data.error);
        }

        updateSummary(data);
        updateRecentRequests(data.recent_requests || []);
        updateClientStats(data.client_stats || []);    } catch (error) {
        console.error('Error updating stats:', error);
        showError('Proxy-Server nicht erreichbar. Bitte überprüfen Sie die Verbindung.');
    }
}

// Initial update
updateStats();

// Set update interval
setInterval(updateStats, UPDATE_INTERVAL);

// Update the displayed interval
document.querySelector('.update-interval').textContent = UPDATE_INTERVAL / 1000;
