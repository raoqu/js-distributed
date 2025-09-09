package script

import (
	"log"
	"sync"
	"time"

	"github.com/puzpuzpuz/xsync/v4"
)

// ScriptCache represents a local cache for scripts
type ScriptCache struct {
	scripts     *xsync.Map[string, *ScriptEntry]
	initOnce    sync.Once
	initialized bool
	Store       ScriptStore
}

// ScriptEntry represents a cached script with metadata
type ScriptEntry struct {
	Name      string
	Code      string
	UpdatedAt time.Time
}

// Global script cache instance
var scriptCache *ScriptCache

// GetScriptCache returns the singleton script cache instance
func NewScriptCache(store ScriptStore) *ScriptCache {
	return &ScriptCache{
		scripts: xsync.NewMap[string, *ScriptEntry](),
		Store:   store,
	}
}

// Initialize loads all scripts from Redis into the local cache
func (sc *ScriptCache) Initialize() error {
	var initErr error
	sc.initOnce.Do(func() {
		sc.Store.Load(func(scriptName string, scriptCode string) {
			sc.scripts.Store(scriptName, &ScriptEntry{
				Name:      scriptName,
				Code:      scriptCode,
				UpdatedAt: time.Now(),
			})
		})
		log.Printf("Script cache initialized with %d scripts", sc.scripts.Size())
		sc.initialized = true
	})

	return initErr
}

// GetScript retrieves a script from the cache, falling back to Redis if not found
func (sc *ScriptCache) GetScript(name string) (string, error) {
	// Try to get from cache first
	if entry, ok := sc.scripts.Load(name); ok && entry != nil {
		return entry.Code, nil
	}

	// Not in cache, try to get from Redis
	code, err := sc.Store.Get(name)
	if err != nil {
		return "", err
	}

	// Store in cache for future use
	sc.scripts.Store(name, &ScriptEntry{
		Name:      name,
		Code:      code,
		UpdatedAt: time.Now(),
	})

	return code, nil
}

// StoreScript stores a script in both the cache and Redis
func (sc *ScriptCache) StoreScript(name string, code string) error {
	// Store in Redis first
	err := sc.Store.Save(name, code)
	if err != nil {
		return err
	}

	// Then update the cache
	sc.scripts.Store(name, &ScriptEntry{
		Name:      name,
		Code:      code,
		UpdatedAt: time.Now(),
	})

	return nil
}

// DeleteScript removes a script from both the cache and Redis
func (sc *ScriptCache) DeleteScript(name string) error {
	// Delete from Redis first
	err := sc.Store.Delete(name)
	if err != nil {
		return err
	}

	// Then remove from cache
	sc.scripts.Delete(name)

	return nil
}

// ListScripts returns all script names from the cache
func (sc *ScriptCache) ListScripts() ([]string, error) {
	// If cache is not initialized, fall back to Redis
	if !sc.initialized {
		return sc.Store.List()
	}

	// Get all script names from cache
	scripts := make([]string, 0)
	sc.scripts.Range(func(key string, value *ScriptEntry) bool {
		scripts = append(scripts, key)
		return true
	})

	return scripts, nil
}

// ScriptExists checks if a script exists in the cache or Redis
func (sc *ScriptCache) ScriptExists(name string) (bool, error) {
	// Check cache first
	if _, ok := sc.scripts.Load(name); ok {
		return true, nil
	}

	// Not in cache, check Redis
	exists, err := sc.Store.Exists(name)
	if err != nil {
		return false, err
	}

	// If it exists in Redis but not in cache, load it into cache
	if exists {
		code, err := sc.Store.Get(name)
		if err != nil {
			return true, nil // Still return true since it exists in Redis
		}

		sc.scripts.Store(name, &ScriptEntry{
			Name:      name,
			Code:      code,
			UpdatedAt: time.Now(),
		})
	}

	return exists, nil
}
