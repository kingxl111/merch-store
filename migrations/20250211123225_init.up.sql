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

CREATE TABLE inventories (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users.user_id,
    item INT
)

CREATE TABLE transactions (
                              id SERIAL PRIMARY KEY,
                              sender_id INT NOT NULL REFERENCES users(id),
                              receiver_id INT REFERENCES users(id),
                              amount INT NOT NULL,
                              type VARCHAR(50) NOT NULL CHECK (type IN ('purchase', 'transfer')),
                              item VARCHAR(255),
                              created_at TIMESTAMP DEFAULT now()
);

INSERT INTO merch (name, price) VALUES
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
