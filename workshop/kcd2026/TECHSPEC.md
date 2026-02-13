# TECHSPEC: Leveraging Day-2 DB Operations with Temporal & Postgres 18

## 1. Overview

This workshop module demonstrates how to leverage **Temporal** workflows to orchestrate complex "Day-2" database operations for **Postgres 18**, ensuring reliability, zero downtime, and safety. We will use a suite of modern Go-based tools—**pgroll**, **sqldef**, and **pgstream**—to handle schema migrations, idempotency, and replication.

**Goal:** Build a resilient system where database maintenance, scaling, and failover are handled as durable, code-defined workflows.

## 2. Technology Stack

- **Language:** Go v1.25.x (using `slog`, generics, `synctest`)
- **Orchestration:** Temporal (Go SDK)
- **Database:** Postgres 18 (leveraging `uuidv7`, `AIO`, `logical replication`)
- **DB Driver:** `pgx/v5` (or latest)
- **Tools:**
  - **pgroll:** Zero-downtime schema migrations (expand/contract).
  - **sqldef:** Idempotent declarative schema management.
  - **pgstream:** Postgres-to-Postgres replication (CDC).
  - **pg0:** Call [pg0](https://github.com/vectorize-io/pg0) using the Go script library

## 3. Architecture

The system consists of a **Temporal Worker** that manages the lifecycle of Postgres instances and executes operations against them.

### Components
1.  **Controller (Temporal Worker):** Runs workflows to provision DBs, apply migrations, and manage replication.
2.  **pg0 (Primary DB):** The initial standalone Postgres 18 instance.
3.  **pg-replica-N (Read Replicas):** Additional instances synced via `pgstream`.
4.  **App (Workload Generator):** A simple Go API that writes to Primary and reads from Replicas.

## 4. Scenarios & Demos

### 4.0. Setup: The `pg0` Library
**Objective:** Create a reusable Go library (`pkg/pg0`) to spin up a standalone Postgres 18 instance programmatically.
- **Implementation:**
  - Call pg0 binary using Go script library; review deply [pg0](https://github.com/vectorize-io/pg0)
  - **Config:** Enable `wal_level = logical` (required for `pgroll` and `pgstream`).
  - **Output:** Connection string, PID, and lifecycle hooks (Start/Stop).

### 4.1. Demo #1: Schema Management Two Ways

#### Demo #1a: Zero-Downtime Migration with `pgroll`
**Workflow:** `MigrateWithPgrollWorkflow`
1.  **Start:** Workflow receives a schema change request (e.g., "add column `email` to `users`").
2.  **Init:** `pgroll init` to set up internal tables.
3.  **Start Migration:** `pgroll start <file>` (creates new schema version, backfills data).
4.  **Verification:** Workflow queries both "old" and "new" schema versions to prove dual accessibility.
5.  **Complete:** `pgroll complete` (switches traffic to new version).
**Key Takeaway:** Safe, reversible migrations without locking.

#### Demo #1b: Idempotent Setup with `sqldef`
**Workflow:** `EnsureSchemaWorkflow`
1.  **Input:** A raw SQL file defining the *desired* state.
2.  **Execution:** Run `psqldef` (dry-run first, then apply).
3.  **Idempotency:** Run the workflow twice; the second run should result in no-op.
**Key Takeaway:** GitOps-friendly schema management.

### 4.2. Demo #2: Read Replica with `pgstream`
**Workflow:** `ProvisionReplicaWorkflow`
1.  **Provision:** Spin up a second empty Postgres instance (`pg-replica-1`) using `pg0` library.
2.  **Stream Init:** Start `pgstream` producer on Primary and consumer on Replica.
3.  **Snapshot:** `pgstream` takes initial snapshot of Primary.
4.  **Sync:** Continuous CDC streaming ensures Replica matches Primary.
**Key Takeaway:** Fast, granular replication setup without complex WAL archiving.

### 4.3. Demo #3: Read/Write Splitting & Scaling
**Workflow:** `ScaleReadReplicasWorkflow`
1.  **App Logic:**
    - **Writes:** Direct to Primary (`pgx` pool).
    - **Reads:** Round-robin load balancing across registered Replicas.
2.  **Scaling:**
    - Trigger workflow to add `pg-replica-2`.
    - Once `pgstream` reports "synced", add new replica URL to the App's read-pool configuration (dynamically updated).
**Key Takeaway:** Dynamic infrastructure scaling orchestrated by code.

### 4.4. Demo #4: Failover Orchestration (No Data Loss)
**Workflow:** `FailoverWorkflow`
1.  **Scenario:** Primary (`pg0`) is flagged for maintenance or simulates a crash.
2.  **Pause Writes:** App enters "buffer mode" (queues writes in memory or Temporal signals).
3.  **Catchup:** Ensure `pg-replica-1` has processed all stream events.
4.  **Promote:** Reconfigure `pg-replica-1` to be the new Primary.
5.  **Redirect:** Update App config to point writes to new Primary.
6.  **Resume:** App flushes buffered writes.
**Key Takeaway:** Orchestrated failover ensures consistency where manual intervention often fails.

### 4.5. Demo #5: Maintenance & Read-Only Modes
**Workflow:** `MaintenanceModeWorkflow`
1.  **Signal:** Admin triggers "Maintenance Mode".
2.  **Action:**
    - App rejects write requests (HTTP 503 or custom error).
    - Read requests continue to be served from Replicas.
3.  **Operation:** Perform heavy maintenance (e.g., `VACUUM FULL`, OS upgrade) on Primary.
4.  **Resume:** Signal "Resume"; App accepts writes again.
**Key Takeaway:** Graceful degradation of service controlled by workflow state.

## 5. Implementation Plan

### Directory Structure

```
workshop/kcd2026/
├── pkg/
│ ├── pg0/ # Library to spin up PG instances
│ ├── workload/ # The sample Go API (App)
│ └── workflows/ # Temporal workflow definitions
├── cmd/
│ ├── worker/ # Temporal Worker main
│ ├── starter/ # CLI to trigger workflows
│ └── app/ # The sample App main
├── migrations/ # SQL files for sqldef/pgroll
└── configs/ # Postgres configs
```


### Key Go Packages
- **`github.com/jackc/pgx/v5`**: For all DB interactions.
- **`go.temporal.io/sdk`**: For workflows/activities.
- **`github.com/xataio/pgroll`**: Embedded or exec'd binary.
- **`github.com/xataio/pgstream`**: Embedded or exec'd binary.
- **`github.com/k0kubun/sqldef`**: Embedded or exec'd binary.

### Next Steps
1.  Implement `pkg/pg0` to verify Postgres 18 startup.
2.  Create basic Temporal Worker shell.
3.  Implement Demo #1a & #1b activities.
4.  Build the `pgstream` integration for Demo #2.

### Future Roadmap

- Complex traffic shaping using pgdog