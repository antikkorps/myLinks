CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    url TEXT NOT NULL,
    title TEXT,
    description TEXT,
    image TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_links_url ON links(url);
CREATE INDEX idx_links_created_at ON links(created_at);
