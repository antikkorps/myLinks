import type { $Fetch, FetchOptions } from 'ofetch'

// Paths where a 401 means "your credentials are wrong" (legitimate
// auth response), not "your access token expired". We must not try to
// refresh on these — that would loop or hide real errors.
const NO_REFRESH_PATHS = ['/auth/login', '/auth/refresh', '/auth/logout', '/auth/register']

function shouldSkipRefresh(request: unknown): boolean {
  const url = typeof request === 'string' ? request : (request as Request).url
  return NO_REFRESH_PATHS.some(p => url.includes(p))
}

export default defineNuxtPlugin((nuxtApp) => {
  const config = useRuntimeConfig()

  const baseOptions: FetchOptions = {
    baseURL: config.public.apiBase,
    credentials: 'include',
  }

  // Single-flight refresh: many parallel 401s share one /auth/refresh call.
  let refreshPromise: Promise<void> | null = null

  function refresh(): Promise<void> {
    if (!refreshPromise) {
      refreshPromise = $fetch('/auth/refresh', {
        ...baseOptions,
        method: 'POST',
      })
        .then(() => undefined)
        .finally(() => {
          refreshPromise = null
        })
    }
    return refreshPromise
  }

  const api: $Fetch = (async (request: any, options: any = {}) => {
    try {
      return await $fetch(request, { ...baseOptions, ...options })
    } catch (err: any) {
      const status = err?.response?.status ?? err?.statusCode
      if (status !== 401 || shouldSkipRefresh(request)) {
        throw err
      }
      try {
        await refresh()
      } catch {
        // Refresh failed: user is truly logged out. Clear state and surface the original 401.
        useAuthStore().user = null
        throw err
      }
      // Refresh succeeded — retry the original request once.
      return await $fetch(request, { ...baseOptions, ...options })
    }
  }) as $Fetch

  return {
    provide: { api },
  }
})

declare module '#app' {
  interface NuxtApp {
    $api: $Fetch
  }
}
