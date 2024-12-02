package main

import (
	"fmt"
	"log"
	"os"

	"grist-cli/gristapi"
)

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		log.Fatal("Merci de passer des arguments")
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
							{
								fmt.Println("Affichage de toutes les orgs accessibles")
								gristapi.DisplayOrgs()
							}
						case 3:
							orgId := args[2]
							fmt.Printf("Affichage des détails de l'org n°%s\n", orgId)
							gristapi.DisplayOrg(orgId)
						}
					}
				case "doc":
					{
						switch len(args) {
						case 3:
							{
								docId := args[2]
								fmt.Printf("Affichage du document %s", docId)
								gristapi.DisplayDoc(docId)
							}
						case 4:
							{
								if args[3] == "access" {
									docId := args[2]
									gristapi.DisplayDocAccess(docId)
								}

							}
						}
					}
				default:
					log.Fatal("Je ne comprend pas votre demande")
				}
			}
		}
	case "purge":
		{
			if len(args) > 2 {
				switch arg2 := args[1]; arg2 {
				case "doc":
					{
						if len(os.Args) == 4 {
							docId := os.Args[3]
							fmt.Printf("Purge du document %s\n", docId)
							gristapi.PurgeDoc(docId)
						}
					}
				}
			}
		}
	default:
		log.Fatal("Argument non compris")
	}

}
