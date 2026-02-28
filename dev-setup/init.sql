USE testdb;

CREATE TABLE IF NOT EXISTS orders (
  id          INT AUTO_INCREMENT PRIMARY KEY,
  worker_id   INT          DEFAULT NULL,
  status      VARCHAR(20)  DEFAULT 'ready',
  data        VARCHAR(255) DEFAULT NULL,
  created_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Seed a fixed set of rows that workers will compete over
INSERT INTO orders (status, data) VALUES
  ('ready', 'row A'),
  ('ready', 'row B'),
  ('ready', 'row C'),
  ('ready', 'row D'),
  ('ready', 'row E');
