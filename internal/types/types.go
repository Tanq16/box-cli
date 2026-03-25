package types

import "time"

type BoxItem struct {
	Type           string          `json:"type"`
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Size           *int64          `json:"size,omitempty"`
	SHA1           string          `json:"sha1,omitempty"`
	Parent         *BoxRef         `json:"parent,omitempty"`
	ModifiedAt     *string         `json:"modified_at,omitempty"`
	SharedLink     *BoxSharedLink  `json:"shared_link,omitempty"`
	ItemCollection *BoxFolderItems `json:"item_collection,omitempty"`
}

type BoxRef struct {
	Type string `json:"type,omitempty"`
	ID   string `json:"id"`
}

type BoxFolderItems struct {
	TotalCount int       `json:"total_count"`
	Entries    []BoxItem `json:"entries"`
	Offset     int       `json:"offset"`
	Limit      int       `json:"limit"`
}

type BoxError struct {
	Type    string `json:"type"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type BoxItemDisplay struct {
	ID           string
	Name         string
	ModifiedTime string
	Size         int64
	Type         string
}

type UploadSession struct {
	ID                string                  `json:"id"`
	Type              string                  `json:"type"`
	SessionExpiresAt  *time.Time              `json:"session_expires_at,omitempty"`
	PartSize          int64                   `json:"part_size"`
	TotalParts        int                     `json:"total_parts"`
	NumPartsProcessed int                     `json:"num_parts_processed"`
	SessionEndpoints  *UploadSessionEndpoints `json:"session_endpoints,omitempty"`
}

type UploadSessionEndpoints struct {
	UploadPart string `json:"upload_part"`
	Commit     string `json:"commit"`
	Abort      string `json:"abort"`
	ListParts  string `json:"list_parts"`
	Status     string `json:"status"`
}

type UploadPart struct {
	PartID string `json:"part_id"`
	Offset int64  `json:"offset"`
	Size   int64  `json:"size"`
	SHA1   string `json:"sha1"`
}

type UploadPartResponse struct {
	Part UploadPart `json:"part"`
}

type FileTree struct {
	Files map[string]FileInfo
	Dirs  map[string]*FileTree
}

type FileInfo struct {
	Path string
	Hash string
	ID   string
}

type IndexItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Type         string `json:"type"`
	Size         int64  `json:"size"`
	ModifiedTime string `json:"modified_time"`
}

type IndexStore struct {
	Provider  string      `json:"provider"`
	RootPath  string      `json:"root_path"`
	Timestamp time.Time   `json:"timestamp"`
	Items     []IndexItem `json:"items"`
}

type Collaboration struct {
	ID           string           `json:"id"`
	Type         string           `json:"type"`
	Item         *CollabItem      `json:"item,omitempty"`
	AccessibleBy *CollabAccessor  `json:"accessible_by,omitempty"`
	Role         string           `json:"role"`
	Status       string           `json:"status"`
	ExpiresAt    *string          `json:"expires_at,omitempty"`
	InviteEmail  *string          `json:"invite_email,omitempty"`
	CreatedBy    *CollabAccessor  `json:"created_by,omitempty"`
	CreatedAt    *string          `json:"created_at,omitempty"`
	ModifiedAt   *string          `json:"modified_at,omitempty"`
}

type CollabItem struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type CollabAccessor struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Login string `json:"login,omitempty"`
}

type CollaborationList struct {
	TotalCount int             `json:"total_count"`
	Entries    []Collaboration `json:"entries"`
	Offset     int             `json:"offset"`
	Limit      int             `json:"limit"`
}

type SearchResults struct {
	TotalCount int       `json:"total_count"`
	Entries    []BoxItem `json:"entries"`
	Offset     int       `json:"offset"`
	Limit      int       `json:"limit"`
}

type SearchOptions struct {
	Query        string
	Type         string
	Extensions   []string
	FolderID     string
	CreatedAfter string
	CreatedBefore string
	UpdatedAfter  string
	UpdatedBefore string
	SizeMin      int64
	SizeMax      int64
	Owner        string
	Sort         string
	Limit        int
}

type BoxSharedLink struct {
	URL               string              `json:"url,omitempty"`
	DownloadURL       string              `json:"download_url,omitempty"`
	VanityURL         string              `json:"vanity_url,omitempty"`
	VanityName        string              `json:"vanity_name,omitempty"`
	Access            string              `json:"access,omitempty"`
	EffectiveAccess   string              `json:"effective_access,omitempty"`
	EffectivePermission string            `json:"effective_permission,omitempty"`
	IsPasswordEnabled bool                `json:"is_password_enabled,omitempty"`
	UnsharedAt        *string             `json:"unshared_at,omitempty"`
	DownloadCount     int                 `json:"download_count,omitempty"`
	PreviewCount      int                 `json:"preview_count,omitempty"`
	Permissions       *SharedLinkPerms    `json:"permissions,omitempty"`
}

type SharedLinkPerms struct {
	CanDownload bool `json:"can_download"`
	CanPreview  bool `json:"can_preview"`
	CanEdit     bool `json:"can_edit"`
}

type BoxFileList struct {
	TotalCount int       `json:"total_count"`
	Entries    []BoxItem `json:"entries"`
}
