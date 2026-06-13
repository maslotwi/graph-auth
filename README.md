# Graph-Auth-Server: Architecture & Flow Summary

## Overview
`graph-auth-server` is a high-performance, passwordless Single Sign-On (SSO) Identity Provider. It replaces traditional relational databases and heavy, opinionated OAuth2 frameworks with a custom Graph-backed session tree, microsecond caching, and blazing-fast Go routing.

## Core Infrastructure

* **Neo4j (The Source of Truth):** Manages long-lived sessions as nodes in a directed graph. This enables cascading scope inheritance and instant hereditary revocation. For example, revoking a User's Phone session via a Cypher query will instantly cascade and sever any PC sessions that were spawned by that phone.
* **Redis (The Hot Cache):** Stores short-lived OTP delegation codes, OAuth authorization codes, and temporary JWT session validations (`jti`) for microsecond read access without hitting the graph database.
* **Fiber v3 (The Gateway):** Handles all HTTP requests, serves dynamic OpenAPI specifications via Scalar, and orchestrates the SSO logic using `fasthttp`.
* **Cryptography:** Issues standard signed JWT Access Tokens for downstream microservices to validate statelessly via the server's public RSA/ECDSA keys.

## Core Flows

### 1. Peer-to-Peer Session Delegation (Passwordless Login)
Instead of typing passwords, unauthenticated devices gain access by being "sponsored" by an already authenticated device in the graph.
1. **Device A (Active)** requests to sponsor a new device. The server generates a secure 6-digit code / QR URL and stores it in Redis.
2. **Device B (New)** submits the code.
3. The server executes a Cypher query to spawn `(Device B)` directly underneath `(Device A)` in the Neo4j graph, linking them and inheriting restricted access scopes.

### 2. The OAuth2 SSO Flow
Standardized so external apps (e.g., "Service A") can integrate with the Identity Provider using standard OAuth2 Authorization Code logic.
1. **`GET /oauth/authorize`**: Service A redirects the user here. The API checks the device state. If unauthenticated, they are routed to the Delegation UI (Flow 1). If authenticated, they are routed to the Consent UI.
2. **`POST /oauth/confirm`**: The user clicks "Approve". The server generates a temporary Auth Code stored in Redis and redirects the browser back to Service A.
3. **`POST /oauth/token`**: Service A server trades the Auth Code for a signed JWT Access Token, caching the session state in Redis for rapid validation.

## Automated Tooling
The project features a highly customized, cross-platform build pipeline that handles code formatting, OpenAPI spec generation (via Swaggo), and documentation presentation.

```bash
# Formats Go code, aligns comments with awk, generates Swagger docs, and boots the server
make run   

# Compiles the production binary
make build 
```

To bypass native bugs in Swaggo's `swag fmt` command, the pipeline utilizes a strict `awk` script to enforce perfect 20-column spacing on all declarative Swagger comments during generation. The final spec is dynamically injected with environment variables at runtime and served interactively via the Scalar UI CDN.