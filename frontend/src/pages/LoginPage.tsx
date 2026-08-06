import { useState, type FormEvent } from 'react'

import { useAuth } from '../hooks/useAuth'

export default function LoginPage() {
  const { login } = useAuth()
  const [loginValue, setLoginValue] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await login(loginValue, password)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось войти')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex h-full items-center justify-center px-4">
      <form onSubmit={handleSubmit} className="card w-full max-w-sm space-y-5 p-8">
        <div className="space-y-1">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-accent-500 font-bold text-white">
            BS
          </div>
          <h1 className="pt-3 text-lg font-semibold text-slate-100">Вход</h1>
          <p className="text-sm text-slate-500">Ассистент по настройке ботов BlueSales</p>
        </div>

        <div className="space-y-3">
          <label className="block space-y-1.5">
            <span className="text-xs font-medium uppercase tracking-wide text-slate-500">Логин</span>
            <input
              className="input"
              value={loginValue}
              onChange={(event) => setLoginValue(event.target.value)}
              autoComplete="username"
              autoFocus
            />
          </label>

          <label className="block space-y-1.5">
            <span className="text-xs font-medium uppercase tracking-wide text-slate-500">Пароль</span>
            <input
              className="input"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete="current-password"
            />
          </label>
        </div>

        {error && (
          <div className="rounded-lg border border-red-900/60 bg-red-950/40 px-3 py-2 text-sm text-red-300">
            {error}
          </div>
        )}

        <button className="btn-primary w-full" type="submit" disabled={submitting}>
          {submitting ? 'Входим…' : 'Войти'}
        </button>
      </form>
    </div>
  )
}
