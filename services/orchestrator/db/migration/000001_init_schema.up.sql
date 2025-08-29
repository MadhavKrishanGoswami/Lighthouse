-- Enums 
CREATE TYPE update_stage AS ENUM (
  'pulling',
  'starting',
  'running',
  'health_check',
  'completed',
  'rollback',
  'failed'
);

-- 2. Hosts table
CREATE TABLE hosts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  mac_address varchar UNIQUE NOT NULL,
  hostname varchar UNIQUE NOT NULL,
  ip_address varchar UNIQUE NOT NULL,
  last_heartbeat timestamptz,
  created_at timestamptz DEFAULT now()
);

-- 3. Containers table
CREATE TABLE containers (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  container_uid varchar UNIQUE NOT NULL,
  host_id uuid NOT NULL,
  name varchar NOT NULL,
  image varchar NOT NULL,
  ports jsonb DEFAULT '[]'::jsonb,
  env_vars text[],
  volumes text[],
  network varchar,
  watch boolean DEFAULT TRUE,
  created_at timestamptz DEFAULT now()
);

-- 5. Update status table
CREATE TABLE update_status (
  id SERIAL PRIMARY KEY,
  image varchar NOT NULL,
  host_id uuid NOT NULL,
  stage update_stage NOT NULL,
  logs text,
  created_at timestamptz DEFAULT now()
);

-- 6. Comments
COMMENT ON COLUMN hosts.id IS 'Primary key for hosts. Root entity key.';
COMMENT ON COLUMN containers.host_id IS 'FK → hosts.id. A container belongs to a host.';
COMMENT ON COLUMN update_status.host_id IS 'FK → hosts.id. Update status directly tied to a host.';

-- 7. Foreign keys
ALTER TABLE containers ADD FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE;
ALTER TABLE update_status ADD FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE;

