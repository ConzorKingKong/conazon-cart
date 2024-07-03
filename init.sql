CREATE SCHEMA cart;
CREATE TABLE cart.cart (
  id SERIAL PRIMARY KEY NOT NULL,
  created_at TIMESTAMP DEFAULT NOW() NOT NULL,
  updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
  user_id INT NOT NULL,
  product_id INT NOT NULL,
  quantity INT DEFAULT 1 NOT NULL,
  status VARCHAR(255) DEFAULT 'active' NOT NULL
);
CREATE UNIQUE INDEX idx_user_product ON cart.cart (user_id, product_id) WHERE status = 'active';
insert into cart.cart (user_id, product_id, quantity, status) values (1, 1, 1, 'active');
insert into cart.cart (user_id, product_id, quantity, status) values (1, 2, 1, 'active');
insert into cart.cart (user_id, product_id, quantity, status) values (1, 3, 1, 'active');