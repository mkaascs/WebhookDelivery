ALTER TABLE subscriptions DROP CONSTRAINT subscriptions_endpoint_id_fkey;
ALTER TABLE deliveries DROP CONSTRAINT deliveries_endpoint_id_fkey;
ALTER TABLE deliveries DROP CONSTRAINT deliveries_event_id_fkey;

ALTER TABLE endpoints ALTER COLUMN id TYPE text USING id::text;
ALTER TABLE events ALTER COLUMN id TYPE text USING id::text;

ALTER TABLE subscriptions
ALTER COLUMN id TYPE text USING id::text,
    ALTER COLUMN endpoint_id TYPE text USING endpoint_id::text;

ALTER TABLE deliveries
ALTER COLUMN id TYPE text USING id::text,
    ALTER COLUMN endpoint_id TYPE text USING endpoint_id::text,
    ALTER COLUMN event_id TYPE text USING event_id::text;

ALTER TABLE subscriptions
    ADD CONSTRAINT subscriptions_endpoint_id_fkey
        FOREIGN KEY (endpoint_id) REFERENCES endpoints(id) ON DELETE CASCADE;

ALTER TABLE deliveries
    ADD CONSTRAINT deliveries_endpoint_id_fkey
        FOREIGN KEY (endpoint_id) REFERENCES endpoints(id) ON DELETE CASCADE,
    ADD CONSTRAINT deliveries_event_id_fkey
        FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;