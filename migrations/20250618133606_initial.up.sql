CREATE TABLE customers (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    email VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
CREATE TABLE service_providers (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    email VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
CREATE TABLE ratings (
    id VARCHAR PRIMARY KEY,
    customer_id VARCHAR NOT NULL REFERENCES customers(id),
    service_provider_id VARCHAR NOT NULL REFERENCES service_providers(id),
    rating_value INTEGER NOT NULL,
    comment TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_ratings_service_provider ON ratings(service_provider_id);
CREATE INDEX idx_ratings_created_at ON ratings(created_at);