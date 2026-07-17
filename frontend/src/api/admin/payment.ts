/**
 * Admin Payment API endpoints
 * Handles payment management operations for administrators
 */

import { apiClient } from '../client'
import type {
  DashboardStats,
  PaymentOrder,
  PaymentChannel,
  SubscriptionPlan,
  ProviderInstance,
  AdminLuckyWheelConfigResponse,
  LuckyWheelStats,
  LuckyWheelConfig,
  AdminRechargeActivityConfigResponse,
  RechargeActivityStats,
  RechargeActivityConfig,
  RechargeActivityDrawRecord,
  AdminFirstRechargeConfigResponse,
  AdminMemberLevelConfigResponse,
  FirstRechargeBulkChanceMode,
  FirstRechargeBulkChanceResult,
  FirstRechargeConfig,
  MemberLevelConfig,
} from '@/types/payment'
import type { BasePaginationResponse } from '@/types'

/** Admin-facing payment config returned by GET /admin/payment/config */
export interface AdminPaymentConfig {
  enabled: boolean
  min_amount: number
  max_amount: number
  daily_limit: number
  order_timeout_minutes: number
  max_pending_orders: number
  enabled_payment_types: string[]
  balance_disabled: boolean
  balance_recharge_multiplier: number
  subscription_usd_to_cny_rate: number
  recharge_fee_rate: number
  load_balance_strategy: string
  product_name_prefix: string
  product_name_suffix: string
  help_image_url: string
  help_text: string
  test_recharge_enabled: boolean
}

/** Fields accepted by PUT /admin/payment/config (all optional via pointer semantics) */
export interface UpdatePaymentConfigRequest {
  enabled?: boolean
  min_amount?: number
  max_amount?: number
  daily_limit?: number
  order_timeout_minutes?: number
  max_pending_orders?: number
  enabled_payment_types?: string[]
  balance_disabled?: boolean
  balance_recharge_multiplier?: number
  subscription_usd_to_cny_rate?: number
  recharge_fee_rate?: number
  load_balance_strategy?: string
  product_name_prefix?: string
  product_name_suffix?: string
  help_image_url?: string
  help_text?: string
  test_recharge_enabled?: boolean
}

export interface RefundResult {
  success: boolean
  warning?: string
  require_force?: boolean
  balance_deducted?: number
  subscription_days_deducted?: number
}

