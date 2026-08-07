import { Navigate, Route, Routes } from 'react-router-dom'
import { Spin } from '@gravity-ui/uikit'

import Layout from './components/Layout'
import { useAuth } from './hooks/useAuth'
import ChatsPage from './pages/ChatsPage'
import DocumentsPage from './pages/DocumentsPage'
import LoginPage from './pages/LoginPage'

// TODO: вернуть маршрут базы знаний вместе с пунктом навигации в Layout.
const SHOW_KNOWLEDGE_BASE = false

export default function App() {
  const { user, loading } = useAuth()

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center bg-surface-950">
        <Spin size="l" />
      </div>
    )
  }

  if (!user) {
    return (
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    )
  }

  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/chats" element={<ChatsPage />} />
        <Route path="/chats/:chatId" element={<ChatsPage />} />
        {SHOW_KNOWLEDGE_BASE && <Route path="/documents" element={<DocumentsPage />} />}
        <Route path="*" element={<Navigate to="/chats" replace />} />
      </Route>
    </Routes>
  )
}
