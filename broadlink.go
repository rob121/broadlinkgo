package broadlinkgo

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
	"sort"
)

const defaultTimeout = 5 // seconds

// Broadlink keeps a track of all the devices and sockets.
type Broadlink struct {
	timeout int // in seconds
	devices []*device
	lookup  map[string]*device
}

// NewBroadlink creates and initializes a new Broadlink struct.
func NewBroadlink() Broadlink {
	b := Broadlink{
		timeout: defaultTimeout,
		lookup:  make(map[string]*device),
	}
	return b
}

// WithTimeout sets the timeout for all subsequent read operations.
func (b *Broadlink) WithTimeout(t int) *Broadlink {
	b.timeout = t
	return b
}

// Count returns the number of devices that were discovered.
func (b Broadlink) Count() int {
	return len(b.devices)
}

func (b Broadlink) DeviceTypes() map[int]string {

	var kd = make(map[int]string)
	for _, d := range knownDevices {

		if d.supported == true {

			kd[d.deviceType] = d.name

		}

	}

	return kd
}

// Discover will populate the Broadlink struct with a device.
func (b *Broadlink) DiscoverHost(host string) error {
	conn, err := net.ListenPacket("udp4", "")
	if err != nil {
		return fmt.Errorf("could not bind UDP listener: %v", err)
	}
	defer conn.Close()

	log.Printf("Listening to address %v", conn.LocalAddr().String())
	err = sendHelloPacketToHost(conn, host)
	if err != nil {
		return fmt.Errorf("error sending broadcast packet: %v", err)
	}
	b.readPacket(conn)

	return nil
}

// Discover will populate the Broadlink struct with a slice of Devices.
func (b *Broadlink) Discover() error {
	return b.DiscoverHost("255.255.255.255")
}

// Learn sends a learn command to the specified device. If id is an empty string it selects the first device.
func (b *Broadlink) Learn(id string) (string, error) {
	d, err := b.deviceIsCapableOfIR(id)
	if err != nil {
		return "", err
	}

	resp, err := d.learn()
	if err != nil {
		return "", fmt.Errorf("error while calling learn: %v", err)
	}

	log.Print("Learn successful")
	return hex.EncodeToString(resp.Data), nil
}

// LearnRF sends an RF Sweep command to the specified device. If id is an empty string it selects the first device.
func (b *Broadlink) LearnRF(id string) (string, error) {
	d, err := b.deviceIsCapableOfRF(id)
	if err != nil {
		return "", err
	}

	resp, err := d.learnRF()
	if err != nil {
		return "", fmt.Errorf("error while calling learn RF: %v", err)
	}

	log.Print("Learn RF successful")
	return hex.EncodeToString(resp.Data), nil
}

// Execute looks at the device type and decides if it should call send() or
// setPowerState().
func (b *Broadlink) Execute(id, s string) error {
	d, err := b.deviceExistsAndIsKnown(id)
	if err != nil {
		return err
	}
	devChar := isKnownDevice(d.deviceType)
	if devChar.power {
		l := len(s)
		if l != 1 && l != 2 {
			return fmt.Errorf("device %v is a power outlet and can only accept the data of 0, 00, 1, or 01 - got %v instead", d.mac.String(), s)
		}
		return d.setPowerState(s)
	}
	if devChar.ir || devChar.rf {
		return d.sendString(s)
	}
	return fmt.Errorf("device %v device type %v (0x%04x) is not capable of power control, IR, and RF", d.mac.String(), d.deviceType, d.deviceType)
}

// GetPowerState queries a WiFi-enabled power outlet and returns its state (on or off).
func (b *Broadlink) GetPowerState(id string) (bool, error) {
	d, err := b.deviceIsCapableOfPowerControl(id)
	if err != nil {
		return false, err
	}
	return d.getPowerState()
}

func (b *Broadlink) AddManualDevice(ip string, macs string, deviceType int) error {

	mac, err := net.ParseMAC(macs)

	if err != nil {

	}

	devChar := isKnownDevice(deviceType)
	if !devChar.supported {
		return fmt.Errorf("device type %v (0x%04x) is not supported", deviceType, deviceType)
	}
	d, err := newDevice(ip, mac, b.timeout, devChar)
	if err != nil {
		return err
	}
	if b.getDevice(d.remoteAddr) != nil {
		log.Printf("A device with IP %v already exists - skipping manual add", d.remoteAddr)
		return nil
	}
	hw := d.mac.String()
	if (len(hw) > 0) && (b.getDevice(hw) != nil) {
		log.Printf("A device with MAC %v already exists - skipping manual add", hw)
	}
	b.devices = append(b.devices, d)
	
	if len(hw) > 0 {
		b.lookup[strings.ToLower(hw)] = d
	}else{
		
	b.lookup[d.remoteAddr] = d

	
	}

	return nil
}

