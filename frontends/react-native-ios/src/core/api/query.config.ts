// ===================
// © AngelaMos | 2026
// query.config.ts
// ===================

import { createAsyncStoragePersister } from '@tanstack/query-async-storage-persister'
import { MutationCache, QueryCache, QueryClient } from '@tanstack/react-query'
import { QUERY_CONFIG } from '@/core/config'
import { queryClientStorage } from '@/core/storage'
import { ApiError, ApiErrorCode } from './errors'

const NO_RETRY_ERROR_CODES: readonly ApiErrorCode[] = [
  ApiErrorCode.AUTHENTICATION_ERROR,
  ApiErrorCode.AUTHORIZATION_ERROR,
  ApiErrorCode.NOT_FOUND,
  ApiErrorCode.VALIDATION_ERROR,
] as const

const shouldRetryQuery = (failureCount: number, error: Error): boolean => {
  if (error instanceof ApiError) {
    if (NO_RETRY_ERROR_CODES.includes(error.code)) {
      return false
    }
  }
  return failureCount < QUERY_CONFIG.RETRY.DEFAULT
}

const calculateRetryDelay = (attemptIndex: number): number => {
  const baseDelay = 1000
  const maxDelay = 30000
  return Math.min(baseDelay * 2 ** attemptIndex, maxDelay)
}

let showToast: ((message: string) => void) | null = null

export const setToastHandler = (handler: (message: string) => void): void => {
  showToast = handler
}

const handleQueryCacheError = (
  error: Error,
  query: { state: { data: unknown } }
): void => {
  if (query.state.data !== undefined) {
    const message =
      error instanceof ApiError
        ? error.getUserMessage()
        : 'Background update failed'
    showToast?.(message)
  }
}

const handleMutationCacheError = (
  error: Error,
  _variables: unknown,
  _context: unknown,
  mutation: { options: { onError?: unknown } }
): void => {
  if (mutation.options.onError === undefined) {
    const message =
      error instanceof ApiError ? error.getUserMessage() : 'Operation failed'
    showToast?.(message)
  }
}

export const QUERY_STRATEGIES = {
  standard: {
    staleTime: QUERY_CONFIG.STALE_TIME.USER,
    gcTime: QUERY_CONFIG.GC_TIME.DEFAULT,
  },
  frequent: {
    staleTime: QUERY_CONFIG.STALE_TIME.FREQUENT,
    gcTime: QUERY_CONFIG.GC_TIME.DEFAULT,
    refetchInterval: QUERY_CONFIG.STALE_TIME.FREQUENT,
  },
  static: {
    staleTime: QUERY_CONFIG.STALE_TIME.STATIC,
    gcTime: QUERY_CONFIG.GC_TIME.LONG,
    refetchOnMount: false,
    refetchOnWindowFocus: false,
  },
  auth: {
    staleTime: QUERY_CONFIG.STALE_TIME.USER,
    gcTime: QUERY_CONFIG.GC_TIME.DEFAULT,
    retry: QUERY_CONFIG.RETRY.NONE,
  },
} as const

export type QueryStrategy = keyof typeof QUERY_STRATEGIES

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: QUERY_CONFIG.STALE_TIME.USER,
      gcTime: QUERY_CONFIG.GC_TIME.DEFAULT,
      retry: shouldRetryQuery,
      retryDelay: calculateRetryDelay,
      networkMode: 'offlineFirst',
    },
    mutations: {
      retry: QUERY_CONFIG.RETRY.NONE,
      networkMode: 'offlineFirst',
    },
  },
  queryCache: new QueryCache({
    onError: handleQueryCacheError,
  }),
  mutationCache: new MutationCache({
    onError: handleMutationCacheError,
  }),
})

export const queryClientPersister = createAsyncStoragePersister({
  storage: queryClientStorage,
  key: 'query-cache',
})
