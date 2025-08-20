CREATE TABLE "hosts" (
  "id" serial PRIMARY KEY,
  "hostname" varchar UNIQUE NOT NULL,
  "ip_address" varchar UNIQUE NOT NULL,
  "created_at" timestamp DEFAULT (now())
);

CREATE TABLE "containers" (
  "id" serial PRIMARY KEY,
  "host_id" int NOT NULL,
  "name" varchar NOT NULL,
  "image" varchar NOT NULL,
  "created_at" timestamp DEFAULT (now())
);

CREATE TABLE "deployments" (
  "id" serial PRIMARY KEY,
  "container_id" int NOT NULL,
  "version" varchar NOT NULL,
  "status" varchar NOT NULL,
  "created_at" timestamp DEFAULT (now())
);

CREATE TABLE "update_status" (
  "id" serial PRIMARY KEY,
  "deployment_id" int NOT NULL,
  "stage" varchar NOT NULL,
  "message" text,
  "created_at" timestamp DEFAULT (now())
);

CREATE INDEX "idx_containers_host_id" ON "containers" ("host_id");

CREATE INDEX "idx_deployments_container_id" ON "deployments" ("container_id");

CREATE INDEX "idx_update_status_deployment_id" ON "update_status" ("deployment_id");

COMMENT ON COLUMN "deployments"."status" IS 'pending, running, success, failed, rollback';

COMMENT ON COLUMN "update_status"."stage" IS 'pulling, starting, health_check, completed, rollback, failed';

ALTER TABLE "containers" ADD FOREIGN KEY ("host_id") REFERENCES "hosts" ("id");

ALTER TABLE "deployments" ADD FOREIGN KEY ("container_id") REFERENCES "containers" ("id");

ALTER TABLE "update_status" ADD FOREIGN KEY ("deployment_id") REFERENCES "deployments" ("id");
