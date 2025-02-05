package mpi3mr

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var showNoLogJData = `
{
  "Controllers": [
    {
      "Command Status": { "CLI Version": "008.0011.0000.0014 Sep 26, 2024", "Operating system": "Linux6.11.0-8-generic",
        "Status Code": 0,
        "Status": "Success",
        "Description": "None"
      },
      "Response Data": {
        "Number of Controllers": 1,
        "Host Name": "ubuntu-server",
        "Operating System ": "Linux6.11.0-8-generic",
        "SL8 Library Version": "08.1113.0000",
        "System Overview": [
          {
            "Ctrl": 0,
            "Product Name": "Cisco Tri-Mode 24G SAS RAID Controller w/4GB Cache",
            "SASAddress": "0X52CF89BD43A7AF80",
            "Personality": "RAID",
            "Status": "Optimal",
            "PD(s)": 1,
            "VD(s)": 1,
            "VNOpt": 0,
            "EPack": "Optimal",
            "SerialNumber": "SPD5001205                      "
          }
        ]
      }
    }
  ]
}
`

func TestStorCli2CmdShow(t *testing.T) {
	controllers, err := parseShowOutput([]byte(showNoLogJData))
	if err != nil {
		t.Fatalf("Failed to parse show output: %v", err)
	}

	if len(controllers) != 1 {
		t.Fatalf("Expected '1' controller, got %d", len(controllers))
	}
}

var showC0ShowNoLogJData = `
{
  "Controllers": [
    {
      "Command Status": {
        "CLI Version": "008.0011.0000.0014 Sep 26, 2024",
        "Operating system": "Linux6.11.0-8-generic",
        "Controller": "0",
        "Status": "Success",
        "Description": "None"
      },
      "Response Data": {
        "Product Name": "Cisco Tri-Mode 24G SAS RAID Controller w/4GB Cache",
        "Board Name": "UCSC-RAID-HP                    ",
        "Board Assembly": "03-50146-00006                  ",
        "Board Tracer Number": "SPD5001205                      ",
        "Board Revision": "00006   ",
        "Chip Name": "SAS4116W                        ",
        "Chip Revision": "B0      ",
        "Package Version": "8.6.2.0-00065-00001",
        "Firmware Version": "8.6.2.0-00000-00001",
        "Firmware Security Version Number": "00.00.00.00",
        "NVDATA Version": "06.0B.00.0D",
        "Driver Name": "mpi3mr",
        "Driver Version": "8.9.1.0.51",
        "SAS Address": "0x52cf89bd43a7af80",
        "Serial Number": "SPD5001205                      ",
        "Controller Time(LocalTime yyyy/mm/dd hh:mm:sec)": "2025/02/03 22:50:27",
        "System Time(LocalTime yyyy/mm/dd hh:mm:sec)": "2025/02/03 22:50:27",
        "Board Mfg Date(yyyy/mm/dd)": "2023/12/20",
        "Controller Personality": "RAID",
        "Max PCIe Link Rate": "0x08 (16GT/s)",
        "Max PCIe Port Width": 16,
        "PCI Address": "00:01:00:0",
        "PCIe Link Width": "X16 Lane(s)",
        "Current Max PCI Link Speed": "16GT/s",
        "Current PCIe Port Width": 16,
        "SAS/SATA": "SAS/SATA-6G, SAS-12G, SAS-22.5G",
        "PCIe": "PCIE-2.5GT, PCIE-5GT, PCIE-8GT, PCIE-16GT",
        "PCI Vendor ID": "0x1000",
        "PCI Device ID": "0x00A5",
        "PCI Subsystem Vendor ID": "0x1137",
        "PCI Subsystem ID": "0x2EB",
        "Security Protocol": "SPDM-1.1.0,1.0.0",
        "PCI Slot Number": 11,
        "Drive Groups": 1,
        "TOPOLOGY": [
          {
            "DG": 0,
            "Span": "-",
            "Row": "-",
            "EID:Slot": "-",
            "PID": "-",
            "Type": "RAID0",
            "State": "-",
            "Status": "-",
            "BT": "N",
            "Size": "893.137 GiB",
            "PDC": "dflt",
            "Secured": "N",
            "FSpace": "N"
          },
          {
            "DG": 0,
            "Span": 0,
            "Row": "-",
            "EID:Slot": "-",
            "PID": "-",
            "Type": "RAID0",
            "State": "-",
            "Status": "-",
            "BT": "N",
            "Size": "893.137 GiB",
            "PDC": "dflt",
            "Secured": "N",
            "FSpace": "N"
          },
          {
            "DG": 0,
            "Span": 0,
            "Row": 0,
            "EID:Slot": "322:2",
            "PID": 312,
            "Type": "DRIVE",
            "State": "Conf",
            "Status": "Online",
            "BT": "N",
            "Size": "893.137 GiB",
            "PDC": "dflt",
            "Secured": "N",
            "FSpace": "-"
          }
        ],
        "Virtual Drives": 1,
        "VD LIST": [
          {
            "DG/VD": "0/1",
            "TYPE": "RAID0",
            "State": "Optl",
            "Access": "RW",
            "CurrentCache": "NR,WB",
            "DefaultCache": "NR,WB",
            "Size": "893.137 GiB",
            "Name": ""
          }
        ],
        "Physical Drives": 1,
        "PD LIST": [
          {
            "EID:Slt": "322:2",
            "PID": 312,
            "State": "Conf",
            "Status": "Online",
            "DG": 0,
            "Size": "893.137 GiB",
            "Intf": "SATA",
            "Med": "SSD",
            "SED_Type": "-",
            "SeSz": "512B",
            "Model": "SAMSUNG MZ7L3960HCJR-00AK1",
            "Sp": "U",
            "LU/NS Count": 1,
            "Alt-EID": "-"
          }
        ],
        "LU/NS LIST": [
          {
            "PID": 312,
            "LUN/NSID": "0/-",
            "Index": 255,
            "Status": "Online",
            "Size": "893.137 GiB"
          }
        ],
        "Enclosures": 1,
        "Enclosure List": [
          {
            "EID": 322,
            "State": "OK",
            "DeviceType": "Logical Enclosure",
            "Slots": 10,
            "PD": 1,
            "Partner-EID": "-",
            "Multipath": "No",
            "PS": 0,
            "Fans": 0,
            "TSs": 0,
            "Alms": 0,
            "SIM": 0,
            "ProdID": "VirtualSES      "
          }
        ],
        "Energy Pack Info": [
          {
            "Type": "Supercap",
            "SubType": "FBU345",
            "Voltage(mV)": 8937,
            "Temperature(C)": 23,
            "Status": "Optimal"
          }
        ]
      }
    }
  ]
}
`

