package resource

import (
	"errors"
	"fmt"
	"math"
	"terraform-provider-eci/internal/api"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func addResourceError(
	diags *diag.Diagnostics,
	summary string,
	resourceId string,
	err error,
) {
	var detail string
	if resourceId == "" {
		detail = fmt.Sprintf("reason: %s", err.Error())
	} else {
		detail = fmt.Sprintf("reason: %s (resource id: %s)", err.Error(), resourceId)
	}

	(*diags).AddError(summary, detail)
}

func isResourceDeleted(err error, resourceKey string, deletedStatus string) (string, error) {
	if err == nil {
		return "successfully deleted", nil
	}

	var apiError *api.APIError
	if !errors.As(err, &apiError) {
		return "", err
	}

	switch apiError.HttpCode {
	case 404:
		return "resource does not exist", nil

	case 409:
		if apiError.IsCode("unexpected_status") &&
			apiError.Detail != nil {
			resource, resourceExists := (*apiError.Detail)[resourceKey]

			if resourceExists {
				if resource.(map[string]interface{})["status"] == deletedStatus {
					return "resource is already deleted", nil
				}
			}
		}
	}

	return "", err
}

func waitStatus(
	getStatus func() (*string, error), targetStatuses []string, maxRetry int,
) (*string, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	for retryIndex := 0; retryIndex < maxRetry; retryIndex++ {
		status, err := getStatus()

		if err != nil {
			diags.Append(diag.NewWarningDiagnostic(
				"failed to get status",
				fmt.Sprintf("retry: %d (err: %s)", retryIndex, err.Error()),
			))
			continue
		}

		for _, targetStatus := range targetStatuses {
			if targetStatus == *status {
				return &targetStatus, diags
			}
		}

		time.Sleep(time.Duration(min(0.5+math.Pow(2, float64(retryIndex)), 15)) * time.Second)
	}

	diags.Append(diag.NewErrorDiagnostic("unexpected status", "reached maximum retry"))
	return nil, diags
}