func (b Broadlink) DeviceExists(id string) bool { 
	
	d := b.getDevice(id)
	
	if(d!=nil){
		
		return true
		
	}
	
	return false
	
	
}

func (b Broadlink) getDevice(id string) *device {
	d, ok := b.lookup[strings.ToLower(id)]
	if !ok {
		return nil
	}
	return d
}

func SortMapStringSlice(m map[string][]string) map[string][]string {
	
out := make(map[string][]string)	

keys := make([]string, 0, len(m))

for k := range m {
        keys = append(keys, k)
}

sort.Strings(keys)

for _, k := range keys {
        
        
        out[k]=m[k]
}

	
return out
	
}

func (b Broadlink) DeviceIds() map[string][]string {

	var lkp = make(map[string][]string)

	for k, v := range b.lookup {
		
		typ := strconv.Itoa(v.deviceType)

		lkp[k] = []string{v.remoteAddr,typ}

	}
	
	lkp = SortMapStringSlice(lkp)

	return lkp

}

func (b *Broadlink) readPacket(conn net.PacketConn) {
	var buf [1024]byte
	if b.timeout <= 0 {
		b.timeout = defaultTimeout
	}
	for {
		conn.SetReadDeadline(time.Now().Add(time.Duration(b.timeout) * time.Second))
		plen, remote, err := conn.ReadFrom(buf[:])
		if err != nil {
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				break
			}
			log.Printf("Error reading UDP packet: %v", err)
		}
		log.Printf("Received packet of length %v bytes from %v", plen, remote.String())
		if plen < 0x40 {
			log.Print("Ignoring packet because it is too short")
			return
		}
		var mac net.HardwareAddr
		mac = append(mac, buf[0x3f])
		mac = append(mac, buf[0x3e])
		mac = append(mac, buf[0x3d])
		mac = append(mac, buf[0x3c])
		mac = append(mac, buf[0x3b])
		mac = append(mac, buf[0x3a])

		deviceType := (int)(buf[0x34]) | ((int)(buf[0x35]) << 8)

		b.addDevice(remote, mac, deviceType)
	}
}

func (b *Broadlink) addDevice(remote net.Addr, mac net.HardwareAddr, deviceType int) {
	remoteAddr := remote.String()
	if strings.Contains(remoteAddr, ":") {
		remoteAddr = remoteAddr[:strings.Index(remoteAddr, ":")]
	}
	devChar := isKnownDevice(deviceType)
	if !devChar.known {
		log.Printf("Unknown device (0x%04x) at address %v, MAC %v", deviceType, remoteAddr, mac.String())
		return
	}
	if !devChar.supported {
		log.Printf("Unsupported %v (0x%04x) found at address %v, MAC %v", devChar.name, deviceType, remoteAddr, mac.String())
	}

	_, ipOK := b.lookup[strings.ToLower(remoteAddr)]
	_, macOK := b.lookup[strings.ToLower(mac.String())]
	if ipOK || macOK {
		log.Printf("We already know about %v, MAC %v - skipping", remoteAddr, mac.String())
		return
	}
	log.Printf("Found a supported %v, device type %d (0x%04x) at address %v, MAC %v", devChar.name, deviceType, deviceType, remoteAddr, mac.String())
	dev, err := newDevice(remoteAddr, mac, b.timeout, devChar)
	if err != nil {
		log.Printf("Error creating new device: %v", err)
		return
	}
	b.devices = append(b.devices, dev)
	//b.lookup[strings.ToLower(remoteAddr)] = dev
	//log.Println(mac.String())
	// one device lookup is fine
	b.lookup[strings.ToLower(mac.String())] = dev
}

func (b *Broadlink) RemoveDevice(key string){
	
	
	delete(b.lookup,key)
	
	
}

func sendHelloPacketToHost(conn net.PacketConn, host string) error {
	ip, port, err := parseIPAndPort(conn.LocalAddr().String())
	if err != nil {
		return err
	}

	var packet [0x30]byte

	t := currentTime()
	copy(packet[0x08:], t[:])
	copy(packet[0x18:], ip[:])
	copy(packet[0x1c:], port[:])
	packet[0x26] = 6
	checksum := calculateChecksum(packet[:])
	copy(packet[0x20:], checksum[:])

	return sendPacket(packet[:], conn, fmt.Sprintf("%s:80", host))
}

