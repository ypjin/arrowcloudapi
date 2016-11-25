package models

// TopRepo holds information about repository that accessed most
type TopRepo struct {
	RepoName    string `json:"name"`
	AccessCount int64  `json:"count"`
	//	Creator     string `json:"creator"`
}
