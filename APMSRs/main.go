package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const jsonURL string = "https://tron.f5net.com/sr/search/?owner=ENE&product=APM,Edge&format=json"
const loginURL string = "https://tron.f5net.com/login/"
const refererURL string = "https://tron.f5net.com/login/?next=/sr/search/?owner=ENE&product=APM,Edge&format=json"
const nextURL string = "/sr/search/?owner=ENE&amp;product=APM%2CEdge&amp;format=json"

// Example JSON - Showing 2 cases in the queue
// {
//     "data": [
//         {
//             "account": "Stanford Hospital",
//             "area": "",
//             "aspenns_closed": false,
//             "aspenns_name": "varina-1",
//             "aspenns_root": "1-3748581311",
//             "closed": null,
//             "cloud_platform": "",
//             "contact": "T.Cole@F5.com",
//             "contact_first": "Timothy",
//             "contact_last": "Cole",
//             "created": "2018-01-27T07:22:14-08:00",
//             "csp_created": null,
//             "csp_updated": null,
//             "entitlement_type": "Premium",
//             "has_children": false,
//             "has_first_response": false,
//             "has_orders": false,
//             "has_skd_session": false,
//             "hot": false,
//             "last_modified": "2018-01-27T07:24:50-08:00",
//             "last_modified_by": "TCOLE",
//             "last_resolved": null,
//             "location_country": "UNITED STATES",
//             "location_state": "CA",
//             "managed_escalation": false,
//             "originator": "TCOLE",
//             "owner": "ENE",
//             "parent": "1-3748581311",
//             "parent_name": "varina",
//             "phone": "(206) 272-6319",
//             "platform": "F5-VPR-LTM-C2400-AC, F100",
//             "platform_id": "F100",
//             "platform_part_number": "400-0028-10",
//             "premium_plus": false,
//             "premium_plus_sdm_only": false,
//             "problem_statement": "tmm cored: ** SIGILL ** fault addr: 0x57000aad144d",
//             "product": "BIG-IP APM",
//             "rowid": "1-1QBH2H7",
//             "security_flag": false,
//             "serial": "chs411447s",
//             "serial_parent_id": "ZWJOGBAQ",
//             "service_level": null,
//             "service_provider": false,
//             "severity": "3 - Performance Degraded",
//             "site": "227217",
//             "slm": "03:42",
//             "slm_due": "2018-01-27T11:22:14-08:00",
//             "slm_indicator": "SLM greater than 45 min",
//             "special_sla": false,
//             "sr_number": "1-3768175051",
//             "status": "Open",
//             "subarea": "",
//             "substatus": "New",
//             "version": "12.1.1"
//         },
//         {
//             "account": "Mondi AG",
//             "area": "",
//             "aspenns_closed": false,
//             "aspenns_name": "amorita-1",
//             "aspenns_root": "1-3768590917",
//             "closed": null,
//             "cloud_platform": "",
//             "contact": "P.Stefopoulos@F5.com",
//             "contact_first": "Periklis",
//             "contact_last": "Stefopoulos",
//             "created": "2018-01-27T05:35:25-08:00",
//             "csp_created": null,
//             "csp_updated": null,
//             "entitlement_type": "Premium",
//             "has_children": false,
//             "has_first_response": false,
//             "has_orders": false,
//             "has_skd_session": false,
//             "hot": false,
//             "last_modified": "2018-01-27T05:35:33-08:00",
//             "last_modified_by": "STEFOPOULOS",
//             "last_resolved": null,
//             "location_country": "AUSTRIA",
//             "location_state": "",
//             "managed_escalation": false,
//             "originator": "STEFOPOULOS",
//             "owner": "ENE",
//             "parent": "1-3768590917",
//             "parent_name": "amorita",
//             "phone": null,
//             "platform": "F5-BIG-BT-I7800, C118",
//             "platform_id": "C118",
//             "platform_part_number": "500-0003-03",
//             "premium_plus": false,
//             "premium_plus_sdm_only": false,
//             "problem_statement": "z101-13.1.0.1-Regression of ID534378 in 13.1.0 version",
//             "product": "BIG-IP APM",
//             "rowid": "1-1QBRP47",
//             "security_flag": false,
//             "serial": "f5-luqr-xbqr",
//             "serial_parent_id": "ZAQNDEXG",
//             "service_level": null,
//             "service_provider": false,
//             "severity": "4 - General Assistance",
//             "site": "381882",
//             "slm": "21:55",
//             "slm_due": "2018-01-28T05:35:25-08:00",
//             "slm_indicator": "SLM greater than 45 min",
//             "special_sla": false,
//             "sr_number": "1-3768670951",
//             "status": "Open",
//             "subarea": "",
//             "substatus": "New",
//             "version": "13.1.0"
//         }
//     ],
//     "metadata": {
//         "filters": {
//             "F5 Product": "~LIKE *APM* OR ~LIKE *EDGE*",
//             "Owner": "=ENE",
//             "Status": "IS NULL OR <> Closed"
//         },
//         "retrieved": "2018-01-27T15:39:42.002478Z",
//         "truncated": false,
//         "version": [
//             1,
//             0
//         ]
//     }
// }

