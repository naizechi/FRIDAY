<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template>
  <div
    class="flex min-h-dvh flex-col items-center justify-center bg-surface px-4 py-8 pt-[max(2rem,env(safe-area-inset-top))] pb-[max(2rem,env(safe-area-inset-bottom))]"
  >
    <div class="w-full max-w-sm">
      <p class="mb-8 text-center text-xl font-semibold tracking-tight text-ink-950">FRIDAY</p>
      <RouterView />
    </div>
  </div>
</template>
