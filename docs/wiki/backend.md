# Backend Architecture Documentation

**FastAPI Template | 2025**

## Overview

The `/backend` FastAPI application built with modern Python patterns and 2025 best practices. It features async throughout, clean architecture, comprehensive security, and enterprise level observability.

## Tech Stack

- **Framework**: FastAPI (async)
- **Python**: 3.12+
- **Database**: PostgreSQL 16+ with asyncpg driver
- **ORM**: SQLAlchemy 2.0+ (async, typed syntax)
- **Migrations**: Alembic (async)
- **Authentication**: JWT (PyJWT + Argon2id password hashing)
- **Validation**: Pydantic v2
- **Server**: Uvicorn (dev), Gunicorn + Uvicorn workers (prod)
- **Caching**: Redis (rate limiting, optional caching)
- **Package Manager**: uv (Astral)
- **Logging**: Structlog (JSON in production)
- **Testing**: Pytest (async)

## Architecture

### Clean Architecture Pattern

The backend follows a layered architecture with clear separation of concerns:

```
Routes (API Layer)
    ↓
Services (Business Logic)
    ↓
Repositories (Data Access)
    ↓
Models (Database Entities)
```

**Key Principles:**
- Dependencies point inward
- Business logic isolated in services
- Data access abstracted in repositories
- Type safety throughout with Mypy strict mode

### Project Structure

```
backend/
├── src/
│   ├── core/              # Framework & infrastructure
│   │   ├── api/           # API configuration
│   │   ├── dependencies.py # FastAPI dependencies
│   │   ├── database.py    # DB session management
│   │   ├── security.py    # Auth & crypto
│   │   ├── exceptions.py  # Custom exceptions
│   │   ├── logging.py     # Structured logging
│   │   └── rate_limit.py  # Rate limiting
│   ├── models/            # SQLAlchemy ORM models
│   │   ├── Base.py        # Base mixins
│   │   ├── User.py        # User model
│   │   └── RefreshToken.py # Token model
│   ├── repositories/      # Data access layer
│   │   ├── base.py        # Generic repository
│   │   ├── user.py        # User repository
│   │   └── refresh_token.py # Token repository
│   ├── services/          # Business logic layer
│   │   ├── auth.py        # Authentication service
│   │   └── user.py        # User service
│   ├── routes/            # FastAPI route handlers
│   │   ├── health.py      # Health checks
│   │   ├── auth.py        # Auth endpoints
│   │   ├── user.py        # User endpoints
│   │   └── admin.py       # Admin endpoints
│   ├── schemas/           # Pydantic request/response
│   │   ├── base.py        # Base schemas
│   │   ├── auth.py        # Auth schemas
│   │   └── user.py        # User schemas
│   ├── middleware/        # Custom middleware
│   │   └── correlation.py # Request tracing
│   ├── config.py          # Configuration management
│   ├── factory.py         # Application factory
│   └── main.py            # Entry point
├── alembic/              # Database migrations
│   ├── env.py            # Migration config
│   └── versions/         # Migration files
├── tests/                # Test suite
│   └── [test files]
├── conftest.py           # Pytest fixtures
├── pyproject.toml        # Dependencies & tooling
└── alembic.ini          # Alembic config
```

## Core Components

### 1. Application Factory (`src/factory.py`)

**Pattern**: Factory pattern with lifespan management

```python
async def lifespan(app: FastAPI):
    # Startup
    await DatabaseSessionManager().init()
    setup_logging()

    yield

    # Shutdown
    await DatabaseSessionManager().close()
```

**Features:**
- Conditional OpenAPI (disabled in production)
- CORS middleware with environment-specific rules
- Rate limiting (Redis or in-memory fallback)
- Correlation ID middleware for distributed tracing
- Centralized exception handling
- API versioning (`/v1` prefix)

**Middleware Stack:**
1. CorrelationIdMiddleware (request tracking)
2. CORSMiddleware (cross-origin)
3. SlowAPI rate limiter

### 2. Configuration (`src/config.py`)

