CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    origin_url text NOT NULL,
    short_url text,
    custom_url text UNIQUE
    );
