-- Add up migration script here
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID REFERENCES carts(id) ON DELETE SET NULL, -- Allow cart to be deleted even if order exists. Consider your requirements
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,  -- Cascade delete
    address_id UUID NOT NULL REFERENCES addresses(id) ON DELETE CASCADE, -- Cascade delete
    total NUMERIC(10, 2) NOT NULL,
    updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);