package login

import (
	"fmt"

	"github.com/meinside/steemit-go"
)

const (
	apiName = "login_api"
)

// Login
func Login(c *steemit.Client, username, passwd string) (success bool, err error) {
	req := c.NewRequest(apiName, "login", []interface{}{username, passwd})

	var res steemit.Response
	res, err = c.SendRequest(req)
	if err != nil {
		return false, err
	}

	if res.Result != nil {
		switch res.Result.(type) {
		case bool:
			return res.Result.(bool), nil
		default:
			return false, fmt.Errorf("Result is in unexpected type %T", res.Result)
		}
	} else {
		return false, fmt.Errorf("Failed to login (%s)", res.Error.Message)
	}
}

// Get API id by name
func GetApiByName(c *steemit.Client, name string) (apiId int, err error) {
	req := c.NewRequest(apiName, "get_api_by_name", []interface{}{name})

	var res steemit.Response
	res, err = c.SendRequest(req)
	if err != nil {
		return -1, err
	}

	if res.Result != nil {
		switch res.Result.(type) {
		case float64:
			return int(res.Result.(float64)), nil
		default:
			return -1, fmt.Errorf("Result is in unexpected type %T", res.Result)
		}
	} else {
		return -1, fmt.Errorf("No result for %s (%s)", name, res.Error.Message)
	}
}

// Get version of this API
func GetVersion(c *steemit.Client) (versions map[string]interface{}, err error) {
	req := c.NewRequest(apiName, "get_version", []interface{}{})

	var res steemit.Response
	res, err = c.SendRequest(req)
	if err != nil {
		return nil, err
	}

	if res.Result != nil {
		switch res.Result.(type) {
		case map[string]interface{}:
			return res.Result.(map[string]interface{}), nil
		default:
			return nil, fmt.Errorf("Result is in unexpected type %T", res.Result)
		}
	}

	return nil, fmt.Errorf("Failed to get version (%s)", res.Error.Message)
}
