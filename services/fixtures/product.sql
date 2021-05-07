CREATE TABLE recipes (
    id serial PRIMARY KEY,
    name text not null,
    ingredients text[] not null,
    isHalal boolean not null,
    isVegetarian boolean not null,
    description text not null,
    rating float not null
);