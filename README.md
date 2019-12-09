# PaySuper Management API

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-brightgreen.svg)](https://www.gnu.org/licenses/gpl-3.0) 
[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/paysuper/paysuper-management-api/issues)
[![Build Status](https://travis-ci.org/paysuper/paysuper-management-api.svg?branch=master)](https://travis-ci.org/paysuper/paysuper-management-api) 
[![codecov](https://codecov.io/gh/paysuper/paysuper-management-api/branch/master/graph/badge.svg)](https://codecov.io/gh/paysuper/paysuper-management-api) 
[![Go Report Card](https://goreportcard.com/badge/github.com/paysuper/paysuper-management-api)](https://goreportcard.com/report/github.com/paysuper/paysuper-management-api) 
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/paysuper/paysuper-management-api)

PaySuper is a unique, simple payment toolkit designed to make developers self-reliant. It’s an open-source payment service with a highly customizable payment form, an intuitive API, and comprehensible, eye-catching reports.

PaySuper Management API is a REST API backend for the [Dashboard](https://github.com/paysuper/paysuper-management-server) and the [Payment Form](https://github.com/paysuper/paysuper-payment-form).

|   | PaySuper Service Architecture
:---: | :---
✨ | **Checkout integration.** [PaySuper JS SDK](https://github.com/paysuper/paysuper-js-sdk) is designed to integrate a Checkout Form on a merchant's website or a game client.
💵 | **Frontend for a payment form.** [PaySuper Checkout Form](https://github.com/paysuper/paysuper-payment-form) is a frontend for a sigle-page application with a payment form.
📊 | **Frontend for a merchant.** [PaySuper Dashboard](https://github.com/paysuper/paysuper-dashboard) is the BFF server and frontend to interact with all PaySuper related features for merchants.
🔧 | **API Backend.** [PaySuper Management API](https://github.com/paysuper/paysuper-management-api) is a REST API backend for the [PaySuper Dashboard](https://github.com/paysuper/paysuper-management-server) and the [PaySuper Checkout Form](https://github.com/paysuper/paysuper-payment-form). Public API methods are documented in the [API Reference](https://docs.pay.super.com/api).
💳 | **Payment processing.** [Billing Server](https://github.com/paysuper/paysuper-billing-server) is a micro-service that provides with any payment processing business logic.

***

## Table of Contents

- [API Reference](#api-reference)
- [Developing](#developing)
    - [Branches](#branches)
    - [Versioning](#versioning)
- [Tests](#tests)
- [Terms](#terms)
- [Contributing](#contributing-support-feature-requests)
- [License](#license)

## API Reference

PaySuper Management API consists of public API methods which paths start with the `/api/v1/` and are documented in the [API Reference](https://docs.pay.super.com/api).

This project also contains internal API methods which paths start with `/system/` and `/admin/`.

## Developing

### Branches

We use the [GitFlow](https://nvie.com/posts/a-successful-git-branching-model) as a branching model for Git.

### Versioning

PaySuper Management API uses the endpoint versioning. The current version is `/v1`.

## Tests

Every API method is covered by tests. The tests classes located in the same directory `internal/handlers` with code classes and have suffix `_test` at the end of its titles.

Test resources located in the `test` directory.

## Terms

### Accounting currency

`PSP Currency` - used to save the amount of the payment transaction in the PSP accounting currency. PSP currency can be set using the environment variable named "PSP_ACCOUNTING_CURRENCY"

`Merchant currency` -  used to save the amount of the payment transaction in the merchant (projects owner) accounting currency. Merchant currency can be set using merchant settings in PSP control panel.

`Payment system currency` - used to save the amount of the payment transaction in the payment system (payment methods owner) accounting currency. Payment system currency can be set using payment system settings in PSP admin panel.

## Contributing, Support, Feature Requests
If you like this project then you can put a ⭐️ on it. It means a lot to us.

If you have an idea of how to improve PaySuper (or any of the product parts) or have general feedback, you're welcome to submit a [feature request](../../issues/new?assignees=&labels=&template=feature_request.md&title=).

Chances are, you like what we have already but you may require a custom integration, a special license or something else big and specific to your needs. We're generally open to such conversations.

If you have a question and can't find the answer yourself, you can [raise an issue](../../issues/new?assignees=&labels=&template=support-request.md&title=I+have+a+question+about+%3Cthis+and+that%3E+%5BSupport%5D) and describe what exactly you're trying to do. We'll do our best to reply in a meaningful time.

We feel that a welcoming community is important and we ask that you follow PaySuper's [Open Source Code of Conduct](https://github.com/paysuper/code-of-conduct/blob/master/README.md) in all interactions with the community.

PaySuper welcomes contributions from anyone and everyone. Please refer to [our contribution guide to learn more](CONTRIBUTING.md).

## License

The project is available as open source under the terms of the [GPL v3 License](https://www.gnu.org/licenses/gpl-3.0).
