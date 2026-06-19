import { useEffect, useRef, useState, useCallback, type ReactNode } from 'react'
import { getToken } from './auth'
import { WSContext, type WSContextValue } from './use-ws'
import type { WSMessage, MessageHandler } from './ws-types'

interface Props {
  children: ReactNode
}

export function WebSocketProvider({ children }: Props) {
  const [connected, setConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const handlersRef = useRef<Map<string, Set<MessageHandler>>>(new Map())
  const reconnectAttempts = useRef(0)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

  const subscribe = useCallback((type: string, handler: MessageHandler) => {
    let handlers = handlersRef.current.get(type)
    if (!handlers) {
      handlers = new Set()
      handlersRef.current.set(type, handlers)
    }
    handlers.add(handler)
    return () => { handlers!.delete(handler) }
  }, [])

  useEffect(() => {
    async function doConnect() {
      // Build WebSocket URL
      let wsHost = location.host
      // In Wails desktop mode, connect to local Gin server
      if (window.go?.main?.WailsApp) {
        try {
          const port = await window.go.main.WailsApp.GetAPIPort()
          wsHost = `127.0.0.1:${port}`
        } catch {
          wsHost = '127.0.0.1:8080'
        }
      }
      const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
      const token = getToken()
      let url = `${proto}//${wsHost}/ws`
      if (token) {
        url += `?token=${token}`
      }

      const ws = new WebSocket(url)
      wsRef.current = ws

      ws.onopen = () => {
        setConnected(true)
        reconnectAttempts.current = 0
      }

      ws.onmessage = (event) => {
        try {
          const msg: WSMessage = JSON.parse(event.data)
          // Dispatch to type-specific handlers
          const handlers = handlersRef.current.get(msg.type)
          if (handlers) {
            handlers.forEach(h => h(msg))
          }
          // Dispatch to wildcard handlers
          const wildcardHandlers = handlersRef.current.get('*')
          if (wildcardHandlers) {
            wildcardHandlers.forEach(h => h(msg))
          }
        } catch {
          // Ignore parse errors
        }
      }

      ws.onclose = () => {
        setConnected(false)
        wsRef.current = null

        // Exponential backoff reconnect
        const attempts = reconnectAttempts.current
        if (attempts < 10) {
          const delay = Math.min(1000 * Math.pow(1.5, attempts), 30000)
          reconnectAttempts.current = attempts + 1
          reconnectTimer.current = setTimeout(doConnect, delay)
        }
      }

      ws.onerror = () => {
        ws.close()
      }
    }

    doConnect()

    return () => {
      clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [])

  // Heartbeat: send ping every 30s
  useEffect(() => {
    const interval = setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ type: 'ping' }))
      }
    }, 30000)
    return () => clearInterval(interval)
  }, [])

  const value: WSContextValue = { connected, subscribe }

  return (
    <WSContext.Provider value={value}>
      {children}
    </WSContext.Provider>
  )
}
