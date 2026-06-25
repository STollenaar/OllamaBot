CREATE TABLE IF NOT EXISTS contexts (
    user_id VARCHAR,
    model_name VARCHAR REFERENCES models(name),
    system_prompt VARCHAR,
    context INTEGER [],
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    PRIMARY KEY (user_id, model_name)
)