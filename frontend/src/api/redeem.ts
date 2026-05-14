/**
 * Redeem code API endpoints
 * Handles redeem code redemption for users
 */

import { apiClient } from './client'
import type { RedeemCodeRequest } from '@/types'

export interface RedeemHistoryItem {
  id: number
  code: string
  type: string
  value: number
  status: string
  used_at: string
  created_at: string
  source?: string
  title?: string
  // Notes from admin for admin_balance/admin_concurrency types
  notes?: string
  // Subscription-specific fields
  group_id?: number
  validity_days?: number
  group?: {
    id: number
    name: string
  }
}

export interface WeeklyQuotaInfo {
  enabled: boolean
  amount: number
  status: 'claimable' | 'claimed' | 'disabled'
  window_started_at: string
  window_ends_at: string
  claimed_at?: string
  next_claim_at?: string
  total_claim_count: number
  total_claim_amount: number
}

export interface WeeklyQuotaClaimResult {
  message: string
  type: string
  value: number
  new_balance: number
  claimed_at: string
  window_started_at: string
  window_ends_at: string
  next_claim_at: string
}

/**
 * Redeem a code
 * @param code - Redeem code string
 * @returns Redemption result with updated balance or concurrency
 */
export async function redeem(code: string): Promise<{
  message: string
  type: string
  value: number
  new_balance?: number
  new_concurrency?: number
}> {
  const payload: RedeemCodeRequest = { code }

  const { data } = await apiClient.post<{
    message: string
    type: string
    value: number
    new_balance?: number
    new_concurrency?: number
  }>('/redeem', payload)

  return data
}

/**
 * Get user's redemption history
 * @returns List of redeemed codes
 */
export async function getHistory(): Promise<RedeemHistoryItem[]> {
  const { data } = await apiClient.get<RedeemHistoryItem[]>('/redeem/history')
  return data
}

export async function getWeeklyQuota(): Promise<WeeklyQuotaInfo> {
  const { data } = await apiClient.get<WeeklyQuotaInfo>('/redeem/weekly-quota')
  return data
}

export async function claimWeeklyQuota(): Promise<WeeklyQuotaClaimResult> {
  const { data } = await apiClient.post<WeeklyQuotaClaimResult>('/redeem/weekly-quota/claim')
  return data
}

export const redeemAPI = {
  redeem,
  getHistory,
  getWeeklyQuota,
  claimWeeklyQuota
}

export default redeemAPI
