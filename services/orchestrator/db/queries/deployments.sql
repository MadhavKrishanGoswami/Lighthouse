-- name: InsertDeployment :one
-- Inserts a deployment for a container on a host.
-- If a deployment already exists for the same container+host and is still pending/running,
-- do NOT overwrite it; otherwise, update target_image and status.
INSERT INTO deployments (
    container_id,
    host_id,
    target_image,
    status
) VALUES ($1, $2, $3, $4)
ON CONFLICT (container_id, host_id) 
DO UPDATE SET
    target_image = EXCLUDED.target_image,
    status = EXCLUDED.status
RETURNING *;

