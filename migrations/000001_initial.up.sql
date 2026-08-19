BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE lab_role AS ENUM ('registrar', 'tester', 'reviewer', 'manager');
CREATE TYPE request_status AS ENUM ('draft', 'submitted', 'completed');
CREATE TYPE sample_status AS ENUM ('registered', 'split', 'in_testing', 'sealed', 'disposed');
CREATE TYPE custody_status AS ENUM ('pending', 'confirmed', 'reversed');
CREATE TYPE method_status AS ENUM ('draft', 'published', 'archived');
CREATE TYPE task_status AS ENUM ('pending', 'in_progress', 'review', 'approved', 'rejected');

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username text NOT NULL UNIQUE,
    display_name text NOT NULL,
    password_hash text NOT NULL,
    role lab_role NOT NULL,
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE refresh_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id),
    token_digest char(64) NOT NULL UNIQUE,
    user_agent text NOT NULL DEFAULT '',
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (expires_at > created_at)
);
CREATE INDEX refresh_sessions_user_active_idx ON refresh_sessions(user_id, expires_at) WHERE revoked_at IS NULL;

CREATE TABLE commissions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    reference text UNIQUE,
    organization text NOT NULL DEFAULT '',
    contact_name text NOT NULL DEFAULT '',
    contact_phone text NOT NULL DEFAULT '',
    material_grade text NOT NULL DEFAULT '',
    batch_number text NOT NULL DEFAULT '',
    sample_description text NOT NULL DEFAULT '',
    requirements text NOT NULL DEFAULT '',
    received_at timestamptz,
    status request_status NOT NULL DEFAULT 'draft',
    revision integer NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_by uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX commissions_cursor_idx ON commissions(created_at DESC, id DESC);

CREATE TABLE locations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code text NOT NULL UNIQUE,
    name text NOT NULL,
    active boolean NOT NULL DEFAULT true
);

CREATE TABLE samples (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    commission_id uuid NOT NULL UNIQUE REFERENCES commissions(id),
    number text NOT NULL UNIQUE,
    description text NOT NULL,
    total_quantity numeric(20,6) NOT NULL CHECK (total_quantity > 0),
    unit text NOT NULL,
    status sample_status NOT NULL DEFAULT 'registered',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE subsamples (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    sample_id uuid NOT NULL REFERENCES samples(id),
    parent_id uuid REFERENCES subsamples(id),
    label text NOT NULL,
    purpose text NOT NULL,
    quantity numeric(20,6) NOT NULL CHECK (quantity > 0),
    unit text NOT NULL,
    location_id uuid REFERENCES locations(id),
    status sample_status NOT NULL DEFAULT 'registered',
    current_holder uuid REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(sample_id, label),
    CHECK (parent_id IS NULL OR parent_id <> id)
);
CREATE INDEX subsamples_sample_idx ON subsamples(sample_id, created_at, id);

CREATE TABLE custody_transfers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    subsample_id uuid NOT NULL REFERENCES subsamples(id),
    action text NOT NULL CHECK (action IN ('pickup','transfer','return','seal','reversal')),
    from_user_id uuid NOT NULL REFERENCES users(id),
    to_user_id uuid NOT NULL REFERENCES users(id),
    from_location_id uuid REFERENCES locations(id),
    to_location_id uuid REFERENCES locations(id),
    condition text NOT NULL,
    notes text NOT NULL DEFAULT '',
    status custody_status NOT NULL DEFAULT 'pending',
    original_id uuid REFERENCES custody_transfers(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    confirmed_at timestamptz,
    CHECK (from_user_id <> to_user_id),
    CHECK ((action = 'reversal') = (original_id IS NOT NULL))
);
CREATE INDEX custody_recipient_pending_idx ON custody_transfers(to_user_id, created_at, id) WHERE status = 'pending';

CREATE TABLE methods (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code text NOT NULL,
    name text NOT NULL,
    material_scope text NOT NULL,
    version integer NOT NULL CHECK (version > 0),
    status method_status NOT NULL DEFAULT 'draft',
    formula text NOT NULL,
    precision smallint NOT NULL CHECK (precision BETWEEN 0 AND 12),
    decision_rule text NOT NULL,
    created_by uuid NOT NULL REFERENCES users(id),
    published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(code, version),
    CHECK ((status = 'published') = (published_at IS NOT NULL) OR status <> 'published')
);

CREATE TABLE method_fields (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    method_id uuid NOT NULL REFERENCES methods(id) ON DELETE CASCADE,
    name text NOT NULL,
    label text NOT NULL,
    field_type text NOT NULL CHECK (field_type IN ('number','text','boolean')),
    unit text NOT NULL DEFAULT '',
    minimum numeric(30,12),
    maximum numeric(30,12),
    required boolean NOT NULL DEFAULT false,
    sort_order integer NOT NULL,
    UNIQUE(method_id, name),
    CHECK (minimum IS NULL OR maximum IS NULL OR minimum <= maximum)
);

CREATE TABLE test_tasks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    commission_id uuid NOT NULL REFERENCES commissions(id),
    subsample_id uuid NOT NULL REFERENCES subsamples(id),
    method_id uuid NOT NULL REFERENCES methods(id),
    method_version integer NOT NULL,
    assignee_id uuid NOT NULL REFERENCES users(id),
    executor_id uuid REFERENCES users(id),
    status task_status NOT NULL DEFAULT 'pending',
    current_round integer NOT NULL DEFAULT 0 CHECK (current_round >= 0),
    lock_version integer NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX test_tasks_assignee_cursor_idx ON test_tasks(assignee_id, status, created_at DESC, id DESC);

CREATE TABLE test_rounds (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id uuid NOT NULL REFERENCES test_tasks(id),
    round_number integer NOT NULL CHECK (round_number > 0),
    result numeric(30,12) NOT NULL,
    decision text NOT NULL,
    reason text NOT NULL DEFAULT '',
    created_by uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(task_id, round_number)
);

CREATE TABLE raw_readings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    round_id uuid NOT NULL REFERENCES test_rounds(id),
    field_name text NOT NULL,
    decimal_value numeric(40,18),
    text_value text,
    boolean_value boolean,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(round_id, field_name),
    CHECK (num_nonnulls(decimal_value, text_value, boolean_value) = 1)
);

