package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Cấu trúc cấu hình server
type ServerConfig struct {
	Server struct {
		Port string `json:"port"`
	} `json:"server"`
}

// Cấu trúc cấu hình cơ sở dữ liệu
type PostgresConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

type DBConfig struct {
	DBType   string         `json:"db_type"`
	Postgres PostgresConfig `json:"postgres"`
}

// Định nghĩa struct User
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"-"` // Không trả về mật khẩu trong JSON
}

var DB *gorm.DB

// Hàm đọc file cấu hình server
func loadServerConfig() (ServerConfig, error) {
	var config ServerConfig
	data, err := os.ReadFile("config/environments.json")
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

// Hàm đọc file cấu hình cơ sở dữ liệu
func loadDBConfig() (DBConfig, error) {
	var config DBConfig
	data, err := os.ReadFile("config/database.json")
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

// Hàm kết nối với PostgreSQL
func connectPostgres(config PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func main() {
	serverConfig, err := loadServerConfig()
	if err != nil {
		log.Fatal("Failed to load server config:", err)
	} else {
		fmt.Println("Server running on port: %s\n", serverConfig.Server.Port)
	}
	// Đọc file config database
	dbConfig, err := loadDBConfig()
	if err != nil {
		log.Fatal("Failed to load database config:", err)
	}

	// Kết nối với PostgreSQL
	DB, err = connectPostgres(dbConfig.Postgres)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Tự động migrate bảng User
	DB.AutoMigrate(&User{})

	// Tạo router cho API
	r := gin.Default()

	// Middleware CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Đăng ký user
	r.POST("/register", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Kiểm tra email đã tồn tại chưa
		var existingUser User
		if err := DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Email already exists",
				"status":  "failed",
				"details": "The email address provided has already registered!"})
			return
		}

		// Mã hóa mật khẩu
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = string(hashedPassword)

		// Lưu user vào database
		if err := DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create user"})
			return
		}

		// Trả về kết quả thành công
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "User created successfully",
			"data": gin.H{
				"userId": user.ID,
				"email":  user.Email,
			},
		})
	})
	// Khởi động server backend với cổng từ environments.json
	//fmt.Printf("Server running on port: %s\n", serverConfig.Server.Port)
	// Khởi động server backend với cổng từ environments.json
	r.Run(":" + serverConfig.Server.Port)
}
