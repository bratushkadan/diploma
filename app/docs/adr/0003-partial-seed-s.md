# 3. Partial SEED(S)

Date: 2025-02-22

## Status

Accepted

## Context

There's a need for an evolutionary project design and documentation framework.

## Decision

Use (partial) SEED(S). All of the features from SEED(S) is welcome for each microservice, except for the excess sequence diagrams for **each** service, which is deemed to be an overkill in this project.

Sequence diagrams are increasingly useful for microservices communication. Otherwise, they're only useful for complex service logic, which is rare for this project at the time of writing.

## Consequences

More resources are needed to follow SEED(S), however, services will have a more robust design.
