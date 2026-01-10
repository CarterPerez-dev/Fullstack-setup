/**
 * @AngelaMos | 2026
 * _layout.tsx
 */

import { Stack } from 'expo-router'
import type React from 'react'

export default function AuthLayout(): React.ReactElement {
  return (
    <Stack
      screenOptions={{
        headerShown: false,
        contentStyle: { backgroundColor: '#121212' },
        animation: 'slide_from_right',
      }}
    />
  )
}
