package controller

import (
	"fmt"
	"github.com/glvd/go-admin/context"
	"github.com/glvd/go-admin/modules/auth"
	"github.com/glvd/go-admin/modules/file"
	"github.com/glvd/go-admin/modules/language"
	"github.com/glvd/go-admin/modules/menu"
	"github.com/glvd/go-admin/plugins/admin/modules"
	"github.com/glvd/go-admin/plugins/admin/modules/constant"
	"github.com/glvd/go-admin/plugins/admin/modules/guard"
	"github.com/glvd/go-admin/plugins/admin/modules/table"
	"github.com/glvd/go-admin/template"
	"github.com/glvd/go-admin/template/types"
	template2 "html/template"
	"net/http"
)

// ShowForm show form page.
func ShowForm(ctx *context.Context) {
	param := guard.GetShowFormParam(ctx)
	showForm(ctx, "", param.Prefix, param.Id, param.GetUrl(), param.GetInfoUrl(), "")
}

func showForm(ctx *context.Context, alert template2.HTML, prefix string, id string, url, infoUrl string, editUrl string) {

	table.RefreshTableList()
	panel := table.Get(prefix)

	formData, groupFormData, groupHeaders, title, description, err := panel.GetDataFromDatabaseWithId(id)

	if err != nil && alert == "" {
		alert = aAlert().SetTitle(template2.HTML(`<i class="icon fa fa-warning"></i> ` + language.Get("error") + `!`)).
			SetTheme("warning").
			SetContent(template2.HTML(err.Error())).
			GetContent()
	}

	user := auth.Auth(ctx)

	referer := ctx.Headers("Referer")

	if referer != "" && !modules.IsInfoUrl(referer) && !modules.IsEditUrl(referer, ctx.Query("__prefix")) {
		infoUrl = referer
	}

	tmpl, tmplName := aTemplate().GetTemplate(isPjax(ctx))
	buf := template.Execute(tmpl, tmplName, user, types.Panel{
		Content: alert + formContent(aForm().
			SetContent(formData).
			SetTabContents(groupFormData).
			SetTabHeaders(groupHeaders).
			SetPrefix(config.PrefixFixSlash()).
			SetPrimaryKey(panel.GetPrimaryKey().Name).
			SetUrl(url).
			SetToken(authSrv().AddToken()).
			SetInfoUrl(infoUrl).
			SetOperationFooter(formFooter()).
			SetHeader(panel.GetForm().HeaderHtml).
			SetFooter(panel.GetForm().FooterHtml)),
		Description: description,
		Title:       title,
	}, config, menu.GetGlobalMenu(user).SetActiveClass(config.URLRemovePrefix(ctx.Path())))

	ctx.HTML(http.StatusOK, buf.String())

	if editUrl != "" {
		ctx.AddHeader(constant.PjaxUrlHeader, editUrl)
	}
}

func EditForm(ctx *context.Context) {

	param := guard.GetEditFormParam(ctx)

	if param.HasAlert() {
		showForm(ctx, param.Alert, param.Prefix, param.Id, param.GetUrl(), param.GetInfoUrl(), param.GetEditUrl())
		return
	}

	// process uploading files, only support local storage for now.
	if len(param.MultiForm.File) > 0 {
		err := file.GetFileEngine(config.FileUploadEngine.Name).Upload(param.MultiForm)
		if err != nil {
			alert := aAlert().SetTitle(template2.HTML(`<i class="icon fa fa-warning"></i> ` + language.Get("error") + `!`)).
				SetTheme("warning").
				SetContent(template2.HTML(err.Error())).
				GetContent()
			showForm(ctx, alert, param.Prefix, param.Id, param.GetUrl(), param.GetInfoUrl(), param.GetEditUrl())
			return
		}
	}

	err := param.Panel.UpdateDataFromDatabase(param.Value())
	if err != nil {
		alert := aAlert().SetTitle(template2.HTML(`<i class="icon fa fa-warning"></i> ` + language.Get("error") + `!`)).
			SetTheme("warning").
			SetContent(template2.HTML(err.Error())).
			GetContent()
		showForm(ctx, alert, param.Prefix, param.Id, param.GetUrl(), param.GetInfoUrl(), param.GetEditUrl())
		return
	}

	if !param.FromList {
		ctx.HTML(http.StatusOK, fmt.Sprintf(`<script>location.href="%s"</script>`, param.PreviousPath))
		ctx.AddHeader(constant.PjaxUrlHeader, param.PreviousPath)
		return
	}

	editUrl := modules.AorB(param.Panel.GetEditable(), param.GetEditUrl(), "")
	deleteUrl := modules.AorB(param.Panel.GetDeletable(), param.GetDeleteUrl(), "")
	exportUrl := modules.AorB(param.Panel.GetExportable(), param.GetExportUrl(), "")
	newUrl := modules.AorB(param.Panel.GetCanAdd(), param.GetNewUrl(), "")
	infoUrl := param.GetInfoUrl()
	updateUrl := modules.AorB(param.Panel.GetEditable(), param.GetUpdateUrl(), "")

	buf := showTable(ctx, param.Panel, param.Path, param.Param, exportUrl, newUrl, deleteUrl, infoUrl, editUrl, updateUrl)

	ctx.HTML(http.StatusOK, buf.String())
	ctx.AddHeader(constant.PjaxUrlHeader, param.PreviousPath)
}
