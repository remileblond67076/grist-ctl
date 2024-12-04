package gristapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"sort"
	"strconv"
	"sync"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

type User struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Access       string `json:"access"`
	ParentAccess string `json:"parentAccess"`
}

type Org struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	CreatedAt string `json:"createdAt"`
}

type Workspace struct {
	Id                 int    `json:"id"`
	Name               string `json:"name"`
	CreatedAt          string `json:"createdAt"`
	Docs               []Doc  `json:"docs"`
	IsSupportWorkspace string `json:"isSupportWorkspace"`
	OrgDomain          string `json:"orgDomain"`
	Org                Org    `json:"org"`
	Access             string `json:"access"`
}

type EntityAccess struct {
	MaxInheritedRole string `json:"maxInheritedRole"`
	Users            []User `json:"users"`
}

type Doc struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	IsPinned  bool      `json:"isPinned"`
	Workspace Workspace `json:"workspace"`
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

func init() {
	if os.Getenv("GRIST_TOKEN") == "" || os.Getenv("GRIST_URL") == "" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("Erreur lors de la lecture du fichier de configuration\n", err)
		}
	}
}

func get(myRequest string) string {
	// Envoi d'une requête HTTP GET à l'API REST de Grist
	// Retourne le corps de la réponse
	client := &http.Client{}

	url := fmt.Sprintf("%s/api/%s", os.Getenv("GRIST_URL"), myRequest)
	bearer := "Bearer " + os.Getenv("GRIST_TOKEN")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating request %s: %s", url, err)
	}
	req.Header.Add("Authorization", bearer)

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request %s: %s", url, err)
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		log.Fatal("HTTP Error %s: %s", url, resp.Status)
	}

	// Read the HTTP response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response %s: %s", url, err)
	}
	return string(body)
}

func post(myRequest string, data []byte) string {
	// Envoi d'une requête HTTP POST à l'API REST de Grist avec une charge de données
	// Retourne le corps de la réponse
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/%s", os.Getenv("GRIST_URL"), myRequest)
	bearer := "Bearer " + os.Getenv("GRIST_TOKEN")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Fatal("Error creating request %s: %s", url, err)
	}
	req.Header.Add("Authorization", bearer)

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request %s: %s", url, err)
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		log.Fatal("HTTP Error %s: %s", url, resp.Status)
	}

	// Read the HTTP response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response %s: %s", url, err)
	}
	return string(body)

}

func delete(myRequest string) string {
	// Envoi d'une requête HTTP DELETE à l'API REST de Grist avec une charge de données
	// Retourne le corps de la réponse
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/%s", os.Getenv("GRIST_URL"), myRequest)
	bearer := "Bearer " + os.Getenv("GRIST_TOKEN")

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Fatal("Error creating request %s: %s", url, err)
	}
	req.Header.Add("Authorization", bearer)

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request %s: %s", url, err)
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		log.Fatal("HTTP Error %s: %s", url, resp.Status)
	}

	// Read the HTTP response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response %s: %s", url, err)
	}
	return string(body)
}

func GetOrgs() []Org {
	// Récupère la liste des organisations
	myOrgs := []Org{}
	response := get("orgs")
	json.Unmarshal([]byte(response), &myOrgs)
	return myOrgs
}

func GetOrg(idOrg string) Org {
	// Récupère l'organisation dont l'identifiant est passé en paramètre
	myOrg := Org{}
	response := get("orgs/" + idOrg)
	json.Unmarshal([]byte(response), &myOrg)
	return myOrg
}

func GetOrgAccess(idOrg string) []User {
	// Récupère la liste des utilisateurs de l'organisation dont l'identifiant est passé en paramètre
	var lstUsers EntityAccess
	url := fmt.Sprintf("orgs/%s/access", idOrg)
	response := get(url)
	json.Unmarshal([]byte(response), &lstUsers)
	return lstUsers.Users
}

func DisplayOrgAccess(idOrg string) {
	// Affiche la liste des utilisateurs ayant accès à une organisation
	lstUsers := GetOrgAccess(idOrg)
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
	fmt.Fprintln(w, "Id\tNom")
	for _, user := range lstUsers {
		fmt.Fprintf(w, "%s\t%s\t%s\n", user.Email, user.Name, user.Access)
	}
	w.Flush()
	fmt.Printf("%d utilisateurs\n", len(lstUsers))
}

func GetOrgWorkspaces(orgId int) []Workspace {
	// Récupère les information sur une organisation particulière
	lstWorkspaces := []Workspace{}
	response := get("orgs/" + strconv.Itoa(orgId) + "/workspaces")
	json.Unmarshal([]byte(response), &lstWorkspaces)
	return lstWorkspaces
}

