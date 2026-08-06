import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { api } from '../api/client'
import type { Document } from '../api/types'
import KnowledgeBasePanel from '../components/KnowledgeBasePanel'
import TagInput from '../components/TagInput'

interface Draft {
  title: string
  categories: string[]
  body: string
}

const emptyDraft: Draft = { title: '', categories: [], body: '' }

export default function DocumentsPage() {
  const queryClient = useQueryClient()
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState('')
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [draft, setDraft] = useState<Draft>(emptyDraft)
  const [error, setError] = useState<string | null>(null)

  const documentsQuery = useQuery({
    queryKey: ['documents', search, category],
    queryFn: () => api.listDocuments({ search, category }),
  })
  const categoriesQuery = useQuery({ queryKey: ['categories'], queryFn: api.listCategories })

  const selected = useMemo(
    () => documentsQuery.data?.find((doc) => doc.id === selectedId) ?? null,
    [documentsQuery.data, selectedId],
  )

  useEffect(() => {
    if (selected) {
      setDraft({ title: selected.title, categories: selected.categories, body: selected.body })
    }
  }, [selected])

  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: ['documents'] })
    void queryClient.invalidateQueries({ queryKey: ['categories'] })
    void queryClient.invalidateQueries({ queryKey: ['kb-status'] })
  }

  const saveMutation = useMutation({
    mutationFn: async () => {
      if (selectedId) return api.updateDocument(selectedId, draft)
      return api.createDocument(draft)
    },
    onSuccess: (doc: Document) => {
      setError(null)
      setSelectedId(doc.id)
      invalidate()
    },
    onError: (err: Error) => setError(err.message),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteDocument(id),
    onSuccess: () => {
      setSelectedId(null)
      setDraft(emptyDraft)
      invalidate()
    },
    onError: (err: Error) => setError(err.message),
  })

  const startNew = () => {
    setSelectedId(null)
    setDraft(emptyDraft)
    setError(null)
  }

  const dirty = selected
    ? selected.title !== draft.title ||
      selected.body !== draft.body ||
      selected.categories.join('\u0000') !== draft.categories.join('\u0000')
    : draft.title.trim() !== '' || draft.body !== '' || draft.categories.length > 0

  return (
    <div className="flex h-full min-h-0">
      <aside className="flex w-80 shrink-0 flex-col border-r border-surface-700 bg-surface-900/50">
        <div className="space-y-2 border-b border-surface-700 p-3">
          <button className="btn-primary w-full" onClick={startNew}>
            + Новый документ
          </button>
          <input
            className="input"
            placeholder="Поиск по заголовку и тексту"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
          <div className="flex flex-wrap gap-1">
            <CategoryChip active={category === ''} onClick={() => setCategory('')} label="Все" />
            {categoriesQuery.data?.map((item) => (
              <CategoryChip
                key={item}
                active={category === item}
                onClick={() => setCategory(category === item ? '' : item)}
                label={item}
              />
            ))}
          </div>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto p-2">
          {documentsQuery.isLoading && <p className="p-2 text-sm text-slate-500">Загрузка…</p>}
          {documentsQuery.data?.length === 0 && (
            <p className="p-2 text-sm text-slate-500">Документов нет</p>
          )}
          {documentsQuery.data?.map((doc) => (
            <button
              key={doc.id}
              onClick={() => setSelectedId(doc.id)}
              className={`mb-1 w-full rounded-lg px-3 py-2 text-left transition-colors ${
                doc.id === selectedId ? 'bg-surface-700' : 'hover:bg-surface-800'
              }`}
            >
              <div className="truncate text-sm text-slate-200">{doc.title}</div>
              <div className="mt-0.5 truncate text-xs text-slate-500">
                {doc.categories.length > 0 ? doc.categories.join(' · ') : 'без категорий'}
              </div>
            </button>
          ))}
        </div>
      </aside>

      <section className="flex min-h-0 min-w-0 flex-1 flex-col">
        <KnowledgeBasePanel />

        <div className="flex min-h-0 flex-1 flex-col gap-3 p-4">
          <div className="flex items-center gap-2">
            <input
              className="input flex-1 text-base"
              placeholder="Человекочитаемый заголовок документа"
              value={draft.title}
              onChange={(event) => setDraft({ ...draft, title: event.target.value })}
            />
            <button
              className="btn-primary"
              disabled={!draft.title.trim() || !dirty || saveMutation.isPending}
              onClick={() => saveMutation.mutate()}
            >
              {saveMutation.isPending ? 'Сохраняем…' : selectedId ? 'Сохранить' : 'Создать'}
            </button>
            {selectedId && (
              <button
                className="btn-danger"
                disabled={deleteMutation.isPending}
                onClick={() => {
                  if (confirm('Удалить документ?')) deleteMutation.mutate(selectedId)
                }}
              >
                Удалить
              </button>
            )}
          </div>

          <TagInput
            values={draft.categories}
            suggestions={categoriesQuery.data ?? []}
            onChange={(categories) => setDraft({ ...draft, categories })}
          />

          {error && (
            <div className="rounded-lg border border-red-900/60 bg-red-950/40 px-3 py-2 text-sm text-red-300">
              {error}
            </div>
          )}

          <textarea
            className="input min-h-0 flex-1 resize-none font-mono text-[13px] leading-relaxed"
            placeholder="Тело документа: пример настройки бота в BlueSales"
            value={draft.body}
            onChange={(event) => setDraft({ ...draft, body: event.target.value })}
          />

          <div className="flex justify-between text-xs text-slate-500">
            <span>{draft.body.length.toLocaleString('ru-RU')} символов</span>
            {dirty && <span className="text-amber-400">есть несохранённые изменения</span>}
          </div>
        </div>
      </section>
    </div>
  )
}

function CategoryChip({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className={`rounded-md border px-2 py-0.5 text-xs transition-colors ${
        active
          ? 'border-accent-500 bg-accent-500/15 text-accent-400'
          : 'border-surface-600 bg-surface-800 text-slate-400 hover:text-slate-200'
      }`}
    >
      {label}
    </button>
  )
}
