CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users
(
    id            uuid                  DEFAULT uuid_generate_v4() PRIMARY KEY,
    login         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP    NULL,
    deleted_at    TIMESTAMP    NULL,
    CONSTRAINT login_uniq UNIQUE (login)
);