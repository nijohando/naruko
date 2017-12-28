package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	log "github.com/cihub/seelog"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nijohando/naruko/config"
	"github.com/nijohando/naruko/twelite"
	"github.com/spf13/viper"
)

// Shadow represents json data structure for root node
type Shadow struct {
	State ShadowState `json:"state"`
}

// ShadowState represents json data structure for state node
type ShadowState struct {
	Reported ShadowReport `json:"reported"`
}

// ShadowReport represents json data structure for reported node
type ShadowReport struct {
	Timestamp          string
	Lqi                uint8
	PowerSupplyVoltage uint16
	SensorMode         uint16
	X                  int16
	Y                  int16
	Z                  int16
}

const (
	dateFormat = "2006/1/2 15:04:05"
)

var client mqtt.Client

func createRequest() (*http.Request, error) {
	url := viper.GetString(config.AWSIoTEndpoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	accessKeyID := viper.GetString(config.AccessKeyID)
	secretAccessKey := viper.GetString(config.SecretAccessKey)
	cred := credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
	signer := v4.Signer{
		Credentials: cred,
	}
	awsIoTRegion := viper.GetString(config.AWSIoTRegion)
	awsIoTServiceName := viper.GetString(config.AWSIoTServiceName)
	awsIoTMQTTSignedRequestExpires := viper.GetDuration(config.AWSIoTMQTTSignedRequestExpires)
	signer.Presign(req, nil, awsIoTServiceName, awsIoTRegion, awsIoTMQTTSignedRequestExpires*time.Second, time.Now())
	return req, nil
}

func newMQTTClient() (mqtt.Client, error) {
	req, err := createRequest()
	if err != nil {
		return nil, err
	}
	awsIoTMQTTClientID := viper.GetString(config.AWSIoTMQTTClientID)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(req.URL.String())
	opts.SetClientID(awsIoTMQTTClientID)
	opts.SetAutoReconnect(false)
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Warnf("Connection is lost. (Reason: %q)\n", err)
		log.Info("Reconnecting...")
		err = initClient()
		if err != nil {
			panic(err)
		}
	})
	return mqtt.NewClient(opts), nil
}

func initClient() error {
	log.Info("Initialize client.")
	var err error
	client, err = newMQTTClient()
	if err != nil {
		return err
	}
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func main() {
	//mqtt.DEBUG = log.New(os.Stdout, "", 0)
	//mqtt.WARN = log.New(os.Stdout, "", 0)
	//mqtt.ERROR = log.New(os.Stdout, "", 0)
	err := initClient()
	if err != nil {
		log.Errorf("Failed to initialize client. %q", err)
		panic(err)
	}
	log.Info("Create TWELITE session.")
	session, err := twelite.NewSession()
	if err != nil {
		log.Errorf("Failed to create TWELITE session. %q", err)
		panic(err)
	}
	log.Info("Waiting to receive sensor data from MONOSTICK.")
	for {
		data := session.Read()
		switch v := data.(type) {
		case *twelite.Acceleration:
			log.Infof("Got data. Timestamp:%q, Lqi:%d, ChildID:%q, PowerSupplyVoltage:%d, SensorMode:%d, X:%d, Y:%d, Z:%d",
				v.Timestamp.Format(dateFormat), v.Lqi, v.ChildID, v.PowerSupplyVoltage, v.SensorMode, v.X, v.Y, v.Z)
			shadow := Shadow{
				State: ShadowState{
					Reported: ShadowReport{
						Timestamp:          v.Timestamp.Format(dateFormat),
						Lqi:                v.Lqi,
						PowerSupplyVoltage: v.PowerSupplyVoltage,
						SensorMode:         v.SensorMode,
						X:                  v.X,
						Y:                  v.Y,
						Z:                  v.Z,
					},
				},
			}
			b, err := json.Marshal(shadow)
			if err != nil {
				panic(err)
			}
			token := client.Publish("$aws/things/door1/shadow/update", 0, false, string(b))
			if token.Wait() && token.Error() != nil {
				log.Errorf("Failed to publish. %q", token.Error())
			}
		case *twelite.Error:
			log.Errorf("Got an error. %q", data)
		}
	}
}
