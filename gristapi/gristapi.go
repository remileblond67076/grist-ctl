package gristapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gristctl/common"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
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

type UserRole struct {
	Email string
	Role  string
}

func init() {
	home := os.Getenv("HOME")
	configFile := filepath.Join(home, ".gristctl")
	if os.Getenv("GRIST_TOKEN") == "" || os.Getenv("GRIST_URL") == "" {
		err := godotenv.Load(configFile)
		if err != nil {
			log.Fatalf("Error reading configuration file : %s\n", err)
		}
	}
}

func httpRequest(action string, myRequest string, data *bytes.Buffer) (string, int) {
	// Sending an HTTP request to Grist's REST API
	// Returns response body

	client := &http.Client{}
	url := fmt.Sprintf("%s/api/%s", os.Getenv("GRIST_URL"), myRequest)
	bearer := "Bearer " + os.Getenv("GRIST_TOKEN")

	req, err := http.NewRequest(action, url, data)
	if err != nil {
		log.Fatalf("Error creating request %s: %s", url, err)
	}
	req.Header.Add("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request %s: %s", url, err)
	}
	defer resp.Body.Close()

	// Read the HTTP response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response %s: %s", url, err)
	}
	return string(body), resp.StatusCode
}

func httpGet(myRequest string, data string) string {
	// Send an HTTP GET request to Grist's REST API
	// Returns the response body
	dataBody := bytes.NewBuffer([]byte(data))
	body, status := httpRequest("GET", myRequest, dataBody)
	if status != http.StatusOK {
		fmt.Printf("Return code from %s : %d (%s)\n", myRequest, status, body)
	}
	return body
}

func httpPost(myRequest string, data string) (string, int) {
	// Sends an HTTP POST request to Grist's REST API with a data load
	// Return the response body

	dataBody := bytes.NewBuffer([]byte(data))
	body, status := httpRequest("POST", myRequest, dataBody)
	return body, status
}

func httpPatch(myRequest string, data string) (string, int) {
	// Sends an HTTP POST request to Grist's REST API with a data load
	// Return the response body

	dataBody := bytes.NewBuffer([]byte(data))
	body, status := httpRequest("PATCH", myRequest, dataBody)
	return body, status
}

func httpDelete(myRequest string, data string) (string, int) {
	// Send an HTTP DELETE request to Grist's REST API with a data load
	// Return the response body

	dataBody := bytes.NewBuffer([]byte(data))
	body, status := httpRequest("DELETE", myRequest, dataBody)
	return body, status
}

func GetOrgs() []Org {
	// Retrieves the list of organizations

	myOrgs := []Org{}
	response := httpGet("orgs", "")
	json.Unmarshal([]byte(response), &myOrgs)
	return myOrgs
}

func GetOrg(idOrg string) Org {
	// Retrieves the organization whose identifier is passed in parameter

	myOrg := Org{}
	response := httpGet("orgs/"+idOrg, "")
	json.Unmarshal([]byte(response), &myOrg)
	return myOrg
}

func GetOrgAccess(idOrg string) []User {
	// Retrieves the list of users in the organization whose ID is passed in parameter

	var lstUsers EntityAccess
	url := fmt.Sprintf("orgs/%s/access", idOrg)
	response := httpGet(url, "")
	json.Unmarshal([]byte(response), &lstUsers)
	return lstUsers.Users
}

func DisplayOrgAccess(idOrg string) {
	// Displays the list of users with access to an organization

	lstUsers := GetOrgAccess(idOrg)
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, '|', 0)
	fmt.Fprintln(w, "Id\tNom")
	for _, user := range lstUsers {
		fmt.Fprintf(w, "%s\t%s\t%s\n", user.Email, user.Name, user.Access)
	}
	w.Flush()
	fmt.Printf("%d utilisateurs\n", len(lstUsers))
}

func GetOrgWorkspaces(orgId int) []Workspace {
	// Retrieves information on a specific organization
	lstWorkspaces := []Workspace{}
	response := httpGet("orgs/"+strconv.Itoa(orgId)+"/workspaces", "")
	json.Unmarshal([]byte(response), &lstWorkspaces)
	return lstWorkspaces
}

