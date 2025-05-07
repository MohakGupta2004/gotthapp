package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [project-name]",
	Short: "Create a new Goth Stack project",
	Long:  `Create a new Goth Stack project with the specified name in the current directory.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		if projectName == "" {
			fmt.Println("Error: Project name cannot be empty")
			os.Exit(1)
		}

		// Get GitHub username
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your GitHub username: ")
		githubUsername, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading GitHub username: %v\n", err)
			os.Exit(1)
		}
		githubUsername = strings.TrimSpace(githubUsername)
		if githubUsername == "" {
			fmt.Println("Error: GitHub username cannot be empty")
			os.Exit(1)
		}

		// Get the current working directory
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		// Create the target directory
		targetDir := filepath.Join(currentDir, projectName)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			fmt.Printf("Error creating target directory: %v\n", err)
			os.Exit(1)
		}

		// Create a temporary directory for cloning
		tempDir, err := os.MkdirTemp("", "goth-template-*")
		if err != nil {
			fmt.Printf("Error creating temporary directory: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(tempDir)

		// Clone the template repository
		cloneCmd := exec.Command("git", "clone", "https://github.com/dtg-lucifer/goth-stack-starter-template.git", tempDir)
		if err := cloneCmd.Run(); err != nil {
			fmt.Printf("Error cloning template repository: %v\n", err)
			os.Exit(1)
		}

		// Remove .git directory
		gitDir := filepath.Join(tempDir, ".git")
		if err := os.RemoveAll(gitDir); err != nil {
			fmt.Printf("Error removing .git directory: %v\n", err)
			os.Exit(1)
		}

		// Update go.mod file
		goModPath := filepath.Join(tempDir, "go.mod")
		goModContent, err := os.ReadFile(goModPath)
		if err != nil {
			fmt.Printf("Error reading go.mod file: %v\n", err)
			os.Exit(1)
		}

		// Replace the module path with user's GitHub username and project name
		newModulePath := fmt.Sprintf("github.com/%s/%s", githubUsername, projectName)
		updatedContent := strings.Replace(string(goModContent), "github.com/dtg-lucifer/goth-stack-starter", newModulePath, -1)

		if err := os.WriteFile(goModPath, []byte(updatedContent), 0644); err != nil {
			fmt.Printf("Error updating go.mod file: %v\n", err)
			os.Exit(1)
		}

		// Update import paths in all Go and Templ files
		err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && (strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".templ")) {
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				// Replace all import paths
				updatedContent := strings.ReplaceAll(string(content),
					"github.com/dtg-lucifer/goth-stack-starter",
					newModulePath)

				if err := os.WriteFile(path, []byte(updatedContent), 0644); err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			fmt.Printf("Error updating import paths: %v\n", err)
			os.Exit(1)
		}

		// Copy the template
		if err := copyDir(tempDir, targetDir); err != nil {
			fmt.Printf("Error copying template: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created new Goth Stack project '%s'!\n", projectName)
		fmt.Printf("Next steps:\n")
		fmt.Printf("1. cd %s\n", projectName)
		fmt.Printf("2. make setup\n")
		fmt.Printf("3. make dev\n")
	},
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Skip .git directory and other hidden files
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			content, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, content, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(createCmd)
}
