<template>
  <component :is="ActiveHomeView" />
</template>

<script setup lang="ts">
import HomeViewLegacy from './HomeView.legacy.vue'
import HomeViewRedesign from './HomeView.redesign.vue'

// Flip to false to restore the legacy /home page immediately.
const USE_HOME_REDESIGN = true

const ActiveHomeView = USE_HOME_REDESIGN ? HomeViewRedesign : HomeViewLegacy
</script>
