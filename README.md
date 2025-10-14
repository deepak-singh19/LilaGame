# 🎮 Lila Games - Tic-Tac-Toe

A real-time multiplayer Tic-Tac-Toe game built with React, TypeScript, and Nakama game server.

## ✨ Features

- 🎯 **Real-time Multiplayer**: Play against other players in real-time using WebSockets
- 🎮 **Two Game Modes**: 
  - Classic (3x3 board)
  - Advanced (5x5 board)
- 🏆 **Leaderboard System**: Track wins, losses, and scores
- ⚡ **Fast Matchmaking**: Quick opponent matching system
- 📱 **Responsive Design**: Works on desktop and mobile devices
- 🔄 **Auto-reconnection**: Handles connection drops gracefully

## 🏗️ Architecture

### Backend (Go + Nakama)
- **Game Server**: Nakama with custom Go plugins
- **Database**: PostgreSQL for persistent data
- **Matchmaking**: Custom queue-based system
- **Real-time**: WebSocket communication
- **Scoring**: Win (+10), Draw (+1), Loss (-5) points

### Frontend (React + TypeScript)
- **Framework**: React 19 with TypeScript
- **Styling**: Tailwind CSS
- **Build Tool**: Vite
- **Game Client**: Nakama JavaScript SDK

## 🚀 Quick Start

### Prerequisites
- Go 1.19+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL (or use Docker)

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/deepak-singh19/Lila.git
   cd Lila
   ```

2. **Start the backend**
   ```bash
   cd backend
   docker-compose up -d
   ```

3. **Start the frontend**
   ```bash
   cd frontend/tac-tac-toe
   npm install
   npm run dev
   ```

4. **Open your browser**
   - Frontend: http://localhost:5173
   - Nakama Console: http://localhost:7351/console

## 🎮 How to Play

1. **Choose Game Mode**: Select Classic (3x3) or Advanced (5x5)
2. **Find Match**: The system will automatically match you with another player
3. **Make Moves**: Click on empty cells to place your symbol (X or O)
4. **Win Conditions**: Get 3 in a row (Classic) or 5 in a row (Advanced)
5. **View Leaderboard**: Check your ranking and statistics

## 🛠️ Development

### Backend Structure
```
backend/
├── main.go              # Main Nakama module
├── matchmaking.go       # Matchmaking logic
├── match.go            # Game match handler
├── leaderboard.go      # Leaderboard system
├── auth.go             # Authentication
├── Dockerfile          # Container configuration
└── docker-compose.yml  # Local development setup
```

### Frontend Structure
```
frontend/tac-tac-toe/
├── src/
│   ├── component/      # React components
│   │   ├── ticTacToe.tsx
│   │   ├── leaderboard.tsx
│   │   └── joinRoom.tsx
│   ├── services/       # Nakama service
│   │   └── nakama.ts
│   └── App.tsx
├── package.json
└── vite.config.ts
```

## 🚀 Deployment

### Google Cloud Platform (Recommended)
```bash
# Deploy backend
cd backend
./deploy.sh

# Deploy frontend
cd frontend/tac-tac-toe
npm run build
vercel --prod
```

### Docker Compose
```bash
docker-compose -f docker-compose.prod.yml up -d
```

## 🔧 Configuration

### Environment Variables
```bash
# Frontend (.env)
VITE_NAKAMA_SERVER_KEY=defaultkey
VITE_NAKAMA_HOST=your-backend-url
VITE_NAKAMA_PORT=7350
VITE_NAKAMA_SSL=false

# Backend
NAKAMA_DATABASE_URL=postgres://user:pass@host:port/db
```

## 📊 Game Rules

### Classic Mode (3x3)
- First player to get 3 symbols in a row wins
- Rows, columns, or diagonals count
- Draw if board is full with no winner

### Advanced Mode (5x5)
- First player to get 5 symbols in a row wins
- Same win conditions as Classic but on larger board
- More strategic gameplay

## 🏆 Scoring System

- **Win**: +10 points
- **Draw**: +1 point  
- **Loss**: -5 points

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Nakama](https://heroiclabs.com/) - Game server backend
- [React](https://reactjs.org/) - Frontend framework
- [Tailwind CSS](https://tailwindcss.com/) - Styling
- [Vite](https://vitejs.dev/) - Build tool

## 📞 Support

If you have any questions or issues, please open an issue on GitHub.

---

Made with ❤️ by [Deepak Singh](https://github.com/deepak-singh19)
