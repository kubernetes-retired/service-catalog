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

package model

// Proxy configuration looks like this:
// version: latest
// port: 8080
// backend:
//   ip: localhost
//   port: 8080

type Backend struct {
	Ip   string `json:"ip"`
	Port string `json:"port"`
}

type Proxy struct {
	Version string  `json:"version"`
	Port    string  `json:"port"`
	Backend Backend `json:"backend"`
}
