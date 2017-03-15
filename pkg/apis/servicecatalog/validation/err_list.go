package validation

import (
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/util/validation/field"
)

func appendToErrListAndLog(list field.ErrorList, toAppend ...*field.Error) field.ErrorList {
	if len(toAppend) > 0 {
		logErrList(toAppend...)
		list = append(list, toAppend...)
	}
	return list
}

func logErrList(list ...*field.Error) {
	for i, err := range list {
		glog.Errorf("%d: %#v", i, *err)
	}
}
