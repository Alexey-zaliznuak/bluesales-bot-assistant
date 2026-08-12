import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { settings } from '@gravity-ui/date-utils'
import { ThemeProvider } from '@gravity-ui/uikit'
import { BrowserRouter } from 'react-router-dom'

import '@gravity-ui/uikit/styles/fonts.css'
import '@gravity-ui/uikit/styles/styles.css'
import App from './App'
import { AuthProvider } from './hooks/useAuth'
import './index.css'

const routerBase = import.meta.env.BASE_URL.replace(/\/$/, '') || '/'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
      refetchOnWindowFocus: false,
      staleTime: 15_000,
    },
  },
})

function renderApp() {
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <ThemeProvider theme="light" lang="ru">
        <QueryClientProvider client={queryClient}>
          <BrowserRouter basename={routerBase}>
            <AuthProvider>
              <App />
            </AuthProvider>
          </BrowserRouter>
        </QueryClientProvider>
      </ThemeProvider>
    </StrictMode>,
  )
}

void settings.loadLocale('ru').then(renderApp, renderApp)
