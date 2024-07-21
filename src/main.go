package main

//使用myrest005的帳號
import (
	"TradingSystem/src/routes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// formatFloat64 用于格式化浮点数为不使用科学记号的字符串
func formatFloat64(f interface{}) string {
	var strValue string
	switch v := f.(type) {
	case float64:
		// 使用 strconv.FormatFloat 避免科学记号
		strValue = strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		// 如果是 float32，同样使用 strconv.FormatFloat
		strValue = strconv.FormatFloat(float64(v), 'f', -1, 32)
	default:
		return fmt.Sprintf("%v", f)
	}

	// 根据传入的精度四舍五入
	//format := fmt.Sprintf("%%.%df", 6)
	//value := fmt.Sprintf(format, strValue)

	// 去掉尾随的零和小数点
	//value = strings.TrimRight(value, "0")
	//value = strings.TrimRight(value, ".")
	return strValue
}

func main() {
	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"subtract": func(a, b int) int {
			return a - b
		},
		"add": func(a, b int) int {
			return a + b
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

	routes.RegisterRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server started at :%s", port)
	log.Fatal(r.Run(":" + port))
}