**Pattern**: Pydantic Settings with validation

```python
class Settings(BaseSettings):
    # Environment
    ENVIRONMENT: Literal["development", "production"]
    DEBUG: bool = False

    # Database
    POSTGRES_USER: str
    POSTGRES_PASSWORD: SecretStr
    DATABASE_URL: str
    DB_POOL_SIZE: int = 10
    DB_MAX_OVERFLOW: int = 20

    # JWT
    JWT_SECRET_KEY: SecretStr
    JWT_ALGORITHM: str = "HS256"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 15
    REFRESH_TOKEN_EXPIRE_DAYS: int = 7

    # Security
    CORS_ORIGINS: list[str]
    ALLOWED_HOSTS: list[str]
```

**Production Validators:**
- Ensures `DEBUG=False` in production
- Prevents wildcard CORS in production
- Validates JWT algorithm selection
- Type-safe with `SecretStr` for sensitive data

**Environment Variables**: All settings loaded from `.env` file

### 3. Database Layer (`src/core/database.py`)

**Pattern**: Singleton session manager with dual engines

```python
class DatabaseSessionManager:
    _instance = None
    _async_engine: AsyncEngine
    _sync_engine: Engine
```

**Async Engine (Runtime):**
- Driver: `asyncpg`
- Connection pooling with health checks (`pool_pre_ping`)
- Configurable pool size, overflow, timeout
- Fast connection recycling (1 hour)

**Sync Engine (Migrations):**
- Driver: `psycopg2`
- Used by Alembic CLI tools
- Same pooling strategy

**Session Management:**
```python
async with session_manager.session() as session:
    # Auto-commit on success
    # Auto-rollback on exception
    result = await session.execute(query)
```

**Key Settings:**
- `expire_on_commit=False` - Prevents lazy loading issues
- `echo=True` in debug mode - SQL query logging

### 4. Security (`src/core/security.py`)

#### Password Hashing

**Algorithm**: Argon2id (OWASP recommended)
- Memory-hard (resistant to GPU cracking)
- Configurable cost parameters
- Async hashing (thread pool, non-blocking)
- Automatic rehashing on login (if parameters outdated)
- Timing attack prevention (dummy hash on non-existent users)

```python
async def hash_password(password: str) -> str:
    return await asyncio.to_thread(pwd_context.hash, password)

async def verify_password(plain: str, hashed: str) -> bool:
    return await asyncio.to_thread(pwd_context.verify, plain, hashed)
```

#### JWT Token System

**Dual Token Strategy:**

1. **Access Tokens** (Short-lived: 15 minutes)
   - Stored client-side (memory, not localStorage)
   - Contains: `sub` (user_id), `exp`, `token_version`
   - Invalidated via `token_version` field on user

2. **Refresh Tokens** (Long-lived: 7 days)
   - Stored in HttpOnly cookies (XSS protection)
   - SHA-256 hashed in database (never raw)
   - Contains: `family_id` (replay detection)
   - Revocable per token or per user

**Token Rotation:**
- Each refresh generates new access + refresh token
- Old refresh token revoked immediately
- Family ID tracks token lineage

**Replay Attack Protection:**
```python
# If revoked token reused → revoke entire family
if token.is_revoked:
    await repo.revoke_family(token.family_id)
    raise TokenRevokedError()
```

**Cookie Security:**
```python
response.set_cookie(
    key="refresh_token",
    value=token,
    httponly=True,           # XSS protection
    secure=settings.SECURE_COOKIES,  # HTTPS only in prod
    samesite="strict",       # CSRF protection
    max_age=604800,          # 7 days
    path="/v1/auth"          # Scoped to auth routes
)
```

**Global Token Revocation:**
```python
# Increment token_version → invalidates all access tokens
user.token_version += 1
await session.commit()
```

### 5. Database Models

#### Base Mixins (`src/models/Base.py`)

**UUIDMixin:**
```python
id: Mapped[uuid.UUID] = mapped_column(
    UUID(as_uuid=True),
    primary_key=True,
    default=uuid7,  # Time-sortable, distributed-friendly
)
```

