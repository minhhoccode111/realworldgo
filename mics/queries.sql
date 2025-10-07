-- INFO: I'd like to write sql queries inside nvim haha

-- create a user
INSERT INTO users (email, username, password)
VALUES (
  'minhhoccode111@gmail.com',
  'minhhoccode111',
  '$2a$10$I9ZdFZ1OMx.LO3dnmv65DO344FPoaUj8LXXv01jzmIgIKFAqF5uia'
);

-- 7c6ecf9d-0c0d-43f7-959f-f397706a760e
-- d89b1945-3193-435d-90c6-b6da95317893
-- 430e5f27-3aaf-48d5-b9ab-a4f364dd8119

-- select user by either id, email, username
select id, email, username, password, bio, image, created_at, updated_at
from users where id::text = 'd89b1945-3193-435d-90c6-b6da95317893' or email = '' or username = '';

-- check if a tag exists
select exists (select 1 from tags where name = 'tai');

-- count tags
select t.name, at.tag_id, count(*) from article_tags at
left join tags t on at.tag_id = t.id
group by at.tag_id, t.name;

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
group by a.id;

-- query all articles of a tag
select t.name, a.slug, u.username from tags t
left join article_tags at on t.id = at.tag_id
left join articles a on at.article_id = a.id
left join users u on u.id = a.author_id
where t.name = 'tai';

-- check if a user (id) follows another user (username)
select exists (
  select 1 from follows
  where follower_id = 'd89b1945-3193-435d-90c6-b6da95317893'
  and following_id = (
    select id from users
    where username = 'asd0'
  )
);

-- a user follow another user by username
insert into follows (follower_id, following_id)
values (
  'd89b1945-3193-435d-90c6-b6da95317893',
  (select id from users where username = 'minhhoccode111')
);

-- a user unfollow another user by username
delete from follows
where follower_id = (select id from users where username = 'asd0')
and following_id = (select id from users where username = 'minhhoccode111');

-- generate fake data by letting a user favorite all articles
insert into favorites (user_id, article_id)
select u.id, a.id
from users u
cross join articles a
where u.username = 'minhhoccode111';

-- a user favorite an article by its slug
insert into favorites (user_id, article_id)
  select u.id, a.id
from users u, articles a
  where u.username = 'asd0' and a.slug = 'slug';

-- a user unfavorite an article by its slug
delete from favorites f
where f.article_id = (select id from articles where slug = 'slug')
and f.user_id = (select id from users where username = 'asd0');

-- generate fake data by letting a user favorite all articles of other users except itself
insert into favorites (user_id, article_id)
select u.id, a.id
from users u
cross join articles a
where u.username = 'asd0' and a.author_id != u.id;

-- count duplicated title articles
SELECT title, COUNT(*) AS count_articles
FROM articles
WHERE deleted_at IS NULL
GROUP BY title
ORDER BY count_articles DESC;

-- count favorites of articles
select a.slug, count(*) as favorites_count
from articles a
left join favorites f on f.article_id = a.id
group by a.slug
order by favorites_count desc;

--------------------------------------------------------------------------------
-- list articles, filter by tags, author, favorited, limit, offset
-- and return multiple articles with their tags and author profile
-- ordered by most recent first and doesn't have deleted_at

-- basic with unauthenticated users
select a.slug, a.title, a.description, a.created_at, a.updated_at, false as favorited,
  u.username, u.bio, false as following,
  array_agg(t.name) filter (where t.name is not null) as tags,
  count(distinct f.user_id) as favorites_count
from articles a
left join users u on a.author_id = u.id
left join article_tags at on at.article_id = a.id
left join tags t on t.id = at.tag_id
left join favorites f on f.article_id = a.id
where a.deleted_at is null
group by a.id, u.id
order by a.created_at desc
limit 20
offset 0;

-- advanced with authenticated users
-- explain: authenticated user: minhhoccode111, id = 'd89b1945-3193-435d-90c6-b6da95317893'
-- filter all articles written by 'asd0' and favorited by 'minhhoccode111' and have tag 'sao'
-- NOTE: this hurt my brain bruh
select a.slug, a.title, a.description, a.created_at, a.updated_at,
  (select exists
    (select 1 from favorites where user_id::text = 'd89b1945-3193-435d-90c6-b6da95317893' and article_id = a.id)
  ) as favorited,
  u.username, u.bio, u.image,
  (select exists
    (select 1 from follows where follower_id::text = 'd89b1945-3193-435d-90c6-b6da95317893' and following_id = u.id)
  ) as following,
  coalesce(array_agg(distinct t.name) filter (where t.name is not null), '{}') as tags,
  count(distinct f.user_id) as favorites_count,
  count(*) over() as articles_count -- count all articles after filtering and before applying limit
