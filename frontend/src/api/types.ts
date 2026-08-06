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
  documentsCount: number
  charsCount: number
  isActive: boolean
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
  reasoningEffort: string
  openrouterKeySet: boolean
}

export interface KBSyncResult {
  snapshot: KBSnapshot
}

export interface KnowledgeBaseInfo {
  snapshotId: string
  documentsCount: number
  charsCount: number
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
