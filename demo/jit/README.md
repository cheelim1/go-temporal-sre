# JIT Access Demo for MongoDB Atlas

This project demonstrates Just-In-Time (JIT) access management for MongoDB Atlas using Go, Temporal, and a simple HTTP API.

## Environment Variables

Set the following environment variables before running the project:
- `ATLAS_PUBLIC_KEY`  – Your MongoDB Atlas API public key.
- `ATLAS_PRIVATE_KEY` – Your MongoDB Atlas API private key.
- `ATLAS_PROJECT_ID`  – Your MongoDB Atlas Project (Group) ID.
- `TEMPORAL_HOST`     – Temporal server host (default: `localhost:7233`).
- `TEMPORAL_NAMESPACE`– Temporal namespace (default: `default`).
- `PORT`              – HTTP server port (default: `8080`).

### Ref
- https://learn.temporal.io/getting_started/go/dev_environment/
- https://pkg.go.dev/github.com/mongodb/atlas-sdk-go