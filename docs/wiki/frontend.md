# Frontend Architecture Documentation

**React + TypeScript + SCSS Template | 2025**

## Overview

This frontend is a modern, production grade React application built with TypeScript and 2025 best practices. It features a complete design system, robust state management, type-safe API integration, and blazing-fast development experience with Vite 7.

## Tech Stack

- **Framework**: React 19.2.1
- **Language**: TypeScript 5.9+ (strict mode)
- **Build Tool**: Vite 7 (Rolldown variant - next-gen bundler)
- **Routing**: React Router v7.1.1
- **Server State**: TanStack Query v5.90.12
- **Client State**: Zustand v5.0.9
- **HTTP Client**: Axios v1.13.0
- **Validation**: Zod v4.1.13
- **UI Library**: Custom components + design system
- **Styling**: SASS v1.95.0 (modern compiler API) + CSS Modules
- **Notifications**: Sonner v2.0.7
- **Error Boundaries**: React Error Boundary v6.0.0
- **Code Quality**: Biome v2.3.8 (linting & formatting)
- **Package Manager**: pnpm (fast, efficient)

## Architecture

### Clean Architecture Pattern

```
Pages (UI Layer)
    ↓
Hooks (API Integration)
    ↓
API Client (HTTP Layer)
    ↓
Stores (State Management)
    ↓
Types (Data Contracts)
```

**Key Principles:**
- Component isolation with CSS Modules
- Type safety with Zod schemas
- Separation of server vs client state
- Route-based code splitting
- Design token system for consistency

### Project Structure

```
frontend/
├── src/
│   ├── core/              # Framework & infrastructure
│   │   ├── app/           # App shell & routing
│   │   │   ├── routers.tsx         # Route definitions
│   │   │   ├── protected-route.tsx # Auth guard
│   │   │   ├── shell.tsx           # Layout shell
│   │   │   ├── shell.module.scss   # Shell styles
│   │   │   └── toast.module.scss   # Toast styles
│   │   ├── api/           # API configuration
│   │   │   ├── api.config.ts       # Axios instance
│   │   │   ├── query.config.ts     # Query client
│   │   │   └── errors.ts           # Error handling
│   │   └── lib/           # Core utilities
│   │       ├── auth.store.ts       # Auth state
│   │       └── ui.store.ts         # UI state
│   ├── api/               # Domain API layer
│   │   ├── hooks/         # TanStack Query hooks
│   │   │   ├── useAuth.ts
│   │   │   ├── useUsers.ts
│   │   │   └── useAdmin.ts
│   │   └── types/         # API types & schemas
│   │       ├── auth.types.ts
│   │       └── user.types.ts
│   ├── pages/             # Route components
│   │   └── home/
│   │       ├── index.tsx
│   │       └── home.module.scss
│   ├── styles/            # Design system
│   │   ├── _tokens.scss   # Design tokens
│   │   ├── _mixins.scss   # SCSS mixins
│   │   ├── _fonts.scss    # Typography
│   │   └── _reset.scss    # CSS reset
│   ├── App.tsx            # App root
│   ├── main.tsx           # Entry point
│   ├── config.ts          # App configuration
│   └── styles.scss        # Global styles
├── public/                # Static assets
│   └── assets/
│       └── icons/         # Favicons & PWA icons
├── index.html             # HTML entry
├── vite.config.ts         # Vite configuration
├── tsconfig.json          # TypeScript config
├── biome.json             # Linter config
├── stylelint.config.js    # SCSS linter
└── package.json           # Dependencies
```

## Core Components

### 1. Application Entry (`src/main.tsx`, `src/App.tsx`)

**main.tsx** - Minimal entry point:
```tsx
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
```

**App.tsx** - Provider setup:
```tsx
import { QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider } from 'react-router';
import { Toaster } from 'sonner';
import { router } from '@/core/app/routers';
import { queryClient } from '@/core/api/query.config';

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
      <Toaster position="top-right" richColors />
      {import.meta.env.DEV && <ReactQueryDevtools />}
    </QueryClientProvider>
  );
}
```

### 2. Routing System (`src/core/app/routers.tsx`)

**Pattern**: React Router v7 with lazy loading

