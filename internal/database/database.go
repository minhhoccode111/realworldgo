package database

import (
	. "auth/internal/model"
	. "auth/internal/utils"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

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

	// CountUsers returns the number of users in the database for pagination.
	CountUsers(ctx context.Context, filter string, isGetAll bool) (int, error)

	// SelectUsers returns a slice of users from the database for pagination.
	SelectUsers(ctx context.Context, limit, offset int, filter string, isGetAll bool) ([]*UserDTO, error)

	// NOTE: GetUserById and GetUserByEmail have to return User model because sometimes we need password to update user

	// SelectUserById returns a user from the database by its ID.
	SelectUserById(ctx context.Context, id string) (*User, error)
	// SelectUserByEmail returns a user from the database by its email.
	SelectUserByEmail(ctx context.Context, email string) (*User, error)

	// InsertUser inserts a new user into the database.
	InsertUser(ctx context.Context, user *User) error

	// UpdateUser updates the email of a user in the database.
	UpdateUser(ctx context.Context, id string, email string) (*UserDTO, error)

	// UpdateUserPassword updates the password of a user in the database.
	UpdateUserPassword(ctx context.Context, id string, password string) error

	// UpdateUserStatus updates the status of a user in the database.
	UpdateUserStatus(ctx context.Context, id string, isActive bool) error

	// DeleteUserById deletes a user from the database by its ID.
	DeleteUserById(ctx context.Context, id string) error
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

func (s *service) CountUsers(ctx context.Context, filter string, isGetAll bool) (int, error) {
	var count int
	var err error
	if isGetAll {
		err = s.db.QueryRowContext(ctx, `
		select count(*) from users
		where email ilike '%' || $1 || '%'
		`, filter).Scan(&count)
	} else {
		err = s.db.QueryRowContext(ctx, `
		select count(*) from users
		where email ilike '%' || $1 || '%'
		and is_active = true
		`, filter).Scan(&count)
	}
	if err != nil {
		log.Printf("Database error: %v", err)
		return 0, fmt.Errorf("Datebase error when count users: %v", err)
	}
	return count, nil
}

func (s *service) SelectUsers(ctx context.Context, limit int, offset int, filter string, isGetAll bool) ([]*UserDTO, error) {
	var rows *sql.Rows
	var err error
	if isGetAll {
		rows, err = s.db.QueryContext(ctx, `
		select id, email, is_active, role from users
		where email ilike '%' || $1 || '%'
		limit $2 offset $3
		`, filter, limit, offset)
	} else {
		rows, err = s.db.QueryContext(ctx, `
		select id, email, is_active, role from users
		where email ilike '%' || $1 || '%'
		and is_active = true
		limit $2 offset $3
		`, filter, limit, offset)
	}
	if err != nil {
		log.Printf("Error select users: %v", err)
		return nil, fmt.Errorf("Error select users: %v", err)
	}
	defer rows.Close()
	var users = []*UserDTO{}
	for rows.Next() {
		var user UserDTO
		err := rows.Scan(&user.Id, &user.Email, &user.IsActive, &user.Role)
		if err != nil {
			log.Printf("Error Scan UserDTO: %v", err)
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

func (s *service) SelectUserById(ctx context.Context, id string) (*User, error) {
	var user User
	err := s.db.QueryRowContext(ctx, `
		select id, email, is_active, role, password
		from users
		where id = $1
		`,
		id,
	).
		Scan(&user.Id, &user.Email, &user.IsActive, &user.Role, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *service) SelectUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := s.db.QueryRowContext(ctx, `
		select id, email, is_active, role, password
		from users
		where email = $1
		`, email).
		Scan(&user.Id, &user.Email, &user.IsActive, &user.Role, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *service) InsertUser(ctx context.Context, user *User) error {
	hashedPassword, err := HashedPassword(user.Password)
	if err != nil {
		log.Printf("Error hashing %v: %v", user.Password, err)
		return fmt.Errorf("Error hashing %v: %v", user.Password, err)
	}
	row := s.db.QueryRowContext(ctx, `
		insert into users(email, is_active, role, password)
		values($1, $2, $3, $4)
		returning id
		`,
		user.Email,
		user.IsActive,
		user.Role,
		hashedPassword,
	)
	// pass generated id back to user
	if err := row.Scan(&user.Id); err != nil {
		log.Printf("Error insert user: %v", err)
		return fmt.Errorf("Error insert user: %v", err)
	}
	return nil
}

func (s *service) UpdateUser(ctx context.Context, id string, email string) (*UserDTO, error) {
	result := s.db.QueryRowContext(ctx, `
		update users
		set email = $1
		where id = $2
		returning id, role, email, is_active
		`,
		email,
		id,
	)
	var updatedUser UserDTO
	if err := result.Scan(&updatedUser.Id, &updatedUser.Role, &updatedUser.Email, &updatedUser.IsActive); err != nil {
		log.Printf("Error update user: %v", err)
		return nil, err
	}
	return &updatedUser, nil
}

func (s *service) UpdateUserPassword(ctx context.Context, id string, password string) error {
	hashedPassword, err := HashedPassword(password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return fmt.Errorf("Error hashing password: %v", err)
	}
	result, err := s.db.ExecContext(ctx, `
		update users
		set password = $1
		where id = $2
		`,
		hashedPassword,
		id,
	)
	if err != nil {
		log.Printf("database error when update user password: %v", err)
		return fmt.Errorf("database error when update user password: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("failed to get rows affected: %v", err)
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		log.Printf("error when update user password: %v", sql.ErrNoRows)
		return sql.ErrNoRows
	}
	return nil
}

func (s *service) UpdateUserStatus(ctx context.Context, id string, isActive bool) error {
	result, err := s.db.ExecContext(ctx, `
		update users
		set is_active = $1
		where id = $2
		`,
		isActive,
		id,
	)
	if err != nil {
		log.Printf("database error update user status: %v", err)
		return fmt.Errorf("database error update user status: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("failed to get rows affected: %v", err)
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		log.Printf("error when update user password: %v", sql.ErrNoRows)
		return sql.ErrNoRows
	}
	return nil
}

func (s *service) DeleteUserById(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `
		delete from users
		where id = $1
		`,
		id,
	)
	if err != nil {
		log.Printf("database error delete user by id: %v", err)
		return fmt.Errorf("database error delete user by id: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("failed to get rows affected: %v", err)
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		log.Printf("error when delete user: %v", sql.ErrNoRows)
		return sql.ErrNoRows
	}
	return nil
}
