-- =============================================================================
-- Migration 2 — Auth, folders, tags + extension of the links table
-- =============================================================================

-- -----------------------------------------------------------------------------
-- USERS
-- -----------------------------------------------------------------------------
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email       TEXT NOT NULL,
    last_name   TEXT NOT NULL,
    first_name  TEXT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP
);

-- Email uniqueness:
--   - case-insensitive (LOWER)
--   - partial: does not block re-registration after a soft-delete
CREATE UNIQUE INDEX idx_users_email_lower
    ON users (LOWER(email))
    WHERE deleted_at IS NULL;

-- -----------------------------------------------------------------------------
-- ACCOUNTS — authentication methods for a user (1 user → N accounts)
-- -----------------------------------------------------------------------------
CREATE TABLE accounts (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider          TEXT NOT NULL,            -- 'password' | 'github' | 'google' | ...
    provider_user_id  TEXT NOT NULL,            -- email (password) or external ID (OAuth)
    password_hash     TEXT,                     -- NULL for OAuth, NOT NULL for password
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- A given (provider, provider_user_id) pair can only be linked to one user.
    -- Prevents the same GitHub account from being linked to two different users.
    UNIQUE (provider, provider_user_id),

    -- Consistency: password_hash present IFF provider='password'
    CONSTRAINT accounts_password_hash_consistency CHECK (
        (provider = 'password' AND password_hash IS NOT NULL)
        OR
        (provider <> 'password' AND password_hash IS NULL)
    )
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);

-- -----------------------------------------------------------------------------
-- REFRESH TOKENS — one row per active session
-- -----------------------------------------------------------------------------
CREATE TABLE refresh_tokens (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- We store the SHA-256 hash of the token, never the raw token.
    -- If the DB leaks, the hashes are unusable.
    token_hash    TEXT NOT NULL UNIQUE,

    expires_at    TIMESTAMP NOT NULL,
    revoked_at    TIMESTAMP,                    -- NULL = active

    -- Device info for the "connected devices" screen and targeted revocation
    user_agent    TEXT,
    ip            INET,                         -- native Postgres type for IPv4/IPv6
    last_used_at  TIMESTAMP,

    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- -----------------------------------------------------------------------------
-- FOLDERS — link organization (flat for V1)
-- -----------------------------------------------------------------------------
CREATE TABLE folders (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP
);

CREATE INDEX idx_folders_user_id ON folders(user_id);

-- -----------------------------------------------------------------------------
-- TAGS — scoped per user (a personal tag is not shared with others)
-- -----------------------------------------------------------------------------
CREATE TABLE tags (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP,

    -- A user cannot have two tags with the same name
    UNIQUE (user_id, name)
);

-- -----------------------------------------------------------------------------
-- LINKS — extending the existing table (created in migration 1)
-- -----------------------------------------------------------------------------
ALTER TABLE links
    ADD COLUMN user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ADD COLUMN folder_id  UUID REFERENCES folders(id) ON DELETE SET NULL,
    ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN deleted_at TIMESTAMP;

CREATE INDEX idx_links_user_id ON links(user_id);
CREATE INDEX idx_links_folder_id ON links(folder_id);

-- -----------------------------------------------------------------------------
-- LINK_TAGS — many-to-many pivot
-- -----------------------------------------------------------------------------
CREATE TABLE link_tags (
    link_id  UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    tag_id   UUID NOT NULL REFERENCES tags(id)  ON DELETE CASCADE,
    PRIMARY KEY (link_id, tag_id)
);

CREATE INDEX idx_link_tags_tag_id ON link_tags(tag_id);
-- (no need for an index on link_id: the composite PK (link_id, tag_id) already covers it)
