# README

## Objective

Rethink use of Kafka by leveraging a fast real-time buffer layer using VictoriaLogs; 
and the rest using Iceberg 

Inspiration; see Netflix latest setup of using ClickHouse for hot buffer
 and Iceberg for everything else.

## Approach

Input as a structured log; will get columnar + timebased indexing for FREE.

All events will start with a event time component. Can be leveraged for real life business
case that needs this capabilities (e.g. like fraud).  Once the data reaches the S3-like
bucket (S3, R2, Tigris); the Iceberg layer will be a good first cut for further transformation
down the line; lineage etc.

After that; transformation to the Iceberg layer can be done at 1 minute interval 
using something like Temporal (or any other durable execution; e.g. Cloudflare Workflows)

Further layers can be larger micro-bacth at the 15 mins level.

## Questions

To replace Kafka-like; a faster interval of 1-5s might be expected but might be harder.
Will the AutoMQ release of their buffer layer implementation be a better one?

## Limits

Temporal has limits --> https://docs.temporal.io/cloud/limits#:~:text=Per%20message%20gRPC%20limit%E2%80%8B&text=Each%20gRPC%20message%20received%20has%20a%20limit,all%20gRPC%20endpoints%20across%20the%20Temporal%20Platform.

Primarily the gRPC messages have ti fit weithin 4MB; so there should be validations.

Teams are know to usually try to abuse and be thoughtless; logging huge blbs of data.
This is esp. worse in model like info; where there .  These should be sent to blob storage


Generally anything >=1MB should be flagged and tagged as a guard rail and should have high bar
to override the limit.


