CREATE DATABASE IF NOT EXISTS auction_db
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'auction_user'@'%'
    IDENTIFIED BY 'auction_password';

GRANT ALL PRIVILEGES ON auction_db.* TO 'auction_user'@'%';

FLUSH PRIVILEGES;