**TimestampMixin:**
```python
created_at: Mapped[datetime] = mapped_column(
    DateTime(timezone=True),
    server_default=func.now()
)
updated_at: Mapped[datetime] = mapped_column(
    DateTime(timezone=True),
    server_default=func.now(),
    onupdate=func.now()
)
```

**SoftDeleteMixin:**
```python
deleted_at: Mapped[datetime | None]
is_deleted: Mapped[bool] = mapped_column(default=False)
```

#### User Model (`src/models/User.py`)

**Fields:**
- `email` - Unique, indexed, case-insensitive
- `hashed_password` - Argon2id hash (max 1024 chars)
- `full_name` - Optional
- `is_active` - Account enabled/disabled
- `is_verified` - Email verification status
- `role` - Enum (USER, ADMIN, UNKNOWN)
- `token_version` - Global token revocation counter

**Relationships:**
```python
refresh_tokens: Mapped[list["RefreshToken"]] = relationship(
    back_populates="user",
    cascade="all, delete-orphan",  # Delete tokens with user
    lazy="raise"  # Prevent N+1, force explicit loading
)
```

**SafeEnum Pattern:**
```python
@staticmethod
def _enum_value(value: Any, enum_class: type[Enum]) -> Any:
    # Stores enum VALUES (not names) in DB
    # Prevents breakage on enum refactoring
    return value.value if isinstance(value, enum_class) else value
```

#### RefreshToken Model (`src/models/RefreshToken.py`)

**Fields:**
- `token_hash` - SHA-256 hash (never store raw tokens)
- `user_id` - Foreign key to User
- `family_id` - UUID for token family tracking
- `expires_at` - Indexed for efficient cleanup
- `is_revoked`, `revoked_at` - Soft revocation
- `device_id`, `device_name`, `ip_address` - Optional tracking

**Properties:**
```python
@property
def is_expired(self) -> bool:
    return datetime.now(timezone.utc) > self.expires_at

@property
def is_valid(self) -> bool:
    return not self.is_revoked and not self.is_expired
```

### 6. Repository Pattern (`src/repositories/`)

**BaseRepository (`base.py`):**

Generic repository with TypeVar for type safety:
```python
ModelType = TypeVar("ModelType")

class BaseRepository(Generic[ModelType]):
    def __init__(self, session: AsyncSession, model: type[ModelType]):
        self.session = session
        self.model = model

    async def get_by_id(self, id: UUID) -> ModelType | None
    async def get_multi(self, *, skip: int = 0, limit: int = 100)
    async def create(self, obj_in: dict[str, Any]) -> ModelType
    async def update(self, db_obj: ModelType, obj_in: dict)
    async def delete(self, id: UUID) -> None
```

**Pattern:**
- Uses `flush()` + `refresh()` instead of `commit()`
- Commit handled by session context manager
- Allows transaction grouping in services

**Specialized Repositories:**

**UserRepository (`user.py`):**
```python
async def get_by_email(self, email: str) -> User | None
async def email_exists(self, email: str) -> bool
async def update_password(self, user: User, new_hash: str)
async def increment_token_version(self, user: User)
```

**RefreshTokenRepository (`refresh_token.py`):**
```python
async def get_valid_by_hash(self, token_hash: str) -> RefreshToken | None
async def revoke_family(self, family_id: UUID)
async def revoke_all_user_tokens(self, user_id: UUID)
async def cleanup_expired(self) -> int
```

### 7. Service Layer (`src/services/`)

**AuthService (`auth.py`):**

Core authentication operations with business logic:

```python
async def authenticate(email: str, password: str) -> User:
    # 1. Fetch user
    # 2. Timing-safe password check (dummy hash if user not found)
    # 3. Auto-rehash if password params outdated
    # 4. Return user or raise InvalidCredentials
```

