package dynamicx

import (
	"encoding/json"
	"fmt"

	"github.com/kubevm.io/vink/pkg/k8s/client/clientset/versioned/scheme"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func FromUnstructured[T any](src *unstructured.Unstructured) (T, error) {
	var zero T

	data, err := UnstructuredToJSONBytes(src)
	if err != nil {
		return zero, err
	}

	var dst T
	if err := json.Unmarshal(data, &dst); err != nil {
		return zero, err
	}

	return dst, nil
}

func FromObject[T any](obj any) (T, error) {
	var zero T

	src, err := InterfaceToUnstructured[T](obj)
	if err != nil {
		return zero, err
	}

	return FromUnstructured[T](src)
}

func UnstructuredToJSONBytes(obj *unstructured.Unstructured) ([]byte, error) {
	obj.SetManagedFields(nil)
	jsonBytes, err := json.Marshal(obj.Object)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

func InterfaceToUnstructured[T any](obj any) (*unstructured.Unstructured, error) {
	var zero T

	c, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	un := &unstructured.Unstructured{}
	un.SetUnstructuredContent(c)

	if len(un.GetKind()) == 0 || len(un.GetAPIVersion()) == 0 {
		_, gvk := ResolveGVRAndGVK(zero)
		un.SetAPIVersion(gvk.GroupVersion().String())
		un.SetKind(gvk.Kind)
	}

	return un, nil
}

func ProtoToUnstructured(obj any) (*unstructured.Unstructured, error) {
	// Protobuf -> JSON
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	// JSON -> map[string]any
	var content map[string]any
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, err
	}

	un := &unstructured.Unstructured{}
	un.SetUnstructuredContent(content)
	return un, nil
}

func GetGVK(obj runtime.Object) (schema.GroupVersionKind, error) {
	gvks, _, _ := scheme.Scheme.ObjectKinds(obj)
	if len(gvks) < 1 {
		return schema.GroupVersionKind{}, fmt.Errorf("no gvk found")
	}
	return gvks[0], nil
}
