'use client';

import { useState, useEffect } from 'react';
import { serverStream, clientStream, bidirectionalStream, checkHealth } from '@/lib/grpc-api';

export default function Home() {
  const [connected, setConnected] = useState(false);
  
  // Pattern 1: Server Streaming
  const [streamName, setStreamName] = useState('Bob');
  const [streamMessages, setStreamMessages] = useState<string[]>([]);
  const [streaming, setStreaming] = useState(false);
  const [streamCleanup, setStreamCleanup] = useState<(() => void) | null>(null);
  
  // Pattern 3: Client Streaming
  const [clientStreamName, setClientStreamName] = useState('');
  const [namesList, setNamesList] = useState<string[]>([]);
  const [clientStreamResult, setClientStreamResult] = useState('');
  
  // Pattern 4: Bidirectional
  const [chatInput, setChatInput] = useState('');
  const [chatMessages, setChatMessages] = useState<Array<{text: string, sent: boolean}>>([]);
  const [wsClient, setWsClient] = useState<any>(null);
  const [wsConnected, setWsConnected] = useState(false);
  
  // Check health on mount
  useEffect(() => {
    const check = async () => {
      const healthy = await checkHealth();
      setConnected(healthy);
    };
    check();
    const interval = setInterval(check, 5000);
    return () => clearInterval(interval);
  }, []);
  
  // Pattern 1: Server Streaming
  const handleServerStream = () => {
    if (streaming && streamCleanup) {
      streamCleanup();
      setStreaming(false);
      return;
    }
    
    setStreamMessages([]);
    setStreaming(true);
    
    const cleanup = serverStream(
      streamName,
      (data) => {
        setStreamMessages(prev => [...prev, data.message]);
      },
      () => {
        setStreaming(false);
        setStreamMessages(prev => [...prev, '✓ Stream completed']);
      },
      (error) => {
        setStreamMessages(prev => [...prev, '✗ Error: ' + error.message]);
        setStreaming(false);
      }
    );
    
    setStreamCleanup(() => cleanup);
  };
  
  // Pattern 3: Client Streaming
  const addName = () => {
    if (clientStreamName.trim()) {
      setNamesList([...namesList, clientStreamName.trim()]);
      setClientStreamName('');
    }
  };
  
  const removeName = (index: number) => {
    setNamesList(namesList.filter((_, i) => i !== index));
  };
  
  const sendClientStream = async () => {
    if (namesList.length === 0) return;
    
    setClientStreamResult('Sending...');
    try {
      const response = await clientStream(namesList);
      setClientStreamResult(response.message);
    } catch (error) {
      setClientStreamResult('Error: ' + (error as Error).message);
    }
  };
  
  const clearNames = () => {
    setNamesList([]);
    setClientStreamResult('');
  };
  
  // Pattern 3: Bidirectional Streaming
  const connectBidirectional = () => {
    if (wsConnected && wsClient) {
      wsClient.close();
      return;
    }
    
    const client = bidirectionalStream(
      (data) => {
        setChatMessages(prev => [...prev, { text: data.message, sent: false }]);
      },
      () => {
        setWsConnected(true);
        setChatMessages(prev => [...prev, { text: 'Connected!', sent: false }]);
      },
      () => {
        setWsConnected(false);
        setChatMessages(prev => [...prev, { text: 'Disconnected', sent: false }]);
      },
      (error) => {
        setChatMessages(prev => [...prev, { text: 'Error: ' + error.message, sent: false }]);
      }
    );
    
    setWsClient(client);
  };
  
  const sendChatMessage = () => {
    if (!chatInput.trim() || !wsClient || !wsConnected) return;
    
    setChatMessages(prev => [...prev, { text: chatInput, sent: true }]);
    wsClient.send(chatInput);
    setChatInput('');
  };
  
  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-600 to-blue-600 p-8">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <header className="bg-white rounded-2xl shadow-2xl p-8 mb-8">
          <h1 className="text-4xl font-bold text-center bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent mb-2">
            ⚡ gRPC Communication Patterns
          </h1>
          <p className="text-center text-gray-600 text-lg mb-4">
            Next.js Frontend + Go Backend + HTTP/2 + gRPC
          </p>
          <div className="text-center">
            <span className={`inline-flex items-center px-4 py-2 rounded-full ${connected ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
              <span className={`w-3 h-3 rounded-full mr-2 ${connected ? 'bg-green-500' : 'bg-red-500'} animate-pulse`}></span>
              {connected ? 'Connected to gRPC Server ✓' : 'Disconnected from Server ✗'}
            </span>
          </div>
        </header>
        
        {/* Patterns Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          
          {/* Pattern 1: Server Streaming */}
          <div className="bg-white rounded-2xl shadow-xl p-6 hover:shadow-2xl transition-all duration-300 transform hover:-translate-y-1">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-2xl font-bold text-green-600">1️⃣ Server Streaming</h2>
              <span className="px-3 py-1 bg-green-100 text-green-800 rounded-full text-sm font-semibold">
                Request → Multiple
              </span>
            </div>
            <p className="text-gray-600 mb-4">Server sends continuous stream of data</p>
            
            <div className="space-y-4">
              <input
                type="text"
                value={streamName}
                onChange={(e) => setStreamName(e.target.value)}
                placeholder="Enter your name"
                className="w-full px-4 py-3 border-2 border-gray-300 rounded-lg focus:border-green-500 focus:outline-none"
              />
              <div className="flex gap-2">
                <button
                  onClick={handleServerStream}
                  className={`flex-1 ${streaming ? 'bg-red-500 hover:bg-red-600' : 'bg-gradient-to-r from-green-500 to-teal-500'} text-white py-3 rounded-lg font-semibold hover:shadow-lg transition-all`}
                >
                  {streaming ? 'Stop Stream' : 'Start Streaming'}
                </button>
              </div>
              
              <div className="h-48 overflow-y-auto p-4 bg-gray-50 border-2 border-gray-200 rounded-lg">
                <h4 className="font-semibold text-gray-800 mb-2">Messages:</h4>
                {streamMessages.length === 0 ? (
                  <p className="text-gray-400">No messages yet...</p>
                ) : (
                  <div className="space-y-2">
                    {streamMessages.map((msg, i) => (
                      <div key={i} className="p-2 bg-white rounded animate-fade-in">
                        {msg}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
            
            <div className="mt-4 p-3 bg-gray-50 rounded-lg">
              <p className="text-sm text-gray-600"><strong>Use Cases:</strong> Live stock prices, Real-time notifications</p>
            </div>
          </div>
          
          {/* Pattern 2: Client Streaming */}
          <div className="bg-white rounded-2xl shadow-xl p-6 hover:shadow-2xl transition-all duration-300 transform hover:-translate-y-1">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-2xl font-bold text-blue-600">2️⃣ Client Streaming</h2>
              <span className="px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm font-semibold">
                Multiple → Response
              </span>
            </div>
            <p className="text-gray-600 mb-4">Client sends multiple messages, server responds once</p>
            
            <div className="space-y-4">
              <div className="flex gap-2">
                <input
                  type="text"
                  value={clientStreamName}
                  onChange={(e) => setClientStreamName(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && addName()}
                  placeholder="Enter name and press Add"
                  className="flex-1 px-4 py-3 border-2 border-gray-300 rounded-lg focus:border-blue-500 focus:outline-none"
                />
                <button
                  onClick={addName}
                  className="px-6 bg-blue-500 text-white rounded-lg font-semibold hover:bg-blue-600 transition-all"
                >
                  Add
                </button>
              </div>
              
              <div className="min-h-[100px] p-4 bg-gray-50 border-2 border-gray-200 rounded-lg">
                <h4 className="font-semibold text-gray-800 mb-2">Names to send ({namesList.length}):</h4>
                {namesList.length === 0 ? (
                  <p className="text-gray-400">No names added yet...</p>
                ) : (
                  <div className="flex flex-wrap gap-2">
                    {namesList.map((name, i) => (
                      <span key={i} className="px-3 py-1 bg-blue-500 text-white rounded-full text-sm flex items-center gap-2">
                        {name}
                        <button onClick={() => removeName(i)} className="hover:text-red-200 font-bold">×</button>
                      </span>
                    ))}
                  </div>
                )}
              </div>
              
              <div className="flex gap-2">
                <button
                  onClick={sendClientStream}
                  disabled={namesList.length === 0}
                  className="flex-1 bg-gradient-to-r from-blue-500 to-purple-500 text-white py-3 rounded-lg font-semibold hover:shadow-lg transition-all disabled:opacity-50"
                >
                  Send All
                </button>
                <button
                  onClick={clearNames}
                  className="px-6 bg-gray-300 text-gray-700 rounded-lg font-semibold hover:bg-gray-400 transition-all"
                >
                  Clear
                </button>
              </div>
              
              {clientStreamResult && (
                <div className="p-4 bg-blue-50 border-2 border-blue-200 rounded-lg">
                  <h4 className="font-semibold text-blue-800 mb-2">Server Response:</h4>
                  <p className="text-gray-800">{clientStreamResult}</p>
                </div>
              )}
            </div>
            
            <div className="mt-4 p-3 bg-gray-50 rounded-lg">
              <p className="text-sm text-gray-600"><strong>Use Cases:</strong> File uploads, Batch inserts, IoT data</p>
            </div>
          </div>
          
          {/* Pattern 3: Bidirectional Streaming */}
          <div className="bg-white rounded-2xl shadow-xl p-6 hover:shadow-2xl transition-all duration-300 transform hover:-translate-y-1">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-2xl font-bold text-orange-600">3️⃣ Bidirectional</h2>
              <span className="px-3 py-1 bg-orange-100 text-orange-800 rounded-full text-sm font-semibold">
                Multiple ↔ Multiple
              </span>
            </div>
            <p className="text-gray-600 mb-4">Real-time bidirectional communication</p>
            
            <div className="space-y-4">
              <div className="flex gap-2">
                <button
                  onClick={connectBidirectional}
                  className={`flex-1 ${wsConnected ? 'bg-red-500 hover:bg-red-600' : 'bg-gradient-to-r from-orange-500 to-pink-500'} text-white py-3 rounded-lg font-semibold hover:shadow-lg transition-all`}
                >
                  {wsConnected ? 'Disconnect' : 'Connect'}
                </button>
                <div className={`px-4 py-3 rounded-lg font-semibold ${wsConnected ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}`}>
                  {wsConnected ? 'Connected ✓' : 'Disconnected'}
                </div>
        </div>
              
              <div className="h-64 overflow-y-auto p-4 bg-gray-50 border-2 border-gray-200 rounded-lg">
                {chatMessages.length === 0 ? (
                  <p className="text-gray-400">No messages yet. Connect to start chatting...</p>
                ) : (
                  <div className="space-y-2">
                    {chatMessages.map((msg, i) => (
                      <div
                        key={i}
                        className={`p-3 rounded-lg animate-fade-in ${
                          msg.sent
                            ? 'bg-gradient-to-r from-orange-500 to-pink-500 text-white ml-8'
                            : 'bg-white border-2 border-gray-300 mr-8'
                        }`}
                      >
                        {msg.text}
                      </div>
                    ))}
                  </div>
                )}
              </div>
              
              <div className="flex gap-2">
                <input
                  type="text"
                  value={chatInput}
                  onChange={(e) => setChatInput(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && sendChatMessage()}
                  placeholder="Type a message..."
                  disabled={!wsConnected}
                  className="flex-1 px-4 py-3 border-2 border-gray-300 rounded-lg focus:border-orange-500 focus:outline-none disabled:bg-gray-100"
                />
                <button
                  onClick={sendChatMessage}
                  disabled={!wsConnected || !chatInput.trim()}
                  className="px-6 bg-gradient-to-r from-orange-500 to-pink-500 text-white rounded-lg font-semibold hover:shadow-lg transition-all disabled:opacity-50"
                >
                  Send
                </button>
              </div>
            </div>
            
            <div className="mt-4 p-3 bg-gray-50 rounded-lg">
              <p className="text-sm text-gray-600"><strong>Use Cases:</strong> Chat apps, Real-time collaboration, Games</p>
            </div>
          </div>
          
        </div>
        
        {/* Performance Section */}
        <div className="mt-8 bg-white rounded-2xl shadow-xl p-8">
          <h2 className="text-3xl font-bold text-center mb-6 text-purple-600">⚡ Performance Comparison</h2>
          <div className="grid grid-cols-2 gap-8 max-w-2xl mx-auto">
            <div className="text-center">
              <h3 className="text-xl font-semibold mb-3">REST/JSON</h3>
              <div className="h-16 bg-gradient-to-r from-red-400 to-red-600 rounded-lg flex items-center justify-center text-white font-bold text-lg">
                100ms
              </div>
            </div>
            <div className="text-center">
              <h3 className="text-xl font-semibold mb-3">gRPC</h3>
              <div className="h-16 bg-gradient-to-r from-green-400 to-green-600 rounded-lg flex items-center justify-center text-white font-bold text-lg w-1/5">
                20ms
              </div>
            </div>
          </div>
          <p className="text-center text-xl font-bold text-purple-600 mt-6">🚀 gRPC is 5-10x faster than REST/JSON!</p>
        </div>
        
        {/* Footer */}
        <footer className="mt-8 text-center text-white">
          <p className="text-lg">Built with Next.js + Go + gRPC + HTTP/2 | Real-time Communication</p>
        </footer>
      </div>
    </div>
  );
}
