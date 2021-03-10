package broadlinkgo

type deviceCharacteristics struct {
	deviceType        int
	known             bool
	name              string
	supported         bool
	ir                bool
	rf                bool
	power             bool
	requestHeader     []byte
	codeSendingHeader []byte
}

type knownDevice struct {
	deviceType        int
	name              string
	supported         bool
	ir                bool
	rf                bool
	power             bool
	requestHeader     []byte
	codeSendingHeader []byte
}

var knownDevices = []knownDevice{
	knownDevice{deviceType: 0x2737, name: "Broadlink RM Mini", supported: true, ir: true, rf: false, power: false},
	knownDevice{deviceType: 0x27c2, name: "Broadlink RM Mini 3", supported: true, ir: true, rf: false, power: false},
	knownDevice{deviceType: 0x273d, name: "Broadlink RM Pro Phicom", supported: true, ir: true, rf: false, power: false},
	knownDevice{deviceType: 0x2712, name: "Broadlink RM2", supported: true, ir: true, rf: false, power: false},
	knownDevice{deviceType: 0x2783, name: "Broadlink RM2 Home Plus", supported: true, ir: true, rf: false, power: false},
	knownDevice{deviceType: 0x277c, name: "Broadlink RM2 Home Plus GDT", supported: true, ir: true, rf: false, power: false},
	knownDevice{deviceType: 0x278f, name: "Broadlink RM Mini Shate", supported: true, ir: true, rf: false, power: false},
	knownDevice{deviceType: 0x272a, name: "Broadlink RM2 Pro Plus", supported: true, ir: true, rf: true, power: false},
	knownDevice{deviceType: 0x2787, name: "Broadlink RM2 Pro Plus v2", supported: true, ir: true, rf: true, power: false},
	knownDevice{deviceType: 0x278b, name: "Broadlink RM2 Pro Plus BL", supported: true, ir: true, rf: true, power: false},
	knownDevice{deviceType: 0x279d, name: "Broadlink RM3 Pro Plus", supported: true, ir: true, rf: true, power: false},
	knownDevice{deviceType: 0x6026, name: "Broadlink RM4 Pro Plus", supported: true, ir: true, rf: true, power: false, requestHeader: []byte{0x04, 0x00}, codeSendingHeader: []byte{0xd0, 0x00}},
	knownDevice{deviceType: 0x27a9, name: "Broadlink RM3 Pro Plus v2", supported: true, ir: true, rf: true, power: false},
	knownDevice{deviceType: 0, name: "Broadlink SP1", supported: true, ir: false, rf: false, power: true},
	knownDevice{deviceType: 0x2711, name: "Broadlink SP2", supported: true, ir: false, rf: false, power: true},
	knownDevice{deviceType: 0x2719, name: "Honeywell SP2", supported: true, ir: false, rf: false, power: true},
	knownDevice{deviceType: 0x7919, name: "Honeywell SP2", supported: true, ir: false, rf: false, power: true},
	knownDevice{deviceType: 0x271a, name: "Honeywell SP2", supported: true, ir: false, rf: false, power: true},
	knownDevice{deviceType: 0x791a, name: "Honeywell SP2", supported: true, ir: false, rf: false, power: true},
	knownDevice{deviceType: 0x2733, name: "OEM Branded SP Mini", supported: false},
	knownDevice{deviceType: 0x273e, name: "OEM Branded SP Mini", supported: false},
	knownDevice{deviceType: 0x2720, name: "Broadlink SP Mini", supported: false},
	knownDevice{deviceType: 0x753e, name: "Broadlink SP 3", supported: false},
	knownDevice{deviceType: 0x2728, name: "Broadlink SPMini 2", supported: false},
	knownDevice{deviceType: 0x2736, name: "Broadlink SPMini Plus", supported: false},
	knownDevice{deviceType: 0x2714, name: "Broadlink A1", supported: false},
	knownDevice{deviceType: 0x4eb5, name: "Broadlink MP1", supported: false},
	knownDevice{deviceType: 0x2722, name: "Broadlink S1 (SmartOne Alarm Kit)", supported: false},
	knownDevice{deviceType: 0x4e4d, name: "Dooya DT360E (DOOYA_CURTAIN_V2) or Hysen Heating Controller", supported: false},
}

func isKnownDevice(dt int) deviceCharacteristics {
	resp := deviceCharacteristics{}
	for _, d := range knownDevices {
		if dt == d.deviceType {
			resp.deviceType = d.deviceType
			resp.known = true
			resp.name = d.name
			resp.supported = d.supported
			resp.ir = d.ir
			resp.rf = d.rf
			resp.power = d.power
			resp.requestHeader = d.requestHeader
			resp.codeSendingHeader = d.codeSendingHeader
			break
		}
	}
	return resp
}
