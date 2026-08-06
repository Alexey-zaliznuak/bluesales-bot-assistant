/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        surface: {
          950: '#0b0f16',
          900: '#111722',
          800: '#18202e',
          700: '#222c3d',
          600: '#2f3a4d',
        },
        accent: {
          400: '#4da3ff',
          500: '#2b8bff',
          600: '#1a6fd6',
        },
      },
      fontFamily: {
        sans: ['Inter', 'Segoe UI', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Consolas', 'monospace'],
      },
    },
  },
  plugins: [],
}
