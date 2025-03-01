package web

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"log"
)

func SetupRouter(handler *Handler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	rate, err := limiter.NewRateFromFormatted("100-M")
	if err != nil {
		log.Fatalf("Failed to create rate: %v", err)
	}

	store := memory.NewStore()
	rateMiddleware := mgin.NewMiddleware(limiter.New(store, rate))

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3003", "https://tmd-nanana.com"}
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	r.Use(
		gzip.Gzip(gzip.DefaultCompression),
		cors.New(corsConfig),
		rateMiddleware,
		SecurityMiddleware(),
	)

	api := r.Group("/api/v1")
	{
		api.GET("/chats/:chatID/messages", handler.GetChatMessages)
		api.GET("/files/*objectName", handler.GetFile)
		api.GET("/chats", handler.GetChats)
	}

	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/public/index.html")
	})

	return r
}

func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	}
}
