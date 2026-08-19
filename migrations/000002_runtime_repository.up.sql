CREATE TABLE runtime_entities (
    kind text NOT NULL,
    id text NOT NULL,
    payload jsonb NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (kind, id)
);
CREATE INDEX runtime_entities_kind_idx ON runtime_entities(kind, updated_at DESC);
