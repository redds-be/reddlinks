// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0

package database

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	ID        uuid.UUID
	CreatedAt time.Time
	ExpireAt  time.Time
	Url       string
	Short     string
}