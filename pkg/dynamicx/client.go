package dynamicx

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

func NewClient[T any](dyclient dynamic.Interface) *Client[T] {
	return &Client[T]{
		dyclient: dyclient,
	}
}

type Client[T any] struct {
	dyclient dynamic.Interface
}

func (c *Client[T]) Get(ctx context.Context, namespace, name string) (T, error) {
	var zero T

	gvr, _ := ResolveGVRAndGVK(zero)
	unstruct, err := c.dyclient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return zero, err
	}

	return FromUnstructured[T](unstruct)
}

func (c *Client[T]) Create(ctx context.Context, in T) (T, error) {
	var zero T

	unstruct, err := ProtoToUnstructured(in)
	if err != nil {
		return zero, err
	}
	gvr, _ := ResolveGVRAndGVK(zero)
	unstruct, err = c.dyclient.Resource(gvr).Namespace(unstruct.GetNamespace()).Create(ctx, unstruct, metav1.CreateOptions{})
	if err != nil {
		return zero, err
	}

	return FromUnstructured[T](unstruct)
}

func (c *Client[T]) Update(ctx context.Context, in T) (T, error) {
	var zero T

	unstruct, err := ProtoToUnstructured(in)
	if err != nil {
		return zero, err
	}

	gvr, _ := ResolveGVRAndGVK(zero)
	unstruct, err = c.dyclient.Resource(gvr).Namespace(unstruct.GetNamespace()).Update(ctx, unstruct, metav1.UpdateOptions{})
	if err != nil {
		return zero, err
	}

	return FromUnstructured[T](unstruct)
}

func (c *Client[T]) Delete(ctx context.Context, namespace, name string) error {
	var zero T
	gvr, _ := ResolveGVRAndGVK(zero)
	return c.dyclient.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *Client[T]) List(ctx context.Context, namespace string) ([]T, error) {
	var zero T

	gvr, _ := ResolveGVRAndGVK(zero)
	unstructList, err := c.dyclient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]T, 0, len(unstructList.Items))
	for _, unstruct := range unstructList.Items {
		tpl, err := FromUnstructured[T](&unstruct)
		if err != nil {
			return nil, err
		}
		result = append(result, tpl)
	}
	return result, nil
}
