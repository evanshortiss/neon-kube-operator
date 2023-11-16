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

func branchSpecToCreateRequestBody(branchSpec *neontechv1alpha1.BranchSpec) map[string]any {
	body := make(map[string]any)

	if branchSpec.ParentId != nil || branchSpec.ParentStartPoint != nil || branchSpec.ParentStartPoint.Lsn != nil || branchSpec.ParentStartPoint.Timestamp != nil {
		branch := make(map[string]any)

		if branchSpec.ParentId != nil {
			branch["parent_id"] = branchSpec.ParentId
		}

		// TODO: name????

		if branchSpec.ParentStartPoint.Lsn != nil {
			branch["parent_lsn"] = branchSpec.ParentStartPoint.Lsn
		}

		if branchSpec.ParentStartPoint.Timestamp != nil {
			branch["parent_timestamp"] = branchSpec.ParentStartPoint.Timestamp
		}

		body["branch"] = branch
	}

	return body
}

func (c *Client) CreateBranch(ctx context.Context, branchSpec *neontechv1alpha1.BranchSpec) (map[string]any, error) {
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches", branchSpec.ProjectId)

	reqData, err := json.Marshal(branchSpecToCreateRequestBody(branchSpec))
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
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("Failed to create branch: %s", resp.Status)
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
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s", branch.Spec.ProjectId, branch.Name)
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
		return nil, fmt.Errorf("Failed to create branch: %s", resp.Status)
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
	BranchNotFound = errors.New("Branch not found")
)

func (c *Client) GetBranch(ctx context.Context, name string, spec *neontechv1alpha1.BranchSpec) (map[string]any, error) {
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s", spec.ProjectId, name)
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
			return nil, BranchNotFound
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
