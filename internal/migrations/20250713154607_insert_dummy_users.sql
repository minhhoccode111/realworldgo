-- +goose Up
-- +goose StatementBegin
INSERT INTO users (email, password, last_name, first_name, is_active, role)
SELECT * FROM (VALUES
  ('minhhoccode111@gmail.com', '$2a$10$I9ZdFZ1OMx.LO3dnmv65DO344FPoaUj8LXXv01jzmIgIKFAqF5uia', 'Minh', 'Dang', true, 'admin'::user_role),
  ('asd0@gmail.com', '$2a$10$I9ZdFZ1OMx.LO3dnmv65DO344FPoaUj8LXXv01jzmIgIKFAqF5uia', 'Dummy', 'Account', true, 'user'::user_role)
) AS t(email, password, last_name, first_name, is_active, role)
WHERE NOT EXISTS (SELECT 1 FROM users);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE email IN ('minhhoccode111@gmail.com', 'asd@gmail.com');
-- +goose StatementEnd
