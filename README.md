# GRIST-ctl - Command Line Interface (CLI) for Grist

**[Grist](https://www.getgrist.com/)** is a versatile platform for creating and managing custom data applications. It blends the capabilities of a relational database with the adaptability of a spreadsheet, empowering users to design advanced data workflows, collaborate in real-time, and automate tasks—all without requiring code.

![GRIST logo](grist-logo.png)

**gristctl** is a command-line utility designed for interacting with Grist. It allows users to automate and manage tasks related to Grist documents, including creating, updating, listing, deleting documents, and retrieving data from them.

## Installation

To get started with `gristctl`, follow the steps below to install the tool on your machine.

### Prerequisites

- If you want to build from sources, ensure you have a [working installation of Go](https://golang.org/doc/install) (version 1.23 or later).
- You should also have access to a Grist instance.

### Installing from exec files

Download exec files from [release](https://github.com/Ville-Eurometropole-Strasbourg/gristctl/releases). Extract the archive and copy the `gristctl` file corresponding to your runtime environment into a directory in your PATH.

### Installing from Source

To install `gristctl` from source:

1. Clone the repository:

    ```bash
    git clone https://github.com/Ville-Eurometropole-Strasbourg/gristctl.git
    ```

2. Navigate to the `gristctl` directory:

    ```bash
    cd gristctl
    ```

3. Build the tool:

    ```bash
    go build
    ```

4. Once the build completes, you can move the binary (`gristctl`) to a directory included in your `PATH`, for example:

    ```bash
    sudo mv gristctl /usr/local/bin/
    ```

### Configuring

#### Interactively

You can configure `gristctl` with the following command :

```bash
gristctl config
```

#### Manually
Create a `.gristctl` file in your home directory containing the following information:

```ini
GRIST_TOKEN="user session token"
GRIST_URL="https://<GRIST server URL, without /api>"
```


## Usage

List of commands :

- `config`: configure gristctl
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

### Configure grist connexion

```bash
$ gristctl config
----------------------------------------------------------------------------------
Setting the url and token for access to the grist server (/Users/me/.gristctl)
----------------------------------------------------------------------------------
Actual URL : https://wpgrist.cus.fr
Token : ✅
Would you like to configure (Y/N) ?
y
Grist server URL (https://......... without '/' in the end): https://grist.numerique.gouv.fr 
User token : secrettoken
Url : https://grist.numerique.gouv.fr --- Token: secrettoken
Is it OK (Y/N) ? y
Config saved in /Users/me/.gristctl
```

### List Grist organization

To list all available Grist organization:

```bash
gristctl get org
```

### Displays information about an organization

Example : get organization n°3 information, including the list of his workspaces :

```bash
gristctl get org 3
```

### Fetch Data from a workspace

To fetch data from a Grist workspace with ID 1234, including the list of his documents:

```bash
gristctl get workspace 1234
```

### Delete a workspace

To delete a Grist workspace with ID 1234:

```bash
gristctl delete workspace 1234
```

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

## Contributing

We welcome contributions to gristctl. If you find a bug or want to improve the tool, feel free to open an issue or submit a pull request.

### Steps for contributing

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Commit your changes.
4. Push your branch and create a pull request.

Please ensure that your code adheres to the project's coding style and includes tests where applicable.

## License

This project is licensed under the MIT License