```tsx
import { createBrowserRouter } from 'react-router';
import ProtectedRoute from './protected-route';
import Shell from './shell';
import { UserRole } from '@/api/types/auth.types';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <Shell />,
    children: [
      {
        index: true,
        lazy: () => import('@/pages/home'),
      },
      {
        path: 'dashboard',
        element: <ProtectedRoute />,
        lazy: () => import('@/pages/dashboard'),
      },
      {
        path: 'settings',
        element: <ProtectedRoute />,
        lazy: () => import('@/pages/settings'),
      },
      {
        path: 'admin/*',
        element: <ProtectedRoute allowedRoles={[UserRole.ADMIN]} />,
        lazy: () => import('@/pages/admin'),
      },
    ],
  },
  {
    path: '/login',
    lazy: () => import('@/pages/login'),
  },
  {
    path: '/register',
    lazy: () => import('@/pages/register'),
  },
  {
    path: '/unauthorized',
    lazy: () => import('@/pages/unauthorized'),
  },
]);
```

**Features:**
- Route-based code splitting (lazy loading)
- Protected routes with role-based access
- Nested routes with shared layout
- Shell wrapper for authenticated pages

### 3. Protected Routes (`src/core/app/protected-route.tsx`)

**Pattern**: Authentication guard with role-based authorization

```tsx
import { Navigate, Outlet, useLocation } from 'react-router';
import { useIsAuthenticated, useHasRole } from '@/core/lib/auth.store';
import { UserRole } from '@/api/types/auth.types';

interface ProtectedRouteProps {
  allowedRoles?: UserRole[];
}

export default function ProtectedRoute({ allowedRoles }: ProtectedRouteProps) {
  const isAuthenticated = useIsAuthenticated();
  const hasRole = useHasRole();
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (allowedRoles && !hasRole(allowedRoles)) {
    return <Navigate to="/unauthorized" replace />;
  }

  return <Outlet />;
}
```

**Features:**
- Authentication check
- Role-based authorization
- Redirect with state preservation
- Granular access control per route

### 4. Application Shell (`src/core/app/shell.tsx`)

**Pattern**: Layout wrapper with sidebar + header

```tsx
import { Suspense } from 'react';
import { Outlet } from 'react-router';
import { ErrorBoundary } from 'react-error-boundary';
import styles from './shell.module.scss';

export default function Shell() {
  return (
    <div className={styles.shell}>
      <aside className={styles.sidebar}>
        {/* Navigation */}
      </aside>
      <div className={styles.main}>
        <header className={styles.header}>
          {/* App header */}
        </header>
        <main className={styles.content}>
          <ErrorBoundary fallback={<ErrorFallback />}>
            <Suspense fallback={<LoadingSpinner />}>
              <Outlet />
            </Suspense>
          </ErrorBoundary>
        </main>
      </div>
    </div>
  );
}
```

**Layout:**
- Fixed sidebar (260px wide, responsive)
- Sticky header (64px tall)
- Scrollable content area
- Error boundaries for resilience
- Suspense for lazy-loaded routes

**SCSS Architecture (`shell.module.scss`):**
```scss
.shell {
  display: grid;
  grid-template-columns: var(--sidebar-width, 260px) 1fr;
  grid-template-rows: 100dvh;

  @include breakpoint-down(md) {
    grid-template-columns: 1fr;
  }
}

.sidebar {
  position: sticky;
  top: 0;
  height: 100dvh;
  overflow-y: auto;
}

.main {
  display: flex;
  flex-direction: column;
  height: 100dvh;
}

.header {
  position: sticky;
  top: 0;
  height: var(--header-height, 64px);
  z-index: $z-header;
}

.content {
  flex: 1;
  overflow-y: auto;
}
```

### 5. State Management

#### Zustand Stores

**Auth Store (`src/core/lib/auth.store.ts`):**

