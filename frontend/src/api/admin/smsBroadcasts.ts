import { apiClient } from '../client'
import type { BasePaginationResponse } from '@/types'

export type SMSBroadcastMode = 'freeform' | 'template'
export type SMSBroadcastStatus = 'draft' | 'queued' | 'running' | 'succeeded' | 'failed' | 'canceled'
export type SMSBroadcastTemplateVarSource = 'phone_number' | 'email' | 'username'

export interface SMSBroadcastTemplateVarRow {
  key: string
  value?: string
  source?: SMSBroadcastTemplateVarSource
}

export interface SMSBroadcastAudience {
  user_ids?: number[]
  status?: string
  role?: string
  search?: string
  group_name?: string
  attributes?: Record<number, string>
}

export interface SMSBroadcastCampaign {
  id: number
  title?: string
  mode?: SMSBroadcastMode
  template_id?: string
  body?: string
  template_vars?: Record<string, string>
  template_var_rows?: SMSBroadcastTemplateVarRow[]
  status?: SMSBroadcastStatus
  audience?: SMSBroadcastAudience
  total_recipients?: number
  sent_count?: number
  failed_count?: number
  skipped_count?: number
  error_message?: string | null
  created_at?: string
  started_at?: string | null
  finished_at?: string | null
}

export interface SMSBroadcastRecipientPreview {
  user_id: number
  phone_number: string
  raw_phone: string
  rendered_body?: string
  status?: string
  error_message?: string | null
  sent_at?: string | null
}

export interface SMSBroadcastPreviewResponse {
  total: number
  sample: SMSBroadcastRecipientPreview[]
}

export async function preview(audience: SMSBroadcastAudience): Promise<SMSBroadcastPreviewResponse> {
  const { data } = await apiClient.post<SMSBroadcastPreviewResponse>('/admin/sms-broadcasts/preview', { audience })
  return data
}

export async function create(request: {
  title: string
  template_id: string
  audience: SMSBroadcastAudience
  vars?: SMSBroadcastTemplateVarRow[]
}): Promise<SMSBroadcastCampaign> {
  const { data } = await apiClient.post<SMSBroadcastCampaign>('/admin/sms-broadcasts', request)
  return data
}

export async function list(page = 1, pageSize = 20): Promise<BasePaginationResponse<SMSBroadcastCampaign>> {
  const { data } = await apiClient.get<BasePaginationResponse<SMSBroadcastCampaign>>('/admin/sms-broadcasts', {
    params: { page, page_size: pageSize }
  })
  return data
}

export async function getById(id: number): Promise<SMSBroadcastCampaign> {
  const { data } = await apiClient.get<SMSBroadcastCampaign>(`/admin/sms-broadcasts/${id}`)
  return data
}

export async function cancel(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>(`/admin/sms-broadcasts/${id}/cancel`)
  return data
}

export interface SMSBroadcastRecipientsResponse {
  items: SMSBroadcastRecipientPreview[]
  total: number
  page: number
  page_size: number
  pages: number
}

export async function getRecipients(
  id: number,
  page = 1,
  pageSize = 20,
  status?: string
): Promise<SMSBroadcastRecipientsResponse> {
  const { data } = await apiClient.get<SMSBroadcastRecipientsResponse>(`/admin/sms-broadcasts/${id}/recipients`, {
    params: { page, page_size: pageSize, status }
  })
  return data
}

const smsBroadcastsAPI = {
  preview,
  create,
  list,
  getById,
  cancel,
  getRecipients
}

export default smsBroadcastsAPI
