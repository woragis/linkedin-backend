package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Connection struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	RequesterID uuid.UUID `gorm:"type:uuid;not null;index"`
	AddresseeID uuid.UUID `gorm:"type:uuid;not null;index"`
	Status      string    `gorm:"not null;default:'pending'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Connection) TableName() string { return "connections" }

type Post struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	AuthorID  uuid.UUID  `gorm:"type:uuid;not null;index"`
	Body      string     `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	Author   Profile    `gorm:"foreignKey:AuthorID;references:UserID"`
	Reactions []Reaction `gorm:"foreignKey:PostID"`
	Comments  []Comment  `gorm:"foreignKey:PostID"`
}

func (Post) TableName() string { return "posts" }

type Reaction struct {
	PostID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Kind      string    `gorm:"not null;default:'like'"`
	CreatedAt time.Time
}

func (Reaction) TableName() string { return "reactions" }

type Comment struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	PostID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	AuthorID  uuid.UUID  `gorm:"type:uuid;not null"`
	Body      string     `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	Author Profile `gorm:"foreignKey:AuthorID;references:UserID"`
}

func (Comment) TableName() string { return "comments" }

type Event struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    *uuid.UUID
	EventType string    `gorm:"column:event_type;not null"`
	Payload   datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt time.Time
}

func (Event) TableName() string { return "events" }

type OutboxJob struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	JobType     string     `gorm:"column:job_type;not null"`
	Payload     datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt   time.Time
	ProcessedAt *time.Time
	LastError   *string
}

func (OutboxJob) TableName() string { return "outbox_jobs" }
