/**
 * Payment System Type Definitions
 */

// ==================== Enums / Union Types ====================

export type OrderStatus =
  | 'PENDING'
  | 'PAID'
  | 'RECHARGING'
  | 'COMPLETED'
  | 'EXPIRED'
  | 'CANCELLED'
  | 'FAILED'
  | 'REFUND_REQUESTED'
  | 'REFUNDING'
  | 'PARTIALLY_REFUNDED'
  | 'REFUNDED'
  | 'REFUND_FAILED'

export type PaymentType = 'alipay' | 'wxpay' | 'alipay_direct' | 'wxpay_direct' | 'stripe' | 'easypay' | 'airwallex'

export type OrderType = 'balance' | 'subscription'

// ==================== Configuration ====================

export interface PaymentConfig {
  payment_enabled: boolean
  min_amount: number
  max_amount: number
  daily_limit: number
  max_pending_orders: number
  order_timeout_minutes: number
  balance_disabled: boolean
  balance_recharge_multiplier: number
  enabled_payment_types: PaymentType[]
  help_image_url: string
  help_text: string
  stripe_publishable_key: string
}

export interface MethodLimit {
  currency?: string
  daily_limit: number
  daily_used: number
  daily_remaining: number
  single_min: number
  single_max: number
  fee_rate: number
  available: boolean
}

/** Response from /payment/limits API */
export interface MethodLimitsResponse {
  methods: Record<string, MethodLimit>
  global_min: number  // widest min across all methods; 0 = no minimum
  global_max: number  // widest max across all methods; 0 = no maximum
}

/** Response from /payment/checkout-info API — single call for the payment page */
export interface CheckoutInfoResponse {
  methods: Record<string, MethodLimit>
  global_min: number
  global_max: number
  plans: SubscriptionPlan[]
  balance_disabled: boolean
  balance_recharge_multiplier: number
  recharge_fee_rate: number
  help_text: string
  help_image_url: string
  stripe_publishable_key: string
  /** When true, Alipay payments on mobile always show the QR code instead of redirecting */
  alipay_force_qrcode?: boolean
}

// ==================== Orders ====================

export interface PaymentOrder {
  id: number
  user_id: number
  amount: number
  pay_amount: number
  currency?: string
  fee_rate: number
  payment_type: string
  out_trade_no: string
  status: OrderStatus
  order_type: OrderType
  created_at: string
  expires_at: string
  paid_at?: string
  completed_at?: string
  refund_amount: number
  refund_reason?: string
  refund_requested_at?: string
  refund_requested_by?: number
  refund_request_reason?: string
  plan_id?: number
  provider_instance_id?: string
}

// ==================== Plans & Channels ====================

export interface SubscriptionPlan {
  id: number
  group_id: number
  group_platform?: string
  group_name?: string
  rate_multiplier?: number
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
  supported_model_scopes?: string[]
  name: string
  description: string
  price: number
  original_price?: number
  validity_days: number
  validity_unit: string
  /** Stored as JSON string in backend; API layer should parse before use */
  features: string[]
  for_sale: boolean
  sort_order: number
}

export interface PaymentChannel {
  id: number
  group_id?: number
  name: string
  platform: string
  rate_multiplier: number
  description: string
  models: string[]
  features: string[]
  enabled: boolean
}

// ==================== Providers ====================

export interface ProviderInstance {
  id: number
  provider_key: string
  name: string
  config: Record<string, string>
  supported_types: string[]
  enabled: boolean
  payment_mode: string
  refund_enabled: boolean
  allow_user_refund: boolean
  limits: string
  sort_order: number
}

// ==================== Request / Response ====================

export interface CreateOrderRequest {
  amount: number
  payment_type: string
  order_type: string
  plan_id?: number
  return_url?: string
  payment_source?: string
  openid?: string
  wechat_resume_token?: string
  is_mobile?: boolean
}

export type CreateOrderResultType = 'order_created' | 'oauth_required' | 'jsapi_ready'

export interface WechatOAuthInfo {
  authorize_url?: string
  appid?: string
  openid?: string
  scope?: string
  state?: string
  redirect_url?: string
}

export interface WechatJSAPIPayload {
  appId?: string
  timeStamp?: string
  nonceStr?: string
  package?: string
  signType?: string
  paySign?: string
}

export interface CreateOrderResult {
  order_id: number
  amount: number
  pay_url?: string
  qr_code?: string
  client_secret?: string
  intent_id?: string
  currency?: string
  country_code?: string
  payment_env?: string
  pay_amount: number
  fee_rate: number
  expires_at: string
  result_type?: CreateOrderResultType
  payment_type?: string
  out_trade_no?: string
  payment_mode?: string
  resume_token?: string
  oauth?: WechatOAuthInfo
  jsapi?: WechatJSAPIPayload
  jsapi_payload?: WechatJSAPIPayload
}

export interface DashboardStats {
  today_amount: number
  total_amount: number
  today_count: number
  total_count: number
  avg_amount: number
  daily_series: { date: string; amount: number; count: number }[]
  payment_methods: { type: string; amount: number; count: number }[]
  top_users: { user_id: number; email: string; amount: number }[]
}

