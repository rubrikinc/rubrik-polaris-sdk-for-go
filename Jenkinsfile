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
        go 'go-1.17.3'
    }
    triggers {
        cron(env.BRANCH_NAME == 'main' ? '@midnight' : '')
    }
    environment {
        // Polaris credentials.
        RUBRIK_POLARIS_SERVICEACCOUNT_FILE = credentials('tf-sdk-test-polaris-service-account')

        // AWS credentials.
        TEST_AWSACCOUNT_FILE  = credentials('tf-sdk-test-aws-account')
        AWS_ACCESS_KEY_ID     = credentials('tf-sdk-test-access-key')
        AWS_SECRET_ACCESS_KEY = credentials('tf-sdk-test-secret-key')
        AWS_DEFAULT_REGION    = 'us-east-2'

        // Azure credentials.
        TEST_AZURESUBSCRIPTION_FILE     = credentials('tf-sdk-test-azure-subscription')
        AZURE_SERVICEPRINCIPAL_LOCATION = credentials('tf-sdk-test-azure-service-principal')

        // GCP credentials.
        TEST_GCPPROJECT_FILE           = credentials('tf-sdk-test-gcp-project')
        GOOGLE_APPLICATION_CREDENTIALS = credentials('tf-sdk-test-gcp-service-account')

        // Run integration tests with the nightly build.
        TEST_INTEGRATION = currentBuild.getBuildCauses('hudson.triggers.TimerTrigger$TimerTriggerCause').size()
    }
    stages {
        stage('Lint') {
            steps {
                sh 'go vet ./...'
                sh 'go run honnef.co/go/tools/cmd/staticcheck@latest ./...'
                sh 'bash -c "diff -u <(echo -n) <(gofmt -d .)"'
            }
        }
        stage('Build') {
            steps {
                sh 'CGO_ENABLED=0 go build ./...'
            }
        }
        stage('Pre-test') {
            when { not { environment name: 'TEST_INTEGRATION', value: '0' } }
            steps {
                sh 'go run ./internal/cmd/testenv -precheck'
            }
        }
        stage('Test') {
            steps {
                sh 'CGO_ENABLED=0 go test -count=1 -coverprofile=coverage.txt -timeout=120m -v ./...'
            }
        }
        stage('Coverage') {
            environment {
                GOPATH = "/tmp/go"
            }
            steps {
                sh 'go run github.com/t-yuki/gocover-cobertura@latest < coverage.txt > coverage.xml'
                cobertura coberturaReportFile: 'coverage.xml'
            }
        }
    }
    post {
        always {
            script {
                if (currentBuild.getBuildCauses('hudson.triggers.TimerTrigger$TimerTriggerCause').size() > 0) {
                    sh 'go run ./internal/cmd/testenv -cleanup'
                }
            }
        }
        success {
            script {
                if (currentBuild.getBuildCauses('hudson.triggers.TimerTrigger$TimerTriggerCause').size() > 0) {
                    slackSend(
                        channel: '#rubrik-polaris-sdk-for-go',
                        color: 'good',
                        message: "The pipeline ${currentBuild.fullDisplayName} succeeded (runtime: ${currentBuild.durationString.minus(' and counting')})\n${currentBuild.absoluteUrl}"
                    )
                    slackSend(
                        channel: '#terraform-provider-development',
                        color: 'good',
                        message: "The pipeline ${currentBuild.fullDisplayName} succeeded (runtime: ${currentBuild.durationString.minus(' and counting')})\n${currentBuild.absoluteUrl}"
                    )
                }
            }
        }
        failure {
            script {
                if (currentBuild.getBuildCauses('hudson.triggers.TimerTrigger$TimerTriggerCause').size() > 0) {
                    slackSend(
                        channel: '#rubrik-polaris-sdk-for-go',
                        color: 'danger',
                        message: "The pipeline ${currentBuild.fullDisplayName} failed (runtime: ${currentBuild.durationString.minus(' and counting')})\n${currentBuild.absoluteUrl}"
                    )
                    slackSend(
                        channel: '#terraform-provider-development',
                        color: 'danger',
                        message: "The pipeline ${currentBuild.fullDisplayName} failed (runtime: ${currentBuild.durationString.minus(' and counting')})\n${currentBuild.absoluteUrl}"
                    )
                }
            }
        }
    }
}
