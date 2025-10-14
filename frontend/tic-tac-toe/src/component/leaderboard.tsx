import React, { useState, useEffect } from 'react';
import { nakamaService } from '../services/nakama';

interface LeaderboardEntry {
  rank: number;
  username: string;
  score: number;
  games_won: number;
  games_lost: number;
  games_drawn: number;
  win_rate?: number;
}

interface LeaderboardProps {
  onClose: () => void;
}

const Leaderboard: React.FC<LeaderboardProps> = ({ onClose }) => {
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');
  const [activeTab, setActiveTab] = useState<'overall' | 'weekly'>('overall');

  useEffect(() => {
    loadLeaderboard();
  }, [activeTab]);

  const loadLeaderboard = async () => {
    try {
      setLoading(true);
      setError('');
      
      const rpcName = activeTab === 'overall' ? 'get_leaderboard' : 'get_weekly_leaderboard';
      
      const data = await nakamaService.makeRpcCall(rpcName, {});
      console.log("=== Leaderboard Data ===", {
        rpcName,
        data,
        entries: data.entries,
        firstEntry: data.entries?.[0]
      });
      setEntries(data.entries || []);
    } catch (err) {
      setError(`Failed to load leaderboard: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-teal-600 rounded-lg p-6 w-96 max-h-96 overflow-y-auto">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-bold text-white">Leaderboard</h2>
          <button
            onClick={onClose}
            className="text-white hover:text-gray-300 text-2xl"
          >
            ×
          </button>
        </div>

        {/* Tabs */}
        <div className="flex mb-4">
          <button
            onClick={() => setActiveTab('overall')}
            className={`px-4 py-2 rounded-l ${
              activeTab === 'overall'
                ? 'bg-teal-500 text-white'
                : 'bg-teal-700 text-gray-300'
            }`}
          >
            Overall
          </button>
          <button
            onClick={() => setActiveTab('weekly')}
            className={`px-4 py-2 rounded-r ${
              activeTab === 'weekly'
                ? 'bg-teal-500 text-white'
                : 'bg-teal-700 text-gray-300'
            }`}
          >
            Weekly
          </button>
        </div>

        {/* Content */}
        {loading ? (
          <div className="text-center text-white">Loading...</div>
        ) : error ? (
          <div className="text-center text-red-300">{error}</div>
        ) : entries.length === 0 ? (
          <div className="text-center text-gray-300">No entries yet</div>
        ) : (
          <div className="space-y-2">
            {entries.map((entry, index) => (
              <div
                key={index}
                className="flex justify-between items-center p-2 bg-teal-700 rounded"
              >
                <div className="flex items-center space-x-3">
                  <span className="font-bold text-yellow-400">#{entry.rank}</span>
                  <span className="text-white">{entry.username}</span>
                </div>
                <div className="text-right text-sm text-gray-300">
                  <div>Score: {entry.score}</div>
                  <div>{entry.games_won}W / {entry.games_lost}L / {entry.games_drawn}D</div>
                  {entry.win_rate !== undefined && (
                    <div className="text-xs text-gray-400">
                      Win Rate: {entry.win_rate.toFixed(1)}%
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}

        <div className="mt-4 text-center">
          <button
            onClick={loadLeaderboard}
            className="px-4 py-2 bg-teal-500 text-white rounded hover:bg-teal-400 transition-colors"
          >
            Refresh
          </button>
        </div>
      </div>
    </div>
  );
};

export default Leaderboard;
