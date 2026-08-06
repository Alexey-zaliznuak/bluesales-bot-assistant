import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

import type { Attachment, Usage } from '../api/types'
import JsonFileCard, { isJsonResponse } from './JsonFileCard'

interface Props {
  role: 'user' | 'assistant' | 'system'
  content: string
  attachments?: Attachment[]
  reasoning?: string
  usage?: Usage | null
  error?: string | null
  streaming?: boolean
}

export default function MessageItem({
  role,
  content,
  attachments = [],
  reasoning,
  usage,
  error,
  streaming,
}: Props) {
  const isUser = role === 'user'
  const jsonResponse = !isUser && isJsonResponse(content, streaming)

  return (
    <div className={`flex gap-3 ${isUser ? 'justify-end' : 'justify-start'}`}>
      {!isUser && (
        <div className="mt-1 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-xs font-semibold text-accent-600">
          AI
        </div>
      )}

      <div className={`min-w-0 max-w-[min(46rem,85%)] space-y-2`}>
        {reasoning && <ReasoningBlock text={reasoning} streaming={streaming} />}

        <div
          className={`rounded-2xl px-4 py-3 ${
            isUser
              ? 'bg-accent-500 text-white shadow-sm'
              : 'border border-surface-700 bg-white text-slate-100 shadow-sm'
          }`}
        >
          {jsonResponse ? (
            <JsonFileCard content={content} streaming={streaming} />
          ) : content ? (
            <div className={`chat-markdown ${isUser ? 'chat-markdown-user' : ''}`}>
              <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
            </div>
          ) : streaming ? (
            <TypingDots />
          ) : (
            <span className="text-sm italic text-slate-500">пустой ответ</span>
          )}

          {attachments.length > 0 && (
            <div className="mt-2 flex flex-wrap gap-1.5 border-t border-white/10 pt-2">
              {attachments.map((file) => (
                <span key={file.filename} className={`badge gap-1.5 ${isUser ? 'border-white/20 bg-white/10 text-white' : ''}`}>
                  <span className="text-slate-400">файл</span>
                  <span className="font-mono">{file.filename}</span>
                  <span className="text-slate-500">{formatSize(file.size)}</span>
                </span>
              ))}
            </div>
          )}
        </div>

        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700">
            {error}
          </div>
        )}

        {usage && (
          <div className="flex flex-wrap gap-x-3 gap-y-1 px-1 text-[11px] text-slate-500">
            <span>промпт: {usage.promptTokens.toLocaleString('ru-RU')}</span>
            <span>ответ: {usage.completionTokens.toLocaleString('ru-RU')}</span>
            {usage.reasoningTokens > 0 && <span>рассуждения: {usage.reasoningTokens.toLocaleString('ru-RU')}</span>}
            {usage.cost != null && <span>${usage.cost.toFixed(5)}</span>}
          </div>
        )}
      </div>
    </div>
  )
}

function ReasoningBlock({ text, streaming }: { text: string; streaming?: boolean }) {
  const [open, setOpen] = useState(false)

  return (
    <div className="rounded-xl border border-dashed border-surface-600 bg-white">
      <button
        className="flex w-full items-center gap-2 px-3 py-2 text-left text-xs text-slate-400 hover:text-slate-200"
        onClick={() => setOpen(!open)}
      >
        <span>{open ? '▾' : '▸'}</span>
        <span>Рассуждения модели</span>
        {streaming && <span className="text-accent-600">идут…</span>}
      </button>
      {open && (
        <div className="max-h-64 overflow-y-auto whitespace-pre-wrap break-words border-t border-surface-700 px-3 py-2 text-xs text-slate-400">
          {text}
        </div>
      )}
    </div>
  )
}

function TypingDots() {
  return (
    <div className="flex gap-1 py-1">
      {[0, 150, 300].map((delay) => (
        <span
          key={delay}
          className="h-1.5 w-1.5 animate-bounce rounded-full bg-slate-500"
          style={{ animationDelay: `${delay}ms` }}
        />
      ))}
    </div>
  )
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} Б`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} КБ`
  return `${(bytes / (1024 * 1024)).toFixed(1)} МБ`
}
