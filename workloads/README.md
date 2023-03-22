# Workload Generators

The workload generators leverage the [Locust](https://locust.io) python framework for generating load against the exposed REST APIs of the sample services.

The generators can be run via a web interface or the command line.

If using the web interface, you can point your web browser to the exposed port to set the workload's concurrency in terms of "users". Then the load runs until the test is stopped in the browser.

Various charts are provided by the web interface to indicate the performance of the load test.

If you do not want to use the web interface, the command line options specify the user concurrency, as well as a run time. Statistics are printed on the
command line for the test.

> **NOTE:** If deploying the workloads on GKE, only the web interface is supported.

More information on these workloads can be found [in the docs](../docs/workloads.md)

