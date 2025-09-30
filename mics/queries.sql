-- INFO: I'd like to write sql queries inside nvim haha

-- create a user
INSERT INTO users (email, username, password)
VALUES (
  'minhhoccode111@gmail.com',
  'minhhoccode111',
  '$2a$10$I9ZdFZ1OMx.LO3dnmv65DO344FPoaUj8LXXv01jzmIgIKFAqF5uia'
);

-- select user by either id, email, username
select id, email, username, password, bio, image, created_at, updated_at
from users where id::text = 'd89b1945-3193-435d-90c6-b6da95317893' or email = '' or username = '';

-- check if a tag exists
select exists (select 1 from tags where name = 'tai');

-- query all tags of an article
select slug, title, description, body, name from articles a
left join article_tags at on a.id = at.article_id
left join tags t on at.tag_id = t.id
where a.slug = 'slug';

-- better version of above written by AI
select a.slug, a.title, a.description, a.body,
  string_agg(t.name, ', ') as tags
from articles a
left join article_tags at on a.id = at.article_id
left join tags t on at.tag_id = t.id
where a.title = 'title cannot be empty'
group by a.id, a.slug, a.title, a.description, a.body;

-- query all articles of a tag
select t.name, a.slug, u.username from tags t
left join article_tags at on t.id = at.tag_id
left join articles a on at.article_id = a.id
left join users u on u.id = a.author_id
where t.name = 'tai';

-- check if a user (id) follows another user (username)
select exists (
  select 1 from follows
  where follower_id = '7c6ecf9d-0c0d-43f7-959f-f397706a760e'
  and following_id = (
    select id from users
    where username = 'minhhoccode111'
  )
);


-- a user follow another user by username
insert into follows (follower_id, following_id)
values (
  'd89b1945-3193-435d-90c6-b6da95317893',
  (select id from users where username = 'minhhoccode111')
);
