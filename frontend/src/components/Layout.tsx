import { NavLink, Outlet } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'

import { api } from '../api/client'
import { useAuth } from '../hooks/useAuth'

export default function Layout() {
  const { user, logout } = useAuth()
  const { data: status } = useQuery({ queryKey: ['kb-status'], queryFn: api.kbStatus })

  return (
    <div className="flex h-full flex-col">
      <header className="flex shrink-0 items-center gap-6 border-b border-surface-700 bg-surface-900 px-6 py-3">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-accent-500 text-sm font-bold text-white">
            BS
          </div>
          <span className="text-sm font-semibold text-slate-100">BlueSales Bot Assistant</span>
        </div>

        <nav className="flex items-center gap-1">
          <NavItem to="/chats" label="Чаты" />
          <NavItem to="/documents" label="База знаний" />
        </nav>

        <div className="ml-auto flex items-center gap-3 text-xs text-slate-400">
          {status && (
            <span className="badge font-mono" title={`Кэш: ${status.cacheMode}, TTL ${status.cacheTtl}`}>
              {status.model} · {status.reasoningEffort}
            </span>
          )}
          {status && !status.openrouterKeySet && (
            <span className="badge border-amber-800 bg-amber-950/50 text-amber-300">нет API-ключа</span>
          )}
          <span className="text-slate-300">{user?.login}</span>
          <button className="btn-ghost" onClick={() => void logout()}>
            Выйти
          </button>
        </div>
      </header>

      <main className="min-h-0 flex-1">
        <Outlet />
      </main>
    </div>
  )
}

function NavItem({ to, label }: { to: string; label: string }) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        `rounded-lg px-3 py-1.5 text-sm transition-colors ${
          isActive ? 'bg-surface-700 text-slate-100' : 'text-slate-400 hover:bg-surface-800 hover:text-slate-200'
        }`
      }
    >
      {label}
    </NavLink>
  )
}
