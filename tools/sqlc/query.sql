-- name: GetSourceByID :one
SELECT *
    FROM sources
    WHERE id = $1
LIMIT 1;

-- name: AddNews :one
INSERT INTO news (
    id,
    source_id,
    title,
    content,
    published_at,
    ingested_at,
    content_hash,
    analysis,
    content_embedding
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    ON CONFLICT (content_hash) DO NOTHING
RETURNING id;

-- name: LookupByHash :one
SELECT 1
    FROM news
    WHERE content_hash = $1
LIMIT 1;

-- name: LookupEmbedding :one
SELECT EXISTS (
    SELECT 1 FROM news
    WHERE content_embedding <=> $1 < 0.15
    AND published_at > now() - interval '24 hours'
    LIMIT 1
);