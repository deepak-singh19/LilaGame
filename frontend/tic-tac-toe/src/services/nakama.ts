import { Client, Session } from "@heroiclabs/nakama-js";
import type { Socket, MatchmakerMatched, MatchData, MatchPresenceEvent } from "@heroiclabs/nakama-js";

// Game state enum
export enum GameState {
  DISCONNECTED = "disconnected",
  CONNECTING = "connecting", 
  CONNECTED = "connected",
  MATCHMAKING = "matchmaking",
  IN_MATCH = "in_match"
}

// Game mode enum
export enum GameMode {
  CLASSIC = "classic",
  ADVANCED = "advanced"
}

// Player info interface
export interface PlayerInfo {
  id: string;
  symbol: string;
}

// Match info interface
export interface MatchInfo {
  id: string;
  size: number;
  players: PlayerInfo[];
  currentPlayer: string;
  currentTurn: string;
  board: string[][];
  winner?: string;
  isDraw?: boolean;
}

// Move data interface
export interface MoveData {
  row: number;
  col: number;
}

// State data interface
export interface StateData {
  board: string[][];
  turn: string;
  winner?: string;
  size: number;
  mode: string;
  players: { [userId: string]: string }; // userID -> symbol
}

// Error data interface
export interface ErrorData {
  msg: string;
}

// Opcodes for match data
export const Opcodes = {
  MOVE: 1,
  STATE: 2,
  ERROR: 3,
  MATCH_FOUND: 4,  // Match found notification
  LEADERBOARD: 5,  // Leaderboard update
  PLAYER_JOIN: 6,  // Player joined match
  PLAYER_LEAVE: 7  // Player left match
};

// Nakama service class
export class NakamaService {
  private client: Client;
  private session: Session | null = null;
  private socket: Socket | null = null;
  private socketConnected = false;
  private gameState: GameState = GameState.DISCONNECTED;
  private currentMatch: MatchInfo | null = null;

  // Event callbacks
  private onStateChange?: (state: GameState) => void;
  private onMatchUpdate?: (match: MatchInfo) => void;
  private onError?: (error: string) => void;

  constructor() {
    // Create client with environment-based configuration
    const serverKey = import.meta.env.VITE_NAKAMA_SERVER_KEY || "defaultkey";
    const host = import.meta.env.VITE_NAKAMA_HOST || "127.0.0.1";
    const port = import.meta.env.VITE_NAKAMA_PORT || "7350";
    
    // Auto-detect SSL requirement: use SSL if explicitly set, or if page is served over HTTPS
    const isHTTPS = window.location.protocol === 'https:';
    const useSSL = import.meta.env.VITE_NAKAMA_SSL === "true" || 
                   (import.meta.env.VITE_NAKAMA_SSL !== "false" && isHTTPS);

    this.client = new Client(serverKey, host, port, useSSL);

    console.log("Nakama client created with config:", {
      serverkey: serverKey,
      host: host,
      port: port,
      useSSL: useSSL,
      isHTTPS: isHTTPS,
      protocol: window.location.protocol
    });
  }

  // Set event callbacks
  setOnStateChange(callback: (state: GameState) => void) {
    this.onStateChange = callback;
  }

  setOnMatchUpdate(callback: (match: MatchInfo) => void) {
    this.onMatchUpdate = callback;
  }

  setOnError(callback: (error: string) => void) {
    this.onError = callback;
  }

  // Get current game state
  getGameState(): GameState {
    return this.gameState;
  }

  // Get current match
  getCurrentMatch(): MatchInfo | null {
    return this.currentMatch;
  }

  // Get current user ID
  getCurrentUserId(): string | null {
    return this.session?.user_id || null;
  }

  // Get socket instance
  getSocket(): Socket | null {
    return this.socket;
  }

  // Check if socket is connected
  isSocketConnected(): boolean {
    return this.socketConnected;
  }

  // Get current session
  getSession(): Session | null {
    return this.session;
  }

  // Make RPC call
  async makeRpcCall(rpcName: string, payload: any = {}): Promise<any> {
    if (!this.session) {
      throw new Error('Not authenticated');
    }
    
    if (!this.socketConnected) {
      throw new Error('Not connected to server');
    }

    const response = await this.client.rpc(this.session, rpcName, payload);
    if (typeof response.payload === 'string') {
      return JSON.parse(response.payload);
    } else {
      return response.payload || {};
    }
  }

