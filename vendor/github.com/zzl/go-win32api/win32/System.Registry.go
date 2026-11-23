package win32

import (
	"syscall"
	"unsafe"
)

type (
	HKEY = uintptr
)

const (
	HKEY_CLASSES_ROOT                               HKEY   = ^HKEY(0x7fffffff)
	HKEY_CURRENT_USER                               HKEY   = ^HKEY(0x7ffffffe)
	HKEY_LOCAL_MACHINE                              HKEY   = ^HKEY(0x7ffffffd)
	HKEY_USERS                                      HKEY   = ^HKEY(0x7ffffffc)
	HKEY_PERFORMANCE_DATA                           HKEY   = ^HKEY(0x7ffffffb)
	HKEY_PERFORMANCE_TEXT                           HKEY   = ^HKEY(0x7fffffaf)
	HKEY_PERFORMANCE_NLSTEXT                        HKEY   = ^HKEY(0x7fffff9f)
	HKEY_CURRENT_CONFIG                             HKEY   = ^HKEY(0x7ffffffa)
	HKEY_DYN_DATA                                   HKEY   = ^HKEY(0x7ffffff9)
	HKEY_CURRENT_USER_LOCAL_SETTINGS                HKEY   = ^HKEY(0x7ffffff8)
	RRF_SUBKEY_WOW6464KEY                           uint32 = 0x10000
	RRF_SUBKEY_WOW6432KEY                           uint32 = 0x20000
	RRF_WOW64_MASK                                  uint32 = 0x30000
	RRF_NOEXPAND                                    uint32 = 0x10000000
	RRF_ZEROONFAILURE                               uint32 = 0x20000000
	REG_PROCESS_APPKEY                              uint32 = 0x1
	REG_USE_CURRENT_SECURITY_CONTEXT                uint32 = 0x2
	PROVIDER_KEEPS_VALUE_LENGTH                     uint32 = 0x1
	REG_MUI_STRING_TRUNCATE                         uint32 = 0x1
	REG_SECURE_CONNECTION                           uint32 = 0x1
	REGSTR_KEY_CLASS                                string = "Class"
	REGSTR_KEY_CONFIG                               string = "Config"
	REGSTR_KEY_ENUM                                 string = "Enum"
	REGSTR_KEY_ROOTENUM                             string = "Root"
	REGSTR_KEY_BIOSENUM                             string = "BIOS"
	REGSTR_KEY_ACPIENUM                             string = "ACPI"
	REGSTR_KEY_PCMCIAENUM                           string = "PCMCIA"
	REGSTR_KEY_PCIENUM                              string = "PCI"
	REGSTR_KEY_VPOWERDENUM                          string = "VPOWERD"
	REGSTR_KEY_ISAENUM                              string = "ISAPnP"
	REGSTR_KEY_EISAENUM                             string = "EISA"
	REGSTR_KEY_LOGCONFIG                            string = "LogConfig"
	REGSTR_KEY_SYSTEMBOARD                          string = "*PNP0C01"
	REGSTR_KEY_APM                                  string = "*PNP0C05"
	REGSTR_KEY_INIUPDATE                            string = "IniUpdate"
	REG_KEY_INSTDEV                                 string = "Installed"
	REGSTR_KEY_DOSOPTCDROM                          string = "CD-ROM"
	REGSTR_KEY_DOSOPTMOUSE                          string = "MOUSE"
	REGSTR_KEY_KNOWNDOCKINGSTATES                   string = "Hardware Profiles"
	REGSTR_KEY_DEVICEPARAMETERS                     string = "Device Parameters"
	REGSTR_KEY_DRIVERPARAMETERS                     string = "Driver Parameters"
	REGSTR_DEFAULT_INSTANCE                         string = "0000"
	REGSTR_PATH_SETUP                               string = "Software\\Microsoft\\Windows\\CurrentVersion"
	REGSTR_PATH_DRIVERSIGN                          string = "Software\\Microsoft\\Driver Signing"
	REGSTR_PATH_NONDRIVERSIGN                       string = "Software\\Microsoft\\Non-Driver Signing"
	REGSTR_PATH_DRIVERSIGN_POLICY                   string = "Software\\Policies\\Microsoft\\Windows NT\\Driver Signing"
	REGSTR_PATH_NONDRIVERSIGN_POLICY                string = "Software\\Policies\\Microsoft\\Windows NT\\Non-Driver Signing"
	REGSTR_PATH_PIFCONVERT                          string = "Software\\Microsoft\\Windows\\CurrentVersion\\PIFConvert"
	REGSTR_PATH_MSDOSOPTS                           string = "Software\\Microsoft\\Windows\\CurrentVersion\\MS-DOSOptions"
	REGSTR_PATH_NOSUGGMSDOS                         string = "Software\\Microsoft\\Windows\\CurrentVersion\\NoMSDOSWarn"
	REGSTR_PATH_NEWDOSBOX                           string = "Software\\Microsoft\\Windows\\CurrentVersion\\MS-DOSSpecialConfig"
	REGSTR_PATH_RUNONCE                             string = "Software\\Microsoft\\Windows\\CurrentVersion\\RunOnce"
	REGSTR_PATH_RUNONCEEX                           string = "Software\\Microsoft\\Windows\\CurrentVersion\\RunOnceEx"
	REGSTR_PATH_RUN                                 string = "Software\\Microsoft\\Windows\\CurrentVersion\\Run"
	REGSTR_PATH_RUNSERVICESONCE                     string = "Software\\Microsoft\\Windows\\CurrentVersion\\RunServicesOnce"
	REGSTR_PATH_RUNSERVICES                         string = "Software\\Microsoft\\Windows\\CurrentVersion\\RunServices"
	REGSTR_PATH_EXPLORER                            string = "Software\\Microsoft\\Windows\\CurrentVersion\\Explorer"
	REGSTR_PATH_PROPERTYSYSTEM                      string = "Software\\Microsoft\\Windows\\CurrentVersion\\PropertySystem"
	REGSTR_PATH_DETECT                              string = "Software\\Microsoft\\Windows\\CurrentVersion\\Detect"
	REGSTR_PATH_APPPATHS                            string = "Software\\Microsoft\\Windows\\CurrentVersion\\App Paths"
	REGSTR_PATH_UNINSTALL                           string = "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall"
	REGSTR_PATH_REALMODENET                         string = "Software\\Microsoft\\Windows\\CurrentVersion\\Network\\Real Mode Net"
	REGSTR_PATH_NETEQUIV                            string = "Software\\Microsoft\\Windows\\CurrentVersion\\Network\\Equivalent"
	REGSTR_PATH_CVNETWORK                           string = "Software\\Microsoft\\Windows\\CurrentVersion\\Network"
	REGSTR_PATH_WMI_SECURITY                        string = "System\\CurrentControlSet\\Control\\Wmi\\Security"
	REGSTR_PATH_RELIABILITY                         string = "Software\\Microsoft\\Windows\\CurrentVersion\\Reliability"
	REGSTR_PATH_RELIABILITY_POLICY                  string = "Software\\Policies\\Microsoft\\Windows NT\\Reliability"
	REGSTR_PATH_RELIABILITY_POLICY_SHUTDOWNREASONUI string = "ShutdownReasonUI"
	REGSTR_PATH_RELIABILITY_POLICY_SNAPSHOT         string = "Snapshot"
	REGSTR_PATH_RELIABILITY_POLICY_REPORTSNAPSHOT   string = "ReportSnapshot"
	REGSTR_PATH_REINSTALL                           string = "SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Reinstall"
	REGSTR_PATH_NT_CURRENTVERSION                   string = "Software\\Microsoft\\Windows NT\\CurrentVersion"
	REGSTR_PATH_VOLUMECACHE                         string = "Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\VolumeCaches"
	REGSTR_VAL_DISPLAY                              string = "display"
	REGSTR_PATH_IDCONFIGDB                          string = "System\\CurrentControlSet\\Control\\IDConfigDB"
	REGSTR_PATH_CRITICALDEVICEDATABASE              string = "System\\CurrentControlSet\\Control\\CriticalDeviceDatabase"
	REGSTR_PATH_CLASS                               string = "System\\CurrentControlSet\\Services\\Class"
	REGSTR_PATH_DISPLAYSETTINGS                     string = "Display\\Settings"
	REGSTR_PATH_FONTS                               string = "Display\\Fonts"
	REGSTR_PATH_ENUM                                string = "Enum"
	REGSTR_PATH_ROOT                                string = "Enum\\Root"
	REGSTR_PATH_CURRENTCONTROLSET                   string = "System\\CurrentControlSet"
	REGSTR_PATH_SYSTEMENUM                          string = "System\\CurrentControlSet\\Enum"
	REGSTR_PATH_HWPROFILES                          string = "System\\CurrentControlSet\\Hardware Profiles"
	REGSTR_PATH_HWPROFILESCURRENT                   string = "System\\CurrentControlSet\\Hardware Profiles\\Current"
	REGSTR_PATH_CLASS_NT                            string = "System\\CurrentControlSet\\Control\\Class"
	REGSTR_PATH_PER_HW_ID_STORAGE                   string = "Software\\Microsoft\\Windows NT\\CurrentVersion\\PerHwIdStorage"
	REGSTR_PATH_DEVICE_CLASSES                      string = "System\\CurrentControlSet\\Control\\DeviceClasses"
	REGSTR_PATH_CODEVICEINSTALLERS                  string = "System\\CurrentControlSet\\Control\\CoDeviceInstallers"
	REGSTR_PATH_BUSINFORMATION                      string = "System\\CurrentControlSet\\Control\\PnP\\BusInformation"
	REGSTR_PATH_SERVICES                            string = "System\\CurrentControlSet\\Services"
	REGSTR_PATH_VXD                                 string = "System\\CurrentControlSet\\Services\\VxD"
	REGSTR_PATH_IOS                                 string = "System\\CurrentControlSet\\Services\\VxD\\IOS"
	REGSTR_PATH_VMM                                 string = "System\\CurrentControlSet\\Services\\VxD\\VMM"
	REGSTR_PATH_VPOWERD                             string = "System\\CurrentControlSet\\Services\\VxD\\VPOWERD"
	REGSTR_PATH_VNETSUP                             string = "System\\CurrentControlSet\\Services\\VxD\\VNETSUP"
	REGSTR_PATH_NWREDIR                             string = "System\\CurrentControlSet\\Services\\VxD\\NWREDIR"
	REGSTR_PATH_NCPSERVER                           string = "System\\CurrentControlSet\\Services\\NcpServer\\Parameters"
	REGSTR_PATH_VCOMM                               string = "System\\CurrentControlSet\\Services\\VxD\\VCOMM"
	REGSTR_PATH_IOARB                               string = "System\\CurrentControlSet\\Services\\Arbitrators\\IOArb"
	REGSTR_PATH_ADDRARB                             string = "System\\CurrentControlSet\\Services\\Arbitrators\\AddrArb"
	REGSTR_PATH_DMAARB                              string = "System\\CurrentControlSet\\Services\\Arbitrators\\DMAArb"
	REGSTR_PATH_IRQARB                              string = "System\\CurrentControlSet\\Services\\Arbitrators\\IRQArb"
	REGSTR_PATH_CODEPAGE                            string = "System\\CurrentControlSet\\Control\\Nls\\Codepage"
	REGSTR_PATH_FILESYSTEM                          string = "System\\CurrentControlSet\\Control\\FileSystem"
	REGSTR_PATH_FILESYSTEM_NOVOLTRACK               string = "System\\CurrentControlSet\\Control\\FileSystem\\NoVolTrack"
	REGSTR_PATH_CDFS                                string = "System\\CurrentControlSet\\Control\\FileSystem\\CDFS"
	REGSTR_PATH_WINBOOT                             string = "System\\CurrentControlSet\\Control\\WinBoot"
	REGSTR_PATH_INSTALLEDFILES                      string = "System\\CurrentControlSet\\Control\\InstalledFiles"
	REGSTR_PATH_VMM32FILES                          string = "System\\CurrentControlSet\\Control\\VMM32Files"
	REGSTR_MAX_VALUE_LENGTH                         uint32 = 0x100
	REGSTR_KEY_DEVICE_PROPERTIES                    string = "Properties"
	REGSTR_VAL_SERVICE                              string = "Service"
	REGSTR_VAL_CLASSGUID                            string = "ClassGUID"
	REGSTR_VAL_DISABLECOUNT                         string = "DisableCount"
	REGSTR_VAL_DOCKSTATE                            string = "DockState"
	REGSTR_VAL_DEVICE_INSTANCE                      string = "DeviceInstance"
	REGSTR_VAL_SYMBOLIC_LINK                        string = "SymbolicLink"
	REGSTR_VAL_DEFAULT                              string = "Default"
	REGSTR_VAL_LOWERFILTERS                         string = "LowerFilters"
	REGSTR_VAL_UPPERFILTERS                         string = "UpperFilters"
	REGSTR_VAL_LOCATION_INFORMATION                 string = "LocationInformation"
	REGSTR_VAL_UI_NUMBER                            string = "UINumber"
	REGSTR_VAL_UI_NUMBER_DESC_FORMAT                string = "UINumberDescFormat"
	REGSTR_VAL_CAPABILITIES                         string = "Capabilities"
	REGSTR_VAL_ADDRESS                              string = "Address"
	REGSTR_VAL_DEVICE_TYPE                          string = "DeviceType"
	REGSTR_VAL_DEVICE_CHARACTERISTICS               string = "DeviceCharacteristics"
	REGSTR_VAL_DEVICE_SECURITY_DESCRIPTOR           string = "Security"
	REGSTR_VAL_DEVICE_EXCLUSIVE                     string = "Exclusive"
	REGSTR_VAL_RESOURCE_PICKER_TAGS                 string = "ResourcePickerTags"
	REGSTR_VAL_RESOURCE_PICKER_EXCEPTIONS           string = "ResourcePickerExceptions"
	REGSTR_VAL_CUSTOM_PROPERTY_CACHE_DATE           string = "CustomPropertyCacheDate"
	REGSTR_VAL_CUSTOM_PROPERTY_HW_ID_KEY            string = "CustomPropertyHwIdKey"
	REGSTR_VAL_LAST_UPDATE_TIME                     string = "LastUpdateTime"
	REGSTR_VAL_CONTAINERID                          string = "ContainerID"
	REGSTR_VAL_EJECT_PRIORITY                       string = "EjectPriority"
	REGSTR_KEY_CONTROL                              string = "Control"
	REGSTR_VAL_ACTIVESERVICE                        string = "ActiveService"
	REGSTR_VAL_LINKED                               string = "Linked"
	REGSTR_VAL_PHYSICALDEVICEOBJECT                 string = "PhysicalDeviceObject"
	REGSTR_VAL_REMOVAL_POLICY                       string = "RemovalPolicy"
	REGSTR_KEY_FILTERS                              string = "Filters"
	REGSTR_VAL_LOWER_FILTER_DEFAULT_LEVEL           string = "LowerFilterDefaultLevel"
	REGSTR_VAL_UPPER_FILTER_DEFAULT_LEVEL           string = "UpperFilterDefaultLevel"
	REGSTR_KEY_LOWER_FILTER_LEVEL_DEFAULT           string = "*Lower"
	REGSTR_KEY_UPPER_FILTER_LEVEL_DEFAULT           string = "*Upper"
	REGSTR_VAL_LOWER_FILTER_LEVELS                  string = "LowerFilterLevels"
	REGSTR_VAL_UPPER_FILTER_LEVELS                  string = "UpperFilterLevels"
	REGSTR_VAL_CURRENT_VERSION                      string = "CurrentVersion"
	REGSTR_VAL_CURRENT_BUILD                        string = "CurrentBuildNumber"
	REGSTR_VAL_CURRENT_CSDVERSION                   string = "CSDVersion"
	REGSTR_VAL_CURRENT_TYPE                         string = "CurrentType"
	REGSTR_VAL_BITSPERPIXEL                         string = "BitsPerPixel"
	REGSTR_VAL_RESOLUTION                           string = "Resolution"
	REGSTR_VAL_DPILOGICALX                          string = "DPILogicalX"
	REGSTR_VAL_DPILOGICALY                          string = "DPILogicalY"
	REGSTR_VAL_DPIPHYSICALX                         string = "DPIPhysicalX"
	REGSTR_VAL_DPIPHYSICALY                         string = "DPIPhysicalY"
	REGSTR_VAL_REFRESHRATE                          string = "RefreshRate"
	REGSTR_VAL_DISPLAYFLAGS                         string = "DisplayFlags"
	REGSTR_PATH_CONTROLPANEL                        string = "Control Panel"
	REGSTR_PATH_CONTROLSFOLDER                      string = "Software\\Microsoft\\Windows\\CurrentVersion\\Controls Folder"
	REGSTR_VAL_DOSCP                                string = "OEMCP"
	REGSTR_VAL_WINCP                                string = "ACP"
	REGSTR_VAL_DONTUSEMEM                           string = "DontAllocLastMem"
	REGSTR_VAL_SYSTEMROOT                           string = "SystemRoot"
	REGSTR_VAL_BOOTCOUNT                            string = "BootCount"
	REGSTR_VAL_REALNETSTART                         string = "RealNetStart"
	REGSTR_VAL_MEDIA                                string = "MediaPath"
	REGSTR_VAL_CONFIG                               string = "ConfigPath"
	REGSTR_VAL_DEVICEPATH                           string = "DevicePath"
	REGSTR_VAL_SRCPATH                              string = "SourcePath"
	REGSTR_VAL_DRIVERCACHEPATH                      string = "DriverCachePath"
	REGSTR_VAL_OLDWINDIR                            string = "OldWinDir"
	REGSTR_VAL_SETUPFLAGS                           string = "SetupFlags"
	REGSTR_VAL_REGOWNER                             string = "RegisteredOwner"
	REGSTR_VAL_REGORGANIZATION                      string = "RegisteredOrganization"
	REGSTR_VAL_LICENSINGINFO                        string = "LicensingInfo"
	REGSTR_VAL_OLDMSDOSVER                          string = "OldMSDOSVer"
	REGSTR_VAL_FIRSTINSTALLDATETIME                 string = "FirstInstallDateTime"
	REGSTR_VAL_INSTALLTYPE                          string = "InstallType"
	IT_COMPACT                                      uint32 = 0x0
	IT_TYPICAL                                      uint32 = 0x1
	IT_PORTABLE                                     uint32 = 0x2
	IT_CUSTOM                                       uint32 = 0x3
	REGSTR_VAL_WRAPPER                              string = "Wrapper"
	REGSTR_KEY_RUNHISTORY                           string = "RunHistory"
	REGSTR_VAL_LASTALIVEINTERVAL                    string = "TimeStampInterval"
	REGSTR_VAL_DIRTYSHUTDOWN                        string = "DirtyShutdown"
	REGSTR_VAL_DIRTYSHUTDOWNTIME                    string = "DirtyShutdownTime"
	REGSTR_VAL_BT                                   string = "6005BT"
	REGSTR_VAL_LASTCOMPUTERNAME                     string = "LastComputerName"
	REGSTR_VAL_LASTALIVEBT                          string = "LastAliveBT"
	REGSTR_VAL_LASTALIVESTAMP                       string = "LastAliveStamp"
	REGSTR_VAL_LASTALIVESTAMPFORCED                 string = "LastAliveStampForced"
	REGSTR_VAL_LASTALIVESTAMPINTERVAL               string = "LastAliveStampInterval"
	REGSTR_VAL_LASTALIVESTAMPPOLICYINTERVAL         string = "LastAliveStampPolicyInterval"
	REGSTR_VAL_LASTALIVEUPTIME                      string = "LastAliveUptime"
	REGSTR_VAL_LASTALIVEPMPOLICY                    string = "LastAlivePMPolicy"
	REGSTR_VAL_REASONCODE                           string = "ReasonCode"
	REGSTR_VAL_COMMENT                              string = "Comment"
	REGSTR_VAL_SHUTDOWNREASON                       string = "ShutdownReason"
	REGSTR_VAL_SHUTDOWNREASON_CODE                  string = "ShutdownReasonCode"
	REGSTR_VAL_SHUTDOWNREASON_COMMENT               string = "ShutdownReasonComment"
	REGSTR_VAL_SHUTDOWNREASON_PROCESS               string = "ShutdownReasonProcess"
	REGSTR_VAL_SHUTDOWNREASON_USERNAME              string = "ShutdownReasonUserName"
	REGSTR_VAL_SHOWREASONUI                         string = "ShutdownReasonUI"
	REGSTR_VAL_SHUTDOWN_IGNORE_PREDEFINED           string = "ShutdownIgnorePredefinedReasons"
	REGSTR_VAL_SHUTDOWN_STATE_SNAPSHOT              string = "ShutdownStateSnapshot"
	REGSTR_KEY_SETUP                                string = "\\Setup"
	REGSTR_VAL_BOOTDIR                              string = "BootDir"
	REGSTR_VAL_WINBOOTDIR                           string = "WinbootDir"
	REGSTR_VAL_WINDIR                               string = "WinDir"
	REGSTR_VAL_APPINSTPATH                          string = "AppInstallPath"
	REGSTR_KEY_EBDFILESLOCAL                        string = "EBDFilesLocale"
	REGSTR_KEY_EBDFILESKEYBOARD                     string = "EBDFilesKeyboard"
	REGSTR_KEY_EBDAUTOEXECBATLOCAL                  string = "EBDAutoexecBatLocale"
	REGSTR_KEY_EBDAUTOEXECBATKEYBOARD               string = "EBDAutoexecBatKeyboard"
	REGSTR_KEY_EBDCONFIGSYSLOCAL                    string = "EBDConfigSysLocale"
	REGSTR_KEY_EBDCONFIGSYSKEYBOARD                 string = "EBDConfigSysKeyboard"
	REGSTR_VAL_POLICY                               string = "Policy"
	REGSTR_VAL_BEHAVIOR_ON_FAILED_VERIFY            string = "BehaviorOnFailedVerify"
	DRIVERSIGN_NONE                                 uint32 = 0x0
	DRIVERSIGN_WARNING                              uint32 = 0x1
	DRIVERSIGN_BLOCKING                             uint32 = 0x2
	REGSTR_VAL_MSDOSMODE                            string = "MSDOSMode"
	REGSTR_VAL_MSDOSMODEDISCARD                     string = "Discard"
	REGSTR_VAL_DOSOPTGLOBALFLAGS                    string = "GlobalFlags"
	DOSOPTGF_DEFCLEAN                               int32  = 1
	REGSTR_VAL_DOSOPTFLAGS                          string = "Flags"
	REGSTR_VAL_OPTORDER                             string = "Order"
	REGSTR_VAL_CONFIGSYS                            string = "Config.Sys"
	REGSTR_VAL_AUTOEXEC                             string = "Autoexec.Bat"
	REGSTR_VAL_STDDOSOPTION                         string = "StdOption"
	REGSTR_VAL_DOSOPTTIP                            string = "TipText"
	DOSOPTF_DEFAULT                                 int32  = 1
	DOSOPTF_SUPPORTED                               int32  = 2
	DOSOPTF_ALWAYSUSE                               int32  = 4
	DOSOPTF_USESPMODE                               int32  = 8
	DOSOPTF_PROVIDESUMB                             int32  = 16
	DOSOPTF_NEEDSETUP                               int32  = 32
	DOSOPTF_INDOSSTART                              int32  = 64
	DOSOPTF_MULTIPLE                                int32  = 128
	SUF_FIRSTTIME                                   int32  = 1
	SUF_EXPRESS                                     int32  = 2
	SUF_BATCHINF                                    int32  = 4
	SUF_CLEAN                                       int32  = 8
	SUF_INSETUP                                     int32  = 16
	SUF_NETSETUP                                    int32  = 32
	SUF_NETHDBOOT                                   int32  = 64
	SUF_NETRPLBOOT                                  int32  = 128
	SUF_SBSCOPYOK                                   int32  = 256
	REGSTR_VAL_DOSPAGER                             string = "DOSPager"
	REGSTR_VAL_VXDGROUPS                            string = "VXDGroups"
	REGSTR_VAL_VPOWERDFLAGS                         string = "Flags"
	VPDF_DISABLEPWRMGMT                             uint32 = 0x1
	VPDF_FORCEAPM10MODE                             uint32 = 0x2
	VPDF_SKIPINTELSLCHECK                           uint32 = 0x4
	VPDF_DISABLEPWRSTATUSPOLL                       uint32 = 0x8
	VPDF_DISABLERINGRESUME                          uint32 = 0x10
	VPDF_SHOWMULTIBATT                              uint32 = 0x20
	BIF_SHOWSIMILARDRIVERS                          uint32 = 0x1
	BIF_RAWDEVICENEEDSDRIVER                        uint32 = 0x2
	REGSTR_VAL_WORKGROUP                            string = "Workgroup"
	REGSTR_VAL_DIRECTHOST                           string = "DirectHost"
	REGSTR_VAL_FILESHARING                          string = "FileSharing"
	REGSTR_VAL_PRINTSHARING                         string = "PrintSharing"
	REGSTR_VAL_FIRSTNETDRIVE                        string = "FirstNetworkDrive"
	REGSTR_VAL_MAXCONNECTIONS                       string = "MaxConnections"
	REGSTR_VAL_APISUPPORT                           string = "APISupport"
	REGSTR_VAL_MAXRETRY                             string = "MaxRetry"
	REGSTR_VAL_MINRETRY                             string = "MinRetry"
	REGSTR_VAL_SUPPORTLFN                           string = "SupportLFN"
	REGSTR_VAL_SUPPORTBURST                         string = "SupportBurst"
	REGSTR_VAL_SUPPORTTUNNELLING                    string = "SupportTunnelling"
	REGSTR_VAL_FULLTRACE                            string = "FullTrace"
	REGSTR_VAL_READCACHING                          string = "ReadCaching"
	REGSTR_VAL_SHOWDOTS                             string = "ShowDots"
	REGSTR_VAL_GAPTIME                              string = "GapTime"
	REGSTR_VAL_SEARCHMODE                           string = "SearchMode"
	REGSTR_VAL_SHELLVERSION                         string = "ShellVersion"
	REGSTR_VAL_MAXLIP                               string = "MaxLIP"
	REGSTR_VAL_PRESERVECASE                         string = "PreserveCase"
	REGSTR_VAL_OPTIMIZESFN                          string = "OptimizeSFN"
	REGSTR_VAL_NCP_BROWSEMASTER                     string = "BrowseMaster"
	REGSTR_VAL_NCP_USEPEERBROWSING                  string = "Use_PeerBrowsing"
	REGSTR_VAL_NCP_USESAP                           string = "Use_Sap"
	REGSTR_VAL_PCCARD_POWER                         string = "EnablePowerManagement"
	REGSTR_VAL_WIN31FILESYSTEM                      string = "Win31FileSystem"
	REGSTR_VAL_PRESERVELONGNAMES                    string = "PreserveLongNames"
	REGSTR_VAL_DRIVEWRITEBEHIND                     string = "DriveWriteBehind"
	REGSTR_VAL_ASYNCFILECOMMIT                      string = "AsyncFileCommit"
	REGSTR_VAL_PATHCACHECOUNT                       string = "PathCache"
	REGSTR_VAL_NAMECACHECOUNT                       string = "NameCache"
	REGSTR_VAL_CONTIGFILEALLOC                      string = "ContigFileAllocSize"
	REGSTR_VAL_FREESPACERATIO                       string = "FreeSpaceRatio"
	REGSTR_VAL_VOLIDLETIMEOUT                       string = "VolumeIdleTimeout"
	REGSTR_VAL_BUFFIDLETIMEOUT                      string = "BufferIdleTimeout"
	REGSTR_VAL_BUFFAGETIMEOUT                       string = "BufferAgeTimeout"
	REGSTR_VAL_NAMENUMERICTAIL                      string = "NameNumericTail"
	REGSTR_VAL_READAHEADTHRESHOLD                   string = "ReadAheadThreshold"
	REGSTR_VAL_DOUBLEBUFFER                         string = "DoubleBuffer"
	REGSTR_VAL_SOFTCOMPATMODE                       string = "SoftCompatMode"
	REGSTR_VAL_DRIVESPINDOWN                        string = "DriveSpinDown"
	REGSTR_VAL_FORCEPMIO                            string = "ForcePMIO"
	REGSTR_VAL_FORCERMIO                            string = "ForceRMIO"
	REGSTR_VAL_LASTBOOTPMDRVS                       string = "LastBootPMDrvs"
	REGSTR_VAL_ACSPINDOWNPREVIOUS                   string = "ACSpinDownPrevious"
	REGSTR_VAL_BATSPINDOWNPREVIOUS                  string = "BatSpinDownPrevious"
	REGSTR_VAL_VIRTUALHDIRQ                         string = "VirtualHDIRQ"
	REGSTR_VAL_SRVNAMECACHECOUNT                    string = "ServerNameCacheMax"
	REGSTR_VAL_SRVNAMECACHE                         string = "ServerNameCache"
	REGSTR_VAL_SRVNAMECACHENETPROV                  string = "ServerNameCacheNumNets"
	REGSTR_VAL_AUTOMOUNT                            string = "AutoMountDrives"
	REGSTR_VAL_COMPRESSIONMETHOD                    string = "CompressionAlgorithm"
	REGSTR_VAL_COMPRESSIONTHRESHOLD                 string = "CompressionThreshold"
	REGSTR_VAL_ACDRIVESPINDOWN                      string = "ACDriveSpinDown"
	REGSTR_VAL_BATDRIVESPINDOWN                     string = "BatDriveSpinDown"
	REGSTR_VAL_CDCACHESIZE                          string = "CacheSize"
	REGSTR_VAL_CDPREFETCH                           string = "Prefetch"
	REGSTR_VAL_CDPREFETCHTAIL                       string = "PrefetchTail"
	REGSTR_VAL_CDRAWCACHE                           string = "RawCache"
	REGSTR_VAL_CDEXTERRORS                          string = "ExtendedErrors"
	REGSTR_VAL_CDSVDSENSE                           string = "SVDSense"
	REGSTR_VAL_CDSHOWVERSIONS                       string = "ShowVersions"
	REGSTR_VAL_CDCOMPATNAMES                        string = "MSCDEXCompatNames"
	REGSTR_VAL_CDNOREADAHEAD                        string = "NoReadAhead"
	REGSTR_VAL_SCSI                                 string = "SCSI\\"
	REGSTR_VAL_ESDI                                 string = "ESDI\\"
	REGSTR_VAL_FLOP                                 string = "FLOP\\"
	REGSTR_VAL_DISK                                 string = "GenDisk"
	REGSTR_VAL_CDROM                                string = "GenCD"
	REGSTR_VAL_TAPE                                 string = "TAPE"
	REGSTR_VAL_SCANNER                              string = "SCANNER"
	REGSTR_VAL_FLOPPY                               string = "FLOPPY"
	REGSTR_VAL_SCSITID                              string = "SCSITargetID"
	REGSTR_VAL_SCSILUN                              string = "SCSILUN"
	REGSTR_VAL_REVLEVEL                             string = "RevisionLevel"
	REGSTR_VAL_PRODUCTID                            string = "ProductId"
	REGSTR_VAL_PRODUCTTYPE                          string = "ProductType"
	REGSTR_VAL_DEVTYPE                              string = "DeviceType"
	REGSTR_VAL_REMOVABLE                            string = "Removable"
	REGSTR_VAL_CURDRVLET                            string = "CurrentDriveLetterAssignment"
	REGSTR_VAL_USRDRVLET                            string = "UserDriveLetterAssignment"
	REGSTR_VAL_SYNCDATAXFER                         string = "SyncDataXfer"
	REGSTR_VAL_AUTOINSNOTE                          string = "AutoInsertNotification"
	REGSTR_VAL_DISCONNECT                           string = "Disconnect"
	REGSTR_VAL_INT13                                string = "Int13"
	REGSTR_VAL_PMODE_INT13                          string = "PModeInt13"
	REGSTR_VAL_USERSETTINGS                         string = "AdapterSettings"
	REGSTR_VAL_NOIDE                                string = "NoIDE"
	REGSTR_VAL_DISKCLASSNAME                        string = "DiskDrive"
	REGSTR_VAL_CDROMCLASSNAME                       string = "CDROM"
	REGSTR_VAL_FORCELOAD                            string = "ForceLoadPD"
	REGSTR_VAL_FORCEFIFO                            string = "ForceFIFO"
	REGSTR_VAL_FORCECL                              string = "ForceChangeLine"
	REGSTR_VAL_NOUSECLASS                           string = "NoUseClass"
	REGSTR_VAL_NOINSTALLCLASS                       string = "NoInstallClass"
	REGSTR_VAL_NODISPLAYCLASS                       string = "NoDisplayClass"
	REGSTR_VAL_SILENTINSTALL                        string = "SilentInstall"
	REGSTR_VAL_FSFILTERCLASS                        string = "FSFilterClass"
	REGSTR_KEY_PCMCIA_CLASS                         string = "PCMCIA"
	REGSTR_KEY_SCSI_CLASS                           string = "SCSIAdapter"
	REGSTR_KEY_PORTS_CLASS                          string = "ports"
	REGSTR_KEY_MEDIA_CLASS                          string = "MEDIA"
	REGSTR_KEY_DISPLAY_CLASS                        string = "Display"
	REGSTR_KEY_KEYBOARD_CLASS                       string = "Keyboard"
	REGSTR_KEY_MOUSE_CLASS                          string = "Mouse"
	REGSTR_KEY_MONITOR_CLASS                        string = "Monitor"
	REGSTR_KEY_MODEM_CLASS                          string = "Modem"
	REGSTR_VAL_PCMCIA_OPT                           string = "Options"
	PCMCIA_OPT_HAVE_SOCKET                          int32  = 1
	PCMCIA_OPT_AUTOMEM                              int32  = 4
	PCMCIA_OPT_NO_SOUND                             int32  = 8
	PCMCIA_OPT_NO_AUDIO                             int32  = 16
	PCMCIA_OPT_NO_APMREMOVE                         int32  = 32
	REGSTR_VAL_PCMCIA_MEM                           string = "Memory"
	PCMCIA_DEF_MEMBEGIN                             uint32 = 0xc0000
	PCMCIA_DEF_MEMEND                               uint32 = 0xffffff
	PCMCIA_DEF_MEMLEN                               uint32 = 0x1000
	REGSTR_VAL_PCMCIA_ALLOC                         string = "AllocMemWin"
	REGSTR_VAL_PCMCIA_ATAD                          string = "ATADelay"
	REGSTR_VAL_PCMCIA_SIZ                           string = "MinRegionSize"
	PCMCIA_DEF_MIN_REGION                           uint32 = 0x10000
	REGSTR_VAL_P1284MDL                             string = "Model"
	REGSTR_VAL_P1284MFG                             string = "Manufacturer"
	REGSTR_VAL_ISAPNP                               string = "ISAPNP"
	REGSTR_VAL_ISAPNP_RDP_OVERRIDE                  string = "RDPOverRide"
	REGSTR_VAL_PCI                                  string = "PCI"
	REGSTR_PCI_OPTIONS                              string = "Options"
	REGSTR_PCI_DUAL_IDE                             string = "PCIDualIDE"
	PCI_OPTIONS_USE_BIOS                            int32  = 1
	PCI_OPTIONS_USE_IRQ_STEERING                    int32  = 2
	AGP_FLAG_NO_1X_RATE                             int32  = 1
	AGP_FLAG_NO_2X_RATE                             int32  = 2
	AGP_FLAG_NO_4X_RATE                             int32  = 4
	AGP_FLAG_NO_8X_RATE                             int32  = 8
	AGP_FLAG_REVERSE_INITIALIZATION                 int32  = 128
	AGP_FLAG_NO_SBA_ENABLE                          int32  = 256
	AGP_FLAG_NO_FW_ENABLE                           int32  = 512
	AGP_FLAG_SPECIAL_TARGET                         int32  = 1048575
	AGP_FLAG_SPECIAL_RESERVE                        int32  = 1015808
	REGSTR_KEY_CRASHES                              string = "Crashes"
	REGSTR_KEY_DANGERS                              string = "Dangers"
	REGSTR_KEY_DETMODVARS                           string = "DetModVars"
	REGSTR_KEY_NDISINFO                             string = "NDISInfo"
	REGSTR_VAL_PROTINIPATH                          string = "ProtIniPath"
	REGSTR_VAL_RESOURCES                            string = "Resources"
	REGSTR_VAL_CRASHFUNCS                           string = "CrashFuncs"
	REGSTR_VAL_CLASS                                string = "Class"
	REGSTR_VAL_CLASSDESC                            string = "ClassDesc"
	REGSTR_VAL_DEVDESC                              string = "DeviceDesc"
	REGSTR_VAL_BOOTCONFIG                           string = "BootConfig"
	REGSTR_VAL_DETFUNC                              string = "DetFunc"
	REGSTR_VAL_DETFLAGS                             string = "DetFlags"
	REGSTR_VAL_COMPATIBLEIDS                        string = "CompatibleIDs"
	REGSTR_VAL_DETCONFIG                            string = "DetConfig"
	REGSTR_VAL_VERIFYKEY                            string = "VerifyKey"
	REGSTR_VAL_COMINFO                              string = "ComInfo"
	REGSTR_VAL_INFNAME                              string = "InfName"
	REGSTR_VAL_CARDSPECIFIC                         string = "CardSpecific"
	REGSTR_VAL_NETOSTYPE                            string = "NetOSType"
	REGSTR_DATA_NETOS_NDIS                          string = "NDIS"
	REGSTR_DATA_NETOS_ODI                           string = "ODI"
	REGSTR_DATA_NETOS_IPX                           string = "IPX"
	REGSTR_VAL_MFG                                  string = "Mfg"
	REGSTR_VAL_SCAN_ONLY_FIRST                      string = "ScanOnlyFirstDrive"
	REGSTR_VAL_SHARE_IRQ                            string = "ForceIRQSharing"
	REGSTR_VAL_NONSTANDARD_ATAPI                    string = "NonStandardATAPI"
	REGSTR_VAL_IDE_FORCE_SERIALIZE                  string = "ForceSerialization"
	REGSTR_VAL_MAX_HCID_LEN                         uint32 = 0x400
	REGSTR_VAL_HWREV                                string = "HWRevision"
	REGSTR_VAL_ENABLEINTS                           string = "EnableInts"
	REGDF_NOTDETIO                                  uint32 = 0x1
	REGDF_NOTDETMEM                                 uint32 = 0x2
	REGDF_NOTDETIRQ                                 uint32 = 0x4
	REGDF_NOTDETDMA                                 uint32 = 0x8
	REGDF_NEEDFULLCONFIG                            uint32 = 0x10
	REGDF_GENFORCEDCONFIG                           uint32 = 0x20
	REGDF_NODETCONFIG                               uint32 = 0x8000
	REGDF_CONFLICTIO                                uint32 = 0x10000
	REGDF_CONFLICTMEM                               uint32 = 0x20000
	REGDF_CONFLICTIRQ                               uint32 = 0x40000
	REGDF_CONFLICTDMA                               uint32 = 0x80000
	REGDF_MAPIRQ2TO9                                uint32 = 0x100000
	REGDF_NOTVERIFIED                               uint32 = 0x80000000
	REGSTR_VAL_APMBIOSVER                           string = "APMBiosVer"
	REGSTR_VAL_APMFLAGS                             string = "APMFlags"
	REGSTR_VAL_SLSUPPORT                            string = "SLSupport"
	REGSTR_VAL_MACHINETYPE                          string = "MachineType"
	REGSTR_VAL_SETUPMACHINETYPE                     string = "SetupMachineType"
	REGSTR_MACHTYPE_UNKNOWN                         string = "Unknown"
	REGSTR_MACHTYPE_IBMPC                           string = "IBM PC"
	REGSTR_MACHTYPE_IBMPCJR                         string = "IBM PCjr"
	REGSTR_MACHTYPE_IBMPCCONV                       string = "IBM PC Convertible"
	REGSTR_MACHTYPE_IBMPCXT                         string = "IBM PC/XT"
	REGSTR_MACHTYPE_IBMPCXT_286                     string = "IBM PC/XT 286"
	REGSTR_MACHTYPE_IBMPCAT                         string = "IBM PC/AT"
	REGSTR_MACHTYPE_IBMPS2_25                       string = "IBM PS/2-25"
	REGSTR_MACHTYPE_IBMPS2_30_286                   string = "IBM PS/2-30 286"
	REGSTR_MACHTYPE_IBMPS2_30                       string = "IBM PS/2-30"
	REGSTR_MACHTYPE_IBMPS2_50                       string = "IBM PS/2-50"
	REGSTR_MACHTYPE_IBMPS2_50Z                      string = "IBM PS/2-50Z"
	REGSTR_MACHTYPE_IBMPS2_55SX                     string = "IBM PS/2-55SX"
	REGSTR_MACHTYPE_IBMPS2_60                       string = "IBM PS/2-60"
	REGSTR_MACHTYPE_IBMPS2_65SX                     string = "IBM PS/2-65SX"
	REGSTR_MACHTYPE_IBMPS2_70                       string = "IBM PS/2-70"
	REGSTR_MACHTYPE_IBMPS2_P70                      string = "IBM PS/2-P70"
	REGSTR_MACHTYPE_IBMPS2_70_80                    string = "IBM PS/2-70/80"
	REGSTR_MACHTYPE_IBMPS2_80                       string = "IBM PS/2-80"
	REGSTR_MACHTYPE_IBMPS2_90                       string = "IBM PS/2-90"
	REGSTR_MACHTYPE_IBMPS1                          string = "IBM PS/1"
	REGSTR_MACHTYPE_PHOENIX_PCAT                    string = "Phoenix PC/AT Compatible"
	REGSTR_MACHTYPE_HP_VECTRA                       string = "HP Vectra"
	REGSTR_MACHTYPE_ATT_PC                          string = "AT&T PC"
	REGSTR_MACHTYPE_ZENITH_PC                       string = "Zenith PC"
	REGSTR_VAL_APMMENUSUSPEND                       string = "APMMenuSuspend"
	APMMENUSUSPEND_DISABLED                         uint32 = 0x0
	APMMENUSUSPEND_ENABLED                          uint32 = 0x1
	APMMENUSUSPEND_UNDOCKED                         uint32 = 0x2
	APMMENUSUSPEND_NOCHANGE                         uint32 = 0x80
	REGSTR_VAL_APMACTIMEOUT                         string = "APMACTimeout"
	REGSTR_VAL_APMBATTIMEOUT                        string = "APMBatTimeout"
	APMTIMEOUT_DISABLED                             uint32 = 0x0
	REGSTR_VAL_APMSHUTDOWNPOWER                     string = "APMShutDownPower"
	REGSTR_VAL_BUSTYPE                              string = "BusType"
	REGSTR_VAL_CPU                                  string = "CPU"
	REGSTR_VAL_NDP                                  string = "NDP"
	REGSTR_VAL_PNPBIOSVER                           string = "PnPBIOSVer"
	REGSTR_VAL_PNPSTRUCOFFSET                       string = "PnPStrucOffset"
	REGSTR_VAL_PCIBIOSVER                           string = "PCIBIOSVer"
	REGSTR_VAL_HWMECHANISM                          string = "HWMechanism"
	REGSTR_VAL_LASTPCIBUSNUM                        string = "LastPCIBusNum"
	REGSTR_VAL_CONVMEM                              string = "ConvMem"
	REGSTR_VAL_EXTMEM                               string = "ExtMem"
	REGSTR_VAL_COMPUTERNAME                         string = "ComputerName"
	REGSTR_VAL_BIOSNAME                             string = "BIOSName"
	REGSTR_VAL_BIOSVERSION                          string = "BIOSVersion"
	REGSTR_VAL_BIOSDATE                             string = "BIOSDate"
	REGSTR_VAL_MODEL                                string = "Model"
	REGSTR_VAL_SUBMODEL                             string = "Submodel"
	REGSTR_VAL_REVISION                             string = "Revision"
	REGSTR_VAL_FIFODEPTH                            string = "FIFODepth"
	REGSTR_VAL_RDINTTHRESHOLD                       string = "RDIntThreshold"
	REGSTR_VAL_WRINTTHRESHOLD                       string = "WRIntThreshold"
	REGSTR_VAL_PRIORITY                             string = "Priority"
	REGSTR_VAL_DRIVER                               string = "Driver"
	REGSTR_VAL_FUNCDESC                             string = "FunctionDesc"
	REGSTR_VAL_FORCEDCONFIG                         string = "ForcedConfig"
	REGSTR_VAL_CONFIGFLAGS                          string = "ConfigFlags"
	REGSTR_VAL_CSCONFIGFLAGS                        string = "CSConfigFlags"
	CONFIGFLAG_DISABLED                             uint32 = 0x1
	CONFIGFLAG_REMOVED                              uint32 = 0x2
	CONFIGFLAG_MANUAL_INSTALL                       uint32 = 0x4
	CONFIGFLAG_IGNORE_BOOT_LC                       uint32 = 0x8
	CONFIGFLAG_NET_BOOT                             uint32 = 0x10
	CONFIGFLAG_REINSTALL                            uint32 = 0x20
	CONFIGFLAG_FAILEDINSTALL                        uint32 = 0x40
	CONFIGFLAG_CANTSTOPACHILD                       uint32 = 0x80
	CONFIGFLAG_OKREMOVEROM                          uint32 = 0x100
	CONFIGFLAG_NOREMOVEEXIT                         uint32 = 0x200
	CONFIGFLAG_FINISH_INSTALL                       uint32 = 0x400
	CONFIGFLAG_NEEDS_FORCED_CONFIG                  uint32 = 0x800
	CONFIGFLAG_NETBOOT_CARD                         uint32 = 0x1000
	CONFIGFLAG_PARTIAL_LOG_CONF                     uint32 = 0x2000
	CONFIGFLAG_SUPPRESS_SURPRISE                    uint32 = 0x4000
	CONFIGFLAG_VERIFY_HARDWARE                      uint32 = 0x8000
	CONFIGFLAG_FINISHINSTALL_UI                     uint32 = 0x10000
	CONFIGFLAG_FINISHINSTALL_ACTION                 uint32 = 0x20000
	CONFIGFLAG_BOOT_DEVICE                          uint32 = 0x40000
	CONFIGFLAG_NEEDS_CLASS_CONFIG                   uint32 = 0x80000
	CSCONFIGFLAG_BITS                               uint32 = 0x7
	CSCONFIGFLAG_DISABLED                           uint32 = 0x1
	CSCONFIGFLAG_DO_NOT_CREATE                      uint32 = 0x2
	CSCONFIGFLAG_DO_NOT_START                       uint32 = 0x4
	DMSTATEFLAG_APPLYTOALL                          uint32 = 0x1
	REGSTR_VAL_ROOT_DEVNODE                         string = "HTREE\\ROOT\\0"
	REGSTR_VAL_RESERVED_DEVNODE                     string = "HTREE\\RESERVED\\0"
	REGSTR_PATH_MULTI_FUNCTION                      string = "MF"
	REGSTR_VAL_RESOURCE_MAP                         string = "ResourceMap"
	REGSTR_PATH_CHILD_PREFIX                        string = "Child"
	NUM_RESOURCE_MAP                                uint32 = 0x100
	REGSTR_VAL_MF_FLAGS                             string = "MFFlags"
	MF_FLAGS_EVEN_IF_NO_RESOURCE                    uint32 = 0x1
	MF_FLAGS_NO_CREATE_IF_NO_RESOURCE               uint32 = 0x2
	MF_FLAGS_FILL_IN_UNKNOWN_RESOURCE               uint32 = 0x4
	MF_FLAGS_CREATE_BUT_NO_SHOW_DISABLED            uint32 = 0x8
	REGSTR_VAL_EISA_RANGES                          string = "EISARanges"
	REGSTR_VAL_EISA_FUNCTIONS                       string = "EISAFunctions"
	REGSTR_VAL_EISA_FUNCTIONS_MASK                  string = "EISAFunctionsMask"
	REGSTR_VAL_EISA_FLAGS                           string = "EISAFlags"
	REGSTR_VAL_EISA_SIMULATE_INT15                  string = "EISASimulateInt15"
	EISAFLAG_NO_IO_MERGE                            uint32 = 0x1
	EISAFLAG_SLOT_IO_FIRST                          uint32 = 0x2
	EISA_NO_MAX_FUNCTION                            uint32 = 0xff
	NUM_EISA_RANGES                                 uint32 = 0x4
	REGSTR_VAL_DRVDESC                              string = "DriverDesc"
	REGSTR_VAL_DEVLOADER                            string = "DevLoader"
	REGSTR_VAL_STATICVXD                            string = "StaticVxD"
	REGSTR_VAL_PROPERTIES                           string = "Properties"
	REGSTR_VAL_MANUFACTURER                         string = "Manufacturer"
	REGSTR_VAL_EXISTS                               string = "Exists"
	REGSTR_VAL_CMENUMFLAGS                          string = "CMEnumFlags"
	REGSTR_VAL_CMDRIVFLAGS                          string = "CMDrivFlags"
	REGSTR_VAL_ENUMERATOR                           string = "Enumerator"
	REGSTR_VAL_DEVICEDRIVER                         string = "DeviceDriver"
	REGSTR_VAL_PORTNAME                             string = "PortName"
	REGSTR_VAL_INFPATH                              string = "InfPath"
	REGSTR_VAL_INFSECTION                           string = "InfSection"
	REGSTR_VAL_INFSECTIONEXT                        string = "InfSectionExt"
	REGSTR_VAL_POLLING                              string = "Polling"
	REGSTR_VAL_DONTLOADIFCONFLICT                   string = "DontLoadIfConflict"
	REGSTR_VAL_PORTSUBCLASS                         string = "PortSubClass"
	REGSTR_VAL_NETCLEAN                             string = "NetClean"
	REGSTR_VAL_IDE_NO_SERIALIZE                     string = "IDENoSerialize"
	REGSTR_VAL_NOCMOSORFDPT                         string = "NoCMOSorFDPT"
	REGSTR_VAL_COMVERIFYBASE                        string = "COMVerifyBase"
	REGSTR_VAL_MATCHINGDEVID                        string = "MatchingDeviceId"
	REGSTR_VAL_DRIVERDATE                           string = "DriverDate"
	REGSTR_VAL_DRIVERDATEDATA                       string = "DriverDateData"
	REGSTR_VAL_DRIVERVERSION                        string = "DriverVersion"
	REGSTR_VAL_LOCATION_INFORMATION_OVERRIDE        string = "LocationInformationOverride"
	REGSTR_KEY_OVERRIDE                             string = "Override"
	REGSTR_VAL_CONFIGMG                             string = "CONFIGMG"
	REGSTR_VAL_SYSDM                                string = "SysDM"
	REGSTR_VAL_SYSDMFUNC                            string = "SysDMFunc"
	REGSTR_VAL_PRIVATE                              string = "Private"
	REGSTR_VAL_PRIVATEFUNC                          string = "PrivateFunc"
	REGSTR_VAL_DETECT                               string = "Detect"
	REGSTR_VAL_DETECTFUNC                           string = "DetectFunc"
	REGSTR_VAL_ASKFORCONFIG                         string = "AskForConfig"
	REGSTR_VAL_ASKFORCONFIGFUNC                     string = "AskForConfigFunc"
	REGSTR_VAL_WAITFORUNDOCK                        string = "WaitForUndock"
	REGSTR_VAL_WAITFORUNDOCKFUNC                    string = "WaitForUndockFunc"
	REGSTR_VAL_REMOVEROMOKAY                        string = "RemoveRomOkay"
	REGSTR_VAL_REMOVEROMOKAYFUNC                    string = "RemoveRomOkayFunc"
	REGSTR_VAL_CURCONFIG                            string = "CurrentConfig"
	REGSTR_VAL_FRIENDLYNAME                         string = "FriendlyName"
	REGSTR_VAL_CURRENTCONFIG                        string = "CurrentConfig"
	REGSTR_VAL_MAP                                  string = "Map"
	REGSTR_VAL_ID                                   string = "CurrentID"
	REGSTR_VAL_DOCKED                               string = "CurrentDockedState"
	REGSTR_VAL_CHECKSUM                             string = "CurrentChecksum"
	REGSTR_VAL_HWDETECT                             string = "HardwareDetect"
	REGSTR_VAL_INHIBITRESULTS                       string = "InhibitResults"
	REGSTR_VAL_PROFILEFLAGS                         string = "ProfileFlags"
	REGSTR_KEY_PCMCIA                               string = "PCMCIA\\"
	REGSTR_KEY_PCUNKNOWN                            string = "UNKNOWN_MANUFACTURER"
	REGSTR_VAL_PCSSDRIVER                           string = "Driver"
	REGSTR_KEY_PCMTD                                string = "MTD-"
	REGSTR_VAL_PCMTDRIVER                           string = "MTD"
	REGSTR_VAL_HARDWAREID                           string = "HardwareID"
	REGSTR_VAL_INSTALLER                            string = "Installer"
	REGSTR_VAL_INSTALLER_32                         string = "Installer32"
	REGSTR_VAL_INSICON                              string = "Icon"
	REGSTR_VAL_ENUMPROPPAGES                        string = "EnumPropPages"
	REGSTR_VAL_ENUMPROPPAGES_32                     string = "EnumPropPages32"
	REGSTR_VAL_BASICPROPERTIES                      string = "BasicProperties"
	REGSTR_VAL_BASICPROPERTIES_32                   string = "BasicProperties32"
	REGSTR_VAL_COINSTALLERS_32                      string = "CoInstallers32"
	REGSTR_VAL_PRIVATEPROBLEM                       string = "PrivateProblem"
	REGSTR_KEY_CURRENT                              string = "Current"
	REGSTR_KEY_DEFAULT                              string = "Default"
	REGSTR_KEY_MODES                                string = "Modes"
	REGSTR_VAL_MODE                                 string = "Mode"
	REGSTR_VAL_BPP                                  string = "BPP"
	REGSTR_VAL_HRES                                 string = "HRes"
	REGSTR_VAL_VRES                                 string = "VRes"
	REGSTR_VAL_FONTSIZE                             string = "FontSize"
	REGSTR_VAL_DRV                                  string = "drv"
	REGSTR_VAL_GRB                                  string = "grb"
	REGSTR_VAL_VDD                                  string = "vdd"
	REGSTR_VAL_VER                                  string = "Ver"
	REGSTR_VAL_MAXRES                               string = "MaxResolution"
	REGSTR_VAL_DPMS                                 string = "DPMS"
	REGSTR_VAL_RESUMERESET                          string = "ResumeReset"
	REGSTR_KEY_SYSTEM                               string = "System"
	REGSTR_KEY_USER                                 string = "User"
	REGSTR_VAL_DPI                                  string = "dpi"
	REGSTR_VAL_PCICOPTIONS                          string = "PCICOptions"
	PCIC_DEFAULT_IRQMASK                            uint32 = 0x4eb8
	PCIC_DEFAULT_NUMSOCKETS                         uint32 = 0x0
	REGSTR_VAL_PCICIRQMAP                           string = "PCICIRQMap"
	REGSTR_PATH_APPEARANCE                          string = "Control Panel\\Appearance"
	REGSTR_PATH_LOOKSCHEMES                         string = "Control Panel\\Appearance\\Schemes"
	REGSTR_VAL_CUSTOMCOLORS                         string = "CustomColors"
	REGSTR_PATH_SCREENSAVE                          string = "Control Panel\\Desktop"
	REGSTR_VALUE_USESCRPASSWORD                     string = "ScreenSaveUsePassword"
	REGSTR_VALUE_SCRPASSWORD                        string = "ScreenSave_Data"
	REGSTR_VALUE_LOWPOWERTIMEOUT                    string = "ScreenSaveLowPowerTimeout"
	REGSTR_VALUE_POWEROFFTIMEOUT                    string = "ScreenSavePowerOffTimeout"
	REGSTR_VALUE_LOWPOWERACTIVE                     string = "ScreenSaveLowPowerActive"
	REGSTR_VALUE_POWEROFFACTIVE                     string = "ScreenSavePowerOffActive"
	REGSTR_PATH_WINDOWSAPPLETS                      string = "Software\\Microsoft\\Windows\\CurrentVersion\\Applets"
	REGSTR_PATH_SYSTRAY                             string = "Software\\Microsoft\\Windows\\CurrentVersion\\Applets\\SysTray"
	REGSTR_VAL_SYSTRAYSVCS                          string = "Services"
	REGSTR_VAL_SYSTRAYBATFLAGS                      string = "PowerFlags"
	REGSTR_VAL_SYSTRAYPCCARDFLAGS                   string = "PCMCIAFlags"
	REGSTR_PATH_NETWORK_USERSETTINGS                string = "Network"
	REGSTR_KEY_NETWORK_PERSISTENT                   string = "\\Persistent"
	REGSTR_KEY_NETWORK_RECENT                       string = "\\Recent"
	REGSTR_VAL_REMOTE_PATH                          string = "RemotePath"
	REGSTR_VAL_USER_NAME                            string = "UserName"
	REGSTR_VAL_PROVIDER_NAME                        string = "ProviderName"
	REGSTR_VAL_CONNECTION_TYPE                      string = "ConnectionType"
	REGSTR_VAL_UPGRADE                              string = "Upgrade"
	REGSTR_KEY_LOGON                                string = "\\Logon"
	REGSTR_VAL_MUSTBEVALIDATED                      string = "MustBeValidated"
	REGSTR_VAL_RUNLOGINSCRIPT                       string = "ProcessLoginScript"
	REGSTR_KEY_NETWORKPROVIDER                      string = "\\NetworkProvider"
	REGSTR_VAL_AUTHENT_AGENT                        string = "AuthenticatingAgent"
	REGSTR_VAL_PREFREDIR                            string = "PreferredRedir"
	REGSTR_VAL_AUTOSTART                            string = "AutoStart"
	REGSTR_VAL_AUTOLOGON                            string = "AutoLogon"
	REGSTR_VAL_NETCARD                              string = "Netcard"
	REGSTR_VAL_TRANSPORT                            string = "Transport"
	REGSTR_VAL_DYNAMIC                              string = "Dynamic"
	REGSTR_VAL_TRANSITION                           string = "Transition"
	REGSTR_VAL_STATICDRIVE                          string = "StaticDrive"
	REGSTR_VAL_LOADHI                               string = "LoadHi"
	REGSTR_VAL_LOADRMDRIVERS                        string = "LoadRMDrivers"
	REGSTR_VAL_SETUPN                               string = "SetupN"
	REGSTR_VAL_SETUPNPATH                           string = "SetupNPath"
	REGSTR_VAL_WRKGRP_FORCEMAPPING                  string = "WrkgrpForceMapping"
	REGSTR_VAL_WRKGRP_REQUIRED                      string = "WrkgrpRequired"
	REGSTR_PATH_CURRENT_CONTROL_SET                 string = "System\\CurrentControlSet\\Control"
	REGSTR_VAL_CURRENT_USER                         string = "Current User"
	REGSTR_PATH_PWDPROVIDER                         string = "System\\CurrentControlSet\\Control\\PwdProvider"
	REGSTR_VAL_PWDPROVIDER_PATH                     string = "ProviderPath"
	REGSTR_VAL_PWDPROVIDER_DESC                     string = "Description"
	REGSTR_VAL_PWDPROVIDER_CHANGEPWD                string = "ChangePassword"
	REGSTR_VAL_PWDPROVIDER_CHANGEPWDHWND            string = "ChangePasswordHwnd"
	REGSTR_VAL_PWDPROVIDER_GETPWDSTATUS             string = "GetPasswordStatus"
	REGSTR_VAL_PWDPROVIDER_ISNP                     string = "NetworkProvider"
	REGSTR_VAL_PWDPROVIDER_CHANGEORDER              string = "ChangeOrder"
	REGSTR_PATH_POLICIES                            string = "Software\\Microsoft\\Windows\\CurrentVersion\\Policies"
	REGSTR_PATH_UPDATE                              string = "System\\CurrentControlSet\\Control\\Update"
	REGSTR_VALUE_ENABLE                             string = "Enable"
	REGSTR_VALUE_VERBOSE                            string = "Verbose"
	REGSTR_VALUE_NETPATH                            string = "NetworkPath"
	REGSTR_VALUE_DEFAULTLOC                         string = "UseDefaultNetLocation"
	REGSTR_KEY_NETWORK                              string = "Network"
	REGSTR_KEY_PRINTERS                             string = "Printers"
	REGSTR_KEY_WINOLDAPP                            string = "WinOldApp"
	REGSTR_KEY_EXPLORER                             string = "Explorer"
	REGSTR_VAL_NOFILESHARING                        string = "NoFileSharing"
	REGSTR_VAL_NOPRINTSHARING                       string = "NoPrintSharing"
	REGSTR_VAL_NOFILESHARINGCTRL                    string = "NoFileSharingControl"
	REGSTR_VAL_NOPRINTSHARINGCTRL                   string = "NoPrintSharingControl"
	REGSTR_VAL_HIDESHAREPWDS                        string = "HideSharePwds"
	REGSTR_VAL_DISABLEPWDCACHING                    string = "DisablePwdCaching"
	REGSTR_VAL_ALPHANUMPWDS                         string = "AlphanumPwds"
	REGSTR_VAL_NETSETUP_DISABLE                     string = "NoNetSetup"
	REGSTR_VAL_NETSETUP_NOCONFIGPAGE                string = "NoNetSetupConfigPage"
	REGSTR_VAL_NETSETUP_NOIDPAGE                    string = "NoNetSetupIDPage"
	REGSTR_VAL_NETSETUP_NOSECURITYPAGE              string = "NoNetSetupSecurityPage"
	REGSTR_VAL_SYSTEMCPL_NOVIRTMEMPAGE              string = "NoVirtMemPage"
	REGSTR_VAL_SYSTEMCPL_NODEVMGRPAGE               string = "NoDevMgrPage"
	REGSTR_VAL_SYSTEMCPL_NOCONFIGPAGE               string = "NoConfigPage"
	REGSTR_VAL_SYSTEMCPL_NOFILESYSPAGE              string = "NoFileSysPage"
	REGSTR_VAL_DISPCPL_NODISPCPL                    string = "NoDispCPL"
	REGSTR_VAL_DISPCPL_NOBACKGROUNDPAGE             string = "NoDispBackgroundPage"
	REGSTR_VAL_DISPCPL_NOSCRSAVPAGE                 string = "NoDispScrSavPage"
	REGSTR_VAL_DISPCPL_NOAPPEARANCEPAGE             string = "NoDispAppearancePage"
	REGSTR_VAL_DISPCPL_NOSETTINGSPAGE               string = "NoDispSettingsPage"
	REGSTR_VAL_SECCPL_NOSECCPL                      string = "NoSecCPL"
	REGSTR_VAL_SECCPL_NOPWDPAGE                     string = "NoPwdPage"
	REGSTR_VAL_SECCPL_NOADMINPAGE                   string = "NoAdminPage"
	REGSTR_VAL_SECCPL_NOPROFILEPAGE                 string = "NoProfilePage"
	REGSTR_VAL_PRINTERS_HIDETABS                    string = "NoPrinterTabs"
	REGSTR_VAL_PRINTERS_NODELETE                    string = "NoDeletePrinter"
	REGSTR_VAL_PRINTERS_NOADD                       string = "NoAddPrinter"
	REGSTR_VAL_WINOLDAPP_DISABLED                   string = "Disabled"
	REGSTR_VAL_WINOLDAPP_NOREALMODE                 string = "NoRealMode"
	REGSTR_VAL_NOENTIRENETWORK                      string = "NoEntireNetwork"
	REGSTR_VAL_NOWORKGROUPCONTENTS                  string = "NoWorkgroupContents"
	REGSTR_VAL_UNDOCK_WITHOUT_LOGON                 string = "UndockWithoutLogon"
	REGSTR_VAL_MINPWDLEN                            string = "MinPwdLen"
	REGSTR_VAL_PWDEXPIRATION                        string = "PwdExpiration"
	REGSTR_VAL_WIN31PROVIDER                        string = "Win31Provider"
	REGSTR_VAL_DISABLEREGTOOLS                      string = "DisableRegistryTools"
	REGSTR_PATH_WINLOGON                            string = "Software\\Microsoft\\Windows\\CurrentVersion\\Winlogon"
	REGSTR_VAL_LEGALNOTICECAPTION                   string = "LegalNoticeCaption"
	REGSTR_VAL_LEGALNOTICETEXT                      string = "LegalNoticeText"
	REGSTR_VAL_DRIVE_SPINDOWN                       string = "NoDispSpinDown"
	REGSTR_VAL_SHUTDOWN_FLAGS                       string = "ShutdownFlags"
	REGSTR_VAL_RESTRICTRUN                          string = "RestrictRun"
	REGSTR_KEY_POL_USERS                            string = "Users"
	REGSTR_KEY_POL_COMPUTERS                        string = "Computers"
	REGSTR_KEY_POL_USERGROUPS                       string = "UserGroups"
	REGSTR_KEY_POL_DEFAULT                          string = ".default"
	REGSTR_KEY_POL_USERGROUPDATA                    string = "GroupData\\UserGroups\\Priority"
	REGSTR_PATH_TIMEZONE                            string = "System\\CurrentControlSet\\Control\\TimeZoneInformation"
	REGSTR_VAL_TZBIAS                               string = "Bias"
	REGSTR_VAL_TZDLTBIAS                            string = "DaylightBias"
	REGSTR_VAL_TZSTDBIAS                            string = "StandardBias"
	REGSTR_VAL_TZACTBIAS                            string = "ActiveTimeBias"
	REGSTR_VAL_TZDLTFLAG                            string = "DaylightFlag"
	REGSTR_VAL_TZSTDSTART                           string = "StandardStart"
	REGSTR_VAL_TZDLTSTART                           string = "DaylightStart"
	REGSTR_VAL_TZDLTNAME                            string = "DaylightName"
	REGSTR_VAL_TZSTDNAME                            string = "StandardName"
	REGSTR_VAL_TZNOCHANGESTART                      string = "NoChangeStart"
	REGSTR_VAL_TZNOCHANGEEND                        string = "NoChangeEnd"
	REGSTR_VAL_TZNOAUTOTIME                         string = "DisableAutoDaylightTimeSet"
	REGSTR_PATH_FLOATINGPOINTPROCESSOR              string = "HARDWARE\\DESCRIPTION\\System\\FloatingPointProcessor"
	REGSTR_PATH_FLOATINGPOINTPROCESSOR0             string = "HARDWARE\\DESCRIPTION\\System\\FloatingPointProcessor\\0"
	REGSTR_PATH_COMPUTRNAME                         string = "System\\CurrentControlSet\\Control\\ComputerName\\ComputerName"
	REGSTR_VAL_COMPUTRNAME                          string = "ComputerName"
	REGSTR_PATH_SHUTDOWN                            string = "System\\CurrentControlSet\\Control\\Shutdown"
	REGSTR_VAL_FORCEREBOOT                          string = "ForceReboot"
	REGSTR_VAL_SETUPPROGRAMRAN                      string = "SetupProgramRan"
	REGSTR_VAL_DOES_POLLING                         string = "PollingSupportNeeded"
	REGSTR_PATH_KNOWNDLLS                           string = "System\\CurrentControlSet\\Control\\SessionManager\\KnownDLLs"
	REGSTR_PATH_KNOWN16DLLS                         string = "System\\CurrentControlSet\\Control\\SessionManager\\Known16DLLs"
	REGSTR_PATH_CHECKVERDLLS                        string = "System\\CurrentControlSet\\Control\\SessionManager\\CheckVerDLLs"
	REGSTR_PATH_WARNVERDLLS                         string = "System\\CurrentControlSet\\Control\\SessionManager\\WarnVerDLLs"
	REGSTR_PATH_HACKINIFILE                         string = "System\\CurrentControlSet\\Control\\SessionManager\\HackIniFiles"
	REGSTR_PATH_CHECKBADAPPS                        string = "System\\CurrentControlSet\\Control\\SessionManager\\CheckBadApps"
	REGSTR_PATH_APPPATCH                            string = "System\\CurrentControlSet\\Control\\SessionManager\\AppPatches"
	REGSTR_PATH_CHECKBADAPPS400                     string = "System\\CurrentControlSet\\Control\\SessionManager\\CheckBadApps400"
	REGSTR_PATH_KNOWNVXDS                           string = "System\\CurrentControlSet\\Control\\SessionManager\\KnownVxDs"
	REGSTR_VAL_UNINSTALLER_DISPLAYNAME              string = "DisplayName"
	REGSTR_VAL_UNINSTALLER_COMMANDLINE              string = "UninstallString"
	REGSTR_VAL_REINSTALL_DISPLAYNAME                string = "DisplayName"
	REGSTR_VAL_REINSTALL_STRING                     string = "ReinstallString"
	REGSTR_VAL_REINSTALL_DEVICEINSTANCEIDS          string = "DeviceInstanceIds"
	REGSTR_PATH_DESKTOP                             string = "Control Panel\\Desktop"
	REGSTR_PATH_MOUSE                               string = "Control Panel\\Mouse"
	REGSTR_PATH_KEYBOARD                            string = "Control Panel\\Keyboard"
	REGSTR_PATH_COLORS                              string = "Control Panel\\Colors"
	REGSTR_PATH_SOUND                               string = "Control Panel\\Sound"
	REGSTR_PATH_METRICS                             string = "Control Panel\\Desktop\\WindowMetrics"
	REGSTR_PATH_ICONS                               string = "Control Panel\\Icons"
	REGSTR_PATH_CURSORS                             string = "Control Panel\\Cursors"
	REGSTR_PATH_CHECKDISK                           string = "Software\\Microsoft\\Windows\\CurrentVersion\\Applets\\Check Drive"
	REGSTR_PATH_CHECKDISKSET                        string = "Settings"
	REGSTR_PATH_CHECKDISKUDRVS                      string = "NoUnknownDDErrDrvs"
	REGSTR_PATH_FAULT                               string = "Software\\Microsoft\\Windows\\CurrentVersion\\Fault"
	REGSTR_VAL_FAULT_LOGFILE                        string = "LogFile"
	REGSTR_PATH_AEDEBUG                             string = "Software\\Microsoft\\Windows NT\\CurrentVersion\\AeDebug"
	REGSTR_VAL_AEDEBUG_DEBUGGER                     string = "Debugger"
	REGSTR_VAL_AEDEBUG_AUTO                         string = "Auto"
	REGSTR_PATH_GRPCONV                             string = "Software\\Microsoft\\Windows\\CurrentVersion\\GrpConv"
	REGSTR_VAL_REGITEMDELETEMESSAGE                 string = "Removal Message"
	REGSTR_PATH_LASTCHECK                           string = "Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\LastCheck"
	REGSTR_PATH_LASTOPTIMIZE                        string = "Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\LastOptimize"
	REGSTR_PATH_LASTBACKUP                          string = "Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\LastBackup"
	REGSTR_PATH_CHKLASTCHECK                        string = "Software\\Microsoft\\Windows\\CurrentVersion\\Applets\\Check Drive\\LastCheck"
	REGSTR_PATH_CHKLASTSURFAN                       string = "Software\\Microsoft\\Windows\\CurrentVersion\\Applets\\Check Drive\\LastSurfaceAnalysis"
	DTRESULTOK                                      uint32 = 0x0
	DTRESULTFIX                                     uint32 = 0x1
	DTRESULTPROB                                    uint32 = 0x2
	DTRESULTPART                                    uint32 = 0x3
	REGSTR_KEY_SHARES                               string = "Software\\Microsoft\\Windows\\CurrentVersion\\Network\\LanMan"
	REGSTR_VAL_SHARES_FLAGS                         string = "Flags"
	REGSTR_VAL_SHARES_TYPE                          string = "Type"
	REGSTR_VAL_SHARES_PATH                          string = "Path"
	REGSTR_VAL_SHARES_REMARK                        string = "Remark"
	REGSTR_VAL_SHARES_RW_PASS                       string = "Parm1"
	REGSTR_VAL_SHARES_RO_PASS                       string = "Parm2"
	REGSTR_PATH_PRINT                               string = "System\\CurrentControlSet\\Control\\Print"
	REGSTR_PATH_PRINTERS                            string = "System\\CurrentControlSet\\Control\\Print\\Printers"
	REGSTR_PATH_PROVIDERS                           string = "System\\CurrentControlSet\\Control\\Print\\Providers"
	REGSTR_PATH_MONITORS                            string = "System\\CurrentControlSet\\Control\\Print\\Monitors"
	REGSTR_PATH_ENVIRONMENTS                        string = "System\\CurrentControlSet\\Control\\Print\\Environments"
	REGSTR_VAL_START_ON_BOOT                        string = "StartOnBoot"
	REGSTR_VAL_PRINTERS_MASK                        string = "PrintersMask"
	REGSTR_VAL_DOS_SPOOL_MASK                       string = "DOSSpoolMask"
	REGSTR_KEY_CURRENT_ENV                          string = "\\Windows 4.0"
	REGSTR_KEY_DRIVERS                              string = "\\Drivers"
	REGSTR_KEY_PRINT_PROC                           string = "\\Print Processors"
	REGSTR_PATH_EVENTLABELS                         string = "AppEvents\\EventLabels"
	REGSTR_PATH_SCHEMES                             string = "AppEvents\\Schemes"
	REGSTR_PATH_MULTIMEDIA_AUDIO                    string = "Software\\Microsoft\\Multimedia\\Audio"
	REGSTR_KEY_JOYCURR                              string = "CurrentJoystickSettings"
	REGSTR_KEY_JOYSETTINGS                          string = "JoystickSettings"
	REGSTR_VAL_JOYUSERVALUES                        string = "JoystickUserValues"
	REGSTR_VAL_JOYCALLOUT                           string = "JoystickCallout"
	REGSTR_VAL_JOYNCONFIG                           string = "Joystick%dConfiguration"
	REGSTR_VAL_JOYNOEMNAME                          string = "Joystick%dOEMName"
	REGSTR_VAL_JOYNOEMCALLOUT                       string = "Joystick%dOEMCallout"
	REGSTR_VAL_JOYOEMCALLOUT                        string = "OEMCallout"
	REGSTR_VAL_JOYOEMNAME                           string = "OEMName"
	REGSTR_VAL_JOYOEMDATA                           string = "OEMData"
	REGSTR_VAL_JOYOEMXYLABEL                        string = "OEMXYLabel"
	REGSTR_VAL_JOYOEMZLABEL                         string = "OEMZLabel"
	REGSTR_VAL_JOYOEMRLABEL                         string = "OEMRLabel"
	REGSTR_VAL_JOYOEMPOVLABEL                       string = "OEMPOVLabel"
	REGSTR_VAL_JOYOEMULABEL                         string = "OEMULabel"
	REGSTR_VAL_JOYOEMVLABEL                         string = "OEMVLabel"
	REGSTR_VAL_JOYOEMTESTMOVEDESC                   string = "OEMTestMoveDesc"
	REGSTR_VAL_JOYOEMTESTBUTTONDESC                 string = "OEMTestButtonDesc"
	REGSTR_VAL_JOYOEMTESTMOVECAP                    string = "OEMTestMoveCap"
	REGSTR_VAL_JOYOEMTESTBUTTONCAP                  string = "OEMTestButtonCap"
	REGSTR_VAL_JOYOEMTESTWINCAP                     string = "OEMTestWinCap"
	REGSTR_VAL_JOYOEMCALCAP                         string = "OEMCalCap"
	REGSTR_VAL_JOYOEMCALWINCAP                      string = "OEMCalWinCap"
	REGSTR_VAL_JOYOEMCAL1                           string = "OEMCal1"
	REGSTR_VAL_JOYOEMCAL2                           string = "OEMCal2"
	REGSTR_VAL_JOYOEMCAL3                           string = "OEMCal3"
	REGSTR_VAL_JOYOEMCAL4                           string = "OEMCal4"
	REGSTR_VAL_JOYOEMCAL5                           string = "OEMCal5"
	REGSTR_VAL_JOYOEMCAL6                           string = "OEMCal6"
	REGSTR_VAL_JOYOEMCAL7                           string = "OEMCal7"
	REGSTR_VAL_JOYOEMCAL8                           string = "OEMCal8"
	REGSTR_VAL_JOYOEMCAL9                           string = "OEMCal9"
	REGSTR_VAL_JOYOEMCAL10                          string = "OEMCal10"
	REGSTR_VAL_JOYOEMCAL11                          string = "OEMCal11"
	REGSTR_VAL_JOYOEMCAL12                          string = "OEMCal12"
	REGSTR_VAL_AUDIO_BITMAP                         string = "bitmap"
	REGSTR_VAL_AUDIO_ICON                           string = "icon"
	REGSTR_PATH_DEVICEINSTALLER                     string = "Software\\Microsoft\\Windows\\CurrentVersion\\Device Installer"
	REGSTR_PATH_DIFX                                string = "Software\\Microsoft\\Windows\\CurrentVersion\\DIFX"
	REGSTR_VAL_SEARCHOPTIONS                        string = "SearchOptions"
	REGSTR_PATH_BIOSINFO                            string = "System\\CurrentControlSet\\Control\\BiosInfo"
	REGSTR_PATH_PCIIR                               string = "System\\CurrentControlSet\\Control\\Pnp\\PciIrqRouting"
	REGSTR_VAL_OPTIONS                              string = "Options"
	REGSTR_VAL_STAT                                 string = "Status"
	REGSTR_VAL_TABLE_STAT                           string = "TableStatus"
	REGSTR_VAL_MINIPORT_STAT                        string = "MiniportStatus"
	PIR_OPTION_ENABLED                              uint32 = 0x1
	PIR_OPTION_REGISTRY                             uint32 = 0x2
	PIR_OPTION_MSSPEC                               uint32 = 0x4
	PIR_OPTION_REALMODE                             uint32 = 0x8
	PIR_OPTION_DEFAULT                              uint32 = 0xf
	PIR_STATUS_ERROR                                uint32 = 0x0
	PIR_STATUS_ENABLED                              uint32 = 0x1
	PIR_STATUS_DISABLED                             uint32 = 0x2
	PIR_STATUS_MAX                                  uint32 = 0x3
	PIR_STATUS_TABLE_REGISTRY                       uint32 = 0x0
	PIR_STATUS_TABLE_MSSPEC                         uint32 = 0x1
	PIR_STATUS_TABLE_REALMODE                       uint32 = 0x2
	PIR_STATUS_TABLE_NONE                           uint32 = 0x3
	PIR_STATUS_TABLE_ERROR                          uint32 = 0x4
	PIR_STATUS_TABLE_BAD                            uint32 = 0x5
	PIR_STATUS_TABLE_SUCCESS                        uint32 = 0x6
	PIR_STATUS_TABLE_MAX                            uint32 = 0x7
	PIR_STATUS_MINIPORT_NORMAL                      uint32 = 0x0
	PIR_STATUS_MINIPORT_COMPATIBLE                  uint32 = 0x1
	PIR_STATUS_MINIPORT_OVERRIDE                    uint32 = 0x2
	PIR_STATUS_MINIPORT_NONE                        uint32 = 0x3
	PIR_STATUS_MINIPORT_ERROR                       uint32 = 0x4
	PIR_STATUS_MINIPORT_NOKEY                       uint32 = 0x5
	PIR_STATUS_MINIPORT_SUCCESS                     uint32 = 0x6
	PIR_STATUS_MINIPORT_INVALID                     uint32 = 0x7
	PIR_STATUS_MINIPORT_MAX                         uint32 = 0x8
	REGSTR_PATH_LASTGOOD                            string = "System\\LastKnownGoodRecovery\\LastGood"
	REGSTR_PATH_LASTGOODTMP                         string = "System\\LastKnownGoodRecovery\\LastGood.Tmp"
	LASTGOOD_OPERATION                              uint32 = 0xff
	LASTGOOD_OPERATION_NOPOSTPROC                   uint32 = 0x0
	LASTGOOD_OPERATION_DELETE                       uint32 = 0x1
)

