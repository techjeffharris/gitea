// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repository

import (
	"context"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/notification"
	repo_module "code.gitea.io/gitea/modules/repository"
)

// GenerateRepository generates a repository from a template
func GenerateRepository(doer, owner *models.User, templateRepo *models.Repository, opts models.GenerateRepoOptions) (_ *models.Repository, err error) {
	if !doer.IsAdmin && !owner.CanCreateRepo() {
		return nil, models.ErrReachLimitOfRepo{
			Limit: owner.MaxRepoCreation,
		}
	}

	var generateRepo *models.Repository
	if err = db.WithTx(func(ctx context.Context) error {
		generateRepo, err = repo_module.GenerateRepository(ctx, doer, owner, templateRepo, opts)
		if err != nil {
			return err
		}

		// Git Content
		if opts.GitContent && !templateRepo.IsEmpty {
			if err = repo_module.GenerateGitContent(ctx, templateRepo, generateRepo); err != nil {
				return err
			}
		}

		// Topics
		if opts.Topics {
			if err = models.GenerateTopics(ctx, templateRepo, generateRepo); err != nil {
				return err
			}
		}

		// Git Hooks
		if opts.GitHooks {
			if err = models.GenerateGitHooks(ctx, templateRepo, generateRepo); err != nil {
				return err
			}
		}

		// Webhooks
		if opts.Webhooks {
			if err = models.GenerateWebhooks(ctx, templateRepo, generateRepo); err != nil {
				return err
			}
		}

		// Avatar
		if opts.Avatar && len(templateRepo.Avatar) > 0 {
			if err = models.GenerateAvatar(ctx, templateRepo, generateRepo); err != nil {
				return err
			}
		}

		// Issue Labels
		if opts.IssueLabels {
			if err = models.GenerateIssueLabels(ctx, templateRepo, generateRepo); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		if generateRepo != nil && generateRepo.ID > 0 {
			if errDelete := models.DeleteRepository(doer, owner.ID, generateRepo.ID); errDelete != nil {
				log.Error("Rollback deleteRepository: %v", errDelete)
			}
		}
		return nil, err
	}

	notification.NotifyCreateRepository(doer, owner, generateRepo)

	return generateRepo, nil
}
