# GRIST-cli : exploitation de l'API GRIST

## Configuration

Mettre en place un fichier `.end` contenant les informations suivantes :

```
GRIST_TOKEN="clef de session"
GRIST_URL="https://wpgrist.cus.fr"
```

## Usage

Liste des commandes utilisables :

- `grist-cli get org` : liste des organisations
- `grist-cli get org <id>` : détails d'une organisation
- `grist-cli get doc <id>` : détails d'un document
- `grist-cli get doc <id> access` : liste des droits d'accès au document
- `grist-cli purge doc <id>`: purge l'historique d'un document (conserve les 3 dernières opérations)
- `grist-cli get workspace <id>`: détails sur un workspace
- `grist-cli get workspace <id> access`: liste des droits d'accès à un workspace