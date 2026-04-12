package rendertest

metadata: {
	name:             "test-module"
	namespace:        "default"
	description:      "Minimal test module for render bridge"
	modulePath:       "test.example/render-test"
	defaultNamespace: "default"
	version:          "0.1.0"
	uuid:             "00000000-0000-0000-0000-000000000001"
	fqn:              "test.example/render-test/test-module:0.1.0"
	labels: {}
	annotations: {}
}

#config: {
	message: *"default-hello" | string
}

values: #config

components: {
	web: {
		metadata: {
			labels: {}
			annotations: {}
		}
		data: values
	}
}
