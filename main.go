package main

import (
	"fmt"
	"log"
	"os"

	"grist-cli/gristapi"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Erreur lors de la lecture du fichier de configuration\n", err)
	}

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
						if len(args) > 2 {
							docId := args[2]
							fmt.Printf("Affichage du document %s", docId)
							gristapi.DisplayDoc(docId)
						} else {
							log.Fatal("Merci de préciser la référence du document")
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

					}
				}
			}
		}
	default:
		log.Fatal("Argument non compris")
	}

}
