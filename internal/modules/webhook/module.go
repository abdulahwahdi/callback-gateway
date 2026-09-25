// Package webhook is the candi module that receives every payment-gateway
// webhook callback, durably logs it (source + raw body), fans it out to
// Kafka for every interested downstream service, and serves a dashboard API
// for tracking all of it.
package webhook

import (
	"webhook-middleware/internal/modules/webhook/delivery/resthandler"
	"webhook-middleware/internal/modules/webhook/delivery/workerhandler"
	"webhook-middleware/pkg/shared/usecase"

	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
)

const (
	moduleName types.Module = "Webhook"
)

// Module model
type Module struct {
	restHandler    interfaces.RESTHandler
	grpcHandler    interfaces.GRPCHandler
	graphqlHandler interfaces.GraphQLHandler

	workerHandlers map[types.Worker]interfaces.WorkerHandler
	serverHandlers map[types.Server]interfaces.ServerHandler
}

// NewModule module constructor
func NewModule(deps dependency.Dependency) *Module {
	var mod Module
	mod.restHandler = resthandler.NewRestHandler(usecase.GetSharedUsecase(), deps)

	mod.workerHandlers = map[types.Worker]interfaces.WorkerHandler{
		types.Scheduler: workerhandler.NewCronHandler(usecase.GetSharedUsecase(), deps),
	}

	mod.serverHandlers = map[types.Server]interfaces.ServerHandler{}

	return &mod
}

// RESTHandler method
func (m *Module) RESTHandler() interfaces.RESTHandler {
	return m.restHandler
}

// GRPCHandler method
func (m *Module) GRPCHandler() interfaces.GRPCHandler {
	return m.grpcHandler
}

// GraphQLHandler method
func (m *Module) GraphQLHandler() interfaces.GraphQLHandler {
	return m.graphqlHandler
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType types.Worker) interfaces.WorkerHandler {
	return m.workerHandlers[workerType]
}

// ServerHandler additional server type (another rest framework, p2p, and many more)
func (m *Module) ServerHandler(serverType types.Server) interfaces.ServerHandler {
	return m.serverHandlers[serverType]
}

// Name get module name
func (m *Module) Name() types.Module {
	return moduleName
}
