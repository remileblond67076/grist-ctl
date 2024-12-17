# GRIST-ctl - Command Line Interface for Grist

**[Grist](https://www.getgrist.com/)** is a flexible platform for building and managing custom data applications. It combines the power of a relational database with the flexibility of a spreadsheet, allowing users to build sophisticated data workflows, collaborate in real-time, and automate processes—all in a no-code environment. Whether you're building a CRM, project management tool, or custom data tracker, Grist provides a powerful interface for data management.

![GRIST logo](grist-logo.png)

**gristctl** is a command-line tool for interacting with Grist. It enables users to automate and manage various tasks related to Grist documents, such as creating, updating, listing, and deleting documents, as well as fetching data from them.

## Installation

To get started with `gristctl`, follow the steps below to install the tool on your machine.

### Prerequisites

- Ensure you have a [working installation of Go](https://golang.org/doc/install) (version 1.16 or later).
- You should also have access to a Grist instance.

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

Create a `.gristctl` file in your home directory containing the following information:

```ini
GRIST_TOKEN="user session token"
GRIST_URL="https://<GRIST server URL, without /api>"
```

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

### Steps for contributing:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Commit your changes.
4. Push your branch and create a pull request.

Please ensure that your code adheres to the project's coding style and includes tests where applicable.

## License

This project is licensed under the MIT License.