package repository

import (
	"context"

	"github.com/localrepo/api-server/internal/models"
	"gorm.io/gorm"
)

type TeamRepository interface {
	CreateOrganization(ctx context.Context, org *models.Organization) error
	GetOrganization(ctx context.Context, id uint) (*models.Organization, error)
	CreateTeam(ctx context.Context, team *models.Team) error
	AddTeamMember(ctx context.Context, member *models.TeamMember) error
	SetTeamRepoPermission(ctx context.Context, perm *models.TeamRepositoryPermission) error
	GetUserTeamRole(ctx context.Context, userID, teamID uint) (models.TeamRole, error)
	GetUserRepoRole(ctx context.Context, userID, repoID uint) (string, error)
}

type teamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) CreateOrganization(ctx context.Context, org *models.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

func (r *teamRepository) GetOrganization(ctx context.Context, id uint) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).Preload("Owner").First(&org, id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *teamRepository) CreateTeam(ctx context.Context, team *models.Team) error {
	return r.db.WithContext(ctx).Create(team).Error
}

func (r *teamRepository) AddTeamMember(ctx context.Context, member *models.TeamMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *teamRepository) SetTeamRepoPermission(ctx context.Context, perm *models.TeamRepositoryPermission) error {
	return r.db.WithContext(ctx).Save(perm).Error
}

func (r *teamRepository) GetUserTeamRole(ctx context.Context, userID, teamID uint) (models.TeamRole, error) {
	var tm models.TeamMember
	err := r.db.WithContext(ctx).Where("user_id = ? AND team_id = ?", userID, teamID).First(&tm).Error
	if err != nil {
		return "", err
	}
	return tm.Role, nil
}

func (r *teamRepository) GetUserRepoRole(ctx context.Context, userID, repoID uint) (string, error) {
	var perm models.TeamRepositoryPermission
	err := r.db.WithContext(ctx).
		Joins("JOIN team_members ON team_members.team_id = team_repository_permissions.team_id").
		Where("team_members.user_id = ? AND team_repository_permissions.repository_id = ?", userID, repoID).
		First(&perm).Error

	if err != nil {
		return "", err
	}
	return perm.Role, nil
}
