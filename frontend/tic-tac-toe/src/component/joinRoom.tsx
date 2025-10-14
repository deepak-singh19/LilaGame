import React, { useState, useEffect } from 'react'
import { nakamaService, GameState } from '../services/nakama'

interface JoinRoomProps {
  onJoin: () => void;
}

const JoinRoom: React.FC<JoinRoomProps> = ({ onJoin }) => {
  const [name, setName] = useState('')
  const [isConnecting, setIsConnecting] = useState(false)
  const [error, setError] = useState('')
  const [gameState, setGameState] = useState<GameState>(GameState.DISCONNECTED)

  useEffect(() => {
    // Set up Nakama service callbacks
    nakamaService.setOnStateChange(setGameState)
    nakamaService.setOnError(setError)

    return () => {
      // Cleanup on unmount - don't disconnect automatically
      // nakamaService.disconnect()
    }
  }, [])

  const handleConnect = async () => {
    if (!name.trim()) {
      setError('Please enter your name')
      return
    }

    try {
      setIsConnecting(true)
      setError('')
      
      // Generate a unique device ID (6-128 bytes, alphanumeric + underscores/hyphens)
      const sanitizedName = name.toLowerCase().replace(/[^a-z0-9]/g, '').substring(0, 8);
      const timestamp = Date.now().toString().slice(-6); // Last 6 digits
      const randomSuffix = Math.random().toString(36).substring(2, 5); // 3 random chars
      const deviceId = `${sanitizedName}_${timestamp}_${randomSuffix}`;
      
      await nakamaService.connect(deviceId, name)
      onJoin()
    } catch (err) {
      setError(`Failed to connect: ${err}`)
    } finally {
      setIsConnecting(false)
    }
  }

  const getStatusMessage = () => {
    switch (gameState) {
      case GameState.CONNECTING:
        return 'Connecting to server...'
      case GameState.CONNECTED:
        return 'Connected! Ready to play.'
      case GameState.MATCHMAKING:
        return 'Looking for opponents...'
      default:
        return 'Enter your name to start playing'
    }
  }

  const getStatusColor = () => {
    switch (gameState) {
      case GameState.CONNECTING:
      case GameState.MATCHMAKING:
        return 'text-yellow-400'
      case GameState.CONNECTED:
        return 'text-green-400'
      default:
        return 'text-amber-50'
    }
  }

  return (
    <div className="flex flex-col justify-center items-center w-full h-screen bg-[#0C141D] px-[100px]">
      <div className="w-4/5 self-start">
        <p className="text-white text-2xl font-bold">Who are you?</p>
        <p className={`text-sm mt-2 ${getStatusColor()}`}>
          {getStatusMessage()}
        </p>
      </div>

      <div className="flex justify-center items-center border-2 w-full h-[200px] border-amber-50 mt-4">
        <input 
          type="text" 
          className="w-4/5 h-2/5 bg-[#1B2837] text-amber-50 text-2xl px-[20px] rounded" 
          value={name} 
          onChange={(e) => setName(e.target.value)}
          placeholder="Enter your name"
          disabled={isConnecting || gameState === GameState.CONNECTED}
          onKeyPress={(e) => e.key === 'Enter' && handleConnect()}
        />
      </div>

      {error && (
        <div className="w-4/5 mt-2 p-2 bg-red-600 text-white text-sm rounded">
          {error}
        </div>
      )}

      <div className='flex w-full justify-end py-[20px]'>
        <button 
          className={`w-[200px] h-[50px] px-[20px] cursor-pointer rounded transition-colors ${
            isConnecting || gameState === GameState.CONNECTED
              ? 'bg-gray-600 cursor-not-allowed'
              : 'bg-[#12BDAC] hover:bg-[#0FA896]'
          }`}
          onClick={handleConnect}
          disabled={isConnecting || gameState === GameState.CONNECTED}
        >
          {isConnecting ? 'Connecting...' : 
           gameState === GameState.CONNECTED ? 'Connected!' : 'Continue'}
        </button>
      </div>
    </div>
  )
}

export default JoinRoom
