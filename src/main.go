package main

//使用myrest005的帳號
import (
	"TradingSystem/src/common"
	"TradingSystem/src/middleware"
	"TradingSystem/src/routes"
	"TradingSystem/src/services"
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// formatFloat64 用于格式化浮点数为不使用科学记号的字符串
func formatFloat64(round int, f float64) string {
	value := common.Decimal(f, round)
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func main() {
	// 創建 Gin 路由器
	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"subtract": func(a, b int) int {
			return a - b
		},
		"add": func(a, b int) int {
			return a + b
		},
		"timesf": func(a, b float64) float64 {
			return common.Decimal(a*b, 6)
		},
		"iterate": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		"formatFloat64": formatFloat64,
	})

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

	r.Use(middleware.ErrorHandlingMiddleware())
	routes.RegisterRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 創建 HTTP 服務器
	var addr string
	if common.GetEnvironmentSetting().Env == common.Dev {
		addr = "127.0.0.1:" + port
	} else {
		addr = ":" + port
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 在 goroutine 中啟動服務器
	go func() {
		log.Printf("Server started at %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中斷信號以優雅地關閉服務器
	quit := make(chan os.Signal, 1)
	// 捕獲 SIGINT (Ctrl+C) 和 SIGTERM 信號
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 創建一個帶有超時的 context，給服務器一些時間來完成正在處理的請求
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 關閉服務器
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// 清理資源
	log.Println("Flushing logs and cleaning up resources...")
	services.FlushLogging()
	if err := services.CloseLogging(); err != nil {
		log.Printf("Error closing logging client: %v", err)
	}

	log.Println("Server exited")
}
