# Dynamic CRM Engine

A high-performance, dynamic schema CRM core engine built in Go. This engine utilizes an asynchronous, thread-safe concurrent worker pool to batch-process event ingestions and translates dynamic JSON queries directly into highly optimized PostgreSQL `JSONB` operations utilizing `GIN` index containment.

## Architectural Highlights

* **Decoupled Write Pipeline:** Incoming payloads are validated in-memory against their dynamic definitions before being offloaded to a concurrent channel-backed bus.
* **Automated Batching:** Workers drain the ingestion queue and flush payloads to PostgreSQL using optimized batch routines (`pgx.Batch`), minimizing database connection overhead.
* **Deterministic Backpressure Processing:** Includes queue capacity validation, yielding instant HTTP `503 Service Unavailable` signals when processing limits are met to prevent cascading worker-pool depletion.
* **Dynamic Read Path Translation:** A custom JSON-to-SQL AST compiler translates structural object logic directly into native PostgreSQL `@>` containment and numerical string extraction operators.
* **High Performance Indexes:** Designed natively to execute dynamic filtering within O(log N) time complexities by matching query bounds against a pre-allocated structural `GIN` index.

---

## Directory Structure

```text
├── cmd
│   └── api
│       ├── main.go            # Dependency injection, setup, and HTTP routing
│       └── migrations
│           └── 001_init.sql   # Core relational database schemas
├── internal
│   ├── database
│   │   ├── db.go              # PostgreSQL pool connection configurations
│   │   ├── entity_repo.go     # Target data store mutations & execution pipelines
│   │   └── schema_repo.go     # Structural type definitions storage
│   ├── eventbus
│   │   └── bus.go             # Background batch workers and backpressure mechanisms
│   ├── query
│   │   └── builder.go         # Dynamic JSON-to-SQL AST compiler logic
│   ├── response
│   │   └── json.go            # Consistent JSON messaging & error rendering package
│   ├── schema
│   │   └── validator.go       # In-memory structural data layout enforcement
│   └── workflow
│       ├── evaluator.go       # Rule criteria AST parser core
│       └── mutator.go         # Extensible context mutations engine
├── Makefile                   # Project automated task definitions
├── go.mod                     # Go dependencies management manifest
└── docker-compose.yml         # Containerized database backing services mapping
```
## Core Schema Layout
The engine enforces strong typing at the system boundary while maintaining absolute schema flexibility inside the application database storage layer:
```sql
CREATE TABLE IF NOT EXISTS schemas (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    structure JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_id INT REFERENCES schemas(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Core optimization index enabling sub-millisecond JSON structural containment queries
CREATE INDEX IF NOT EXISTS idx_entities_data ON entities USING GIN(data);
```
## Getting Started
**Prerequisites**
* Go 1.25.10

* Docker & Docker Compose

**Operational Commands:**
```bash
# 1. Spin up the containerized database backing environment
make db-up

# 2. Execute the comprehensive unit & integration test suites
make test

# 3. Compile and initialize the localized server instance
make run
```
## Technical API Contracts
<h3>1. Register a New Custom Data Layout</h3>

* Endpoint: `POST /schemas`

* Payload Type: `application/json`

<h4>Example Request Body:</h4>

```JSON
{
  "name": "leads",
  "fields": {
    "company": { "type": "string", "required": true },
    "score": { "type": "int", "required": false },
    "profile": {
      "type": "object",
      "required": true,
      "properties": {
        "status": { "type": "string", "required": true },
        "employee_count": { "type": "int", "required": false }
      }
    }
  }
}
```
<h3>2. Ingest Data Instances (Write Pipeline)</h3>

* Endpoint: `POST /entities/leads`

* Processing Strategy: Asynchronous thread-safe ingest backing up to a buffer max boundary of 500 entries per batch before execution.

* Payload Type: `application/json`

<h4>Example Request Body:</h4>

```JSON
{
  "company": "Acme Corp",
  "score": 95,
  "profile": {
    "status": "active",
    "employee_count": 120
  }
}
```
<h3>3. Query Structural Document Data (Read Pipeline)</h3>

* Endpoint: `POST /query`

* Execution Profile: Compiles payload arrays directly into optimized parameterized index bounds queries utilizing deep path text operators (`->>`) and index node matchers (`@>`).

* Payload Type: `application/json`

<h4>Example Request Body:</h4>

```JSON
{
  "schema_id": 1,
  "conditions": [
    { "field": "profile.status", "op": "==", "value": "active" },
    { "field": "score", "op": ">", "value": 90 }
  ],
  "limit": 10,
  "offset": 0
}
```
