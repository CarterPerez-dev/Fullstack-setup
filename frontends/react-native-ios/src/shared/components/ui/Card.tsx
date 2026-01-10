/**
 * @AngelaMos | 2026
 * Card.tsx
 */

import { type GetProps, Stack, styled, Text } from 'tamagui'

export const Card = styled(Stack, {
  name: 'Card',
  backgroundColor: '$black',
  borderWidth: 1,
  borderColor: '$borderDefault',
  borderRadius: '$4',
  padding: '$8',
  width: '100%',
  maxWidth: 400,

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
  marginBottom: '$6',
})

export const CardTitle = styled(Text, {
  name: 'CardTitle',
  fontSize: 24,
  fontWeight: '600',
  color: '$white',
  marginBottom: '$2',
})

export const CardSubtitle = styled(Text, {
  name: 'CardSubtitle',
  fontSize: 14,
  color: '$textLight',
})

export const CardContent = styled(Stack, {
  name: 'CardContent',
  flexDirection: 'column',
  gap: '$5',
})

export const CardFooter = styled(Stack, {
  name: 'CardFooter',
  flexDirection: 'row',
  alignItems: 'center',
  justifyContent: 'center',
  marginTop: '$6',
})

export const CardFooterText = styled(Text, {
  name: 'CardFooterText',
  fontSize: 14,
  color: '$textLight',
  textAlign: 'center',
})

export const CardFooterLink = styled(Text, {
  name: 'CardFooterLink',
  fontSize: 14,
  color: '$accent',
  textDecorationLine: 'underline',
})

export type CardProps = GetProps<typeof Card>
