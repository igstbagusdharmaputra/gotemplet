package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/spf13/cobra"
	cryptossh "golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

var (
	templatePath, subTemplatePath, dataPath, outputPath string
	branch, gitUser, gitPass, sshKeyPath                string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gotemplet",
	Short: "Template Generator CLI",
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Clone a Git repository, process templates, and rename files/directories",
	RunE:  runGenerate,
}

func init() {
	generateCmd.Flags().StringVarP(&templatePath, "templatePath", "t", "", "Git repository URL or local directory")
	generateCmd.Flags().StringVarP(&subTemplatePath, "subTemplatePath", "s", "", "Sub-template path")
	generateCmd.Flags().StringVarP(&dataPath, "dataPath", "d", "", "Path to JSON/YAML data file")
	generateCmd.Flags().StringVarP(&outputPath, "outputPath", "o", "", "Output directory")
	generateCmd.Flags().StringVarP(&branch, "branch", "b", "main", "Git branch to clone")
	generateCmd.Flags().StringVarP(&gitUser, "gitUser", "u", os.Getenv("GIT_USER"), "Git username for HTTP Basic Auth")
	generateCmd.Flags().StringVarP(&gitPass, "gitPass", "p", os.Getenv("GIT_PASS"), "Git password for HTTP Basic Auth")
	generateCmd.Flags().StringVarP(&sshKeyPath, "sshKeyPath", "k", "", "Path to SSH private key for authentication")

	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	if isGitURL(templatePath) {
		fmt.Println("Cloning repository...")
		clonedPath, err := cloneRepo(templatePath, branch)
		if err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
		templatePath = clonedPath
	}

	return processTemplates()
}

func isGitURL(url string) bool {
	return strings.HasPrefix(url, "http") || strings.HasPrefix(url, "ssh://") || strings.HasPrefix(url, "git@")
}

func cloneRepo(repoURL, branch string) (string, error) {
	dir := filepath.Join(os.TempDir(), "tmpl_repo")
	_ = os.RemoveAll(dir)

	auth, err := getGitAuth(repoURL)
	if err != nil {
		return "", err
	}

	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL:           repoURL,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Auth:          auth,
		SingleBranch:  true,
		Depth:         1,
	})

	if err != nil {
		return "", fmt.Errorf("git clone failed: %w", err)
	}

	return dir, nil
}

func getGitAuth(repoURL string) (transport.AuthMethod, error) {
	if strings.HasPrefix(repoURL, "http") {
		if gitUser != "" && gitPass != "" {
			return &http.BasicAuth{Username: gitUser, Password: gitPass}, nil
		}
	} else if strings.HasPrefix(repoURL, "ssh") {
		if sshKeyPath != "" {
			return loadSSHKey(sshKeyPath)
		}
		return ssh.NewSSHAgentAuth("git")
	}
	return nil, nil
}

func loadSSHKey(sshKeyPath string) (transport.AuthMethod, error) {
	key, err := os.ReadFile(sshKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSH key: %w", err)
	}

	signer, err := cryptossh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH key: %w", err)
	}

	return &ssh.PublicKeys{User: "git", Signer: signer}, nil
}

func loadData() (map[string]interface{}, error) {
	if dataPath == "" {
		return nil, nil
	}

	b, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file: %w", err)
	}

	var data map[string]interface{}
	switch {
	case strings.HasSuffix(dataPath, ".yaml"), strings.HasSuffix(dataPath, ".yml"):
		err = yaml.Unmarshal(b, &data)
	case strings.HasSuffix(dataPath, ".json"):
		err = json.Unmarshal(b, &data)
	default:
		return nil, errors.New("unsupported data format")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse data file: %w", err)
	}

	return data, nil
}

// Function map for template processing
var funcMap = template.FuncMap{
	"env": func(key, def string) string {
		if val, exists := os.LookupEnv(key); exists {
			return val
		}
		return def
	},
}

func processTemplates() error {
	data, err := loadData()
	if err != nil {
		return err
	}

	templateDir := filepath.Join(templatePath, subTemplatePath)
	if subTemplatePath == "" {
		templateDir = templatePath
	}

	return filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		newPath, err := processPath(path, templateDir, outputPath, data)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return os.MkdirAll(newPath, os.ModePerm)
		}

		return renderFile(path, newPath, data)
	})
}

func processPath(oldPath, templateDir, outputDir string, data map[string]interface{}) (string, error) {
	relPath, _ := filepath.Rel(templateDir, oldPath)

	var buf bytes.Buffer
	tmpl, err := template.New("path").Funcs(funcMap).Parse(relPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse path template: %w", err)
	}

	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute path template: %w", err)
	}

	newPath := filepath.Join(outputDir, buf.String())
	return newPath, nil
}

func renderFile(src, dst string, data map[string]interface{}) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	tmpl, err := template.New(filepath.Base(src)).Funcs(funcMap).Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse file template: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}
