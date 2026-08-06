import { useState, type FormEvent } from 'react'
import { Alert, Button, PasswordInput, TextInput } from '@gravity-ui/uikit'

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
    <div className="relative flex h-full items-center justify-center overflow-hidden bg-surface-950 px-4">
      <div className="absolute inset-x-0 top-0 h-1 bg-accent-500" />
      <div className="absolute left-1/2 top-24 h-72 w-72 -translate-x-1/2 rounded-full bg-blue-100/70 blur-3xl" />
      <form onSubmit={handleSubmit} className="card relative w-full max-w-md space-y-6 p-9">
        <div className="space-y-1">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-accent-500 font-bold text-white shadow-sm">
            BS
          </div>
          <h1 className="pt-4 text-2xl font-semibold tracking-tight text-slate-100">Добро пожаловать</h1>
          <p className="text-sm text-slate-500">Корпоративный AI-ассистент BlueSales</p>
        </div>

        <div className="space-y-4">
          <TextInput
            label="Логин"
            size="xl"
            value={loginValue}
            onUpdate={setLoginValue}
            autoComplete="username"
            autoFocus
          />
          <PasswordInput
            label="Пароль"
            size="xl"
            value={password}
            onUpdate={setPassword}
            autoComplete="current-password"
          />
        </div>

        {error && <Alert theme="danger" message={error} />}

        <Button view="action" size="xl" width="max" type="submit" loading={submitting}>
          {submitting ? 'Входим…' : 'Войти'}
        </Button>
      </form>
    </div>
  )
}
