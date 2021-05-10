CREATE TABLE recipes (
    id serial PRIMARY KEY,
    name text not null,
    ingredients text[] not null,
    isHalal boolean not null,
    isVegetarian boolean not null,
    description text not null,
    rating float not null
);

INSERT INTO recipes (name, ingredients, isHalal, isVegetarian, description, rating) VALUES
('fried rice', array['rice', 'margarine', 'egg'], true, true, 'fried rice description', 4.0),
('fried rice', array['rice', 'margarine', 'egg'], true, false, 'fried rice description', 4.0),
('fried rice', array['rice', 'margarine', 'egg'], false, false, 'fried rice description', 4.0),
('fried rice 1', array['rice', 'margarine', 'salt'], true, true, 'fried rice description', 4.1),
('fried rice 2', array['rice', 'margarine', 'ketchup'], true, true, 'fried rice description', 4.2);