package main

import (
	"fmt"
	"os"
	"strconv"

	"gristctl/gristapi"
)

func help() {
	fmt.Println(`-----------------------------
GRIST : interrogation des API
-----------------------------
Commandes acceptées :
- gristctl get org : liste des organisations
- gristctl get org <id> : détails d'une organisation
- gristctl get doc <id> : détails d'un document
- gristctl get doc <id> access : liste des droits d'accès au document
- gristctl purge doc <id>: purge l'historique d'un document (conserve les 3 dernières opérations)
- gristctl get workspace <id>: détails sur un workspace
- gristctl get workspace <id> access: liste des droits d'accès à un workspace
- gristctl delete workspace <id> : suppression d'un workspace`)
	os.Exit(0)
}

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		help()
	}

	switch arg1 := args[0]; arg1 {
	case "get":
		{
			if len(args) > 1 {
				switch arg2 := args[1]; arg2 {
				case "org":
					{
						switch nb := len(args); nb {
						case 2:
							gristapi.DisplayOrgs()
						case 3:
							orgId := args[2]
							gristapi.DisplayOrg(orgId)
						case 4:
							switch args[3] {
							case "access":
								orgId := args[2]
								gristapi.DisplayOrgAccess(orgId)
							default:
								help()
							}
						default:
							help()
						}
					}
				case "doc":
					{
						switch len(args) {
						case 3:
							docId := args[2]
							fmt.Printf("Affichage du document %s", docId)
							gristapi.DisplayDoc(docId)
						case 4:
							if args[3] == "access" {
								docId := args[2]
								gristapi.DisplayDocAccess(docId)
							}
						default:
							help()
						}
					}
				case "workspace":
					{
						switch len(args) {
						case 3:
							worskspaceId, err := strconv.Atoi(args[2])
							if err == nil {
								gristapi.DisplayWorkspace(worskspaceId)
							}
						case 4:
							if args[3] == "access" {
								worskspaceId, err := strconv.Atoi(args[2])
								if err == nil {
									gristapi.DisplayWorkspaceAccess(worskspaceId)
								}
							}
						default:
							help()
						}
					}
				default:
					help()
				}
			}
		}
	case "purge":
		{
			if len(args) > 2 {
				switch arg2 := args[1]; arg2 {
				case "doc":
					if len(os.Args) == 4 {
						docId := os.Args[3]
						fmt.Printf("Purge du document %s\n", docId)
						gristapi.PurgeDoc(docId)
					}
				default:
					help()
				}
			}
		}
	case "delete":
		{
			if len(args) > 2 {
				switch arg2 := args[1]; arg2 {
				case "workspace":
					if len(args) == 3 {
						arg3, err := strconv.Atoi(args[2])
						if err == nil {
							gristapi.DeleteWorkspace(arg3)
						}
					} else {
						help()
					}
				default:
					help()
				}
			}
		}
	default:
		help()
	}

}
