import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

import JsonFileCard from './JsonFileCard'

interface Props {
  content: string
  isUser?: boolean
  streaming?: boolean
}

type Segment =
  | { type: 'markdown'; content: string }
  | { type: 'json'; content: string }

export default function ChatMarkdown({ content, isUser = false, streaming = false }: Props) {
  const markdown = normalizeRawJson(content)
  const segments = splitLargeJsonBlocks(markdown)

  return (
    <div className={`chat-markdown ${isUser ? 'chat-markdown-user' : ''}`}>
      {segments.map((segment, index) =>
        segment.type === 'json' ? (
          <JsonFileCard key={index} content={segment.content} streaming={streaming} />
        ) : (
          <ReactMarkdown key={index} remarkPlugins={[remarkGfm]}>
            {segment.content}
          </ReactMarkdown>
        ),
      )}
    </div>
  )
}

function normalizeRawJson(content: string): string {
  const value = content.trimStart()
  if (!value.startsWith('{') && !value.startsWith('[')) return content
  return `\`\`\`json\n${content.trim()}\n\`\`\``
}

function splitLargeJsonBlocks(markdown: string): Segment[] {
  const pattern = /```json[^\r\n]*\r?\n([\s\S]*?)(?:```|$)/gi
  const segments: Segment[] = []
  let cursor = 0

  for (const match of markdown.matchAll(pattern)) {
    const json = match[1].replace(/\r?\n$/, '')
    if (countLines(json) <= 50) continue

    const position = match.index ?? 0
    if (position > cursor) {
      segments.push({ type: 'markdown', content: markdown.slice(cursor, position) })
    }
    segments.push({ type: 'json', content: json })
    cursor = position + match[0].length
  }

  if (cursor < markdown.length) {
    segments.push({ type: 'markdown', content: markdown.slice(cursor) })
  }

  return segments.length > 0 ? segments : [{ type: 'markdown', content: markdown }]
}

function countLines(value: string): number {
  return value === '' ? 0 : value.split(/\r?\n/).length
}
