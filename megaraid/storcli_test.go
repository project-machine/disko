package megaraid

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

var tableData1 = `
-------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC     Size Name
-------------------------------------------------------------
0/0   RAID0 Optl  RW     Yes     NRWTD -   OFF 2.181 TB HDD1
-------------------------------------------------------------
`

var tableData2 = `
EID:Slt DID State  DG       Size Intf Med SED PI SeSz Model            Sp Type
-------------------------------------------------------------------------------
62:1     10 JBOD   -  931.512 GB SAS  HDD N   N  512B ST1000NX0453     U  -
62:2      8 JBOD   -  931.512 GB SAS  HDD N   N  512B ST1000NX0453     U  -
62:3     13 UBUnsp -        0 KB SAS  HDD N   N  512B MZ6ER400HAGL/003 U  -
62:4     14 JBOD   -  372.611 GB SAS  SSD N   N  512B KPM51VUG400G     U  -
-------------------------------------------------------------------------------
`

var tableData3 = `
----------------------------------------------------------------------------
EID:Slt DID State DG     Size Intf Med SED PI SeSz Model            Sp Type
----------------------------------------------------------------------------
134:6     3 Onln   3 2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -
----------------------------------------------------------------------------
`

func TestParseTableData(t *testing.T) {
	for i, d := range []struct {
		tabledata string
		expected  []map[string]string
	}{
		{tableData1, []map[string]string{{
			"DG/VD": "0/0", "TYPE": "RAID0", "State": "Optl", "Access": "RW",
			"Consist": "Yes", "Cache": "NRWTD", "Cac": "-", "sCC": "OFF",
			"Size": "2.181 TB", "Name": "HDD1"}}},
		{tableData2, []map[string]string{
			{"EID:Slt": "62:1", "DID": "10", "State": "JBOD", "DG": "-",
				"Size": "931.512 GB", "Intf": "SAS", "Med": "HDD", "SED": "N", "PI": "N",
				"SeSz": "512B", "Model": "ST1000NX0453", "Sp": "U", "Type": "-"},
			{"EID:Slt": "62:2", "DID": "8", "State": "JBOD", "DG": "-",
				"Size": "931.512 GB", "Intf": "SAS", "Med": "HDD", "SED": "N", "PI": "N",
				"SeSz": "512B", "Model": "ST1000NX0453", "Sp": "U", "Type": "-"},
			{"EID:Slt": "62:3", "DID": "13", "State": "UBUnsp", "DG": "-",
				"Size": "0 KB", "Intf": "SAS", "Med": "HDD", "SED": "N", "PI": "N",
				"SeSz": "512B", "Model": "MZ6ER400HAGL/003", "Sp": "U", "Type": "-"},
			{"EID:Slt": "62:4", "DID": "14", "State": "JBOD", "DG": "-",
				"Size": "372.611 GB", "Intf": "SAS", "Med": "SSD", "SED": "N", "PI": "N",
				"SeSz": "512B", "Model": "KPM51VUG400G", "Sp": "U", "Type": "-"}}},
		{tableData3, []map[string]string{
			{"EID:Slt": "134:6", "DID": "3", "State": "Onln", "DG": "3",
				"Size": "2.181 TB", "Intf": "SAS", "Med": "HDD", "SED": "N", "PI": "N",
				"SeSz": "4 KB", "Model": "ST2400MM0129", "Sp": "U", "Type": "-"},
		}},
	} {
		found := parseTableData(strings.Split(d.tabledata, "\n"))
		if len(found) != len(d.expected) {
			t.Errorf("entry %d had %d elements, expected %d", i, len(found), len(d.expected))
			continue
		}

		for j := range found {
			if !reflect.DeepEqual(d.expected[j], found[j]) {
				t.Errorf("entry %d line %d differed:\n found: %#v\n expct: %#v ", i, j, found[j], d.expected[j])
			}
		}
	}
}

func TestParseCxShow(t *testing.T) {
	tables := []struct {
		pNum, vNum int
		name, data string
	}{
		{3, 2, "raidCxShow", raidCxShow},
		{2, 0, "noraidsupportCxShow", noraidsupportCxShow},
		{3, 0, "jbodCxShow", jbodCxShow},
	}

	for _, d := range tables {
		vds, pds, err := parseCxShow(d.data)

		if err != nil {
			t.Errorf("Failed parsing '%s' with parseCxShow: %s", d.name, err)
		}

		if len(vds) != d.vNum {
			t.Errorf("%s: Expected %d vds found %d\n", d.name, d.vNum, len(vds))
		}

		if len(pds) != d.pNum {
			t.Errorf("%s: Expected %d pds found %d\n", d.name, d.pNum, len(pds))
		}
	}
}

func TestParseCxNoController(t *testing.T) {
	_, _, err := parseCxShow(dallNoController)

	if err != ErrNoController {
		t.Fatalf("parseCxShow dallNoController expected ErrNoController found: %s",
			err)
	}
}

func TestForeignDrive(t *testing.T) {
	_, drives, err := parseCxShow(foreignCxShow)

	if err != nil {
		t.Fatalf("failed parsing: %s", err)
	}

	foreign := -2
	if drives[0].DriveGroup != foreign {
		t.Errorf("Drive 0 drivegroup=%d expected %d", drives[0].DriveGroup, foreign)
	}
}

