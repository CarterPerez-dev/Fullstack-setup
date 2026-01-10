/**
 * ©AngelaMos | 2025
 * index.tsx
 */

import { useUser } from '@/core/lib'
import styles from './dashboard.module.scss'

const AVAILABLE_STORES = [
  {
    name: 'useUser()',
    file: 'core/lib/auth.store.ts',
    description: 'Get current authenticated user',
  },
  {
    name: 'useIsAuthenticated()',
    file: 'core/lib/auth.store.ts',
    description: 'Check if user is logged in',
  },
  {
    name: 'useIsAdmin()',
    file: 'core/lib/auth.store.ts',
    description: 'Check if user has admin role',
  },
]

const SUGGESTED_FEATURES = [
  'User stats and metrics',
  'Recent activity feed',
  'Quick actions',
  'Charts and analytics',
  'Notifications overview',
  'Task/project summary',
]

export function Component(): React.ReactElement {
  const user = useUser()

  return (
    <div className={styles.page}>
      <div className={styles.container}>
        <div className={styles.header}>
          <h1 className={styles.title}>
            Welcome{user?.full_name ? `, ${user.full_name}` : ''}
          </h1>
          <p className={styles.subtitle}>
            Template page — build your dashboard here
          </p>
        </div>

        <div className={styles.userCard}>
          <div className={styles.avatar}>
            {user?.full_name?.[0]?.toUpperCase() ??
              user?.email?.[0]?.toUpperCase() ??
              'U'}
          </div>
          <div className={styles.userInfo}>
            <span className={styles.userName}>{user?.full_name ?? 'User'}</span>
            <span className={styles.userEmail}>{user?.email}</span>
            <span className={styles.userRole}>{user?.role}</span>
          </div>
        </div>

        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Available Stores</h2>
          <div className={styles.grid}>
            {AVAILABLE_STORES.map((store) => (
              <div key={store.name} className={styles.card}>
                <code className={styles.hookName}>{store.name}</code>
                <p className={styles.description}>{store.description}</p>
                <span className={styles.file}>{store.file}</span>
              </div>
            ))}
          </div>
        </section>

        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Suggested Features</h2>
          <ul className={styles.list}>
            {SUGGESTED_FEATURES.map((feature) => (
              <li key={feature}>{feature}</li>
            ))}
          </ul>
        </section>
      </div>
    </div>
  )
}

Component.displayName = 'Dashboard'
