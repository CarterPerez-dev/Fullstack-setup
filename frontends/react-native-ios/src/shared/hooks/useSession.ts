/**
 * @AngelaMos | 2026
 * useSession.ts
 */

import { useAuthStore } from '@/core/lib'

interface UseSessionReturn {
  isAuthenticated: boolean
  isLoading: boolean
  hasHydrated: boolean
  isReady: boolean
}

export const useSession = (): UseSessionReturn => {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const isLoading = useAuthStore((s) => s.isLoading)
  const hasHydrated = useAuthStore((s) => s.hasHydrated)

  const isReady = hasHydrated && !isLoading

  return {
    isAuthenticated,
    isLoading,
    hasHydrated,
    isReady,
  }
}

export const useRequireAuth = (): boolean => {
  const { isAuthenticated, isReady } = useSession()
  return isReady && isAuthenticated
}

export const useIsGuest = (): boolean => {
  const { isAuthenticated, isReady } = useSession()
  return isReady && !isAuthenticated
}