// enums

// enum
type REG_VALUE_TYPE uint32

const (
	REG_NONE                       REG_VALUE_TYPE = 0
	REG_SZ                         REG_VALUE_TYPE = 1
	REG_EXPAND_SZ                  REG_VALUE_TYPE = 2
	REG_BINARY                     REG_VALUE_TYPE = 3
	REG_DWORD                      REG_VALUE_TYPE = 4
	REG_DWORD_LITTLE_ENDIAN        REG_VALUE_TYPE = 4
	REG_DWORD_BIG_ENDIAN           REG_VALUE_TYPE = 5
	REG_LINK                       REG_VALUE_TYPE = 6
	REG_MULTI_SZ                   REG_VALUE_TYPE = 7
	REG_RESOURCE_LIST              REG_VALUE_TYPE = 8
	REG_FULL_RESOURCE_DESCRIPTOR   REG_VALUE_TYPE = 9
	REG_RESOURCE_REQUIREMENTS_LIST REG_VALUE_TYPE = 10
	REG_QWORD                      REG_VALUE_TYPE = 11
	REG_QWORD_LITTLE_ENDIAN        REG_VALUE_TYPE = 11
)

// enum
// flags
type REG_SAM_FLAGS uint32

const (
	KEY_QUERY_VALUE        REG_SAM_FLAGS = 1
	KEY_SET_VALUE          REG_SAM_FLAGS = 2
	KEY_CREATE_SUB_KEY     REG_SAM_FLAGS = 4
	KEY_ENUMERATE_SUB_KEYS REG_SAM_FLAGS = 8
	KEY_NOTIFY             REG_SAM_FLAGS = 16
	KEY_CREATE_LINK        REG_SAM_FLAGS = 32
	KEY_WOW64_32KEY        REG_SAM_FLAGS = 512
	KEY_WOW64_64KEY        REG_SAM_FLAGS = 256
	KEY_WOW64_RES          REG_SAM_FLAGS = 768
	KEY_READ               REG_SAM_FLAGS = 131097
	KEY_WRITE              REG_SAM_FLAGS = 131078
	KEY_EXECUTE            REG_SAM_FLAGS = 131097
	KEY_ALL_ACCESS         REG_SAM_FLAGS = 983103
)

