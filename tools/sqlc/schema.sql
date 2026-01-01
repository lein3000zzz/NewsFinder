CREATE TABLE sources (
    id UUID PRIMARY KEY,

    name VARCHAR(255) NOT NULL UNIQUE,

    credibility REAL NOT NULL CHECK (credibility >= 0 AND credibility <= 1),

    active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE news (
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

