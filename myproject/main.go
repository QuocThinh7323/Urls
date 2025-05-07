package main

import (
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

// URL là model đại diện cho bảng trong cơ sở dữ liệu
type URL struct {
	ID          uint   `gorm:"primaryKey"`
	ShortLink   string `gorm:"type:varchar(255);uniqueIndex"`
	OriginalURL string
	VisitCount  int
}

func initDB() {
	var err error

	// Chuỗi kết nối MySQL (Data Source Name - DSN)
	dsn := "root:my-secret-pw@tcp(127.0.0.1:3306)/go_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	// Tự động tạo bảng nếu chưa tồn tại
	err = db.AutoMigrate(&URL{})
	if err != nil {
		panic("Failed to migrate database!")
	}
}

func generateShortLink() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	short := make([]rune, 6)
	for i := range short {
		short[i] = letters[rand.Intn(len(letters))]
	}
	return string(short)
}

func shortenURL(c *gin.Context) {
	var request struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	shortLink := generateShortLink()

	// Lưu URL vào cơ sở dữ liệu
	url := URL{
		ShortLink:   shortLink,
		OriginalURL: request.URL,
		VisitCount:  0,
	}
	if err := db.Create(&url).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"short_url": "http://localhost:8080/" + shortLink})
}

func redirectURL(c *gin.Context) {
	shortLink := c.Param("shortLink")

	// Tìm URL gốc trong cơ sở dữ liệu và tăng bộ đếm
	var url URL
	if err := db.First(&url, "short_link = ?", shortLink).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	// Tăng số lượt truy cập
	db.Model(&url).Update("visit_count", url.VisitCount+1)

	c.Redirect(http.StatusFound, url.OriginalURL)
}

func homePage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to the URL Shortener API!",
		"usage":   "Use POST /shorten to shorten a URL and GET /:shortLink to redirect.",
	})
}

func renderHomePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func main() {
	r := gin.Default()
	
	// Cấu hình template và static files
	r.LoadHTMLGlob("templates/*")
	
	// Thử kết nối database sau khi đã thiết lập các route cơ bản
	initDB()
	
	// Thêm route cho trang chủ
	r.GET("/", renderHomePage)
	r.GET("/index", renderHomePage)  // Thêm route thay thế
	r.GET("/indexus", renderHomePage) // Giữ nguyên route cũ

	r.POST("/shorten", shortenURL)
	r.GET("/:shortLink", redirectURL)
	
	r.Run(":8080")
}