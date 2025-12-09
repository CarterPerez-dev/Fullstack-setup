# Nginx Configuration Documentation

**Reverse Proxy & Static File Server | 2025**

## Overview

This template uses Nginx as a high-performance reverse proxy and static file server. It handles routing between the frontend and backend, implements security headers, rate limiting, and optimizes asset delivery with caching strategies.

## Architecture

### Dual Configuration Strategy

**Development** (`infra/nginx/dev.nginx`):
- Proxies to Vite dev server (HMR support)
- Proxies API requests to FastAPI backend
- WebSocket support for Vite HMR
- No caching (for development iteration)

**Production** (`infra/nginx/prod.nginx`):
- Serves static frontend assets (built React app)
- Proxies API requests to FastAPI backend
- Aggressive caching for static assets
- Security headers
- Performance optimizations

### Request Flow

```
Client Request
    ↓
Nginx (Port 80)
    ↓
    ├─→ /api/*        → Backend (FastAPI:8000)
    ├─→ /api/ws/*     → Backend WebSocket
    ├─→ /health       → Nginx health check
    └─→ /*            → Frontend (Vite dev OR static files)
```

## Configuration Files

### 1. Base Configuration (`infra/nginx/nginx.conf`)

**Shared across dev and production environments**

#### Process Management

```nginx
user nginx;
worker_processes auto;           # One per CPU core
worker_rlimit_nofile 65535;      # Max open files per worker

error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;
```

**Why `worker_processes auto`?**
- Automatically detects CPU cores
- Optimal for horizontal scaling
- No manual configuration needed

#### Event Configuration

```nginx
events {
    worker_connections 4096;      # Connections per worker
    multi_accept on;              # Accept multiple connections at once
    use epoll;                    # Linux-specific, efficient event model
}
```

**Capacity Calculation:**
- Max connections = `worker_processes` × `worker_connections`
- Example: 4 cores × 4096 = 16,384 concurrent connections

#### HTTP Configuration

**MIME Types:**
```nginx
include /etc/nginx/mime.types;
default_type application/octet-stream;
```

**WebSocket Support:**
```nginx
map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}
```
- Detects WebSocket upgrade requests
- Used for Vite HMR and API WebSockets

#### Upstream Definitions

**Backend (FastAPI):**
```nginx
upstream backend {
    server backend:8000 max_fails=3 fail_timeout=30s;
    keepalive 32;                 # Connection pool
    keepalive_requests 1000;      # Requests per keepalive connection
    keepalive_timeout 60s;        # Keep connection alive for 60s
}
```

**Frontend Dev (Vite):**
```nginx
upstream frontend_dev {
    server frontend:5173;
    keepalive 8;
}
```

**Benefits of Keepalive:**
- Reuses TCP connections
- Reduces handshake overhead
- Improves backend throughput by ~50%

#### Rate Limiting

```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=1r/s;
limit_conn_zone $binary_remote_addr zone=conn_limit:10m;
limit_req_status 429;
```

**Zones:**
- `api_limit`: 10 requests/second per IP (general API)
- `auth_limit`: 1 request/second per IP (login protection)
- `conn_limit`: Max concurrent connections per IP

**Memory:**
- `10m` = 10 MB of shared memory
- Stores ~160,000 IP addresses

#### Logging Format

```nginx
log_format main_timed '$remote_addr - $remote_user [$time_local] '
                      '"$request" $status $body_bytes_sent '
                      '"$http_referer" "$http_user_agent" '
                      'rt=$request_time uct="$upstream_connect_time" '
                      'uht="$upstream_header_time" urt="$upstream_response_time"';
```

**Key Metrics:**
- `rt` = Total request time
- `uct` = Time to connect to upstream
- `uht` = Time to receive first byte from upstream
- `urt` = Time to receive full response

#### Performance Tuning

