/**
 * Copyright (c) 2025 Michael Lechner
 * This software is released under the MIT License.
 * See the LICENSE file for further details.
 */

// Update interval for refreshing statistics (in milliseconds)
const UPDATE_INTERVAL = 5000;

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
}

/**
 * Updates the recent requests table
 * @param {Array} requests - Array of recent request objects
 */
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

/**
 * Updates the client statistics table
 * @param {Array} clients - Array of client statistics objects
 */
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

/**
 * Fetches and updates all statistics
 */
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
        updateClientStats(data.client_stats || []);
    } catch (error) {
        console.error('Error updating stats:', error);
        showError('Proxy server not reachable. Please check your connection.');
    }
}

// Initialize statistics display
updateStats();

// Set up automatic updates
setInterval(updateStats, UPDATE_INTERVAL);

// Update the displayed refresh interval
document.querySelector('.update-interval').textContent = UPDATE_INTERVAL / 1000;
