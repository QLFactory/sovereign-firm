-- 001_init.down.sql
-- Rollback core schema

DROP TRIGGER IF EXISTS projects_updated_at ON projects;
DROP TRIGGER IF EXISTS users_updated_at ON users;
DROP TRIGGER IF EXISTS tenants_updated_at ON tenants;

DROP FUNCTION IF EXISTS update_updated_at();

DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS artifacts;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tenants;
