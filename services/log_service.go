package services

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogService struct {
	logger *zap.Logger
}

func NewLogService(level string, filePath string) (*LogService, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{filePath}
	
	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, err
	}
	config.Level = zap.NewAtomicLevelAt(logLevel)

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &LogService{logger: logger}, nil
}

func (s *LogService) Info(msg string, fields ...zap.Field) {
	s.logger.Info(msg, fields...)
}

func (s *LogService) Error(msg string, fields ...zap.Field) {
	s.logger.Error(msg, fields...)
}

func (s *LogService) Debug(msg string, fields ...zap.Field) {
	s.logger.Debug(msg, fields...)
}

func (s *LogService) Sync() error {
	return s.logger.Sync()
} 