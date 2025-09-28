CREATE SEQUENCE seq_history START 1;

CREATE TABLE IF NOT EXISTS history (
    id INTEGER PRIMARY KEY DEFAULT NEXTVAL('seq_history'),
    model_name VARCHAR REFERENCES models(name),
    prompt VARCHAR
);