package smartpqi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"machinerun.io/disko"
)

var arcConfGetConfig = `
Controllers found: 1
----------------------------------------------------------------------
Controller information
----------------------------------------------------------------------
   Controller Status                          : Optimal
   Controller Mode                            : Mixed
   Channel description                        : SCSI
   Controller Model                           : Cisco 24G TriMode M1 RAID 4GB FBWC 16D UCSC-RAID-M1L16
   Vendor ID                                  : 0x9005
   Device ID                                  : 0x028F
   Subsystem Vendor ID                        : 0x1137
   Subsystem Device ID                        : 0x02F9
   Controller Serial Number                   : 3137F30003A
   Controller World Wide Name                 : 50000D1E01787B00
   Physical Slot                              : 16
   Temperature                                : 33 C/ 91 F (Normal)
   Negotiated PCIe Data Rate                  : PCIe 4.0 x16(31504 MB/s)
   PCI Address (Domain:Bus:Device:Function)   : 0:4b:0:0
   Number of Ports                            : 2
   Internal Port Count                        : 2
   External Port Count                        : 0
   Defunct disk drive count                   : 0
   NCQ status                                 : Enabled
   Queue Depth                                : Automatic
   Monitor and Performance Delay              : 60 minutes
   Elevator Sort                              : Enabled
   Degraded Mode Performance Optimization     : Disabled
   Latency                                    : Disabled
   Post Prompt Timeout                        : 0 seconds
   Boot Controller                            : False
   Primary Boot Volume                        : Logical device 2(Logical Drive 3)
   Secondary Boot Volume                      : Logical device 2(Logical Drive 3)
   Driver Name                                : smartpqi
   Driver Supports SSD I/O Bypass             : Yes
   NVMe Supported                             : Yes
   NVMe Configuration Supported               : Yes
   Manufacturing Part Number                  : Not Applicable
   Manufacturing Spare Part Number            : Not Applicable
   Manufacturing Wellness Log                 : Not Applicable
   Manufacturing SKU Number                   : Not Applicable
   Manufacturing Model                        : Not Applicable
   NVRAM Checksum Status                      : Passed
   Sanitize Lock Setting                      : None
   Expander Minimum Scan Duration             : 0 seconds
   Expander Scan Time-out                     : 350 seconds
   Active PCIe Maximum Read Request Size      : 2048 bytes
   Pending PCIe Maximum Read Request Size     : Not Applicable
   PCIe Maximum Payload Size                  : 512 bytes
   Persistent Event Log Policy                : Oldest
   UEFI Health Reporting Mode                 : Enabled
   Reboot Required Reasons                    : Not Available
   -------------------------------------------------------------------
   Power Settings
   -------------------------------------------------------------------
   Power Consumption                          : 16891 milliWatts
   Current Power Mode                         : Maximum Performance
   Pending Power Mode                         : Not Applicable
   Survival Mode                              : Enabled
   -------------------------------------------------------------------
   Cache Properties
   -------------------------------------------------------------------
   Cache Status                               : Temporarily Disabled
   Cache State                                : Disabled Flashlight Capacitor Charge Is Low
   Cache State Details                        : Temporarily Disabled(Reason: Charge level of capacitor attached to cache module is low)
   Cache Serial Number                        : Not Applicable
   Cache memory                               : 3644 MB
   Read Cache Percentage                      : 10 percent
   Write Cache Percentage                     : 90 percent
   No-Battery Write Cache                     : Disabled
   Wait for Cache Room                        : Disabled
   Write Cache Bypass Threshold Size          : 1040 KB
   -------------------------------------------------------------------
   Green Backup Information
   -------------------------------------------------------------------
   Backup Power Status                        : Not Fully Charged
   Battery/Capacitor Pack Count               : 1
   Hardware Error                             : No Error
   Power Type                                 : Supercap
   Current Temperature                        : 19 deg C
   Maximum Temperature                        : 60 deg C
   Threshold Temperature                      : 50 deg C
   Voltage                                    : 3885 milliVolts
   Maximum Voltage                            : 5208 milliVolts
   Current                                    : 0 milliAmps
   Health Status                              : 100 percent
   Relative Charge                            : 0 percent
   -------------------------------------------------------------------
   Physical Drive Write Cache Policy Information
   -------------------------------------------------------------------
   Configured Drives                          : Default
   Unconfigured Drives                        : Default
   HBA Drives                                 : Default
   -------------------------------------------------------------------
   maxCache Properties
   -------------------------------------------------------------------
   maxCache Version                           : 4
   maxCache RAID5 WriteBack Enabled           : Enabled
   -------------------------------------------------------------------
   RAID Properties
   -------------------------------------------------------------------
   Logical devices/Failed/Degraded            : 3/0/0
   Spare Activation Mode                      : Failure
   Background consistency check               : Idle
   Consistency Check Delay                    : 3 seconds
   Parallel Consistency Check Supported       : Enabled
   Parallel Consistency Check Count           : 1
   Inconsistency Repair Policy                : Disabled
   Consistency Check Inconsistency Notify     : Disabled
   Rebuild Priority                           : High
   Expand Priority                            : Medium
   -------------------------------------------------------------------
   Controller Version Information
   -------------------------------------------------------------------
   Firmware                                   : 03.01.24.082-p
   Driver                                     : Linux 2.1.12-055
   Hardware Revision                          : B
   Hardware Minor Revision                    : 1
   SEEPROM Version                            : 0
   CPLD Revision                              : 1
   -------------------------------------------------------------------
   SED Encryption Properties
   -------------------------------------------------------------------
   SED Encryption                             : Off
   Key Mode                                   : None
   SED Encryption Status                      : Not Applicable
   SED Operation In Progress                  : Not Applicable
   Master Key Identifier                      : Not Applicable
   SED Controller Password Status             : Not Configured
   Countdown Timer                            : Not Applicable
   Attempts Left                              : Not Applicable
   -------------------------------------------------------------------
   Controller maxCrypto Information
   -------------------------------------------------------------------
   maxCrypto Supported                                     : Yes
   maxCrypto Status                                        : Disabled
   License Installed                                       : Not Installed
   Express Local maxCrypto                                 : Not Configured
   Controller Password                                     : Not Configured
   Crypto Officer Password                                 : Not Configured
   User Password                                           : Not Configured
   Allow New Plaintext Logical device(s)                   : Not Applicable
   Key Management Mode                                     : Not Configured
   Master Key                                              : Not Configured
   Remote Mode Master Key Mismatch                         : No
   Master Key Reset in Progress                            : No
   Local Key Cache                                         : Not Configured
   FW Locked for Update                                    : No
   Controller Locked                                       : No
   Has Suspended Controller Password                       : No
   Logical Drive(s) Locked For Missing Controller Password : No
   Password Recovery Parameters Set                        : No
   SSD I/O Bypass Mixing                                   : Supported
   maxCache Mixing                                         : Supported
   Skip Controller Password                                : Disabled
   Controller Password Unlock Attempts Remaining           : 0
   Crypto Account Password Unlock Attempts Remaining       : 0
   User Account Password Unlock Attempts Remaining         : 0
   Number of maxCrypto Physical devices                    : 0
   Number of maxCrypto Data Logical devices                : 0
   Number of maxCrypto Foreign Logical devices without key : 0
   Number of maxCrypto Logical devices with maxCrypto off  : 0
   -------------------------------------------------------------------
   Temperature Sensors Information
   -------------------------------------------------------------------
   Sensor ID                                  : 0
   Current Value                              : 21 deg C
   Max Value Since Powered On                 : 21 deg C
   Location                                   : Inlet Ambient

   Sensor ID                                  : 1
   Current Value                              : 33 deg C
   Max Value Since Powered On                 : 33 deg C
   Location                                   : ASIC

   Sensor ID                                  : 2
   Current Value                              : 27 deg C
   Max Value Since Powered On                 : 27 deg C
   Location                                   : Top

   Sensor ID                                  : 3
   Current Value                              : 26 deg C
   Max Value Since Powered On                 : 26 deg C
   Location                                   : Bottom

   -------------------------------------------------------------------
   Out Of Band Interface Settings
   -------------------------------------------------------------------
   OOB Interface                              : MCTP
   Pending OOB Interface                      : MCTP
   I2C Address                                : 0xDE
   Pending I2C Address                        : 0xDE
   -------------------------------------------------------------------
   PBSI
   -------------------------------------------------------------------
   I2C Clock Speed                            : Not Applicable
   I2C Clock Stretching                       : Not Applicable
   Pending I2C Clock Speed                    : Not Applicable
   Pending I2C Clock Stretching               : Not Applicable
   -------------------------------------------------------------------
   MCTP
   -------------------------------------------------------------------
   SMBus Device Type                          : Fixed
   SMBus Channel                              : Disabled
   Static EIDs Use On Initialization          : Disabled
   VDM Notification                           : Enabled
   Pending SMBus Device Type                  : Fixed
   Pending SMBus Channel                      : Disabled
   Pending Static EIDs Use On Initialization  : Disabled
   Pending VDM Notification                   : Enabled
   -------------------------------------------------------------------
   Controller SPDM Setting Information
   -------------------------------------------------------------------
   Version                                    : 0x10
   Endpoint ID                                : 0x09
   Crypto Timeout Exponent                    : 20
   Authority Key ID                           : DA:20:0B:C9:6B:E0:81:DC:C5:D8:96:00:CD:0E:3C:F7:59:DB:3C:A0
   Slot 0                                     : Valid and Sealed
   Slot 1                                     : Available
   Slot 2                                     : Available
   Slot 3                                     : Available
   Slot 4                                     : Available
   Slot 5                                     : Available
   Slot 6                                     : Available
   Slot 7                                     : Available
   -------------------------------------------------------------------
   Capabilities
   -------------------------------------------------------------------
      Digests And Certificate                 : Supported
      Challenge                               : Supported
      Measurements With Signature             : Supported
   -------------------------------------------------------------------
   Connector information
   -------------------------------------------------------------------
   Connector #0
      Connector name                          : CN2
      Connection Number                       : 0
      Functional Mode                         : Mixed
      Connector Location                      : Internal
      SAS Address                             : 50000D1E01787B00
      Current Discovery Protocol              : UBM
      Pending Discovery Protocol              : Not Applicable


   Connector #1
      Connector name                          : CN3
      Connection Number                       : 1
      Functional Mode                         : Mixed
      Connector Location                      : Internal
      SAS Address                             : 50000D1E01787B0C
      Current Discovery Protocol              : AutoDetect
      Pending Discovery Protocol              : Not Applicable



----------------------------------------------------------------------
Array Information
----------------------------------------------------------------------
Array Number 0
   Name                                       : A
   Status                                     : Ok
   Interface                                  : SAS
   Total Size                                 : 572325 MB
   Unused Size                                : 0 MB
   Block Size                                 : 512 Bytes
   Array Utilization                          : 100.00% Used, 0.00% Unused
   Type                                       : Data
   Transformation Status                      : Not Applicable
   Spare Rebuild Mode                         : Dedicated
   SSD I/O Bypass                             : Not Applicable
   SED Encryption                             : Disabled
--------------------------------------------------------
   Array Logical Device Information
--------------------------------------------------------
   Logical ID                                 : Status (RAID, Interface, SizeMB) Name
--------------------------------------------------------
   Logical 0                                  : Optimal (0, Data, 572293 MB) Logical Drive 1
--------------------------------------------------------
   Array Physical Device Information
--------------------------------------------------------
   Device ID                                  : Availability (SizeMB, Protocol, Type, Connector ID, Location) Serial Number
--------------------------------------------------------
   Device 0                                   : Present (572325MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:1) 63M0A0BYFJPF


Array Number 1
   Name                                       : B
   Status                                     : Ok
   Interface                                  : SAS
   Total Size                                 : 1144641 MB
   Unused Size                                : 0 MB
   Block Size                                 : 512 Bytes
   Array Utilization                          : 100.00% Used, 0.00% Unused
   Type                                       : Data
   Transformation Status                      : Not Applicable
   Spare Rebuild Mode                         : Dedicated
   SSD I/O Bypass                             : Not Applicable
   SED Encryption                             : Disabled
--------------------------------------------------------
   Array Logical Device Information
--------------------------------------------------------
   Logical ID                                 : Status (RAID, Interface, SizeMB) Name
--------------------------------------------------------
   Logical 1                                  : Optimal (0, Data, 1144609 MB) Logical Drive 2
--------------------------------------------------------
   Array Physical Device Information
--------------------------------------------------------
   Device ID                                  : Availability (SizeMB, Protocol, Type, Connector ID, Location) Serial Number
--------------------------------------------------------
   Device 1                                   : Present (1144641MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:2) 59M0A06CFJRG


Array Number 2
   Name                                       : C
   Status                                     : Ok
   Interface                                  : SAS
   Total Size                                 : 1144641 MB
   Unused Size                                : 0 MB
   Block Size                                 : 512 Bytes
   Array Utilization                          : 100.00% Used, 0.00% Unused
   Type                                       : Data
   Transformation Status                      : Not Applicable
   Spare Rebuild Mode                         : Dedicated
   SSD I/O Bypass                             : Not Applicable
   SED Encryption                             : Disabled
--------------------------------------------------------
   Array Logical Device Information
--------------------------------------------------------
   Logical ID                                 : Status (RAID, Interface, SizeMB) Name
--------------------------------------------------------
   Logical 2                                  : Optimal (0, Data, 1144609 MB) Logical Drive 3
--------------------------------------------------------
   Array Physical Device Information
--------------------------------------------------------
   Device ID                                  : Availability (SizeMB, Protocol, Type, Connector ID, Location) Serial Number
--------------------------------------------------------
   Device 2                                   : Present (1144641MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:3) WFK076DT0000E821CET2



--------------------------------------------------------
Logical device information
--------------------------------------------------------
Logical Device number 0
   Logical Device name                        : Logical Drive 1
   Disk Name                                  : /dev/sdd (Disk0) (Bus: 1, Target: 0, Lun: 0)
   Block Size of member drives                : 512 Bytes
   Array                                      : 0
   RAID level                                 : 0
   Status of Logical Device                   : Optimal
   Size                                       : 572293 MB
   Stripe-unit size                           : 128 KB
   Full Stripe Size                           : 128 KB
   Interface Type                             : Serial Attached SCSI
   Device Type                                : Data
   Boot Type                                  : None
   Heads                                      : 255
   Sectors Per Track                          : 32
   Cylinders                                  : 65535
   Caching                                    : Enabled
   Mount Points                               : Not Mounted
   LD Acceleration Method                     : Controller Cache
   SED Encryption                             : Disabled
   Volume Unique Identifier                   : 600508B1001CB9F0FE988FC40CD17395
--------------------------------------------------------
   Consistency Check Information
--------------------------------------------------------
   Consistency Check Status                   : Not Applicable
   Last Consistency Check Completion Time     : Not Applicable
   Last Consistency Check Duration            : Not Applicable
--------------------------------------------------------
   Array Physical Device Information
--------------------------------------------------------
   Device ID                                  : Availability (SizeMB, Protocol, Type, Connector ID, Location) Serial Number
--------------------------------------------------------
   Device 0                                   : Present (572325MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:1) 63M0A0BYFJPF


Logical Device number 1
   Logical Device name                        : Logical Drive 2
   Disk Name                                  : /dev/sde (Disk0) (Bus: 1, Target: 0, Lun: 1)
   Block Size of member drives                : 512 Bytes
   Array                                      : 1
   RAID level                                 : 0
   Status of Logical Device                   : Optimal
   Size                                       : 1144609 MB
   Stripe-unit size                           : 128 KB
   Full Stripe Size                           : 128 KB
   Interface Type                             : Serial Attached SCSI
   Device Type                                : Data
   Boot Type                                  : None
   Heads                                      : 255
   Sectors Per Track                          : 32
   Cylinders                                  : 65535
   Caching                                    : Enabled
   Mount Points                               : /boot/efi 1075 MB  Partition Number 1
   LD Acceleration Method                     : Controller Cache
   SED Encryption                             : Disabled
   Volume Unique Identifier                   : 600508B1001CFE283CA96614826A7F85
--------------------------------------------------------
   Consistency Check Information
--------------------------------------------------------
   Consistency Check Status                   : Not Applicable
   Last Consistency Check Completion Time     : Not Applicable
   Last Consistency Check Duration            : Not Applicable
--------------------------------------------------------
   Array Physical Device Information
--------------------------------------------------------
   Device ID                                  : Availability (SizeMB, Protocol, Type, Connector ID, Location) Serial Number
--------------------------------------------------------
   Device 1                                   : Present (1144641MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:2) 59M0A06CFJRG


Logical Device number 2
   Logical Device name                        : Logical Drive 3
   Disk Name                                  : /dev/sdc (Disk0) (Bus: 1, Target: 0, Lun: 2)
   Block Size of member drives                : 512 Bytes
   Array                                      : 2
   RAID level                                 : 0
   Status of Logical Device                   : Optimal
   Size                                       : 1144609 MB
   Stripe-unit size                           : 128 KB
   Full Stripe Size                           : 128 KB
   Interface Type                             : Serial Attached SCSI
   Device Type                                : Data
   Boot Type                                  : Primary and Secondary
   Heads                                      : 255
   Sectors Per Track                          : 32
   Cylinders                                  : 65535
   Caching                                    : Enabled
   Mount Points                               : Not Mounted
   LD Acceleration Method                     : Controller Cache
   SED Encryption                             : Disabled
   Volume Unique Identifier                   : 600508B1001C6DB81E7099960E5B5796
--------------------------------------------------------
   Consistency Check Information
--------------------------------------------------------
   Consistency Check Status                   : Not Applicable
   Last Consistency Check Completion Time     : Not Applicable
   Last Consistency Check Duration            : Not Applicable
--------------------------------------------------------
   Array Physical Device Information
--------------------------------------------------------
   Device ID                                  : Availability (SizeMB, Protocol, Type, Connector ID, Location) Serial Number
--------------------------------------------------------
   Device 2                                   : Present (1144641MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:3) WFK076DT0000E821CET2



----------------------------------------------------------------------
Physical Device information
----------------------------------------------------------------------
   Channel #0:
      Device #0
         Device is a Hard drive
         State                                : Online
         Drive has stale RIS data             : False
         Block Size                           : 512 Bytes
         Physical Block Size                  : 512 Bytes
         Transfer Speed                       : SAS 12.0 Gb/s
         Reported Channel,Device(T:L)         : 0,0(0:0)
         Reported Location                    : Backplane 0, Slot 1(Connector 0:CN2)
         Array                                : 0
         Vendor                               : TOSHIBA
         Model                                : AL15SEB060N
         Firmware                             : 5703
         Serial number                        : 63M0A0BYFJPF
         World-wide name                      : 5000039C983AD88A
         Reserved Size                        : 32768 KB
         Used Size                            : 572293 MB
         Unused Size                          : 0 MB
         Total Size                           : 572325 MB
         Write Cache                          : Disabled (write-through)
         S.M.A.R.T.                           : No
         S.M.A.R.T. warnings                  : 0
         SSD                                  : No
         Boot Type                            : None
         Rotational Speed                     : 10500 RPM
         Current Temperature                  : 20 deg C
         Maximum Temperature                  : 20 deg C
         Threshold Temperature                : 65 deg C
         PHY Count                            : 2
         Drive Configuration Type             : Data
         Drive Exposed to OS                  : False
         Sanitize Erase Support               : True
         Sanitize Lock Freeze Support         : False
         Sanitize Lock Anti-Freeze Support    : False
         Sanitize Lock Setting                : None
         Drive Unique ID                      : 5000039C983AD889
         Drive SKU Number                     : Not Applicable
         Drive Part Number                    : Not Applicable
         Last Failure Reason                  : No Failure
      ----------------------------------------------------------------
      Device Phy Information
      ----------------------------------------------------------------
         Phy #0
            Negotiated Physical Link Rate     : 12 Gbps
            Negotiated Logical Link Rate      : 12 Gbps
            Maximum Link Rate                 : 12 Gbps
         Phy #1
            Negotiated Physical Link Rate     : unknown
            Negotiated Logical Link Rate      : unknown
            Maximum Link Rate                 : 12 Gbps

      ----------------------------------------------------------------
      Device Error Counters
      ----------------------------------------------------------------
         Aborted Commands                     : 0
         Bad Target Errors                    : 0
         Ecc Recovered Read Errors            : 0
         Failed Read Recovers                 : 0
         Failed Write Recovers                : 0
         Format Errors                        : 0
         Hardware Errors                      : 0
         Hard Read Errors                     : 0
         Hard Write Errors                    : 0
         Hot Plug Count                       : 0
         Media Failures                       : 0
         Not Ready Errors                     : 0
         Other Time Out Errors                : 0
         Predictive Failures                  : 0
         Retry Recovered Read Errors          : 0
         Retry Recovered Write Errors         : 0
         Scsi Bus Faults                      : 6
         Sectors Reads                        : 0
         Sectors Written                      : 0
         Service Hours                        : 44

      Device #1
         Device is a Hard drive
         State                                : Online
         Drive has stale RIS data             : False
         Block Size                           : 512 Bytes
         Physical Block Size                  : 512 Bytes
         Transfer Speed                       : SAS 12.0 Gb/s
         Reported Channel,Device(T:L)         : 0,1(1:0)
         Reported Location                    : Backplane 0, Slot 2(Connector 0:CN2)
         Array                                : 1
         Vendor                               : TOSHIBA
         Model                                : AL15SEB120N
         Firmware                             : 5701
         Serial number                        : 59M0A06CFJRG
         World-wide name                      : 50000399686BA672
         Reserved Size                        : 32768 KB
         Used Size                            : 1144609 MB
         Unused Size                          : 0 MB
         Total Size                           : 1144641 MB
         Write Cache                          : Disabled (write-through)
         S.M.A.R.T.                           : No
         S.M.A.R.T. warnings                  : 0
         SSD                                  : No
         Boot Type                            : None
         Rotational Speed                     : 10500 RPM
         Current Temperature                  : 22 deg C
         Maximum Temperature                  : 22 deg C
         Threshold Temperature                : 65 deg C
         PHY Count                            : 2
         Drive Configuration Type             : Data
         Drive Exposed to OS                  : False
         Sanitize Erase Support               : False
         Drive Unique ID                      : 50000399686BA671
         Drive SKU Number                     : Not Applicable
         Drive Part Number                    : Not Applicable
         Last Failure Reason                  : No Failure
      ----------------------------------------------------------------
      Device Phy Information
      ----------------------------------------------------------------
         Phy #0
            Negotiated Physical Link Rate     : 12 Gbps
            Negotiated Logical Link Rate      : 12 Gbps
            Maximum Link Rate                 : 12 Gbps
         Phy #1
            Negotiated Physical Link Rate     : unknown
            Negotiated Logical Link Rate      : unknown
            Maximum Link Rate                 : 12 Gbps

      ----------------------------------------------------------------
      Device Error Counters
      ----------------------------------------------------------------
         Aborted Commands                     : 0
         Bad Target Errors                    : 0
         Ecc Recovered Read Errors            : 0
         Failed Read Recovers                 : 0
         Failed Write Recovers                : 0
         Format Errors                        : 0
         Hardware Errors                      : 0
         Hard Read Errors                     : 0
         Hard Write Errors                    : 0
         Hot Plug Count                       : 0
         Media Failures                       : 0
         Not Ready Errors                     : 0
         Other Time Out Errors                : 0
         Predictive Failures                  : 0
         Retry Recovered Read Errors          : 0
         Retry Recovered Write Errors         : 0
         Scsi Bus Faults                      : 0
         Sectors Reads                        : 0
         Sectors Written                      : 0
         Service Hours                        : 44

      Device #2
         Device is a Hard drive
         State                                : Online
         Drive has stale RIS data             : False
         Block Size                           : 512 Bytes
         Physical Block Size                  : 512 Bytes
         Transfer Speed                       : SAS 12.0 Gb/s
         Reported Channel,Device(T:L)         : 0,2(2:0)
         Reported Location                    : Backplane 0, Slot 3(Connector 0:CN2)
         Array                                : 2
         Vendor                               : SEAGATE
         Model                                : ST1200MM0009
         Firmware                             : CN03
         Serial number                        : WFK076DT0000E821CET2
         World-wide name                      : 5000C500A137C8C9
         Reserved Size                        : 32768 KB
         Used Size                            : 1144609 MB
         Unused Size                          : 0 MB
         Total Size                           : 1144641 MB
         Write Cache                          : Disabled (write-through)
         S.M.A.R.T.                           : No
         S.M.A.R.T. warnings                  : 0
         SSD                                  : No
         Boot Type                            : None
         Rotational Speed                     : 10500 RPM
         Current Temperature                  : 22 deg C
         Maximum Temperature                  : 22 deg C
         Threshold Temperature                : 60 deg C
         PHY Count                            : 2
         Drive Configuration Type             : Data
         Drive Exposed to OS                  : False
         Sanitize Erase Support               : True
         Sanitize Lock Freeze Support         : False
         Sanitize Lock Anti-Freeze Support    : False
         Sanitize Lock Setting                : None
         Drive Unique ID                      : 5000C500A137C8CB
         Drive SKU Number                     : Not Applicable
         Drive Part Number                    : Not Applicable
         Last Failure Reason                  : No Failure
      ----------------------------------------------------------------
      Device Phy Information
      ----------------------------------------------------------------
         Phy #0
            Negotiated Physical Link Rate     : 12 Gbps
            Negotiated Logical Link Rate      : 12 Gbps
            Maximum Link Rate                 : 12 Gbps
         Phy #1
            Negotiated Physical Link Rate     : unknown
            Negotiated Logical Link Rate      : unknown
            Maximum Link Rate                 : 12 Gbps

      ----------------------------------------------------------------
      Device Error Counters
      ----------------------------------------------------------------
         Aborted Commands                     : 0
         Bad Target Errors                    : 0
         Ecc Recovered Read Errors            : 0
         Failed Read Recovers                 : 0
         Failed Write Recovers                : 0
         Format Errors                        : 0
         Hardware Errors                      : 0
         Hard Read Errors                     : 0
         Hard Write Errors                    : 0
         Hot Plug Count                       : 0
         Media Failures                       : 0
         Not Ready Errors                     : 0
         Other Time Out Errors                : 0
         Predictive Failures                  : 0
         Retry Recovered Read Errors          : 0
         Retry Recovered Write Errors         : 0
         Scsi Bus Faults                      : 0
         Sectors Reads                        : 0
         Sectors Written                      : 0
         Service Hours                        : 44

   Channel #2:
      Device #0
         Device is an Enclosure Services Device
         Reported Channel,Device(T:L)         : 2,0(0:0)
         Enclosure ID                         : 0
         Enclosure Logical Identifier         : 50000D1E01787B10
         Type                                 : SES2
         Vendor                               : Cisco
         Model                                : Virtual SGPIO
         Firmware                             : 0124
         Status of Enclosure Services Device
            Speaker status                    : Not Available

     Backplane:
      Device #0
         Device is an UBM Controller
         Backplane ID                         : 0
         UBM Controller ID                    : 0
         Type                                 : UBM
         Firmware                             : 0.2
         Device Code                          : 3792052480
         PCI Vendor ID                        : 0x1000


----------------------------------------------------------------------
maxCache information
----------------------------------------------------------------------
   No maxCache Array found


Command completed successfully.
`

