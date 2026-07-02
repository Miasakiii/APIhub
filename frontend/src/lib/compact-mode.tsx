import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'

type CompactMode = 'full' | 'icons' | 'hidden'

interface CompactModeContextValue {
  mode: CompactMode
  setMode: (mode: CompactMode) => void
  toggle: () => void
}

export const CompactModeContext = createContext<CompactModeContextValue>({
  mode: 'full',
  setMode: () => {},
  toggle: () => {},
})

export function CompactModeProvider({ children }: { children: ReactNode }) {
  const [mode, setModeState] = useState<CompactMode>(() => {
    const saved = localStorage.getItem('apihub-compact-mode')
    if (saved === 'full' || saved === 'icons' || saved === 'hidden') return saved
    return 'full'
  })

  useEffect(() => {
    localStorage.setItem('apihub-compact-mode', mode)
  }, [mode])

  const setMode = (m: CompactMode) => setModeState(m)
  const toggle = () => setModeState((prev) => {
    if (prev === 'full') return 'icons'
    if (prev === 'icons') return 'hidden'
    return 'full'
  })

  return (
    <CompactModeContext.Provider value={{ mode, setMode, toggle }}>
      {children}
    </CompactModeContext.Provider>
  )
}

export function useCompactMode() {
  return useContext(CompactModeContext)
}

export type { CompactMode }