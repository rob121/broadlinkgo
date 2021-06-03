package broadlinkgo

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

const learnTimeout = 30 // seconds
const sendRetries = 3

// ResponseType denotes the type of payload.
type ResponseType int

// Enumerations of PayloadType.
const (
	Unknown ResponseType = iota
	AuthOK
	DeviceError
	Temperature
	CommandOK
	RawData
	RawRFData
	RawRFData2
)

// Response represents a decrypted payload from the device.
type Response struct {
	Type ResponseType
	Data []byte
}

type device struct {
	conn              *net.PacketConn
	remoteAddr        string
	timeout           int
	deviceType        int
	mac               net.HardwareAddr
	count             int
	key               []byte
	iv                []byte
	id                []byte
	requestHeader     []byte
	codeSendingHeader []byte
}

type unencryptedRequest struct {
	command byte
	payload []byte
}

func newDevice(remoteAddr string, mac net.HardwareAddr, timeout int, devChar deviceCharacteristics) (*device, error) {
	rand.Seed(time.Now().Unix())
	d := &device{
		remoteAddr:        remoteAddr,
		timeout:           timeout,
		deviceType:        devChar.deviceType,
		mac:               mac,
		count:             rand.Intn(0xffff),
		key:               []byte{0x09, 0x76, 0x28, 0x34, 0x3f, 0xe9, 0x9e, 0x23, 0x76, 0x5c, 0x15, 0x13, 0xac, 0xcf, 0x8b, 0x02},
		iv:                []byte{0x56, 0x2e, 0x17, 0x99, 0x6d, 0x09, 0x3d, 0x28, 0xdd, 0xb3, 0xba, 0x69, 0x5a, 0x2e, 0x6f, 0x58},
		id:                []byte{0, 0, 0, 0},
		requestHeader:     devChar.requestHeader,
		codeSendingHeader: devChar.codeSendingHeader,
	}

	//Logger.Printf("%#v",d)

	resp, err := d.serverRequest(authenticatePayload())
	d.close()
	if err != nil {
		Logger.Printf("%#v",err)
		return d, fmt.Errorf("error making authentication request: %v", err)
	}
	if resp.Type == DeviceError {
		return d, errors.New("device responded with an error code during authentication")
	}
	if resp.Type != AuthOK {
		return d, fmt.Errorf("did not get an affirmative response to the authenticaton request - expected %v but got %v instead", AuthOK, resp.Type)
	}

	return d, nil
}

// newManualDevice lets you create a device by specifying a key and id,
// skipping the authentication phase. All fields aside from the mac address are
// mandatory.
func newManualDevice(ip, mac, key, id string, timeout, deviceType int) (*device, error) {
	parsedip := net.ParseIP(ip)
	if parsedip.String() == "<nil>" {
		return nil, fmt.Errorf("%v is not a valid IP address", ip)
	}
	skipmac := false
	parsedmac, err := net.ParseMAC(mac)
	if err != nil {
		skipmac = true
	}
	keyhex, err := hex.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("key %v is an invalid hex string: %v", key, err)
	}
	if len(keyhex) != 16 {
		return nil, fmt.Errorf("key has length of %v bytes - it should have a length of 16 bytes", len(keyhex))
	}
	idhex, err := hex.DecodeString(id)
	if err != nil {
		return nil, fmt.Errorf("id %v is an invalid hex string: %v", id, err)
	}
	if len(idhex) != 4 {
		return nil, fmt.Errorf("id has length of %v bytes - it should have a length of 4 bytes", len(idhex))
	}

	rand.Seed(time.Now().Unix())
	d := &device{
		remoteAddr: parsedip.String(),
		timeout:    timeout,
		deviceType: deviceType,
		count:      rand.Intn(0xffff),
		key:        keyhex,
		iv:         []byte{0x56, 0x2e, 0x17, 0x99, 0x6d, 0x09, 0x3d, 0x28, 0xdd, 0xb3, 0xba, 0x69, 0x5a, 0x2e, 0x6f, 0x58},
		id:         idhex,
	}
	if !skipmac {
		d.mac = parsedmac
	}

	return d, nil
}

