package example

import (
	"github.com/glvd/go-admin/context"
	"github.com/glvd/go-admin/modules/auth"
	"github.com/glvd/go-admin/modules/db"
	"github.com/glvd/go-admin/modules/service"
)

func InitRouter(prefix string, srv service.List) *context.App {

	app := context.NewApp()
	route := app.Group(prefix)
	route.GET("/example", auth.Middleware(db.GetConnection(srv)), TestHandler)

	return app
}
