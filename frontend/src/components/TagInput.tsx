import { useState, type KeyboardEvent } from 'react'

interface Props {
  values: string[]
  suggestions: string[]
  onChange: (values: string[]) => void
}

export default function TagInput({ values, suggestions, onChange }: Props) {
  const [input, setInput] = useState('')

  const add = (raw: string) => {
    const value = raw.trim()
    if (!value) return
    if (values.some((item) => item.toLowerCase() === value.toLowerCase())) {
      setInput('')
      return
    }
    onChange([...values, value])
    setInput('')
  }

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter' || event.key === ',') {
      event.preventDefault()
      add(input)
    } else if (event.key === 'Backspace' && input === '' && values.length > 0) {
      onChange(values.slice(0, -1))
    }
  }

  const unused = suggestions.filter(
    (item) => !values.some((value) => value.toLowerCase() === item.toLowerCase()),
  )

  return (
    <div className="space-y-1.5">
      <div className="flex flex-wrap items-center gap-1.5 rounded-lg border border-surface-600 bg-white px-2 py-1.5 shadow-sm focus-within:border-accent-500 focus-within:ring-2 focus-within:ring-blue-100">
        {values.map((value) => (
          <span key={value} className="badge gap-1 border-blue-200 bg-blue-50 text-accent-600">
            {value}
            <button
              className="text-slate-500 hover:text-slate-200"
              onClick={() => onChange(values.filter((item) => item !== value))}
              aria-label={`Убрать категорию ${value}`}
            >
              ×
            </button>
          </span>
        ))}
        <input
          className="min-w-32 flex-1 bg-transparent px-1 py-0.5 text-sm text-slate-100 placeholder:text-slate-500 focus:outline-none"
          placeholder="Категории: Enter или запятая"
          value={input}
          onChange={(event) => setInput(event.target.value)}
          onKeyDown={handleKeyDown}
          onBlur={() => add(input)}
        />
      </div>

      {unused.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {unused.slice(0, 12).map((item) => (
            <button
              key={item}
              className="rounded-md border border-dashed border-surface-600 px-2 py-0.5 text-xs text-slate-500 hover:text-slate-200"
              onClick={() => add(item)}
            >
              + {item}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