```tsx
import { create } from 'zustand';
import { persist, devtools } from 'zustand/middleware';
import type { User, UserRole } from '@/api/types/auth.types';

interface AuthState {
  // State
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  // Actions
  login: (user: User, token: string) => void;
  logout: () => void;
  setAccessToken: (token: string) => void;
  updateUser: (user: User) => void;
  setLoading: (loading: boolean) => void;
}

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      (set) => ({
        // Initial state
        user: null,
        accessToken: null,
        isAuthenticated: false,
        isLoading: true,

        // Actions
        login: (user, token) => set({
          user,
          accessToken: token,
          isAuthenticated: true,
        }),
        logout: () => set({
          user: null,
          accessToken: null,
          isAuthenticated: false,
        }),
        setAccessToken: (token) => set({ accessToken: token }),
        updateUser: (user) => set({ user }),
        setLoading: (loading) => set({ isLoading: loading }),
      }),
      {
        name: STORAGE_KEYS.AUTH,
        partialize: (state) => ({
          user: state.user,
          isAuthenticated: state.isAuthenticated,
        }), // Don't persist token (security)
      }
    ),
    { name: 'Auth Store' }
  )
);

// Granular selectors
export const useUser = () => useAuthStore((s) => s.user);
export const useIsAuthenticated = () => useAuthStore((s) => s.isAuthenticated);
export const useIsAdmin = () => useAuthStore((s) => s.user?.role === UserRole.ADMIN);
export const useHasRole = () => {
  const user = useUser();
  return (roles: UserRole[]) => user && roles.includes(user.role);
};
```

**Features:**
- Persistent auth state (localStorage)
- DevTools integration
- Granular selectors (avoid re-renders)
- Security: Token not persisted (memory only)

**UI Store (`src/core/lib/ui.store.ts`):**

```tsx
interface UIState {
  // Theme
  theme: 'light' | 'dark' | 'system';
  setTheme: (theme: 'light' | 'dark' | 'system') => void;

  // Sidebar
  sidebarOpen: boolean;
  sidebarCollapsed: boolean;
  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
  toggleSidebarCollapsed: () => void;
}
```

**Features:**
- Theme management (light/dark/system)
- Sidebar state (open/collapsed)
- Persistent preferences

### 6. API Layer

#### Axios Client (`src/core/api/api.config.ts`)

**Configuration:**
```tsx
import axios from 'axios';
import { useAuthStore } from '@/core/lib/auth.store';

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api',
  timeout: 15000,
  withCredentials: true, // Send cookies
  headers: {
    'Content-Type': 'application/json',
  },
});
```

**Request Interceptor** - Add Bearer token:
```tsx
apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
```

**Response Interceptor** - Token refresh on 401:
```tsx
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // If 401 and not already retrying
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      // Add to refresh queue (batch concurrent requests)
      if (!isRefreshing) {
        isRefreshing = true;
        try {
          const { data } = await axios.post('/api/v1/auth/refresh', null, {
            withCredentials: true,
          });
          useAuthStore.getState().setAccessToken(data.access_token);
          isRefreshing = false;
          processQueue(null, data.access_token);
        } catch (err) {
          processQueue(err, null);
          useAuthStore.getState().logout();
          window.location.href = '/login';
          return Promise.reject(err);
        }
      }

      // Wait for refresh to complete
      return new Promise((resolve, reject) => {
        failedQueue.push({ resolve, reject });
      }).then((token) => {
        originalRequest.headers.Authorization = `Bearer ${token}`;
        return axios(originalRequest);
      });
    }

    return Promise.reject(transformError(error));
  }
);
```

**Features:**
- Automatic token refresh on 401
- Refresh queue (batches concurrent requests)
- Auto-logout on refresh failure
- Error transformation

#### Query Client (`src/core/api/query.config.ts`)

**Configuration:**
```tsx
import { QueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { ApiError } from './errors';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => {
        // No retry for auth/validation errors
        if (error instanceof ApiError) {
          if (['AUTHENTICATION_ERROR', 'VALIDATION_ERROR'].includes(error.code)) {
            return false;
          }
        }
        return failureCount < 3;
      },
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      staleTime: QUERY_CONFIG.staleTime.standard, // 5 minutes
      gcTime: QUERY_CONFIG.gcTime.standard, // 30 minutes
    },
    mutations: {
      retry: false,
      onError: (error) => {
        if (error instanceof ApiError) {
          toast.error(error.message);
        }
      },
    },
  },
});
```

**Query Strategies** (defined in `config.ts`):
```tsx
export const QUERY_CONFIG = {
  staleTime: {
    standard: 5 * 60 * 1000,      // 5 minutes
    frequent: 30 * 1000,           // 30 seconds
    static: Number.POSITIVE_INFINITY, // Never stale
    auth: 5 * 60 * 1000,           // 5 minutes
  },
  gcTime: {
    standard: 30 * 60 * 1000,      // 30 minutes
    frequent: 5 * 60 * 1000,       // 5 minutes
    static: 60 * 60 * 1000,        // 1 hour
    auth: 10 * 60 * 1000,          // 10 minutes
  },
  retry: {
    max: 3,
    delay: 1000,
  },
  refetchOnWindowFocus: {
    frequent: true,
    standard: false,
  },
};
```

