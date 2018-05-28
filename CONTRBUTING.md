# How to contribute
We love pull requests from everyone. Please take a look at the contributing
guidelines below before making a pull request.

## Getting started
Make sure you have cloned the project under `$GOPATH/src` by using the `go get` command.

## Formatting your code
This project uses `gofmt` as standard code style. You can learn more about gofmt [here](https://blog.golang.org/go-fmt-your-code).

## Writing tests
Make sure to accompany your code with tests with the highest possible code coverage, that way we
can be sure your code works.

## Running tests
To run all your tests, run 
```
go test ./...
```
Tests should also be tested against raciness by using the `-race` flag

Make sure all the CI tests pass. We won't merge the PR where the CI tests don't pass as the release process would be blocked on it.