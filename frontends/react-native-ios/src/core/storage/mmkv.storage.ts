// ===================
// © AngelaMos | 2026
// mmkv.storage.ts
// ===================

import { MMKV } from 'react-native-mmkv'
import type { StateStorage } from 'zustand/middleware'
import { APP_CONFIG } from '@/core/config'

export const mmkv = new MMKV({
  id: APP_CONFIG.MMKV_STORAGE_ID,
})

export const zustandMMKVStorage: StateStorage = {
  getItem: (name: string): string | null => {
    const value = mmkv.getString(name)
    return value ?? null
  },

  setItem: (name: string, value: string): void => {
    mmkv.set(name, value)
  },

  removeItem: (name: string): void => {
    mmkv.delete(name)
  },
}

export const queryClientMMKVStorage = {
  getItem: (key: string): string | null => {
    const value = mmkv.getString(key)
    return value ?? null
  },

  setItem: (key: string, value: string): void => {
    mmkv.set(key, value)
  },

  removeItem: (key: string): void => {
    mmkv.delete(key)
  },
}