#### Error Handling (`src/core/api/errors.ts`)

**Custom Error Class:**
```tsx
export class ApiError extends Error {
  constructor(
    message: string,
    public code: ErrorCode,
    public status?: number,
    public details?: Record<string, unknown>
  ) {
    super(message);
    this.name = 'ApiError';
  }

  static fromAxiosError(error: AxiosError): ApiError {
    const status = error.response?.status;
    const data = error.response?.data as any;

    switch (status) {
      case 400:
        return new ApiError(
          data?.detail || 'Invalid request',
          'VALIDATION_ERROR',
          status,
          data
        );
      case 401:
        return new ApiError(
          'Authentication required',
          'AUTHENTICATION_ERROR',
          status
        );
      case 403:
        return new ApiError(
          'Access denied',
          'AUTHORIZATION_ERROR',
          status
        );
      case 404:
        return new ApiError(
          data?.detail || 'Resource not found',
          'NOT_FOUND',
          status
        );
      case 409:
        return new ApiError(
          data?.detail || 'Resource conflict',
          'CONFLICT',
          status
        );
      case 429:
        return new ApiError(
          'Too many requests',
          'RATE_LIMITED',
          status
        );
      case 500:
      case 502:
      case 503:
        return new ApiError(
          'Server error. Please try again later.',
          'SERVER_ERROR',
          status
        );
      default:
        return new ApiError(
          'An unexpected error occurred',
          'UNKNOWN_ERROR',
          status
        );
    }
  }
}

// Type registration for TanStack Query
declare module '@tanstack/react-query' {
  interface Register {
    defaultError: ApiError;
  }
}
```

### 7. API Hooks (TanStack Query)

#### Auth Hooks (`src/api/hooks/useAuth.ts`)

**Fetch Current User:**
```tsx
export function useCurrentUser() {
  const isAuthenticated = useIsAuthenticated();

  return useQuery({
    queryKey: QUERY_KEYS.auth.me(),
    queryFn: async () => {
      const { data } = await apiClient.get<UserResponse>(
        API_ENDPOINTS.auth.me
      );
      return isValidUserResponse(data) ? data : null;
    },
    enabled: isAuthenticated,
    staleTime: QUERY_CONFIG.staleTime.auth,
  });
}
```

**Login Mutation:**
```tsx
export function useLogin() {
  const login = useAuthStore((s) => s.login);
  const navigate = useNavigate();

  return useMutation({
    mutationFn: async (credentials: LoginRequest) => {
      // OAuth2 form-data format
      const formData = new URLSearchParams();
      formData.append('username', credentials.email);
      formData.append('password', credentials.password);

      const { data } = await apiClient.post<TokenWithUserResponse>(
        API_ENDPOINTS.auth.login,
        formData,
        { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
      );
      return data;
    },
    onSuccess: (data) => {
      login(data.user, data.access_token);
      toast.success('Login successful');
      navigate('/dashboard');
    },
    onError: (error: ApiError) => {
      toast.error(error.message);
    },
  });
}
```

**Logout Mutation:**
```tsx
export function useLogout() {
  const logout = useAuthStore((s) => s.logout);
  const navigate = useNavigate();

  return useMutation({
    mutationFn: async () => {
      await apiClient.post(API_ENDPOINTS.auth.logout);
    },
    onSuccess: () => {
      logout();
      queryClient.clear(); // Clear all cached queries
      toast.success('Logged out successfully');
      navigate('/login');
    },
  });
}
```

**Password Change:**
```tsx
export function useChangePassword() {
  const logout = useAuthStore((s) => s.logout);

  return useMutation({
    mutationFn: async (data: PasswordChangeRequest) => {
      await apiClient.post(API_ENDPOINTS.auth.changePassword, data);
    },
    onSuccess: () => {
      toast.success('Password changed. Please log in again.');
      logout();
      queryClient.clear();
    },
  });
}
```

