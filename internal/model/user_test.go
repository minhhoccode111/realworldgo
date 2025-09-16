package model_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"auth/internal/model"
)

var _ = Describe("User Model", func() {
	Describe("Role Constants", func() {
		It("defines RoleAdmin as \"admin\"", func() {
			Expect(model.RoleAdmin).To(Equal(model.Role("admin")))
		})

		It("defines RoleUser as \"user\"", func() {
			Expect(model.RoleUser).To(Equal(model.Role("user")))
		})
	})
})