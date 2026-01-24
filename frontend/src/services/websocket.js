// src/services/websocket.js

import { usePlayerStore } from '@/stores/player';

let socket = null;
let reconnectTimeoutId = null;

export const websocketService = {
  // --- MODIFIED: The connect function now accepts username and password ---
  connect(username, password) {
    // If credentials are not provided, we cannot connect.
    if (!username || !password) {
      console.error('WebSocket connection attempt failed: username or password not provided.');
      return;
    }

    // Prevent duplicate connections
    if (socket && socket.readyState === WebSocket.OPEN) {
      console.log('WebSocket is already connected.');
      return;
    }

    // Clear any pending reconnect attempts before creating a new connection
    if (reconnectTimeoutId) {
      clearTimeout(reconnectTimeoutId);
      reconnectTimeoutId = null;
    }

    // --- NEW: Build the authenticated WebSocket URL ---
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsHost = window.location.host; // Use the current host, Vite will proxy this
    const wsUrl = `${wsProtocol}//${wsHost}/ws?username=${encodeURIComponent(username)}&password=${encodeURIComponent(password)}`;

    console.log('Connecting to WebSocket with URL:', wsUrl);
    socket = new WebSocket(wsUrl);

    socket.onopen = () => {
      console.log('WebSocket connected');
    };

    socket.onmessage = (event) => {
      try {
        const state = JSON.parse(event.data);
        const playerStore = usePlayerStore();
        playerStore.setGlobalState(state);
      } catch (e) {
        console.error("Error parsing WebSocket message:", e);
      }
    };

    socket.onclose = () => {
      console.log('WebSocket disconnected. Attempting to reconnect...');
      socket = null; // Clear the old socket instance
      // --- MODIFIED: Use the store to re-initiate the connection process ---
      // This ensures we always have the correct credentials.
      reconnectTimeoutId = setTimeout(() => {
        const playerStore = usePlayerStore();
        if (playerStore.isAuthenticated) {
          playerStore.initializeAuthAndConnect();
        }
      }, 5000);
    };

    socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      // The onclose event will typically fire after an error, triggering the reconnect logic.
      if (socket) {
        socket.close();
      }
    };
  },

  disconnect() {
    // When manually disconnecting, clear any scheduled reconnects.
    if (reconnectTimeoutId) {
      clearTimeout(reconnectTimeoutId);
      reconnectTimeoutId = null;
    }
    if (socket) {
      // Prevent the onclose handler from firing the reconnect logic
      socket.onclose = null;
      socket.close();
      socket = null;
      console.log('WebSocket disconnected manually.');
    }
  },
};
