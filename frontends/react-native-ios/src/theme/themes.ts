/**
 * @AngelaMos | 2026
 * themes.ts
 */

import { tokens } from './tokens'

export const darkTheme = {
  background: tokens.color.bgDefault,
  backgroundHover: tokens.color.bgSurface75,
  backgroundPress: tokens.color.bgSurface100,
  backgroundFocus: tokens.color.bgSurface100,

  color: tokens.color.textDefault,
  colorHover: tokens.color.textDefault,
  colorPress: tokens.color.textLight,
  colorFocus: tokens.color.textDefault,

  borderColor: tokens.color.borderDefault,
  borderColorHover: tokens.color.borderStrong,
  borderColorPress: tokens.color.borderStrong,
  borderColorFocus: tokens.color.accent,

  placeholderColor: tokens.color.textMuted,
}

export const themes = {
  dark: darkTheme,
}
