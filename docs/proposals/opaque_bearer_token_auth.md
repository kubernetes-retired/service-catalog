# Overview
This proposal adds an opaque bearer token support in addition to 
currently supported basic auth.

# Problem
In addition to basic auth, we need to support other kinds of authentication for brokers. 
`Bearer` token is one of the most common authentication schemes, which can be supported by
Service Catalog in a generic and easy way described below.
This generic approach should be compatible with any authentication protocol using `Bearer` tokens:
- OAuth 2.0
- Any JWT-based protocol, including vendor-specific ones

# Background
The prerequisite for adding any authentication schemes other than basic auth, is the proposal 
[additional_auth.md](https://github.com/rifung/service-catalog/blob/3c0e8c7d599ea6dc96d0a5aa53ae5fbe7107cb89/docs/proposals/additional_auth.md)
which provides the generic `AuthInterface` that can be implemented for adding new auth schemes:
```go
type AuthInterface interface {
	AddAuth(*http.Request) error
	ParseSecret(kubernetes.Interface, Secret) error
}
```
To distinguish different auth schemes used by brokers, we need a `SecretType` field:
```go
type BrokerAuthInfo struct {
	AuthSecret *v1.ObjectReference
	SecretType string // "basic", "opaque"
}
```

# Proposal

## BrokerAuthInfo
Define the "`opaque`" constant for the `SecretType` field in `BrokerAuthInfo` object as an identifier for 
opaque bearer token.

## Broker Request
Introduce the `OpaqueBearerTokenAuth` implementation of the `AuthInterface` described in [pull request #1037](https://github.com/kubernetes-incubator/service-catalog/pull/1037) 
with the following details:
- Service Catalog examines the `SecretType` field of `BrokerAuthInfo` object, finds the "opaque" value, 
which maps to the corresponding `OpaqueBearerTokenAuth` implementation of the `AuthInterface` responsible for this 
type of secret
- Service Catalog fetches the Secret the reference to which is specified in the Broker spec, and extract the token 
stored in the "`token`" key of the secret (by convention)
- Service Catalog adds an `Authorization` header to every broker request as `Authorization: Bearer $TOKEN`

## Token Generation and Maintenance
Token generation and maintenance is out of scope of Service Catalog. Possible ways of covering this part:
- User manually fetches the token and creates a secret containing it
  - works only for long-lived tokens ("static" token)
- There is a separate "Token Maintainer" process watching on all brokers and periodically creating/updating the secrets 
with new tokens to mitigate their expiration (short-lived tokens could be valid only for 10 minutes, for example)
  - Token Maintainer process maintains secrets only for the brokers with `SecretType` set to "opaque"
  - Token Maintainer process is free to use any extra features it needs, such as labels and annotations on `Broker` 
  objects for extra filtering and specific parameters

# Limitations
- There is a single token per broker supported only, which suggests that a token should not have any specifics of 
particular request type, and should be suitable for any request to the broker.
- A single token could be used multiple times (until it expires)