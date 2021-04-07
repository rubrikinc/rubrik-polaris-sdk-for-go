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
    environment {
        CGO_ENABLED = 0

        // Schedule a run at midnight for the main branch.
        NIGHTLY = sh(script: 'if [[ $BRANCH_NAME == "main" ]]; then echo "@midnight"; fi', returnStdout: true).trim()
    }
    triggers {
        cron($NIGHTLY)
    }
    stages {
        stage('Lint') {
            steps {
                sh 'go vet ./...'
            }
        }
        stage('Build') {
            steps {
                sh 'go build ./...'
            }
        }
        stage('Test') {
            environment {
                INTEGRATION = currentBuild.getBuildCauses('jenkins.branch.BranchEventCause').size()
            }
            steps {
                sh 'echo "Integration build: ${INTEGRATION}"'
                sh 'go test -cover -timeout=1m ./...'
            }
        }
    }
}
