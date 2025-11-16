-- 001_create_sales_table.up.sql
CREATE TABLE IF NOT EXISTS sales (
                                     id BIGSERIAL PRIMARY KEY,
                                     kind TEXT NOT NULL CHECK (kind IN ('income','expense')), -- тип записи
                                     amount NUMERIC(18,2) NOT NULL CHECK (amount >= 0),        -- абсолютная сумма
                                     category TEXT,
                                     note TEXT,
                                     created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sales_created_at ON sales (created_at);
CREATE INDEX IF NOT EXISTS idx_sales_category ON sales (category);