var arcConfGetConfigLD = `
Logical Device number 0
   Logical Device name                        : Logical Drive 1
   Disk Name                                  : /dev/sdd (Disk0) (Bus: 1, Target: 0, Lun: 0)
   Block Size of member drives                : 512 Bytes
   Array                                      : 0
   RAID level                                 : 0
   Status of Logical Device                   : Optimal
   Size                                       : 572293 MB
   Stripe-unit size                           : 128 KB
   Full Stripe Size                           : 128 KB
   Interface Type                             : Serial Attached ATA
   Device Type                                : Data
   Boot Type                                  : None
   Heads                                      : 255
   Sectors Per Track                          : 32
   Cylinders                                  : 65535
   Caching                                    : Enabled
   Mount Points                               : Not Mounted
   LD Acceleration Method                     : Controller Cache
   SED Encryption                             : Disabled
   Volume Unique Identifier                   : 600508B1001CB9F0FE988FC40CD17395
--------------------------------------------------------
   Consistency Check Information
--------------------------------------------------------
   Consistency Check Status                   : Not Applicable
   Last Consistency Check Completion Time     : Not Applicable
   Last Consistency Check Duration            : Not Applicable
--------------------------------------------------------
   Array Physical Device Information
--------------------------------------------------------
   Device ID                                  : Availability (SizeMB, Protocol, Type, Connector ID, Location) Serial Number
--------------------------------------------------------
   Device 0                                   : Present (572325MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:1) 63M0A0BYFJPF


Logical Device number 1
   Logical Device name                        : Logical Drive 2
   Disk Name                                  : /dev/sde (Disk0) (Bus: 1, Target: 0, Lun: 1)
   Block Size of member drives                : 512 Bytes
   Array                                      : 1
   RAID level                                 : 0
   Status of Logical Device                   : Optimal
   Size                                       : 1144609 MB
   Stripe-unit size                           : 128 KB
   Full Stripe Size                           : 128 KB
   Interface Type                             : Serial Attached SCSI
   Device Type                                : Data
   Boot Type                                  : None
   Heads                                      : 255
   Sectors Per Track                          : 32
   Cylinders                                  : 65535
   Caching                                    : Enabled
   Mount Points                               : /boot/efi 1075 MB  Partition Number 1
   LD Acceleration Method                     : Controller Cache
   SED Encryption                             : Disabled
   Volume Unique Identifier                   : 600508B1001CFE283CA96614826A7F85
--------------------------------------------------------
   Consistency Check Information
--------------------------------------------------------
   Consistency Check Status                   : Not Applicable
   Last Consistency Check Completion Time     : Not Applicable
   Last Consistency Check Duration            : Not Applicable
--------------------------------------------------------
   Array Physical Device Information
--------------------------------------------------------
   Device ID                                  : Availability (SizeMB, Protocol, Type, Connector ID, Location) Serial Number
--------------------------------------------------------
   Device 1                                   : Present (1144641MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:2) 59M0A06CFJRG


Logical Device number 2
   Logical Device name                        : Logical Drive 3
   Disk Name                                  : /dev/sdc (Disk0) (Bus: 1, Target: 0, Lun: 2)
   Block Size of member drives                : 512 Bytes
   Array                                      : 2
   RAID level                                 : 0
   Status of Logical Device                   : Optimal
   Size                                       : 1144609 MB
   Stripe-unit size                           : 128 KB
   Full Stripe Size                           : 128 KB
   Interface Type                             : Serial Attached SCSI
   Device Type                                : Data
   Boot Type                                  : Primary and Secondary
   Heads                                      : 255
   Sectors Per Track                          : 32
   Cylinders                                  : 65535
   Caching                                    : Enabled
   Mount Points                               : Not Mounted
   LD Acceleration Method                     : Controller Cache
   SED Encryption                             : Disabled
   Volume Unique Identifier                   : 600508B1001C6DB81E7099960E5B5796
--------------------------------------------------------
   Consistency Check Information
--------------------------------------------------------
   Consistency Check Status                   : Not Applicable
   Last Consistency Check Completion Time     : Not Applicable
   Last Consistency Check Duration            : Not Applicable
--------------------------------------------------------
   Array Physical Device Information
--------------------------------------------------------
   Device ID                                  : Availability (SizeMB, Protocol, Type, Connector ID, Location) Serial Number
--------------------------------------------------------
   Device 2                                   : Present (1144641MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:3) WFK076DT0000E821CET2



Command completed successfully.
`

