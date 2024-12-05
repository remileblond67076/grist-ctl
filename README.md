# GRIST-cli : exploitation de l'API GRIST

## Configuration

Mettre en place un fichier `.end` contenant les informations suivantes :

```ini
GRIST_TOKEN="clef de session"
GRIST_URL="https://<url du serveur GRIST avant /api>"
```

## Usage

Liste des commandes utilisables :

- `grist-cli get org` : liste des organisations
- `grist-cli get org <id>` : détails d'une organisation
- `grist-cli get doc <id>` : détails d'un document
- `grist-cli get doc <id> access` : liste des droits d'accès au document
- `grist-cli purge doc <id> [<nombre d'états à conserver>]`: purge l'historique d'un document (conserve les 3 dernières opérations par défaut)
- `grist-cli get workspace <id>`: détails sur un workspace
- `grist-cli get workspace <id> access`: liste des droits d'accès à un workspace
- `gristctl delete workspace <id>` : suppression d'un workspace

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
