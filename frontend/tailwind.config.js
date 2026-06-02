/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#fffbeb',
          100: '#fef3c7',
          400: '#f59e0b',
          500: '#d97706',
          600: '#b45309',
          700: '#92400e',
        },
        apricot: {
          50: '#FFF8F1',
          100: '#FEF0E2',
          200: '#FDE4CC',
          300: '#FBD0A5',
          400: '#F5BC82',
          500: '#E8A87C',
          600: '#D4956A',
        },
      },
    },
  },
  plugins: [],
}
