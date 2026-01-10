/**
 * @AngelaMos | 2026
 * _layout.tsx
 */

import { PersistQueryClientProvider } from '@tanstack/react-query-persist-client'
import { Stack } from 'expo-router'
import { StatusBar } from 'expo-status-bar'
import type React from 'react'
import { useEffect } from 'react'
import { GestureHandlerRootView } from 'react-native-gesture-handler'
import { SafeAreaProvider } from 'react-native-safe-area-context'
import { TamaguiProvider } from 'tamagui'
import {
  queryClient,
  queryClientPersister,
  setAuthFailureHandler,
} from '@/core/api'
import { useAuthStore } from '@/core/lib'
import { useAppStateFocusManager, useOnlineManager } from '@/shared/hooks'
import config from '../tamagui.config'

function AppProviders({
  children,
}: {
  children: React.ReactNode
}): React.ReactElement {
  useAppStateFocusManager()
  useOnlineManager()

  const logout = useAuthStore((s) => s.logout)

  useEffect(() => {
    setAuthFailureHandler(() => {
      logout()
    })
  }, [logout])

  return (
    <TamaguiProvider config={config} defaultTheme="dark">
      <PersistQueryClientProvider
        client={queryClient}
        persistOptions={{ persister: queryClientPersister }}
      >
        {children}
      </PersistQueryClientProvider>
    </TamaguiProvider>
  )
}

export default function RootLayout(): React.ReactElement {
  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <SafeAreaProvider>
        <AppProviders>
          <StatusBar style="light" />
          <Stack
            screenOptions={{
              headerShown: false,
              contentStyle: { backgroundColor: '#121212' },
              animation: 'fade',
            }}
          />
        </AppProviders>
      </SafeAreaProvider>
    </GestureHandlerRootView>
  )
}