  // Join a match
  async joinMatch(matchId: string): Promise<void> {
    try {
      console.log("=== Joining Match ===", { matchId });
      
      if (!this.socket || !this.socketConnected) {
        throw new Error("Socket not connected");
      }

      // Join the match
      const match = await this.socket.joinMatch(matchId);
      console.log("=== Match Joined Successfully ===", { matchId, match });
      
      // Update match info
      this.currentMatch = {
        id: matchId,
        size: 3, // Default size
        players: [], // Will be populated by match data
        currentPlayer: '',
        currentTurn: '',
        board: Array(3).fill(null).map(() => Array(3).fill(''))
      };
      
      console.log("=== Match Info Updated ===", this.currentMatch);
      
    } catch (error) {
      console.error("=== Failed to Join Match ===", { matchId, error });
      throw error;
    }
  }

  // Update state and notify listeners
  private updateState(newState: GameState) {
    console.log("Updating state from", this.gameState, "to", newState);
    this.gameState = newState;
    this.onStateChange?.(newState);
  }

  // Connect to Nakama
  async connect(deviceId: string, username?: string): Promise<void> {
    try {
      console.log("=== Starting Connection Process ===", {
        deviceId,
        username,
        currentState: this.gameState,
        hasSession: !!this.session,
        hasSocket: !!this.socket,
        timestamp: new Date().toISOString()
      });

      this.updateState(GameState.CONNECTING);

      // Try to authenticate with the provided username
      let finalUsername = username;
      let attempts = 0;
      const maxAttempts = 3;

      while (attempts < maxAttempts) {
        try {
          console.log("=== Attempting Authentication ===", {
            attempt: attempts + 1,
            username: finalUsername,
            timestamp: new Date().toISOString()
          });

          this.session = await this.client.authenticateDevice(deviceId, true, finalUsername);

          console.log("=== Authentication Successful ===", {
            userId: this.session.user_id,
            username: this.session.username,
            hasToken: !!this.session.token,
            timestamp: new Date().toISOString()
          });
          break;
        } catch (err: any) {
          attempts++;
          if (err.message?.includes("username") && attempts < maxAttempts) {
            finalUsername = `${username}_${Math.random().toString(36).substr(2, 9)}`;
            console.log("=== Username Conflict, Retrying ===", {
              originalUsername: username,
              newUsername: finalUsername,
              timestamp: new Date().toISOString()
            });
            continue;
          }
          throw err;
        }
      }

      if (!this.session) {
        throw new Error("Failed to create session");
      }

      console.log("=== Creating WebSocket ===");
      // Create socket with default configuration
      this.socket = this.client.createSocket(false);

      console.log("=== Socket Created ===", {
        timestamp: new Date().toISOString()
      });

      // Event handlers will be set up in the connection promise

      console.log("=== Attempting WebSocket Connection ===", {
        hasSession: !!this.session,
        socketExists: !!this.socket,
        timestamp: new Date().toISOString()
      });

      // Connect with explicit promise handling and debug logging
      await new Promise<void>((resolve, reject) => {
        const connectTimeout = setTimeout(() => {
          console.error("=== WebSocket Connection Timeout ===", {
            timestamp: new Date().toISOString()
          });
          reject(new Error("WebSocket connection timeout"));
        }, 10000);

        // Event handlers will be set up after successful connection

        // Start the connection
        console.log("=== Starting WebSocket Connection ===", {
          sessionToken: this.session?.token?.substring(0, 20) + '...',
          timestamp: new Date().toISOString()
        });

        // Connect with session
        this.socket!.connect(this.session!, false)
          .then(() => {
            console.log("=== Socket Connect Promise Resolved ===", {
              timestamp: new Date().toISOString()
            });
            clearTimeout(connectTimeout);
            this.socketConnected = true;
            this.setupSocketHandlers();
            resolve();
          })
          .catch((error: any) => {
            clearTimeout(connectTimeout);
            console.error("=== Socket Connect Promise Rejected ===", {
              error: error.message || error,
              timestamp: new Date().toISOString()
            });
            reject(error);
          });
      });

      this.updateState(GameState.CONNECTED);

    } catch (err) {
      const error = err as Error;
      console.error("=== Connection Process Failed ===", {
        error: error.message || String(error),
        gameState: this.gameState,
        socketState: this.socket ? {
          connected: this.socketConnected
        } : 'NO_SOCKET',
        timestamp: new Date().toISOString()
      });

      // Handle specific error cases
      let errorMessage = `Connection failed: ${error}`;
      if (error instanceof Response) {
        if (error.status === 409) {
          errorMessage = "Username already taken. Please try a different name.";
        } else if (error.status === 400) {
          errorMessage = "Invalid request. Please check your input.";
        } else if (error.status === 401) {
          errorMessage = "Authentication failed. Please try again.";
        }
      }

      this.updateState(GameState.DISCONNECTED);
      throw new Error(errorMessage);
    }
  }

