# SCENARIO

Having 1k concurrency input of random size; reject lager than 4k size?

## Client cut off

While submitting; it gets terminated the node ..

## Broker cut-off 

While receving; the node gets removed without acknowledgement ..


## CDC of Database

- This senario should be discouraged as events should only represent the Ubiquitous Language (UL)
- Possibly have the before and after data; just liek seen in DynamoDB stream ..

## Business Domain UL

- The data extracted out via Outbox pattern
- Should have the full 
- MUST never drop it; as it is important domain business context being represented
- MUST have the exact time of when the event gets triggerred; not when it is published, consumerd etc 
- 

## Client side events (button push, interaction)

- Volume is much more; can have much more fsilure and should be non-blocking ..
- Traditionally, this is managed via Segment
- Should be fire-forget ..
- Buffered; if have dropped data is not end of world; for higher volume?
- Can choose but must acknowledge paying more for this durability; MUST be jsutified

## Fast WAL using Shared Storage

- https://www.automq.com/docs/automq/architecture/overview#shared-storage-architecture