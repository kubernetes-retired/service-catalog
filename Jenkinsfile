#!groovy

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

// Jenkins continuous integration
//
// Parameters Jenkins needs to / can supply:
//
// TEST_PROJECT:   Google Cloud Project ID of the project to use for testing.
// TEST_ZONE:      GCP Zone in which to create test GKE cluster
// TEST_ACCOUNT:   GCP service account credentials (JSON file) to use for testing.
// GERRIT_ACCOUNT: GCP service account credentials (from metadata) to use for
//                 Gerrit API access.

// Verify required parameters
if (! params.TEST_PROJECT) {
  error 'Missing required parameter TEST_PROJECT'
}

if (! params.TEST_ACCOUNT) {
  error 'Missing required parameter TEST_ACCOUNT'
}

if (! params.GERRIT_ACCOUNT) {
  error 'Missing required parameter GERRIT_ACCOUNT'
}

def test_project = params.TEST_PROJECT
def test_account = params.TEST_ACCOUNT
def test_zone    = params.TEST_ZONE ?: 'us-west1-b'
def namespace    = 'catalog'
def root_path    = 'src/github.com/kubernetes-incubator/service-catalog'

def notifyBuild(buildStatus) {
  def gerrit_credentials = params.GERRIT_ACCOUNT
  def gerrit_url         = 'https://plori-review.googlesource.com'

  buildStatus = buildStatus ?: 'SUCCESS'
  def message = "Jenkins ${env.JOB_NAME} build ${currentBuild.displayName} status: " +
      "${buildStatus}. View log at: ${currentBuild.absoluteUrl}consoleFull"

  def BUILD_RESULTS = [SUCCESS: 1, FAILURE: -1, STARTED: 0]
  def verified = BUILD_RESULTS.get(buildStatus)
  if (verified == null) verified = -1

  def label = verified == 0 ? '' : """, "labels": { "Verified": ${verified}}"""
  def payload = """{"message": "${message}"${label}}"""

  withCredentials([[$class: 'UsernamePasswordMultiBinding', credentialsId: "source:${gerrit_credentials}",
      passwordVariable: 'PASS', usernameVariable: 'USER']]) {

    sh """
CHANGE_ID="\$(git log --format=%B -n 1 HEAD | awk '/^Change-Id: / {print \$2}')"
HEAD_SHA="\$(git rev-parse --verify HEAD)"

curl --request POST --silent \
     --header 'Content-type: application/json' \
     --header 'Authorization: Bearer ${env.PASS}' \
     --url "${gerrit_url}/a/changes/\${CHANGE_ID}/revisions/\${HEAD_SHA}/review" \
     --data '${payload}'
"""
  }
}

node {
  // Checkout the source code.
  checkout scm

  env.GOPATH = env.WORKSPACE
  env.ROOT = "${env.WORKSPACE}/${root_path}"
  env.KUBECONFIG = "${env.ROOT}/kubeconfig"

  dir([path: env.ROOT]) {
    // Run build.
    notifyBuild('STARTED')

    def clustername = "jenkins-" + sh([returnStdout: true, script: '''openssl rand \
        -base64 100 | tr -dc a-z0-9 | cut -c -25''']).trim()

    // These are done in parallel since creating the cluster takes a while, and the build
    // doesn't depend on it.
    parallel(
      'Initialize Kubernetes cluster': {
        withCredentials([file(credentialsId: "${test_account}", variable: 'TEST_SERVICE_ACCOUNT')]) {
          sh """${env.ROOT}/script/init_cluster.sh ${clustername} \
                --project ${test_project} \
                --zone ${test_zone} \
                --credentials ${env.TEST_SERVICE_ACCOUNT}"""
        }
      },
      'Build': {
        try {
          sh """${env.ROOT}/script/build.sh --no-docker-compile \
                --project ${test_project}"""
        } catch (Exception e) {
          currentBuild.result = 'FAILURE'
          notifyBuild(currentBuild.result)
        }
      }
    )

    if (currentBuild.result == 'FAILURE') {
      sh """${env.ROOT}/script/cleanup_cluster.sh ${clustername} \
            --project ${test_project} \
            --zone ${test_zone}"""
      error 'Build failed.'
    }

    // Run end-2-end tests on the deployed cluster.
    try {
      sh """${env.ROOT}/script/test_deploy.sh \
            --project ${test_project} \
            --namespace ${namespace}
      """
    } catch (Exception e) {
      currentBuild.result = 'FAILURE'
      notifyBuild(currentBuild.result)
      error 'End-to-end tests failed.'
    } finally {
      sh """${env.ROOT}/script/cleanup_cluster.sh ${clustername} \
            --project ${test_project} \
            --zone ${test_zone}"""
    }

    notifyBuild(currentBuild.result)
  }
}
