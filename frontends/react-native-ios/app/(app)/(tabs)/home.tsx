/**
 * @AngelaMos | 2026
 * home.tsx
 */

import { useCurrentUser } from '@/api/hooks'
import { DottedBackground } from '@/shared/components'
import type React from 'react'
import { SafeAreaView } from 'react-native-safe-area-context'
import { Stack, Text, YStack } from 'tamagui'

export default function HomeScreen(): React.ReactElement {
  const { data: user } = useCurrentUser()

  return (
    <DottedBackground>
      <SafeAreaView style={{ flex: 1 }} edges={['top']}>
        <YStack flex={1} padding="$6">
        <Stack marginBottom="$6">
          <Text
            fontSize={26}
            fontWeight="600"
            color="$textDefault"
            marginBottom="$2"
          >
            Welcome{user?.full_name ? `, ${user.full_name}` : ''}
          </Text>
          <Text fontSize={16} color="$textLighter">
            Your dashboard
          </Text>
        </Stack>

        <Stack
          backgroundColor="$bgSurface100"
          borderWidth={1}
          borderColor="$borderDefault"
          borderRadius="$4"
          padding="$5"
        >
          <Text fontSize={14} color="$textLight">
            This is a template home screen. Customize it for your app.
          </Text>
        </Stack>
        </YStack>
      </SafeAreaView>
    </DottedBackground>
  )
}
