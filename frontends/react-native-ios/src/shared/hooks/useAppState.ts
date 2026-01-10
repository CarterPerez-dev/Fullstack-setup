/**
 * @AngelaMos | 2026
 * useAppState.ts
 */

import { focusManager } from '@tanstack/react-query'
import { useEffect } from 'react'
import { AppState, type AppStateStatus } from 'react-native'

interface UseAppStateOptions {
  onForeground?: () => void | Promise<void>
  onBackground?: () => void | Promise<void>
}

export const useAppState = (options?: UseAppStateOptions): void => {
  useEffect(() => {
    const handleAppStateChange = (nextAppState: AppStateStatus): void => {
      if (nextAppState === 'active') {
        focusManager.setFocused(true)
        options?.onForeground?.()
      } else {
        focusManager.setFocused(false)
        options?.onBackground?.()
      }
    }

    const subscription = AppState.addEventListener('change', handleAppStateChange)

    return (): void => {
      subscription.remove()
    }
  }, [options])
}

export const useAppStateFocusManager = (): void => {
  useEffect(() => {
    const handleAppStateChange = (nextAppState: AppStateStatus): void => {
      focusManager.setFocused(nextAppState === 'active')
    }

    const subscription = AppState.addEventListener('change', handleAppStateChange)

    return (): void => {
      subscription.remove()
    }
  }, [])
}
