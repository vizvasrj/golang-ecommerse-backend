-- Add up migration script here
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    content_type VARCHAR(255),  -- Use varchar
    description TEXT NOT NULL,  -- Could be longer
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);