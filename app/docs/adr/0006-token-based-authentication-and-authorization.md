# 6. Token-based authentication and authorization

Date: 2025-02-23

## Status

Accepted

## Context

There's a need to create AuthNZ model for the system.

## Decision

Use token-based AuthNZ: refresh and access tokens based on JWT. Tokens are signed using asymmetric cryptography, allowing services to check identity and authorize requests without centralization, allowing for higher level of availability. Refresh tokens can be revoked.

## Consequences

Higher level of availability comes at a cost of security: access tokens expire only after 30 minutes (to make Serverless model cost-efficient).
