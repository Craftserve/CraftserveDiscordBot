package services

import (
	"context"
	"csrvbot/internal/repos"
	"csrvbot/pkg/logger"
)

type SavedroleService struct {
	UserRepo repos.UserRepo
}

func NewSavedRoleService(userRepo *repos.UserRepo) *SavedroleService {
	return &SavedroleService{
		UserRepo: *userRepo,
	}
}

func (h SavedroleService) UpdateMemberSavedRoles(ctx context.Context, memberRoles []string, memberId, guildId string) {
	log := logger.GetLoggerFromContext(ctx)
	savedRoles, err := h.UserRepo.GetRolesForMember(ctx, guildId, memberId)
	if err != nil {
		log.WithError(err).Error("UpdateMemberSavedRoles Error while getting saved roles")
		return
	}
	var savedRolesIds []string
	for _, role := range savedRoles {
		savedRolesIds = append(savedRolesIds, role.RoleId)
	}

	for _, memberRole := range memberRoles {
		found := false
		for i, savedRole := range savedRolesIds {
			if savedRole == memberRole {
				found = true
				savedRolesIds[i] = ""
				break
			}
		}
		if !found {
			log.Debugf("Adding role %s of member %s to saved roles in database", memberRole, memberId)
			err = h.UserRepo.AddRoleForMember(ctx, guildId, memberId, memberRole)
			if err != nil {
				log.WithError(err).Error("UpdateMemberSavedRoles Error while saving new role info", err)
				continue
			}
		}
	}

	for _, savedRole := range savedRolesIds {
		if savedRole != "" {
			log.Debugf("Removing role %s of member %s from saved roles in database", savedRole, memberId)
			err = h.UserRepo.RemoveRoleForMember(ctx, guildId, memberId, savedRole)
			if err != nil {
				log.WithError(err).Error("UpdateMemberSavedRoles Error while deleting info about member role", err)
				continue
			}
		}
	}
}
