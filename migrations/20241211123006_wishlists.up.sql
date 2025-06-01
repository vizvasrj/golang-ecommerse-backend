-- Add up migration script here
CREATE TABLE wishlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE, -- Cascade delete
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,  -- Cascade delete
    is_liked BOOLEAN,
    updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

