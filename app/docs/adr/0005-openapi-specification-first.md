# 5. OpenAPI Specification First

Date: 2025-02-23

## Status

Accepted

## Context

There's a need to fix the API contract between the platform and the consumer. There's a need for API documentation.

## Decision

Center API development around the OpenAPI Specification (OAS).

OAS gives the following advantages:
- API gateway integration (automatic requests validation);
- Server code generation (http handlers, request and response models, http endpoints and their respective methods - a lot of code can be generated);
- Client code generation (platform API consumers can leverage tools to generate the code that would help in interacting with the platform API: code generation will create code for API Operations, API request and response models);
- Code documentation (robust code documentation useful to both the API consumers and API developer).

## Consequences

API is tied to one particular tooling (OAS).