var testc0VAllShowAllJOut = `
{
  "Controllers": [
    {
      "Command Status": {
        "CLI Version": "008.0011.0000.0014 Sep 26, 2024",
        "Operating system": "Linux6.11.0-8-generic",
        "Controller": "0",
        "Status": "Success",
        "Description": "None"
      },
      "Response Data": {
        "Virtual Drives": [
          {
            "VD Info": {
              "DG/VD": "0/1",
              "TYPE": "RAID0",
              "State": "Optl",
              "Access": "RW",
              "CurrentCache": "NR,WB",
              "DefaultCache": "NR,WB",
              "Size": "893.137 GiB",
              "Name": ""
            },
            "PDs": [
              {
                "EID:Slt": "322:2",
                "PID": 312,
                "State": "Conf",
                "Status": "Online",
                "DG": 0,
                "Size": "893.137 GiB",
                "Intf": "SATA",
                "Med": "SSD",
                "SED_Type": "-",
                "SeSz": "512B",
                "Model": "SAMSUNG MZ7L3960HCJR-00AK1",
                "Sp": "U",
                "LU/NS Count": 1,
                "Alt-EID": "-"
              }
            ],
            "VD Properties": {
              "Strip Size": "64 KiB",
              "Block Size": 4096,
              "Number of Blocks": 234130688,
              "Span Depth": 1,
              "Number of Drives": 1,
              "Drive Write Cache Policy": "Default",
              "Default Power Save Policy": "Default",
              "Current Power Save Policy": "Default",
              "Access Policy Status": "VD has desired access",
              "Auto BGI": "Off",
              "Secured": "No",
              "Init State": "No Init",
              "Consistent": "Yes",
              "Morphing": "No",
              "Cache Preserved": "No",
              "Bad Block Exists": "No",
              "VD Ready for OS Requests": "Yes",
              "Reached LD BBM failure threshold": "No",
              "Supported Erase Types": "Simple, Normal, Thorough",
              "Exposed to OS": "Yes",
              "Creation Time(LocalTime yyyy/mm/dd hh:mm:sec)": "2025/01/29 22:07:51",
              "Default Cachebypass Mode": "Cachebypass Not Performed For Any IOs",
              "Current Cachebypass Mode": "Cachebypass Not Performed For Any IOs",
              "SCSI NAA Id": "62cf89bd43a7af80679aa6b751349581",
              "OS Drive Name": "/dev/sdg",
              "Current Unmap Status": "No",
              "Current WriteSame Unmap Status": "No",
              "LU/NS Count used per PD": 1,
              "Data Format for I/O": "None",
              "Serial Number": "0081953451b7a69a6780afa743bd89cf"
            }
          }
        ]
      }
    }
  ]
}
`