type jData struct {
	SR []dataEntries `json:"data"`
}

type dataEntries struct {
	Account      string `json:"account"`
	ContactFirst string `json:"contact_first"`
	ContactLast  string `json:"contact_last"`
	Hot          bool   `json:"hot"`
	Location     string `json:"location"`
	Title        string `json:"problem_statement"`
	Severity     string `json:"severity"`
	SRNumber     string `json:"sr_number"`
}

func main() {
	username := os.Args[1:2][0]
	password := os.Args[2:][0]
	seconds, err := strconv.Atoi(os.Args[3:][0])
	if err != nil {
		log.Fatal(err)
	}

	for {

		// request the url and get back a response. Go will follow up to 10 redirects
		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		}
		client := &http.Client{Transport: transCfg}
		resp, err := client.Get(jsonURL)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		// Check the response body for a form action of /login/ if we got a 200 back
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		if strings.Contains(string(body), "/login/") {
			// POST to /login/ make sure to get the csrfmiddleware token and send that as a cookie

			// Setup the post params
			// params are username, password, csrfmiddlewaretoken, next
			re := regexp.MustCompile(`csrfmiddlewaretoken.*value=\'([a-zA-Z0-9].*)\'`)
			csrfmiddlewaretoken := re.FindStringSubmatch(string(body))[1]

			// if we see the logon page then we need to authenticate
			jar, err := cookiejar.New(nil)
			if err != nil {
				log.Fatal(err)
			}
			httpClient := http.Client{Jar: jar}

			data := url.Values{}
			data.Set("username", string(username))
			data.Add("password", password)
			data.Add("csrfmiddlewaretoken", csrfmiddlewaretoken)
			data.Add("next", nextURL)
			// res, err := httpClient.PostForm(loginURL, data)
			req, err := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
			if err != nil {
				log.Fatal(err)
			}
			req.Header.Set("Referer", refererURL)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(resp.Cookies()[0])
			res, err := httpClient.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			defer res.Body.Close()

			bod, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err)
			}

			// fmt.Println(string(bod))

			// parse out the json using the data struct
			var jsonData jData
			json.Unmarshal(bod, &jsonData)
			// fmt.Println(jsonData)
			// fmt.Println(jsonData.SR[0].Account)

			//Display the unmarshalled data
			for _, sr := range jsonData.SR {
				fmt.Println("*****", sr.SRNumber, "*****")
				fmt.Println("Account Name:", sr.Account)
				fmt.Println("Problem:", sr.Title)
				fmt.Println("Contact Name:", sr.ContactFirst, sr.ContactLast)
				fmt.Println("Severity:", sr.Severity)
				fmt.Println("Location:", sr.Location)
				fmt.Println("Is it Hot:", sr.Hot)
				fmt.Println("")
			}

		}

		// wait 10 seconds before checking again
		time.Sleep(time.Second * time.Duration(seconds))
		fmt.Println("resuming...")
		//fmt.Println(resp.StatusCode)
		//fmt.Println(resp.Cookies()[0])
	}
}