```python
async def login(email: str, password: str) -> tuple[str, User]:
    # 1. Authenticate user
    # 2. Generate access token + refresh token
    # 3. Store refresh token in DB (hashed)
    # 4. Return (access_token, user)
```

```python
async def refresh_tokens(refresh_token: str) -> tuple[str, str]:
    # 1. Hash and lookup token
    # 2. Validate (not revoked, not expired)
    # 3. Check user token_version
    # 4. Revoke old token
    # 5. Generate new access + refresh tokens
    # 6. Replay attack check (revoke family if old token reused)
    # 7. Return (new_access_token, new_refresh_token)
```

```python
async def logout(refresh_token: str):
    # Revoke single refresh token
```

```python
async def logout_all(user_id: UUID):
    # Increment token_version (invalidates all access tokens)
    # Revoke all refresh tokens in DB
```

**UserService (`user.py`):**

```python
async def create_user(user_in: UserCreate) -> User:
    # 1. Check email uniqueness
    # 2. Hash password
    # 3. Create user
```

```python
async def change_password(user: User, current: str, new: str):
    # 1. Verify current password
    # 2. Hash new password
    # 3. Update password
    # 4. Increment token_version (logout all devices)
```

```python
async def list_users(skip: int, limit: int) -> tuple[list[User], int]:
    # Paginated user list with total count
```

**Admin operations** - Bypass validation, full field control

### 8. API Routes

#### Health Endpoints (`src/routes/health.py`)

**`GET /health`** - Basic health check
```json
{"status": "healthy"}
```

**`GET /health/detailed`** - Component health
```json
{
  "status": "healthy",
  "database": "healthy",
  "redis": "healthy"
}
```

#### Auth Endpoints (`src/routes/auth.py`)

**`POST /v1/auth/login`** - OAuth2 password flow
- Rate limited: 1 req/s
- Body: `username` (email), `password` (form-data)
- Returns: Access token + user data
- Sets: `refresh_token` cookie (HttpOnly)

**`POST /v1/auth/refresh`** - Rotate tokens
- Reads: `refresh_token` cookie
- Returns: New access token
- Sets: New `refresh_token` cookie

**`POST /v1/auth/logout`** - Logout current device
- Requires: Auth
- Revokes: Current refresh token

**`POST /v1/auth/logout-all`** - Logout all devices
- Requires: Auth
- Increments: `token_version`
- Revokes: All refresh tokens

**`GET /v1/auth/me`** - Current user
- Requires: Auth
- Returns: User data

**`POST /v1/auth/change-password`** - Change password
- Requires: Auth
- Body: `current_password`, `new_password`
- Effect: Logs out all devices

#### User Endpoints (`src/routes/user.py`)

**`POST /v1/users`** - Register (public)
- Body: `email`, `password`, `full_name?`
- Returns: Created user

**`GET /v1/users/{user_id}`** - Get user
- Requires: Auth
- Returns: User data

**`PATCH /v1/users/me`** - Update own profile
- Requires: Auth
- Body: `full_name?`, `email?`
- Returns: Updated user

#### Admin Endpoints (`src/routes/admin.py`)

All routes require `ADMIN` role:

**`GET /v1/admin/users`** - List users (paginated)
- Query: `skip`, `limit`
- Returns: User list + total count

**`POST /v1/admin/users`** - Create user
- Body: Full user schema (including `role`, `is_active`)
- Returns: Created user

**`GET /v1/admin/users/{user_id}`** - Get user
**`PATCH /v1/admin/users/{user_id}`** - Update user
**`DELETE /v1/admin/users/{user_id}`** - Hard delete user

### 9. Schemas (Pydantic)

**Base Schemas (`src/schemas/base.py`):**
```python
class BaseSchema(BaseModel):
    model_config = ConfigDict(
        from_attributes=True,  # ORM mode
        str_strip_whitespace=True,
    )

class BaseResponseSchema(BaseSchema):
    id: UUID
    created_at: datetime
    updated_at: datetime
```

