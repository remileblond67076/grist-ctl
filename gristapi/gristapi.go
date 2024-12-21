package gristapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-gota/gota/dataframe"
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

func GetConfig() string {
	// Apply config and return the config file path
	home := os.Getenv("HOME")
	configFile := filepath.Join(home, ".gristctl")
	if os.Getenv("GRIST_TOKEN") == "" || os.Getenv("GRIST_URL") == "" {
		err := godotenv.Load(configFile)
		if err != nil {
			fmt.Printf("Error reading configuration file : %s\n", err)
		}
	}
	return configFile
}

func init() {
	GetConfig()
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

func GetDocAccess(docId string) EntityAccess {
	// Returns the list of users with access to the document

	var lstUsers EntityAccess
	url := fmt.Sprintf("docs/%s/access", docId)
	response := httpGet(url, "")
	json.Unmarshal([]byte(response), &lstUsers)
	return lstUsers
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

func ExportDocGrist(docId string) {
	// Export doc in Grist format (Sqlite)
	url := fmt.Sprintf("docs/%s/download", docId)
	file := httpGet(url, "")
	fmt.Println(file)
}

func ExportDocExcel(docId string) {
	// Export doc in Excel format (XLSX)
	url := fmt.Sprintf("docs/%s/download/xlsx", docId)
	file := httpGet(url, "")
	fmt.Println(file)
}

func GetTableContent(docId string, tableName string) dataframe.DataFrame {
	// Returns table content as Dataframe
	url := fmt.Sprintf("docs/%s/tables/%s/records", docId, tableName)
	response := httpGet(url, "")
	type GristRecord struct {
		ID     int            `json:"id"`
		Fields map[string]any `json:"fields"`
	}

	type GristResponse struct {
		Records []GristRecord `json:"records"`
	}

	var gristResponse GristResponse
	err := json.Unmarshal([]byte(response), &gristResponse)
	if err != nil {
		log.Fatalf("Erreur lors du décodage du JSON: %v", err)
	}

	df := dataframe.LoadStructs(gristResponse.Records)
	return df
}
