<template>
  <div class="min-h-screen bg-gray-50 px-4 py-10 dark:bg-dark-900">
    <div class="mx-auto max-w-md">
      <div class="card p-6 text-center">
        <div
          class="mx-auto flex h-10 w-10 items-center justify-center rounded-full bg-red-100 text-red-600 dark:bg-red-500/10 dark:text-red-300"
        >
          !
        </div>
        <h1 class="mt-4 text-lg font-semibold text-gray-900 dark:text-white">
          Casdoor 登录失败
        </h1>
        <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
          {{ errorMessage }}
        </p>
        <RouterLink class="btn btn-primary mt-6" to="/login">
          返回登录页
        </RouterLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()

function firstQueryValue(value: unknown): string {
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

const errorCode = computed(() => firstQueryValue(route.query.error).trim())
const errorMessage = computed(() => {
  switch (errorCode.value) {
    case 'account_conflict':
      return '邮箱和手机号命中了不同账号，请联系管理员处理。'
    case 'registration_disabled':
      return '当前未开放新账号注册，请使用已有账号登录。'
    case 'missing_ticket':
      return '缺少登录凭证，请重新发起 Casdoor 登录。'
    case 'exchange_failed':
      return '登录凭证已失效或兑换失败，请重新发起 Casdoor 登录。'
    default:
      return '请重新发起 Casdoor 登录。'
  }
})
</script>
