package wiki

import (
	"html/template"
	"net/http"
	"path"
	"strings"

	"github.com/MerryMage/libellus/common"
)

type Wiki struct {
	config *common.Config

	pageTemplate *template.Template
}

func NewWiki(config *common.Config) *Wiki {
	return &Wiki{
		config:       config,
		pageTemplate: template.Must(template.New("pageTemplate").Parse(config.StaticData.String("wiki/page_template.html"))),
	}
}

func (wiki *Wiki) invalidPathResponse(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	w.Write([]byte("404"))
}

func validatePath(p *string) bool {
	*p = path.Clean(*p)

	if (*p)[0] != '/' {
		return false
	}

	for _, ch := range *p {
		if (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') && (ch < '0' || ch > '9') && ch != '/' && ch != '-' {
			return false
		}
	}

	return true
}

func (wiki *Wiki) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !validatePath(&path) {
		wiki.invalidPathResponse(w, r)
		return
	}

	if path == "/refresh" {
		wiki.config.WikiData.RefreshState()
		w.Write([]byte("we're very refreshed"))
		return
	}

	page, ok := wiki.config.WikiData.LookupPage(path)
	if !ok {
		w.Write([]byte("!ok"))
		return
	}

	rendered := RenderedPage{
		Authorized: wiki.config.Authentication.IsAuthenticated(r),
		Title:      page.Title,
		Path:       RenderedPath(strings.Split(path[1:], "/")),
	}

	for _, v := range page.Children {
		rendered.Subpages = append(rendered.Subpages, RenderedSubpage{
			Path:  v,
			Title: v,
		})
	}

	for _, kid := range page.ActualKnowledges {
		k := wiki.RenderKnowledge(kid)
		rendered.Knowledges = append(rendered.Knowledges, k)
	}

	wiki.pageTemplate.Execute(w, rendered)
}
