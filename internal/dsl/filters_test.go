package dsl

import "testing"

func TestApplyFilter_AllFilters(t *testing.T) {
	tests := []struct {
		name   string
		filter string
		in     string
		want   string
	}{
		// snake_case
		{"snake_camel", "snake_case", "MyProject", "my_project"},
		{"snake_kebab", "snake_case", "my-project", "my_project"},
		{"snake_snake", "snake_case", "my_project", "my_project"},
		{"snake_spaces", "snake_case", "My Project API", "my_project_api"},

		// pascal_case
		{"pascal_kebab", "pascal_case", "my-project", "MyProject"},
		{"pascal_snake", "pascal_case", "my_project", "MyProject"},
		{"pascal_spaces", "pascal_case", "my project api", "MyProjectApi"},
		{"pascal_acronym", "pascal_case", "myAPIClient", "MyApiClient"},

		// camel_case
		{"camel_kebab", "camel_case", "my-project", "myProject"},
		{"camel_snake", "camel_case", "my_project", "myProject"},
		{"camel_spaces", "camel_case", "my project api", "myProjectApi"},
		{"camel_acronym", "camel_case", "MyAPIClient", "myApiClient"},

		// kebab_case
		{"kebab_camel", "kebab_case", "MyProject", "my-project"},
		{"kebab_snake", "kebab_case", "my_project", "my-project"},
		{"kebab_spaces", "kebab_case", "My Project API", "my-project-api"},
		{"kebab_kebab", "kebab_case", "my-project", "my-project"},

		// upper/lower
		{"upper", "upper", "hello", "HELLO"},
		{"upper_mixed", "upper", "MyProject", "MYPROJECT"},
		{"lower", "lower", "HELLO", "hello"},
		{"lower_mixed", "lower", "MyProject", "myproject"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyFilter(tt.filter, tt.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestApplyFilter_UnknownFilter(t *testing.T) {
	_, err := ApplyFilter("nope", "x")
	if err == nil {
		t.Fatalf("expected error")
	}
}
