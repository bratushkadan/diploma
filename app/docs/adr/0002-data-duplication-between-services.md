# 2. Data duplication between services

Date: 2025-02-21

## Status

Accepted

## Context

There are three patterns I managed to discern: 1) write data to message broker transactionally with the writing data to the DMBS - more control, less coupling, more expensive compute-wise and more expensive developer time-wise; 2) use CDC feature to load changefeed to a topic with messages and subscribe on CDC updates (Debezium), handling them on the downstream application side - less control, less coupling, less expensive; 3) asynchronous replication - WAL is replicated across upstream and downstream DMBSs, leaving me with single database across services limitation, but saving precious development time.

## Decision

I will use *asynchronous replication* for services that *have a possibility* of using YDB for the persistence - YDB is a great DBMS and I am already too coupled to the Yandex Cloud infrastructure, there's no further need to reduce the infrastructure coupling. *Asynchronous replication* incurs no other costs for me, it accelerated the development speed.

I will use *CDC with the downstream service Debezium format messages processing* in the Catalog service (the only service that will use OpenSearch instead of YDB), and any other service that will not leverage YDB DMBS.

## Consequences

Asynchronous Replication saves development time, but increases coupling on YDB. *CDC downstream service Debezium format messages processing* may lead to data loss/insonsistencies (e.g. update and delete events are processed by different processing units, leading to a data race or possibly corruping OpenSearch database state), but is a necessary technological approach.
