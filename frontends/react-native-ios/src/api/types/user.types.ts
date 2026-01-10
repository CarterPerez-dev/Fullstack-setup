/**
 * @AngelaMos | 2026
 * user.types.ts
 */

import { z } from 'zod'
import { PASSWORD_CONSTRAINTS } from '@/core/config'
import { userResponseSchema } from './auth.types'

export { type UserResponse, UserRole, userResponseSchema } from './auth.types'

export const userListResponseSchema = z.object({
  items: z.array(userResponseSchema),
  total: z.number(),
  page: z.number(),
  size: z.number(),
})

export const userCreateRequestSchema = z.object({
  email: z.string().email(),
  password: z
    .string()
    .min(PASSWORD_CONSTRAINTS.MIN_LENGTH)
    .max(PASSWORD_CONSTRAINTS.MAX_LENGTH),
  full_name: z.string().max(255).nullable().optional(),
})

export const userUpdateRequestSchema = z.object({
  full_name: z.string().max(255).nullable().optional(),
})

export const paginationParamsSchema = z.object({
  page: z.number().min(1),
  size: z.number().min(1).max(100),
})

export type UserListResponse = z.infer<typeof userListResponseSchema>
export type UserCreateRequest = z.infer<typeof userCreateRequestSchema>
export type UserUpdateRequest = z.infer<typeof userUpdateRequestSchema>
export type PaginationParams = z.infer<typeof paginationParamsSchema>

export const isValidUserListResponse = (
  data: unknown
): data is UserListResponse => {
  if (data === null || data === undefined) return false
  if (typeof data !== 'object') return false

  const result = userListResponseSchema.safeParse(data)
  return result.success
}

export const isValidUserCreateRequest = (
  data: unknown
): data is UserCreateRequest => {
  if (data === null || data === undefined) return false
  if (typeof data !== 'object') return false

  const result = userCreateRequestSchema.safeParse(data)
  return result.success
}

export class UserResponseError extends Error {
  readonly endpoint?: string

  constructor(message: string, endpoint?: string) {
    super(message)
    this.name = 'UserResponseError'
    this.endpoint = endpoint
    Object.setPrototypeOf(this, UserResponseError.prototype)
  }
}

export const USER_ERROR_MESSAGES = {
  INVALID_USER_RESPONSE: 'Invalid user data from server',
  INVALID_USER_LIST_RESPONSE: 'Invalid user list from server',
  USER_NOT_FOUND: 'User not found',
  EMAIL_ALREADY_EXISTS: 'Email already exists',
  FAILED_TO_CREATE: 'Failed to create user',
  FAILED_TO_UPDATE: 'Failed to update user',
} as const

export const USER_SUCCESS_MESSAGES = {
  PROFILE_UPDATED: 'Profile updated successfully',
  REGISTERED: 'Registration successful!',
} as const

export type UserErrorMessage =
  (typeof USER_ERROR_MESSAGES)[keyof typeof USER_ERROR_MESSAGES]
export type UserSuccessMessage =
  (typeof USER_SUCCESS_MESSAGES)[keyof typeof USER_SUCCESS_MESSAGES]
