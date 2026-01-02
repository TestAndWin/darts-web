import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';

export default function ActiveGame({ gameId, onExit }) {
  const [game, setGame] = useState(null);
  const [users, setUsers] = useState({});
  const [multiplier, setMultiplier] = useState(1);
  const [sending, setSending] = useState(false);
  const [gameStats, setGameStats] = useState(null);
  const [currentTurnThrows, setCurrentTurnThrows] = useState([]);

  const loadUsers = useCallback(async () => {
    const u = await api.getUsers();
    const map = {};
    u.forEach(user => map[user.id] = user.name);
    setUsers(map);
  }, []);

  const loadGame = useCallback(async () => {
    try {
      const g = await api.getGame(gameId);
      setGame(g);
    } catch (e) { console.error(e); }
  }, [gameId]);

  const loadGameStatistics = useCallback(async () => {
    try {
      const stats = await api.getGameStatistics(gameId);
      setGameStats(stats);
    } catch (e) {
      console.error('Failed to load statistics:', e);
    }
  }, [gameId]);

  useEffect(() => {
    // Initial Load
    loadGame();
    loadUsers();
  }, [loadGame, loadUsers]);

  useEffect(() => {
    // Load statistics when game finishes
    if (game?.status === 'FINISHED' && !gameStats) {
      loadGameStatistics();
    }
  }, [game?.status, gameStats, loadGameStatistics]);

  useEffect(() => {
    // Clear current turn throws when player changes
    if (game?.current_turn) {
      setCurrentTurnThrows([]);
    }
  }, [game?.current_turn?.player_index]);

  const formatThrow = (throwData) => {
    const { points, multiplier } = throwData;

    // Miss
    if (points === 0) return 'Miss';

    // Bull
    if (points === 25) {
      return multiplier === 2 ? 'Bulls Eye' : 'Bull';
    }

    // Regular throws
    if (multiplier === 3) return `T${points}`;
    if (multiplier === 2) return `D${points}`;
    return `${points}`;
  };

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

      // Add throw to current turn throws
      setCurrentTurnThrows(prev => [...prev, { points: point, multiplier: realMultiplier }]);

      setMultiplier(1); // Reset multiplier after throw
      loadGame(); // Refresh immediately
    } catch (e) {
      alert("Error processing throw: " + e.message);
    } finally {
      setSending(false);
    }
  }

  if (!game) return <div className="p-8 text-center">Loading game...</div>;

  if (game.status === 'FINISHED') {
    return (
      <div className="max-w-4xl mx-auto p-8">
        {/* Winner Announcement */}
        <div className="bg-gradient-to-r from-darts-gold to-yellow-500 rounded-xl p-8 text-center mb-8 shadow-2xl">
          <h2 className="text-4xl font-black text-slate-900 mb-2">Game Over!</h2>
          <p className="text-3xl font-bold text-slate-800">
            {users[game.winner_id]} Wins!
          </p>
        </div>

        {/* Statistics Section */}
        {gameStats ? (
          <div className="bg-white rounded-xl shadow-lg p-6 mb-6">
            <h3 className="text-2xl font-bold text-slate-800 mb-6 text-center">
              Game Statistics
            </h3>

            {/* Overall Stats Cards */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
              {gameStats.players.map(player => (
                <div key={player.user_id} className="border-2 border-slate-200 rounded-lg p-6">
                  <h4 className="text-xl font-bold text-slate-700 mb-4">
                    {player.user_name}
                  </h4>
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-slate-600">3-Dart Average:</span>
                      <span className="font-bold text-darts-blue">
                        {player.overall_stats.average_3_dart.toFixed(2)}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-600">Total Points:</span>
                      <span className="font-bold">{player.overall_stats.total_points}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-600">Total Throws:</span>
                      <span className="font-bold">{player.overall_stats.total_throws}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* Per-Set Breakdown (if more than 1 set) */}
            {gameStats.total_sets_played > 1 && (
              <div>
                <h4 className="text-lg font-bold text-slate-700 mb-4">Set-by-Set Breakdown</h4>
                {gameStats.players.map(player => (
                  <div key={player.user_id} className="mb-6">
                    <h5 className="font-semibold text-slate-600 mb-2">{player.user_name}</h5>
                    <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                      {player.set_stats.map(set => (
                        <div
                          key={set.set_number}
                          className={`p-4 rounded-lg border-2 ${set.won_set
                              ? 'bg-green-50 border-green-300'
                              : 'bg-slate-50 border-slate-200'
                            }`}
                        >
                          <div className="text-sm font-bold text-slate-700 mb-2">
                            Set {set.set_number} {set.won_set && '✓'}
                          </div>
                          <div className="text-xs space-y-1">
                            <div>Avg: {set.average_3_dart.toFixed(1)}</div>
                            <div>{set.total_points} pts</div>
                            <div>{set.total_throws} throws</div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        ) : (
          <div className="text-center text-slate-500 mb-6">Loading statistics...</div>
        )}

        {/* Back Button */}
        <div className="text-center">
          <button
            onClick={onExit}
            className="px-8 py-4 bg-darts-blue text-white text-lg font-bold rounded-lg hover:bg-blue-700 transition shadow-lg"
          >
            Back to Menu
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto pb-2 px-2 landscape:md:flex landscape:md:h-[calc(100vh-120px)] landscape:md:gap-3 landscape:md:p-2 landscape:md:pb-2 landscape:md:mx-5 landscape:md:max-w-none">
      {/* Header Info */}
      <div className="flex justify-between items-center mb-3 sm:mb-6 bg-white p-3 sm:p-4 rounded-xl shadow-sm w-full landscape:md:hidden">
        <div className="flex flex-col gap-1">
          <div className="text-sm sm:text-base text-slate-500 font-semibold">
            Matches (Best of {game.settings.best_of_sets})
          </div>
          <div className="text-xs sm:text-sm text-slate-400">
            {game.settings.double_out ? '🎯 Double Out' : 'Straight Out'}
          </div>
        </div>
        <button onClick={onExit} className="text-xs sm:text-sm text-slate-400 hover:text-red-500">Exit Game</button>
      </div>

      {/* Left Panel: Player Scores */}
      <div className="landscape:md:w-[30%] landscape:md:flex landscape:md:flex-col landscape:md:overflow-y-auto">
        {/* Compact header for landscape mode */}
        <div className="hidden landscape:md:block mb-3">
          <div className="bg-white p-3 rounded-xl shadow-sm">
            <div className="flex justify-between items-center mb-2">
              <div>
                <div className="text-sm font-semibold text-slate-500">
                  Best of {game.settings.best_of_sets}
                </div>
                <div className="text-xs text-slate-400">
                  {game.settings.double_out ? '🎯 Double Out' : 'Straight Out'}
                </div>
              </div>
              <button
                onClick={onExit}
                className="text-xs text-slate-400 hover:text-red-500 px-3 py-1 rounded hover:bg-slate-100"
              >
                Exit Game
              </button>
            </div>
          </div>
        </div>

        {/* Scoreboard */}
        <div className="grid grid-cols-2 gap-2 sm:gap-4 mb-4 sm:mb-8 landscape:md:grid-cols-1 landscape:md:mb-0 landscape:md:gap-3 w-full">
          {game.players.map((p, idx) => {
            const isCurrent = idx === game.current_turn.player_index;
            return (
              <div key={p.user_id} className={`relative p-3 sm:p-6 rounded-2xl border-2 transition-all duration-300 landscape:md:p-4 landscape:md:h-[120px] ${isCurrent ? 'bg-darts-blue text-white border-darts-blue shadow-lg sm:scale-105 landscape:md:scale-100 z-10' : 'bg-white text-slate-800 border-slate-100'}`}>
                <div className="flex justify-between items-center mb-1 sm:mb-2">
                  <div className="flex items-center gap-2 flex-1 min-w-0">
                    <span className="text-base sm:text-xl font-bold truncate">{users[p.user_id]}</span>
                    {isCurrent && (
                      <div className="flex gap-1">
                        {[...Array(3)].map((_, i) => (
                          <div key={i} className={`w-2 h-2 sm:w-3 sm:h-3 rounded-full ${i < game.current_turn.throw_number ? 'bg-darts-gold' : 'bg-white/20'}`} />
                        ))}
                      </div>
                    )}
                    {isCurrent && currentTurnThrows.length > 0 && (
                      <span className="text-xs sm:text-sm opacity-80 ml-2">
                        {currentTurnThrows.map(formatThrow).join(', ')}
                      </span>
                    )}
                  </div>
                  <div className="text-xs sm:text-sm opacity-80 ml-2 flex-shrink-0">Sets: {p.sets_won}</div>
                </div>
                <div className="text-4xl sm:text-6xl font-black mb-1 sm:mb-2 text-center">
                  {p.current_points}
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Control Pad */}
      <div className="fixed bottom-0 left-0 right-0 flex justify-center pointer-events-none px-4 md:px-8 landscape:md:!px-0 landscape:md:static landscape:md:block landscape:md:w-[70%] landscape:md:h-fit landscape:md:self-start landscape:md:pointer-events-auto">
        <div className="w-full max-w-4xl px-2 pb-2 landscape:md:max-w-none landscape:md:px-0 landscape:md:pb-0">
          <div className="bg-slate-800 rounded-3xl shadow-2xl pointer-events-auto">
            <div className="px-2 py-3 sm:py-4 landscape:md:px-4 landscape:md:pt-3 landscape:md:pb-3 landscape:md:flex landscape:md:flex-col">
              {/* Multipliers */}
              <div className="flex gap-3 sm:gap-4 justify-center mb-3 sm:mb-4 landscape:md:gap-8 landscape:md:mb-6">
                <button
                  onClick={() => setMultiplier(multiplier === 2 ? 1 : 2)}
                  className={`px-4 sm:px-6 py-1.5 sm:py-2 rounded-full font-bold text-sm sm:text-base transition landscape:md:px-16 landscape:md:py-5 landscape:md:text-2xl ${multiplier === 2 ? 'bg-darts-gold text-slate-900' : 'bg-slate-700 text-slate-300'}`}
                >
                  DOUBLE
                </button>
                <button
                  onClick={() => setMultiplier(multiplier === 3 ? 1 : 3)}
                  className={`px-4 sm:px-6 py-1.5 sm:py-2 rounded-full font-bold text-sm sm:text-base transition landscape:md:px-16 landscape:md:py-5 landscape:md:text-2xl ${multiplier === 3 ? 'bg-darts-gold text-slate-900' : 'bg-slate-700 text-slate-300'}`}
                >
                  TRIPLE
                </button>
              </div>

              {/* Numbers */}
              <div className="grid grid-cols-5 gap-1.5 sm:gap-2 landscape:md:gap-4">
                {/* 1-20 in standard layout or grid? Grid is easier for touch. */}
                {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20].map(n => (
                  <button
                    key={n}
                    onClick={() => handleThrow(n)}
                    disabled={sending}
                    className="bg-slate-700 text-white font-bold text-lg sm:text-xl py-3 sm:py-4 rounded-lg active:scale-95 transition hover:bg-slate-600 disabled:opacity-50 landscape:md:text-4xl landscape:md:py-4 landscape:md:px-4"
                  >
                    {n}
                  </button>
                ))}
              </div>

              {/* Zero / Bull / Bulls Eye */}
              <div className="grid grid-cols-3 gap-1.5 sm:gap-2 mt-1.5 sm:mt-2 landscape:md:gap-4 landscape:md:mt-6">
                <button onClick={() => handleThrow(0)} className="bg-red-900/50 text-red-200 font-bold text-sm sm:text-base py-2 sm:py-3 rounded-lg hover:bg-red-900/70 landscape:md:text-2xl landscape:md:py-4">MISS</button>
                <button onClick={() => handleThrow(25, 1)} className="bg-green-700/50 text-green-200 font-bold text-sm sm:text-base py-2 sm:py-3 rounded-lg hover:bg-green-700/70 landscape:md:text-2xl landscape:md:py-4">
                  <div className="text-xs opacity-75 landscape:md:text-base">25</div>
                  <div>BULL</div>
                </button>
                <button onClick={() => handleThrow(25, 2)} className="bg-green-900/70 text-green-200 font-bold text-sm sm:text-base py-2 sm:py-3 rounded-lg hover:bg-green-900 border-2 border-green-400/30 landscape:md:text-2xl landscape:md:py-4">
                  <div className="text-xs opacity-75 landscape:md:text-base">50</div>
                  <div>BULLS EYE</div>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Spacer for fixed bottom */}
      <div className="h-[420px] sm:h-80 landscape:md:hidden"></div>
    </div>
  )
}