from articles a
left join users u on a.author_id = u.id
left join article_tags at on at.article_id = a.id
left join tags t on t.id = at.tag_id
left join favorites f on f.article_id = a.id
left join users u2 on f.user_id = u2.id
where a.deleted_at is null
  -- and ('' = 'asd0' or u.username = 'asd0') -- filter by author username, skip if empty string
  -- and ('' = 'minhhoccode111' or u2.username = 'minhhoccode111') -- filter by favorited username, skip if empty string
  -- and ('' = 'sao' or exists (select 1 from article_tags at2
  --     left join tags t2 on at2.tag_id = t2.id
  --     where at2.article_id = a.id and t2.name = 'sao')) -- filter by tag, skip if empty string
group by a.id, u.id
order by a.created_at desc
limit 10
offset 0;
--------------------------------------------------------------------------------

-- get /feed, recent articles from people you followed
select a.slug, a.title, a.description, a.created_at, a.updated_at,
  (select exists
    (select 1 from favorites where user_id::text = 'd89b1945-3193-435d-90c6-b6da95317893' and article_id = a.id)
  ) as favorited,
  u.username, u.bio, u.image,
  coalesce(array_agg(distinct t.name) filter (where t.name is not null), '{}') as tags,
  count(distinct f.user_id) as favorites_count,
  count(*) over() as articles_count
from articles a
left join users u on a.author_id = u.id
left join article_tags at on at.article_id = a.id
left join tags t on t.id = at.tag_id
left join favorites f on f.article_id = a.id
left join users u2 on f.user_id = u2.id
where a.deleted_at is null
  and (select exists
    (select 1 from follows where follower_id::text = 'd89b1945-3193-435d-90c6-b6da95317893'
      and following_id = u.id)
  )
group by a.id, u.id
order by a.created_at desc
limit 10
offset 0;

-- select an article detail (with body) by slug, get tags, favorited, favorites_count, author profile, personalize with current user, deleted_at is null
select a.slug, a.title, a.description, a.body, a.created_at, a.updated_at,
  coalesce(array_agg(distinct t.name) filter (where t.name is not null), '{}') as tags,
  (select exists
    (select 1 from favorites where article_id = a.id and user_id::text = '7c6ecf9d-0c0d-43f7-959f-f397706a760e')
  ) as favorited,
  (count(distinct f.user_id)) as favorites_count,
  u.username, u.bio, u.image,
  (select exists
    (select 1 from follows where a.author_id = following_id and follower_id::text = '7c6ecf9d-0c0d-43f7-959f-f397706a760e')
  ) as following
from articles a
left join users u on a.author_id = u.id
left join article_tags at on at.article_id = a.id
left join tags t on at.tag_id = t.id
left join favorites f on f.article_id = a.id
where a.deleted_at is null and slug = 'slug'
group by a.id, u.id;

-- select an article only
select id, author_id, slug, title, description, body, created_at, updated_at
from articles
where slug = 'slug'
and deleted_at is null;

-- update an article
update articles
set slug = 'slug', title = 'title', description = 'description', body = 'body'
where id::text = '';

-- delete an article by setting its deleted_at to now()
update articles
set deleted_at = now()
where id::text = '' and author_id::text = '';

-- create comment on an article with slug, authorId, body and article deleted_at is null
insert into comments (article_id, author_id, body)
values (
  (select id from users where username = 'asd0'),
  (select id from articles where slug = 'title-cannot-be-empty-9' and deleted_at is null),
  'body 0'
)
returning id;

-- select a comment, author profile and personalize with current user
select c.id, c.body, c.created_at,
  u.username, u.bio, u.image,
  (select exists (
    select 1 from follows
    where follower_id::text = ''
    and following_id = c.author_id
  )) as following
from comments c
left join users u on u.id = c.author_id
left join articles a on a.id = c.article_id
where c.deleted_at is null
and a.deleted_at is null
and c.id = 'da1b0dc3-e2a5-4930-9e5d-1dd6f7884717';

-- select all comments of an article, author profile and personalize with current user
select c.id, c.body, c.created_at,
  u.username, u.bio, u.image,
  (select exists (
    select 1 from follows
    where follower_id::text = 'd89b1945-3193-435d-90c6-b6da95317893'
    and following_id = c.author_id
  )) as following,
  count(*) over() as comments_count
from comments c
left join users u on u.id = c.author_id
left join articles a on a.id = c.article_id
where c.deleted_at is null
and a.deleted_at is null
and a.slug = ''
group by a.id, u.id, c.id
order by c.created_at -- latest bottom
limit 10
offset 0;

-- delete a comment in an article by setting its deleted_at to now()
update comments c
set c.deleted_at = now()
where c.author_id = ''
and exists (
  select 1 from articles a
  where a.id = c.article_id
  and a.slug = ''
  and a.deleted_at is null
)
and c.id = ''
and c.deleted_at is null;

-- current user favorite an active article by its slug
insert into favorites (user_id, article_id)
values (
  'd89b1945-3193-435d-90c6-b6da95317893',
  (select id from articles where slug = 'slug' and deleted_at is null)
);

-- current user unfavorite an article by its slug
delete from favorites
where user_id = ''
and article_id = (
  select id from articles
  where slug = ''
  and deleted_at is null
);

-- get all tags
select distinct t.name,
  count(*) over() as tags_count
from tags t
order by t.name
limit 5
offset 0;