func GetWorkspace(workspaceId int) Workspace {
	// Recovers a workspace
	workspace := Workspace{}
	url := fmt.Sprintf("workspaces/%d", workspaceId)
	response := httpGet(url, "")
	json.Unmarshal([]byte(response), &workspace)
	return workspace
}

func DeleteWorkspace(workspaceId int) {
	// Delete a workspace

	url := fmt.Sprintf("workspaces/%d", workspaceId)
	response, status := httpDelete(url, "")
	if status == http.StatusOK {
		fmt.Printf("Workspace %d deleted\t✅\n", workspaceId)
	} else {
		fmt.Printf("Unable to delete workspace %d : %s ❗️\n", workspaceId, response)
	}
}

func DeleteDoc(docId string) {
	// Delete a document

	url := fmt.Sprintf("docs/%s", docId)
	response, status := httpDelete(url, "")
	if status == http.StatusOK {
		fmt.Printf("Document %s deleted\t✅\n", docId)
	} else {
		fmt.Printf("Unable to delete document %s : %s ❗️", docId, response)
	}
}

func DeleteUser(userId int) {
	// Delete a user

	url := fmt.Sprintf("users/%d", userId)
	response, status := httpDelete(url, `{"name": ""}`)

	var message string
	switch status {
	case 200:
		message = "The account has been deleted successfully"
	case 400:
		message = "The passed user name does not match the one retrieved from the database given the passed user id"
	case 403:
		message = "The caller is not allowed to delete this account"
	case 404:
		message = "The user is not found"
	}
	fmt.Println(message)
	if status != http.StatusOK {
		fmt.Printf("ERREUR: %s\n", response)
	}
}

func GetWorkspaceAccess(workspaceId int) EntityAccess {
	// Workspace access rights query

	workspaceAccess := EntityAccess{}
	url := fmt.Sprintf("workspaces/%d/access", workspaceId)
	response := httpGet(url, "")
	json.Unmarshal([]byte(response), &workspaceAccess)
	return workspaceAccess
}

func GetDoc(docId string) Doc {
	// Retrieves information about a specific document

	doc := Doc{}
	url := "docs/" + docId
	response := httpGet(url, "")
	json.Unmarshal([]byte(response), &doc)
	return doc
}

func GetDocTables(docId string) Tables {
	// Retrieves the list of tables contained in a document
	tables := Tables{}
	url := "docs/" + docId + "/tables"
	response := httpGet(url, "")
	json.Unmarshal([]byte(response), &tables)

	return tables
}

func GetTableColumns(docId string, tableId string) TableColumns {
	// Retrieves a list of table columns

	columns := TableColumns{}
	url := "docs/" + docId + "/tables/" + tableId + "/columns"
	response := httpGet(url, "")
	json.Unmarshal([]byte(response), &columns)

	return columns
}

func GetTableRows(docId string, tableId string) TableRows {
	// Retrieves records from a table

	rows := TableRows{}
	url := "docs/" + docId + "/tables/" + tableId + "/data"
	response := httpGet(url, "")
	json.Unmarshal([]byte(response), &rows)

	return rows
}

func DisplayDoc(docId string) {
	// Displays detailed information about a document
	// - Document name
	// - Number of tables
	// For each table :
	// - Number of columns
	// - Number of rows
	// - List of columns

	doc := GetDoc(docId)

	type TableDetails struct {
		name       string
		nb_rows    int
		nb_cols    int
		cols_names []string
	}

	title := color.New(color.FgRed).SprintFunc()
	pinned := ""
	if doc.IsPinned {
		pinned = "📌"
	}
	common.DisplayTitle(fmt.Sprintf("Document %s (%s) %s", title(doc.Name), doc.Id, pinned))

	var tables Tables = GetDocTables(docId)
	fmt.Printf("Contains %d tables\n", len(tables.Tables))
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
		ligne := fmt.Sprintf("- %s : %d lines, %d colomns\n", title(table_details.name), table_details.nb_rows, table_details.nb_cols)
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
	// Displays the list of accessible organizations

	lstOrgs := GetOrgs()
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
	fmt.Fprintln(w, "Id\tNom")
	for _, org := range lstOrgs {
		fmt.Fprintf(w, "%d\t%s\n", org.Id, org.Name)
	}
	w.Flush()
}

