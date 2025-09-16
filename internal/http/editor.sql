-- to test after the migration that drop default of updated_at column

-- create a dummy user
INSERT INTO users (email, password, is_active, role)
VALUES (
  'asd0@gmail.com',
  '$2a$10$I9ZdFZ1OMx.LO3dnmv65DO344FPoaUj8LXXv01jzmIgIKFAqF5uia', -- "asdasd"
  true,
  'user'::user_role
  );

-- create a dummy user
INSERT INTO users (email, password, is_active, role)
VALUES (
  'asd8@gmail.com',
  '$2a$10$I9ZdFZ1OMx.LO3dnmv65DO344FPoaUj8LXXv01jzmIgIKFAqF5uia', -- "asdasd"
  false,
  'user'::user_role
  );


select * from users where email ilike '%' || 'm' || '%';

