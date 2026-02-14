# Reference Plan C: OpenAPI-Generated Go Client for Box

> Full Go client generated from the Box OpenAPI 3.0 specification using `openapi-generator` v7.19.0.
> Generated code lives at `/Users/tanishqrupaal/box-go-generated/` (package `boxapi`).

---

## Table of Contents

1. [Generation Details](#1-generation-details)
2. [Directory Structure](#2-directory-structure)
3. [Dependencies](#3-dependencies)
4. [Client Setup & Authentication](#4-client-setup--authentication)
5. [Available API Services (73 total)](#5-available-api-services-73-total)
6. [Key Model Structs](#6-key-model-structs)
7. [Usage Examples](#7-usage-examples)
8. [How to Integrate with Custom Auth](#8-how-to-integrate-with-custom-auth)
9. [Pros and Cons](#9-pros-and-cons)

---

## 1. Generation Details

**Source spec:** Box OpenAPI 3.0 spec from `https://github.com/box/box-openapi` (`openapi.json`)
**Downloaded to:** `/Users/tanishqrupaal/box-openapi.json` (1.7MB)

**Generation command:**
```bash
openapi-generator generate \
  -i /Users/tanishqrupaal/box-openapi.json \
  -g go \
  -o /Users/tanishqrupaal/box-go-generated/ \
  --package-name boxapi
```

**Generator version:** openapi-generator v7.19.0 (installed via Homebrew)

**Output stats:**
- 565 model files (`model_*.go`)
- 73 API service files (`api_*.go`)
- 642 total `.go` files
- Package name: `boxapi`

---

## 2. Directory Structure

```
box-go-generated/
├── go.mod                          # Module definition
├── go.sum                          # Dependency checksums
├── client.go                       # APIClient struct with all service pointers
├── configuration.go                # Config: base URL, auth, HTTP settings
├── response.go                     # Response wrapper type
├── utils.go                        # Helper functions
├── api/
│   └── openapi.yaml                # Converted OpenAPI spec
├── docs/                           # Markdown docs for every API & model
├── test/                           # Test stubs
│
├── api_files.go                    # FilesAPI service
├── api_folders.go                  # FoldersAPI service
├── api_collaborations.go           # CollaborationsAPI service
├── api_collaborations_list.go      # CollaborationsListAPI service
├── api_comments.go                 # CommentsAPI service
├── api_users.go                    # UsersAPI service
├── api_groups.go                   # GroupsAPI service
├── api_group_memberships.go        # GroupMembershipsAPI service
├── api_search.go                   # SearchAPI service
├── api_uploads.go                  # UploadsAPI service (simple)
├── api_uploads_chunked.go          # UploadsChunkedAPI service
├── api_downloads.go                # DownloadsAPI service
├── api_trashed_items.go            # TrashedItemsAPI service
├── api_trashed_files.go            # TrashedFilesAPI service
├── api_trashed_folders.go          # TrashedFoldersAPI service
├── api_shared_links_files.go       # SharedLinksFilesAPI service
├── api_shared_links_folders.go     # SharedLinksFoldersAPI service
├── api_events.go                   # EventsAPI service
├── api_webhooks.go                 # WebhooksAPI service
├── api_tasks.go                    # TasksAPI service
├── ... (73 total api_*.go files)
│
├── model_file.go                   # File struct
├── model_file__mini.go             # FileMini struct
├── model_file__full.go             # FileFull struct
├── model_folder.go                 # Folder struct
├── model_folder__mini.go           # FolderMini struct
├── model_folder__full.go           # FolderFull struct
├── model_user.go                   # User struct
├── model_user__mini.go             # UserMini struct
├── model_user__full.go             # UserFull struct
├── model_collaboration.go          # Collaboration struct
├── model_upload_session.go         # UploadSession struct
├── model_upload_part.go            # UploadPart struct
├── ... (565 total model_*.go files)
```

---

## 3. Dependencies

```go
// go.mod
module github.com/GIT_USER_ID/GIT_REPO_ID

go 1.23

require (
    github.com/stretchr/testify v1.10.0    // testing
    golang.org/x/oauth2 v0.27.0            // OAuth2 support (built-in!)
    gopkg.in/validator.v2 v2.0.1           // struct validation
)
```

The generated client **natively supports OAuth2** via `golang.org/x/oauth2`.

---

## 4. Client Setup & Authentication

### Configuration

From `configuration.go`:

```go
type Configuration struct {
    Host             string            `json:"host,omitempty"`
    Scheme           string            `json:"scheme,omitempty"`
    DefaultHeader    map[string]string `json:"defaultHeader,omitempty"`
    UserAgent        string            `json:"userAgent,omitempty"`
    Debug            bool              `json:"debug,omitempty"`
    Servers          ServerConfigurations
    OperationServers map[string]ServerConfigurations
    HTTPClient       *http.Client
}
```

### Auth Context Keys

```go
var (
    // ContextOAuth2 takes an oauth2.TokenSource as authentication
    ContextOAuth2 = contextKey("token")

    // ContextServerIndex uses a server configuration from the index
    ContextServerIndex = contextKey("serverIndex")
)
```

### APIClient

From `client.go` — the main client struct with all 73 service pointers:

```go
type APIClient struct {
    cfg    *Configuration
    common service

    // API Services (relevant subset):
    CollaborationsAPI     *CollaborationsAPIService
    CollaborationsListAPI *CollaborationsListAPIService
    CommentsAPI           *CommentsAPIService
    DownloadsAPI          *DownloadsAPIService
    EventsAPI             *EventsAPIService
    FilesAPI              *FilesAPIService
    FoldersAPI            *FoldersAPIService
    GroupMembershipsAPI   *GroupMembershipsAPIService
    GroupsAPI             *GroupsAPIService
    SearchAPI             *SearchAPIService
    SharedLinksFilesAPI   *SharedLinksFilesAPIService
    SharedLinksFoldersAPI *SharedLinksFoldersAPIService
    TrashedFilesAPI       *TrashedFilesAPIService
    TrashedFoldersAPI     *TrashedFoldersAPIService
    TrashedItemsAPI       *TrashedItemsAPIService
    UploadsAPI            *UploadsAPIService
    UploadsChunkedAPI     *UploadsChunkedAPIService
    UsersAPI              *UsersAPIService
    // ... 55 more services
}
```

---

## 5. Available API Services (73 total)

### Services relevant to our tool:

| Service | File | Description |
|---------|------|-------------|
| `FilesAPI` | `api_files.go` | File CRUD, copy, move, update |
| `FoldersAPI` | `api_folders.go` | Folder CRUD, list items, copy |
| `UploadsAPI` | `api_uploads.go` | Simple file upload |
| `UploadsChunkedAPI` | `api_uploads_chunked.go` | Chunked upload sessions |
| `DownloadsAPI` | `api_downloads.go` | File download |
| `CollaborationsAPI` | `api_collaborations.go` | Collaboration CRUD |
| `CollaborationsListAPI` | `api_collaborations_list.go` | List collaborations |
| `CommentsAPI` | `api_comments.go` | Comment CRUD |
| `SearchAPI` | `api_search.go` | Content search |
| `UsersAPI` | `api_users.go` | User CRUD |
| `GroupsAPI` | `api_groups.go` | Group CRUD |
| `GroupMembershipsAPI` | `api_group_memberships.go` | Group membership management |
| `TrashedItemsAPI` | `api_trashed_items.go` | List trashed items |
| `TrashedFilesAPI` | `api_trashed_files.go` | Get/restore/delete trashed files |
| `TrashedFoldersAPI` | `api_trashed_folders.go` | Get/restore/delete trashed folders |
| `SharedLinksFilesAPI` | `api_shared_links_files.go` | Shared links on files |
| `SharedLinksFoldersAPI` | `api_shared_links_folders.go` | Shared links on folders |

### All 73 services:

```
AIAPI, AIStudioAPI, AppItemAssociationsAPI, AuthorizationAPI,
BoxSignRequestsAPI, BoxSignTemplatesAPI,
ClassificationsAPI, ClassificationsOnFilesAPI, ClassificationsOnFoldersAPI,
CollaborationsAPI, CollaborationsListAPI, CollectionsAPI, CommentsAPI,
DevicePinnersAPI,
DomainRestrictionsForCollaborationsAPI, DomainRestrictionsUserExemptionsAPI,
DownloadsAPI, EmailAliasesAPI, EventsAPI,
FileRequestsAPI, FileVersionLegalHoldsAPI, FileVersionRetentionsAPI,
FileVersionsAPI, FilesAPI, FolderLocksAPI, FoldersAPI,
GroupMembershipsAPI, GroupsAPI,
IntegrationMappingsAPI, InvitesAPI,
LegalHoldPoliciesAPI, LegalHoldPolicyAssignmentsAPI,
MetadataCascadePoliciesAPI, MetadataInstancesFilesAPI,
MetadataInstancesFoldersAPI, MetadataTaxonomiesAPI, MetadataTemplatesAPI,
RecentItemsAPI,
RetentionPoliciesAPI, RetentionPolicyAssignmentsAPI,
SearchAPI, SessionTerminationAPI,
SharedLinksAppItemsAPI, SharedLinksFilesAPI, SharedLinksFoldersAPI,
SharedLinksWebLinksAPI,
ShieldInformationBarrierReportsAPI, ShieldInformationBarrierSegmentMembersAPI,
ShieldInformationBarrierSegmentRestrictionsAPI, ShieldInformationBarrierSegmentsAPI,
ShieldInformationBarriersAPI,
SkillsAPI,
StandardAndZonesStoragePoliciesAPI, StandardAndZonesStoragePolicyAssignmentsAPI,
TaskAssignmentsAPI, TasksAPI,
TermsOfServiceAPI, TermsOfServiceUserStatusesAPI,
TransferFoldersAPI,
TrashedFilesAPI, TrashedFoldersAPI, TrashedItemsAPI, TrashedWebLinksAPI,
UploadsAPI, UploadsChunkedAPI,
UserAvatarsAPI, UsersAPI,
WatermarksFilesAPI, WatermarksFoldersAPI,
WebLinksAPI, WebhooksAPI, WorkflowsAPI, ZipDownloadsAPI
```

---

## 6. Key Model Structs

### FileFull

```go
// model_file__full.go
type FileFull struct {
    Id                string                          `json:"id"`
    Etag              NullableString                  `json:"etag,omitempty"`
    Type              string                          `json:"type"`
    SequenceId        NullableString                  `json:"sequence_id,omitempty"`
    Name              *string                         `json:"name,omitempty"`
    Sha1              *string                         `json:"sha1,omitempty"`
    FileVersion       *FileVersionMini                `json:"file_version,omitempty"`
    Description       *string                         `json:"description,omitempty"`
    Size              *int32                          `json:"size,omitempty"`
    PathCollection    *FileAllOfPathCollection         `json:"path_collection,omitempty"`
    CreatedAt         *time.Time                      `json:"created_at,omitempty"`
    ModifiedAt        *time.Time                      `json:"modified_at,omitempty"`
    TrashedAt         NullableTime                    `json:"trashed_at,omitempty"`
    PurgedAt          NullableTime                    `json:"purged_at,omitempty"`
    ContentCreatedAt  NullableTime                    `json:"content_created_at,omitempty"`
    ContentModifiedAt NullableTime                    `json:"content_modified_at,omitempty"`
    CreatedBy         *UserMini                       `json:"created_by,omitempty"`
    ModifiedBy        *UserMini                       `json:"modified_by,omitempty"`
    OwnedBy           *UserMini                       `json:"owned_by,omitempty"`
    SharedLink        NullableFileAllOfSharedLink      `json:"shared_link,omitempty"`
    Parent            NullableFolderMini              `json:"parent,omitempty"`
    ItemStatus        *string                         `json:"item_status,omitempty"`
    VersionNumber     *string                         `json:"version_number,omitempty"`
    CommentCount      *int32                          `json:"comment_count,omitempty"`
    Permissions       *FileFullAllOfPermissions        `json:"permissions,omitempty"`
    Tags              []string                        `json:"tags,omitempty"`
    Lock              NullableFileFullAllOfLock        `json:"lock,omitempty"`
    Extension         *string                         `json:"extension,omitempty"`
    IsPackage         *bool                           `json:"is_package,omitempty"`
    // ... more fields (representations, metadata, etc.)
}
```

### UserFull

```go
// model_user__full.go
type UserFull struct {
    Id                            string                              `json:"id"`
    Type                          string                              `json:"type"`
    Name                          *string                             `json:"name,omitempty"`
    Login                         *string                             `json:"login,omitempty"`
    CreatedAt                     *time.Time                          `json:"created_at,omitempty"`
    ModifiedAt                    *time.Time                          `json:"modified_at,omitempty"`
    Language                      *string                             `json:"language,omitempty"`
    Timezone                      *string                             `json:"timezone,omitempty"`
    SpaceAmount                   *int64                              `json:"space_amount,omitempty"`
    SpaceUsed                     *int64                              `json:"space_used,omitempty"`
    MaxUploadSize                 *int64                              `json:"max_upload_size,omitempty"`
    Status                        *string                             `json:"status,omitempty"`
    JobTitle                      *string                             `json:"job_title,omitempty"`
    Phone                         *string                             `json:"phone,omitempty"`
    Address                       *string                             `json:"address,omitempty"`
    AvatarUrl                     *string                             `json:"avatar_url,omitempty"`
    NotificationEmail             NullableUserAllOfNotificationEmail  `json:"notification_email,omitempty"`
    Role                          *string                             `json:"role,omitempty"`
    TrackingCodes                 []TrackingCode                      `json:"tracking_codes,omitempty"`
    CanSeeManagedUsers            *bool                               `json:"can_see_managed_users,omitempty"`
    IsSyncEnabled                 *bool                               `json:"is_sync_enabled,omitempty"`
    IsExternalCollabRestricted    *bool                               `json:"is_external_collab_restricted,omitempty"`
    IsExemptFromDeviceLimits      *bool                               `json:"is_exempt_from_device_limits,omitempty"`
    IsExemptFromLoginVerification *bool                               `json:"is_exempt_from_login_verification,omitempty"`
    Enterprise                    *UserFullAllOfEnterprise            `json:"enterprise,omitempty"`
    MyTags                        []string                            `json:"my_tags,omitempty"`
    Hostname                      *string                             `json:"hostname,omitempty"`
    IsPlatformAccessOnly          *bool                               `json:"is_platform_access_only,omitempty"`
    ExternalAppUserId             *string                             `json:"external_app_user_id,omitempty"`
}
```

### Collaboration

```go
// model_collaboration.go
type Collaboration struct {
    Id                           string                                      `json:"id"`
    Type                         string                                      `json:"type"`
    Item                         NullableCollaborationItem                   `json:"item,omitempty"`
    AppItem                      NullableAppItem                             `json:"app_item,omitempty"`
    AccessibleBy                 *CollaborationAccessGrantee                 `json:"accessible_by,omitempty"`
    InviteEmail                  NullableString                              `json:"invite_email,omitempty"`
    Role                         *string                                     `json:"role,omitempty"`
    ExpiresAt                    NullableTime                                `json:"expires_at,omitempty"`
    IsAccessOnly                 *bool                                       `json:"is_access_only,omitempty"`
    Status                       *string                                     `json:"status,omitempty"`
    AcknowledgedAt               *time.Time                                  `json:"acknowledged_at,omitempty"`
    CreatedBy                    *UserCollaborations                         `json:"created_by,omitempty"`
    CreatedAt                    *time.Time                                  `json:"created_at,omitempty"`
    ModifiedAt                   *time.Time                                  `json:"modified_at,omitempty"`
    AcceptanceRequirementsStatus *CollaborationAcceptanceRequirementsStatus  `json:"acceptance_requirements_status,omitempty"`
}
```

### UploadSession

```go
// model_upload_session.go
type UploadSession struct {
    Id                *string                         `json:"id,omitempty"`
    Type              *string                         `json:"type,omitempty"`
    SessionExpiresAt  *time.Time                      `json:"session_expires_at,omitempty"`
    PartSize          *int64                          `json:"part_size,omitempty"`
    TotalParts        *int32                          `json:"total_parts,omitempty"`
    NumPartsProcessed *int32                          `json:"num_parts_processed,omitempty"`
    SessionEndpoints  *UploadSessionSessionEndpoints  `json:"session_endpoints,omitempty"`
}
```

### SearchAPI Request Builder

```go
// api_search.go
type ApiGetSearchRequest struct {
    ctx                      context.Context
    ApiService               *SearchAPIService
    query                    *string
    scope                    *string
    fileExtensions           *[]string
    createdAtRange           *[]string
    updatedAtRange           *[]string
    sizeRange                *[]int32
    ownerUserIds             *[]string
    recentUpdaterUserIds     *[]string
    ancestorFolderIds        *[]string
    contentTypes             *[]string
    type_                    *string
    trashContent             *string
    mdfilters                *[]MetadataFilter
    sort                     *string
    direction                *string
    limit                    *int64
    includeRecentSharedLinks *bool
    fields                   *[]string
    offset                   *int64
    deletedUserIds           *[]string
    deletedAtRange           *[]string
}

// Builder pattern methods:
func (r ApiGetSearchRequest) Query(query string) ApiGetSearchRequest { ... }
func (r ApiGetSearchRequest) Scope(scope string) ApiGetSearchRequest { ... }
func (r ApiGetSearchRequest) FileExtensions(ext []string) ApiGetSearchRequest { ... }
func (r ApiGetSearchRequest) Type_(t string) ApiGetSearchRequest { ... }
func (r ApiGetSearchRequest) AncestorFolderIds(ids []string) ApiGetSearchRequest { ... }
func (r ApiGetSearchRequest) Limit(limit int64) ApiGetSearchRequest { ... }
func (r ApiGetSearchRequest) Offset(offset int64) ApiGetSearchRequest { ... }
func (r ApiGetSearchRequest) Execute() (*SearchResultsOrSearchResultsWithSharedLinks, *http.Response, error) { ... }
```

---

## 7. Usage Examples

### Initialize client with OAuth token

```go
import (
    "context"
    boxapi "github.com/your/module/box-go-generated"
    "golang.org/x/oauth2"
)

// Create the token source (reuse ANBU's auth code)
tokenSource := oauthConfig.TokenSource(ctx, token)

// Create API client
cfg := boxapi.NewConfiguration()
apiClient := boxapi.NewAPIClient(cfg)

// Inject OAuth into context
ctx = context.WithValue(context.Background(), boxapi.ContextOAuth2, tokenSource)
```

### List folder items

```go
items, resp, err := apiClient.FoldersAPI.
    GetFoldersIdItems(ctx, "0").  // "0" = root
    Fields([]string{"type", "name", "id", "size", "modified_at"}).
    Limit(1000).
    Offset(0).
    Execute()
```

### Create collaboration

```go
collab, resp, err := apiClient.CollaborationsAPI.
    PostCollaborations(ctx).
    PostCollaborationsRequest(boxapi.PostCollaborationsRequest{
        Item: boxapi.PostCollaborationsRequestItem{
            Type: boxapi.PtrString("folder"),
            Id:   boxapi.PtrString("12345"),
        },
        AccessibleBy: boxapi.PostCollaborationsRequestAccessibleBy{
            Type:  "user",
            Login: boxapi.PtrString("user@example.com"),
        },
        Role: "viewer",
    }).
    Execute()
```

### Search

```go
results, resp, err := apiClient.SearchAPI.
    GetSearch(ctx).
    Query("quarterly report").
    Type_("file").
    FileExtensions([]string{"pdf", "docx"}).
    Limit(50).
    Execute()
```

### Create chunked upload session

```go
session, resp, err := apiClient.UploadsChunkedAPI.
    PostFilesUploadSessions(ctx).
    PostFilesUploadSessionsRequest(boxapi.PostFilesUploadSessionsRequest{
        FolderId: "12345",
        FileSize: 104857600,  // 100MB
        FileName: "large-file.zip",
    }).
    Execute()
// Then use session.PartSize to chunk the file and upload parts
```

### List users

```go
users, resp, err := apiClient.UsersAPI.
    GetUsers(ctx).
    FilterTerm("john").
    UserType("managed").
    Limit(100).
    Execute()
```

### Trash operations

```go
// List trashed items
trashed, resp, err := apiClient.TrashedItemsAPI.
    GetFoldersTrashItems(ctx).
    Limit(100).
    Execute()

// Restore a file from trash
restored, resp, err := apiClient.TrashedFilesAPI.
    PostFilesId(ctx, "67890").
    PostFilesIdRequest(boxapi.PostFilesIdRequest{
        Name:   boxapi.PtrString("restored-file.pdf"),
        Parent: &boxapi.PostFilesIdRequestParent{Id: boxapi.PtrString("0")},
    }).
    Execute()

// Permanently delete
resp, err := apiClient.TrashedFilesAPI.
    DeleteFilesIdTrash(ctx, "67890").
    Execute()
```

---

## 8. How to Integrate with Custom Auth

The generated client already supports OAuth2 natively via context injection. To integrate with ANBU-style auth:

```go
package main

import (
    "context"
    boxapi "github.com/your/module/generated"
    "golang.org/x/oauth2"
)

func main() {
    // 1. Reuse ANBU's GetBoxClient logic to get an OAuth token
    token, err := getBoxOAuthToken(oauthConfig) // from auth.go

    // 2. Create token source
    tokenSource := oauthConfig.TokenSource(context.Background(), token)

    // 3. Create generated API client
    cfg := boxapi.NewConfiguration()
    client := boxapi.NewAPIClient(cfg)

    // 4. Set OAuth context
    ctx := context.WithValue(context.Background(), boxapi.ContextOAuth2, tokenSource)

    // 5. Use any API service
    folders, _, _ := client.FoldersAPI.GetFoldersIdItems(ctx, "0").Execute()
}
```

**Alternative:** Pass the oauth2 HTTP client directly:

```go
cfg := boxapi.NewConfiguration()
cfg.HTTPClient = oauth2.NewClient(ctx, tokenSource)  // from ANBU's GetBoxClient
client := boxapi.NewAPIClient(cfg)
// Now all requests automatically include the Bearer token
```

---

## 9. Pros and Cons

### Pros

- **Complete API coverage** — 73 API services, 565 model types, every Box endpoint
- **Typed structs** — no manual struct writing for request/response types
- **Builder pattern** — fluent API for setting query params, headers, body
- **OAuth2 built-in** — natively supports `golang.org/x/oauth2`
- **Always up-to-date** — regenerate from latest spec anytime Box updates their API
- **Documentation** — auto-generated docs in `docs/` directory

### Cons

- **Verbose** — 642 Go files, much of which you'll never use
- **Generated code style** — heavy use of pointers, nullable types, getter/setter boilerplate
- **Large binary impact** — unused code gets compiled in (unless you cherry-pick files)
- **Less control** — error handling, retry logic, and HTTP behavior are harder to customize
- **Module path** — defaults to `github.com/GIT_USER_ID/GIT_REPO_ID`, needs updating

### Recommendation

**Best approach for the new tool:**

1. **Cherry-pick model files** — Copy only the `model_*.go` files you need (File, Folder, User, Group, Collaboration, Comment, UploadSession, etc.) into your project. This gives you typed structs without the bloat.

2. **Write your own HTTP client** — Use ANBU's proven patterns (`net/http` + `oauth2`) for the actual API calls. This gives you full control over error handling, retries, and rate limiting.

3. **Reference the generated API files** — Use `api_*.go` as documentation for the correct URL paths, query params, and request body structures. But write simpler Go code instead of using the generated client directly.

4. **Regenerate as needed** — When Box adds new API endpoints, re-run the generator and cherry-pick updated structs.

This hybrid approach gives you the type safety of the generated code with the simplicity and control of hand-written HTTP calls.
