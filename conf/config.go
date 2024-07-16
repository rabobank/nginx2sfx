package conf

import (
	"encoding/json"
	"fmt"
	"github.com/rabobank/nginx2sfx/model"
	"os"
	"strconv"
)

var (
	VERSION string
	COMMIT  string

	DebugStr             = os.Getenv("NGINX2SFX_DEBUG")
	Debug                bool
	InputFile            = os.Getenv("NGINX2SFX_INPUTFILE")
	SfxUrl               = os.Getenv("NGINX2SFX_URL")
	SkipSslValidationStr = os.Getenv("NGINX2SFX_SKIP_SSL_VALIDATION")
	SkipSslValidation    bool
	SfxToken                   = os.Getenv("NGINX2SFX_TOKEN")
	BatchSizeStr               = os.Getenv("NGINX2SFX_BATCH_SIZE")
	BatchSize                  = 100
	BatchIntervalStr           = os.Getenv("NGINX2SFX_BATCH_INTERVAL")
	BatchInterval        int64 = 5
	UriAsDimensionStr          = os.Getenv("NGINX2SFX_URI_AS_DIMENSION")
	UriAsDimension       bool
	VcapApp              model.VcapApplication
	CfEnv                string
	CfInstanceIndex      string
)

const SfxTimeout = 3

func EnvironmentComplete() {
	envComplete := true
	Debug = false
	if DebugStr != "false" && DebugStr != "" {
		Debug = true
	}
	if InputFile == "" {
		InputFile = "logs/nginx2sfx.log"
	}
	if SfxUrl == "" {
		envComplete = false
		fmt.Println("missing envvar: NGINX2SFX_URL")
	}
	if BatchSizeStr != "" {
		var err error
		BatchSize, err = strconv.Atoi(BatchSizeStr)
		if err != nil {
			fmt.Printf("failed reading envvar NGINX2SFX_BATCHSIZE, err: %s\n", err)
			envComplete = false
		}
	}
	if BatchIntervalStr != "" {
		var err error
		BatchInterval, err = strconv.ParseInt(BatchIntervalStr, 0, 0)
		if err != nil {
			fmt.Printf("failed reading envvar NGINX2SFX_BATCHINTERVAL, err: %s\n", err)
			envComplete = false
		}
	}
	SkipSslValidation = false
	if SkipSslValidationStr == "true" {
		SkipSslValidation = true
	}

	UriAsDimension = false
	if UriAsDimensionStr == "true" {
		UriAsDimension = true
	}

	// get optional envvars, that could be used as dimensions in the SignalFx metrics
	vcapAppStr := os.Getenv("VCAP_APPLICATION")
	if vcapAppStr != "" {
		err := json.Unmarshal([]byte(vcapAppStr), &VcapApp)
		if err != nil && Debug {
			fmt.Printf("failed to json decode envvar VCAP_APPLICATION, error: %s", err)
		}
		CfEnv = os.Getenv("RABOPCF_SYSTEM_ENV")
		CfInstanceIndex = os.Getenv("CF_INSTANCE_INDEX")
	}

	// try to get the SfxToken from credhub
	vcapServicesString := os.Getenv("VCAP_SERVICES")
	if vcapServicesString != "" {
		vcapServices := model.VcapServices{}
		if err := json.Unmarshal([]byte(vcapServicesString), &vcapServices); err != nil {
			fmt.Printf("could not get SfxToken from credhub, error: %s\n", err)
		} else {
			for _, service := range vcapServices.Credhub {
				if service.InstanceName == "sfxtoken" {
					SfxToken = service.Credentials.Token
					if Debug {
						fmt.Println("got SfxToken from credhub")
					}
				}
			}
		}
	} else {
		if SfxToken == "" {
			envComplete = false
			fmt.Println("missing envvar: NGINX2SFX_TOKEN, and also no \"sfxtoken \" credhub service instance bound")
		}
	}

	if Debug {
		fmt.Printf("using environment variables:\nNGINX2SFX_INPUTFILE:%s\nNGINX2SFX_URL:%s\nNGINX2SFX_TOKEN:<redacted>\nNGINX2SFX_BATCH_SIZE:%d\nNGINX2SFX_BATCH_INTERVAL:%d\nNGINX2SFX_URI_AS_DIMENSION:%t\n", InputFile, SfxUrl, BatchSize, BatchInterval, UriAsDimension)
	}

	if !envComplete {
		fmt.Println("one or more required envvars missing, aborting...")
		os.Exit(8)
	}
}
