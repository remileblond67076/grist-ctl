// SPDX-FileCopyrightText: 2024 Ville Eurométropole Strasbourg
//
// SPDX-License-Identifier: MIT

// Common tools for Grist
package gristtools

import (
	"bufio"
	"fmt"
	"gristctl/common"
	"gristctl/gristapi"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/go-gota/gota/dataframe"
	"github.com/olekukonko/tablewriter"
)

// Display help message and quit
func Help() {
	common.DisplayTitle("GRIST : API querying")
	type command struct {
		cmd  string
		help string
	}

	cmdColor := color.New(color.FgRed).SprintFunc()
	commands := []command{
		{"config", "configure url & token of Grist server"},
		{"delete doc <id>", "delete a document"},
		{"delete user <id>", "delete a user"},
		{"delete workspace <id>", "delete a workspace"},
		{"get doc <id> access", "list of document access rights"},
		{"get doc <id> excel", "export document as <workspace name>_<doc name>.xlsx Excel file"},
		{"get doc <id> grist", "export document as <workspace name>_<doc name>.grist Grist file"},
		{"get doc <id> table <tableName>", "export content of a document's table as a CSV file (xlsx) in stdout"},
		{"get doc <id>", "document details"},
		{"get org <id>", "organization details"},
		{"get org", "organization list"},
		{"get users", "displays all user rights"},
		{"get workspace <id> access", "list of workspace access rights"},
		{"get workspace <id>", "workspace details"},
		{"import users", "imports users from standard input"},
		{"purge doc <id> [<number of states to keep>]", "purges document history (retains last 3 operations by default)"},
		{"version", "displays the version of the program"},
	}
	// Sort commands by name
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].cmd < commands[j].cmd
	})

	fmt.Println("Accepted orders :")
	for _, command := range commands {
		fmt.Printf("- %s : %s\n", cmdColor(command.cmd), command.help)
	}
	os.Exit(0)
}

// Displays the version of the program
func Version(version string) {
	fmt.Println("Version : ", version)
}

/*
Configure Grist envfile (url and api token)
Interactive filling the `.gristctl` file
*/
func Config() {
	configFile := gristapi.GetConfig()
	common.DisplayTitle(fmt.Sprintf("Setting the url and token for access to the grist server (%s)", configFile))
	fmt.Printf("Actual URL : %s\n", os.Getenv("GRIST_URL"))
	token := ""
	for i := 0; i < len(os.Getenv("GRIST_TOKEN")); i++ {
		token += "•"
	}
	fmt.Printf("Token : %s\n", token)
	testConnect := "❌"
	if gristapi.TestConnection() {
		testConnect = "✅"
	}
	fmt.Printf("Connection : %s\n", testConnect)

	if common.Confirm("Would you like to configure (Y/N) ?") {
		var url string
		urlSet := false
		for urlSet == false {
			fmt.Print("Grist server URL (that starts with https:// and without '/' in the end): ")
			fmt.Scanln(&url)

			// Test if url is well formatted
			urlOk, _ := regexp.MatchString(`^https?://.*[^/]$`, url)
			urlSet = urlOk
		}
		fmt.Print("User token (API key) : ")
		var token string
		fmt.Scanln(&token)
		if common.Confirm(fmt.Sprintf("Url : %s --- Token: %s\nIs it OK (Y/N) ? ", url, token)) {
			f, err := os.Create(configFile)
			if err != nil {
				fmt.Printf("Error on creating %s file (%s)", configFile, err)
				os.Exit(-1)
			}
			defer f.Close()
			config := fmt.Sprintf("GRIST_URL=\"%s\"\nGRIST_TOKEN=\"%s\"\n", url, token)
			f.WriteString(config)
			f.Close()
			fmt.Printf("Config saved in %s\n", configFile)

			// Test the configuration by connecting to the server
			nbOrgs := len(gristapi.GetOrgs())
			if nbOrgs <= 0 {
				fmt.Println("Error on connecting to the server. The config looks wrong.")
				os.Exit(-1)
			}
		}
	}
}

/*
User role translation

Returns the role explanation corresponding to its code
*/
func TranslateRole(roleCode string) string {
	var role string
	switch roleCode {
	case "":
		role = "No inheritance of rights from upper level"
	case "owners":
		role = "Full inheritance of rights from the next level up"
	case "editors":
		role = "Inherit display and edit rights from higher level"
	case "viewers":
		role = "Inheritance of consultation rights from higher level"
	default:
		role = fmt.Sprintf("Inheritance level : %s\n", roleCode)
	}
	return role
}

