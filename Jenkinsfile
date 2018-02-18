pipeline {
  agent none

  environment {
    CONTAINER_TAG = 'latest'
    GIT_CREDENTIALS = "ssh-key-jenkins-bot"
  }

  stages {
    stage('Build') {
      agent {
        docker {
          image 'docker.io/controlplane/gcloud-sdk:latest'
          args '-v /var/run/docker.sock:/var/run/docker.sock ' +
            '--user=root ' +
            '--cap-drop=ALL ' +
            '--cap-add=DAC_OVERRIDE'
        }
      }

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
        DOCKER_HUB_PASSWORD = credentials('docker-hub-controlplane')
      }

      steps {
        ansiColor('xterm') {
          sh 'echo "${DOCKER_HUB_PASSWORD}" | docker login ' +
            '--username "controlplane" ' +
            '--password-stdin'
          sh 'make push CONTAINER_TAG="${CONTAINER_TAG}"'
        }
      }
    }
  }
}
