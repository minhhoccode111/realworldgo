package database_test

import (
	"context"
	"database/sql"
	"errors"

	"auth/internal/database"
	"auth/internal/model"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Database Service", func() {
	var (
		db   *sql.DB
		mock sqlmock.Sqlmock
		s    database.Service
		ctx  context.Context
		err  error
	)

	BeforeEach(func() {
		db, mock, err = sqlmock.New()
		Expect(err).ToNot(HaveOccurred(), "sqlmock.New should not return an error")
		s = database.NewService(db)
		ctx = context.Background()
	})

	AfterEach(func() {
		Expect(mock.ExpectationsWereMet()).ToNot(HaveOccurred(), "There were unfulfilled expectations")
	})

	Describe("Health", func() {
		Context("when the database is up", func() {
			BeforeEach(func() {
				mock.ExpectPing()
			})

			It("returns an 'up' status", func() {
				stats := s.Health()
				Expect(stats["status"]).To(Equal("up"))
				Expect(stats["message"]).To(Equal("It's healthy"))
			})
		})

		Context("when the database is down", func() {
			BeforeEach(func() {
				mock.ExpectPing().WillReturnError(errors.New("db connection error"))
			})

			PIt("returns a 'down' status and an error message", func() {
				stats := s.Health()
				Expect(stats["status"]).To(Equal("down"))
				Expect(stats["error"]).To(Equal("db down: db connection error"))
			})
		})
	})

	Describe("Close", func() {
		Context("when the database closes successfully", func() {
			BeforeEach(func() {
				mock.ExpectClose().WillReturnError(nil)
			})

			It("returns no error", func() {
				err := s.Close("testdb")
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the database fails to close", func() {
			BeforeEach(func() {
				mock.ExpectClose().WillReturnError(errors.New("close error"))
			})

			It("returns an error", func() {
				err := s.Close("testdb")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("close error"))
			})
		})
	})

	Describe("CountUsers", func() {
		Context("when isGetAll is true", func() {
			It("returns the correct count for a given filter", func() {
				filter := "test"
				isGetAll := true
				expectedCount := 5

				mock.ExpectQuery(`
				select count\(\*\) from users
				where email ilike '%' \|\| \$1 \|\| '%'
				`).WithArgs(filter).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

				count, err := s.CountUsers(ctx, filter, isGetAll)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(expectedCount))
			})
		})

		Context("when isGetAll is false", func() {
			It("returns the correct count for active users with a given filter", func() {
				filter := "test"
				isGetAll := false
				expectedCount := 3

				mock.ExpectQuery(`
				select count\(\*\) from users
				where email ilike '%' \|\| \$1 \|\| '%'
				and is_active = true
				`).WithArgs(filter).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

				count, err := s.CountUsers(ctx, filter, isGetAll)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(expectedCount))
			})
		})

		Context("when a database error occurs", func() {
			It("returns an error", func() {
				filter := "test"
				isGetAll := true

				mock.ExpectQuery(`
				select count\(\*\) from users
				where email ilike '%' \|\| \$1 \|\| '%'
				`).WithArgs(filter).WillReturnError(errors.New("db error"))

				_, err := s.CountUsers(ctx, filter, isGetAll)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("db error"))
			})
		})
	})

	Describe("SelectUsers", func() {
		Context("when isGetAll is true", func() {
			It("returns the correct list of users", func() {
				filter := "test"
				limit := 2
				offset := 0
				isGetAll := true
				expectedUsers := []*model.UserDTO{
					{Id: "1", Email: "test1@example.com", IsActive: true, Role: model.RoleUser},
					{Id: "2", Email: "test2@example.com", IsActive: true, Role: model.RoleAdmin},
				}

				rows := sqlmock.NewRows([]string{"id", "email", "is_active", "role"}).
					AddRow("1", "test1@example.com", true, model.RoleUser).
					AddRow("2", "test2@example.com", true, model.RoleAdmin)

				mock.ExpectQuery(`
				select id, email, is_active, role from users
				where email ilike '%' \|\| \$1 \|\| '%'
				limit \$2 offset \$3
				`).WithArgs(filter, limit, offset).WillReturnRows(rows)

				users, err := s.SelectUsers(ctx, limit, offset, filter, isGetAll)
				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(HaveLen(len(expectedUsers)))
				Expect(users).To(ConsistOf(expectedUsers))
			})
		})

		Context("when isGetAll is false", func() {
			It("returns the correct list of active users", func() {
				filter := "test"
				limit := 1
				offset := 0
				isGetAll := false
				expectedUsers := []*model.UserDTO{
					{Id: "1", Email: "test1@example.com", IsActive: true, Role: model.RoleUser},
				}

				rows := sqlmock.NewRows([]string{"id", "email", "is_active", "role"}).
					AddRow("1", "test1@example.com", true, model.RoleUser)

				mock.ExpectQuery(`
				select id, email, is_active, role from users
				where email ilike '%' \|\| \$1 \|\| '%'
				and is_active = true
				`).WithArgs(filter, limit, offset).WillReturnRows(rows)

				users, err := s.SelectUsers(ctx, limit, offset, filter, isGetAll)
				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(HaveLen(len(expectedUsers)))
				Expect(users).To(ConsistOf(expectedUsers))
			})
		})

		Context("when a database error occurs", func() {
			It("returns an error", func() {
				filter := "test"
				limit := 1
				offset := 0
				isGetAll := true

				mock.ExpectQuery(`
				select id, email, is_active, role from users
				where email ilike '%' \|\| \$1 \|\| '%'
				limit \$2 offset \$3
				`).WithArgs(filter, limit, offset).WillReturnError(errors.New("db error"))

				_, err := s.SelectUsers(ctx, limit, offset, filter, isGetAll)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("db error"))
			})
		})
	})

	Describe("SelectUserById", func() {
		Context("when the user exists", func() {
			It("returns the user", func() {
				id := "123"
				expectedUser := &model.User{
					Id:       "123",
					Email:    "test@example.com",
					IsActive: true,
					Role:     model.RoleUser,
					Password: "hashedpassword",
				}

				rows := sqlmock.NewRows([]string{"id", "email", "is_active", "role", "password"}).
					AddRow(expectedUser.Id, expectedUser.Email, expectedUser.IsActive, expectedUser.Role, expectedUser.Password)

				mock.ExpectQuery(`
				select id, email, is_active, role, password
				from users
				where id = \$1
				`).WithArgs(id).WillReturnRows(rows)

				user, err := s.SelectUserById(ctx, id)
				Expect(err).ToNot(HaveOccurred())
				Expect(user).To(Equal(expectedUser))
			})
		})

		Context("when the user does not exist", func() {
			It("returns sql.ErrNoRows", func() {
				id := "nonexistent"

				mock.ExpectQuery(`
				select id, email, is_active, role, password
				from users
				where id = \$1
				`).WithArgs(id).WillReturnError(sql.ErrNoRows)

				_, err := s.SelectUserById(ctx, id)
				Expect(err).To(MatchError(sql.ErrNoRows))
			})
		})

		Context("when a database error occurs", func() {
			It("returns an error", func() {
				id := "123"

				mock.ExpectQuery(`
				select id, email, is_active, role, password
				from users
				where id = \$1
				`).WithArgs(id).WillReturnError(errors.New("db error"))

				_, err := s.SelectUserById(ctx, id)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("db error"))
			})
		})
	})

	Describe("SelectUserByEmail", func() {
		Context("when the user exists", func() {
			It("returns the user", func() {
				email := "test@example.com"
				expectedUser := &model.User{
					Id:       "123",
					Email:    "test@example.com",
					IsActive: true,
					Role:     model.RoleUser,
					Password: "hashedpassword",
				}

				rows := sqlmock.NewRows([]string{"id", "email", "is_active", "role", "password"}).
					AddRow(expectedUser.Id, expectedUser.Email, expectedUser.IsActive, expectedUser.Role, expectedUser.Password)

				mock.ExpectQuery(`
				select id, email, is_active, role, password
				from users
				where email = \$1
				`).WithArgs(email).WillReturnRows(rows)

				user, err := s.SelectUserByEmail(ctx, email)
				Expect(err).ToNot(HaveOccurred())
				Expect(user).To(Equal(expectedUser))
			})
		})

		Context("when the user does not exist", func() {
			It("returns sql.ErrNoRows", func() {
				email := "nonexistent@example.com"

				mock.ExpectQuery(`
				select id, email, is_active, role, password
				from users
				where email = \$1
				`).WithArgs(email).WillReturnError(sql.ErrNoRows)

				_, err := s.SelectUserByEmail(ctx, email)
				Expect(err).To(MatchError(sql.ErrNoRows))
			})
		})

		Context("when a database error occurs", func() {
			It("returns an error", func() {
				email := "test@example.com"

				mock.ExpectQuery(`
				select id, email, is_active, role, password
				from users
				where email = \$1
				`).WithArgs(email).WillReturnError(errors.New("db error"))

				_, err := s.SelectUserByEmail(ctx, email)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("db error"))
			})
		})
	})

	Describe("InsertUser", func() {
		Context("when the insert is successful", func() {
			It("inserts the user and returns no error", func() {
				user := &model.User{
					Email:    "newuser@example.com",
					IsActive: true,
					Role:     model.RoleUser,
					Password: "Password123!",
				}
				expectedID := "generated-id-123"

				mock.ExpectQuery(`
				insert into users\(email, is_active, role, password\)
				values\(\$1, \$2, \$3, \$4\)
				returning id
				`).WithArgs(user.Email, user.IsActive, user.Role, sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

				err := s.InsertUser(ctx, user)
				Expect(err).ToNot(HaveOccurred())
				Expect(user.Id).To(Equal(expectedID))
			})
		})

		Context("when a database error occurs", func() {
			It("returns an error", func() {
				user := &model.User{
					Email:    "error@example.com",
					IsActive: true,
					Role:     model.RoleUser,
					Password: "Password123!",
				}

				mock.ExpectQuery(`
				insert into users\(email, is_active, role, password\)
				values\(\$1, \$2, \$3, \$4\)
				returning id
				`).WithArgs(user.Email, user.IsActive, user.Role, sqlmock.AnyArg()).WillReturnError(errors.New("db error"))

				err := s.InsertUser(ctx, user)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("db error"))
			})
		})
	})

	Describe("UpdateUser", func() {
		Context("when the update is successful", func() {
			It("updates the user and returns the updated DTO", func() {
				id := "123"
				newEmail := "updated@example.com"
				expectedUserDTO := &model.UserDTO{
					Id:       "123",
					Email:    "updated@example.com",
					IsActive: true,
					Role:     model.RoleUser,
				}

				rows := sqlmock.NewRows([]string{"id", "role", "email", "is_active"}).
					AddRow(expectedUserDTO.Id, expectedUserDTO.Role, expectedUserDTO.Email, expectedUserDTO.IsActive)

				mock.ExpectQuery(`
				update users
				set email = \$1
				where id = \$2
				returning id, role, email, is_active
				`).WithArgs(newEmail, id).WillReturnRows(rows)

				userDTO, err := s.UpdateUser(ctx, id, newEmail)
				Expect(err).ToNot(HaveOccurred())
				Expect(userDTO).To(Equal(expectedUserDTO))
			})
		})

		Context("when a database error occurs", func() {
			It("returns an error", func() {
				id := "123"
				newEmail := "updated@example.com"

				mock.ExpectQuery(`
				update users
				set email = \$1
				where id = \$2
				returning id, role, email, is_active
				`).WithArgs(newEmail, id).WillReturnError(errors.New("db error"))

				_, err := s.UpdateUser(ctx, id, newEmail)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("db error"))
			})
		})
	})

	Describe("UpdateUserPassword", func() {
		Context("when the update is successful", func() {
			It("updates the user's password and returns no error", func() {
				id := "123"
				newPassword := "NewPassword123!"

				mock.ExpectExec(`
				update users
				set password = \$1
				where id = \$2
				`).WithArgs(sqlmock.AnyArg(), id).WillReturnResult(sqlmock.NewResult(1, 1))

				err := s.UpdateUserPassword(ctx, id, newPassword)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the user is not found", func() {
			It("returns sql.ErrNoRows", func() {
				id := "nonexistent"
				newPassword := "NewPassword123!"

				mock.ExpectExec(`
				update users
				set password = \$1
				where id = \$2

				`).WithArgs(sqlmock.AnyArg(), id).WillReturnResult(sqlmock.NewResult(1, 0))

				err := s.UpdateUserPassword(ctx, id, newPassword)
				Expect(err).To(MatchError(sql.ErrNoRows))
			})
		})

		Context("when a database error occurs", func() {
			It("returns an error", func() {
				id := "123"
				newPassword := "NewPassword123!"

				mock.ExpectExec(`
				update users
				set password = \$1
				where id = \$2
				`).WithArgs(sqlmock.AnyArg(), id).WillReturnError(errors.New("db error"))

				err := s.UpdateUserPassword(ctx, id, newPassword)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("db error"))
			})
		})
	})

	Describe("UpdateUserStatus", func() {
		Context("when the update is successful", func() {
			It("updates the user's status to active", func() {
				id := "123"
				isActive := true

				mock.ExpectExec(`
				update users
				set is_active = \$1
				where id = \$2
				`).WithArgs(isActive, id).WillReturnResult(sqlmock.NewResult(1, 1))

				err := s.UpdateUserStatus(ctx, id, isActive)
				Expect(err).ToNot(HaveOccurred())
			})

			It("updates the user's status to inactive", func() {
				id := "123"
				isActive := false

				mock.ExpectExec(`
				update users
				set is_active = \$1
				where id = \$2
				`).WithArgs(isActive, id).WillReturnResult(sqlmock.NewResult(1, 1))

				err := s.UpdateUserStatus(ctx, id, isActive)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the user is not found", func() {
			It("returns sql.ErrNoRows", func() {
				id := "nonexistent"
				isActive := true

				mock.ExpectExec(`
				update users
				set is_active = \$1
				where id = \$2
				`).WithArgs(isActive, id).WillReturnResult(sqlmock.NewResult(1, 0))

				err := s.UpdateUserStatus(ctx, id, isActive)
				Expect(err).To(MatchError(sql.ErrNoRows))
			})
		})

		Context("when a database error occurs", func() {
			It("returns an error", func() {
				id := "123"
				isActive := true

				mock.ExpectExec(`
				update users
				set is_active = \$1
				where id = \$2
				`).WithArgs(isActive, id).WillReturnError(errors.New("db error"))

				err := s.UpdateUserStatus(ctx, id, isActive)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("db error"))
			})
		})
	})

	Describe("DeleteUserById", func() {
		Context("when the delete is successful", func() {
			It("deletes the user and returns no error", func() {
				id := "123"

				mock.ExpectExec(`
				delete from users
				where id = \$1
				`).WithArgs(id).WillReturnResult(sqlmock.NewResult(1, 1))

				err := s.DeleteUserById(ctx, id)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the user is not found", func() {
			It("returns sql.ErrNoRows", func() {
				id := "nonexistent"

				mock.ExpectExec(`
				delete from users
				where id = \$1
				`).WithArgs(id).WillReturnResult(sqlmock.NewResult(1, 0))

				err := s.DeleteUserById(ctx, id)
				Expect(err).To(MatchError(sql.ErrNoRows))
			})
		})

		Context("when a database error occurs", func() {
			It("returns an error", func() {
				id := "123"

				mock.ExpectExec(`
				delete from users
				where id = \$1
				`).WithArgs(id).WillReturnError(errors.New("db error"))

				err := s.DeleteUserById(ctx, id)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("db error"))
			})
		})
	})
})