import { useRef, useState, type ChangeEvent, type KeyboardEvent } from 'react'

interface Props {
  disabled: boolean
  streaming: boolean
  onSend: (text: string, files: File[]) => void
  onStop: () => void
}

const ACCEPT = '.txt,.md,.csv,.json,.yaml,.yml,.log,.xml,.sql,.ini,.conf'

export default function Composer({ disabled, streaming, onSend, onStop }: Props) {
  const [text, setText] = useState('')
  const [files, setFiles] = useState<File[]>([])
  const fileInputRef = useRef<HTMLInputElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const submit = () => {
    if (disabled || streaming) return
    if (!text.trim() && files.length === 0) return
    onSend(text, files)
    setText('')
    setFiles([])
    if (textareaRef.current) textareaRef.current.style.height = 'auto'
  }

  const handleKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      submit()
    }
  }

  const handleFiles = (event: ChangeEvent<HTMLInputElement>) => {
    const selected = Array.from(event.target.files ?? [])
    setFiles((current) => [...current, ...selected])
    event.target.value = ''
  }

  return (
    <div className="border-t border-surface-700 bg-surface-900/60 p-3">
      {files.length > 0 && (
        <div className="mb-2 flex flex-wrap gap-1.5">
          {files.map((file, index) => (
            <span key={`${file.name}-${index}`} className="badge gap-1.5">
              <span className="font-mono">{file.name}</span>
              <button
                className="text-slate-500 hover:text-slate-200"
                onClick={() => setFiles(files.filter((_, i) => i !== index))}
                aria-label={`Убрать ${file.name}`}
              >
                ×
              </button>
            </span>
          ))}
        </div>
      )}

      <div className="flex items-end gap-2">
        <button
          className="btn-secondary h-10 px-3"
          onClick={() => fileInputRef.current?.click()}
          disabled={disabled}
          title="Прикрепить текстовые файлы"
        >
          Файл
        </button>
        <input
          ref={fileInputRef}
          type="file"
          multiple
          accept={ACCEPT}
          className="hidden"
          onChange={handleFiles}
        />

        <textarea
          ref={textareaRef}
          className="input max-h-48 min-h-10 flex-1 resize-none py-2.5"
          rows={1}
          placeholder={disabled ? 'Создайте чат, чтобы начать' : 'Сообщение… Enter — отправить, Shift+Enter — перенос'}
          value={text}
          disabled={disabled}
          onChange={(event) => {
            setText(event.target.value)
            const el = event.target
            el.style.height = 'auto'
            el.style.height = `${Math.min(el.scrollHeight, 192)}px`
          }}
          onKeyDown={handleKeyDown}
        />

        {streaming ? (
          <button className="btn-danger h-10" onClick={onStop}>
            Остановить
          </button>
        ) : (
          <button className="btn-primary h-10" onClick={submit} disabled={disabled}>
            Отправить
          </button>
        )}
      </div>

      <p className="mt-1.5 px-1 text-[11px] text-slate-600">
        Поддерживаются текстовые файлы: {ACCEPT.replaceAll(',', ' ')}
      </p>
    </div>
  )
}
