/**
 * @AngelaMos | 2026
 * login.tsx
 */

import { useLogin } from '@/api/hooks'
import { loginRequestSchema } from '@/api/types'
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

export default function LoginScreen(): React.ReactElement {
  const login = useLogin()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = (): void => {
    setError(null)

    const result = loginRequestSchema.safeParse({
      username: email,
      password,
    })

    if (!result.success) {
      const firstError = result.error.issues[0]
      setError(firstError.message)
      haptics.error()
      return
    }

    login.mutate(result.data, {
      onSuccess: () => {
        haptics.success()
        router.replace('/(app)/(tabs)/home')
      },
      onError: (err) => {
        setError(err.message)
        haptics.error()
      },
    })
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
              <CardTitle>Login</CardTitle>
              <CardSubtitle>Welcome back</CardSubtitle>
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
                textContentType="password"
              />

              {error && (
                <Text fontSize={12} color="$errorDefault">
                  {error}
                </Text>
              )}

              <Stack marginTop="$2">
                <Button
                  fullWidth
                  loading={login.isPending}
                  onPress={handleSubmit}
                >
                  {login.isPending ? 'Logging in...' : 'Login'}
                </Button>
              </Stack>
            </CardContent>

            <CardFooter>
              <CardFooterText>
                Don't have an account?{' '}
                <Link href="/(auth)/register" asChild>
                  <CardFooterLink>Sign up</CardFooterLink>
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
