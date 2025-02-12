CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    coins INT NOT NULL DEFAULT 1000,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    item_type VARCHAR(255) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
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
    price INT NOT NULL CHECK (price > 0),
    stock INT NOT NULL DEFAULT 0 CHECK (stock >= 0)
);

-- CREATE INDEX idx_inventory_user ON inventory(user_id);
-- CREATE INDEX idx_transactions_from ON coin_transactions(from_user_id);
-- CREATE INDEX idx_transactions_to ON coin_transactions(to_user_id);