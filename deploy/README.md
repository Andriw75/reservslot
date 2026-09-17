# Deploy — ReservSlot

Sistema de reserva de ambientes deportivos.

```
ReservSlot/
├── back/            # API Go (chi) — puerto 8080, SSE en /events
├── front/           # Angular 21 (standalone)
└── deploy/
    ├── nginx.conf   # server block para nginx (puerto 5000)
    └── README.md
```

---

## 1. Entornos del frontend

### Desarrollo (`front/src/environments/environment.ts`)

```ts
export const environment = {
  API_URL: 'http://localhost:8080',
  MEDIA_URL: 'http://localhost:8080/public',
};
```

El backend sirve las imágenes en `/public/*`, así que en desarrollo no hace falta nginx.

### Producción (`front/src/environments/environment.production.ts`)

```ts
export const environment = {
  API_URL: '/api',
  MEDIA_URL: '/public',
};
```

`npm run build` usa `environment.production.ts` automáticamente (ver `angular.json` → `fileReplacements`).
La SPA se publica en la raíz (`baseHref: "/"`).

---

## 2. Backend (Go)

```powershell
cd back
go build -o server.exe ./cmd/server
.\server.exe
```

Variables en `back/.env`:

```
SERVER_PORT=:8080
DATABASE_URL=postgresql://usuario:password@host:5432/reservslot?sslmode=require
PUBLIC_DIR=.
JWT_SECRET=...
```

- `PUBLIC_DIR` es la raíz web; el código agrega `/public` y dentro `/imagenes`.
  Con `PUBLIC_DIR=.` las imágenes quedan en `back/public/imagenes`.

Servidor: `http://localhost:8080` — health en `/health`.

### Base de datos

Los archivos SQL viven en `back/db/` y se aplican **manualmente** en el servidor
(el backend no ejecuta migraciones ni verifica el esquema al arrancar):

```
back/db/schema.sql   # esquema completo (tablas, funciones, triggers, índices)
back/db/seed.sql     # usuario inicial (usuario / placeholder-password)
```

Puedes pegarlos en el editor SQL de Neon, o ejecutarlos con psql:

```powershell
psql "$env:DATABASE_URL" -f back/db/schema.sql
psql "$env:DATABASE_URL" -f back/db/seed.sql
```

---

## 3. Frontend (Angular)

```powershell
cd front
npm install
npm run build
```

El build se genera en `front/dist/front/browser`.

```powershell
Copy-Item -Recurse -Force .\dist\front\browser\* "C:\nginx\nginx-1.28.0\html\reservslot\"
```

Para desarrollo: `npm start` (http://localhost:4200, requiere CORS hacia `:8080`).

---

## 4. nginx

1. Copia `deploy/nginx.conf` dentro del bloque `http { ... }` de tu `nginx.conf`.
2. Recarga:

```powershell
nginx -s reload
```

| Ruta        | Origen                                              |
| ----------- | --------------------------------------------------- |
| `/`         | SPA Angular (`html/reservslot`)                     |
| `/public/`  | Imágenes (`html/reservslot/public/imagenes`)        |
| `/api/`     | Proxy → `127.0.0.1:8080` (REST)                     |
| `/api/events` | Proxy SSE (buffering off, timeout 3600s)          |

> `proxy_pass http://127.0.0.1:8080/;` (con `/` final) hace que
> `/api/auth/token` llegue al backend como `/auth/token`.

---

## 5. API — endpoints principales

Público:

- `GET /categorias`, `GET /categorias/{id}`
- `GET /ambientes`, `GET /ambientes/{id}`
- `GET /ambientes/{id}/info`
- `GET /ambientes/{id}/ocupaciones?from=&to=` (sin parámetros → semana actual America/Lima)
- `GET /events?topics=categorias,ambientes,ambiente:1:info,ambiente:1:ocupaciones` (SSE)

Requiere sesión (cookie JWT):

- `POST /auth/token` (form-urlencoded `username`, `password`), `POST /auth/logout`, `GET /auth/me`
- `POST|PUT|PATCH|DELETE /categorias[...]`
- `POST|PUT|PATCH|DELETE /ambientes[...]`, `PUT|PATCH /ambientes/{id}/info`
- `POST /ambientes/{id}/imagenes` (multipart, campo `images`, máx 10 MB c/u)
- `DELETE /ambientes/{id}/imagenes/{imagenID}`
- `POST|PUT|PATCH|DELETE /ocupaciones[...]`

Usuario inicial (mock): `usuario` / `placeholder-password`.

---

## 6. Tiempo real (SSE)

Un único `EventSource` multiplexado. El cliente declara los topics al conectar:

```
GET /events?topics=categorias,ambientes,ambiente:1:info,ambiente:1:ocupaciones
```

Eventos: `snapshot`, `categoria.upsert`, `categoria.delete`, `ambiente.upsert`,
`ambiente.delete`, `ambiente_info.upsert`, `ocupacion.upsert`, `ocupacion.delete`.

Payload: `{ "topic": "...", "data": ... }`. Las ocupaciones públicas se emiten
solo si intersecan la semana actual.
