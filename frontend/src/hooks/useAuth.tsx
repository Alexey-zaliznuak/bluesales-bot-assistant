import { createContext, useCallback, useContext, useMemo, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import { ApiError, api } from '../api/client'
import type { User } from '../api/types'

interface AuthContextValue {
  user: User | null
  loading: boolean
  login: (login: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['me'],
    queryFn: async () => {
      try {
        return await api.me()
      } catch (error) {
        if (error instanceof ApiError && error.status === 401) return null
        throw error
      }
    },
    staleTime: 60_000,
  })

  const login = useCallback(
    async (loginValue: string, password: string) => {
      const user = await api.login(loginValue, password)
      queryClient.setQueryData(['me'], user)
    },
    [queryClient],
  )

  const logout = useCallback(async () => {
    await api.logout()
    queryClient.setQueryData(['me'], null)
    queryClient.clear()
  }, [queryClient])

  const value = useMemo<AuthContextValue>(
    () => ({ user: data ?? null, loading: isLoading, login, logout }),
    [data, isLoading, login, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuth используется вне AuthProvider')
  return context
}