/*
var testParseVirtPropObj = VirtualDrives{
	VDInfo: VirtualDrive{
		DGVD:         "0/1",
		Type:         "RAID0",
		State:        "Optl",
		Access:       "RW",
		CurrentCache: "NR,WB",
		DefaultCache: "NR,WB",
		Size:         "893.137 GiB",
		Name:         "",
	},
	PhysicalDrives: []PhysicalDrive{
		{
			EIDSlot:    "322:2",
			PID:        312,
			State:      "Conf",
			Status:     "Online",
			DG:         0,
			Size:       "893.137 GiB",
			Interface:  "SATA",
			Medium:     "SSD",
			SEDType:    "-",
			SectorSize: "512B",
			Model:      "SAMSUNG MZ7L3960HCJR-00AK1",
			SP:         "U",
			LUNSCount:  1,
			AltEID:     "-",
		},
	},
	VDProperties: VirtualDriveProperties{
		StripSize:                    "64 KiB",
		BlockSize:                    4096,
		NumberOfBlocks:               234130688,
		SpanDepth:                    1,
		NumberOfDrives:               1,
		DriveWriteCachePolicy:        "Default",
		DefaultPowerSavePolicy:       "Default",
		CurrentPowerSavePolicy:       "Default",
		AccessPolicyStatus:           "VD has desired access",
		AutoBGI:                      "Off",
		Secured:                      "No",
		InitState:                    "No Init",
		Consistent:                   "Yes",
		Morphing:                     "No",
		CachePreserved:               "No",
		BadBlockExists:               "No",
		VDReadyForOSRequests:         "Yes",
		ReachedLDBBMFailureThreshold: "No",
		SupportedEraseTypes:          "Simple, Normal, Thorough",
		ExposedToOS:                  "Yes",
		CreationTimeLocalTimeString:  "2025/01/29 22:07:51",
		DefaultCachebypassMode:       "Cachebypass Not Performed For Any IOs",
		CurrentCachebypassMode:       "Cachebypass Not Performed For Any IOs",
		SCSINAAId:                    "62cf89bd43a7af80679aa6b751349581",
		OSDriveName:                  "/dev/sdg",
		CurrentUnmapStatus:           "No",
		CurrentWriteSameUnmapStatus:  "No",
		LUNSCountUsedPerPD:           1,
		DataFormatForIO:              "None",
		SerialNumber:                 "0081953451b7a69a6780afa743bd89cf",
	},
}
var testParsePhysicalDrivesObj = PhysicalDrive{
	EIDSlot:    "322:2",
	PID:        312,
	State:      "Conf",
	Status:     "Online",
	DG:         0,
	Size:       "893.137 GiB",
	Interface:  "SATA",
	Medium:     "SSD",
	SEDType:    "-",
	SectorSize: "512B",
	Model:      "SAMSUNG MZ7L3960HCJR-00AK1",
	SP:         "U",
	LUNSCount:  1,
	AltEID:     "-",
}

func TestStorCli2ParseVirtProperties(t *testing.T) {
	ctrlID := 0
	vdMap, err := parseVirtProperties(ctrlID, []byte(testc0VAllShowAllJOut))
	if err != nil {
		t.Fatalf("Error parsing virt properties: %s", err)
	}

	if len(vdMap) != 1 {
		t.Fatalf("expected 1 virtual drive, got %d", len(vdMap))
	}

	want := testParseVirtPropObj.VDProperties
	for _, got := range vdMap {
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("got:\n%+v\nwant:\n+%v", got, want)
		}
	}
}
*/

