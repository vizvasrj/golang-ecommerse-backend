-- Add up migration script here
CREATE TABLE receipts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE, -- Cascade delete
    amount NUMERIC(10, 2) NOT NULL,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    payment_provider TEXT,
    provider_data JSONB,
    payment_status VARCHAR(20) NOT NULL
);
