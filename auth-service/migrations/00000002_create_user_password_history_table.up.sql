CREATE TABLE "user_password_history" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    hashed_password VARCHAR(255) NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION manage_password_history()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM "user_password_history"
    WHERE id IN (
        SELECT id FROM "user_password_history"
        WHERE user_id = NEW.user_id
        ORDER BY updated_at ASC
        LIMIT 1 OFFSET 5
    );

    NEW.updated_at = CURRENT_TIMESTAMP;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER manage_user_password_history
AFTER INSERT ON "user_password_history"
FOR EACH ROW
EXECUTE PROCEDURE manage_password_history();