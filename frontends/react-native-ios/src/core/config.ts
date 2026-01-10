// ===================
// © AngelaMos | 2026
// config.ts
// ===================

export const APP_CONFIG = {
  DEV_IP: '192.168.1.167',
  DEV_PORT: '5426',
  PROD_API_URL: 'https://api.example.com',

  KEYCHAIN_SERVICE: 'com.yourcompany.app.auth',
  AUTH_STORAGE_KEY: 'app-auth',
  MMKV_STORAGE_ID: 'app-storage',

  ACCENT_COLOR: '#C15F3C',
} as const

const API_VERSION = 'v1'

export const API_ENDPOINTS = {
  AUTH: {
    LOGIN: `/${API_VERSION}/auth/login`,
    REFRESH: `/${API_VERSION}/auth/refresh`,
    REFRESH_MOBILE: `/${API_VERSION}/auth/refresh-mobile`,
    LOGOUT: `/${API_VERSION}/auth/logout`,
    LOGOUT_ALL: `/${API_VERSION}/auth/logout-all`,
    ME: `/${API_VERSION}/auth/me`,
    CHANGE_PASSWORD: `/${API_VERSION}/auth/change-password`,
  },
  USERS: {
    BASE: `/${API_VERSION}/users`,
    BY_ID: (id: string) => `/${API_VERSION}/users/${id}`,
    ME: `/${API_VERSION}/users/me`,
    REGISTER: `/${API_VERSION}/users`,
  },
} as const

export const QUERY_KEYS = {
  AUTH: {
    ALL: ['auth'] as const,
    ME: () => [...QUERY_KEYS.AUTH.ALL, 'me'] as const,
  },
  USERS: {
    ALL: ['users'] as const,
    BY_ID: (id: string) => [...QUERY_KEYS.USERS.ALL, 'detail', id] as const,
    ME: () => [...QUERY_KEYS.USERS.ALL, 'me'] as const,
  },
} as const

export const QUERY_CONFIG = {
  STALE_TIME: {
    USER: 1000 * 60 * 5,
    STATIC: Number.POSITIVE_INFINITY,
    FREQUENT: 1000 * 30,
  },
  GC_TIME: {
    DEFAULT: 1000 * 60 * 30,
    LONG: 1000 * 60 * 60,
  },
  RETRY: {
    DEFAULT: 3,
    NONE: 0,
  },
} as const

export const HTTP_STATUS = {
  OK: 200,
  CREATED: 201,
  NO_CONTENT: 204,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  CONFLICT: 409,
  TOO_MANY_REQUESTS: 429,
  INTERNAL_SERVER: 500,
} as const

export const PASSWORD_CONSTRAINTS = {
  MIN_LENGTH: 8,
  MAX_LENGTH: 128,
} as const

export const PAGINATION = {
  DEFAULT_PAGE: 1,
  DEFAULT_SIZE: 20,
  MAX_SIZE: 100,
} as const

export type ApiEndpoint = typeof API_ENDPOINTS
export type QueryKey = typeof QUERY_KEYS
