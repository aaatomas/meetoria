CREATE TABLE IF NOT EXISTS email_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL UNIQUE,
    correlation_id UUID NOT NULL,
    organization_id UUID,
    booking_id UUID,
    recipient_email VARCHAR(255) NOT NULL,
    template VARCHAR(100) NOT NULL,
    variables JSONB DEFAULT '{}',
    subject VARCHAR(500),
    provider VARCHAR(50) NOT NULL,
    provider_message_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    retry_count INT NOT NULL DEFAULT 0,
    sent_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_messages_correlation ON email_messages(correlation_id);
CREATE INDEX idx_email_messages_status ON email_messages(status);

CREATE TABLE IF NOT EXISTS email_provider_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_message_id UUID REFERENCES email_messages(id),
    provider VARCHAR(50) NOT NULL,
    request_payload JSONB,
    response_payload JSONB,
    status_code INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
