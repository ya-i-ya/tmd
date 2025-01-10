package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"tmd/pkg/cfg"
)

type DB struct {
	Conn *gorm.DB
}

func NewDB(configuration *cfg.Config) (*DB, error) {
	if configuration.Database.Dialect != "postgres" {
		return nil, fmt.Errorf("unsupported database dialect: %s", configuration.Database.Dialect)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		configuration.Database.Host,
		configuration.Database.Port,
		configuration.Database.User,
		configuration.Database.Password,
		configuration.Database.DBName,
		configuration.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(&User{}, &Chat{}, &ChatUser{}, &Message{}); err != nil {
		return nil, fmt.Errorf("failed to auto migrate users: %w", err)
	}

	return &DB{Conn: db}, nil
}

func (db *DB) Shutdown() error {
	sqlDB, err := db.Conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
