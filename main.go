package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hpcloud/tail"
	"github.com/rabobank/nginx2sfx/conf"
	"github.com/rabobank/nginx2sfx/model"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type mapKey struct {
	statusCode     string
	uri            string
	method         string
	serverProtocol string
	metricName     string
}

var (
	lastProcessed int64
	metricBuffer  = map[mapKey]float64{}
	tlsConfig     = tls.Config{InsecureSkipVerify: conf.SkipSslValidation}
	client        = http.Client{Transport: &http.Transport{TLSClientConfig: &tlsConfig}, Timeout: time.Duration(conf.SfxTimeout) * time.Second}
)

func main() {
	fmt.Printf("nginx2sfx, CommitHash %s, Version %s\n", conf.COMMIT, conf.VERSION)
	// check envvars (and exit if incomplete)
	conf.EnvironmentComplete()

	// check for existence of the file:
	for {
		if _, err := os.Stat(conf.InputFile); os.IsNotExist(err) {
			fmt.Printf("waiting for file %s to appear...\n", conf.InputFile)
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	if fileChannel, err := tail.TailFile(conf.InputFile, tail.Config{Follow: true, Logger: tail.DiscardingLogger}); err != nil {
		panic(err)
	} else {
		lineCount := 0      // used to keep track when to send a batch of metrics to Sfx
		totalLineCount := 0 // used to keep track when to truncate the file
		for line := range fileChannel.Lines {
			lineCount++
			totalLineCount++
			updateMetrics(line.Text)
			if lineCount >= conf.BatchSize || time.Now().Unix()-lastProcessed >= conf.BatchInterval {
				send2sfx()
				lineCount = 0
				lastProcessed = time.Now().Unix()
				metricBuffer = make(map[mapKey]float64) // clear the buffer, so all metrics start with 0 again
				// truncate the file if it gets too big
				if totalLineCount > 5000 {
					totalLineCount = 0
					if err = os.Truncate(conf.InputFile, 0); err != nil {
						log.Printf("Failed to truncate file %s: %s", conf.InputFile, err)
					}
				}
			}
		}
	}
}

func updateMetrics(logLine string) {
	nginxLog := model.LogLine{}
	if err := json.Unmarshal([]byte(logLine), &nginxLog); err != nil {
		fmt.Printf("failed to parse logline: %s : %s\n", logLine, err)
	} else {
		if !conf.UriAsDimension {
			nginxLog.Uri = "" // if we don't want to use the uri as a dimension, we set it to an empty string
		}
		// The statuscode, uri, method and serverProtocol can vary, so they are in the key of the map. When sending the metrics to Sfx we also add all the "static" dimensions like app_name, space_name...
		if element, found := metricBuffer[mapKey{nginxLog.Status, nginxLog.Uri, nginxLog.Method, nginxLog.ServerProtocol, "nginx_http_requests_count"}]; found {
			metricBuffer[mapKey{nginxLog.Status, nginxLog.Uri, nginxLog.Method, nginxLog.ServerProtocol, "nginx_http_requests_count"}] = element + 1
		} else {
			metricBuffer[mapKey{nginxLog.Status, nginxLog.Uri, nginxLog.Method, nginxLog.ServerProtocol, "nginx_http_requests_count"}] = 1
		}
		if element, found := metricBuffer[mapKey{nginxLog.Status, nginxLog.Uri, nginxLog.Method, nginxLog.ServerProtocol, "nginx_http_requests_totalTime"}]; found {
			metricBuffer[mapKey{nginxLog.Status, nginxLog.Uri, nginxLog.Method, nginxLog.ServerProtocol, "nginx_http_requests_totalTime"}] = element + nginxLog.RequestTime
		} else {
			metricBuffer[mapKey{nginxLog.Status, nginxLog.Uri, nginxLog.Method, nginxLog.ServerProtocol, "nginx_http_requests_totalTime"}] = nginxLog.RequestTime
		}
		if element, found := metricBuffer[mapKey{nginxLog.Status, nginxLog.Uri, nginxLog.Method, nginxLog.ServerProtocol, "nginx_http_requests_totalBytes"}]; found {
			metricBuffer[mapKey{nginxLog.Status, nginxLog.Uri, nginxLog.Method, nginxLog.ServerProtocol, "nginx_http_requests_totalBytes"}] = element + nginxLog.BodyBytesSent
		} else {
			metricBuffer[mapKey{nginxLog.Status, nginxLog.Uri, nginxLog.Method, nginxLog.ServerProtocol, "nginx_http_requests_totalBytes"}] = nginxLog.BodyBytesSent
		}
	}
}

func send2sfx() {
	counterMetrics := model.CounterMetrics{}
	var dimensions model.Dimensions
	for ix, metric := range metricBuffer {
		if conf.UriAsDimension {
			dimensions = model.Dimensions{Uri: ix.uri, Method: ix.method, ServerProtocol: ix.serverProtocol, StatusCode: ix.statusCode, Cfenv: conf.CfEnv, CfInstanceIndex: conf.CfInstanceIndex, CfAppName: conf.VcapApp.Name, CfAppId: conf.VcapApp.ApplicationID, CfSpaceName: conf.VcapApp.SpaceName, CfOrgName: conf.VcapApp.OrganizationName}
		} else {
			dimensions = model.Dimensions{Method: ix.method, ServerProtocol: ix.serverProtocol, StatusCode: ix.statusCode, Cfenv: conf.CfEnv, CfInstanceIndex: conf.CfInstanceIndex, CfAppName: conf.VcapApp.Name, CfAppId: conf.VcapApp.ApplicationID, CfSpaceName: conf.VcapApp.SpaceName, CfOrgName: conf.VcapApp.OrganizationName}
		}

		sfxMetric := model.CounterMetric{
			Metric:     ix.metricName,
			Value:      metric,
			Dimensions: dimensions,
		}
		counterMetrics.Counter = append(counterMetrics.Counter, sfxMetric)
	}

	postBodyBytes, err := json.Marshal(counterMetrics)
	if err != nil {
		fmt.Println(err)
		return
	}

	if conf.Debug {
		fmt.Printf("sending to SignalFx: %s\n", string(postBodyBytes))
	}

	req, _ := http.NewRequest("POST", conf.SfxUrl, bytes.NewBuffer(postBodyBytes))
	req.Header.Add("X-SF-Token", conf.SfxToken)
	req.Header.Add("Content-Type", "application/json")
	var resp *http.Response
	resp, err = client.Do(req)
	if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
		// do nothing, all went well
	} else {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			errString := fmt.Sprintf("failed to send log, response from POST to %s: %s\n%s", conf.SfxUrl, resp.Status, body)
			fmt.Println(errString)
			err = errors.New(errString)
		} else {
			errString := fmt.Sprintf("failed to send log, response from POST to %s: %v", conf.SfxUrl, err)
			fmt.Println(errString)
			err = errors.New(errString)
		}
	}
	lastProcessed = time.Now().Unix()
}
