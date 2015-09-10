// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"unicode"
)

const (
	defaultBaseConfigURL   = "http://127.0.0.1/api/v1/hook/"
	defaultSuffixList      = "_DATA,_CONFIG"
	suffixListOverride     = "BP_HOOK_DATA_SUFFIX_LIST"
	bluePlanetPrefix       = "BP_"
	bluePlanetDataSuffix   = "_DATA"
	bluePlanetConfigSuffix = "_CONFIG"
	bluePlanetOverride     = "BP_HOOK_URL_REDIRECT_"
	postBodyType           = "application/json"
)

var displayOnly = flag.Bool("n", false, "display REST call, but don't actually make it")
var verbose = flag.Bool("v", false, "output verbose logging information")

// hasDataSuffix - returns true if the given string (s) ends in one of the strings specified
// as a suffix, else returns false
func hasDataSuffix(s string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

// ignore - returns true if the environment variable is essentially in a black list of
// variables used for configuration and not to be passed to the REST interface
func ignore(s string) bool {
	switch s {
	case suffixListOverride:
		return true
	}
	return false
}

func main() {

	flag.Parse()

	if *verbose {
		log.Printf("verbose logging is 'true'\n")
		log.Printf("display only actions is '%v'\n", *displayOnly)
		log.Printf("default data suffix list is '%s'\n", defaultSuffixList)
	}

	base := path.Base(os.Args[0])
	name := strings.Replace(strings.ToUpper(base), "-", "_", -1)
	targetURL := defaultBaseConfigURL + base
	overrideVar := bluePlanetOverride + name

	// Check to see if the data suffix list is overrider by an environment variable and then convert the given list,
	// or the default to a string array for checking against later down the code
	suffixList := os.Getenv(suffixListOverride)
	if suffixList == "" {
		suffixList = defaultSuffixList
	} else if *verbose {
		log.Printf("overriding data suffix list to '%s'\n", suffixList)
	}
	// Support both space and comma separated lists
	suffixes := strings.FieldsFunc(suffixList,
		func(r rune) bool { return unicode.IsSpace(r) || r == ',' })

	if *verbose {
		log.Printf("default target URL set to '%s'\n", targetURL)
		log.Printf("environment variable to override target URL set to '%s'\n", overrideVar)
	}

	env := os.Environ()
	bpData := make(map[string]interface{})
	for _, s := range env {
		// Rather annoying that we get the environment as an array of "=" separated strings given that I really want
		// to turn this into data, so split the string on "=" so we have a key and a value
		nv := strings.SplitN(s, "=", 2)
		if strings.HasPrefix(nv[0], bluePlanetPrefix) {
			// Before we add this into the data to push to the hook, lets check to see if the environment
			// overrides the target URL. The var to override would be BP_HOOK_URL_REDIRECT_<name>. The name
			// will be the name of the BP hook such as southbound-update, which is also the name of the executable.
			if *verbose {
				log.Printf("processing environment variable '%s'\n", nv[0])
			}
			if !ignore(nv[0]) {
				if nv[0] == overrideVar {
					targetURL = nv[1]
					if *verbose {
						log.Printf("overriding target URL from environment to '%s'\n", targetURL)
					}
				} else if hasDataSuffix(nv[0], suffixes) {
					// If we have a data var then this "should be" string version of a JSON object or array, so lets
					// convert it to a proper JSON structure.
					if *verbose {
						log.Printf("processing value of '%s' as JSON data\n", nv[0])
					}
					var obj interface{}
					if err := json.Unmarshal([]byte(nv[1]), &obj); err != nil {
						// If we can't parse it, then log the warning and set it as a string value
						log.Printf("WARN: unable to unmarshal data value '%s' as JSON, passing vaue as string instead: %s\n", nv[1], err)
						bpData[nv[0]] = nv[1]
					} else {
						bpData[nv[0]] = obj
					}
				} else {
					bpData[nv[0]] = nv[1]
				}
			}
		} else if *verbose {
			log.Printf("ignoring environment variable '%s' because it does not start with prefix '%s'\n",
				nv[0], bluePlanetPrefix)
		}
	}

	if *verbose {
		log.Printf("final target URL is '%s'\n", targetURL)
	}

	if b, err := json.Marshal(bpData); err != nil {
		log.Fatalf("ERROR: unable to marshal data to JSON: %s", err)
	} else {
		if *displayOnly || *verbose {
			log.Printf("POST: %s '%s'\n", targetURL, string(b))
		}

		if !*displayOnly {
			resp, err := http.Post(targetURL, postBodyType, strings.NewReader(string(b)))
			if err != nil {
				log.Fatalf("ERROR: unable to POST to URL '%s': %s\n", targetURL, err)
			} else {
				// Check the response code to see if it is a non-200 value
				if (int)(resp.StatusCode/100) != 2 {
					log.Fatalf("ERROR: Unsuccessful response code for POST to URL '%s', received '%s'\n", targetURL, resp.Status)
				}
			}
		}
	}
}
