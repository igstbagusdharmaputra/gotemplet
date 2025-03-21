# Template Generator CLI (`gotemplet`)

## Overview
`gotemplet` is a CLI tool that clones a Git repository, processes template files, and renames files/directories based on template expressions. It also supports environment variables in templates.

## Installation
To install, clone this repository and build the executable:

```sh
git clone <repository-url>
cd gotemplet
go build -o gotemplet
```

## Usage

### Generate Template-Based Files
```sh
./gotemplet generate --templatePath <repo-url-or-local-dir> \
                     --subTemplatePath <sub-dir> \
                     --dataPath <data-file> \
                     --outputPath <output-dir> \
                     --branch <branch-name> \
                     --gitUser <username> \
                     --gitPass <password> \
                     --sshKeyPath <ssh-key>
```

### Options
- `--templatePath` (`-t`): Git repository URL or local template directory.
- `--subTemplatePath` (`-s`): Subdirectory within the template repository.
- `--dataPath` (`-d`): Path to a JSON/YAML file containing data for template rendering.
- `--outputPath` (`-o`): Output directory where processed templates are saved.
- `--branch` (`-b`): The Git branch to clone (default: `main`).
- `--gitUser` (`-u`): Git username for HTTP authentication.
- `--gitPass` (`-p`): Git password for HTTP authentication.
- `--sshKeyPath` (`-k`): Path to an SSH private key for authentication.

## Using Environment Variables in Templates
You can use environment variables inside your templates like this:

```yaml
app_name: "{{ env `APP_NAME` `default-app` }}"
port: "{{ env `APP_PORT` `8080` }}"
```

- If `APP_NAME` is set in the environment, it will use that value.
- If `APP_NAME` is not set, it will default to `"default-app"`.
- If `APP_PORT` is not set, it will default to `8080`.

### Example Usage
#### Using a Local Template Directory
```sh
./gotemplet generate -t ./templates -d data.yaml -o ./output
```

#### Cloning a Git Repository and Processing Templates
```sh
./gotemplet generate -t git@github.com:user/repo.git -d data.json -o ./output -b develop
```

#### Using SSH Authentication
```sh
./gotemplet generate -t git@github.com:user/repo.git -d data.yaml -o ./output -k ~/.ssh/id_rsa
```

## Handling Missing Environment Variables
If an environment variable is missing, the program will:
1. Use the default value if provided.
2. Return an error if no default is provided.

Example:
```yaml
database_url: "{{ env `DATABASE_URL` }}"
```
If `DATABASE_URL` is not set, the program will return an error.

## License
This project is licensed under the MIT License.

