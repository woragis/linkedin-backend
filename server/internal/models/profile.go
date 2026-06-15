package models

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	Slug      string    `gorm:"uniqueIndex;not null" json:"slug"`
	FullName  string    `gorm:"column:full_name;not null" json:"full_name"`
	Headline  string    `gorm:"not null;default:''" json:"headline"`
	Bio       string    `gorm:"not null;default:''" json:"bio"`
	Location  string    `gorm:"not null;default:''" json:"location"`
	BirthYear *int      `gorm:"column:birth_year" json:"birth_year,omitempty"`
	AvatarURL *string   `gorm:"column:avatar_url" json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func (Profile) TableName() string { return "profiles" }

type Institution struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Slug      string    `gorm:"uniqueIndex;not null" json:"slug"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

func (Institution) TableName() string { return "institutions" }

type Company struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Slug      string    `gorm:"uniqueIndex;not null" json:"slug"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

func (Company) TableName() string { return "companies" }

type Skill struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Slug      string    `gorm:"uniqueIndex;not null" json:"slug"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

func (Skill) TableName() string { return "skills" }

type Education struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	InstitutionID uuid.UUID `gorm:"type:uuid;not null" json:"institution_id"`
	FieldOfStudy  string    `gorm:"column:field_of_study;not null;default:''" json:"field_of_study"`
	Degree        string    `gorm:"not null;default:''" json:"degree"`
	StartYear     *int      `gorm:"column:start_year" json:"start_year,omitempty"`
	EndYear       *int      `gorm:"column:end_year" json:"end_year,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`

	Institution Institution `gorm:"foreignKey:InstitutionID" json:"institution,omitempty"`
}

func (Education) TableName() string { return "educations" }

type Experience struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	CompanyID   uuid.UUID `gorm:"type:uuid;not null" json:"company_id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `gorm:"not null;default:''" json:"description"`
	StartYear   *int      `gorm:"column:start_year" json:"start_year,omitempty"`
	EndYear     *int      `gorm:"column:end_year" json:"end_year,omitempty"`
	IsCurrent   bool      `gorm:"column:is_current;not null;default:false" json:"is_current"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`

	Company Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}

func (Experience) TableName() string { return "experiences" }

type UserSkill struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	SkillID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"skill_id"`
	CreatedAt time.Time `json:"created_at,omitempty"`

	Skill Skill `gorm:"foreignKey:SkillID" json:"skill,omitempty"`
}

func (UserSkill) TableName() string { return "user_skills" }
