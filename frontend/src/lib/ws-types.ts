// WebSocket message types matching the backend

export interface WSMessage {
  type: string
  timestamp: string
  data?: unknown
}

export interface UsageUpdateData {
  request_count: number
  input_tokens: number
  output_tokens: number
  cost_usd: number
}

export interface AlertData {
  level: string
  title: string
  message: string
}

export interface SyncProgressData {
  provider_id: string
  status: string
  progress: number
  processed_keys: number
  total_keys: number
}

export interface SyncCompleteData {
  provider_id: string
}

export interface SyncErrorData {
  provider_id: string
  error: string
}

export type MessageHandler = (msg: WSMessage) => void
