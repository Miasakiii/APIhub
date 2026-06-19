import { clearToken, getAuthHeaders } from './lib/auth'

const BASE = '/api/v1'

type OnUnauthorized = () => void
let onUnauthorized: OnUnauthorized | null = null

export function setUnauthorizedHandler(handler: OnUnauthorized) {
  onUnauthorized = handler
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    ...getAuthHeaders(),
    ...(init?.headers as Record<string, string> | undefined),
  }
  if (init?.body && !headers['Content-Type']) {
    headers['Content-Type'] = 'application/json'
  }

  const res = await fetch(`${BASE}${path}`, { ...init, headers })

  if (res.status === 401) {
    clearToken()
    onUnauthorized?.()
    throw new Error('未授权，请重新登录')
  }

  if (!res.ok) {
    let msg = `${res.status} ${res.statusText}`
    try {
      const body = await res.json()
      if (body?.error) msg = body.error
    } catch {
      /* ignore */
    }
    throw new Error(msg)
  }

  if (res.status === 204) return undefined as T
  return res.json()
}

async function get<T>(path: string): Promise<T> {
  return request<T>(path)
}

async function post<T>(path: string, body?: unknown): Promise<T> {
  return request<T>(path, {
    method: 'POST',
    body: body ? JSON.stringify(body) : undefined,
  })
}

async function put<T>(path: string, body?: unknown): Promise<T> {
  return request<T>(path, {
    method: 'PUT',
    body: body ? JSON.stringify(body) : undefined,
  })
}

async function del(path: string): Promise<void> {
  await request<void>(path, { method: 'DELETE' })
}

// --- Providers ---
export interface Provider {
  id: string
  name: string
  type: string
  base_url?: string
  console_url?: string
  topup_url?: string
  docs_url?: string
  syncer?: string
  enabled: boolean
}

export interface APIKey {
  id: string
  provider_id: string
  key_hash: string
  name: string
  source: string
  status: string
  balance_usd: number
  last_checked?: string
}

export interface UsageRecord {
  id: string
  provider_id: string
  model: string
  input_tokens: number
  output_tokens: number
  cache_read: number
  cache_create: number
  cost_usd: number
  source: string
  timestamp: string
}

export interface DailyStats {
  id: string
  provider_id: string
  model: string
  source: string
  date: string
  request_count: number
  input_tokens: number
  output_tokens: number
  cache_read: number
  cache_create: number
  cost_usd: number
}

export interface Alert {
  id: string
  name: string
  type: string
  provider_id?: string
  api_key_id?: string
  threshold: number
  unit?: string
  enabled: boolean
  last_triggered_at?: string
  created_at?: string
}

export interface AlertHistory {
  id: string
  alert_id: string
  message: string
  level: string
  created_at: string
}

export interface Subscription {
  id: string
  provider_id: string
  plan_name: string
  price: number
  currency: string
  billing_cycle: string
  quota_type: string
  quota_total: number
  quota_used: number
  start_date?: string
  renew_date?: string
  auto_renew: boolean
  status: string
  source?: string
  notes?: string
  created_at?: string
  provider?: Provider
}

export interface SyncState {
  id: string
  source: string
  last_sync?: string
  offset: number
  status: string
  error?: string
}

export interface UsageSession {
  id: string
  provider_id: string
  model: string
  source: string
  started_at: string
  ended_at: string
  duration_ms: number
  request_count: number
  input_tokens: number
  output_tokens: number
  cache_read: number
  cache_create: number
  cost_usd: number
}

export interface ActivityBucket {
  id: string
  bucket_start: string
  provider_id: string
  model: string
  request_count: number
  input_tokens: number
  output_tokens: number
  cache_read: number
  cache_create: number
  cost_usd: number
}

export interface SessionStats {
  total_sessions: number
  avg_duration_ms: number
  avg_cost_usd: number
  total_cost_usd: number
  total_requests: number
}

export interface Agent {
  id: string
  name: string
  type: string
  icon?: string
  created_at?: string
  updated_at?: string
}

export interface ScanFinding {
  provider_type: string
  name: string
  base_url: string
  masked_key: string
  source: string
  config_path: string
}

export interface ScanImportResult {
  name: string
  provider_id?: string
  key_id?: string
  status: 'created' | 'skipped' | 'error'
  message?: string
}

export interface AuthConfig {
  enabled: boolean
  allow_register: boolean
}

