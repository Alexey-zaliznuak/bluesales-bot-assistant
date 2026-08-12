import { lazy, Suspense } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { Spin } from '@gravity-ui/uikit'

import Layout from './components/Layout'
import { useAuth } from './hooks/useAuth'
import ChatsPage from './pages/ChatsPage'
import DocumentsPage from './pages/DocumentsPage'
import LandingPage from './pages/LandingPage'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'

// TODO: вернуть маршрут базы знаний вместе с пунктом навигации в Layout.
const SHOW_KNOWLEDGE_BASE = false
const AdminPage = lazy(() => import('./pages/AdminPage'))

export default function App() {
  const { user, loading } = useAuth()

  if (loading) {
    return <PageLoader />
  }

  if (!user) {
    return (
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    )
  }

  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route element={<Layout />}>
        <Route path="/chats" element={<ChatsPage />} />
        <Route path="/chats/:chatId" element={<ChatsPage />} />
        {user.isAdmin && (
          <Route
            path="/admin"
            element={
              <Suspense fallback={<PageLoader />}>
                <AdminPage />
              </Suspense>
            }
          />
        )}
        {SHOW_KNOWLEDGE_BASE && <Route path="/documents" element={<DocumentsPage />} />}
        <Route path="*" element={<Navigate to="/chats" replace />} />
      </Route>
    </Routes>
  )
}

function PageLoader() {
  return (
    <div className="flex h-full items-center justify-center bg-surface-950">
      <Spin size="l" />
    </div>
  )
}
