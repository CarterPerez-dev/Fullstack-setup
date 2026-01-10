// ===================
// © AngelaMos | 2026
// secure.storage.ts
// ===================

import * as SecureStore from 'expo-secure-store'
import { APP_CONFIG } from '@/core/config'

const KEYCHAIN_OPTIONS: SecureStore.SecureStoreOptions = {
  keychainService: APP_CONFIG.KEYCHAIN_SERVICE,
  keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
}

export const SECURE_KEYS = {
  REFRESH_TOKEN: 'refresh_token',
  BIOMETRIC_ENABLED: 'biometric_enabled',
} as const

type SecureKey = (typeof SECURE_KEYS)[keyof typeof SECURE_KEYS]

export const secureStorage = {
  async getItem(key: SecureKey): Promise<string | null> {
    try {
      return await SecureStore.getItemAsync(key, KEYCHAIN_OPTIONS)
    } catch {
      return null
    }
  },

  async setItem(key: SecureKey, value: string): Promise<boolean> {
    try {
      await SecureStore.setItemAsync(key, value, KEYCHAIN_OPTIONS)
      return true
    } catch {
      return false
    }
  },

  async deleteItem(key: SecureKey): Promise<boolean> {
    try {
      await SecureStore.deleteItemAsync(key, KEYCHAIN_OPTIONS)
      return true
    } catch {
      return false
    }
  },

  async clearAll(): Promise<void> {
    const keys = Object.values(SECURE_KEYS)
    await Promise.all(keys.map((key) => this.deleteItem(key)))
  },
}
