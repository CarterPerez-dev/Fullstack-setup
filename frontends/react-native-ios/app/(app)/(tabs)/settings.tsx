/**
 * @AngelaMos | 2026
 * settings.tsx
 */

import { ChevronRight, LogOut, Shield, User } from 'lucide-react-native'
import type React from 'react'
import { SafeAreaView } from 'react-native-safe-area-context'
import { Stack, Switch, Text, YStack } from 'tamagui'
import { useCurrentUser, useLogout } from '@/api/hooks'
import { useBiometricsEnabled, useUIStore } from '@/core/lib'
import { getBiometricLabel, useBiometrics } from '@/shared/hooks'
import { haptics } from '@/shared/utils'
import { colors } from '@/theme/tokens'

function SettingsRow({
  icon,
  label,
  onPress,
  rightElement,
}: {
  icon: React.ReactNode
  label: string
  onPress?: () => void
  rightElement?: React.ReactNode
}): React.ReactElement {
  return (
    <Stack
      flexDirection="row"
      alignItems="center"
      justifyContent="space-between"
      padding="$4"
      borderBottomWidth={1}
      borderBottomColor="$borderMuted"
      pressStyle={onPress ? { backgroundColor: '$bgSurface75' } : undefined}
      onPress={onPress}
    >
      <Stack flexDirection="row" alignItems="center" gap="$3">
        {icon}
        <Text fontSize={14} color="$textDefault">
          {label}
        </Text>
      </Stack>
      {rightElement ??
        (onPress && <ChevronRight size={18} color={colors.textMuted.val} />)}
    </Stack>
  )
}

export default function SettingsScreen(): React.ReactElement {
  const logout = useLogout()
  const { data: user } = useCurrentUser()
  const { biometryType, isAvailable, isEnrolled } = useBiometrics()
  const biometricsEnabled = useBiometricsEnabled()
  const setBiometricsEnabled = useUIStore((s) => s.setBiometricsEnabled)

  const handleLogout = (): void => {
    haptics.medium()
    logout.mutate()
  }

  const handleBiometricsToggle = (enabled: boolean): void => {
    haptics.selection()
    setBiometricsEnabled(enabled)
  }

  const canUseBiometrics = isAvailable && isEnrolled

  return (
    <SafeAreaView
      style={{ flex: 1, backgroundColor: colors.bgDefault.val }}
      edges={['top']}
    >
      <YStack flex={1} padding="$6">
        <Stack marginBottom="$6">
          <Text
            fontSize={26}
            fontWeight="600"
            color="$textDefault"
            marginBottom="$2"
          >
            Settings
          </Text>
          <Text fontSize={14} color="$textLighter">
            {user?.email}
          </Text>
        </Stack>

        <Stack
          backgroundColor="$bgSurface100"
          borderWidth={1}
          borderColor="$borderDefault"
          borderRadius="$4"
          overflow="hidden"
          marginBottom="$6"
        >
          <SettingsRow
            icon={<User size={18} color={colors.textLight.val} />}
            label="Profile"
            onPress={() => {}}
          />

          {canUseBiometrics && (
            <SettingsRow
              icon={<Shield size={18} color={colors.textLight.val} />}
              label={getBiometricLabel(biometryType)}
              rightElement={
                <Switch
                  size="$3"
                  checked={biometricsEnabled}
                  onCheckedChange={handleBiometricsToggle}
                  backgroundColor={
                    biometricsEnabled ? '$accent' : '$bgSurface200'
                  }
                >
                  <Switch.Thumb backgroundColor="$white" />
                </Switch>
              }
            />
          )}
        </Stack>

        <Stack
          backgroundColor="$bgSurface100"
          borderWidth={1}
          borderColor="$borderDefault"
          borderRadius="$4"
          overflow="hidden"
        >
          <SettingsRow
            icon={<LogOut size={18} color={colors.errorDefault.val} />}
            label="Logout"
            onPress={handleLogout}
          />
        </Stack>
      </YStack>
    </SafeAreaView>
  )
}
