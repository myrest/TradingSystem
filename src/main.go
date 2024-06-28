package main

//使用myrest005的帳號
import (
	"TradingSystem/src/routes"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	// 设置会话存储
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	//設定HTML樣板目錄
	templatesDir := filepath.Join(wd, "templates")
	r.LoadHTMLGlob(filepath.Join(templatesDir, "**/*")) // Add this line to load HTML templates

	//設定靜態路由
	staticDir := filepath.Join(wd, "static")
	r.Static("/static", staticDir)
	r.StaticFile("/favicon.ico", filepath.Join(staticDir, "favicon.ico"))

	routes.RegisterRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server started at :%s", port)
	log.Fatal(r.Run(":" + port))
}