var expectedControllerBytes = `
{
  "Controller": {
    "ID": 0,
    "PhysicalDrives": {
      "0": {
        "EID:Slt": "322:2",
        "PID": 0,
        "State": "Conf",
        "Status": "Online",
        "DG": 0,
        "Size": "893.137 GiB",
        "Intf": "SATA",
        "Med": "SSD",
        "SED_Type": "-",
        "SeSz": "512B",
        "Model": "SAMSUNG MZ7L3960HCJR-00AK1",
        "Sp": "U",
        "LU/NS Count": 1,
        "Alt-EID": "-"
      }
    },
    "VirtualDrives": {
      "0/1": {
        "DG/VD": "0/1",
        "TYPE": "RAID0",
        "State": "Optl",
        "Access": "RW",
        "CurrentCache": "NR,WB",
        "DefaultCache": "NR,WB",
        "Size": "893.137 GiB",
        "Properties": {
          "Strip Size": "64 KiB",
          "Block Size": 4096,
          "Number of Blocks": 234130688,
          "Span Depth": 1,
          "Number of Drives": 1,
          "Drive Write Cache Policy": "Default",
          "Default Power Save Policy": "Default",
          "Current Power Save Policy": "Default",
          "Access Policy Status": "VD has desired access",
          "Auto BGI": "Off",
          "Secured": "No",
          "Init State": "No Init",
          "Consistent": "Yes",
          "Morphing": "No",
          "Cache Preserved": "No",
          "Bad Block Exists": "No",
          "VD Ready for OS Requests": "Yes",
          "Reached LD BBM failure threshold": "No",
          "Supported Erase Types": "Simple, Normal, Thorough",
          "Exposed to OS": "Yes",
          "Creation Time(LocalTime yyyy/mm/dd hh:mm:sec)": "2025/01/29 22:07:51",
          "Default Cachebypass Mode": "Cachebypass Not Performed For Any IOs",
          "Current Cachebypass Mode": "Cachebypass Not Performed For Any IOs",
          "SCSI NAA Id": "62cf89bd43a7af80679aa6b751349581",
          "OS Drive Name": "/dev/sdg",
          "Current Unmap Status": "No",
          "Current WriteSame Unmap Status": "No",
          "LU/NS Count used per PD": 1,
          "Data Format for I/O": "None",
          "Serial Number": "0081953451b7a69a6780afa743bd89cf"
        },
        "PhysicalDrives": [
          {
            "EID:Slt": "322:2",
            "PID": 0,
            "State": "Conf",
            "Status": "Online",
            "DG": 0,
            "Size": "893.137 GiB",
            "Intf": "SATA",
            "Med": "SSD",
            "SED_Type": "-",
            "SeSz": "512B",
            "Model": "SAMSUNG MZ7L3960HCJR-00AK1",
            "Sp": "U",
            "LU/NS Count": 1,
            "Alt-EID": "-"
          }
        ]
      }
    }
  }
}
`

func TestStorCli2NewController(t *testing.T) {
	ctrlID := 0

	cxShowNoLogJOut := []byte(showC0ShowNoLogJData)
	cxVallShowAllNoLogJOut := []byte(testc0VAllShowAllJOut)

	got, err := newController(ctrlID, cxShowNoLogJOut, cxVallShowAllNoLogJOut)
	if err != nil {
		t.Fatalf("failed to create new controller")
	}

	// we wrap the controller in a struct so we can capture the full struture in JSON
	type Payload struct {
		Controller Controller `json:"Controller"`
	}
	p := Payload{}
	p.Controller = got

	// Load the expectec controller
	want := Payload{}
	if err := json.Unmarshal([]byte(expectedControllerBytes), &want); err != nil {
		t.Fatalf("Failed to unmarshal expected controller: %s", err)
	}

	if diff := cmp.Diff(p, want); diff != "" {
		t.Errorf("got:\n%+v\nwant:\n+%v", got, want)
	}
}
