package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"gristctl/common"
	"gristctl/gristapi"
)

func help() {
	common.DisplayTitle("GRIST : interrogation des API")
	fmt.Println(`Commandes acceptées :
- get org : liste des organisations
- get org <id> : détails d'une organisation
- get doc <id> : détails d'un document
- get doc <id> access : liste des droits d'accès au document
- purge doc <id> [<nombre d'états à conserver>]: purge l'historique d'un document (conserve les 3 dernières opérations par défaut)
- get workspace <id>: détails sur un workspace
- get workspace <id> access: liste des droits d'accès à un workspace
- delete workspace <id> : suppression d'un workspace
- delete user <id> : suppression d'un utilisateur
- import users : importe des utilisateurs dont la liste est envoyée sur l'entrée standard
- get users : affiche l'ensemble des droits utilisateurs`)
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
				case "users":
					gristapi.DisplayUserMatrix()
				default:
					help()
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
							help()
						}
					}
					gristapi.PurgeDoc(docId, nbHisto)
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
						idWorkspace, err := strconv.Atoi(args[2])
						if err == nil {
							gristapi.DeleteWorkspace(idWorkspace)
						}
					} else {
						help()
					}
				case "user":
					if len(args) == 3 {
						idUser, err := strconv.Atoi(args[2])
						if err == nil {
							gristapi.DeleteUser(idUser)
						}
					} else {
						help()
					}
				case "doc":
					if len(args) == 3 {
						docId := args[2]
						gristapi.DeleteDoc(docId)
					}
				default:
					help()
				}
			}
		}
	case "import":
		if len(args) > 1 {
			switch args[1] {
			case "users":
				common.DisplayTitle("Import des utilisateurs")
				fmt.Println("Format des données attendues en entrée standard : <mail>;<org id>;<workspace name>;<role>")

				// Lecture des données en entrée standard
				scanner := bufio.NewScanner(os.Stdin)
				var wg sync.WaitGroup
				for scanner.Scan() {
					line := scanner.Text()
					data := strings.Split(line, ";")
					if len(data) == 4 {
						email := data[0]
						orgId, errOrg := strconv.Atoi(data[1])
						if errOrg != nil {
							fmt.Printf("ERREUR : l'id d'organisation devrait être un entier (%s)\n", data[1])
						}
						workspaceName := data[2]

						role := data[3]
						wg.Add(1)
						func() {
							defer wg.Done()
							gristapi.ImportUser(
								email,
								orgId,
								workspaceName,
								role,
							)
						}()
					} else {
						fmt.Printf("Ligne mal formatée : %s", line)
					}
				}
				wg.Wait()

				if scanner.Err() != nil {
					// Handle error.
				}
			default:
				help()
			}
		}
	default:
		help()
	}

}
