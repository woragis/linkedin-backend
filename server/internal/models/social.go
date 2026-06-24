package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Connection struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RequesterID uuid.UUID `gorm:"type:uuid;not null;index" json:"requester_id"`
	AddresseeID uuid.UUID `gorm:"type:uuid;not null;index" json:"addressee_id"`
	Status      string    `gorm:"not null;default:'pending'" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Connection) TableName() string { return "connections" }

type Post struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	AuthorID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"author_id"`
	Body      string     `gorm:"not null" json:"body"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	Author    Profile    `gorm:"foreignKey:AuthorID;references:UserID" json:"author,omitempty"`
	Reactions []Reaction `gorm:"foreignKey:PostID" json:"-"`
	Comments  []Comment  `gorm:"foreignKey:PostID" json:"-"`
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
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	PostID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"post_id"`
	AuthorID        uuid.UUID  `gorm:"type:uuid;not null" json:"author_id"`
	ParentCommentID *uuid.UUID `gorm:"type:uuid;index" json:"parent_comment_id,omitempty"`
	Body            string     `gorm:"not null" json:"body"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`

	Author Profile `gorm:"foreignKey:AuthorID;references:UserID" json:"author,omitempty"`
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