/*
Import users from a list sent to standard input (stdin)

CSV input file has to be formatied with the following columns, separated with ';' :
- mail
- org id
- Workspace name
- role

Missing workspaces will be created on import.
*/
func ImportUsers() {
	common.DisplayTitle("Import users from stdin")
	fmt.Println("Expected data format : <mail>;<org id>;<workspace name>;<role>")

	scanner := bufio.NewScanner(os.Stdin)
	type userAccess struct {
		Mail          string
		OrgId         int
		WorkspaceName string
		Role          string
	}
	lstUserAccess := []userAccess{}
	for scanner.Scan() {
		line := scanner.Text()
		data := strings.Split(line, ";")
		if len(data) == 4 {
			var lineOk bool = true
			newUserAccess := userAccess{}
			newUserAccess.Mail = data[0]
			if !common.IsValidEmail(newUserAccess.Mail) {
				lineOk = false
			}
			orgId, errOrg := strconv.Atoi(data[1])
			if errOrg != nil {
				fmt.Printf("ERROR : org id should be an integer : %s\n", data[1])
				lineOk = false
			}
			newUserAccess.OrgId = orgId
			newUserAccess.WorkspaceName = data[2]
			newUserAccess.Role = data[3]
			if lineOk {
				lstUserAccess = append(lstUserAccess, newUserAccess)
			} else {
				fmt.Printf("ERROR : badly formatted line : %s\n", line)
			}
		} else {
			fmt.Printf("ERROR : badly formatted line (should have 4 columns): %s\n", line)
		}
	}

	if scanner.Err() != nil {
		fmt.Println("Standard input read error")
	}
	usersDf := dataframe.LoadStructs(lstUserAccess)

	workspaces := usersDf.GroupBy("OrgId", "WorkspaceName")
	for group, users := range workspaces.GetGroups() {
		var roles []gristapi.UserRole
		line := strings.Split(group, "_")
		orgId, orgErr := strconv.Atoi(line[0])
		if orgErr != nil {
			Help()
		}
		workspaceId := line[1]
		for i, user := range users.Select([]string{"Mail", "Role"}).Records() {
			if i > 0 {
				newRole := gristapi.UserRole{Email: user[0], Role: user[1]}
				roles = append(roles, newRole)
			}
		}
		gristapi.ImportUsers(orgId, workspaceId, roles)
	}
}

// Displays the list of users witch access to an organization
func DisplayOrgAccess(idOrg string) {

	lstUsers := gristapi.GetOrgAccess(idOrg)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Email", "Name", "Access"})
	for _, user := range lstUsers {
		table.Append([]string{user.Email, user.Name, user.Access})
	}

	table.Render()
}

