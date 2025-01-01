# Contributing Guidelines

## Overview
Thank you for your interest in contributing to the "Minds Toolkit" project! By following
this guideline, we hope to ensure that your contributions are made smoothly and
efficiently. The Go Minds project is licensed under the [Apache 2.0
License](https://github.com/chriscow/minds/blob/master/LICENSE), and we
welcome contributions through GitHub pull requests.

## Reporting Bugs
If you discover a bug, first check the [GitHub Issues page](https://github.com/chriscow/minds/issues) to see if the issue has already been reported. If you're reporting a new issue, please use the "Bug report" template and provide detailed information about the problem, including steps to reproduce it.

## Suggesting Features
If you want to suggest a new feature or improvement, first check the [GitHub Issues page](https://github.com/chriscow/minds/issues) to ensure a similar suggestion hasn't already been made. Use the "Feature request" template to provide a detailed description of your suggestion.

## Reporting Vulnerabilities
If you identify a security concern, please use the "Report a security vulnerability" template on the [GitHub Issues page](https://github.com/chriscow/minds/issues) to share the details. This report will only be viewable to repository maintainers. You will be credited if the advisory is published.

## Questions for Users
If you have questions, please utilize the [GitHub Discussions page](https://github.com/chriscow/minds/discussions).

## Contributing Code
There might already be a similar pull requests submitted! Please search for [pull requests](https://github.com/chriscow/minds/pulls) before creating one.

### Requirements for Merging a Pull Request

The requirements to accept a pull request are as follows:

- All pull requests should be written in Go according to common conventions, formatted with `goimports`, and free of warnings from tools like `golangci-lint`.
- Include tests and ensure all tests pass.
- Maintain test coverage without any reduction.
- All pull requests require approval from at least one Thoughtnet Minds maintainer.

**Note:**  
The merging method for pull requests in this repository is squash merge.

### Creating a Pull Request
- Fork the repository.
- Create a new branch and commit your changes.
- Push that branch to GitHub.
- Start a new Pull Request on GitHub. (Please use the pull request template to provide detailed information.)

**Note:**  
If your changes introduce breaking changes, please prefix your pull request title with "[BREAKING_CHANGES]".

### Code Style
In this project, we adhere to the standard coding style of Go. Your code should
maintain consistency with the rest of the codebase. To achieve this, please
format your code using tools like `goimports` and resolve any syntax or style
issues with `golangci-lint`.

**Run goimports:**
```
go install golang.org/x/tools/cmd/goimports@latest
```

```
goimports -w .
```

**Run golangci-lint:**
```
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

```
golangci-lint run --out-format=github-actions
```

### Unit Test
Please create or update tests relevant to your changes. Ensure all tests run
successfully to verify that your modifications do not adversely affect other
functionalities.

**Run test:**
```
go test -v ./...
```

### Integration Test
Integration tests are requested against the production version of the OpenAI
API. These tests will verify that the library is properly coded against the
actual behavior of the API, and will fail upon any incompatible change in the
API. I use the `_examples` as integration tests as well.

**Notes:** These tests send real network traffic to the OpenAI and Google Gemini API and may reach
rate limits. Temporary network problems may also cause the test to fail.

**Run integration test:**
```
OPENAI_API_KEY=XXX GEMINI_API_KEY=XXX make run-examples
```

---

We wholeheartedly welcome your active participation. Let's build an amazing project together!
