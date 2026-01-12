// ===================
// © AngelaMos | 2026
// app.storage.ts
// ===================

import * as SecureStore from 'expo-secure-store'
import type { StateStorage } from 'zustand/middleware'

export const zustandStorage: StateStorage = {
  getItem: (name: string): string | null => {
    return SecureStore.getItem(name)
  },

  setItem: (name: string, value: string): void => {
    SecureStore.setItem(name, value)
  },

  removeItem: (name: string): void => {
    SecureStore.deleteItemAsync(name)
  },
}

export const queryClientStorage = {
  getItem: (key: string): string | null => {
    return SecureStore.getItem(key)
  },

  setItem: (key: string, value: string): void => {
    SecureStore.setItem(key, value)
  },

  removeItem: (key: string): void => {
    SecureStore.deleteItemAsync(key)
  },
}
