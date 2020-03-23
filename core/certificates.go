package core

// CertificateInfo contains information related to client certificates.
type CertificateInfo struct {
	Name        string                    `json:"name"`
	Permissions []*CertificatePermissions `json:"perms"`
}

// CertificatePermissions contains information about the operations allowed by the certificate.
type CertificatePermissions struct {
	Path       string   `json:"path"`
	Operations []string `json:"operations"`
}
