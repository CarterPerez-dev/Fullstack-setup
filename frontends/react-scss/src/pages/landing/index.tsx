/**
 * ©AngelaMos | 2026
 * index.tsx
 */

import { FiGithub } from 'react-icons/fi'
import { Link } from 'react-router-dom'
import { ROUTES } from '@/config'
import styles from './landing.module.scss'

export function Component(): React.ReactElement {
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <h1 className={styles.title}>Full Stack Template</h1>
        <p className={styles.subtitle}>by Carter Perez</p>
        <a
          href="https://github.com/CarterPerez-dev/fullstack-template"
          target="_blank"
          rel="noopener noreferrer"
          className={styles.github}
          aria-label="View source on GitHub"
        >
          <FiGithub />
        </a>
      </header>

      <div className={styles.content}>
        <p className={styles.description}>
          Boilerplate for medium-large scale applications. Built with modern
          patterns, strict typing, and security best practices.
        </p>

        <div className={styles.sections}>
          <section className={styles.section}>
            <h2 className={styles.sectionTitle}>Frontend</h2>
            <ul className={styles.features}>
              <li>React 19 + TypeScript with strict mode</li>
              <li>TanStack Query for server state caching</li>
              <li>Zustand stores with persistence</li>
              <li>Axios interceptors with auto token refresh</li>
              <li>Zod runtime validation on API responses</li>
              <li>SCSS modules with design tokens</li>
            </ul>
          </section>

          <section className={styles.section}>
            <h2 className={styles.sectionTitle}>Backend</h2>
            <ul className={styles.features}>
              <li>DDD + DI Architecture</li>
              <li>FastAPI with async/await throughout</li>
              <li>SQLAlchemy 2.0+ async with connection pooling</li>
              <li>JWT auth with token rotation and replay detection</li>
              <li>Argon2id hashing with timing safe verification</li>
              <li>Pydantic v2 strict validation</li>
            </ul>
          </section>

          <section className={styles.section}>
            <h2 className={styles.sectionTitle}>Infrastructure</h2>
            <ul className={styles.features}>
              <li>Docker Compose with dev/prod configs</li>
              <li>Nginx reverse proxy with rate limiting</li>
              <li>PostgreSQL 18 + Redis 7</li>
              <li>Health checks and graceful shutdown</li>
            </ul>
          </section>

          <section className={styles.section}>
            <h2 className={styles.sectionTitle}>DevOps</h2>
            <ul className={styles.features}>
              <li>GitHub Actions CI (Ruff, Pylint, Mypy, Biome)</li>
              <li>Strict type checking on both ends</li>
              <li>Alembic async migrations</li>
            </ul>
          </section>
        </div>

        <div className={styles.actions}>
          <Link to={ROUTES.REGISTER} className={styles.button}>
            Try Auth Flow
          </Link>
          <a
            href="/api/docs"
            target="_blank"
            rel="noopener noreferrer"
            className={styles.buttonOutline}
          >
            API Docs
          </a>
        </div>
      </div>
    </div>
  )
}

Component.displayName = 'Landing'
