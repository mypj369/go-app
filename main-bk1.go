package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	//"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// testing
// Cấu trúc User để lưu vào database
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"-"` // Không trả về mật khẩu dưới dạng JSON
}

var DB *gorm.DB

func main() {
	// Kết nối với MariaDB
	//dsn := "root:root@tcp(127.0.0.1:8889)/bokt?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := "user=postgres password=123@123 dbname=bokt host=localhost port=5432 sslmode=disable"
	var err error
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