/*
	Displays detailed information about a document

- Document name
- Number of tables
- For each table :
  - Number of columns
  - Number of rows
  - List of columns
*/
func DisplayDoc(docId string) {
	// Getting the document
	doc := gristapi.GetDoc(docId)

	if doc.Id == "" {
		fmt.Printf("❗️ Document %s not found ❗️\n", docId)
	} else {
		// Document was found

		type TableDetails struct {
			name       string
			nb_rows    int
			nb_cols    int
			cols_names []string
		}

		// Displaying the document name
		titleColor := color.New(color.FgRed).SprintFunc()
		pinned := ""
		if doc.IsPinned {
			pinned = "📌"
		}
		common.DisplayTitle(fmt.Sprintf("Document %s (%s) %s", titleColor(doc.Name), doc.Id, pinned))

		// Getting the doc's tables
		var tables gristapi.Tables = gristapi.GetDocTables(docId)
		fmt.Printf("Contains %d tables :\n", len(tables.Tables))

		// Getting the tables details
		var wg sync.WaitGroup
		var tables_details []TableDetails
		for _, table := range tables.Tables {
			wg.Add(1)
			go func() {
				defer wg.Done()
				table_desc := ""
				columns := gristapi.GetTableColumns(docId, table.Id)
				rows := gristapi.GetTableRows(docId, table.Id)

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

		// Displaying the tables details
		tableView := tablewriter.NewWriter(os.Stdout)
		tableView.SetHeader([]string{"Table", "Nb columns", "Columns", "Rows"})
		for _, table_details := range tables_details {
			for i, col_name := range table_details.cols_names {
				if i == 0 {
					tableView.Append([]string{table_details.name, strconv.Itoa(table_details.nb_cols), col_name, strconv.Itoa(table_details.nb_rows)})
				} else {
					tableView.Append([]string{"", "", col_name, ""})
				}
			}
		}
		tableView.Render()
	}

}

// Displays the list of accessible organizations
func DisplayOrgs() {

	// Getting the list of organizations
	lstOrgs := gristapi.GetOrgs()
	// Sorting the list of organizations by name (lowercase)
	sort.Slice(lstOrgs, func(i, j int) bool {
		return strings.ToLower(lstOrgs[i].Name) < strings.ToLower(lstOrgs[j].Name)
	})
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Name"})
	for _, org := range lstOrgs {
		table.Append([]string{strconv.Itoa(org.Id), org.Name})
	}
	table.Render()
}

// Displays details about an organization
func DisplayOrg(orgId string) {

	type wpDesc struct {
		id     int
		name   string
		nbDoc  int
		nbUser int
	}
	var lstWsDesc []wpDesc

	org := gristapi.GetOrg(orgId)
	if org.Id == 0 {
		fmt.Printf("❗️ Organization %s not found ❗️\n", orgId)
	} else {
		// Org was found
		worskspaces := gristapi.GetOrgWorkspaces(org.Id)
		common.DisplayTitle(fmt.Sprintf("Organization n°%d : %s (%d workspaces)", org.Id, org.Name, len(worskspaces)))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Workspace Id", "Workspace name", "Doc", "Direct users"})
		var wg sync.WaitGroup
		// Retrieving the number of documents and users for each workspace
		for _, ws := range worskspaces {
			func() {
				defer wg.Done()
				wg.Add(1)
				users := gristapi.GetWorkspaceAccess(ws.Id)
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
		// Sorting the list of workspaces by name
		sort.Slice(lstWsDesc, func(i, j int) bool {
			return lstWsDesc[i].name < lstWsDesc[j].name
		})

		// Displaying the list of workspaces
		for _, desc := range lstWsDesc {
			table.Append([]string{strconv.Itoa(desc.id), desc.name, strconv.Itoa(desc.nbDoc), strconv.Itoa(desc.nbUser)})
		}
		table.Render()
	}
}

// Display a Workspace
func DisplayWorkspace(workspaceId int) {

	// Getting the workspace
	ws := gristapi.GetWorkspace(workspaceId)
	if ws.Id == 0 {
		fmt.Printf("❗️ Workspace %d not found ❗️\n", workspaceId)
	} else {
		// Workspace was found
		common.DisplayTitle(fmt.Sprintf("Organization n°%d : \"%s\", workspace n°%d : \"%s\"", ws.Org.Id, ws.Org.Name, ws.Id, ws.Name))

		// Sort the documents by name (lowercase)
		sort.Slice(ws.Docs, func(i, j int) bool {
			return strings.ToLower(ws.Docs[i].Name) < strings.ToLower(ws.Docs[j].Name)
		})

		// Listing the documents
		if len(ws.Docs) > 0 {
			fmt.Printf("Contains %d documents :\n", len(ws.Docs))
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Id", "Name", "Pinned"})
			for _, doc := range ws.Docs {
				pin := ""
				if doc.IsPinned {
					pin = "📌"
				}
				table.Append([]string{doc.Id, doc.Name, pin})
			}
			table.Render()
		} else {
			fmt.Println("No documents")
		}
	}
}

// Displays workspace access rights
func DisplayWorkspaceAccess(workspaceId int) {
	// Getting the workspace
	ws := gristapi.GetWorkspace((workspaceId))
	if ws.Id == 0 {
		fmt.Printf("❗️ Workspace %d not found ❗️\n", workspaceId)
	} else {
		// Workspace was found

		// Displaying the workspace name
		common.DisplayTitle(fmt.Sprintf("Workspace n°%d : %s", ws.Id, ws.Name))

		// Displaying the access rights
		wsa := gristapi.GetWorkspaceAccess(workspaceId)

		// Displaying the MaxInheritedRole
		fmt.Println(TranslateRole(wsa.MaxInheritedRole))

		// Sort users by email (lowercase)
		sort.Slice(wsa.Users, func(i, j int) bool {
			return strings.ToLower(wsa.Users[i].Email) < strings.ToLower(wsa.Users[j].Email)
		})
		nbUsers := len(wsa.Users)
		if nbUsers <= 0 {
			fmt.Println("Accessible to no user")
		} else {
			nbUser := 0
			fmt.Println("\nAccessible to the following users :")
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Id", "Nom", "Email", "Inherited access", "Direct access"})
			for _, user := range wsa.Users {
				if user.Access != "" || user.ParentAccess != "" {
					table.Append([]string{strconv.Itoa(user.Id), user.Name, user.Email, user.ParentAccess, user.Access})
					nbUser += 1
				}
			}
			table.Render()
			fmt.Printf("%d users\n", nbUser)
		}
	}
}

// Displays users with access to a document
func DisplayDocAccess(docId string) {
	// Getting the document
	doc := gristapi.GetDoc(docId)
	if doc.Name == "" {
		fmt.Printf("❗️ Document %s not found ❗️\n", docId)
	} else {
		// Document was found

		// Displaying the document name
		title := fmt.Sprintf("Workspace \"%s\" (n°%d), document \"%s\"", doc.Workspace.Name, doc.Workspace.Id, doc.Name)
		common.DisplayTitle(title)

		// Displaying the access rights
		docAccess := gristapi.GetDocAccess(docId)
		// Sorting users by email (lowercase)
		sort.Slice(docAccess.Users, func(i, j int) bool {
			return strings.ToLower(docAccess.Users[i].Email) < strings.ToLower(docAccess.Users[j].Email)
		})

		fmt.Println(TranslateRole(docAccess.MaxInheritedRole))
		fmt.Printf("\nDirect users:\n")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Id", "Email", "Nom", "Inherited access", "Direct access"})
		for _, user := range docAccess.Users {
			if user.Access != "" {
				table.Append([]string{strconv.Itoa(user.Id), user.Email, user.Name, user.ParentAccess, user.Access})
			}
		}
		table.Render()
	}

}

// Displaying the rights matrix
func DisplayUserMatrix() {
	type userAccess struct {
		Id            int
		Email         string
		Name          string
		OrgId         int
		OrgName       string
		WorkspaceName string
		WokspaceId    int
		ParentAccess  string
		DirectAccess  string
		Access        string
	}
	lstUserAccess := []userAccess{}

	lstOrg := gristapi.GetOrgs()
	for _, org := range lstOrg {
		for _, ws := range gristapi.GetOrgWorkspaces(org.Id) {
			for _, access := range gristapi.GetWorkspaceAccess(ws.Id).Users {
				tmpUserAccess := userAccess{
					Id:            access.Id,
					Email:         access.Email,
					Name:          access.Name,
					OrgId:         org.Id,
					OrgName:       org.Name,
					WorkspaceName: ws.Name,
					WokspaceId:    ws.Id,
					ParentAccess:  access.ParentAccess,
					DirectAccess:  access.Access,
				}
				if access.Access != "" {
					tmpUserAccess.Access = access.Access
				} else {
					if access.ParentAccess != "" {
						tmpUserAccess.Access = access.Access
					}
				}
				if access.Access != "" {
					lstUserAccess = append(lstUserAccess, tmpUserAccess)
				}
			}
		}
	}
	accessDf := dataframe.LoadStructs(lstUserAccess)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Email", "Name", "Org Id", "Org name", "Wokspace id", "Workspace name", "ParentAccess", "DirectAccess", "Access"})
	for email, access := range accessDf.Arrange(dataframe.Sort("Email")).GroupBy("Email").GetGroups() {
		for id, val := range access.Records() {
			if id > 0 {
				line := []string{val[3], email, val[4], val[5], val[6], val[8], val[9], val[7], val[1], val[0]}
				table.Append(line)
			}
		}
	}
	table.Render()
}

// Delete a workspace
func DeleteWorkspace(workspaceId int) {
	if common.Confirm(fmt.Sprintf("Do you really want to delete workspace %d ?", workspaceId)) {
		gristapi.DeleteWorkspace(workspaceId)
	}
}

// Delete a document
func DeleteDoc(docId string) {
	if common.Confirm(fmt.Sprintf("Do you really want to delete document %s ?", docId)) {
		gristapi.DeleteDoc(docId)
	}
}

// Delete a user
func DeleteUser(userId int) {
	if common.Confirm(fmt.Sprintf("Do you really want to delete user %d ?", userId)) {
		gristapi.DeleteUser(userId)
	}
}

// Export a document as a Grist file
func ExportDocGrist(docId string) {
	doc := gristapi.GetDoc(docId)
	if doc.Name != "" {
		gristapi.ExportDocGrist(docId, doc.Workspace.Name+"_"+doc.Name+".grist")
	} else {
		fmt.Printf("❗️ Document %s not found ❗️\n", docId)
	}
}

// Export a document as an Excel file
func ExportDocExcel(docId string) {
	doc := gristapi.GetDoc(docId)
	if doc.Name != "" {
		gristapi.ExportDocExcel(docId, doc.Workspace.Name+"_"+doc.Name+".xlsx")
	} else {
		fmt.Printf("❗️ Document %s not found ❗️\n", docId)
	}
}
