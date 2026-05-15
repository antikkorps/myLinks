export default defineNuxtRouteMiddleware((to) => {
  if (to.path === "/login") return
  if (to.meta.auth === false) return
  const store = useAuthStore()
  if (!store.isAuthenticated) return navigateTo("/login")
})
