/**
 * @AngelaMos | 2026
 * useBiometricAuth.ts
 */

import { useCallback, useEffect, useRef, useState } from 'react'
import { AppState, type AppStateStatus } from 'react-native'
import { useBiometricsEnabled } from '@/core/lib'
import { useBiometrics } from './useBiometrics'

interface UseBiometricAuthReturn {
  isLocked: boolean
  unlock: () => Promise<void>
}

const LOCK_TIMEOUT_MS = 30000

export const useBiometricAuth = (): UseBiometricAuthReturn => {
  const biometricsEnabled = useBiometricsEnabled()
  const { authenticate, isAvailable, isEnrolled } = useBiometrics()

  const [isLocked, setIsLocked] = useState(false)
  const backgroundTime = useRef<number | null>(null)

  const canUseBiometrics = biometricsEnabled && isAvailable && isEnrolled

  const unlock = useCallback(async (): Promise<void> => {
    if (!canUseBiometrics) {
      setIsLocked(false)
      return
    }

    const result = await authenticate('Unlock app')
    if (result.success) {
      setIsLocked(false)
    }
  }, [canUseBiometrics, authenticate])

  useEffect(() => {
    if (!canUseBiometrics) {
      setIsLocked(false)
      return
    }

    const handleAppStateChange = (nextAppState: AppStateStatus): void => {
      if (nextAppState === 'background' || nextAppState === 'inactive') {
        backgroundTime.current = Date.now()
      } else if (nextAppState === 'active') {
        if (backgroundTime.current !== null) {
          const elapsed = Date.now() - backgroundTime.current
          if (elapsed > LOCK_TIMEOUT_MS) {
            setIsLocked(true)
          }
          backgroundTime.current = null
        }
      }
    }

    const subscription = AppState.addEventListener('change', handleAppStateChange)

    return (): void => {
      subscription.remove()
    }
  }, [canUseBiometrics])

  useEffect(() => {
    if (isLocked && canUseBiometrics) {
      unlock()
    }
  }, [isLocked, canUseBiometrics, unlock])

  return { isLocked, unlock }
}