var arcList = `
Controllers found: 3
----------------------------------------------------------------------
Controller information
----------------------------------------------------------------------
   Controller ID             : Status, Slot, Mode, Name, SerialNumber, WWN
----------------------------------------------------------------------
   Controller 1:             : Optimal, Slot 16, Mixed, Cisco 24G TriMode M1 RAID 4GB FBWC 16D UCSC-RAID-M1L16, 3137F30003A, 50000D1E01787B00
   Controller 2:             : Optimal, Slot 9, Mixed, Cisco 24G TriMode M1 RAID 4GB FBWC 16D UCSC-RAID-M1L16, A312J89902X, 50000XKJLKJSJHFG
   Controller 3:             : Optimal, Slot 4, Mixed, Cisco 24G TriMode M1 RAID 4GB FBWC 16D UCSC-RAID-M1L16, B8812N9128Y, 50000ZKJSK8391J9

Command completed successfully.
`

var arcListBad = `
Controllers found: 3
----------------------------------------------------------------------
Controller information
----------------------------------------------------------------------
   Controller ID             : Status, Slot, Mode, Name, SerialNumber, WWN
----------------------------------------------------------------------
   Controller 1:             : Optimal, Slot 16, Mixed, Cisco 24G TriMode M1 RAID 4GB FBWC 16D UCSC-RAID-M1L16, 3137F30003A, 50000D1E01787B00

Command completed successfully.
`