#### User Hooks (`src/api/hooks/useUsers.ts`)

**Register:**
```tsx
export function useRegister() {
  return useMutation({
    mutationFn: async (data: RegisterRequest) => {
      const { data: user } = await apiClient.post<UserResponse>(
        API_ENDPOINTS.users.register,
        data
      );
      return user;
    },
    onSuccess: () => {
      toast.success('Registration successful. Please log in.');
    },
  });
}
```

**Update Profile:**
```tsx
export function useUpdateProfile() {
  const updateUser = useAuthStore((s) => s.updateUser);

  return useMutation({
    mutationFn: async (data: UserUpdate) => {
      const { data: user } = await apiClient.patch<UserResponse>(
        API_ENDPOINTS.users.me,
        data
      );
      return user;
    },
    onSuccess: (user) => {
      updateUser(user);
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.auth.me() });
      toast.success('Profile updated');
    },
  });
}
```

#### Admin Hooks (`src/api/hooks/useAdmin.ts`)

**List Users (Paginated):**
```tsx
export function useAdminUsers(page = 1, limit = 20) {
  return useQuery({
    queryKey: QUERY_KEYS.admin.users.list(page, limit),
    queryFn: async () => {
      const { data } = await apiClient.get<UserListResponse>(
        API_ENDPOINTS.admin.users.list,
        { params: { skip: (page - 1) * limit, limit } }
      );
      return data;
    },
  });
}
```

**Create User (Admin):**
```tsx
export function useAdminCreateUser() {
  return useMutation({
    mutationFn: async (data: UserCreateAdmin) => {
      const { data: user } = await apiClient.post<UserResponse>(
        API_ENDPOINTS.admin.users.create,
        data
      );
      return user;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.admin.users.list() });
      toast.success('User created');
    },
  });
}
```

**Delete User:**
```tsx
export function useAdminDeleteUser() {
  return useMutation({
    mutationFn: async (userId: string) => {
      await apiClient.delete(API_ENDPOINTS.admin.users.byId(userId));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.admin.users.list() });
      toast.success('User deleted');
    },
  });
}
```

### 8. Type System

#### Auth Types (`src/api/types/auth.types.ts`)

**Zod Schemas for Runtime Validation:**
```tsx
import { z } from 'zod';

export const userResponseSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  full_name: z.string().nullable(),
  is_active: z.boolean(),
  is_verified: z.boolean(),
  role: z.enum(['UNKNOWN', 'USER', 'ADMIN']),
  created_at: z.string().datetime(),
  updated_at: z.string().datetime(),
});

export const loginRequestSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
});

export const registerRequestSchema = loginRequestSchema.extend({
  full_name: z.string().optional(),
});
```

**TypeScript Types from Schemas:**
```tsx
export type UserResponse = z.infer<typeof userResponseSchema>;
export type LoginRequest = z.infer<typeof loginRequestSchema>;
export type RegisterRequest = z.infer<typeof registerRequestSchema>;
```

**Validation Helpers:**
```tsx
export function isValidUserResponse(data: unknown): data is UserResponse {
  return userResponseSchema.safeParse(data).success;
}
```

**Error Messages:**
```tsx
export const AUTH_ERRORS = {
  INVALID_CREDENTIALS: 'Invalid email or password',
  TOKEN_EXPIRED: 'Your session has expired. Please log in again.',
  UNAUTHORIZED: 'You are not authorized to access this resource',
} as const;

export const AUTH_SUCCESS = {
  LOGIN: 'Login successful',
  LOGOUT: 'Logged out successfully',
  PASSWORD_CHANGED: 'Password changed successfully',
} as const;
```

### 9. Configuration (`src/config.ts`)

**Centralized Configuration:**

