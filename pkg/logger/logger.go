package logger

import (
	"errors"

	"go.uber.org/zap"
)

var Log *zap.Logger

func Init() error {
	var err error
	Log, err = zap.NewProduction()
	if err != nil {
		return err
	}
	return nil
}

func Sync() error {
	if Log == nil {
		return errors.New("logger not initialized")
	}
	return Log.Sync()
}