**Sendfile & TCP Optimization:**
```nginx
sendfile on;                    # Zero-copy file transfer
tcp_nopush on;                  # Send headers in one packet
tcp_nodelay on;                 # Disable Nagle's algorithm
keepalive_timeout 65;           # Client keepalive
types_hash_max_size 2048;
server_tokens off;              # Hide Nginx version (security)
```

**Buffer Sizes:**
```nginx
client_body_buffer_size 128k;
client_header_buffer_size 16k;
client_max_body_size 10m;       # Max upload size
large_client_header_buffers 4 16k;
```

**Timeouts:**
```nginx
client_body_timeout 12s;        # Time to read client body
client_header_timeout 12s;      # Time to read client headers
send_timeout 10s;               # Response send timeout
```

**Why these timeouts?**
- Prevents slow client attacks
- Frees up resources quickly
- Balances UX and security

#### Gzip Compression

```nginx
gzip on;
gzip_vary on;                   # Add Vary: Accept-Encoding header
gzip_proxied any;               # Compress proxied responses
gzip_comp_level 6;              # Compression level (1-9, 6 is optimal)
gzip_min_length 256;            # Only compress files > 256 bytes
gzip_types
    text/plain
    text/css
    text/xml
    text/javascript
    application/json
    application/javascript
    application/xml+rss
    application/atom+xml
    image/svg+xml;
gzip_disable "msie6";           # Disable for old IE
```

**Compression Ratio:**
- JSON/CSS/JS: ~70-80% reduction
- Saves bandwidth
- Faster page loads

---

### 2. Development Configuration (`infra/nginx/dev.nginx`)

**Optimized for fast iteration and debugging**

#### Server Block

```nginx
server {
    listen 80;
    listen [::]:80;               # IPv6 support
    server_name _;                # Accept any hostname

    access_log /var/log/nginx/access.log main_timed;
    error_log /var/log/nginx/error.log debug;  # Verbose logging

    add_header Cache-Control "no-store, no-cache, must-revalidate" always;
}
```

**Cache-Control:**
- Disables all caching in development
- Ensures fresh content on every request

#### Health Check

```nginx
location /health {
    access_log off;               # Don't log health checks
    return 200 "healthy\n";
    add_header Content-Type text/plain;
}
```

**Usage:**
- Docker healthcheck
- Load balancer health probes
- Monitoring systems

#### API Proxy

```nginx
location /api/ {
    limit_req zone=api_limit burst=50 nodelay;
    limit_conn conn_limit 20;

    proxy_pass http://backend/;
    proxy_http_version 1.1;
    proxy_set_header Connection "";        # Enable keepalive
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    proxy_buffering off;                   # Stream responses (dev)
    proxy_request_buffering off;           # Stream requests (dev)

    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;
}
```

**Rate Limiting:**
- `burst=50`: Allow bursts up to 50 requests
- `nodelay`: Process burst immediately
- Prevents accidental API spam in development

**Buffering Off:**
- Real-time request/response streaming
- Better for debugging
- See logs immediately

#### WebSocket Proxy (API)

```nginx
location /api/ws/ {
    proxy_pass http://backend/ws/;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection $connection_upgrade;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    proxy_read_timeout 3600s;             # 1 hour (long-lived connections)
    proxy_send_timeout 3600s;
    proxy_buffering off;
}
```

**WebSocket Headers:**
- `Upgrade: websocket`
- `Connection: Upgrade`
- Required for WebSocket handshake

#### Vite Dev Server Proxy

```nginx
location / {
    proxy_pass http://frontend_dev;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection $connection_upgrade;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

    proxy_read_timeout 60s;
    proxy_buffering off;                  # Real-time HMR
}
```

**Why WebSocket for Vite?**
- Hot Module Replacement (HMR)
- Fast feedback loop
- No page refresh needed

---

### 3. Production Configuration (`infra/nginx/prod.nginx`)

**Optimized for performance, security, and caching**

