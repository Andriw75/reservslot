-- ============================================================
-- ReservSlot — Datos iniciales (seed)
--   psql "$DATABASE_URL" -f db/seed.sql
--
-- Usuario inicial:
--   usuario:  definido-por-entorno
--   password: placeholder-password   (hash bcrypt, cost 12)
--
-- Los permisos existen en la tabla `permisos` pero NO se enlazan a usuarios:
-- de momento cualquier usuario autenticado puede operar todos los módulos.
-- ============================================================

BEGIN;

INSERT INTO usuarios (username, password_hash, activo)
SELECT 'usuario', '$2a$12$PLACEHOLDER', true
WHERE NOT EXISTS (
  SELECT 1 FROM usuarios WHERE lower(username) = lower('usuario')
);

COMMIT;
