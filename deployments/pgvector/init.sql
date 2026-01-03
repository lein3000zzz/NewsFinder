CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS sources (
    id UUID PRIMARY KEY,

    name VARCHAR(255) NOT NULL UNIQUE,

    credibility REAL NOT NULL CHECK (credibility >= 0 AND credibility <= 1),

    active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS news (
    id UUID PRIMARY KEY,

    source_id UUID NOT NULL REFERENCES sources(id),

    title TEXT,
    content TEXT NOT NULL,

    published_at TIMESTAMPTZ NOT NULL,
    ingested_at TIMESTAMPTZ NOT NULL,

    content_hash TEXT NOT NULL,
    analysis JSONB NOT NULL,
    content_embedding VECTOR(384),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_news_content_hash ON news(content_hash);

CREATE INDEX IF NOT EXISTS idx_news_embedding_hnsw ON news
    USING hnsw (content_embedding vector_cosine_ops);

CREATE INDEX IF NOT EXISTS idx_news_published_at ON news(published_at DESC);

CREATE INDEX IF NOT EXISTS idx_news_source_id ON news(source_id);