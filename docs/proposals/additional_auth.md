# Overview
Currently the only authentication method supported by the service catalog is
basic auth. The user is expected to create a secret containing their
credentials which is then read by the service catalog when required.

# Problem
We would like to be able to support brokers which require other kinds of
authentication. Hopefully we can add support in a way that is generic enough
for new types of authentication to be easy to support.

# Background
Currently the way auth works is
* parse the secret information and generate an `AuthConfig` (`getAuthCredentialsFromBroker`)
* use said `AuthConfig` to generate a ClientConfig and subsequently a broker client
* broker client uses the `AuthConfig` to do some auth stuff on each request if necessary (`prepareAndDo`)

AuthConfig is currently a struct

```
type AuthConfig struct {
	BasicAuthConfig *BasicAuthConfig
}
```
The auth information is a struct inside the `BrokerSpec`

```
type BrokerAuthInfo struct {
	BasicAuthSecret *v1.ObjectReference
}
```

# Proposal
We can add support for other types of auth by adding a new field to the
BrokerAuthInfo struct which identifies what type of auth we want to use. For
simplicity, we use a string here. We can just rename `BasicAuthSecret` to be
`AuthSecret`

```
type BrokerAuthInfo struct {
	AuthSecret *v1.ObjectReference
	SecretType string // the supported values are "basic"
}
```

So for example, to use Basic Auth like we currently do, we could just set
`SecretType` to be "basic"

Then, we replace AuthConfig with an interface, `AuthInterface`

```
type AuthInterface interface {
	AddAuth(*http.Request) error
	ParseSecret(kubernetes.Interface, Secret) error
}
```

BasicAuth would then look like this

```
type BasicAuth struct {
	Username string
	Password string
}

func (basic *BasicAuth) AddAuth(req *http.Request) error {
	req.SetBasicAuth(basic.Username, basic.password)
}

func (basic *BasicAuth) ParseSecret(client kubernetes.Interface, authSecret Secret) {
	usernameBytes, ok := authSecret.Data["username"]
	if !ok {
		return nil, fmt.Errorf("auth secret didn't contain username")
	}

	passwordBytes, ok := authSecret.Data["password"]
	if !ok {
		return nil, fmt.Errorf("auth secret didn't contain password")
	}

	basic.Username = string(usernameBytes)
	basic.Password = string(passwordBytes)
}
```

Finally, we would change `getAuthCredentialsFromBroker` so that it looks at the
`SecretType`, calls the constructor for the corresponding `AuthInterface`, and
uses that to parse the secret. Then, this AuthInterface would be passed to the
broker client and the broker client would call AddAuth before making any request.
