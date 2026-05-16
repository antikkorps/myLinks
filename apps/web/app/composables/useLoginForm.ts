export function useLoginForm() {
  const email = ref("")
  const password = ref("")
  const error = ref("")
  const isLoading = ref(false)
  const authStore = useAuthStore()

  async function submit() {
    error.value = ""
    isLoading.value = true
    try {
      await authStore.login(email.value, password.value)
      await navigateTo("/")
    } catch (err) {
      const code = err instanceof Error ? err.message : "unknown"
      error.value = errorMessages[code] ?? "Une erreur est survenue"
    } finally {
      isLoading.value = false
    }
  }

  return { email, password, error, isLoading, submit }
}
