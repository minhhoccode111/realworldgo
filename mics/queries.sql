-- INFO: I'd like to write sql queries inside nvim haha

-- create a user
INSERT INTO users (email, username, password)
VALUES (
  'minhhoccode111@gmail.com',
  'minhhoccode111',
  '$2a$10$I9ZdFZ1OMx.LO3dnmv65DO344FPoaUj8LXXv01jzmIgIKFAqF5uia'
);

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
