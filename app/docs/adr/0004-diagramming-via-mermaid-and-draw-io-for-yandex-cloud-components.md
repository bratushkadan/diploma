# 4. Diagramming via mermaid and draw.io for Yandex Cloud Components

Date: 2025-02-23

## Status

Accepted

## Context

There's a need for visual representation of system's architecture and components interaction.

## Decision

Use mermaid docs as code tool to render diagrams in Markdown extension files. Use draw.io to create diagrams that have Yandex Cloud components. Draw.io allows for installation of custom icons, thus Yandex Cloud icons can be installated, and diagrams in draw.io become more specific for cloud-based architectures.

Store draw.io diagrams in Git.

## Consequences

Draw.io requires an editor from a project member.
