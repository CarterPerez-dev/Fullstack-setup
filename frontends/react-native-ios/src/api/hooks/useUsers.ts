/**
 * @AngelaMos | 2026
 * useUsers.ts
 */

import {
  type UseMutationResult,
  type UseQueryResult,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import {
  isValidUserResponse,
  USER_ERROR_MESSAGES,
  type UserCreateRequest,
  type UserResponse,
  UserResponseError,
  type UserUpdateRequest,
} from '@/api/types'
import { apiClient, QUERY_STRATEGIES } from '@/core/api'
import { API_ENDPOINTS, QUERY_KEYS } from '@/core/config'
import { useAuthStore } from '@/core/lib'
import { authQueries } from './useAuth'

export const userQueries = {
  all: () => QUERY_KEYS.USERS.ALL,
  byId: (id: string) => QUERY_KEYS.USERS.BY_ID(id),
  me: () => QUERY_KEYS.USERS.ME(),
} as const

const fetchUserById = async (id: string): Promise<UserResponse> => {
  const response = await apiClient.get<unknown>(API_ENDPOINTS.USERS.BY_ID(id))
  const data: unknown = response.data

  if (!isValidUserResponse(data)) {
    throw new UserResponseError(
      USER_ERROR_MESSAGES.INVALID_USER_RESPONSE,
      API_ENDPOINTS.USERS.BY_ID(id)
    )
  }

  return data
}

export const useUser = (id: string): UseQueryResult<UserResponse, Error> => {
  return useQuery({
    queryKey: userQueries.byId(id),
    queryFn: () => fetchUserById(id),
    enabled: id.length > 0,
    ...QUERY_STRATEGIES.standard,
  })
}

const performRegister = async (
  data: UserCreateRequest
): Promise<UserResponse> => {
  const response = await apiClient.post<unknown>(
    API_ENDPOINTS.USERS.REGISTER,
    data
  )
  const responseData: unknown = response.data

  if (!isValidUserResponse(responseData)) {
    throw new UserResponseError(
      USER_ERROR_MESSAGES.INVALID_USER_RESPONSE,
      API_ENDPOINTS.USERS.REGISTER
    )
  }

  return responseData
}

export const useRegister = (): UseMutationResult<
  UserResponse,
  Error,
  UserCreateRequest
> => {
  return useMutation({
    mutationFn: performRegister,
  })
}

const performUpdateProfile = async (
  data: UserUpdateRequest
): Promise<UserResponse> => {
  const response = await apiClient.patch<unknown>(API_ENDPOINTS.USERS.ME, data)
  const responseData: unknown = response.data

  if (!isValidUserResponse(responseData)) {
    throw new UserResponseError(
      USER_ERROR_MESSAGES.INVALID_USER_RESPONSE,
      API_ENDPOINTS.USERS.ME
    )
  }

  return responseData
}

export const useUpdateProfile = (): UseMutationResult<
  UserResponse,
  Error,
  UserUpdateRequest
> => {
  const queryClient = useQueryClient()
  const updateUser = useAuthStore((s) => s.updateUser)

  return useMutation({
    mutationFn: performUpdateProfile,
    onSuccess: (data: UserResponse): void => {
      updateUser(data)

      queryClient.setQueryData(authQueries.me(), data)
      queryClient.setQueryData(userQueries.me(), data)
    },
  })
}
