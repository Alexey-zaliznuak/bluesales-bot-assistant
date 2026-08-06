ALTER TABLE kb_snapshots
    DROP COLUMN IF EXISTS cache_key,
    DROP COLUMN IF EXISTS warmed_at,
    DROP COLUMN IF EXISTS warm_error,
    DROP COLUMN IF EXISTS prompt_tokens,
    DROP COLUMN IF EXISTS cached_tokens;
