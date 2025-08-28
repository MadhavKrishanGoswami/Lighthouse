-- name: InsertContainer :one
-- Inserts or updates container based on container_uid
INSERT INTO containers (
  container_uid,
  host_id,
  name,
  image,
  digest,
  ports,
  env_vars,
  volumes,
  network
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (container_uid)
DO UPDATE SET
  host_id = EXCLUDED.host_id,
  name = EXCLUDED.name,
  image = EXCLUDED.image,
  digest = EXCLUDED.digest,
  ports = EXCLUDED.ports,
  env_vars = EXCLUDED.env_vars,
  volumes = EXCLUDED.volumes,
  network = EXCLUDED.network
RETURNING *;
-- name: DeleteStaleContainersForHost :exec
-- Deletes containers for a given host that are not in the provided list of UIDs.
DELETE FROM containers
WHERE host_id = $1 AND container_uid <> ALL($2::text[]);
-- name: GetallContainersWhereWatched :many
-- Retrieves all containers where watched is true
SELECT * FROM containers WHERE watch = TRUE;
-- name: GetHostbyContainerUID :one
-- Retrieves the host associated with a given container UID 
SELECT h.*
FROM hosts h
JOIN containers c ON h.id = c.host_id
WHERE c.container_uid = $1;
-- name: GetContainerbyContainerUID :one
-- Retrieves a container by its UID
SELECT * FROM containers WHERE container_uid = $1;