// serverRequest sends a request to the device and waits for a response.
func (d *device) serverRequest(req unencryptedRequest) (Response, error) {
	resp := Response{}

	err := d.setupConnection()
	if err != nil {
		return resp, fmt.Errorf("could not setup UDP listener: %v", err)
	}

	//Logger.Printf("Request: %#v",req)


	encryptedReq, err := d.encryptRequest(req)
	if err != nil {
		//Logger.Printf("%#v",err)
		return resp, err
	}

	retries := 0
	for {
		retries++

		err = d.send(encryptedReq)
		if err != nil {
			if retries < sendRetries {
				continue
			}
			return resp, fmt.Errorf("could not send packet: %v", err)
		}

		resp, err = d.readPacket()
		if err != nil {
			if retries < sendRetries {
				continue
			}
			return resp, fmt.Errorf("error while waiting for device response: %v", err)
		}
		return resp, nil
	}
}

func (d *device) close() {
	if d.conn != nil {
		(*d.conn).Close()
		d.conn = nil
	}
}

func (d *device) setupConnection() error {
	if d.conn != nil {
		return nil
	}

	conn, err := net.ListenPacket("udp4", "")
	if err != nil {
		return err
	}

	d.conn = &conn
	return nil
}

func (d *device) encryptRequest(req unencryptedRequest) ([]byte, error) {
	if len(req.payload)%16 != 0 {
		return []byte{}, fmt.Errorf("length of unencrypted request payload must be a multiple of 16 - got %d instead", len(req.payload))
	}
	d.count = (d.count + 1) & 0xffff
	header := make([]byte, 0x38, 0x38)
	header[0x00] = 0x5a
	header[0x01] = 0xa5
	header[0x02] = 0xaa
	header[0x03] = 0x55
	header[0x04] = 0x5a
	header[0x05] = 0xa5
	header[0x06] = 0xaa
	header[0x07] = 0x55
	header[0x24] = 0x2a
	header[0x25] = 0x27
	header[0x26] = req.command
	header[0x28] = (byte)(d.count & 0xff)
	header[0x29] = (byte)(d.count >> 8)
	header[0x2a] = d.mac[5]
	header[0x2b] = d.mac[4]
	header[0x2c] = d.mac[3]
	header[0x2d] = d.mac[2]
	header[0x2e] = d.mac[1]
	header[0x2f] = d.mac[0]
	header[0x30] = d.id[0]
	header[0x31] = d.id[1]
	header[0x32] = d.id[2]
	header[0x33] = d.id[3]

	checksum := 0xbeaf
	for _, v := range req.payload {
		checksum += (int)(v)
		checksum = checksum & 0xffff
	}

	block, err := aes.NewCipher(d.key)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to create new AES cipher: %v", err)
	}
	mode := cipher.NewCBCEncrypter(block, d.iv)
	encryptedPayload := make([]byte, len(req.payload))
	mode.CryptBlocks(encryptedPayload, req.payload)

	packet := make([]byte, len(header)+len(encryptedPayload))
	copy(packet, header)
	copy(packet[len(header):], encryptedPayload)

	packet[0x34] = (byte)(checksum & 0xff)
	packet[0x35] = (byte)(checksum >> 8)

	checksum = 0xbeaf
	for _, v := range packet {
		checksum += (int)(v)
		checksum = checksum & 0xffff
	}
	packet[0x20] = (byte)(checksum & 0xff)
	packet[0x21] = (byte)(checksum >> 8)

	return packet, nil
}

func (d device) send(packet []byte) error {
	if d.conn == nil {
		return errors.New("could not send packet because a connection does not exist")
	}
	destAddr, err := net.ResolveUDPAddr("udp", d.remoteAddr+":80")
	if err != nil {
		return fmt.Errorf("could not resolve device address %v: %v", d.remoteAddr, err)
	}

	//Logger.Printf("Send to %s:\n%#v",d.remoteAddr,packet)

	_, err = (*d.conn).WriteTo(packet, destAddr)
	if err != nil {
		return fmt.Errorf("could not send packet: %v", err)
	}
	return nil
}

