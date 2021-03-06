//Package brake contains tools for setting up and working with the [airbrake](https://airbrake.io/) error monitoring software.
package brake

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/airbrake/gobrake"
	"github.com/seeker-insurance/kit/functools"
	"github.com/seeker-insurance/kit/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	traceDepth = 5

	SeverityError    severity = "error"
	SeverityWarn     severity = "warning"
	SeverityCritical severity = "critical"
)

type severity string

var (
	Airbrake *gobrake.Notifier
	Env      string
)

func init() {
	cobra.OnInitialize(setup)
}

func setup() {
	key := viper.GetString("airbrake_key")
	project := viper.GetInt64("airbrake_project")
	Env = viper.GetString("airbrake_env")

	if len(key) != 0 && project != 0 {
		Airbrake = gobrake.NewNotifier(project, key)
	}
}

func IsSetup() bool {
	return Airbrake != nil
}

func Notify(e error, req *http.Request, sev severity) {
	if IsSetup() {
		notice := gobrake.NewNotice(e, req, traceDepth)
		setNoticeVars(notice, req, sev)
		Airbrake.SendNotice(notice)
		return
	}
	log.Warnf("Error (not reported): %v", e)
}

func NotifyFromChan(errs chan error, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		for e := range errs {
			Notify(e, nil, SeverityError)
		}
		wg.Done()
	}()
}

func setNoticeVars(n *gobrake.Notice, req *http.Request, sev severity) {
	n.Context["environment"] = Env
	n.Context["severity"] = sev
	if expectBody(req) {
		n.Params["body"] = body(req)
	}
}

func body(req *http.Request) interface{} {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return fmt.Sprintf("error reading body: %v", err)
	}

	if len(b) == 0 {
		return "no body"
	}

	if !isJSONContentType(req) {
		return string(b)
	}

	// if !json.Valid(b) {
	// 	return fmt.Sprintf("body is not valid JSON: %s", b)
	// }

	formatted := make(map[string]interface{})
	json.Unmarshal(b, &formatted)

	return formatted
}

func isJSONContentType(req *http.Request) bool {
	cType := req.Header.Get("Content-Type")
	cType = strings.ToLower(cType)
	return strings.Contains(cType, "json")
}

func expectBody(req *http.Request) bool {
	if req == nil {
		return false
	}
	requestsWithBody := []string{http.MethodPost, http.MethodPatch, http.MethodPut}
	return functools.StringSliceContains(requestsWithBody, req.Method)
}
