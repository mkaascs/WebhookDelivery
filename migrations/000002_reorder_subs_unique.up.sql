ALTER TABLE subscriptions
DROP CONSTRAINT subscriptions_endpoint_id_event_type_key;

ALTER TABLE subscriptions
ADD CONSTRAINT subscriptions_event_type_endpoint_id_key
UNIQUE (event_type, endpoint_id);

CREATE INDEX idx_subscriptions_endpoint_id
ON subscriptions(endpoint_id);