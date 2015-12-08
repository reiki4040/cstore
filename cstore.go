package cstore

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/BurntSushi/toml"
)

type Format int

const (
	ENV_HOME = "HOME"
)

const (
	TOML Format = 1 + iota
	JSON
	YAML
)

func NewManager(name, baseDirPath string) (*Manager, error) {
	err := createDir(baseDirPath)
	if err != nil {
		return nil, err
	}

	return &Manager{
		name:        name,
		baseDirPath: baseDirPath,
		csMap:       make(map[string]*CStore),
	}, nil
}

type Manager struct {
	name        string
	baseDirPath string
	csMap       map[string]*CStore
}

func (m *Manager) Name() string {
	return m.name
}

func (m *Manager) New(name string, format Format) (*CStore, error) {
	cs, err := NewCStore(name, m.baseDirPath+string(os.PathSeparator)+name, format)
	if err != nil {
		return nil, err
	}

	if m.csMap == nil {
		m.csMap = make(map[string]*CStore)
	}
	m.csMap[name] = cs

	return cs, nil
}

func (m *Manager) Get(name string) *CStore {
	if m.csMap == nil {
		return nil
	}

	if cs, ok := m.csMap[name]; ok {
		return cs
	} else {
		return nil
	}
}

func (m *Manager) Remove(name string) *CStore {
	if m.csMap == nil {
		return nil
	}

	if cs, ok := m.csMap[name]; ok {
		delete(m.csMap, name)
		return cs
	}

	return nil
}

type Validatable interface {
	Validate() error
}

type Serializable interface {
	Load(p interface{}) error
	Store(p interface{}) error
	Remove() error
}

func Get(v Validatable, s Serializable) error {
	err := GetWithoutValidate(v, s)
	if err != nil {
		return err
	}

	err = v.Validate()
	if err != nil {
		return err
	}

	return nil
}

func GetWithoutValidate(p interface{}, s Serializable) error {
	err := s.Load(p)
	if err != nil {
		return err
	}

	return nil
}

func Save(v Validatable, s Serializable) error {
	err := v.Validate()
	if err != nil {
		return err
	}

	return SaveWithoutValidate(v, s)
}

func SaveWithoutValidate(p interface{}, s Serializable) error {
	err := s.Store(p)
	if err != nil {
		return err
	}

	return nil
}

func NewCStore(name, filePath string, format Format) (*CStore, error) {
	var s Serializable
	switch format {
	case TOML:
		s = &TomlFile{
			FilePath: filePath,
		}
	case JSON:
		s = &JsonFile{
			FilePath: filePath,
		}
	case YAML:
		s = &YamlFile{
			FilePath: filePath,
		}
	default:
		return nil, fmt.Errorf("invalid format type: %d", format)
	}

	cs := &CStore{
		name:       name,
		serializer: s,
	}

	return cs, nil
}

type CStore struct {
	name       string
	serializer Serializable
}

func (cs *CStore) Name() string {
	return cs.name
}

func (cs *CStore) Get(v Validatable) error {
	return Get(v, cs.serializer)
}

func (cs *CStore) Save(v Validatable) error {
	return Save(v, cs.serializer)
}

func (cs *CStore) GetWithoutValidate(p interface{}) error {
	return GetWithoutValidate(p, cs.serializer)
}

func (cs *CStore) SaveWithoutValidate(p interface{}) error {
	return SaveWithoutValidate(p, cs.serializer)
}

func (cs *CStore) Load(p interface{}) error {
	return cs.serializer.Load(p)
}

func (cs *CStore) Store(p interface{}) error {
	return cs.serializer.Store(p)
}

func (cs *CStore) Remove() error {
	return cs.serializer.Remove()
}

func createDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			if !os.IsExist(err) {
				return err
			}
		}
	}

	return nil
}

type TomlFile struct {
	FilePath string
}

func (t *TomlFile) Load(p interface{}) error {
	return LoadFromTomlFile(t.FilePath, p)
}

func (t *TomlFile) Store(p interface{}) error {
	return StoreToTomlFile(t.FilePath, p)
}

func (t *TomlFile) Remove() error {
	return removeFile(t.FilePath)
}

func LoadFromTomlFile(filePath string, p interface{}) error {
	if _, err := toml.DecodeFile(filePath, p); err != nil {
		return err
	}

	return nil
}

func StoreToTomlFile(filePath string, p interface{}) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// wrap with bufio.NewWriter in toml
	enc := toml.NewEncoder(f)
	if err := enc.Encode(p); err != nil {
		return err
	}

	return nil
}

type JsonFile struct {
	FilePath string
}

func (f *JsonFile) Load(p interface{}) error {
	return LoadFromJsonFile(f.FilePath, p)
}

func (f *JsonFile) Store(p interface{}) error {
	return StoreToJsonFile(f.FilePath, p)
}

func (f *JsonFile) Remove() error {
	return removeFile(f.FilePath)
}

func StoreToJsonFile(filePath string, p interface{}) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	enc := json.NewEncoder(w)
	if err := enc.Encode(p); err != nil {
		return err
	}

	return nil
}

func LoadFromJsonFile(filePath string, p interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	dec := json.NewDecoder(bufio.NewReader(file))
	return dec.Decode(p)
}

type YamlFile struct {
	FilePath string
}

func (f *YamlFile) Load(p interface{}) error {
	return LoadFromYamlFile(f.FilePath, p)
}

func (f *YamlFile) Store(p interface{}) error {
	return StoreToYamlFile(f.FilePath, p)
}

func (f *YamlFile) Remove() error {
	return removeFile(f.FilePath)
}

func StoreToYamlFile(filePath string, p interface{}) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	bytes, err := yaml.Marshal(p)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)
	defer w.Flush()

	_, err = w.Write(bytes)
	return err
}

func LoadFromYamlFile(filePath string, p interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	yml, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(yml, p)
}

func removeFile(filePath string) error {
	return os.Remove(filePath)
}
