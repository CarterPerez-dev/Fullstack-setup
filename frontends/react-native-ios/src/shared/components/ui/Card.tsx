/**
 * @AngelaMos | 2026
 * Card.tsx
 */

import { type GetProps, styled } from 'tamagui'
import { Stack, Text } from 'tamagui'

export const Card = styled(Stack, {
  name: 'Card',
  backgroundColor: '$black',
  borderWidth: 1,
  borderColor: '$borderDefault',
  borderRadius: '$3',
  padding: '$6',
  width: '100%',
  maxWidth: 340,

  variants: {
    variant: {
      default: {},
      surface: {
        backgroundColor: '$bgSurface100',
      },
    },
  } as const,

  defaultVariants: {
    variant: 'default',
  },
})

export const CardHeader = styled(Stack, {
  name: 'CardHeader',
  flexDirection: 'column',
  marginBottom: '$4',
})

export const CardTitle = styled(Text, {
  name: 'CardTitle',
  fontSize: 22,
  fontWeight: '600',
  color: '$white',
  marginBottom: '$1',
})

export const CardSubtitle = styled(Text, {
  name: 'CardSubtitle',
  fontSize: 12,
  color: '$textLight',
})

export const CardContent = styled(Stack, {
  name: 'CardContent',
  flexDirection: 'column',
  gap: '$4',
})

export const CardFooter = styled(Stack, {
  name: 'CardFooter',
  flexDirection: 'row',
  alignItems: 'center',
  justifyContent: 'center',
  marginTop: '$4',
})

export const CardFooterText = styled(Text, {
  name: 'CardFooterText',
  fontSize: 12,
  color: '$textLight',
  textAlign: 'center',
})

export const CardFooterLink = styled(Text, {
  name: 'CardFooterLink',
  fontSize: 12,
  color: '$accent',
  textDecorationLine: 'underline',
})

export type CardProps = GetProps<typeof Card>
