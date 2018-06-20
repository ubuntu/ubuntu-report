package metrics

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type filterResult struct {
	r   []string
	err error
}

func filter(r io.Reader, regex string, getAll bool) <-chan filterResult {
	re := regexp.MustCompile(regex)
	scanner := bufio.NewScanner(r)
	results := make(chan filterResult)

	go func(r io.Reader) {
		defer close(results)

		for scanner.Scan() {
			m := re.FindStringSubmatch(scanner.Text())
			if m != nil {

				if getAll {
					results <- filterResult{r: m[1:]}
					continue
				}

				// we want the first non empty match (can be null in case of alternative regexp)
				var match string
				for i := 1; match == "" && i < len(m); i++ {
					match = strings.TrimSpace(m[i])
				}
				results <- filterResult{r: []string{match}}
			}
		}

		if err := scanner.Err(); err != nil {
			results <- filterResult{err: errors.Wrap(err, "error while scanning")}
		}
	}(r)

	return results
}

func filterFirst(r io.Reader, regex string, notFoundOk bool) (string, error) {
	result := <-filter(r, regex, false)
	if !notFoundOk && result.err == nil && len(result.r) < 1 {
		result.err = errors.Errorf("couldn't find any line matching %s", regex)
	}
	if len(result.r) < 1 {
		result.r = []string{""}
	}
	return result.r[0], result.err
}

func filterAll(r io.Reader, regex string) ([]string, error) {
	var results []string
	for result := range filter(r, regex, false) {
		if result.err != nil {
			return nil, result.err
		}
		results = append(results, result.r[0])
	}

	if len(results) < 1 {
		return nil, errors.Errorf("couldn't find any line matching %s", regex)
	}

	return results, nil
}
