import { NavLink, Outlet } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Button, Icon, Label, type IconData } from '@gravity-ui/uikit'
import { ArrowRightFromSquare, Comments, Database } from '@gravity-ui/icons'

import { api } from '../api/client'
import { useAuth } from '../hooks/useAuth'

export default function Layout() {
  const { user, logout } = useAuth()
  const { data: status } = useQuery({ queryKey: ['kb-status'], queryFn: api.kbStatus })

  return (
    <div className="flex h-full flex-col bg-surface-950">
      <header className="flex h-16 shrink-0 items-center gap-7 border-b border-surface-700 bg-white px-6 shadow-sm">
        <div className="flex items-center gap-2">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-accent-500 text-sm font-bold text-white shadow-sm">
            BS
          </div>
          <div>
            <div className="text-sm font-semibold leading-tight text-slate-100">BlueSales</div>
            <div className="text-xs text-slate-500">Bot Assistant</div>
          </div>
        </div>

        <nav className="flex items-center gap-1">
          <NavItem to="/chats" label="Чаты" icon={Comments} />
          <NavItem to="/documents" label="База знаний" icon={Database} />
        </nav>

        <div className="ml-auto flex items-center gap-3 text-xs text-slate-400">
          {status && (
            <Label theme="normal" size="s">
              {status.model} · {status.reasoningEffort}
            </Label>
          )}
          {status && !status.openrouterKeySet && (
            <Label theme="warning" size="s">нет API-ключа</Label>
          )}
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-50 text-xs font-semibold text-accent-600">
            {user?.login.slice(0, 2).toUpperCase()}
          </div>
          <Button view="flat-secondary" size="m" onClick={() => void logout()}>
            <Button.Icon><Icon data={ArrowRightFromSquare} size={16} /></Button.Icon>
            Выйти
          </Button>
        </div>
      </header>

      <main className="min-h-0 flex-1">
        <Outlet />
      </main>
    </div>
  )
}

function NavItem({ to, label, icon }: { to: string; label: string; icon: IconData }) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        `flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
          isActive ? 'bg-blue-50 text-accent-600' : 'text-slate-400 hover:bg-surface-800 hover:text-slate-100'
        }`
      }
    >
      <Icon data={icon} size={16} />
      {label}
    </NavLink>
  )
}
