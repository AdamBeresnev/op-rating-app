package views

import (
	"github.com/a-h/templ"
	"net/http"
)

func Render(w http.ResponseWriter, r *http.Request, component templ.Component) error {
	return component.Render(r.Context(), w)
}