func (d *device) readPacket() (Response, error) {
	var buf [1024]byte
	processedPayload := Response{Type: Unknown}
	if d.conn == nil {
		return processedPayload, errors.New("a connection to the device does not exist")
	}
	(*d.conn).SetReadDeadline(time.Now().Add(time.Duration(d.timeout) * time.Second))
	plen, _, err := (*d.conn).ReadFrom(buf[:])
	if err != nil {
		return processedPayload, fmt.Errorf("error reading UDP packet: %v", err)
	}

	if plen < 0x38+16 {
		return processedPayload, fmt.Errorf("received a packet with a length of %v which is too short", plen)
	}

	encryptedPayload := make([]byte, plen-0x38, plen-0x38)
	copy(encryptedPayload, buf[0x38:plen])

	block, err := aes.NewCipher(d.key)
	if err != nil {
		return processedPayload, fmt.Errorf("error creating new decryption cipher: %v", err)
	}
	payload := make([]byte, len(encryptedPayload), len(encryptedPayload))
	mode := cipher.NewCBCDecrypter(block, d.iv)


	if len(encryptedPayload)%16 != 0 {

		Logger.Printf("%#v",encryptedPayload)
		return processedPayload,errors.New("crypto/cipher: input not full blocks")
	}


	mode.CryptBlocks(payload, encryptedPayload)

	command := buf[0x26]
	Logger.Printf("Command: %#v",command)
	header := d.requestHeader
	if command == 0xe9 {
		copy(d.key, payload[0x04:0x14])
		copy(d.id, payload[:0x04])

		Logger.Printf("Device %v ready - updating to a new key %v and new id %v", d.mac.String(), hex.EncodeToString(d.key), hex.EncodeToString(d.id))
		processedPayload.Type = AuthOK
		return processedPayload, nil
	}

	if command == 0xee || command == 0xef {
		param := payload[len(header)]
		errorCode := (int)(buf[0x22]) | ((int)(buf[0x23]) << 8)
		if errorCode != 0 {
			processedPayload.Type = DeviceError
			return processedPayload, nil
		}
		switch param {
		case 1:
			processedPayload.Type = Temperature
			processedPayload.Data = []byte{(payload[len(header)+0x4]*10 + payload[len(header)+0x5]) / 10}
		case 2:
			processedPayload.Type = CommandOK
		case 4:
			processedPayload.Type = RawData
			processedPayload.Data = make([]byte, len(payload)-len(header)-4, len(payload)-len(header)-4)
			copy(processedPayload.Data, payload[len(header)+4:])
		case 26:
			processedPayload.Data = make([]byte, len(payload)-len(header)-4, len(payload)-len(header)-4)
			copy(processedPayload.Data, payload[len(header)+4:])
			if payload[len(header)+0x4] == 1 {
				processedPayload.Type = RawRFData
			}
		case 27:
			processedPayload.Data = make([]byte, len(payload)-len(header)-4, len(payload)-len(header)-4)
			copy(processedPayload.Data, payload[len(header)+4:])
			if payload[len(header)+0x4] == 1 {
				processedPayload.Type = RawRFData2
			}
		}
		return processedPayload, nil
	}

	log.Printf("Unhandled command %v", command)
	return processedPayload, fmt.Errorf("unhandled command - %v", command)
}

func (d *device) sendString(s string) error {
	data, err := hex.DecodeString(s)
	if err != nil {
		return fmt.Errorf("error converting %v to hex: %v", s, err)
	}
	return d.sendData(data)
}

func (d *device) sendData(data []byte) error {
	header := d.codeSendingHeader
	reqPayload := make([]byte, len(header)+len(data)+4, len(header)+len(data)+4)
	copy(reqPayload, header)
	reqPayload[len(header)] = 0x02
	reqPayload[len(header)+1] = 0x00
	reqPayload[len(header)+2] = 0x00
	reqPayload[len(header)+3] = 0x00
	copy(reqPayload[len(header)+4:], data)
	req := unencryptedRequest{
		command: 0x6a,
		payload: reqPayload,
	}

	defer d.close()
	resp, err := d.serverRequest(req)
	if err != nil {
		return fmt.Errorf("error reading response while trying to send data to device: %v", err)
	}
	if resp.Type == DeviceError {
		return errors.New("device responded with an error code")
	}
	if resp.Type != CommandOK {
		return fmt.Errorf("expected response type %v but got %v instead", CommandOK, resp.Type)
	}
	return nil
}

func (d *device) checkData() (Response, error) {
	resp, err := d.serverRequest(d.checkDataPayload())
	if err != nil {
		return resp, fmt.Errorf("error making CheckData request: %v", err)
	}

	return resp, nil
}