CREATE TABLE retest_requests (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id uuid NOT NULL REFERENCES test_tasks(id),
    round_number integer NOT NULL,
    reason text NOT NULL,
    status text NOT NULL CHECK (status IN ('pending','approved','rejected')),
    decided_by uuid REFERENCES users(id),
    decision_reason text,
    created_at timestamptz NOT NULL DEFAULT now(),
    decided_at timestamptz
);

CREATE TABLE technical_reviews (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id uuid NOT NULL REFERENCES test_tasks(id),
    revision integer NOT NULL,
    reviewer_id uuid NOT NULL REFERENCES users(id),
    decision text NOT NULL CHECK (decision IN ('approve','return')),
    comment text NOT NULL,
    snapshot jsonb NOT NULL,
    difference jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(task_id, revision)
);

CREATE TABLE certificates (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    number text NOT NULL UNIQUE,
    commission_id uuid NOT NULL REFERENCES commissions(id),
    sample_id uuid NOT NULL REFERENCES samples(id),
    status text NOT NULL CHECK (status IN ('issued','voided')),
    verification_digest char(64) NOT NULL UNIQUE,
    content_digest char(64) NOT NULL,
    snapshot jsonb NOT NULL,
    supersedes_id uuid REFERENCES certificates(id),
    issued_by uuid NOT NULL REFERENCES users(id),
    issued_at timestamptz NOT NULL,
    voided_at timestamptz,
    void_reason text,
    CHECK ((status = 'voided') = (voided_at IS NOT NULL))
);

CREATE TABLE attachments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_type text NOT NULL,
    owner_id uuid NOT NULL,
    storage_key text NOT NULL UNIQUE,
    filename text NOT NULL,
    content_type text NOT NULL,
    size_bytes bigint NOT NULL CHECK (size_bytes > 0),
    sha256 char(64) NOT NULL,
    uploaded_by uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE background_jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    kind text NOT NULL,
    payload jsonb NOT NULL,
    attempts integer NOT NULL DEFAULT 0,
    max_attempts integer NOT NULL DEFAULT 5,
    available_at timestamptz NOT NULL DEFAULT now(),
    lease_owner text,
    lease_until timestamptz,
    completed_at timestamptz,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX background_jobs_lease_idx ON background_jobs(available_at, created_at) WHERE completed_at IS NULL;

CREATE TABLE notifications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id),
    kind text NOT NULL,
    title text NOT NULL,
    object_id uuid,
    read_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE audit_events (
    sequence bigserial PRIMARY KEY,
    id uuid NOT NULL UNIQUE,
    actor_id uuid REFERENCES users(id),
    action text NOT NULL,
    object_type text NOT NULL,
    object_id uuid NOT NULL,
    request_id text NOT NULL,
    before_summary jsonb NOT NULL DEFAULT '{}'::jsonb,
    after_summary jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX audit_events_cursor_idx ON audit_events(sequence DESC);

CREATE OR REPLACE FUNCTION reject_audit_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'audit events are append-only';
END;
$$;
CREATE TRIGGER audit_events_no_update BEFORE UPDATE OR DELETE ON audit_events FOR EACH ROW EXECUTE FUNCTION reject_audit_mutation();

COMMIT;