func GetWorkspace(workspaceId int) Workspace {
	// Récupère un workspace
	workspace := Workspace{}
	url := fmt.Sprintf("workspaces/%d", workspaceId)
	response := get(url)
	json.Unmarshal([]byte(response), &workspace)
	return workspace
}

func DeleteWorkspace(workspaceId int) {
	// Supprime un workspace
	url := fmt.Sprintf("workspaces/%d", workspaceId)
	delete(url)
}

func GetWorkspaceAccess(workspaceId int) EntityAccess {
	// Récupère les droits d'accès à un workspace
	workspaceAccess := EntityAccess{}
	url := fmt.Sprintf("workspaces/%d/access", workspaceId)
	response := get(url)
	json.Unmarshal([]byte(response), &workspaceAccess)
	return workspaceAccess
}

func GetDoc(docId string) Doc {
	// Récupère les informations relatives à un document particulier
	doc := Doc{}
	url := "docs/" + docId
	response := get(url)
	json.Unmarshal([]byte(response), &doc)
	return doc
}

func GetDocTables(docId string) Tables {
	// Récupère la liste des tables contenues dans un document
	tables := Tables{}
	url := "docs/" + docId + "/tables"
	response := get(url)
	json.Unmarshal([]byte(response), &tables)

	return tables
}

func GetTableColumns(docId string, tableId string) TableColumns {
	// Récupère la liste des colonnes d'une table
	columns := TableColumns{}
	url := "docs/" + docId + "/tables/" + tableId + "/columns"
	response := get(url)
	json.Unmarshal([]byte(response), &columns)

	return columns
}

func GetTableRows(docId string, tableId string) TableRows {
	// Récupère les données d'une table
	rows := TableRows{}
	url := "docs/" + docId + "/tables/" + tableId + "/data"
	response := get(url)
	json.Unmarshal([]byte(response), &rows)

	return rows
}

func DisplayDoc(docId string) {
	// Affiche les informations détaillées sur un document
	// - Nom du document
	// - Nombre de tables
	//   Pour chacune d'entre elles :
	//   - Nombre de colonnes
	//   - Nombre de ligne
	//   - Liste des colonnes
	doc := GetDoc(docId)

	type TableDetails struct {
		name       string
		nb_rows    int
		nb_cols    int
		cols_names []string
	}

	title := color.New(color.FgRed).SprintFunc()
	fmt.Printf("\nDocument %s (%s)", title(doc.Name), doc.Id)
	if doc.IsPinned {
		fmt.Printf(" - épinglé\n")
	} else {
		fmt.Printf(" - non épinglé\n")
	}

	var tables Tables = GetDocTables(docId)
	fmt.Printf("Contient %d tables\n", len(tables.Tables))
	var wg sync.WaitGroup
	var tables_details []TableDetails
	for _, table := range tables.Tables {
		wg.Add(1)
		go func() {
			defer wg.Done()
			table_desc := ""
			columns := GetTableColumns(docId, table.Id)
			rows := GetTableRows(docId, table.Id)

			var cols_names []string
			for _, col := range columns.Columns {
				cols_names = append(cols_names, col.Id)
			}
			slices.Sort(cols_names)
			for _, col := range cols_names {
				table_desc += fmt.Sprintf("%s ", col)
			}
			table_info := TableDetails{
				name:       table.Id,
				nb_rows:    len(rows.Id),
				nb_cols:    len(columns.Columns),
				cols_names: cols_names,
			}
			tables_details = append(tables_details, table_info)
		}()
	}
	wg.Wait()
	var details []string
	for _, table_details := range tables_details {
		ligne := fmt.Sprintf("- %s : %d lignes, %d colonnes\n", title(table_details.name), table_details.nb_rows, table_details.nb_cols)
		for _, col_name := range table_details.cols_names {
			ligne = ligne + fmt.Sprintf("  - %s\n", col_name)
		}
		details = append(details, ligne)
	}
	sort.Strings(details)
	for _, ligne := range details {
		fmt.Printf("%s", ligne)
	}
}

func DisplayOrgs() {
	// Affiche la liste des organisations accessibles
	lstOrgs := GetOrgs()
	fmt.Printf("%d organisations trouvées:\n\n", len(lstOrgs))
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
	fmt.Fprintln(w, "Id\tNom")
	for _, org := range lstOrgs {
		fmt.Fprintf(w, "%d\t%s\n", org.Id, org.Name)
	}
	w.Flush()
}