func (d *device) checkRFData() (Response, error) {
	resp, err := d.serverRequest(d.checkRFDataPayload())
	if err != nil {
		return resp, fmt.Errorf("error making CheckRFData request: %v", err)
	}

	return resp, nil
}

func (d *device) checkRFData2() (Response, error) {
	resp, err := d.serverRequest(d.checkRFData2Payload())
	if err != nil {
		return resp, fmt.Errorf("error making CheckRFData2 request: %v", err)
	}

	return resp, nil
}

func (d *device) learn() (Response, error) {
	deadline := time.Now().Add(learnTimeout * time.Second)
	defer d.close()
	_, err := d.serverRequest(d.enterLearningPayload())
	if err != nil {
		return Response{}, fmt.Errorf("error making learning request: %v", err)
	}

	for {
		if time.Now().After(deadline) {
			d.cancelLearn()
			return Response{}, errors.New("learning timeout")
		}

		resp, err := d.checkData()
		if err != nil || resp.Type == DeviceError {
			// If err != nil, it's probably because it's just timed out waiting
			// for a response from check data.
			// If resp.Type is DeviceError, we'll ignore it and wait till we
			// receive a response without an error.
			// In any case, we have a learningtimeout so we won't be looping
			// indefinitely.
			continue
		}
		if resp.Type == RawData {
			return resp, nil
		}
	}
}

// Information on the RF learning sequence can be found at:
// https://github.com/mjg59/python-broadlink/issues/87
func (d *device) learnRF() (Response, error) {

	deadline := time.Now().Add(learnTimeout * time.Second)
	defer d.close()
	_, err := d.serverRequest(d.enterRFSweepPayload())
	if err != nil {
		return Response{}, fmt.Errorf("error making learning request: %v", err)
	}
	log.Print("Successfully sent RF frequency sweep command, waiting for long press...")

	state := 0
	for {
		if time.Now().After(deadline) {
			d.cancelLearn()
			return Response{}, errors.New("learning timeout")
		}

		switch state {
		case 0:
			// Keep sending CheckRFData (check frequency) till we receive a
			// valid RawRFData response.
			resp, err := d.checkRFData()
			if err != nil || resp.Type == DeviceError {
				continue
			}
			if resp.Type == RawRFData {
				log.Print("Check frequency successful, proceeding to find RF packet...")
				state = 1
			}
		case 1:
			// Send CheckRFData2 (find RF packet) once then proceed to next stage
			if _, err = d.checkRFData2(); err == nil {
				log.Print("Find RF packet request sent successfully, proceeding to check data...")
				state = 2
			}
		case 2:
			resp, err := d.checkData()
			if err != nil || resp.Type == DeviceError {
				// If err != nil, it's probably because it's just timed out waiting
				// for a response from check data.
				// If resp.Type is DeviceError, we'll ignore it and wait till we
				// receive a response without an error.
				// In any case, we have a learningtimeout so we won't be looping
				// indefinitely.
				Logger.Printf("Learn RF Err: %#v %#v",err,resp)
				continue
			}
			if resp.Type == RawData {
				log.Println("Everything went ok with rf learning")
				return resp, nil
			}
		}
	}
}

func (d *device) checkTemperature() (Response, error) {
	defer d.close()
	resp, err := d.serverRequest(d.checkTemperaturePayload())
	if err != nil {
		return resp, fmt.Errorf("error making check temperature request: %v", err)
	}
	if resp.Type == DeviceError {
		return resp, errors.New("device responded with an error code")
	}
	return resp, nil
}

func (d *device) cancelLearn() {
	//d.sendPacket(cancelLearnPayload())
	//d.close()
	d.serverRequest(d.cancelLearnPayload())
	d.close()
}

func (d *device) setPowerState(data string) error {
	var state bool
	if data == "00" || data == "0" {
		state = false
	} else if data == "01" || data == "1" {
		state = true
	} else {
		return fmt.Errorf("set power state expects an argument of 0, 00, 1, or 01 - got %v instead", data)
	}

	resp, err := d.serverRequest(d.setPowerStatePayload(state))
	d.close()

	if err != nil {
		return fmt.Errorf("error while making server request to set power state: %v", err)
	}
	if resp.Type == DeviceError {
		return errors.New("device responded with an error code")
	}
	if resp.Type != CommandOK {
		return fmt.Errorf("expected response type %v but got %v instead", CommandOK, resp.Type)
	}
	log.Print("Set power state successful")
	return nil
}

