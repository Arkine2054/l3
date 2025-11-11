-- migrations/0001_init.up.sql

CREATE TABLE IF NOT EXISTS events (
                                      id SERIAL PRIMARY KEY,
                                      title TEXT NOT NULL,
                                      date TIMESTAMP NOT NULL,
                                      total_seats INT NOT NULL,
                                      created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bookings (
                                        id SERIAL PRIMARY KEY,
                                        event_id INT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
                                        user_name TEXT NOT NULL,
                                        paid BOOLEAN NOT NULL DEFAULT FALSE,
                                        created_at TIMESTAMP NOT NULL DEFAULT now(),
                                        updated_at TIMESTAMP NULL
);

-- Indexes for read performance
CREATE INDEX IF NOT EXISTS idx_bookings_event_paid ON bookings(event_id, paid);
CREATE INDEX IF NOT EXISTS idx_events_date ON events(date);
