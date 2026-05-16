import { FetchError } from "ofetch"
import { defineStore } from "pinia"
import type { User } from "~/types/user"

export const useAuthStore = defineStore("auth", () => {
  const user = ref<User | null>(null)
  const isAuthenticated = computed(() => user.value !== null)

  // Single-flight: concurrent callers share the same in-flight request.
  let fetchMePromise: Promise<void> | null = null

  async function register(
    email: string,
    password: string,
    firstname: string,
    lastname: string,
  ): Promise<void> {
    const { $api } = useNuxtApp()
    try {
      await $api<User>("/auth/register", {
        method: "POST",
        body: { email, password, first_name: firstname, last_name: lastname },
      })
    } catch (error) {
      if (error instanceof FetchError && error.statusCode === 409) {
        throw new Error("email_already_exists")
      }
      throw new Error("network_error")
    }
  }

  async function login(email: string, password: string): Promise<void> {
    const { $api } = useNuxtApp()
    try {
      const u = await $api<User>("/auth/login", {
        method: "POST",
        body: { email, password },
      })
      user.value = u
      fetchMePromise = null
    } catch (error) {
      if (error instanceof FetchError && error.statusCode === 401) {
        throw new Error("invalid_credentials")
      }
      throw new Error("network_error")
    }
  }

  async function logout(): Promise<void> {
    const { $api } = useNuxtApp()
    try {
      await $api("/auth/logout", { method: "POST" })
    } finally {
      user.value = null
      fetchMePromise = null
    }
  }

  function fetchMe(): Promise<void> {
    let p = fetchMePromise
    if (!p) {
      const { $api } = useNuxtApp()
      p = $api<User>("/auth/me")
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

  return { user, isAuthenticated, register, login, logout, fetchMe }
})
