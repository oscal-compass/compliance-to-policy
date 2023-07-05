/*
Copyright 2023 IBM Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"

	k8sruntime "k8s.io/apimachinery/pkg/runtime"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	goyaml "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	sigyaml "sigs.k8s.io/yaml"
)

var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
}

func GetLogger(name string) *zap.Logger {
	return logger.Named(name)
}

func LoadYaml(path string) ([]*unstructured.Unstructured, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	k8sdec := k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	var objects []*unstructured.Unstructured
	for {
		var objInput map[string]interface{}
		err := dec.Decode(&objInput)
		if errors.Is(err, io.EOF) {
			break
		} else if objInput == nil {
			continue
		} else if err != nil {
			return objects, err
		}
		yamlByte, err := yaml.Marshal(objInput)
		if err != nil {
			return objects, err
		}
		obj := &unstructured.Unstructured{}
		_, gvk, err := k8sdec.Decode(yamlByte, nil, obj)
		_ = gvk
		if err != nil {
			return objects, err
		}
		objects = append(objects, obj)
	}
	return objects, nil
}

func CopyFile(src string, dest string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dest, input, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func MakeDir(path string) (string, error) {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", err
	}
	return path, nil
}

func MakeDirAndWriteObjToYamlFile(dir string, filename string, in interface{}) error {
	if _, err := MakeDir(dir); err != nil {
		return err
	}
	return WriteObjToYamlFile(dir+"/"+filename, in)
}

func WriteObjToYamlFile(path string, in interface{}) error {
	if yamlData, err := sigyaml.Marshal(in); err != nil {
		return err
	} else {
		return os.WriteFile(path, yamlData, os.ModePerm)
	}
}

func WriteObjToYamlFileByGoYaml(path string, in interface{}) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	encoder := goyaml.NewEncoder(file)
	encoder.SetIndent(2)
	return encoder.Encode(in)
}

func WriteObjToJsonFile(path string, in interface{}) error {
	if jsonData, err := json.MarshalIndent(in, "", "\t"); err != nil {
		return err
	} else {
		return os.WriteFile(path, jsonData, os.ModePerm)
	}
}

// Read a yaml file and invoke yaml.Unmarshal(content, out).
// Maps and pointers (to a struct, string, int, etc) are accepted as out
// values.
func LoadYamlFileToObject(path string, out interface{}) error {
	yamlData, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := sigyaml.Unmarshal(yamlData, out); err != nil {
		return err
	}
	return nil
}

func LoadJsonFileToObject(path string, out interface{}) error {
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(jsonData, out); err != nil {
		return err
	}
	return nil
}

func LoadYamlFileToK8sTypedObject(path string, out interface{}) error {
	yamlData, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := utilyaml.Unmarshal(yamlData, out); err != nil {
		return err
	}
	return nil
}

func LoadByteToK8sTypedObject(yamlData []byte, out interface{}) error {
	if err := utilyaml.Unmarshal(yamlData, out); err != nil {
		return err
	}
	return nil
}

func ToK8sTypedObject(unstructured *unstructured.Unstructured, typedObject interface{}) error {
	return k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, typedObject)
}

func ToK8sUnstructedObject(typedObject interface{}) (unstructured.Unstructured, error) {
	var (
		err error
		u   unstructured.Unstructured
	)
	u.Object, err = k8sruntime.DefaultUnstructuredConverter.ToUnstructured(typedObject)
	return u, err
}

type filenameCreator struct {
	currentFilenames map[string]bool
	fnameExt         string
	opts             FilenameCreatorOption
}

type FilenameCreatorOption struct {
	UnlabelToZero bool
}

func NewFilenameCreator(fnameExt string, opts *FilenameCreatorOption) filenameCreator {
	var _opts FilenameCreatorOption
	if opts == nil {
		_opts = FilenameCreatorOption{
			UnlabelToZero: false,
		}
	} else {
		_opts = *opts
	}
	return filenameCreator{
		currentFilenames: map[string]bool{},
		fnameExt:         fnameExt,
		opts:             _opts,
	}
}

func (fc *filenameCreator) Get(fname string) string {
	suffix := 0
	var _fname string
	if fc.opts.UnlabelToZero {
		_fname = fmt.Sprintf("%s%s", fname, fc.fnameExt)
	} else {
		_fname = fmt.Sprintf("%s.%d%s", fname, suffix, fc.fnameExt)
	}
	_, found := fc.currentFilenames[_fname]
	for {
		if !found {
			fc.currentFilenames[_fname] = true
			break
		}
		suffix++
		_fname := fmt.Sprintf("%s.%d.%s", fname, suffix, fc.fnameExt)
		_, found = fc.currentFilenames[_fname]
	}
	return _fname
}

func PathFromPkgDirectory(relativePath string) string {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), relativePath)
	return dir
}

func ChdirFromPkgDirectory(relativePath string) string {
	dir := PathFromPkgDirectory(relativePath)
	if err := os.Chdir(dir); err != nil {
		panic(err)
	}
	return dir
}

type TempDirectory struct {
	tempDir string
}

func NewTempDirectory(tempDir string) TempDirectory {
	dir, err := os.MkdirTemp(tempDir, "tmp-")
	if err != nil {
		panic(err)
	}
	return TempDirectory{tempDir: dir}
}

func (t *TempDirectory) GetTempDir() string {
	return t.tempDir
}
