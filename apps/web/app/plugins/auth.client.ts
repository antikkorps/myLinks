// Hydrate the auth store on app boot by calling /auth/me. This runs before
// any page renders so route middlewares can rely on user.value being set
// (or null) instead of seeing the initial undefined state.
//
// .client suffix → never runs on the server (we are SPA-only anyway).
// Filename "auth" sorts after "api", so $api is always provided by then.
export default defineNuxtPlugin(async () => {
  const auth = useAuthStore()
  await auth.fetchMe()
})
