/**
 * @AngelaMos | 2026
 * profile.tsx
 */

import { useChangePassword, useCurrentUser } from '@/api/hooks'
import { useUpdateProfile } from '@/api/hooks/useUsers'
import { Button, DottedBackground, Input, PasswordInput } from '@/shared/components'
import { haptics } from '@/shared/utils'
import { router } from 'expo-router'
import { ArrowLeft } from 'lucide-react-native'
import type React from 'react'
import { useEffect, useState } from 'react'
import { Alert, KeyboardAvoidingView, Platform, ScrollView } from 'react-native'
import { SafeAreaView } from 'react-native-safe-area-context'
import { Stack, Text, YStack } from 'tamagui'
import { colors } from '@/theme/tokens'

export default function ProfileScreen(): React.ReactElement {
  const { data: user } = useCurrentUser()
  const updateProfile = useUpdateProfile()
  const changePassword = useChangePassword()

  const [fullName, setFullName] = useState('')
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')

  useEffect(() => {
    if (user) {
      setFullName(user.full_name ?? '')
    }
  }, [user])

  const handleSaveProfile = (): void => {
    if (!fullName.trim()) {
      haptics.error()
      Alert.alert('Error', 'Please enter your name')
      return
    }

    updateProfile.mutate(
      { full_name: fullName.trim() },
      {
        onSuccess: () => {
          haptics.success()
          Alert.alert('Success', 'Profile updated')
        },
        onError: (err) => {
          haptics.error()
          Alert.alert('Error', err.message)
        },
      }
    )
  }

  const handleChangePassword = (): void => {
    if (!currentPassword || !newPassword || !confirmPassword) {
      haptics.error()
      Alert.alert('Error', 'Please fill in all password fields')
      return
    }

    if (newPassword !== confirmPassword) {
      haptics.error()
      Alert.alert('Error', 'New passwords do not match')
      return
    }

    if (newPassword.length < 8) {
      haptics.error()
      Alert.alert('Error', 'Password must be at least 8 characters')
      return
    }

    changePassword.mutate(
      { current_password: currentPassword, new_password: newPassword },
      {
        onSuccess: () => {
          haptics.success()
          Alert.alert('Success', 'Password changed')
          setCurrentPassword('')
          setNewPassword('')
          setConfirmPassword('')
        },
        onError: (err) => {
          haptics.error()
          Alert.alert('Error', err.message)
        },
      }
    )
  }

  return (
    <DottedBackground>
      <SafeAreaView style={{ flex: 1 }} edges={['top']}>
        <KeyboardAvoidingView
          behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
          style={{ flex: 1 }}
        >
          <ScrollView
            contentContainerStyle={{ padding: 24 }}
            keyboardShouldPersistTaps="handled"
          >
            <Stack
              flexDirection="row"
              alignItems="center"
              gap="$2"
              marginBottom="$6"
              pressStyle={{ opacity: 0.7 }}
              onPress={() => router.back()}
            >
              <ArrowLeft size={18} color={colors.textLight.val} />
              <Text fontSize={14} color="$textLight">
                Back
              </Text>
            </Stack>

            <Text
              fontSize={22}
              fontWeight="600"
              color="$textDefault"
              marginBottom="$6"
            >
              Profile
            </Text>

            <YStack gap="$6">
              <Stack
                backgroundColor="$bgSurface100"
                borderWidth={1}
                borderColor="$borderDefault"
                borderRadius="$3"
                padding="$4"
              >
                <Text
                  fontSize={14}
                  fontWeight="500"
                  color="$textDefault"
                  marginBottom="$4"
                >
                  Account Info
                </Text>

                <YStack gap="$3">
                  <Input
                    label="Email"
                    value={user?.email ?? ''}
                    editable={false}
                    autoCapitalize="none"
                    keyboardType="email-address"
                  />

                  <Input
                    label="Full Name"
                    value={fullName}
                    onChangeText={setFullName}
                    placeholder="Enter your name"
                  />

                  <Stack marginTop="$2">
                    <Button
                      onPress={handleSaveProfile}
                      loading={updateProfile.isPending}
                    >
                      {updateProfile.isPending ? 'Saving...' : 'Save Changes'}
                    </Button>
                  </Stack>
                </YStack>
              </Stack>

              <Stack
                backgroundColor="$bgSurface100"
                borderWidth={1}
                borderColor="$borderDefault"
                borderRadius="$3"
                padding="$4"
              >
                <Text
                  fontSize={14}
                  fontWeight="500"
                  color="$textDefault"
                  marginBottom="$4"
                >
                  Change Password
                </Text>

                <YStack gap="$3">
                  <PasswordInput
                    label="Current Password"
                    value={currentPassword}
                    onChangeText={setCurrentPassword}
                  />

                  <PasswordInput
                    label="New Password"
                    value={newPassword}
                    onChangeText={setNewPassword}
                  />

                  <PasswordInput
                    label="Confirm New Password"
                    value={confirmPassword}
                    onChangeText={setConfirmPassword}
                  />

                  <Stack marginTop="$2">
                    <Button
                      variant="secondary"
                      onPress={handleChangePassword}
                      loading={changePassword.isPending}
                    >
                      {changePassword.isPending ? 'Updating...' : 'Update Password'}
                    </Button>
                  </Stack>
                </YStack>
              </Stack>
            </YStack>
          </ScrollView>
        </KeyboardAvoidingView>
      </SafeAreaView>
    </DottedBackground>
  )
}
