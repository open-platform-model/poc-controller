package catalog

import (
	provs "opmodel.dev/opm/v1alpha1/providers@v1"
	k8s_provs "opmodel.dev/kubernetes/v1/providers/kubernetes@v1"
	gw_provs "opmodel.dev/gateway_api/v1alpha1/providers/kubernetes@v1"
	cm_provs "opmodel.dev/cert_manager/v1alpha1/providers/kubernetes@v1"
	k8up_provs "opmodel.dev/k8up/v1alpha1/providers/kubernetes@v1"
	mdb_provs "opmodel.dev/mongodb_operator/v1alpha1/providers/kubernetes@v1"
	chop_provs "opmodel.dev/clickhouse_operator/v1alpha1/providers/kubernetes@v1"
	otel_provs "opmodel.dev/otel_collector/v1alpha1/providers/kubernetes@v1"
	chvmm "opmodel.dev/ch_vmm/v1alpha1/providers/kubernetes@v1"
	istio_provs "opmodel.dev/istio/v1alpha1/providers/kubernetes@v1"
)

providers: {
	kubernetes: provs.#Registry["kubernetes"] & {
		#transformers: k8s_provs.#Provider.#transformers
		#transformers: gw_provs.#Provider.#transformers
		#transformers: cm_provs.#Provider.#transformers
		#transformers: k8up_provs.#Provider.#transformers
		#transformers: mdb_provs.#Provider.#transformers
		#transformers: chop_provs.#Provider.#transformers
		#transformers: otel_provs.#Provider.#transformers
		#transformers: chvmm.#Provider.#transformers
		#transformers: istio_provs.#Provider.#transformers
	}
}
