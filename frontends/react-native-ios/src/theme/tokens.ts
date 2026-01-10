/**
 * @AngelaMos | 2026
 * tokens.ts
 */

import { createTokens } from 'tamagui'
import { APP_CONFIG } from '@/core/config'

export const tokens = createTokens({
  color: {
    bgDefault: '#121212',
    bgAlternative: '#0f0f0f',
    bgPage: '#1a1a1a',
    bgCard: '#101010',

    bgShellBase: '#151515',
    bgShellDot: '#1d1d1d',

    bgContentBase: '#0f0f0f',
    bgContentDot: '#171717',

    bgSurface75: '#171717',
    bgSurface100: '#181818',
    bgSurface200: '#242424',
    bgSurface300: '#292929',

    bgControl: '#1a1a1a',
    bgSelection: '#313131',
    bgOverlay: '#242424',
    bgOverlayHover: '#2e2e2e',

    borderMuted: '#242424',
    borderDefault: '#2e2e2e',
    borderStrong: '#363636',
    borderStronger: '#454545',
    borderControl: '#393939',

    textDefault: '#fafafa',
    textLight: '#b4b4b4',
    textLighter: '#898989',
    textMuted: '#4d4d4d',

    white: '#ffffff',
    black: '#000000',

    accent: APP_CONFIG.ACCENT_COLOR,

    errorDefault: '#dc2626',
    errorLight: '#ef4444',

    transparent: 'transparent',
  },

  space: {
    0: 0,
    0.5: 2,
    1: 4,
    1.5: 6,
    2: 8,
    2.5: 10,
    3: 12,
    3.5: 14,
    4: 16,
    5: 20,
    6: 24,
    7: 28,
    8: 32,
    9: 36,
    10: 40,
    12: 48,
    14: 56,
    16: 64,
    20: 80,
    24: 96,
    true: 16,
  },

  size: {
    0: 0,
    0.5: 2,
    1: 4,
    1.5: 6,
    2: 8,
    2.5: 10,
    3: 12,
    3.5: 14,
    4: 16,
    5: 20,
    6: 24,
    7: 28,
    8: 32,
    9: 36,
    10: 40,
    11: 44,
    12: 48,
    14: 56,
    16: 64,
    20: 80,
    true: 16,
  },

  radius: {
    0: 0,
    1: 2,
    2: 4,
    3: 6,
    4: 8,
    5: 10,
    6: 12,
    true: 6,
  },

  zIndex: {
    0: 0,
    1: 100,
    2: 200,
    3: 300,
    4: 400,
    5: 500,
  },
})

export const colors = tokens.color
