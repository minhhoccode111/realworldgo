package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"github.com/lib/pq"
	. "github.com/minhhoccode111/realworldgo/internal/model"
	. "github.com/minhhoccode111/realworldgo/internal/utils"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// WARN: remember, every query must take deleted_at into consideration

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close(dbName string) error

	// CreateUser inserts a new user
	CreateUser(ctx context.Context, newUser *User) error

	// SelectUserById returns a user by id
	SelectUser(ctx context.Context, id, email, username string) (*User, error)

	// UpdateUser updates a user
	UpdateUser(ctx context.Context, newUser *User) error

	// CanSlugBeUSed checks if an article with the given slug already exists
	CanSlugBeUSed(ctx context.Context, articleId, slug string) (bool, error)

	// CreateArticle inserts a new user
	CreateArticle(ctx context.Context, newArticle *Article, tags []string) (string, error)

	// SelectArticles returns a list of articles
	SelectArticles(
		ctx context.Context,
		currentUserId, tag, author, favorited string,
		limit, offset int,
	) (articles []ArticlePreview, articlesCount int, err error)

	// SelectArticlesFeed returns a list of articles
	SelectArticlesFeed(
		ctx context.Context,
		currentUserId string,
		limit, offset int,
	) (articles []ArticlePreview, articlesCount int, err error)

	// SelectArticle
	SelectArticle(ctx context.Context, slug string) (*Article, error)

	// SelectArticleDetail returns an article (with body), tags, favorites count, author, and relationship between current user and the article, author
	SelectArticleDetail(ctx context.Context, currentUserId, slug string) (*ArticleDetail, error)

	// UpdateUser updates the email of a user in the database.
	UpdateArticle(ctx context.Context, newArticle *Article) (string, error)

	// DeleteArticle soft deletes the article
	DeleteArticle(ctx context.Context, currentUserId, slug string) error

	// CreateComment creates a new comment on an article
	CreateComment(ctx context.Context, slug, currentUserId, body string) (string, error)

	// SelectComments returns a list of comments
	SelectComments(
		ctx context.Context,
		currentUserId, slug string,
		limit, offset int,
	) (articles []CommentDetail, articlesCount int, err error)

	// SelectCommentDetail returns a comment (with body), author, and relationship between current user and the author
	SelectCommentDetail(
		ctx context.Context,
		currentUserId, commentId string,
	) (*CommentDetail, error)

	// DeleteComment soft deletes the comment
	DeleteComment(ctx context.Context, currentUserId, slug, commentId string) error

	// IsFollowing checks if the follower is following the followingName
	IsFollowing(ctx context.Context, followerId, followingName string) (bool, error)

	// CreateFollow creates a new follower for the followingUsername
	CreateFollow(ctx context.Context, followerId, followingUsername string) error

	// DeleteFollow deletes a follower for the followingUsername
	DeleteFollow(ctx context.Context, followerId, followingUsername string) error

	// CreateFavorite creates a new favorite for the article
	CreateFavorite(ctx context.Context, currentUserId, slug string) error

	// DeleteFavorite deletes a favorite for the article
	DeleteFavorite(ctx context.Context, currentUserId, slug string) error
}

type service struct {
	db *sql.DB
}

func New(connStr string) Service {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return NewService(db)
}

