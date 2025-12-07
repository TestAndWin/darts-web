import { useState, useEffect } from 'react';
import { api } from '../services/api';

export default function UserStats({ onBack }) {
  const [users, setUsers] = useState([]);
  const [selectedUser, setSelectedUser] = useState(null);
  const [stats, setStats] = useState(null);

  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    try {
      const u = await api.getUsers();
      setUsers(u || []);
    } catch (e) { console.error(e); }
  }

  const selectUser = async (user) => {
    setSelectedUser(user);
    setStats(null);
    try {
      const s = await api.getUserStats(user.id);
      setStats(s);
    } catch (e) { console.error(e); }
  }

  return (
    <div className="max-w-4xl mx-auto">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-3xl font-bold text-slate-800">Player Statistics</h2>
        <button onClick={onBack} className="text-darts-blue hover:underline">Back to Menu</button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* User List */}
        <div className="bg-white p-6 rounded-xl shadow-sm border border-slate-200">
          <h3 className="font-semibold text-slate-500 mb-4 uppercase text-sm">Select Player</h3>
          <div className="space-y-2">
            {users.map(u => (
              <button
                key={u.id}
                onClick={() => selectUser(u)}
                className={`w-full text-left p-3 rounded-lg transition ${selectedUser?.id === u.id ? 'bg-darts-blue text-white font-bold' : 'hover:bg-slate-50 text-slate-700'}`}
              >
                {u.name}
              </button>
            ))}
          </div>
        </div>

        {/* Stats Display */}
        <div className="md:col-span-2">
          {selectedUser ? (
            stats ? (
              <div className="bg-white p-8 rounded-xl shadow-lg border border-slate-200">
                <div className="flex items-center gap-4 mb-8">
                  <div className="w-16 h-16 bg-darts-gold rounded-full flex items-center justify-center text-2xl font-bold text-slate-900">
                    {selectedUser.name[0]}
                  </div>
                  <div>
                    <h3 className="text-2xl font-bold text-slate-900">{selectedUser.name}</h3>
                    <p className="text-slate-500">Player Profile</p>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="p-4 bg-slate-50 rounded-xl border border-slate-100">
                    <div className="text-sm text-slate-500 mb-1">Total Games</div>
                    <div className="text-3xl font-black text-slate-800">{stats.total_games}</div>
                  </div>
                  <div className="p-4 bg-slate-50 rounded-xl border border-slate-100">
                    <div className="text-sm text-slate-500 mb-1">Wins</div>
                    <div className="text-3xl font-black text-green-600">{stats.wins}</div>
                  </div>
                  <div className="p-4 bg-slate-50 rounded-xl border border-slate-100">
                    <div className="text-sm text-slate-500 mb-1">3-Dart Average</div>
                    <div className="text-3xl font-black text-darts-blue">{parseFloat(stats.average_3_dart).toFixed(2)}</div>
                  </div>
                  <div className="p-4 bg-slate-50 rounded-xl border border-slate-100">
                    <div className="text-sm text-slate-500 mb-1">Total Throws</div>
                    <div className="text-3xl font-black text-slate-800">{stats.total_throws}</div>
                  </div>
                </div>
              </div>
            ) : (
              <div className="h-full flex items-center justify-center text-slate-400">Loading stats...</div>
            )
          ) : (
            <div className="h-full flex items-center justify-center bg-slate-50 rounded-xl border-2 border-dashed border-slate-200 text-slate-400">
              Select a player to view statistics
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
