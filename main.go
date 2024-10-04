package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mypj369/go-app/db"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Cấu trúc User để lưu vào database
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"-"` // Không trả về mật khẩu dưới dạng JSON
}

var DB *gorm.DB

func main() {
	// Load cấu hình database từ file config
	config, err := db.LoadDBConfig("config/database.json")
	if err != nil {
		log.Fatal("Error loading database config:", err)
	}

	// Kết nối tới PostgreSQL dựa trên cấu hình đã load
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		config.Postgres.Host,
		config.Postgres.User,
		config.Postgres.Password,
		config.Postgres.DBName,
		config.Postgres.Port,
		config.Postgres.SSLMode,
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Tự động migrate User model vào database
	DB.AutoMigrate(&User{})

	// Tạo router cho API
	r := gin.Default()

	// Thêm middleware CORS vào router
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

		// Kiểm tra xem email đã tồn tại chưa
		var existingUser User
		if err := DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Email already exists",
				"status":  "failed",
				"details": "The email address provided has already registered!"})
			return
		}

		// Mã hóa mật khẩu bằng bcrypt
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

		// Trả về phản hồi JSON thành công với thông tin chi tiết
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "User created successfully",
			"data": gin.H{
				"userId": user.ID,
				"email":  user.Email,
			},
		})
	})

	// Khởi động server backend
	r.Run(":8085")
}