// enum
// flags
type REG_OPEN_CREATE_OPTIONS uint32

const (
	REG_OPTION_RESERVED        REG_OPEN_CREATE_OPTIONS = 0
	REG_OPTION_NON_VOLATILE    REG_OPEN_CREATE_OPTIONS = 0
	REG_OPTION_VOLATILE        REG_OPEN_CREATE_OPTIONS = 1
	REG_OPTION_CREATE_LINK     REG_OPEN_CREATE_OPTIONS = 2
	REG_OPTION_BACKUP_RESTORE  REG_OPEN_CREATE_OPTIONS = 4
	REG_OPTION_OPEN_LINK       REG_OPEN_CREATE_OPTIONS = 8
	REG_OPTION_DONT_VIRTUALIZE REG_OPEN_CREATE_OPTIONS = 16
)

// enum
type REG_CREATE_KEY_DISPOSITION uint32

const (
	REG_CREATED_NEW_KEY     REG_CREATE_KEY_DISPOSITION = 1
	REG_OPENED_EXISTING_KEY REG_CREATE_KEY_DISPOSITION = 2
)

// enum
type REG_SAVE_FORMAT uint32

const (
	REG_STANDARD_FORMAT REG_SAVE_FORMAT = 1
	REG_LATEST_FORMAT   REG_SAVE_FORMAT = 2
	REG_NO_COMPRESSION  REG_SAVE_FORMAT = 4
)

