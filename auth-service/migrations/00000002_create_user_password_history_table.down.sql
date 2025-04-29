
DROP TRIGGER IF EXISTS manage_user_password_history ON "user_password_history";

DROP TRIGGER IF EXISTS update_user_password_updated_at ON "user_password_history";

DROP FUNCTION IF EXISTS manage_password_history();
DROP FUNCTION IF EXISTS update_password_updated_at();

DROP TABLE IF EXISTS "user_password_history";
