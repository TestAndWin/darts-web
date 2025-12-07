import { useState, useEffect } from 'react';
import { api } from '../services/api';

export default function ActiveGame({ gameId, onExit }) {
  const [game, setGame] = useState(null);
  const [users, setUsers] = useState({});
  const [multiplier, setMultiplier] = useState(1);
  const [sending, setSending] = useState(false);

  useEffect(() => {
    // Initial Load
    loadGame();
    loadUsers();

    // Polling for simple live updates (reduced frequency for better performance)
    // TODO: Replace with SSE or WebSocket for production
    const interval = setInterval(loadGame, 3000);
    return () => clearInterval(interval);
  }, [gameId]); // Refresh if ID changes

  const loadUsers = async () => {
    const u = await api.getUsers();
    const map = {};
    u.forEach(user => map[user.id] = user.name);
    setUsers(map);
  }

  const loadGame = async () => {
    try {
      const g = await api.getGame(gameId);
      setGame(g);
    } catch (e) { console.error(e); }
  }

  const handleThrow = async (point, forcedMultiplier = null) => {
    if (sending || !game) return;

    setSending(true);
    try {
      const currentPlayerIndex = game.current_turn.player_index;
      const currentUserId = game.players[currentPlayerIndex].user_id;

      // Use forced multiplier if provided, otherwise use the state multiplier
      let realMultiplier = forcedMultiplier !== null ? forcedMultiplier : multiplier;

      // Special handling for bull and miss if no forced multiplier
      if (forcedMultiplier === null) {
        if (point === 25) {
          // Bull can be single (25) or double (50), but not triple
          realMultiplier = multiplier > 1 ? 2 : 1;
        } else if (point === 0) {
          // Miss is always 1x0
          realMultiplier = 1;
        }
      }

      await api.sendThrow(game.id, currentUserId, point, realMultiplier);

      setMultiplier(1); // Reset multiplier after throw
      loadGame(); // Refresh immediately
    } catch (e) {
      alert("Error processing throw: " + e.message);
    } finally {
      setSending(false);
    }
  }

  if (!game) return <div className="p-8 text-center">Loading game...</div>;

  const currentPlayer = game.players[game.current_turn.player_index];

  if (game.status === 'FINISHED') {
    return (
      <div className="max-w-xl mx-auto p-8 bg-white rounded-xl shadow-lg text-center">
        <h2 className="text-3xl font-bold text-darts-gold mb-4">Game Finished!</h2>
        <p className="text-2xl mb-8">Winner: {users[game.winner_id] || 'Unknown'}</p>
        <button onClick={onExit} className="px-6 py-3 bg-darts-blue text-white rounded-lg">Back to Menu</button>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto">
      {/* Header Info */}
      <div className="flex justify-between items-center mb-6 bg-white p-4 rounded-xl shadow-sm">
        <div className="text-slate-500 font-semibold">
          Matches (Best of {game.settings.best_of_sets})
        </div>
        <button onClick={onExit} className="text-sm text-slate-400 hover:text-red-500">Exit Game</button>
      </div>

      {/* Scoreboard */}
      <div className="grid grid-cols-2 gap-4 mb-8">
        {game.players.map((p, idx) => {
          const isCurrent = idx === game.current_turn.player_index;
          return (
            <div key={p.user_id} className={`relative p-6 rounded-2xl border-2 transition-all duration-300 ${isCurrent ? 'bg-darts-blue text-white border-darts-blue shadow-lg scale-105 z-10' : 'bg-white text-slate-800 border-slate-100'}`}>
              <div className="flex justify-between items-start mb-2">
                <span className="text-xl font-bold truncate pr-4">{users[p.user_id]}</span>
                <div className="text-sm opacity-80">Sets: {p.sets_won}</div>
              </div>
              <div className="text-6xl font-black mb-2 text-center">
                {p.current_points}
              </div>
              {/* Last Throws visualization could go here */}
              {isCurrent && (
                <div className="flex justify-center gap-1 mt-4">
                  {[...Array(3)].map((_, i) => (
                    <div key={i} className={`w-3 h-3 rounded-full ${i < game.current_turn.throw_number ? 'bg-darts-gold' : 'bg-white/20'}`} />
                  ))}
                </div>
              )}
            </div>
          )
        })}
      </div>

      {/* Control Pad */}
      <div className="bg-slate-800 p-4 rounded-t-3xl shadow-2xl fixed bottom-0 left-0 right-0 max-w-4xl mx-auto">
        {/* Multipliers */}
        <div className="flex gap-4 justify-center mb-4">
          <button
            onClick={() => setMultiplier(multiplier === 2 ? 1 : 2)}
            className={`px-6 py-2 rounded-full font-bold transition ${multiplier === 2 ? 'bg-darts-gold text-slate-900' : 'bg-slate-700 text-slate-300'}`}
          >
            DOUBLE
          </button>
          <button
            onClick={() => setMultiplier(multiplier === 3 ? 1 : 3)}
            className={`px-6 py-2 rounded-full font-bold transition ${multiplier === 3 ? 'bg-darts-gold text-slate-900' : 'bg-slate-700 text-slate-300'}`}
          >
            TRIPLE
          </button>
        </div>

        {/* Numbers */}
        <div className="grid grid-cols-5 gap-2 max-w-lg mx-auto">
          {/* 1-20 in standard layout or grid? Grid is easier for touch. */}
          {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20].map(n => (
            <button
              key={n}
              onClick={() => handleThrow(n)}
              disabled={sending}
              className="bg-slate-700 text-white font-bold text-xl py-4 rounded-lg active:scale-95 transition hover:bg-slate-600 disabled:opacity-50"
            >
              {n}
            </button>
          ))}
        </div>

        {/* Zero / Bull / Bulls Eye */}
        <div className="grid grid-cols-3 gap-2 max-w-lg mx-auto mt-2">
          <button onClick={() => handleThrow(0)} className="bg-red-900/50 text-red-200 font-bold py-3 rounded-lg hover:bg-red-900/70">MISS</button>
          <button onClick={() => handleThrow(25, 1)} className="bg-green-700/50 text-green-200 font-bold py-3 rounded-lg hover:bg-green-700/70">
            <div className="text-xs opacity-75">25</div>
            <div>BULL</div>
          </button>
          <button onClick={() => handleThrow(25, 2)} className="bg-green-900/70 text-green-200 font-bold py-3 rounded-lg hover:bg-green-900 border-2 border-green-400/30">
            <div className="text-xs opacity-75">50</div>
            <div>BULLS EYE</div>
          </button>
        </div>
      </div>

      {/* Spacer for fixed bottom */}
      <div className="h-64"></div>
    </div>
  )
}