func (d *device) getPowerState() (bool, error) {
	resp, err := d.serverRequest(d.getPowerStatePayload())
	d.close()

	if err != nil {
		return false, fmt.Errorf("error while making server request to get power state: %v", err)
	}
	if resp.Type == DeviceError {
		return false, errors.New("device responded with an error code")
	}
	if resp.Type != CommandOK {
		return false, fmt.Errorf("expected response type %v but got %v instead", CommandOK, resp.Type)
	}
	if len(resp.Data) < 5 {
		return false, fmt.Errorf("received a response payload of length %v - expected at least 5 bytes", len(resp.Data))
	}
	b := resp.Data[4]
	switch b {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("received unknown state - expected 0 or 1 but got 0x%02x instead", b)
	}
}

func authenticatePayload() unencryptedRequest {
	req := unencryptedRequest{
		command: 0x65,
		payload: make([]byte, 0x50, 0x50),
	}
	req.payload[0x04] = 0x31
	req.payload[0x05] = 0x31
	req.payload[0x06] = 0x31
	req.payload[0x07] = 0x31
	req.payload[0x08] = 0x31
	req.payload[0x09] = 0x31
	req.payload[0x0a] = 0x31
	req.payload[0x0b] = 0x31
	req.payload[0x0c] = 0x31
	req.payload[0x0d] = 0x31
	req.payload[0x0e] = 0x31
	req.payload[0x0f] = 0x31
	req.payload[0x10] = 0x31
	req.payload[0x11] = 0x31
	req.payload[0x12] = 0x31
	req.payload[0x1e] = 0x01
	req.payload[0x2d] = 0x01
	req.payload[0x30] = 'T'
	req.payload[0x31] = 'e'
	req.payload[0x32] = 's'
	req.payload[0x33] = 't'
	req.payload[0x34] = ' '
	req.payload[0x35] = ' '
	req.payload[0x36] = '1'

	return req
}

func (d *device) checkDataPayload() unencryptedRequest {
	return unencryptedRequest{
		command: 0x6a,
		payload: d.basicRequestPayload(0x04),
	}
}

func (d *device) enterLearningPayload() unencryptedRequest {
	return unencryptedRequest{
		command: 0x6a,
		payload: d.basicRequestPayload(0x03),
	}
}

func (d *device) checkTemperaturePayload() unencryptedRequest {
	return unencryptedRequest{
		command: 0x6a,
		payload: d.basicRequestPayload(0x01),
	}
}

func (d *device) cancelLearnPayload() unencryptedRequest {
	return unencryptedRequest{
		command: 0x6a,
		payload: d.basicRequestPayload(0x1e),
	}
}

func (d *device) enterRFSweepPayload() unencryptedRequest {
	return unencryptedRequest{
		command: 0x6a,
		payload: d.basicRequestPayload(0x19),
	}
}

func (d *device) checkRFDataPayload() unencryptedRequest {
	return unencryptedRequest{
		command: 0x6a,
		payload: d.basicRequestPayload(0x1a),
	}
}

func (d *device) checkRFData2Payload() unencryptedRequest {
	return unencryptedRequest{
		command: 0x6a,
		payload: d.basicRequestPayload(0x1b),
	}
}

// Based on the following paragraph from https://blog.ipsumdomus.com/broadlink-smart-home-devices-complete-protocol-hack-bc0b4b397af1:
// Command (16 bytes message id 0x6a payload) always has Get (byte 0x1) or Set
// (byte 0x2) at byte 0 and state (On — 0x1 and Off — 0x0) at byte 4.
func (d *device) setPowerStatePayload(state bool) unencryptedRequest {
	var stateValue byte
	if state {
		stateValue = 0x01
	} else {
		stateValue = 0x00
	}
	p := d.basicRequestPayload(0x02)
	p[4] = stateValue
	return unencryptedRequest{
		command: 0x6a,
		payload: p,
	}
}

func (d *device) getPowerStatePayload() unencryptedRequest {
	return unencryptedRequest{
		command: 0x6a,
		payload: d.basicRequestPayload(0x01),
	}
}

func (d *device) basicRequestPayload(command byte) []byte {
	payload := make([]byte, 16, 16)
	header := d.requestHeader
	copy(payload, header)
	payload[len(header)] = command
	return payload
}
