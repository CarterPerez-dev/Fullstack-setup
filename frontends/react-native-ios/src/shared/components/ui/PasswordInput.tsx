/**
 * @AngelaMos | 2026
 * PasswordInput.tsx
 */

import { Eye, EyeOff } from 'lucide-react-native'
import { forwardRef, type Ref, useState } from 'react'
import type { TextInput as RNTextInput } from 'react-native'
import {
  type GetProps,
  Stack,
  styled,
  Input as TamaguiInput,
  Text,
} from 'tamagui'
import { colors } from '@/theme/tokens'

const InputWrapper = styled(Stack, {
  name: 'InputWrapper',
  position: 'relative',
  width: '100%',
})

const InputFrame = styled(TamaguiInput, {
  name: 'PasswordInputFrame',
  height: 48,
  paddingHorizontal: '$4',
  paddingRight: 48,
  backgroundColor: 'transparent',
  borderWidth: 1,
  borderColor: '$borderDefault',
  borderRadius: '$3',
  fontSize: 16,
  color: '$white',
  placeholderTextColor: '$textMuted',
  outlineWidth: 0,
  width: '100%',

  focusStyle: {
    borderColor: '$borderStrong',
  },

  variants: {
    error: {
      true: {
        borderColor: '$errorDefault',
        focusStyle: {
          borderColor: '$errorDefault',
        },
      },
    },
  } as const,
})

const EyeButton = styled(Stack, {
  name: 'EyeButton',
  position: 'absolute',
  right: '$3',
  top: 0,
  bottom: 0,
  justifyContent: 'center',
  alignItems: 'center',
  cursor: 'pointer',
  pressStyle: {
    opacity: 0.7,
  },
})

const Container = styled(Stack, {
  name: 'PasswordInputContainer',
  flexDirection: 'column',
  gap: '$2',
})

const Label = styled(Text, {
  name: 'Label',
  fontSize: 14,
  fontWeight: '500',
  color: '$white',
})

const ErrorText = styled(Text, {
  name: 'ErrorText',
  fontSize: 12,
  color: '$errorDefault',
})

type InputFrameProps = GetProps<typeof InputFrame>

interface PasswordInputProps
  extends Omit<InputFrameProps, 'ref' | 'secureTextEntry' | 'error'> {
  label?: string
  errorMessage?: string
}

export const PasswordInput = forwardRef<RNTextInput, PasswordInputProps>(
  function PasswordInput({ label, errorMessage, ...props }, ref) {
    const [showPassword, setShowPassword] = useState(false)

    const toggleVisibility = (): void => {
      setShowPassword((prev) => !prev)
    }

    const input = (
      <InputWrapper>
        <InputFrame
          ref={ref as Ref<RNTextInput>}
          secureTextEntry={!showPassword}
          error={!!errorMessage}
          autoCapitalize="none"
          autoCorrect={false}
          {...props}
        />
        <EyeButton onPress={toggleVisibility}>
          {showPassword ? (
            <EyeOff size={20} color={colors.textMuted.val} />
          ) : (
            <Eye size={20} color={colors.textMuted.val} />
          )}
        </EyeButton>
      </InputWrapper>
    )

    if (!label && !errorMessage) {
      return input
    }

    return (
      <Container>
        {label && <Label>{label}</Label>}
        {input}
        {errorMessage && <ErrorText>{errorMessage}</ErrorText>}
      </Container>
    )
  }
)