// enum
type REG_RESTORE_KEY_FLAGS int32

const (
	REG_FORCE_RESTORE       REG_RESTORE_KEY_FLAGS = 8
	REG_WHOLE_HIVE_VOLATILE REG_RESTORE_KEY_FLAGS = 1
)

// enum
// flags
type REG_NOTIFY_FILTER uint32

const (
	REG_NOTIFY_CHANGE_NAME       REG_NOTIFY_FILTER = 1
	REG_NOTIFY_CHANGE_ATTRIBUTES REG_NOTIFY_FILTER = 2
	REG_NOTIFY_CHANGE_LAST_SET   REG_NOTIFY_FILTER = 4
	REG_NOTIFY_CHANGE_SECURITY   REG_NOTIFY_FILTER = 8
	REG_NOTIFY_THREAD_AGNOSTIC   REG_NOTIFY_FILTER = 268435456
)

// enum
// flags
type RRF_RT uint32

const (
	RRF_RT_ANY           RRF_RT = 65535
	RRF_RT_DWORD         RRF_RT = 24
	RRF_RT_QWORD         RRF_RT = 72
	RRF_RT_REG_BINARY    RRF_RT = 8
	RRF_RT_REG_DWORD     RRF_RT = 16
	RRF_RT_REG_EXPAND_SZ RRF_RT = 4
	RRF_RT_REG_MULTI_SZ  RRF_RT = 32
	RRF_RT_REG_NONE      RRF_RT = 1
	RRF_RT_REG_QWORD     RRF_RT = 64
	RRF_RT_REG_SZ        RRF_RT = 2
)

