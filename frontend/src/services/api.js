const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

export const api = {
  getUsers: async () => {
    const res = await fetch(`${API_URL}/users`);
    return res.json();
  },

  createUser: async (name) => {
    const res = await fetch(`${API_URL}/users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    });
    return res.json();
  },

  createGame: async (totalPoints, bestOf, playerIds) => {
    const res = await fetch(`${API_URL}/games`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        total_points: totalPoints,
        best_of: bestOf,
        player_ids: playerIds
      }),
    });
    return res.json();
  },

  getGame: async (id) => {
    const res = await fetch(`${API_URL}/games/${id}`);
    if (!res.ok) throw new Error('Game not found');
    return res.json();
  },

  sendThrow: async (gameId, userId, points, multiplier) => {
    const res = await fetch(`${API_URL}/games/${gameId}/throw`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        user_id: userId,
        points,
        multiplier
      }),
    });
    if (!res.ok) {
      const err = await res.text();
      throw new Error(err);
    }
    return res.json();
  },

  getUserStats: async (userId) => {
    const res = await fetch(`${API_URL}/users/${userId}/stats`);
    if (!res.ok) throw new Error('Failed to load stats');
    return res.json();
  }
};
