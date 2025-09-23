-- to test after the migration that drop default of updated_at column

-- create a dummy user
INSERT INTO users (email, username, password)
VALUES (
  'minhhoccode111@gmail.com',
  'minhhoccode111',
  '$2a$10$I9ZdFZ1OMx.LO3dnmv65DO344FPoaUj8LXXv01jzmIgIKFAqF5uia'
);
