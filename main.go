package main

import (
	"os"
	"strconv"

	"gristctl/gristapi"
	"gristctl/gristtools"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		gristtools.Help()
	}

	switch arg1 := args[0]; arg1 {
	case "config":
		gristtools.Config()
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
								gristtools.Help()
							}
						default:
							gristtools.Help()
						}
					}
				case "doc":
					{
						switch len(args) {
						case 3:
							docId := args[2]
							gristapi.DisplayDoc(docId)
						case 4:
							if args[3] == "access" {
								docId := args[2]
								gristapi.DisplayDocAccess(docId)
							}
						default:
							gristtools.Help()
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
							gristtools.Help()
						}
					}
				case "users":
					gristapi.DisplayUserMatrix()
				default:
					gristtools.Help()
				}
			}
		}
	case "purge":
		{
			if len(args) > 2 {
				switch args[1] {
				case "doc":
					docId := args[2]
					nbHisto := 3
					if len(args) == 4 {
						nb, err := strconv.Atoi(args[3])
						if err == nil {
							nbHisto = nb
						} else {
							gristtools.Help()
						}
					}
					gristapi.PurgeDoc(docId, nbHisto)
				default:
					gristtools.Help()
				}
			}
		}
	case "delete":
		{
			if len(args) > 2 {
				switch arg2 := args[1]; arg2 {
				case "workspace":
					if len(args) == 3 {
						idWorkspace, err := strconv.Atoi(args[2])
						if err == nil {
							gristapi.DeleteWorkspace(idWorkspace)
						}
					} else {
						gristtools.Help()
					}
				case "user":
					if len(args) == 3 {
						idUser, err := strconv.Atoi(args[2])
						if err == nil {
							gristapi.DeleteUser(idUser)
						}
					} else {
						gristtools.Help()
					}
				case "doc":
					if len(args) == 3 {
						docId := args[2]
						gristapi.DeleteDoc(docId)
					}
				default:
					gristtools.Help()
				}
			}
		}
	case "import":
		if len(args) > 1 {
			switch args[1] {
			case "users":
				gristtools.ImportUsers()
			default:
				gristtools.Help()
			}
		}
	default:
		gristtools.Help()
	}

}
