import { useState, useEffect } from 'react';
import { api } from '../services/api';

export default function GameSetup({ onGameStarted }) {
  const [users, setUsers] = useState([]);
  const [selectedUsers, setSelectedUsers] = useState([]);
  const [newUserName, setNewUserName] = useState('');
  const [settings, setSettings] = useState({ points: 501, sets: 3 });
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    initSetup();
  }, []);

  const initSetup = async () => {
    try {
      // 1. Load all users
      const allUsers = await api.getUsers();
      let currentUsers = allUsers || [];
      setUsers(currentUsers);

      // 2. Check URL Params for p1, p2
      const params = new URLSearchParams(window.location.search);
      const p1Name = params.get('p1');
      const p2Name = params.get('p2');

      const preSelected = [];

      const ensureUser = async (name) => {
        if (!name) return null;
        // Check if exists
        const existing = currentUsers.find(u => u.name.toLowerCase() === name.toLowerCase());
        if (existing) return existing;
        // Create
        try {
          const newUser = await api.createUser(name);
          const updatedUsers = [...currentUsers, newUser];
          currentUsers = updatedUsers; // Update local reference
          setUsers(updatedUsers);
          return newUser;
        } catch (e) { console.error("Failed to create user from URL", e); return null; }
      };

      if (p1Name) {
        const u1 = await ensureUser(p1Name);
        if (u1) preSelected.push(u1.id);
      }
      if (p2Name) {
        const u2 = await ensureUser(p2Name);
        if (u2) preSelected.push(u2.id);
      }

      if (preSelected.length > 0) {
        setSelectedUsers(preSelected);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const handleCreateUser = async (e) => {
    e.preventDefault();
    if (!newUserName.trim()) return;
    try {
      const user = await api.createUser(newUserName);
      setUsers([...users, user]);
      setNewUserName('');
    } catch (err) {
      alert('Failed to create user');
    }
  };

  const toggleUser = (id) => {
    if (selectedUsers.includes(id)) {
      setSelectedUsers(selectedUsers.filter(u => u !== id));
    } else {
      if (selectedUsers.length >= 4) return; // Max 4 players
      setSelectedUsers([...selectedUsers, id]);
    }
  };

  const startGame = async () => {
    if (selectedUsers.length < 1) return;
    setLoading(true);
    try {
      const game = await api.createGame(settings.points, settings.sets, selectedUsers);
      onGameStarted(game);
    } catch (e) {
      alert('Failed to start game');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-xl mx-auto p-6 bg-white rounded-xl shadow-lg border border-slate-200">
      <h2 className="text-2xl font-bold text-slate-800 mb-6">New Game Setup</h2>

      {/* Settings */}
      <div className="mb-8 grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-2">Points</label>
          <div className="flex gap-2">
            {[301, 501].map(p => (
              <button
                key={p}
                onClick={() => setSettings({ ...settings, points: p })}
                className={`flex-1 py-2 rounded-lg font-semibold border ${settings.points === p ? 'bg-darts-blue text-white border-darts-blue' : 'text-slate-600 border-slate-300'}`}
              >
                {p}
              </button>
            ))}
          </div>
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-2">Best of (Sets)</label>
          <div className="flex gap-2">
            {[1, 3, 5].map(s => (
              <button
                key={s}
                onClick={() => setSettings({ ...settings, sets: s })}
                className={`flex-1 py-2 rounded-lg font-semibold border ${settings.sets === s ? 'bg-darts-blue text-white border-darts-blue' : 'text-slate-600 border-slate-300'}`}
              >
                {s}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Player Selection */}
      <div className="mb-6">
        <label className="block text-sm font-medium text-slate-700 mb-2">Select Players ({selectedUsers.length}/4)</label>
        <div className="grid grid-cols-2 gap-2 mb-4">
          {users.map(u => (
            <button
              key={u.id}
              onClick={() => toggleUser(u.id)}
              className={`p-3 rounded-lg text-left transition ${selectedUsers.includes(u.id) ? 'bg-darts-blue-light text-darts-blue border border-darts-blue font-bold' : 'bg-slate-50 text-slate-600 hover:bg-slate-100'}`}
            >
              {u.name}
            </button>
          ))}
        </div>

        <form onSubmit={handleCreateUser} className="flex gap-2">
          <input
            type="text"
            value={newUserName}
            onChange={e => setNewUserName(e.target.value)}
            placeholder="Add new player..."
            className="flex-1 px-4 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-darts-blue/50"
          />
          <button type="submit" className="px-4 py-2 bg-slate-800 text-white rounded-lg hover:bg-slate-700 font-semibold">+</button>
        </form>
      </div>

      <button
        disabled={selectedUsers.length === 0 || loading}
        onClick={startGame}
        className="w-full py-4 bg-gradient-to-r from-darts-blue to-blue-600 text-white rounded-xl font-bold text-lg shadow-lg hover:shadow-xl hover:scale-[1.02] transition disabled:opacity-50 disabled:hover:scale-100"
      >
        Start Game
      </button>
    </div>
  );
}
