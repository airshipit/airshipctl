/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package document

import (
	"errors"
	"io"

	"github.com/stretchr/testify/mock"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/fs"
)

var _ document.Bundle = &MockBundle{}

// MockBundle mocks document.Bundle interface
type MockBundle struct {
	mock.Mock
}

// Write mock
func (mb *MockBundle) Write(writer io.Writer) error {
	args := mb.Called(writer)
	return args.Error(0)
}

// SetFileSystem mock
func (mb *MockBundle) SetFileSystem(filesystem fs.FileSystem) error {
	args := mb.Called(filesystem)
	return args.Error(0)
}

// GetFileSystem mock
func (mb *MockBundle) GetFileSystem() fs.FileSystem {
	args := mb.Called()
	val, ok := args.Get(0).(fs.FileSystem)
	if !ok {
		return nil
	}
	return val
}

// Select mock
func (mb *MockBundle) Select(selector document.Selector) ([]document.Document, error) {
	args := mb.Called(selector)
	val, ok := args.Get(0).([]document.Document)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// SelectOne mock
func (mb *MockBundle) SelectOne(selector document.Selector) (document.Document, error) {
	args := mb.Called(selector)
	val, ok := args.Get(0).(document.Document)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// SelectBundle mock
func (mb *MockBundle) SelectBundle(selector document.Selector) (document.Bundle, error) {
	args := mb.Called(selector)
	val, ok := args.Get(0).(document.Bundle)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// SelectByFieldValue mock
func (mb *MockBundle) SelectByFieldValue(path string, condition func(interface{}) bool) (document.Bundle, error) {
	args := mb.Called(path, condition)
	val, ok := args.Get(0).(document.Bundle)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// GetByGvk mock
func (mb *MockBundle) GetByGvk(group, version, kind string) ([]document.Document, error) {
	args := mb.Called(group, version, kind)
	val, ok := args.Get(0).([]document.Document)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// GetByName mock
func (mb *MockBundle) GetByName(name string) (document.Document, error) {
	args := mb.Called(name)
	val, ok := args.Get(0).(document.Document)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// GetByAnnotation mock
func (mb *MockBundle) GetByAnnotation(annotationSelector string) ([]document.Document, error) {
	args := mb.Called(annotationSelector)
	val, ok := args.Get(0).([]document.Document)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// GetByLabel mock
func (mb *MockBundle) GetByLabel(labelSelector string) ([]document.Document, error) {
	args := mb.Called(labelSelector)
	val, ok := args.Get(0).([]document.Document)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// GetAllDocuments mock
func (mb *MockBundle) GetAllDocuments() ([]document.Document, error) {
	args := mb.Called()
	val, ok := args.Get(0).([]document.Document)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

// Append mock
func (mb *MockBundle) Append(doc document.Document) error {
	args := mb.Called(doc)
	return args.Error(0)
}

var (
	// EmptyBundleFactory returns empty MockBundle
	EmptyBundleFactory document.BundleFactoryFunc = func() (document.Bundle, error) {
		return &MockBundle{}, nil
	}
	// ErrorBundleFactory returns error instead of bundle
	ErrorBundleFactory document.BundleFactoryFunc = func() (document.Bundle, error) {
		return nil, errors.New("error")
	}
)
