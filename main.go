package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/sendgrid/platformlib/pflog"
)

var (
	termBold         []byte
	termUnderlineOn  []byte
	termUnderlineOff []byte
	termReset        []byte
)

func main() {
	log := pflog.New("wa", "STDOUT", true)

	var err error
	termBold, err = exec.Command("tput", "bold").CombinedOutput()
	log.AssertNoErr(err)
	termReset, err = exec.Command("tput", "sgr0").CombinedOutput()
	log.AssertNoErr(err)
	termUnderlineOn, err = exec.Command("tput", "smul").CombinedOutput()
	log.AssertNoErr(err)
	termUnderlineOff, err = exec.Command("tput", "rmul").CombinedOutput()
	log.AssertNoErr(err)

	appID := os.Getenv("WOLFRAM_ALPHA_APPID")
	if appID == "" {
		log.Debugf("Env var WOLFRAM_ALPHA_APPID is required")
		os.Exit(1)
	}

	query := strings.Join(os.Args[1:], " ")
	err = sendRequest(appID, query)
	log.AssertNoErr(err)
}

func sendRequest(appID, query string) error {

	// Create client
	client := &http.Client{}

	// Create request
	u := url.Values{
		"appid": {appID},
		"input": {query},
	}
	req, err := http.NewRequest("GET", "https://api.wolframalpha.com/v2/query?"+u.Encode(), nil)
	if err != nil {
		return err
	}

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Read Response Body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("invalid status code: %d\nbody: %s", resp.StatusCode, string(respBody))
	}

	// log.Print(string(respBody))

	return formatResponse(respBody, os.Stdout)
}
 
ty  pe queryResult struct {
	Pods []struct {  
		Title  strin g `xml:"title,attr"`
		SubPod []st ruct {
			PlainT ext string `xml:"plaintext"`
		} `xml:"s ubpod"`
	} `xml:"pod" `
}

func formatResponse(in []byte, w io.Writer) error {
	var r queryResult
	err := xml.Unmarshal(in, &r)
	if err != nil {
		return err
	}

	for _, p := range r.Pods {
		fmt.Fprintf(w, "%s%s%s\n", string(termUnderlineOn), p.Title, string(termUnderlineOff))
		for _, sp := range p.SubPod {
			fmt.Fprintf(w, "%s\n\n", sp.PlainText)
		}
	}
	return nil
}
