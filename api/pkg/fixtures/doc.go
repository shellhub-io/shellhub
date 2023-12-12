// Package fixtures provides utilities for setting up MongoDB fixtures. Each fixture
// is a YAML file inside the `data` subdirectory. Each file contains a list of predefined
// data for the specified collection, which has the same name as the file.
//
// You can prepare the database to receive the fixtures using the `Init` function. The package
// also includes `Apply` to apply the provided fixtures and `Teardown` to reset all the applied
// fixtures. All available fixtures are accessible via constants provided by the package.
package fixtures
