package handlers

import (
	"fmt"
	"net/http"
)

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ok")
}