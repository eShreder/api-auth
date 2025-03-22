package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/user/user-server/pkg/auth"
	"github.com/user/user-server/pkg/database"
	"github.com/user/user-server/pkg/models"
)

type AuthHandler struct {
	DB         *database.DB
	JWTManager *auth.JWTManager
}

type RegisterRequest struct {
	Username     string `json:"username" binding:"required,min=3,max=32"`
	Password  string `json:"password" binding:"required,min=8"`
	InviteToken string `json:"invite_token" binding:"required"`
}

type LoginRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type PublicKeyResponse struct {
	Key string `json:"key"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	inviteToken, err := h.DB.GetInviteToken(req.InviteToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invite token"})
			return
		}
		log.Printf("Error getting invite token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if inviteToken.Used {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invite token already used"})
		return
	}

	_, err = h.DB.GetUserByUsername(req.Username)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with this username already exists"})
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error checking user existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	user := &models.User{
		Username:    req.Username,
		Password: hashedPassword,
	}

	if err := h.DB.CreateUser(user); err != nil {
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if err := h.DB.MarkInviteTokenAsUsed(inviteToken.ID); err != nil {
		log.Printf("Error marking token as used: %v", err)
	}

	accessToken, refreshToken, err := h.JWTManager.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		log.Printf("Error generating tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Сохраняем refresh токен в базе данных
	refreshTokenExpiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 дней
	_, err = h.DB.CreateRefreshToken(user.ID, refreshToken, refreshTokenExpiresAt)
	if err != nil {
		log.Printf("Error saving refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(h.JWTManager.GetTokenTTL().Seconds()),
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	user, err := h.DB.GetUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		log.Printf("Error getting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if err := auth.CheckPassword(req.Password, user.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	accessToken, refreshToken, err := h.JWTManager.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		log.Printf("Error generating tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Сохраняем refresh токен в базе данных
	refreshTokenExpiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 дней
	_, err = h.DB.CreateRefreshToken(user.ID, refreshToken, refreshTokenExpiresAt)
	if err != nil {
		log.Printf("Error saving refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(h.JWTManager.GetTokenTTL().Seconds()),
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	refreshToken, err := h.DB.GetRefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}
		log.Printf("Error getting refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
		return
	}

	user, err := h.DB.GetUserByID(refreshToken.UserID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	accessToken, newRefreshToken, err := h.JWTManager.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		log.Printf("Error generating tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Удаляем старый refresh токен
	if err := h.DB.DeleteRefreshToken(req.RefreshToken); err != nil {
		log.Printf("Error deleting old refresh token: %v", err)
	}

	// Сохраняем новый refresh токен
	refreshTokenExpiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 дней
	_, err = h.DB.CreateRefreshToken(user.ID, newRefreshToken, refreshTokenExpiresAt)
	if err != nil {
		log.Printf("Error saving new refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(h.JWTManager.GetTokenTTL().Seconds()),
	})
}

func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		claims, err := h.JWTManager.VerifyToken(parts[1])
		if err != nil {
			if errors.Is(err, auth.ErrExpiredToken) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	c.JSON(http.StatusOK, gin.H{
		"id":       userID,
		"username": username,
	})
}

func (h *AuthHandler) GetPublicKey(c *gin.Context) {
	publicKeyPEM, err := h.JWTManager.GetPublicKeyPEM()
	if err != nil {
		log.Printf("Error getting public key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, PublicKeyResponse{Key: publicKeyPEM})
} 