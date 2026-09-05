<script setup lang="ts">
import { RouterLink, RouterView, useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

async function onLogout() {
  await auth.logout()
  await router.replace({ name: 'login' })
}
</script>

<template>
  <div class="flex min-h-dvh flex-col md:flex-row">
    <aside
      class="hidden w-56 shrink-0 border-r border-ink-100 bg-surface p-4 pt-[max(1rem,env(safe-area-inset-top))] md:block"
    >
      <p class="mb-6 text-lg font-semibold tracking-tight text-ink-950">FRIDAY</p>
      <nav class="flex flex-col gap-1">
        <RouterLink
          :to="{ name: 'radar' }"
          class="rounded-md px-3 py-2 text-sm font-medium text-ink-700 transition hover:bg-ink-50 hover:text-ink-950"
          active-class="bg-ink-50 text-ink-950"
        >
          Radar
        </RouterLink>
      </nav>
      <button
        type="button"
        class="mt-6 w-full rounded-md px-3 py-2 text-left text-sm font-medium text-ink-500"
        @click="onLogout"
      >
        Sign out
      </button>
    </aside>

    <div class="flex min-w-0 flex-1 flex-col">
      <header
        class="flex items-center justify-between border-b border-ink-100 bg-surface px-4 py-3 pt-[max(0.75rem,env(safe-area-inset-top))] md:hidden"
      >
        <p class="text-base font-semibold tracking-tight text-ink-950">FRIDAY</p>
        <button type="button" class="text-sm font-medium text-ink-500" @click="onLogout">
          Sign out
        </button>
      </header>

      <main class="flex-1 overflow-y-auto px-4 py-4 pb-[max(5rem,calc(4rem+env(safe-area-inset-bottom)))] md:pb-6">
        <RouterView />
      </main>

      <nav
        class="fixed inset-x-0 bottom-0 z-10 flex border-t border-ink-100 bg-surface pb-[env(safe-area-inset-bottom)] md:hidden"
      >
        <RouterLink
          :to="{ name: 'radar' }"
          class="flex flex-1 flex-col items-center justify-center gap-0.5 py-2 text-xs font-medium text-ink-500"
          active-class="text-accent-500"
        >
          Radar
        </RouterLink>
      </nav>
    </div>
  </div>
</template>