func TestParseVirtProperties(t *testing.T) {
	var exLen = 5

	propMap, err := parseVirtProperties(sys0CxVallShowAll)

	if err != nil {
		t.Fatalf("Failed parsing with parseCxDallShow: %s", err)
	}

	if len(propMap) != exLen {
		t.Errorf("Expected %d properties, but only found %d", exLen, len(propMap))
	}

	expected := map[string]string{
		"Strip Size":                   "64 KB",
		"Number of Blocks":             "779296768",
		"VD has Emulated PD":           "Yes",
		"Span Depth":                   "1",
		"Number of Drives Per Span":    "1",
		"Write Cache(initial setting)": "WriteThrough",
		"Disk Cache Policy":            "Disk's Default",
		"Encryption":                   "None",
		"Data Protection":              "None",
		"Active Operations":            "None",
		"Exposed to OS":                "Yes",
		"OS Drive Name":                "/dev/sda",
		"Creation Date":                "12-03-2021",
		"Creation Time":                "12:15:08 AM",
		"Emulation type":               "default",
		"Cachebypass size":             "Cachebypass-64k",
		"Cachebypass Mode":             "Cachebypass Intelligent",
		"Is LD Ready for OS Requests":  "Yes",
		"SCSI NAA Id":                  "6cc167e9730322c027dd6f0c462b44b4",
		"Unmap Enabled":                "No",
	}

	p0 := propMap[0]
	if len(p0) != len(expected) {
		t.Fatalf("Expected %d in propMap[1], got %d", len(expected), len(p0))
	}

	for k, v := range expected {
		if p0[k] != v {
			t.Errorf("propMap[0][%s]: expected '%s' found '%s'", k, v, p0[v])
		}
	}
}

func TestParseVirtPropertiesNone(t *testing.T) {
	propMap, err := parseVirtProperties(jbodCxVallShowAll)

	if err != nil {
		t.Fatalf("Failed parsing with parseVirtPropertie: %s", err)
	}

	if len(propMap) != 0 {
		t.Errorf("Found entries %d, expected none", len(propMap))
	}
}

func TestParseVirtPropertiesNoController(t *testing.T) {
	_, err := parseVirtProperties(vallNoController)

	if err != ErrNoController {
		t.Fatalf("storcli /c0/vall vallNoController expected ErrNoController found: %s",
			err)
	}
}

func TestParseVirtPropertiesUnsupported(t *testing.T) {
	_, _, err := parseCxShow(noraidsupportCxVallShowAll)
	if err != ErrUnsupported {
		t.Fatalf("storcli /c0/vall noraidsupportCxVallShowAll expected ErrUnsupported found: %s", err)
	}
}

func TestNewController(t *testing.T) {
	var exVlen, exPlen, exDGlen = 5, 5, 5
	var cID = 0

	ctrl, err := newController(cID, sys0CxShow, sys0CxVallShowAll)
	if err != nil {
		t.Fatalf("newController failed: %s", err)
	}

	if ctrl.ID != cID {
		t.Errorf("Expected ID %d, found %d", cID, ctrl.ID)
	}

	if len(ctrl.Drives) != exPlen {
		t.Errorf("Expected %d drives, found %d", exPlen, len(ctrl.Drives))
	}

	if len(ctrl.DriveGroups) != exDGlen {
		t.Errorf("Expected %d DriveGroups, found %d", exDGlen, len(ctrl.DriveGroups))
	}

	if len(ctrl.VirtDrives) != exVlen {
		t.Errorf("Expected %d VirtDrives, found %d", exVlen, len(ctrl.VirtDrives))
	}

	for _, data := range []struct {
		vdNum, dgNum int
		isSSD        bool
	}{
		{0, 3, false},
		{1, 4, false},
		{2, 0, false},
		{3, 1, true},
		{4, 2, false},
	} {
		foundNum := ctrl.VirtDrives[data.vdNum].DriveGroup
		if foundNum != data.dgNum {
			t.Errorf("VirtDrive %d had driveGroup ID %d, expected %d",
				data.vdNum, foundNum, data.dgNum)
		}

		found := ctrl.DriveGroups[data.vdNum].IsSSD()
		if found != data.isSSD {
			t.Errorf("DriveGroup %d found IsSSD = %t expected %t",
				data.dgNum, found, data.isSSD)
		}
	}
}

func TestJsonDriveGroupSet(t *testing.T) {
	dgs := DriveGroupSet{
		1: &DriveGroup{
			ID: 1,
			Drives: DriveSet{
				3: &Drive{ID: 3, DriveGroup: 1, Model: "demodrive"},
				4: &Drive{ID: 4, DriveGroup: 1, Model: "demodrive"},
			},
		},
		2: &DriveGroup{
			ID: 2,
			Drives: DriveSet{
				12: &Drive{ID: 12, DriveGroup: 2, Model: "demo-ssd"},
				13: &Drive{ID: 13, DriveGroup: 2, Model: "demo-ssd"},
			},
		},
	}

	jbytes, err := json.Marshal(&dgs)
	if err != nil {
		t.Errorf("Failed Marshal of DriveGroupSet: %s", err)
	}

	jstr := string(jbytes)

	// really bad test.
	if !(strings.Contains(jstr, "[12,13]") || strings.Contains(jstr, "[13,12]")) {
		t.Errorf("expected to find '[12,13]' in Json blob, did not.: %s", jstr)
	}

	// verify that the drives do not get inlined
	if strings.Contains(jstr, "demo-ssd") {
		t.Error("Found demo-ssd in json output... Custom marshaller not used?")
	}
}

func TestMediaTypeString(t *testing.T) {
	for _, d := range []struct {
		mtype  MediaType
		expStr string
	}{
		{UnknownMedia, "UNKNOWN"},
		{HDD, "HDD"},
		{SSD, "SSD"},
	} {
		found := d.mtype.String()
		if found != d.expStr {
			t.Errorf("%d as string got %s expected %s", d.mtype, found, d.expStr)
		}
	}
}

