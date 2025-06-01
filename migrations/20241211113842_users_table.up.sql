-- Add up migration script here
CREATE TYPE user_role AS ENUM ('ROLE ADMIN', 'ROLE MERCHANT', 'ROLE USER');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone_number VARCHAR(20),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    password VARCHAR(255),
    provider VARCHAR(255),
    google_id VARCHAR(255),
    facebook_id VARCHAR(255),
    avatar VARCHAR(255),
    role user_role NOT NULL,
    reset_password_token VARCHAR(255),
    reset_password_expires TIMESTAMP WITH TIME ZONE,
    updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
