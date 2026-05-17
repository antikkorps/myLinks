export default defineNuxtRouteMiddleware(() => {
  const store = useAuthStore()
  if (store.isAuthenticated) return navigateTo("/")
})
