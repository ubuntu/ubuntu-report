## How to contribute
We love your contributions. Take a look at these guidelines to know
how to contribute to ubuntu-report.

## Getting started
Make sure you have the repo cloned under `$GOPATH/src` by using the `go get ... ` command.
You can learn more about this [here](https://golang.org/cmd/go/#hdr-Download_and_install_packages_and_dependencies)

## Formatting using gofmt
`gofmt` is a tool to format your Go code. It makes your code: 
* **Easier to write**: never worry about minor formatting concerns while hacking away,
* **Easier to read**: when all code looks the same you need not mentally convert others' formatting style into something you can understand.
* **Easier to maintain**: mechanical changes to the source don't cause unrelated changes to the file's formatting; diffs show only the real changes.
* **Uncontroversial**: never have a debate about spacing or brace position ever again!

click [here](https://blog.golang.org/go-fmt-your-code) to learn how to gofmt your code.

## Writing your tests
Make sure to accompany all your code with corresponding tests with the highest possible code coverage. This way we can ensure that your code works and life gets easier for everyone. You can place your test file right alongside your code file and run the test using
```
go test path/to/directory
```
where `path/to/directory` is the directory with your code and test file(s).

## CI tests should pass
When your tests pass locally, then you can make a PR. If the CI tests pass, your PR will be merged. Else, you may have to make a few changes to it.