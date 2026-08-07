import { useEffect, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, Icon, Label } from '@gravity-ui/uikit'
import { Comments, Plus, TrashBin } from '@gravity-ui/icons'

import { api, sendMessage } from '../api/client'
import type { Attachment, Usage } from '../api/types'
import Composer from '../components/Composer'
import MessageItem from '../components/MessageItem'

// TODO: вернуть технические сведения о базе знаний и модели.
const SHOW_INTERNAL_UI = false

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
  const messages = chatQuery.data?.messages ?? []

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
  }, [messages.length, streamContent, pendingUser])

  const handleSend = async (text: string, files: File[]) => {
    if (!chatId) return

    const controller = new AbortController()
    abortRef.current = controller

    const pendingAttachments = await Promise.all(
      files.map(async (file) => ({
        filename: file.name,
        size: file.size,
        content: file.name.toLowerCase().endsWith('.json') ? await file.text() : undefined,
      })),
    )
    setPendingUser({ content: text, attachments: pendingAttachments })
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
    <div className="flex h-full min-h-0 bg-surface-950 p-3">
      <aside className="flex w-72 shrink-0 flex-col overflow-hidden rounded-l-xl border border-r-0 border-surface-700 bg-white">
        <div className="space-y-3 border-b border-surface-700 p-4">
          <div>
            <h1 className="text-base font-semibold text-slate-100">Чаты</h1>
            <p className="mt-0.5 text-xs text-slate-500">{chatsQuery.data?.length ?? 0} диалогов</p>
          </div>
          <Button
            view="action"
            size="l"
            width="max"
            onClick={() => createChatMutation.mutate()}
            disabled={createChatMutation.isPending}
            loading={createChatMutation.isPending}
          >
            <Button.Icon><Icon data={Plus} size={16} /></Button.Icon>
            Новый чат
          </Button>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto p-2">
          {chatsQuery.data?.length === 0 && <p className="p-2 text-sm text-slate-500">Чатов пока нет</p>}
          {chatsQuery.data?.map((item) => (
            <div
              key={item.id}
              className={`group mb-1 flex items-center gap-1 rounded-lg border pr-1 transition-colors ${
                item.id === chatId
                  ? 'border-blue-200 bg-blue-50'
                  : 'border-transparent hover:border-surface-700 hover:bg-surface-800'
              }`}
            >
              <button
                className="min-w-0 flex-1 px-3 py-2 text-left"
                onClick={() => navigate(`/chats/${item.id}`)}
              >
                <div className={`truncate text-sm font-medium ${item.id === chatId ? 'text-accent-600' : 'text-slate-200'}`}>
                  {item.title}
                </div>
                <div className="truncate text-xs text-slate-500">
                  {new Date(item.updatedAt).toLocaleString('ru-RU')}
                </div>
              </button>
              <Button
                view="flat-danger"
                size="s"
                className="invisible shrink-0 group-hover:visible"
                onClick={() => {
                  if (confirm('Удалить чат?')) deleteChatMutation.mutate(item.id)
                }}
                aria-label="Удалить чат"
              >
                <Icon data={TrashBin} size={14} />
              </Button>
            </div>
          ))}
        </div>
      </aside>

      <section className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-r-xl border border-surface-700 bg-white">
        {!chatId ? (
          <EmptyState
            hasKB={Boolean(kbStatus?.snapshot)}
            onCreate={() => createChatMutation.mutate()}
            creating={createChatMutation.isPending}
          />
        ) : (
          <>
            <div className="flex min-h-14 shrink-0 flex-wrap items-center gap-3 border-b border-surface-700 px-5 py-2.5">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-50 text-accent-600">
                <Icon data={Comments} size={17} />
              </div>
              <h1 className="truncate text-sm font-semibold text-slate-100">{chat?.title ?? '…'}</h1>
              {SHOW_INTERNAL_UI && (
                <div className="ml-auto flex flex-wrap items-center gap-2 text-xs text-slate-500">
                  {chat?.knowledgeBase ? (
                    <Label theme="success" size="s" title={`Снимок от ${new Date(chat.knowledgeBase.createdAt).toLocaleString('ru-RU')}`}>
                      контекст: {chat.knowledgeBase.documentsCount} докум. ·{' '}
                      {chat.knowledgeBase.charsCount.toLocaleString('ru-RU')} симв.
                    </Label>
                  ) : (
                    <Label theme="warning" size="s">база знаний не подключена</Label>
                  )}
                  <Label theme="normal" size="s">{chat?.model}</Label>
                </div>
              )}
            </div>

            <div ref={scrollRef} className="min-h-0 flex-1 space-y-5 overflow-y-auto bg-surface-950/60 px-6 py-6">
              {messages.length === 0 && !pendingUser && (
                <p className="pt-10 text-center text-sm text-slate-500">
                  Спросите, как настроить бота в BlueSales — ассистент ответит по загруженной базе знаний.
                </p>
              )}

              {messages.map((message) => (
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
                <div className="mx-auto max-w-2xl rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
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

            {SHOW_INTERNAL_UI && !kbStatus?.openrouterKeySet && (
              <div className="border-t border-amber-200 bg-amber-50 px-4 py-2 text-xs text-amber-700">
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
      <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-blue-50 text-accent-600">
        <Icon data={Comments} size={26} />
      </div>
      <h2 className="text-xl font-semibold text-slate-100">Чат с ассистентом BlueSales</h2>
      <p className="max-w-md text-sm text-slate-500">
        Начните новый диалог с ассистентом.
      </p>
      {SHOW_INTERNAL_UI && !hasKB && (
        <p className="max-w-md rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-700">
          База знаний ещё не синхронизирована — чат будет работать без контекста документов.
        </p>
      )}
      <Button view="action" size="l" onClick={onCreate} disabled={creating} loading={creating}>
        <Button.Icon><Icon data={Plus} size={16} /></Button.Icon>
        Новый чат
      </Button>
    </div>
  )
}
