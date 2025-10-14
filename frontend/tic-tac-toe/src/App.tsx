import { useState } from 'react'
import './App.css'
import TicTacToe from './component/ticTacToe'
import JoinRoom from './component/joinRoom'

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  const handleJoin = () => {
    setIsAuthenticated(true)
  }

  return (
    <div className='w-full h-screen'>
      {!isAuthenticated ? (
        <JoinRoom onJoin={handleJoin} />
      ) : (
        <TicTacToe />
      )}
    </div>
  )
}

export default App
