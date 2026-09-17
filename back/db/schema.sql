-- ============================================================
-- ReservSlot — Esquema de base de datos (PostgreSQL)
-- Genera el esquema completo desde cero.
--   psql "$DATABASE_URL" -f db/schema.sql
-- ============================================================

BEGIN;

CREATE EXTENSION IF NOT EXISTS btree_gist;

-- ============================================================
-- Funciones
-- ============================================================

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
$$;

-- ============================================================
-- Autorización
-- ============================================================

CREATE TABLE IF NOT EXISTS permisos (
  codigo varchar(100) PRIMARY KEY,
  descripcion text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS usuarios (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  username varchar(100) NOT NULL,
  password_hash text NOT NULL,
  activo boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT usuarios_username_no_vacio CHECK (btrim(username) <> '')
);

CREATE UNIQUE INDEX IF NOT EXISTS usuarios_username_unico_ci
  ON usuarios (lower(username));

CREATE TABLE IF NOT EXISTS usuario_permisos (
  usuario_id bigint NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
  permiso_codigo varchar(100) NOT NULL REFERENCES permisos(codigo) ON DELETE RESTRICT,
  assigned_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (usuario_id, permiso_codigo)
);

CREATE INDEX IF NOT EXISTS usuario_permisos_permiso_idx
  ON usuario_permisos (permiso_codigo);

INSERT INTO permisos (codigo, descripcion) VALUES
  ('usuarios:read', 'Consultar usuarios'),
  ('usuarios:write', 'Crear, editar y eliminar usuarios'),
  ('catalogo:write', 'Administrar categorías, ambientes e información'),
  ('archivos:write', 'Subir y eliminar imágenes'),
  ('ocupaciones:read', 'Consultar ocupaciones'),
  ('ocupaciones:write', 'Crear, editar y borrar ocupaciones')
ON CONFLICT (codigo) DO NOTHING;

-- ============================================================
-- Catálogo
-- ============================================================

CREATE TABLE IF NOT EXISTS categorias (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(150) NOT NULL,
  descripcion text,
  color varchar(7),
  activo boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT categorias_nombre_no_vacio CHECK (btrim(nombre) <> ''),
  CONSTRAINT categorias_color_hex CHECK (color IS NULL OR color ~ '^#[0-9A-Fa-f]{6}$')
);

CREATE UNIQUE INDEX IF NOT EXISTS categorias_nombre_unico_ci
  ON categorias (lower(nombre));

CREATE TABLE IF NOT EXISTS ambientes (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(255) NOT NULL,
  categoria_id bigint NOT NULL REFERENCES categorias(id) ON DELETE RESTRICT,
  activo boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT ambientes_nombre_no_vacio CHECK (btrim(nombre) <> '')
);

CREATE INDEX IF NOT EXISTS ambientes_categoria_idx ON ambientes (categoria_id);

-- Información extendida del ambiente en JSONB:
-- {
--   "ancho": 20, "alto": 10, "aforo": 30,
--   "descripcion": "Cancha techada",
--   "opciones_adicionales": { "iluminacion": true, "vestidores": true }
-- }
CREATE TABLE IF NOT EXISTS ambiente_info (
  ambiente_id bigint PRIMARY KEY REFERENCES ambientes(id) ON DELETE CASCADE,
  data jsonb NOT NULL DEFAULT '{}'::jsonb,
  activo boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT ambiente_info_data_es_objeto CHECK (jsonb_typeof(data) = 'object')
);

CREATE INDEX IF NOT EXISTS ambiente_info_gin_idx ON ambiente_info USING gin (data);

CREATE TABLE IF NOT EXISTS imagenes (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  ambiente_id bigint NOT NULL REFERENCES ambientes(id) ON DELETE CASCADE,
  archivo varchar(500) NOT NULL,
  descripcion text,
  orden integer NOT NULL DEFAULT 0,
  mime_type varchar(100),
  bytes bigint,
  activo boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT imagenes_archivo_no_vacio CHECK (btrim(archivo) <> ''),
  CONSTRAINT imagenes_orden_no_negativo CHECK (orden >= 0),
  CONSTRAINT imagenes_bytes_positivos CHECK (bytes IS NULL OR bytes > 0)
);

CREATE INDEX IF NOT EXISTS imagenes_ambiente_idx ON imagenes (ambiente_id, orden);

CREATE TABLE IF NOT EXISTS ocupaciones (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  ambiente_id bigint NOT NULL REFERENCES ambientes(id) ON DELETE RESTRICT,
  inicio timestamptz NOT NULL,
  fin timestamptz NOT NULL,
  activo boolean NOT NULL DEFAULT true,
  created_by bigint REFERENCES usuarios(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT ocupaciones_inicio_menor_fin CHECK (inicio < fin)
);

CREATE INDEX IF NOT EXISTS ocupaciones_ambiente_inicio_idx
  ON ocupaciones (ambiente_id, inicio);

CREATE INDEX IF NOT EXISTS ocupaciones_activas_ambiente_inicio_idx
  ON ocupaciones (ambiente_id, inicio)
  WHERE activo = true;

-- Devuelve true si el rango [p_inicio, p_fin) se solapa con otra ocupación
-- activa del mismo ambiente. Es la misma regla que aplica la restricción
-- EXCLUDE: intervalos semiabiertos, de modo que 09:00-10:00 y 10:00-11:00
-- no se consideran solapados.
CREATE OR REPLACE FUNCTION ocupacion_solapa(
  p_ambiente_id bigint,
  p_inicio timestamptz,
  p_fin timestamptz,
  p_ignore_id bigint DEFAULT NULL
)
RETURNS boolean
LANGUAGE sql
STABLE
AS $$
  SELECT EXISTS (
    SELECT 1
    FROM ocupaciones o
    WHERE o.ambiente_id = p_ambiente_id
      AND o.activo = true
      AND (p_ignore_id IS NULL OR o.id <> p_ignore_id)
      AND o.inicio < p_fin
      AND o.fin > p_inicio
  );
$$;

-- Garantía de no solapamiento a nivel de motor (refuerza la validación de la
-- aplicación). Rangos semiabiertos: [inicio, fin).
ALTER TABLE ocupaciones
  DROP CONSTRAINT IF EXISTS ocupaciones_no_solapadas;

ALTER TABLE ocupaciones
  ADD CONSTRAINT ocupaciones_no_solapadas
  EXCLUDE USING gist (
    ambiente_id WITH =,
    tstzrange(inicio, fin, '[)') WITH &&
  )
  WHERE (activo = true);

-- ============================================================
-- Triggers de auditoría temporal
-- ============================================================

DROP TRIGGER IF EXISTS usuarios_set_updated_at ON usuarios;
CREATE TRIGGER usuarios_set_updated_at
BEFORE UPDATE ON usuarios
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS categorias_set_updated_at ON categorias;
CREATE TRIGGER categorias_set_updated_at
BEFORE UPDATE ON categorias
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS ambientes_set_updated_at ON ambientes;
CREATE TRIGGER ambientes_set_updated_at
BEFORE UPDATE ON ambientes
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS ambiente_info_set_updated_at ON ambiente_info;
CREATE TRIGGER ambiente_info_set_updated_at
BEFORE UPDATE ON ambiente_info
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS imagenes_set_updated_at ON imagenes;
CREATE TRIGGER imagenes_set_updated_at
BEFORE UPDATE ON imagenes
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS ocupaciones_set_updated_at ON ocupaciones;
CREATE TRIGGER ocupaciones_set_updated_at
BEFORE UPDATE ON ocupaciones
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
