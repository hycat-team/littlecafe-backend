package di

import "github.com/hycat-team/littlecafe-backend/services/identity/internal/bootstrap"

// newModuleRouters starts empty until a reviewed business contract is implemented.
func newModuleRouters() []bootstrap.ModuleRouter { return nil }

// newModuleWorkers starts empty until an AsyncAPI event contract is approved.
func newModuleWorkers() []bootstrap.ModuleWorker { return nil }

// newGRPCRegistrars keeps the gRPC host healthy without publishing a fake RPC.
func newGRPCRegistrars() []bootstrap.GRPCServerRegistrar { return nil }
