CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE "auth"."users" (
    id BIGINT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    password TEXT NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "auth"."refresh_tokens" (
    id VARCHAR(75) NOT NULL,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    PRIMARY KEY ("id", "user_id"),
    FOREIGN KEY (user_id) REFERENCES "auth"."users" (id) ON DELETE CASCADE
);

