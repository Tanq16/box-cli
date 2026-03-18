package sharedlinkcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/cmd/cmdutil"
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

var slFlags struct {
	itemID   string
	access   string
	password string
}

var SharedLinkCmd = &cobra.Command{
	Use:   "shared-link",
	Short: "Manage shared links on Box items",
}

var slCreateCmd = &cobra.Command{
	Use:   "create <path>",
	Short: "Create a shared link",
	Args:  cobra.MaximumNArgs(1),
	Run: func(c *cobra.Command, args []string) {
		itemID, itemType := cmdutil.ResolveItem(args, slFlags.itemID)
		item, err := api.CreateSharedLink(cmdutil.BoxClient, itemType, itemID, slFlags.access, slFlags.password)
		if err != nil {
			u.PrintFatal("cmd","Failed to create shared link", err)
		}
		if item.SharedLink != nil {
			u.PrintGeneric(fmt.Sprintf("URL:    %s", item.SharedLink.URL))
			u.PrintGeneric(fmt.Sprintf("Access: %s", item.SharedLink.Access))
			if item.SharedLink.IsPasswordEnabled {
				u.PrintGeneric("Password: enabled")
			}
		}
		u.PrintSuccess("cmd","Shared link created")
	},
}

var slGetCmd = &cobra.Command{
	Use:   "get <path>",
	Short: "Get shared link info for an item",
	Args:  cobra.MaximumNArgs(1),
	Run: func(c *cobra.Command, args []string) {
		itemID, itemType := cmdutil.ResolveItem(args, slFlags.itemID)
		item, err := api.GetSharedLink(cmdutil.BoxClient, itemType, itemID)
		if err != nil {
			u.PrintFatal("cmd","Failed to get shared link", err)
		}
		if item.SharedLink == nil {
			u.PrintInfo("cmd","No shared link on this item")
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
	Run: func(c *cobra.Command, args []string) {
		itemID, itemType := cmdutil.ResolveItem(args, slFlags.itemID)
		if err := api.RemoveSharedLink(cmdutil.BoxClient, itemType, itemID); err != nil {
			u.PrintFatal("cmd","Failed to remove shared link", err)
		}
		u.PrintSuccess("cmd","Shared link removed")
	},
}

var slResolveCmd = &cobra.Command{
	Use:   "resolve <url>",
	Short: "Resolve a shared link URL to get item info",
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		item, err := api.ResolveSharedLink(cmdutil.BoxClient, args[0], slFlags.password)
		if err != nil {
			u.PrintFatal("cmd","Failed to resolve shared link", err)
		}
		u.PrintGeneric(fmt.Sprintf("Type: %s", item.Type))
		u.PrintGeneric(fmt.Sprintf("ID:   %s", item.ID))
		u.PrintGeneric(fmt.Sprintf("Name: %s", item.Name))
	},
}

func init() {
	slCreateCmd.Flags().StringVar(&slFlags.itemID, "id", "", "Item ID (instead of path)")
	slGetCmd.Flags().StringVar(&slFlags.itemID, "id", "", "Item ID (instead of path)")
	slRemoveCmd.Flags().StringVar(&slFlags.itemID, "id", "", "Item ID (instead of path)")

	slCreateCmd.Flags().StringVarP(&slFlags.access, "access", "a", "open", "Access level (open, company, collaborators)")
	slCreateCmd.Flags().StringVarP(&slFlags.password, "password", "P", "", "Password for shared link")

	slResolveCmd.Flags().StringVarP(&slFlags.password, "password", "P", "", "Shared link password")

	SharedLinkCmd.AddCommand(slCreateCmd)
	SharedLinkCmd.AddCommand(slGetCmd)
	SharedLinkCmd.AddCommand(slRemoveCmd)
	SharedLinkCmd.AddCommand(slResolveCmd)
}
