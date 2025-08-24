-- name: InsertHost :one
-- Inserts a new host or updates an existing one based on the MAC address.
INSERT INTO hosts (
  mac_address,
  hostname,
  ip_address
) VALUES (
  $1, $2, $3
)
ON CONFLICT (mac_address)
DO UPDATE SET
  hostname = EXCLUDED.hostname,
  ip_address = EXCLUDED.ip_address
RETURNING *;

-- name: GertHostByMacAddress :one
-- Retrieves a host by its mac_address.
SELECT * FROM hosts WHERE mac_address = $1;

-- name: UpdateHostLastHeartbeat :one
-- Updates the last heartbeat timestamp for a host identified by id.
UPDATE hosts SET last_heartbeat = NOW() WHERE id = $1 RETURNING *;

