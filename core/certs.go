// Copyright © 2020 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
)

func DumpCerts(config *ServerConfig) {
	serverCertFile := filepath.Join(config.CertPath, fmt.Sprintf("%s.crt", config.Name))
	serverKeyFile := filepath.Join(config.CertPath, fmt.Sprintf("%s.key", config.Name))
	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		fmt.Printf("Failed to access server certificates: %v\n", err)
		return
	}
	if len(serverCert.Certificate) == 0 {
		fmt.Printf("Certificate file %q does not have expected information\n", serverCertFile)
		return
	}
	cert, err := x509.ParseCertificate(serverCert.Certificate[0])
	if err != nil {
		fmt.Printf("Could not read certificate: %v\n", err)
		return
	}

	fmt.Printf("Server certificate issued by: %s\n", cert.Issuer.CommonName)
	if cert.NotAfter.Before(time.Now()) {
		fmt.Printf("WARNING: server certificate expired at: %v\n", cert.NotAfter)
	} else {
		fmt.Printf("Server certificate expires: %v\n", cert.NotAfter)
	}
	fmt.Printf("Server certificate issued to: %s\n", cert.Subject.CommonName)

	caCertFile := filepath.Join(config.CertPath, "ca.crt")
	caCertData, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		fmt.Printf("Failed to access certificate authority certificate: %v\n", err)
		return
	}
	fmt.Println("")

	for len(caCertData) > 0 {
		var block *pem.Block
		block, caCertData = pem.Decode(caCertData)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		fmt.Printf("Certificate authority certificate is: %s\n", cert.Subject.CommonName)
		if cert.NotAfter.Before(time.Now()) {
			fmt.Printf("WARNING: certificate authority certificate expired at: %v\n", cert.NotAfter)
		} else {
			fmt.Printf("Certificate authority certificate expires: %v\n", cert.NotAfter)
		}
	}
}
