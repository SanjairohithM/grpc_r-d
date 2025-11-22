// gRPC-Web client - Direct connection to gRPC server (no gateway needed!)
// Using @improbable-eng/grpc-web for browser-to-gRPC communication

import { grpc } from '@improbable-eng/grpc-web';

// gRPC-Web endpoint (direct to server, no gateway!)
const GRPC_WEB_URL = process.env.NEXT_PUBLIC_GRPC_WEB_URL || 'http://localhost:8080';

// Service and method names matching proto file
const SERVICE_NAME = 'helloworld.Greeter';

// Message interfaces
export interface HelloRequest {
  name: string;
}

export interface HelloReply {
  message: string;
}

export interface StreamMessage {
  message: string;
}

// Helper to create metadata
function createMetadata(): grpc.Metadata {
  return new grpc.Metadata();
}

// 2. SERVER STREAMING - Server sends multiple responses
export function serverStream(
  name: string,
  onMessage: (data: StreamMessage) => void,
  onComplete?: () => void,
  onError?: (error: Error) => void
): () => void {
  const request: HelloRequest = { name };
  
  const client = grpc.client<HelloRequest, HelloReply>({
    method: `${SERVICE_NAME}/SayHelloServerStream`,
    host: GRPC_WEB_URL,
    transport: grpc.WebsocketTransport(), // Use WebSocket for streaming
  }, {
    transport: grpc.WebsocketTransport(),
  });

  let cancelled = false;

  client.onHeaders((headers: grpc.Metadata) => {
    // Headers received
  });

  client.onMessage((reply: HelloReply) => {
    if (!cancelled) {
      onMessage({ message: reply.message });
    }
  });

  client.onEnd((status: grpc.Code, statusMessage: string, trailers: grpc.Metadata) => {
    if (status === grpc.Code.OK) {
      if (onComplete) onComplete();
    } else {
      if (onError) {
        onError(new Error(`Stream error: ${statusMessage || status}`));
      }
    }
  });

  client.start(createMetadata());
  client.send(request);

  // Return cleanup function
  return () => {
    cancelled = true;
    client.close();
  };
}

// 3. CLIENT STREAMING - Client sends multiple requests
export async function clientStream(names: string[]): Promise<{ message: string }> {
  return new Promise((resolve, reject) => {
    const client = grpc.client<HelloRequest, HelloReply>({
      method: `${SERVICE_NAME}/SayHelloClientStream`,
      host: GRPC_WEB_URL,
      transport: grpc.WebsocketTransport(),
    }, {
      transport: grpc.WebsocketTransport(),
    });

    let response: HelloReply | null = null;

    client.onHeaders((headers: grpc.Metadata) => {
      // Headers received
    });

    client.onMessage((reply: HelloReply) => {
      response = reply;
    });

    client.onEnd((status: grpc.Code, statusMessage: string, trailers: grpc.Metadata) => {
      if (status === grpc.Code.OK) {
        if (response) {
          resolve({ message: response.message });
        } else {
          reject(new Error('No response received'));
        }
      } else {
        reject(new Error(`Client stream error: ${statusMessage || status}`));
      }
    });

    client.start(createMetadata());

    // Send all names
    names.forEach((name, index) => {
      const request: HelloRequest = { name };
      if (index === names.length - 1) {
        // Last message - close the stream
        client.send(request);
        client.finishSend();
      } else {
        client.send(request);
      }
    });
  });
}

// 4. BIDIRECTIONAL STREAMING - Both send multiple messages
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
  const client = grpc.client<HelloRequest, HelloReply>({
    method: `${SERVICE_NAME}/SayHelloBidirectional`,
    host: GRPC_WEB_URL,
    transport: grpc.WebsocketTransport(), // WebSocket required for bidirectional
  }, {
    transport: grpc.WebsocketTransport(),
  });

  let isOpen = false;

  client.onHeaders((headers: grpc.Metadata) => {
    isOpen = true;
    if (onOpen) onOpen();
  });

  client.onMessage((reply: HelloReply) => {
    onMessage({ message: reply.message });
  });

  client.onEnd((status: grpc.Code, statusMessage: string, trailers: grpc.Metadata) => {
    isOpen = false;
    if (status === grpc.Code.OK) {
      if (onClose) onClose();
    } else {
      if (onError) {
        onError(new Error(`Bidirectional stream error: ${statusMessage || status}`));
      }
    }
  });

  client.start(createMetadata());

  return {
    send: (message: string) => {
      if (isOpen) {
        const request: HelloRequest = { name: message };
        client.send(request);
      }
    },
    close: () => {
      client.finishSend();
      client.close();
    },
  };
}

// Health check
export async function checkHealth(): Promise<boolean> {
  return new Promise((resolve) => {
    try {
      const stream = serverStream(
        'HealthCheck',
        () => {
          // Received message - server is healthy
          resolve(true);
        },
        () => {
          resolve(true);
        },
        () => {
          resolve(false);
        }
      );

      // Timeout after 2 seconds
      setTimeout(() => {
        stream(); // Cancel stream
        resolve(false);
      }, 2000);
    } catch {
      resolve(false);
    }
  });
}
