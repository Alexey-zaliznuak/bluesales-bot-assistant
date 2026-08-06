import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { api } from '../api/client'
import type { KBSyncResult } from '../api/types'

export default function KnowledgeBasePanel() {
  const queryClient = useQueryClient()
  const [result, setResult] = useState<KBSyncResult | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [previewOpen, setPreviewOpen] = useState(false)

  const statusQuery = useQuery({ queryKey: ['kb-status'], queryFn: api.kbStatus })
  const status = statusQuery.data

  const syncMutation = useMutation({
    mutationFn: api.kbSync,
    onSuccess: (data) => {
      setError(null)
      setResult(data)
      void queryClient.invalidateQueries({ queryKey: ['kb-status'] })
    },
    onError: (err: Error) => {
      setResult(null)
      setError(err.message)
    },
  })

  const snapshot = status?.snapshot ?? null

  return (
    <div className="border-b border-surface-700 bg-surface-900/50 px-4 py-3">
      <div className="flex flex-wrap items-center gap-3">
        <div className="flex items-center gap-2">
          <span
            className={`h-2 w-2 rounded-full ${
              !status ? 'bg-slate-600' : status.stale ? 'bg-amber-400' : 'bg-emerald-400'
            }`}
          />
          <span className="text-sm font-medium text-slate-200">База знаний</span>
        </div>

        <span className="text-xs text-slate-500">
          {status
            ? status.stale
              ? snapshot
                ? 'изменения не синхронизированы с OpenRouter'
                : 'ещё ни разу не синхронизирована'
              : 'синхронизирована'
            : '…'}
        </span>

        <div className="ml-auto flex items-center gap-2">
          <button className="btn-ghost text-xs" onClick={() => setPreviewOpen(true)}>
            Показать префикс
          </button>
          <button
            className="btn-primary"
            disabled={syncMutation.isPending || status?.documentsCount === 0}
            onClick={() => syncMutation.mutate()}
          >
            {syncMutation.isPending ? 'Синхронизация…' : 'Синхронизировать базу знаний'}
          </button>
        </div>
      </div>

      <div className="mt-2 flex flex-wrap gap-x-5 gap-y-1 text-xs text-slate-500">
        <span>
          Документов: <span className="text-slate-300">{status?.documentsCount ?? '—'}</span>
        </span>
        <span>
          Размер префикса:{' '}
          <span className="text-slate-300">
            {status ? `${status.currentCharsCount.toLocaleString('ru-RU')} симв.` : '—'}
          </span>
        </span>
        {snapshot && (
          <>
            <span>
              Снимок: <span className="font-mono text-slate-300">{snapshot.contentHash.slice(0, 12)}</span>
            </span>
            <span>
              Кэш-ключ: <span className="font-mono text-slate-300">{snapshot.cacheKey}</span>
            </span>
            <span>
              Прогрет:{' '}
              <span className="text-slate-300">
                {snapshot.warmedAt ? new Date(snapshot.warmedAt).toLocaleString('ru-RU') : 'нет'}
              </span>
            </span>
            {snapshot.promptTokens != null && (
              <span>
                Токенов в префиксе: <span className="text-slate-300">{snapshot.promptTokens.toLocaleString('ru-RU')}</span>
              </span>
            )}
          </>
        )}
      </div>

      {result && (
        <div
          className={`mt-2 rounded-lg border px-3 py-2 text-xs ${
            result.warmError
              ? 'border-amber-900/60 bg-amber-950/30 text-amber-300'
              : 'border-emerald-900/60 bg-emerald-950/30 text-emerald-300'
          }`}
        >
          {result.warmError
            ? `Снимок сохранён, но прогрев кэша не удался: ${result.warmError}`
            : result.warmSkipped
              ? `Снимок сохранён. ${result.warmSkipped}`
              : `Снимок сохранён и кэш префикса прогрет: ${result.snapshot.documentsCount} докум., ` +
                `${result.snapshot.promptTokens?.toLocaleString('ru-RU') ?? '?'} токенов в префиксе. ` +
                'Следующие запросы читают его из кэша по сниженной цене.'}
        </div>
      )}

      {(error || snapshot?.warmError) && !result && (
        <div className="mt-2 rounded-lg border border-red-900/60 bg-red-950/40 px-3 py-2 text-xs text-red-300">
          {error ?? snapshot?.warmError}
        </div>
      )}

      {previewOpen && <PreviewModal onClose={() => setPreviewOpen(false)} />}
    </div>
  )
}

function PreviewModal({ onClose }: { onClose: () => void }) {
  const previewQuery = useQuery({ queryKey: ['kb-preview'], queryFn: api.kbPreview })

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-6" onClick={onClose}>
      <div
        className="card flex max-h-[80vh] w-full max-w-4xl flex-col overflow-hidden"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b border-surface-700 px-4 py-3">
          <div>
            <h2 className="text-sm font-semibold text-slate-100">Кэшируемый префикс промпта</h2>
            <p className="text-xs text-slate-500">
              Ровно этот текст уходит первым system-блоком и кэшируется в OpenRouter
            </p>
          </div>
          <button className="btn-ghost" onClick={onClose}>
            Закрыть
          </button>
        </div>
        <pre className="min-h-0 flex-1 overflow-auto whitespace-pre-wrap break-words p-4 font-mono text-xs text-slate-300">
          {previewQuery.isLoading ? 'Загрузка…' : previewQuery.data?.content}
        </pre>
      </div>
    </div>
  )
}
