package web

import (
	"github.com/google/uuid"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"tmd/internal/db"
	"tmd/pkg/minio"
)

type Handler struct {
	DB        *db.DB
	Minio     *minio.Storage
	PageLimit int
}

func NewHandler(db *db.DB, minio *minio.Storage) *Handler {
	return &Handler{
		DB:        db,
		Minio:     minio,
		PageLimit: 50,
	}
}

type MessageResponse struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	MediaURL  string `json:"media_url"`
	CreatedAt string `json:"created_at"`
	Username  string `json:"username"`
}
type ChatResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

func (h *Handler) GetChats(c *gin.Context) {
	var chats []db.Chat
	result := h.DB.Conn.Order("created_at DESC").Find(&chats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	response := make([]ChatResponse, len(chats))
	for i, chat := range chats {
		response[i] = ChatResponse{
			ID:        chat.ID.String(),
			Title:     chat.Title,
			CreatedAt: chat.CreatedAt.Format(time.RFC3339),
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (h *Handler) GetChatMessages(ctx *gin.Context) {
	chatIDstr := ctx.Param("chatID")
	chatUUID, err := uuid.Parse(chatIDstr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat id"})
		return
	}
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	mediaType := ctx.Query("media_type")
	offset := (page - 1) * h.PageLimit

	var messages []db.Message
	var total int64

	query := h.DB.Conn.Model(&db.Message{}).Where("chat_id = ?", chatUUID)

	if mediaType != "" && mediaType != "all" {
		query = query.Where("message_type = ?", mediaType)
	}

	if err := query.Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if err := query.
		Preload("User").
		Order("created_at DESC").
		Limit(h.PageLimit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	response := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		username := ""
		if msg.User != nil {
			username = msg.User.Username
		}
		response[i] = MessageResponse{
			ID:        msg.ID.String(),
			Content:   msg.Content,
			MediaURL:  msg.MediaURL,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
			Username:  username,
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": response,
		"meta": gin.H{
			"page":       page,
			"per_page":   h.PageLimit,
			"total":      total,
			"totalPages": (int(total) + h.PageLimit - 1) / h.PageLimit,
		},
	})
}

func (h *Handler) GetFile(c *gin.Context) {
	objectName := c.Param("objectName")
	if !isValidObjectName(objectName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file path"})
		return
	}

	presignedURL, err := h.Minio.GeneratePresignedURL(objectName, 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": presignedURL})
}

func isValidObjectName(name string) bool {
	return filepath.Base(name) == name && !filepath.IsAbs(name)
}
