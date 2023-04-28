// Copyright 2022 Rubrik, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/appliance"
	polaris_log "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
	}

	applianceID, err := uuid.Parse(os.Args[1])
	if err != nil {
		printHelp()
	}

	logger := polaris_log.NewStandardLogger()
	logger.SetLogLevel(polaris_log.Error)
	if err := polaris.SetLogLevelFromEnv(logger); err != nil {
		log.Fatal(err)
	}

	serviceAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}

	token, err := appliance.TokenFromServiceAccount(serviceAccount, applianceID, logger)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(token)
}

func printHelp() {
	fmt.Printf("%s <appliance-uuid>\n", filepath.Base(os.Args[0]))
	os.Exit(1)
}
