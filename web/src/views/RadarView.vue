<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { fetchHealth } from '@/api/health'
import { useLayersStore } from '@/stores/layers'

const layers = useLayersStore()
const healthStatus = ref<string | null>(null)
const healthError = ref<string | null>(null)

onMounted(async () => {
  try {
    const health = await fetchHealth()
    healthStatus.value = health.status
  } catch {
    healthError.value = 'API is not reachable yet'
  }
})
</script>

<template>
  <section class="mx-auto w-full max-w-lg space-y-4">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight text-ink-950">Radar</h1>
      <p class="mt-1 text-sm text-ink-500">
        Your personal world will appear here. Layers activate from your profile.
      </p>
    </div>

    <div class="rounded-lg border border-ink-100 bg-surface p-4">
      <p class="text-sm font-medium text-ink-900">Active layers</p>
      <ul class="mt-2 space-y-1">
        <li v-for="layer in layers.active" :key="layer" class="text-sm text-ink-700">
          {{ layer }}
        </li>
      </ul>
      <p v-if="layers.active.length === 0" class="mt-2 text-sm text-ink-500">
        No layers yet — complete onboarding later.
      </p>
    </div>

    <div class="rounded-lg border border-dashed border-ink-100 bg-surface/60 p-4">
      <p class="text-sm font-medium text-ink-900">API health</p>
      <p v-if="healthStatus" class="mt-1 text-sm text-ink-700">status: {{ healthStatus }}</p>
      <p v-else-if="healthError" class="mt-1 text-sm text-ink-500">{{ healthError }}</p>
      <p v-else class="mt-1 text-sm text-ink-500">Checking…</p>
    </div>
  </section>
</template>
