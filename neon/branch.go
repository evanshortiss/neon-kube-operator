package neon

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	neontechv1alpha1 "github.com/evanshortiss/neon-kube-operator/api/v1alpha1"
)

func branchSpecToCreateRequestBody(b *neontechv1alpha1.Branch) map[string]any {
	body := make(map[string]any)
	branchSpec := b.Spec
	branch := make(map[string]any)
	branch["name"] = b.Name

	if branchSpec.ParentId != nil {
		branch["parent_id"] = branchSpec.ParentId
	}

	if branchSpec.ParentStartPoint != nil {
		if branchSpec.ParentStartPoint.Lsn != nil {
			branch["parent_lsn"] = branchSpec.ParentStartPoint.Lsn
		}

		if branchSpec.ParentStartPoint.Timestamp != nil {
			branch["parent_timestamp"] = branchSpec.ParentStartPoint.Timestamp
		}
	}

	body["branch"] = branch

	return body
}

func (c *Client) CreateBranch(ctx context.Context, branch *neontechv1alpha1.Branch) (map[string]any, error) {
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches", branch.Spec.ProjectId)

	reqData, err := json.Marshal(branchSpecToCreateRequestBody(branch))
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqData))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 201 { // TODO: add already exists
		return nil, fmt.Errorf("failed to create branch: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	m := make(map[string]any)
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (c *Client) DeleteBranch(ctx context.Context, branch *neontechv1alpha1.Branch) (map[string]any, error) {
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s", branch.Spec.ProjectId, branch.Status.Id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
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
	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		return nil, fmt.Errorf("failed to create branch: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	m := make(map[string]any)
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

var (
	ErrBranchNotFound = errors.New("branch not found")
)

func (c *Client) GetBranch(ctx context.Context, name string, branch *neontechv1alpha1.Branch) (map[string]any, error) {
	if branch.Status.Id == "" {
		return nil, ErrBranchNotFound
	}
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s", branch.Spec.ProjectId, branch.Status.Id)
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
		if resp.StatusCode == 404 {
			return nil, ErrBranchNotFound
		}
		return nil, fmt.Errorf("failed to get branch %s", resp.Status)
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

func NewBranchStatus(response map[string]any) neontechv1alpha1.BranchStatus {
	var branchStatus neontechv1alpha1.BranchStatus
	if branch, ok := response["branch"].(map[string]any); ok {
		branchStatus.Id, _ = branch["id"].(string)
		branchStatus.Name, _ = branch["name"].(string)
		branchStatus.ProjectId, _ = branch["project_id"].(string)
		branchStatus.ParentId, _ = branch["parent_id"].(string)
		branchStatus.ParentLsn, _ = branch["parent_lsn"].(string)
		branchStatus.Primary, _ = branch["primary"].(bool)
		branchStatus.CreatedAt, _ = branch["created_at"].(string)
		branchStatus.UpdatedAt, _ = branch["updated_at"].(string)
	}

	return branchStatus
}

func NewEndpointStatus(response map[string]any) neontechv1alpha1.EndpointStatus {
	var es neontechv1alpha1.EndpointStatus
	if branch, ok := response["endpoint"].(map[string]any); ok {
		es.Id, _ = branch["id"].(string)
		es.CurrentState, _ = branch["current_state"].(string)
		es.PendingState, _ = branch["pending_state"].(string)
		es.Host, _ = branch["host"].(string)
		es.CreatedAt, _ = branch["created_at"].(string)
		es.UpdatedAt, _ = branch["updated_at"].(string)
	}

	return es
}
