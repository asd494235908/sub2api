<template>
  <div class="min-h-screen bg-gray-50 px-4 py-10 dark:bg-dark-900">
    <div class="mx-auto max-w-md">
      <div class="card p-6 text-center">
        <div
          class="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
        <h1 class="mt-4 text-lg font-semibold text-gray-900 dark:text-white">
          Casdoor 登录处理中
        </h1>
        <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
          正在完成登录，请稍候。
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  exchangeCasdoorTicket,
  persistOAuthTokenContext,
  sanitizeSameSiteRedirectPath,
} from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

function firstQueryValue(value: unknown): string {
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

function redirectToError(error: string) {
  return router.replace({
    path: '/auth/casdoor/error',
    query: { error },
  })
}

onMounted(async () => {
  const ticket = firstQueryValue(route.query.ticket).trim()
  if (!ticket) {
    await redirectToError('missing_ticket')
    return
  }

  const redirect = sanitizeSameSiteRedirectPath(firstQueryValue(route.query.redirect)) || '/dashboard'

  try {
    const authResponse = await exchangeCasdoorTicket(ticket)
    persistOAuthTokenContext(authResponse)
    await authStore.setToken(authResponse.access_token)
    await router.replace(redirect)
  } catch {
    await redirectToError('exchange_failed')
  }
})
</script>
