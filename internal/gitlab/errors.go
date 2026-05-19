package gitlab

import (
	"errors"
	"fmt"

	glab "gitlab.com/gitlab-org/api/client-go"
)

var (
	ErrClientNotConfigured = errors.New("gitlab: client is not configured")
	ErrEmptyResponse       = errors.New("gitlab: empty response from server")
	ErrNoResult            = errors.New("gitlab: no result")
)

type ErrUnexpectedResponse struct {
	Status     string
	StatusCode int
	Body       string
}

func (e *ErrUnexpectedResponse) Error() string {
	return fmt.Sprintf("gitlab: unexpected response %s: %s", e.Status, e.Body)
}

func normalizeError(err error) error {
	if err == nil {
		return nil
	}

	var responseErr *glab.ErrorResponse
	if errors.As(err, &responseErr) && responseErr.Response != nil {
		return &ErrUnexpectedResponse{
			Status:     responseErr.Response.Status,
			StatusCode: responseErr.Response.StatusCode,
			Body:       string(responseErr.Body),
		}
	}

	return err
}