**User Schemas (`src/schemas/user.py`):**
```python
class UserCreate(BaseSchema):
    email: EmailStr
    password: str  # Min 8 chars, uppercase + digit validator
    full_name: str | None = None

class UserUpdate(BaseSchema):
    email: EmailStr | None = None
    full_name: str | None = None

class UserResponse(BaseResponseSchema):
    email: EmailStr
    full_name: str | None
    is_active: bool
    is_verified: bool
    role: UserRole

class UserListResponse(BaseSchema):
    items: list[UserResponse]
    total: int
```

### 10. Middleware & Utilities

**Correlation ID Middleware (`src/middleware/correlation.py`):**
- Extracts or generates `X-Correlation-ID` header
- Binds to structlog context (appears in all logs)
- Returns correlation ID in response headers
- Essential for distributed tracing

**Rate Limiting (`src/core/rate_limit.py`):**

**Strategy**: User ID (if authenticated) > IP address

**Storage**: Redis (preferred) with in-memory fallback

**Implementation**:
```python
@limiter.limit("10/minute", key_func=get_rate_limit_key)
async def protected_endpoint():
    # Rate limited per user or IP
```

**Headers Returned:**
- `X-RateLimit-Limit`
- `X-RateLimit-Remaining`
- `X-RateLimit-Reset`

**Structured Logging (`src/core/logging.py`):**

**Library**: Structlog with JSON output

**Processors:**
- ISO timestamps
- Log level formatting
- Exception stack traces
- Context variables (correlation IDs)
- Colored output in development

**Integration**: Captures Uvicorn/FastAPI logs

### 11. Exceptions (`src/core/exceptions.py`)

**Base Exception:**
```python
class BaseAppException(Exception):
    def __init__(
        self,
        status_code: int,
        message: str,
        extra: dict[str, Any] | None = None
    ):
        self.status_code = status_code
        self.message = message
        self.extra = extra or {}
```

**Hierarchy:**
```
BaseAppException
├── ResourceNotFound (404)
│   └── UserNotFound
├── ConflictError (409)
│   └── EmailAlreadyExists
├── AuthenticationError (401)
│   ├── InvalidCredentials
│   ├── TokenError
│   ├── TokenRevokedError
│   └── InactiveUser
├── PermissionDenied (403)
├── ValidationError (422)
└── RateLimitExceeded (429)
```

**Global Handler:**
```python
@app.exception_handler(BaseAppException)
async def handle_app_exception(request, exc):
    return JSONResponse(
        status_code=exc.status_code,
        content={
            "detail": exc.message,
            "type": exc.__class__.__name__
        }
    )
```

### 12. Dependencies (`src/core/dependencies.py`)

**Database Session:**
```python
async def get_session() -> AsyncGenerator[AsyncSession, None]:
    async with session_manager.session() as session:
        yield session

DBSession = Annotated[AsyncSession, Depends(get_session)]
```

**Authentication:**
```python
async def get_current_user(
    session: DBSession,
    token: str = Depends(oauth2_scheme)
) -> User:
    # 1. Decode JWT
    # 2. Extract user_id
    # 3. Fetch user from DB
    # 4. Verify token_version matches
    # 5. Return user or raise 401
```

**Type Aliases:**
```python
CurrentUser = Annotated[User, Depends(get_current_active_user)]
OptionalUser = Annotated[User | None, Depends(get_optional_user)]
ClientIP = Annotated[str, Depends(get_client_ip)]
```

**Authorization:**
```python
class RequireRole:
    def __init__(self, allowed_roles: list[UserRole]):
        self.allowed_roles = allowed_roles

    async def __call__(self, user: CurrentUser) -> User:
        if user.role not in self.allowed_roles:
            raise PermissionDenied()
        return user

# Usage
@router.get("/admin/users", dependencies=[Depends(RequireRole([UserRole.ADMIN]))])
```

### 13. Database Migrations (Alembic)

**Configuration (`alembic/env.py`):**

**Async Migrations:**
```python
def run_migrations_online():
    connectable = async_engine_from_config(
        config.get_section(config.config_ini_section),
        prefix="sqlalchemy.",
        poolclass=pool.NullPool,
    )
    # Run migrations in async context
```

