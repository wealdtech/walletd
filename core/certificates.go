package core

// CertificateInfo contains information related to client certificates
type CertificateInfo struct {
	Name     string   `json:"name"`
	Accounts []string `json:"accounts"`
}
