# GRIST-cli : exploitation de l'API GRIST

## Configuration

Mettre en place un fichier `.end` contenant les informations suivantes :

```ini
GRIST_TOKEN="clef de session"
GRIST_URL="https://<url du serveur GRIST avant /api>"
```

## Usage

### Général

Liste des commandes utilisables :

- `get org` : liste des organisations
- `get org <id>` : détails d'une organisation
- `get doc <id>` : détails d'un document
- `get doc <id> access` : liste des droits d'accès au document
- `purge doc <id> [<nombre d'états à conserver>]`: purge l'historique d'un document (conserve les 3 dernières opérations par défaut)
- `get workspace <id>`: détails sur un workspace
- `get workspace <id> access`: liste des droits d'accès à un workspace
- `delete workspace <id>` : suppression d'un workspace
- `import users` : importe des utilisateurs dont la liste est envoyée sur l'entrée standard
- `get users` : affiche l'ensemble des droits utilisateurs

### Import des utilisateurs depuis un annuaire ActiveDirectory

Extraction de la liste des membres des groupes AD GA_GRIST_PU et GA_GRIST_PA :

```powershell
foreach ($grp in ('a', 'u')) {
    get-adgroupmember ga_grist_p$grp | get-aduser -properties mail, extensionAttribute6, extensionAttribute15 |select-object mail, extensionAttribute6, extensionAttribute15 | export-csv -Path ga_grist_p$grp.csv -NoTypeInformation -Encoding:UTF8
}
```

```bash
cat ga_grist_pu.csv | awk -F',' 'NR>1 {gsub(/"/, "", $0); print tolower($1)";3;Direction-"$2";viewers"}' | ./gristctl import users
cat ga_grist_pu.csv | awk -F',' 'NR>1 {gsub(/"/, "", $0); print tolower($1)";3;Service-"$3";viewers"}' | ./gristctl import users
cat ga_grist_pa.csv | awk -F',' 'NR>1 {gsub(/"/, "", $0); print tolower($1)";3;Direction-"$2";editors"}' | ./gristctl import users
cat ga_grist_pa.csv | awk -F',' 'NR>1 {gsub(/"/, "", $0); print tolower($1)";3;Service-"$3";editors"}' | ./gristctl import users
```

## Intégration

### Tests unitaires

Lancement des tests unitaires à l'aide de la commande suivante :

```bash
go test .
```

### Compilation

Pour construire les binaires pour la plateforme de développement, utiliser la commande suivante :

```bash
go build .
```

Pour construire les binaires pour Windows :

```bash
GOOS=windows GOARCH=amd64 go build .
```
