package controller

import (
	"github.com/glvd/go-admin/context"
	"github.com/glvd/go-admin/modules/logger"
	"github.com/glvd/go-admin/plugins/admin/modules/guard"
	"github.com/glvd/go-admin/plugins/admin/modules/response"
	"github.com/glvd/go-admin/plugins/admin/modules/table"
)

// Delete delete the row from database.
func Delete(ctx *context.Context) {

	param := guard.GetDeleteParam(ctx)

	//token := ctx.FormValue("_t")
	//
	//if !auth.TokenHelper.CheckToken(token) {
	//	ctx.SetStatusCode(http.StatusBadRequest)
	//	ctx.WriteString(`{"code":400, "msg":"delete fail"}`)
	//	return
	//}

	if err := table.Get(param.Prefix).DeleteDataFromDatabase(param.Id); err != nil {
		logger.Error(err)
		response.Error(ctx, "删除失败")
		return
	}

	newToken := authSrv().AddToken()

	response.OkWithData(ctx, map[string]interface{}{
		"token": newToken,
	})
}
