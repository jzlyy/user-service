package services

import (
	"database/sql"
	"errors"
	"log"
	"user-service/database"
	"user-service/models"
)

func GetUserByID(userID int) (*models.User, error) {
	var user models.User
	err := database.DB.QueryRow(`
		SELECT id, username, email, created_at 
		FROM users 
		WHERE id = ?`, userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		log.Printf("Failed to get user by ID: %v", err)
		return nil, err
	}
	return &user, nil
}
