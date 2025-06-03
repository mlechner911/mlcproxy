/**
 * Copyright (c) 2025 Michael Lechner
 * This software is released under the MIT License.
 * See the LICENSE file for further details.
 */

// Update interval for refreshing statistics (in milliseconds)
const UPDATE_INTERVAL = 5000;
const MAX_POINTS = 60; // 5 minutes of history

// Statistics tracking
let lastRequests = 0;
let lastBytesIn = 0;
let lastBytesOut = 0;
let lastClients = 0;

let rateHistory = [];
let bytesInHistory = [];
let bytesOutHistory = [];
let clientsHistory = [];

// Tracking für Min/Max-Werte
let requestsMinMax = { min: Infinity, max: 0 };
let bytesInMinMax = { min: Infinity, max: 0 };
let bytesOutMinMax = { min: Infinity, max: 0 };
let clientsMinMax = { min: Infinity, max: 0 };

/**
 * Formats byte values to human-readable strings using browser locale
 * @param {number} bytes - The number of bytes to format
 * @returns {string} Formatted string with units (e.g., "1.24 MB")
 */
function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    const formatter = new Intl.NumberFormat(navigator.language, {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
    });
    
    const value = bytes / Math.pow(k, i);
    return formatter.format(value) + ' ' + sizes[i];
}

/**
 * Formats timestamps using browser locale settings
 * @param {number} timestamp - Unix timestamp to format
 * @returns {string} Formatted time string
 */