// Blob command outputs from here down.
// storcli /c0/vall show all
var vallNoController = `
CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 5.0.0-37-generic
Controller = 0
Status = Failure
Description = Controller 0 not found
`

var dallNoController = `
CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 5.0.0-37-generic
Controller = 0
Status = Failure
Description = Controller 0 not found
`

// storcli /c0 show all
var raidCxShow = `
Generating detailed summary of the adapter, it may take a while to complete.

CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 5.10.19stock-1
Controller = 0
Status = Success
Description = None

Product Name = Cisco 12G Modular Raid Controller with 2GB cache (max 16 drives)
Serial Number = SK74277921
SAS Address =  5cc167e972c89c80
PCI Address = 00:3c:00:00
System Time = 04/01/2021 15:48:24
Mfg. Date = 10/22/17
Controller Time = 04/01/2021 15:48:24
FW Package Build = 51.10.0-3612
BIOS Version = 7.10.03.1_0x070A0402
FW Version = 5.100.00-3310
Driver Name = megaraid_sas
Driver Version = 07.714.04.00-rc1
Current Personality = RAID-Mode
Vendor Id = 0x1000
Device Id = 0x14
SubVendor Id = 0x1137
SubDevice Id = 0x20E
Host Interface = PCI-E
Device Interface = SAS-12G
Bus Number = 60
Device Number = 0
Function Number = 0
Drive Groups = 2

TOPOLOGY :
========

-----------------------------------------------------------------------------
DG Arr Row EID:Slot DID Type  State BT       Size PDC  PI SED DS3  FSpace TR
-----------------------------------------------------------------------------
 0 -   -   -        -   RAID1 Optl  N  557.861 GB dflt N  N   dflt N      N
 0 0   -   -        -   RAID1 Optl  N  557.861 GB dflt N  N   dflt N      N
 0 0   0   134:2    17  DRIVE Onln  N  557.861 GB dflt N  N   dflt -      N
 0 0   1   134:3    18  DRIVE Onln  N  557.861 GB dflt N  N   dflt -      N
 1 -   -   -        -   RAID0 Optl  N  371.597 GB dflt N  N   dflt N      N
 1 0   -   -        -   RAID0 Optl  N  371.597 GB dflt N  N   dflt N      N
 1 0   0   134:1    16  DRIVE Onln  N  371.597 GB dflt N  N   dflt -      N
-----------------------------------------------------------------------------

DG=Disk Group Index|Arr=Array Index|Row=Row Index|EID=Enclosure Device ID
DID=Device ID|Type=Drive Type|Onln=Online|Rbld=Rebuild|Dgrd=Degraded
Pdgd=Partially degraded|Offln=Offline|BT=Background Task Active
PDC=PD Cache|PI=Protection Info|SED=Self Encrypting Drive|Frgn=Foreign
DS3=Dimmer Switch 3|dflt=Default|Msng=Missing|FSpace=Free Space Present
TR=Transport Ready

Virtual Drives = 2

VD LIST :
=======

---------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC       Size Name
---------------------------------------------------------------
0/0   RAID1 Optl  RW     No      NRWTD -   OFF 557.861 GB
1/1   RAID0 Optl  RW     Yes     NRWTD -   OFF 371.597 GB
---------------------------------------------------------------

EID=Enclosure Device ID| VD=Virtual Drive| DG=Drive Group|Rec=Recovery
Cac=CacheCade|OfLn=OffLine|Pdgd=Partially Degraded|Dgrd=Degraded
Optl=Optimal|RO=Read Only|RW=Read
Write|HD=Hidden|TRANS=TransportReady|B=Blocked|
Consist=Consistent|R=Read Ahead Always|NR=No Read Ahead|WB=WriteBack|
AWB=Always WriteBack|WT=WriteThrough|C=Cached IO|D=Direct IO|sCC=Scheduled
Check Consistency

Physical Drives = 3

PD LIST :
=======

------------------------------------------------------------------------------
EID:Slt DID State DG       Size Intf Med SED PI SeSz Model            Sp Type
------------------------------------------------------------------------------
134:1    16 Onln   1 371.597 GB SAS  SSD N   N  512B PX05SVB040       U  -
134:2    17 Onln   0 557.861 GB SAS  HDD N   N  512B ST600MM0208      U  -
134:3    18 Onln   0 557.861 GB SAS  HDD N   N  512B ST600MM0208      U  -
------------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild

Enclosures = 1

Enclosure LIST :
==============

------------------------------------------------------------------------
EID State Slots PD PS Fans TSs Alms SIM Port# ProdID     VendorSpecific
------------------------------------------------------------------------
134 OK       16  3  0    0   0    0   0 -     virtualSES
------------------------------------------------------------------------

EID=Enclosure Device ID |PD=Physical drive count |PS=Power Supply count|
TSs=Temperature sensor count |Alms=Alarm count |SIM=SIM Count


Cachevault_Info :
===============

------------------------------------
Model  State   Temp Mode MfgDate
------------------------------------
CVPM05 Optimal 32C  -    2017/11/28
------------------------------------

`

