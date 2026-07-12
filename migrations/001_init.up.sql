CREATE TABLE IF NOT EXISTS products (
 id CHAR(36) NOT NULL, seller_id CHAR(36) NOT NULL, title VARCHAR(200) NOT NULL, description TEXT NOT NULL, image_key TEXT NOT NULL,
 starting_price DECIMAL(12,2) NOT NULL, current_price DECIMAL(12,2) NOT NULL, highest_bidder_id CHAR(36) NULL,
 auction_end_at DATETIME(6) NOT NULL, status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
 PRIMARY KEY (id), CONSTRAINT chk_products_starting_price CHECK (starting_price > 0), CONSTRAINT chk_products_current_price CHECK (current_price >= starting_price),
 CONSTRAINT chk_products_status CHECK (status IN ('ACTIVE','EXPIRED','SOLD','CANCELLED')), INDEX idx_products_status_end_at (status, auction_end_at), INDEX idx_products_seller_created_at (seller_id, created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS bids (
 id CHAR(36) NOT NULL, product_id CHAR(36) NOT NULL, bidder_id CHAR(36) NOT NULL, amount DECIMAL(12,2) NOT NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 PRIMARY KEY (id), CONSTRAINT fk_bids_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE, CONSTRAINT chk_bids_amount CHECK (amount > 0),
 INDEX idx_bids_product_amount_created_at (product_id, amount DESC, created_at DESC), INDEX idx_bids_bidder_created_at (bidder_id, created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS sales (
 id CHAR(36) NOT NULL, product_id CHAR(36) NOT NULL, seller_id CHAR(36) NOT NULL, buyer_id CHAR(36) NOT NULL, final_price DECIMAL(12,2) NOT NULL, sold_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 PRIMARY KEY (id), UNIQUE KEY uq_sales_product (product_id), CONSTRAINT fk_sales_product FOREIGN KEY (product_id) REFERENCES products(id), CONSTRAINT chk_sales_final_price CHECK (final_price > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
