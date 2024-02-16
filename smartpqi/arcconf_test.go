package smartpqi

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
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
		`{"ArrayID":0,"Availability":"Online","BlockSize":512,"Channel":0,"ID":0,"Firmware":"5703","Model":"AL15SEB060N","PhysicalBlockSize":512,"Protocol":"","SerialNumber":"63M0A0BYFJPF","SizeMB":572325,"Type":"HDD","Vendor":"TOSHIBA","WriteCache":"Disabled (write-through)"}`,
		`{"ArrayID":1,"Availability":"Online","BlockSize":512,"Channel":0,"ID":1,"Firmware":"5701","Model":"AL15SEB120N","PhysicalBlockSize":512,"Protocol":"","SerialNumber":"59M0A06CFJRG","SizeMB":1144641,"Type":"HDD","Vendor":"TOSHIBA","WriteCache":"Disabled (write-through)"}`,
		`{"ArrayID":2,"Availability":"Online","BlockSize":512,"Channel":0,"ID":2,"Firmware":"CN03","Model":"ST1200MM0009","PhysicalBlockSize":512,"Protocol":"","SerialNumber":"WFK076DT0000E821CET2","SizeMB":1144641,"Type":"HDD","Vendor":"SEAGATE","WriteCache":"Disabled (write-through)"}`,
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

	expectedJSON := `{"ID":1,"PhysicalDrives":{"0":{"ArrayID":0,"Availability":"Online","BlockSize":512,"Channel":0,"ID":0,"Firmware":"5703","Model":"AL15SEB060N","PhysicalBlockSize":512,"Protocol":"","SerialNumber":"63M0A0BYFJPF","SizeMB":572325,"Type":"HDD","Vendor":"TOSHIBA","WriteCache":"Disabled (write-through)"},"1":{"ArrayID":1,"Availability":"Online","BlockSize":512,"Channel":0,"ID":1,"Firmware":"5701","Model":"AL15SEB120N","PhysicalBlockSize":512,"Protocol":"","SerialNumber":"59M0A06CFJRG","SizeMB":1144641,"Type":"HDD","Vendor":"TOSHIBA","WriteCache":"Disabled (write-through)"},"2":{"ArrayID":2,"Availability":"Online","BlockSize":512,"Channel":0,"ID":2,"Firmware":"CN03","Model":"ST1200MM0009","PhysicalBlockSize":512,"Protocol":"","SerialNumber":"WFK076DT0000E821CET2","SizeMB":1144641,"Type":"HDD","Vendor":"SEAGATE","WriteCache":"Disabled (write-through)"}},"LogicalDrives":{"0":{"ArrayID":0,"BlockSize":512,"Caching":"","Devices":[{"ArrayID":0,"Availability":"Online","BlockSize":512,"Channel":0,"ID":0,"Firmware":"5703","Model":"AL15SEB060N","PhysicalBlockSize":512,"Protocol":"","SerialNumber":"63M0A0BYFJPF","SizeMB":572325,"Type":"HDD","Vendor":"TOSHIBA","WriteCache":"Disabled (write-through)"}],"DiskName":"/dev/sdd","ID":0,"InterfaceType":"SCSI","Name":"Logical Drive 1","RAIDLevel":"0","SizeMB":572293},"1":{"ArrayID":1,"BlockSize":512,"Caching":"","Devices":[{"ArrayID":1,"Availability":"Online","BlockSize":512,"Channel":0,"ID":1,"Firmware":"5701","Model":"AL15SEB120N","PhysicalBlockSize":512,"Protocol":"","SerialNumber":"59M0A06CFJRG","SizeMB":1144641,"Type":"HDD","Vendor":"TOSHIBA","WriteCache":"Disabled (write-through)"}],"DiskName":"/dev/sde","ID":1,"InterfaceType":"SCSI","Name":"Logical Drive 2","RAIDLevel":"0","SizeMB":1144609},"2":{"ArrayID":2,"BlockSize":512,"Caching":"","Devices":[{"ArrayID":2,"Availability":"Online","BlockSize":512,"Channel":0,"ID":2,"Firmware":"CN03","Model":"ST1200MM0009","PhysicalBlockSize":512,"Protocol":"","SerialNumber":"WFK076DT0000E821CET2","SizeMB":1144641,"Type":"HDD","Vendor":"SEAGATE","WriteCache":"Disabled (write-through)"}],"DiskName":"/dev/sdc","ID":2,"InterfaceType":"SCSI","Name":"Logical Drive 3","RAIDLevel":"0","SizeMB":1144609}}}`
	expCtrl := Controller{}
	if err := json.Unmarshal([]byte(expectedJSON), &expCtrl); err != nil {
		t.Errorf("failed to unmarshal expected Controller JSON: %s", err)
	}

	if !reflect.DeepEqual(expCtrl, ctrl) {
		t.Errorf("controller differed:\n found: %#v\n expct: %#v ", ctrl, expCtrl)
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
