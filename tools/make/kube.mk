##@ Kubernetes

.PHONY: kube.generate
kube.generate: ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
kube.generate:
	@$(LOG_TARGET)
# @tools/bin/controller-gen-darwin object:headerFile="$(ROOT_DIR)/tools/boilerplate/boilerplate.go.txt" paths="$(ROOT_DIR)/pkg/k8s/apis/vink/v1alpha1/..."
# @tools/bin/controller-gen-darwin crd paths="$(ROOT_DIR)/pkg/k8s/apis/vink/v1alpha1/..." output:crd:dir="$(ROOT_DIR)/manifests/crds"
	@tools/bin/controller-gen-darwin \
		object:headerFile="$(ROOT_DIR)/tools/boilerplate/boilerplate.go.txt" \
		crd:allowDangerousTypes=true \
		paths="$(ROOT_DIR)/pkg/k8s/apis/vink/v1alpha1/..." \
		output:crd:dir="$(ROOT_DIR)/manifests/crds"

	@cd $(ROOT_DIR)/pkg/k8s/apis/vink/v1alpha1
	@source $(ROOT_DIR)/vendor/k8s.io/code-generator/kube_codegen.sh && \
		kube::codegen::gen_client \
			--output-dir $(ROOT_DIR)/pkg/k8s/client \
			--output-pkg github.com/kubevm.io/vink/pkg/k8s/client \
			--boilerplate $(ROOT_DIR)/tools/boilerplate/boilerplate.go.txt \
			--with-watch \
			$(ROOT_DIR)/pkg/k8s/apis

# .PHONY: gen-client
# gen-client: ## Generate client code.
# gen-client:
# 	go run $(GOPATH)/pkg/mod/k8s.io/code-generator@v0.31.1/cmd/client-gen \
# 		--input=$(ROOT_DIR)/pkg/k8s/apis \
# 		--clientset-name=client \
# 		--go-header-file=$(ROOT_DIR)/tools/boilerplate/boilerplate.go.txt \
# 		-i baize.io/api/kube/api \
# 		-o . \
# 		-p github.com/kubevm.io/pkg/k8s \
# 		--input-base github.com/kubevm.io/pkg/k8s
# 	mv baize.io/api/kube/client . && rm -rf "baize.io"
