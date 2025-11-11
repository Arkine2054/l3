CREATE TABLE IF NOT EXISTS notifications (
                                             id BIGSERIAL PRIMARY KEY,
                                             recipient TEXT NOT NULL,
                                             channel TEXT NOT NULL,
                                             message TEXT NOT NULL,
                                             send_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                             status TEXT NOT NULL DEFAULT 'scheduled',
                                             attempts INT NOT NULL DEFAULT 0,
                                             last_error TEXT,
                                             created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
