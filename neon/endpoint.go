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
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RetryError error

var (
	ErrRetryAgain RetryError = errors.New("retry again")
)

func endpointSpecToCreateRequestBody(e *neontechv1alpha1.Endpoint, branchId, projectId string) map[string]any {
	body := make(map[string]any)
	endpoint := make(map[string]any)

	endpointSpec := e.Spec

	endpoint["branch_id"] = branchId
	endpoint["project_id"] = projectId
	endpoint["type"] = endpointSpec.Type
	if endpointSpec.RegionId != nil {
		endpoint["region_id"] = endpointSpec.RegionId
	}
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

func (c *Client) CreateEndpoint(ctx context.Context, k8sClient client.Client, e *neontechv1alpha1.Endpoint) (map[string]any, error) {
	branchId, projectId, err := getBranchProjectId(ctx, k8sClient, e)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/endpoints", projectId)

	reqData, err := json.Marshal(endpointSpecToCreateRequestBody(e, branchId, projectId))
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

func getBranchProjectId(ctx context.Context, k8sClient client.Client, e *neontechv1alpha1.Endpoint) (string, string, error) {
	var branchId, projectId string
	if e.Spec.BranchFrom.BranchRef != "" {
		branch := neontechv1alpha1.Branch{}
		err := k8sClient.Get(ctx, types.NamespacedName{Name: e.Spec.BranchFrom.BranchRef, Namespace: e.Namespace}, &branch)
		if kerrors.IsNotFound(err) {
			return "", "", fmt.Errorf("branch is not found yet, %w", ErrRetryAgain)
		}
		if err != nil {
			return "", "", err
		}
		if !branch.Status.State.Exists() {
			return "", "", fmt.Errorf("branch status is not updated yet, %w", ErrRetryAgain)
		}
		branchId = branch.Status.Id
		projectId = branch.Spec.ProjectId

	} else {
		branchId = e.Spec.BranchFrom.BranchId
		projectId = e.Spec.BranchFrom.ProjectId
	}

	return branchId, projectId, nil
}

func (c *Client) DeleteEndpoint(ctx context.Context, k8sClient client.Client, e *neontechv1alpha1.Endpoint) (map[string]any, error) {
	_, projectId, err := getBranchProjectId(ctx, k8sClient, e)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/endpoints/%s", projectId, e.Status.Id)
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
		return nil, fmt.Errorf("failed to delete endpoint: %s", resp.Status)
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

var ErrEndpointNotFound = errors.New("branch not found")

func (c *Client) GetEndpoint(ctx context.Context, k8sClient client.Client, e *neontechv1alpha1.Endpoint) (map[string]any, error) {
	if e.Status.Id == "" {
		return nil, ErrEndpointNotFound
	}
	_, projectId, err := getBranchProjectId(ctx, k8sClient, e)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/endpoints/%s", projectId, e.Status.Id)
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
			return nil, ErrEndpointNotFound
		}
		return nil, fmt.Errorf("failed to get endpoint: %s", resp.Status)
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