#### Server Block

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name _;

    root /usr/share/nginx/html;           # Static files location
    index index.html;

    access_log /var/log/nginx/access.log main_timed buffer=32k flush=5s;
    error_log /var/log/nginx/error.log warn;
}
```

**Buffered Logging:**
- `buffer=32k`: Batch logs in memory
- `flush=5s`: Write to disk every 5 seconds
- Reduces I/O overhead

#### Security Headers

```nginx
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;
```

**Header Breakdown:**

1. **X-Frame-Options**: Prevents clickjacking
   - `SAMEORIGIN`: Only allow same-origin framing

2. **X-Content-Type-Options**: Prevents MIME sniffing
   - `nosniff`: Browser respects Content-Type header

3. **X-XSS-Protection**: XSS filter (legacy browsers)
   - `1; mode=block`: Block detected XSS attempts

4. **Referrer-Policy**: Controls referrer information
   - `strict-origin-when-cross-origin`: Send full URL for same-origin, origin only for cross-origin

5. **Permissions-Policy**: Restricts browser features
   - Disables geolocation, microphone, camera (adjust per needs)

#### API Proxy (Production)

```nginx
location /api/ {
    limit_req zone=api_limit burst=20 nodelay;
    limit_conn conn_limit 50;

    proxy_pass http://backend/;
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    proxy_buffering on;                   # Buffer responses
    proxy_buffers 8 32k;                  # 8 buffers of 32KB
    proxy_buffer_size 4k;                 # Initial buffer size

    proxy_connect_timeout 60s;
    proxy_send_timeout 30s;               # Tighter timeout
    proxy_read_timeout 30s;
}
```

**Production Differences:**
- Smaller burst (20 vs 50)
- More concurrent connections (50 vs 20)
- Buffering enabled (better performance)
- Tighter timeouts (faster failure detection)

#### Static Assets - Aggressive Caching

```nginx
location /assets/ {
    expires 1y;                           # Cache for 1 year
    add_header Cache-Control "public, immutable";
    access_log off;                       # Don't log static assets
    try_files $uri =404;
}

location ~* \.(jpg|jpeg|png|gif|ico|svg|webp|avif|woff|woff2|ttf|eot|otf)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
    access_log off;
}
```

**Why 1 year?**
- Vite uses content hashes in filenames (`index-abc123.js`)
- File changes = new filename = cache bypass
- `immutable`: Tells browser file never changes

**Cache Headers:**
- `public`: Can be cached by CDN
- `immutable`: File content never changes (safe to cache forever)

#### SPA Fallback

```nginx
location / {
    add_header Cache-Control "no-cache, must-revalidate";
    try_files $uri $uri/ /index.html;
}
```

**Why no cache for index.html?**
- Entry point must always be fresh
- Contains references to hashed assets
- Enables instant deploys

**try_files Directive:**
1. Try exact file match (`$uri`)
2. Try directory (`$uri/`)
3. Fallback to `index.html` (SPA routing)

#### Hidden Files Protection

```nginx
location ~ /\. {
    deny all;                             # Block .env, .git, etc.
    access_log off;
    log_not_found off;
}
```

**Security:**
- Prevents access to `.env`, `.git`, `.htaccess`
- Essential for protecting secrets

---

## Docker Integration

### Development Setup

**docker-compose.yml** (dev):
```yaml
nginx:
  image: nginx:1.27-alpine
  ports:
    - "8420:80"
  volumes:
    - ./infra/nginx/nginx.conf:/etc/nginx/nginx.conf:ro
    - ./infra/nginx/dev.nginx:/etc/nginx/conf.d/default.conf:ro
  depends_on:
    - backend
    - frontend
  networks:
    - frontend
    - backend
```

**Key Points:**
- Mounts config as read-only (`:ro`)
- Connects to both frontend and backend networks
- Depends on services being healthy

### Production Setup

**Dockerfile** (`infra/docker/frontend-builder.prod`):
```dockerfile
# Stage 1: Build frontend
FROM node:22-slim AS builder
WORKDIR /app
RUN pnpm build

