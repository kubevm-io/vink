package filter

import (
	"strings"

	"github.com/kubevm.io/vink/pkg/watcher/utils"
	"k8s.io/apimachinery/pkg/types"
)

type FilterFunc func(obj any) (bool, error)

func TrueFilterFunc() FilterFunc {
	return func(_ any) (bool, error) {
		return true, nil
	}
}

func FilterFuncWithNamespacedName(namespaceName *types.NamespacedName) FilterFunc {
	return func(obj any) (bool, error) {
		metaobj, err := utils.GetMetaObject(obj)
		if err != nil {
			return false, err
		}

		if len(namespaceName.Namespace) > 0 && metaobj.GetNamespace() != namespaceName.Namespace {
			return false, nil
		}
		if len(namespaceName.Name) > 0 && !strings.Contains(metaobj.GetName(), namespaceName.Name) {
			return false, nil
		}

		return true, nil
	}
}

func PassesAllFilters(filterFuncs []FilterFunc, obj any) (bool, error) {
	for _, fn := range filterFuncs {
		ok, err := fn(obj)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}
