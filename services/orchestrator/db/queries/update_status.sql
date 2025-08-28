--- name: InsertUpdateStatus :one
INSERT INTO update_status (
    deployment_id, host_id, stage, logs
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

