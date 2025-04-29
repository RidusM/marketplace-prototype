CREATE SCHEMA IF NOT EXISTS profile_schema;

CREATE TABLE IF NOT EXISTS profile_schema.profiles (
    "profile_id" UUID PRIMARY KEY,
    "user_id" UUID NOT NULL UNIQUE,
    "username" varchar(30) NOT NULL,
    "first_name" varchar(30),
    "middle_name" varchar(30),
    "last_name" varchar(30),
    "phone_number" varchar(30),
    "email" varchar(50) NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now()),
    "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON profile_schema.profiles
    FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_column();