  // Disconnect from Nakama
  async disconnect(): Promise<void> {
    if (this.socket) {
      this.socket.disconnect(false);
      this.socket = null;
    }
    this.session = null;
    this.socketConnected = false;
    this.currentMatch = null;
    this.updateState(GameState.DISCONNECTED);
  }

  // Find a match
  async findMatch(boardSize: number = 3): Promise<void> {
    console.log("=== Starting findMatch ===");
    console.log("Current game state:", this.gameState);
    console.log("Session exists:", !!this.session);
    console.log("Socket exists:", !!this.socket);
    console.log("Socket connected flag:", this.socketConnected);

    if (!this.session) {
      console.error("No session found - throwing error");
      throw new Error("Not connected to Nakama");
    }

    // Check if socket exists and is connected
    if (!this.socket) {
      console.error("No socket found - throwing error");
      throw new Error("WebSocket not initialized");
    }

    // Check if socket is still connected using our flag
    if (!this.socketConnected) {
      console.error("Socket not connected according to our flag");
      console.error("Socket details:", {
        wasEverConnected: this.socketConnected,
        currentGameState: this.gameState
      });

      // Try to reconnect if disconnected
      console.log("Socket is not connected, attempting to reconnect...");
      try {
        // Reset connection flag
        this.socketConnected = false;
        this.updateState(GameState.CONNECTING);

        // Wait for reconnection using promise-based approach
        await this.socket.connect(this.session!, false);
        console.log("Socket reconnected successfully");
        
        // Set connection flag and setup handlers
        this.socketConnected = true;
        this.setupSocketHandlers();
        this.updateState(GameState.CONNECTED);
      } catch (reconnectError) {
        console.error("Failed to reconnect socket:", reconnectError);
        this.updateState(GameState.DISCONNECTED);
        throw new Error("WebSocket disconnected and failed to reconnect. Please refresh and try again.");
      }
    }

    try {
      console.log("=== Starting Matchmaking Process ===", {
        boardSize,
        currentState: this.gameState,
        connected: this.socketConnected
      });

      this.updateState(GameState.MATCHMAKING);

      // Use standard Nakama matchmaking
      const mode = boardSize === 3 ? GameMode.CLASSIC : GameMode.ADVANCED;

      // Verify socket state before matchmaking
      if (!this.socketConnected) {
        console.error("=== Socket Not Ready for Matchmaking ===", {
          connected: this.socketConnected,
          timestamp: new Date().toISOString()
        });
        throw new Error("Socket not ready for matchmaking");
      }

      console.log("=== Starting Matchmaking via RPC ===", {
        mode,
        boardSize,
        timestamp: new Date().toISOString()
      });

      // Call the start_matchmaking RPC
      const response = await this.socket.rpc("start_matchmaking", JSON.stringify({
        mode: mode,
        boardSize: boardSize
      }));

      console.log("=== Matchmaking RPC Response ===", {
        response: response,
        timestamp: new Date().toISOString()
      });

      // Parse the response
      const matchmakingResponse = JSON.parse(response.payload || '{}');
      console.log("=== Parsed Matchmaking Response ===", matchmakingResponse);

      // Check if we got a match ID (match was created) or a ticket (waiting for opponent)
      if (matchmakingResponse.ticket && matchmakingResponse.ticket.includes('.nakama')) {
        // This is a match ID - join the match immediately
        console.log("=== Match Created! Joining Match ===", {
          matchId: matchmakingResponse.ticket,
          mode: matchmakingResponse.mode
        });
        
        // Update state to in_match
        this.updateState(GameState.IN_MATCH);
        
        // Join the match
        await this.joinMatch(matchmakingResponse.ticket);
      } else {
        // This is a ticket - waiting for opponent
        console.log("=== Waiting for Opponent ===", {
          ticket: matchmakingResponse.ticket,
          mode: matchmakingResponse.mode
        });
        
        // The actual matchmaking will be handled by the matchmaker matched event
        console.log("=== Matchmaking Request Sent ===", {
          mode,
          boardSize,
          connected: this.socketConnected,
          timestamp: new Date().toISOString()
        });
      }

    } catch (err) {
      const error = err as Error;
      console.error("=== Matchmaking Process Failed ===", {
        error: error.message || String(error),
        gameState: this.gameState,
        connected: this.socketConnected,
        timestamp: new Date().toISOString()
      });

      this.onError?.(`Matchmaking failed: ${error}`);
      this.updateState(GameState.CONNECTED);
      throw error;
    }
  }

