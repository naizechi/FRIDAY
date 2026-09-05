<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import { ApiError } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const email = ref('')
const password = ref('')
const error = ref<string | null>(null)
const submitting = ref(false)

async function onSubmit() {
  error.value = null
  submitting.value = true
  try {
    await auth.register(email.value.trim(), password.value)
    await router.replace({ name: 'radar' })
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Registration failed'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <form class="space-y-4" @submit.prevent="onSubmit">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight text-ink-950">Create account</h1>
      <p class="mt-1 text-sm text-ink-500">Email and password, at least 8 characters.</p>
    </div>

    <label class="block space-y-1">
      <span class="text-sm font-medium text-ink-700">Email</span>
      <input
        v-model="email"
        type="email"
        autocomplete="email"
        required
        class="w-full rounded-md border border-ink-100 bg-white px-3 py-2 text-sm text-ink-950 outline-none focus:border-accent-500"
      />
    </label>

    <label class="block space-y-1">
      <span class="text-sm font-medium text-ink-700">Password</span>
      <input
        v-model="password"
        type="password"
        autocomplete="new-password"
        required
        minlength="8"
        class="w-full rounded-md border border-ink-100 bg-white px-3 py-2 text-sm text-ink-950 outline-none focus:border-accent-500"
      />
    </label>

    <p v-if="error" class="text-sm text-red-600">{{ error }}</p>

    <button
      type="submit"
      class="w-full rounded-md bg-accent-500 px-3 py-2.5 text-sm font-medium text-white disabled:opacity-60"
      :disabled="submitting"
    >
      {{ submitting ? 'Creating…' : 'Create account' }}
    </button>

    <p class="text-center text-sm text-ink-500">
      Already have an account?
      <RouterLink :to="{ name: 'login' }" class="font-medium text-accent-500">
        Sign in
      </RouterLink>
    </p>
  </form>
</template>
