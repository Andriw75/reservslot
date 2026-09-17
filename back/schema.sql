BEGIN;

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
$$;

CREATE TABLE permisos (
  codigo varchar(100) PRIMARY KEY,
  descripcion text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE usuarios (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  username varchar(100) NOT NULL,
  password_hash text NOT NULL,
  activo boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT usuarios_username_no_vacio CHECK (btrim(username) <> '')
);

CREATE UNIQUE INDEX usuarios_username_unico_ci ON usuarios (lower(username));

CREATE TABLE usuario_permisos (
  usuario_id bigint NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
  permiso_codigo varchar(100) NOT NULL REFERENCES permisos(codigo) ON DELETE RESTRICT,
  assigned_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (usuario_id, permiso_codigo)
);

CREATE INDEX usuario_permisos_permiso_idx ON usuario_permisos (permiso_codigo);

INSERT INTO permisos (codigo, descripcion) VALUES
  ('usuarios:read','Consultar usuarios'),
  ('usuarios:write','Crear, editar y eliminar usuarios'),
  ('catalogo:write','Administrar categorías, ambientes e información'),
  ('archivos:write','Subir y eliminar imágenes'),
  ('ocupaciones:read','Consultar ocupaciones'),
  ('ocupaciones:write','Crear, editar y borrar ocupaciones')
ON CONFLICT (codigo) DO NOTHING;

CREATE TABLE categorias (
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

CREATE UNIQUE INDEX categorias_nombre_unico_ci ON categorias (lower(nombre)) WHERE activo = true;

CREATE TABLE ambientes (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(255) NOT NULL,
  categoria_id bigint NOT NULL REFERENCES categorias(id) ON DELETE RESTRICT,
  activo boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT ambientes_nombre_no_vacio CHECK (btrim(nombre) <> '')
);

CREATE INDEX ambientes_categoria_idx ON ambientes (categoria_id);

CREATE TABLE ambiente_info (
  ambiente_id bigint PRIMARY KEY REFERENCES ambientes(id) ON DELETE CASCADE,
  data jsonb NOT NULL DEFAULT '{}'::jsonb,
  activo boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT ambiente_info_data_es_objeto CHECK (jsonb_typeof(data) = 'object')
);

CREATE TABLE imagenes (
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

CREATE UNIQUE INDEX imagenes_ambiente_orden_unico ON imagenes (ambiente_id, orden) WHERE activo = true;
CREATE INDEX imagenes_ambiente_idx ON imagenes (ambiente_id, orden);

CREATE TABLE ocupaciones (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  ambiente_id bigint NOT NULL REFERENCES ambientes(id) ON DELETE RESTRICT,
  inicio timestamptz NOT NULL,
  fin timestamptz NOT NULL,
  activo boolean NOT NULL DEFAULT true,
  created_by bigint REFERENCES usuarios(id) ON DELETE SET NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT ocupaciones_inicio_menor_fin CHECK (inicio < fin),
  CONSTRAINT ocupaciones_activo_en_solo_activas CHECK (activo IN (true,false))
);

CREATE INDEX ocupaciones_ambiente_inicio_idx ON ocupaciones (ambiente_id, inicio);
CREATE INDEX ocupaciones_activo_ambiente_inicio_idx ON ocupaciones (ambiente_id, inicio) WHERE activo = true;

ALTER TABLE ocupaciones
  ADD CONSTRAINT ocupaciones_no_solapadas
  EXCLUDE USING gist (
    ambiente_id WITH =,
    tstzrange(inicio, fin, '[)') WITH &&
  )
  WHERE (activo = true);

CREATE TRIGGER usuarios_set_updated_at
BEFORE UPDATE ON usuarios
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER categorias_set_updated_at
BEFORE UPDATE ON categorias
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER ambientes_set_updated_at
BEFORE UPDATE ON ambientes
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER ambiente_info_set_updated_at
BEFORE UPDATE ON ambiente_info
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER imagenes_set_updated_at
BEFORE UPDATE ON imagenes
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
