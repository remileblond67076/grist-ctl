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
	fmt.Println(`Accepted orders :
- config : configure url & token of Grist server
- get org : organization list
- get org <id> : organization details
- get workspace <id>: workspace details
- get workspace <id> access: list of workspace access rights
- get doc <id> : document details
- get doc <id> access : list of document access rights
- get doc <id> grist : export document as a Grist file (Sqlite) in stdout
- get doc <id> excel : export document as an Excel file (xlsx) in stdout
- get doc <id> table <tableName> : export content of a document's table as a CSV file (xlsx) in stdout
- get users : displays all user rights
- import users : imports users from standard input
- purge doc <id> [<number of states to keep>]: purges document history (retains last 3 operations by default)
- delete doc <id> : delete a document
- delete user <id> : delete a user
- delete workspace <id> : delete a workspace
- version : displays the version of the program`)
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
	token := "✅"
	if os.Getenv("GRIST_TOKEN") == "" {
		token = "❌"
	}
	fmt.Printf("Token : %s\n", token)
	fmt.Println("Would you like to configure (Y/N) ?")
	var goConfig string
	fmt.Scanln(&goConfig)

	switch response := strings.ToLower(goConfig); response {
	case "y":
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
		fmt.Printf("Url : %s --- Token: %s\nIs it OK (Y/N) ? ", url, token)
		var ok string
		fmt.Scanln(&ok)
		switch strings.ToLower(ok) {
		case "y":
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

		default:
			os.Exit(0)
		}
	default:
		fmt.Println("Keeping all in place...")
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
			newUserAccess := userAccess{}
			newUserAccess.Mail = data[0]
			orgId, errOrg := strconv.Atoi(data[1])
			if errOrg != nil {
				fmt.Printf("ERROR : org id should be an integer : %s\n", data[1])
			}
			newUserAccess.OrgId = orgId
			newUserAccess.WorkspaceName = data[2]
			newUserAccess.Role = data[3]
			lstUserAccess = append(lstUserAccess, newUserAccess)
		} else {
			fmt.Printf("Badly formatted line : %s", line)
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
	doc := gristapi.GetDoc(docId)

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

	var tables gristapi.Tables = gristapi.GetDocTables(docId)
	fmt.Printf("Contains %d tables\n", len(tables.Tables))
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
	var details []string
	for _, table_details := range tables_details {
		ligne := fmt.Sprintf("- %s : %d lines, %d columns\n", title(table_details.name), table_details.nb_rows, table_details.nb_cols)
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

// Displays the list of accessible organizations
func DisplayOrgs() {

	lstOrgs := gristapi.GetOrgs()
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
	worskspaces := gristapi.GetOrgWorkspaces(org.Id)
	common.DisplayTitle(fmt.Sprintf("Organization n°%d : %s (%d workspaces)", org.Id, org.Name, len(worskspaces)))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Workspace Id", "Workspace name", "Doc", "Direct users"})
	var wg sync.WaitGroup
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

	for _, desc := range lstWsDesc {
		table.Append([]string{strconv.Itoa(desc.id), desc.name, strconv.Itoa(desc.nbDoc), strconv.Itoa(desc.nbUser)})
	}
	table.Render()
}

// Affiche des détails d'un Workspace
func DisplayWorkspace(workspaceId int) {

	ws := gristapi.GetWorkspace(workspaceId)
	common.DisplayTitle(fmt.Sprintf("Organization n°%d : \"%s\", workspace n°%d : \"%s\"", ws.Org.Id, ws.Org.Name, ws.Id, ws.Name))

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

// Displays workspace access rights
func DisplayWorkspaceAccess(workspaceId int) {

	ws := gristapi.GetWorkspace((workspaceId))
	common.DisplayTitle(fmt.Sprintf("Workspace n°%d access rights : %s", ws.Id, ws.Name))
	wsa := gristapi.GetWorkspaceAccess(workspaceId)
	fmt.Println(TranslateRole(wsa.MaxInheritedRole))

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

// Displays users with access to a document
func DisplayDocAccess(docId string) {
	doc := gristapi.GetDoc(docId)
	title := fmt.Sprintf("Workspace \"%s\" (n°%d), document \"%s\"", doc.Workspace.Name, doc.Workspace.Id, doc.Name)
	common.DisplayTitle(title)

	docAccess := gristapi.GetDocAccess(docId)
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
