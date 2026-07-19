ALTER TABLE deliveries DROP COLUMN IF EXISTS status;

ALTER TABLE deliveries
ADD COLUMN status TEXT NOT NULL DEFAULT 'pending'
CHECK (status IN ('pending', 'failed', 'delivered'));