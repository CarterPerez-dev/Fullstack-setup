/**
 * @AngelaMos | 2026
 * useOnlineManager.ts
 */

import NetInfo from '@react-native-community/netinfo'
import { onlineManager } from '@tanstack/react-query'
import { useEffect } from 'react'

export const useOnlineManager = (): void => {
  useEffect(() => {
    const unsubscribe = NetInfo.addEventListener((state) => {
      const isOnline =
        state.isConnected !== null &&
        state.isConnected &&
        Boolean(state.isInternetReachable)
      onlineManager.setOnline(isOnline)
    })

    return (): void => {
      unsubscribe()
    }
  }, [])
}
