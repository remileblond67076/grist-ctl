# GRIST-ctl : GRIST API tool

Command-line tool for managing a [Grist](https://www.getgrist.com/) instance.
Based on [Grist's REST APIs](https://support.getgrist.com/api/).

## Usage

List of commands :

- `get org` : organization list
- `get org <id>` : organization details
- `get doc <id>` : document details
- `get doc <id> access` : list of document access rights
- `purge doc <id> [<number of states to keep>]`: purges document history (retains last 3 operations by default)
- `get workspace <id>`: workspace details
- `get workspace <id> access`: list of workspace access rights
- `delete workspace <id>` : delete a workspace
- `delete user <id>` : delete a user
- `import users` : imports users from standard input
- `get users` : displays all user rights

## Configuration

Set up a `$HOME/.gristctl` file containing the following information:

```ini
GRIST_TOKEN="user session token"
GRIST_URL="https://<GRIST server URL, without /api>"
```

## Examples of usages

### Import users from an ActiveDirectory directory

Extract the list of members of AD groups GA_GRIST_PU and GA_GRIST_PA :

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

## Integration

### Unit testing

Launch unit tests with the following command :

```bash
go test -race $(go list ./... | grep -v /vendor/)
```

### Compilation

To build binaries for the development platform, use the following command:

```bash
go build .
```

To build binaries for Windows :

```bash
GOOS=windows GOARCH=amd64 go build .
```
