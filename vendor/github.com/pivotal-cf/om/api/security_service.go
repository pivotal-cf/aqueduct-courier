package api

import (
	"encoding/json"
	"fmt"
)

type certResponse struct {
	Cert string `json:"root_ca_certificate_pem"`
}

func (a Api) GetSecurityRootCACertificate() (string, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/security/root_ca_certificate", nil)
	if err != nil {
		return "", fmt.Errorf("failed to submit request: %w", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return "", err
	}

	var certResponse certResponse
	if err := json.NewDecoder(resp.Body).Decode(&certResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return certResponse.Cert, nil
}
