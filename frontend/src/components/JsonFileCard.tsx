import { useEffect, useMemo, useState } from 'react'
import { ArrowDownToLine, FileCode, Xmark } from '@gravity-ui/icons'
import { Button, Icon, Loader } from '@gravity-ui/uikit'

interface Props {
  content: string
  streaming?: boolean
}

export function isLargeJsonResponse(content: string): boolean {
  const value = content.trimStart()
  const isJson = value.startsWith('{') || value.startsWith('[')

  return isJson && extractJson(content).split(/\r?\n/).length > 50
}

export default function JsonFileCard({ content, streaming = false }: Props) {
  const [open, setOpen] = useState(false)
  const json = extractJson(content)
  const valid = isValidJson(json)
  const filename = 'automation-rules.json'
  const size = new TextEncoder().encode(json).length

  useEffect(() => {
    if (!open) return

    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setOpen(false)
    }
    window.addEventListener('keydown', closeOnEscape)
    return () => window.removeEventListener('keydown', closeOnEscape)
  }, [open])

  const download = () => {
    if (!valid || streaming) return

    const url = URL.createObjectURL(new Blob([json], { type: 'application/json;charset=utf-8' }))
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    link.click()
    URL.revokeObjectURL(url)
  }

  return (
    <>
      <button
        type="button"
        className="group flex w-full min-w-64 items-center gap-3 rounded-xl border border-surface-600 bg-surface-900 p-3 text-left transition-colors hover:border-blue-300 hover:bg-blue-50/50 focus:outline-none focus:ring-2 focus:ring-blue-200"
        onClick={() => setOpen(true)}
      >
        <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-blue-100 text-accent-600">
          <Icon data={FileCode} size={23} />
        </span>
        <span className="min-w-0 flex-1">
          <span className="block truncate text-sm font-medium text-slate-100">{filename}</span>
          <span className="mt-0.5 flex items-center gap-1.5 text-xs text-slate-500">
            {streaming ? (
              <>
                <Loader size="s" />
                Модель ещё формирует JSON
              </>
            ) : valid ? (
              `JSON · ${formatSize(size)}`
            ) : (
              'JSON не завершён'
            )}
          </span>
        </span>
      </button>

      {open && (
        <div className="fixed inset-0 z-[1000]" role="dialog" aria-modal="true" aria-label={`Предпросмотр ${filename}`}>
          <button
            type="button"
            className="absolute inset-0 bg-slate-100/30"
            onClick={() => setOpen(false)}
            aria-label="Закрыть предпросмотр"
          />
          <aside className="absolute inset-y-0 right-0 flex w-full max-w-3xl flex-col border-l border-surface-600 bg-white shadow-2xl">
            <header className="flex shrink-0 items-center gap-3 border-b border-surface-700 px-5 py-4">
              <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-blue-50 text-accent-600">
                <Icon data={FileCode} size={19} />
              </span>
              <div className="min-w-0">
                <h2 className="truncate text-sm font-semibold text-slate-100">{filename}</h2>
                <p className="text-xs text-slate-500">
                  {streaming ? 'Файл ещё формируется' : valid ? `JSON · ${formatSize(size)}` : 'Некорректный или незавершённый JSON'}
                </p>
              </div>
              <div className="ml-auto flex items-center gap-2">
                <Button view="action" onClick={download} disabled={streaming || !valid}>
                  <Button.Icon><Icon data={ArrowDownToLine} size={16} /></Button.Icon>
                  Скачать
                </Button>
                <Button view="flat" onClick={() => setOpen(false)} aria-label="Закрыть">
                  <Icon data={Xmark} size={18} />
                </Button>
              </div>
            </header>

            {streaming && (
              <div className="flex shrink-0 items-center gap-2 border-b border-blue-100 bg-blue-50 px-5 py-2.5 text-xs text-accent-600">
                <Loader size="s" />
                Ответ ещё не закончен — предпросмотр обновляется
              </div>
            )}

            <div className="min-h-0 flex-1 overflow-auto bg-surface-950 p-5">
              {json ? (
                <HighlightedJson value={json} />
              ) : (
                <div className="flex h-full items-center justify-center text-sm text-slate-500">
                  Ожидаем содержимое JSON…
                </div>
              )}
            </div>
          </aside>
        </div>
      )}
    </>
  )
}

function extractJson(content: string): string {
  let value = content.trimStart()
  if (!value.startsWith('```')) return value.trim()

  const lineEnd = value.indexOf('\n')
  if (lineEnd === -1) return ''

  value = value.slice(lineEnd + 1)
  return value.replace(/\n?```\s*$/, '').trim()
}

function isValidJson(value: string): boolean {
  if (!value) return false
  try {
    JSON.parse(value)
    return true
  } catch {
    return false
  }
}

function HighlightedJson({ value }: { value: string }) {
  const tokens = useMemo(() => tokenizeJson(value), [value])

  return (
    <pre className="m-0 min-w-max whitespace-pre font-mono text-[13px] leading-6 text-slate-300">
      {tokens.map((token, index) => (
        <span key={index} className={token.className}>{token.value}</span>
      ))}
    </pre>
  )
}

function tokenizeJson(value: string): Array<{ value: string; className: string }> {
  const pattern = /("(?:\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*"\s*:)|("(?:\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*")|\b(true|false|null)\b|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?|[{}[\],:]/g
  const tokens: Array<{ value: string; className: string }> = []
  let cursor = 0

  for (const match of value.matchAll(pattern)) {
    const position = match.index ?? 0
    if (position > cursor) tokens.push({ value: value.slice(cursor, position), className: '' })

    const token = match[0]
    let className = 'text-slate-500'
    if (match[1]) className = 'text-accent-600'
    else if (match[2]) className = 'text-slate-200'
    else if (/^(true|false|null)$/.test(token)) className = 'text-accent-500'
    else if (/^-?\d/.test(token)) className = 'text-slate-400'

    tokens.push({ value: token, className })
    cursor = position + token.length
  }

  if (cursor < value.length) tokens.push({ value: value.slice(cursor), className: '' })
  return tokens
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} Б`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} КБ`
  return `${(bytes / (1024 * 1024)).toFixed(1)} МБ`
}
