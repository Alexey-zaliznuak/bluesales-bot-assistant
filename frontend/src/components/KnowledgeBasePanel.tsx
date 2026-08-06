import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, Icon } from '@gravity-ui/uikit'
import { ArrowRotateLeft, Eye } from '@gravity-ui/icons'

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
    <div className="border-b border-surface-700 bg-white px-5 py-4">
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
                ? 'изменения не включены в активный снимок'
                : 'ещё ни разу не синхронизирована'
              : 'синхронизирована'
            : '…'}
        </span>

        <div className="ml-auto flex items-center gap-2">
          <Button view="flat-secondary" size="m" onClick={() => setPreviewOpen(true)}>
            <Button.Icon><Icon data={Eye} size={16} /></Button.Icon>
            Показать префикс
          </Button>
          <Button
            view="action"
            size="m"
            disabled={syncMutation.isPending || status?.documentsCount === 0}
            loading={syncMutation.isPending}
            onClick={() => syncMutation.mutate()}
          >
            <Button.Icon><Icon data={ArrowRotateLeft} size={16} /></Button.Icon>
            {syncMutation.isPending ? 'Синхронизация…' : 'Синхронизировать базу знаний'}
          </Button>
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
          </>
        )}
      </div>

      {result && (
        <div className="mt-2 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-xs text-emerald-700">
          Снимок сохранён: {result.snapshot.documentsCount} докум.,{' '}
          {result.snapshot.charsCount.toLocaleString('ru-RU')} симв.
        </div>
      )}

      {error && !result && (
        <div className="mt-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700">
          {error}
        </div>
      )}

      {previewOpen && <PreviewModal onClose={() => setPreviewOpen(false)} />}
    </div>
  )
}

function PreviewModal({ onClose }: { onClose: () => void }) {
  const previewQuery = useQuery({ queryKey: ['kb-preview'], queryFn: api.kbPreview })

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-100/35 p-6 backdrop-blur-sm" onClick={onClose}>
      <div
        className="card flex max-h-[80vh] w-full max-w-4xl flex-col overflow-hidden"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b border-surface-700 px-4 py-3">
          <div>
            <h2 className="text-sm font-semibold text-slate-100">Контекст базы знаний</h2>
            <p className="text-xs text-slate-500">
              Этот текст отправляется модели первым системным сообщением
            </p>
          </div>
          <Button view="flat-secondary" size="m" onClick={onClose}>
            Закрыть
          </Button>
        </div>
        <pre className="min-h-0 flex-1 overflow-auto whitespace-pre-wrap break-words p-4 font-mono text-xs text-slate-300">
          {previewQuery.isLoading ? 'Загрузка…' : previewQuery.data?.content}
        </pre>
      </div>
    </div>
  )
}