/*
var raidCxVallShowAll = `
CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 5.10.19stock-1
Controller = 0
Status = Success
Description = None


/c0/v0 :
======

---------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC       Size Name
---------------------------------------------------------------
0/0   RAID1 Optl  RW     No      NRWTD -   OFF 557.861 GB
---------------------------------------------------------------

EID=Enclosure Device ID| VD=Virtual Drive| DG=Drive Group|Rec=Recovery
Cac=CacheCade|OfLn=OffLine|Pdgd=Partially Degraded|Dgrd=Degraded
Optl=Optimal|RO=Read Only|RW=Read Write|HD=Hidden|TRANS=TransportReady|B=Blocked|
Consist=Consistent|R=Read Ahead Always|NR=No Read Ahead|WB=WriteBack|
AWB=Always WriteBack|WT=WriteThrough|C=Cached IO|D=Direct IO|sCC=Scheduled
Check Consistency


PDs for VD 0 :
============

------------------------------------------------------------------------------
EID:Slt DID State DG       Size Intf Med SED PI SeSz Model            Sp Type
------------------------------------------------------------------------------
134:2    17 Onln   0 557.861 GB SAS  HDD N   N  512B ST600MM0208      U  -
134:3    18 Onln   0 557.861 GB SAS  HDD N   N  512B ST600MM0208      U  -
------------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild


VD0 Properties :
==============
Strip Size = 64 KB
Number of Blocks = 1169920000
VD has Emulated PD = No
Span Depth = 1
Number of Drives Per Span = 2
Write Cache(initial setting) = WriteThrough
Disk Cache Policy = Disk's Default
Encryption = None
Data Protection = None
Active Operations = None
Exposed to OS = Yes
OS Drive Name = /dev/sda
Creation Date = 30-04-2018
Creation Time = 09:03:03 PM
Emulation type = default
Cachebypass size = Cachebypass-64k
Cachebypass Mode = Cachebypass Intelligent
Is LD Ready for OS Requests = Yes
SCSI NAA Id = 6cc167e972c89c80227a4107648269fa
Unmap Enabled = No


/c0/v1 :
======

---------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC       Size Name
---------------------------------------------------------------
1/1   RAID0 Optl  RW     Yes     NRWTD -   OFF 371.597 GB
---------------------------------------------------------------

EID=Enclosure Device ID| VD=Virtual Drive| DG=Drive Group|Rec=Recovery
Cac=CacheCade|OfLn=OffLine|Pdgd=Partially Degraded|Dgrd=Degraded
Optl=Optimal|RO=Read Only|RW=Read Write|HD=Hidden|TRANS=TransportReady|B=Blocked|
Consist=Consistent|R=Read Ahead Always|NR=No Read Ahead|WB=WriteBack|
AWB=Always WriteBack|WT=WriteThrough|C=Cached IO|D=Direct IO|sCC=Scheduled
Check Consistency


PDs for VD 1 :
============

------------------------------------------------------------------------------
EID:Slt DID State DG       Size Intf Med SED PI SeSz Model            Sp Type
------------------------------------------------------------------------------
134:1    16 Onln   1 371.597 GB SAS  SSD N   N  512B PX05SVB040       U  -
------------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild


VD1 Properties :
==============
Strip Size = 64 KB
Number of Blocks = 779296768
VD has Emulated PD = Yes
Span Depth = 1
Number of Drives Per Span = 1
Write Cache(initial setting) = WriteThrough
Disk Cache Policy = Disk's Default
Encryption = None
Data Protection = None
Active Operations = None
Exposed to OS = Yes
OS Drive Name = /dev/sdb
Creation Date = 30-04-2018
Creation Time = 09:03:41 PM
Emulation type = default
Cachebypass size = Cachebypass-64k
Cachebypass Mode = Cachebypass Intelligent
Is LD Ready for OS Requests = Yes
SCSI NAA Id = 6cc167e972c89c80227a412d9c77fc8a
Unmap Enabled = No
`
*/

var noraidsupportCxShow = `
CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 5.10.19stock-1
Controller = 0
Status = Success
Description = None

Product Name = UCSC-SAS-M5HD
Serial Number = SKA2570285
SAS Address =  50027e370b866b00
PCI Address = 00:18:00:00
System Time = 04/01/2021 15:49:56
FW Package Build = 11.00.05.02
FW Version = 11.00.05.00
BIOS Version = 09.21.03.00_11.00.02.00
NVDATA Version = 12.02.00.19
Driver Name = mpt3sas
Driver Version = 35.100.00.00
Bus Number = 24
Device Number = 0
Function Number = 0
Vendor Id = 0x1000
Device Id = 0xAE
SubVendor Id = 0x1137
SubDevice Id = 0x211
Board Name = UCSC-SAS-M5HD
Board Assembly = 03-50037-02008
Board Tracer Number = SKA2570285
Physical Drives = 2

PD LIST :
=======

-----------------------------------------------------------------------
EID:Slt DID State DG     Size Intf Med SED PI SeSz Model            Sp
-----------------------------------------------------------------------
2:1       0 JBOD  -  1.819 TB SAS  HDD N   N  512B ST2000NX0433     U
2:2       1 JBOD  -  1.819 TB SAS  HDD N   N  512B ST2000NX0433     U
-----------------------------------------------------------------------

EID-Enclosure Device ID|Slt-Slot No.|DID-Device ID|DG-DriveGroup
UGood-Unconfigured Good|UBad-Unconfigured Bad|Intf-Interface
Med-Media Type|SED-Self Encryptive Drive|PI-Protection Info
SeSz-Sector Size|Sp-Spun|U-Up|D-Down|T-Transition

Requested Boot Drive = Not Set
`

var noraidsupportCxVallShowAll = `
CLI Version = 007.1507.0000.0000 Sep 18, 2020
Operating system = Linux 5.10.19stock-1
Controller = 0
Status = Failure
Description = Un-supported command
`

