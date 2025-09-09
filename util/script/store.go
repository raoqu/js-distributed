package script

type ScriptLoadCallback func(name string, code string)

type ScriptStore interface {
	Load(callback ScriptLoadCallback)
	Save(name string, code string) error
	Get(name string) (string, error)
	Delete(name string) error
	List() ([]string, error)
	Exists(name string) (bool, error)
}
