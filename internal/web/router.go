package web

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func SetupRouter(handler *Handler) *gin.Engine {
	r := gin.Default()

	rate, _ := limiter.NewRateFromFormatted("100-M")
	rateMiddleware := mgin.NewMiddleware(limiter.New(memory.NewStore(), rate))

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000", "https://tmd-nanana.com"}
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
		api.GET("/files/:objectName", handler.GetFile)
		api.GET("/chats", handler.GetChats)
	}

	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
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
