package config

import (
	"bytes"

	log "github.com/cihub/seelog"
	"github.com/nijohando/naruko/resource"
	"github.com/spf13/viper"
)

const (
	envAccessKeyID     = "ACCESS_KEY_ID"
	envSecretAccessKey = "SECRET_ACCESS_KEY"
	envMonostickDevice = "MONOSTICK_DEVICE"
	envAWSIoTEndpoint  = "AWS_IOT_ENDPOINT"
	// AccessKeyID is a configuration key for AWS access key id
	AccessKeyID = "access-key-id"
	// SecretAccessKey is a configuration key for AWS secret access key
	SecretAccessKey = "secret-access-key"
	// AWSIoTEndpoint is a configuration key for AWS Iot endpoint URL
	AWSIoTEndpoint = "aws.iot.endpoint"
	// AWSIoTRegion is a configuration key for AWS Iot region
	AWSIoTRegion = "aws.iot.region"
	// AWSIoTServiceName is a configuration key for AWS IoT service name
	AWSIoTServiceName = "aws.iot.service-name"
	// AWSIoTMQTTSignedRequestExpires is a configuration key for expiration of AWS signed request
	AWSIoTMQTTSignedRequestExpires = "aws.iot.mqtt.signed-request-expires"
	// AWSIoTMQTTClientID is a configuration key for MQTT client id
	AWSIoTMQTTClientID = "aws.iot.mqtt.client-id"
	// MonostickDevice is a configuration key for monostick device file name
	MonostickDevice = "monostick.device"
	// MonostickBaud is a configuration key for serial baud rates
	MonostickBaud = "monostick.baud"
	// MonostickReadTimeout is a configuration key for timeout when reading MONOSTICK
	MonostickReadTimeout = "monostick.read-timeout"
)

func validateConfig() bool {
	ok := true
	requireConfigError := func(key string) {
		log.Errorf("%s is required", key)
		ok = false
	}
	requireConfig := func(key string) {
		if viper.Get(key) == nil {
			requireConfigError(key)
		}
	}
	requireConfig(AccessKeyID)
	requireConfig(SecretAccessKey)
	requireConfig(AWSIoTEndpoint)
	requireConfig(AWSIoTRegion)
	requireConfig(AWSIoTMQTTClientID)
	requireConfig(MonostickDevice)
	requireConfig(MonostickBaud)
	requireConfig(MonostickReadTimeout)
	return ok
}

func init() {
	logConfig, err := resource.Asset("resource/seelog.xml")
	if err != nil {
		panic(err)
	}
	logger, err := log.LoggerFromConfigAsBytes(logConfig)
	if err != nil {
		panic(err)
	}
	log.ReplaceLogger(logger)
	defer log.Flush()

	viper.BindEnv(AccessKeyID, envAccessKeyID)
	viper.BindEnv(SecretAccessKey, envSecretAccessKey)
	viper.BindEnv(AWSIoTEndpoint, envAWSIoTEndpoint)
	viper.BindEnv(MonostickDevice, envMonostickDevice)

	appConfig, err := resource.Asset("resource/config.yml")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	viper.ReadConfig(bytes.NewBuffer(appConfig))
	if !validateConfig() {
		panic("Failed to configure")
	}
}
