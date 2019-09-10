Write Unit Tests
================================

We have two business components `service.Indexer` & `sync.syncClient` that compose of some supported interfaces.
The interfaces are defined as `subscriber.Subscriber`, `store.Manager`, `blockchain.Client`.
Each interface has an individual functionality that to get more details, please take a look as them.

To be easy to write thoroughly Unit Test for the both important components, we use mixes of some libraries & tools as below:
  * [`testify`](https://github.com/stretchr/testify) testing suite
  * [`mockery`](https://github.com/vektra/mockery) mock generation tool
  * [`ginkgo`](https://github.com/onsi/ginkgo) testing framework
  * [`gomega`](https://github.com/onsi/gomega) matcher library
  
[`testify`](https://github.com/stretchr/testify) testing suite
--------------------------------------------------------------

Intensively use `mock` package to mock real objects.
Package [`mock`](https://godoc.org/github.com/stretchr/testify/mock) provides a system by which it is possible to mock your objects and verify calls are happening as expected.

You can use the [`mockery`](https://github.com/vektra/mockery) tool to autogenerate the mock code against an interface as well, making using mocks much quicker.

[`mockery`](https://github.com/vektra/mockery) mock generation tool
-------------------------------------------------------------------

### Installation

`go get github.com/vektra/mockery/.../`, then `$GOPATH/bin/mockery`

### Most used flags

#### Name

The `-name` option takes either the name or matching regular expression of interface to generate mock(s) for.

#### All

It's common for a big package to have a lot of interfaces, so mockery provides `-all`.
This option will tell mockery to scan all files under the directory named by `-dir` ("." by default)
and generates mocks for any interfaces it finds. This option implies `-recursive=true`.

### Output

mockery always generates files with the package `mocks` to keep things clean and simple.
You can control which mocks directory is used by using `-output`, which defaults to `./mocks`.

[`ginkgo`](https://github.com/onsi/ginkgo) A Go BDD Testing Framework
---------------------------------------------------------------------

- Structure your BDD-style tests expressively:
    - Nestable [`Describe`, `Context` and `When` container blocks](http://onsi.github.io/ginkgo/#organizing-specs-with-containers-describe-and-context)
    - [`BeforeEach` and `AfterEach` blocks](http://onsi.github.io/ginkgo/#extracting-common-setup-beforeeach) for setup and teardown
    - [`It` and `Specify` blocks](http://onsi.github.io/ginkgo/#individual-specs-) that hold your assertions
    
- Straightforward support for third-party testing libraries such as [Gomock](https://code.google.com/p/gomock/) and [Testify](https://github.com/stretchr/testify).  Check out the [docs](http://onsi.github.io/ginkgo/#third-party-integrations) for details.

- Ginkgo is best paired with [Gomega](https://github.com/onsi/gomega)

[`gomega`](https://github.com/onsi/gomega) Ginkgo's Preferred Matcher Library
-----------------------------------------------------------------------------

- A matcher/assertion library. It is best paired with the Ginkgo BDD test framework, but can be adapted for use in other contexts too.