package metrics

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type filterResult struct {
	r   string
	err error
}

func filter(r io.Reader, regex string) <-chan filterResult {
	re := regexp.MustCompile(regex)
	scanner := bufio.NewScanner(r)
	results := make(chan filterResult)

	go func(r io.Reader) {
		defer close(results)

		for scanner.Scan() {
			m := re.FindStringSubmatch(scanner.Text())
			if m != nil {
				// we want the first non empty match (can be null in case of alternative regexp)
				var match string
				for i := 1; match == "" && i < len(m); i++ {
					match = strings.TrimSpace(m[i])
				}
				results <- filterResult{r: match}
			}
		}

		if err := scanner.Err(); err != nil {
			results <- filterResult{err: errors.Wrap(err, "error while scanning")}
		}
	}(r)

	return results
}

func filterFirst(r io.Reader, regex string, notFoundOk bool) (string, error) {
	result := <-filter(r, regex)
	if !notFoundOk && result.err == nil && result.r == "" {
		result.err = errors.Errorf("couldn't find any line matching %s", regex)
	}
	return result.r, result.err
}

func filterAll(r io.Reader, regex string) ([]string, error) {
	var results []string
	for result := range filter(r, regex) {
		if result.err != nil {
			return nil, result.err
		}
		results = append(results, result.r)
	}

	if len(results) < 1 {
		return nil, errors.Errorf("couldn't find any line matching %s", regex)
	}

	return results, nil
}
