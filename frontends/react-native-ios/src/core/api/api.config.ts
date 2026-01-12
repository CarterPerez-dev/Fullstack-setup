// ===================
// © AngelaMos | 2026
// api.config.ts
// ===================

import axios, {
  type AxiosError,
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from 'axios'
import { API_ENDPOINTS, APP_CONFIG, HTTP_STATUS } from '@/core/config'
import { SECURE_KEYS, secureStorage } from '@/core/storage'
import { ApiError, ApiErrorCode, transformAxiosError } from './errors'

interface RequestConfig extends InternalAxiosRequestConfig {
  _retry?: boolean
}

interface RefreshSubscriber {
  resolve: (token: string) => void
  reject: (error: Error) => void
}

interface RefreshResponse {
  access_token: string
  refresh_token: string
}

const getBaseURL = (): string => {
  if (__DEV__) {
    return `http://${APP_CONFIG.DEV_IP}:${APP_CONFIG.DEV_PORT}`
  }
  return APP_CONFIG.PROD_API_URL
}

export const apiClient: AxiosInstance = axios.create({
  baseURL: getBaseURL(),
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

let isRefreshing = false
let refreshSubscribers: RefreshSubscriber[] = []
let accessToken: string | null = null

export const setAccessToken = (token: string | null): void => {
  accessToken = token
}

export const getAccessToken = (): string | null => {
  return accessToken
}

const processRefreshQueue = (error: Error | null, token: string | null): void => {
  refreshSubscribers.forEach((subscriber) => {
    if (error !== null) {
      subscriber.reject(error)
    } else if (token !== null) {
      subscriber.resolve(token)
    }
  })
  refreshSubscribers = []
}

const addRefreshSubscriber = (
  resolve: (token: string) => void,
  reject: (error: Error) => void
): void => {
  refreshSubscribers.push({ resolve, reject })
}

const handleTokenRefresh = async (): Promise<RefreshResponse> => {
  const refreshToken = await secureStorage.getItem(SECURE_KEYS.REFRESH_TOKEN)

  if (!refreshToken) {
    throw new ApiError(
      'No refresh token',
      ApiErrorCode.AUTHENTICATION_ERROR,
      HTTP_STATUS.UNAUTHORIZED
    )
  }

  const response = await apiClient.post<RefreshResponse>(
    API_ENDPOINTS.AUTH.REFRESH_MOBILE,
    { refresh_token: refreshToken }
  )

  if (
    response.data === null ||
    response.data === undefined ||
    typeof response.data !== 'object'
  ) {
    throw new ApiError(
      'Invalid refresh response',
      ApiErrorCode.AUTHENTICATION_ERROR,
      HTTP_STATUS.UNAUTHORIZED
    )
  }

  const { access_token, refresh_token: newRefreshToken } = response.data

  if (typeof access_token !== 'string' || access_token.length === 0) {
    throw new ApiError(
      'Invalid access token',
      ApiErrorCode.AUTHENTICATION_ERROR,
      HTTP_STATUS.UNAUTHORIZED
    )
  }

  await secureStorage.setItem(SECURE_KEYS.REFRESH_TOKEN, newRefreshToken)

  return response.data
}

let onAuthFailure: (() => void) | null = null

export const setAuthFailureHandler = (handler: () => void): void => {
  onAuthFailure = handler
}

const handleAuthFailure = async (): Promise<void> => {
  await secureStorage.deleteItem(SECURE_KEYS.REFRESH_TOKEN)
  setAccessToken(null)
  onAuthFailure?.()
}

apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig): InternalAxiosRequestConfig => {
    const token = getAccessToken()
    if (token !== null && token.length > 0) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error: unknown): Promise<never> => {
    return Promise.reject(error)
  }
)

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError): Promise<unknown> => {
    const originalRequest = error.config as RequestConfig | undefined

    if (originalRequest === undefined) {
      return Promise.reject(transformAxiosError(error))
    }

    const isUnauthorized = error.response?.status === HTTP_STATUS.UNAUTHORIZED
    const isNotRetried = originalRequest._retry !== true
    const isNotAuthEndpoint =
      !originalRequest.url?.includes('refresh') &&
      !originalRequest.url?.includes('login') &&
      !originalRequest.url?.includes('register')

    if (isUnauthorized && isNotRetried && isNotAuthEndpoint) {
      if (isRefreshing) {
        return new Promise<unknown>((resolve, reject) => {
          addRefreshSubscriber(
            (newToken: string): void => {
              originalRequest.headers.Authorization = `Bearer ${newToken}`
              resolve(apiClient(originalRequest))
            },
            (refreshError: Error): void => {
              reject(refreshError)
            }
          )
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        const { access_token: newToken } = await handleTokenRefresh()
        setAccessToken(newToken)
        processRefreshQueue(null, newToken)
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return await apiClient(originalRequest)
      } catch (refreshError: unknown) {
        const apiError =
          refreshError instanceof ApiError
            ? refreshError
            : new ApiError(
                'Session expired',
                ApiErrorCode.AUTHENTICATION_ERROR,
                HTTP_STATUS.UNAUTHORIZED
              )
        processRefreshQueue(apiError, null)
        await handleAuthFailure()
        return Promise.reject(apiError)
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(transformAxiosError(error))
  }
)
