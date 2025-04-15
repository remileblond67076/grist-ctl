package gristtools

import (
	"fmt"
	"gristctl/gristapi"
	"html/template"
	"net/http"
)

func webIndex(w http.ResponseWriter, r *http.Request) {
	type PageData struct {
		Title           string
		Heading         string
		Message         string
		ContentTemplate string
	}

	// Parse the template file
	tmpl, err := template.ParseFiles("templates/base.html", "templates/content_index.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:           "accœil",
		Heading:         "administration de votre instance",
		Message:         "Liste des organisations",
		ContentTemplate: "content_index",
	}

	// Execute the template with the data
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}

func webOrg(w http.ResponseWriter, r *http.Request) {
	type PageData struct {
		Title           string
		Heading         string
		Message         string
		ContentTemplate string
		Orgs            []gristapi.Org
	}

	// Parse the template file
	tmpl, err := template.ParseFiles("templates/base.html", "templates/content_orgs.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:           "accœil",
		Heading:         "administration de votre instance",
		Message:         "Liste des organisations",
		ContentTemplate: "content_orgs",
		Orgs:            gristapi.GetOrgs(),
	}

	// Execute the template with the data
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}

func StartWebServer() {
	// Start the web server
	fmt.Println("Starting web server on port 8080")
	http.HandleFunc("/", webIndex)
	http.HandleFunc("/orgs", webOrg)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.ListenAndServe(":9090", nil)
}
