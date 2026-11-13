package librarysvc

import (
	"context"
	"errors"
	"time"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/library"
)

// actionTimeout is the longest a source action may run before the host stops waiting.
const actionTimeout = 30 * time.Second

// errUnknownAction means the source does not declare the action id.
var errUnknownAction = errors.New("unknown action")

// runAction calls one action a source declares and returns the plugin's toast text and refresh flag.
func (s *service) runAction(ctx context.Context, in contract.ActionRequest) (contract.ActionResult, error) {
	plugin, action, err := findAction(s.sources.Sources(), in.Source, in.Action)
	if err != nil {
		return contract.ActionResult{}, err
	}
	var result contract.ActionResult
	if err := s.sources.CallAction(ctx, plugin, action.Method, in, &result); err != nil {
		return contract.ActionResult{}, err
	}
	return result, nil
}

// findAction returns the plugin that offers the source, and the source's action with the given id.
func findAction(sources []contract.SourceInfo, sourceID, actionID string) (string, contract.SourceAction, error) {
	info, ok := sourceNamed(sources, sourceID)
	if !ok {
		return "", contract.SourceAction{}, library.ErrUnknownSource
	}
	for _, action := range info.Actions {
		if action.ID == actionID {
			return info.Plugin, action, nil
		}
	}
	return "", contract.SourceAction{}, errUnknownAction
}

// sourceNamed returns the first source with the given id.
func sourceNamed(sources []contract.SourceInfo, sourceID string) (contract.SourceInfo, bool) {
	for _, info := range sources {
		if info.SourceID == sourceID {
			return info, true
		}
	}
	return contract.SourceInfo{}, false
}