func DisplayOrg(orgId string) {
	// Affiche les détails sur une organisation
	org := GetOrg(orgId)
	fmt.Printf("Organisation n°%d : %s\n", org.Id, org.Name)
	worskspaces := GetOrgWorkspaces(org.Id)

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
	fmt.Fprintln(w, "Id workspace\tNom Workspace\tNb doc")
	for _, ws := range worskspaces {
		fmt.Fprintf(w, "%d\t%s\t%d\n", ws.Id, ws.Name, len(ws.Docs))
	}
	w.Flush()
	fmt.Printf("%d workspaces\n", len(worskspaces))
}

func DisplayWorkspace(workspaceId int) {
	// Affiche des détails d'un Workspace

	ws := GetWorkspace(workspaceId)
	fmt.Printf("Organisation n°%d : %s\n", ws.Org.Id, ws.Org.Name)
	fmt.Printf("Workspace n°%d : %s\n\n", ws.Id, ws.Name)

	if len(ws.Docs) > 1 {
		fmt.Printf("Contient %d documents :\n", len(ws.Docs))
		w := tabwriter.NewWriter(os.Stdout, 1, 1, 5, ' ', 0)
		fmt.Fprintf(w, "Id\tNom\tÉpinglé\n")
		for _, doc := range ws.Docs {
			pin := ""
			if doc.IsPinned {
				pin = "✅"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", doc.Id, doc.Name, pin)
		}
		w.Flush()
	} else {
		fmt.Println("Ne contient aucun document")
	}
}

func DisplayWorkspaceAccess(workspaceId int) {
	// Affiche les droits d'accès à un workspace

	ws := GetWorkspace((workspaceId))
	wsa := GetWorkspaceAccess(workspaceId)

	fmt.Printf("Droits d'accès au workspace n°%d : %s\n", ws.Id, ws.Name)
	displayMaxInheritedRole(wsa.MaxInheritedRole)

	nbUsers := len(wsa.Users)
	if nbUsers <= 0 {
		fmt.Println("Accessible à aucun utilisateur")
	} else {
		nbUser := 0
		fmt.Println("\nAccessible utilisateurs suivants :")
		w := tabwriter.NewWriter(os.Stdout, 5, 1, 5, ' ', 0)
		fmt.Fprintf(w, "Id\tNom\tEmail\tAccès direct\tAccès hérité\n")
		for _, user := range wsa.Users {
			if user.Access != "" || user.ParentAccess != "" {
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", user.Id, user.Name, user.Email, user.Access, user.ParentAccess)
				nbUser += 1
			}
		}
		w.Flush()
		fmt.Printf("%d utilisateurs\n", nbUser)
	}
}

func GetDocAccess(docId string) EntityAccess {
	// Retourne la liste des utilisateurs ayant accès au document
	var lstUsers EntityAccess
	response := get("docs/" + docId + "/access")
	json.Unmarshal([]byte(response), &lstUsers)
	return lstUsers
}

func displayMaxInheritedRole(maxInheritedRole string) {
	switch maxInheritedRole {
	case "":
		fmt.Println("Aucun héritage de droit depuis le niveau supérieur")
	case "owners":
		fmt.Println("Héritage complet des droits depuis le niveau supérieur")
	case "editors":
		fmt.Println("Héritage des droits d'afficher et éditer depuis le niveau supérieur")
	case "viewers":
		fmt.Println("Héritage des droits de consultation depuis le niveau supérieur")
	default:
		fmt.Printf("Niveau d'héritage : %s\n", maxInheritedRole)
	}
}

func DisplayDocAccess(docId string) {
	doc := GetDoc(docId)
	title := fmt.Sprintf("Workspace \"%s\" (n°%d)\nDocument \"%s\"\n", doc.Workspace.Name, doc.Workspace.Id, doc.Name)
	fmt.Println(title)
	docAccess := GetDocAccess(docId)
	displayMaxInheritedRole(docAccess.MaxInheritedRole)
	fmt.Printf("\nUtilisateurs directs:\n")
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
	fmt.Fprintln(w, "Id\tEmail\tNom\tAccès")
	for _, user := range docAccess.Users {
		if user.Access != "" {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", user.Id, user.Email, user.Name, user.Access)
		}
	}
	w.Flush()
}

func PurgeDoc(docId string) {
	// Purge l'historique d'un document, pour ne conserver que les trois dernières modifications
	url := "docs/" + docId + "/states/remove"
	data := []byte(`{"keep": "3"}`)
	response := post(url, data)
	println(response)
}
