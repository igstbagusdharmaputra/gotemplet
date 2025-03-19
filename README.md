# Template Generator CLI (`gotemplet`)

## Overview
`gotemplet` is a CLI tool that clones a Git repository, processes template files, and renames files/directories based on template expressions.

## Installation
To install, clone this repository and build the executable:

```sh
git clone <repository-url>
cd tmpl
go build -o tmpl
```

## Usage

### Generate Template-Based Files
```sh
./tmpl generate --templatePath <repo-url-or-local-dir> \
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

## Example Usage
### Using a Local Template Directory
```sh
./tmpl generate -t ./templates -d data.yaml -o ./output
```

### Cloning a Git Repository and Processing Templates
```sh
./tmpl generate -t git@github.com:user/repo.git -d data.json -o ./output -b develop
```

### Using SSH Authentication
```sh
./tmpl generate -t git@github.com:user/repo.git -d data.yaml -o ./output -k ~/.ssh/id_rsa
```

## License
This project is licensed under the MIT License.