func DisplayOrg(orgId string) {
	// Displays details about an organization

	type wpDesc struct {
		id     int
		name   string
		nbDoc  int
		nbUser int
	}
	var lstWsDesc []wpDesc

	org := GetOrg(orgId)
	worskspaces := GetOrgWorkspaces(org.Id)
	common.DisplayTitle(fmt.Sprintf("Organization n°%d : %s (%d workspaces)", org.Id, org.Name, len(worskspaces)))

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
	fmt.Fprintln(w, "Workspace Id\tWorkspace name\tDoc\tDirect users")
	var wg sync.WaitGroup
	for _, ws := range worskspaces {
		func() {
			defer wg.Done()
			wg.Add(1)
			users := GetWorkspaceAccess(ws.Id)
			nbUsers := 0
			for _, user := range users.Users {
				if user.Access != "" {
					nbUsers += 1
				}
			}
			lstWsDesc = append(lstWsDesc, wpDesc{ws.Id, ws.Name, len(ws.Docs), nbUsers})
		}()
	}
	wg.Wait()

	for _, desc := range lstWsDesc {
		fmt.Fprintf(w, "%d\t%s\t%d\t%d\n", desc.id, desc.name, desc.nbDoc, desc.nbUser)
	}
	w.Flush()
}

func DisplayWorkspace(workspaceId int) {
	// Affiche des détails d'un Workspace

	ws := GetWorkspace(workspaceId)
	common.DisplayTitle(fmt.Sprintf("Organization n°%d : \"%s\", workspace n°%d : \"%s\"", ws.Org.Id, ws.Org.Name, ws.Id, ws.Name))

	if len(ws.Docs) > 0 {
		fmt.Printf("Contains %d documents :\n", len(ws.Docs))
		w := tabwriter.NewWriter(os.Stdout, 1, 1, 5, ' ', 0)
		fmt.Fprintf(w, "Id\tName\tPinned\n")
		for _, doc := range ws.Docs {
			pin := ""
			if doc.IsPinned {
				pin = "📌"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", doc.Id, doc.Name, pin)
		}
		w.Flush()
	} else {
		fmt.Println("No documents")
	}
}

func DisplayWorkspaceAccess(workspaceId int) {
	// Displays workspace access rights

	ws := GetWorkspace((workspaceId))
	wsa := GetWorkspaceAccess(workspaceId)

	common.DisplayTitle(fmt.Sprintf("Workspace n°%d access rights : %s", ws.Id, ws.Name))
	displayRole(wsa.MaxInheritedRole)

	nbUsers := len(wsa.Users)
	if nbUsers <= 0 {
		fmt.Println("Accessible to no user")
	} else {
		nbUser := 0
		fmt.Println("\nAccessible to the following users :")
		w := tabwriter.NewWriter(os.Stdout, 5, 1, 5, ' ', 0)
		fmt.Fprintf(w, "Id\tNom\tEmail\tInherited access\tDirect access\n")
		for _, user := range wsa.Users {
			if user.Access != "" || user.ParentAccess != "" {
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", user.Id, user.Name, user.Email, user.ParentAccess, user.Access)
				nbUser += 1
			}
		}
		w.Flush()
		fmt.Printf("%d users\n", nbUser)
	}
}

func GetDocAccess(docId string) EntityAccess {
	// Returns the list of users with access to the document

	var lstUsers EntityAccess
	url := fmt.Sprintf("docs/%s/access", docId)
	response := httpGet(url, "")
	json.Unmarshal([]byte(response), &lstUsers)
	return lstUsers
}

func displayRole(role string) {
	// User role translation

	switch role {
	case "":
		fmt.Println("No inheritance of rights from upper level")
	case "owners":
		fmt.Println("Full inheritance of rights from the next level up")
	case "editors":
		fmt.Println("Inherit display and edit rights from higher level")
	case "viewers":
		fmt.Println("Inheritance of consultation rights from higher level")
	default:
		fmt.Printf("Inheritance level : %s\n", role)
	}
}

func DisplayDocAccess(docId string) {
	// Displays users with access to a document

	doc := GetDoc(docId)
	title := fmt.Sprintf("Workspace \"%s\" (n°%d)\nDocument \"%s\"\n", doc.Workspace.Name, doc.Workspace.Id, doc.Name)
	fmt.Println(title)
	docAccess := GetDocAccess(docId)
	displayRole(docAccess.MaxInheritedRole)
	fmt.Printf("\nDirect users:\n")
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
	fmt.Fprintln(w, "Id\tEmail\tNom\tInherited access\tDirect access")
	for _, user := range docAccess.Users {
		if user.Access != "" {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", user.Id, user.Email, user.Name, user.ParentAccess, user.Access)
		}
	}
	w.Flush()
}

func PurgeDoc(docId string, nbHisto int) {
	// Purge a document's history, to retain only the last modifications

	url := "docs/" + docId + "/states/remove"
	data := fmt.Sprintf(`{"keep": "%d"}`, nbHisto)
	_, status := httpPost(url, data)
	if status == http.StatusOK {
		fmt.Printf("History cleared (%d last states) ✅\n", nbHisto)
	}
}

func ImportUsers(orgId int, workspaceName string, users []UserRole) {
	// Import a list of user & role into a workspace

	// Search workspace by name in org
	lstWorkspaces := GetOrgWorkspaces(orgId)
	idWorkspace := 0
	for _, ws := range lstWorkspaces {
		if ws.Name == workspaceName {
			idWorkspace = ws.Id
		}
	}

	if idWorkspace == 0 {
		idWorkspace = CreateWorkspace(orgId, workspaceName)
	}
	if idWorkspace == 0 {
		fmt.Printf("Unable to create workspace %s\n", workspaceName)
	} else {
		url := fmt.Sprintf("workspaces/%d/access", idWorkspace)

		roleLine := []string{}
		for _, role := range users {
			roleLine = append(roleLine, fmt.Sprintf(`"%s": "%s"`, role.Email, role.Role))
		}
		patch := fmt.Sprintf(`{	"delta": { "users": {%s}}}`, strings.Join(roleLine, ","))

		body, status := httpPatch(url, patch)

		var result string
		if status == http.StatusOK {
			result = "✅"
		} else {
			result = fmt.Sprintf("❗️ (%s)", body)
		}
		fmt.Printf("Import %d users in workspace n°%d\t : %s\n", len(users), idWorkspace, result)
	}

}

func CreateWorkspace(orgId int, workspaceName string) int {
	// Create a workspace in an organization

	url := fmt.Sprintf("orgs/%d/workspaces", orgId)
	data := fmt.Sprintf(`{"name":"%s"}`, workspaceName)
	body, status := httpPost(url, data)
	idWorkspace := 0
	if status == http.StatusOK {
		id, err := strconv.Atoi(body)
		if err == nil {
			idWorkspace = id
		}
	}
	return idWorkspace
}

func DisplayUserMatrix() {
	// Displaying the rights matrix

	lstOrg := GetOrgs()
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
	fmt.Fprintf(w, "Id\tName\tAccess\tWokspace id\tWorkspace name\n")
	for _, org := range lstOrg {
		common.DisplayTitle(fmt.Sprintf("Org %s (%d)", org.Name, org.Id))
		for _, ws := range GetOrgWorkspaces(org.Id) {
			for _, user := range GetWorkspaceAccess(ws.Id).Users {
				if user.Access != "" {
					fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%d\n", user.Id, user.Email, user.Access, ws.Name, ws.Id)
				}
			}
		}
	}
	w.Flush()
}