func TestParseBlockSize(t *testing.T) {
	testCases := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"512", 512, false},
		{"4096", 4096, false},
		{"4K", 4096, false},
		{"4k", 4096, false},
		// "0" is rejected by parseBlockSize; the call site accepts it only for
		// failed devices.
		{"0", 0, true},
		{"", 0, true},
		{"abc", 0, true},
		{"1M", 0, true},
		{"1G", 0, true},
	}

	for _, tc := range testCases {
		result, err := parseBlockSize(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseBlockSize(%q): expected error, got %d", tc.input, result)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseBlockSize(%q): unexpected error: %s", tc.input, err)
			continue
		}
		if result != tc.expected {
			t.Errorf("parseBlockSize(%q): got %d, expected %d", tc.input, result, tc.expected)
		}
	}
}

var arcConfGetConfigLD4K = `
Logical Device number 0
   Logical Device name                        : Logical Drive 1
   Disk Name                                  : /dev/sda (Disk0) (Bus: 1, Target: 0, Lun: 0)
   Block Size of member drives                : 4K Bytes
   Array                                      : 0
   RAID level                                 : 0
   Status of Logical Device                   : Optimal
   Size                                       : 572293 MB
   Interface Type                             : Serial Attached SCSI
`

func TestSmartPqiLogicalDevice4KBlockSize(t *testing.T) {
	found, err := parseLogicalDevices(arcConfGetConfigLD4K)
	if err != nil {
		t.Errorf("failed to parse arcconf getconfig with 4K block size: %s", err)
		return
	}
	if len(found) != 1 {
		t.Errorf("expected 1 logical device, found %d", len(found))
		return
	}
	if found[0].BlockSize != 4096 {
		t.Errorf("expected BlockSize 4096, got %d", found[0].BlockSize)
	}
}

func TestSmartPqiList(t *testing.T) {
	found, err := parseList(arcList)
	if err != nil {
		t.Errorf("failed to parse arcconf list command output: %s", err)
	}

	if len(found) != 3 {
		t.Errorf("parseList found %d expected 3", len(found))
	}

	expected := []int{1, 2, 3}
	for i := range found {
		if !reflect.DeepEqual(expected[i], found[i]) {
			t.Errorf("entry %d: found controller ID: %d expected: %d", i, expected[i], found[i])
		}
	}
}

func TestSmartPqiListBad(t *testing.T) {
	_, err := parseList(arcListBad)
	if err == nil {
		t.Errorf("expected error on bad input")
	}

	expected := "mismatched output, found 1 controllers, expected 3"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("expected '%s' message, got %q", expected, err)
	}
}

func TestSmartPqiLogicaDevice(t *testing.T) {
	expectedLDs := []LogicalDevice{}
	expectedLDsJSON := []string{
		`{"ArrayID":0,"BlockSize":512,"Devices":null,"DiskName":"/dev/sdd","ID":0,"InterfaceType":"ATA","Name":"Logical Drive 1","RAIDLevel":"0","SizeMB":572293}`,
		`{"ArrayID":1,"BlockSize":512,"Devices":null,"DiskName":"/dev/sde","ID":1,"InterfaceType":"SCSI","Name":"Logical Drive 2","RAIDLevel":"0","SizeMB":1144609}`,
		`{"ArrayID":2,"BlockSize":512,"Devices":null,"DiskName":"/dev/sdc","ID":2,"InterfaceType":"SCSI","Name":"Logical Drive 3","RAIDLevel":"0","SizeMB":1144609}`,
	}

	for idx, content := range expectedLDsJSON {
		ld := LogicalDevice{}
		if err := json.Unmarshal([]byte(content), &ld); err != nil {
			t.Errorf("failed to unmarshal expected LD JSON index %d: %s", idx, err)
		}
		if len(ld.DiskName) == 0 {
			t.Fatalf("Failed to unmarshal JSON blob correctly")
		}
		expectedLDs = append(expectedLDs, ld)
	}

	if len(expectedLDs) != len(expectedLDsJSON) {
		t.Errorf("failed to marshall expected number of LogicalDevice, found %d, expected %d", len(expectedLDs), len(expectedLDsJSON))
	}

	found, err := parseLogicalDevices(arcConfGetConfigLD)
	if err != nil {
		t.Errorf("failed to parse arcconf getconfig 1 ld: %s", err)
	}

	for i := range found {
		if !reflect.DeepEqual(expectedLDs[i], found[i]) {
			t.Errorf("entry %d logical device differed:\n found: %#v\n expct: %#v ", i, found[i], expectedLDs[i])
		}
	}
}

func TestSmartPqiLogicaDeviceBadInput(t *testing.T) {
	var input = `Logical Device number 0`
	found, err := parseLogicalDevices(input)
	if err == nil {
		t.Errorf("did not return error with bad input")
	}
	if len(found) != 0 {
		t.Errorf("expected 0, found %d", len(found))
	}
	expectedErr := "expected exactly 2 lines of data, found"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected %q, got %q", expectedErr, err)
	}
}

func TestSmartPqiLogicaDeviceBadDeviceNumber(t *testing.T) {
	var input = `
Logical Device number #!
   Logical Device name                        : Logical Drive 1
   Disk Name                                  : /dev/sdd (Disk0) (Bus: 1, Target: 0, Lun: 0)
`
	found, err := parseLogicalDevices(input)
	if err == nil {
		t.Errorf("did not return error with bad output")
	}
	if len(found) != 0 {
		t.Errorf("expected 0, found %d", len(found))
	}
	expectedErr := "error while parsing integer from"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected %q message, got %q", expectedErr, err)
	}
}

func TestSmartPqiLogicaDeviceNumberMultidigit(t *testing.T) {
	var input = `
Logical Device number 127
   Logical Device name                        : Logical Drive 127
   Disk Name                                  : /dev/sdd (Disk0) (Bus: 1, Target: 0, Lun: 0)
   Array                                      : 127
`
	found, err := parseLogicalDevices(input)
	if err != nil {
		t.Errorf("expected err to be nil, got %s", err)
	}
	if len(found) != 1 {
		t.Errorf("expected 1, found %d", len(found))
	}
	if found[0].ArrayID != 127 {
		t.Errorf("expected logical device with ArrayID = 127, got %d", found[0].ArrayID)
	}
}

func TestSmartPqiLogicaDeviceBadAtoiTest(t *testing.T) {
	var inputArray = `
Logical Device number 0
   Logical Device name                        : Logical Drive 1
   Disk Name                                  : /dev/sdd (Disk0) (Bus: 1, Target: 0, Lun: 0)
   Array                                      : err
   Block Size of member drives                : 512 Bytes
   Size                                       : 1144641 MB
`
	var inputBsize = `
Logical Device number 0
   Logical Device name                        : Logical Drive 1
   Disk Name                                  : /dev/sdd (Disk0) (Bus: 1, Target: 0, Lun: 0)
   Array                                      : 1
   Block Size of member drives                : zero Bytes
   Size                                       : 1144641 MB
`
	var inputSize = `
Logical Device number 0
   Logical Device name                        : Logical Drive 1
   Disk Name                                  : /dev/sdd (Disk0) (Bus: 1, Target: 0, Lun: 0)
   Array                                      : 1
   Block Size of member drives                : 512 Bytes
   Size                                       : xxxxx MB
`

	testCases := []struct {
		input       string
		expectedErr string
	}{
		{inputArray, "failed to parse ArrayID from token"},
		{inputBsize, "failed to parse BlockSize from token"},
		{inputSize, "failed to parse Size from token"},
	}

	for idx, testCase := range testCases {
		found, err := parseLogicalDevices(testCase.input)
		if err == nil {
			t.Errorf("did not return error with bad input index %d", idx)
		}
		if len(found) != 0 {
			t.Errorf("expected 0, found %d index %d", len(found), idx)
		}
		if !strings.Contains(err.Error(), testCase.expectedErr) {
			t.Errorf("expected '%s' message, got %q index %d", testCase.expectedErr, err, idx)
		}
	}
}