func sendBroadcastPacket(conn net.PacketConn) error {
	return sendHelloPacketToHost(conn, "255.255.255.255")
}

func sendPacket(p []byte, conn net.PacketConn, dest string) error {
	destAddr, err := net.ResolveUDPAddr("udp", dest)
	if err != nil {
		return fmt.Errorf("could not resolve broadcast address: %v", err)
	}

	_, err = conn.WriteTo(p, destAddr)
	if err != nil {
		return fmt.Errorf("error while writing broadcast message: %v", err)
	}

	return nil
}

func parseIPAndPort(address string) ([4]byte, [2]byte, error) {
	var ip [4]byte
	var port [2]byte

	if !strings.Contains(address, ":") {
		return ip, port, fmt.Errorf("%v is not of the form XXX.XXX.XXX.XXX:XXX", address)
	}

	index := strings.Index(address, ":")
	p, err := strconv.Atoi(address[index+1:])
	if err != nil {
		return [4]byte{}, [2]byte{},

			fmt.Errorf("could not parse port number %v", address[index+1:])
	}
	port[0] = (byte)(p & 0xff)
	port[1] = (byte)(p >> 8)

	components := strings.Split(address[:index], ".")
	if len(components) != 4 {
		return ip, port, fmt.Errorf("%v is not of the form XXX.XXX.XXX.XXX", address[:index])
	}

	for i := 0; i < 4; i++ {
		tmp, err := strconv.Atoi(components[i])
		if err != nil || tmp < 0 || tmp > 255 {
			return ip, port, fmt.Errorf("%v is not a valid IP address", address[:index])
		}
		ip[i] = (byte)(tmp)
	}

	return ip, port, nil
}

func currentTime() [12]byte {
	var b [12]byte

	now := time.Now()
	_, offset := now.Local().Zone()
	offset = offset / 3600

	if offset < 0 {
		b[0] = (byte)(0xff + offset - 1)
		b[1] = 0xff
		b[2] = 0xff
		b[3] = 0xff
	} else {
		b[0] = (byte)(offset)
		b[1] = 0
		b[2] = 0
		b[3] = 0
	}

	year := now.Year()
	b[4] = (byte)(year & 0xff)
	b[5] = (byte)(year >> 8)
	b[6] = (byte)(now.Minute())
	b[7] = (byte)(now.Hour())
	b[8] = (byte)(year % 100)
	b[9] = (byte)(now.Weekday())
	b[10] = (byte)(now.Day())
	b[11] = (byte)(now.Month())

	return b
}

func calculateChecksum(p []byte) [2]byte {
	checksum := 0xbeaf

	for _, v := range p {
		checksum += (int)(v)
	}

	checksum = checksum & 0xffff

	return [2]byte{(byte)(checksum & 0xff), (byte)(checksum >> 8)}
}

func (b Broadlink) deviceExistsAndIsKnown(id string) (*device, error) {
	if len(b.devices) == 0 {
		return nil, fmt.Errorf("no devices")
	}
	var d *device
	if len(id) == 0 {
		d = b.devices[0]
	} else {
		d = b.getDevice(id)
		if d == nil {
			return nil, fmt.Errorf("%v is not a known device", id)
		}
	}
	return d, nil
}

func (b Broadlink) deviceIsCapableOfIR(id string) (*device, error) {
	d, err := b.deviceExistsAndIsKnown(id)
	if err != nil {
		return nil, err
	}

	devChar := isKnownDevice(d.deviceType)
	if !devChar.ir {
		return d, fmt.Errorf("device %v is of device type %v (0x%04x) and is not capable of sending and receiving IR", d.mac.String(), d.deviceType, d.deviceType)
	}
	return d, nil
}

func (b Broadlink) deviceIsCapableOfRF(id string) (*device, error) {
	d, err := b.deviceExistsAndIsKnown(id)
	if err != nil {
		return nil, err
	}

	devChar := isKnownDevice(d.deviceType)
	if !devChar.rf {
		return d, fmt.Errorf("device %v is of device type %v (0x%04x) and is not capable of sending and receiving RF", d.mac.String(), d.deviceType, d.deviceType)
	}
	return d, nil
}

func (b Broadlink) deviceIsCapableOfPowerControl(id string) (*device, error) {
	d, err := b.deviceExistsAndIsKnown(id)
	if err != nil {
		return nil, err
	}

	devChar := isKnownDevice(d.deviceType)
	if !devChar.power {
		return d, fmt.Errorf("device %v is of device type %v (0x%04x) and is not capable of power control", d.mac.String(), d.deviceType, d.deviceType)
	}
	return d, nil
}
