/**
 * @AngelaMos | 2026
 * Button.tsx
 */

import { haptics } from '@/shared/utils'
import type React from 'react'
import { type GetProps, styled } from 'tamagui'
import { Stack, Text } from 'tamagui'

const ButtonFrame = styled(Stack, {
  name: 'Button',
  tag: 'button',
  role: 'button',
  flexDirection: 'row',
  alignItems: 'center',
  justifyContent: 'center',
  height: 44,
  paddingHorizontal: '$4',
  borderRadius: '$2',
  cursor: 'pointer',
  pressStyle: {
    opacity: 0.9,
  },
  disabledStyle: {
    opacity: 0.5,
    cursor: 'not-allowed',
  },

  variants: {
    variant: {
      primary: {
        backgroundColor: '$white',
      },
      secondary: {
        backgroundColor: 'transparent',
        borderWidth: 1,
        borderColor: '$borderDefault',
        pressStyle: {
          backgroundColor: '$bgSurface75',
        },
      },
      accent: {
        backgroundColor: '$accent',
      },
      ghost: {
        backgroundColor: 'transparent',
        pressStyle: {
          backgroundColor: '$bgSurface75',
        },
      },
    },

    fullWidth: {
      true: {
        width: '100%',
      },
    },
  } as const,

  defaultVariants: {
    variant: 'primary',
  },
})

const ButtonText = styled(Text, {
  name: 'ButtonText',
  fontSize: 14,
  fontWeight: '500',
  textAlign: 'center',

  variants: {
    variant: {
      primary: {
        color: '$black',
      },
      secondary: {
        color: '$textDefault',
      },
      accent: {
        color: '$white',
      },
      ghost: {
        color: '$textDefault',
      },
    },
  } as const,

  defaultVariants: {
    variant: 'primary',
  },
})

type ButtonFrameProps = GetProps<typeof ButtonFrame>

interface ButtonProps extends Omit<ButtonFrameProps, 'children'> {
  children: string
  variant?: 'primary' | 'secondary' | 'accent' | 'ghost'
  fullWidth?: boolean
  disabled?: boolean
  loading?: boolean
  onPress?: () => void
}

export function Button({
  children,
  variant = 'primary',
  fullWidth,
  disabled,
  loading,
  onPress,
  ...props
}: ButtonProps): React.ReactElement {
  const handlePress = (): void => {
    if (disabled || loading) return
    haptics.light()
    onPress?.()
  }

  return (
    <ButtonFrame
      variant={variant}
      fullWidth={fullWidth}
      disabled={disabled || loading}
      onPress={handlePress}
      {...props}
    >
      <ButtonText variant={variant}>
        {loading ? 'Loading...' : children}
      </ButtonText>
    </ButtonFrame>
  )
}
