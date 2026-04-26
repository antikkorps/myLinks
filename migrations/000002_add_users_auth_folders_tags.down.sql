-- Rollback for migration 2.
-- Reverse order of creation, respecting FK dependencies.

DROP TABLE IF EXISTS link_tags;

-- Remove the columns added to links
DROP INDEX IF EXISTS idx_links_folder_id;
DROP INDEX IF EXISTS idx_links_user_id;
ALTER TABLE links
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS folder_id,
    DROP COLUMN IF EXISTS user_id;

DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS folders;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS users;
