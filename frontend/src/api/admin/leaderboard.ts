import { apiClient } from '../client'
import type { SimpleUser } from './usage'

export interface LeaderboardSettings {
  ignored_user_ids: number[]
  ignored_users: SimpleUser[]
}

export async function getSettings(): Promise<LeaderboardSettings> {
  const { data } = await apiClient.get<LeaderboardSettings>('/admin/leaderboard/settings')
  return data
}

export async function updateSettings(payload: {
  ignored_user_ids: number[]
}): Promise<LeaderboardSettings> {
  const { data } = await apiClient.put<LeaderboardSettings>('/admin/leaderboard/settings', payload)
  return data
}

export const adminLeaderboardAPI = {
  getSettings,
  updateSettings
}

export default adminLeaderboardAPI
