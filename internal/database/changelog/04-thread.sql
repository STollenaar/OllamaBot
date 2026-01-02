CREATE TABLE IF NOT EXISTS threads (
    thread_id VARCHAR PRIMARY KEY,
    model_name VARCHAR REFERENCES models(name),
    system_prompt VARCHAR,
    context INTEGER [],
)