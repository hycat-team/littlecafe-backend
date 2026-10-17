//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"

	"github.com/hycat-team/littlecafe-backend/services/identity/internal/bootstrap"
	identityconfig "github.com/hycat-team/littlecafe-backend/services/identity/internal/config"
	sharedroot "github.com/hycat-team/littlecafe-backend/shared"
	"github.com/hycat-team/littlecafe-backend/shared/infrastructure/db"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/logger"
)

// InitializeApp resolves the technical bootstrap graph. Business providers are
// added only with an approved contract and workflow specification.
func InitializeApp(cfg *identityconfig.Config, log logger.Interface) (*bootstrap.App, error) {
	wire.Build(
		newSharedConfig,
		db.NewPostgresConnection,
		sharedroot.ProviderSet,
		newModuleRouters,
		bootstrap.NewRouter,
		newGRPCRegistrars,
		bootstrap.NewGRPCServer,
		newModuleWorkers,
		bootstrap.NewWorkers,
		bootstrap.NewApp,
	)
	return nil, nil
}
