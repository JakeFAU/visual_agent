/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        gray: {
          750: '#2d333d',
          850: '#1a1f29',
          950: '#0a0c10',
        },
      },
    },
  },
  plugins: [],
}
