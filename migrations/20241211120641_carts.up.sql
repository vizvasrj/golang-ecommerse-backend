-- Add up migration script here
CREATE TABLE carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,  -- Cascade delete
    updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
