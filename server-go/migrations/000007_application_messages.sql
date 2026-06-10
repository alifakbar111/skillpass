-- Phase 1: Application pipeline messages (company notes to candidates)

CREATE TABLE IF NOT EXISTS application_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    sender_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_type TEXT NOT NULL DEFAULT 'note',
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS application_messages_application_idx ON application_messages(application_id, created_at);
