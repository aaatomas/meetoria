CREATE TABLE IF NOT EXISTS sms_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL UNIQUE,
    correlation_id UUID NOT NULL,
    organization_id UUID,
    booking_id UUID,
    recipient_phone VARCHAR(20) NOT NULL,
    template VARCHAR(100) NOT NULL,
    variables JSONB DEFAULT '{}',
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

CREATE INDEX idx_sms_messages_correlation ON sms_messages(correlation_id);
CREATE INDEX idx_sms_messages_status ON sms_messages(status);

CREATE TABLE IF NOT EXISTS sms_provider_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sms_message_id UUID REFERENCES sms_messages(id),
    provider VARCHAR(50) NOT NULL,
    request_payload JSONB,
    response_payload JSONB,
    status_code INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
