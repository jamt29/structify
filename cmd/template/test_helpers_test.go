package template

// templateMinValidManifestYAML builds a minimal valid scaffold.yaml for tests.
func templateMinValidManifestYAML(name string) string {
	return "" +
		"name: \"" + name + "\"\n" +
		"version: \"0.0.1\"\n" +
		"author: \"test\"\n" +
		"language: \"go\"\n" +
		"architecture: \"clean\"\n" +
		"description: \"test\"\n" +
		"tags: [\"test\"]\n" +
		"inputs: []\n" +
		"files: []\n" +
		"steps: []\n"
}

