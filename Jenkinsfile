#!/usr/bin/env groovy

agentConfigImage = 'docker.io/controlplane/gcloud-sdk:latest'
agentConfigArgs = '-v /var/run/docker.sock:/var/run/docker.sock ' +
  '--user=root ' +
  '--cap-drop=ALL ' +
  '--cap-add=DAC_OVERRIDE ' +
  '--cap-add=CHOWN'

def getDockerImageTag() {
  if (env.GIT_COMMIT == "") {
    error "GIT_COMMIT value was empty at usage. "
  }
  return "${env.BUILD_ID}-${env.GIT_COMMIT}"
}

pipeline {
  agent none

  environment {
    DOCKER_IMAGE_TAG = "${getDockerImageTag()}"
    ENVIRONMENT = 'ops'
    GIT_CREDENTIALS = "ssh-key-jenkins-bot"
  }

  post {
    always {
      node("master") {
        step([$class: 'ClaimPublisher'])
      }
    }
  }

  // stages is "all pipeline stages"
  stages {
    // the name of this stage, represented in the stage view e.g. https://jenkins.ctlplane.io/job/netassert/
    stage('Build') {
      // defines the "agent" aka "jenkins slave"
      agent {
        docker {
          image agentConfigImage
          args agentConfigArgs
        }
      }

      // here is the actual build for this stage
      steps {
        ansiColor('xterm') {
          sh 'make build CONTAINER_TAG="${DOCKER_IMAGE_TAG}"'
        }
      }
    }

    stage('Test - host: localhost') {
      agent {
        docker {
          image agentConfigImage
          args agentConfigArgs
        }
      }

      options {
        timeout(time: 15, unit: 'MINUTES')
        retry(2)
        timestamps()
      }

      environment {
        HOME = "/tmp/home/"
        TEST_FILE = "test/test-localhost-remote.yaml"
      }

      steps {
        ansiColor('xterm') {
          sh "make jenkins TEST_FILE=${TEST_FILE}"
        }
      }
    }

    stage('Test - k8s: GKE') {
      agent {
        docker {
          image agentConfigImage
          args agentConfigArgs
        }
      }

      options {
        timeout(time: 15, unit: 'MINUTES')
        retry(2)
        timestamps()
      }

      environment {
        HOME = "/tmp/home/"
        TEST_FILE = "test/test-localhost-remote.yaml"
      }

      steps {
        ansiColor('xterm') {
          sh """#!/bin/bash
            set -euxo pipefail

            EXIT_CODE=0

            if ! make cluster; then
              make cluster-kill cluster
            fi

            if ! make test; then
              EXIT_CODE=1
            fi

            make kill-cluster

            exit \${EXIT_CODE}
          """
        }
      }
    }

    stage('Push') {
      agent {
        docker {
          image agentConfigImage
          args agentConfigArgs
        }
      }

      environment {
        DOCKER_REGISTRY_CREDENTIALS = credentials("${ENVIRONMENT}_docker_credentials")
      }

      steps {
        ansiColor('xterm') {
          sh """
            echo '${DOCKER_REGISTRY_CREDENTIALS_PSW}' \
            | docker login \
              --username '${DOCKER_REGISTRY_CREDENTIALS_USR}' \
              --password-stdin

            make push CONTAINER_TAG='${DOCKER_IMAGE_TAG}'
          """
        }
      }
    }
  }
}
