-- Add up migration script here
alter table products
    drop column category_id; 

create table product_categories (
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    primary key (product_id, category_id)
);
