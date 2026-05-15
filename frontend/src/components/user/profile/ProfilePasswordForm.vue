<template>
  <div :class="props.embedded ? 'space-y-4' : 'card'">
    <div
      v-if="!props.embedded"
      class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
    >
      <h2 class="text-lg font-medium text-gray-900 dark:text-white">
        {{ t('profile.changePassword') }}
      </h2>
    </div>
    <div :class="props.embedded ? '' : 'px-6 py-6'">
      <form @submit.prevent="handleChangePassword" class="space-y-4">
        <div v-if="props.embedded">
          <p class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('profile.changePassword') }}
          </p>
        </div>
        <div>
          <label for="old_password" class="input-label">
            {{ t('profile.currentPassword') }}
          </label>
          <input
            id="old_password"
            v-model="form.old_password"
            type="password"
            required
            autocomplete="current-password"
            class="input"
          />
        </div>

        <div>
          <label for="new_password" class="input-label">
            {{ t('profile.newPassword') }}
          </label>
          <input
            id="new_password"
            v-model="form.new_password"
            type="password"
            required
            autocomplete="new-password"
            class="input"
          />
          <p class="input-hint">
            {{ t('profile.passwordHint') }}
          </p>
        </div>

        <div>
          <label for="confirm_password" class="input-label">
            {{ t('profile.confirmNewPassword') }}
          </label>
          <input
            id="confirm_password"
            v-model="form.confirm_password"
            type="password"
            required
            autocomplete="new-password"
            class="input"
          />
        </div>

        <div v-if="phoneVerifyEnabled">
          <label for="phone_verify_code" class="input-label">
            {{ t('profile.phoneBinding.codeLabel') }}
          </label>
          <p class="mb-2 text-sm text-gray-500 dark:text-gray-400">
            {{ boundPhoneLabel }}
          </p>
          <div class="flex gap-3">
            <input
              id="phone_verify_code"
              v-model="form.phone_verify_code"
              type="text"
              autocomplete="one-time-code"
              class="input flex-1"
              :disabled="!hasBoundPhone"
            />
            <button
              type="button"
              class="btn btn-secondary whitespace-nowrap"
              :disabled="!hasBoundPhone || sendingPhoneCode || phoneCodeCountdown > 0"
              @click="sendPhoneCode"
            >
              {{ phoneCodeCountdown > 0 ? `${phoneCodeCountdown}s` : (sendingPhoneCode ? t('auth.sendingCode') : t('auth.sendCode')) }}
            </button>
          </div>
        </div>

        <div class="flex justify-end pt-4">
          <button type="submit" :disabled="loading" class="btn btn-primary">
            {{ loading ? t('profile.changingPassword') : t('profile.changePasswordButton') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores'
import { userAPI } from '@/api'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const props = withDefaults(defineProps<{
  embedded?: boolean
  phoneVerifyEnabled?: boolean
}>(), {
  embedded: false,
  phoneVerifyEnabled: false,
})

const loading = ref(false)
const sendingPhoneCode = ref(false)
const phoneCodeCountdown = ref(0)
let phoneCodeTimer: ReturnType<typeof setInterval> | null = null
const hasBoundPhone = computed(() => Boolean(authStore.user?.phone_number?.trim()))
const boundPhoneLabel = computed(() => {
  const phone = authStore.user?.phone_number?.trim()
  return phone
    ? t('profile.phoneBinding.currentPhoneValue', { phone })
    : t('profile.phoneBinding.unbound')
})
const form = ref({
  old_password: '',
  new_password: '',
  confirm_password: '',
  phone_verify_code: ''
})

onMounted(() => {
  if (typeof authStore.refreshUser === 'function') {
    authStore.refreshUser().catch(() => {
      // Best-effort refresh so phone_number is not stale from old localStorage data.
    })
  }
})

onUnmounted(() => {
  clearPhoneCodeTimer()
})

function clearPhoneCodeTimer(): void {
  if (phoneCodeTimer) {
    clearInterval(phoneCodeTimer)
    phoneCodeTimer = null
  }
}

function startPhoneCodeCountdown(seconds: number): void {
  clearPhoneCodeTimer()
  phoneCodeCountdown.value = Math.max(0, seconds)
  if (phoneCodeCountdown.value <= 0) {
    return
  }
  phoneCodeTimer = setInterval(() => {
    if (phoneCodeCountdown.value <= 1) {
      phoneCodeCountdown.value = 0
      clearPhoneCodeTimer()
      return
    }
    phoneCodeCountdown.value -= 1
  }, 1000)
}

function extractCooldownCountdown(error: unknown): number | null {
  const err = error as {
    response?: {
      data?: {
        reason?: string
        metadata?: Record<string, string>
      }
    }
  }
  const responseData = err.response?.data
  if (responseData?.reason !== 'VERIFY_CODE_TOO_FREQUENT') {
    return null
  }
  const countdownRaw = responseData.metadata?.countdown
  if (!countdownRaw) {
    return null
  }
  const countdown = Number.parseInt(countdownRaw, 10)
  return Number.isFinite(countdown) && countdown > 0 ? countdown : null
}

const sendPhoneCode = async () => {
  if (!props.phoneVerifyEnabled) {
    return
  }
  if (!hasBoundPhone.value) {
    appStore.showError(t('profile.phoneBinding.unbound'))
    return
  }
  if (sendingPhoneCode.value || phoneCodeCountdown.value > 0) {
    return
  }
  try {
    sendingPhoneCode.value = true
    const response = await userAPI.sendChangePasswordPhoneCode()
    startPhoneCodeCountdown(response.countdown)
    appStore.showSuccess(t('profile.phoneBinding.sendCodeSuccess'))
  } catch (error: any) {
    const cooldown = extractCooldownCountdown(error)
    if (cooldown) {
      startPhoneCodeCountdown(cooldown)
    }
    appStore.showError(error.response?.data?.detail || t('profile.phoneBinding.sendCodeFailed'))
  } finally {
    sendingPhoneCode.value = false
  }
}

const handleChangePassword = async () => {
  if (form.value.new_password !== form.value.confirm_password) {
    appStore.showError(t('profile.passwordsNotMatch'))
    return
  }

  if (form.value.new_password.length < 8) {
    appStore.showError(t('profile.passwordTooShort'))
    return
  }

  loading.value = true
  try {
    await userAPI.changePassword(form.value.old_password, form.value.new_password, form.value.phone_verify_code)
    form.value = { old_password: '', new_password: '', confirm_password: '', phone_verify_code: '' }
    appStore.showSuccess(t('profile.passwordChangeSuccess'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('profile.passwordChangeFailed'))
  } finally {
    loading.value = false
  }
}
</script>