var jbodCxShow = `
Generating detailed summary of the adapter, it may take a while to complete.

CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 5.8.0-34-generic
Controller = 0
Status = Success
Description = None

Product Name = Cisco 12G SAS Modular Raid Controller
Serial Number = SK74073770
SAS Address =  570708bff8842250
PCI Address = 00:05:00:00
System Time = 04/01/2021 08:49:48
Mfg. Date = 10/05/17
Controller Time = 04/01/2021 15:47:03
FW Package Build = 24.12.1-0110
BIOS Version = 6.30.03.0_4.17.08.00_0xC6130202
FW Version = 4.620.01-7246
Driver Name = megaraid_sas
Driver Version = 07.714.04.00-rc1
Vendor Id = 0x1000
Device Id = 0x5D
SubVendor Id = 0x1137
SubDevice Id = 0xDB
Host Interface = PCI-E
Device Interface = SAS-12G
Bus Number = 5
Device Number = 0
Function Number = 0
JBOD Drives = 3

JBOD LIST :
=========

------------------------------------------------------------------------------
EID:Slt DID State DG       Size Intf Med SED PI SeSz Model            Sp Type
------------------------------------------------------------------------------
62:1     10 JBOD  -  931.512 GB SAS  HDD N   N  512B ST1000NX0453     U  -
62:2      8 JBOD  -  931.512 GB SAS  HDD N   N  512B ST1000NX0453     U  -
62:4     14 JBOD  -  372.611 GB SAS  SSD N   N  512B KPM51VUG400G     U  -
------------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|Onln=Online|
Offln=Offline|Intf=Interface|Med=Media Type|SeSz=Sector Size

Physical Drives = 3

PD LIST :
=======

------------------------------------------------------------------------------
EID:Slt DID State DG       Size Intf Med SED PI SeSz Model            Sp Type
------------------------------------------------------------------------------
62:1     10 JBOD  -  931.512 GB SAS  HDD N   N  512B ST1000NX0453     U  -
62:2      8 JBOD  -  931.512 GB SAS  HDD N   N  512B ST1000NX0453     U  -
62:4     14 JBOD  -  372.611 GB SAS  SSD N   N  512B KPM51VUG400G     U  -
------------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild

Enclosures = 1

Enclosure LIST :
==============

--------------------------------------------------------------------
EID State Slots PD PS Fans TSs Alms SIM Port# ProdID VendorSpecific
--------------------------------------------------------------------
 62 OK        8  3  0    0   0    0   1 -     SGPIO
--------------------------------------------------------------------

EID=Enclosure Device ID |PD=Physical drive count |PS=Power Supply count|
TSs=Temperature sensor count |Alms=Alarm count |SIM=SIM Count
`

var jbodCxVallShowAll = `
CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 5.8.0-34-generic
Controller = 0
Status = Success
Description = No VD's have been configured.
`

var sys0CxShow = `
Generating detailed summary of the adapter, it may take a while to complete.

CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 4.14.211stock-1
Controller = 0
Status = Success
Description = None

Product Name = Cisco 12G Modular Raid Controller with 2GB cache (max 16 drives)
Serial Number = SK84978884
SAS Address =  5cc167e9730322c0
PCI Address = 00:3c:00:00
System Time = 04/01/2021 18:24:24
Mfg. Date = 12/14/18
Controller Time = 04/01/2021 18:24:23
FW Package Build = 51.10.0-3612
BIOS Version = 7.10.03.1_0x070A0402
FW Version = 5.100.00-3310
Driver Name = megaraid_sas
Driver Version = 07.702.06.00-rc1
Current Personality = RAID-Mode
Vendor Id = 0x1000
Device Id = 0x14
SubVendor Id = 0x1137
SubDevice Id = 0x20E
Host Interface = PCI-E
Device Interface = SAS-12G
Bus Number = 60
Device Number = 0
Function Number = 0
Drive Groups = 5

TOPOLOGY :
========

-----------------------------------------------------------------------------
DG Arr Row EID:Slot DID Type  State BT       Size PDC  PI SED DS3  FSpace TR
-----------------------------------------------------------------------------
 0 -   -   -        -   RAID0 Optl  N    2.181 TB enbl N  N   dflt N      N
 0 0   -   -        -   RAID0 Optl  N    2.181 TB enbl N  N   dflt N      N
 0 0   0   134:3    2   DRIVE Onln  N    2.181 TB enbl N  N   dflt -      N
 1 -   -   -        -   RAID0 Optl  N    2.181 TB enbl N  N   dflt N      N
 1 0   -   -        -   RAID0 Optl  N    2.181 TB enbl N  N   dflt N      N
 1 0   0   134:4    1   DRIVE Onln  N    2.181 TB enbl N  N   dflt -      N
 2 -   -   -        -   RAID0 Optl  N    2.181 TB enbl N  N   dflt N      N
 2 0   -   -        -   RAID0 Optl  N    2.181 TB enbl N  N   dflt N      N
 2 0   0   134:5    8   DRIVE Onln  N    2.181 TB enbl N  N   dflt -      N
 3 -   -   -        -   RAID0 Optl  N  371.597 GB dflt N  N   dflt N      N
 3 0   -   -        -   RAID0 Optl  N  371.597 GB dflt N  N   dflt N      N
 3 0   0   134:2    0   DRIVE Onln  N  371.597 GB dflt N  N   dflt -      N
 4 -   -   -        -   RAID0 Optl  N    2.181 TB enbl N  N   dflt N      N
 4 0   -   -        -   RAID0 Optl  N    2.181 TB enbl N  N   dflt N      N
 4 0   0   134:6    6   DRIVE Onln  N    2.181 TB enbl N  N   dflt -      N
-----------------------------------------------------------------------------

DG=Disk Group Index|Arr=Array Index|Row=Row Index|EID=Enclosure Device ID
DID=Device ID|Type=Drive Type|Onln=Online|Rbld=Rebuild|Dgrd=Degraded
Pdgd=Partially degraded|Offln=Offline|BT=Background Task Active
PDC=PD Cache|PI=Protection Info|SED=Self Encrypting Drive|Frgn=Foreign
DS3=Dimmer Switch 3|dflt=Default|Msng=Missing|FSpace=Free Space Present
TR=Transport Ready

Virtual Drives = 5

VD LIST :
=======

---------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC       Size Name
---------------------------------------------------------------
3/0   RAID0 Optl  RW     Yes     NRWTD -   OFF 371.597 GB VD02
4/1   RAID0 Optl  RW     Yes     NRWBD -   OFF   2.181 TB VD06
0/2   RAID0 Optl  RW     Yes     NRWBD -   OFF   2.181 TB VD03
1/3   RAID0 Optl  RW     Yes     NRWBD -   OFF   2.181 TB VD04
2/4   RAID0 Optl  RW     Yes     NRWBD -   OFF   2.181 TB VD05
---------------------------------------------------------------

EID=Enclosure Device ID| VD=Virtual Drive| DG=Drive Group|Rec=Recovery
Cac=CacheCade|OfLn=OffLine|Pdgd=Partially Degraded|Dgrd=Degraded
Optl=Optimal|RO=Read Only|RW=Read Write|HD=Hidden|TRANS=TransportReady|B=Blocked|
Consist=Consistent|R=Read Ahead Always|NR=No Read Ahead|WB=WriteBack|
AWB=Always WriteBack|WT=WriteThrough|C=Cached IO|D=Direct IO|sCC=Scheduled
Check Consistency

Physical Drives = 5

PD LIST :
=======

------------------------------------------------------------------------------
EID:Slt DID State DG       Size Intf Med SED PI SeSz Model            Sp Type
------------------------------------------------------------------------------
134:2     0 Onln   3 371.597 GB SAS  SSD N   N  512B PX05SVB040       U  -
134:3     2 Onln   0   2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -
134:4     1 Onln   1   2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -
134:5     8 Onln   2   2.181 TB SAS  HDD Y   N  4 KB ST2400MM0149     U  -
134:6     6 Onln   4   2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -
------------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild

Enclosures = 1

Enclosure LIST :
==============

------------------------------------------------------------------------
EID State Slots PD PS Fans TSs Alms SIM Port# ProdID     VendorSpecific
------------------------------------------------------------------------
134 OK       16  5  0    0   0    0   0 -     virtualSES
------------------------------------------------------------------------

EID=Enclosure Device ID |PD=Physical drive count |PS=Power Supply count|
TSs=Temperature sensor count |Alms=Alarm count |SIM=SIM Count


Cachevault_Info :
===============

------------------------------------
Model  State   Temp Mode MfgDate
------------------------------------
CVPM05 Optimal 34C  -    2018/10/16
------------------------------------
`

