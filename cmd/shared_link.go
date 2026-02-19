package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/internal/utils"
)

var (
	slItemID   string
	slAccess   string
	slPassword string
)

var sharedLinkCmd = &cobra.Command{
	Use:   "shared-link",
	Short: "Manage shared links on Box items",
}

var slCreateCmd = &cobra.Command{
	Use:   "create <path>",
	Short: "Create a shared link",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		itemID, itemType := resolveItem(args, slItemID)
		item, err := api.CreateSharedLink(boxClient, itemType, itemID, slAccess, slPassword)
		if err != nil {
			u.PrintFatal("Failed to create shared link", err)
		}
		if item.SharedLink != nil {
			u.PrintGeneric(fmt.Sprintf("URL:    %s", item.SharedLink.URL))
			u.PrintGeneric(fmt.Sprintf("Access: %s", item.SharedLink.Access))
			if item.SharedLink.IsPasswordEnabled {
				u.PrintGeneric("Password: enabled")
			}
		}
		u.PrintSuccess("Shared link created")
	},
}

var slGetCmd = &cobra.Command{
	Use:   "get <path>",
	Short: "Get shared link info for an item",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		itemID, itemType := resolveItem(args, slItemID)
		item, err := api.GetSharedLink(boxClient, itemType, itemID)
		if err != nil {
			u.PrintFatal("Failed to get shared link", err)
		}
		if item.SharedLink == nil {
			u.PrintInfo("No shared link on this item")
			return
		}
		u.PrintGeneric(fmt.Sprintf("URL:    %s", item.SharedLink.URL))
		u.PrintGeneric(fmt.Sprintf("Access: %s", item.SharedLink.Access))
		if item.SharedLink.IsPasswordEnabled {
			u.PrintGeneric("Password: enabled")
		}
	},
}

var slRemoveCmd = &cobra.Command{
	Use:   "remove <path>",
	Short: "Remove shared link from an item",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		itemID, itemType := resolveItem(args, slItemID)
		if err := api.RemoveSharedLink(boxClient, itemType, itemID); err != nil {
			u.PrintFatal("Failed to remove shared link", err)
		}
		u.PrintSuccess("Shared link removed")
	},
}

var slResolveCmd = &cobra.Command{
	Use:   "resolve <url>",
	Short: "Resolve a shared link URL to get item info",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		item, err := api.ResolveSharedLink(boxClient, args[0], slPassword)
		if err != nil {
			u.PrintFatal("Failed to resolve shared link", err)
		}
		u.PrintGeneric(fmt.Sprintf("Type: %s", item.Type))
		u.PrintGeneric(fmt.Sprintf("ID:   %s", item.ID))
		u.PrintGeneric(fmt.Sprintf("Name: %s", item.Name))
	},
}

func init() {
	// Shared --id flag on create/get/remove (alternative to path)
	slCreateCmd.Flags().StringVar(&slItemID, "id", "", "Item ID (instead of path)")
	slGetCmd.Flags().StringVar(&slItemID, "id", "", "Item ID (instead of path)")
	slRemoveCmd.Flags().StringVar(&slItemID, "id", "", "Item ID (instead of path)")

	// Create-specific flags
	slCreateCmd.Flags().StringVarP(&slAccess, "access", "a", "open", "Access level (open, company, collaborators)")
	slCreateCmd.Flags().StringVarP(&slPassword, "password", "P", "", "Password for shared link")

	// Resolve-specific flag
	slResolveCmd.Flags().StringVarP(&slPassword, "password", "P", "", "Shared link password")

	sharedLinkCmd.AddCommand(slCreateCmd)
	sharedLinkCmd.AddCommand(slGetCmd)
	sharedLinkCmd.AddCommand(slRemoveCmd)
	sharedLinkCmd.AddCommand(slResolveCmd)
	rootCmd.AddCommand(sharedLinkCmd)
}
