package utils

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func GetMetaObject(obj any) (metav1.Object, error) {
	if accessor, ok := obj.(metav1.Object); ok {
		return accessor, nil
	}
	if ro, ok := obj.(runtime.Object); ok {
		accessor, err := meta.Accessor(ro)
		if err != nil {
			return nil, err
		}
		return accessor, nil
	}
	return nil, fmt.Errorf("object has no MetaObject")
}
