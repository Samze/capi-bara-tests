# CAPI BARAS - Banausic Acceptance & Regression Avoidance Suite
BARAs are a test suite to supplement [CATS](https://github.com/cloudfoundry/cf-acceptance-tests). While CATS focuses on a happy-path tests for major CF features, BARAS are broader.

BARAS is home to tests that couldn't go anywhere else:
- End-to-end feature interaction tests
    - e.g. can I configure sidecars in server-side manifests?
- Bleeding edge CF API features
    - e.g. I want to test my new API resource, but there's not stable CLI commands yet.
- Integration-level regression tests
    - e.g. this bug is impossible to reproduce in unit tests!
- Integration tests that concern themselves with some specific CF implementations (kubectl, nginx)

## Test Setup
### Prerequisites for running BARAS

- Same as for [CATS](https://github.com/cloudfoundry/cf-acceptance-tests),
  with the following exceptions:

- Check out a copy of `capi-bara-tests`
  and make sure that it is added to your `$GOPATH`.
  The recommended way to do this is to run:

  ```bash
  go get -d github.com/cloudfoundry/capi-bara-tests
  ```

  You will receive a warning:
  `no buildable Go source files`.
  This can be ignored, as there is only test code in the package.

### Updating `go` dependencies
- Same as for [CATS](https://github.com/cloudfoundry/cf-acceptance-tests).

## Test Configuration
- Similar to [CATS](https://github.com/cloudfoundry/cf-acceptance-tests).

Example config for CF for VMs:
```bash
cat > integration_config.json <<EOF
{
  "api": "api.bosh-lite.com",
  "apps_domain": "bosh-lite.com",
  "admin_user": "admin",
  "admin_password": "admin",
  "skip_ssl_validation": true
}
EOF
export CONFIG=$PWD/integration_config.json
```

Example config for CF for Kubernetes:
```bash
cat > integration_config.json <<EOF
{
  "api": "api.k8s.example.com",
  "apps_domain": "apps.k8s.example.com",
  "admin_user": "admin",
  "admin_password": "admin",
  "skip_ssl_validation": true,
  "infrastructure": "kubernetes",
  "gcloud_project_name": "gcp-project-name",
  "cluster_zone": "gcp-zone-eg-us-west1-a",
  "cluster_name": "gke-cluster-name",
  "cf_push_timeout": 480,
  "python_buildpack_name": "paketo-community/python",
  "ruby_buildpack_name": "paketo-buildpacks/ruby",
  "java_buildpack_name": "paketo-buildpacks/java",
  "go_buildpack_name": "paketo-buildpacks/go",
  "nodejs_buildpack_name": "paketo-buildpacks/nodejs",
  "staticfile_buildpack_name": "paketo-community/staticfile",
  "binary_buildpack_name": "paketo-buildpacks/procfile"
}
EOF
export CONFIG=$PWD/integration_config.json
```

## Test Execution
To execute all test groups, run the following from the root directory of cf-acceptance-tests:
```bash
./bin/test
```

##### Parallel execution
To execute all test groups, and have tests run in parallel across four processes one would run:

```bash
./bin/test -nodes=4
```

Be careful with this number, as it's effectively "how many apps to push at once", as nearly every example pushes an app.


##### Focusing Test Groups
If you are already familiar with CATs you probably know that there are many test groups. You may not wish to run all the tests in all contexts, and sometimes you may want to focus individual test groups to pinpoint a failure. To execute a specific group of acceptance tests, e.g. `routing/`, edit your [`integration_config.json`](#test-configuration) file and set all `include_*` values to `false` except for `include_routing` then run the following:

```bash
./bin/test
```

To execute tests in a single file use an `FDescribe` block around the tests in that file:
```go
var _ = BackendCompatibilityDescribe("Backend Compatibility", func() {
  FDescribe("Focused tests", func() { // Add this line here
  // ... rest of file
  }) // Close here
})

```

The test group names correspond to directory names.

##### Verbose Output
To see verbose output from `ginkgo`, use the `-v` flag.

```bash
./bin/test -v
```

You can of course combine the `-v` flag with the `-nodes=N` flag.

## Contributing

- See [CATS](https://github.com/cloudfoundry/cf-acceptance-tests).

### Code Conventions

- See [CATS](https://github.com/cloudfoundry/cf-acceptance-tests).

