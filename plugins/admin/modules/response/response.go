package response

import (
	"github.com/glvd/go-admin/context"
	"github.com/glvd/go-admin/modules/auth"
	"github.com/glvd/go-admin/modules/config"
	"github.com/glvd/go-admin/modules/db"
	"github.com/glvd/go-admin/modules/language"
	"github.com/glvd/go-admin/modules/menu"
	"github.com/glvd/go-admin/plugins/admin/modules/constant"
	"github.com/glvd/go-admin/template"
	"github.com/glvd/go-admin/template/types"
	template2 "html/template"
	"net/http"
)

func Ok(ctx *context.Context) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"code": 200,
		"msg":  "ok",
	})
}

func OkWithData(ctx *context.Context, data map[string]interface{}) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"code": 200,
		"msg":  "ok",
		"data": data,
	})
}

func BadRequest(ctx *context.Context, msg string) {
	ctx.JSON(http.StatusBadRequest, map[string]interface{}{
		"code": 400,
		"msg":  language.Get(msg),
	})
}

func Alert(ctx *context.Context, config config.Config, desc, title, msg string, conn db.Connection) {
	user := auth.Auth(ctx)

	alert := template.Get(config.Theme).Alert().
		SetTitle(template2.HTML(`<i class="icon fa fa-warning"></i> ` + language.Get("error") + `!`)).
		SetTheme("warning").
		SetContent(template2.HTML(msg)).
		GetContent()

	tmpl, tmplName := template.Get(config.Theme).GetTemplate(ctx.Headers(constant.PjaxHeader) == "true")
	buf := template.Execute(tmpl, tmplName, user, types.Panel{
		Content:     alert,
		Description: desc,
		Title:       title,
	}, config, menu.GetGlobalMenu(user, conn).SetActiveClass(config.URLRemovePrefix(ctx.Path())))
	ctx.HTML(http.StatusOK, buf.String())
}

func Error(ctx *context.Context, msg string) {
	ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
		"code": 500,
		"msg":  language.Get(msg),
	})
}
