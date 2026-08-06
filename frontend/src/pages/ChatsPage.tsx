import { useEffect, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { api, sendMessage } from '../api/client'
import type { Attachment, Usage } from '../api/types'
import Composer from '../components/Composer'
import MessageItem from '../components/MessageItem'

interface PendingUser {
  content: string
  attachments: Attachment[]
}

export default function ChatsPage() {
  const { chatId } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [pendingUser, setPendingUser] = useState<PendingUser | null>(null)
  const [streamContent, setStreamContent] = useState('')
  const [streamReasoning, setStreamReasoning] = useState('')
  const [streamUsage, setStreamUsage] = useState<Usage | null>(null)
  const [streamError, setStreamError] = useState<string | null>(null)
  const [streaming, setStreaming] = useState(false)
  const abortRef = useRef<AbortController | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)

  const chatsQuery = useQuery({ queryKey: ['chats'], queryFn: api.listChats })
  const chatQuery = useQuery({
    queryKey: ['chat', chatId],
    queryFn: () => api.getChat(chatId!),
    enabled: Boolean(chatId),
  })
  const kbStatusQuery = useQuery({ queryKey: ['kb-status'], queryFn: api.kbStatus })

  const createChatMutation = useMutation({
    mutationFn: () => api.createChat(),
    onSuccess: (chat) => {
      void queryClient.invalidateQueries({ queryKey: ['chats'] })
      navigate(`/chats/${chat.id}`)
    },
  })

  const deleteChatMutation = useMutation({
    mutationFn: (id: string) => api.deleteChat(id),
    onSuccess: (_data, id) => {
      void queryClient.invalidateQueries({ queryKey: ['chats'] })
      if (id === chatId) navigate('/chats')
    },
  })

  // Сброс состояния стрима при переключении чата.
  useEffect(() => {
    abortRef.current?.abort()
    abortRef.current = null
    setPendingUser(null)
    setStreamContent('')
    setStreamReasoning('')
    setStreamUsage(null)
    setStreamError(null)
    setStreaming(false)
  }, [chatId])

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight, behavior: 'smooth' })
  }, [chatQuery.data?.messages.length, streamContent, pendingUser])

  const handleSend = async (text: string, files: File[]) => {
    if (!chatId) return

    const controller = new AbortController()
    abortRef.current = controller

    setPendingUser({
      content: text,
      attachments: files.map((file) => ({ filename: file.name, size: file.size })),
    })
    setStreamContent('')
    setStreamReasoning('')
    setStreamUsage(null)
    setStreamError(null)
    setStreaming(true)

    try {
      await sendMessage(
        chatId,
        text,
        files,
        {
          onDelta: (delta) => setStreamContent((current) => current + delta),
          onReasoning: (delta) => setStreamReasoning((current) => current + delta),
          onUsage: (usage) => setStreamUsage(usage),
          onError: (message) => setStreamError(message),
          onChat: () => void queryClient.invalidateQueries({ queryKey: ['chats'] }),
        },
        controller.signal,
      )
      await queryClient.invalidateQueries({ queryKey: ['chat', chatId] })
      await queryClient.invalidateQueries({ queryKey: ['chats'] })
      setPendingUser(null)
      setStreamContent('')
      setStreamReasoning('')
      setStreamUsage(null)
    } catch (error) {
      if (controller.signal.aborted) {
        await queryClient.invalidateQueries({ queryKey: ['chat', chatId] })
        setPendingUser(null)
        setStreamContent('')
      } else {
        setStreamError(error instanceof Error ? error.message : 'Не удалось отправить сообщение')
      }
    } finally {
      setStreaming(false)
      abortRef.current = null
    }
  }

  const chat = chatQuery.data
  const kbStatus = kbStatusQuery.data

  return (
    <div className="flex h-full min-h-0">
      <aside className="flex w-72 shrink-0 flex-col border-r border-surface-700 bg-surface-900/50">
        <div className="border-b border-surface-700 p-3">
          <button
            className="btn-primary w-full"
            onClick={() => createChatMutation.mutate()}
            disabled={createChatMutation.isPending}
          >
            + Новый чат
          </button>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto p-2">
          {chatsQuery.data?.length === 0 && <p className="p-2 text-sm text-slate-500">Чатов пока нет</p>}
          {chatsQuery.data?.map((item) => (
            <div
              key={item.id}
              className={`group mb-1 flex items-center gap-1 rounded-lg pr-1 transition-colors ${
                item.id === chatId ? 'bg-surface-700' : 'hover:bg-surface-800'
              }`}
            >
              <button
                className="min-w-0 flex-1 px-3 py-2 text-left"
                onClick={() => navigate(`/chats/${item.id}`)}
              >
                <div className="truncate text-sm text-slate-200">{item.title}</div>
                <div className="truncate text-xs text-slate-500">
                  {new Date(item.updatedAt).toLocaleString('ru-RU')}
                </div>
              </button>
              <button
                className="hidden shrink-0 rounded px-2 py-1 text-slate-500 hover:text-red-400 group-hover:block"
                onClick={() => {
                  if (confirm('Удалить чат?')) deleteChatMutation.mutate(item.id)
                }}
                aria-label="Удалить чат"
              >
                ×
              </button>
            </div>
          ))}
        </div>
      </aside>

      <section className="flex min-h-0 min-w-0 flex-1 flex-col">
        {!chatId ? (
          <EmptyState
            hasKB={Boolean(kbStatus?.snapshot)}
            onCreate={() => createChatMutation.mutate()}
            creating={createChatMutation.isPending}
          />
        ) : (
          <>
            <div className="flex shrink-0 flex-wrap items-center gap-3 border-b border-surface-700 px-4 py-2.5">
              <h1 className="truncate text-sm font-medium text-slate-100">{chat?.title ?? '…'}</h1>
              <div className="ml-auto flex flex-wrap items-center gap-2 text-xs text-slate-500">
                {chat?.knowledgeBase ? (
                  <span className="badge" title={`Снимок от ${new Date(chat.knowledgeBase.createdAt).toLocaleString('ru-RU')}`}>
                    контекст: {chat.knowledgeBase.documentsCount} докум. ·{' '}
                    {chat.knowledgeBase.charsCount.toLocaleString('ru-RU')} симв.
                  </span>
                ) : (
                  <span className="badge border-amber-800 bg-amber-950/50 text-amber-300">
                    база знаний не подключена
                  </span>
                )}
                <span className="badge font-mono">{chat?.model}</span>
              </div>
            </div>

            <div ref={scrollRef} className="min-h-0 flex-1 space-y-5 overflow-y-auto px-4 py-5">
              {chat?.messages.length === 0 && !pendingUser && (
                <p className="pt-10 text-center text-sm text-slate-500">
                  Спросите, как настроить бота в BlueSales — ассистент ответит по загруженной базе знаний.
                </p>
              )}

              {chat?.messages.map((message) => (
                <MessageItem
                  key={message.id}
                  role={message.role}
                  content={message.content}
                  attachments={message.attachments}
                  usage={message.usage}
                  error={message.error}
                />
              ))}

              {pendingUser && (
                <MessageItem role="user" content={pendingUser.content} attachments={pendingUser.attachments} />
              )}

              {(streaming || streamContent) && (
                <MessageItem
                  role="assistant"
                  content={streamContent}
                  reasoning={streamReasoning || undefined}
                  usage={streamUsage}
                  streaming={streaming}
                />
              )}

              {streamError && (
                <div className="mx-auto max-w-2xl rounded-lg border border-red-900/60 bg-red-950/40 px-3 py-2 text-sm text-red-300">
                  {streamError}
                </div>
              )}
            </div>

            <Composer
              disabled={!chatId || !kbStatus?.openrouterKeySet}
              streaming={streaming}
              onSend={(text, files) => void handleSend(text, files)}
              onStop={() => abortRef.current?.abort()}
            />

            {!kbStatus?.openrouterKeySet && (
              <div className="border-t border-amber-900/50 bg-amber-950/30 px-4 py-2 text-xs text-amber-300">
                Не задан OPENROUTER_API_KEY — отправка сообщений отключена.
              </div>
            )}
          </>
        )}
      </section>
    </div>
  )
}

function EmptyState({
  hasKB,
  onCreate,
  creating,
}: {
  hasKB: boolean
  onCreate: () => void
  creating: boolean
}) {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-4 px-6 text-center">
      <h2 className="text-lg font-medium text-slate-200">Чат с ассистентом BlueSales</h2>
      <p className="max-w-md text-sm text-slate-500">
        Новый чат берёт текущий снимок базы знаний как кэшируемый контекст. Пересборка базы позже не меняет
        контекст уже начатых чатов.
      </p>
      {!hasKB && (
        <p className="max-w-md rounded-lg border border-amber-900/60 bg-amber-950/30 px-3 py-2 text-xs text-amber-300">
          База знаний ещё не синхронизирована — чат будет работать без контекста документов.
        </p>
      )}
      <button className="btn-primary" onClick={onCreate} disabled={creating}>
        + Новый чат
      </button>
    </div>
  )
}
