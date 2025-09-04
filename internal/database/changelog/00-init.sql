CREATE TABLE IF NOT EXISTS platforms (
    id VARCHAR PRIMARY KEY,
    name VARCHAR,
    buying_power INTEGER
);

CREATE TABLE IF NOT EXISTS models (name VARCHAR PRIMARY KEY);

CREATE TABLE IF NOT EXISTS platform_models (
    platform_id VARCHAR REFERENCES platforms(id),
    model_name VARCHAR REFERENCES models(name),
    tokens INTEGER,
    PRIMARY KEY (platform_id, model_name)
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID DEFAULT uuid() PRIMARY KEY,
    user_id VARCHAR,
    platform_id VARCHAR REFERENCES platforms(id),
    model_name VARCHAR REFERENCES models(name),
    amount INTEGER,
    date TIMESTAMP,
    status VARCHAR
);