  // Make a move using WebSocket
  async makeMove(row: number, col: number): Promise<void> {
    if (!this.socket || !this.currentMatch) {
      throw new Error("Not in a match");
    }

    if (!this.socketConnected) {
      throw new Error("Socket not connected");
    }

    const moveData: MoveData = { row, col };
    console.log("=== Sending Move ===", { moveData, matchId: this.currentMatch.id });
    
    try {
      // Send match data using Nakama's sendMatchState method
      await this.socket.sendMatchState(
        this.currentMatch.id,
        Opcodes.MOVE,
        JSON.stringify(moveData)
      );
      console.log("=== Move Sent Successfully ===", moveData);
    } catch (error) {
      console.error("=== Failed to Send Move ===", { error, moveData });
      throw new Error(`Failed to send move: ${error}`);
    }
  }

  // Leave current match
  leaveMatch(): void {
    if (this.currentMatch) {
      console.log("Leaving match:", this.currentMatch.id);
      this.currentMatch = null;
      this.updateState(GameState.CONNECTED);
    }
  }

  // Set up socket event handlers
  private setupSocketHandlers(): void {
    if (!this.socket) return;

    console.log("Setting up WebSocket event handlers");

    // Note: Nakama socket connection is handled by promise, not onopen event

    // Socket error
    this.socket.onerror = (error: any) => {
      console.error("=== WebSocket Error Event ===", {
        error: error.message || error,
        wasConnected: this.socketConnected,
        currentGameState: this.gameState,
        hasSession: !!this.session,
        timestamp: new Date().toISOString()
      });
      this.onError?.(`Connection error: ${error}`);
    };

    // Matchmaker matched event
    this.socket.onmatchmakermatched = (matchmakerMatched: MatchmakerMatched) => {
      console.log("=== Matchmaker Matched ===", {
        matchId: matchmakerMatched.match_id,
        matchmakerMatched: matchmakerMatched,
        connected: this.socketConnected,
        currentGameState: this.gameState,
        timestamp: new Date().toISOString()
      });

      if (this.socket && this.socketConnected) {
        console.log("=== Attempting to Join Match ===");
        
        // Update state to in_match before joining
        this.updateState(GameState.IN_MATCH);
        
        this.socket.joinMatch(matchmakerMatched.match_id)
          .then((match) => {
            console.log("=== Successfully Joined Match ===", { matchId: matchmakerMatched.match_id, match });
            
            // Update current match info
            this.currentMatch = {
              id: matchmakerMatched.match_id,
              size: 3, // Default size
              players: [], // Will be populated by match data
              currentPlayer: '',
              currentTurn: '',
              board: Array(3).fill(null).map(() => Array(3).fill(''))
            };
            
            console.log("=== Match Info Updated ===", this.currentMatch);
            this.onMatchUpdate?.(this.currentMatch);
          })
          .catch((error: any) => {
            console.error("=== Failed to Join Match ===", {
              error: error.message || error,
              matchId: matchmakerMatched.match_id,
              timestamp: new Date().toISOString()
            });
            this.onError?.(`Failed to join match: ${error.message || error}`);
            this.updateState(GameState.CONNECTED);
          });
      } else {
        console.error("=== Cannot Join Match - Socket Not Ready ===", {
          connected: this.socketConnected,
          timestamp: new Date().toISOString()
        });
        this.onError?.("Cannot join match - socket not ready");
      }
    };

    // Match data event
    this.socket.onmatchdata = (matchData: MatchData) => {
      console.log("=== Match Data Received ===", {
        matchId: matchData.match_id,
        opCode: matchData.op_code,
        data: matchData.data,
        presence: matchData.presence,
        timestamp: new Date().toISOString()
      });

      this.handleMatchData(matchData);
    };

    // Match presence event
    this.socket.onmatchpresence = (matchPresence: MatchPresenceEvent) => {
      console.log("=== Match Presence Event ===", {
        matchId: matchPresence.match_id,
        joins: matchPresence.joins,
        leaves: matchPresence.leaves,
        timestamp: new Date().toISOString()
      });

      // Handle player joins/leaves
      if (matchPresence.joins && matchPresence.joins.length > 0) {
        console.log("Players joined:", matchPresence.joins.map(p => p.username));
      }
      if (matchPresence.leaves && matchPresence.leaves.length > 0) {
        console.log("Players left:", matchPresence.leaves.map(p => p.username));
      }
    };

    // Notification handler
    this.socket.onnotification = (notification: any) => {
      console.log("=== Notification Received ===", {
        notification: notification,
        timestamp: new Date().toISOString()
      });

      try {
        // Handle different notification content types
        let data;
        if (typeof notification.content === 'string') {
          data = JSON.parse(notification.content);
        } else if (typeof notification.content === 'object') {
          data = notification.content;
        } else {
          console.log("=== Unknown Notification Content Type ===", {
            type: typeof notification.content,
            content: notification.content,
            timestamp: new Date().toISOString()
          });
          return;
        }

        if (data.type === "match_created") {
          console.log("=== Match Created Notification ===", {
            matchId: data.match_id,
            mode: data.mode,
            timestamp: new Date().toISOString()
          });

          // Update state to in_match
          this.updateState(GameState.IN_MATCH);
          
          // Join the match
          this.joinMatch(data.match_id);
        }
      } catch (error) {
        console.error("=== Failed to Parse Notification ===", {
          error: error,
          notification: notification,
          timestamp: new Date().toISOString()
        });
      }
    };

    // Note: Nakama socket doesn't have onopen event, connection is handled by promise
  }

