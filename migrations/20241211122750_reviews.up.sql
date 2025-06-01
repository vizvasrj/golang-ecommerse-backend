-- Add up migration script here
create type review_status as enum ('Rejected', 'Approved', 'Waiting Approval');

CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,  -- Cascade delete
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Cascade delete
    title TEXT NOT NULL,
    rating NUMERIC(3,2) NOT NULL,
    review TEXT NOT NULL,
    is_recommended BOOLEAN NOT NULL,
    status review_status NOT NULL,
    updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
