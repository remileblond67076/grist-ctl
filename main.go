package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

type Org struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	CreatedAt string `json:"createdAt"`
}

type Workspace struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	Docs      []Doc  `json:"docs"`
}

type Doc struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	IsPinned bool   `json:"isPinned"`
}

type Table struct {
	Id string `json:"id"`
}

type Tables struct {
	Tables []Table `json:"tables"`
}

type TableColumn struct {
	Id string `json:"id"`
}

type TableColumns struct {
	Columns []TableColumn `json:"columns"`
}

type TableRows struct {
	Id []uint `json:"id"`
}

func get(myRequest string) string {
	// Envoi d'une requête HTTP GET à l'API REST de Grist
	// Retourne le corps de la réponse
	client := &http.Client{}

	url := "https://wpgrist.cus.fr/api/" + myRequest
	bearer := "Bearer " + os.Getenv("TOKEN")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating request: %s", err)
	}
	req.Header.Add("Authorization", bearer)

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request: %s", err)
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		log.Fatal("HTTP Error: %s", resp.Status)
	}

	// Read the HTTP response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response: %s", err)
	}
	return string(body)
}

func getOrgs() []Org {
	// Récupère la liste des organisations
	myOrgs := []Org{}
	response := get("orgs")
	json.Unmarshal([]byte(response), &myOrgs)
	return myOrgs
}

func getOrg(idOrg string) Org {
	// Récupère l'organisation dont l'identifiant est passé en paramètre
	myOrg := Org{}
	response := get("orgs/" + idOrg)
	json.Unmarshal([]byte(response), &myOrg)
	return myOrg
}

func getOrgWorkspaces(id int) []Workspace {
	lstWorkspaces := []Workspace{}
	response := get("orgs/" + strconv.Itoa(id) + "/workspaces")
	json.Unmarshal([]byte(response), &lstWorkspaces)
	return lstWorkspaces
}

func getWorkspaceDocs(idWorkspace int) []Doc {
	lstDocs := []Doc{}
	url := "workspaces/" + strconv.Itoa((idWorkspace)) + "/docs"
	response := get(url)
	json.Unmarshal([]byte(response), &lstDocs)
	return lstDocs
}

func getDoc(id string) Doc {
	doc := Doc{}
	url := "docs/" + id
	response := get(url)
	json.Unmarshal([]byte(response), &doc)
	return doc
}

func getDocTables(id string) Tables {
	tables := Tables{}
	url := "docs/" + id + "/tables"
	response := get(url)
	json.Unmarshal([]byte(response), &tables)

	return tables
}

func getTableColuns(docId string, columnId string) TableColumns {
	columns := TableColumns{}
	url := "docs/" + docId + "/tables/" + columnId + "/columns"
	response := get(url)
	json.Unmarshal([]byte(response), &columns)

	return columns
}

func getTableRows(docId string, columnId string) TableRows {
	rows := TableRows{}
	url := "docs/" + docId + "/tables/" + columnId + "/data"
	response := get(url)
	json.Unmarshal([]byte(response), &rows)

	return rows
}

func getDocDetails(docId string) {
	// Affiche les informations détaillées sur un document
	// - Nom du document
	// - Nombre de tables
	//   Pour chacune d'entre elles :
	//   - Nombre de colonnes
	//   - Nombre de ligne
	//   - Lisre des colonnes
	doc := getDoc(docId)

	title := color.New(color.FgRed).SprintFunc()
	fmt.Printf("\nDocument '%s' (%s)\n", title(doc.Name), doc.Id)

	var tables Tables = getDocTables(docId)
	fmt.Printf("Contient %d tables\n", len(tables.Tables))
	var wg sync.WaitGroup
	for _, table := range tables.Tables {
		wg.Add(1)
		go func() {
			defer wg.Done()
			table_desc := ""
			columns := getTableColuns(docId, table.Id)
			rows := getTableRows(docId, table.Id)

			var table_names []string
			for _, col := range columns.Columns {
				table_names = append(table_names, col.Id)
			}
			slices.Sort(table_names)
			for _, col := range table_names {
				table_desc += fmt.Sprintf("%s ", col)
			}
			fmt.Printf("- %s : %d lignes, %d colonnes (%s)\n", title(table.Id), len(rows.Id), len(columns.Columns), strings.Join(table_names, ", "))
		}()
	}
	wg.Wait()
}

func getOrgsDetails() {
	for _, org := range getOrgs() {
		fmt.Printf("- Org n°%d (%s):\n", org.Id, org.Name)
		worskspaces := getOrgWorkspaces(org.Id)
		fmt.Printf("  %d workspaces : \n", len(worskspaces))
		for _, ws := range worskspaces {
			if len(ws.Docs) > 0 {
				var lst_docs []string
				for _, doc := range ws.Docs {
					lst_docs = append(lst_docs, doc.Name)
				}
				slices.Sort(lst_docs)
				fmt.Printf("  - n°%d %s: %d documents (%s)\n", ws.Id, ws.Name, len(ws.Docs), strings.Join(lst_docs, ", "))
			}
		}
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Erreur lors de la lecture du fichier de configuration\n", err)
	}
	var mainOrgId string = os.Getenv("ORGID")
	fmt.Printf("Organisation principale (n°%d) : %s\n", mainOrgId, getOrg(mainOrgId).Name)

	getOrgsDetails()

	docId := os.Getenv("DOCID")
	getDocDetails(docId)

}