// NewService creates a new database service with the given *sql.DB instance.
// This function is primarily intended for testing purposes, allowing a mock database
// to be injected.
func NewService(db *sql.DB) Service {
	return &service{db: db}
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Printf("db down: %v", err) // Log the error instead of terminating
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close(dbName string) error {
	log.Printf("Disconnected from database: %s", dbName)
	return s.db.Close()
}

func (s *service) CreateUser(ctx context.Context, newUser *User) error {
	hashedPassword, err := HashedPassword(newUser.Password)
	if err != nil {
		log.Printf("Error hashing %v: %v", newUser.Password, err)
		return fmt.Errorf("Error hashing %v: %v", newUser.Password, err)
	}
	row := s.db.QueryRowContext(ctx, `
		INSERT INTO users(email, username, password, bio, image)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id
		`,
		newUser.Email,
		newUser.Username,
		hashedPassword,
		newUser.Bio,
		newUser.Image,
	)
	err = row.Scan(&newUser.Id)
	return err
}

func (s *service) SelectUser(ctx context.Context, id, email, username string) (*User, error) {
	var user User
	var row *sql.Row

	switch {
	case id != "":
		row = s.db.QueryRowContext(ctx, `
		SELECT id, email, username, password, bio, image, created_at, updated_at
		FROM users WHERE id = $1`, id,
		)
	case email != "":
		row = s.db.QueryRowContext(ctx, `
		SELECT id, email, username, password, bio, image, created_at, updated_at
		FROM users WHERE email = $1`, email,
		)
	case username != "":
		row = s.db.QueryRowContext(ctx, `
		SELECT id, email, username, password, bio, image, created_at, updated_at
		FROM users WHERE username = $1`, username,
		)
	default:
		return nil, fmt.Errorf("Must provide either id, email or username")
	}

	if err := row.Scan(
		&user.Id,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.Bio,
		&user.Image,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *service) UpdateUser(ctx context.Context, newUser *User) error {
	hashedPassword, err := HashedPassword(newUser.Password)
	if err != nil {
		log.Printf("Error hashing %v: %v", newUser.Password, err)
		return fmt.Errorf("Error hashing %v: %v", newUser.Password, err)
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE users SET
		email = $1,
		username = $2,
		password = $3,
		image = $4,
		bio = $5
		WHERE id = $6
		`,
		newUser.Email,
		newUser.Username,
		hashedPassword,
		newUser.Image,
		newUser.Bio,
		newUser.Id,
	)
	return err
}

func (s *service) CanSlugBeUSed(ctx context.Context, articleId, slug string) (bool, error) {
	// INFO: how about articles that have deleted_at? Should we take into account?
	// if an article try to update with its same old slug, we can skip
	query := `
	select exists (
		select 1 from articles
		where id::text <> $1 and slug = $2
	)
	`
	var existed bool
	err := s.db.QueryRowContext(ctx, query, articleId, slug).Scan(&existed)
	if err != nil {
		return false, err
	}
	return existed, nil
}

func (s *service) CreateArticle(
	ctx context.Context,
	newArticle *Article,
	tags []string,
) (newSlug string, err error) {
	// 1. insert new article to db to generate id
	// 2. create a list of tags, will return error if unique constraint fail
	// 3. create rows in junction table between article and tags
	// we can apply concurrency for 1. and 2.
	// and we also need ACID transaction to make sure every query succeed

	var tx *sql.Tx
	tx, err = s.db.Begin()
	if err != nil {
		return "", err
	}

	// defer a rollback in case of an error or panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback() // rollback transaction if error occurs
		}
	}()

	// TODO: add concurrency with goroutines and channels to improve performance

	// insert an article, return its id
	baseSlug := slug.Make(newArticle.Title)
	newArticle.Slug = baseSlug
	for i := 0; ; i++ {
		var existed bool
		existed, err = s.CanSlugBeUSed(ctx, "", newArticle.Slug)
		if err != nil {
			return "", err
		}

		if !existed {
			break
		}

		newArticle.Slug = baseSlug + "-" + strconv.Itoa(i)
	}

	err = s.db.QueryRowContext(ctx, `
		INSERT INTO articles (author_id, slug, title, description, body)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
		`,
		newArticle.AuthorId,
		newArticle.Slug,
		newArticle.Title,
		newArticle.Description,
		newArticle.Body,
	).Scan(&newArticle.Id, &newArticle.CreatedAt, &newArticle.UpdatedAt)
	if err != nil {
		return "", err
	}

	// TODO: if we have large number of tags, we need to use batch insert to improve performance
	for _, tagName := range tags {
		// insert or get existing tag
		var tagId string
		// `ON CONFLICT (name)`: if insert conflict on `name` col (unique constraint)
		// `DO UPDATE SET name=EXCLUDED.name`: update the name to the new name (same)
		// `EXCLUDED`: represent our newly inserted row
		err := s.db.QueryRowContext(ctx, `
				INSERT INTO tags (name)
				VALUES ($1)
				ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name
				RETURNING id
				`, tagName,
		).Scan(&tagId)
		if err != nil {
			return "", err
		}

		// insert junction
		_, err = s.db.ExecContext(ctx, `
			insert into article_tags (article_id, tag_id)
			values ($1, $2)
			on conflict do nothing
			`,
			newArticle.Id,
			tagId,
		)
		if err != nil {
			return "", err
		}
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return newArticle.Slug, nil
}

func (s *service) SelectArticles(
	ctx context.Context,
	currentUserId, tag, author, favorited string,
	limit, offset int,
) (articles []ArticlePreview, articlesCount int, err error) {
	query := `
		select a.slug, a.title, a.description, a.created_at, a.updated_at,
		  (select exists
			(select 1 from favorites where user_id::text = $1 and article_id = a.id)
		  ) as favorited,
		  u.username, u.bio, u.image,
		  (select exists
			(select 1 from follows where follower_id::text = $1 and following_id = u.id)
		  ) as following,
		  coalesce(array_agg(distinct t.name) filter (where t.name is not null), '{}') as tags,
		  count(distinct f.user_id) as favorites_count,
		  count(*) over() as articles_count -- count all articles match before applying limit
		from articles a
		left join users u on a.author_id = u.id
		left join article_tags at on at.article_id = a.id
		left join tags t on t.id = at.tag_id
		left join favorites f on f.article_id = a.id
		left join users uf on f.user_id = uf.id
		where a.deleted_at is null
		  and ('' = $2 or u.username = $2) -- author, skip if empty
		  and ('' = $3 or uf.username = $3) -- favorited, skip if empty
		  and ('' = $4 or exists (select 1 from article_tags at2
			  left join tags t2 on at2.tag_id = t2.id
			  where at2.article_id = a.id and t2.name = $4)) -- tag, skip if empty
		group by a.id, u.id
		order by a.created_at desc
		limit $5
		offset $6;
	`

	/*
		example output:
		          slug           |         title         |         description         |          created_at           |          updated_at           | favorited | username | bio | image | following |     tags     | favorites_count | articles_count
		-------------------------+-----------------------+-----------------------------+-------------------------------+-------------------------------+-----------+----------+-----+-------+-----------+--------------+-----------------+----------------
		 title-cannot-be-empty-6 | title cannot be empty | description cannot be empty | 2025-10-02 13:38:00.168028+00 | 2025-10-02 13:38:00.168028+00 | t         | asd0     |     |       | t         | {tai,vi,sao} |               1 |              1
	*/

	rows, err := s.db.QueryContext(
		ctx,
		query,
		currentUserId,
		author,
		favorited,
		tag,
		limit,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var a ArticlePreview
		var tags pq.StringArray
		err = rows.Scan(
			&a.Slug,
			&a.Title,
			&a.Description,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.Favorited,
			&a.Author.Username,
			&a.Author.Bio,
			&a.Author.Image,
			&a.Author.Following,
			&tags,
			&a.FavoritesCount,
			&articlesCount,
		)
		if err != nil {
			return nil, 0, err
		}
		a.TagList = []string(tags)
		articles = append(articles, a)
	}

	err = rows.Err()
	if err != nil {
		return nil, 0, err
	}

	return articles, articlesCount, nil
}

func (s *service) SelectArticlesFeed(
	ctx context.Context,
	currentUserId string,
	limit, offset int,
) (articles []ArticlePreview, articlesCount int, err error) {
	query := `
		select a.slug, a.title, a.description, a.created_at, a.updated_at,
		  (select exists
			(select 1 from favorites where user_id::text = $1 and article_id = a.id)
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
			(select 1 from follows where follower_id::text = $1
			  and following_id = u.id)
		  )
		group by a.id, u.id
		order by a.created_at desc
		limit $2
		offset $3;
	`

	/*
		example output:
		          slug           |         title         |         description         |          created_at           |          updated_at           | favorited | username | bio | image |     tags     | favorites_count | articles_count
		-------------------------+-----------------------+-----------------------------+-------------------------------+-------------------------------+-----------+----------+-----+-------+--------------+-----------------+----------------
		 title-cannot-be-empty-8 | title cannot be empty | description cannot be empty | 2025-10-02 13:38:13.768144+00 | 2025-10-02 13:38:13.768144+00 | t         | asd0     |     |       | {tag}        |               1 |              3
	*/

	rows, err := s.db.QueryContext(
		ctx,
		query,
		currentUserId,
		limit,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var a ArticlePreview
		var tags pq.StringArray
		err = rows.Scan(
			&a.Slug,
			&a.Title,
			&a.Description,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.Favorited,
			&a.Author.Username,
			&a.Author.Bio,
			&a.Author.Image,
			// &a.Author.Following,
			&tags,
			&a.FavoritesCount,
			&articlesCount,
		)
		if err != nil {
			return nil, 0, err
		}
		a.TagList = []string(tags)
		a.Author.Following = true
		articles = append(articles, a)
	}

	err = rows.Err()
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, err
}

func (s *service) SelectArticle(ctx context.Context, slug string) (*Article, error) {
	query := `
	select id, author_id, slug, title, description, body, created_at, updated_at
	from articles
	where slug = $1
	and deleted_at is null
	`
	var a Article
	err := s.db.QueryRowContext(ctx, query, slug).Scan(
		&a.Id,
		&a.AuthorId,
		&a.Slug,
		&a.Title,
		&a.Description,
		&a.Body,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *service) SelectArticleDetail(
	ctx context.Context,
	currentUserId, slug string,
) (*ArticleDetail, error) {
	query := `
		select a.slug, a.title, a.description, a.body, a.created_at, a.updated_at,
		  coalesce(array_agg(distinct t.name) filter (where t.name is not null), '{}') as tags,
		  (select exists
			(select 1 from favorites where article_id = a.id and user_id::text = $1)
		  ) as favorited,
		  (count(distinct f.user_id)) as favorites_count,
		  u.username, u.bio, u.image,
		  (select exists
			(select 1 from follows where a.author_id = following_id and follower_id::text = $1)
		  ) as following
		from articles a
		left join users u on a.author_id = u.id
		left join article_tags at on at.article_id = a.id
		left join tags t on at.tag_id = t.id
		left join favorites f on f.article_id = a.id
		where a.deleted_at is null and slug = $2
		group by a.id, u.id;
	`

	/*
		example output:
		 slug | title |         description         |         body         |          created_at          |          updated_at          |     tags     | favorited | favorites_count |    username    |      bio      |                     image                      | following
		------+-------+-----------------------------+----------------------+------------------------------+------------------------------+--------------+-----------+-----------------+----------------+---------------+------------------------------------------------+-----------
		 slug | slug  | description cannot be empty | body cannot be empty | 2025-10-05 05:23:47.80455+00 | 2025-10-05 05:23:47.80455+00 | {sao,tai,vi} | t         |               3 | minhhoccode111 | i like golang | https://www.w3schools.com/howto/img_avatar.png | t
	*/

	a := ArticleDetail{}
	var tags pq.StringArray
	err := s.db.QueryRowContext(ctx, query, currentUserId, slug).Scan(
		&a.Slug,
		&a.Title,
		&a.Description,
		&a.Body,
		&a.CreatedAt,
		&a.UpdatedAt,
		&tags,
		&a.Favorited,
		&a.FavoritesCount,
		&a.Author.Username,
		&a.Author.Bio,
		&a.Author.Image,
		&a.Author.Following,
	)
	if err != nil {
		return nil, err
	}

	a.TagList = []string(tags)
	return &a, nil
}

func (s *service) UpdateArticle(ctx context.Context, newArticle *Article) (string, error) {
	var err error
	baseSlug := slug.Make(newArticle.Title)
	newArticle.Slug = baseSlug
	for i := 0; ; i++ {
		var existed bool
		existed, err = s.CanSlugBeUSed(ctx, newArticle.Id, newArticle.Slug)
		if err != nil {
			return "", err
		}

		if !existed {
			break
		}

		newArticle.Slug = baseSlug + "-" + strconv.Itoa(i)
	}

	query := `
		update articles
		set slug = $1, title = $2, description = $3, body = $4
		where id = $5 and deleted_at is null;
	`

	_, err = s.db.ExecContext(ctx, query,
		newArticle.Slug,
		newArticle.Title,
		newArticle.Description,
		newArticle.Body,
		newArticle.Id,
	)
	if err != nil {
		return "", err
	}

	return newArticle.Slug, nil
}

func (s *service) DeleteArticle(ctx context.Context, currentUserId, slug string) error {
	query := `
	update articles
	set deleted_at = now()
	where author_id = $1 and slug = $2
	`

	_, err := s.db.ExecContext(ctx, query, currentUserId, slug)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) CreateComment(
	ctx context.Context,
	currentUserId, slug, body string,
) (string, error) {
	query := `
		insert into comments (author_id, article_id, body)
		values (
		  $1,
		  (select id from articles where slug = $2 and deleted_at is null),
		  $3
		)
		returning id;
	`

	var commentId string
	err := s.db.QueryRowContext(ctx, query, currentUserId, slug, body).Scan(&commentId)
	if err != nil {
		return "", err
	}

	return commentId, nil
}

func (s *service) SelectComments(
	ctx context.Context,
	currentUserId, slug string,
	limit, offset int,
) (comments []CommentDetail, commentsCount int, err error) {
	query := `
		select c.id, c.body, c.created_at,
		  u.username, u.bio, u.image,
		  (select exists (
			select 1 from follows
			where follower_id::text = $1
			and following_id = c.author_id
		  )) as following,
		  count(*) over() as comments_count
		from comments c
		left join users u on u.id = c.author_id
		left join articles a on a.id = c.article_id
		where c.deleted_at is null
		and a.deleted_at is null
		and a.slug = $2
		group by a.id, u.id, c.id
		order by c.created_at
		limit $3
		offset $4;
	`

	/*
		example query output:
		                  id                  |               body               |          created_at           |    username    |      bio      |                     image                      | following | comments_count
		--------------------------------------+----------------------------------+-------------------------------+----------------+---------------+------------------------------------------------+-----------+----------------
		 4925ab3e-9faf-4d6c-9343-ac9291ac56d9 | asd1 created a comment           | 2025-10-07 11:37:18.00979+00  | minhhoccode111 | i like golang | https://www.w3schools.com/howto/img_avatar.png | t         |             12
	*/

	rows, err := s.db.QueryContext(ctx, query, currentUserId, slug, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	comments = []CommentDetail{}
	for rows.Next() {
		var c CommentDetail
		err := rows.Scan(
			&c.Id,
			&c.Body,
			&c.CreatedAt,
			&c.Author.Username,
			&c.Author.Bio,
			&c.Author.Image,
			&c.Author.Following,
			&commentsCount,
		)
		if err != nil {
			return nil, 0, err
		}

		comments = append(comments, c)
	}

	return comments, commentsCount, nil
}

func (s *service) SelectCommentDetail(
	ctx context.Context,
	currentUserId, commentId string,
) (*CommentDetail, error) {
	query := `
		select c.id, c.body, c.created_at,
		  u.username, u.bio, u.image,
		  (select exists (
			select 1 from follows
			where follower_id::text = $1
			and following_id = c.author_id
		  )) as following
		from comments c
		left join users u on u.id = c.author_id
		left join articles a on a.id = c.article_id
		where c.deleted_at is null
		and a.deleted_at is null
		and c.id = $2;
	`

	/*
		example query output:
		                  id                  |  body  |          created_at           | username | bio | image | following
		--------------------------------------+--------+-------------------------------+----------+-----+-------+-----------
		 da1b0dc3-e2a5-4930-9e5d-1dd6f7884717 | body 0 | 2025-10-06 13:45:43.116717+00 | asd0     |     |       | f
	*/

	commentDetail := CommentDetail{}
	err := s.db.QueryRowContext(ctx, query, currentUserId, commentId).Scan(
		&commentDetail.Id,
		&commentDetail.Body,
		&commentDetail.CreatedAt,
		&commentDetail.Author.Username,
		&commentDetail.Author.Bio,
		&commentDetail.Author.Image,
		&commentDetail.Author.Following,
	)
	if err != nil {
		return nil, err
	}

	return &commentDetail, nil
}

func (s *service) DeleteComment(ctx context.Context, currentUserId, slug, commentId string) error {
	query := `
		update comments
		set deleted_at = now()
		where author_id = $1
		and exists (
		  select 1 from articles
		  where id = article_id
		  and slug = $2
		  and deleted_at is null
		)
		and id = $3;
	`

	result, err := s.db.ExecContext(ctx, query, currentUserId, slug, commentId)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("zero rows affected")
	}

	return nil
}

func (s *service) IsFollowing(ctx context.Context, followerId, followingName string) (bool, error) {
	var following bool
	err := s.db.QueryRowContext(ctx, `
		select exists (
			select 1 from follows
			where follower_id = $1
			and following_id = (
				select id from users
				where username = $2
			)
		)`,
		followerId,
		followingName,
	).Scan(&following)
	if err != nil {
		return false, err
	}
	return following, nil
}

func (s *service) CreateFollow(
	ctx context.Context,
	followerId string,
	followingUsername string,
) error {
	_, err := s.db.ExecContext(ctx, `
		insert into follows (follower_id, following_id)
		values (
			$1,
			(select id from users where username = $2)
		)
		on conflict do nothing
		`,
		followerId,
		followingUsername,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteFollow(
	ctx context.Context,
	followerId string,
	followingUsername string,
) error {
	_, err := s.db.ExecContext(ctx, `
		delete from follows
		where follower_id = $1
		and following_id = (
			select id from users
			where username = $2
		)`,
		followerId,
		followingUsername,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) CreateFavorite(ctx context.Context, currentUserId, slug string) error {
	query := `
		insert into favorites (user_id, article_id)
		values (
		  $1,
		  (select id from articles where slug = $2 and deleted_at is null)
		);
	`

	_, err := s.db.QueryContext(ctx, query, currentUserId, slug)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteFavorite(ctx context.Context, currentUserId, slug string) error {
	query := `
		delete from favorites
		where user_id = $1
		and article_id = (
		  select id from articles
		  where slug = $2
		  and deleted_at is null
		);
	`

	_, err := s.db.QueryContext(ctx, query, currentUserId, slug)
	if err != nil {
		return err
	}

	return nil
}
