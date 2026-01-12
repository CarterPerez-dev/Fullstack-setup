/**
 * @AngelaMos | 2026
 * register.tsx
 */

import { useRegister } from '@/api/hooks'
import { userCreateRequestSchema } from '@/api/types'
import { PASSWORD_CONSTRAINTS } from '@/core/config'
import {
  Button,
  Card,
  CardContent,
  CardFooter,
  CardFooterLink,
  CardFooterText,
  CardHeader,
  CardSubtitle,
  CardTitle,
  DottedBackground,
  Input,
  PasswordInput,
} from '@/shared/components'
import { haptics } from '@/shared/utils'
import { Link, router } from 'expo-router'
import type React from 'react'
import { useState } from 'react'
import { KeyboardAvoidingView, Platform, ScrollView } from 'react-native'
import { Stack, Text, YStack } from 'tamagui'
import { z } from 'zod'

const registerFormSchema = userCreateRequestSchema
  .extend({
    confirmPassword: z
      .string()
      .min(
        PASSWORD_CONSTRAINTS.MIN_LENGTH,
        `Password must be at least ${PASSWORD_CONSTRAINTS.MIN_LENGTH} characters`
      )
      .max(PASSWORD_CONSTRAINTS.MAX_LENGTH),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  })

export default function RegisterScreen(): React.ReactElement {
  const register = useRegister()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = (): void => {
    setError(null)

    const result = registerFormSchema.safeParse({
      email,
      password,
      confirmPassword,
    })

    if (!result.success) {
      const firstError = result.error.issues[0]
      setError(firstError.message)
      haptics.error()
      return
    }

    register.mutate(
      { email: result.data.email, password: result.data.password },
      {
        onSuccess: () => {
          haptics.success()
          router.replace('/(auth)/login')
        },
        onError: (err) => {
          setError(err.message)
          haptics.error()
        },
      }
    )
  }

  return (
    <DottedBackground>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        style={{ flex: 1 }}
      >
        <ScrollView
          contentContainerStyle={{
            flexGrow: 1,
            justifyContent: 'center',
            padding: 16,
            paddingTop: 60,
          }}
          keyboardShouldPersistTaps="handled"
        >
          <YStack alignItems="center">
            <Card>
            <CardHeader>
              <CardTitle>Sign up</CardTitle>
              <CardSubtitle>Create a new account</CardSubtitle>
            </CardHeader>

            <CardContent>
              <Input
                label="Email"
                placeholder="xxx@example.com"
                value={email}
                onChangeText={setEmail}
                autoCapitalize="none"
                autoCorrect={false}
                keyboardType="email-address"
                textContentType="emailAddress"
              />

              <PasswordInput
                label="Password"
                value={password}
                onChangeText={setPassword}
                textContentType="newPassword"
              />

              <PasswordInput
                label="Repeat Password"
                value={confirmPassword}
                onChangeText={setConfirmPassword}
                textContentType="newPassword"
              />

              {error && (
                <Text fontSize={12} color="$errorDefault">
                  {error}
                </Text>
              )}

              <Stack marginTop="$2">
                <Button
                  fullWidth
                  loading={register.isPending}
                  onPress={handleSubmit}
                >
                  {register.isPending ? 'Creating account...' : 'Sign up'}
                </Button>
              </Stack>
            </CardContent>

            <CardFooter>
              <CardFooterText>
                Already have an account?{' '}
                <Link href="/(auth)/login" asChild>
                  <CardFooterLink>Login</CardFooterLink>
                </Link>
              </CardFooterText>
            </CardFooter>
            </Card>
          </YStack>
        </ScrollView>
      </KeyboardAvoidingView>
    </DottedBackground>
  )
}
