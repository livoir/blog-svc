package e2e

import (
	"context"
	"livoir-blog/internal/domain"
)

func (suite *E2ETestSuite) insertAdmin() {
	err := suite.repoProvider.AdministratorRepository.Insert(context.Background(), &domain.Administrator{
		FullName:     "admin",
		Email:        "admin@example.com",
		PasswordHash: "hashedPassword",
	})
	suite.Assert().Nil(err)
}
