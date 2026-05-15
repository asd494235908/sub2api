<template>
  <div :class="props.embedded ? 'space-y-4' : 'card'">
    <div
      v-if="!props.embedded"
      class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
    >
      <h2 class="text-lg font-medium text-gray-900 dark:text-white">
        {{ t('profile.editProfile') }}
      </h2>
    </div>
    <div :class="props.embedded ? '' : 'px-6 py-6'">
      <form @submit.prevent="handleUpdateProfile" class="space-y-4">
        <div v-if="props.embedded">
          <p class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('profile.editProfile') }}
          </p>
        </div>
        <div>
          <label for="username" class="input-label">
            {{ t('profile.username') }}
          </label>
          <input
            id="username"
            v-model="username"
            type="text"
            class="input"
            :placeholder="t('profile.enterUsername')"
          />
        </div>

        <div v-if="phoneVerifyEnabled" class="space-y-3">
          <div>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ phoneSectionTitle }}
            </p>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ currentPhoneLabel }}
            </p>
          </div>
          <div>
            <label for="phone_number" class="input-label">
              {{ t('profile.phoneBinding.phoneLabel') }}
            </label>
            <input
              id="phone_number"
              v-model="phoneNumber"
              type="tel"
              class="input"
              :placeholder="t('profile.phoneBinding.phonePlaceholder')"
            />
          </div>
          <div>
            <label for="phone_verify_code" class="input-label">
              {{ t('profile.phoneBinding.codeLabel') }}
            </label>
            <div class="flex gap-3">
              <input
                id="phone_verify_code"
                v-model="phoneVerifyCode"
                type="text"
                class="input flex-1"
                autocomplete="one-time-code"
                :placeholder="t('profile.phoneBinding.codePlaceholder')"
              />
              <button
                type="button"
                class="btn btn-secondary whitespace-nowrap"
                :disabled="sendingPhoneCode || phoneCodeCountdown > 0"
                data-testid="profile-phone-send-code"
                @click="sendPhoneCode"
              >
                {{ phoneCodeCountdown > 0 ? `${phoneCodeCountdown}s` : (sendingPhoneCode ? t('auth.sendingCode') : t('auth.sendCode')) }}
              </button>
            </div>
          </div>
        </div>

        <div class="flex justify-end pt-4">
          <button type="submit" :disabled="loading" class="btn btn-primary">
            {{ loading ? t('profile.updating') : t('profile.updateProfile') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { userAPI } from '@/api'

const props = withDefaults(defineProps<{
  initialUsername: string
  embedded?: boolean
  phoneVerifyEnabled?: boolean
}>(), {
  embedded: false,
  phoneVerifyEnabled: false,
})

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const username = ref(props.initialUsername)
const loading = ref(false)
const phoneNumber = ref('')
const phoneVerifyCode = ref('')
const sendingPhoneCode = ref(false)
const phoneCodeCountdown = ref(0)
let phoneCodeTimer: ReturnType<typeof setInterval> | null = null

const hasBoundPhone = computed(() => Boolean(authStore.user?.phone_number?.trim()))
const phoneSectionTitle = computed(() =>
  hasBoundPhone.value ? t('profile.phoneBinding.rebindTitle') : t('profile.phoneBinding.bindTitle')
)
const currentPhoneLabel = computed(() => {
  const phone = authStore.user?.phone_number?.trim()
  return phone
    ? t('profile.phoneBinding.currentPhoneValue', { phone })
    : t('profile.phoneBinding.unbound')
})

watch(() => props.initialUsername, (val) => {
  username.value = val
})

watch(
  () => authStore.user?.phone_number,
  (val) => {
    phoneNumber.value = val?.trim() || ''
  },
  { immediate: true }
)

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
  if (!phoneNumber.value.trim()) {
    appStore.showError(t('profile.phoneBinding.phoneRequired'))
    return
  }
  if (sendingPhoneCode.value || phoneCodeCountdown.value > 0) {
    return
  }

  try {
    sendingPhoneCode.value = true
    const response = await userAPI.sendPhoneBindingCode(phoneNumber.value.trim())
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

const handleUpdateProfile = async () => {
  if (!username.value.trim()) {
    appStore.showError(t('profile.usernameRequired'))
    return
  }
  if (!props.phoneVerifyEnabled) {
    loading.value = true
    try {
      const updatedUser = await userAPI.updateProfile({ username: username.value })
      authStore.user = updatedUser
      appStore.showSuccess(t('profile.updateSuccess'))
    } catch (error: any) {
      appStore.showError(error.response?.data?.detail || t('profile.updateFailed'))
    } finally {
      loading.value = false
    }
    return
  }
  if (!phoneNumber.value.trim()) {
    appStore.showError(t('profile.phoneBinding.phoneRequired'))
    return
  }
  if (!phoneVerifyCode.value.trim()) {
    appStore.showError(t('profile.phoneBinding.codeRequired'))
    return
  }

  loading.value = true
  try {
    await userAPI.updateProfile({ username: username.value })
    const updatedUser = await userAPI.bindPhoneNumber({
      phone_number: phoneNumber.value.trim(),
      phone_verify_code: phoneVerifyCode.value.trim(),
    })
    authStore.user = updatedUser
    phoneVerifyCode.value = ''
    appStore.showSuccess(
      hasBoundPhone.value ? t('profile.phoneBinding.rebindSuccess') : t('profile.phoneBinding.bindSuccess')
    )
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('profile.phoneBinding.bindFailed'))
  } finally {
    loading.value = false
  }
}
</script>
