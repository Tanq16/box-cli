package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tanq16/box/internal/api"
	"github.com/tanq16/box/internal/types"
)

var (
	collabItemType  string
	collabItemID    string
	collabRole      string
	collabUserEmail string
	collabUserID    string
	collabGroupID   string
	collabStatus    string
)

var collabCmd = &cobra.Command{
	Use:   "collab",
	Short: "Manage collaborations on Box items",
}

var collabCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a collaboration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if collabItemType == "" || collabItemID == "" || collabRole == "" {
			return fmt.Errorf("--item-type, --item-id, and --role are required")
		}
		collab, err := api.CreateCollaboration(boxClient, collabItemType, collabItemID, collabRole, collabUserEmail, collabUserID, collabGroupID)
		if err != nil {
			return err
		}
		printCollab(collab)
		return nil
	},
}

var collabGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a collaboration by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		collab, err := api.GetCollaboration(boxClient, args[0])
		if err != nil {
			return err
		}
		printCollab(collab)
		return nil
	},
}

var collabUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a collaboration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		collab, err := api.UpdateCollaboration(boxClient, args[0], collabRole, collabStatus)
		if err != nil {
			return err
		}
		printCollab(collab)
		return nil
	},
}

var collabDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a collaboration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := api.DeleteCollaboration(boxClient, args[0]); err != nil {
			return err
		}
		fmt.Println("Collaboration deleted.")
		return nil
	},
}

var collabPendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List pending collaborations",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := api.ListPendingCollaborations(boxClient)
		if err != nil {
			return err
		}
		if len(list.Entries) == 0 {
			fmt.Println("No pending collaborations.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tROLE\tSTATUS\tITEM\tACCESSIBLE BY")
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
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", c.ID, c.Role, c.Status, itemStr, accessStr)
		}
		w.Flush()
		return nil
	},
}

func printCollab(c *types.Collaboration) {
	fmt.Printf("ID:     %s\n", c.ID)
	fmt.Printf("Role:   %s\n", c.Role)
	fmt.Printf("Status: %s\n", c.Status)
	if c.Item != nil {
		fmt.Printf("Item:   %s %s (%s)\n", c.Item.Type, c.Item.ID, c.Item.Name)
	}
	if c.AccessibleBy != nil {
		fmt.Printf("User:   %s (%s)\n", c.AccessibleBy.Name, c.AccessibleBy.Login)
	}
}

func init() {
	collabCreateCmd.Flags().StringVar(&collabItemType, "item-type", "", "Item type (file or folder)")
	collabCreateCmd.Flags().StringVar(&collabItemID, "item-id", "", "Item ID")
	collabCreateCmd.Flags().StringVar(&collabRole, "role", "", "Role (editor, viewer, etc.)")
	collabCreateCmd.Flags().StringVar(&collabUserEmail, "user-email", "", "User email")
	collabCreateCmd.Flags().StringVar(&collabUserID, "user-id", "", "User ID")
	collabCreateCmd.Flags().StringVar(&collabGroupID, "group-id", "", "Group ID")

	collabUpdateCmd.Flags().StringVar(&collabRole, "role", "", "New role")
	collabUpdateCmd.Flags().StringVar(&collabStatus, "status", "", "New status (accepted, rejected)")

	collabCmd.AddCommand(collabCreateCmd)
	collabCmd.AddCommand(collabGetCmd)
	collabCmd.AddCommand(collabUpdateCmd)
	collabCmd.AddCommand(collabDeleteCmd)
	collabCmd.AddCommand(collabPendingCmd)
	rootCmd.AddCommand(collabCmd)
}
