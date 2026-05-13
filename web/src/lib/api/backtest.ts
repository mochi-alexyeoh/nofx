import type {
  BacktestRunsResponse,
  BacktestStartConfig,
  BacktestRunMetadata,
  BacktestStatusPayload,
  BacktestEquityPoint,
  BacktestTradeEvent,
  BacktestMetrics,
  BacktestKlinesResponse,
  DecisionRecord,
} from '../../types'
import { API_BASE, httpClient } from './helpers'

export const backtestApi = {
  async getBacktestRuns(params?: {
    state?: string
    search?: string
    limit?: number
    offset?: number
  }): Promise<BacktestRunsResponse> {
    const query = new URLSearchParams()
    if (params?.state) query.set('state', params.state)
    if (params?.search) query.set('search', params.search)
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const result = await httpClient.get<BacktestRunsResponse>(
      `${API_BASE}/backtest/runs${query.toString() ? `?${query}` : ''}`
    )
    if (!result.success || !result.data) throw new Error('Failed to fetch backtest runs')
    return result.data
  },

  async startBacktest(config: BacktestStartConfig): Promise<BacktestRunMetadata> {
    const result = await httpClient.post<BacktestRunMetadata>(`${API_BASE}/backtest/start`, { config })
    if (!result.success || !result.data) throw new Error('Failed to start backtest')
    return result.data
  },

  async pauseBacktest(runId: string): Promise<BacktestRunMetadata> {
    const result = await httpClient.post<BacktestRunMetadata>(`${API_BASE}/backtest/pause`, { run_id: runId })
    if (!result.success || !result.data) throw new Error('Failed to pause backtest')
    return result.data
  },

  async resumeBacktest(runId: string): Promise<BacktestRunMetadata> {
    const result = await httpClient.post<BacktestRunMetadata>(`${API_BASE}/backtest/resume`, { run_id: runId })
    if (!result.success || !result.data) throw new Error('Failed to resume backtest')
    return result.data
  },

  async stopBacktest(runId: string): Promise<BacktestRunMetadata> {
    const result = await httpClient.post<BacktestRunMetadata>(`${API_BASE}/backtest/stop`, { run_id: runId })
    if (!result.success || !result.data) throw new Error('Failed to stop backtest')
    return result.data
  },

  async updateBacktestLabel(runId: string, label: string): Promise<BacktestRunMetadata> {
    const result = await httpClient.post<BacktestRunMetadata>(`${API_BASE}/backtest/label`, { run_id: runId, label })
    if (!result.success || !result.data) throw new Error('Failed to update backtest label')
    return result.data
  },

  async deleteBacktestRun(runId: string): Promise<void> {
    const result = await httpClient.post(`${API_BASE}/backtest/delete`, { run_id: runId })
    if (!result.success) throw new Error(result.message || 'Failed to delete backtest run')
  },

  async getBacktestStatus(runId: string): Promise<BacktestStatusPayload> {
    const result = await httpClient.get<BacktestStatusPayload>(`${API_BASE}/backtest/status?run_id=${runId}`)
    if (!result.success || !result.data) throw new Error('Failed to fetch backtest status')
    return result.data
  },

  async getBacktestEquity(runId: string, timeframe?: string, limit?: number): Promise<BacktestEquityPoint[]> {
    const query = new URLSearchParams({ run_id: runId })
    if (timeframe) query.set('tf', timeframe)
    if (limit) query.set('limit', String(limit))
    const result = await httpClient.get<BacktestEquityPoint[]>(`${API_BASE}/backtest/equity?${query}`)
    if (!result.success || !result.data) throw new Error('Failed to fetch backtest equity')
    return result.data
  },

  async getBacktestTrades(runId: string, limit = 200): Promise<BacktestTradeEvent[]> {
    const query = new URLSearchParams({ run_id: runId, limit: String(limit) })
    const result = await httpClient.get<BacktestTradeEvent[]>(`${API_BASE}/backtest/trades?${query}`)
    if (!result.success || !result.data) throw new Error('Failed to fetch backtest trades')
    return result.data
  },

  async getBacktestMetrics(runId: string): Promise<BacktestMetrics> {
    const result = await httpClient.get<BacktestMetrics>(`${API_BASE}/backtest/metrics?run_id=${runId}`)
    if (!result.success || !result.data) throw new Error('Failed to fetch backtest metrics')
    return result.data
  },

  async getBacktestKlines(runId: string, symbol: string, timeframe?: string): Promise<BacktestKlinesResponse> {
    const query = new URLSearchParams({ run_id: runId, symbol })
    if (timeframe) query.set('timeframe', timeframe)
    const result = await httpClient.get<BacktestKlinesResponse>(`${API_BASE}/backtest/klines?${query}`)
    if (!result.success || !result.data) throw new Error('Failed to fetch backtest klines')
    return result.data
  },

  async getBacktestTrace(runId: string, cycle?: number): Promise<DecisionRecord> {
    const query = new URLSearchParams({ run_id: runId })
    if (cycle !== undefined) query.set('cycle', String(cycle))
    const result = await httpClient.get<DecisionRecord>(`${API_BASE}/backtest/trace?${query}`)
    if (!result.success || !result.data) throw new Error('Failed to fetch backtest trace')
    return result.data
  },

  async getBacktestDecisions(runId: string, limit = 20, offset = 0): Promise<DecisionRecord[]> {
    const query = new URLSearchParams({ run_id: runId, limit: String(limit), offset: String(offset) })
    const result = await httpClient.get<DecisionRecord[]>(`${API_BASE}/backtest/decisions?${query}`)
    if (!result.success || !result.data) throw new Error('Failed to fetch backtest decisions')
    return result.data
  },

  async exportBacktest(runId: string): Promise<Blob> {
    const response = await fetch(`${API_BASE}/backtest/export?run_id=${encodeURIComponent(runId)}`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token') || ''}`,
      },
    })
    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || 'Failed to export backtest')
    }
    return response.blob()
  },
}