```tsx
export const API_ENDPOINTS = {
  auth: {
    login: '/v1/auth/login',
    logout: '/v1/auth/logout',
    logoutAll: '/v1/auth/logout-all',
    refresh: '/v1/auth/refresh',
    me: '/v1/auth/me',
    changePassword: '/v1/auth/change-password',
  },
  users: {
    register: '/v1/users',
    me: '/v1/users/me',
    byId: (id: string) => `/v1/users/${id}`,
  },
  admin: {
    users: {
      list: '/v1/admin/users',
      create: '/v1/admin/users',
      byId: (id: string) => `/v1/admin/users/${id}`,
    },
  },
} as const;

export const QUERY_KEYS = {
  auth: {
    all: ['auth'] as const,
    me: () => [...QUERY_KEYS.auth.all, 'me'] as const,
  },
  users: {
    all: ['users'] as const,
    byId: (id: string) => [...QUERY_KEYS.users.all, id] as const,
  },
  admin: {
    users: {
      all: ['admin', 'users'] as const,
      list: (page?: number, limit?: number) =>
        [...QUERY_KEYS.admin.users.all, 'list', { page, limit }] as const,
      byId: (id: string) => [...QUERY_KEYS.admin.users.all, id] as const,
    },
  },
} as const;

export const ROUTES = {
  HOME: '/',
  LOGIN: '/login',
  REGISTER: '/register',
  DASHBOARD: '/dashboard',
  SETTINGS: '/settings',
  ADMIN: '/admin',
  UNAUTHORIZED: '/unauthorized',
} as const;

export const STORAGE_KEYS = {
  AUTH: 'auth-storage',
  UI: 'ui-storage',
} as const;
```

### 10. Styling Architecture

#### Design Token System (`src/styles/_tokens.scss`)

**Modern OKLCH Color System:**
```scss
// OKLCH = Perceptually uniform, wider gamut than RGB
$blue-500: oklch(0.58 0.20 250); // Primary blue
$gray-100: oklch(0.98 0.01 270); // Light gray
$gray-900: oklch(0.25 0.02 270); // Dark gray

// 11-step scales for each color
$blue-50, $blue-100, ..., $blue-950
$green-50, $green-100, ..., $green-950
$red-50, $red-100, ..., $red-950
```

**Semantic CSS Variables:**
```scss
:root {
  // Colors
  --color-bg: #{$gray-50};
  --color-text: #{$gray-900};
  --color-primary: #{$blue-500};
  --color-success: #{$green-500};
  --color-error: #{$red-500};
  --color-warning: #{$yellow-500};

  // Spacing (8px base)
  --spacing-xs: 0.5rem;   // 8px
  --spacing-sm: 0.75rem;  // 12px
  --spacing-md: 1rem;     // 16px
  --spacing-lg: 1.5rem;   // 24px
  --spacing-xl: 2rem;     // 32px

  // Typography (Major Third scale: 1.25)
  --text-xs: 0.64rem;     // 10px
  --text-sm: 0.8rem;      // 13px
  --text-base: 1rem;      // 16px
  --text-lg: 1.25rem;     // 20px
  --text-xl: 1.563rem;    // 25px
  --text-2xl: 1.953rem;   // 31px

  // Shadows
  --shadow-sm: 0 1px 2px oklch(0 0 0 / 0.05);
  --shadow-md: 0 4px 6px oklch(0 0 0 / 0.07);
  --shadow-lg: 0 10px 15px oklch(0 0 0 / 0.10);

  // Border radius
  --radius-sm: 0.25rem;   // 4px
  --radius-md: 0.5rem;    // 8px
  --radius-lg: 0.75rem;   // 12px
  --radius-xl: 1rem;      // 16px
  --radius-full: 9999px;

  // Z-index layers
  --z-dropdown: 1000;
  --z-sticky: 1020;
  --z-fixed: 1030;
  --z-modal-backdrop: 1040;
  --z-modal: 1050;
  --z-popover: 1060;
  --z-tooltip: 1070;
  --z-toast: 1080;
}
```

**Breakpoints:**
```scss
$breakpoints: (
  xs: 360px,
  sm: 640px,
  md: 768px,
  lg: 1024px,
  xl: 1280px,
  2xl: 1536px,
  3xl: 1920px,
);
```

#### SCSS Mixins (`src/styles/_mixins.scss`)

**Responsive Breakpoints:**
```scss
@mixin breakpoint-up($size) {
  @media (min-width: map-get($breakpoints, $size)) {
    @content;
  }
}

@mixin breakpoint-down($size) {
  @media (max-width: map-get($breakpoints, $size) - 1px) {
    @content;
  }
}
```

**Flexbox Utilities:**
```scss
@mixin flex-center {
  display: flex;
  align-items: center;
  justify-content: center;
}

@mixin flex-between {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
```

**Accessibility:**
```scss
@mixin sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  border: 0;
}

@mixin focus-ring {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
```

