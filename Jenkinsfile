// MIT License
//
// Copyright (c) 2021 Rubrik
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy
//  of this software and associated documentation files (the "Software"), to deal
//  in the Software without restriction, including without limitation the rights
//  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//  copies of the Software, and to permit persons to whom the Software is
//  furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all
//  copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//  SOFTWARE.

pipeline {
    agent any
    tools {
        go 'go-1.16.2'
    }
    triggers {
        cron(env.BRANCH_NAME == 'main' ? '@midnight' : '')
    }
    stages {
        stage('Lint') {
            steps {
                sh 'go vet ./...'
            }
        }
        stage('Build') {
            steps {
                sh 'CGO_ENABLED=0 go build ./...'
            }
        }
        stage('Test') {
            environment {
                // AWS credentials.
                AWS_ACCESS_KEY_ID     = credentials('sdk-aws-access-key')
                AWS_SECRET_ACCESS_KEY = credentials('sdk-aws-secret-key')
                AWS_DEFAULT_REGION    = "us-east-2"

                // GCP credentials.
                GOOGLE_APPLICATION_CREDENTIALS = credentials('sdk-gcp-service-account')

                // Polaris credentials.
                RUBRIK_POLARIS_SERVICEACCOUNT_FILE = credentials('sdk-polaris-service-account')

                // Run integration tests with the nightly build.
                SDK_INTEGRATION = currentBuild.getBuildCauses('hudson.triggers.TimerTrigger$TimerTriggerCause').size()
            }
            steps {
                sh 'CGO_ENABLED=0 go test -count=1 -coverprofile=coverage.txt -v ./...'
            }
        }
        stage('Coverage') {
            environment {
                GOPATH = "/tmp/go"
            }
            steps {
                sh 'go get github.com/t-yuki/gocover-cobertura'
                sh '${GOPATH}/bin/gocover-cobertura < coverage.txt > coverage.xml'
                cobertura coberturaReportFile: 'coverage.xml'
            }
        }
    }
}
