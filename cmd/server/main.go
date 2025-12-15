// cmd/server/main.go

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/rs/cors" // 1. 导入新库
	"github.com/yeeeck/sync-jukebox/internal/api"
	"github.com/yeeeck/sync-jukebox/internal/db"
	"github.com/yeeeck/sync-jukebox/internal/state"
	"github.com/yeeeck/sync-jukebox/internal/websocket"
)

const (
	dbPath      = "./jukebox.db"
	mediaDir    = "./media"
	frontendDir = "./frontend/dist"
	serverAddr  = ":8080"
)

func main() {
	// ... (数据库、Hub、状态管理器的初始化代码保持不变) ...
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		log.Fatalf("Failed to create media directory: %v", err)
	}

	database, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("DB initialization failed: %v", err)
	}
	defer database.Close()

	hub := websocket.NewHub()
	go hub.Run()

	stateManager, err := state.NewManager(database, hub)
	if err != nil {
		log.Fatalf("State manager initialization failed: %v", err)
	}

	// 初始化 API 和路由
	apiHandler := api.New(database, stateManager, hub, mediaDir)
	mux := http.NewServeMux()
	apiHandler.RegisterRoutes(mux)
	mux.Handle("/", http.FileServer(http.Dir(frontendDir)))

	// 2. 配置 CORS 中间件
	// 这里我们允许了常见的前端开发服务器源。
	// 在生产环境中，你应该将其替换为你的前端域名。
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://10.8.0.10:5173",
			"http://localhost:5173", // Vite (Vue) 默认
			"http://localhost:3000", // Create React App 默认
			"http://localhost:4200", // Angular 默认
			// "https://your-production-frontend.com", // 生产环境域名
		},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions}, // 允许的方法
		AllowedHeaders: []string{"Content-Type", "Authorization"},                     // 允许的请求头
		Debug:          true,                                                          // 在控制台打印调试信息，方便排查问题
	})

	// 3. 将 CORS 中间件应用到我们的路由
	handler := c.Handler(mux)

	// 启动服务器
	log.Printf("SyncJukebox v2.0 server starting on %s with CORS enabled", serverAddr)
	log.Printf("Serving frontend from: %s", frontendDir)
	log.Printf("Serving media from: %s", mediaDir)

	// 4. 使用包装后的 handler
	if err := http.ListenAndServe(serverAddr, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
