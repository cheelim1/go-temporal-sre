# RESEARCH

## AutoMQ

- Has Iceberg + scehma ... might be intersting ..
- For those with 1s level sync .. must be justified?
- Uses EBS as the WAL + buffer?

## Bufstrea

- Stateless Broker per AZ
- Iceberg native support; even S3 Tables??
- Handles 100 GB/s of writes when using Spanner instead of etcd --> https://buf.build/blog/bufstream-on-spanner
- Automatic schema management for Iceberg ..

## VictoriaLogs

- Still can be fed from AutoMQ; is it worth it?
- dcd

## Estuary + Gazette

- Gazette Broker - https://gazette.readthedocs.io/en/latest/brokers-concepts.html
- Eatuary Flow - https://github.com/estuary/flow

## Transformation

- Cloudflare PIpeline ..
- dlt
- hamilton

## Lineage

- SQLMesh --> https://estuary.dev/blog/dbt-alternatives/