function formatDate(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleTimeString(navigator.language, {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

/**
 * Creates a clone of a template element
 * @param {string} templateId - ID selector of the template to clone
 * @returns {DocumentFragment|null} Cloned template or null if not found
 */
function cloneTemplate(templateId) {
    const template = document.querySelector(templateId);
    if (!template) {
        console.error(`Template ${templateId} not found`);
        return null;
    }
    return template.content.cloneNode(true);
}

/**
 * Displays an error message in the summary section
 * @param {string} message - Error message to display
 */
function showError(message) {
    const summary = document.querySelector('.summary');
    const errorElement = cloneTemplate('#error-template');
    if (!errorElement) {
        console.error('Internal error: Error template not found');
        return;
    }
    
    summary.innerHTML = '';
    errorElement.querySelector('[data-placeholder="message"]').textContent = message;
    summary.appendChild(errorElement);
}

/**
 * Updates the min/max values in the stat card
 * @param {HTMLElement} container - The stat card container
 * @param {number} value - Current value
 * @param {Object} minMax - Object containing min and max values
 * @param {Function} formatter - Function to format the value
 */
function updateMinMax(container, value, minMax, formatter = (v) => v.toString()) {
    minMax.min = Math.min(minMax.min, value);
    minMax.max = Math.max(minMax.max, value);
    
    const minElement = container.querySelector('.min-max .min');
    const maxElement = container.querySelector('.min-max .max');
    
    if (minElement && maxElement) {
        minElement.textContent = `Min: ${formatter(minMax.min)}`;
        maxElement.textContent = `Max: ${formatter(minMax.max)}`;
    }
}

/**
 * Updates the request rate graph
 * @param {number} totalRequests - Current total requests
 */
function updateRequestRate(totalRequests) {
    if (lastRequests === 0) {
        lastRequests = totalRequests;
        return;
    }

    // Calculate requests per second
    const requestDiff = totalRequests - lastRequests;
    const ratePerSecond = requestDiff / (UPDATE_INTERVAL / 1000);
    lastRequests = totalRequests;

    // Update history
    rateHistory.push(ratePerSecond);
    if (rateHistory.length > MAX_POINTS) {
        rateHistory.shift();
    }

    // Find all required elements
    const pathElement = document.querySelector('.rate-line');
    const container = pathElement?.closest('.stat-card');
    const svgElement = container?.querySelector('.rate-graph');
    
    if (!pathElement || !container || !svgElement) return;

    // Update current rate value and min/max
    const rateValue = container.querySelector('.rate-value');
    if (rateValue) {
        rateValue.textContent = ratePerSecond.toFixed(1) + '/s';
    }

    updateMinMax(container, ratePerSecond, requestsMinMax, v => v.toFixed(1) + '/s');

    // Get SVG dimensions and calculate points
    const svgWidth = svgElement.width.baseVal.value;
    const svgHeight = svgElement.height.baseVal.value;
    const maxRate = Math.max(1, ...rateHistory);

    const points = rateHistory.map((rate, index) => {
        const x = (index / (MAX_POINTS - 1)) * svgWidth;
        const y = svgHeight - (rate / maxRate) * svgHeight;
        return `${x},${y}`;
    });

    // Update path
    pathElement.setAttribute('d', `M ${points.join(' L ')}`);
}

/**
 * Aktualisiert den Traffic-In-Graphen
 * @param {number} bytesIn - Aktuelle eingehende Bytes
 */
function updateTrafficIn(bytesIn) {
    if (lastBytesIn === 0) {
        lastBytesIn = bytesIn;
        return;
    }

    const bytesDiff = bytesIn - lastBytesIn;
    const ratePerSecond = bytesDiff / (UPDATE_INTERVAL / 1000);
    lastBytesIn = bytesIn;

    bytesInHistory.push(ratePerSecond);
    if (bytesInHistory.length > MAX_POINTS) {
        bytesInHistory.shift();
    }

    const pathElement = document.querySelector('.traffic-in');
    const container = pathElement?.closest('.stat-card');
    const svgElement = container?.querySelector('.rate-graph');
    
    if (!pathElement || !container || !svgElement) return;

    updateMinMax(container, ratePerSecond, bytesInMinMax, formatBytes);

    const svgWidth = svgElement.width.baseVal.value;
    const svgHeight = svgElement.height.baseVal.value;
    const maxRate = Math.max(1, ...bytesInHistory);

    const points = bytesInHistory.map((rate, index) => {
        const x = (index / (MAX_POINTS - 1)) * svgWidth;
        const y = svgHeight - (rate / maxRate) * svgHeight;
        return `${x},${y}`;
    });

    pathElement.setAttribute('d', `M ${points.join(' L ')}`);

    const rateValue = container.querySelector('.rate-value');
    if (rateValue) {
        rateValue.textContent = formatBytes(ratePerSecond) + '/s';
    }
}

/**
 * Aktualisiert den Traffic-Out-Graphen
 * @param {number} bytesOut - Aktuelle ausgehende Bytes
 */
function updateTrafficOut(bytesOut) {
    if (lastBytesOut === 0) {
        lastBytesOut = bytesOut;
        return;
    }

    const bytesDiff = bytesOut - lastBytesOut;
    const ratePerSecond = bytesDiff / (UPDATE_INTERVAL / 1000);
    lastBytesOut = bytesOut;

    bytesOutHistory.push(ratePerSecond);
    if (bytesOutHistory.length > MAX_POINTS) {
        bytesOutHistory.shift();
    }

    const pathElement = document.querySelector('.traffic-out');
    const container = pathElement?.closest('.stat-card');
    const svgElement = container?.querySelector('.rate-graph');
    
    if (!pathElement || !container || !svgElement) return;

    updateMinMax(container, ratePerSecond, bytesOutMinMax, formatBytes);

    const svgWidth = svgElement.width.baseVal.value;
    const svgHeight = svgElement.height.baseVal.value;
    const maxRate = Math.max(1, ...bytesOutHistory);

    const points = bytesOutHistory.map((rate, index) => {
        const x = (index / (MAX_POINTS - 1)) * svgWidth;
        const y = svgHeight - (rate / maxRate) * svgHeight;
        return `${x},${y}`;
    });

    pathElement.setAttribute('d', `M ${points.join(' L ')}`);

    const rateValue = container.querySelector('.rate-value');
    if (rateValue) {
        rateValue.textContent = formatBytes(ratePerSecond) + '/s';
    }
}

/**
 * Aktualisiert den Client-Trend-Graphen
 * @param {number} activeClients - Aktuelle Anzahl aktiver Clients
 */
function updateClientTrend(activeClients) {
    const clientDiff = activeClients - lastClients;
    lastClients = activeClients;

    clientsHistory.push(activeClients);
    if (clientsHistory.length > MAX_POINTS) {
        clientsHistory.shift();
    }

    const pathElement = document.querySelector('.clients');
    const container = pathElement?.closest('.stat-card');
    const svgElement = container?.querySelector('.rate-graph');
    
    if (!pathElement || !container || !svgElement) return;

    updateMinMax(container, activeClients, clientsMinMax);

    const svgWidth = svgElement.width.baseVal.value;
    const svgHeight = svgElement.height.baseVal.value;
    const maxClients = Math.max(1, ...clientsHistory);

    const points = clientsHistory.map((clients, index) => {
        const x = (index / (MAX_POINTS - 1)) * svgWidth;
        const y = svgHeight - (clients / maxClients) * svgHeight;
        return `${x},${y}`;
    });

    pathElement.setAttribute('d', `M ${points.join(' L ')}`);

    const rateValue = container.querySelector('.rate-value');
    if (rateValue) {
        rateValue.textContent = activeClients.toString();
    }
    
    const trendIndicator = container.querySelector('.trend-indicator');
    if (trendIndicator) {
        if (clientDiff > 0) {
            trendIndicator.textContent = '↑';
            trendIndicator.className = 'trend-indicator up';
        } else if (clientDiff < 0) {
            trendIndicator.textContent = '↓';
            trendIndicator.className = 'trend-indicator down';
        } else {
            trendIndicator.textContent = '→';
            trendIndicator.className = 'trend-indicator stable';
        }
    }
}

/**
 * Updates the summary cards with current statistics
 * @param {Object} stats - Statistics data object
 */
function updateSummary(stats) {
    const summary = document.querySelector('.summary');
    summary.innerHTML = '';
    
    const templates = [
        { id: '#total-requests-template', value: stats.total_requests },
        { id: '#traffic-in-template', value: formatBytes(stats.total_bytes_in) },
        { id: '#traffic-out-template', value: formatBytes(stats.total_bytes_out) },
        { id: '#active-clients-template', value: stats.active_clients }
    ];

    templates.forEach(({ id, value }) => {
        const element = cloneTemplate(id);
        if (element) {
            element.querySelector('[data-placeholder="value"]').textContent = value;
            summary.appendChild(element);
        }
    });

    // Update request rate after template is added
    updateRequestRate(stats.total_requests);
    updateTrafficIn(stats.total_bytes_in);
    updateTrafficOut(stats.total_bytes_out);
    updateClientTrend(stats.active_clients);
}

/**
 * Groups identical requests and adds a counter
 * @param {Array} requests - Array of request objects
 * @returns {Array} Grouped requests with count
 */
function groupIdenticalRequests(requests) {
    const grouped = new Map();
    
    requests.forEach(req => {
        const key = `${req.method}|${req.host}|${req.path}|${req.status}`;
        if (!grouped.has(key)) {
            grouped.set(key, {
                ...req,
                count: 1,
                bytes_total: req.bytes_in + req.bytes_out
            });
        } else {
            const existing = grouped.get(key);
            existing.count++;
            existing.bytes_total += (req.bytes_in + req.bytes_out);
            // Update timestamp if newer
            if (req.timestamp > existing.timestamp) {
                existing.timestamp = req.timestamp;
            }
        }
    });
    
    return Array.from(grouped.values())
        .sort((a, b) => b.timestamp - a.timestamp);
}

/**
 * Updates the recent requests table
 * @param {Array} requests - Array of recent request objects
 */
function updateRecentRequests(requests) {
    const tbody = document.querySelector('.recent-requests tbody');
    if (!tbody) return;
    
    const groupedRequests = groupIdenticalRequests(requests);
    
    tbody.innerHTML = groupedRequests.map(req => `
        <tr>
            <td>${formatDate(req.timestamp)}</td>
            <td>${req.client_ip}</td>
            <td>${req.method}</td>
            <td>${req.host}</td>
            <td>${req.path}</td>
            <td>${req.status}</td>
            <td>${formatBytes(req.bytes_total)}</td>            <td>${req.count > 1 ? `<small>${req.count}×</small>` : ''}</td>
        </tr>
    `).join('');
}

/**
 * Updates the client statistics table
 * @param {Array} clients - Array of client statistics objects
 */
function updateClientStats(clients) {
    const tbody = document.querySelector('.client-stats tbody');
    if (!tbody) return;

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

/**
 * Fetches and updates all statistics
 */
async function updateStats() {
    try {
        const response = await fetch('./stats.json');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const stats = await response.json();
        updateSummary(stats);
        updateClientStats(stats.client_stats || []);
        updateRecentRequests(stats.recent_requests || []);
        updateLastUpdateTime();
    } catch (error) {
        console.error('Error updating stats:', error);
        showError('Verbindung zum Proxy-Server nicht möglich. Bitte überprüfen Sie die Verbindung.');
    }
}

// Initialize statistics display
updateStats();

/**
 * Updates the last update time display
 */
function updateLastUpdateTime() {
    const element = document.querySelector('.last-update');
    if (element) {
        element.textContent = formatDate(Date.now());
    }
}

// Set up automatic updates
setInterval(updateStats, UPDATE_INTERVAL);

/**
 * Manages theme switching functionality
 */
const themeManager = {
    storageKey: 'preferred-theme',
    
    /**
     * Initializes theme manager
     */
    init() {
        this.select = document.getElementById('theme-select');
        if (!this.select) return;
        
        // Load saved preference
        const savedTheme = localStorage.getItem(this.storageKey) || 'auto';
        this.select.value = savedTheme;
        
        // Set initial theme
        this.setTheme(savedTheme);
        
        // Listen for changes
        this.select.addEventListener('change', () => {
            const theme = this.select.value;
            this.setTheme(theme);
            localStorage.setItem(this.storageKey, theme);
        });
        
        // Listen for system theme changes
        if (window.matchMedia) {
            window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
                if (this.select.value === 'auto') {
                    this.setTheme('auto');
                }
            });
        }
    },
    
    /**
     * Sets the theme
     * @param {string} theme - Theme to set ('auto', 'light', or 'dark')
     */
    setTheme(theme) {
        if (theme === 'auto') {
            // Check system preference
            const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
            document.documentElement.setAttribute('data-theme', prefersDark ? 'dark' : 'light');
        } else {
            document.documentElement.setAttribute('data-theme', theme);
        }
    }
};

// Initialize theme manager after DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    themeManager.init();
});
