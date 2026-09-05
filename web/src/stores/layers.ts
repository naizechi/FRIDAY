import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

/** Placeholder until profile/layers API exists (E2–E4). */
export const useLayersStore = defineStore('layers', () => {
  const names = ref<string[]>(['radar'])

  const active = computed(() => names.value)

  return { active }
})
