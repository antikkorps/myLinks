-- =============================================================================
-- Migration 4 — Trigram search on links (and tag names)
-- =============================================================================
-- pg_trgm enables GIN-indexed substring search (ILIKE '%q%'), which is the
-- right fit for an as-you-type search box: partial words, typo tolerance,
-- no whole-word/lexeme semantics (unlike tsvector full-text search).
--
-- Note: a trigram index only kicks in for patterns of length >= 3. Shorter
-- queries fall back to a sequential scan — correct, just not index-accelerated.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_links_title_trgm
    ON links USING gin (title gin_trgm_ops);

CREATE INDEX idx_links_url_trgm
    ON links USING gin (url gin_trgm_ops);

CREATE INDEX idx_links_description_trgm
    ON links USING gin (description gin_trgm_ops);

CREATE INDEX idx_tags_name_trgm
    ON tags USING gin (name gin_trgm_ops);
