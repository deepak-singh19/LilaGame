# Multiplayer Tic-Tac-Toe Backend

A complete multiplayer Tic-Tac-Toe game backend built with Nakama and Go, featuring real-time gameplay, matchmaking, leaderboards, and Google Cloud deployment.

## Features

### ✅ Core Requirements
- **Device-based Authentication** - JWT token-based authentication using device IDs
- **WebSocket Communication** - Real-time bidirectional communication
- **Server-Authoritative Gameplay** - All game logic validated on the server
- **Matchmaking System** - Queue-based matchmaking for 2 game modes (Classic 3x3, Advanced 5x5)
- **Leaderboard System** - Player rankings and performance tracking
- **Google Cloud Deployment** - Scalable deployment with Cloud Run and Cloud SQL

### 🎮 Game Features
- **Two Game Modes**:
  - Classic: 3x3 board (traditional tic-tac-toe)
  - Advanced: 5x5 board (extended gameplay)
- **Real-time Multiplayer** - Up to 2 players per match
- **Game State Management** - Server validates all moves and maintains game state
- **Win Detection** - Automatic win/loss/draw detection
- **Player Statistics** - Track wins, losses, draws, and performance

### 🔧 Technical Features
- **Go Plugin Architecture** - Custom Nakama plugins for game logic
- **PostgreSQL Database** - Persistent storage for user data and leaderboards
- **RESTful APIs** - RPC endpoints for authentication, matchmaking, and leaderboards
- **Error Handling** - Comprehensive error handling and logging
- **Scalable Architecture** - Designed for horizontal scaling

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Nakama        │    │   PostgreSQL    │
│   (Web/Mobile)  │◄──►│   Server        │◄──►│   Database      │
│                 │    │                 │    │                 │
│ - WebSocket     │    │ - Go Plugins    │    │ - User Data     │
│ - REST API      │    │ - Matchmaking   │    │ - Leaderboards  │
│ - Authentication│    │ - Game Logic    │    │ - Statistics    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## API Endpoints

### Authentication
- `POST /device_auth` - Authenticate with device ID
- `POST /authenticate` - Authenticate with JWT token

### Matchmaking
- `POST /start_matchmaking` - Start matchmaking for a game mode
- `POST /stop_matchmaking` - Stop current matchmaking
- `GET /matchmaking_status` - Get current matchmaking status

### Game
- `WebSocket /match/{match_id}` - Join a game match
- `POST /move` - Make a move in the game

### Leaderboards
- `GET /get_leaderboard` - Get overall leaderboard
- `GET /get_weekly_leaderboard` - Get weekly leaderboard
- `GET /get_player_stats` - Get player statistics

## WebSocket Messages

### Client → Server
```json
{
  "opcode": 1,
  "data": {
    "row": 0,
    "col": 1
  }
}
```

### Server → Client
```json
{
  "opcode": 2,
  "data": {
    "board": [["X", "O", ""], ["", "X", ""], ["", "", ""]],
    "turn": "O",
    "winner": "",
    "size": 3,
    "mode": "classic",
    "players": {
      "user1": "X",
      "user2": "O"
    }
  }
}
```

## Local Development

### Prerequisites
- Docker and Docker Compose
- Go 1.23.5+
- Nakama 3.26.0+

### Setup
1. Clone the repository
2. Navigate to the backend directory
3. Start the services:
   ```bash
   docker compose up --build
   ```

### Services
- **Nakama Server**: `http://localhost:7350`
- **Nakama Console**: `http://localhost:7351` (admin/password)
- **PostgreSQL**: `localhost:5432`

## Google Cloud Deployment

### Prerequisites
- Google Cloud SDK installed
- Project with billing enabled
- Required APIs enabled

### Deploy
1. Update `deploy.sh` with your project ID
2. Run the deployment script:
   ```bash
   ./deploy.sh
   ```

### Services Used
- **Cloud Run** - Serverless container hosting
- **Cloud SQL** - Managed PostgreSQL database
- **Cloud Build** - CI/CD pipeline
- **Container Registry** - Docker image storage

## Configuration

### Environment Variables
- `NAKAMA_DATABASE_URL` - PostgreSQL connection string
- `NAKAMA_CONSOLE_USERNAME` - Admin console username
- `NAKAMA_CONSOLE_PASSWORD` - Admin console password
- `NAKAMA_SOCKET_SERVER_KEY` - WebSocket server key

### Game Modes
- **Classic**: 3x3 board, traditional rules
- **Advanced**: 5x5 board, extended gameplay

## Testing

### Manual Testing
1. Start the server
2. Use the Nakama console to test RPC functions
3. Use WebSocket clients to test real-time gameplay

### API Testing
```bash
# Authenticate
curl -X POST http://localhost:7350/v2/account/authenticate/device \
  -H "Content-Type: application/json" \
  -d '{"device_id": "test-device-123"}'

# Start matchmaking
curl -X POST http://localhost:7350/v2/rpc/start_matchmaking \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"mode": "classic"}'
```

## Monitoring

### Logs
- Application logs: `docker compose logs nakama`
- Database logs: `docker compose logs postgres`

### Metrics
- Game sessions
- Matchmaking queue length
- Leaderboard updates
- Error rates

## Security

### Authentication
- Device-based authentication
- JWT tokens with expiration
- Secure password hashing

### Data Protection
- Input validation
- SQL injection prevention
- XSS protection

## Performance

### Scalability
- Horizontal scaling with Cloud Run
- Database connection pooling
- Efficient matchmaking algorithms

### Optimization
- Minimal memory footprint
- Fast game state updates
- Optimized database queries

## Troubleshooting

### Common Issues
1. **Protobuf version mismatch**: Ensure Go version matches Nakama runtime
2. **Database connection**: Check PostgreSQL is running and accessible
3. **WebSocket connection**: Verify firewall and network settings

### Debug Mode
Enable debug logging in `local.yml`:
```yaml
logger:
  level: "DEBUG"
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

For issues and questions:
- Create an issue in the repository
- Check the Nakama documentation
- Review the logs for error details
