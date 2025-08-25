-- 1. ENUM types
CREATE TYPE deployment_status AS ENUM (
  'pending',
  'running',
  'success',
  'failed',
  'rollback_initiated',
  'rollback_complete'
);

CREATE TYPE update_stage AS ENUM (
  'pulling',
  'starting',
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
  digest varchar NOT NULL,
  ports text[],
  env_vars text[],
  volumes text[],
  network varchar,
  watch boolean DEFAULT TRUE,
  created_at timestamptz DEFAULT now()
);

-- 4. Deployments table
CREATE TABLE deployments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  container_id uuid NOT NULL,
  host_id uuid NOT NULL,
  target_image varchar NOT NULL,
  configuration text,
  status deployment_status NOT NULL DEFAULT 'pending',
  created_at timestamptz DEFAULT now()
);

-- 5. Update status table
CREATE TABLE update_status (
  id SERIAL PRIMARY KEY,
  deployment_id uuid NOT NULL,
  host_id uuid NOT NULL,
  stage update_stage NOT NULL,
  logs text,
  created_at timestamptz DEFAULT now()
);

-- 6. Comments
COMMENT ON COLUMN hosts.id IS 'Primary key for hosts. Root entity key.';
COMMENT ON COLUMN containers.host_id IS 'FK → hosts.id. A container belongs to a host.';
COMMENT ON COLUMN deployments.host_id IS 'FK → hosts.id. Deployment tied to a host.';
COMMENT ON COLUMN update_status.host_id IS 'FK → hosts.id. Update status directly tied to a host.';

-- 7. Foreign keys
ALTER TABLE containers ADD FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE;
ALTER TABLE deployments ADD FOREIGN KEY (container_id) REFERENCES containers(id) ON DELETE CASCADE;
ALTER TABLE deployments ADD FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE;
ALTER TABLE update_status ADD FOREIGN KEY (deployment_id) REFERENCES deployments(id) ON DELETE CASCADE;
ALTER TABLE update_status ADD FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE;

