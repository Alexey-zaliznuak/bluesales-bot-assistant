import type {
  Chat,
  ChatDetail,
  Document,
  KBStatus,
  KBSyncResult,
  Message,
  Usage,
  User,
} from './types'

const apiBase = `${import.meta.env.BASE_URL}api`

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
    this.name = 'ApiError'
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`${apiBase}${path}`, {
    credentials: 'include',
    ...init,
    headers: {
      ...(init.body && !(init.body instanceof FormData) ? { 'Content-Type': 'application/json' } : {}),
      ...init.headers,
    },
  })

  if (!response.ok) {
    let message = `Ошибка ${response.status}`
    try {
      const body = await response.json()
      if (body?.error) message = body.error
    } catch {
      // тело не JSON — оставляем код статуса
    }
    throw new ApiError(response.status, message)
  }

  if (response.status === 204) return undefined as T
  return (await response.json()) as T
}

export const api = {
  login: (login: string, password: string) =>
    request<User>('/auth/login', { method: 'POST', body: JSON.stringify({ login, password }) }),
  logout: () => request<{ status: string }>('/auth/logout', { method: 'POST' }),
  me: () => request<User>('/auth/me'),

  listDocuments: (params: { search?: string; category?: string } = {}) => {
    const query = new URLSearchParams()
    if (params.search) query.set('search', params.search)
    if (params.category) query.set('category', params.category)
    const suffix = query.toString() ? `?${query}` : ''
    return request<Document[]>(`/documents${suffix}`)
  },
  listCategories: () => request<string[]>('/documents/categories'),
  createDocument: (payload: { title: string; categories: string[]; body: string }) =>
    request<Document>('/documents', { method: 'POST', body: JSON.stringify(payload) }),
  updateDocument: (id: string, payload: { title?: string; categories?: string[]; body?: string }) =>
    request<Document>(`/documents/${id}`, { method: 'PATCH', body: JSON.stringify(payload) }),
  deleteDocument: (id: string) => request<void>(`/documents/${id}`, { method: 'DELETE' }),

  kbStatus: () => request<KBStatus>('/kb/status'),
  kbPreview: () =>
    request<{ content: string; hash: string; documentsCount: number; charsCount: number }>(
      '/kb/preview',
    ),
  kbSync: () => request<KBSyncResult>('/kb/sync', { method: 'POST' }),

  listChats: () => request<Chat[]>('/chats'),
  createChat: (title?: string) =>
    request<ChatDetail>('/chats', { method: 'POST', body: JSON.stringify({ title: title ?? '' }) }),
  getChat: (id: string) => request<ChatDetail>(`/chats/${id}`),
  renameChat: (id: string, title: string) =>
    request<{ status: string }>(`/chats/${id}`, { method: 'PATCH', body: JSON.stringify({ title }) }),
  deleteChat: (id: string) => request<void>(`/chats/${id}`, { method: 'DELETE' }),
}

export interface StreamHandlers {
  onUserMessage?: (message: Message) => void
  onChat?: (chat: { id: string; title: string }) => void
  onDelta?: (delta: string) => void
  onReasoning?: (delta: string) => void
  onUsage?: (usage: Usage) => void
  onDone?: (message: Message) => void
  onError?: (error: string) => void
}

// sendMessage читает SSE-поток ответа вручную: EventSource не умеет POST
// и не передаёт файлы.
export async function sendMessage(
  chatId: string,
  content: string,
  files: File[],
  handlers: StreamHandlers,
  signal?: AbortSignal,
): Promise<void> {
  const form = new FormData()
  form.set('content', content)
  files.forEach((file) => form.append('files', file))

  const response = await fetch(`${apiBase}/chats/${chatId}/messages`, {
    method: 'POST',
    credentials: 'include',
    body: form,
    signal,
  })

  if (!response.ok || !response.body) {
    let message = `Ошибка ${response.status}`
    try {
      const body = await response.json()
      if (body?.error) message = body.error
    } catch {
      // не JSON
    }
    throw new ApiError(response.status, message)
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  const dispatch = (event: string, raw: string) => {
    if (!raw) return
    let payload: unknown
    try {
      payload = JSON.parse(raw)
    } catch {
      return
    }
    switch (event) {
      case 'user_message':
        handlers.onUserMessage?.(payload as Message)
        break
      case 'chat':
        handlers.onChat?.(payload as { id: string; title: string })
        break
      case 'delta':
        handlers.onDelta?.((payload as { delta: string }).delta)
        break
      case 'reasoning':
        handlers.onReasoning?.((payload as { delta: string }).delta)
        break
      case 'usage':
        handlers.onUsage?.(payload as Usage)
        break
      case 'done':
        handlers.onDone?.(payload as Message)
        break
      case 'error':
        handlers.onError?.((payload as { error: string }).error)
        break
    }
  }

  const flushFrames = () => {
    let boundary = buffer.indexOf('\n\n')
    while (boundary !== -1) {
      const frame = buffer.slice(0, boundary)
      buffer = buffer.slice(boundary + 2)

      let event = 'message'
      const dataLines: string[] = []
      for (const line of frame.split('\n')) {
        if (line.startsWith('event:')) event = line.slice(6).trim()
        else if (line.startsWith('data:')) dataLines.push(line.slice(5).trim())
      }
      dispatch(event, dataLines.join('\n'))

      boundary = buffer.indexOf('\n\n')
    }
  }

  for (;;) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    flushFrames()
  }
  buffer += decoder.decode()
  flushFrames()
}
