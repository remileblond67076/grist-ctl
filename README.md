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