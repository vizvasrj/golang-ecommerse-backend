CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    image_url TEXT,
    image_key TEXT,
    description TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    taxable BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    brand_id UUID REFERENCES brands(id) ON DELETE SET NULL,  -- Set brand_id to NULL if the referenced brand is deleted
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,  -- Set category_id to NULL if the referenced category is deleted
    merchant_id UUID REFERENCES merchants(id) ON DELETE CASCADE,  -- Deletes product if the referenced merchant is deleted
    updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
