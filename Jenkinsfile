pipeline {
    agent any
    tools {
        go 'go-1.16.2'
    }
    stages {
        statge('lint') {
            sh 'go vet ./...'
        }
        stage('build') {
            sh 'CGO_ENABLED=0 go build ./...'
        }
        stage('test') {
            sh 'CGO_ENABLED=0 go test -cover ./...'
        }
    }
}
