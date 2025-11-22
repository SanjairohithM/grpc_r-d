// API client for gRPC gateway
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081/api';

export interface UnaryResponse {
  message: string;
}

export interface StreamMessage {
  message: string;
}

// 1. SERVER STREAMING - Server sends multiple responses
export function serverStream(
  name: string,
  onMessage: (data: StreamMessage) => void,
  onComplete?: () => void,
  onError?: (error: Error) => void
): () => void {
  const eventSource = new EventSource(`${API_BASE}/server-stream?name=${encodeURIComponent(name)}`);
  
  eventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      onMessage(data);
    } catch (error) {
      console.error('Parse error:', error);
    }
  };
  
  eventSource.addEventListener('done', () => {
    if (onComplete) onComplete();
    eventSource.close();
  });
  
  eventSource.onerror = (error) => {
    if (onError) onError(new Error('Stream error'));
    eventSource.close();
  };
  
  // Return cleanup function
  return () => eventSource.close();
}

// 2. CLIENT STREAMING - Client sends multiple requests
export async function clientStream(names: string[]): Promise<UnaryResponse> {
  const response = await fetch(`${API_BASE}/client-stream`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(names),
  });
  
  if (!response.ok) {
    throw new Error('Client stream failed');
  }
  
  return response.json();
}

// 3. BIDIRECTIONAL STREAMING - WebSocket for two-way communication
export interface BidirectionalClient {
  send: (message: string) => void;
  close: () => void;
}

export function bidirectionalStream(
  onMessage: (data: StreamMessage) => void,
  onOpen?: () => void,
  onClose?: () => void,
  onError?: (error: Error) => void
): BidirectionalClient {
  const wsUrl = API_BASE.replace('http://', 'ws://').replace('https://', 'wss://');
  const ws = new WebSocket(`${wsUrl}/bidirectional`);
  
  ws.onopen = () => {
    console.log('WebSocket connected');
    if (onOpen) onOpen();
  };
  
  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      onMessage(data);
    } catch (error) {
      console.error('Parse error:', error);
    }
  };
  
  ws.onclose = () => {
    console.log('WebSocket closed');
    if (onClose) onClose();
  };
  
  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
    if (onError) onError(new Error('WebSocket error'));
  };
  
  return {
    send: (message: string) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ name: message }));
      }
    },
    close: () => ws.close(),
  };
}

// Health check - uses server stream endpoint to check connectivity
export async function checkHealth(): Promise<boolean> {
  try {
    // Try to connect to server stream endpoint (quick check)
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 2000); // 2 second timeout
    
    const response = await fetch(`${API_BASE}/server-stream?name=HealthCheck`, {
      method: 'GET',
      signal: controller.signal,
    });
    
    clearTimeout(timeoutId);
    return response.ok || response.status === 200;
  } catch {
    return false;
  }
}

