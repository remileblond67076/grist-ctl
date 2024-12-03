package main

import (
	"fmt"
	"grist-cli/gristapi"
	"testing"
)

func TestConnect(t *testing.T) {
	orgs := gristapi.GetOrgs()
	nbOrgs := len(orgs)

	if nbOrgs < 2 {
		t.Errorf("Je ne trouve que %d organisations", nbOrgs)
	}

	for i, org := range orgs {
		orgId := fmt.Sprintf("%d", org.Id)
		if gristapi.GetOrg(orgId).Name != orgs[i].Name {
			t.Error("Ne trouve pas l'organisation principale")
		}

		workspaces := gristapi.GetOrgWorkspaces(org.Id)

		if len(workspaces) < 1 {
			t.Errorf("Aucun Workspace dans l'organisation %d", org.Id)
		}

		for i, workspace := range workspaces {
			if workspace.OrgDomain != org.Domain {
				t.Errorf("Workspace %d : le domaine du workspace %s ne correspond pas à %s", workspace.Id, workspace.OrgDomain, org.Domain)
			}

			myWorkspace := gristapi.GetWorkspace(workspace.Id)
			if myWorkspace.Name != workspace.Name {
				t.Errorf("Workspace n°%d : les noms ne correspondent pas (%s/%s)", workspace.Id, workspace.Name, myWorkspace.Name)
			}

			if workspace.Name != workspaces[i].Name {
				t.Error("Mauvaise correspondance des noms de Workspaces")
			}

			for i, doc := range workspace.Docs {
				if doc.Name != workspace.Docs[i].Name {
					t.Errorf("Document n°%s : non correspondance des noms (%s/%s)", doc.Id, doc.Name, workspace.Docs[i].Name)
				}

				// Un document doit avoir au moins une table
				tables := gristapi.GetDocTables(doc.Id)
				if len(tables.Tables) < 1 {
					t.Errorf("Le document n°%s ne contient pas de table", doc.Name)
				}
				for _, table := range tables.Tables {
					// Une table doit avoir au moins une colonne
					cols := gristapi.GetTableColumns(doc.Id, table.Id)
					if len(cols.Columns) < 1 {
						t.Errorf("La table %s du document %s ne contient pas de colonne", table.Id, doc.Id)
					}
				}
			}

		}
	}

}
