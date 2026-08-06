/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        surface: {
          950: '#f6f8fb',
          900: '#ffffff',
          800: '#f2f5f8',
          700: '#e5e9f0',
          600: '#d5dbe5',
        },
        accent: {
          400: '#3b82f6',
          500: '#2563eb',
          600: '#1d4ed8',
        },
        slate: {
          100: '#111827',
          200: '#1f2937',
          300: '#374151',
          400: '#4b5563',
          500: '#6b7280',
          600: '#9ca3af',
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
