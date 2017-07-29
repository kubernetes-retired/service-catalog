#!/bin/bash
# sc-setup: implement the steps documented in the "Service-Catalog Quick Deploy".
#           https://docs.google.com/document/d/1mcK4de5OIqpRhBmCf8G4Z-PdXJ5Io1W6L8Gcz5lmAN4/edit?ts=595ea422#heading=h.jvdpxx8x2tol
# Note: the local cluster is assumed to be running.
# Args:
#   $1= path to s-c repo. Default is "~/go/src/k8s.io/service-catalog" or KPATH if defined.

# Executes the passed-in kubectl cmd looking for all containers to be Running and returns an error if
# the loop times-out, or if the cmd generates an error. The expected number of containers is the 2nd
# number in the "x/y" string found under the Ready column.
# args: $1= the command to execute which produces the "x/y" containers ready string for the target pod
#       $2= the total wait time in seconds, default=5
function waitForContainerRunning() {
  cmd="$1"
  maxSleep=${2:-5}
  elapsed=0
  err=0

  for (( elapsed=0; elapsed <= maxSleep; )) ; do
    out="$(eval $cmd)"
    err=$?
    (( err != 0 )) && break
    expect="${out#*/}"
    have="${out%/*}"
    (( have == expect )) && break
    increment=$(((elapsed/15)+1)) # sleep an extra second every 15s
    ((elapsed+=increment))
    echo "...waiting $elapsed sec for $expect containers to be Ready, have $have..."
    sleep $increment
  done

  (( elapsed > maxSleep )) && err=2
  return $err
}

# Executes the passed-in kubectl cmd and compares its output with the passed-in match string.
# Returns an error if the loop times-out or if the cmd generates an error. The output produced by
# the command is expected to be santitized, eg. removing spaces/tabs to better handle json or yaml.
# args: $1= the command to execute which produces the string to be matched against. Note this
#           string is expected to have all whitespaces removed, eg. `tr -d "[:space:]"`
#       $2= the match string
#       $3= resource object type (subject of `kubectl` eg, "pod")
#       $4= the total wait time in seconds, default=5
# Note: the passed-in match string needs to have all spaces removed.
function waitForMatch() {
  cmd="$1"
  match="$2"
  object="$3"
  maxSleep=${4:-5}
  err=0

  for (( elapsed=0; elapsed <= maxSleep; )) ; do
    out="$(eval $cmd)"
    err=$?
    (( err != 0 )) && break
    [[ "$out" == "$match" ]] && break
    increment=$(((elapsed/15)+1)) # sleep an extra second every 15s
    ((elapsed+=increment))
    echo "...waiting $elapsed sec for $object match. Have: \"$out\", expect: \"$match\"..."
    sleep $increment
  done

  (( elapsed > maxSleep )) && err=2
  return $err
}

# Waits for the passed-in cmd to return an exit code of 0, or up to the specified time-out. Returns an
# error if the wait times-out of the cmd returns an error.
# args: $1= the command to execute which produces the "x/y" containers ready string for the target pod
#       $2= the total wait time in seconds, default=5
function waitForCmdSuccess() {
  cmd="$1"
  maxSleep=${2:-5}
  elapsed=0
  err=0

  for (( elapsed=0; elapsed <= maxSleep; )) ; do
    out="$(eval $cmd)"
    err=$?
    (( err == 0 )) && break
    increment=$(((elapsed/15)+1)) # sleep an extra second every 15s
    ((elapsed+=increment))
    echo "...waiting $elapsed sec for cmd \"$cmd\" to succeed..."
    sleep $increment
  done

  (( elapsed > maxSleep )) && err=2
  return $err
}

##
## *** main ***
##

# path to s-c repo
sc_path="$1"
[[ -z "$sc_path" && -n "$KPATH" ]] && sc_path="$(dirname $KPATH)/service-catalog"
[[ -z "$sc_path" ]] && sc_path="~/go/src/k8s.io/service-catalog"

echo
echo "******* begin service-catalog setup ***********"
echo "  assumes local-up-cluster is running"
echo "  s-c repo path: $sc_path"

echo
echo "** Verifying pre-requisits"
if [[ -z "$GOPATH" ]] ; then
  echo "GOPATH variable missing"
  exit 1
fi
if [[ ! -d "$sc_path" ]] ; then
  echo "$sc_path is not a directory or does not exist"
  exit 1
fi
if [[ ! -f "/usr/local/bin/helm" ]] ; then
  echo "helm binary missing, install helm:"
  echo "wget -P /tmp/ https://storage.googleapis.com/kubernetes-helm/helm-v2.5.0-linux-amd64.tar.gz && "
  echo " tar -zxvf /tmp/helm-v2.5.0-linux-amd64.tar.gz -C /tmp/ && "
  echo " mv /tmp/linux-amd64/helm /usr/local/bin/"
  exit 1
