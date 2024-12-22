package gristtools

import (
	"bufio"
	"fmt"
	"gristctl/common"
	"gristctl/gristapi"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/go-gota/gota/dataframe"
	"github.com/olekukonko/tablewriter"
)

func Help() {
	common.DisplayTitle("GRIST : API querying")
	fmt.Println(`Accepted orders :
- config : configure url & token of Grist server
- get org : organization list
- get org <id> : organization details
- get doc <id> : document details
- get doc <id> access : list of document access rights
- get doc <id> grist : export document as a Grist file (Sqlite) in stdout
- get doc <id> excel : export document as an Excel file (xlsx) in stdout
- purge doc <id> [<number of states to keep>]: purges document history (retains last 3 operations by default)
- get workspace <id>: workspace details
- get workspace <id> access: list of workspace access rights
- delete workspace <id> : delete a workspace
- delete user <id> : delete a user
- import users : imports users from standard input
- get users : displays all user rights`)
	os.Exit(0)
}

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
		fmt.Print("Grist server URL (https://......... without '/' in the end): ")
		var url string
		fmt.Scanln(&url)
		fmt.Print("User token : ")
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

			fmt.Printf("Config saved in %s\n", configFile)
		default:
			os.Exit(0)
		}
	default:
		fmt.Println("On ne fait rien...")
	}
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

func ImportUsers() {
	// Import users from CSV file from stdin
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
				newRole := gristapi.UserRole{user[0], user[1]}
				roles = append(roles, newRole)
			}
		}
		gristapi.ImportUsers(orgId, workspaceId, roles)
	}
}

func DisplayOrgAccess(idOrg string) {
	// Displays the list of users with access to an organization

	lstUsers := gristapi.GetOrgAccess(idOrg)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Email", "Name", "Access"})
	for _, user := range lstUsers {
		table.Append([]string{user.Email, user.Name, user.Access})
	}

	table.Render()
}

func DisplayDoc(docId string) {
	// Displays detailed information about a document
	// - Document name
	// - Number of tables
	// For each table :
	// - Number of columns
	// - Number of rows
	// - List of columns

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

func DisplayTableRecords(docId string, tableName string) {
	fmt.Printf("Doc %s - table %s\n", docId, tableName)
	df := gristapi.GetTableContent(docId, tableName)

	if df.Err != nil {
		fmt.Println("Error:", df.Err)
		return
	}
}

func DisplayOrgs() {
	// Displays the list of accessible organizations

	lstOrgs := gristapi.GetOrgs()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Name"})
	for _, org := range lstOrgs {
		table.Append([]string{strconv.Itoa(org.Id), org.Name})
	}
	table.Render()
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

func DisplayWorkspace(workspaceId int) {
	// Affiche des détails d'un Workspace

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

func DisplayWorkspaceAccess(workspaceId int) {
	// Displays workspace access rights

	ws := gristapi.GetWorkspace((workspaceId))
	common.DisplayTitle(fmt.Sprintf("Workspace n°%d access rights : %s", ws.Id, ws.Name))
	wsa := gristapi.GetWorkspaceAccess(workspaceId)
	displayRole(wsa.MaxInheritedRole)

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

func DisplayDocAccess(docId string) {
	// Displays users with access to a document

	doc := gristapi.GetDoc(docId)
	title := fmt.Sprintf("Workspace \"%s\" (n°%d), document \"%s\"", doc.Workspace.Name, doc.Workspace.Id, doc.Name)
	common.DisplayTitle(title)

	docAccess := gristapi.GetDocAccess(docId)
	displayRole(docAccess.MaxInheritedRole)
	fmt.Printf("\nDirect users:\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Emai", "Nom", "Inherited access", "Direct access"})
	for _, user := range docAccess.Users {
		if user.Access != "" {
			table.Append([]string{strconv.Itoa(user.Id), user.Email, user.Name, user.ParentAccess, user.Access})
		}
	}
	table.Render()
}

func DisplayUserMatrix() {
	// Displaying the rights matrix

	lstOrg := gristapi.GetOrgs()
	for _, org := range lstOrg {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Id", "Name", "Email", "Access", "ParentAccess", "Workspace name", "Wokspace id"})
		common.DisplayTitle(fmt.Sprintf("Org \"%s\" (%d)", org.Name, org.Id))
		for _, ws := range gristapi.GetOrgWorkspaces(org.Id) {
			users := dataframe.LoadStructs(gristapi.GetWorkspaceAccess(ws.Id).Users).Arrange(dataframe.Sort("Email"))
			for line := 0; line < users.Nrow(); line++ {
				ok := false
				row := make([]string, users.Ncol()+2)
				for colId, colName := range users.Names() {
					value := fmt.Sprintf("%v", users.Elem(line, colId))
					row[colId] = value

					// Only keep accessfull lines
					if strings.Contains(colName, "Access") {
						if value != "" {
							ok = true
						}
					}
				}
				row[users.Ncol()] = ws.Name
				row[users.Ncol()+1] = strconv.Itoa(ws.Id)
				if ok {
					table.Append(row)
				}
			}
		}
		table.Render()
	}
}