**Text Utilities:**
```scss
@mixin truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@mixin line-clamp($lines: 2) {
  display: -webkit-box;
  -webkit-line-clamp: $lines;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
```

#### CSS Modules Pattern

**Component Example (`home.module.scss`):**
```scss
@use '@/styles/tokens' as *;
@use '@/styles/mixins' as *;

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: var(--spacing-lg);

  @include breakpoint-down(md) {
    padding: var(--spacing-md);
  }
}

.title {
  font-size: var(--text-2xl);
  font-weight: 700;
  color: var(--color-text);
  margin-bottom: var(--spacing-md);
}

.button {
  padding: var(--spacing-sm) var(--spacing-lg);
  background: var(--color-primary);
  color: white;
  border-radius: var(--radius-md);
  transition: opacity 0.2s;

  &:hover {
    opacity: 0.9;
  }

  &:focus-visible {
    @include focus-ring;
  }
}
```

**Usage in Component:**
```tsx
import styles from './home.module.scss';

export function HomePage() {
  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Welcome</h1>
      <button className={styles.button}>Get Started</button>
    </div>
  );
}
```

### 11. Build Configuration

#### Vite Config (`vite.config.ts`)

```tsx
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        api: 'modern-compiler', // New Sass API
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
    },
  },
  build: {
    target: 'es2022',
    minify: 'esbuild',
    sourcemap: 'hidden',
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom', 'react-router'],
          'vendor-query': ['@tanstack/react-query'],
          'vendor-state': ['zustand'],
        },
      },
    },
  },
});
```

**Features:**
- Path aliases (`@/*`)
- Modern SCSS compiler
- Dev proxy to backend
- Vendor chunk splitting
- ES2022 target (modern browsers)

#### TypeScript Config (`tsconfig.app.json`)

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "jsx": "react-jsx",
    "strict": true,
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

### 12. Code Quality

#### Biome (`biome.json`)

**Fast, Rust-based linter and formatter:**

```json
{
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true,
      "correctness": {
        "useExhaustiveDependencies": "warn",
        "noUnusedVariables": "error"
      },
      "a11y": {
        "useButtonType": "error",
        "useValidAriaProps": "error"
      },
      "security": {
        "noDangerouslySetInnerHtml": "warn"
      }
    }
  },
  "formatter": {
    "enabled": true,
    "indentStyle": "space",
    "indentWidth": 2,
    "lineWidth": 100
  }
}
```

#### Stylelint (`stylelint.config.js`)

**SCSS linting:**
```js
export default {
  extends: [
    'stylelint-config-standard-scss',
    'stylelint-config-prettier-scss',
  ],
  rules: {
    'selector-class-pattern': '^[a-z][a-zA-Z0-9]*$', // camelCase
    'scss/at-import-partial-extension': null,
  },
};
```

## Development Workflow

**Start Development Server:**
```bash
cd frontend
pnpm install
pnpm dev
```

**Available at**: `http://localhost:5173`

**HMR**: Fast refresh on file changes

**Linting:**
```bash
pnpm lint        # Run Biome
pnpm lint:fix    # Auto-fix issues
pnpm lint:css    # SCSS linting
```

**Type Checking:**
```bash
pnpm type-check
```

**Build:**
```bash
pnpm build       # Production build
pnpm preview     # Preview production build
```

## Production Build

**Optimizations:**
- Tree shaking (removes unused code)
- Code splitting (lazy routes + vendor chunks)
- Minification (esbuild)
- Asset optimization (images, fonts)
- Hidden sourcemaps

**Build Output:**
```
dist/
├── assets/
│   ├── index-[hash].js      # Main bundle
│   ├── vendor-react-[hash].js
│   ├── vendor-query-[hash].js
│   └── [route]-[hash].js    # Lazy route chunks
├── index.html
└── assets/icons/
```

**Deployment:**
- Nginx serves static files
- API proxied to backend
- Aggressive caching for assets (1 year)
- No cache for index.html (SPA routing)

## Environment Variables

**Development (`.env`):**
```bash
VITE_API_URL=/api
VITE_APP_TITLE=My App
```

**Production:**
- Set via Docker build args
- Example: `--build-arg VITE_API_URL=https://api.example.com`

