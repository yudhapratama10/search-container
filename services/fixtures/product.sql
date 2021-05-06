CREATE TABLE product(
    id INTEGER NOT NULL, 
    name CHARACTER VARYING(255) NOT NULL, 
    price INTEGER NOT NULL
);

CREATE SEQUENCE product_id_seq AS INTEGER START WITH 1 INCREMENT BY 1;

ALTER TABLE ONLY product ALTER COLUMN id SET DEFAULT nextval('product_id_seq'::regclass);

ALTER TABLE ONLY product ADD CONSTRAINT product_pkey PRIMARY KEY (id);

INSERT INTO product (name, price) VALUES
('samsung galaxy note s9', 10540000),
('samsung galaxy note s9 2', 10550000),
('samsung galaxy note s9 3', 10560000),
('amd ryzen 9 3900xt official', 6700000),
('amd ryzen 9 3900xt official 2', 6800000),
('amd ryzen 9 3900xt official 3', 6900000),
('amd ryzen 9 3950xt official', 1150000),
('amd ryzen 9 3950xt official 2', 11600000),
('amd ryzen 9 3950xt official 3', 11700000),
('nvidia rtx 3080 official', 17500000),
('nvidia rtx 3080 official 2', 17600000),
('nvidia rtx 3080 official 3', 17700000),
('nvidia rtx 3070 official', 13500000),
('nvidia rtx 3070 official 2', 13600000),
('nvidia rtx 3070 official 3', 13700000),
('philips air fryer HD9252', 1597000),
('philips air fryer HD9252 2', 1598000),
('philips air fryer HD9252 3', 1599000),
('bakeware loaf pan non-stick', 81400),
('bakeware loaf pan non-stick 2', 81500),
('bakeware loaf pan non-stick 3', 81600),
('cheese keju cheddar prochiz 2KG', 172700),
('cheese keju cheddar prochiz 2KG 2', 172800),
('cheese keju cheddar prochiz 2KG 3', 172900),
('cleo kitten food 1,2 kg', 65300),
('cleo kitten food 1,2 kg 2', 65400),
('cleo kitten food 1,2 kg 3', 65500);
