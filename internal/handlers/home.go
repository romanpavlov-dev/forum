package handlers

import (
	"log"
	"net/http"
	"text/template"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Println("Forbidden Method")
		writeJSONError(w, "Method is not Allowed", http.StatusMethodNotAllowed)
		return
	}

	tmp, err := template.ParseFiles("web/index.html")
	if err != nil {
		http.Error(w, "Failed to Parse the tmp", http.StatusInternalServerError)
		return
	}

	if err := tmp.Execute(w, nil); err != nil {
		http.Error(w, "Failed to Exec the tmp", http.StatusInternalServerError)
		return
	}

}
