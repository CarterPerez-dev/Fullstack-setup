/**
 * @AngelaMos | 2026
 * useAuth.ts
 */

import {
  type UseMutationResult,
  type UseQueryResult,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { router } from 'expo-router'
import {
  AUTH_ERROR_MESSAGES,
  AuthResponseError,
  isValidLogoutAllResponse,
  isValidMobileLoginResponse,
  isValidUserResponse,
  type LoginRequest,
  type LogoutAllResponse,
  type MobileLoginResponse,
  type PasswordChangeRequest,
  type UserResponse,
} from '@/api/types'
import { apiClient, QUERY_STRATEGIES, setAccessToken } from '@/core/api'
import { API_ENDPOINTS, QUERY_KEYS } from '@/core/config'
import { useAuthStore } from '@/core/lib'
import { SECURE_KEYS, secureStorage } from '@/core/storage'

export const authQueries = {
  all: () => QUERY_KEYS.AUTH.ALL,
  me: () => QUERY_KEYS.AUTH.ME(),
} as const

const fetchCurrentUser = async (): Promise<UserResponse> => {
  const response = await apiClient.get<unknown>(API_ENDPOINTS.AUTH.ME)
  const data: unknown = response.data

  if (!isValidUserResponse(data)) {
    throw new AuthResponseError(
      AUTH_ERROR_MESSAGES.INVALID_USER_RESPONSE,
      API_ENDPOINTS.AUTH.ME
    )
  }

  return data
}

export const useCurrentUser = (): UseQueryResult<UserResponse, Error> => {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)

  return useQuery({
    queryKey: authQueries.me(),
    queryFn: fetchCurrentUser,
    enabled: isAuthenticated,
    ...QUERY_STRATEGIES.auth,
  })
}

const performLogin = async (
  credentials: LoginRequest
): Promise<MobileLoginResponse> => {
  const formData = new URLSearchParams()
  formData.append('username', credentials.username)
  formData.append('password', credentials.password)

  const response = await apiClient.post<unknown>(
    API_ENDPOINTS.AUTH.LOGIN,
    formData,
    {
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    }
  )

  const data: unknown = response.data

  if (!isValidMobileLoginResponse(data)) {
    throw new AuthResponseError(
      AUTH_ERROR_MESSAGES.INVALID_LOGIN_RESPONSE,
      API_ENDPOINTS.AUTH.LOGIN
    )
  }

  return data
}

export const useLogin = (): UseMutationResult<
  MobileLoginResponse,
  Error,
  LoginRequest
> => {
  const queryClient = useQueryClient()
  const login = useAuthStore((s) => s.login)

  return useMutation({
    mutationFn: performLogin,
    onSuccess: async (data: MobileLoginResponse): Promise<void> => {
      setAccessToken(data.access_token)
      await secureStorage.setItem(SECURE_KEYS.REFRESH_TOKEN, data.refresh_token)

      login(data.user)

      queryClient.setQueryData(authQueries.me(), data.user)
    },
  })
}

const performLogout = async (): Promise<void> => {
  await apiClient.post(API_ENDPOINTS.AUTH.LOGOUT)
}

export const useLogout = (): UseMutationResult<void, Error, void> => {
  const queryClient = useQueryClient()
  const logout = useAuthStore((s) => s.logout)

  return useMutation({
    mutationFn: performLogout,
    onSuccess: async (): Promise<void> => {
      setAccessToken(null)
      await secureStorage.deleteItem(SECURE_KEYS.REFRESH_TOKEN)

      logout()

      queryClient.removeQueries({ queryKey: authQueries.all() })

      router.replace('/(auth)/login')
    },
    onError: async (): Promise<void> => {
      setAccessToken(null)
      await secureStorage.deleteItem(SECURE_KEYS.REFRESH_TOKEN)

      logout()

      queryClient.removeQueries({ queryKey: authQueries.all() })

      router.replace('/(auth)/login')
    },
  })
}

const performLogoutAll = async (): Promise<LogoutAllResponse> => {
  const response = await apiClient.post<unknown>(API_ENDPOINTS.AUTH.LOGOUT_ALL)
  const data: unknown = response.data

  if (!isValidLogoutAllResponse(data)) {
    throw new AuthResponseError(
      AUTH_ERROR_MESSAGES.INVALID_LOGOUT_RESPONSE,
      API_ENDPOINTS.AUTH.LOGOUT_ALL
    )
  }

  return data
}

export const useLogoutAll = (): UseMutationResult<
  LogoutAllResponse,
  Error,
  void
> => {
  const queryClient = useQueryClient()
  const logout = useAuthStore((s) => s.logout)

  return useMutation({
    mutationFn: performLogoutAll,
    onSuccess: async (): Promise<void> => {
      setAccessToken(null)
      await secureStorage.deleteItem(SECURE_KEYS.REFRESH_TOKEN)

      logout()

      queryClient.removeQueries({ queryKey: authQueries.all() })

      router.replace('/(auth)/login')
    },
  })
}

const performPasswordChange = async (
  data: PasswordChangeRequest
): Promise<void> => {
  await apiClient.post(API_ENDPOINTS.AUTH.CHANGE_PASSWORD, data)
}

export const useChangePassword = (): UseMutationResult<
  void,
  Error,
  PasswordChangeRequest
> => {
  return useMutation({
    mutationFn: performPasswordChange,
  })
}
