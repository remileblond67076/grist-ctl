package gristtools

import (
	"bufio"
	"fmt"
	"gristctl/common"
	"os"
	"strconv"
	"strings"

	"github.com/go-gota/gota/dataframe"
)

func Help() {
	common.DisplayTitle("GRIST : API querying")
	fmt.Println(`Accepted orders :
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
		line := strings.Split(group, "_")
		orgId, orgErr := strconv.Atoi(line[0])
		if orgErr != nil {
			Help()
		}
		workspaceId := line[1]
		fmt.Printf("Org: %d, Workspace : %s\n", orgId, workspaceId)
		for i, user := range users.Select([]string{"Mail", "Role"}).Records() {
			if i > 0 {
				fmt.Printf("  - %s --> %s\n", user[0], user[1])
			}
		}
	}
	// var wg sync.WaitGroup
	// func() {
	// wg.Add(1)
	// 	defer wg.Done()
	// 	gristapi.ImportUser(
	// 		email,
	// 		orgId,
	// 		workspaceName,
	// 		role,
	// 	)
	// }()
	// wg.Wait()
}
