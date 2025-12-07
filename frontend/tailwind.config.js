/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'darts-blue': '#186aa1',
        'darts-blue-light': '#c0eaf5',
        'darts-gold': '#ffd700',
      },
    },
  },
  plugins: [],
}
