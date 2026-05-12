import { defineStore } from 'pinia'
import type { User } from '~/types/user'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const isAuthenticated = computed(() => user.value !== null)

  // Single-flight: concurrent callers share the same in-flight request.
  let fetchMePromise: Promise<void> | null = null

  async function login(email: string, password: string): Promise<void> {
    const { $api } = useNuxtApp()
    const u = await $api<User>('/auth/login', {
      method: 'POST',
      body: { email, password },
    })
    user.value = u
    fetchMePromise = null
  }

  async function logout(): Promise<void> {
    const { $api } = useNuxtApp()
    try {
      await $api('/auth/logout', { method: 'POST' })
    } finally {
      user.value = null
      fetchMePromise = null
    }
  }

  function fetchMe(): Promise<void> {
    let p = fetchMePromise
    if (!p) {
      const { $api } = useNuxtApp()
      p = $api<User>('/auth/me')
        .then((u) => {
          user.value = u
        })
        .catch(() => {
          user.value = null
        })
      fetchMePromise = p
    }
    return p
  }

  return { user, isAuthenticated, login, logout, fetchMe }
})
