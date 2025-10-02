package remote_config

import (
	"configuration-management-service/internal/remote_config/handler"
	"configuration-management-service/internal/remote_config/repository"
	"configuration-management-service/internal/remote_config/service"
	"configuration-management-service/internal/remote_config/validator"
	"database/sql"

	"github.com/labstack/echo/v4"
)

type IModule interface {
	RegisterRoute(g *echo.Group, writeLimit echo.MiddlewareFunc)
}

type module struct {
	srv       service.IService
	repo      repository.IRepo
	h         handler.IHandler
	validator validator.ISchemaValidator
}

func New(
	repo repository.IRepo,
	schemaValidator validator.ISchemaValidator,
) IModule {
	srv := service.NewService(repo, schemaValidator)
	h := handler.NewHandler(srv)

	return &module{
		srv:       srv,
		repo:      repo,
		h:         h,
		validator: schemaValidator,
	}
}

func NewWithDB(db *sql.DB) IModule {
	schemaValidator := validator.NewSchemaValidator()
	repo := repository.NewRepo(db)
	return New(repo, schemaValidator)
}

func InitModule(db *sql.DB) IModule {
	return NewWithDB(db)
}

func (m *module) RegisterRoute(g *echo.Group, writeLimit echo.MiddlewareFunc) {
	if g == nil {
		return
	}

	cfgs := g.Group("/configs")
	cfgs.POST("", m.h.Create, writeLimit)
	cfgs.PUT("/:name", m.h.Update, writeLimit)
	cfgs.GET("/:name", m.h.Get)
	cfgs.GET("/:name/versions", m.h.List)
	cfgs.POST("/:name/rollback", m.h.Rollback)
}
