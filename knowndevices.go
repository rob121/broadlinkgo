package broadlinkgo

import "errors"

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

type KnownDevice struct {
	DeviceType        int
	Name              string
	Supported         bool
	Ir                bool
	Rf                bool
	Power             bool
	RequestHeader     []byte
	CodeSendingHeader []byte
}

var knownDevices = []KnownDevice{
	KnownDevice{DeviceType: 0x2737, Name: "Broadlink RM Mini", Supported: true, Ir: true, Rf: false, Power: false},
	KnownDevice{DeviceType: 0x27c2, Name: "Broadlink RM Mini 3", Supported: true, Ir: true, Rf: false, Power: false},
	KnownDevice{DeviceType: 0x5f36, Name: "Broadlink RM Mini 3 (RM4 update)", Supported: true, Ir: true, Rf: true, Power: false, RequestHeader: []byte{0x04, 0x00}, CodeSendingHeader: []byte{0xd0, 0x00}},
	KnownDevice{DeviceType: 0x273d, Name: "Broadlink RM Pro Phicom", Supported: true, Ir: true, Rf: false, Power: false},
	KnownDevice{DeviceType: 0x2712, Name: "Broadlink RM2", Supported: true, Ir: true, Rf: false, Power: false},
	KnownDevice{DeviceType: 0x2783, Name: "Broadlink RM2 Home Plus", Supported: true, Ir: true, Rf: false, Power: false},
	KnownDevice{DeviceType: 0x277c, Name: "Broadlink RM2 Home Plus GDT", Supported: true, Ir: true, Rf: false, Power: false},
	KnownDevice{DeviceType: 0x278f, Name: "Broadlink RM Mini Shate", Supported: true, Ir: true, Rf: false, Power: false},
	KnownDevice{DeviceType: 0x272a, Name: "Broadlink RM2 Pro Plus", Supported: true, Ir: true, Rf: true, Power: false},
	KnownDevice{DeviceType: 0x2787, Name: "Broadlink RM2 Pro Plus v2", Supported: true, Ir: true, Rf: true, Power: false},
	KnownDevice{DeviceType: 0x278b, Name: "Broadlink RM2 Pro Plus BL", Supported: true, Ir: true, Rf: true, Power: false},
	KnownDevice{DeviceType: 0x279d, Name: "Broadlink RM3 Pro Plus", Supported: true, Ir: true, Rf: true, Power: false},
	KnownDevice{DeviceType: 0x6026, Name: "Broadlink RM4 Pro V1", Supported: true, Ir: true, Rf: true, Power: false, RequestHeader: []byte{0x04, 0x00}, CodeSendingHeader: []byte{0xd0, 0x00}},
	KnownDevice{DeviceType: 0x61a2, Name: "Broadlink RM4 Pro V2", Supported: true, Ir: true, Rf: true, Power: false, RequestHeader: []byte{0x04, 0x00}, CodeSendingHeader: []byte{0xd0, 0x00}},
	KnownDevice{DeviceType: 0x649b, Name: "Broadlink RM4 Pro V3", Supported: true, Ir: true, Rf: true, Power: false, RequestHeader: []byte{0x04, 0x00}, CodeSendingHeader: []byte{0xd0, 0x00}},
	KnownDevice{DeviceType: 0x653c, Name: "Broadlink RM4 Pro V4", Supported: true, Ir: true, Rf: true, Power: false, RequestHeader: []byte{0x04, 0x00}, CodeSendingHeader: []byte{0xd0, 0x00}},
	KnownDevice{DeviceType: 0x6026, Name: "Broadlink RM4 Pro Plus", Supported: true, Ir: true, Rf: true, Power: false, RequestHeader: []byte{0x04, 0x00}, CodeSendingHeader: []byte{0xd0, 0x00}},
	KnownDevice{DeviceType: 0x27a9, Name: "Broadlink RM3 Pro Plus v2", Supported: true, Ir: true, Rf: true, Power: false},
	KnownDevice{DeviceType: 0, Name: "Broadlink SP1", Supported: true, Ir: false, Rf: false, Power: true},
	KnownDevice{DeviceType: 0x2711, Name: "Broadlink SP2", Supported: true, Ir: false, Rf: false, Power: true},
	KnownDevice{DeviceType: 0x2719, Name: "Honeywell SP2", Supported: true, Ir: false, Rf: false, Power: true},
	KnownDevice{DeviceType: 0x7919, Name: "Honeywell SP2", Supported: true, Ir: false, Rf: false, Power: true},
	KnownDevice{DeviceType: 0x271a, Name: "Honeywell SP2", Supported: true, Ir: false, Rf: false, Power: true},
	KnownDevice{DeviceType: 0x791a, Name: "Honeywell SP2", Supported: true, Ir: false, Rf: false, Power: true},
	KnownDevice{DeviceType: 0x2733, Name: "OEM Branded SP Mini", Supported: false},
	KnownDevice{DeviceType: 0x273e, Name: "OEM Branded SP Mini", Supported: false},
	KnownDevice{DeviceType: 0x2720, Name: "Broadlink SP Mini", Supported: false},
	KnownDevice{DeviceType: 0x753e, Name: "Broadlink SP 3", Supported: false},
	KnownDevice{DeviceType: 0x2728, Name: "Broadlink SPMini 2", Supported: false},
	KnownDevice{DeviceType: 0x2736, Name: "Broadlink SPMini Plus", Supported: false},
	KnownDevice{DeviceType: 0x2714, Name: "Broadlink A1", Supported: false},
	KnownDevice{DeviceType: 0x4eb5, Name: "Broadlink MP1", Supported: false},
	KnownDevice{DeviceType: 0x2722, Name: "Broadlink S1 (SmartOne Alarm Kit)", Supported: false},
	KnownDevice{DeviceType: 0x4e4d, Name: "Dooya DT360E (DOOYA_CURTAIN_V2) or Hysen Heating Controller", Supported: false},
}

//force add a device, useful in testing for new devices, error if the same name label is used

func AddKnownDevice(kd KnownDevice) error{



	for _, d := range knownDevices {

		if(d.Name == kd.Name) {

			return errors.New("Unable to add device with the same Name")


		}

	}

	knownDevices = append(knownDevices,kd)

	return nil
}


func isKnownDevice(dt int) deviceCharacteristics {
	resp := deviceCharacteristics{}
	for _, d := range knownDevices {
		if dt == d.DeviceType {
			resp.deviceType = d.DeviceType
			resp.known = true
			resp.name = d.Name
			resp.supported = d.Supported
			resp.ir = d.Ir
			resp.rf = d.Rf
			resp.power = d.Power
			resp.requestHeader = d.RequestHeader
			resp.codeSendingHeader = d.CodeSendingHeader
			break
		}
	}
	return resp
}
