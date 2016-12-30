package azure

const azureAPI = "https://graph.azure.com/v2.8/"

// User is a Azure user.
//
// Note that user ids are unique to each app.
// ref: https://docs.microsoft.com/en-us/azure/active-directory/active-directory-v2-tokens#id-tokens
type User struct {
	TenantID          string `json:"tid"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	ID                string `json:"sub"`
	OrganizationID    int    `json:"oid"`
}