export const api = {
  auth: {
    config: () => get<AuthConfig>('/auth/config'),
    register: (username: string, password: string) =>
      post<{ id: string; username: string; token?: string }>('/auth/register', { username, password }),
    login: (username: string, password: string) =>
      post<{ token: string; username: string }>('/auth/login', { username, password }),
    me: () => get<{ id: string } | { auth: string }>('/auth/me'),
  },

  providers: {
    list: () => get<Provider[]>('/providers'),
    create: (p: Partial<Provider>) => post<Provider>('/providers', p),
    delete: (id: string) => del(`/providers/${id}`),
    detail: (id: string) => get<{ provider: Provider; keys: APIKey[]; total_cost: number; total_requests: number }>(`/providers/${id}`),
  },

  keys: {
    list: () => get<APIKey[]>('/keys'),
    create: (provider_id: string, key: string, name: string) => post('/keys', { provider_id, key, name }),
    decrypt: (id: string) => get<{ key: string }>(`/keys/${id}/decrypt`),
    revoke: (id: string) => post(`/keys/${id}/revoke`),
    delete: (id: string) => del(`/keys/${id}`),
  },

  usage: {
    summary: () => get<{ total_cost_usd: number; total_tokens: number; total_requests: number; unique_models: number; unique_keys: number }>('/usage/summary'),
    list: (params?: { provider_id?: string; model?: string; source?: string; date?: string; page?: number; page_size?: number }) => {
      const q = new URLSearchParams(Object.entries(params || {}).filter(([, v]) => v) as [string, string][])
      return get<{ records: UsageRecord[]; total: number; page: number; page_size: number }>(`/usage?${q}`)
    },
  },

  stats: {
    daily: (params?: { model?: string; provider_id?: string }) => {
      const q = new URLSearchParams(Object.entries(params || {}).filter(([, v]) => v) as [string, string][])
      return get<DailyStats[]>(`/stats/daily?${q}`)
    },
    costTrend: () => get<Array<{ date: string; cost_usd: number; tokens: number; request_count: number }>>('/stats/cost-trend'),
    modelBreakdown: () => get<Array<{ model: string; total_cost_usd: number; total_tokens: number; request_count: number }>>('/stats/model-breakdown'),
  },

  alerts: {
    list: () => get<Alert[]>('/alerts'),
    create: (a: Partial<Alert>) => post<Alert>('/alerts', a),
    update: (id: string, a: Partial<Alert>) => put(`/alerts/${id}`, a),
    delete: (id: string) => del(`/alerts/${id}`),
    history: (alert_id?: string) => {
      const q = alert_id ? `?alert_id=${alert_id}` : ''
      return get<AlertHistory[]>(`/alerts/history${q}`)
    },
  },

  subscriptions: {
    list: () => get<Subscription[]>('/subscriptions'),
    create: (s: Partial<Subscription>) => post<Subscription>('/subscriptions', s),
    get: (id: string) => get<Subscription>(`/subscriptions/${id}`),
    update: (id: string, s: Partial<Subscription>) => put(`/subscriptions/${id}`, s),
    delete: (id: string) => del(`/subscriptions/${id}`),
  },

  frequency: {
    hourly: () => get<{ heatmap: number[][]; days: number }>('/frequency/hourly'),
    peakQps: () => get<{ peak_qps: number; peak_minute: string; avg_qps: number; peak_count: number }>('/frequency/peak-qps'),
    today: () => get<{ hourly: number[]; date: string }>('/frequency/today'),
  },

  sync: {
    status: () => get<SyncState[]>('/sync/status'),
    trigger: (provider_id: string) => post(`/sync/${provider_id}`),
    triggerCCSwitch: () => post('/sync/ccswitch'),
    syncers: () => get<{ syncers: string[] }>('/syncers'),
  },

  export: {
    csv: async () => {
      const headers = getAuthHeaders()
      const res = await fetch(`${BASE}/export/csv`, { headers })
      if (!res.ok) throw new Error('导出失败')
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `apihub-usage-${new Date().toISOString().slice(0, 10)}.csv`
      a.click()
      URL.revokeObjectURL(url)
    },
  },

  sessions: {
    list: (params?: { provider_id?: string; model?: string; source?: string; from?: string; to?: string; page?: number; page_size?: number }) => {
      const q = new URLSearchParams(Object.entries(params || {}).filter(([, v]) => v) as [string, string][])
      return get<{ sessions: UsageSession[]; total: number; page: number; page_size: number }>(`/sessions?${q}`)
    },
    stats: () => get<SessionStats>('/sessions/stats'),
    buckets: (params?: { from?: string; to?: string; provider_id?: string; model?: string }) => {
      const q = new URLSearchParams(Object.entries(params || {}).filter(([, v]) => v) as [string, string][])
      return get<{ buckets: ActivityBucket[] }>(`/sessions/buckets?${q}`)
    },
    hourly: (date?: string) => {
      const q = date ? `?date=${date}` : ''
      return get<{ date: string; buckets: ActivityBucket[] }>(`/sessions/hourly${q}`)
    },
  },

  scan: {
    run: () => post<{ findings: ScanFinding[]; total: number }>('/scan'),
    import: (indices?: number[]) =>
      post<{ results: ScanImportResult[]; total: number }>('/scan/import', { indices: indices || [] }),
  },

  agents: {
    list: () => get<Agent[]>('/agents'),
    create: (a: Partial<Agent>) => post<Agent>('/agents', a),
    get: (id: string) => get<Agent>(`/agents/${id}`),
    update: (id: string, a: Partial<Agent>) => put(`/agents/${id}`, a),
    delete: (id: string) => del(`/agents/${id}`),
  },
}
