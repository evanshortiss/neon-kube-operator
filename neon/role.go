package neon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) GetFirstRole(ctx context.Context, projectId, branchId string) (string, error) {
	if branchId == "" {
		return "", ErrBranchNotFound
	}
	roles, err := c.GetRoles(ctx, projectId, branchId)
	if err != nil {
		return "", err
	}
	rolesObj, ok := roles["roles"].([]any)
	if !ok {
		return "", fmt.Errorf("incorrect body roles")
	}
	if len(rolesObj) == 0 {
		return "", fmt.Errorf("no role found")
	}
	role, ok := rolesObj[0].(map[string]any)
	if !ok {
		return "", fmt.Errorf("incorrect body role")
	}
	name, ok := role["name"].(string)
	if !ok {
		return "", fmt.Errorf("no name of role")
	}
	return name, nil
}

func (c *Client) GetRoles(ctx context.Context, projectId, branchId string) (map[string]any, error) {
	if branchId == "" {
		return nil, ErrBranchNotFound
	}
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s/roles", projectId, branchId)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get roles %s", resp.Status)
	}
	m := make(map[string]any)
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (c *Client) GetRolePassword(ctx context.Context, projectId, branchId, role string) (string, error) {
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s/roles/%s/reveal_password", projectId, branchId, role)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get roles %s", resp.Status)
	}
	m := make(map[string]any)
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return "", err
	}
	password, ok := m["password"].(string)
	if !ok {
		return "", fmt.Errorf("no password in body")
	}
	return password, nil
}
