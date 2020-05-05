// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import (
	"io/ioutil"
	"os"
	"reflect"
)

// Patch sets a variable to a temporary value for the duration of the test.
//
// It sets the value pointed to by the given destination to the given
// value, which must be assignable to the element type of the destination.
//
// At the end of the test (see "Deferred execution" in the package docs), the
// destination is set back to its original value.
func (c *C) Patch(dest, value interface{}) {
	destv := reflect.ValueOf(dest).Elem()
	oldv := reflect.New(destv.Type()).Elem()
	oldv.Set(destv)
	valuev := reflect.ValueOf(value)
	if !valuev.IsValid() {
		// This isn't quite right when the destination type is not
		// nilable, but it's better than the complex alternative.
		valuev = reflect.Zero(destv.Type())
	}
	destv.Set(valuev)
	c.cleanup(func() {
		destv.Set(oldv)
	})
}

// Setenv sets an environment variable to a temporary value for the
// duration of the test.
//
// At the end of the test (see "Deferred execution" in the package docs), the
// environment variable is returned to its original value.
func (c *C) Setenv(name, val string) {
	c.setenv(name, val, true)
}

// Unsetenv unsets an environment variable for the duration of a test.
func (c *C) Unsetenv(name string) {
	c.setenv(name, "", false)
}

// setenv sets or unsets an environment variable to a temporary value for the
// duration of the test
func (c *C) setenv(name, val string, valOK bool) {
	oldVal, oldOK := os.LookupEnv(name)
	if valOK {
		os.Setenv(name, val)
	} else {
		os.Unsetenv(name)
	}
	c.cleanup(func() {
		if oldOK {
			os.Setenv(name, oldVal)
		} else {
			os.Unsetenv(name)
		}
	})
}

// Mkdir makes a temporary directory and returns its name.
//
//
// At the end of the test (see "Deferred execution" in the package docs), the
// directory and its contents are removed.
func (c *C) Mkdir() string {
	name, err := ioutil.TempDir("", "quicktest-")
	c.Assert(err, Equals, nil)
	c.cleanup(func() {
		err := os.RemoveAll(name)
		c.Check(err, Equals, nil)
	})
	return name
}

// cleanup uses Cleanup when it can, falling back to using Defer.
func (c *C) cleanup(f func()) {
	if tb, ok := c.TB.(cleaner); ok {
		tb.Cleanup(f)
	} else {
		c.Defer(f)
	}
}
