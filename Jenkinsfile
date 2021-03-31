pipeline {
    agent any
    tools {
        go 'go-1.16.2'
    }
    stages {
        stage('lint') {
            steps {
                sh 'go vet ./...'
            }
        }
        stage('build') {
            steps {
                sh 'CGO_ENABLED=0 go build ./...'
            }
        }
        stage('test') {
            steps {
                sh 'CGO_ENABLED=0 go test -cover ./...'
            }
        }
    }
}
