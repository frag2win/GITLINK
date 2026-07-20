package models

import "gorm.io/gorm"

type Organization struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string `json:"description"`
	OwnerID     uint   `gorm:"not null;index"`
	Owner       User   `gorm:"foreignKey:OwnerID"`
}

type TeamRole string

const (
	TeamRoleOwner      TeamRole = "Owner"
	TeamRoleAdmin      TeamRole = "Admin"
	TeamRoleMaintainer TeamRole = "Maintainer"
	TeamRoleDeveloper  TeamRole = "Developer"
	TeamRoleReporter   TeamRole = "Reporter"
	TeamRoleGuest      TeamRole = "Guest"
)

type Team struct {
	gorm.Model
	OrganizationID uint         `gorm:"not null;index"`
	Organization   Organization `gorm:"foreignKey:OrganizationID"`
	Name           string       `gorm:"not null"`
	Description    string       `json:"description"`
	Members        []TeamMember `gorm:"foreignKey:TeamID"`
}

type TeamMember struct {
	gorm.Model
	TeamID uint     `gorm:"not null;index"`
	UserID uint     `gorm:"not null;index"`
	User   User     `gorm:"foreignKey:UserID"`
	Role   TeamRole `gorm:"type:varchar(32);not null;default:'Developer'"`
}

type TeamRepositoryPermission struct {
	gorm.Model
	TeamID       uint   `gorm:"not null;index"`
	RepositoryID uint   `gorm:"not null;index"`
	Role         string `gorm:"type:varchar(32);not null"` // "admin", "write", "read"
}
