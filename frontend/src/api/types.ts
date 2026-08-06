export interface User {
  id: string
  login: string
}

export interface Document {
  id: string
  title: string
  categories: string[]
  body: string
  createdAt: string
  updatedAt: string
}

export interface KBSnapshot {
  id: string
  contentHash: string
  cacheKey: string
  documentsCount: number
  charsCount: number
  isActive: boolean
  warmedAt: string | null
  warmError: string | null
  promptTokens: number | null
  cachedTokens: number | null
  createdAt: string
}

export interface KBStatus {
  snapshot: KBSnapshot | null
  documentsCount: number
  lastDocumentUpdate: string | null
  currentHash: string
  currentCharsCount: number
  stale: boolean
  model: string
  cacheMode: string
  cacheTtl: string
  reasoningEffort: string
  openrouterKeySet: boolean
}

export interface KBSyncResult {
  snapshot: KBSnapshot
  warmed: boolean
  warmSkipped?: string
  warmError?: string
}

export interface KnowledgeBaseInfo {
  snapshotId: string
  documentsCount: number
  charsCount: number
  warmedAt: string | null
  createdAt: string
}

export interface Chat {
  id: string
  kbSnapshotId: string | null
  title: string
  model: string
  createdAt: string
  updatedAt: string
}

export interface Attachment {
  filename: string
  size: number
  content?: string
}

export interface Usage {
  promptTokens: number
  completionTokens: number
  totalTokens: number
  cachedTokens: number
  reasoningTokens: number
  cost?: number
}

export interface Message {
  id: string
  chatId: string
  role: 'user' | 'assistant' | 'system'
  content: string
  attachments: Attachment[]
  usage: Usage | null
  error: string | null
  createdAt: string
}

export interface ChatDetail extends Chat {
  knowledgeBase: KnowledgeBaseInfo | null
  messages: Message[]
}
