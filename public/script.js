// Configuration
const API_BASE = 'http://localhost:3000/api';
let eventSource = null;
let websocket = null;
let namesArray = [];

// Initialize
window.onload = function() {
    checkGRPCStatus();
    setInterval(checkGRPCStatus, 5000);
};

// Check gRPC server status
async function checkGRPCStatus() {
    const indicator = document.getElementById('grpcStatus');
    const statusText = document.getElementById('statusText');
    
    try {
        const response = await fetch(`${API_BASE}/unary`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: 'StatusCheck' })
        });
        
        if (response.ok) {
            indicator.className = 'status-indicator connected';
            statusText.textContent = 'Connected to gRPC Server ✓';
        } else {
            throw new Error('Server error');
        }
    } catch (error) {
        indicator.className = 'status-indicator disconnected';
        statusText.textContent = 'Disconnected from Server ✗';
    }
}

// ===========================================
// 1. UNARY RPC
// ===========================================
async function testUnary() {
    const name = document.getElementById('unaryName').value;
    const resultDiv = document.getElementById('unaryResult');
    
    if (!name) {
        showResult(resultDiv, 'Please enter a name', 'error');
        return;
    }
    
    resultDiv.textContent = 'Sending request...';
    resultDiv.className = 'result';
    
    try {
        const response = await fetch(`${API_BASE}/unary`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: name })
        });
        
        if (!response.ok) throw new Error('Request failed');
        
        const data = await response.json();
        showResult(resultDiv, data.message, 'success');
    } catch (error) {
        showResult(resultDiv, `Error: ${error.message}`, 'error');
    }
}

// ===========================================
// 2. SERVER STREAMING
// ===========================================
function testServerStream() {
    const name = document.getElementById('streamName').value;
    const resultDiv = document.getElementById('streamResult');
    
    if (!name) {
        showResult(resultDiv, 'Please enter a name', 'error');
        return;
    }
    
    // Close existing stream
    stopServerStream();
    
    resultDiv.textContent = 'Connecting to stream...\n';
    resultDiv.className = 'result stream-result';
    
    // Create EventSource for Server-Sent Events
    eventSource = new EventSource(`${API_BASE}/server-stream?name=${encodeURIComponent(name)}`);
    
    let messageCount = 0;
    
    eventSource.onmessage = function(event) {
        messageCount++;
        const data = JSON.parse(event.data);
        appendToResult(resultDiv, `[${messageCount}] ${data.message}\n`, 'success');
    };
    
    eventSource.addEventListener('done', function(event) {
        appendToResult(resultDiv, '\n✓ Stream completed!\n', 'success');
        stopServerStream();
    });
    
    eventSource.onerror = function(error) {
        appendToResult(resultDiv, '\n✗ Stream error or completed\n', 'error');
        stopServerStream();
    };
}

function stopServerStream() {
    if (eventSource) {
        eventSource.close();
        eventSource = null;
    }
}

// ===========================================
// 3. CLIENT STREAMING
// ===========================================
function addName() {
    const input = document.getElementById('clientStreamName');
    const name = input.value.trim();
    
    if (!name) return;
    
    namesArray.push(name);
    input.value = '';
    updateNamesList();
}

function removeName(index) {
    namesArray.splice(index, 1);
    updateNamesList();
}

function clearNames() {
    namesArray = [];
    updateNamesList();
    document.getElementById('clientStreamResult').textContent = '';
}

function updateNamesList() {
    const container = document.getElementById('namesList');
    
    if (namesArray.length === 0) {
        container.innerHTML = '<div style="color: #999; padding: 10px;">No names added yet</div>';
        return;
    }
    
    container.innerHTML = namesArray.map((name, index) => `
        <div class="name-tag">
            ${name}
            <span class="remove" onclick="removeName(${index})">×</span>
        </div>
    `).join('');
}

async function sendClientStream() {
    const resultDiv = document.getElementById('clientStreamResult');
    
    if (namesArray.length === 0) {
        showResult(resultDiv, 'Please add at least one name', 'error');
        return;
    }
    
    resultDiv.textContent = `Sending ${namesArray.length} names to server...`;
    resultDiv.className = 'result';
    
    try {
        const response = await fetch(`${API_BASE}/client-stream`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(namesArray)
        });
        
        if (!response.ok) throw new Error('Request failed');
        
        const data = await response.json();
        showResult(resultDiv, data.message, 'success');
        
        // Clear after successful send
        setTimeout(() => {
            clearNames();
        }, 2000);
    } catch (error) {
        showResult(resultDiv, `Error: ${error.message}`, 'error');
    }
}

// ===========================================
// 4. BIDIRECTIONAL STREAMING (WebSocket)
// ===========================================
function connectBidirectional() {
    const statusSpan = document.getElementById('wsStatus');
    const messagesDiv = document.getElementById('chatMessages');
    
    if (websocket && websocket.readyState === WebSocket.OPEN) {
        addChatMessage('Already connected', 'received');
        return;
    }
    
    // Connect to WebSocket
    websocket = new WebSocket('ws://localhost:3000/api/bidirectional');
    
    websocket.onopen = function() {
        statusSpan.textContent = 'Connected ✓';
        statusSpan.className = 'connected';
        addChatMessage('Connected to bidirectional stream!', 'received');
    };
    
    websocket.onmessage = function(event) {
        const data = JSON.parse(event.data);
        addChatMessage(data.message, 'received');
    };
    
    websocket.onerror = function(error) {
        addChatMessage('Connection error!', 'received');
        statusSpan.textContent = 'Error ✗';
        statusSpan.className = '';
    };
    
    websocket.onclose = function() {
        statusSpan.textContent = 'Disconnected';
        statusSpan.className = '';
        addChatMessage('Disconnected from server', 'received');
    };
}

function disconnectBidirectional() {
    if (websocket) {
        websocket.close();
        websocket = null;
    }
}

function sendChatMessage() {
    const input = document.getElementById('chatInput');
    const message = input.value.trim();
    
    if (!message) return;
    
    if (!websocket || websocket.readyState !== WebSocket.OPEN) {
        addChatMessage('Not connected! Please connect first.', 'received');
        return;
    }
    
    // Send to server
    websocket.send(JSON.stringify({ name: message }));
    
    // Add to chat
    addChatMessage(message, 'sent');
    
    // Clear input
    input.value = '';
}

function handleChatKeyPress(event) {
    if (event.key === 'Enter') {
        sendChatMessage();
    }
}

function addChatMessage(message, type) {
    const messagesDiv = document.getElementById('chatMessages');
    const messageEl = document.createElement('div');
    messageEl.className = `chat-message ${type}`;
    messageEl.textContent = message;
    messagesDiv.appendChild(messageEl);
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
}

// ===========================================
// UTILITY FUNCTIONS
// ===========================================
function showResult(element, text, type) {
    element.textContent = text;
    element.className = `result ${type}`;
}

function appendToResult(element, text, type) {
    element.textContent += text;
    if (type) {
        element.className = `result stream-result ${type}`;
    }
    element.scrollTop = element.scrollHeight;
}

// Cleanup on page unload
window.onbeforeunload = function() {
    stopServerStream();
    disconnectBidirectional();
};

// Handle Enter key for inputs
document.getElementById('unaryName').addEventListener('keypress', function(e) {
    if (e.key === 'Enter') testUnary();
});

document.getElementById('streamName').addEventListener('keypress', function(e) {
    if (e.key === 'Enter') testServerStream();
});

document.getElementById('clientStreamName').addEventListener('keypress', function(e) {
    if (e.key === 'Enter') addName();
});

// Initialize names list
updateNamesList();