// structs

type Val_context struct {
	Valuelen      int32
	Value_context unsafe.Pointer
	Val_buff_ptr  unsafe.Pointer
}

type PvalueA struct {
	Pv_valuename     PSTR
	Pv_valuelen      int32
	Pv_value_context unsafe.Pointer
	Pv_type          uint32
}

type Pvalue = PvalueW
type PvalueW struct {
	Pv_valuename     PWSTR
	Pv_valuelen      int32
	Pv_value_context unsafe.Pointer
	Pv_type          uint32
}

type Provider_info struct {
	Pi_R0_1val     PQUERYHANDLER
	Pi_R0_allvals  PQUERYHANDLER
	Pi_R3_1val     PQUERYHANDLER
	Pi_R3_allvals  PQUERYHANDLER
	Pi_flags       uint32
	Pi_key_context unsafe.Pointer
}

type VALENTA struct {
	Ve_valuename PSTR
	Ve_valuelen  uint32
	Ve_valueptr  uintptr
	Ve_type      REG_VALUE_TYPE
}

type VALENT = VALENTW
type VALENTW struct {
	Ve_valuename PWSTR
	Ve_valuelen  uint32
	Ve_valueptr  uintptr
	Ve_type      REG_VALUE_TYPE
}

type DSKTLSYSTEMTIME struct {
	WYear         uint16
	WMonth        uint16
	WDayOfWeek    uint16
	WDay          uint16
	WHour         uint16
	WMinute       uint16
	WSecond       uint16
	WMilliseconds uint16
	WResult       uint16
}

