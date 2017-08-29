# Installing Service Catalog on Clusters Running Kubernetes 1.7 and Above

Kubernetes 1.7 or higher clusters run the 
[API Aggregator](https://kubernetes.io/docs/concepts/api-extension/apiserver-aggregation/),
which is a specialized proxy server that sits in front of the core API Server.

The aggregator allows user-defined, Kubernetes compatible API servers to come 
and go inside the cluster, and register themselves on demand to augment the 
externally facing API that kubernetes offers.

Instead of requiring the end-user to access multiple API servers, the API 
aggregation system allows many servers to run inside the cluster, and combines
all of their APIs into one externally facing API. 

This system is very useful from an end-user's perspective, as it allows the 
client to use a single API point with familiar, consistent tooling, 
authentication and authorization.

The Service Catalog utilizes API aggregation to present its API.

# Step 1 - Generate TLS Certificates

We provide a script to do all of the steps needed to set up TLS certificates
that the aggregation system uses. If you'd like to read how to do this setup
manually, please see the 
[manual API aggregation setup document](./manual-api-aggregation-setup.md).

Otherwise, read on for automated instructions.

First, create a directory in which certificates will be generated:

```console
mkdir certs
cd certs
```

We'll assume that you're operating from this `docs/certs` directory for the 
remainder of this document.

Next, install the `cfssl` toolchain (which the following script uses):

```console
go get -u github.com/cloudflare/cfssl/cmd/...
```

Finally, create the certs:

```console
source ../../contrib/svc-cat-apiserver-aggregation-tls-setup.sh
```

# Step 2 - Install the Helm Chart

Use helm to install the Service Catalog, associating it with the
configured name ${HELM_NAME}, and into the specified namespace." This
command also enables authentication and aggregation and provides the
keys we just generated inline.

The installation commands vary slightly between Linux and Mac OS X because of
the versions of the `base64` command (Linux has GNU base64, Mac OS X has BSD 
base64). If you're installing from a Linux based machine, run this:

```
helm install ../../charts/catalog \
    --name ${HELM_RELEASE_NAME} --namespace ${SVCCAT_NAMESPACE} \
    --set apiserver.auth.enabled=true \
        --set useAggregator=true \
        --set apiserver.tls.ca=$(base64 --wrap 0 ${SC_SERVING_CA}) \
        --set apiserver.tls.cert=$(base64 --wrap 0 ${SC_SERVING_CERT}) \
        --set apiserver.tls.key=$(base64 --wrap 0 ${SC_SERVING_KEY})
```

If you're on a Mac OS X based machine, run this:

```
helm install ../../charts/catalog \
    --name ${HELM_RELEASE_NAME} --namespace ${SVCCAT_NAMESPACE} \
    --set apiserver.auth.enabled=true \
        --set useAggregator=true \
        --set apiserver.tls.ca=$(base64 ${SC_SERVING_CA}) \
        --set apiserver.tls.cert=$(base64 ${SC_SERVING_CERT}) \
        --set apiserver.tls.key=$(base64 ${SC_SERVING_KEY})
```
