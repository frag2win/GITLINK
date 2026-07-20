package service

import (
	"context"

	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/repository"
)

type TeamService interface {
	CreateOrganization(ctx context.Context, name, description string, ownerID uint) (*models.Organization, error)
	CreateTeam(ctx context.Context, orgID uint, name, description string) (*models.Team, error)
	AddMember(ctx context.Context, teamID, userID uint, role models.TeamRole) error
	SetRepoPermission(ctx context.Context, teamID, repoID uint, role string) error
}

type teamService struct {
	teamRepo repository.TeamRepository
}

func NewTeamService(teamRepo repository.TeamRepository) TeamService {
	return &teamService{teamRepo: teamRepo}
}

func (s *teamService) CreateOrganization(ctx context.Context, name, description string, ownerID uint) (*models.Organization, error) {
	org := &models.Organization{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
	}
	if err := s.teamRepo.CreateOrganization(ctx, org); err != nil {
		return nil, err
	}
	return org, nil
}

func (s *teamService) CreateTeam(ctx context.Context, orgID uint, name, description string) (*models.Team, error) {
	team := &models.Team{
		OrganizationID: orgID,
		Name:           name,
		Description:    description,
	}
	if err := s.teamRepo.CreateTeam(ctx, team); err != nil {
		return nil, err
	}
	return team, nil
}

func (s *teamService) AddMember(ctx context.Context, teamID, userID uint, role models.TeamRole) error {
	member := &models.TeamMember{
		TeamID: teamID,
		UserID: userID,
		Role:   role,
	}
	return s.teamRepo.AddTeamMember(ctx, member)
}

func (s *teamService) SetRepoPermission(ctx context.Context, teamID, repoID uint, role string) error {
	perm := &models.TeamRepositoryPermission{
		TeamID:       teamID,
		RepositoryID: repoID,
		Role:         role,
	}
	return s.teamRepo.SetTeamRepoPermission(ctx, perm)
}
