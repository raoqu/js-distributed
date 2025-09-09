package script

import (
	"context"
	"errors"
	"log"
	"main/util"
)

type ScriptRedisStore struct {
	Group string
	Redis *util.RedisClient
}

func NewScriptRedisStore(groupName string, redis *util.RedisClient) *ScriptRedisStore {
	return &ScriptRedisStore{Group: groupName, Redis: redis}
}

func (s *ScriptRedisStore) Load(callback ScriptLoadCallback) {
	scriptNames, err := s.List()
	if err != nil {
		log.Printf("Failed to list scripts from Redis: %v", err)
		return
	}

	for _, name := range scriptNames {
		code, err := s.Get(name)
		if err != nil {
			log.Printf("Failed to get script '%s' from Redis: %v", name, err)
			continue
		}

		callback(name, code)
	}
}

func (s *ScriptRedisStore) Save(scriptName string, scriptCode string) error {
	if s.Redis == nil {
		return errors.New("redis client not initialized")
	}

	err := s.Redis.SetHValue(s.Group, scriptName, scriptCode)
	if err != nil {
		log.Printf("Failed to store script %s in Redis: %v", scriptName, err)
		return err
	}

	return nil
}

func (s *ScriptRedisStore) Get(scriptName string) (string, error) {
	if s.Redis == nil {
		return "", errors.New("redis client not initialized")
	}

	scriptCode, err := s.Redis.GetHValue(s.Group, scriptName)
	if err != nil {
		return "", err
	}

	return scriptCode, nil
}

func (s *ScriptRedisStore) Delete(scriptName string) error {
	if s.Redis == nil {
		return errors.New("redis client not initialized")
	}

	err := s.Redis.HDel(s.Group, scriptName)
	if err != nil {
		log.Printf("Failed to delete script %s from Redis: %v", scriptName, err)
		return err
	}

	return nil
}

func (s *ScriptRedisStore) List() ([]string, error) {
	if s.Redis == nil {
		return nil, errors.New("redis client not initialized")
	}

	scriptNames, err := s.Redis.HKeys(s.Group)
	if err != nil {
		log.Printf("Failed to list scripts from Redis: %v", err)
		return nil, err
	}

	return scriptNames, nil
}

// ScriptExists checks if a script exists in Redis
func (s *ScriptRedisStore) Exists(scriptName string) (bool, error) {
	if s.Redis == nil {
		return false, errors.New("redis client not initialized")
	}

	exists, err := s.Redis.Client.HExists(context.Background(), s.Group, scriptName).Result()
	if err != nil {
		return false, err
	}

	return exists, nil
}