func TestArcconfParseConf(t *testing.T) {
	expectedPDs := []PhysicalDevice{}
	expectedPDsJSON := []string{
		`{"ArrayID":0,"Availability":"Online","BlockSize":512,"Channel":0,"ID":0,"Firmware":"5703","Model":"AL15SEB060N","PhysicalBlockSize":512,"Protocol":"SAS","SerialNumber":"63M0A0BYFJPF","SizeMB":572325,"Type":"HDD","Vendor":"TOSHIBA","WriteCache":"Disabled (write-through)"}`,
		`{"ArrayID":1,"Availability":"Online","BlockSize":512,"Channel":0,"ID":1,"Firmware":"5701","Model":"AL15SEB120N","PhysicalBlockSize":512,"Protocol":"SAS","SerialNumber":"59M0A06CFJRG","SizeMB":1144641,"Type":"HDD","Vendor":"TOSHIBA","WriteCache":"Disabled (write-through)"}`,
		`{"ArrayID":2,"Availability":"Online","BlockSize":512,"Channel":0,"ID":2,"Firmware":"CN03","Model":"ST1200MM0009","PhysicalBlockSize":512,"Protocol":"SAS","SerialNumber":"WFK076DT0000E821CET2","SizeMB":1144641,"Type":"HDD","Vendor":"SEAGATE","WriteCache":"Disabled (write-through)"}`,
	}

	for idx, content := range expectedPDsJSON {
		pd := PhysicalDevice{}
		if err := json.Unmarshal([]byte(content), &pd); err != nil {
			t.Errorf("failed to unmarshal expected PD JSON index %d: %s", idx, err)
		}
		if len(pd.SerialNumber) == 0 {
			t.Fatalf("Failed to unmarshal JSON blob correctly")
		}
		expectedPDs = append(expectedPDs, pd)
	}

	if len(expectedPDs) != len(expectedPDsJSON) {
		t.Errorf("failed to marshall expected number of PhysicalDevices, found %d, expected %d", len(expectedPDs), len(expectedPDsJSON))
	}

	_, found, err := parseGetConf(arcConfGetConfig)
	if err != nil {
		t.Errorf("failed to parse arcconf getconfig 1: %s", err)
	}

	for i := range found {
		if !reflect.DeepEqual(expectedPDs[i], found[i]) {
			t.Errorf("entry %d physical device differed:\n found: %#v\n expct: %#v ", i, found[i], expectedPDs[i])
		}
	}
}

func TestSmartPqiPhysicalDeviceBadFormats(t *testing.T) {
	var inputShort = `
      Device #0
`
	var inputDevID = `
      Device #err
         Device is a Hard drive
         State                                : Online
         Drive has stale RIS data             : False
         Block Size                           : 512 Bytes
`
	var inputBsize = `
      Device #0
         State                                : Online
         Device is a Hard drive
         Drive has stale RIS data             : False
         Block Size                           : XXX Bytes
`
	var inputPhyBsize = `
      Device #0
         Physical Block Size                  : abc Bytes
`
	var inputTotalSize = `
      Device #0
         Total Size                           : XXX MB
`

	testCases := []struct {
		input       string
		expectedErr string
	}{
		{inputShort, "expected more than 2 lines of data"},
		{inputDevID, "error finding start of PhysicalDevice in data"},
		{inputBsize, "failed to parse Block Size from token"},
		{inputPhyBsize, "failed to parse Physical Block Size from token"},
		{inputTotalSize, "failed to parse Total Size from token"},
	}

	for idx, testCase := range testCases {
		found, err := parsePhysicalDevices(testCase.input)
		if err == nil {
			t.Errorf("did not return error with bad input index %d", idx)
		}
		if len(found) != 0 {
			t.Errorf("expected 0, found %d index %d", len(found), idx)
		}
		if !strings.Contains(err.Error(), testCase.expectedErr) {
			t.Errorf("expected '%s' message, got %q index %d", testCase.expectedErr, err, idx)
		}
	}
}

// Failed drives report "Block Size : 0 Bytes"; arcconf cannot query the drive.
// Parsing must succeed and leave BlockSize at 0 for a device with State=Failed.
func TestParsePhysicalDeviceBlockSizeZeroFailed(t *testing.T) {
	input := `
Physical Device information
----------------------------------------------------------------------
   Channel #0:
      Device #0
         Device is a Hard drive
         State                                : Failed
         Block Size                           : 0
         Physical Block Size                  : 0
         Reported Channel,Device(T:L)         : 0,0(0:0)
         Array                                : 0
         Vendor                               : Not Available
         Model                                : Not Available
         Firmware                             : Not Available
         Serial number                        : Not Available
         Total Size                           : 0 MB
         Write Cache                          : Unknown
         SSD                                  : No
`
	pDevs, err := parsePhysicalDevices(input)
	if err != nil {
		t.Fatalf("unexpected error parsing failed physical device with Block Size 0: %s", err)
	}
	if len(pDevs) != 1 {
		t.Fatalf("expected 1 physical device, got %d", len(pDevs))
	}
	if pDevs[0].BlockSize != 0 {
		t.Errorf("expected BlockSize 0, got %d", pDevs[0].BlockSize)
	}
	if pDevs[0].PhysicalBlockSize != 0 {
		t.Errorf("expected PhysicalBlockSize 0, got %d", pDevs[0].PhysicalBlockSize)
	}
}

