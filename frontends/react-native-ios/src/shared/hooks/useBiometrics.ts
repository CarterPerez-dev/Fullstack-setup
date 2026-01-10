/**
 * @AngelaMos | 2026
 * useBiometrics.ts
 */

import * as LocalAuthentication from 'expo-local-authentication'
import { useCallback, useEffect, useState } from 'react'

type BiometryType = 'facial' | 'fingerprint' | null

interface BiometricResult {
  success: boolean
  error?: string
}

interface UseBiometricsReturn {
  isAvailable: boolean
  isEnrolled: boolean
  biometryType: BiometryType
  authenticate: (promptMessage?: string) => Promise<BiometricResult>
  checkBiometricSupport: () => Promise<void>
}

export const useBiometrics = (): UseBiometricsReturn => {
  const [isAvailable, setIsAvailable] = useState(false)
  const [isEnrolled, setIsEnrolled] = useState(false)
  const [biometryType, setBiometryType] = useState<BiometryType>(null)

  const checkBiometricSupport = useCallback(async (): Promise<void> => {
    const [hasHardware, enrolled, supportedTypes] = await Promise.all([
      LocalAuthentication.hasHardwareAsync(),
      LocalAuthentication.isEnrolledAsync(),
      LocalAuthentication.supportedAuthenticationTypesAsync(),
    ])

    setIsAvailable(hasHardware)
    setIsEnrolled(enrolled)

    if (
      supportedTypes.includes(
        LocalAuthentication.AuthenticationType.FACIAL_RECOGNITION
      )
    ) {
      setBiometryType('facial')
    } else if (
      supportedTypes.includes(LocalAuthentication.AuthenticationType.FINGERPRINT)
    ) {
      setBiometryType('fingerprint')
    } else {
      setBiometryType(null)
    }
  }, [])

  useEffect(() => {
    checkBiometricSupport()
  }, [checkBiometricSupport])

  const authenticate = useCallback(
    async (promptMessage = 'Verify your identity'): Promise<BiometricResult> => {
      if (!isAvailable || !isEnrolled) {
        return { success: false, error: 'Biometrics unavailable' }
      }

      const result = await LocalAuthentication.authenticateAsync({
        promptMessage,
        cancelLabel: 'Cancel',
        fallbackLabel: 'Use Passcode',
        disableDeviceFallback: false,
      })

      if (result.success) {
        return { success: true }
      }

      return {
        success: false,
        error: result.error,
      }
    },
    [isAvailable, isEnrolled]
  )

  return {
    isAvailable,
    isEnrolled,
    biometryType,
    authenticate,
    checkBiometricSupport,
  }
}

export const getBiometricLabel = (biometryType: BiometryType): string => {
  switch (biometryType) {
    case 'facial':
      return 'Face ID'
    case 'fingerprint':
      return 'Touch ID'
    default:
      return 'Biometrics'
  }
}
