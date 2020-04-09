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

func TestParseCxDallShow(t *testing.T) {
	var exVlen, exPlen = 5, 5

	vds, pds, err := parseCxDallShow(storeCliCxDallShowAllBlob)

	if err != nil {
		t.Fatalf("Failed parsing with parseCxDallShow: %s", err)
	}

	if len(vds) != exVlen {
		t.Errorf("Expected %d vds found %d\n", exVlen, len(vds))
	}

	if len(pds) != exPlen {
		t.Errorf("Expected %d pds found %d\n", exPlen, len(pds))
	}
}

func TestParseCxDallShowNotConfigured(t *testing.T) {
	var exVlen, exPlen = 0, 4

	vds, pds, err := parseCxDallShow(dumpNotConfigured)

	if err != nil {
		t.Fatalf("Failed parsing with dumpNotConfigured: %s", err)
	}

	if len(vds) != exVlen {
		t.Errorf("Expected %d vds found %d\n", exVlen, len(vds))
	}

	if len(pds) != exPlen {
		t.Errorf("Expected %d pds found %d\n", exPlen, len(pds))
	}

	// nolint:gomnd
	expected := DriveSet{
		13: &Drive{ID: 13, EID: 62, Slot: 3, DriveGroup: -1,
			MediaType: HDD, Model: "MZ6ER400HAGL/003", State: "UBUnsp"},
		8: &Drive{ID: 8, EID: 62, Slot: 2, DriveGroup: -1,
			MediaType: HDD, Model: "ST1000NX0453", State: "JBOD"},
		10: &Drive{ID: 10, EID: 62, Slot: 1, DriveGroup: -1,
			MediaType: HDD, Model: "ST1000NX0453", State: "JBOD"},
		14: &Drive{ID: 14, EID: 62, Slot: 4, DriveGroup: -1,
			MediaType: SSD, Model: "KPM51VUG400G", State: "JBOD"},
	}
	for dID, d := range expected {
		if !d.IsEqual(*pds[dID]) {
			t.Errorf("%d not equal\n", dID)
		}
	}
}

func TestParseCxDallNoController(t *testing.T) {
	_, _, err := parseCxDallShow(dallNoController)

	if err != ErrNoController {
		t.Fatalf("storcli /c0/dall dallNoController expected ErrNoController found: %s",
			err)
	}
}

func TestForeignDriveDall(t *testing.T) {
	_, drives, err := parseCxDallShow(foreignDallBlob)

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

	propMap, err := parseVirtProperties(storeCliCxVallShowAllBlob)

	if err != nil {
		t.Fatalf("Failed parsing with parseCxDallShow: %s", err)
	}

	if len(propMap) != exLen {
		t.Errorf("Expected %d properties, but only found %d", exLen, len(propMap))
	}

	expected := map[string]string{
		"Strip Size":                   "64 KB",
		"Number of Blocks":             "585691648",
		"VD has Emulated PD":           "No",
		"Span Depth":                   "1",
		"Number of Drives Per Span":    "1",
		"Write Cache(initial setting)": "WriteThrough",
		"Disk Cache Policy":            "Disk's Default",
		"Encryption":                   "None",
		"Data Protection":              "None",
		"Active Operations":            "None",
		"Exposed to OS":                "Yes",
		"OS Drive Name":                "/dev/sdb",
		"Creation Date":                "05-08-2019",
		"Creation Time":                "11:57:05 PM",
		"Emulation type":               "default",
		"Cachebypass size":             "Cachebypass-64k",
		"Cachebypass Mode":             "Cachebypass Intelligent",
		"Is LD Ready for OS Requests":  "Yes",
		"SCSI NAA Id":                  "6cc167e97319bec024db7ed14575954b",
		"Unmap Enabled":                "No",
	}

	p1 := propMap[1]
	if len(p1) != len(expected) {
		t.Fatalf("Expected %d in propMap[1], got %d", len(expected), len(p1))
	}

	for k, v := range expected {
		if p1[k] != v {
			t.Errorf("propMap[1][%s]: expected '%s' found '%s'", k, v, p1[v])
		}
	}
}