// func types

type PQUERYHANDLER = uintptr
type PQUERYHANDLER_func = func(keycontext unsafe.Pointer, val_list *Val_context, num_vals uint32, outputbuffer unsafe.Pointer, total_outlen *uint32, input_blen uint32) uint32

var (
	pRegCloseKey                 uintptr
	pRegOverridePredefKey        uintptr
	pRegOpenUserClassesRoot      uintptr
	pRegOpenCurrentUser          uintptr
	pRegDisablePredefinedCache   uintptr
	pRegDisablePredefinedCacheEx uintptr
	pRegConnectRegistryA         uintptr
	pRegConnectRegistryW         uintptr
	pRegConnectRegistryExA       uintptr
	pRegConnectRegistryExW       uintptr
	pRegCreateKeyA               uintptr
	pRegCreateKeyW               uintptr
	pRegCreateKeyExA             uintptr
	pRegCreateKeyExW             uintptr
	pRegCreateKeyTransactedA     uintptr
	pRegCreateKeyTransactedW     uintptr
	pRegDeleteKeyA               uintptr
	pRegDeleteKeyW               uintptr
	pRegDeleteKeyExA             uintptr
	pRegDeleteKeyExW             uintptr
	pRegDeleteKeyTransactedA     uintptr
	pRegDeleteKeyTransactedW     uintptr
	pRegDisableReflectionKey     uintptr
	pRegEnableReflectionKey      uintptr
	pRegQueryReflectionKey       uintptr
	pRegDeleteValueA             uintptr
	pRegDeleteValueW             uintptr
	pRegEnumKeyA                 uintptr
	pRegEnumKeyW                 uintptr
	pRegEnumKeyExA               uintptr
	pRegEnumKeyExW               uintptr
	pRegEnumValueA               uintptr
	pRegEnumValueW               uintptr
	pRegFlushKey                 uintptr
	pRegGetKeySecurity           uintptr
	pRegLoadKeyA                 uintptr
	pRegLoadKeyW                 uintptr
	pRegNotifyChangeKeyValue     uintptr
	pRegOpenKeyA                 uintptr
	pRegOpenKeyW                 uintptr
	pRegOpenKeyExA               uintptr
	pRegOpenKeyExW               uintptr
	pRegOpenKeyTransactedA       uintptr
	pRegOpenKeyTransactedW       uintptr
	pRegQueryInfoKeyA            uintptr
	pRegQueryInfoKeyW            uintptr
	pRegQueryValueA              uintptr
	pRegQueryValueW              uintptr
	pRegQueryMultipleValuesA     uintptr
	pRegQueryMultipleValuesW     uintptr
	pRegQueryValueExA            uintptr
	pRegQueryValueExW            uintptr
	pRegReplaceKeyA              uintptr
	pRegReplaceKeyW              uintptr
	pRegRestoreKeyA              uintptr
	pRegRestoreKeyW              uintptr
	pRegRenameKey                uintptr
	pRegSaveKeyA                 uintptr
	pRegSaveKeyW                 uintptr
	pRegSetKeySecurity           uintptr
	pRegSetValueA                uintptr
	pRegSetValueW                uintptr
	pRegSetValueExA              uintptr
	pRegSetValueExW              uintptr
	pRegUnLoadKeyA               uintptr
	pRegUnLoadKeyW               uintptr
	pRegDeleteKeyValueA          uintptr
	pRegDeleteKeyValueW          uintptr
	pRegSetKeyValueA             uintptr
	pRegSetKeyValueW             uintptr
	pRegDeleteTreeA              uintptr
	pRegDeleteTreeW              uintptr
	pRegCopyTreeA                uintptr
	pRegGetValueA                uintptr
	pRegGetValueW                uintptr
	pRegCopyTreeW                uintptr
	pRegLoadMUIStringA           uintptr
	pRegLoadMUIStringW           uintptr
	pRegLoadAppKeyA              uintptr
	pRegLoadAppKeyW              uintptr
	pRegSaveKeyExA               uintptr
	pRegSaveKeyExW               uintptr
)

func RegCloseKey(hKey HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegCloseKey, libAdvapi32, "RegCloseKey")
	ret, _, _ := syscall.SyscallN(addr, hKey)
	return WIN32_ERROR(ret)
}

func RegOverridePredefKey(hKey HKEY, hNewHKey HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegOverridePredefKey, libAdvapi32, "RegOverridePredefKey")
	ret, _, _ := syscall.SyscallN(addr, hKey, hNewHKey)
	return WIN32_ERROR(ret)
}

func RegOpenUserClassesRoot(hToken HANDLE, dwOptions uint32, samDesired uint32, phkResult *HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegOpenUserClassesRoot, libAdvapi32, "RegOpenUserClassesRoot")
	ret, _, _ := syscall.SyscallN(addr, hToken, uintptr(dwOptions), uintptr(samDesired), uintptr(unsafe.Pointer(phkResult)))
	return WIN32_ERROR(ret)
}

func RegOpenCurrentUser(samDesired uint32, phkResult *HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegOpenCurrentUser, libAdvapi32, "RegOpenCurrentUser")
	ret, _, _ := syscall.SyscallN(addr, uintptr(samDesired), uintptr(unsafe.Pointer(phkResult)))
	return WIN32_ERROR(ret)
}

func RegDisablePredefinedCache() WIN32_ERROR {
	addr := lazyAddr(&pRegDisablePredefinedCache, libAdvapi32, "RegDisablePredefinedCache")
	ret, _, _ := syscall.SyscallN(addr)
	return WIN32_ERROR(ret)
}

func RegDisablePredefinedCacheEx() WIN32_ERROR {
	addr := lazyAddr(&pRegDisablePredefinedCacheEx, libAdvapi32, "RegDisablePredefinedCacheEx")
	ret, _, _ := syscall.SyscallN(addr)
	return WIN32_ERROR(ret)
}

func RegConnectRegistryA(lpMachineName PSTR, hKey HKEY, phkResult *HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegConnectRegistryA, libAdvapi32, "RegConnectRegistryA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMachineName)), hKey, uintptr(unsafe.Pointer(phkResult)))
	return WIN32_ERROR(ret)
}

var RegConnectRegistry = RegConnectRegistryW

func RegConnectRegistryW(lpMachineName PWSTR, hKey HKEY, phkResult *HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegConnectRegistryW, libAdvapi32, "RegConnectRegistryW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMachineName)), hKey, uintptr(unsafe.Pointer(phkResult)))
	return WIN32_ERROR(ret)
}

