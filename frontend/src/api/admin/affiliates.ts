/**
 * Admin Affiliate API endpoints
 * Manage per-user affiliate (邀请返利) configurations:
 * exclusive invite codes (overrides aff_code) and exclusive rebate rates.
 */

import { apiClient } from '../client'
import type { AffiliateWithdrawal, AffiliateWithdrawSettings, PaginatedResponse } from '@/types'

export interface AffiliateAdminEntry {
  user_id: number
  email: string
  username: string
  aff_code: string
  aff_code_custom: boolean
  aff_rebate_rate_percent?: number | null
  aff_count: number
}

export interface AffiliateInviterEntry {
  user_id: number
  email: string
  username: string
  aff_code: string
  aff_count: number
  total_rebate: number
  identity_status?: string | null
  identity_expires_at?: string | null
}

export interface AffiliateInvitee {
  user_id: number
  email: string
  username: string
  created_at?: string
  total_rebate: number
  paid_amount?: number
  risk_flagged?: boolean
  risk_reason?: string
  identity_status?: string | null
  identity_expires_at?: string | null
}

export interface AffiliateIdentityConfig {
  inviter_rate_multiplier: number
  invitee_rate_multiplier: number
  duration_hours: number
  qualified_invitee_count: number
  qualified_pay_amount: number
  eligible_order_types: Array<'balance' | 'subscription' | string>
  fingerprint_enforcement_enabled: boolean
  max_accounts_per_fingerprint_hash: number
}

export interface AdminAffiliateIdentityConfigResponse {
  enabled: boolean
  config: AffiliateIdentityConfig
}

export interface ListAffiliateWithdrawalsParams {
  page?: number
  page_size?: number
  search?: string
  status?: string
}

export interface ListAffiliateUsersParams {
  page?: number
  page_size?: number
  search?: string
}

export interface ListAffiliateInvitersParams {
  page?: number
  page_size?: number
  search?: string
  start_at?: string
  end_at?: string
  timezone?: string
}

export interface ListAffiliateRecordsParams {
  page?: number
  page_size?: number
  search?: string
  start_at?: string
  end_at?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
  timezone?: string
}

export interface AffiliateInviteRecord {
  inviter_id: number
  inviter_email: string
  inviter_username: string
  invitee_id: number
  invitee_email: string
  invitee_username: string
  aff_code: string
  total_rebate: number
  created_at: string
}

export interface AffiliateRebateRecord {
  order_id: number
  out_trade_no: string
  inviter_id: number
  inviter_email: string
  inviter_username: string
  invitee_id: number
  invitee_email: string
  invitee_username: string
  order_amount: number
  pay_amount: number
  rebate_base_amount?: number | null
  rebate_amount: number
  payment_type: string
  order_status: string
  created_at: string
}

export interface AffiliateTransferRecord {
  ledger_id: number
  user_id: number
  user_email: string
  username: string
  amount: number
  balance_after?: number | null
  available_quota_after?: number | null
  frozen_quota_after?: number | null
  history_quota_after?: number | null
  snapshot_available: boolean
  created_at: string
}

export interface AffiliateUserOverview {
  user_id: number
  email: string
  username: string
  aff_code: string
  rebate_rate_percent: number
  invited_count: number
  rebated_invitee_count: number
  available_quota: number
  history_quota: number
}

export interface UpdateAffiliateUserRequest {
  aff_code?: string
  aff_rebate_rate_percent?: number | null
  /** Set true to explicitly clear the per-user rate (sets it to NULL). */
  clear_rebate_rate?: boolean
}

export interface BatchSetRateRequest {
  user_ids: number[]
  aff_rebate_rate_percent?: number | null
  /** Set true to clear rates instead of setting. */
  clear?: boolean
}

export interface SimpleUser {
  id: number
  email: string
  username: string
}

export interface CreateInviteRelationRequest {
  inviter_user_id: number
  invitee_user_id: number
  overwrite: boolean
}

export interface CreateInviteRelationResponse {
  inviter_user_id: number
  invitee_user_id: number
  overwritten: boolean
  previous_inviter_user_id?: number | null
}

export async function listUsers(
  params: ListAffiliateUsersParams = {},
): Promise<PaginatedResponse<AffiliateAdminEntry>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateAdminEntry>>(
    '/admin/affiliates/users',
    {
      params: {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
        search: params.search ?? '',
      },
    },
  )
  return data
}

export async function lookupUsers(q: string): Promise<SimpleUser[]> {
  const { data } = await apiClient.get<SimpleUser[]>(
    '/admin/affiliates/users/lookup',
    { params: { q } },
  )
  return data
}

export async function listInviters(
  params: ListAffiliateInvitersParams = {},
): Promise<PaginatedResponse<AffiliateInviterEntry>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateInviterEntry>>(
    '/admin/affiliates/inviters',
    {
      params: {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
        search: params.search ?? '',
        start_at: params.start_at || undefined,
        end_at: params.end_at || undefined,
        timezone: params.timezone || undefined,
      },
    },
  )
  return data
}

export async function listInviterInvitees(
  userId: number,
): Promise<AffiliateInvitee[]> {
  const { data } = await apiClient.get<AffiliateInvitee[]>(
    `/admin/affiliates/inviters/${userId}/invitees`,
  )
  return data
}

export async function createInviteRelation(
  payload: CreateInviteRelationRequest,
): Promise<CreateInviteRelationResponse> {
  const { data } = await apiClient.post<CreateInviteRelationResponse>(
    '/admin/affiliates/invites',
    payload,
  )
  return data
}

