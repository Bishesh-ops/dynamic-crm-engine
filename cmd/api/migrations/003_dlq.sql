CREATE TABLE IF NOT EXISTS dead_letter_queue (
    id SEREAL PRIMARY KEY,
    schema_id INT NOT NULL,
    schema_name VARCHAR(100),
    payload JSONB NOT NULL,
    error_reason TEXT NOT NULL,
    failed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_dlq_unresolved ON dead_letter_queue(resolved) WHERE resolved = FALSE;
