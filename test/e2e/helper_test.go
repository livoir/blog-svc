package e2e

import (
	"context"
	"livoir-blog/internal/domain"
)

func (suite *E2ETestSuite) insertAdmin(fullName, email string) {
	err := suite.repoProvider.AdministratorRepository.Insert(context.Background(), &domain.Administrator{
		FullName:     fullName,
		Email:        email,
		PasswordHash: "hashed_password",
	})
	suite.Assert().Nil(err)
}
