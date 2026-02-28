#!/bin/bash
set -euo pipefail

WORKER_ID=${WORKER_ID:-1}
DB_HOST=${DB_HOST:-db}
MYSQL="mariadb -h${DB_HOST} -uroot -pSecret1234 testdb"

echo "[worker ${WORKER_ID}] waiting for database..."
until $MYSQL -e "SELECT 1" &>/dev/null; do
  sleep 1
done
echo "[worker ${WORKER_ID}] connected, starting work loop"

while true; do
  # Pick a random existing row to contend on
  ROW_ID=$(( (RANDOM % 5) + 1 ))

  $MYSQL 2>&1 <<EOF || echo "[worker ${WORKER_ID}] transaction failed (deadlock or lock timeout), retrying..."
START TRANSACTION;

-- Acquire a row-level lock so other workers must wait
SELECT id, status, data
FROM orders
WHERE id = ${ROW_ID}
FOR UPDATE;

-- Simulate work being done while holding the lock
SELECT SLEEP(3);

-- Update the locked row
UPDATE orders
SET worker_id = ${WORKER_ID},
    status    = 'processing',
    data      = CONCAT('handled by worker ${WORKER_ID} at ', NOW())
WHERE id = ${ROW_ID};

-- Insert an audit/log entry
INSERT INTO orders (worker_id, status, data)
VALUES (${WORKER_ID}, 'done', CONCAT('worker ${WORKER_ID} finished row ${ROW_ID} at ', NOW()));

COMMIT;
EOF

  echo "[worker ${WORKER_ID}] finished iteration (row ${ROW_ID})"
  sleep 1
done
