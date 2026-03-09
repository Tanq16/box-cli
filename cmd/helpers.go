package cmd

import (
	"github.com/tanq16/box/internal/api"
	u "github.com/tanq16/box/utils"
)

func resolveItemByID(id string) (string, string) {
	info, err := api.GetFileInfo(boxClient, id)
	if err == nil {
		return info.ID, info.Type
	}
	fInfo, err := api.GetFolderInfo(boxClient, id)
	if err == nil {
		return fInfo.ID, fInfo.Type
	}
	u.PrintFatal("cmd","Failed to resolve item ID", err)
	return "", ""
}

func resolveItem(args []string, idFlag string) (string, string) {
	if idFlag != "" {
		return resolveItemByID(idFlag)
	}
	if len(args) == 0 {
		u.PrintFatal("cmd","Must specify a path or --id", nil)
	}
	itemID, itemType, err := api.ResolvePath(boxClient, args[0], "")
	if err != nil {
		u.PrintFatal("cmd","Failed to resolve path", err)
	}
	return itemID, itemType
}
