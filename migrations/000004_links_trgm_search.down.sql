DROP INDEX IF EXISTS idx_tags_name_trgm;
DROP INDEX IF EXISTS idx_links_description_trgm;
DROP INDEX IF EXISTS idx_links_url_trgm;
DROP INDEX IF EXISTS idx_links_title_trgm;

DROP EXTENSION IF EXISTS pg_trgm;
