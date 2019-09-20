---
title: Serving Certificates, Authentication, and Authorization
layout: docwithnav
---

This document outlines how the service catalog handles authentication.

The service catalog Helm chart's defaults paired with most Kubernetes
distributions will automatically set up all authentication details correctly. 
This documentation, therefore, exists for the benefit of
those who wish to develop and advanced understanding of this topic and those
who have a need to address various outlying scenarios.

## Certificates Overview

Certificates are used for parties to identify themselves to one another.

CA (Certificate Authority) certificates are used to delegate trust.
Whenever something trusts the CA, it can trust any certificates *signed*
by the CA.

If a certificate is not signed by a separate CA, it is instead
*self-signed*. A self-signed certificate must either be trusted directly
(instead of being trusted indirectly by trusting a CA), or not trusted at
all.  Generally, our client CA certificates will be self-signed, since
they represent the "root" of our trust relationship: clients must
inherently trust the CA.

The service catalog Helm chart automatically generates new CA. 
This CA signs "serving" certificates, which are used to encrypt communication 
over HTTPS.There is generally no need to override this.

### Generating certificates

In the common case  CA certificate referenced above already
exist as part of the installation.

In case you need to generate any of the CA certificate pairs mentioned
above yourself, the Kubernetes documentation has [detailed
instructions](https://kubernetes.io/docs/admin/authentication/#creating-certificates)
on how to create certificates several different ways.

Service Catalog Helm chart uses build-in [Sprig’s](https://github.com/Masterminds/sprig) functions to generate 
all needed certificates used by Webhook Server:  
```
{{- $ca := genCA "service-catalog-webhook-ca" 3650 }}
{{- $cn := printf "%s-webhook" (include "fullname" .) }}
{{- $altName1 := printf "%s.%s" $cn .Release.Namespace }}
{{- $altName2 := printf "%s.%s.svc" $cn .Release.Namespace }}
{{- $cert := genSignedCert $cn nil (list $altName1 $altName2) 3650 $ca }}
```

## Authentication

CRDs always use the same authentication and authorization as the built-in resources of your API Server.
If you use RBAC for authorization, most RBAC roles will not grant access to the new resources (except the cluster-admin role or any role created with wildcard rules). 
You’ll need to explicitly grant access to the new resources.

### Client Certificate Authentication

Client certificate authentication authenticates clients who connect using
certificates signed by a given CA (as specified by the *client CA
certificate*).  This same mechanism is also generally used by the main
Kubernetes API server.

Generally, the default admin user in a cluster connects with client
certificate authentication.  Additionally, off-cluster non-human clients
often use client certificate authentication.

See the [x509 client
certificates](https://kubernetes.io/docs/admin/authentication/#x509-client-certs)
section of the Kubernetes documentation for more information.

