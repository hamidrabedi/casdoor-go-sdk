package casdoorsdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

func (c *Client) WebAuthnSignupBegin() ([]byte, error) {
	url := c.GetUrl("webauthn/signup/begin", nil)

	respBytes, err := c.doGetBytesRawWithoutCheck(url)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(respBytes, &response); err == nil {
		if response.Status == "error" {
			return nil, errors.New(response.Msg)
		}

		if response.Status == "ok" && response.Data != nil {
			dataBytes, err := json.Marshal(response.Data)
			if err != nil {
				return nil, err
			}
			return dataBytes, nil
		}
	}

	return respBytes, nil
}

func (c *Client) WebAuthnSignupFinish(credentialCreationResponse []byte) error {
	url := c.GetUrl("webauthn/signup/finish", nil)

	contentType := "application/json"
	body := bytes.NewReader(credentialCreationResponse)

	respBytes, err := c.DoPostBytesRaw(url, contentType, body)
	if err != nil {
		return err
	}

	var response Response
	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		return err
	}

	if response.Status != "ok" {
		return fmt.Errorf(response.Msg)
	}

	return nil
}
