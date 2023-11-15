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
	"fmt"
	"os"

	"github.com/chriskievit/ado-cli/util/config"
	"github.com/chriskievit/ado-cli/util/stdin"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Configure the organization and PAT for your Azure DevOps account",
	Long: `Configure the organization and PAT (Personal Access Token) for your Azure DevOps account.

The PAT is used to authenticate with Azure DevOps and must have the following permissions:

- Work Items (Read & Write)
- Code (Read & Write)`,
	Run: func(cmd *cobra.Command, args []string) {
		// Read existing configuration
		url := config.GetOrganizationUrl()
		pat := config.GetPat()

		// Read required configuration from stdin.
		// Because of the sensitivity of the parameters, we don't want to use flags.
		url = stdin.ReadInput("Please input your organization URL (i.e. https://dev.azure.com/<myorg>)", url)
		config.SetOrganizationUrl(url)
		pat = stdin.ReadInput("Please input your PAT", pat)
		config.SetPat(pat)

		// Validate connection
		fmt.Println("Validating connection to Azure DevOps...")
		success, err := validateConnection(url, pat)
		if !success || err != nil {
			fmt.Println("Connection failed. Please check your configuration and try again.")
			os.Exit(1)
		}

		fmt.Println("Configuration successfully initialized.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func validateConnection(url string, pat string) (bool, error) {
	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(url, pat)

	ctx := context.Background()

	// Create a client to interact with the Core area
	client, err := core.NewClient(ctx, connection)
	if err != nil {
		fmt.Println("Error creating client: ", err)
		return false, err
	}

	// Get first page of the list of team projects for your organization
	responseValue, err := client.GetProjects(ctx, core.GetProjectsArgs{})
	if err != nil {
		_ = responseValue
		fmt.Printf("Error validating connection: %s\n", err)
		return false, err
	}

	return true, nil
}
