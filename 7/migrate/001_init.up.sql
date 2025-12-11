CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     username TEXT UNIQUE NOT NULL,
                                     password TEXT NOT NULL,
                                     role TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS items (
                                     id SERIAL PRIMARY KEY,
                                     name TEXT NOT NULL,
                                     quantity INT NOT NULL DEFAULT 0,
                                     price NUMERIC(10,2) NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS history (
                                       id SERIAL PRIMARY KEY,
                                       item_id INT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
                                       user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                       action TEXT NOT NULL,
                                       old_data JSONB,
                                       new_data JSONB,
                                       timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users(username, password, role)
VALUES ('admin', '$2a$12$K1CusRQiC/6gTd/PylPIvOOmU29YoabStLjtLsRzYgNgy3re5NqTq', 'admin')
ON CONFLICT (username) DO NOTHING;
-- пароль: admin
