-- Add up migration script here
alter table cart_items add column status varchar(255) not null default 'Not_ordered';

