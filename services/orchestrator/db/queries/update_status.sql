-- name: InsertUpdateStatus :one
-- Updates the status of a deployment.
INSERT INTO update_status (
  host_id,
  stage,
  logs,
  image
  ) 
VALUES (
  $1, $2, $3, $4
) RETURNING *;
