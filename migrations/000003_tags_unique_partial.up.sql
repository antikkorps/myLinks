ALTER TABLE tags DROP CONSTRAINT tags_user_id_name_key;

CREATE UNIQUE INDEX tags_user_id_name_active_idx
    ON tags (user_id, name)
    WHERE deleted_at IS NULL;