fi
if ! rpm -q rpm -q socat >/dev/null ; then
  echo "socket cat pkg missing:"
  echo "yum install -y socat"
  exit 1
fi

cd $sc_path

# check kube-dns is running (assumed to be local cluster)
echo
echo "** ensure kube-dns is running..."
waitForContainerRunning "kubectl get pod -n kube-system | grep kube-dns | awk '{print \$2}'" 30
if (( $? != 0 )) ; then
  echo "not all kube-dns containers are running"
  exit 1
fi

# cluster role steps
echo
echo "** setting up clusterRoleBindings; ignore errors if these roles already exist..."
kubectl create clusterrolebinding add-on-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
kubectl create clusterrolebinding catalog --clusterrole=cluster-admin --serviceaccount=catalog:default
kubectl create clusterrolebinding ups-broker --clusterrole=cluster-admin --serviceaccount=ups-broker:default

# start helm steps
echo
echo "--> helm init"
echo
helm init
waitForContainerRunning "kubectl get pod -n kube-system | grep tiller-deploy | awk '{print \$2}'" 30
if (( $? != 0 )) ; then
  echo "tiller container is not running"
  exit 1
fi

# deploy service catalog
echo
echo "--> helm install charts/catalog --name catalog --namespace catalog"
echo
helm install charts/catalog --name catalog --namespace catalog
waitForContainerRunning "kubectl get pod -n catalog | grep catalog-catalog-apiserver | awk '{print \$2}'" 30
if (( $? != 0 )) ; then
  echo "catalog-apiserver pod is not running"
  exit 1
fi
waitForContainerRunning "kubectl get pod -n catalog | grep catalog-catalog-manager | awk '{print \$2}'" 30
if (( $? != 0 )) ; then
  echo "catalog-manager pod is not running"
  exit 1
fi

# config kubectl to talk to s-c api server in local cluster
ip=$(hostname -i)
kubectl config set-cluster service-catalog --server=http://$ip:30080
kubectl config set-context service-catalog --cluster=service-catalog

# deploy object broker
echo
echo "--> helm install charts/ups-broker --name ups-broker --namespace ups-broker"
echo
helm install charts/ups-broker --name ups-broker --namespace ups-broker
echo
echo "--> kubectl get -n ups-broker service,deployment"
echo
waitForCmdSuccess "kubectl get -n ups-broker service,deployment" 15
(( $? != 0 )) && exit 1
# capture cluster-ip
cluster_ip="$(kubectl get -n ups-broker service | grep ups-broker | awk '{print $3}')"

# create broker resource
echo "--> kubectl --context=service-catalog create -f contrib/examples/walkthrough/ups-broker.yaml"
echo
kubectl --context=service-catalog create -f contrib/examples/walkthrough/ups-broker.yaml
waitForMatch \
  'kubectl --context=service-catalog get brokers ups-broker -o yaml |grep reason: |tr -d "[:space:]"' \
  "reason:FetchedCatalog" "ups-broker" 90
if (( $? != 0 )) ; then
  echo "catalog was not fetched"
  exit 1
fi
echo
echo "--> kubectl --context=service-catalog get serviceclasses"
echo
kubectl --context=service-catalog get serviceclasses

# create service instance of the ServiceClass
echo
echo "--> kubectl create namespace test-ns"
echo
kubectl create namespace test-ns
kubectl --context=service-catalog create -f contrib/examples/walkthrough/ups-instance.yaml
kubectl --context=service-catalog get instances -n test-ns
waitForMatch \
  'kubectl --context=service-catalog get instances -n test-ns -o yaml|grep reason: |tr -d "[:space:]"' \
  "reason:ProvisionedSuccessfully" "service-instance" 30
if (( $? != 0 )) ; then
  echo "ServiceInstance was not created"
  exit 1
fi

# create binding and verify resulting secret
echo
echo "--> kubectl --context=service-catalog create -f contrib/examples/walkthrough/ups-binding.yaml"
echo
kubectl --context=service-catalog create -f contrib/examples/walkthrough/ups-binding.yaml
secretName="$(grep secretName: contrib/examples/walkthrough/ups-binding.yaml | awk '{print $2}')"
waitForCmdSuccess "kubectl -n test-ns get secret $secretName" 15
(( $? != 0 )) && exit 1

echo
echo "******* done! ***********"
echo "Cluster broker ip: $cluster_ip"
echo
