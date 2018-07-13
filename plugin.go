package main

import (
	"github.com/kadende/kadende-interfaces/pkg/types"
	"github.com/kadende/kadende-interfaces/spi/instance"
	log "github.com/Sirupsen/logrus"
	"path"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"os"
)

type plugin struct{
	dir string
}

// Anything that can be unmarshalled into a generic JSON map
type spec map[string]interface{}

type fileInstance struct {
	instance.Spec
}

func (f plugin) Validate(req *types.Any) error {
	log.Debugln("validate", req.String())
	spec := spec{}

	if err := req.Decode(&spec); err != nil{
		return err
	}
	log.Debugln("validated:", spec)
	return nil
}

func (f plugin) Provision(spec instance.Spec) (*instance.ID, error) {
	fileId := instance.ID(uuid.Must(uuid.NewV4(), nil ).String())

	specString, err := types.AnyValue(spec)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(path.Join(f.dir, string(fileId)), specString.Bytes(), 0644)

	if err != nil{
		return nil, err
	}
	return &fileId, nil
}

func (f plugin) Label(id instance.ID, labels map[string]string) (err error) {
	// read file content
	data, err := ioutil.ReadFile(path.Join(f.dir, string(id)))
	if err != nil{
		return
	}

	var spec instance.Spec
	json.Unmarshal(data, &spec)
	for key, value := range labels{
		spec.Tags[key] = value
	}
	specString, err := types.AnyValue(spec)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(path.Join(f.dir, string(id)), specString.Bytes(), 0644)
	return
}

func (f plugin) Destroy(id instance.ID, context instance.Context) (err error) {
	err = os.Remove(path.Join(f.dir, string(id)))
	return
}

func (f plugin) DescribeInstances(labels map[string]string, properties bool) (instances []instance.Description, err error)  {
	instances = []instance.Description{}
	err = filepath.Walk(f.dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir(){
			// read file content
			data, err := ioutil.ReadFile(path)
			if err != nil{
				return err
			}

			// convert file content to spec interface
			var spec instance.Spec
			err = json.Unmarshal(data, &spec)
			if err != nil {
				return err
			}

			validInstance := true
			for key, value := range labels{
				if spec.Tags[key] != value{
					validInstance = false
					break
				}
			}

			if validInstance{
				instances = append(instances, instance.Description{ID: instance.ID(info.Name()), Tags: spec.Tags})
			}

		}
		return nil
	})
	return
}


func NewPlugin(dir string) instance.Plugin {
	return plugin{dir: dir}
}
