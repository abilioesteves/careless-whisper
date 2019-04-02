package misc

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/abilioesteves/goh/gohtypes"
	"github.com/sirupsen/logrus"

	"github.com/abilioesteves/goh/gohclient"
)

// HydraClient holds data and methods to communicate with an hydra service instance
type HydraClient struct {
	BaseURL    *url.URL
	HTTPClient *gohclient.Default
}

// AcceptLoginRequestPayload holds the data to communicate with hydra's accept login api
type AcceptLoginRequestPayload struct {
	Subject     string `json:"subject"`
	Remember    bool   `json:"remember"`
	RememberFor int    `json:"remember_for"`
	ACR         string `json:"acr"`
}

// AcceptConsentRequestPayload holds the data to communicate with hydra's accept consent api
type AcceptConsentRequestPayload struct {
	GrantScope               []string            `json:"grant_scope"`
	GrantAccessTokenAudience []string            `json:"grant_access_token_audience"`
	Remember                 bool                `json:"remember"`
	RememberFor              int                 `json:"remember_for"`
	Session                  TokenSessionPayload `json:"session"`
}

// TokenSessionPayload holds additional data to be carried with the created token
type TokenSessionPayload struct {
	IDToken     interface{} `json:"id_token"`
	AccessToken interface{} `json:"access_token"`
}

// RejectConsentRequestPayload holds the data to communicate with hydra's reject consent api
type RejectConsentRequestPayload struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// Init initializes a hydra client
func (hydra *HydraClient) Init(hydraEndpoint string) *HydraClient {
	hydra.BaseURL, _ = url.Parse(hydraEndpoint)
	hydra.BaseURL.Path = path.Join(hydra.BaseURL.Path, "/oauth2/auth/requests/")
	hydra.HTTPClient = gohclient.New("application/json", "application/json")

	logrus.Infof("Hydra enpoint url: %v", hydra.BaseURL.String())
	return hydra
}

// GetLoginRequestInfo retrieves information to drive decisions over how to deal with the login request
func (hydra *HydraClient) GetLoginRequestInfo(challenge string) map[string]interface{} {
	return hydra.get("login", challenge)
}

// AcceptLoginRequest sends an accept login request to hydra
func (hydra *HydraClient) AcceptLoginRequest(challenge string, payload AcceptLoginRequestPayload) map[string]interface{} {
	data, _ := json.Marshal(&payload)
	return hydra.put("login", challenge, "accept", data)
}

// GetConsentRequestInfo retrieves information to drive decisions over how to deal with the consent request
func (hydra *HydraClient) GetConsentRequestInfo(challenge string) map[string]interface{} {
	return hydra.get("consent", challenge)
}

// AcceptConsentRequest sends an accept login request to hydra
func (hydra *HydraClient) AcceptConsentRequest(challenge string, payload AcceptConsentRequestPayload) map[string]interface{} {
	data, _ := json.Marshal(&payload)
	return hydra.put("consent", challenge, "accept", data)
}

// RejectConsentRequest sends a reject login request to hydra
func (hydra *HydraClient) RejectConsentRequest(challenge string, payload RejectConsentRequestPayload) map[string]interface{} {
	data, _ := json.Marshal(&payload)
	return hydra.put("consent", challenge, "reject", data)
}

func (hydra *HydraClient) get(flow, challenge string) map[string]interface{} {
	u, _ := url.Parse(hydra.BaseURL.String())
	u.Path = path.Join(u.Path, flow, url.QueryEscape(challenge))
	logrus.Debugf("url: '%v'", u.String())
	return hydra.treatResponse(hydra.HTTPClient.Get(u.String()))
}

func (hydra *HydraClient) put(flow, challenge, action string, data []byte) map[string]interface{} {
	u, _ := url.Parse(hydra.BaseURL.String())
	u.Path = path.Join(u.Path, flow, url.QueryEscape(challenge), action)
	logrus.Debugf("url: '%v'", u.String())
	return hydra.treatResponse(hydra.HTTPClient.Put(u.String(), data))
}

func (hydra *HydraClient) treatResponse(resp *http.Response, data []byte, err error) map[string]interface{} {
	if err == nil {
		if resp.StatusCode >= 200 && resp.StatusCode <= 302 {
			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err == nil {
				return result
			}
			panic(gohtypes.Error{Code: 500, Err: err, Message: "Error while decoding hydra's response bytes"})
		}
	}
	panic(gohtypes.Error{Code: 500, Err: err, Message: "Error while communicating with Hydra"})
}