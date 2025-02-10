CREATE TABLE merch (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    price INT NOT NULL
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    balance INT DEFAULT 0
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    sender_id INT NOT NULL REFERENCES users(id),
    receiver_id INT REFERENCES users(id),
    amount INT NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('purchase', 'transfer')),
    item VARCHAR(255),
    created_at TIMESTAMP DEFAULT now()
);
