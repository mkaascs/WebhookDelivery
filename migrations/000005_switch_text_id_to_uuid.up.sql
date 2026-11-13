TRUNCATE deliveries, subscriptions, events, endpoints;

ALTER TABLE subscriptions DROP CONSTRAINT subscriptions_endpoint_id_fkey;
ALTER TABLE deliveries DROP CONSTRAINT deliveries_endpoint_id_fkey;
ALTER TABLE deliveries DROP CONSTRAINT deliveries_event_id_fkey;

ALTER TABLE endpoints ALTER COLUMN id TYPE uuid USING id::uuid;
ALTER TABLE events ALTER COLUMN id TYPE uuid USING id::uuid;

ALTER TABLE subscriptions
ALTER COLUMN id TYPE uuid USING id::uuid,
    ALTER COLUMN endpoint_id TYPE uuid USING endpoint_id::uuid;

ALTER TABLE deliveries
ALTER COLUMN id TYPE uuid USING id::uuid,
    ALTER COLUMN endpoint_id TYPE uuid USING endpoint_id::uuid,
    ALTER COLUMN event_id TYPE uuid USING event_id::uuid;

ALTER TABLE subscriptions
    ADD CONSTRAINT subscriptions_endpoint_id_fkey
        FOREIGN KEY (endpoint_id) REFERENCES endpoints(id) ON DELETE CASCADE;

ALTER TABLE deliveries
    ADD CONSTRAINT deliveries_endpoint_id_fkey
        FOREIGN KEY (endpoint_id) REFERENCES endpoints(id) ON DELETE CASCADE,
    ADD CONSTRAINT deliveries_event_id_fkey
        FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;