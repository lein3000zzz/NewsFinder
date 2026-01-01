-- TODO: Пример, взят из гайда, сделать под себя. https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html
-- name: GetSources :many
SELECT * FROM sources;

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
)
SELECT id, source_id, title, content, published_at, ingested_at, content_hash, analysis, content_embedding
FROM json_populate_record(NULL::news, $1::jsonb)
    ON CONFLICT (content_hash) DO NOTHING
RETURNING id;

-- name: LookupByHash :one
SELECT 1
FROM news
WHERE content_hash = $1
LIMIT 1;