export async function updateUserSettings(
  userId: number,
  payload: UpdateAffiliateUserRequest,
): Promise<{ user_id: number }> {
  const { data } = await apiClient.put<{ user_id: number }>(
    `/admin/affiliates/users/${userId}`,
    payload,
  )
  return data
}

export async function clearUserSettings(
  userId: number,
): Promise<{ user_id: number }> {
  const { data } = await apiClient.delete<{ user_id: number }>(
    `/admin/affiliates/users/${userId}`,
  )
  return data
}

export async function batchSetRate(
  payload: BatchSetRateRequest,
): Promise<{ affected: number }> {
  const { data } = await apiClient.post<{ affected: number }>(
    '/admin/affiliates/users/batch-rate',
    payload,
  )
  return data
}

function recordParams(params: ListAffiliateRecordsParams = {}) {
  return {
    page: params.page ?? 1,
    page_size: params.page_size ?? 20,
    search: params.search ?? '',
    start_at: params.start_at || undefined,
    end_at: params.end_at || undefined,
    sort_by: params.sort_by || undefined,
    sort_order: params.sort_order || undefined,
    timezone: params.timezone || undefined,
  }
}

export async function listInviteRecords(
  params: ListAffiliateRecordsParams = {},
): Promise<PaginatedResponse<AffiliateInviteRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateInviteRecord>>(
    '/admin/affiliates/invites',
    { params: recordParams(params) },
  )
  return data
}

export async function listRebateRecords(
  params: ListAffiliateRecordsParams = {},
): Promise<PaginatedResponse<AffiliateRebateRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateRebateRecord>>(
    '/admin/affiliates/rebates',
    { params: recordParams(params) },
  )
  return data
}

export async function listTransferRecords(
  params: ListAffiliateRecordsParams = {},
): Promise<PaginatedResponse<AffiliateTransferRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateTransferRecord>>(
    '/admin/affiliates/transfers',
    { params: recordParams(params) },
  )
  return data
}

export async function getUserOverview(
  userId: number,
): Promise<AffiliateUserOverview> {
  const { data } = await apiClient.get<AffiliateUserOverview>(
    `/admin/affiliates/users/${userId}/overview`,
  )
  return data
}

export async function getIdentityConfig(): Promise<AdminAffiliateIdentityConfigResponse> {
  const { data } = await apiClient.get<AdminAffiliateIdentityConfigResponse>(
    '/admin/affiliates/identity-config',
  )
  return data
}

export async function updateIdentityConfig(
  payload: AdminAffiliateIdentityConfigResponse,
): Promise<AdminAffiliateIdentityConfigResponse> {
  const { data } = await apiClient.put<AdminAffiliateIdentityConfigResponse>(
    '/admin/affiliates/identity-config',
    payload,
  )
  return data
}

export async function getWithdrawSettings(): Promise<AffiliateWithdrawSettings> {
  const { data } = await apiClient.get<AffiliateWithdrawSettings>(
    '/admin/affiliates/withdraw-settings',
  )
  return data
}

export async function updateWithdrawSettings(
  payload: AffiliateWithdrawSettings,
): Promise<AffiliateWithdrawSettings> {
  const { data } = await apiClient.put<AffiliateWithdrawSettings>(
    '/admin/affiliates/withdraw-settings',
    payload,
  )
  return data
}

export async function listWithdrawals(
  params: ListAffiliateWithdrawalsParams = {},
): Promise<PaginatedResponse<AffiliateWithdrawal>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateWithdrawal>>(
    '/admin/affiliates/withdrawals',
    {
      params: {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
        search: params.search ?? '',
        status: params.status || undefined,
      },
    },
  )
  return data
}

export async function approveWithdrawal(id: number, note = ''): Promise<AffiliateWithdrawal> {
  const { data } = await apiClient.post<AffiliateWithdrawal>(
    `/admin/affiliates/withdrawals/${id}/approve`,
    { note },
  )
  return data
}

export async function rejectWithdrawal(id: number, reason: string): Promise<AffiliateWithdrawal> {
  const { data } = await apiClient.post<AffiliateWithdrawal>(
    `/admin/affiliates/withdrawals/${id}/reject`,
    { reason },
  )
  return data
}

export async function markWithdrawalPaid(
  id: number,
  payload: { payout_channel: string; payout_trade_no?: string; admin_note?: string },
): Promise<AffiliateWithdrawal> {
  const { data } = await apiClient.post<AffiliateWithdrawal>(
    `/admin/affiliates/withdrawals/${id}/paid`,
    payload,
  )
  return data
}

export async function markWithdrawalFailed(id: number, reason: string): Promise<AffiliateWithdrawal> {
  const { data } = await apiClient.post<AffiliateWithdrawal>(
    `/admin/affiliates/withdrawals/${id}/fail`,
    { reason },
  )
  return data
}

export const affiliatesAPI = {
  listUsers,
  createInviteRelation,
  listInviters,
  listInviterInvitees,
  lookupUsers,
  updateUserSettings,
  clearUserSettings,
  batchSetRate,
  listInviteRecords,
  listRebateRecords,
  listTransferRecords,
  getUserOverview,
  getIdentityConfig,
  updateIdentityConfig,
  getWithdrawSettings,
  updateWithdrawSettings,
  listWithdrawals,
  approveWithdrawal,
  rejectWithdrawal,
  markWithdrawalPaid,
  markWithdrawalFailed,
}

export default affiliatesAPI
