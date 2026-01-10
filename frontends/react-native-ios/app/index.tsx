import { Redirect } from 'expo-router'
import type React from 'react'
import { ActivityIndicator, View } from 'react-native'
import { useSession } from '@/shared/hooks'
import { colors } from '@/theme/tokens'

export default function Index(): React.ReactElement {
  const { isAuthenticated, isReady } = useSession()

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

  if (isAuthenticated) {
    return <Redirect href="/(app)/(tabs)/home" />
  }

  return <Redirect href="/(auth)/login" />
}
