package api

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type APIClient struct {
	restyClient    *resty.Client
	token          string
	baseURL        string
	pathPrefix     string
	OrganizationId string
	zoneId         string
}

var _ error = &APIError{}

type APIError struct {
	HttpCode int
	Code     *string
	Message  *string
	Detail   *map[string]interface{}
}

func (e *APIError) IsCode(code string) bool {
	return e.Code != nil && *e.Code == code
}

func getValue[T any](p *T) any {
	if p != nil {
		return *p
	}
	return "<nil>"
}

func (e *APIError) Error() string {
	return fmt.Sprintf(
		"code: %v, message: %v, http_code: %d, detail: %v",
		getValue(e.Code), getValue(e.Message), e.HttpCode, getValue(e.Detail),
	)
}

func NewAPIClient(
	token string,
	baseURL string,
	pathPrefix string,
	zoneId string,
	debug bool,
) (*APIClient, error) {
	client := resty.New().
		SetDebug(debug).
		SetBaseURL(baseURL).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", token))

	result, err := handleAPIResponse[OrganizationGetResponse](
		client.R().
			SetResult(&OrganizationGetResponse{}).
			Get(fmt.Sprintf("%s/user/organization", pathPrefix)),
	)

	if err != nil {
		return nil, err
	}

	return &APIClient{
		restyClient:    client,
		token:          token,
		baseURL:        baseURL,
		pathPrefix:     pathPrefix,
		OrganizationId: result.Id.String(),
		zoneId:         zoneId,
	}, nil
}

func makeAPIError(resp *resty.Response) (*APIError, error) {
	var data map[string]interface{}

	err := json.Unmarshal(resp.Body(), &data)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse error message: %s, http status: %s (%w) ",
			resp.String(),
			resp.Status(),
			err,
		)
	}

	code, _ := data["code"].(string)
	message, _ := data["message"].(string)
	maybeDetail, detailExists := data["detail"]

	if detailExists {
		switch detail := maybeDetail.(type) {
		case map[string]interface{}:
			return nil, &APIError{
				HttpCode: resp.StatusCode(),
				Code:     &code,
				Message:  &message,
				Detail:   &detail,
			}
		}
	}

	return nil, &APIError{
		HttpCode: resp.StatusCode(),
		Code:     &code,
		Message:  &message,
		Detail:   nil,
	}
}

func detectError(resp *resty.Response, err error) error {
	if err != nil {
		return err
	}

	if resp.StatusCode() == 200 {
		return nil
	}

	apiError, err := makeAPIError(resp)
	if err != nil {
		return err
	}

	return apiError
}

func handleAPIResponse[T any](resp *resty.Response, err error) (*T, error) {
	if err = detectError(resp, err); err != nil {
		return nil, err
	}

	return resp.Result().(*T), nil
}

func handleListAPIResponse[T any](resp *resty.Response, err error) ([]T, error) {
	if err = detectError(resp, err); err != nil {
		return nil, err
	}

	return *resp.Result().(*[]T), nil
}

func setIfNotNil[T any](m map[string]interface{}, key string, ptr *T) {
	if ptr != nil {
		m[key] = *ptr
	}
}

func setStrIfNotNil(m map[string]string, key string, ptr *string) {
	if ptr != nil {
		m[key] = *ptr
	}
}