export const adminPaymentAPI = {
  // ==================== Config ====================

  /** Get payment configuration (admin view) */
  getConfig() {
    return apiClient.get<AdminPaymentConfig>('/admin/payment/config')
  },

  /** Update payment configuration */
  updateConfig(data: UpdatePaymentConfigRequest) {
    return apiClient.put('/admin/payment/config', data)
  },

  // ==================== Dashboard ====================

  /** Get payment dashboard statistics */
  getDashboard(days?: number) {
    return apiClient.get<DashboardStats>('/admin/payment/dashboard', {
      params: days ? { days } : undefined
    })
  },

  /** Get lucky wheel admin configuration */
  getLuckyWheelConfig() {
    return apiClient.get<AdminLuckyWheelConfigResponse>('/admin/payment/lucky-wheel/config')
  },

  /** Update lucky wheel admin configuration */
  updateLuckyWheelConfig(data: { enabled: boolean; config: LuckyWheelConfig }) {
    return apiClient.put<AdminLuckyWheelConfigResponse>('/admin/payment/lucky-wheel/config', data)
  },

  /** Get lucky wheel admin stats */
  getLuckyWheelStats() {
    return apiClient.get<LuckyWheelStats>('/admin/payment/lucky-wheel/stats')
  },

  /** Get recharge activity admin configuration */
  getRechargeActivityConfig() {
    return apiClient.get<AdminRechargeActivityConfigResponse>('/admin/payment/recharge-activity/config')
  },

  /** Update recharge activity admin configuration */
  updateRechargeActivityConfig(data: { enabled: boolean; config: RechargeActivityConfig }) {
    return apiClient.put<AdminRechargeActivityConfigResponse>('/admin/payment/recharge-activity/config', data)
  },

  /** Get recharge activity admin stats */
  getRechargeActivityStats(params?: { page?: number; page_size?: number; user_keyword?: string }) {
    return apiClient.get<RechargeActivityStats>('/admin/payment/recharge-activity/stats', { params })
  },

  /** Update manual fulfillment status for a recharge activity draw record */
  updateRechargeActivityRecordFulfillment(id: number, data: { status: 'pending' | 'fulfilled'; note?: string }) {
    return apiClient.put<RechargeActivityDrawRecord>(`/admin/payment/recharge-activity/records/${id}/fulfillment`, data)
  },

  getFirstRechargeConfig() {
    return apiClient.get<AdminFirstRechargeConfigResponse>('/admin/payment/first-recharge/config')
  },

  updateFirstRechargeConfig(data: { enabled: boolean; config: FirstRechargeConfig }) {
    return apiClient.put<AdminFirstRechargeConfigResponse>('/admin/payment/first-recharge/config', data)
  },

  grantFirstRechargeChance(data: { user_id: number; tier_id: string; chances: number; note?: string }) {
    return apiClient.post('/admin/payment/first-recharge/chances', data)
  },

  bulkUpdateFirstRechargeChances(data: { tier_id: string; chances: number; mode: FirstRechargeBulkChanceMode; note?: string }) {
    return apiClient.post<FirstRechargeBulkChanceResult>('/admin/payment/first-recharge/chances/bulk', data)
  },

  getMemberLevelConfig() {
    return apiClient.get<AdminMemberLevelConfigResponse>('/admin/payment/member-level/config')
  },

  updateMemberLevelConfig(data: { enabled: boolean; config: MemberLevelConfig }) {
    return apiClient.put<AdminMemberLevelConfigResponse>('/admin/payment/member-level/config', data)
  },

  // ==================== Orders ====================

  /** Get all orders (paginated, with filters) */
  getOrders(params?: {
    page?: number
    page_size?: number
    status?: string
    payment_type?: string
    user_id?: number
    keyword?: string
    start_date?: string
    end_date?: string
    order_type?: string
  }) {
    return apiClient.get<BasePaginationResponse<PaymentOrder>>('/admin/payment/orders', { params })
  },

  /** Get a specific order by ID */
  getOrder(id: number) {
    return apiClient.get<PaymentOrder>(`/admin/payment/orders/${id}`)
  },

  /** Cancel an order (admin) */
  cancelOrder(id: number) {
    return apiClient.post(`/admin/payment/orders/${id}/cancel`)
  },

  /** Retry recharge for a failed order */
  retryRecharge(id: number) {
    return apiClient.post(`/admin/payment/orders/${id}/retry`)
  },

  /** Process a refund */
  refundOrder(id: number, data: { amount: number; reason: string; deduct_balance?: boolean; force?: boolean }) {
    return apiClient.post<RefundResult>(`/admin/payment/orders/${id}/refund`, data)
  },

  /** Query and finalize a pending refund */
  queryRefund(id: number) {
    return apiClient.post<RefundResult>(`/admin/payment/orders/${id}/refund/query`)
  },

  // ==================== Channels ====================

  /** Get all payment channels */
  getChannels() {
    return apiClient.get<PaymentChannel[]>('/admin/payment/channels')
  },

  /** Create a payment channel */
  createChannel(data: Partial<PaymentChannel>) {
    return apiClient.post<PaymentChannel>('/admin/payment/channels', data)
  },

  /** Update a payment channel */
  updateChannel(id: number, data: Partial<PaymentChannel>) {
    return apiClient.put<PaymentChannel>(`/admin/payment/channels/${id}`, data)
  },

  /** Delete a payment channel */
  deleteChannel(id: number) {
    return apiClient.delete(`/admin/payment/channels/${id}`)
  },

  // ==================== Subscription Plans ====================

  /** Get all subscription plans */
  getPlans() {
    return apiClient.get<SubscriptionPlan[]>('/admin/payment/plans')
  },

  /** Create a subscription plan */
  createPlan(data: Record<string, unknown>) {
    return apiClient.post<SubscriptionPlan>('/admin/payment/plans', data)
  },

  /** Update a subscription plan */
  updatePlan(id: number, data: Record<string, unknown>) {
    return apiClient.put<SubscriptionPlan>(`/admin/payment/plans/${id}`, data)
  },

  /** Delete a subscription plan */
  deletePlan(id: number) {
    return apiClient.delete(`/admin/payment/plans/${id}`)
  },

  // ==================== Provider Instances ====================

  /** Get all provider instances */
  getProviders() {
    return apiClient.get<ProviderInstance[]>('/admin/payment/providers')
  },

  /** Create a provider instance */
  createProvider(data: Partial<ProviderInstance>) {
    return apiClient.post<ProviderInstance>('/admin/payment/providers', data)
  },

  /** Update a provider instance */
  updateProvider(id: number, data: Partial<ProviderInstance>) {
    return apiClient.put<ProviderInstance>(`/admin/payment/providers/${id}`, data)
  },

  /** Delete a provider instance */
  deleteProvider(id: number) {
    return apiClient.delete(`/admin/payment/providers/${id}`)
  }
}

export default adminPaymentAPI
