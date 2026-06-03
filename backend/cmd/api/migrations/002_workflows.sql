CREATE TABLE IF NOT EXISTS workflows (
    id            SERIAL PRIMARY KEY,
    name          VARCHAR(100) UNIQUE NOT NULL,
    target_schema VARCHAR(100) NOT NULL,
    is_active     BOOLEAN DEFAULT TRUE,
    condition     JSONB NOT NULL,
    actions       JSONB NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);