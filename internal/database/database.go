package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"github.com/minhhoccode111/realworldgo/internal/model"
	"github.com/minhhoccode111/realworldgo/internal/utils"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close(dbName string) error

	// SelectUserById returns a user from the database by its ID.
	SelectUserById(ctx context.Context, id string) (*model.User, error)

	// SelectUserByEmail returns a user from the database by its email.
	SelectUserByEmail(ctx context.Context, email string) (*model.User, error)

	// CreateUser inserts a new user into the database.
	CreateUser(ctx context.Context, newUser *model.User) error

	// UpdateUser updates the email of a user in the database.
	UpdateUser(ctx context.Context, newUser *model.User) error

	// IsSlugExisted checks if an article with the given slug already exists in the database
	IsSlugExisted(ctx context.Context, slug string) (bool, error)

	// CreateArticle inserts a new user into the database.
	CreateArticle(ctx context.Context, newArticle *model.Article, tags []string) error

	// IsFollowing checks if the follower is following the followingUsername
	IsFollowing(ctx context.Context, followerId string, followingUsername string) (bool, error)
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

func (s *service) SelectUserById(ctx context.Context, userId string) (*model.User, error) {
	var user model.User
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, email, username, password, bio, image, created_at, updated_at
		FROM users WHERE id = $1 `, userId,
	).Scan(
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

func (s *service) SelectUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, email, username, password, bio, image, created_at, updated_at
		FROM users WHERE email = $1 `, email,
	).Scan(
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

func (s *service) CreateUser(ctx context.Context, newUser *model.User) error {
	hashedPassword, err := utils.HashedPassword(newUser.Password)
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

func (s *service) UpdateUser(ctx context.Context, newUser *model.User) error {
	hashedPassword, err := utils.HashedPassword(newUser.Password)
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

func (s *service) IsSlugExisted(ctx context.Context, slug string) (bool, error) {
	var existed bool
	err := s.db.QueryRowContext(ctx, `
		select exists (
		select 1 from articles where slug = $1
		)
		`).Scan(&existed)
	if err != nil {
		return false, err
	}
	return existed, nil
}

func (s *service) CreateArticle(
	ctx context.Context,
	newArticle *model.Article,
	tags []string,
) (err error) {
	// 1. insert new article to db to generate id
	// 2. create a list of tags, will return error if unique constraint fail
	// 3. create rows in junction table between article and tags
	// we can apply concurrency for 1. and 2.
	// and we also need ACID transaction to make sure every query succeed

	var tx *sql.Tx
	tx, err = s.db.Begin()
	if err != nil {
		return err
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
		existed, err = s.IsSlugExisted(ctx, newArticle.Slug)
		if err != nil {
			return err
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
		return err
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
			return err
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
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *service) IsFollowing(
	ctx context.Context,
	followerId string,
	followingUsername string,
) (bool, error) {
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
		followingUsername,
	).Scan(&following)
	if err != nil {
		return false, err
	}
	return following, nil
}
