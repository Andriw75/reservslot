# ReservSlot

Sistema de reserva de ambientes deportivos: catálogo público, detalle con programación, panel de administración y actualizaciones en tiempo real.

Proyecto de portafolio desarrollado con Angular 21 y Go. No incluye una instancia pública; se ejecuta localmente.

## Características

- Catálogo público de ambientes agrupados por categoría.
- Detalle de ambiente con imágenes, información y programación semanal.
- Panel de administración con autenticación JWT en cookie httpOnly.
- CRUD de categorías, ambientes, información y ocupaciones.
- Subida, ordenación y eliminación de imágenes.
- Actualizaciones en tiempo real mediante Server-Sent Events (SSE).
- Enlaces para compartir un ambiente.
- API REST en Go con PostgreSQL.

## Stack

| Capa | Tecnología |
| --- | --- |
| Frontend | Angular 21 (standalone), TypeScript, CSS |
| Backend | Go, chi, pgx, JWT, bcrypt |
| Base de datos | PostgreSQL 16+ |
| Tiempo real | Server-Sent Events (SSE) |
| Servidor | nginx (opcional en desarrollo) |

## Arquitectura

```mermaid
flowchart LR
    U[Usuario] --> N[Angular SPA]
    N -->|REST /api| G[API Go]
    N -->|SSE /events| G
    N -->|imágenes /public| G
    G --> P[(PostgreSQL)]
```

## Estructura

```text
ReservSlot/
├── back/            # API Go (chi), autenticación, repositorios y SSE
│   ├── api/         # rutas y handlers
│   ├── cmd/         # server, create_user, generate_secrets
│   ├── db/          # schema.sql y seed.sql
│   ├── infrastructure/
│   ├── services/
│   └── domain/
├── front/           # Angular 21 (standalone)
│   └── src/
└── deploy/          # configuración de nginx y guía de despliegue
```

## Requisitos

- Go 1.25 o superior.
- Node.js 22 o superior y npm 11 o superior.
- PostgreSQL 16 o superior con la extensión `btree_gist` (incluida en `contrib`).

## Puesta en marcha

### 1. Base de datos

Crea la base y aplica el esquema:

```powershell
$env:DATABASE_URL="postgresql://usuario:password@localhost:5432/reservslot?sslmode=disable"
psql "$env:DATABASE_URL" -f back/db/schema.sql
psql "$env:DATABASE_URL" -f back/db/seed.sql
```

`seed.sql` no crea usuarios: las credenciales nunca se guardan en el repositorio.

### 2. Configuración del backend

```powershell
Copy-Item back\.env.example back\.env
cd back
go run ./cmd/generate_secrets.go
```

Copia el valor generado en `JWT_SECRET` dentro de `back/.env` y ajusta `DATABASE_URL` y `PUBLIC_DIR`.

### 3. Primer usuario

Define las credenciales solo en el entorno y ejecuta el comando:

```powershell
cd back
$env:ADMIN_USERNAME="admin"
$env:ADMIN_PASSWORD="<contraseña-segura>"
go run ./cmd/create_user
```

En Linux o macOS:

```bash
cd back
ADMIN_USERNAME=admin ADMIN_PASSWORD='<contraseña-segura>' go run ./cmd/create_user
```

El comando crea el usuario si no existe o actualiza su contraseña si ya existe.

### 4. Backend

```powershell
cd back
go run ./cmd/server
```

La API queda disponible en `http://localhost:8080` y el health check en `/health`.

### 5. Frontend

```powershell
cd front
npm install
npm start
```

Abre `http://localhost:4200` e inicia sesión con el usuario creado en el paso 3.

## Variables de entorno

| Variable | Descripción | Ejemplo |
| --- | --- | --- |
| `SERVER_PORT` | Puerto del backend | `:8080` |
| `DATABASE_URL` | Cadena de conexión a PostgreSQL | `postgresql://...` |
| `DB_TYPE` / `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | Alternativa a `DATABASE_URL` | `postgres` / `localhost` / `5432` |
| `DB_SSLMODE` | Modo SSL de PostgreSQL | `require` |
| `JWT_SECRET` | Secreto para firmar tokens; obligatorio | generado con `generate_secrets` |
| `JWT_EXPIRATION_MINUTES` | Duración del token | `60` |
| `JWT_RENEW_MINUTES` | Ventana de renovación | `20` |
| `COOKIE_NAME` | Nombre de la cookie de sesión | `reservslot_token` |
| `COOKIE_SECURE` | `true` en producción con HTTPS | `false` |
| `CORS_ORIGIN` | Origen permitido en desarrollo | `http://localhost:4200` |
| `PUBLIC_DIR` | Raíz de archivos públicos | `.` |
| `X_ACCEL_REDIRECT_PREFIX` | Prefijo opcional para nginx | vacío |

## API

Público:

- `GET /categorias`, `GET /categorias/{id}`
- `GET /ambientes`, `GET /ambientes/{id}`
- `GET /ambientes/{id}/info`
- `GET /ambientes/{id}/ocupaciones?from=&to=`
- `GET /events?topics=...` (SSE)
- `GET /health`

Requiere sesión (cookie JWT):

- `POST /auth/token`, `POST /auth/logout`, `GET /auth/me`
- CRUD de `/categorias`, `/ambientes`, `/ambientes/{id}/info` y `/ocupaciones`
- `POST /ambientes/{id}/imagenes` y `DELETE /ambientes/{id}/imagenes/{imagenID}`

La referencia completa está en [`deploy/README.md`](deploy/README.md).

## Seguridad

- `back/.env` está ignorado por Git y nunca debe publicarse.
- No hay usuarios ni contraseñas en el código: el primer usuario se crea con variables de entorno.
- El backend no arranca si `JWT_SECRET` está vacío.
- En producción usa HTTPS y `COOKIE_SECURE=true`.
- Rota las credenciales y secretos si sospechas que se expusieron.

## Despliegue

La guía de nginx y despliegue manual está en [`deploy/README.md`](deploy/README.md). No se publica ninguna URL de producción en este repositorio.

## Capturas

<!-- Añadir aquí las capturas del catálogo, detalle y panel de administración. -->

## Licencia

MIT © 2026 Andriw75. Consulta [`LICENSE`](LICENSE).