var sys0CxVallShowAll = `
CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 4.14.211stock-1
Controller = 0
Status = Success
Description = None


/c0/v0 :
======

---------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC       Size Name
---------------------------------------------------------------
3/0   RAID0 Optl  RW     Yes     NRWTD -   OFF 371.597 GB VD02
---------------------------------------------------------------

EID=Enclosure Device ID| VD=Virtual Drive| DG=Drive Group|Rec=Recovery
Cac=CacheCade|OfLn=OffLine|Pdgd=Partially Degraded|Dgrd=Degraded
Optl=Optimal|RO=Read Only|RW=Read Write|HD=Hidden|TRANS=TransportReady|B=Blocked|
Consist=Consistent|R=Read Ahead Always|NR=No Read Ahead|WB=WriteBack|
AWB=Always WriteBack|WT=WriteThrough|C=Cached IO|D=Direct IO|sCC=Scheduled
Check Consistency


PDs for VD 0 :
============

------------------------------------------------------------------------------
EID:Slt DID State DG       Size Intf Med SED PI SeSz Model            Sp Type
------------------------------------------------------------------------------
134:2     0 Onln   3 371.597 GB SAS  SSD N   N  512B PX05SVB040       U  -
------------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild


VD0 Properties :
==============
Strip Size = 64 KB
Number of Blocks = 779296768
VD has Emulated PD = Yes
Span Depth = 1
Number of Drives Per Span = 1
Write Cache(initial setting) = WriteThrough
Disk Cache Policy = Disk's Default
Encryption = None
Data Protection = None
Active Operations = None
Exposed to OS = Yes
OS Drive Name = /dev/sda
Creation Date = 12-03-2021
Creation Time = 12:15:08 AM
Emulation type = default
Cachebypass size = Cachebypass-64k
Cachebypass Mode = Cachebypass Intelligent
Is LD Ready for OS Requests = Yes
SCSI NAA Id = 6cc167e9730322c027dd6f0c462b44b4
Unmap Enabled = No


/c0/v1 :
======

-------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC     Size Name
-------------------------------------------------------------
4/1   RAID0 Optl  RW     Yes     NRWBD -   OFF 2.181 TB VD06
-------------------------------------------------------------

EID=Enclosure Device ID| VD=Virtual Drive| DG=Drive Group|Rec=Recovery
Cac=CacheCade|OfLn=OffLine|Pdgd=Partially Degraded|Dgrd=Degraded
Optl=Optimal|RO=Read Only|RW=Read Write|HD=Hidden|TRANS=TransportReady|B=Blocked|
Consist=Consistent|R=Read Ahead Always|NR=No Read Ahead|WB=WriteBack|
AWB=Always WriteBack|WT=WriteThrough|C=Cached IO|D=Direct IO|sCC=Scheduled
Check Consistency


PDs for VD 1 :
============

----------------------------------------------------------------------------
EID:Slt DID State DG     Size Intf Med SED PI SeSz Model            Sp Type
----------------------------------------------------------------------------
134:6     6 Onln   4 2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -
----------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild


VD1 Properties :
==============
Strip Size = 64 KB
Number of Blocks = 585691648
VD has Emulated PD = No
Span Depth = 1
Number of Drives Per Span = 1
Write Cache(initial setting) = WriteBack
Disk Cache Policy = Enabled
Encryption = None
Data Protection = None
Active Operations = None
Exposed to OS = Yes
OS Drive Name = /dev/sdb
Creation Date = 12-03-2021
Creation Time = 12:15:47 AM
Emulation type = default
Cachebypass size = Cachebypass-64k
Cachebypass Mode = Cachebypass Intelligent
Is LD Ready for OS Requests = Yes
SCSI NAA Id = 6cc167e9730322c027dd6f33813a86de
Unmap Enabled = No


/c0/v2 :
======

-------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC     Size Name
-------------------------------------------------------------
0/2   RAID0 Optl  RW     Yes     NRWBD -   OFF 2.181 TB VD03
-------------------------------------------------------------

EID=Enclosure Device ID| VD=Virtual Drive| DG=Drive Group|Rec=Recovery
Cac=CacheCade|OfLn=OffLine|Pdgd=Partially Degraded|Dgrd=Degraded
Optl=Optimal|RO=Read Only|RW=Read Write|HD=Hidden|TRANS=TransportReady|B=Blocked|
Consist=Consistent|R=Read Ahead Always|NR=No Read Ahead|WB=WriteBack|
AWB=Always WriteBack|WT=WriteThrough|C=Cached IO|D=Direct IO|sCC=Scheduled
Check Consistency


PDs for VD 2 :
============

----------------------------------------------------------------------------
EID:Slt DID State DG     Size Intf Med SED PI SeSz Model            Sp Type
----------------------------------------------------------------------------
134:3     2 Onln   0 2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -
----------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild


VD2 Properties :
==============
Strip Size = 64 KB
Number of Blocks = 585691648
VD has Emulated PD = No
Span Depth = 1
Number of Drives Per Span = 1
Write Cache(initial setting) = WriteBack
Disk Cache Policy = Enabled
Encryption = None
Data Protection = None
Active Operations = None
Exposed to OS = Yes
OS Drive Name = /dev/sdc
Creation Date = 21-07-2020
Creation Time = 07:06:08 PM
Emulation type = default
Cachebypass size = Cachebypass-64k
Cachebypass Mode = Cachebypass Intelligent
Is LD Ready for OS Requests = Yes
SCSI NAA Id = 6cc167e9730322c026a9f9205fc6ee24
Unmap Enabled = No


/c0/v3 :
======

-------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC     Size Name
-------------------------------------------------------------
1/3   RAID0 Optl  RW     Yes     NRWBD -   OFF 2.181 TB VD04
-------------------------------------------------------------

EID=Enclosure Device ID| VD=Virtual Drive| DG=Drive Group|Rec=Recovery
Cac=CacheCade|OfLn=OffLine|Pdgd=Partially Degraded|Dgrd=Degraded
Optl=Optimal|RO=Read Only|RW=Read Write|HD=Hidden|TRANS=TransportReady|B=Blocked|
Consist=Consistent|R=Read Ahead Always|NR=No Read Ahead|WB=WriteBack|
AWB=Always WriteBack|WT=WriteThrough|C=Cached IO|D=Direct IO|sCC=Scheduled
Check Consistency


PDs for VD 3 :
============

----------------------------------------------------------------------------
EID:Slt DID State DG     Size Intf Med SED PI SeSz Model            Sp Type
----------------------------------------------------------------------------
134:4     1 Onln   1 2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -
----------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild


VD3 Properties :
==============
Strip Size = 64 KB
Number of Blocks = 585691648
VD has Emulated PD = No
Span Depth = 1
Number of Drives Per Span = 1
Write Cache(initial setting) = WriteBack
Disk Cache Policy = Enabled
Encryption = None
Data Protection = None
Active Operations = None
Exposed to OS = Yes
OS Drive Name = /dev/sdd
Creation Date = 21-07-2020
Creation Time = 07:06:48 PM
Emulation type = default
Cachebypass size = Cachebypass-64k
Cachebypass Mode = Cachebypass Intelligent
Is LD Ready for OS Requests = Yes
SCSI NAA Id = 6cc167e9730322c026a9f9489aaef7b2
Unmap Enabled = No


/c0/v4 :
======

-------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC     Size Name
-------------------------------------------------------------
2/4   RAID0 Optl  RW     Yes     NRWBD -   OFF 2.181 TB VD05
-------------------------------------------------------------

EID=Enclosure Device ID| VD=Virtual Drive| DG=Drive Group|Rec=Recovery
Cac=CacheCade|OfLn=OffLine|Pdgd=Partially Degraded|Dgrd=Degraded
Optl=Optimal|RO=Read Only|RW=Read Write|HD=Hidden|TRANS=TransportReady|B=Blocked|
Consist=Consistent|R=Read Ahead Always|NR=No Read Ahead|WB=WriteBack|
AWB=Always WriteBack|WT=WriteThrough|C=Cached IO|D=Direct IO|sCC=Scheduled
Check Consistency


PDs for VD 4 :
============

----------------------------------------------------------------------------
EID:Slt DID State DG     Size Intf Med SED PI SeSz Model            Sp Type
----------------------------------------------------------------------------
134:5     8 Onln   2 2.181 TB SAS  HDD Y   N  4 KB ST2400MM0149     U  -
----------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild


VD4 Properties :
==============
Strip Size = 64 KB
Number of Blocks = 585691648
VD has Emulated PD = No
Span Depth = 1
Number of Drives Per Span = 1
Write Cache(initial setting) = WriteBack
Disk Cache Policy = Enabled
Encryption = None
Data Protection = None
Active Operations = None
Exposed to OS = Yes
OS Drive Name = /dev/sde
Creation Date = 12-03-2021
Creation Time = 12:04:32 AM
Emulation type = default
Cachebypass size = Cachebypass-64k
Cachebypass Mode = Cachebypass Intelligent
Is LD Ready for OS Requests = Yes
SCSI NAA Id = 6cc167e9730322c027dd6c90047c42a1
Unmap Enabled = No
`

