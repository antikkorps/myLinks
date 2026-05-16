export function useRegisterForm() {
  const email = ref("")
  const password = ref("")
  const firstname = ref("")
  const lastname = ref("")
  const error = ref("")
  const isLoading = ref(false)
  const authStore = useAuthStore()

  async function submit() {
    error.value = ""
    isLoading.value = true
    try {
      await authStore.register(
        email.value,
        password.value,
        firstname.value,
        lastname.value,
      )
      await navigateTo("/login?registered=1")
    } catch (err) {
      const code = err instanceof Error ? err.message : "unknown"
      error.value = errorMessages[code] ?? "Une erreur est survenue"
    } finally {
      isLoading.value = false
    }
  }

  return { email, password, firstname, lastname, error, isLoading, submit }
}
