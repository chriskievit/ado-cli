/*
Copyright © 2023 Chris Kievit chris.kievit@gmail.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/chriskievit/ado-cli/model"
	"github.com/chriskievit/ado-cli/util/config"
	"github.com/go-git/go-git/v5"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	adogit "github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/spf13/cobra"
)

var WorkItemId int
var Name string
var Clean bool

// linkCmd represents the link command
var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Link an existing branch or create a new branch and link it to a work item",
	Long: `Link an existing branch or create a new branch and link it to a work item.
Creating a new branch will use a clean version of the default branch.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if local directory is a git repository
		repoDefinition := validateGitRepository()

		// Check if work item exists
		projectDefinition := getWorkItem()

		// Link the branch
		linkBranch(repoDefinition, projectDefinition)
	},
}

func init() {
	rootCmd.AddCommand(linkCmd)

	linkCmd.Flags().IntVarP(&WorkItemId, "work-item", "w", 0, "work item ID to link to branch")
	linkCmd.MarkFlagRequired("work-item")

	linkCmd.Flags().BoolVarP(&Clean, "clean", "c", true, "will branch from a clean version of the default branch")
	linkCmd.Flags().StringVarP(&Name, "name", "n", "", "name of the new branch. if not specified, the work item title will be used")
}

func validateGitRepository() model.RepoDefinition {
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting working directory: ", err)
		os.Exit(1)
	}

	repo, err := git.PlainOpen(workingDir)
	if err != nil {
		_ = repo
		fmt.Printf("\033[0;31mThe current directory isn't a valid git repository.\n\033[0m")
		os.Exit(1)
	}

	remoteUrl := ""
	headRef, err := repo.Head()
	remotes, err := repo.Remotes()
	for _, remote := range remotes {
		for _, url := range remote.Config().URLs {
			found, err := regexp.MatchString(".*dev\\.azure\\.com(:v\\d{1}){0,1}\\/"+config.GetOrganizationName(), url)
			if err != nil {
			}

			if found {
				remoteUrl = url
				fmt.Printf("\u001b[33mFound remote: %s\n\033[0m", url)
				break
			}
		}
	}
	fmt.Println("Currently on branch: " + headRef.Name().Short())

	if remoteUrl == "" {
		fmt.Printf("\033[0;31mCould not find Azure DevOps remote.\n\033[0m")
		os.Exit(1)
	}

	parts := strings.Split(remoteUrl, "/")
	projectName := parts[len(parts)-2]
	project, err := getProject(projectName)

	if err != nil {
		fmt.Printf("\033[0;31mCould not find Azure DevOps project (%s) based on current remote.\n\033[0m", projectName)
		os.Exit(1)
	}
	fmt.Printf("\u001b[32mFound project: %s (%s)\n\033[0m", *project.Name, *project.Id)

	repositoryName := parts[len(parts)-1]
	repository, err := getRepository(repositoryName)

	if err != nil {
		fmt.Printf("\033[0;31mCould not find Azure DevOps repository (%s) based on current remote.\n\033[0m", repositoryName)
		os.Exit(1)
	}
	fmt.Printf("\u001b[32mFound repository: %s (%s)\n\033[0m", *repository.Name, *repository.Id)

	return model.RepoDefinition{
		ProjectId:    project.Id.String(),
		RepositoryId: repository.Id.String(),
		BranchName:   headRef.Name().Short(),
	}
}

func getWorkItem() model.WorkItemDefinition {
	// Read existing configuration
	url := config.GetOrganizationUrl()
	pat := config.GetPat()

	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(url, pat)

	ctx := context.Background()

	// Create a client to interact with the Core area
	client, err := workitemtracking.NewClient(ctx, connection)
	if err != nil {
		fmt.Println("Error creating client: ", err)
		os.Exit(1)
	}

	fmt.Println("Getting work item...")
	// Get first page of the list of team projects for your organization
	responseValue, err := client.GetWorkItem(ctx, workitemtracking.GetWorkItemArgs{Id: &WorkItemId})
	if err != nil {
		_ = responseValue
		fmt.Printf("Error getting workitem: %s\n", err)
		os.Exit(1)
	}

	fields := *responseValue.Fields
	name := fields["System.Title"]
	fmt.Printf("\u001b[32mFound work item: %s\n\033[0m", name)

	return model.WorkItemDefinition{
		WorkItemId:  WorkItemId,
		ProjectName: fields["System.TeamProject"].(string),
	}
}

func getProject(name string) (core.TeamProjectReference, error) {
	// Read existing configuration
	url := config.GetOrganizationUrl()
	pat := config.GetPat()

	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(url, pat)

	ctx := context.Background()

	// Create a client to interact with the Core area
	client, err := core.NewClient(ctx, connection)
	if err != nil {
		fmt.Println("Error creating client: ", err)
	}

	fmt.Println("Getting project...")
	// Get first page of the list of team projects for your organization
	responseValue, err := client.GetProjects(ctx, core.GetProjectsArgs{})
	if err != nil {
		_ = responseValue
		fmt.Printf("Error validating connection: %s\n", err)
	}

	projects := responseValue.Value
	for _, project := range projects {
		if *project.Name == name {
			return project, nil
		}
	}

	return core.TeamProjectReference{}, errors.New("Project not found")
}

func getRepository(name string) (adogit.GitRepository, error) {
	// Read existing configuration
	url := config.GetOrganizationUrl()
	pat := config.GetPat()

	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(url, pat)

	ctx := context.Background()

	// Create a client to interact with the Core area
	client, err := adogit.NewClient(ctx, connection)
	if err != nil {
		fmt.Println("Error creating client: ", err)
	}

	fmt.Println("Getting repository...")
	// Get first page of the list of team projects for your organization
	responseValue, err := client.GetRepositories(ctx, adogit.GetRepositoriesArgs{})
	if err != nil {
		_ = responseValue
		fmt.Printf("Error validating connection: %s\n", err)
	}

	repositories := *responseValue
	for _, repository := range repositories {
		if *repository.Name == name {
			return repository, nil
		}
	}

	return adogit.GitRepository{}, errors.New("Repository not found")
}

func linkBranch(repoDefinition model.RepoDefinition, projectDefinition model.WorkItemDefinition) {
	// Read existing configuration
	url := config.GetOrganizationUrl()
	pat := config.GetPat()

	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(url, pat)

	ctx := context.Background()

	// Create a client to interact with the Core area
	client, err := workitemtracking.NewClient(ctx, connection)
	if err != nil {
		fmt.Println("Error creating client: ", err)
		os.Exit(1)
	}

	fmt.Println("Updating work item...")
	path := "/relations/-"
	// Get first page of the list of team projects for your organization
	responseValue, err := client.UpdateWorkItem(ctx, workitemtracking.UpdateWorkItemArgs{
		Id: &projectDefinition.WorkItemId,
		Document: &[]webapi.JsonPatchOperation{
			{
				Op:   &webapi.OperationValues.Add,
				Path: &path,
				Value: map[string]interface{}{
					"rel": "ArtifactLink",
					"url": fmt.Sprintf("vstfs:///Git/Ref/%s/%s/GB%s",
						repoDefinition.ProjectId,
						repoDefinition.RepositoryId,
						repoDefinition.BranchName),
					"attributes": map[string]string{
						"name":    "Branch",
						"comment": "Linked via ado-cli",
					},
				},
			},
		},
	})

	if err != nil {
		_ = responseValue
		fmt.Printf("\033[0;31mUnable to link branch to work item: %s\n\033[0m", err)
		os.Exit(1)
	}

	fmt.Printf("\u001b[32mSuccessfully linked branch to work item.\n\033[0m")
}
