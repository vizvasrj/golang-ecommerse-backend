-- Add down migration script here
alter table products
    add column category_id UUID REFERENCES categories(id) ON DELETE SET NULL;

drop table product_categories;