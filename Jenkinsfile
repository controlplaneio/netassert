GIT_CREDENTIALS = "ssh-key-jenkins-bot"

node {
  ansiColor('xterm') {

    stage('Checkout') {
      git url: 'ssh://git@github.com:controlplaneio/netassert',
        changelog: false,
        branch: 'master',
        credentialsId: "${GIT_CREDENTIALS}"
      }

    stage('Build') {
      sh "command -v make &>/dev/null || yum install -yt make"
      sh "make jenkins"
    }

  }
}
