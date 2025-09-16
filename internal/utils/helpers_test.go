package utils_test

import (
	"auth/internal/config"
	"auth/internal/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"auth/internal/utils"
)

var _ = Describe("Helpers", func() {
	Describe("IsValidEmail", func() {
		Context("when the email is valid", func() {
			It("returns the email and no error for a simple valid email", func() {
				email := "test@example.com"
				got, err := utils.IsValidEmail(email)
				Expect(err).ToNot(HaveOccurred())
				Expect(got).To(Equal(email))
			})

			It("returns the email and no error for an email with a plus sign", func() {
				email := "test+alias@example.com"
				got, err := utils.IsValidEmail(email)
				Expect(err).ToNot(HaveOccurred())
				Expect(got).To(Equal(email))
			})

			It("returns the email and no error for an email with a dot in the local part", func() {
				email := "first.last@example.com"
				got, err := utils.IsValidEmail(email)
				Expect(err).ToNot(HaveOccurred())
				Expect(got).To(Equal(email))
			})

			It("returns the email and no error for an email with a subdomain", func() {
				email := "test@sub.example.com"
				got, err := utils.IsValidEmail(email)
				Expect(err).ToNot(HaveOccurred())
				Expect(got).To(Equal(email))
			})

			It("trims leading/trailing spaces and returns the valid email", func() {
				email := "  test@example.com  "
				expectedEmail := "test@example.com"
				got, err := utils.IsValidEmail(email)
				Expect(err).ToNot(HaveOccurred())
				Expect(got).To(Equal(expectedEmail))
			})
		})

		Context("when the email is invalid", func() {
			It("returns an error for an empty email", func() {
				email := ""
				got, err := utils.IsValidEmail(email)
				Expect(err).To(HaveOccurred())
				Expect(got).To(BeEmpty())
			})

			It("returns an error for an email without @", func() {
				email := "testexample.com"
				got, err := utils.IsValidEmail(email)
				Expect(err).To(HaveOccurred())
				Expect(got).To(BeEmpty())
			})

			It("returns an error for an email without a domain", func() {
				email := "test@"
				got, err := utils.IsValidEmail(email)
				Expect(err).To(HaveOccurred())
				Expect(got).To(BeEmpty())
			})

			It("returns an error for an email without a top-level domain", func() {
				email := "test@example"
				got, err := utils.IsValidEmail(email)
				Expect(err).To(HaveOccurred())
				Expect(got).To(BeEmpty())
			})

			It("returns an error for an email with invalid characters", func() {
				email := "test!@example.com"
				got, err := utils.IsValidEmail(email)
				Expect(err).To(HaveOccurred())
				Expect(got).To(BeEmpty())
			})
		})
	})

	Describe("UserToUserDTO", func() {
		It("correctly converts a User model to a UserDTO", func() {
			user := &model.User{
				Id:       "123",
				Email:    "test@example.com",
				IsActive: true,
				Role:     model.RoleUser,
			}

			dto := utils.UserToUserDTO(user)

			Expect(dto.Id).To(Equal(user.Id))
			Expect(dto.Email).To(Equal(user.Email))
			Expect(dto.IsActive).To(Equal(user.IsActive))
			Expect(dto.Role).To(Equal(user.Role))
			
		})
	})

	Describe("HashedPassword and ValidatePassword", func() {
		var password string
		BeforeEach(func() {
			password = "mysecretpassword"
		})

		It("successfully hashes a password", func() {
			hashedPassword, err := utils.HashedPassword(password)
			Expect(err).ToNot(HaveOccurred())
			Expect(hashedPassword).ToNot(BeEmpty())
		})

		It("validates a correct password", func() {
			hashedPassword, _ := utils.HashedPassword(password)
			Expect(utils.ValidatePassword(hashedPassword, password)).To(BeTrue())
		})

		It("does not validate an incorrect password", func() {
			hashedPassword, _ := utils.HashedPassword(password)
			Expect(utils.ValidatePassword(hashedPassword, "wrongpassword")).To(BeFalse())
		})

		It("does not validate with an empty plain password", func() {
			hashedPassword, _ := utils.HashedPassword(password)
			Expect(utils.ValidatePassword(hashedPassword, "")).To(BeFalse())
		})

		It("does not validate with an empty hashed password", func() {
			Expect(utils.ValidatePassword("", password)).To(BeFalse())
		})
	})

	Describe("IsValidPassword", func() {
		Context("when the password is valid", func() {
			It("returns the password and no error for a strong password", func() {
				password := "Password123!"
				got, err := utils.IsValidPassword(password)
				Expect(err).ToNot(HaveOccurred())
				Expect(got).To(Equal(password))
			})

			It("trims leading/trailing spaces and returns the valid password", func() {
				password := "  Password123!  "
				expectedPassword := "Password123!"
				got, err := utils.IsValidPassword(password)
				Expect(err).ToNot(HaveOccurred())
				Expect(got).To(Equal(expectedPassword))
			})
		})

		Context("when the password is invalid", func() {
			It("returns an error for a password that is too short", func() {
				password := "Pass1!"
				got, err := utils.IsValidPassword(password)
				Expect(err).To(HaveOccurred())
				Expect(got).To(BeEmpty())
			})

			It("returns an error for a password with no uppercase letter", func() {
				password := "password123!"
				got, err := utils.IsValidPassword(password)
				Expect(err).To(HaveOccurred())
				Expect(got).To(BeEmpty())
			})

			It("returns an error for a password with no lowercase letter", func() {
				password := "PASSWORD123!"
				got, err := utils.IsValidPassword(password)
				Expect(err).To(HaveOccurred())
				Expect(got).To(BeEmpty())
			})

			It("returns an error for a password with no digit", func() {
				password := "Password!!"
				got, err := utils.IsValidPassword(password)
				Expect(err).To(HaveOccurred())
				Expect(got).To(BeEmpty())
			})

			It("returns an error for a password with no special character", func() {
				password := "Password123"
				got, err := utils.IsValidPassword(password)
				Expect(err).To(HaveOccurred())
				Expect(got).To(BeEmpty())
			})

			It("returns an error for an empty password", func() {
				password := ""
				got, err := utils.IsValidPassword(password)
				Expect(err).To(HaveOccurred())
				Expect(got).To(BeEmpty())
			})
		})
	})

	Describe("GenerateJWT", func() {
		var ( 
			jwtConfig config.JWTConfig
			userDTO   *model.UserDTO
		)

		BeforeEach(func() {
			jwtConfig = config.JWTConfig{
				Secret:     "supersecretjwtkey",
				Expiration: 3600, // 1 hour
				Issuer:     "test-issuer",
			}

			userDTO = &model.UserDTO{
				Id:       "user123",
				Email:    "user@example.com",
				IsActive: true,
				Role:     model.RoleUser,
			}
		})

		It("generates a valid JWT token", func() {
			tokenString, err := utils.GenerateJWT(jwtConfig, userDTO)
			Expect(err).ToNot(HaveOccurred())
			Expect(tokenString).ToNot(BeEmpty())

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				Expect(token.Method).To(BeAssignableToTypeOf(&jwt.SigningMethodHMAC{}), "Unexpected signing method")
				return []byte(jwtConfig.Secret), nil
			}, jwt.WithLeeway(time.Minute))

			Expect(err).ToNot(HaveOccurred())
			Expect(token.Valid).To(BeTrue(), "Generated token should be valid")

			claims, ok := token.Claims.(jwt.MapClaims)
			Expect(ok).To(BeTrue(), "Could not get claims from token")

			Expect(claims["userId"]).To(Equal(userDTO.Id))
			Expect(claims["iss"]).To(Equal(jwtConfig.Issuer))

			exp := int64(claims["exp"].(float64))
			Expect(exp).To(BeNumerically(">=", time.Now().Unix()), "Token expiration time should be in the future")
		})

		It("returns an error if JWT secret is empty", func() {
			jwtConfig.Secret = ""
			tokenString, err := utils.GenerateJWT(jwtConfig, userDTO)
			Expect(err).To(HaveOccurred())
			Expect(tokenString).To(BeEmpty())
		})
	})
})