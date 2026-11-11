package plugins

import (
	"context"
	"encoding/json"

	"light-launcher/core/internal/contract"
)

const maxSecretKeyLength = 200

type secretParams struct {
	Key   string  `json:"key"`
	Value *string `json:"value"`
}

// serviceSecretsGet returns {"value": secret} for this plugin, with a null value when it is not set.
func serviceSecretsGet(ctx context.Context, env *pluginEnv, params json.RawMessage) (any, error) {
	req, err := decodeParams[secretParams](params)
	if err != nil {
		return nil, err
	}
	if err := checkSecretKey(req.Key); err != nil {
		return nil, err
	}
	store, err := secretStore(env)
	if err != nil {
		return nil, err
	}
	value, found, err := store.Get(ctx, env.id, req.Key)
	if err != nil {
		return nil, internalError(err.Error())
	}
	if !found {
		return valueReply{Value: nil}, nil
	}
	return valueReply{Value: value}, nil
}

// valueReply is the protocol shape of secrets.get and storage.get results.
type valueReply struct {
	Value any `json:"value"`
}

func serviceSecretsSet(ctx context.Context, env *pluginEnv, params json.RawMessage) (any, error) {
	req, err := decodeParams[secretParams](params)
	if err != nil {
		return nil, err
	}
	if err := checkSecretKey(req.Key); err != nil {
		return nil, err
	}
	if req.Value == nil {
		return nil, invalidParams("value is required")
	}
	store, err := secretStore(env)
	if err != nil {
		return nil, err
	}
	if err := store.Set(ctx, env.id, req.Key, *req.Value); err != nil {
		return nil, internalError(err.Error())
	}
	return nil, nil
}

func serviceSecretsDelete(ctx context.Context, env *pluginEnv, params json.RawMessage) (any, error) {
	req, err := decodeParams[secretParams](params)
	if err != nil {
		return nil, err
	}
	if err := checkSecretKey(req.Key); err != nil {
		return nil, err
	}
	store, err := secretStore(env)
	if err != nil {
		return nil, err
	}
	if err := store.Delete(ctx, env.id, req.Key); err != nil {
		return nil, internalError(err.Error())
	}
	return nil, nil
}

func checkSecretKey(key string) error {
	if key == "" || len(key) > maxSecretKeyLength {
		return invalidParams("key must be 1 to 200 characters")
	}
	return nil
}

func secretStore(env *pluginEnv) (contract.Secrets, error) {
	if env.deps.secrets == nil {
		return nil, internalError("secrets are not available")
	}
	return env.deps.secrets, nil
}
