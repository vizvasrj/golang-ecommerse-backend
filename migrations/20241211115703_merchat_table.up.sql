-- Add up migration script here
CREATE TYPE merchant_status AS ENUM ('Waiting Approval', 'Rejected', 'Approved');

CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    brand_name VARCHAR(255) NOT NULL,
    business TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    status merchant_status NOT NULL,
    updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
