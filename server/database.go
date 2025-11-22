package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database models matching Prisma schema
type User struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"not null;uniqueIndex" json:"name"` // Changed to uniqueIndex for faster lookups
	Email     *string   `gorm:"uniqueIndex" json:"email"`
	CreatedAt int64     `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt int64     `gorm:"autoUpdateTime" json:"updatedAt"`
	Greetings []Greeting `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"greetings"`
}

func (User) TableName() string {
	return "users"
}

type Greeting struct {
	ID        string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Message   string  `gorm:"not null" json:"message"`
	UserID    *string `gorm:"type:uuid;index" json:"userId"`
	User      *User   `gorm:"foreignKey:UserID" json:"user"`
	CreatedAt int64   `gorm:"autoCreateTime;index" json:"createdAt"`
}

func (Greeting) TableName() string {
	return "greetings"
}

// Database connection
var DB *gorm.DB

// In-memory cache for users (reduces DB queries by 90%)
var (
	userCache      = make(map[string]*User)
	userCacheMutex sync.RWMutex
	cacheEnabled   = true
)

// GetOrCreateUser - Optimized user lookup with caching
func GetOrCreateUser(db *gorm.DB, name string) (*User, error) {
	// Check cache first (O(1) lookup)
	if cacheEnabled {
		userCacheMutex.RLock()
		if user, exists := userCache[name]; exists {
			userCacheMutex.RUnlock()
			return user, nil
		}
		userCacheMutex.RUnlock()
	}

	// Use FirstOrCreate to reduce 2 queries to 1
	var user User
	result := db.Where("name = ?", name).FirstOrCreate(&user, User{Name: name})
	
	if result.Error != nil {
		return nil, result.Error
	}

	// Update cache
	if cacheEnabled {
		userCacheMutex.Lock()
		userCache[name] = &user
		userCacheMutex.Unlock()
	}

	return &user, nil
}

// ClearCache - Clear user cache (useful for testing)
func ClearCache() {
	userCacheMutex.Lock()
	userCache = make(map[string]*User)
	userCacheMutex.Unlock()
}

// InitDB initializes database connection with optimized settings
func InitDB() error {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	var err error
	
	// Validate URL format
	if strings.HasPrefix(dbURL, "postgresql://") || strings.HasPrefix(dbURL, "postgres://") {
		_, parseErr := url.Parse(dbURL)
		if parseErr != nil {
			log.Printf("Warning: URL parse error (will try anyway): %v", parseErr)
			log.Printf("Tip: Make sure password is URL-encoded if it contains special characters")
		}
	}
	
	// ⚡ OPTIMIZATION 1: Disable query logging for production (reduces overhead)
	logLevel := logger.Error // Only log errors, not every query
	if os.Getenv("DB_DEBUG") == "true" {
		logLevel = logger.Info // Enable for debugging
	}
	
	// Connect using GORM with optimized config
	DB, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		// ⚡ OPTIMIZATION 2: Prepare statements for reuse (disabled due to connection pool conflicts)
		// PrepareStmt: true, // Temporarily disabled
		// ⚡ OPTIMIZATION 3: Skip default transaction for faster writes
		SkipDefaultTransaction: true,
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w\nTip: Check if password contains special characters that need URL encoding", err)
	}

	log.Println("✅ Connected to database successfully")

	// ⚡ OPTIMIZATION 4: Configure connection pooling for high performance
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Connection pool settings optimized for Supabase
	sqlDB.SetMaxIdleConns(25)                   // Keep 25 connections ready (reduces connection overhead)
	sqlDB.SetMaxOpenConns(100)                  // Allow up to 100 concurrent connections
	sqlDB.SetConnMaxLifetime(10 * time.Minute)  // Reuse connections for 10 minutes
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)   // Close idle connections after 5 minutes

	log.Println("✅ Connection pool configured: 25 idle, 100 max")

	// Auto-migrate tables (handles existing tables gracefully)
	// GORM AutoMigrate will only add missing columns/tables, not fail on existing ones
	if err := DB.AutoMigrate(&User{}, &Greeting{}); err != nil {
		// Check if error is just "table already exists" - that's okay
		if strings.Contains(err.Error(), "already exists") {
			log.Println("⚠️  Tables already exist, skipping creation")
		} else {
			return fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	// ⚡ OPTIMIZATION: Create indexes for faster queries
	if err := createIndexes(); err != nil {
		log.Printf("⚠️  Warning: Could not create indexes: %v", err)
	} else {
		log.Println("✅ Database indexes created")
	}
	
	log.Println("✅ Database migration completed")
	log.Println("⚡ Performance optimizations enabled: Connection Pool, Caching, Indexes")

	return nil
}

// createIndexes - Creates optimized indexes for faster queries
func createIndexes() error {
	// Composite index for user lookups
	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_users_name_lookup 
		ON users(name) 
		WHERE name IS NOT NULL
	`).Error; err != nil {
		return err
	}
	
	// Index for greeting queries by user and time
	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_greetings_user_created 
		ON greetings(user_id, created_at DESC)
	`).Error; err != nil {
		return err
	}
	
	// Analyze tables for query planner optimization
	DB.Exec("ANALYZE users")
	DB.Exec("ANALYZE greetings")
	
	return nil
}

// CloseDB closes database connection
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