# Stage 2: Nginx production
FROM nginx:1.27-alpine AS production
COPY --from=builder /app/dist /usr/share/nginx/html
COPY infra/nginx/nginx.conf /etc/nginx/nginx.conf
COPY infra/nginx/prod.nginx /etc/nginx/conf.d/default.conf
```

**Production Features:**
- Multi-stage build (smaller image)
- Static files baked into image
- No external volumes (immutable)

**Healthcheck:**
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80/health || exit 1
```

---

## Performance Optimizations

### 1. Connection Pooling

**Upstream Keepalive:**
```nginx
upstream backend {
    keepalive 32;
    keepalive_requests 1000;
}
```

**Benefit**: Reduces latency by ~10-30ms per request

### 2. Sendfile & Zero-Copy

```nginx
sendfile on;
tcp_nopush on;
```

**How it works:**
- `sendfile`: Kernel transfers file directly to socket (no user-space copy)
- `tcp_nopush`: Sends file + headers in one packet
- **Result**: ~40% faster static file delivery

### 3. Gzip Compression

```nginx
gzip on;
gzip_comp_level 6;
gzip_types text/css application/javascript;
```

**Savings:**
- CSS: ~75% size reduction
- JavaScript: ~70% size reduction
- JSON: ~80% size reduction

### 4. HTTP/1.1 Keepalive

```nginx
proxy_http_version 1.1;
proxy_set_header Connection "";
```

**Why HTTP/1.1?**
- Persistent connections to backend
- Reduces TCP handshakes
- Better throughput

### 5. Buffering Strategy

**Development**: Buffering off (real-time debugging)
**Production**: Buffering on (better performance)

```nginx
# Production
proxy_buffering on;
proxy_buffers 8 32k;
```

**Benefit**: Frees backend threads faster, handles slow clients

---

## Rate Limiting Details

### API Rate Limit

```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

location /api/ {
    limit_req zone=api_limit burst=50 nodelay;
}
```

