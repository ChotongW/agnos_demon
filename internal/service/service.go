package service

import (
	"agnos_demo/internal/database"

	"github.com/sirupsen/logrus"
)

type Service struct {
	Logger *logrus.Logger
	DB     database.DB
}

type ServiceOptions struct {
}

func NewService(logger *logrus.Logger, db database.DB, opts *ServiceOptions) (*Service, error) {
	return &Service{
		Logger: logger,
		DB:     db,
	}, nil
}
