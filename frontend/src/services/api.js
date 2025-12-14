// Use relative path by default to work with BASE_PATH in production
// In development, Vite proxy or explicit VITE_API_URL can override this
const API_URL = import.meta.env.VITE_API_URL || '/darts/api';

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
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || 'Failed to create user');
    }
    return data;
  },

  deleteUser: async (userId) => {
    const res = await fetch(`${API_URL}/users/${userId}`, {
      method: 'DELETE',
    });
    if (!res.ok) {
      const error = await res.json();
      throw new Error(error.error || 'Failed to delete user');
    }
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
