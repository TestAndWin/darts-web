
import { useState } from 'react'
import './App.css'
import GameSetup from './components/GameSetup'
import ActiveGame from './components/ActiveGame'
import UserStats from './components/UserStats'

function App() {
  const [view, setView] = useState('home'); // home, setup, game, stats
  const [activeGameId, setActiveGameId] = useState(null);

  const startSetup = () => setView('setup');
  const showStats = () => setView('stats');

  const handleGameStarted = (game) => {
    setActiveGameId(game.id);
    setView('game');
  }

  const handleExit = () => {
    setActiveGameId(null);
    setView('home');
  }

  return (
    <div className="min-h-screen bg-slate-100 font-sans text-slate-900">
      {/* Navbar/Header */}
      <nav className="bg-white border-b border-slate-200 px-6 py-4 flex items-center justify-between shadow-sm">
        <div className="flex items-center gap-2 cursor-pointer" onClick={handleExit}>
          <div className="w-8 h-8 bg-darts-blue rounded-full flex items-center justify-center text-white font-bold">D</div>
          <span className="font-bold text-xl tracking-tight text-slate-800">Darts Web</span>
        </div>
      </nav>

      {/* Main Content */}
      <main className="p-4 md:p-8">
        {view === 'home' && (
          <div className="max-w-4xl mx-auto mt-20 text-center">
            <h1 className="text-5xl md:text-6xl font-black text-slate-900 mb-6 tracking-tight">
              Master the <span className="text-darts-blue">Oche</span>
            </h1>
            <p className="text-xl text-slate-500 mb-12 max-w-2xl mx-auto">
              Professional grade serving system for your Darts matches. Track scores, stats, and games in real-time.
            </p>

            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <button onClick={startSetup} className="px-8 py-4 bg-darts-blue text-white rounded-xl text-lg font-bold shadow-lg shadow-blue-500/30 hover:bg-blue-700 hover:scale-105 transition transform">
                Start New Game
              </button>
              <button onClick={showStats} className="px-8 py-4 bg-white text-slate-700 rounded-xl text-lg font-bold border border-slate-200 hover:bg-slate-50 transition">
                Player Stats
              </button>
            </div>
          </div>
        )}

        {view === 'setup' && (
          <GameSetup onGameStarted={handleGameStarted} />
        )}

        {view === 'game' && activeGameId && (
          <ActiveGame gameId={activeGameId} onExit={handleExit} />
        )}

        {view === 'stats' && (
          <UserStats onBack={handleExit} />
        )}
      </main>
    </div>
  )
}

export default App

