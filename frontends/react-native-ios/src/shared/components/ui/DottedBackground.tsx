// ===================
// © AngelaMos | 2026
// DottedBackground.tsx
// ===================

import { colors } from '@/theme/tokens'
import type React from 'react'
import { Dimensions, StyleSheet, View } from 'react-native'
import Svg, { Circle, Defs, Pattern, Rect } from 'react-native-svg'

interface DottedBackgroundProps {
  children: React.ReactNode
  dotColor?: string
  dotSize?: number
  dotSpacing?: number
}

export function DottedBackground({
  children,
  dotColor = colors.bgShellDot.val,
  dotSize = 1,
  dotSpacing = 16,
}: DottedBackgroundProps): React.ReactElement {
  const { width, height } = Dimensions.get('window')

  return (
    <View style={styles.container}>
      <Svg
        style={StyleSheet.absoluteFill}
        width={width}
        height={height}
        preserveAspectRatio="none"
      >
        <Defs>
          <Pattern
            id="dotPattern"
            patternUnits="userSpaceOnUse"
            width={dotSpacing}
            height={dotSpacing}
          >
            <Circle cx={dotSpacing / 2} cy={dotSpacing / 2} r={dotSize} fill={dotColor} />
          </Pattern>
        </Defs>
        <Rect
          x="0"
          y="0"
          width="100%"
          height="100%"
          fill={colors.bgDefault.val}
        />
        <Rect x="0" y="0" width="100%" height="100%" fill="url(#dotPattern)" />
      </Svg>
      <View style={styles.content}>{children}</View>
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  content: {
    flex: 1,
  },
})