**Custom Renderers:**
```python
# SafeEnum → sa.Enum with values
def render_safe_enum(type_, object_, autogen_context):
    return f"sa.Enum({', '.join(repr(v.value) for v in type_.enum_class)})"

# DateTime → Always timezone-aware
def render_datetime(type_, object_, autogen_context):
    return "sa.DateTime(timezone=True)"
```

**Auto-detection**: Imports all models from `Base.metadata`

**Creating Migrations:**
```bash
# Auto-generate migration
uv run alembic revision --autogenerate -m "description"

# Apply migrations
uv run alembic upgrade head

# Rollback
uv run alembic downgrade -1
```

### 14. Testing (`conftest.py`)

**Test Database:**
- SQLite in-memory (fast, isolated)
- StaticPool (shared connection across threads)

**Session Fixture:**
```python
@pytest.fixture
async def db_session():
    # Transaction with savepoint
    # Rollback after test (no pollution)
    async with engine.connect() as conn:
        await conn.begin()
        await conn.begin_nested()  # Savepoint

        session = AsyncSession(conn, join_transaction_mode="create_savepoint")
        yield session

        await session.close()
        await conn.rollback()
```

**Factories:**
```python
class UserFactory:
    _counter = 0

    @classmethod
    async def create(cls, session, **kwargs):
        cls._counter += 1
        user = User(
            email=f"user{cls._counter}@example.com",
            hashed_password=await hash_password("password"),
            **kwargs
        )
        session.add(user)
        await session.flush()
        return user
```

**Fixtures Provided:**
- `test_user`, `admin_user`, `inactive_user`
- `access_token`, `admin_access_token`
- `refresh_token_pair` (raw token + DB record)
- `client` (TestClient with DB override)

### 15. Development Tooling

**pyproject.toml** includes:

**Ruff** (Linting):
- Fast Python linter
- Rules: pycodestyle, pyflakes, bugbear, security
- Auto-fixes where safe

**Mypy** (Type Checking):
- Strict mode enabled
- Pydantic plugin
- SQLAlchemy plugin

**Pytest**:
- Async mode auto-enabled
- Coverage tracking
- Fixtures auto-discovery

**Pre-commit Hooks**:
- Ruff linting
- Mypy type checking
- Trailing whitespace removal

## API Documentation

**Swagger UI**: `/docs` (dev only)
**ReDoc**: `/redoc` (dev only)
**OpenAPI JSON**: `/openapi.json` (dev only)

In production, OpenAPI endpoints are disabled for security.

## Environment Variables

Required:
```bash
# Environment
ENVIRONMENT=development|production
DEBUG=true|false

# Database
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your-password
POSTGRES_DB=app_db
POSTGRES_HOST=db
POSTGRES_PORT=5432

# Security
JWT_SECRET_KEY=your-secret-key-here
CORS_ORIGINS=http://localhost:5173,http://localhost:3000

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=  # Optional
```

Optional (with defaults):
```bash
# JWT
JWT_ALGORITHM=HS256
ACCESS_TOKEN_EXPIRE_MINUTES=15
REFRESH_TOKEN_EXPIRE_DAYS=7

# Database Pool
DB_POOL_SIZE=10
DB_MAX_OVERFLOW=20
DB_POOL_TIMEOUT=30
DB_POOL_RECYCLE=3600

# Rate Limiting
RATE_LIMIT_ENABLED=true
```

## Running Locally

**Development:**
```bash
# Start services
docker compose -f dev.compose.yml up -d

# Run migrations
docker compose -f dev.compose.yml exec backend uv run alembic upgrade head

# View logs
docker compose -f dev.compose.yml logs -f backend

# Backend available at http://localhost:5420
# Swagger UI at http://localhost:5420/docs
```

**Without Docker:**
```bash
cd backend

# Install dependencies
uv sync

# Run migrations
uv run alembic upgrade head

# Start server (hot reload)
uv run uvicorn src.main:app --reload --port 8000
```

