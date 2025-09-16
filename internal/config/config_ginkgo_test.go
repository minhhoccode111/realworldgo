package config_test

import (
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"auth/internal/config"
)

var _ = Describe("Config", func() {
	var originalEnv map[string]string

	BeforeEach(func() {
		// Save original environment variables
		originalEnv = make(map[string]string)
		for _, e := range os.Environ() {
			pair := strings.SplitN(e, "=", 2)
			if len(pair) == 2 {
				originalEnv[pair[0]] = pair[1]
			}
		}
		// Clear relevant environment variables before each test
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_READ_TIMEOUT")
		os.Unsetenv("SERVER_WRITE_TIMEOUT")
		os.Unsetenv("SERVER_IDLE_TIMEOUT")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USERNAME")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_SSL_MODE")
		os.Unsetenv("DB_SCHEMA")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_EXPIRATION")
		os.Unsetenv("JWT_ISSUER")
		os.Unsetenv("ACCESS_CONTROL_ALLOW_ORIGIN")
	})

	AfterEach(func() {
		// Restore original environment variables
		for k := range originalEnv {
			os.Unsetenv(k)
		}
		for k, v := range originalEnv {
			os.Setenv(k, v)
		}
	})

	Describe("Load", func() {
		Context("when no environment variables are set", func() {
			It("loads default configuration values", func() {
				cfg, err := config.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg.Server.Host).To(Equal("localhost"))
				Expect(cfg.Server.Port).To(Equal(8000))
				Expect(cfg.Server.ReadTimeout).To(Equal(10 * time.Second))
				Expect(cfg.Server.WriteTimeout).To(Equal(10 * time.Second))
				Expect(cfg.Server.IdleTimeout).To(Equal(60 * time.Second))
				Expect(cfg.Database.Name).To(Equal("authz"))
				Expect(cfg.Database.Host).To(Equal("localhost"))
				Expect(cfg.Database.Port).To(Equal(5432))
				Expect(cfg.Database.Username).To(Equal("postgres"))
				Expect(cfg.Database.Password).To(Equal("Bruh0!0!"))
				Expect(cfg.Database.SSLMode).To(Equal("disable"))
				Expect(cfg.Database.Schema).To(Equal("public"))
				Expect(cfg.JWT.Secret).To(Equal("ai33yUUcmRPI64hq06ViG0404On-nMebsCtY4nTFqOg"))
				Expect(cfg.JWT.Expiration).To(Equal(24 * time.Hour))
				Expect(cfg.JWT.Issuer).To(Equal("myapp"))
				Expect(cfg.CORS.AllowedOrigins).To(ConsistOf("*"))
			})
		})

		Context("when environment variables are set", func() {
			BeforeEach(func() {
				os.Setenv("SERVER_HOST", "127.0.0.1")
				os.Setenv("SERVER_PORT", "9000")
				os.Setenv("SERVER_READ_TIMEOUT", "5s")
				os.Setenv("DB_NAME", "testdb")
				os.Setenv("JWT_SECRET", "newsecret")
				os.Setenv("JWT_EXPIRATION", "1h")
				os.Setenv("ACCESS_CONTROL_ALLOW_ORIGIN", "http://localhost:3000,https://example.com")
			})

			It("loads configuration values from environment variables", func() {
				cfg, err := config.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg.Server.Host).To(Equal("127.0.0.1"))
				Expect(cfg.Server.Port).To(Equal(9000))
				Expect(cfg.Server.ReadTimeout).To(Equal(5 * time.Second))
				Expect(cfg.Database.Name).To(Equal("testdb"))
				Expect(cfg.JWT.Secret).To(Equal("newsecret"))
				Expect(cfg.JWT.Expiration).To(Equal(1 * time.Hour))
				Expect(cfg.CORS.AllowedOrigins).To(ConsistOf("http://localhost:3000", "https://example.com"))
			})
		})

		Context("when JWT_SECRET is empty", func() {
			BeforeEach(func() {
				os.Setenv("JWT_SECRET", "")
			})

			It("returns an error", func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("JWT_SECRET is required"))
			})
		})

		Context("when DB_PASSWORD is empty", func() {
			BeforeEach(func() {
				os.Setenv("DB_PASSWORD", "")
			})

			It("returns an error", func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("DB_PASSWORD is required"))
			})
		})

		Context("when SERVER_PORT is invalid", func() {
			BeforeEach(func() {
				os.Setenv("SERVER_PORT", "99999")
			})

			It("returns an error", func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("SERVER_PORT must be between 1 and 65535"))
			})
		})

		Context("when ACCESS_CONTROL_ALLOW_ORIGIN contains invalid origin", func() {
			BeforeEach(func() {
				os.Setenv("ACCESS_CONTROL_ALLOW_ORIGIN", "invalid-url")
			})

			It("returns an error", func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("ACCESS_CONTROL_ALLOW_ORIGIN contains invalid origin"))
			})
		})
	})
})