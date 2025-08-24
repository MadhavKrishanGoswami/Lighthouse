-- name: InsertContainer :one
-- Inserts or updates container based on container_uid
INSERT INTO containers (
  container_uid,
  host_id,
  name,
  image,
  ports,
  env_vars,
  volumes,
  network
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (container_uid)
DO UPDATE SET
  host_id = EXCLUDED.host_id,
  name = EXCLUDED.name,
  image = EXCLUDED.image,
  ports = EXCLUDED.ports,
  env_vars = EXCLUDED.env_vars,
  volumes = EXCLUDED.volumes,
  network = EXCLUDED.network
RETURNING *;
-- name: GetallContainersWhereWatched :many
-- Retrieves all containers where watched is true
SELECT * FROM containers WHERE watch = TRUE;
