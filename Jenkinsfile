pipeline {
  agent none

  environment {
    CONTAINER_TAG = 'latest'
    ENVIRONMENT = 'ops'
    GIT_CREDENTIALS = "ssh-key-jenkins-bot"
  }

  // stages is "all pipeline stages"
  stages {
    // the name of this stage, represented in the stage view e.g. https://jenkins.ctlplane.io/job/netassert/
    stage('Build') {
      // defines the "agent" aka "jenkins slave"
      agent {
        docker {
          // always run in this image, it's got latest kubectl and is based from a google-managed image
          image 'docker.io/controlplane/gcloud-sdk:latest'
          args '-v /var/run/docker.sock:/var/run/docker.sock ' +
            '--user=root ' +
            '--cap-drop=ALL ' +
            '--cap-add=DAC_OVERRIDE'
        }
      }

      // here is the actual build for this stage
      steps {
        ansiColor('xterm') {
          sh 'make build CONTAINER_TAG="${CONTAINER_TAG}"'
        }
      }
    }

    stage('Test') {
      agent {
        docker {
          image 'docker.io/controlplane/gcloud-sdk:latest'
          args '-v /var/run/docker.sock:/var/run/docker.sock ' +
            '--user=root ' +
            '--cap-drop=ALL ' +
            '--cap-add=DAC_OVERRIDE'
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

    stage('Push') {
      agent {
        docker {
          image 'docker.io/controlplane/gcloud-sdk:latest'
          args '-v /var/run/docker.sock:/var/run/docker.sock ' +
            '--user=root ' +
            '--cap-drop=ALL ' +
            '--cap-add=DAC_OVERRIDE'
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

            make push CONTAINER_TAG='${CONTAINER_TAG}'
          """
        }
      }
    }
  }
}