**Usage in Code:**
```tsx
const apiUrl = import.meta.env.VITE_API_URL;
const isDev = import.meta.env.DEV;
const isProd = import.meta.env.PROD;
```

## Performance Optimizations

1. **Route-based Code Splitting** - Lazy load pages
2. **Vendor Chunk Splitting** - Separate React, Query, Zustand
3. **Stale-While-Revalidate** - TanStack Query caching
4. **Optimistic Updates** - Instant UI feedback
5. **Memoization** - Granular Zustand selectors
6. **Image Optimization** - Lazy loading, WebP format
7. **CSS Modules** - Scoped styles, no conflicts
8. **Modern Bundler** - Rolldown (faster than Rollup)

## Security Best Practices

1. **No Token in localStorage** - Stored in memory (Zustand, not persisted)
2. **HttpOnly Cookies** - Refresh token protected from XSS
3. **CORS Validation** - Backend validates origins
4. **Content Security Policy** - Meta tag in HTML
5. **Type Validation** - Zod schemas for API responses
6. **Error Handling** - Never expose sensitive info
7. **Secure Dependencies** - Regular audits (`pnpm audit`)

## Common Patterns

**Adding a New Page:**

1. Create page component in `src/pages/[name]/index.tsx`
2. Add route in `src/core/app/routers.tsx` with lazy loading
3. Add to navigation menu
4. Create styles in `[name].module.scss`

**Adding an API Endpoint:**

1. Add endpoint to `API_ENDPOINTS` in `src/config.ts`
2. Add query key to `QUERY_KEYS` in `src/config.ts`
3. Create Zod schema in `src/api/types/`
4. Create hook in `src/api/hooks/`
5. Use hook in component

**Creating a Protected Route:**

```tsx
{
  path: '/admin',
  element: <ProtectedRoute allowedRoles={[UserRole.ADMIN]} />,
  lazy: () => import('@/pages/admin'),
}
```

**Optimistic Updates:**

```tsx
const updateMutation = useMutation({
  mutationFn: updateUser,
  onMutate: async (newData) => {
    await queryClient.cancelQueries({ queryKey: QUERY_KEYS.users.me() });
    const previous = queryClient.getQueryData(QUERY_KEYS.users.me());
    queryClient.setQueryData(QUERY_KEYS.users.me(), newData);
    return { previous };
  },
  onError: (err, newData, context) => {
    queryClient.setQueryData(QUERY_KEYS.users.me(), context?.previous);
  },
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: QUERY_KEYS.users.me() });
  },
});
```

## Troubleshooting

**CORS errors:**
- Check `CORS_ORIGINS` on backend
- Verify `withCredentials: true` in Axios
- Ensure cookies sent with requests

**Token refresh loop:**
- Check `_retry` flag in interceptor
- Verify refresh endpoint returns new token
- Clear localStorage and cookies

**Styles not applying:**
- Check CSS Module import syntax
- Verify SCSS syntax (use modern compiler)
- Inspect element to see applied classes

**Type errors:**
- Run `pnpm type-check`
- Check Zod schema matches API response
- Verify TypeScript version compatibility

## File Paths Reference

| Component | Path |
|-----------|------|
| Entry Point | `src/main.tsx` |
| App Root | `src/App.tsx` |
| Config | `src/config.ts` |
| Routers | `src/core/app/routers.tsx` |
| Shell | `src/core/app/shell.tsx` |
| Protected Route | `src/core/app/protected-route.tsx` |
| Auth Store | `src/core/lib/auth.store.ts` |
| UI Store | `src/core/lib/ui.store.ts` |
| API Client | `src/core/api/api.config.ts` |
| Query Client | `src/core/api/query.config.ts` |
| Error Handling | `src/core/api/errors.ts` |
| Auth Hooks | `src/api/hooks/useAuth.ts` |
| User Hooks | `src/api/hooks/useUsers.ts` |
| Admin Hooks | `src/api/hooks/useAdmin.ts` |
| Auth Types | `src/api/types/auth.types.ts` |
| User Types | `src/api/types/user.types.ts` |
| Design Tokens | `src/styles/_tokens.scss` |
| Mixins | `src/styles/_mixins.scss` |
| Global Styles | `src/styles.scss` |

---

**This frontend is production-ready and follows 2025 best practices for React applications with TypeScript, featuring modern tooling, robust state management, and a comprehensive design system.**
