// Copyright (c) 2019 Ashley Jeffs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package test

import (
	"fmt"

	"golang.org/x/sync/errgroup"
)

//------------------------------------------------------------------------------

// Definition of a group of tests for a Benthos config file.
type Definition struct {
	Parallel bool   `yaml:"parallel"`
	Cases    []Case `yaml:"tests"`
}

// ExampleDefinition returns a Definition containing an example case.
func ExampleDefinition() Definition {
	return Definition{
		Parallel: true,
		Cases:    []Case{NewCase()},
	}
}

//------------------------------------------------------------------------------

// Execute attempts to run a test definition on a target config file. Returns
// an array of test failures or an error.
func (d Definition) Execute(filepath string) ([]CaseFailure, error) {
	procsProvider := NewProcessorsProvider(filepath)
	if d.Parallel {
		// Warm the cache of processor configs.
		for _, c := range d.Cases {
			if _, err := procsProvider.getConfs(c.TargetProcessors, c.Environment); err != nil {
				return nil, err
			}
		}
	}

	var totalFailures []CaseFailure
	if !d.Parallel {
		for i, c := range d.Cases {
			failures, err := c.Execute(procsProvider)
			if err != nil {
				return nil, fmt.Errorf("test case %v failed: %v", i, err)
			}
			totalFailures = append(totalFailures, failures...)
		}
	} else {
		var g errgroup.Group

		failureSlices := make([][]CaseFailure, len(d.Cases))
		for i, c := range d.Cases {
			i := i
			c := c
			g.Go(func() error {
				failures, err := c.Execute(procsProvider)
				if err != nil {
					return fmt.Errorf("test case %v failed: %v", i, err)
				}
				failureSlices[i] = failures
				return nil
			})
		}

		// Wait for all test cases to complete.
		if err := g.Wait(); err != nil {
			return nil, err
		}

		for _, fs := range failureSlices {
			totalFailures = append(totalFailures, fs...)
		}
	}

	return totalFailures, nil
}

//------------------------------------------------------------------------------
