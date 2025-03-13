package e2e

import (
	"context"
	"livoir-blog/internal/domain"
	"time"
)

func (suite *E2ETestSuite) insertAdmin(fullName, email string) {
	err := suite.repoProvider.AdministratorRepository.Insert(context.Background(), &domain.Administrator{
		ID:           "idadmin",
		FullName:     fullName,
		Email:        email,
		PasswordHash: "hashed_password",
	})
	suite.Assert().Nil(err)
}

func (suite *E2ETestSuite) getAccessToken(email string) (string, error) {
	tokenData := &domain.TokenData{
		UserID:    "idadmin",
		Email:     email,
		IssuedAt:  time.Now().Unix(),
		ExpiredAt: time.Now().Add(time.Hour).Unix(),
	}
	return suite.repoProvider.TokenRepository.Generate(context.Background(), tokenData)
}
