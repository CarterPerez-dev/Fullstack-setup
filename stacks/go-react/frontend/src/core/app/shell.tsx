/**
 * ©AngelaMos | 2025
 * shell.tsx
 */

import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { GiCardAceClubs, GiCardJoker, GiExitDoor } from 'react-icons/gi'
import { LuChevronLeft, LuChevronRight, LuMenu, LuShield } from 'react-icons/lu'
import { Link, NavLink, Outlet, useLocation } from 'react-router-dom'
import { useLogout } from '@/api/hooks'
import { ROUTES } from '@/config'
import { useIsAdmin, useUIStore, useUser } from '@/core/lib'
import styles from './shell.module.scss'

const NAV_ITEMS = [
  { path: ROUTES.DASHBOARD, label: 'Dashboard', icon: GiCardJoker },
  { path: ROUTES.SETTINGS, label: 'Settings', icon: GiCardAceClubs },
]

const ADMIN_NAV_ITEM = {
  path: ROUTES.ADMIN.USERS,
  label: 'Admin',
  icon: LuShield,
}

function ShellErrorFallback({ error }: { error: Error }): React.ReactElement {
  return (
    <div className={styles.error}>
      <h2>Something went wrong</h2>
      <pre>{error.message}</pre>
    </div>
  )
}

function ShellLoading(): React.ReactElement {
  return <div className={styles.loading}>Loading...</div>
}

function getPageTitle(pathname: string, isAdmin: boolean): string {
  if (isAdmin && pathname === ADMIN_NAV_ITEM.path) {
    return ADMIN_NAV_ITEM.label
  }
  const item = NAV_ITEMS.find((i) => i.path === pathname)
  return item?.label ?? 'Dashboard'
}

export function Shell(): React.ReactElement {
  const location = useLocation()
  const { sidebarOpen, sidebarCollapsed, toggleSidebar, toggleSidebarCollapsed } =
    useUIStore()
  const { mutate: logout } = useLogout()
  const isAdmin = useIsAdmin()
  const user = useUser()

  const pageTitle = getPageTitle(location.pathname, isAdmin)
  const avatarLetter =
    user?.full_name?.[0]?.toUpperCase() ?? user?.email?.[0]?.toUpperCase() ?? 'U'

  return (
    <div className={styles.shell}>
      <aside
        className={`${styles.sidebar} ${sidebarOpen ? styles.open : ''} ${sidebarCollapsed ? styles.collapsed : ''}`}
      >
        <div className={styles.sidebarHeader}>
          <span className={styles.logo}>NavBar Template</span>
          <button
            type="button"
            className={styles.collapseBtn}
            onClick={toggleSidebarCollapsed}
            aria-label={sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          >
            {sidebarCollapsed ? <LuChevronRight /> : <LuChevronLeft />}
          </button>
        </div>

        <nav className={styles.nav}>
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              className={({ isActive }) =>
                `${styles.navItem} ${isActive ? styles.active : ''}`
              }
              onClick={() => sidebarOpen && toggleSidebar()}
            >
              <item.icon className={styles.navIcon} />
              <span className={styles.navLabel}>{item.label}</span>
            </NavLink>
          ))}
          {isAdmin && (
            <NavLink
              to={ADMIN_NAV_ITEM.path}
              className={({ isActive }) =>
                `${styles.navItem} ${styles.adminItem} ${isActive ? styles.active : ''}`
              }
              onClick={() => sidebarOpen && toggleSidebar()}
            >
              <ADMIN_NAV_ITEM.icon className={styles.navIcon} />
              <span className={styles.navLabel}>{ADMIN_NAV_ITEM.label}</span>
            </NavLink>
          )}
        </nav>

        <div className={styles.sidebarFooter}>
          <button
            type="button"
            className={styles.logoutBtn}
            onClick={() => logout()}
          >
            <GiExitDoor className={styles.logoutIcon} />
            <span className={styles.logoutText}>Logout</span>
          </button>
        </div>
      </aside>

      {sidebarOpen && (
        <button
          type="button"
          className={styles.overlay}
          onClick={toggleSidebar}
          onKeyDown={(e) => e.key === 'Escape' && toggleSidebar()}
          aria-label="Close sidebar"
        />
      )}

      <div
        className={`${styles.main} ${sidebarCollapsed ? styles.collapsed : ''}`}
      >
        <header className={styles.header}>
          <div className={styles.headerLeft}>
            <button
              type="button"
              className={styles.menuBtn}
              onClick={toggleSidebar}
              aria-label="Toggle menu"
            >
              <LuMenu />
            </button>
            <h1 className={styles.pageTitle}>{pageTitle}</h1>
          </div>

          <div className={styles.headerRight}>
            <Link to={ROUTES.SETTINGS} className={styles.avatar}>
              {avatarLetter}
            </Link>
          </div>
        </header>

        <main className={styles.content}>
          <ErrorBoundary FallbackComponent={ShellErrorFallback}>
            <Suspense fallback={<ShellLoading />}>
              <Outlet />
            </Suspense>
          </ErrorBoundary>
        </main>
      </div>
    </div>
  )
}
