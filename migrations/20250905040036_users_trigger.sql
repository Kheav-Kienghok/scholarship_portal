-- +goose Up
-- +goose StatementBegin

CREATE OR REPLACE FUNCTION enforce_max_two_admins()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.role = 'admin' THEN
        IF (SELECT COUNT(*) FROM users WHERE role = 'admin') >= 2 THEN
            RAISE EXCEPTION 'Maximum number of admins reached';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_max_two_admins
BEFORE INSERT OR UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION enforce_max_two_admins();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_max_two_admins ON users;
DROP FUNCTION IF EXISTS enforce_max_two_admins;
-- +goose StatementEnd