**Parameters:**
- `rate=10r/s`: Steady-state rate
- `burst=50`: Allow bursts up to 50 requests
- `nodelay`: Process burst immediately (don't delay)

**Example:**
- Client sends 100 requests instantly
- First 50 processed immediately (burst)
- Next 10 processed over 1 second
- Remaining 40 get 429 (Too Many Requests)

### Auth Rate Limit

```nginx
limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=1r/s;
```

**Purpose**: Prevent brute-force login attacks

### Connection Limit

```nginx
limit_conn_zone $binary_remote_addr zone=conn_limit:10m;
limit_conn conn_limit 20;  # Dev
limit_conn conn_limit 50;  # Prod
```

**Purpose**: Prevent connection exhaustion attacks

---

## Security Best Practices

### 1. Hide Server Version

```nginx
server_tokens off;
```

**Before**: `Server: nginx/1.27.0`
**After**: `Server: nginx`

### 2. Security Headers

- **X-Frame-Options**: Clickjacking protection
- **X-Content-Type-Options**: MIME sniffing prevention
- **X-XSS-Protection**: XSS filter (legacy)
- **Referrer-Policy**: Referrer leakage control
- **Permissions-Policy**: Feature restrictions

### 3. Hidden Files Protection

```nginx
location ~ /\. {
    deny all;
}
```

### 4. Request Size Limits

```nginx
client_max_body_size 10m;
```

**Purpose**: Prevent large upload DoS attacks

### 5. Timeouts

```nginx
client_body_timeout 12s;
client_header_timeout 12s;
send_timeout 10s;
```

**Purpose**: Prevent slow-loris attacks

---

## Monitoring & Debugging

### Access Logs

**Location**: `/var/log/nginx/access.log`

**Format**: Custom `main_timed` with timing metrics

**Example Entry:**
```
192.168.1.1 - - [09/Dec/2025:10:30:45 +0000] "GET /api/users HTTP/1.1" 200 1234
"-" "Mozilla/5.0" rt=0.123 uct="0.005" uht="0.010" urt="0.108"
```

**Key Metrics:**
- `rt=0.123`: Total request time (123ms)
- `uct=0.005`: Connect time (5ms)
- `uht=0.010`: Time to first byte (10ms)
- `urt=0.108`: Backend response time (108ms)

### Error Logs

**Location**: `/var/log/nginx/error.log`

**Levels**: `debug`, `info`, `notice`, `warn`, `error`, `crit`, `alert`, `emerg`

**Common Errors:**
- `upstream timed out`: Backend too slow
- `connect() failed`: Backend not reachable
- `no live upstreams`: All backends failed health checks

### Health Check Endpoint

**URL**: `http://localhost/health`

**Response**: `200 OK` with body `healthy`

**Usage:**
```bash
# Docker healthcheck
wget --spider http://localhost/health

# Kubernetes liveness probe
httpGet:
  path: /health
  port: 80
```

---

## Troubleshooting

### Issue: 502 Bad Gateway

**Causes:**
1. Backend not running
2. Backend not listening on correct port
3. Network connectivity issue
4. Backend crashed

**Debug Steps:**
```bash
# Check backend is running
docker ps | grep backend

# Check backend logs
docker logs backend

# Test backend directly
curl http://backend:8000/health

# Check Nginx error log
docker logs nginx
```

### Issue: 504 Gateway Timeout

**Causes:**
1. Backend response too slow
2. Timeout settings too aggressive
3. Backend deadlock

**Solutions:**
```nginx
# Increase timeouts
proxy_connect_timeout 120s;
proxy_read_timeout 120s;
```

### Issue: 429 Too Many Requests

**Causes:**
1. Client exceeding rate limits
2. Rate limits too strict

**Solutions:**
```nginx
# Increase burst size
limit_req zone=api_limit burst=100 nodelay;

# Increase rate
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=20r/s;
```

### Issue: WebSocket Connection Failed

**Causes:**
1. Missing WebSocket headers
2. Timeout too short
3. Buffering enabled

**Check Config:**
```nginx
proxy_http_version 1.1;
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection $connection_upgrade;
proxy_buffering off;
proxy_read_timeout 3600s;
```

### Issue: Static Files Not Caching

**Causes:**
1. Incorrect `Cache-Control` headers
2. Browser ignoring cache
3. Files not in `/assets/` directory

**Check Headers:**
```bash
curl -I http://localhost/assets/index-abc123.js
# Should see: Cache-Control: public, immutable
# Should see: Expires: [1 year from now]
```

---

## Configuration Validation

**Test config syntax:**
```bash
# Local test
nginx -t

# Docker test
docker compose exec nginx nginx -t
```

**Reload config (no downtime):**
```bash
# Local
nginx -s reload

# Docker
docker compose exec nginx nginx -s reload
```

---

## Performance Tuning

### For High Traffic

```nginx
worker_processes auto;
worker_connections 8192;        # Increase connections
worker_rlimit_nofile 100000;    # Increase file limit

upstream backend {
    keepalive 64;                # More keepalive connections
    keepalive_requests 10000;    # More requests per connection
}
```

### For Large Files

```nginx
client_max_body_size 100m;      # Allow larger uploads
proxy_request_buffering off;    # Stream uploads
proxy_buffering off;            # Stream downloads
```

### For CDN Integration

```nginx
# Add CDN-friendly headers
add_header X-Cache-Status $upstream_cache_status;
add_header X-Content-Type-Options "nosniff" always;

# Enable Vary header
gzip_vary on;
```

---

## File Paths Reference

| File | Path |
|------|------|
| Base Config | `infra/nginx/nginx.conf` |
| Dev Config | `infra/nginx/dev.nginx` |
| Prod Config | `infra/nginx/prod.nginx` |
| Dev Compose | `dev.compose.yml` |
| Prod Compose | `compose.yml` |
| Prod Dockerfile | `infra/docker/frontend-builder.prod` |

---

**This Nginx configuration is production-ready with performance optimizations, security hardening, and comprehensive monitoring capabilities for 2025 web applications.**
