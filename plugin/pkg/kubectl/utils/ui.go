/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import "fmt"

// Green will print the specified string in green text
func Green(str string) string {
	return fmt.Sprintf("\x1b[32;1m%s\x1b[0m", str)
}

// Red will print the specified string in red text
func Red(str string) string {
	return fmt.Sprintf("\x1b[31;1m%s\x1b[0m", str)
}

// Entity will print the specified string in bold text
func Entity(str string) string {
	return fmt.Sprintf("\x1b[36;1m%s\x1b[0m", str)
}

// Error will print the specified error string in red text
func Error(msg string) {
	fmt.Printf("%s\n\n%s\n\n", Red("ERROR"), msg)
}

// Ok will print "OK" in green
func Ok() {
	fmt.Printf("%s\n\n", Green("OK"))
}
