package cmdutil

import (
	"strings"

	"github.com/tanq16/box/internal/api"
	"github.com/tanq16/box/internal/client"
	u "github.com/tanq16/box/utils"
)

var BoxClient *client.BoxClient

func Confirm(prompt string) bool {
	answer, err := u.PromptInput(prompt+" [y/N]", "")
	if err != nil {
		return false
	}
	answer = strings.ToLower(answer)
	return answer == "y" || answer == "yes"
}

func ResolveItemByID(id string) (string, string) {
	info, err := api.GetFileInfo(BoxClient, id)
	if err == nil {
		return info.ID, info.Type
	}
	fInfo, err := api.GetFolderInfo(BoxClient, id)
	if err == nil {
		return fInfo.ID, fInfo.Type
	}
	u.PrintFatal("Failed to resolve item ID", err)
	return "", ""
}

func ResolveItem(args []string, idFlag string) (string, string) {
	if idFlag != "" {
		return ResolveItemByID(idFlag)
	}
	if len(args) == 0 {
		u.PrintFatal("Must specify a path or --id", nil)
	}
	itemID, itemType, err := api.ResolvePath(BoxClient, args[0], "")
	if err != nil {
		u.PrintFatal("Failed to resolve path", err)
	}
	return itemID, itemType
}
