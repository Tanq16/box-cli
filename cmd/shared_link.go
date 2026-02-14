package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
)

var (
	slType     string
	slID       string
	slAccess   string
	slPassword string
)

var sharedLinkCmd = &cobra.Command{
	Use:   "shared-link",
	Short: "Manage shared links on Box items",
}

var slCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a shared link",
	RunE: func(cmd *cobra.Command, args []string) error {
		if slType == "" || slID == "" {
			return fmt.Errorf("--type and --id are required")
		}
		access := slAccess
		if access == "" {
			access = "open"
		}
		item, err := api.CreateSharedLink(boxClient, slType, slID, access, slPassword)
		if err != nil {
			return err
		}
		if item.SharedLink != nil {
			fmt.Printf("URL:    %s\n", item.SharedLink.URL)
			fmt.Printf("Access: %s\n", item.SharedLink.Access)
			if item.SharedLink.IsPasswordEnabled {
				fmt.Println("Password: enabled")
			}
		}
		return nil
	},
}

var slGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get shared link info for an item",
	RunE: func(cmd *cobra.Command, args []string) error {
		if slType == "" || slID == "" {
			return fmt.Errorf("--type and --id are required")
		}
		item, err := api.GetSharedLink(boxClient, slType, slID)
		if err != nil {
			return err
		}
		if item.SharedLink == nil {
			fmt.Println("No shared link on this item.")
			return nil
		}
		fmt.Printf("URL:    %s\n", item.SharedLink.URL)
		fmt.Printf("Access: %s\n", item.SharedLink.Access)
		if item.SharedLink.IsPasswordEnabled {
			fmt.Println("Password: enabled")
		}
		return nil
	},
}

var slRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove shared link from an item",
	RunE: func(cmd *cobra.Command, args []string) error {
		if slType == "" || slID == "" {
			return fmt.Errorf("--type and --id are required")
		}
		if err := api.RemoveSharedLink(boxClient, slType, slID); err != nil {
			return err
		}
		fmt.Println("Shared link removed.")
		return nil
	},
}

var slResolveCmd = &cobra.Command{
	Use:   "resolve <url>",
	Short: "Resolve a shared link URL to get item info",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		item, err := api.ResolveSharedLink(boxClient, args[0], slPassword)
		if err != nil {
			return err
		}
		fmt.Printf("Type: %s\n", item.Type)
		fmt.Printf("ID:   %s\n", item.ID)
		fmt.Printf("Name: %s\n", item.Name)
		return nil
	},
}

func init() {
	slCreateCmd.Flags().StringVar(&slType, "type", "", "Item type (file or folder)")
	slCreateCmd.Flags().StringVar(&slID, "id", "", "Item ID")
	slCreateCmd.Flags().StringVar(&slAccess, "access", "open", "Access level (open, company, collaborators)")
	slCreateCmd.Flags().StringVar(&slPassword, "password", "", "Password for shared link")

	slGetCmd.Flags().StringVar(&slType, "type", "", "Item type (file or folder)")
	slGetCmd.Flags().StringVar(&slID, "id", "", "Item ID")

	slRemoveCmd.Flags().StringVar(&slType, "type", "", "Item type (file or folder)")
	slRemoveCmd.Flags().StringVar(&slID, "id", "", "Item ID")

	slResolveCmd.Flags().StringVar(&slPassword, "password", "", "Shared link password")

	sharedLinkCmd.AddCommand(slCreateCmd)
	sharedLinkCmd.AddCommand(slGetCmd)
	sharedLinkCmd.AddCommand(slRemoveCmd)
	sharedLinkCmd.AddCommand(slResolveCmd)
	rootCmd.AddCommand(sharedLinkCmd)
}
