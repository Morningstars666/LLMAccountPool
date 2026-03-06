package models

import (
	"fmt"
	"llmaccountpool/config"
	"llmaccountpool/utils"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func getDefaultDatabaseDSN(dbType config.DatabaseType) string {
	host := config.GetEnv("DB_HOST", "localhost")
	port := config.GetEnv("DB_PORT", "")
	user := config.GetEnv("DB_USER", "postgres")
	password := config.GetEnv("DB_PASSWORD", "")
	sslmode := config.GetEnv("DB_SSLMODE", "disable")

	if dbType == config.PostgresType {
		if port == "" {
			port = "5432"
		}
		return fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s TimeZone=UTC",
			host, port, user, password, sslmode,
		)
	}

	if dbType == config.MySQLType {
		if port == "" {
			port = "3306"
		}
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=UTC",
			user, password, host, port,
		)
	}

	return ""
}

func ensureDatabase(cfg *config.Config) {
	if cfg.DatabaseType == config.SQLiteType {
		return
	}

	defaultDSN := getDefaultDatabaseDSN(cfg.DatabaseType)
	defaultDB, err := gorm.Open(getGormDialector(cfg.DatabaseType, defaultDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to default database: %v", err)
	}

	dbName := config.GetEnv("DB_NAME", "llmaccountpool")

	var count int64
	if cfg.DatabaseType == config.PostgresType {
		defaultDB.Raw("SELECT COUNT(*) FROM pg_database WHERE datname = ?", dbName).Scan(&count)
		if count == 0 {
			defaultDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, dbName))
			log.Printf("Database '%s' created successfully", dbName)
		}
	} else if cfg.DatabaseType == config.MySQLType {
		defaultDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
		log.Printf("Database '%s' ensured", dbName)
	}

	sqlDB, _ := defaultDB.DB()
	sqlDB.Close()
}

func getGormDialector(dbType config.DatabaseType, dsn string) gorm.Dialector {
	switch dbType {
	case config.PostgresType:
		return postgres.Open(dsn)
	case config.MySQLType:
		return mysql.Open(dsn)
	case config.SQLiteType:
		fallthrough
	default:
		return sqlite.Open(dsn)
	}
}

func InitDB(cfg *config.Config) {
	var err error

	log.Printf("Using database type: %s", cfg.DatabaseType)

	ensureDatabase(cfg)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	DB, err = gorm.Open(getGormDialector(cfg.DatabaseType, cfg.DatabaseURL), gormConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	switch cfg.DatabaseType {
	case config.PostgresType, config.MySQLType:
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)

		if cfg.DatabaseType == config.PostgresType {
			sqlDB.Exec("SET default_transaction_isolation = 'read committed'")
			log.Println("PostgreSQL transaction isolation set to READ COMMITTED")
		} else if cfg.DatabaseType == config.MySQLType {
			sqlDB.Exec("SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED")
			log.Println("MySQL transaction isolation set to READ COMMITTED")
		}

		log.Printf("Connection pool configured: max_open=%d, max_idle=%d, lifetime=%ds, idle_time=%ds",
			cfg.MaxOpenConns, cfg.MaxIdleConns, cfg.ConnMaxLifetime, cfg.ConnMaxIdleTime)

	case config.SQLiteType:
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

		log.Println("Applying SQLite optimizations")
		if cfg.EnableWALMode {
			sqlDB.Exec("PRAGMA journal_mode=WAL")
			log.Println("WAL mode enabled")
		}
		if cfg.BusyTimeout > 0 {
			sqlDB.Exec(fmt.Sprintf("PRAGMA busy_timeout=%d", cfg.BusyTimeout))
			log.Printf("Busy timeout set to %dms", cfg.BusyTimeout)
		}
		sqlDB.Exec("PRAGMA synchronous=NORMAL")
		sqlDB.Exec("PRAGMA cache_size=-64000")
		sqlDB.Exec("PRAGMA temp_store=MEMORY")
	}

	if err := DB.AutoMigrate(
		&User{},
		&ExternalModel{},
		&RequestSource{},
		&APIKey{},
		&UsageRecord{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database tables migrated successfully")

	var count int64
	if err := DB.Model(&User{}).Count(&count).Error; err != nil {
		log.Printf("CRITICAL: Failed to count users: %v", err)
		log.Printf("This may indicate database connection or permission issues")
		log.Printf("Attempting to continue, but default user creation may fail...")
		count = 0
	}

	log.Printf("Current user count in database: %d", count)

	if count == 0 {
		log.Println("No users found, creating default admin user...")

		defaultPassword := config.GetEnv("DEFAULT_ADMIN_PASSWORD", "admin123")
		if defaultPassword == "admin123" {
			log.Println("⚠️  Using default password 'admin123'. Set DEFAULT_ADMIN_PASSWORD environment variable for production!")
		}

		hash, err := utils.HashPassword(defaultPassword)
		if err != nil {
			log.Fatalf("Failed to hash default password: %v", err)
		}

		log.Printf("Password hashed successfully (length: %d bytes)", len(hash))

		defaultUser := User{
			Username: "admin",
			Password: hash,
		}

		if err := DB.Create(&defaultUser).Error; err != nil {
			log.Fatalf("Failed to create default admin user: %v", err)
		}

		log.Printf("✓ Default admin user created successfully (ID: %d)", defaultUser.ID)
		log.Printf("  Username: admin")
		if defaultPassword == "admin123" {
			log.Printf("  Password: admin123 (WARNING: Change this immediately in production!)")
		} else {
			log.Printf("  Password: [configured via DEFAULT_ADMIN_PASSWORD]")
		}

		log.Printf("User creation audit:")
		log.Printf("  - Timestamp: %s", time.Now().UTC().Format(time.RFC3339))
		log.Printf("  - Action: Auto-create default admin")
		log.Printf("  - Database type: %s", cfg.DatabaseType)

		var verifyUser User
		if err := DB.Where("username = ?", "admin").First(&verifyUser).Error; err != nil {
			log.Printf("Warning: Failed to verify created user: %v", err)
		} else {
			log.Printf("✓ User verification successful - Username: %s, ID: %d", verifyUser.Username, verifyUser.ID)
		}
	} else {
		log.Printf("Found %d existing user(s), skipping default user creation", count)

		var firstUser User
		if err := DB.First(&firstUser).Error; err == nil {
			log.Printf("First user in database: %s (ID: %d)", firstUser.Username, firstUser.ID)
		}
	}

	log.Println("Database initialized successfully")
}
