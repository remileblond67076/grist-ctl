package gristtools

import (
	"bufio"
	"fmt"
	"gristctl/common"
	"gristctl/gristapi"
	"os"
	"strconv"
	"strings"

	"github.com/go-gota/gota/dataframe"
)

func Help() {
	common.DisplayTitle("GRIST : API querying")
	fmt.Println(`Accepted orders :
- config : configure url & token of Grist server
- get org : organization list
- get org <id> : organization details
- get doc <id> : document details
- get doc <id> access : list of document access rights
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
