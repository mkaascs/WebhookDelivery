ALTER TABLE subscriptions
DROP CONSTRAINT subscriptions_event_type_endpoint_id_key;

DROP INDEX idx_subscriptions_endpoint_id;

ALTER TABLE subscriptions
ADD CONSTRAINT subscriptions_endpoint_id_event_type_key
UNIQUE (endpoint_id, event_type);