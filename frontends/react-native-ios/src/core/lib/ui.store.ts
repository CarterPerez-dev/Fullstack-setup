/**
 * @AngelaMos | 2026
 * ui.store.ts
 */

import { create } from 'zustand'
import { createJSONStorage, persist } from 'zustand/middleware'
import { zustandMMKVStorage } from '@/core/storage'

interface UIState {
  biometricsEnabled: boolean
}

interface UIActions {
  setBiometricsEnabled: (enabled: boolean) => void
}

type UIStore = UIState & UIActions

export const useUIStore = create<UIStore>()(
  persist(
    (set) => ({
      biometricsEnabled: false,

      setBiometricsEnabled: (enabled) => set({ biometricsEnabled: enabled }),
    }),
    {
      name: 'ui-storage',
      storage: createJSONStorage(() => zustandMMKVStorage),
    }
  )
)

export const useBiometricsEnabled = (): boolean =>
  useUIStore((s) => s.biometricsEnabled)
