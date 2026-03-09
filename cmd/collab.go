package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	"github.com/tanq16/box/internal/types"
	u "github.com/tanq16/box/utils"
)

var collabCreateFlags struct {
	id      string
	role    string
	userID  string
	groupID string
}

var collabUpdateFlags struct {
	role   string
	status string
}

var collabCmd = &cobra.Command{
	Use:   "collab",
	Short: "Manage collaborations on Box items",
}

var collabCreateCmd = &cobra.Command{
	Use:   "create <path> <email> [--role viewer]",
	Short: "Create a collaboration on a file or folder",
	Long: `Share a file or folder with a user by email.

The most common usage is path + email:
  box collab create /Documents/project user@example.com
  box collab create /Documents/project user@example.com --role editor

For advanced cases, use --user-id or --group-id instead of a positional email.`,
	Args: cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var itemID, itemType string

		if collabCreateFlags.id != "" {
			itemID, itemType = resolveItemByID(collabCreateFlags.id)
			if len(args) < 1 && collabCreateFlags.userID == "" && collabCreateFlags.groupID == "" {
				u.PrintFatal("cmd","Must specify an email, --user-id, or --group-id", nil)
			}
		} else {
			if len(args) < 1 {
				u.PrintFatal("cmd","Must specify a path (or use --id)", nil)
			}
			itemID, itemType = resolveItem(args[:1], "")

			if len(args) < 2 && collabCreateFlags.userID == "" && collabCreateFlags.groupID == "" {
				u.PrintFatal("cmd","Must specify an email, --user-id, or --group-id", nil)
			}
		}

		var email string
		if collabCreateFlags.id != "" && len(args) >= 1 {
			email = args[0]
		} else if collabCreateFlags.id == "" && len(args) >= 2 {
			email = args[1]
		}

		collab, err := api.CreateCollaboration(boxClient, itemType, itemID, collabCreateFlags.role, email, collabCreateFlags.userID, collabCreateFlags.groupID)
		if err != nil {
			u.PrintFatal("cmd","Failed to create collaboration", err)
		}
		printCollab(collab)
		u.PrintSuccess("cmd","Collaboration created")
	},
}

var collabGetCmd = &cobra.Command{
	Use:   "get <collab-id>",
	Short: "Get a collaboration by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		collab, err := api.GetCollaboration(boxClient, args[0])
		if err != nil {
			u.PrintFatal("cmd","Failed to get collaboration", err)
		}
		printCollab(collab)
	},
}

var collabUpdateCmd = &cobra.Command{
	Use:   "update <collab-id>",
	Short: "Update a collaboration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		collab, err := api.UpdateCollaboration(boxClient, args[0], collabUpdateFlags.role, collabUpdateFlags.status)
		if err != nil {
			u.PrintFatal("cmd","Failed to update collaboration", err)
		}
		printCollab(collab)
		u.PrintSuccess("cmd","Collaboration updated")
	},
}

var collabDeleteCmd = &cobra.Command{
	Use:   "delete <collab-id>",
	Short: "Delete a collaboration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := api.DeleteCollaboration(boxClient, args[0]); err != nil {
			u.PrintFatal("cmd","Failed to delete collaboration", err)
		}
		u.PrintSuccess("cmd","Collaboration deleted")
	},
}

var collabPendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List pending collaborations",
	Run: func(cmd *cobra.Command, args []string) {
		list, err := api.ListPendingCollaborations(boxClient)
		if err != nil {
			u.PrintFatal("cmd","Failed to list pending collaborations", err)
		}
		if len(list.Entries) == 0 {
			u.PrintInfo("cmd","No pending collaborations")
			return
		}

		headers := []string{"ID", "ROLE", "STATUS", "ITEM", "ACCESSIBLE BY"}
		var rows [][]string
		for _, c := range list.Entries {
			itemStr := "-"
			if c.Item != nil {
				itemStr = fmt.Sprintf("%s:%s", c.Item.Type, c.Item.ID)
			}
			accessStr := "-"
			if c.AccessibleBy != nil {
				accessStr = c.AccessibleBy.Login
				if accessStr == "" {
					accessStr = c.AccessibleBy.Name
				}
			}
			rows = append(rows, []string{c.ID, c.Role, c.Status, itemStr, accessStr})
		}
		u.PrintTable(headers, rows)
	},
}

func printCollab(c *types.Collaboration) {
	u.PrintGeneric(fmt.Sprintf("ID:     %s", c.ID))
	u.PrintGeneric(fmt.Sprintf("Role:   %s", c.Role))
	u.PrintGeneric(fmt.Sprintf("Status: %s", c.Status))
	if c.Item != nil {
		u.PrintGeneric(fmt.Sprintf("Item:   %s %s (%s)", c.Item.Type, c.Item.ID, c.Item.Name))
	}
	if c.AccessibleBy != nil {
		u.PrintGeneric(fmt.Sprintf("User:   %s (%s)", c.AccessibleBy.Name, c.AccessibleBy.Login))
	}
}

func init() {
	collabCreateCmd.Flags().StringVar(&collabCreateFlags.id, "id", "", "Item ID (instead of path)")
	collabCreateCmd.Flags().StringVarP(&collabCreateFlags.role, "role", "r", "viewer", "Permission role (viewer, editor, co-owner, etc.)")
	collabCreateCmd.Flags().StringVar(&collabCreateFlags.userID, "user-id", "", "User ID (instead of email)")
	collabCreateCmd.Flags().StringVar(&collabCreateFlags.groupID, "group-id", "", "Group ID (instead of email)")

	collabUpdateCmd.Flags().StringVarP(&collabUpdateFlags.role, "role", "r", "", "New role")
	collabUpdateCmd.Flags().StringVarP(&collabUpdateFlags.status, "status", "s", "", "New status (accepted, rejected)")

	collabCmd.AddCommand(collabCreateCmd)
	collabCmd.AddCommand(collabGetCmd)
	collabCmd.AddCommand(collabUpdateCmd)
	collabCmd.AddCommand(collabDeleteCmd)
	collabCmd.AddCommand(collabPendingCmd)
	rootCmd.AddCommand(collabCmd)
}
