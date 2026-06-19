import { createContext, useContext, useEffect } from 'react'
import type { MessageHandler } from './ws-types'

export interface WSContextValue {
  connected: boolean
  subscribe: (type: string, handler: MessageHandler) => () => void
}

export const WSContext = createContext<WSContextValue>({
  connected: false,
  subscribe: () => () => {},
})

export function useWebSocket() {
  return useContext(WSContext)
}

// Convenience hook for subscribing to a specific message type
export function useWSMessage(type: string, handler: MessageHandler) {
  const { subscribe } = useWebSocket()
  useEffect(() => subscribe(type, handler), [type, handler, subscribe])
}
