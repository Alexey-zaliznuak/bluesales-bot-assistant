import { useState, type FormEvent } from 'react'
import { Alert, Button, PasswordInput, TextInput } from '@gravity-ui/uikit'
import { Link } from 'react-router-dom'

import { useAuth } from '../hooks/useAuth'

export default function RegisterPage() {
  const { register } = useAuth()
  const [loginValue, setLoginValue] = useState('')
  const [password, setPassword] = useState('')
  const [passwordConfirmation, setPasswordConfirmation] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault()
    setError(null)

    const normalizedLogin = loginValue.trim()
    const loginLength = Array.from(normalizedLogin).length
    if (loginLength < 3 || loginLength > 64) {
      setError('Логин должен содержать от 3 до 64 символов')
      return
    }
    if (Array.from(password).length < 8) {
      setError('Пароль должен содержать не менее 8 символов')
      return
    }
    if (new TextEncoder().encode(password).length > 72) {
      setError('Пароль не должен превышать 72 байта')
      return
    }
    if (password !== passwordConfirmation) {
      setError('Пароли не совпадают')
      return
    }

    setSubmitting(true)
    try {
      await register(normalizedLogin, password)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось зарегистрироваться')
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
          <h1 className="pt-4 text-2xl font-semibold tracking-tight text-slate-100">
            Создание аккаунта
          </h1>
          <p className="text-sm text-slate-500">Зарегистрируйтесь для работы с AI-ассистентом</p>
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
            autoComplete="new-password"
          />
          <PasswordInput
            label="Повторите пароль"
            size="xl"
            value={passwordConfirmation}
            onUpdate={setPasswordConfirmation}
            autoComplete="new-password"
          />
        </div>

        {error && <Alert theme="danger" message={error} />}

        <div className="space-y-3">
          <Button view="action" size="xl" width="max" type="submit" loading={submitting}>
            {submitting ? 'Создаём аккаунт…' : 'Зарегистрироваться'}
          </Button>
          <p className="text-center text-sm text-slate-500">
            Уже есть аккаунт?{' '}
            <Link className="font-medium text-accent-600 hover:text-accent-700" to="/login">
              Войти
            </Link>
          </p>
        </div>
      </form>
    </div>
  )
}