func TestParseVirtPropertiesNone(t *testing.T) {
	propMap, err := parseVirtProperties(vallNotConfigured)

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

func TestNewController(t *testing.T) {
	var exVlen, exPlen, exDGlen = 5, 5, 5
	var cID = 0

	ctrl, err := newController(cID, storeCliCxDallShowAllBlob, storeCliCxVallShowAllBlob)
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
		{4, 4, true},
		{3, 3, false},
		{2, 2, false},
		{1, 1, false},
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
	//nolint: gomnd
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

// output of storecli /c0/dall show all
var storeCliCxDallShowAllBlob = `
CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 4.14.164stock-2
Controller = 0
Status = Success
Description = Show Drive Group Succeeded


TOPOLOGY :
========

-----------------------------------------------------------------------------
DG Arr Row EID:Slot DID Type  State BT       Size PDC  PI SED DS3  FSpace TR 
-----------------------------------------------------------------------------
 0 -   -   -        -   RAID0 Optl  N    2.181 TB dflt N  N   dflt N      N  
 0 0   -   -        -   RAID0 Optl  N    2.181 TB dflt N  N   dflt N      N  
 0 0   0   134:3    2   DRIVE Onln  N    2.181 TB dflt N  N   dflt -      N  
 1 -   -   -        -   RAID0 Optl  N    2.181 TB dflt N  N   dflt N      N  
 1 0   -   -        -   RAID0 Optl  N    2.181 TB dflt N  N   dflt N      N  
 1 0   0   134:4    4   DRIVE Onln  N    2.181 TB dflt N  N   dflt -      N  
 2 -   -   -        -   RAID0 Optl  N    2.181 TB dflt N  N   dflt N      N  
 2 0   -   -        -   RAID0 Optl  N    2.181 TB dflt N  N   dflt N      N  
 2 0   0   134:5    1   DRIVE Onln  N    2.181 TB dflt N  N   dflt -      N  
 3 -   -   -        -   RAID0 Optl  N    2.181 TB dflt N  N   dflt N      N  
 3 0   -   -        -   RAID0 Optl  N    2.181 TB dflt N  N   dflt N      N  
 3 0   0   134:6    3   DRIVE Onln  N    2.181 TB dflt N  N   dflt -      N  
 4 -   -   -        -   RAID0 Optl  N  371.597 GB dflt N  N   dflt N      N  
 4 0   -   -        -   RAID0 Optl  N  371.597 GB dflt N  N   dflt N      N  
 4 0   0   134:2    0   DRIVE Onln  N  371.597 GB dflt N  N   dflt -      N  
-----------------------------------------------------------------------------

DG=Disk Group Index|Arr=Array Index|Row=Row Index|EID=Enclosure Device ID
DID=Device ID|Type=Drive Type|Onln=Online|Rbld=Rebuild|Dgrd=Degraded
Pdgd=Partially degraded|Offln=Offline|BT=Background Task Active
PDC=PD Cache|PI=Protection Info|SED=Self Encrypting Drive|Frgn=Foreign
DS3=Dimmer Switch 3|dflt=Default|Msng=Missing|FSpace=Free Space Present
TR=Transport Ready


VD LIST :
=======

---------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC       Size Name 
---------------------------------------------------------------
0/0   RAID0 Optl  RW     Yes     NRWTD -   OFF   2.181 TB HDD1 
1/1   RAID0 Optl  RW     Yes     NRWTD -   OFF   2.181 TB HDD2 
2/2   RAID0 Optl  RW     Yes     NRWTD -   OFF   2.181 TB HDD3 
3/3   RAID0 Optl  RW     Yes     NRWTD -   OFF   2.181 TB HDD4 
4/4   RAID0 Optl  RW     Yes     NRWTD -   OFF 371.597 GB SSD1 
---------------------------------------------------------------

EID=Enclosure Device ID| VD=Virtual Drive| DG=Drive Group|Rec=Recovery
Cac=CacheCade|OfLn=OffLine|Pdgd=Partially Degraded|Dgrd=Degraded
Optl=Optimal|RO=Read Only|RW=Read Write|HD=Hidden|TRANS=TransportReady|B=Blocked|
Consist=Consistent|R=Read Ahead Always|NR=No Read Ahead|WB=WriteBack|
AWB=Always WriteBack|WT=WriteThrough|C=Cached IO|D=Direct IO|sCC=Scheduled
Check Consistency

Total VD Count = 5

DG Drive LIST :
=============

------------------------------------------------------------------------------
EID:Slt DID State DG       Size Intf Med SED PI SeSz Model            Sp Type 
------------------------------------------------------------------------------
134:3     2 Onln   0   2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -    
134:4     4 Onln   1   2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -    
134:5     1 Onln   2   2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -    
134:6     3 Onln   3   2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -    
134:2     0 Onln   4 371.597 GB SAS  SSD N   N  512B KPM51VUG400G     U  -    
------------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild

Total Drive Count = 5
`

// this has a 'F' for a DriveGroup (foreign)
var foreignDallBlob = `
CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 4.14.174stock-1
Controller = 0
Status = Success
Description = Show Drive Group Succeeded


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

Total VD Count = 3

DG Drive LIST :
=============

----------------------------------------------------------------------------
EID:Slt DID State DG     Size Intf Med SED PI SeSz Model            Sp Type 
----------------------------------------------------------------------------
134:3     2 Onln   0 2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -    
134:4     3 Onln   1 2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -    
134:5     1 Onln   2 2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -    
----------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild

Total Drive Count = 3

UN-CONFIGURED DRIVE LIST :
========================

------------------------------------------------------------------------------
EID:Slt DID State DG       Size Intf Med SED PI SeSz Model            Sp Type 
------------------------------------------------------------------------------
134:2     0 UGood F  371.597 GB SAS  SSD N   N  512B KPM51VUG400G     U  -    
134:6     4 UGood F    2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -    
------------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild

Unconfigured Drive Count = 2
	`

// output of storcli /c0/vall show all
var storeCliCxVallShowAllBlob = `
CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 4.14.164stock-2
Controller = 0
Status = Success
Description = None


/c0/v0 :
======

-------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC     Size Name 
-------------------------------------------------------------
0/0   RAID0 Optl  RW     Yes     NRWTD -   OFF 2.181 TB HDD1 
-------------------------------------------------------------

EID=Enclosure Device ID| VD=Virtual Drive| DG=Drive Group|Rec=Recovery
Cac=CacheCade|OfLn=OffLine|Pdgd=Partially Degraded|Dgrd=Degraded
Optl=Optimal|RO=Read Only|RW=Read Write|HD=Hidden|TRANS=TransportReady|B=Blocked|
Consist=Consistent|R=Read Ahead Always|NR=No Read Ahead|WB=WriteBack|
AWB=Always WriteBack|WT=WriteThrough|C=Cached IO|D=Direct IO|sCC=Scheduled
Check Consistency


PDs for VD 0 :
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


VD0 Properties :
==============
Strip Size = 64 KB
Number of Blocks = 585691648
VD has Emulated PD = No
Span Depth = 1
Number of Drives Per Span = 1
Write Cache(initial setting) = WriteThrough
Disk Cache Policy = Disk's Default
Encryption = None
Data Protection = None
Active Operations = None
Exposed to OS = Yes
OS Drive Name = /dev/sda
Creation Date = 05-08-2019
Creation Time = 11:56:46 PM
Emulation type = default
Cachebypass size = Cachebypass-64k
Cachebypass Mode = Cachebypass Intelligent
Is LD Ready for OS Requests = Yes
SCSI NAA Id = 6cc167e97319bec024db7ebe29ff7906
Unmap Enabled = No


/c0/v1 :
======

-------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC     Size Name 
-------------------------------------------------------------
1/1   RAID0 Optl  RW     Yes     NRWTD -   OFF 2.181 TB HDD2 
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
134:4     4 Onln   1 2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -    
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
Write Cache(initial setting) = WriteThrough
Disk Cache Policy = Disk's Default
Encryption = None
Data Protection = None
Active Operations = None
Exposed to OS = Yes
OS Drive Name = /dev/sdb
Creation Date = 05-08-2019
Creation Time = 11:57:05 PM
Emulation type = default
Cachebypass size = Cachebypass-64k
Cachebypass Mode = Cachebypass Intelligent
Is LD Ready for OS Requests = Yes
SCSI NAA Id = 6cc167e97319bec024db7ed14575954b
Unmap Enabled = No


/c0/v2 :
======

-------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC     Size Name 
-------------------------------------------------------------
2/2   RAID0 Optl  RW     Yes     NRWTD -   OFF 2.181 TB HDD3 
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
134:5     1 Onln   2 2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -    
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
Write Cache(initial setting) = WriteThrough
Disk Cache Policy = Disk's Default
Encryption = None
Data Protection = None
Active Operations = None
Exposed to OS = Yes
OS Drive Name = /dev/sdc
Creation Date = 05-08-2019
Creation Time = 11:57:24 PM
Emulation type = default
Cachebypass size = Cachebypass-64k
Cachebypass Mode = Cachebypass Intelligent
Is LD Ready for OS Requests = Yes
SCSI NAA Id = 6cc167e97319bec024db7ee462d8e595
Unmap Enabled = No


/c0/v3 :
======

-------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC     Size Name 
-------------------------------------------------------------
3/3   RAID0 Optl  RW     Yes     NRWTD -   OFF 2.181 TB HDD4 
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
134:6     3 Onln   3 2.181 TB SAS  HDD N   N  4 KB ST2400MM0129     U  -    
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
Write Cache(initial setting) = WriteThrough
Disk Cache Policy = Disk's Default
Encryption = None
Data Protection = None
Active Operations = None
Exposed to OS = Yes
OS Drive Name = /dev/sdd
Creation Date = 05-08-2019
Creation Time = 11:58:09 PM
Emulation type = default
Cachebypass size = Cachebypass-64k
Cachebypass Mode = Cachebypass Intelligent
Is LD Ready for OS Requests = Yes
SCSI NAA Id = 6cc167e97319bec024db7f11a5648b7c
Unmap Enabled = No


/c0/v4 :
======

---------------------------------------------------------------
DG/VD TYPE  State Access Consist Cache Cac sCC       Size Name 
---------------------------------------------------------------
4/4   RAID0 Optl  RW     Yes     NRWTD -   OFF 371.597 GB SSD1 
---------------------------------------------------------------

EID=Enclosure Device ID| VD=Virtual Drive| DG=Drive Group|Rec=Recovery
Cac=CacheCade|OfLn=OffLine|Pdgd=Partially Degraded|Dgrd=Degraded
Optl=Optimal|RO=Read Only|RW=Read Write|HD=Hidden|TRANS=TransportReady|B=Blocked|
Consist=Consistent|R=Read Ahead Always|NR=No Read Ahead|WB=WriteBack|
AWB=Always WriteBack|WT=WriteThrough|C=Cached IO|D=Direct IO|sCC=Scheduled
Check Consistency


PDs for VD 4 :
============

------------------------------------------------------------------------------
EID:Slt DID State DG       Size Intf Med SED PI SeSz Model            Sp Type 
------------------------------------------------------------------------------
134:2     0 Onln   4 371.597 GB SAS  SSD N   N  512B KPM51VUG400G     U  -    
------------------------------------------------------------------------------

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
OS Drive Name = /dev/sde
Creation Date = 05-08-2019
Creation Time = 11:58:27 PM
Emulation type = default
Cachebypass size = Cachebypass-64k
Cachebypass Mode = Cachebypass Intelligent
Is LD Ready for OS Requests = Yes
SCSI NAA Id = 6cc167e97319bec024db7f23c023279c
Unmap Enabled = No
`

// sudo ./storcli /c0/dall show all
// atom-lab-4
var dumpNotConfigured = `
CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 5.0.0-37-generic
Controller = 0
Status = Success
Description = Show Drive Group Succeeded

Drive Group not found.

UN-CONFIGURED DRIVE LIST :
========================

-------------------------------------------------------------------------------
EID:Slt DID State  DG       Size Intf Med SED PI SeSz Model            Sp Type 
-------------------------------------------------------------------------------
62:3     13 UBUnsp -        0 KB SAS  HDD N   N  512B MZ6ER400HAGL/003 U  -    
62:2      8 JBOD   -  931.512 GB SAS  HDD N   N  512B ST1000NX0453     U  -    
62:1     10 JBOD   -  931.512 GB SAS  HDD N   N  512B ST1000NX0453     U  -    
62:4     14 JBOD   -  372.611 GB SAS  SSD N   N  512B KPM51VUG400G     U  -    
-------------------------------------------------------------------------------

EID=Enclosure Device ID|Slt=Slot No.|DID=Device ID|DG=DriveGroup
DHS=Dedicated Hot Spare|UGood=Unconfigured Good|GHS=Global Hotspare
UBad=Unconfigured Bad|Onln=Online|Offln=Offline|Intf=Interface
Med=Media Type|SED=Self Encryptive Drive|PI=Protection Info
SeSz=Sector Size|Sp=Spun|U=Up|D=Down|T=Transition|F=Foreign
UGUnsp=UGood Unsupported|UGShld=UnConfigured shielded|HSPShld=Hotspare shielded
CFShld=Configured shielded|Cpybck=CopyBack|CBShld=Copyback Shielded
UBUnsp=UBad Unsupported|Rbld=Rebuild

Unconfigured Drive Count = 4
`

// sudo ./storcli /c0/vall show all
var vallNotConfigured = `
CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 5.0.0-37-generic
Controller = 0
Status = Success
Description = No VD's have been configured.
`

// storcli /c0/vall show all
var vallNoController = `
CLI Version = 007.1211.0000.0000 Nov 07, 2019
Operating system = Linux 5.0.0-37-generic
Controller = 0
Status = Failure
Description = Controller 0 not found
`

// storcli /c0/dall show all
var dallNoController = vallNoController