## Testing

```bash
# Run all tests
uv run pytest

# With coverage
uv run pytest --cov=src --cov-report=html

# Specific test
uv run pytest tests/test_auth.py -v

# Type checking
uv run mypy src

# Linting
uv run ruff check src
uv run ruff format src
```

## Production Deployment

**Build:**
```bash
docker build -f infra/docker/fastapi.prod -t backend:latest ./backend
```

**Run:**
```bash
docker compose -f compose.yml up -d
```

**Configuration:**
- 4 Gunicorn workers with Uvicorn workers
- Max 1000 requests per worker (memory leak protection)
- Health checks every 30s
- Runs as non-root user (`appuser`)
- Compiled Python bytecode (startup optimization)

## Security Considerations

1. **Passwords**: Argon2id with high cost parameters
2. **Tokens**: Short-lived access, long-lived refresh (HttpOnly)
3. **Token Rotation**: New refresh token on every refresh
4. **Replay Detection**: Family ID tracks token lineage
5. **CORS**: Strict origin validation in production
6. **Rate Limiting**: Per-user or IP-based
7. **SQL Injection**: SQLAlchemy parameterized queries
8. **XSS**: HttpOnly cookies, no raw token exposure
9. **CSRF**: SameSite=strict cookies
10. **Timing Attacks**: Constant-time password comparison

## Performance Optimizations

1. **Async Throughout**: Non-blocking I/O
2. **Connection Pooling**: Reuse DB connections
3. **Lazy Loading Prevention**: `lazy="raise"` on relationships
4. **Indexed Queries**: Indexes on email, expires_at, etc.
5. **Password Hashing**: Thread pool (doesn't block event loop)
6. **Redis Caching**: Fast rate limit lookups
7. **Connection Keep-Alive**: HTTP/1.1 persistent connections
8. **Batch Operations**: Repositories support bulk operations

## Common Patterns

**Adding a New Endpoint:**

1. Create schema in `src/schemas/`
2. Add repository method in `src/repositories/`
3. Add service method in `src/services/`
4. Add route in `src/routes/`
5. Register router in `src/factory.py`

**Adding a New Model:**

1. Create model in `src/models/` (inherit from `Base`)
2. Add repository in `src/repositories/`
3. Import in `alembic/env.py` for auto-detection
4. Generate migration: `alembic revision --autogenerate`
5. Review and apply: `alembic upgrade head`

**Custom Authentication:**

Extend `get_current_user` dependency or create new dependency with different logic.

**Background Tasks:**

Use FastAPI's `BackgroundTasks` or integrate Celery for heavy workloads.

## Troubleshooting

**Database connection errors:**
- Check `DATABASE_URL` format
- Ensure PostgreSQL is running
- Verify network connectivity (Docker networks)

**JWT errors:**
- Verify `JWT_SECRET_KEY` is set and consistent
- Check token expiration times
- Ensure `token_version` hasn't changed

**Migration errors:**
- Check `alembic/env.py` imports all models
- Review generated migration before applying
- Use `alembic downgrade` to rollback

**Rate limit errors:**
- Check Redis connectivity
- Verify rate limit configuration
- Falls back to in-memory if Redis unavailable

## File Paths Reference

| Component | Path |
|-----------|------|
| Entry Point | `src/__main__.py` |
| App Factory | `src/factory.py` |
| Config | `src/config.py` |
| Database | `src/core/database.py` |
| Security | `src/core/security.py` |
| Models | `src/models/` |
| Repositories | `src/repositories/` |
| Services | `src/services/` |
| Routes | `src/routes/` |
| Schemas | `src/schemas/` |
| Dependencies | `src/core/dependencies.py` |
| Exceptions | `src/core/exceptions.py` |
| Middleware | `src/middleware/` |
| Migrations | `alembic/` |
| Tests | `tests/` |
| Test Config | `conftest.py` |

---

**This backend is production-ready and follows 2025 best practices for scalability, security, and maintainability.**
