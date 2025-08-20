CREATE TYPE "deployment_status" AS ENUM (
  'pending',
  'running',
  'success',
  'failed',
  'rollback'
);

CREATE TYPE "update_stage" AS ENUM (
  'pulling',
  'starting',
  'health_check',
  'completed',
  'rollback',
  'failed'
);

CREATE TABLE "hosts" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "mac_address" varchar UNIQUE NOT NULL,
  "hostname" varchar UNIQUE NOT NULL,
  "ip_address" varchar UNIQUE NOT NULL,
  "created_at" timestamptz DEFAULT (now())
);

CREATE TABLE "containers" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "container_uid" varchar UNIQUE NOT NULL,
  "host_id" uuid NOT NULL,
  "name" varchar NOT NULL,
  "image" varchar NOT NULL,
  "created_at" timestamptz DEFAULT (now())
);

CREATE TABLE "deployments" (
  "id" serial PRIMARY KEY,
  "container_id" uuid NOT NULL,
  "version" varchar NOT NULL,
  "status" deployment_status NOT NULL,
  "created_at" timestamptz DEFAULT (now())
);

CREATE TABLE "update_status" (
  "id" serial PRIMARY KEY,
  "deployment_id" int NOT NULL,
  "stage" update_stage NOT NULL,
  "message" text,
  "created_at" timestamptz DEFAULT (now())
);

CREATE INDEX "idx_containers_host_id" ON "containers" ("host_id");

CREATE UNIQUE INDEX "uq_containers_host_name" ON "containers" ("host_id", "name");

CREATE INDEX "idx_deployments_container_id" ON "deployments" ("container_id");

CREATE INDEX "idx_update_status_deployment_id" ON "update_status" ("deployment_id");

COMMENT ON COLUMN "hosts"."id" IS 'Internal primary key';

COMMENT ON COLUMN "hosts"."mac_address" IS 'User-provided MAC address or host ID';

COMMENT ON COLUMN "containers"."id" IS 'Internal primary key';

COMMENT ON COLUMN "containers"."container_uid" IS 'User-provided container ID (Docker UID)';

ALTER TABLE "containers" ADD FOREIGN KEY ("host_id") REFERENCES "hosts" ("id");

ALTER TABLE "deployments" ADD FOREIGN KEY ("container_id") REFERENCES "containers" ("id");

ALTER TABLE "update_status" ADD FOREIGN KEY ("deployment_id") REFERENCES "deployments" ("id");