  // Handle incoming match data
  private handleMatchData(matchData: MatchData): void {
    try {
      const rawData = new TextDecoder().decode(matchData.data);
      console.log("=== Raw Match Data ===", {
        matchId: matchData.match_id,
        opCode: matchData.op_code,
        rawData: rawData,
        dataLength: matchData.data.length,
        hasCurrentMatch: !!this.currentMatch,
        currentMatchId: this.currentMatch?.id,
        timestamp: new Date().toISOString()
      });
      
      const data = JSON.parse(rawData);
      console.log("=== Parsed Match Data ===", {
        matchId: matchData.match_id,
        opCode: matchData.op_code,
        data: data,
        hasCurrentMatch: !!this.currentMatch,
        currentMatchId: this.currentMatch?.id,
        timestamp: new Date().toISOString()
      });

      console.log("=== Processing Opcode ===", {
        opCode: matchData.op_code,
        opCodeType: typeof matchData.op_code,
        MATCH_FOUND: Opcodes.MATCH_FOUND,
        isMatchFound: matchData.op_code === Opcodes.MATCH_FOUND
      });

      switch (matchData.op_code) {
        case Opcodes.MOVE:
          this.handleMove(data);
          break;
        case Opcodes.STATE:
          this.handleStateUpdate(data);
          break;
        case Opcodes.ERROR:
          this.handleError(data);
          break;
        case Opcodes.MATCH_FOUND:
          console.log("=== Handling MATCH_FOUND ===", data);
          this.handleMatchFound(data);
          break;
        case Opcodes.LEADERBOARD:
          this.handleLeaderboard(data);
          break;
        case Opcodes.PLAYER_JOIN:
          this.handlePlayerJoin(data);
          break;
        case Opcodes.PLAYER_LEAVE:
          this.handlePlayerLeave(data);
          break;
        default:
          console.log("Unknown opcode:", matchData.op_code);
      }
    } catch (error) {
      console.error("Failed to parse match data:", error);
    }
  }

