package models

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	UserID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	Slug       string    `gorm:"uniqueIndex;not null"`
	FullName   string    `gorm:"column:full_name;not null"`
	Headline   string    `gorm:"not null;default:''"`
	Bio        string    `gorm:"not null;default:''"`
	Location   string    `gorm:"not null;default:''"`
	BirthYear  *int      `gorm:"column:birth_year"`
	AvatarURL  *string   `gorm:"column:avatar_url"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (Profile) TableName() string { return "profiles" }

type Institution struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"not null"`
	Slug      string    `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time
}

func (Institution) TableName() string { return "institutions" }

type Company struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"not null"`
	Slug      string    `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time
}

func (Company) TableName() string { return "companies" }

type Skill struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"not null"`
	Slug      string    `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time
}

func (Skill) TableName() string { return "skills" }

type Education struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index"`
	InstitutionID uuid.UUID `gorm:"type:uuid;not null"`
	FieldOfStudy  string    `gorm:"column:field_of_study;not null;default:''"`
	Degree        string    `gorm:"not null;default:''"`
	StartYear     *int      `gorm:"column:start_year"`
	EndYear       *int      `gorm:"column:end_year"`
	CreatedAt     time.Time
	UpdatedAt     time.Time

	Institution Institution `gorm:"foreignKey:InstitutionID"`
}

func (Education) TableName() string { return "educations" }

type Experience struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
	CompanyID   uuid.UUID `gorm:"type:uuid;not null"`
	Title       string    `gorm:"not null"`
	Description string    `gorm:"not null;default:''"`
	StartYear   *int      `gorm:"column:start_year"`
	EndYear     *int      `gorm:"column:end_year"`
	IsCurrent   bool      `gorm:"column:is_current;not null;default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Company Company `gorm:"foreignKey:CompanyID"`
}

func (Experience) TableName() string { return "experiences" }

type UserSkill struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	SkillID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time

	Skill Skill `gorm:"foreignKey:SkillID"`
}

func (UserSkill) TableName() string { return "user_skills" }