// this has a 'F' for a DriveGroup (foreign)
// it is put together from old '/c0/dall show' output to
// look like '/c0 show all' would.
var foreignCxShow = `
CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 4.14.174stock-1
Controller = 0
Status = Success
Description = Show Drive Group Succeeded

Product Name = Cisco 12G Modular Raid Controller with 2GB cache (max 16 drives)
Serial Number = SK74277921
SAS Address =  5cc167e972c89c80
PCI Address = 00:3c:00:00
System Time = 04/01/2021 15:48:24
Mfg. Date = 10/22/17
Controller Time = 04/01/2021 15:48:24
FW Package Build = 51.10.0-3612
BIOS Version = 7.10.03.1_0x070A0402
FW Version = 5.100.00-3310
Driver Name = megaraid_sas
Driver Version = 07.714.04.00-rc1
Current Personality = RAID-Mode
Vendor Id = 0x1000
Device Id = 0x14
SubVendor Id = 0x1137
SubDevice Id = 0x20E
Host Interface = PCI-E
Device Interface = SAS-12G
Bus Number = 60
Device Number = 0
Function Number = 0
Drive Groups = 2

TOPOLOGY :
========

---------------------------------------------------------------------------
DG Arr Row EID:Slot DID Type  State BT     Size PDC  PI SED DS3  FSpace TR
---------------------------------------------------------------------------
 0 -   -   -        -   RAID0 Optl  N  2.181 TB dflt N  N   dflt N      N
 0 0   -   -        -   RAID0 Optl  N  2.181 TB dflt N  N   dflt N      N
 0 0   0   134:3    2   DRIVE Onln  N  2.181 TB dflt N  N   dflt -      N
 1 -   -   -        -   RAID0 Optl  N  2.181 TB dflt N  N   dflt N      N
 1 0   -   -        -   RAID0 Optl  N  2.181 TB dflt N  N   dflt N      N
 1 0   0   134:4    3   DRIVE Onln  N  2.181 TB dflt N  N   dflt -      N
 2 -   -   -        -   RAID0 Optl  N  2.181 TB dflt N  N   dflt N      N
 2 0   -   -        -   RAID0 Optl  N  2.181 TB dflt N  N   dflt N      N
 2 0   0   134:5    1   DRIVE Onln  N  2.181 TB dflt N  N   dflt -      N
---------------------------------------------------------------------------

DG=Disk Group Index|Arr=Array Index|Row=Row Index|EID=Enclosure Device ID
DID=Device ID|Type=Drive Type|Onln=Online|Rbld=Rebuild|Dgrd=Degraded
Pdgd=Partially degraded|Offln=Offline|BT=Background Task Active
PDC=PD Cache|PI=Protection Info|SED=Self Encrypting Drive|Frgn=Foreign
DS3=Dimmer Switch 3|dflt=Default|Msng=Missing|FSpace=Free Space Present
TR=Transport Ready

Virtual Drives = 2

VD LIST :
=======

----------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC     Size Name
----------------------------------------------------------------
0/0   RAID0 Optl  RW     Yes     NRWTD -   OFF 2.181 TB RAID0_3
1/1   RAID0 Optl  RW     Yes     NRWTD -   OFF 2.181 TB RAID0_4
2/2   RAID0 Optl  RW     Yes     NRWTD -   OFF 2.181 TB RAID0_5
----------------------------------------------------------------

EID=Enclosure Device ID| VD=Virtual Drive| DG=Drive Group|Rec=Recovery
Cac=CacheCade|OfLn=OffLine|Pdgd=Partially Degraded|Dgrd=Degraded
Optl=Optimal|RO=Read Only|RW=Read Write|HD=Hidden|TRANS=TransportReady|B=Blocked|
Consist=Consistent|R=Read Ahead Always|NR=No Read Ahead|WB=WriteBack|
AWB=Always WriteBack|WT=WriteThrough|C=Cached IO|D=Direct IO|sCC=Scheduled
Check Consistency

Physical Drives = 3

PD LIST :
=======

------------------------------------------------------------------------------
EID:Slt DID State DG        Size Intf Med SED PI SeSz Model            Sp Type
------------------------------------------------------------------------------
134:3     2 Onln   0    2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -
134:4     3 Onln   1    2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -
134:5     1 Onln   2    2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -
134:2     0 UGood  F  371.597 GB SAS  SSD N   N  512B KPM51VUG400G     U  -
134:6     4 UGood  F    2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -
------------------------------------------------------------------------------
------------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild

Enclosures = 1

Enclosure LIST :
==============

------------------------------------------------------------------------
EID State Slots PD PS Fans TSs Alms SIM Port# ProdID     VendorSpecific
------------------------------------------------------------------------
134 OK       16  3  0    0   0    0   0 -     virtualSES
------------------------------------------------------------------------

EID=Enclosure Device ID |PD=Physical drive count |PS=Power Supply count|
TSs=Temperature sensor count |Alms=Alarm count |SIM=SIM Count

Cachevault_Info :
===============

------------------------------------
Model  State   Temp Mode MfgDate
------------------------------------
CVPM05 Optimal 34C  -    2018/10/16
------------------------------------
`
