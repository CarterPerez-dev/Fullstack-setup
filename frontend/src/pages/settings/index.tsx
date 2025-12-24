/**
 * ©AngelaMos | 2025
 * index.tsx
 */

import styles from './settings.module.scss'

const AVAILABLE_HOOKS = [
  {
    name: 'useUpdateProfile()',
    file: 'api/hooks/useUsers.ts',
    description: 'Update user profile (full_name)',
    endpoint: 'PATCH /api/v1/users/me',
  },
  {
    name: 'useChangePassword()',
    file: 'api/hooks/useAuth.ts',
    description: 'Change password (current + new)',
    endpoint: 'POST /api/v1/auth/change-password',
  },
]

const AVAILABLE_STORES = [
  {
    name: 'useAuthStore()',
    file: 'core/lib/auth.store.ts',
    description: 'Access user state, logout, updateUser',
  },
  {
    name: 'useUser()',
    file: 'core/lib/auth.store.ts',
    description: 'Get current user from store',
  },
]

export function Component(): React.ReactElement {
  return (
    <div className={styles.page}>
      <div className={styles.container}>
        <div className={styles.header}>
          <h1 className={styles.title}>Settings</h1>
          <p className={styles.subtitle}>
            Template page — available hooks and stores for building your settings
            UI
          </p>
        </div>

        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Available Hooks</h2>
          <div className={styles.grid}>
            {AVAILABLE_HOOKS.map((hook) => (
              <div key={hook.name} className={styles.card}>
                <code className={styles.hookName}>{hook.name}</code>
                <p className={styles.description}>{hook.description}</p>
                <div className={styles.meta}>
                  <span className={styles.file}>{hook.file}</span>
                  <span className={styles.endpoint}>{hook.endpoint}</span>
                </div>
              </div>
            ))}
          </div>
        </section>

        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Available Stores</h2>
          <div className={styles.grid}>
            {AVAILABLE_STORES.map((store) => (
              <div key={store.name} className={styles.card}>
                <code className={styles.hookName}>{store.name}</code>
                <p className={styles.description}>{store.description}</p>
                <div className={styles.meta}>
                  <span className={styles.file}>{store.file}</span>
                </div>
              </div>
            ))}
          </div>
        </section>

        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Suggested Features</h2>
          <ul className={styles.list}>
            <li>Profile form (full name, avatar)</li>
            <li>Change password form</li>
            <li>Email preferences</li>
            <li>Theme toggle (dark/light)</li>
            <li>Notification settings</li>
            <li>Delete account</li>
          </ul>
        </section>
      </div>
    </div>
  )
}

Component.displayName = 'Settings'
