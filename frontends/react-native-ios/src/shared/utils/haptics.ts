/**
 * @AngelaMos | 2026
 * haptics.ts
 */

import * as Haptics from 'expo-haptics'

export const haptics = {
  selection: (): void => {
    Haptics.selectionAsync()
  },

  light: (): void => {
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light)
  },

  medium: (): void => {
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium)
  },

  heavy: (): void => {
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Heavy)
  },

  success: (): void => {
    Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success)
  },

  warning: (): void => {
    Haptics.notificationAsync(Haptics.NotificationFeedbackType.Warning)
  },

  error: (): void => {
    Haptics.notificationAsync(Haptics.NotificationFeedbackType.Error)
  },
} as const
