package postgres

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host     string `yaml:"postgres_host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port     string `yaml:"postgres_port" env:"POSTGRES_PORT" env-default:"5434"`
	Database string `yaml:"postgres_db" env:"POSTGRES_DB" env-default:"postgres"`
	User     string `yaml:"postgres_user" env:"POSTGRES_USER" env-default:"root"`
	Password string `yaml:"postgres_password" env:"POSTGRES_PASSWORD" env-default:"1234"`
}

func New(config Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		config.Host,
		config.User,
		config.Password,
		config.Database,
		config.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("unable to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