export interface LuckyWheelAmountTier {
  id: string
  name: string
  min_amount: number
  max_amount?: number | null
  min_multiplier: number
  max_multiplier: number
  draw_count: number
}

export interface LuckyWheelInviteBonusConfig {
  enabled: boolean
  qualifying_amount: number
  bonus_per_invitee: number
  max_bonus: number
  consume_policy: string
}

export interface LuckyWheelGoldenWindowConfig {
  enabled: boolean
  timezone: string
  start_time: string
  end_time: string
  min_amount: number
  extra_draws: number
  daily_quota: number
}

export interface LuckyWheelConfig {
  eligible_order_types: OrderType[]
  multiplier_step: number
  global_max_multiplier: number
  intro_text: string
  rules_title: string
  rules_items: string[]
  prizes: LuckyWheelPrize[]
  tiers: LuckyWheelTier[]
  amount_tiers: LuckyWheelAmountTier[]
  invite_bonus: LuckyWheelInviteBonusConfig
  golden_window: LuckyWheelGoldenWindowConfig
}

export interface LuckyWheelPrize {
  id: string
  name: string
  reward_amount: number
  enabled: boolean
}

export interface LuckyWheelTier {
  id: string
  name: string
  min_amount: number
  max_amount?: number | null
  prize_weights: Record<string, number>
}

export interface LuckyWheelDrawRecord {
  id: number
  session_id: number
  user_id: number
  draw_index: number
  base_multiplier: number
  invite_bonus_multiplier: number
  final_multiplier: number
  is_best: boolean
  created_at: string
}

export interface LuckyWheelSession {
  id: number
  user_id: number
  source_order_id: number
  source_order_type: OrderType
  source_pay_amount: number
  matched_tier_id: string
  matched_tier_name: string
  min_multiplier: number
  max_multiplier: number
  total_draws: number
  completed_draws: number
  remaining_draws: number
  best_multiplier: number
  invite_bonus_multiplier: number
  golden_window_extra_draws: number
  settled: boolean
  settled_bonus_amount?: number | null
  settled_at?: string | null
  draw_records?: LuckyWheelDrawRecord[]
  created_at: string
  updated_at: string
}

export interface LuckyWheelSummary {
  enabled: boolean
  config: LuckyWheelConfig
  active_session?: LuckyWheelSession | null
  pending_sessions: LuckyWheelSession[]
  history_sessions: LuckyWheelSession[]
}

export interface LuckyWheelDrawResult {
  session_id: number
  draw_record: LuckyWheelDrawRecord
  best_multiplier: number
  remaining_draws: number
  settled: boolean
  settled_bonus_amount?: number | null
  session?: LuckyWheelSession | null
}

export interface AdminLuckyWheelConfigResponse {
  enabled: boolean
  config: LuckyWheelConfig
}

export interface LuckyWheelMultiplierStat {
  multiplier: number
  draw_count: number
}

export interface LuckyWheelStats {
  enabled: boolean
  total_sessions: number
  pending_sessions: number
  settled_sessions: number
  total_bonus_amount: number
  recent_sessions: LuckyWheelSession[]
  multiplier_stats: LuckyWheelMultiplierStat[]
  golden_window_used_today: number
  golden_window_daily_quota: number
}

export interface RechargeActivityPrize {
  id: string
  name: string
  reward_amount: number
  reward_description: string
  probability: number
  min_pay_amount: number
  enabled: boolean
  sort_order: number
}

export interface RechargeActivityConfig {
  eligible_order_types: OrderType[]
  intro_text: string
  rules_title: string
  rules_items: string[]
  prizes: RechargeActivityPrize[]
}

export interface RechargeActivityChance {
  id: number
  user_id: number
  source_order_id: number
  source_order_type: OrderType
  source_pay_amount: number
  drawn: boolean
  drawn_at?: string | null
  created_at: string
  updated_at: string
}

export interface RechargeActivityDrawRecord {
  id: number
  chance_id: number
  user_id: number
  user_email?: string
  user_name?: string
  source_order_id: number
  prize_id: string
  prize_name: string
  reward_amount: number
  reward_description: string
  probability: number
  min_pay_amount: number
  prize_snapshot: string
  eligible_prize_ids: string[]
  fulfillment_status: 'pending' | 'fulfilled'
  fulfillment_note: string
  fulfilled_at?: string | null
  fulfilled_by?: number | null
  created_at: string
}

export interface RechargeActivitySummary {
  enabled: boolean
  config: RechargeActivityConfig
  pending_chances: RechargeActivityChance[]
  history_records: RechargeActivityDrawRecord[]
}

export interface RechargeActivityDrawResult {
  chance_id: number
  record: RechargeActivityDrawRecord
  chance?: RechargeActivityChance | null
}

export interface AdminRechargeActivityConfigResponse {
  enabled: boolean
  config: RechargeActivityConfig
}

export interface RechargeActivityStats {
  enabled: boolean
  total_chances: number
  pending_chances: number
  drawn_chances: number
  pending_fulfillments: number
  fulfilled_records: number
  total_reward_amount: number
  recent_records: RechargeActivityDrawRecord[]
  recent_records_total: number
  recent_records_page: number
  recent_records_page_size: number
  recent_records_keyword: string
}
