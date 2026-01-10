/**
 * @AngelaMos | 2026
 * _layout.tsx
 */

import { Redirect, Stack } from 'expo-router'
import { Lock } from 'lucide-react-native'
import type React from 'react'
import { ActivityIndicator, Pressable, View } from 'react-native'
import { Text, YStack } from 'tamagui'
import {
  getBiometricLabel,
  useBiometricAuth,
  useBiometrics,
  useSession,
} from '@/shared/hooks'
import { colors } from '@/theme/tokens'

function LockScreen({ onUnlock }: { onUnlock: () => void }): React.ReactElement {
  const { biometryType } = useBiometrics()

  return (
    <View
      style={{
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        backgroundColor: colors.bgDefault.val,
      }}
    >
      <YStack alignItems="center" gap="$6">
        <Lock size={48} color={colors.textMuted.val} />
        <Text fontSize={18} fontWeight="600" color="$textDefault">
          App Locked
        </Text>
        <Pressable onPress={onUnlock}>
          <Text fontSize={14} color="$accent">
            Tap to unlock with {getBiometricLabel(biometryType)}
          </Text>
        </Pressable>
      </YStack>
    </View>
  )
}

export default function AppLayout(): React.ReactElement {
  const { isAuthenticated, isReady } = useSession()
  const { isLocked, unlock } = useBiometricAuth()

  if (!isReady) {
    return (
      <View
        style={{
          flex: 1,
          justifyContent: 'center',
          alignItems: 'center',
          backgroundColor: colors.bgDefault.val,
        }}
      >
        <ActivityIndicator size="large" color={colors.accent.val} />
      </View>
    )
  }

  if (!isAuthenticated) {
    return <Redirect href="/(auth)/login" />
  }

  if (isLocked) {
    return <LockScreen onUnlock={unlock} />
  }

  return (
    <Stack
      screenOptions={{
        headerShown: false,
        contentStyle: { backgroundColor: '#121212' },
      }}
    />
  )
}
