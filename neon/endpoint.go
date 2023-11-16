package neon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	neontechv1alpha1 "github.com/evanshortiss/neon-kube-operator/api/v1alpha1"
)

func endpointSpecToCreateRequestBody(endpointSpec *neontechv1alpha1.EndpointSpec) map[string]any {
	body := make(map[string]any)
	endpoint := make(map[string]any)

	endpoint["branch_id"] = endpointSpec.BranchId
	if endpointSpec.RegionId != nil {
		endpoint["region_id"] = endpointSpec.RegionId
	}
	endpoint["type"] = endpointSpec.Type
	if endpointSpec.Settings != nil {
		endpoint["settings"] = endpointSpec.Settings
	}
	if endpointSpec.AutoscalingLimitMinCu != nil {
		endpoint["autoscaling_limit_min_cu"] = endpointSpec.AutoscalingLimitMinCu
	}
	if endpointSpec.AutoscalingLimitMaxCu != nil {
		endpoint["autoscaling_limit_max_cu"] = endpointSpec.AutoscalingLimitMaxCu
	}
	if endpointSpec.Provisioner != nil {
		endpoint["provisioner"] = endpointSpec.Provisioner
	}
	if endpointSpec.PoolerEnabled != nil {
		endpoint["pooler_enabled"] = endpointSpec.PoolerEnabled
	}
	if endpointSpec.PoolerMode != nil {
		endpoint["pooler_mode"] = endpointSpec.PoolerMode
	}
	if endpointSpec.Disabled != nil {
		endpoint["disabled"] = endpointSpec.Disabled
	}
	if endpointSpec.PasswordlessAccess != nil {
		endpoint["passwordless_access"] = endpointSpec.PasswordlessAccess
	}
	if endpointSpec.SuspendTimeoutSeconds != nil {
		endpoint["suspend_timeout_seconds"] = endpointSpec.SuspendTimeoutSeconds
	}

	body["endpoint"] = endpoint

	return body
}

func (c *Client) CreateEndpoint(ctx context.Context, endpointSpec *neontechv1alpha1.EndpointSpec) (map[string]any, error) {
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/endpoints", endpointSpec.ProjectId)

	reqData, err := json.Marshal(endpointSpec)
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
		return nil, fmt.Errorf("Failed to create endpoint: %s", resp.Status)
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

func (c *Client) DeleteEndpoint(ctx context.Context, endpoint *neontechv1alpha1.Endpoint) (map[string]any, error) {
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/endpoints/%s", endpoint.Spec.ProjectId, endpoint.Name)
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
		return nil, fmt.Errorf("Failed to delete endpoint: %s", resp.Status)
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

func (c *Client) GetEndpoint(ctx context.Context, name string, spec *neontechv1alpha1.EndpointSpec) (map[string]any, error) {
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/endpoints/%s", spec.ProjectId, name)
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
		return nil, fmt.Errorf("Failed to get endpoint: %s", resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	m := make(map[string]any)
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