func RegConnectRegistryExA(lpMachineName PSTR, hKey HKEY, Flags uint32, phkResult *HKEY) int32 {
	addr := lazyAddr(&pRegConnectRegistryExA, libAdvapi32, "RegConnectRegistryExA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMachineName)), hKey, uintptr(Flags), uintptr(unsafe.Pointer(phkResult)))
	return int32(ret)
}

var RegConnectRegistryEx = RegConnectRegistryExW

func RegConnectRegistryExW(lpMachineName PWSTR, hKey HKEY, Flags uint32, phkResult *HKEY) int32 {
	addr := lazyAddr(&pRegConnectRegistryExW, libAdvapi32, "RegConnectRegistryExW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMachineName)), hKey, uintptr(Flags), uintptr(unsafe.Pointer(phkResult)))
	return int32(ret)
}

func RegCreateKeyA(hKey HKEY, lpSubKey PSTR, phkResult *HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegCreateKeyA, libAdvapi32, "RegCreateKeyA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(phkResult)))
	return WIN32_ERROR(ret)
}

var RegCreateKey = RegCreateKeyW

func RegCreateKeyW(hKey HKEY, lpSubKey PWSTR, phkResult *HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegCreateKeyW, libAdvapi32, "RegCreateKeyW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(phkResult)))
	return WIN32_ERROR(ret)
}

func RegCreateKeyExA(hKey HKEY, lpSubKey PSTR, Reserved uint32, lpClass PSTR, dwOptions REG_OPEN_CREATE_OPTIONS, samDesired REG_SAM_FLAGS, lpSecurityAttributes *SECURITY_ATTRIBUTES, phkResult *HKEY, lpdwDisposition *REG_CREATE_KEY_DISPOSITION) WIN32_ERROR {
	addr := lazyAddr(&pRegCreateKeyExA, libAdvapi32, "RegCreateKeyExA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(Reserved), uintptr(unsafe.Pointer(lpClass)), uintptr(dwOptions), uintptr(samDesired), uintptr(unsafe.Pointer(lpSecurityAttributes)), uintptr(unsafe.Pointer(phkResult)), uintptr(unsafe.Pointer(lpdwDisposition)))
	return WIN32_ERROR(ret)
}

var RegCreateKeyEx = RegCreateKeyExW

func RegCreateKeyExW(hKey HKEY, lpSubKey PWSTR, Reserved uint32, lpClass PWSTR, dwOptions REG_OPEN_CREATE_OPTIONS, samDesired REG_SAM_FLAGS, lpSecurityAttributes *SECURITY_ATTRIBUTES, phkResult *HKEY, lpdwDisposition *REG_CREATE_KEY_DISPOSITION) WIN32_ERROR {
	addr := lazyAddr(&pRegCreateKeyExW, libAdvapi32, "RegCreateKeyExW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(Reserved), uintptr(unsafe.Pointer(lpClass)), uintptr(dwOptions), uintptr(samDesired), uintptr(unsafe.Pointer(lpSecurityAttributes)), uintptr(unsafe.Pointer(phkResult)), uintptr(unsafe.Pointer(lpdwDisposition)))
	return WIN32_ERROR(ret)
}

func RegCreateKeyTransactedA(hKey HKEY, lpSubKey PSTR, Reserved uint32, lpClass PSTR, dwOptions REG_OPEN_CREATE_OPTIONS, samDesired REG_SAM_FLAGS, lpSecurityAttributes *SECURITY_ATTRIBUTES, phkResult *HKEY, lpdwDisposition *REG_CREATE_KEY_DISPOSITION, hTransaction HANDLE, pExtendedParemeter unsafe.Pointer) WIN32_ERROR {
	addr := lazyAddr(&pRegCreateKeyTransactedA, libAdvapi32, "RegCreateKeyTransactedA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(Reserved), uintptr(unsafe.Pointer(lpClass)), uintptr(dwOptions), uintptr(samDesired), uintptr(unsafe.Pointer(lpSecurityAttributes)), uintptr(unsafe.Pointer(phkResult)), uintptr(unsafe.Pointer(lpdwDisposition)), hTransaction, uintptr(pExtendedParemeter))
	return WIN32_ERROR(ret)
}

var RegCreateKeyTransacted = RegCreateKeyTransactedW

func RegCreateKeyTransactedW(hKey HKEY, lpSubKey PWSTR, Reserved uint32, lpClass PWSTR, dwOptions REG_OPEN_CREATE_OPTIONS, samDesired REG_SAM_FLAGS, lpSecurityAttributes *SECURITY_ATTRIBUTES, phkResult *HKEY, lpdwDisposition *REG_CREATE_KEY_DISPOSITION, hTransaction HANDLE, pExtendedParemeter unsafe.Pointer) WIN32_ERROR {
	addr := lazyAddr(&pRegCreateKeyTransactedW, libAdvapi32, "RegCreateKeyTransactedW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(Reserved), uintptr(unsafe.Pointer(lpClass)), uintptr(dwOptions), uintptr(samDesired), uintptr(unsafe.Pointer(lpSecurityAttributes)), uintptr(unsafe.Pointer(phkResult)), uintptr(unsafe.Pointer(lpdwDisposition)), hTransaction, uintptr(pExtendedParemeter))
	return WIN32_ERROR(ret)
}

func RegDeleteKeyA(hKey HKEY, lpSubKey PSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegDeleteKeyA, libAdvapi32, "RegDeleteKeyA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)))
	return WIN32_ERROR(ret)
}

var RegDeleteKey = RegDeleteKeyW

func RegDeleteKeyW(hKey HKEY, lpSubKey PWSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegDeleteKeyW, libAdvapi32, "RegDeleteKeyW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)))
	return WIN32_ERROR(ret)
}

func RegDeleteKeyExA(hKey HKEY, lpSubKey PSTR, samDesired uint32, Reserved uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegDeleteKeyExA, libAdvapi32, "RegDeleteKeyExA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(samDesired), uintptr(Reserved))
	return WIN32_ERROR(ret)
}

var RegDeleteKeyEx = RegDeleteKeyExW

func RegDeleteKeyExW(hKey HKEY, lpSubKey PWSTR, samDesired uint32, Reserved uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegDeleteKeyExW, libAdvapi32, "RegDeleteKeyExW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(samDesired), uintptr(Reserved))
	return WIN32_ERROR(ret)
}

func RegDeleteKeyTransactedA(hKey HKEY, lpSubKey PSTR, samDesired uint32, Reserved uint32, hTransaction HANDLE, pExtendedParameter unsafe.Pointer) WIN32_ERROR {
	addr := lazyAddr(&pRegDeleteKeyTransactedA, libAdvapi32, "RegDeleteKeyTransactedA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(samDesired), uintptr(Reserved), hTransaction, uintptr(pExtendedParameter))
	return WIN32_ERROR(ret)
}

var RegDeleteKeyTransacted = RegDeleteKeyTransactedW

func RegDeleteKeyTransactedW(hKey HKEY, lpSubKey PWSTR, samDesired uint32, Reserved uint32, hTransaction HANDLE, pExtendedParameter unsafe.Pointer) WIN32_ERROR {
	addr := lazyAddr(&pRegDeleteKeyTransactedW, libAdvapi32, "RegDeleteKeyTransactedW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(samDesired), uintptr(Reserved), hTransaction, uintptr(pExtendedParameter))
	return WIN32_ERROR(ret)
}

func RegDisableReflectionKey(hBase HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegDisableReflectionKey, libAdvapi32, "RegDisableReflectionKey")
	ret, _, _ := syscall.SyscallN(addr, hBase)
	return WIN32_ERROR(ret)
}

func RegEnableReflectionKey(hBase HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegEnableReflectionKey, libAdvapi32, "RegEnableReflectionKey")
	ret, _, _ := syscall.SyscallN(addr, hBase)
	return WIN32_ERROR(ret)
}

func RegQueryReflectionKey(hBase HKEY, bIsReflectionDisabled *BOOL) WIN32_ERROR {
	addr := lazyAddr(&pRegQueryReflectionKey, libAdvapi32, "RegQueryReflectionKey")
	ret, _, _ := syscall.SyscallN(addr, hBase, uintptr(unsafe.Pointer(bIsReflectionDisabled)))
	return WIN32_ERROR(ret)
}

func RegDeleteValueA(hKey HKEY, lpValueName PSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegDeleteValueA, libAdvapi32, "RegDeleteValueA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpValueName)))
	return WIN32_ERROR(ret)
}

var RegDeleteValue = RegDeleteValueW

func RegDeleteValueW(hKey HKEY, lpValueName PWSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegDeleteValueW, libAdvapi32, "RegDeleteValueW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpValueName)))
	return WIN32_ERROR(ret)
}

func RegEnumKeyA(hKey HKEY, dwIndex uint32, lpName PSTR, cchName uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegEnumKeyA, libAdvapi32, "RegEnumKeyA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(dwIndex), uintptr(unsafe.Pointer(lpName)), uintptr(cchName))
	return WIN32_ERROR(ret)
}

var RegEnumKey = RegEnumKeyW

func RegEnumKeyW(hKey HKEY, dwIndex uint32, lpName PWSTR, cchName uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegEnumKeyW, libAdvapi32, "RegEnumKeyW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(dwIndex), uintptr(unsafe.Pointer(lpName)), uintptr(cchName))
	return WIN32_ERROR(ret)
}

func RegEnumKeyExA(hKey HKEY, dwIndex uint32, lpName PSTR, lpcchName *uint32, lpReserved *uint32, lpClass PSTR, lpcchClass *uint32, lpftLastWriteTime *FILETIME) WIN32_ERROR {
	addr := lazyAddr(&pRegEnumKeyExA, libAdvapi32, "RegEnumKeyExA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(dwIndex), uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(lpcchName)), uintptr(unsafe.Pointer(lpReserved)), uintptr(unsafe.Pointer(lpClass)), uintptr(unsafe.Pointer(lpcchClass)), uintptr(unsafe.Pointer(lpftLastWriteTime)))
	return WIN32_ERROR(ret)
}

var RegEnumKeyEx = RegEnumKeyExW

func RegEnumKeyExW(hKey HKEY, dwIndex uint32, lpName PWSTR, lpcchName *uint32, lpReserved *uint32, lpClass PWSTR, lpcchClass *uint32, lpftLastWriteTime *FILETIME) WIN32_ERROR {
	addr := lazyAddr(&pRegEnumKeyExW, libAdvapi32, "RegEnumKeyExW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(dwIndex), uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(lpcchName)), uintptr(unsafe.Pointer(lpReserved)), uintptr(unsafe.Pointer(lpClass)), uintptr(unsafe.Pointer(lpcchClass)), uintptr(unsafe.Pointer(lpftLastWriteTime)))
	return WIN32_ERROR(ret)
}

func RegEnumValueA(hKey HKEY, dwIndex uint32, lpValueName PSTR, lpcchValueName *uint32, lpReserved *uint32, lpType *uint32, lpData *byte, lpcbData *uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegEnumValueA, libAdvapi32, "RegEnumValueA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(dwIndex), uintptr(unsafe.Pointer(lpValueName)), uintptr(unsafe.Pointer(lpcchValueName)), uintptr(unsafe.Pointer(lpReserved)), uintptr(unsafe.Pointer(lpType)), uintptr(unsafe.Pointer(lpData)), uintptr(unsafe.Pointer(lpcbData)))
	return WIN32_ERROR(ret)
}

var RegEnumValue = RegEnumValueW

func RegEnumValueW(hKey HKEY, dwIndex uint32, lpValueName PWSTR, lpcchValueName *uint32, lpReserved *uint32, lpType *uint32, lpData *byte, lpcbData *uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegEnumValueW, libAdvapi32, "RegEnumValueW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(dwIndex), uintptr(unsafe.Pointer(lpValueName)), uintptr(unsafe.Pointer(lpcchValueName)), uintptr(unsafe.Pointer(lpReserved)), uintptr(unsafe.Pointer(lpType)), uintptr(unsafe.Pointer(lpData)), uintptr(unsafe.Pointer(lpcbData)))
	return WIN32_ERROR(ret)
}

func RegFlushKey(hKey HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegFlushKey, libAdvapi32, "RegFlushKey")
	ret, _, _ := syscall.SyscallN(addr, hKey)
	return WIN32_ERROR(ret)
}

func RegGetKeySecurity(hKey HKEY, SecurityInformation uint32, pSecurityDescriptor PSECURITY_DESCRIPTOR, lpcbSecurityDescriptor *uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegGetKeySecurity, libAdvapi32, "RegGetKeySecurity")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(SecurityInformation), uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(unsafe.Pointer(lpcbSecurityDescriptor)))
	return WIN32_ERROR(ret)
}

func RegLoadKeyA(hKey HKEY, lpSubKey PSTR, lpFile PSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegLoadKeyA, libAdvapi32, "RegLoadKeyA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(lpFile)))
	return WIN32_ERROR(ret)
}

var RegLoadKey = RegLoadKeyW

func RegLoadKeyW(hKey HKEY, lpSubKey PWSTR, lpFile PWSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegLoadKeyW, libAdvapi32, "RegLoadKeyW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(lpFile)))
	return WIN32_ERROR(ret)
}

func RegNotifyChangeKeyValue(hKey HKEY, bWatchSubtree BOOL, dwNotifyFilter REG_NOTIFY_FILTER, hEvent HANDLE, fAsynchronous BOOL) WIN32_ERROR {
	addr := lazyAddr(&pRegNotifyChangeKeyValue, libAdvapi32, "RegNotifyChangeKeyValue")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(bWatchSubtree), uintptr(dwNotifyFilter), hEvent, uintptr(fAsynchronous))
	return WIN32_ERROR(ret)
}

func RegOpenKeyA(hKey HKEY, lpSubKey PSTR, phkResult *HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegOpenKeyA, libAdvapi32, "RegOpenKeyA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(phkResult)))
	return WIN32_ERROR(ret)
}

var RegOpenKey = RegOpenKeyW

func RegOpenKeyW(hKey HKEY, lpSubKey PWSTR, phkResult *HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegOpenKeyW, libAdvapi32, "RegOpenKeyW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(phkResult)))
	return WIN32_ERROR(ret)
}

func RegOpenKeyExA(hKey HKEY, lpSubKey PSTR, ulOptions uint32, samDesired REG_SAM_FLAGS, phkResult *HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegOpenKeyExA, libAdvapi32, "RegOpenKeyExA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(ulOptions), uintptr(samDesired), uintptr(unsafe.Pointer(phkResult)))
	return WIN32_ERROR(ret)
}

var RegOpenKeyEx = RegOpenKeyExW

func RegOpenKeyExW(hKey HKEY, lpSubKey PWSTR, ulOptions uint32, samDesired REG_SAM_FLAGS, phkResult *HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegOpenKeyExW, libAdvapi32, "RegOpenKeyExW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(ulOptions), uintptr(samDesired), uintptr(unsafe.Pointer(phkResult)))
	return WIN32_ERROR(ret)
}

func RegOpenKeyTransactedA(hKey HKEY, lpSubKey PSTR, ulOptions uint32, samDesired REG_SAM_FLAGS, phkResult *HKEY, hTransaction HANDLE, pExtendedParemeter unsafe.Pointer) WIN32_ERROR {
	addr := lazyAddr(&pRegOpenKeyTransactedA, libAdvapi32, "RegOpenKeyTransactedA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(ulOptions), uintptr(samDesired), uintptr(unsafe.Pointer(phkResult)), hTransaction, uintptr(pExtendedParemeter))
	return WIN32_ERROR(ret)
}

var RegOpenKeyTransacted = RegOpenKeyTransactedW

func RegOpenKeyTransactedW(hKey HKEY, lpSubKey PWSTR, ulOptions uint32, samDesired REG_SAM_FLAGS, phkResult *HKEY, hTransaction HANDLE, pExtendedParemeter unsafe.Pointer) WIN32_ERROR {
	addr := lazyAddr(&pRegOpenKeyTransactedW, libAdvapi32, "RegOpenKeyTransactedW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(ulOptions), uintptr(samDesired), uintptr(unsafe.Pointer(phkResult)), hTransaction, uintptr(pExtendedParemeter))
	return WIN32_ERROR(ret)
}

func RegQueryInfoKeyA(hKey HKEY, lpClass PSTR, lpcchClass *uint32, lpReserved *uint32, lpcSubKeys *uint32, lpcbMaxSubKeyLen *uint32, lpcbMaxClassLen *uint32, lpcValues *uint32, lpcbMaxValueNameLen *uint32, lpcbMaxValueLen *uint32, lpcbSecurityDescriptor *uint32, lpftLastWriteTime *FILETIME) WIN32_ERROR {
	addr := lazyAddr(&pRegQueryInfoKeyA, libAdvapi32, "RegQueryInfoKeyA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpClass)), uintptr(unsafe.Pointer(lpcchClass)), uintptr(unsafe.Pointer(lpReserved)), uintptr(unsafe.Pointer(lpcSubKeys)), uintptr(unsafe.Pointer(lpcbMaxSubKeyLen)), uintptr(unsafe.Pointer(lpcbMaxClassLen)), uintptr(unsafe.Pointer(lpcValues)), uintptr(unsafe.Pointer(lpcbMaxValueNameLen)), uintptr(unsafe.Pointer(lpcbMaxValueLen)), uintptr(unsafe.Pointer(lpcbSecurityDescriptor)), uintptr(unsafe.Pointer(lpftLastWriteTime)))
	return WIN32_ERROR(ret)
}

var RegQueryInfoKey = RegQueryInfoKeyW

func RegQueryInfoKeyW(hKey HKEY, lpClass PWSTR, lpcchClass *uint32, lpReserved *uint32, lpcSubKeys *uint32, lpcbMaxSubKeyLen *uint32, lpcbMaxClassLen *uint32, lpcValues *uint32, lpcbMaxValueNameLen *uint32, lpcbMaxValueLen *uint32, lpcbSecurityDescriptor *uint32, lpftLastWriteTime *FILETIME) WIN32_ERROR {
	addr := lazyAddr(&pRegQueryInfoKeyW, libAdvapi32, "RegQueryInfoKeyW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpClass)), uintptr(unsafe.Pointer(lpcchClass)), uintptr(unsafe.Pointer(lpReserved)), uintptr(unsafe.Pointer(lpcSubKeys)), uintptr(unsafe.Pointer(lpcbMaxSubKeyLen)), uintptr(unsafe.Pointer(lpcbMaxClassLen)), uintptr(unsafe.Pointer(lpcValues)), uintptr(unsafe.Pointer(lpcbMaxValueNameLen)), uintptr(unsafe.Pointer(lpcbMaxValueLen)), uintptr(unsafe.Pointer(lpcbSecurityDescriptor)), uintptr(unsafe.Pointer(lpftLastWriteTime)))
	return WIN32_ERROR(ret)
}

func RegQueryValueA(hKey HKEY, lpSubKey PSTR, lpData PSTR, lpcbData *int32) WIN32_ERROR {
	addr := lazyAddr(&pRegQueryValueA, libAdvapi32, "RegQueryValueA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(lpData)), uintptr(unsafe.Pointer(lpcbData)))
	return WIN32_ERROR(ret)
}

var RegQueryValue = RegQueryValueW

func RegQueryValueW(hKey HKEY, lpSubKey PWSTR, lpData PWSTR, lpcbData *int32) WIN32_ERROR {
	addr := lazyAddr(&pRegQueryValueW, libAdvapi32, "RegQueryValueW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(lpData)), uintptr(unsafe.Pointer(lpcbData)))
	return WIN32_ERROR(ret)
}

func RegQueryMultipleValuesA(hKey HKEY, val_list *VALENTA, num_vals uint32, lpValueBuf PSTR, ldwTotsize *uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegQueryMultipleValuesA, libAdvapi32, "RegQueryMultipleValuesA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(val_list)), uintptr(num_vals), uintptr(unsafe.Pointer(lpValueBuf)), uintptr(unsafe.Pointer(ldwTotsize)))
	return WIN32_ERROR(ret)
}

var RegQueryMultipleValues = RegQueryMultipleValuesW

func RegQueryMultipleValuesW(hKey HKEY, val_list *VALENTW, num_vals uint32, lpValueBuf PWSTR, ldwTotsize *uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegQueryMultipleValuesW, libAdvapi32, "RegQueryMultipleValuesW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(val_list)), uintptr(num_vals), uintptr(unsafe.Pointer(lpValueBuf)), uintptr(unsafe.Pointer(ldwTotsize)))
	return WIN32_ERROR(ret)
}

func RegQueryValueExA(hKey HKEY, lpValueName PSTR, lpReserved *uint32, lpType *REG_VALUE_TYPE, lpData *byte, lpcbData *uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegQueryValueExA, libAdvapi32, "RegQueryValueExA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpValueName)), uintptr(unsafe.Pointer(lpReserved)), uintptr(unsafe.Pointer(lpType)), uintptr(unsafe.Pointer(lpData)), uintptr(unsafe.Pointer(lpcbData)))
	return WIN32_ERROR(ret)
}

var RegQueryValueEx = RegQueryValueExW

func RegQueryValueExW(hKey HKEY, lpValueName PWSTR, lpReserved *uint32, lpType *REG_VALUE_TYPE, lpData *byte, lpcbData *uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegQueryValueExW, libAdvapi32, "RegQueryValueExW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpValueName)), uintptr(unsafe.Pointer(lpReserved)), uintptr(unsafe.Pointer(lpType)), uintptr(unsafe.Pointer(lpData)), uintptr(unsafe.Pointer(lpcbData)))
	return WIN32_ERROR(ret)
}

func RegReplaceKeyA(hKey HKEY, lpSubKey PSTR, lpNewFile PSTR, lpOldFile PSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegReplaceKeyA, libAdvapi32, "RegReplaceKeyA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(lpNewFile)), uintptr(unsafe.Pointer(lpOldFile)))
	return WIN32_ERROR(ret)
}

var RegReplaceKey = RegReplaceKeyW

func RegReplaceKeyW(hKey HKEY, lpSubKey PWSTR, lpNewFile PWSTR, lpOldFile PWSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegReplaceKeyW, libAdvapi32, "RegReplaceKeyW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(lpNewFile)), uintptr(unsafe.Pointer(lpOldFile)))
	return WIN32_ERROR(ret)
}

func RegRestoreKeyA(hKey HKEY, lpFile PSTR, dwFlags REG_RESTORE_KEY_FLAGS) WIN32_ERROR {
	addr := lazyAddr(&pRegRestoreKeyA, libAdvapi32, "RegRestoreKeyA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpFile)), uintptr(dwFlags))
	return WIN32_ERROR(ret)
}

var RegRestoreKey = RegRestoreKeyW

func RegRestoreKeyW(hKey HKEY, lpFile PWSTR, dwFlags REG_RESTORE_KEY_FLAGS) WIN32_ERROR {
	addr := lazyAddr(&pRegRestoreKeyW, libAdvapi32, "RegRestoreKeyW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpFile)), uintptr(dwFlags))
	return WIN32_ERROR(ret)
}

func RegRenameKey(hKey HKEY, lpSubKeyName PWSTR, lpNewKeyName PWSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegRenameKey, libAdvapi32, "RegRenameKey")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKeyName)), uintptr(unsafe.Pointer(lpNewKeyName)))
	return WIN32_ERROR(ret)
}

func RegSaveKeyA(hKey HKEY, lpFile PSTR, lpSecurityAttributes *SECURITY_ATTRIBUTES) WIN32_ERROR {
	addr := lazyAddr(&pRegSaveKeyA, libAdvapi32, "RegSaveKeyA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpFile)), uintptr(unsafe.Pointer(lpSecurityAttributes)))
	return WIN32_ERROR(ret)
}

var RegSaveKey = RegSaveKeyW

func RegSaveKeyW(hKey HKEY, lpFile PWSTR, lpSecurityAttributes *SECURITY_ATTRIBUTES) WIN32_ERROR {
	addr := lazyAddr(&pRegSaveKeyW, libAdvapi32, "RegSaveKeyW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpFile)), uintptr(unsafe.Pointer(lpSecurityAttributes)))
	return WIN32_ERROR(ret)
}

func RegSetKeySecurity(hKey HKEY, SecurityInformation uint32, pSecurityDescriptor PSECURITY_DESCRIPTOR) WIN32_ERROR {
	addr := lazyAddr(&pRegSetKeySecurity, libAdvapi32, "RegSetKeySecurity")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(SecurityInformation), uintptr(unsafe.Pointer(pSecurityDescriptor)))
	return WIN32_ERROR(ret)
}

func RegSetValueA(hKey HKEY, lpSubKey PSTR, dwType REG_VALUE_TYPE, lpData PSTR, cbData uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegSetValueA, libAdvapi32, "RegSetValueA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(dwType), uintptr(unsafe.Pointer(lpData)), uintptr(cbData))
	return WIN32_ERROR(ret)
}

var RegSetValue = RegSetValueW

func RegSetValueW(hKey HKEY, lpSubKey PWSTR, dwType REG_VALUE_TYPE, lpData PWSTR, cbData uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegSetValueW, libAdvapi32, "RegSetValueW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(dwType), uintptr(unsafe.Pointer(lpData)), uintptr(cbData))
	return WIN32_ERROR(ret)
}

func RegSetValueExA(hKey HKEY, lpValueName PSTR, Reserved uint32, dwType REG_VALUE_TYPE, lpData *byte, cbData uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegSetValueExA, libAdvapi32, "RegSetValueExA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpValueName)), uintptr(Reserved), uintptr(dwType), uintptr(unsafe.Pointer(lpData)), uintptr(cbData))
	return WIN32_ERROR(ret)
}

var RegSetValueEx = RegSetValueExW

func RegSetValueExW(hKey HKEY, lpValueName PWSTR, Reserved uint32, dwType REG_VALUE_TYPE, lpData *byte, cbData uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegSetValueExW, libAdvapi32, "RegSetValueExW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpValueName)), uintptr(Reserved), uintptr(dwType), uintptr(unsafe.Pointer(lpData)), uintptr(cbData))
	return WIN32_ERROR(ret)
}

func RegUnLoadKeyA(hKey HKEY, lpSubKey PSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegUnLoadKeyA, libAdvapi32, "RegUnLoadKeyA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)))
	return WIN32_ERROR(ret)
}

var RegUnLoadKey = RegUnLoadKeyW

func RegUnLoadKeyW(hKey HKEY, lpSubKey PWSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegUnLoadKeyW, libAdvapi32, "RegUnLoadKeyW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)))
	return WIN32_ERROR(ret)
}

func RegDeleteKeyValueA(hKey HKEY, lpSubKey PSTR, lpValueName PSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegDeleteKeyValueA, libAdvapi32, "RegDeleteKeyValueA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(lpValueName)))
	return WIN32_ERROR(ret)
}

var RegDeleteKeyValue = RegDeleteKeyValueW

func RegDeleteKeyValueW(hKey HKEY, lpSubKey PWSTR, lpValueName PWSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegDeleteKeyValueW, libAdvapi32, "RegDeleteKeyValueW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(lpValueName)))
	return WIN32_ERROR(ret)
}

func RegSetKeyValueA(hKey HKEY, lpSubKey PSTR, lpValueName PSTR, dwType uint32, lpData unsafe.Pointer, cbData uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegSetKeyValueA, libAdvapi32, "RegSetKeyValueA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(lpValueName)), uintptr(dwType), uintptr(lpData), uintptr(cbData))
	return WIN32_ERROR(ret)
}

var RegSetKeyValue = RegSetKeyValueW

func RegSetKeyValueW(hKey HKEY, lpSubKey PWSTR, lpValueName PWSTR, dwType uint32, lpData unsafe.Pointer, cbData uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegSetKeyValueW, libAdvapi32, "RegSetKeyValueW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(lpValueName)), uintptr(dwType), uintptr(lpData), uintptr(cbData))
	return WIN32_ERROR(ret)
}

func RegDeleteTreeA(hKey HKEY, lpSubKey PSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegDeleteTreeA, libAdvapi32, "RegDeleteTreeA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)))
	return WIN32_ERROR(ret)
}

var RegDeleteTree = RegDeleteTreeW

func RegDeleteTreeW(hKey HKEY, lpSubKey PWSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegDeleteTreeW, libAdvapi32, "RegDeleteTreeW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpSubKey)))
	return WIN32_ERROR(ret)
}

func RegCopyTreeA(hKeySrc HKEY, lpSubKey PSTR, hKeyDest HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegCopyTreeA, libAdvapi32, "RegCopyTreeA")
	ret, _, _ := syscall.SyscallN(addr, hKeySrc, uintptr(unsafe.Pointer(lpSubKey)), hKeyDest)
	return WIN32_ERROR(ret)
}

func RegGetValueA(hkey HKEY, lpSubKey PSTR, lpValue PSTR, dwFlags RRF_RT, pdwType *uint32, pvData unsafe.Pointer, pcbData *uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegGetValueA, libAdvapi32, "RegGetValueA")
	ret, _, _ := syscall.SyscallN(addr, hkey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(lpValue)), uintptr(dwFlags), uintptr(unsafe.Pointer(pdwType)), uintptr(pvData), uintptr(unsafe.Pointer(pcbData)))
	return WIN32_ERROR(ret)
}

var RegGetValue = RegGetValueW

func RegGetValueW(hkey HKEY, lpSubKey PWSTR, lpValue PWSTR, dwFlags RRF_RT, pdwType *uint32, pvData unsafe.Pointer, pcbData *uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegGetValueW, libAdvapi32, "RegGetValueW")
	ret, _, _ := syscall.SyscallN(addr, hkey, uintptr(unsafe.Pointer(lpSubKey)), uintptr(unsafe.Pointer(lpValue)), uintptr(dwFlags), uintptr(unsafe.Pointer(pdwType)), uintptr(pvData), uintptr(unsafe.Pointer(pcbData)))
	return WIN32_ERROR(ret)
}

var RegCopyTree = RegCopyTreeW

func RegCopyTreeW(hKeySrc HKEY, lpSubKey PWSTR, hKeyDest HKEY) WIN32_ERROR {
	addr := lazyAddr(&pRegCopyTreeW, libAdvapi32, "RegCopyTreeW")
	ret, _, _ := syscall.SyscallN(addr, hKeySrc, uintptr(unsafe.Pointer(lpSubKey)), hKeyDest)
	return WIN32_ERROR(ret)
}

func RegLoadMUIStringA(hKey HKEY, pszValue PSTR, pszOutBuf PSTR, cbOutBuf uint32, pcbData *uint32, Flags uint32, pszDirectory PSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegLoadMUIStringA, libAdvapi32, "RegLoadMUIStringA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(pszValue)), uintptr(unsafe.Pointer(pszOutBuf)), uintptr(cbOutBuf), uintptr(unsafe.Pointer(pcbData)), uintptr(Flags), uintptr(unsafe.Pointer(pszDirectory)))
	return WIN32_ERROR(ret)
}

var RegLoadMUIString = RegLoadMUIStringW

func RegLoadMUIStringW(hKey HKEY, pszValue PWSTR, pszOutBuf PWSTR, cbOutBuf uint32, pcbData *uint32, Flags uint32, pszDirectory PWSTR) WIN32_ERROR {
	addr := lazyAddr(&pRegLoadMUIStringW, libAdvapi32, "RegLoadMUIStringW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(pszValue)), uintptr(unsafe.Pointer(pszOutBuf)), uintptr(cbOutBuf), uintptr(unsafe.Pointer(pcbData)), uintptr(Flags), uintptr(unsafe.Pointer(pszDirectory)))
	return WIN32_ERROR(ret)
}

func RegLoadAppKeyA(lpFile PSTR, phkResult *HKEY, samDesired uint32, dwOptions uint32, Reserved uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegLoadAppKeyA, libAdvapi32, "RegLoadAppKeyA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFile)), uintptr(unsafe.Pointer(phkResult)), uintptr(samDesired), uintptr(dwOptions), uintptr(Reserved))
	return WIN32_ERROR(ret)
}

var RegLoadAppKey = RegLoadAppKeyW

func RegLoadAppKeyW(lpFile PWSTR, phkResult *HKEY, samDesired uint32, dwOptions uint32, Reserved uint32) WIN32_ERROR {
	addr := lazyAddr(&pRegLoadAppKeyW, libAdvapi32, "RegLoadAppKeyW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFile)), uintptr(unsafe.Pointer(phkResult)), uintptr(samDesired), uintptr(dwOptions), uintptr(Reserved))
	return WIN32_ERROR(ret)
}

func RegSaveKeyExA(hKey HKEY, lpFile PSTR, lpSecurityAttributes *SECURITY_ATTRIBUTES, Flags REG_SAVE_FORMAT) WIN32_ERROR {
	addr := lazyAddr(&pRegSaveKeyExA, libAdvapi32, "RegSaveKeyExA")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpFile)), uintptr(unsafe.Pointer(lpSecurityAttributes)), uintptr(Flags))
	return WIN32_ERROR(ret)
}

var RegSaveKeyEx = RegSaveKeyExW

func RegSaveKeyExW(hKey HKEY, lpFile PWSTR, lpSecurityAttributes *SECURITY_ATTRIBUTES, Flags REG_SAVE_FORMAT) WIN32_ERROR {
	addr := lazyAddr(&pRegSaveKeyExW, libAdvapi32, "RegSaveKeyExW")
	ret, _, _ := syscall.SyscallN(addr, hKey, uintptr(unsafe.Pointer(lpFile)), uintptr(unsafe.Pointer(lpSecurityAttributes)), uintptr(Flags))
	return WIN32_ERROR(ret)
}
