BEGIN;
DROP TABLE IF EXISTS audit_events, notifications, background_jobs, attachments, certificates, technical_reviews, retest_requests, raw_readings, test_rounds, test_tasks, method_fields, methods, custody_transfers, subsamples, samples, locations, commissions, refresh_sessions, users CASCADE;
DROP TYPE IF EXISTS task_status, method_status, custody_status, sample_status, request_status, lab_role;
DROP FUNCTION IF EXISTS reject_audit_mutation();
COMMIT;