// An Online device reporting "Block Size : 0" is unexpected and should still
// be treated as a parse error.
func TestParsePhysicalDeviceBlockSizeZeroOnlineErrors(t *testing.T) {
	input := `
Physical Device information
----------------------------------------------------------------------
   Channel #0:
      Device #0
         Device is a Hard drive
         State                                : Online
         Block Size                           : 0
         Physical Block Size                  : 512
         Reported Channel,Device(T:L)         : 0,0(0:0)
         Array                                : 0
         Vendor                               : TOSHIBA
         Model                                : AL15SEB060N
         Firmware                             : 5703
         Serial number                        : DEADBEEF0001
         Total Size                           : 572325 MB
         Write Cache                          : Disabled (write-through)
         SSD                                  : No
`
	_, err := parsePhysicalDevices(input)
	if err == nil {
		t.Fatal("expected error for Online device with Block Size 0, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse Block Size") {
		t.Errorf("expected 'failed to parse Block Size' in error, got %q", err)
	}
}

func TestSmartPqiInterface(t *testing.T) {
	arc := ArcConf()
	if arc == nil {
		t.Errorf("expected ArcConf pointer, found nil")
	}
}

func TestSmartPqiNewController(t *testing.T) {
	cID := 1
	ctrl, err := newController(cID, arcConfGetConfig)
	if err != nil {
		t.Errorf("unexpected error creating new controller: %s", err)
	}

	expectedJSON := `{"ID":1,"PhysicalDrives":{"0":{"ArrayID":0,"Availability":"Online","BlockSize":512,"Channel":0,"ID":0,"Firmware":"5703","Model":"AL15SEB060N","PhysicalBlockSize":512,"Protocol":"SAS","SerialNumber":"63M0A0BYFJPF","SizeMB":572325,"Type":"HDD","Vendor":"TOSHIBA","WriteCache":"Disabled (write-through)"},"1":{"ArrayID":1,"Availability":"Online","BlockSize":512,"Channel":0,"ID":1,"Firmware":"5701","Model":"AL15SEB120N","PhysicalBlockSize":512,"Protocol":"SAS","SerialNumber":"59M0A06CFJRG","SizeMB":1144641,"Type":"HDD","Vendor":"TOSHIBA","WriteCache":"Disabled (write-through)"},"2":{"ArrayID":2,"Availability":"Online","BlockSize":512,"Channel":0,"ID":2,"Firmware":"CN03","Model":"ST1200MM0009","PhysicalBlockSize":512,"Protocol":"SAS","SerialNumber":"WFK076DT0000E821CET2","SizeMB":1144641,"Type":"HDD","Vendor":"SEAGATE","WriteCache":"Disabled (write-through)"}},"LogicalDrives":{"0":{"ArrayID":0,"BlockSize":512,"Caching":"","Devices":[{"ArrayID":0,"Availability":"Online","BlockSize":512,"Channel":0,"ID":0,"Firmware":"5703","Model":"AL15SEB060N","PhysicalBlockSize":512,"Protocol":"SAS","SerialNumber":"63M0A0BYFJPF","SizeMB":572325,"Type":"HDD","Vendor":"TOSHIBA","WriteCache":"Disabled (write-through)"}],"DiskName":"/dev/sdd","ID":0,"InterfaceType":"SCSI","Name":"Logical Drive 1","RAIDLevel":"0","SizeMB":572293},"1":{"ArrayID":1,"BlockSize":512,"Caching":"","Devices":[{"ArrayID":1,"Availability":"Online","BlockSize":512,"Channel":0,"ID":1,"Firmware":"5701","Model":"AL15SEB120N","PhysicalBlockSize":512,"Protocol":"SAS","SerialNumber":"59M0A06CFJRG","SizeMB":1144641,"Type":"HDD","Vendor":"TOSHIBA","WriteCache":"Disabled (write-through)"}],"DiskName":"/dev/sde","ID":1,"InterfaceType":"SCSI","Name":"Logical Drive 2","RAIDLevel":"0","SizeMB":1144609},"2":{"ArrayID":2,"BlockSize":512,"Caching":"","Devices":[{"ArrayID":2,"Availability":"Online","BlockSize":512,"Channel":0,"ID":2,"Firmware":"CN03","Model":"ST1200MM0009","PhysicalBlockSize":512,"Protocol":"SAS","SerialNumber":"WFK076DT0000E821CET2","SizeMB":1144641,"Type":"HDD","Vendor":"SEAGATE","WriteCache":"Disabled (write-through)"}],"DiskName":"/dev/sdc","ID":2,"InterfaceType":"SCSI","Name":"Logical Drive 3","RAIDLevel":"0","SizeMB":1144609}}}`
	expCtrl := Controller{}
	if err := json.Unmarshal([]byte(expectedJSON), &expCtrl); err != nil {
		t.Errorf("failed to unmarshal expected Controller JSON: %s", err)
	}

	if !reflect.DeepEqual(expCtrl, ctrl) {
		t.Errorf("controller differed:\n found: %#v\n expct: %#v ", ctrl, expCtrl)
	}
}

// arcconf GETCONFIG output with mixed SSD/HDD arrays and 4K native sector drives.
// Confidential identifiers (serial numbers, WWNs) replaced with synthetic values.
// Unparsed sections (sensors, SPDM, error counters, phy info) trimmed for clarity.
// This tests that the parser handles multiple LDs in the same \n\n\n chunk.
var arcConfGetConfigMixedArrays = `
Controllers found: 1
----------------------------------------------------------------------
Controller information
----------------------------------------------------------------------
   Controller Status                                     : Optimal
   Controller Mode                                       : Mixed
   Controller Model                                      : Cisco 24G TriMode M1 RAID 4GB FBWC 16D UCSC-RAID-M1L16
   Driver Name                                           : smartpqi
   Connector #0
      Connector name                                     : CN2


   Connector #1
      Connector name                                     : CN3



----------------------------------------------------------------------
Array Information
----------------------------------------------------------------------
Array Number 0
   Name                                                  : A
   Interface                                             : SATA SSD
   Block Size                                            : 512 Bytes
   Device 0                                              : Present (915715MB, SATA, SSD, Connector:CN2, Backplane:0, Slot:1) SN_SSD_0


Array Number 1
   Name                                                  : B
   Interface                                             : SAS 4K
   Block Size                                            : 4K Bytes
   Device 1                                              : Present (2289272MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:2) SN_HDD_1
   Device 2                                              : Present (2289272MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:3) SN_HDD_2
   Device 3                                              : Present (2289272MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:4) SN_HDD_3
   Device 4                                              : Present (2289272MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:6) SN_HDD_4
   Device 5                                              : Present (2289272MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:7) SN_HDD_5
   Device 8                                              : Present (2289272MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:5) SN_HDD_8



--------------------------------------------------------
Logical device information
--------------------------------------------------------
Logical Device number 0
   Logical Device name                                   : RAID0_1
   Disk Name                                             : /dev/sdb (Disk0) (Bus: 1, Target: 0, Lun: 0)
   Block Size of member drives                           : 512 Bytes
   Array                                                 : 0
   RAID level                                            : 0
   Status of Logical Device                              : Optimal
   Size                                                  : 915527 MB
   Interface Type                                        : SATA SSD
   Device 0                                              : Present (915715MB, SATA, SSD, Connector:CN2, Backplane:0, Slot:1) SN_SSD_0

Logical Device number 1
   Logical Device name                                   : RAID0_234567
   Disk Name                                             : /dev/sda (Disk0) (Bus: 1, Target: 0, Lun: 1)
   Block Size of member drives                           : 4K Bytes
   Array                                                 : 1
   RAID level                                            : 0
   Status of Logical Device                              : Optimal
   Size                                                  : 13732910 MB
   Interface Type                                        : SAS 4K
   Device 1                                              : Present (2289272MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:2) SN_HDD_1
   Device 8                                              : Present (2289272MB, SAS, HDD, Connector:CN2, Backplane:0, Slot:5) SN_HDD_8


----------------------------------------------------------------------
Physical Device information
----------------------------------------------------------------------
   Channel #0:
      Device #0
         Device is a Hard drive
         State                                           : Online
         Block Size                                      : 512 Bytes
         Physical Block Size                             : 4K Bytes
         Negotiated Transfer Speed                       : SATA 6.0 Gb/s
         Reported Channel,Device(T:L)                    : 0,0(0:0)
         Array                                           : 0
         Vendor                                          : ATA
         Model                                           : Micron_5300_MTFDDAK960TDS
         Firmware                                        : D3MC000
         Serial number                                   : SN_SSD_0
         Total Size                                      : 915715 MB
         Write Cache                                     : Disabled (write-through)
         SSD                                             : Yes

      Device #1
         Device is a Hard drive
         State                                           : Online
         Block Size                                      : 4K Bytes
         Physical Block Size                             : 4K Bytes
         Negotiated Transfer Speed                       : SAS 12.0 Gb/s
         Reported Channel,Device(T:L)                    : 0,1(1:0)
         Array                                           : 1
         Vendor                                          : TOSHIBA
         Model                                           : AL15SEB24EP
         Firmware                                        : 5704
         Serial number                                   : SN_HDD_1
         Total Size                                      : 2289272 MB
         Write Cache                                     : Disabled (write-through)
         SSD                                             : No

      Device #2
         Device is a Hard drive
         State                                           : Online
         Block Size                                      : 4K Bytes
         Physical Block Size                             : 4K Bytes
         Negotiated Transfer Speed                       : SAS 12.0 Gb/s
         Reported Channel,Device(T:L)                    : 0,2(2:0)
         Array                                           : 1
         Vendor                                          : TOSHIBA
         Model                                           : AL15SEB24EP
         Firmware                                        : 5704
         Serial number                                   : SN_HDD_2
         Total Size                                      : 2289272 MB
         Write Cache                                     : Enabled
         SSD                                             : No

      Device #3
         Device is a Hard drive
         State                                           : Online
         Block Size                                      : 4K Bytes
         Physical Block Size                             : 4K Bytes
         Negotiated Transfer Speed                       : SAS 12.0 Gb/s
         Reported Channel,Device(T:L)                    : 0,3(3:0)
         Array                                           : 1
         Vendor                                          : TOSHIBA
         Model                                           : AL15SEB24EP
         Firmware                                        : 5704
         Serial number                                   : SN_HDD_3
         Total Size                                      : 2289272 MB
         Write Cache                                     : Disabled (write-through)
         SSD                                             : No

      Device #4
         Device is a Hard drive
         State                                           : Online
         Block Size                                      : 4K Bytes
         Physical Block Size                             : 4K Bytes
         Negotiated Transfer Speed                       : SAS 12.0 Gb/s
         Reported Channel,Device(T:L)                    : 0,4(4:0)
         Array                                           : 1
         Vendor                                          : TOSHIBA
         Model                                           : AL15SEB24EP
         Firmware                                        : 5704
         Serial number                                   : SN_HDD_4
         Total Size                                      : 2289272 MB
         Write Cache                                     : Disabled (write-through)
         SSD                                             : No

      Device #5
         Device is a Hard drive
         State                                           : Online
         Block Size                                      : 4K Bytes
         Physical Block Size                             : 4K Bytes
         Negotiated Transfer Speed                       : SAS 12.0 Gb/s
         Reported Channel,Device(T:L)                    : 0,5(5:0)
         Array                                           : 1
         Vendor                                          : TOSHIBA
         Model                                           : AL15SEB24EP
         Firmware                                        : 5704
         Serial number                                   : SN_HDD_5
         Total Size                                      : 2289272 MB
         Write Cache                                     : Disabled (write-through)
         SSD                                             : No

      Device #8
         Device is a Hard drive
         State                                           : Online
         Block Size                                      : 4K Bytes
         Physical Block Size                             : 4K Bytes
         Negotiated Transfer Speed                       : SAS 12.0 Gb/s
         Reported Channel,Device(T:L)                    : 0,8(8:0)
         Array                                           : 1
         Vendor                                          : TOSHIBA
         Model                                           : AL15SEB24EP
         Firmware                                        : 5704
         Serial number                                   : SN_HDD_8
         Total Size                                      : 2289272 MB
         Write Cache                                     : Disabled (write-through)
         SSD                                             : No

   Channel #2:
      Device #0
         Device is an Enclosure Services Device
         Reported Channel,Device(T:L)                    : 2,0(0:0)
         Enclosure ID                                    : 0
         Type                                            : SES2
         Vendor                                          : Cisco

     Backplane:
      Device #0
         Device is an UBM Controller
         Backplane ID                                    : 0
         Type                                            : UBM
         Firmware                                        : 0.2 (dec), 0.2 (hex)


----------------------------------------------------------------------
maxCache information
----------------------------------------------------------------------
   No maxCache Array found


Command completed successfully.
`

func TestSmartPqiGetConfigMixedArrays(t *testing.T) {
	ctrl, err := newController(1, arcConfGetConfigMixedArrays)
	if err != nil {
		t.Fatalf("failed to parse arcconf getconfig: %s", err)
	}

	// --- Logical Drives ---

	expectedLDs := []struct {
		id            int
		name          string
		diskName      string
		blockSize     int
		arrayID       int
		raidLevel     string
		sizeMB        int
		interfaceType string
		isSSD         bool
		numDevices    int
	}{
		{0, "RAID0_1", "/dev/sdb", 512, 0, "0", 915527, "ATA", true, 1},
		{1, "RAID0_234567", "/dev/sda", 4096, 1, "0", 13732910, "SCSI", false, 6},
	}

	if len(ctrl.LogicalDrives) != len(expectedLDs) {
		t.Fatalf("expected %d logical drives, got %d", len(expectedLDs), len(ctrl.LogicalDrives))
	}

	for _, exp := range expectedLDs {
		ld := ctrl.LogicalDrives[exp.id]
		if ld == nil {
			t.Fatalf("missing logical drive %d", exp.id)
		}
		if ld.ID != exp.id {
			t.Errorf("LD%d ID: got %d, want %d", exp.id, ld.ID, exp.id)
		}
		if ld.Name != exp.name {
			t.Errorf("LD%d Name: got %q, want %q", exp.id, ld.Name, exp.name)
		}
		if ld.DiskName != exp.diskName {
			t.Errorf("LD%d DiskName: got %q, want %q", exp.id, ld.DiskName, exp.diskName)
		}
		if ld.BlockSize != exp.blockSize {
			t.Errorf("LD%d BlockSize: got %d, want %d", exp.id, ld.BlockSize, exp.blockSize)
		}
		if ld.ArrayID != exp.arrayID {
			t.Errorf("LD%d ArrayID: got %d, want %d", exp.id, ld.ArrayID, exp.arrayID)
		}
		if ld.RAIDLevel != exp.raidLevel {
			t.Errorf("LD%d RAIDLevel: got %q, want %q", exp.id, ld.RAIDLevel, exp.raidLevel)
		}
		if ld.SizeMB != exp.sizeMB {
			t.Errorf("LD%d SizeMB: got %d, want %d", exp.id, ld.SizeMB, exp.sizeMB)
		}
		if ld.InterfaceType != exp.interfaceType {
			t.Errorf("LD%d InterfaceType: got %q, want %q", exp.id, ld.InterfaceType, exp.interfaceType)
		}
		if ld.IsSSD() != exp.isSSD {
			t.Errorf("LD%d IsSSD: got %v, want %v", exp.id, ld.IsSSD(), exp.isSSD)
		}
		if len(ld.Devices) != exp.numDevices {
			t.Errorf("LD%d linked devices: got %d, want %d", exp.id, len(ld.Devices), exp.numDevices)
		}
	}

	// --- Physical Drives ---

	expectedPDs := []struct {
		id                int
		channel           int
		arrayID           int
		availability      string
		blockSize         int
		physicalBlockSize int
		protocol          string
		vendor            string
		model             string
		firmware          string
		serialNumber      string
		sizeMB            int
		mediaType         MediaType
		writeCache        string
	}{
		{0, 0, 0, "Online", 512, 4096, "SATA", "ATA", "Micron_5300_MTFDDAK960TDS", "D3MC000", "SN_SSD_0", 915715, SSD, "Disabled (write-through)"},
		{1, 0, 1, "Online", 4096, 4096, "SAS", "TOSHIBA", "AL15SEB24EP", "5704", "SN_HDD_1", 2289272, HDD, "Disabled (write-through)"},
		{2, 0, 1, "Online", 4096, 4096, "SAS", "TOSHIBA", "AL15SEB24EP", "5704", "SN_HDD_2", 2289272, HDD, "Enabled"},
		{3, 0, 1, "Online", 4096, 4096, "SAS", "TOSHIBA", "AL15SEB24EP", "5704", "SN_HDD_3", 2289272, HDD, "Disabled (write-through)"},
		{4, 0, 1, "Online", 4096, 4096, "SAS", "TOSHIBA", "AL15SEB24EP", "5704", "SN_HDD_4", 2289272, HDD, "Disabled (write-through)"},
		{5, 0, 1, "Online", 4096, 4096, "SAS", "TOSHIBA", "AL15SEB24EP", "5704", "SN_HDD_5", 2289272, HDD, "Disabled (write-through)"},
		{8, 0, 1, "Online", 4096, 4096, "SAS", "TOSHIBA", "AL15SEB24EP", "5704", "SN_HDD_8", 2289272, HDD, "Disabled (write-through)"},
	}

	if len(ctrl.PhysicalDrives) != len(expectedPDs) {
		t.Fatalf("expected %d physical drives, got %d", len(expectedPDs), len(ctrl.PhysicalDrives))
	}

	for _, exp := range expectedPDs {
		pd := ctrl.PhysicalDrives[exp.id]
		if pd == nil {
			t.Fatalf("missing physical drive %d", exp.id)
		}
		if pd.ID != exp.id {
			t.Errorf("PD%d ID: got %d, want %d", exp.id, pd.ID, exp.id)
		}
		if pd.Channel != exp.channel {
			t.Errorf("PD%d Channel: got %d, want %d", exp.id, pd.Channel, exp.channel)
		}
		if pd.ArrayID != exp.arrayID {
			t.Errorf("PD%d ArrayID: got %d, want %d", exp.id, pd.ArrayID, exp.arrayID)
		}
		if pd.Availability != exp.availability {
			t.Errorf("PD%d Availability: got %q, want %q", exp.id, pd.Availability, exp.availability)
		}
		if pd.BlockSize != exp.blockSize {
			t.Errorf("PD%d BlockSize: got %d, want %d", exp.id, pd.BlockSize, exp.blockSize)
		}
		if pd.PhysicalBlockSize != exp.physicalBlockSize {
			t.Errorf("PD%d PhysicalBlockSize: got %d, want %d", exp.id, pd.PhysicalBlockSize, exp.physicalBlockSize)
		}
		if pd.Protocol != exp.protocol {
			t.Errorf("PD%d Protocol: got %q, want %q", exp.id, pd.Protocol, exp.protocol)
		}
		if pd.Vendor != exp.vendor {
			t.Errorf("PD%d Vendor: got %q, want %q", exp.id, pd.Vendor, exp.vendor)
		}
		if pd.Model != exp.model {
			t.Errorf("PD%d Model: got %q, want %q", exp.id, pd.Model, exp.model)
		}
		if pd.Firmware != exp.firmware {
			t.Errorf("PD%d Firmware: got %q, want %q", exp.id, pd.Firmware, exp.firmware)
		}
		if pd.SerialNumber != exp.serialNumber {
			t.Errorf("PD%d SerialNumber: got %q, want %q", exp.id, pd.SerialNumber, exp.serialNumber)
		}
		if pd.SizeMB != exp.sizeMB {
			t.Errorf("PD%d SizeMB: got %d, want %d", exp.id, pd.SizeMB, exp.sizeMB)
		}
		if pd.Type != exp.mediaType {
			t.Errorf("PD%d Type: got %v, want %v", exp.id, pd.Type, exp.mediaType)
		}
		if pd.WriteCache != exp.writeCache {
			t.Errorf("PD%d WriteCache: got %q, want %q", exp.id, pd.WriteCache, exp.writeCache)
		}
	}
}

func TestSmartPqiNewControllerBadInput(t *testing.T) {
	var inputEmpty = ``
	var inputEmptyLines = `


`
	testCases := []struct {
		input       string
		expectedErr string
	}{
		{inputEmpty, "failed to parse arcconf getconfig output: cannot parse an empty string"},
		{inputEmptyLines, "failed to parse arcconf getconfig output: expected more than 3 lines of data in input"},
	}

	cID := 1
	for idx, testCase := range testCases {
		_, err := newController(cID, testCase.input)
		if err == nil {
			t.Errorf("did not return error with bad input index %d", idx)
			continue
		}

		if !strings.Contains(err.Error(), testCase.expectedErr) {
			t.Errorf("expected '%s' message, got %q index %d", testCase.expectedErr, err, idx)
		}
	}
}

// jbodCtrlWithPDs: one HDD, one SSD, one NVMe keyed by distinct serials.
func jbodCtrlWithPDs() Controller {
	return Controller{
		ID: 1,
		PhysicalDrives: DriveSet{
			10: &PhysicalDevice{ID: 10, SerialNumber: "SN-HDD-1", Type: HDD},
			20: &PhysicalDevice{ID: 20, SerialNumber: "SN-SSD-1", Type: SSD},
			30: &PhysicalDevice{ID: 30, SerialNumber: "SN-NVME-1", Type: NVME},
		},
	}
}

func TestJBODDiskTypeSerialHDDShort(t *testing.T) {
	ud := disko.UdevInfo{
		Properties: map[string]string{"ID_SERIAL_SHORT": "SN-HDD-1"},
	}

	dType, ok := jbodDiskTypeFromSerial([]Controller{jbodCtrlWithPDs()}, ud)
	require.True(t, ok, "expected ok=true when ID_SERIAL_SHORT matches")
	assert.Equal(t, disko.HDD, dType, "jbodDiskTypeFromSerial")
}

// ID_SERIAL is the fallback when ID_SERIAL_SHORT is absent.
func TestJBODDiskTypeSerialSSDFallback(t *testing.T) {
	ud := disko.UdevInfo{
		Properties: map[string]string{"ID_SERIAL": "SN-SSD-1"},
	}

	dType, ok := jbodDiskTypeFromSerial([]Controller{jbodCtrlWithPDs()}, ud)
	require.True(t, ok, "expected ok=true when ID_SERIAL matches as fallback")
	assert.Equal(t, disko.SSD, dType, "jbodDiskTypeFromSerial")
}

func TestJBODDiskTypeSerialNVME(t *testing.T) {
	ud := disko.UdevInfo{
		Properties: map[string]string{"ID_SERIAL_SHORT": "SN-NVME-1"},
	}

	dType, ok := jbodDiskTypeFromSerial([]Controller{jbodCtrlWithPDs()}, ud)
	require.True(t, ok, "expected ok=true for NVME serial match")
	assert.Equal(t, disko.NVME, dType, "jbodDiskTypeFromSerial")
}

// No ID_SERIAL* property in udev.
func TestJBODDiskTypeSerialNoSerial(t *testing.T) {
	ud := disko.UdevInfo{
		Properties: map[string]string{"ID_MODEL": "SomeModel"},
	}

	_, ok := jbodDiskTypeFromSerial([]Controller{jbodCtrlWithPDs()}, ud)
	assert.False(t, ok, "expected ok=false when udev carries no ID_SERIAL*")
}

// Serial not reported by any controller.
func TestJBODDiskTypeSerialNoMatch(t *testing.T) {
	ud := disko.UdevInfo{
		Properties: map[string]string{"ID_SERIAL_SHORT": "SN-UNKNOWN"},
	}

	_, ok := jbodDiskTypeFromSerial([]Controller{jbodCtrlWithPDs()}, ud)
	assert.False(t, ok, "expected ok=false when no PhysicalDevice matches")
}

// Duplicate serial across controllers is ambiguous.
func TestJBODDiskTypeSerialAmbiguous(t *testing.T) {
	dup := Controller{
		ID: 2,
		PhysicalDrives: DriveSet{
			99: &PhysicalDevice{ID: 99, SerialNumber: "SN-HDD-1", Type: SSD},
		},
	}

	ud := disko.UdevInfo{
		Properties: map[string]string{"ID_SERIAL_SHORT": "SN-HDD-1"},
	}

	_, ok := jbodDiskTypeFromSerial([]Controller{jbodCtrlWithPDs(), dup}, ud)
	assert.False(t, ok, "expected ok=false on duplicate SerialNumber across controllers")
}

// Matched PD with UnknownMedia cannot be classified.
func TestJBODDiskTypeSerialUnknownMedia(t *testing.T) {
	ctrl := Controller{
		ID: 3,
		PhysicalDrives: DriveSet{
			10: &PhysicalDevice{ID: 10, SerialNumber: "SN-UNK", Type: UnknownMedia},
		},
	}

	ud := disko.UdevInfo{
		Properties: map[string]string{"ID_SERIAL_SHORT": "SN-UNK"},
	}

	_, ok := jbodDiskTypeFromSerial([]Controller{ctrl}, ud)
	assert.False(t, ok, "expected ok=false when matched PD has UnknownMedia")
}

// Blank controller-side SerialNumber never matches.
func TestJBODDiskTypeSerialEmptyPDSerial(t *testing.T) {
	ctrl := Controller{
		ID: 4,
		PhysicalDrives: DriveSet{
			10: &PhysicalDevice{ID: 10, SerialNumber: "   ", Type: SSD},
		},
	}

	ud := disko.UdevInfo{
		Properties: map[string]string{"ID_SERIAL_SHORT": "SN-X"},
	}

	_, ok := jbodDiskTypeFromSerial([]Controller{ctrl}, ud)
	assert.False(t, ok, "expected ok=false when PhysicalDevice has blank SerialNumber")
}

func TestJBODDiskTypeSerialViaIDSCSISerial(t *testing.T) {
	ud := disko.UdevInfo{
		Properties: map[string]string{"ID_SCSI_SERIAL": "SN-HDD-1"},
	}

	dType, ok := jbodDiskTypeFromSerial([]Controller{jbodCtrlWithPDs()}, ud)
	require.True(t, ok, "expected ok=true when ID_SCSI_SERIAL matches")
	assert.Equal(t, disko.HDD, dType, "jbodDiskTypeFromSerial")
}

// isSoftArcconfErr must recognise the three soft sentinels whether bare
// or wrapped with fmt.Errorf("%w"), so wrapped per-controller errors are
// still routed through the udev fallback instead of propagating as fatal.
func TestIsSoftArcconfErr(t *testing.T) {
	hard := errors.New("arcconf blew up")

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"bare ErrNoArcconf", ErrNoArcconf, true},
		{"bare ErrNoController", ErrNoController, true},
		{"bare ErrUnsupported", ErrUnsupported, true},
		{"wrapped ErrNoArcconf", fmt.Errorf("controller id %d: %w", 0, ErrNoArcconf), true},
		{"wrapped ErrNoController", fmt.Errorf("controller id %d: %w", 1, ErrNoController), true},
		{"wrapped ErrUnsupported", fmt.Errorf("controller id %d: %w", 2, ErrUnsupported), true},
		{"unrelated hard error", hard, false},
		{"wrapped hard error", fmt.Errorf("controller id %d: %w", 3, hard), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isSoftArcconfErr(tc.err),
				"isSoftArcconfErr(%v)", tc.err)
		})
	}
}

