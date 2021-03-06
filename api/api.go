package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/mobingilabs/go-modaemon/config"
	"github.com/mobingilabs/go-modaemon/server_config"
)

type client struct {
	client    *http.Client
	config    *config.Config
	tokenType string
	token     string
}

type StsToken struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

func NewClient(conf *config.Config) (*client, error) {
	c := &client{
		config: conf,
		client: &http.Client{},
	}

	err := c.getAccessToken()

	return c, err
}

func (c *client) GetServerConfig() (*serverConfig.Config, error) {
	values := url.Values{}
	values.Set("stack_id", c.config.StackID)

	res, err := c.get("/v2/alm/serverconfig", values)
	if err != nil {
		return nil, err
	}

	conf := &serverConfig.Config{}
	err = json.Unmarshal(res, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func (c *client) GetStsToken() (*StsToken, error) {
	values := url.Values{}
	values.Set("user_id", c.config.UserID)
	values.Set("stack_id", c.config.StackID)

	res, err := c.get("/v2/alm/sts", values)
	if err != nil {
		return nil, err
	}

	stsToken := &StsToken{}
	err = json.Unmarshal(res, stsToken)
	if err != nil {
		return nil, err
	}

	return stsToken, nil
}

func (c *client) SendInstanceStatus(serverID, status string) error {
	values := url.Values{}
	values.Set("instance_id", serverID)
	values.Set("stack_id", c.config.StackID)
	values.Set("status", status)

	_, err := c.post("/v2/alm/instance/status", values)
	return err
}

func (c *client) get(path string, values url.Values) ([]byte, error) {
	req, err := http.NewRequest("GET", c.config.APIHost+path, nil)
	if c.token != "" && c.tokenType != "" {
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", c.tokenType, c.token))
	}

	req.URL.RawQuery = values.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return res, errors.New(resp.Status)
	} else {
		return res, nil
	}
}

func (c *client) post(path string, values url.Values) ([]byte, error) {
	req, err := http.NewRequest("POST", c.config.APIHost+path, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.token != "" && c.tokenType != "" {
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", c.tokenType, c.token))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return res, errors.New(resp.Status)
	} else {
		return res, nil
	}
}

func (c *client) getAccessToken() error {
	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", c.config.StackID)
	values.Set("client_secret", c.config.AuthorizationToken)

	res, err := c.post("/v2/access_token", values)
	if err != nil {
		return err
	}

	var tokenInfo map[string]interface{}

	err = json.Unmarshal(res, &tokenInfo)
	if err != nil {
		return err
	}

	c.tokenType = tokenInfo["token_type"].(string)
	c.token = tokenInfo["access_token"].(string)

	return nil
}