  // Handle move data
  private handleMove(data: MoveData): void {
    console.log("Move received:", data);
    // Update game state based on move
    if (this.currentMatch) {
      // This would update the board state
      this.onMatchUpdate?.(this.currentMatch);
    }
  }

  // Handle state update from server
  private handleStateUpdate(stateData: StateData): void {
    if (!this.currentMatch) {
      console.log("No current match to update - creating temporary match");
      // Create a temporary match if we don't have one yet
      this.currentMatch = {
        id: 'temp',
        size: stateData.size || 3,
        players: [],
        currentPlayer: '',
        currentTurn: '',
        board: Array(stateData.size || 3).fill(null).map(() => Array(stateData.size || 3).fill(''))
      };
    }

    console.log("=== State Update Received ===", {
      turn: stateData.turn,
      board: stateData.board,
      winner: stateData.winner,
      size: stateData.size,
      mode: stateData.mode,
      players: stateData.players,
      timestamp: new Date().toISOString()
    });

    // Update current match with new state
    this.currentMatch.currentPlayer = stateData.turn || '';
    this.currentMatch.currentTurn = stateData.turn || '';
    this.currentMatch.board = stateData.board || Array(3).fill(null).map(() => Array(3).fill(''));
    this.currentMatch.winner = stateData.winner || undefined;
    this.currentMatch.size = stateData.size || 3;
    
    // Convert players map to array format
    if (stateData.players) {
      this.currentMatch.players = Object.entries(stateData.players).map(([userId, symbol]) => ({
        id: userId,
        symbol: symbol
      }));
      console.log("=== Players Converted ===", {
        originalPlayers: stateData.players,
        convertedPlayers: this.currentMatch.players,
        currentUserId: this.getCurrentUserId()
      });
    }

    // Check for game end
    const isDraw = !stateData.winner && stateData.board && 
      stateData.board.every(row => row.every(cell => cell !== ''));

    this.currentMatch.isDraw = isDraw;

    console.log("=== Match Updated ===", {
      currentPlayer: this.currentMatch.currentPlayer,
      currentTurn: this.currentMatch.currentTurn,
      board: this.currentMatch.board,
      winner: this.currentMatch.winner,
      players: this.currentMatch.players,
      isDraw: this.currentMatch.isDraw
    });

    // Notify listeners
    this.onMatchUpdate?.(this.currentMatch);

    // Check for game end
    if (stateData.winner) {
      console.log("Game ended - Winner:", stateData.winner);
      this.updateState(GameState.CONNECTED);
    } else if (isDraw) {
      console.log("Game ended - Draw");
      this.updateState(GameState.CONNECTED);
    }
  }

  // Handle error from server
  private handleError(errorData: ErrorData): void {
    console.error("Game error:", errorData.msg);
    this.onError?.(errorData.msg);
  }

  // Handle match found notification
  private handleMatchFound(data: any): void {
    console.log("=== Match Found ===", data);
    // This is just a notification that a match was found
    // The actual game state will come via STATE opcode
  }

  // Handle leaderboard update
  private handleLeaderboard(data: any): void {
    console.log("=== Leaderboard Update ===", data);
    // Handle leaderboard updates if needed
  }

  // Handle player join
  private handlePlayerJoin(data: any): void {
    console.log("=== Player Joined ===", data);
    if (this.currentMatch && data.player) {
      this.currentMatch.players.push(data.player);
      this.onMatchUpdate?.(this.currentMatch);
    }
  }

  // Handle player leave
  private handlePlayerLeave(data: any): void {
    console.log("=== Player Left ===", data);
    if (this.currentMatch && data.playerId) {
      this.currentMatch.players = this.currentMatch.players.filter(p => p.id !== data.playerId);
      this.onMatchUpdate?.(this.currentMatch);
    }
  }
}

// Export singleton instance
export const nakamaService = new NakamaService();
export default nakamaService;