// arcConfGetConfigDegradedRAID5 is a synthetic fixture that mimics a degraded
// RAID5 configuration. It exercises two problems in the same output:
//   - Device #2 and #6 report "Block Size : 0 Bytes" (failed drives)
//   - Device #0 and #7 are unassigned (no "Array :" field); they must NOT be
//     linked to LD 0 so that IsSSD() is not corrupted by unrelated HDD drives.
var arcConfGetConfigDegradedRAID5 = `
Controllers found: 1
----------------------------------------------------------------------
Controller information
----------------------------------------------------------------------
   Controller Status                                     : Optimal
   Controller Mode                                       : Mixed
   Controller Model                                      : Cisco 24G TriMode M1 RAID 4GB FBWC 16D UCSC-RAID-M1L16
   Driver Name                                           : smartpqi


----------------------------------------------------------------------
Array Information
----------------------------------------------------------------------
Array Number 0
   Name                                                  : A
   Status                                                : Has Failed Physical Drive
   Interface                                             : SAS SSD
   Total Size                                            : 18314162 MB
   Block Size                                            : 512 Bytes
   Type                                                  : Data


--------------------------------------------------------
Logical device information
--------------------------------------------------------
Logical Device number 0
   Logical Device name                                   : RAID5_123456
   Disk Name                                             : /dev/sdc (Disk0) (Bus: 1, Target: 0, Lun: 0)
   Block Size of member drives                           : 512 Bytes
   Array                                                 : 0
   RAID level                                            : 5
   Status of Logical Device                              : Interim Recovery
   Parity Initialization Status                          : Completed
   Size                                                  : 15258789 MB
   Interface Type                                        : SAS SSD
   Consistency Check Status                              : Not Running
   Last Consistency Check Completion Time                : Thu Aug 28 2025 06:46:07 UTC
   Last Consistency Check Duration                       : 2 hours 40 minutes 37 seconds


----------------------------------------------------------------------
Physical Device information
----------------------------------------------------------------------
   Channel #0:
      Device #0
         Device is a Hard drive
         State                                           : Ready
         Block Size                                      : 512 Bytes
         Physical Block Size                             : 4K Bytes
         Negotiated Transfer Speed                       : SATA 6.0 Gb/s
         Reported Channel,Device(T:L)                    : 0,0(0:0)
         Vendor                                          : ATA
         Model                                           : TestSSD-1T-A
         Firmware                                        : F0000
         Serial number                                   : TSTSER0000001
         Total Size                                      : 915715 MB
         Write Cache                                     : Enabled (write-back)
         SSD                                             : Yes

      Device #1
         Device is a Hard drive
         State                                           : Online
         Block Size                                      : 512 Bytes
         Physical Block Size                             : 4K Bytes
         Negotiated Transfer Speed                       : SAS 22.5 Gb/s
         Reported Channel,Device(T:L)                    : 0,1(1:0)
         Array                                           : 0
         Vendor                                          : KIOXIA
         Model                                           : TestSSD-3T-B
         Firmware                                        : 0001
         Serial number                                   : TSTSER0000002
         Total Size                                      : 3052360 MB
         Write Cache                                     : Disabled (write-through)
         SSD                                             : Yes

      Device #2
         Device is a Hard drive
         State                                           : Failed
         Block Size                                      : 0 Bytes
         Physical Block Size                             : 0 Bytes
         Reported Channel,Device(T:L)                    : 0,2(2:0)
         Array                                           : 0
         Vendor                                          : Not Available
         Model                                           : Not Available
         Firmware                                        : Not Available
         Serial number                                   : Not Available
         Total Size                                      : 0 MB
         Write Cache                                     : Unknown
         SSD                                             : No

      Device #3
         Device is a Hard drive
         State                                           : Online
         Block Size                                      : 512 Bytes
         Physical Block Size                             : 4K Bytes
         Negotiated Transfer Speed                       : SAS 22.5 Gb/s
         Reported Channel,Device(T:L)                    : 0,3(3:0)
         Array                                           : 0
         Vendor                                          : KIOXIA
         Model                                           : TestSSD-3T-B
         Firmware                                        : 0001
         Serial number                                   : TSTSER0000003
         Total Size                                      : 3052360 MB
         Write Cache                                     : Disabled (write-through)
         SSD                                             : Yes

      Device #7
         Device is a Hard drive
         State                                           : Ready
         Block Size                                      : 4K Bytes
         Physical Block Size                             : 4K Bytes
         Negotiated Transfer Speed                       : SAS 12.0 Gb/s
         Reported Channel,Device(T:L)                    : 0,7(7:0)
         Vendor                                          : TOSHIBA
         Model                                           : TestHDD-2T-A
         Firmware                                        : 5000
         Serial number                                   : TSTSER0000007
         Total Size                                      : 2289272 MB
         Write Cache                                     : Disabled (write-through)
         SSD                                             : No

----------------------------------------------------------------------
maxCache information
----------------------------------------------------------------------
   No maxCache Array found


Command completed successfully.
`

