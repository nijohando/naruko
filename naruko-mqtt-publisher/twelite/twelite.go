package twelite

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/cihub/seelog"

	"github.com/nijohando/naruko/naruko-mqtt-publisher/config"
	"github.com/spf13/viper"
	"github.com/tarm/serial"
)

type (
	// Session represents the interface that operates a connection
	Session interface {
		Read() interface{}
		Close()
	}
	// session represents a new connection with the device
	session struct {
		port    *serial.Port
		buff    *bytes.Buffer
		channel chan interface{}
		closed  bool
	}

	// Acceleration represents sensor data from TWELITE 2525A
	Acceleration struct {
		Timestamp          time.Time
		Lqi                uint8  // Value of LQI
		ChildID            string // The last 8 digits of child device's MAC address
		PowerSupplyVoltage uint16 // Power supply voltage[mV]
		SensorMode         uint16 // Sensor mode
		X                  int16  // Acceleration X
		Y                  int16  // Acceleration Y
		Z                  int16  // Acceleration Z
	}
	// Error represents a parse failure of sensor data from TWELITE 2525A
	Error struct {
		Msg string
	}
)

const (
	packetDataSeparator = ";"
)

// NewSession establishes a new session with the device
func NewSession() (Session, error) {
	device := viper.GetString(config.MonostickDevice)
	baud := viper.GetInt(config.MonostickBaud)
	readTimeout := viper.GetDuration(config.MonostickReadTimeout)
	c := &serial.Config{Name: device, Baud: baud, ReadTimeout: readTimeout * time.Second}
	p, err := serial.OpenPort(c)
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer(make([]byte, 0, 256))
	ch := make(chan interface{})
	s := &session{port: p, buff: b, channel: ch, closed: false}
	go listen(s)
	return s, nil
}

func (s *session) Read() interface{} {
	return <-s.channel
}

// Close session
func (s *session) Close() {
	s.closed = true
	err := s.port.Close()
	if err != nil {
		log.Errorf("Failed to close session. %q", err)
	}
}

func listen(s *session) {
	defer close(s.channel)
	work := make([]byte, 256)
	eol := []byte("\r\n")
	for {
		if s.closed {
			break
		}
		n, _ := s.port.Read(work)
		if n == 0 {
			continue
		}
		_, err := s.buff.Write(work[:n])
		if err != nil {
			log.Errorf("Failed to write buffer. %q", err)
		}
		b := s.buff.Bytes()
		if bytes.Contains(b, eol) {
			packets := bytes.Split(b, eol)
			len := len(packets)
			for _, p := range packets[:len-1] {
				packet := string(p[:])
				acceleration, err := parsePacket(packet)
				if err != nil {
					s.channel <- &Error{Msg: err.Error()}
				} else if acceleration != nil {
					s.channel <- acceleration
				}
			}
			s.buff.Reset()
			s.buff.Write(packets[len-1])
		}
	}
}

func parsePacket(packet string) (*Acceleration, error) {
	if !strings.HasPrefix(packet, packetDataSeparator) || !strings.HasSuffix(packet, packetDataSeparator) {
		return nil, fmt.Errorf("Invalid packet prefix. %s", packet)
	}
	s := strings.Split(packet[1:len(packet)-1], ";")
	l := len(s)
	if l == 1 {
		return nil, nil
	}
	if l != 14 {
		return nil, fmt.Errorf("Invalid packet data length. %d, %q", len(s), s)
	}
	wk := s[2]
	lqi, err := strconv.ParseUint(wk, 10, 8)
	if err != nil {
		return nil, fmt.Errorf("Invalid lqi. %s", wk)
	}
	childID := s[4]
	wk = s[5]
	powerSupplyVoltage, err := strconv.ParseUint(wk, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("Invalid powerSupplyVoltage. %s", wk)
	}
	wk = s[6]
	sensorMode, err := strconv.ParseUint(wk, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("Invalid sensorMode. %s", wk)
	}
	wk = s[11]
	x, err := strconv.ParseInt(wk, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("Invalid x . %s", wk)
	}
	wk = s[12]
	y, err := strconv.ParseInt(wk, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("Invalid y . %s", wk)
	}
	wk = s[13]
	z, err := strconv.ParseInt(wk, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("Invalid z . %s", wk)
	}
	a := &Acceleration{
		Timestamp:          time.Now().UTC(),
		Lqi:                uint8(lqi),
		ChildID:            childID,
		PowerSupplyVoltage: uint16(powerSupplyVoltage),
		SensorMode:         uint16(sensorMode),
		X:                  int16(x),
		Y:                  int16(y),
		Z:                  int16(z),
	}
	return a, nil
}
