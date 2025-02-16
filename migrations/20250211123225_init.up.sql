CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    coins INT NOT NULL DEFAULT 1000 CHECK (coins >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    item_type VARCHAR(255) NOT NULL,
    quantity INT NOT NULL,
    UNIQUE(user_id, item_type)
);

CREATE TABLE coin_transactions (
    id SERIAL PRIMARY KEY,
    from_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    to_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    amount INT NOT NULL CHECK (amount > 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE shop_items (
    id SERIAL PRIMARY KEY,
    type VARCHAR(255) UNIQUE NOT NULL,
    price INT NOT NULL CHECK (price > 0)
);

INSERT INTO shop_items (type, price) VALUES
    ('t-shirt', 80),
    ('cup', 20),
    ('book', 50),
    ('pen', 10),
    ('powerbank', 200),
    ('hoody', 300),
    ('umbrella', 200),
    ('socks', 10),
    ('wallet', 50),
    ('pink-hoody', 500);

-- CREATE INDEX idx_inventory_user ON inventory(user_id);
-- CREATE INDEX idx_transactions_from ON coin_transactions(from_user_id);
-- CREATE INDEX idx_transactions_to ON coin_transactions(to_user_id);

-- SELECT pid, age(clock_timestamp(), query_start), state, wait_event, query
-- FROM pg_stat_activity
-- WHERE state != 'idle' AND wait_event IS NOT NULL;