// TestNewControllerDegradedRAID5 verifies that:
//  1. Failed drives with "Block Size : 0" do not cause a parse error.
//  2. Unassigned drives (no "Array :" field) are not linked to any logical
//     drive, so IsSSD() correctly reflects only the drives in the array.
func TestNewControllerDegradedRAID5(t *testing.T) {
	ctrl, err := newController(1, arcConfGetConfigDegradedRAID5)
	require.NoError(t, err, "newController must not fail on degraded RAID5 output")

	// One logical drive.
	require.Len(t, ctrl.LogicalDrives, 1)
	ld := ctrl.LogicalDrives[0]
	require.NotNil(t, ld)
	assert.Equal(t, "/dev/sdc", ld.DiskName)
	assert.Equal(t, "5", ld.RAIDLevel)

	// The LD must only contain drives explicitly assigned to Array 0:
	// Device #1 (SSD), Device #2 (failed HDD), Device #3 (SSD).
	// Unassigned Device #0 (SSD) and Device #7 (HDD) must NOT be linked.
	assert.Len(t, ld.Devices, 3, "only Array-0 drives should be linked to LD 0")

	linkedIDs := make(map[int]bool)
	for _, pd := range ld.Devices {
		linkedIDs[pd.ID] = true
	}
	assert.True(t, linkedIDs[1], "Device #1 must be linked")
	assert.True(t, linkedIDs[2], "Device #2 (failed) must be linked")
	assert.True(t, linkedIDs[3], "Device #3 must be linked")
	assert.False(t, linkedIDs[0], "unassigned Device #0 must not be linked")
	assert.False(t, linkedIDs[7], "unassigned Device #7 must not be linked")

	// IsSSD() must be false because the failed drive (Device #2) reports HDD.
	assert.False(t, ld.IsSSD(), "IsSSD must be false: failed placeholder is HDD")

	// Physical drive inventory: all 5 devices should be present.
	assert.Len(t, ctrl.PhysicalDrives, 5)

	// Failed drive's BlockSize parsed as 0 — no error, field is 0.
	pd2 := ctrl.PhysicalDrives[2]
	require.NotNil(t, pd2)
	assert.Equal(t, 0, pd2.BlockSize)
	assert.Equal(t, 0, pd2.PhysicalBlockSize)
	assert.Equal(t, "Failed", pd2.Availability)

	// Unassigned drives have ArrayID -1.
	pd0 := ctrl.PhysicalDrives[0]
	require.NotNil(t, pd0)
	assert.Equal(t, -1, pd0.ArrayID, "unassigned drive ArrayID must be -1")

	pd7 := ctrl.PhysicalDrives[7]
	require.NotNil(t, pd7)
	assert.Equal(t, -1, pd7.ArrayID, "unassigned drive ArrayID must be -1")
}
