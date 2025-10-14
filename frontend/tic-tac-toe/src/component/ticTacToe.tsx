import React, { useState, useEffect } from "react";
import { nakamaService, GameState, type MatchInfo } from "../services/nakama";
import Leaderboard from "./leaderboard";

interface Player {
  name: string;
  label: "you" | "opp";
  symbol: "X" | "O";
}

const TicTacToe: React.FC = () => {
  const [boardSize, setBoardSize] = useState(3);
  const [board, setBoard] = useState<string[][]>(
    Array(3).fill(null).map(() => Array(3).fill(""))
  );
  const [currentPlayer, setCurrentPlayer] = useState<"X" | "O">("X");
  const [winner, setWinner] = useState<string | null>(null);
  const [gameState, setGameState] = useState<GameState>(GameState.CONNECTED);
  const [, setMatchInfo] = useState<MatchInfo | null>(null);
  const [error, setError] = useState<string>("");
  const [isMyTurn, setIsMyTurn] = useState(false);
  const [mySymbol, setMySymbol] = useState<"X" | "O" | null>(null);
  const [myUserId, setMyUserId] = useState<string | null>(null);
  const [showModeSelection, setShowModeSelection] = useState(true);
  const [showLeaderboard, setShowLeaderboard] = useState(false);
  const [gameResult, setGameResult] = useState<{
    type: 'win' | 'lose' | 'draw';
    message: string;
  } | null>(null);

  const players: Player[] = [
    { name: "You", label: "you", symbol: mySymbol || "X" },
    { name: "Opponent", label: "opp", symbol: mySymbol === "X" ? "O" : "X" },
  ];

  useEffect(() => {
    // Get current user ID
    const userId = nakamaService.getCurrentUserId();
    setMyUserId(userId);

    // Set up Nakama service callbacks
    nakamaService.setOnStateChange(setGameState);
    nakamaService.setOnMatchUpdate((match) => {
      // Get current user ID fresh each time
      const currentUserId = nakamaService.getCurrentUserId();
      
      console.log("=== Match Update Callback ===", {
        myUserId,
        currentUserId,
        matchPlayers: match.players,
        currentTurn: match.currentTurn,
        match
      });
      
      setMatchInfo(match);
      setBoard(match.board);
      setCurrentPlayer(match.currentTurn as "X" | "O");
      setWinner(match.winner || null);
      setBoardSize(match.size);
      
      // Check for game end and show result notification
      if (match.winner) {
        console.log("=== Game Result Debug ===", {
          winner: match.winner,
          mySymbol: mySymbol,
          currentUserId: currentUserId,
          matchPlayers: match.players
        });
        
        // Find the current player's symbol from the match data
        const currentPlayerSymbol = currentUserId ? 
          match.players.find(p => p.id === currentUserId)?.symbol : null;
        
        console.log("Current player symbol from match:", currentPlayerSymbol);
        
        // Use either mySymbol or the symbol from match data
        const playerSymbol = mySymbol || currentPlayerSymbol;
        const isMyWin = match.winner === playerSymbol;
        
        console.log("Final check - Winner:", match.winner, "Player Symbol:", playerSymbol, "Is my win?", isMyWin);
        
        setGameResult({
          type: isMyWin ? 'win' : 'lose',
          message: isMyWin ? '🎉 You Won!' : '😔 You Lost'
        });
        
        // Auto-hide notification after 5 seconds
        setTimeout(() => setGameResult(null), 5000);
      } else if (match.isDraw) {
        setGameResult({
          type: 'draw',
          message: '🤝 It\'s a Draw!'
        });
        
        // Auto-hide notification after 5 seconds
        setTimeout(() => setGameResult(null), 5000);
      }
      
      // Update myUserId if it's not set
      if (!myUserId && currentUserId) {
        setMyUserId(currentUserId);
      }
      
      // Find my symbol and set turn based on currentTurn
      if (currentUserId) {
        const me = match.players.find(p => p.id === currentUserId);
        console.log("=== Player Assignment ===", {
          currentUserId,
          me,
          currentTurn: match.currentTurn,
          isMyTurn: me ? match.currentTurn === me.symbol : false
        });
        
        if (me && (me.symbol === 'X' || me.symbol === 'O')) {
          setMySymbol(me.symbol);
          setIsMyTurn(match.currentTurn === me.symbol);
        }
      } else {
        console.log("No current user ID available");
      }
    });
    nakamaService.setOnError(setError);

    // Don't auto-start matchmaking, wait for user to select mode

    return () => {
      // Cleanup on unmount
      nakamaService.leaveMatch();
    };
  }, []);

  const findMatch = async (mode: "classic" | "advanced") => {
    try {
      setError("");
      
      // Check if we're actually connected
      if (nakamaService.getGameState() !== GameState.CONNECTED) {
        throw new Error("Not connected to server. Please refresh and try again.");
      }

      // Check if WebSocket is active
      if (!nakamaService.isSocketConnected()) {
        throw new Error("WebSocket disconnected. Please refresh and try again.");
      }

      console.log("Current game state:", nakamaService.getGameState());
      console.log("Session exists:", !!nakamaService.getSession());
      
      setShowModeSelection(false);
      const size = mode === "classic" ? 3 : 5;
      setBoardSize(size);
      setBoard(Array(size).fill(null).map(() => Array(size).fill("")));
      await nakamaService.findMatch(size);
    } catch (err) {
      setError(`Failed to find match: ${err}`);
      setShowModeSelection(true);
    }
  };

  const handleClick = async (row: number, col: number) => {
    // Check if move is valid
    if (board[row][col] || winner || !isMyTurn || gameState !== GameState.IN_MATCH) {
      return;
    }

    try {
      await nakamaService.makeMove(row, col);
    } catch (err) {
      setError(`Failed to make move: ${err}`);
    }
  };

  const handleLeaveMatch = () => {
    nakamaService.leaveMatch();
    setMatchInfo(null);
    setBoard(Array(boardSize).fill(null).map(() => Array(boardSize).fill("")));
    setCurrentPlayer("X");
    setWinner(null);
    setIsMyTurn(false);
    setMySymbol(null);
  };

  const handleNewGame = () => {
    handleLeaveMatch();
    // Reset user ID to ensure it's available for next game
    const userId = nakamaService.getCurrentUserId();
    setMyUserId(userId);
    setGameResult(null); // Clear any previous game result
    setShowModeSelection(true);
  };

  const getStatusMessage = () => {
    if (error) return error;
    if (gameState === GameState.DISCONNECTED) return "Not connected - Please join a room first";
    if (winner) return `Winner: ${winner}`;
    if (gameState === GameState.MATCHMAKING) return "Looking for opponents...";
    if (gameState === GameState.IN_MATCH) {
      if (isMyTurn) return `Your turn (${currentPlayer})`;
      return `Opponent's turn (${currentPlayer})`;
    }
    if (gameState === GameState.CONNECTED) return "Connected - Ready to play!";
    return "Ready to play";
  };

  const getStatusColor = () => {
    if (error) return "text-red-400";
    if (gameState === GameState.DISCONNECTED) return "text-orange-400";
    if (winner) return "text-yellow-400";
    if (gameState === GameState.MATCHMAKING) return "text-yellow-400";
    if (gameState === GameState.IN_MATCH) {
      return isMyTurn ? "text-green-400" : "text-blue-400";
    }
    if (gameState === GameState.CONNECTED) return "text-green-400";
    return "text-white";
  };

  // Show mode selection if not in a match
  if (showModeSelection) {
    return (
      <div className="flex flex-col h-full w-full bg-teal-500 text-white">
        <div className="flex flex-col items-center justify-center h-full">
          <h2 className="text-2xl font-bold mb-8">Choose Game Mode</h2>
          <div className="flex gap-4">
            <button
              onClick={() => findMatch("classic")}
              className="px-8 py-4 bg-teal-600 text-white rounded-lg text-lg font-semibold hover:bg-teal-700 transition-colors"
            >
              Classic (3x3)
            </button>
            <button
              onClick={() => findMatch("advanced")}
              className="px-8 py-4 bg-teal-600 text-white rounded-lg text-lg font-semibold hover:bg-teal-700 transition-colors"
            >
              Advanced (5x5)
            </button>
          </div>
          {error && (
            <div className="mt-4 p-3 bg-red-600 text-white text-sm rounded">
              {error}
            </div>
          )}
          <button 
            onClick={() => setShowLeaderboard(true)}
            className="mt-4 px-6 py-2 bg-yellow-600 text-white rounded-md text-sm hover:bg-yellow-700 transition-colors"
          >
            View Leaderboard
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full w-full bg-teal-500 text-white">
      {/* Status */}
      <div className="px-6 pt-4 text-center">
        <p className={`text-sm ${getStatusColor()}`}>
          {getStatusMessage()}
        </p>
      </div>

      {/* Players */}
      <div className="flex justify-between px-6 pt-2 text-sm">
        <div className="flex flex-col items-center">
          <span className="font-bold">{players[0].name}</span>
          <span className="text-xs">({players[0].symbol})</span>
        </div>
        <div className="flex flex-col items-center">
          <span className="font-bold">{players[1].name}</span>
          <span className="text-xs">({players[1].symbol})</span>
        </div>
      </div>

      {/* Board */}
      <div className="flex justify-center mt-6">
        <div
          className="relative grid"
          style={{
            gridTemplateColumns: `repeat(${boardSize}, 1fr)`,
            gridTemplateRows: `repeat(${boardSize}, 1fr)`,
            width: "320px",
            height: "320px",
          }}
        >
          {/* Vertical lines */}
          {Array.from({ length: boardSize - 1 }).map((_, i) => (
            <div
              key={`v-${i}`}
              className="absolute top-0 bottom-0 w-0.5 bg-white"
              style={{ left: `${((i + 1) / boardSize) * 100}%` }}
            />
          ))}

          {/* Horizontal lines */}
          {Array.from({ length: boardSize - 1 }).map((_, i) => (
            <div
              key={`h-${i}`}
              className="absolute left-0 right-0 h-0.5 bg-white"
              style={{ top: `${((i + 1) / boardSize) * 100}%` }}
            />
          ))}

          {/* Clickable cells */}
          {board.map((row, i) =>
            row.map((cell, j) => (
              <button
                key={`${i}-${j}`}
                onClick={() => handleClick(i, j)}
                disabled={Boolean(cell) || Boolean(winner) || !isMyTurn || gameState !== GameState.IN_MATCH}
                className={`flex items-center justify-center text-3xl font-bold transition-colors ${
                  cell || winner || !isMyTurn || gameState !== GameState.IN_MATCH
                    ? "cursor-not-allowed opacity-50"
                    : "cursor-pointer hover:bg-teal-400"
                }`}
              >
                {cell === "X" && <span className="text-red-700">X</span>}
                {cell === "O" && <span className="text-white">O</span>}
              </button>
            ))
          )}
        </div>
      </div>

      {/* Game Result Notification */}
      {gameResult && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className={`p-8 rounded-lg text-center text-white text-2xl font-bold ${
            gameResult.type === 'win' ? 'bg-green-600' : 
            gameResult.type === 'lose' ? 'bg-red-600' : 
            'bg-yellow-600'
          }`}>
            <div className="mb-4">{gameResult.message}</div>
            <div className="text-sm opacity-75">
              {gameResult.type === 'win' && 'Score +10 points!'}
              {gameResult.type === 'lose' && 'Score -5 points'}
              {gameResult.type === 'draw' && 'Score +1 point'}
            </div>
            <button
              onClick={() => setGameResult(null)}
              className="mt-4 px-4 py-2 bg-white bg-opacity-20 rounded hover:bg-opacity-30 transition-colors"
            >
              Close
            </button>
          </div>
        </div>
      )}

      {/* Controls */}
      <div className="flex justify-center gap-4 mt-6">
        <button 
          onClick={handleLeaveMatch}
          className="px-6 py-2 bg-teal-600 text-gray-200 rounded-md text-sm hover:bg-teal-700 transition-colors"
        >
          Leave Match
        </button>
        <button 
          onClick={handleNewGame}
          className="px-6 py-2 bg-teal-600 text-gray-200 rounded-md text-sm hover:bg-teal-700 transition-colors"
        >
          New Game
        </button>
        <button 
          onClick={() => setShowLeaderboard(true)}
          className="px-6 py-2 bg-yellow-600 text-white rounded-md text-sm hover:bg-yellow-700 transition-colors"
        >
          Leaderboard
        </button>
      </div>

      {/* Leaderboard Modal */}
      {showLeaderboard && (
        <Leaderboard onClose={() => setShowLeaderboard(false)} />
      )}
    </div>
  );
};

export default TicTacToe;
