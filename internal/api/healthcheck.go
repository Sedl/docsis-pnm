package api

import (
	"fmt"
	"net/http"
)

func (*Api) healthStatus(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprint(w, "OK")
}
