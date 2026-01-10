/**
 * @AngelaMos | 2026
 * _layout.tsx
 */

import { Tabs } from 'expo-router'
import { Home, Settings } from 'lucide-react-native'
import type React from 'react'
import { colors } from '@/theme/tokens'

export default function TabsLayout(): React.ReactElement {
  return (
    <Tabs
      screenOptions={{
        headerShown: false,
        tabBarStyle: {
          backgroundColor: colors.bgDefault.val,
          borderTopColor: colors.borderDefault.val,
          borderTopWidth: 1,
          height: 84,
          paddingBottom: 28,
          paddingTop: 12,
        },
        tabBarActiveTintColor: colors.accent.val,
        tabBarInactiveTintColor: colors.textMuted.val,
        tabBarLabelStyle: {
          fontSize: 12,
          fontWeight: '500',
        },
      }}
    >
      <Tabs.Screen
        name="home"
        options={{
          title: 'Home',
          tabBarIcon: ({ color, size }) => <Home size={size} color={color} />,
        }}
      />
      <Tabs.Screen
        name="settings"
        options={{
          title: 'Settings',
          tabBarIcon: ({ color, size }) => <Settings size={size} color={color} />,
        }}
      />
    </Tabs>
  )
}
