/*
Copyright 2018 The Kubernetes Authors.

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

package tableconvertor

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metatable "k8s.io/apimachinery/pkg/api/meta/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
)

// RowFunction is a function that maps an object (e.g. ServiceInstance)
// to an array of values that get printed in columnar form when a user
// runs kubectl get.
type RowFunction func(obj runtime.Object, meta metav1.Object, name, age string) ([]interface{}, error)

// NewTableConvertor creates a TableConvertor with the specified columns
// and RowFunction, which is used to map an object to those columns.
func NewTableConvertor(columnDefinitions []metav1beta1.TableColumnDefinition, rowFunction RowFunction) rest.TableConvertor {
	return &convertor{columnDefinitions, rowFunction}
}

type convertor struct {
	columnDefinitions []metav1beta1.TableColumnDefinition
	rowFunction       RowFunction
}

func (c *convertor) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1beta1.Table, error) {
	table := &metav1beta1.Table{
		ColumnDefinitions: c.columnDefinitions,
	}
	if m, err := meta.ListAccessor(obj); err == nil {
		table.ResourceVersion = m.GetResourceVersion()
		table.SelfLink = m.GetSelfLink()
		table.Continue = m.GetContinue()
	} else {
		if m, err := meta.CommonAccessor(obj); err == nil {
			table.ResourceVersion = m.GetResourceVersion()
			table.SelfLink = m.GetSelfLink()
		}
	}

	var err error
	table.Rows, err = metatable.MetaToTableRow(obj, c.rowFunction)
	return table, err
}
