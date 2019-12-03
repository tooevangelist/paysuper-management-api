# PaySuper Management API

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-brightgreen.svg)](https://www.gnu.org/licenses/gpl-3.0) [![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/paysuper/paysuper-management-api/issues)

[![Build Status](https://travis-ci.org/paysuper/paysuper-management-api.svg?branch=master)](https://travis-ci.org/paysuper/paysuper-management-api) [![codecov](https://codecov.io/gh/paysuper/paysuper-management-api/branch/master/graph/badge.svg)](https://codecov.io/gh/paysuper/paysuper-management-api) [![Go Report Card](https://goreportcard.com/badge/github.com/paysuper/paysuper-management-api)](https://goreportcard.com/report/github.com/paysuper/paysuper-management-api) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/paysuper/paysuper-management-api)

PaySuper is a unique, simple payment toolkit designed to make developers self-reliant. It’s an open-source payment service with a highly customizable payment form, an intuitive API, and comprehensible, eye-catching reports.

PaySuper Management API is a REST API backend for the [Dashboard](https://github.com/paysuper/paysuper-management-server) and the [Payment Form](https://github.com/paysuper/paysuper-payment-form). And you can just proxy all requests to the [Billing Server](https://github.com/paysuper/paysuper-billing-server) micro-service that will provide you with any payment processing business logic.

## Table of Contents

- [Developing](#developing)
    - [Branches](#branches)
    - [Built With](#built-with)
    - [Prerequisites](#prerequisites)
    - [Setting up Dev](#setting-up-dev)
    - [Building](#building)
    - [Deploying](#deploying)
- [Versioning](#versioning)
- [Configuration](#configuration)
- [Tests](#tests)
- [API Reference](#api-reference)
- [Tests](#tests)
- [Terms](#terms)
- [Contributing](#contributing)
- [License](#license)

## Developing

### Branches

The `master` branch of this repository contains the latest stable release of this component.

### Built With

??? List main libraries, frameworks used including versions (React, Angular etc...)

### Prerequisites

??? What is needed to set up the dev environment. For instance, global dependencies or any other tools. include download links.

### Setting up Dev

??? Here's a brief intro about what a developer must do in order to start developing
the project further:

```shell
git clone https://github.com/your/your-project.git
cd your-project/
packagemanager install
```

And state what happens step-by-step. If there is any virtual environment, local server or database feeder needed, explain here.

### Building

??? If your project needs some additional steps for the developer to build the
project after some code changes, state them here. for example:

```shell
./configure
make
make install
```

Here again you should state what actually happens when the code above gets
executed.

### Deploying

??? Give instructions on how to build and release a new version
In case there's some step you have to take that publishes this project to a
server, this is the right time to state it.

```shell
packagemanager deploy your-project -s server.com -u username -p password
```

And again you'd need to tell what the previous code actually does.

## Versioning

PaySuper Management API uses the endpoint versioning. The current version is `/v1`.

## Configuration

??? Here you should write what are all of the configurations a user can enter when
using the project.

## Tests

Every API method is covered by tests. The tests classes located in the same directory `internal/handlers` with code classes and have suffix `_test` at the end of its titles.

Test resources located in the `test` directory.

## API Reference

PaySuper Management API consists of public API methods which starts from the `/api/v1/` and are documented in the [API Reference](https://docs.stg.pay.super.com/api).

This project also contains internal API methods which starts from `/system/` and `/admin/`.

## Terms

### Accounting currency

`PSP Currency` - used to save the amount of the payment transaction in the PSP accounting currency. PSP currency can be set using the environment variable named "PSP_ACCOUNTING_CURRENCY"

`Merchant currency` -  used to save the amount of the payment transaction in the merchant (projects owner) accounting currency. Merchant currency can be set using merchant settings in PSP control panel.

`Payment system currency` - used to save the amount of the payment transaction in the payment system (payment methods owner) accounting currency. Payment system currency can be set using payment system settings in PSP admin panel.

## Contributing

If you like this project then you can put a ⭐️ on it.

We welcome contributions to PaySuper of any kind including documentation, suggestions, bug reports, pull requests etc. We would love to hear from you. In general, we follow the "fork-and-pull" Git workflow.

We feel that a welcoming community is important and we ask that you follow the PaySuper's [Open Source Code of Conduct](https://github.com/paysuper/code-of-conduct/blob/master/README.md) in all interactions with the community.

## License

The project is available as open source under the terms of the [GPL v3 License](https://www.gnu.org/licenses/gpl-3.0).