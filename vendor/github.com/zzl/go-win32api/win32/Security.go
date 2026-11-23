package win32

import (
	"syscall"
	"unsafe"
)

type (
	HDIAGNOSTIC_DATA_QUERY_SESSION         = uintptr
	HDIAGNOSTIC_REPORT                     = uintptr
	HDIAGNOSTIC_EVENT_TAG_DESCRIPTION      = uintptr
	HDIAGNOSTIC_EVENT_PRODUCER_DESCRIPTION = uintptr
	HDIAGNOSTIC_EVENT_CATEGORY_DESCRIPTION = uintptr
	HDIAGNOSTIC_RECORD                     = uintptr
	NCRYPT_DESCRIPTOR_HANDLE               = uintptr
	NCRYPT_STREAM_HANDLE                   = uintptr
	SAFER_LEVEL_HANDLE                     = uintptr
	SC_HANDLE                              = uintptr
	PSECURITY_DESCRIPTOR                   = unsafe.Pointer
)

const (
	WszCERTENROLLSHAREPATH               string = "CertSrv\\CertEnroll"
	CwcHRESULTSTRING                     uint32 = 0x28
	SzLBRACE                             string = "{"
	SzRBRACE                             string = "}"
	WszLBRACE                            string = "{"
	WszRBRACE                            string = "}"
	SzLPAREN                             string = "("
	SzRPAREN                             string = ")"
	WszLPAREN                            string = "("
	WszRPAREN                            string = ")"
	CVT_SECONDS                          uint32 = 0x1
	CwcFILENAMESUFFIXMAX                 uint32 = 0x14
	WszFCSAPARM_SERVERDNSNAME            string = "%1"
	WszFCSAPARM_SERVERSHORTNAME          string = "%2"
	WszFCSAPARM_SANITIZEDCANAME          string = "%3"
	WszFCSAPARM_CERTFILENAMESUFFIX       string = "%4"
	WszFCSAPARM_DOMAINDN                 string = "%5"
	WszFCSAPARM_CONFIGDN                 string = "%6"
	WszFCSAPARM_SANITIZEDCANAMEHASH      string = "%7"
	WszFCSAPARM_CRLFILENAMESUFFIX        string = "%8"
	WszFCSAPARM_CRLDELTAFILENAMESUFFIX   string = "%9"
	WszFCSAPARM_DSCRLATTRIBUTE           string = "%10"
	WszFCSAPARM_DSCACERTATTRIBUTE        string = "%11"
	WszFCSAPARM_DSUSERCERTATTRIBUTE      string = "%12"
	WszFCSAPARM_DSKRACERTATTRIBUTE       string = "%13"
	WszFCSAPARM_DSCROSSCERTPAIRATTRIBUTE string = "%14"
)

// enums

// enum
// flags
type TOKEN_PRIVILEGES_ATTRIBUTES uint32

const (
	SE_PRIVILEGE_ENABLED            TOKEN_PRIVILEGES_ATTRIBUTES = 2
	SE_PRIVILEGE_ENABLED_BY_DEFAULT TOKEN_PRIVILEGES_ATTRIBUTES = 1
	SE_PRIVILEGE_REMOVED            TOKEN_PRIVILEGES_ATTRIBUTES = 4
	SE_PRIVILEGE_USED_FOR_ACCESS    TOKEN_PRIVILEGES_ATTRIBUTES = 2147483648
)

// enum
type LOGON32_PROVIDER uint32

const (
	LOGON32_PROVIDER_DEFAULT LOGON32_PROVIDER = 0
	LOGON32_PROVIDER_WINNT50 LOGON32_PROVIDER = 3
	LOGON32_PROVIDER_WINNT40 LOGON32_PROVIDER = 2
)

// enum
// flags
type CREATE_RESTRICTED_TOKEN_FLAGS uint32

const (
	DISABLE_MAX_PRIVILEGE CREATE_RESTRICTED_TOKEN_FLAGS = 1
	SANDBOX_INERT         CREATE_RESTRICTED_TOKEN_FLAGS = 2
	LUA_TOKEN             CREATE_RESTRICTED_TOKEN_FLAGS = 4
	WRITE_RESTRICTED      CREATE_RESTRICTED_TOKEN_FLAGS = 8
)

// enum
type LOGON32_LOGON uint32

const (
	LOGON32_LOGON_BATCH             LOGON32_LOGON = 4
	LOGON32_LOGON_INTERACTIVE       LOGON32_LOGON = 2
	LOGON32_LOGON_NETWORK           LOGON32_LOGON = 3
	LOGON32_LOGON_NETWORK_CLEARTEXT LOGON32_LOGON = 8
	LOGON32_LOGON_NEW_CREDENTIALS   LOGON32_LOGON = 9
	LOGON32_LOGON_SERVICE           LOGON32_LOGON = 5
	LOGON32_LOGON_UNLOCK            LOGON32_LOGON = 7
)

// enum
// flags
type ACE_FLAGS uint32

const (
	CONTAINER_INHERIT_ACE              ACE_FLAGS = 2
	FAILED_ACCESS_ACE_FLAG             ACE_FLAGS = 128
	INHERIT_ONLY_ACE                   ACE_FLAGS = 8
	INHERITED_ACE                      ACE_FLAGS = 16
	NO_PROPAGATE_INHERIT_ACE           ACE_FLAGS = 4
	OBJECT_INHERIT_ACE                 ACE_FLAGS = 1
	SUCCESSFUL_ACCESS_ACE_FLAG         ACE_FLAGS = 64
	SUB_CONTAINERS_AND_OBJECTS_INHERIT ACE_FLAGS = 3
	SUB_CONTAINERS_ONLY_INHERIT        ACE_FLAGS = 2
	SUB_OBJECTS_ONLY_INHERIT           ACE_FLAGS = 1
	INHERIT_NO_PROPAGATE               ACE_FLAGS = 4
	INHERIT_ONLY                       ACE_FLAGS = 8
	NO_INHERITANCE                     ACE_FLAGS = 0
)

// enum
// flags
type OBJECT_SECURITY_INFORMATION uint32

const (
	ATTRIBUTE_SECURITY_INFORMATION        OBJECT_SECURITY_INFORMATION = 32
	BACKUP_SECURITY_INFORMATION           OBJECT_SECURITY_INFORMATION = 65536
	DACL_SECURITY_INFORMATION             OBJECT_SECURITY_INFORMATION = 4
	GROUP_SECURITY_INFORMATION            OBJECT_SECURITY_INFORMATION = 2
	LABEL_SECURITY_INFORMATION            OBJECT_SECURITY_INFORMATION = 16
	OWNER_SECURITY_INFORMATION            OBJECT_SECURITY_INFORMATION = 1
	PROTECTED_DACL_SECURITY_INFORMATION   OBJECT_SECURITY_INFORMATION = 2147483648
	PROTECTED_SACL_SECURITY_INFORMATION   OBJECT_SECURITY_INFORMATION = 1073741824
	SACL_SECURITY_INFORMATION             OBJECT_SECURITY_INFORMATION = 8
	SCOPE_SECURITY_INFORMATION            OBJECT_SECURITY_INFORMATION = 64
	UNPROTECTED_DACL_SECURITY_INFORMATION OBJECT_SECURITY_INFORMATION = 536870912
	UNPROTECTED_SACL_SECURITY_INFORMATION OBJECT_SECURITY_INFORMATION = 268435456
)

// enum
// flags
type SECURITY_AUTO_INHERIT_FLAGS uint32

const (
	SEF_AVOID_OWNER_CHECK             SECURITY_AUTO_INHERIT_FLAGS = 16
	SEF_AVOID_OWNER_RESTRICTION       SECURITY_AUTO_INHERIT_FLAGS = 4096
	SEF_AVOID_PRIVILEGE_CHECK         SECURITY_AUTO_INHERIT_FLAGS = 8
	SEF_DACL_AUTO_INHERIT             SECURITY_AUTO_INHERIT_FLAGS = 1
	SEF_DEFAULT_DESCRIPTOR_FOR_OBJECT SECURITY_AUTO_INHERIT_FLAGS = 4
	SEF_DEFAULT_GROUP_FROM_PARENT     SECURITY_AUTO_INHERIT_FLAGS = 64
	SEF_DEFAULT_OWNER_FROM_PARENT     SECURITY_AUTO_INHERIT_FLAGS = 32
	SEF_MACL_NO_EXECUTE_UP            SECURITY_AUTO_INHERIT_FLAGS = 1024
	SEF_MACL_NO_READ_UP               SECURITY_AUTO_INHERIT_FLAGS = 512
	SEF_MACL_NO_WRITE_UP              SECURITY_AUTO_INHERIT_FLAGS = 256
	SEF_SACL_AUTO_INHERIT             SECURITY_AUTO_INHERIT_FLAGS = 2
)

// enum
type ACE_REVISION uint32

const (
	ACL_REVISION    ACE_REVISION = 2
	ACL_REVISION_DS ACE_REVISION = 4
)

// enum
type TOKEN_MANDATORY_POLICY_ID uint32

const (
	TOKEN_MANDATORY_POLICY_OFF             TOKEN_MANDATORY_POLICY_ID = 0
	TOKEN_MANDATORY_POLICY_NO_WRITE_UP     TOKEN_MANDATORY_POLICY_ID = 1
	TOKEN_MANDATORY_POLICY_NEW_PROCESS_MIN TOKEN_MANDATORY_POLICY_ID = 2
	TOKEN_MANDATORY_POLICY_VALID_MASK      TOKEN_MANDATORY_POLICY_ID = 3
)

// enum
// flags
type SYSTEM_AUDIT_OBJECT_ACE_FLAGS uint32

const (
	ACE_OBJECT_TYPE_PRESENT           SYSTEM_AUDIT_OBJECT_ACE_FLAGS = 1
	ACE_INHERITED_OBJECT_TYPE_PRESENT SYSTEM_AUDIT_OBJECT_ACE_FLAGS = 2
)

// enum
// flags
type CLAIM_SECURITY_ATTRIBUTE_FLAGS uint32

const (
	CLAIM_SECURITY_ATTRIBUTE_NON_INHERITABLE      CLAIM_SECURITY_ATTRIBUTE_FLAGS = 1
	CLAIM_SECURITY_ATTRIBUTE_VALUE_CASE_SENSITIVE CLAIM_SECURITY_ATTRIBUTE_FLAGS = 2
	CLAIM_SECURITY_ATTRIBUTE_USE_FOR_DENY_ONLY    CLAIM_SECURITY_ATTRIBUTE_FLAGS = 4
	CLAIM_SECURITY_ATTRIBUTE_DISABLED_BY_DEFAULT  CLAIM_SECURITY_ATTRIBUTE_FLAGS = 8
	CLAIM_SECURITY_ATTRIBUTE_DISABLED             CLAIM_SECURITY_ATTRIBUTE_FLAGS = 16
	CLAIM_SECURITY_ATTRIBUTE_MANDATORY            CLAIM_SECURITY_ATTRIBUTE_FLAGS = 32
)

// enum
type CLAIM_SECURITY_ATTRIBUTE_VALUE_TYPE uint16

const (
	CLAIM_SECURITY_ATTRIBUTE_TYPE_INT64        CLAIM_SECURITY_ATTRIBUTE_VALUE_TYPE = 1
	CLAIM_SECURITY_ATTRIBUTE_TYPE_UINT64       CLAIM_SECURITY_ATTRIBUTE_VALUE_TYPE = 2
	CLAIM_SECURITY_ATTRIBUTE_TYPE_STRING       CLAIM_SECURITY_ATTRIBUTE_VALUE_TYPE = 3
	CLAIM_SECURITY_ATTRIBUTE_TYPE_OCTET_STRING CLAIM_SECURITY_ATTRIBUTE_VALUE_TYPE = 16
	CLAIM_SECURITY_ATTRIBUTE_TYPE_FQBN         CLAIM_SECURITY_ATTRIBUTE_VALUE_TYPE = 4
	CLAIM_SECURITY_ATTRIBUTE_TYPE_SID          CLAIM_SECURITY_ATTRIBUTE_VALUE_TYPE = 5
	CLAIM_SECURITY_ATTRIBUTE_TYPE_BOOLEAN      CLAIM_SECURITY_ATTRIBUTE_VALUE_TYPE = 6
)

// enum
// flags
type TOKEN_ACCESS_MASK uint32

const (
	TOKEN_DELETE                 TOKEN_ACCESS_MASK = 65536
	TOKEN_READ_CONTROL           TOKEN_ACCESS_MASK = 131072
	TOKEN_WRITE_DAC              TOKEN_ACCESS_MASK = 262144
	TOKEN_WRITE_OWNER            TOKEN_ACCESS_MASK = 524288
	TOKEN_ACCESS_SYSTEM_SECURITY TOKEN_ACCESS_MASK = 16777216
	TOKEN_ASSIGN_PRIMARY         TOKEN_ACCESS_MASK = 1
	TOKEN_DUPLICATE              TOKEN_ACCESS_MASK = 2
	TOKEN_IMPERSONATE            TOKEN_ACCESS_MASK = 4
	TOKEN_QUERY                  TOKEN_ACCESS_MASK = 8
	TOKEN_QUERY_SOURCE           TOKEN_ACCESS_MASK = 16
	TOKEN_ADJUST_PRIVILEGES      TOKEN_ACCESS_MASK = 32
	TOKEN_ADJUST_GROUPS          TOKEN_ACCESS_MASK = 64
	TOKEN_ADJUST_DEFAULT         TOKEN_ACCESS_MASK = 128
	TOKEN_ADJUST_SESSIONID       TOKEN_ACCESS_MASK = 256
	TOKEN_ALL_ACCESS             TOKEN_ACCESS_MASK = 983295
)

// enum
type ENUM_PERIOD int32

const (
	ENUM_PERIOD_INVALID ENUM_PERIOD = -1
	ENUM_PERIOD_SECONDS ENUM_PERIOD = 0
	ENUM_PERIOD_MINUTES ENUM_PERIOD = 1
	ENUM_PERIOD_HOURS   ENUM_PERIOD = 2
	ENUM_PERIOD_DAYS    ENUM_PERIOD = 3
	ENUM_PERIOD_WEEKS   ENUM_PERIOD = 4
	ENUM_PERIOD_MONTHS  ENUM_PERIOD = 5
	ENUM_PERIOD_YEARS   ENUM_PERIOD = 6
)

// enum
type SID_NAME_USE int32

const (
	SidTypeUser           SID_NAME_USE = 1
	SidTypeGroup          SID_NAME_USE = 2
	SidTypeDomain         SID_NAME_USE = 3
	SidTypeAlias          SID_NAME_USE = 4
	SidTypeWellKnownGroup SID_NAME_USE = 5
	SidTypeDeletedAccount SID_NAME_USE = 6
	SidTypeInvalid        SID_NAME_USE = 7
	SidTypeUnknown        SID_NAME_USE = 8
	SidTypeComputer       SID_NAME_USE = 9
	SidTypeLabel          SID_NAME_USE = 10
	SidTypeLogonSession   SID_NAME_USE = 11
)

// enum
type WELL_KNOWN_SID_TYPE int32

const (
	WinNullSid                                    WELL_KNOWN_SID_TYPE = 0
	WinWorldSid                                   WELL_KNOWN_SID_TYPE = 1
	WinLocalSid                                   WELL_KNOWN_SID_TYPE = 2
	WinCreatorOwnerSid                            WELL_KNOWN_SID_TYPE = 3
	WinCreatorGroupSid                            WELL_KNOWN_SID_TYPE = 4
	WinCreatorOwnerServerSid                      WELL_KNOWN_SID_TYPE = 5
	WinCreatorGroupServerSid                      WELL_KNOWN_SID_TYPE = 6
	WinNtAuthoritySid                             WELL_KNOWN_SID_TYPE = 7
	WinDialupSid                                  WELL_KNOWN_SID_TYPE = 8
	WinNetworkSid                                 WELL_KNOWN_SID_TYPE = 9
	WinBatchSid                                   WELL_KNOWN_SID_TYPE = 10
	WinInteractiveSid                             WELL_KNOWN_SID_TYPE = 11
	WinServiceSid                                 WELL_KNOWN_SID_TYPE = 12
	WinAnonymousSid                               WELL_KNOWN_SID_TYPE = 13
	WinProxySid                                   WELL_KNOWN_SID_TYPE = 14
	WinEnterpriseControllersSid                   WELL_KNOWN_SID_TYPE = 15
	WinSelfSid                                    WELL_KNOWN_SID_TYPE = 16
	WinAuthenticatedUserSid                       WELL_KNOWN_SID_TYPE = 17
	WinRestrictedCodeSid                          WELL_KNOWN_SID_TYPE = 18
	WinTerminalServerSid                          WELL_KNOWN_SID_TYPE = 19
	WinRemoteLogonIdSid                           WELL_KNOWN_SID_TYPE = 20
	WinLogonIdsSid                                WELL_KNOWN_SID_TYPE = 21
	WinLocalSystemSid                             WELL_KNOWN_SID_TYPE = 22
	WinLocalServiceSid                            WELL_KNOWN_SID_TYPE = 23
	WinNetworkServiceSid                          WELL_KNOWN_SID_TYPE = 24
	WinBuiltinDomainSid                           WELL_KNOWN_SID_TYPE = 25
	WinBuiltinAdministratorsSid                   WELL_KNOWN_SID_TYPE = 26
	WinBuiltinUsersSid                            WELL_KNOWN_SID_TYPE = 27
	WinBuiltinGuestsSid                           WELL_KNOWN_SID_TYPE = 28
	WinBuiltinPowerUsersSid                       WELL_KNOWN_SID_TYPE = 29
	WinBuiltinAccountOperatorsSid                 WELL_KNOWN_SID_TYPE = 30
	WinBuiltinSystemOperatorsSid                  WELL_KNOWN_SID_TYPE = 31
	WinBuiltinPrintOperatorsSid                   WELL_KNOWN_SID_TYPE = 32
	WinBuiltinBackupOperatorsSid                  WELL_KNOWN_SID_TYPE = 33
	WinBuiltinReplicatorSid                       WELL_KNOWN_SID_TYPE = 34
	WinBuiltinPreWindows2000CompatibleAccessSid   WELL_KNOWN_SID_TYPE = 35
	WinBuiltinRemoteDesktopUsersSid               WELL_KNOWN_SID_TYPE = 36
	WinBuiltinNetworkConfigurationOperatorsSid    WELL_KNOWN_SID_TYPE = 37
	WinAccountAdministratorSid                    WELL_KNOWN_SID_TYPE = 38
	WinAccountGuestSid                            WELL_KNOWN_SID_TYPE = 39
	WinAccountKrbtgtSid                           WELL_KNOWN_SID_TYPE = 40
	WinAccountDomainAdminsSid                     WELL_KNOWN_SID_TYPE = 41
	WinAccountDomainUsersSid                      WELL_KNOWN_SID_TYPE = 42
	WinAccountDomainGuestsSid                     WELL_KNOWN_SID_TYPE = 43
	WinAccountComputersSid                        WELL_KNOWN_SID_TYPE = 44
	WinAccountControllersSid                      WELL_KNOWN_SID_TYPE = 45
	WinAccountCertAdminsSid                       WELL_KNOWN_SID_TYPE = 46
	WinAccountSchemaAdminsSid                     WELL_KNOWN_SID_TYPE = 47
	WinAccountEnterpriseAdminsSid                 WELL_KNOWN_SID_TYPE = 48
	WinAccountPolicyAdminsSid                     WELL_KNOWN_SID_TYPE = 49
	WinAccountRasAndIasServersSid                 WELL_KNOWN_SID_TYPE = 50
	WinNTLMAuthenticationSid                      WELL_KNOWN_SID_TYPE = 51
	WinDigestAuthenticationSid                    WELL_KNOWN_SID_TYPE = 52
	WinSChannelAuthenticationSid                  WELL_KNOWN_SID_TYPE = 53
	WinThisOrganizationSid                        WELL_KNOWN_SID_TYPE = 54
	WinOtherOrganizationSid                       WELL_KNOWN_SID_TYPE = 55
	WinBuiltinIncomingForestTrustBuildersSid      WELL_KNOWN_SID_TYPE = 56
	WinBuiltinPerfMonitoringUsersSid              WELL_KNOWN_SID_TYPE = 57
	WinBuiltinPerfLoggingUsersSid                 WELL_KNOWN_SID_TYPE = 58
	WinBuiltinAuthorizationAccessSid              WELL_KNOWN_SID_TYPE = 59
	WinBuiltinTerminalServerLicenseServersSid     WELL_KNOWN_SID_TYPE = 60
	WinBuiltinDCOMUsersSid                        WELL_KNOWN_SID_TYPE = 61
	WinBuiltinIUsersSid                           WELL_KNOWN_SID_TYPE = 62
	WinIUserSid                                   WELL_KNOWN_SID_TYPE = 63
	WinBuiltinCryptoOperatorsSid                  WELL_KNOWN_SID_TYPE = 64
	WinUntrustedLabelSid                          WELL_KNOWN_SID_TYPE = 65
	WinLowLabelSid                                WELL_KNOWN_SID_TYPE = 66
	WinMediumLabelSid                             WELL_KNOWN_SID_TYPE = 67
	WinHighLabelSid                               WELL_KNOWN_SID_TYPE = 68
	WinSystemLabelSid                             WELL_KNOWN_SID_TYPE = 69
	WinWriteRestrictedCodeSid                     WELL_KNOWN_SID_TYPE = 70
	WinCreatorOwnerRightsSid                      WELL_KNOWN_SID_TYPE = 71
	WinCacheablePrincipalsGroupSid                WELL_KNOWN_SID_TYPE = 72
	WinNonCacheablePrincipalsGroupSid             WELL_KNOWN_SID_TYPE = 73
	WinEnterpriseReadonlyControllersSid           WELL_KNOWN_SID_TYPE = 74
	WinAccountReadonlyControllersSid              WELL_KNOWN_SID_TYPE = 75
	WinBuiltinEventLogReadersGroup                WELL_KNOWN_SID_TYPE = 76
	WinNewEnterpriseReadonlyControllersSid        WELL_KNOWN_SID_TYPE = 77
	WinBuiltinCertSvcDComAccessGroup              WELL_KNOWN_SID_TYPE = 78
	WinMediumPlusLabelSid                         WELL_KNOWN_SID_TYPE = 79
	WinLocalLogonSid                              WELL_KNOWN_SID_TYPE = 80
	WinConsoleLogonSid                            WELL_KNOWN_SID_TYPE = 81
	WinThisOrganizationCertificateSid             WELL_KNOWN_SID_TYPE = 82
	WinApplicationPackageAuthoritySid             WELL_KNOWN_SID_TYPE = 83
	WinBuiltinAnyPackageSid                       WELL_KNOWN_SID_TYPE = 84
	WinCapabilityInternetClientSid                WELL_KNOWN_SID_TYPE = 85
	WinCapabilityInternetClientServerSid          WELL_KNOWN_SID_TYPE = 86
	WinCapabilityPrivateNetworkClientServerSid    WELL_KNOWN_SID_TYPE = 87
	WinCapabilityPicturesLibrarySid               WELL_KNOWN_SID_TYPE = 88
	WinCapabilityVideosLibrarySid                 WELL_KNOWN_SID_TYPE = 89
	WinCapabilityMusicLibrarySid                  WELL_KNOWN_SID_TYPE = 90
	WinCapabilityDocumentsLibrarySid              WELL_KNOWN_SID_TYPE = 91
	WinCapabilitySharedUserCertificatesSid        WELL_KNOWN_SID_TYPE = 92
	WinCapabilityEnterpriseAuthenticationSid      WELL_KNOWN_SID_TYPE = 93
	WinCapabilityRemovableStorageSid              WELL_KNOWN_SID_TYPE = 94
	WinBuiltinRDSRemoteAccessServersSid           WELL_KNOWN_SID_TYPE = 95
	WinBuiltinRDSEndpointServersSid               WELL_KNOWN_SID_TYPE = 96
	WinBuiltinRDSManagementServersSid             WELL_KNOWN_SID_TYPE = 97
	WinUserModeDriversSid                         WELL_KNOWN_SID_TYPE = 98
	WinBuiltinHyperVAdminsSid                     WELL_KNOWN_SID_TYPE = 99
	WinAccountCloneableControllersSid             WELL_KNOWN_SID_TYPE = 100
	WinBuiltinAccessControlAssistanceOperatorsSid WELL_KNOWN_SID_TYPE = 101
	WinBuiltinRemoteManagementUsersSid            WELL_KNOWN_SID_TYPE = 102
	WinAuthenticationAuthorityAssertedSid         WELL_KNOWN_SID_TYPE = 103
	WinAuthenticationServiceAssertedSid           WELL_KNOWN_SID_TYPE = 104
	WinLocalAccountSid                            WELL_KNOWN_SID_TYPE = 105
	WinLocalAccountAndAdministratorSid            WELL_KNOWN_SID_TYPE = 106
	WinAccountProtectedUsersSid                   WELL_KNOWN_SID_TYPE = 107
	WinCapabilityAppointmentsSid                  WELL_KNOWN_SID_TYPE = 108
	WinCapabilityContactsSid                      WELL_KNOWN_SID_TYPE = 109
	WinAccountDefaultSystemManagedSid             WELL_KNOWN_SID_TYPE = 110
	WinBuiltinDefaultSystemManagedGroupSid        WELL_KNOWN_SID_TYPE = 111
	WinBuiltinStorageReplicaAdminsSid             WELL_KNOWN_SID_TYPE = 112
	WinAccountKeyAdminsSid                        WELL_KNOWN_SID_TYPE = 113
	WinAccountEnterpriseKeyAdminsSid              WELL_KNOWN_SID_TYPE = 114
	WinAuthenticationKeyTrustSid                  WELL_KNOWN_SID_TYPE = 115
	WinAuthenticationKeyPropertyMFASid            WELL_KNOWN_SID_TYPE = 116
	WinAuthenticationKeyPropertyAttestationSid    WELL_KNOWN_SID_TYPE = 117
	WinAuthenticationFreshKeyAuthSid              WELL_KNOWN_SID_TYPE = 118
	WinBuiltinDeviceOwnersSid                     WELL_KNOWN_SID_TYPE = 119
)

// enum
type ACL_INFORMATION_CLASS int32

const (
	AclRevisionInformation ACL_INFORMATION_CLASS = 1
	AclSizeInformation     ACL_INFORMATION_CLASS = 2
)

// enum
type AUDIT_EVENT_TYPE int32

const (
	AuditEventObjectAccess           AUDIT_EVENT_TYPE = 0
	AuditEventDirectoryServiceAccess AUDIT_EVENT_TYPE = 1
)

// enum
type SECURITY_IMPERSONATION_LEVEL int32

const (
	SecurityAnonymous      SECURITY_IMPERSONATION_LEVEL = 0
	SecurityIdentification SECURITY_IMPERSONATION_LEVEL = 1
	SecurityImpersonation  SECURITY_IMPERSONATION_LEVEL = 2
	SecurityDelegation     SECURITY_IMPERSONATION_LEVEL = 3
)

// enum
type TOKEN_TYPE int32

const (
	TokenPrimary       TOKEN_TYPE = 1
	TokenImpersonation TOKEN_TYPE = 2
)

// enum
type TOKEN_ELEVATION_TYPE int32

const (
	TokenElevationTypeDefault TOKEN_ELEVATION_TYPE = 1
	TokenElevationTypeFull    TOKEN_ELEVATION_TYPE = 2
	TokenElevationTypeLimited TOKEN_ELEVATION_TYPE = 3
)

// enum
type TOKEN_INFORMATION_CLASS int32

const (
	TokenUser                            TOKEN_INFORMATION_CLASS = 1
	TokenGroups                          TOKEN_INFORMATION_CLASS = 2
	TokenPrivileges                      TOKEN_INFORMATION_CLASS = 3
	TokenOwner                           TOKEN_INFORMATION_CLASS = 4
	TokenPrimaryGroup                    TOKEN_INFORMATION_CLASS = 5
	TokenDefaultDacl                     TOKEN_INFORMATION_CLASS = 6
	TokenSource                          TOKEN_INFORMATION_CLASS = 7
	TokenType                            TOKEN_INFORMATION_CLASS = 8
	TokenImpersonationLevel              TOKEN_INFORMATION_CLASS = 9
	TokenStatistics                      TOKEN_INFORMATION_CLASS = 10
	TokenRestrictedSids                  TOKEN_INFORMATION_CLASS = 11
	TokenSessionId                       TOKEN_INFORMATION_CLASS = 12
	TokenGroupsAndPrivileges             TOKEN_INFORMATION_CLASS = 13
	TokenSessionReference                TOKEN_INFORMATION_CLASS = 14
	TokenSandBoxInert                    TOKEN_INFORMATION_CLASS = 15
	TokenAuditPolicy                     TOKEN_INFORMATION_CLASS = 16
	TokenOrigin                          TOKEN_INFORMATION_CLASS = 17
	TokenElevationType                   TOKEN_INFORMATION_CLASS = 18
	TokenLinkedToken                     TOKEN_INFORMATION_CLASS = 19
	TokenElevation                       TOKEN_INFORMATION_CLASS = 20
	TokenHasRestrictions                 TOKEN_INFORMATION_CLASS = 21
	TokenAccessInformation               TOKEN_INFORMATION_CLASS = 22
	TokenVirtualizationAllowed           TOKEN_INFORMATION_CLASS = 23
	TokenVirtualizationEnabled           TOKEN_INFORMATION_CLASS = 24
	TokenIntegrityLevel                  TOKEN_INFORMATION_CLASS = 25
	TokenUIAccess                        TOKEN_INFORMATION_CLASS = 26
	TokenMandatoryPolicy                 TOKEN_INFORMATION_CLASS = 27
	TokenLogonSid                        TOKEN_INFORMATION_CLASS = 28
	TokenIsAppContainer                  TOKEN_INFORMATION_CLASS = 29
	TokenCapabilities                    TOKEN_INFORMATION_CLASS = 30
	TokenAppContainerSid                 TOKEN_INFORMATION_CLASS = 31
	TokenAppContainerNumber              TOKEN_INFORMATION_CLASS = 32
	TokenUserClaimAttributes             TOKEN_INFORMATION_CLASS = 33
	TokenDeviceClaimAttributes           TOKEN_INFORMATION_CLASS = 34
	TokenRestrictedUserClaimAttributes   TOKEN_INFORMATION_CLASS = 35
	TokenRestrictedDeviceClaimAttributes TOKEN_INFORMATION_CLASS = 36
	TokenDeviceGroups                    TOKEN_INFORMATION_CLASS = 37
	TokenRestrictedDeviceGroups          TOKEN_INFORMATION_CLASS = 38
	TokenSecurityAttributes              TOKEN_INFORMATION_CLASS = 39
	TokenIsRestricted                    TOKEN_INFORMATION_CLASS = 40
	TokenProcessTrustLevel               TOKEN_INFORMATION_CLASS = 41
	TokenPrivateNameSpace                TOKEN_INFORMATION_CLASS = 42
	TokenSingletonAttributes             TOKEN_INFORMATION_CLASS = 43
	TokenBnoIsolation                    TOKEN_INFORMATION_CLASS = 44
	TokenChildProcessFlags               TOKEN_INFORMATION_CLASS = 45
	TokenIsLessPrivilegedAppContainer    TOKEN_INFORMATION_CLASS = 46
	TokenIsSandboxed                     TOKEN_INFORMATION_CLASS = 47
	MaxTokenInfoClass                    TOKEN_INFORMATION_CLASS = 48
)

// enum
type MANDATORY_LEVEL int32

const (
	MandatoryLevelUntrusted     MANDATORY_LEVEL = 0
	MandatoryLevelLow           MANDATORY_LEVEL = 1
	MandatoryLevelMedium        MANDATORY_LEVEL = 2
	MandatoryLevelHigh          MANDATORY_LEVEL = 3
	MandatoryLevelSystem        MANDATORY_LEVEL = 4
	MandatoryLevelSecureProcess MANDATORY_LEVEL = 5
	MandatoryLevelCount         MANDATORY_LEVEL = 6
)

// structs

type SECURITY_ATTRIBUTES struct {
	NLength              uint32
	LpSecurityDescriptor unsafe.Pointer
	BInheritHandle       BOOL
}

type LLFILETIME_Anonymous struct {
	Data [1]uint64
}

func (this *LLFILETIME_Anonymous) Ll() *int64 {
	return (*int64)(unsafe.Pointer(this))
}

func (this *LLFILETIME_Anonymous) LlVal() int64 {
	return *(*int64)(unsafe.Pointer(this))
}

func (this *LLFILETIME_Anonymous) Ft() *FILETIME {
	return (*FILETIME)(unsafe.Pointer(this))
}

func (this *LLFILETIME_Anonymous) FtVal() FILETIME {
	return *(*FILETIME)(unsafe.Pointer(this))
}

type LLFILETIME struct {
	LLFILETIME_Anonymous
}

type GENERIC_MAPPING struct {
	GenericRead    uint32
	GenericWrite   uint32
	GenericExecute uint32
	GenericAll     uint32
}

type LUID_AND_ATTRIBUTES struct {
	Luid       LUID
	Attributes TOKEN_PRIVILEGES_ATTRIBUTES
}

type SID_IDENTIFIER_AUTHORITY struct {
	Value [6]byte
}

type SID struct {
	Revision            byte
	SubAuthorityCount   byte
	IdentifierAuthority SID_IDENTIFIER_AUTHORITY
	SubAuthority        [1]uint32
}

type SE_SID struct {
	Data [17]uint32
}

func (this *SE_SID) Sid() *SID {
	return (*SID)(unsafe.Pointer(this))
}

func (this *SE_SID) SidVal() SID {
	return *(*SID)(unsafe.Pointer(this))
}

func (this *SE_SID) Buffer() *[68]byte {
	return (*[68]byte)(unsafe.Pointer(this))
}

func (this *SE_SID) BufferVal() [68]byte {
	return *(*[68]byte)(unsafe.Pointer(this))
}

type SID_AND_ATTRIBUTES struct {
	Sid        PSID
	Attributes uint32
}

type SID_AND_ATTRIBUTES_HASH struct {
	SidCount uint32
	SidAttr  *SID_AND_ATTRIBUTES
	Hash     [32]uintptr
}

type ACL struct {
	AclRevision byte
	Sbz1        byte
	AclSize     uint16
	AceCount    uint16
	Sbz2        uint16
}

type ACE_HEADER struct {
	AceType  byte
	AceFlags byte
	AceSize  uint16
}

type ACCESS_ALLOWED_ACE struct {
	Header   ACE_HEADER
	Mask     uint32
	SidStart uint32
}

type ACCESS_DENIED_ACE struct {
	Header   ACE_HEADER
	Mask     uint32
	SidStart uint32
}

type SYSTEM_AUDIT_ACE struct {
	Header   ACE_HEADER
	Mask     uint32
	SidStart uint32
}

type SYSTEM_ALARM_ACE struct {
	Header   ACE_HEADER
	Mask     uint32
	SidStart uint32
}

type SYSTEM_RESOURCE_ATTRIBUTE_ACE struct {
	Header   ACE_HEADER
	Mask     uint32
	SidStart uint32
}

type SYSTEM_SCOPED_POLICY_ID_ACE struct {
	Header   ACE_HEADER
	Mask     uint32
	SidStart uint32
}

type SYSTEM_MANDATORY_LABEL_ACE struct {
	Header   ACE_HEADER
	Mask     uint32
	SidStart uint32
}

type SYSTEM_PROCESS_TRUST_LABEL_ACE struct {
	Header   ACE_HEADER
	Mask     uint32
	SidStart uint32
}

type SYSTEM_ACCESS_FILTER_ACE struct {
	Header   ACE_HEADER
	Mask     uint32
	SidStart uint32
}

type ACCESS_ALLOWED_OBJECT_ACE struct {
	Header              ACE_HEADER
	Mask                uint32
	Flags               SYSTEM_AUDIT_OBJECT_ACE_FLAGS
	ObjectType          syscall.GUID
	InheritedObjectType syscall.GUID
	SidStart            uint32
}

type ACCESS_DENIED_OBJECT_ACE struct {
	Header              ACE_HEADER
	Mask                uint32
	Flags               SYSTEM_AUDIT_OBJECT_ACE_FLAGS
	ObjectType          syscall.GUID
	InheritedObjectType syscall.GUID
	SidStart            uint32
}

type SYSTEM_AUDIT_OBJECT_ACE struct {
	Header              ACE_HEADER
	Mask                uint32
	Flags               SYSTEM_AUDIT_OBJECT_ACE_FLAGS
	ObjectType          syscall.GUID
	InheritedObjectType syscall.GUID
	SidStart            uint32
}

type SYSTEM_ALARM_OBJECT_ACE struct {
	Header              ACE_HEADER
	Mask                uint32
	Flags               uint32
	ObjectType          syscall.GUID
	InheritedObjectType syscall.GUID
	SidStart            uint32
}

type ACCESS_ALLOWED_CALLBACK_ACE struct {
	Header   ACE_HEADER
	Mask     uint32
	SidStart uint32
}

type ACCESS_DENIED_CALLBACK_ACE struct {
	Header   ACE_HEADER
	Mask     uint32
	SidStart uint32
}

type SYSTEM_AUDIT_CALLBACK_ACE struct {
	Header   ACE_HEADER
	Mask     uint32
	SidStart uint32
}

type SYSTEM_ALARM_CALLBACK_ACE struct {
	Header   ACE_HEADER
	Mask     uint32
	SidStart uint32
}

type ACCESS_ALLOWED_CALLBACK_OBJECT_ACE struct {
	Header              ACE_HEADER
	Mask                uint32
	Flags               SYSTEM_AUDIT_OBJECT_ACE_FLAGS
	ObjectType          syscall.GUID
	InheritedObjectType syscall.GUID
	SidStart            uint32
}

type ACCESS_DENIED_CALLBACK_OBJECT_ACE struct {
	Header              ACE_HEADER
	Mask                uint32
	Flags               SYSTEM_AUDIT_OBJECT_ACE_FLAGS
	ObjectType          syscall.GUID
	InheritedObjectType syscall.GUID
	SidStart            uint32
}

type SYSTEM_AUDIT_CALLBACK_OBJECT_ACE struct {
	Header              ACE_HEADER
	Mask                uint32
	Flags               SYSTEM_AUDIT_OBJECT_ACE_FLAGS
	ObjectType          syscall.GUID
	InheritedObjectType syscall.GUID
	SidStart            uint32
}

type SYSTEM_ALARM_CALLBACK_OBJECT_ACE struct {
	Header              ACE_HEADER
	Mask                uint32
	Flags               SYSTEM_AUDIT_OBJECT_ACE_FLAGS
	ObjectType          syscall.GUID
	InheritedObjectType syscall.GUID
	SidStart            uint32
}

type ACL_REVISION_INFORMATION struct {
	AclRevision uint32
}

type ACL_SIZE_INFORMATION struct {
	AceCount      uint32
	AclBytesInUse uint32
	AclBytesFree  uint32
}

type SECURITY_DESCRIPTOR struct {
	Revision byte
	Sbz1     byte
	Control  uint16
	Owner    PSID
	Group    PSID
	Sacl     *ACL
	Dacl     *ACL
}

type OBJECT_TYPE_LIST struct {
	Level      uint16
	Sbz        uint16
	ObjectType *syscall.GUID
}

type PRIVILEGE_SET struct {
	PrivilegeCount uint32
	Control        uint32
	Privilege      [1]LUID_AND_ATTRIBUTES
}

type ACCESS_REASONS struct {
	Data [32]uint32
}

type SE_SECURITY_DESCRIPTOR struct {
	Size               uint32
	Flags              uint32
	SecurityDescriptor PSECURITY_DESCRIPTOR
}

type SE_ACCESS_REQUEST struct {
	Size                    uint32
	SeSecurityDescriptor    *SE_SECURITY_DESCRIPTOR
	DesiredAccess           uint32
	PreviouslyGrantedAccess uint32
	PrincipalSelfSid        PSID
	GenericMapping          *GENERIC_MAPPING
	ObjectTypeListCount     uint32
	ObjectTypeList          *OBJECT_TYPE_LIST
}

type SE_ACCESS_REPLY struct {
	Size            uint32
	ResultListCount uint32
	GrantedAccess   *uint32
	AccessStatus    *uint32
	AccessReason    *ACCESS_REASONS
	Privileges      **PRIVILEGE_SET
}

type TOKEN_USER struct {
	User SID_AND_ATTRIBUTES
}

type TOKEN_GROUPS struct {
	GroupCount uint32
	Groups     [1]SID_AND_ATTRIBUTES
}

type TOKEN_PRIVILEGES struct {
	PrivilegeCount uint32
	Privileges     [1]LUID_AND_ATTRIBUTES
}

type TOKEN_OWNER struct {
	Owner PSID
}

type TOKEN_PRIMARY_GROUP struct {
	PrimaryGroup PSID
}

type TOKEN_DEFAULT_DACL struct {
	DefaultDacl *ACL
}

type TOKEN_USER_CLAIMS struct {
	UserClaims unsafe.Pointer
}

type TOKEN_DEVICE_CLAIMS struct {
	DeviceClaims unsafe.Pointer
}

type TOKEN_GROUPS_AND_PRIVILEGES struct {
	SidCount            uint32
	SidLength           uint32
	Sids                *SID_AND_ATTRIBUTES
	RestrictedSidCount  uint32
	RestrictedSidLength uint32
	RestrictedSids      *SID_AND_ATTRIBUTES
	PrivilegeCount      uint32
	PrivilegeLength     uint32
	Privileges          *LUID_AND_ATTRIBUTES
	AuthenticationId    LUID
}

type TOKEN_LINKED_TOKEN struct {
	LinkedToken HANDLE
}

type TOKEN_ELEVATION struct {
	TokenIsElevated uint32
}

type TOKEN_MANDATORY_LABEL struct {
	Label SID_AND_ATTRIBUTES
}

type TOKEN_MANDATORY_POLICY struct {
	Policy TOKEN_MANDATORY_POLICY_ID
}

type TOKEN_ACCESS_INFORMATION struct {
	SidHash            *SID_AND_ATTRIBUTES_HASH
	RestrictedSidHash  *SID_AND_ATTRIBUTES_HASH
	Privileges         *TOKEN_PRIVILEGES
	AuthenticationId   LUID
	TokenType          TOKEN_TYPE
	ImpersonationLevel SECURITY_IMPERSONATION_LEVEL
	MandatoryPolicy    TOKEN_MANDATORY_POLICY
	Flags              uint32
	AppContainerNumber uint32
	PackageSid         PSID
	CapabilitiesHash   *SID_AND_ATTRIBUTES_HASH
	TrustLevelSid      PSID
	SecurityAttributes unsafe.Pointer
}

type TOKEN_AUDIT_POLICY struct {
	PerUserPolicy [30]byte
}

type TOKEN_SOURCE struct {
	SourceName       [8]CHAR
	SourceIdentifier LUID
}

type TOKEN_STATISTICS struct {
	TokenId            LUID
	AuthenticationId   LUID
	ExpirationTime     int64
	TokenType          TOKEN_TYPE
	ImpersonationLevel SECURITY_IMPERSONATION_LEVEL
	DynamicCharged     uint32
	DynamicAvailable   uint32
	GroupCount         uint32
	PrivilegeCount     uint32
	ModifiedId         LUID
}

type TOKEN_CONTROL struct {
	TokenId          LUID
	AuthenticationId LUID
	ModifiedId       LUID
	TokenSource      TOKEN_SOURCE
}

type TOKEN_ORIGIN struct {
	OriginatingLogonSession LUID
}

type TOKEN_APPCONTAINER_INFORMATION struct {
	TokenAppContainer PSID
}

type CLAIM_SECURITY_ATTRIBUTE_FQBN_VALUE struct {
	Version uint64
	Name    PWSTR
}

type CLAIM_SECURITY_ATTRIBUTE_OCTET_STRING_VALUE struct {
	PValue      unsafe.Pointer
	ValueLength uint32
}

type CLAIM_SECURITY_ATTRIBUTE_V1_Values struct {
	Data [1]uint64
}

func (this *CLAIM_SECURITY_ATTRIBUTE_V1_Values) PInt64() **int64 {
	return (**int64)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_V1_Values) PInt64Val() *int64 {
	return *(**int64)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_V1_Values) PUint64() **uint64 {
	return (**uint64)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_V1_Values) PUint64Val() *uint64 {
	return *(**uint64)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_V1_Values) PpString() **PWSTR {
	return (**PWSTR)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_V1_Values) PpStringVal() *PWSTR {
	return *(**PWSTR)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_V1_Values) PFqbn() **CLAIM_SECURITY_ATTRIBUTE_FQBN_VALUE {
	return (**CLAIM_SECURITY_ATTRIBUTE_FQBN_VALUE)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_V1_Values) PFqbnVal() *CLAIM_SECURITY_ATTRIBUTE_FQBN_VALUE {
	return *(**CLAIM_SECURITY_ATTRIBUTE_FQBN_VALUE)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_V1_Values) POctetString() **CLAIM_SECURITY_ATTRIBUTE_OCTET_STRING_VALUE {
	return (**CLAIM_SECURITY_ATTRIBUTE_OCTET_STRING_VALUE)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_V1_Values) POctetStringVal() *CLAIM_SECURITY_ATTRIBUTE_OCTET_STRING_VALUE {
	return *(**CLAIM_SECURITY_ATTRIBUTE_OCTET_STRING_VALUE)(unsafe.Pointer(this))
}

type CLAIM_SECURITY_ATTRIBUTE_V1 struct {
	Name       PWSTR
	ValueType  CLAIM_SECURITY_ATTRIBUTE_VALUE_TYPE
	Reserved   uint16
	Flags      uint32
	ValueCount uint32
	Values     CLAIM_SECURITY_ATTRIBUTE_V1_Values
}

type CLAIM_SECURITY_ATTRIBUTE_RELATIVE_V1_Values struct {
	Data [1]uint32
}

func (this *CLAIM_SECURITY_ATTRIBUTE_RELATIVE_V1_Values) PInt64() *[1]uint32 {
	return (*[1]uint32)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_RELATIVE_V1_Values) PInt64Val() [1]uint32 {
	return *(*[1]uint32)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_RELATIVE_V1_Values) PUint64() *[1]uint32 {
	return (*[1]uint32)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_RELATIVE_V1_Values) PUint64Val() [1]uint32 {
	return *(*[1]uint32)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_RELATIVE_V1_Values) PpString() *[1]uint32 {
	return (*[1]uint32)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_RELATIVE_V1_Values) PpStringVal() [1]uint32 {
	return *(*[1]uint32)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_RELATIVE_V1_Values) PFqbn() *[1]uint32 {
	return (*[1]uint32)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_RELATIVE_V1_Values) PFqbnVal() [1]uint32 {
	return *(*[1]uint32)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_RELATIVE_V1_Values) POctetString() *[1]uint32 {
	return (*[1]uint32)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTE_RELATIVE_V1_Values) POctetStringVal() [1]uint32 {
	return *(*[1]uint32)(unsafe.Pointer(this))
}

type CLAIM_SECURITY_ATTRIBUTE_RELATIVE_V1 struct {
	Name       uint32
	ValueType  CLAIM_SECURITY_ATTRIBUTE_VALUE_TYPE
	Reserved   uint16
	Flags      CLAIM_SECURITY_ATTRIBUTE_FLAGS
	ValueCount uint32
	Values     CLAIM_SECURITY_ATTRIBUTE_RELATIVE_V1_Values
}

type CLAIM_SECURITY_ATTRIBUTES_INFORMATION_Attribute struct {
	Data [1]uint64
}

func (this *CLAIM_SECURITY_ATTRIBUTES_INFORMATION_Attribute) PAttributeV1() **CLAIM_SECURITY_ATTRIBUTE_V1 {
	return (**CLAIM_SECURITY_ATTRIBUTE_V1)(unsafe.Pointer(this))
}

func (this *CLAIM_SECURITY_ATTRIBUTES_INFORMATION_Attribute) PAttributeV1Val() *CLAIM_SECURITY_ATTRIBUTE_V1 {
	return *(**CLAIM_SECURITY_ATTRIBUTE_V1)(unsafe.Pointer(this))
}

type CLAIM_SECURITY_ATTRIBUTES_INFORMATION struct {
	Version        uint16
	Reserved       uint16
	AttributeCount uint32
	Attribute      CLAIM_SECURITY_ATTRIBUTES_INFORMATION_Attribute
}

type SECURITY_QUALITY_OF_SERVICE struct {
	Length              uint32
	ImpersonationLevel  SECURITY_IMPERSONATION_LEVEL
	ContextTrackingMode byte
	EffectiveOnly       BOOLEAN
}

type SE_IMPERSONATION_STATE struct {
	Token         unsafe.Pointer
	CopyOnOpen    BOOLEAN
	EffectiveOnly BOOLEAN
	Level         SECURITY_IMPERSONATION_LEVEL
}

type SECURITY_CAPABILITIES struct {
	AppContainerSid PSID
	Capabilities    *SID_AND_ATTRIBUTES
	CapabilityCount uint32
	Reserved        uint32
}

type QUOTA_LIMITS struct {
	PagedPoolLimit        uintptr
	NonPagedPoolLimit     uintptr
	MinimumWorkingSetSize uintptr
	MaximumWorkingSetSize uintptr
	PagefileLimit         uintptr
	TimeLimit             int64
}

// func types

type PLSA_AP_CALL_PACKAGE_UNTRUSTED = uintptr
type PLSA_AP_CALL_PACKAGE_UNTRUSTED_func = func(ClientRequest unsafe.Pointer, ProtocolSubmitBuffer unsafe.Pointer, ClientBufferBase unsafe.Pointer, SubmitBufferLength uint32, ProtocolReturnBuffer unsafe.Pointer, ReturnBufferLength *uint32, ProtocolStatus *int32) NTSTATUS

type SEC_THREAD_START = uintptr
type SEC_THREAD_START_func = func(lpThreadParameter unsafe.Pointer) uint32

var (
	pAccessCheck                                        uintptr
	pAccessCheckAndAuditAlarmW                          uintptr
	pAccessCheckByType                                  uintptr
	pAccessCheckByTypeResultList                        uintptr
	pAccessCheckByTypeAndAuditAlarmW                    uintptr
	pAccessCheckByTypeResultListAndAuditAlarmW          uintptr
	pAccessCheckByTypeResultListAndAuditAlarmByHandleW  uintptr
	pAddAccessAllowedAce                                uintptr
	pAddAccessAllowedAceEx                              uintptr
	pAddAccessAllowedObjectAce                          uintptr
	pAddAccessDeniedAce                                 uintptr
	pAddAccessDeniedAceEx                               uintptr
	pAddAccessDeniedObjectAce                           uintptr
	pAddAce                                             uintptr
	pAddAuditAccessAce                                  uintptr
	pAddAuditAccessAceEx                                uintptr
	pAddAuditAccessObjectAce                            uintptr
	pAddMandatoryAce                                    uintptr
	pAddResourceAttributeAce                            uintptr
	pAddScopedPolicyIDAce                               uintptr
	pAdjustTokenGroups                                  uintptr
	pAdjustTokenPrivileges                              uintptr
	pAllocateAndInitializeSid                           uintptr
	pAllocateLocallyUniqueId                            uintptr
	pAreAllAccessesGranted                              uintptr
	pAreAnyAccessesGranted                              uintptr
	pCheckTokenMembership                               uintptr
	pCheckTokenCapability                               uintptr
	pGetAppContainerAce                                 uintptr
	pCheckTokenMembershipEx                             uintptr
	pConvertToAutoInheritPrivateObjectSecurity          uintptr
	pCopySid                                            uintptr
	pCreatePrivateObjectSecurity                        uintptr
	pCreatePrivateObjectSecurityEx                      uintptr
	pCreatePrivateObjectSecurityWithMultipleInheritance uintptr
	pCreateRestrictedToken                              uintptr
	pCreateWellKnownSid                                 uintptr
	pEqualDomainSid                                     uintptr
	pDeleteAce                                          uintptr
	pDestroyPrivateObjectSecurity                       uintptr
	pDuplicateToken                                     uintptr
	pDuplicateTokenEx                                   uintptr
	pEqualPrefixSid                                     uintptr
	pEqualSid                                           uintptr
	pFindFirstFreeAce                                   uintptr
	pFreeSid                                            uintptr
	pGetAce                                             uintptr
	pGetAclInformation                                  uintptr
	pGetFileSecurityW                                   uintptr
	pGetKernelObjectSecurity                            uintptr
	pGetLengthSid                                       uintptr
	pGetPrivateObjectSecurity                           uintptr
	pGetSecurityDescriptorControl                       uintptr
	pGetSecurityDescriptorDacl                          uintptr
	pGetSecurityDescriptorGroup                         uintptr
	pGetSecurityDescriptorLength                        uintptr
	pGetSecurityDescriptorOwner                         uintptr
	pGetSecurityDescriptorRMControl                     uintptr
	pGetSecurityDescriptorSacl                          uintptr
	pGetSidIdentifierAuthority                          uintptr
	pGetSidLengthRequired                               uintptr
	pGetSidSubAuthority                                 uintptr
	pGetSidSubAuthorityCount                            uintptr
	pGetTokenInformation                                uintptr
	pGetWindowsAccountDomainSid                         uintptr
	pImpersonateAnonymousToken                          uintptr
	pImpersonateLoggedOnUser                            uintptr
	pImpersonateSelf                                    uintptr
	pInitializeAcl                                      uintptr
	pInitializeSecurityDescriptor                       uintptr
	pInitializeSid                                      uintptr
	pIsTokenRestricted                                  uintptr
	pIsValidAcl                                         uintptr
	pIsValidSecurityDescriptor                          uintptr
	pIsValidSid                                         uintptr
	pIsWellKnownSid                                     uintptr
	pMakeAbsoluteSD                                     uintptr
	pMakeSelfRelativeSD                                 uintptr
	pMapGenericMask                                     uintptr
	pObjectCloseAuditAlarmW                             uintptr
	pObjectDeleteAuditAlarmW                            uintptr
	pObjectOpenAuditAlarmW                              uintptr
	pObjectPrivilegeAuditAlarmW                         uintptr
	pPrivilegeCheck                                     uintptr
	pPrivilegedServiceAuditAlarmW                       uintptr
	pQuerySecurityAccessMask                            uintptr
	pRevertToSelf                                       uintptr
	pSetAclInformation                                  uintptr
	pSetFileSecurityW                                   uintptr
	pSetKernelObjectSecurity                            uintptr
	pSetPrivateObjectSecurity                           uintptr
	pSetPrivateObjectSecurityEx                         uintptr
	pSetSecurityAccessMask                              uintptr
	pSetSecurityDescriptorControl                       uintptr
	pSetSecurityDescriptorDacl                          uintptr
	pSetSecurityDescriptorGroup                         uintptr
	pSetSecurityDescriptorOwner                         uintptr
	pSetSecurityDescriptorRMControl                     uintptr
	pSetSecurityDescriptorSacl                          uintptr
	pSetTokenInformation                                uintptr
	pSetCachedSigningLevel                              uintptr
	pGetCachedSigningLevel                              uintptr
	pSetUserObjectSecurity                              uintptr
	pGetUserObjectSecurity                              uintptr
	pAccessCheckAndAuditAlarmA                          uintptr
	pAccessCheckByTypeAndAuditAlarmA                    uintptr
	pAccessCheckByTypeResultListAndAuditAlarmA          uintptr
	pAccessCheckByTypeResultListAndAuditAlarmByHandleA  uintptr
	pObjectOpenAuditAlarmA                              uintptr
	pObjectPrivilegeAuditAlarmA                         uintptr
	pObjectCloseAuditAlarmA                             uintptr
	pObjectDeleteAuditAlarmA                            uintptr
	pPrivilegedServiceAuditAlarmA                       uintptr
	pAddConditionalAce                                  uintptr
	pSetFileSecurityA                                   uintptr
	pGetFileSecurityA                                   uintptr
	pLookupAccountSidA                                  uintptr
	pLookupAccountSidW                                  uintptr
	pLookupAccountNameA                                 uintptr
	pLookupAccountNameW                                 uintptr
	pLookupPrivilegeValueA                              uintptr
	pLookupPrivilegeValueW                              uintptr
	pLookupPrivilegeNameA                               uintptr
	pLookupPrivilegeNameW                               uintptr
	pLookupPrivilegeDisplayNameA                        uintptr
	pLookupPrivilegeDisplayNameW                        uintptr
	pLogonUserA                                         uintptr
	pLogonUserW                                         uintptr
	pLogonUserExA                                       uintptr
	pLogonUserExW                                       uintptr
)

func AccessCheck(pSecurityDescriptor PSECURITY_DESCRIPTOR, ClientToken HANDLE, DesiredAccess uint32, GenericMapping *GENERIC_MAPPING, PrivilegeSet *PRIVILEGE_SET, PrivilegeSetLength *uint32, GrantedAccess *uint32, AccessStatus *int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAccessCheck, libAdvapi32, "AccessCheck")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)), ClientToken, uintptr(DesiredAccess), uintptr(unsafe.Pointer(GenericMapping)), uintptr(unsafe.Pointer(PrivilegeSet)), uintptr(unsafe.Pointer(PrivilegeSetLength)), uintptr(unsafe.Pointer(GrantedAccess)), uintptr(unsafe.Pointer(AccessStatus)))
	return BOOL(ret), WIN32_ERROR(err)
}

var AccessCheckAndAuditAlarm = AccessCheckAndAuditAlarmW

func AccessCheckAndAuditAlarmW(SubsystemName PWSTR, HandleId unsafe.Pointer, ObjectTypeName PWSTR, ObjectName PWSTR, SecurityDescriptor PSECURITY_DESCRIPTOR, DesiredAccess uint32, GenericMapping *GENERIC_MAPPING, ObjectCreation BOOL, GrantedAccess *uint32, AccessStatus *int32, pfGenerateOnClose *int32) BOOL {
	addr := lazyAddr(&pAccessCheckAndAuditAlarmW, libAdvapi32, "AccessCheckAndAuditAlarmW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), uintptr(unsafe.Pointer(ObjectTypeName)), uintptr(unsafe.Pointer(ObjectName)), uintptr(unsafe.Pointer(SecurityDescriptor)), uintptr(DesiredAccess), uintptr(unsafe.Pointer(GenericMapping)), uintptr(ObjectCreation), uintptr(unsafe.Pointer(GrantedAccess)), uintptr(unsafe.Pointer(AccessStatus)), uintptr(unsafe.Pointer(pfGenerateOnClose)))
	return BOOL(ret)
}

func AccessCheckByType(pSecurityDescriptor PSECURITY_DESCRIPTOR, PrincipalSelfSid PSID, ClientToken HANDLE, DesiredAccess uint32, ObjectTypeList *OBJECT_TYPE_LIST, ObjectTypeListLength uint32, GenericMapping *GENERIC_MAPPING, PrivilegeSet *PRIVILEGE_SET, PrivilegeSetLength *uint32, GrantedAccess *uint32, AccessStatus *int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAccessCheckByType, libAdvapi32, "AccessCheckByType")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(unsafe.Pointer(PrincipalSelfSid)), ClientToken, uintptr(DesiredAccess), uintptr(unsafe.Pointer(ObjectTypeList)), uintptr(ObjectTypeListLength), uintptr(unsafe.Pointer(GenericMapping)), uintptr(unsafe.Pointer(PrivilegeSet)), uintptr(unsafe.Pointer(PrivilegeSetLength)), uintptr(unsafe.Pointer(GrantedAccess)), uintptr(unsafe.Pointer(AccessStatus)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AccessCheckByTypeResultList(pSecurityDescriptor PSECURITY_DESCRIPTOR, PrincipalSelfSid PSID, ClientToken HANDLE, DesiredAccess uint32, ObjectTypeList *OBJECT_TYPE_LIST, ObjectTypeListLength uint32, GenericMapping *GENERIC_MAPPING, PrivilegeSet *PRIVILEGE_SET, PrivilegeSetLength *uint32, GrantedAccessList *uint32, AccessStatusList *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAccessCheckByTypeResultList, libAdvapi32, "AccessCheckByTypeResultList")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(unsafe.Pointer(PrincipalSelfSid)), ClientToken, uintptr(DesiredAccess), uintptr(unsafe.Pointer(ObjectTypeList)), uintptr(ObjectTypeListLength), uintptr(unsafe.Pointer(GenericMapping)), uintptr(unsafe.Pointer(PrivilegeSet)), uintptr(unsafe.Pointer(PrivilegeSetLength)), uintptr(unsafe.Pointer(GrantedAccessList)), uintptr(unsafe.Pointer(AccessStatusList)))
	return BOOL(ret), WIN32_ERROR(err)
}

var AccessCheckByTypeAndAuditAlarm = AccessCheckByTypeAndAuditAlarmW

func AccessCheckByTypeAndAuditAlarmW(SubsystemName PWSTR, HandleId unsafe.Pointer, ObjectTypeName PWSTR, ObjectName PWSTR, SecurityDescriptor PSECURITY_DESCRIPTOR, PrincipalSelfSid PSID, DesiredAccess uint32, AuditType AUDIT_EVENT_TYPE, Flags uint32, ObjectTypeList *OBJECT_TYPE_LIST, ObjectTypeListLength uint32, GenericMapping *GENERIC_MAPPING, ObjectCreation BOOL, GrantedAccess *uint32, AccessStatus *int32, pfGenerateOnClose *int32) BOOL {
	addr := lazyAddr(&pAccessCheckByTypeAndAuditAlarmW, libAdvapi32, "AccessCheckByTypeAndAuditAlarmW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), uintptr(unsafe.Pointer(ObjectTypeName)), uintptr(unsafe.Pointer(ObjectName)), uintptr(unsafe.Pointer(SecurityDescriptor)), uintptr(unsafe.Pointer(PrincipalSelfSid)), uintptr(DesiredAccess), uintptr(AuditType), uintptr(Flags), uintptr(unsafe.Pointer(ObjectTypeList)), uintptr(ObjectTypeListLength), uintptr(unsafe.Pointer(GenericMapping)), uintptr(ObjectCreation), uintptr(unsafe.Pointer(GrantedAccess)), uintptr(unsafe.Pointer(AccessStatus)), uintptr(unsafe.Pointer(pfGenerateOnClose)))
	return BOOL(ret)
}

var AccessCheckByTypeResultListAndAuditAlarm = AccessCheckByTypeResultListAndAuditAlarmW

func AccessCheckByTypeResultListAndAuditAlarmW(SubsystemName PWSTR, HandleId unsafe.Pointer, ObjectTypeName PWSTR, ObjectName PWSTR, SecurityDescriptor PSECURITY_DESCRIPTOR, PrincipalSelfSid PSID, DesiredAccess uint32, AuditType AUDIT_EVENT_TYPE, Flags uint32, ObjectTypeList *OBJECT_TYPE_LIST, ObjectTypeListLength uint32, GenericMapping *GENERIC_MAPPING, ObjectCreation BOOL, GrantedAccessList *uint32, AccessStatusList *uint32, pfGenerateOnClose *int32) BOOL {
	addr := lazyAddr(&pAccessCheckByTypeResultListAndAuditAlarmW, libAdvapi32, "AccessCheckByTypeResultListAndAuditAlarmW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), uintptr(unsafe.Pointer(ObjectTypeName)), uintptr(unsafe.Pointer(ObjectName)), uintptr(unsafe.Pointer(SecurityDescriptor)), uintptr(unsafe.Pointer(PrincipalSelfSid)), uintptr(DesiredAccess), uintptr(AuditType), uintptr(Flags), uintptr(unsafe.Pointer(ObjectTypeList)), uintptr(ObjectTypeListLength), uintptr(unsafe.Pointer(GenericMapping)), uintptr(ObjectCreation), uintptr(unsafe.Pointer(GrantedAccessList)), uintptr(unsafe.Pointer(AccessStatusList)), uintptr(unsafe.Pointer(pfGenerateOnClose)))
	return BOOL(ret)
}

var AccessCheckByTypeResultListAndAuditAlarmByHandle = AccessCheckByTypeResultListAndAuditAlarmByHandleW

func AccessCheckByTypeResultListAndAuditAlarmByHandleW(SubsystemName PWSTR, HandleId unsafe.Pointer, ClientToken HANDLE, ObjectTypeName PWSTR, ObjectName PWSTR, SecurityDescriptor PSECURITY_DESCRIPTOR, PrincipalSelfSid PSID, DesiredAccess uint32, AuditType AUDIT_EVENT_TYPE, Flags uint32, ObjectTypeList *OBJECT_TYPE_LIST, ObjectTypeListLength uint32, GenericMapping *GENERIC_MAPPING, ObjectCreation BOOL, GrantedAccessList *uint32, AccessStatusList *uint32, pfGenerateOnClose *int32) BOOL {
	addr := lazyAddr(&pAccessCheckByTypeResultListAndAuditAlarmByHandleW, libAdvapi32, "AccessCheckByTypeResultListAndAuditAlarmByHandleW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), ClientToken, uintptr(unsafe.Pointer(ObjectTypeName)), uintptr(unsafe.Pointer(ObjectName)), uintptr(unsafe.Pointer(SecurityDescriptor)), uintptr(unsafe.Pointer(PrincipalSelfSid)), uintptr(DesiredAccess), uintptr(AuditType), uintptr(Flags), uintptr(unsafe.Pointer(ObjectTypeList)), uintptr(ObjectTypeListLength), uintptr(unsafe.Pointer(GenericMapping)), uintptr(ObjectCreation), uintptr(unsafe.Pointer(GrantedAccessList)), uintptr(unsafe.Pointer(AccessStatusList)), uintptr(unsafe.Pointer(pfGenerateOnClose)))
	return BOOL(ret)
}

func AddAccessAllowedAce(pAcl *ACL, dwAceRevision uint32, AccessMask uint32, pSid PSID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddAccessAllowedAce, libAdvapi32, "AddAccessAllowedAce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceRevision), uintptr(AccessMask), uintptr(unsafe.Pointer(pSid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AddAccessAllowedAceEx(pAcl *ACL, dwAceRevision uint32, AceFlags ACE_FLAGS, AccessMask uint32, pSid PSID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddAccessAllowedAceEx, libAdvapi32, "AddAccessAllowedAceEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceRevision), uintptr(AceFlags), uintptr(AccessMask), uintptr(unsafe.Pointer(pSid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AddAccessAllowedObjectAce(pAcl *ACL, dwAceRevision uint32, AceFlags ACE_FLAGS, AccessMask uint32, ObjectTypeGuid *syscall.GUID, InheritedObjectTypeGuid *syscall.GUID, pSid PSID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddAccessAllowedObjectAce, libAdvapi32, "AddAccessAllowedObjectAce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceRevision), uintptr(AceFlags), uintptr(AccessMask), uintptr(unsafe.Pointer(ObjectTypeGuid)), uintptr(unsafe.Pointer(InheritedObjectTypeGuid)), uintptr(unsafe.Pointer(pSid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AddAccessDeniedAce(pAcl *ACL, dwAceRevision uint32, AccessMask uint32, pSid PSID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddAccessDeniedAce, libAdvapi32, "AddAccessDeniedAce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceRevision), uintptr(AccessMask), uintptr(unsafe.Pointer(pSid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AddAccessDeniedAceEx(pAcl *ACL, dwAceRevision uint32, AceFlags ACE_FLAGS, AccessMask uint32, pSid PSID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddAccessDeniedAceEx, libAdvapi32, "AddAccessDeniedAceEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceRevision), uintptr(AceFlags), uintptr(AccessMask), uintptr(unsafe.Pointer(pSid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AddAccessDeniedObjectAce(pAcl *ACL, dwAceRevision uint32, AceFlags ACE_FLAGS, AccessMask uint32, ObjectTypeGuid *syscall.GUID, InheritedObjectTypeGuid *syscall.GUID, pSid PSID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddAccessDeniedObjectAce, libAdvapi32, "AddAccessDeniedObjectAce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceRevision), uintptr(AceFlags), uintptr(AccessMask), uintptr(unsafe.Pointer(ObjectTypeGuid)), uintptr(unsafe.Pointer(InheritedObjectTypeGuid)), uintptr(unsafe.Pointer(pSid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AddAce(pAcl *ACL, dwAceRevision uint32, dwStartingAceIndex uint32, pAceList unsafe.Pointer, nAceListLength uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddAce, libAdvapi32, "AddAce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceRevision), uintptr(dwStartingAceIndex), uintptr(pAceList), uintptr(nAceListLength))
	return BOOL(ret), WIN32_ERROR(err)
}

func AddAuditAccessAce(pAcl *ACL, dwAceRevision uint32, dwAccessMask uint32, pSid PSID, bAuditSuccess BOOL, bAuditFailure BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddAuditAccessAce, libAdvapi32, "AddAuditAccessAce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceRevision), uintptr(dwAccessMask), uintptr(unsafe.Pointer(pSid)), uintptr(bAuditSuccess), uintptr(bAuditFailure))
	return BOOL(ret), WIN32_ERROR(err)
}

func AddAuditAccessAceEx(pAcl *ACL, dwAceRevision uint32, AceFlags ACE_FLAGS, dwAccessMask uint32, pSid PSID, bAuditSuccess BOOL, bAuditFailure BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddAuditAccessAceEx, libAdvapi32, "AddAuditAccessAceEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceRevision), uintptr(AceFlags), uintptr(dwAccessMask), uintptr(unsafe.Pointer(pSid)), uintptr(bAuditSuccess), uintptr(bAuditFailure))
	return BOOL(ret), WIN32_ERROR(err)
}

func AddAuditAccessObjectAce(pAcl *ACL, dwAceRevision uint32, AceFlags ACE_FLAGS, AccessMask uint32, ObjectTypeGuid *syscall.GUID, InheritedObjectTypeGuid *syscall.GUID, pSid PSID, bAuditSuccess BOOL, bAuditFailure BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddAuditAccessObjectAce, libAdvapi32, "AddAuditAccessObjectAce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceRevision), uintptr(AceFlags), uintptr(AccessMask), uintptr(unsafe.Pointer(ObjectTypeGuid)), uintptr(unsafe.Pointer(InheritedObjectTypeGuid)), uintptr(unsafe.Pointer(pSid)), uintptr(bAuditSuccess), uintptr(bAuditFailure))
	return BOOL(ret), WIN32_ERROR(err)
}

func AddMandatoryAce(pAcl *ACL, dwAceRevision ACE_REVISION, AceFlags ACE_FLAGS, MandatoryPolicy uint32, pLabelSid PSID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddMandatoryAce, libAdvapi32, "AddMandatoryAce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceRevision), uintptr(AceFlags), uintptr(MandatoryPolicy), uintptr(unsafe.Pointer(pLabelSid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AddResourceAttributeAce(pAcl *ACL, dwAceRevision uint32, AceFlags ACE_FLAGS, AccessMask uint32, pSid PSID, pAttributeInfo *CLAIM_SECURITY_ATTRIBUTES_INFORMATION, pReturnLength *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddResourceAttributeAce, libKernel32, "AddResourceAttributeAce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceRevision), uintptr(AceFlags), uintptr(AccessMask), uintptr(unsafe.Pointer(pSid)), uintptr(unsafe.Pointer(pAttributeInfo)), uintptr(unsafe.Pointer(pReturnLength)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AddScopedPolicyIDAce(pAcl *ACL, dwAceRevision uint32, AceFlags ACE_FLAGS, AccessMask uint32, pSid PSID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddScopedPolicyIDAce, libKernel32, "AddScopedPolicyIDAce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceRevision), uintptr(AceFlags), uintptr(AccessMask), uintptr(unsafe.Pointer(pSid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AdjustTokenGroups(TokenHandle HANDLE, ResetToDefault BOOL, NewState *TOKEN_GROUPS, BufferLength uint32, PreviousState *TOKEN_GROUPS, ReturnLength *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAdjustTokenGroups, libAdvapi32, "AdjustTokenGroups")
	ret, _, err := syscall.SyscallN(addr, TokenHandle, uintptr(ResetToDefault), uintptr(unsafe.Pointer(NewState)), uintptr(BufferLength), uintptr(unsafe.Pointer(PreviousState)), uintptr(unsafe.Pointer(ReturnLength)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AdjustTokenPrivileges(TokenHandle HANDLE, DisableAllPrivileges BOOL, NewState *TOKEN_PRIVILEGES, BufferLength uint32, PreviousState *TOKEN_PRIVILEGES, ReturnLength *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAdjustTokenPrivileges, libAdvapi32, "AdjustTokenPrivileges")
	ret, _, err := syscall.SyscallN(addr, TokenHandle, uintptr(DisableAllPrivileges), uintptr(unsafe.Pointer(NewState)), uintptr(BufferLength), uintptr(unsafe.Pointer(PreviousState)), uintptr(unsafe.Pointer(ReturnLength)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AllocateAndInitializeSid(pIdentifierAuthority *SID_IDENTIFIER_AUTHORITY, nSubAuthorityCount byte, nSubAuthority0 uint32, nSubAuthority1 uint32, nSubAuthority2 uint32, nSubAuthority3 uint32, nSubAuthority4 uint32, nSubAuthority5 uint32, nSubAuthority6 uint32, nSubAuthority7 uint32, pSid *PSID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAllocateAndInitializeSid, libAdvapi32, "AllocateAndInitializeSid")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pIdentifierAuthority)), uintptr(nSubAuthorityCount), uintptr(nSubAuthority0), uintptr(nSubAuthority1), uintptr(nSubAuthority2), uintptr(nSubAuthority3), uintptr(nSubAuthority4), uintptr(nSubAuthority5), uintptr(nSubAuthority6), uintptr(nSubAuthority7), uintptr(unsafe.Pointer(pSid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AllocateLocallyUniqueId(Luid *LUID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAllocateLocallyUniqueId, libAdvapi32, "AllocateLocallyUniqueId")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Luid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AreAllAccessesGranted(GrantedAccess uint32, DesiredAccess uint32) BOOL {
	addr := lazyAddr(&pAreAllAccessesGranted, libAdvapi32, "AreAllAccessesGranted")
	ret, _, _ := syscall.SyscallN(addr, uintptr(GrantedAccess), uintptr(DesiredAccess))
	return BOOL(ret)
}

func AreAnyAccessesGranted(GrantedAccess uint32, DesiredAccess uint32) BOOL {
	addr := lazyAddr(&pAreAnyAccessesGranted, libAdvapi32, "AreAnyAccessesGranted")
	ret, _, _ := syscall.SyscallN(addr, uintptr(GrantedAccess), uintptr(DesiredAccess))
	return BOOL(ret)
}

func CheckTokenMembership(TokenHandle HANDLE, SidToCheck PSID, IsMember *BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCheckTokenMembership, libAdvapi32, "CheckTokenMembership")
	ret, _, err := syscall.SyscallN(addr, TokenHandle, uintptr(unsafe.Pointer(SidToCheck)), uintptr(unsafe.Pointer(IsMember)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CheckTokenCapability(TokenHandle HANDLE, CapabilitySidToCheck PSID, HasCapability *BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCheckTokenCapability, libKernel32, "CheckTokenCapability")
	ret, _, err := syscall.SyscallN(addr, TokenHandle, uintptr(unsafe.Pointer(CapabilitySidToCheck)), uintptr(unsafe.Pointer(HasCapability)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetAppContainerAce(Acl *ACL, StartingAceIndex uint32, AppContainerAce unsafe.Pointer, AppContainerAceIndex *uint32) BOOL {
	addr := lazyAddr(&pGetAppContainerAce, libKernel32, "GetAppContainerAce")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Acl)), uintptr(StartingAceIndex), uintptr(AppContainerAce), uintptr(unsafe.Pointer(AppContainerAceIndex)))
	return BOOL(ret)
}

func CheckTokenMembershipEx(TokenHandle HANDLE, SidToCheck PSID, Flags uint32, IsMember *BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCheckTokenMembershipEx, libKernel32, "CheckTokenMembershipEx")
	ret, _, err := syscall.SyscallN(addr, TokenHandle, uintptr(unsafe.Pointer(SidToCheck)), uintptr(Flags), uintptr(unsafe.Pointer(IsMember)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ConvertToAutoInheritPrivateObjectSecurity(ParentDescriptor PSECURITY_DESCRIPTOR, CurrentSecurityDescriptor PSECURITY_DESCRIPTOR, NewSecurityDescriptor *PSECURITY_DESCRIPTOR, ObjectType *syscall.GUID, IsDirectoryObject BOOLEAN, GenericMapping *GENERIC_MAPPING) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pConvertToAutoInheritPrivateObjectSecurity, libAdvapi32, "ConvertToAutoInheritPrivateObjectSecurity")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ParentDescriptor)), uintptr(unsafe.Pointer(CurrentSecurityDescriptor)), uintptr(unsafe.Pointer(NewSecurityDescriptor)), uintptr(unsafe.Pointer(ObjectType)), uintptr(IsDirectoryObject), uintptr(unsafe.Pointer(GenericMapping)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CopySid(nDestinationSidLength uint32, pDestinationSid PSID, pSourceSid PSID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCopySid, libAdvapi32, "CopySid")
	ret, _, err := syscall.SyscallN(addr, uintptr(nDestinationSidLength), uintptr(unsafe.Pointer(pDestinationSid)), uintptr(unsafe.Pointer(pSourceSid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreatePrivateObjectSecurity(ParentDescriptor PSECURITY_DESCRIPTOR, CreatorDescriptor PSECURITY_DESCRIPTOR, NewDescriptor *PSECURITY_DESCRIPTOR, IsDirectoryObject BOOL, Token HANDLE, GenericMapping *GENERIC_MAPPING) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreatePrivateObjectSecurity, libAdvapi32, "CreatePrivateObjectSecurity")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ParentDescriptor)), uintptr(unsafe.Pointer(CreatorDescriptor)), uintptr(unsafe.Pointer(NewDescriptor)), uintptr(IsDirectoryObject), Token, uintptr(unsafe.Pointer(GenericMapping)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreatePrivateObjectSecurityEx(ParentDescriptor PSECURITY_DESCRIPTOR, CreatorDescriptor PSECURITY_DESCRIPTOR, NewDescriptor *PSECURITY_DESCRIPTOR, ObjectType *syscall.GUID, IsContainerObject BOOL, AutoInheritFlags SECURITY_AUTO_INHERIT_FLAGS, Token HANDLE, GenericMapping *GENERIC_MAPPING) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreatePrivateObjectSecurityEx, libAdvapi32, "CreatePrivateObjectSecurityEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ParentDescriptor)), uintptr(unsafe.Pointer(CreatorDescriptor)), uintptr(unsafe.Pointer(NewDescriptor)), uintptr(unsafe.Pointer(ObjectType)), uintptr(IsContainerObject), uintptr(AutoInheritFlags), Token, uintptr(unsafe.Pointer(GenericMapping)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreatePrivateObjectSecurityWithMultipleInheritance(ParentDescriptor PSECURITY_DESCRIPTOR, CreatorDescriptor PSECURITY_DESCRIPTOR, NewDescriptor *PSECURITY_DESCRIPTOR, ObjectTypes **syscall.GUID, GuidCount uint32, IsContainerObject BOOL, AutoInheritFlags SECURITY_AUTO_INHERIT_FLAGS, Token HANDLE, GenericMapping *GENERIC_MAPPING) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreatePrivateObjectSecurityWithMultipleInheritance, libAdvapi32, "CreatePrivateObjectSecurityWithMultipleInheritance")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ParentDescriptor)), uintptr(unsafe.Pointer(CreatorDescriptor)), uintptr(unsafe.Pointer(NewDescriptor)), uintptr(unsafe.Pointer(ObjectTypes)), uintptr(GuidCount), uintptr(IsContainerObject), uintptr(AutoInheritFlags), Token, uintptr(unsafe.Pointer(GenericMapping)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateRestrictedToken(ExistingTokenHandle HANDLE, Flags CREATE_RESTRICTED_TOKEN_FLAGS, DisableSidCount uint32, SidsToDisable *SID_AND_ATTRIBUTES, DeletePrivilegeCount uint32, PrivilegesToDelete *LUID_AND_ATTRIBUTES, RestrictedSidCount uint32, SidsToRestrict *SID_AND_ATTRIBUTES, NewTokenHandle *HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateRestrictedToken, libAdvapi32, "CreateRestrictedToken")
	ret, _, err := syscall.SyscallN(addr, ExistingTokenHandle, uintptr(Flags), uintptr(DisableSidCount), uintptr(unsafe.Pointer(SidsToDisable)), uintptr(DeletePrivilegeCount), uintptr(unsafe.Pointer(PrivilegesToDelete)), uintptr(RestrictedSidCount), uintptr(unsafe.Pointer(SidsToRestrict)), uintptr(unsafe.Pointer(NewTokenHandle)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateWellKnownSid(WellKnownSidType WELL_KNOWN_SID_TYPE, DomainSid PSID, pSid PSID, cbSid *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateWellKnownSid, libAdvapi32, "CreateWellKnownSid")
	ret, _, err := syscall.SyscallN(addr, uintptr(WellKnownSidType), uintptr(unsafe.Pointer(DomainSid)), uintptr(unsafe.Pointer(pSid)), uintptr(unsafe.Pointer(cbSid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func EqualDomainSid(pSid1 PSID, pSid2 PSID, pfEqual *BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEqualDomainSid, libAdvapi32, "EqualDomainSid")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSid1)), uintptr(unsafe.Pointer(pSid2)), uintptr(unsafe.Pointer(pfEqual)))
	return BOOL(ret), WIN32_ERROR(err)
}

func DeleteAce(pAcl *ACL, dwAceIndex uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDeleteAce, libAdvapi32, "DeleteAce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceIndex))
	return BOOL(ret), WIN32_ERROR(err)
}

func DestroyPrivateObjectSecurity(ObjectDescriptor *PSECURITY_DESCRIPTOR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDestroyPrivateObjectSecurity, libAdvapi32, "DestroyPrivateObjectSecurity")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ObjectDescriptor)))
	return BOOL(ret), WIN32_ERROR(err)
}

func DuplicateToken(ExistingTokenHandle HANDLE, ImpersonationLevel SECURITY_IMPERSONATION_LEVEL, DuplicateTokenHandle *HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDuplicateToken, libAdvapi32, "DuplicateToken")
	ret, _, err := syscall.SyscallN(addr, ExistingTokenHandle, uintptr(ImpersonationLevel), uintptr(unsafe.Pointer(DuplicateTokenHandle)))
	return BOOL(ret), WIN32_ERROR(err)
}

func DuplicateTokenEx(hExistingToken HANDLE, dwDesiredAccess TOKEN_ACCESS_MASK, lpTokenAttributes *SECURITY_ATTRIBUTES, ImpersonationLevel SECURITY_IMPERSONATION_LEVEL, TokenType TOKEN_TYPE, phNewToken *HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDuplicateTokenEx, libAdvapi32, "DuplicateTokenEx")
	ret, _, err := syscall.SyscallN(addr, hExistingToken, uintptr(dwDesiredAccess), uintptr(unsafe.Pointer(lpTokenAttributes)), uintptr(ImpersonationLevel), uintptr(TokenType), uintptr(unsafe.Pointer(phNewToken)))
	return BOOL(ret), WIN32_ERROR(err)
}

func EqualPrefixSid(pSid1 PSID, pSid2 PSID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEqualPrefixSid, libAdvapi32, "EqualPrefixSid")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSid1)), uintptr(unsafe.Pointer(pSid2)))
	return BOOL(ret), WIN32_ERROR(err)
}

func EqualSid(pSid1 PSID, pSid2 PSID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEqualSid, libAdvapi32, "EqualSid")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSid1)), uintptr(unsafe.Pointer(pSid2)))
	return BOOL(ret), WIN32_ERROR(err)
}

func FindFirstFreeAce(pAcl *ACL, pAce unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pFindFirstFreeAce, libAdvapi32, "FindFirstFreeAce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(pAce))
	return BOOL(ret), WIN32_ERROR(err)
}

func FreeSid(pSid PSID) unsafe.Pointer {
	addr := lazyAddr(&pFreeSid, libAdvapi32, "FreeSid")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSid)))
	return (unsafe.Pointer)(ret)
}

func GetAce(pAcl *ACL, dwAceIndex uint32, pAce unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetAce, libAdvapi32, "GetAce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceIndex), uintptr(pAce))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetAclInformation(pAcl *ACL, pAclInformation unsafe.Pointer, nAclInformationLength uint32, dwAclInformationClass ACL_INFORMATION_CLASS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetAclInformation, libAdvapi32, "GetAclInformation")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(pAclInformation), uintptr(nAclInformationLength), uintptr(dwAclInformationClass))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetFileSecurity = GetFileSecurityW

func GetFileSecurityW(lpFileName PWSTR, RequestedInformation uint32, pSecurityDescriptor PSECURITY_DESCRIPTOR, nLength uint32, lpnLengthNeeded *uint32) BOOL {
	addr := lazyAddr(&pGetFileSecurityW, libAdvapi32, "GetFileSecurityW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFileName)), uintptr(RequestedInformation), uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(nLength), uintptr(unsafe.Pointer(lpnLengthNeeded)))
	return BOOL(ret)
}

func GetKernelObjectSecurity(Handle HANDLE, RequestedInformation uint32, pSecurityDescriptor PSECURITY_DESCRIPTOR, nLength uint32, lpnLengthNeeded *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetKernelObjectSecurity, libAdvapi32, "GetKernelObjectSecurity")
	ret, _, err := syscall.SyscallN(addr, Handle, uintptr(RequestedInformation), uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(nLength), uintptr(unsafe.Pointer(lpnLengthNeeded)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetLengthSid(pSid PSID) uint32 {
	addr := lazyAddr(&pGetLengthSid, libAdvapi32, "GetLengthSid")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSid)))
	return uint32(ret)
}

func GetPrivateObjectSecurity(ObjectDescriptor PSECURITY_DESCRIPTOR, SecurityInformation uint32, ResultantDescriptor PSECURITY_DESCRIPTOR, DescriptorLength uint32, ReturnLength *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetPrivateObjectSecurity, libAdvapi32, "GetPrivateObjectSecurity")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ObjectDescriptor)), uintptr(SecurityInformation), uintptr(unsafe.Pointer(ResultantDescriptor)), uintptr(DescriptorLength), uintptr(unsafe.Pointer(ReturnLength)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetSecurityDescriptorControl(pSecurityDescriptor PSECURITY_DESCRIPTOR, pControl *uint16, lpdwRevision *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetSecurityDescriptorControl, libAdvapi32, "GetSecurityDescriptorControl")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(unsafe.Pointer(pControl)), uintptr(unsafe.Pointer(lpdwRevision)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetSecurityDescriptorDacl(pSecurityDescriptor PSECURITY_DESCRIPTOR, lpbDaclPresent *int32, pDacl **ACL, lpbDaclDefaulted *int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetSecurityDescriptorDacl, libAdvapi32, "GetSecurityDescriptorDacl")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(unsafe.Pointer(lpbDaclPresent)), uintptr(unsafe.Pointer(pDacl)), uintptr(unsafe.Pointer(lpbDaclDefaulted)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetSecurityDescriptorGroup(pSecurityDescriptor PSECURITY_DESCRIPTOR, pGroup *PSID, lpbGroupDefaulted *int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetSecurityDescriptorGroup, libAdvapi32, "GetSecurityDescriptorGroup")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(unsafe.Pointer(pGroup)), uintptr(unsafe.Pointer(lpbGroupDefaulted)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetSecurityDescriptorLength(pSecurityDescriptor PSECURITY_DESCRIPTOR) uint32 {
	addr := lazyAddr(&pGetSecurityDescriptorLength, libAdvapi32, "GetSecurityDescriptorLength")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)))
	return uint32(ret)
}

func GetSecurityDescriptorOwner(pSecurityDescriptor PSECURITY_DESCRIPTOR, pOwner *PSID, lpbOwnerDefaulted *int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetSecurityDescriptorOwner, libAdvapi32, "GetSecurityDescriptorOwner")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(unsafe.Pointer(pOwner)), uintptr(unsafe.Pointer(lpbOwnerDefaulted)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetSecurityDescriptorRMControl(SecurityDescriptor PSECURITY_DESCRIPTOR, RMControl *byte) uint32 {
	addr := lazyAddr(&pGetSecurityDescriptorRMControl, libAdvapi32, "GetSecurityDescriptorRMControl")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SecurityDescriptor)), uintptr(unsafe.Pointer(RMControl)))
	return uint32(ret)
}

func GetSecurityDescriptorSacl(pSecurityDescriptor PSECURITY_DESCRIPTOR, lpbSaclPresent *int32, pSacl **ACL, lpbSaclDefaulted *int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetSecurityDescriptorSacl, libAdvapi32, "GetSecurityDescriptorSacl")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(unsafe.Pointer(lpbSaclPresent)), uintptr(unsafe.Pointer(pSacl)), uintptr(unsafe.Pointer(lpbSaclDefaulted)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetSidIdentifierAuthority(pSid PSID) (*SID_IDENTIFIER_AUTHORITY, WIN32_ERROR) {
	addr := lazyAddr(&pGetSidIdentifierAuthority, libAdvapi32, "GetSidIdentifierAuthority")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSid)))
	return (*SID_IDENTIFIER_AUTHORITY)(unsafe.Pointer(ret)), WIN32_ERROR(err)
}

func GetSidLengthRequired(nSubAuthorityCount byte) uint32 {
	addr := lazyAddr(&pGetSidLengthRequired, libAdvapi32, "GetSidLengthRequired")
	ret, _, _ := syscall.SyscallN(addr, uintptr(nSubAuthorityCount))
	return uint32(ret)
}

func GetSidSubAuthority(pSid PSID, nSubAuthority uint32) (*uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetSidSubAuthority, libAdvapi32, "GetSidSubAuthority")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSid)), uintptr(nSubAuthority))
	return (*uint32)(unsafe.Pointer(ret)), WIN32_ERROR(err)
}

func GetSidSubAuthorityCount(pSid PSID) (*byte, WIN32_ERROR) {
	addr := lazyAddr(&pGetSidSubAuthorityCount, libAdvapi32, "GetSidSubAuthorityCount")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSid)))
	return (*byte)(unsafe.Pointer(ret)), WIN32_ERROR(err)
}

func GetTokenInformation(TokenHandle HANDLE, TokenInformationClass TOKEN_INFORMATION_CLASS, TokenInformation unsafe.Pointer, TokenInformationLength uint32, ReturnLength *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetTokenInformation, libAdvapi32, "GetTokenInformation")
	ret, _, err := syscall.SyscallN(addr, TokenHandle, uintptr(TokenInformationClass), uintptr(TokenInformation), uintptr(TokenInformationLength), uintptr(unsafe.Pointer(ReturnLength)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetWindowsAccountDomainSid(pSid PSID, pDomainSid PSID, cbDomainSid *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetWindowsAccountDomainSid, libAdvapi32, "GetWindowsAccountDomainSid")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSid)), uintptr(unsafe.Pointer(pDomainSid)), uintptr(unsafe.Pointer(cbDomainSid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ImpersonateAnonymousToken(ThreadHandle HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pImpersonateAnonymousToken, libAdvapi32, "ImpersonateAnonymousToken")
	ret, _, err := syscall.SyscallN(addr, ThreadHandle)
	return BOOL(ret), WIN32_ERROR(err)
}

func ImpersonateLoggedOnUser(hToken HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pImpersonateLoggedOnUser, libAdvapi32, "ImpersonateLoggedOnUser")
	ret, _, err := syscall.SyscallN(addr, hToken)
	return BOOL(ret), WIN32_ERROR(err)
}

func ImpersonateSelf(ImpersonationLevel SECURITY_IMPERSONATION_LEVEL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pImpersonateSelf, libAdvapi32, "ImpersonateSelf")
	ret, _, err := syscall.SyscallN(addr, uintptr(ImpersonationLevel))
	return BOOL(ret), WIN32_ERROR(err)
}

func InitializeAcl(pAcl *ACL, nAclLength uint32, dwAclRevision uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInitializeAcl, libAdvapi32, "InitializeAcl")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(nAclLength), uintptr(dwAclRevision))
	return BOOL(ret), WIN32_ERROR(err)
}

func InitializeSecurityDescriptor(pSecurityDescriptor PSECURITY_DESCRIPTOR, dwRevision uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInitializeSecurityDescriptor, libAdvapi32, "InitializeSecurityDescriptor")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(dwRevision))
	return BOOL(ret), WIN32_ERROR(err)
}

func InitializeSid(Sid PSID, pIdentifierAuthority *SID_IDENTIFIER_AUTHORITY, nSubAuthorityCount byte) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInitializeSid, libAdvapi32, "InitializeSid")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Sid)), uintptr(unsafe.Pointer(pIdentifierAuthority)), uintptr(nSubAuthorityCount))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsTokenRestricted(TokenHandle HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsTokenRestricted, libAdvapi32, "IsTokenRestricted")
	ret, _, err := syscall.SyscallN(addr, TokenHandle)
	return BOOL(ret), WIN32_ERROR(err)
}

func IsValidAcl(pAcl *ACL) BOOL {
	addr := lazyAddr(&pIsValidAcl, libAdvapi32, "IsValidAcl")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)))
	return BOOL(ret)
}

func IsValidSecurityDescriptor(pSecurityDescriptor PSECURITY_DESCRIPTOR) BOOL {
	addr := lazyAddr(&pIsValidSecurityDescriptor, libAdvapi32, "IsValidSecurityDescriptor")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)))
	return BOOL(ret)
}

func IsValidSid(pSid PSID) BOOL {
	addr := lazyAddr(&pIsValidSid, libAdvapi32, "IsValidSid")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSid)))
	return BOOL(ret)
}

func IsWellKnownSid(pSid PSID, WellKnownSidType WELL_KNOWN_SID_TYPE) BOOL {
	addr := lazyAddr(&pIsWellKnownSid, libAdvapi32, "IsWellKnownSid")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSid)), uintptr(WellKnownSidType))
	return BOOL(ret)
}

func MakeAbsoluteSD(pSelfRelativeSecurityDescriptor PSECURITY_DESCRIPTOR, pAbsoluteSecurityDescriptor PSECURITY_DESCRIPTOR, lpdwAbsoluteSecurityDescriptorSize *uint32, pDacl *ACL, lpdwDaclSize *uint32, pSacl *ACL, lpdwSaclSize *uint32, pOwner PSID, lpdwOwnerSize *uint32, pPrimaryGroup PSID, lpdwPrimaryGroupSize *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pMakeAbsoluteSD, libAdvapi32, "MakeAbsoluteSD")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSelfRelativeSecurityDescriptor)), uintptr(unsafe.Pointer(pAbsoluteSecurityDescriptor)), uintptr(unsafe.Pointer(lpdwAbsoluteSecurityDescriptorSize)), uintptr(unsafe.Pointer(pDacl)), uintptr(unsafe.Pointer(lpdwDaclSize)), uintptr(unsafe.Pointer(pSacl)), uintptr(unsafe.Pointer(lpdwSaclSize)), uintptr(unsafe.Pointer(pOwner)), uintptr(unsafe.Pointer(lpdwOwnerSize)), uintptr(unsafe.Pointer(pPrimaryGroup)), uintptr(unsafe.Pointer(lpdwPrimaryGroupSize)))
	return BOOL(ret), WIN32_ERROR(err)
}

func MakeSelfRelativeSD(pAbsoluteSecurityDescriptor PSECURITY_DESCRIPTOR, pSelfRelativeSecurityDescriptor PSECURITY_DESCRIPTOR, lpdwBufferLength *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pMakeSelfRelativeSD, libAdvapi32, "MakeSelfRelativeSD")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAbsoluteSecurityDescriptor)), uintptr(unsafe.Pointer(pSelfRelativeSecurityDescriptor)), uintptr(unsafe.Pointer(lpdwBufferLength)))
	return BOOL(ret), WIN32_ERROR(err)
}

func MapGenericMask(AccessMask *uint32, GenericMapping *GENERIC_MAPPING) {
	addr := lazyAddr(&pMapGenericMask, libAdvapi32, "MapGenericMask")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(AccessMask)), uintptr(unsafe.Pointer(GenericMapping)))
}

var ObjectCloseAuditAlarm = ObjectCloseAuditAlarmW

func ObjectCloseAuditAlarmW(SubsystemName PWSTR, HandleId unsafe.Pointer, GenerateOnClose BOOL) BOOL {
	addr := lazyAddr(&pObjectCloseAuditAlarmW, libAdvapi32, "ObjectCloseAuditAlarmW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), uintptr(GenerateOnClose))
	return BOOL(ret)
}

var ObjectDeleteAuditAlarm = ObjectDeleteAuditAlarmW

func ObjectDeleteAuditAlarmW(SubsystemName PWSTR, HandleId unsafe.Pointer, GenerateOnClose BOOL) BOOL {
	addr := lazyAddr(&pObjectDeleteAuditAlarmW, libAdvapi32, "ObjectDeleteAuditAlarmW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), uintptr(GenerateOnClose))
	return BOOL(ret)
}

var ObjectOpenAuditAlarm = ObjectOpenAuditAlarmW

func ObjectOpenAuditAlarmW(SubsystemName PWSTR, HandleId unsafe.Pointer, ObjectTypeName PWSTR, ObjectName PWSTR, pSecurityDescriptor PSECURITY_DESCRIPTOR, ClientToken HANDLE, DesiredAccess uint32, GrantedAccess uint32, Privileges *PRIVILEGE_SET, ObjectCreation BOOL, AccessGranted BOOL, GenerateOnClose *int32) BOOL {
	addr := lazyAddr(&pObjectOpenAuditAlarmW, libAdvapi32, "ObjectOpenAuditAlarmW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), uintptr(unsafe.Pointer(ObjectTypeName)), uintptr(unsafe.Pointer(ObjectName)), uintptr(unsafe.Pointer(pSecurityDescriptor)), ClientToken, uintptr(DesiredAccess), uintptr(GrantedAccess), uintptr(unsafe.Pointer(Privileges)), uintptr(ObjectCreation), uintptr(AccessGranted), uintptr(unsafe.Pointer(GenerateOnClose)))
	return BOOL(ret)
}

var ObjectPrivilegeAuditAlarm = ObjectPrivilegeAuditAlarmW

func ObjectPrivilegeAuditAlarmW(SubsystemName PWSTR, HandleId unsafe.Pointer, ClientToken HANDLE, DesiredAccess uint32, Privileges *PRIVILEGE_SET, AccessGranted BOOL) BOOL {
	addr := lazyAddr(&pObjectPrivilegeAuditAlarmW, libAdvapi32, "ObjectPrivilegeAuditAlarmW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), ClientToken, uintptr(DesiredAccess), uintptr(unsafe.Pointer(Privileges)), uintptr(AccessGranted))
	return BOOL(ret)
}

func PrivilegeCheck(ClientToken HANDLE, RequiredPrivileges *PRIVILEGE_SET, pfResult *int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pPrivilegeCheck, libAdvapi32, "PrivilegeCheck")
	ret, _, err := syscall.SyscallN(addr, ClientToken, uintptr(unsafe.Pointer(RequiredPrivileges)), uintptr(unsafe.Pointer(pfResult)))
	return BOOL(ret), WIN32_ERROR(err)
}

var PrivilegedServiceAuditAlarm = PrivilegedServiceAuditAlarmW

func PrivilegedServiceAuditAlarmW(SubsystemName PWSTR, ServiceName PWSTR, ClientToken HANDLE, Privileges *PRIVILEGE_SET, AccessGranted BOOL) BOOL {
	addr := lazyAddr(&pPrivilegedServiceAuditAlarmW, libAdvapi32, "PrivilegedServiceAuditAlarmW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(unsafe.Pointer(ServiceName)), ClientToken, uintptr(unsafe.Pointer(Privileges)), uintptr(AccessGranted))
	return BOOL(ret)
}

func QuerySecurityAccessMask(SecurityInformation uint32, DesiredAccess *uint32) {
	addr := lazyAddr(&pQuerySecurityAccessMask, libAdvapi32, "QuerySecurityAccessMask")
	syscall.SyscallN(addr, uintptr(SecurityInformation), uintptr(unsafe.Pointer(DesiredAccess)))
}

func RevertToSelf() (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pRevertToSelf, libAdvapi32, "RevertToSelf")
	ret, _, err := syscall.SyscallN(addr)
	return BOOL(ret), WIN32_ERROR(err)
}

func SetAclInformation(pAcl *ACL, pAclInformation unsafe.Pointer, nAclInformationLength uint32, dwAclInformationClass ACL_INFORMATION_CLASS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetAclInformation, libAdvapi32, "SetAclInformation")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(pAclInformation), uintptr(nAclInformationLength), uintptr(dwAclInformationClass))
	return BOOL(ret), WIN32_ERROR(err)
}

var SetFileSecurity = SetFileSecurityW

func SetFileSecurityW(lpFileName PWSTR, SecurityInformation uint32, pSecurityDescriptor PSECURITY_DESCRIPTOR) BOOL {
	addr := lazyAddr(&pSetFileSecurityW, libAdvapi32, "SetFileSecurityW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFileName)), uintptr(SecurityInformation), uintptr(unsafe.Pointer(pSecurityDescriptor)))
	return BOOL(ret)
}

func SetKernelObjectSecurity(Handle HANDLE, SecurityInformation uint32, SecurityDescriptor PSECURITY_DESCRIPTOR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetKernelObjectSecurity, libAdvapi32, "SetKernelObjectSecurity")
	ret, _, err := syscall.SyscallN(addr, Handle, uintptr(SecurityInformation), uintptr(unsafe.Pointer(SecurityDescriptor)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetPrivateObjectSecurity(SecurityInformation uint32, ModificationDescriptor PSECURITY_DESCRIPTOR, ObjectsSecurityDescriptor *PSECURITY_DESCRIPTOR, GenericMapping *GENERIC_MAPPING, Token HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetPrivateObjectSecurity, libAdvapi32, "SetPrivateObjectSecurity")
	ret, _, err := syscall.SyscallN(addr, uintptr(SecurityInformation), uintptr(unsafe.Pointer(ModificationDescriptor)), uintptr(unsafe.Pointer(ObjectsSecurityDescriptor)), uintptr(unsafe.Pointer(GenericMapping)), Token)
	return BOOL(ret), WIN32_ERROR(err)
}

func SetPrivateObjectSecurityEx(SecurityInformation uint32, ModificationDescriptor PSECURITY_DESCRIPTOR, ObjectsSecurityDescriptor *PSECURITY_DESCRIPTOR, AutoInheritFlags SECURITY_AUTO_INHERIT_FLAGS, GenericMapping *GENERIC_MAPPING, Token HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetPrivateObjectSecurityEx, libAdvapi32, "SetPrivateObjectSecurityEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(SecurityInformation), uintptr(unsafe.Pointer(ModificationDescriptor)), uintptr(unsafe.Pointer(ObjectsSecurityDescriptor)), uintptr(AutoInheritFlags), uintptr(unsafe.Pointer(GenericMapping)), Token)
	return BOOL(ret), WIN32_ERROR(err)
}

func SetSecurityAccessMask(SecurityInformation uint32, DesiredAccess *uint32) {
	addr := lazyAddr(&pSetSecurityAccessMask, libAdvapi32, "SetSecurityAccessMask")
	syscall.SyscallN(addr, uintptr(SecurityInformation), uintptr(unsafe.Pointer(DesiredAccess)))
}

func SetSecurityDescriptorControl(pSecurityDescriptor PSECURITY_DESCRIPTOR, ControlBitsOfInterest uint16, ControlBitsToSet uint16) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetSecurityDescriptorControl, libAdvapi32, "SetSecurityDescriptorControl")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(ControlBitsOfInterest), uintptr(ControlBitsToSet))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetSecurityDescriptorDacl(pSecurityDescriptor PSECURITY_DESCRIPTOR, bDaclPresent BOOL, pDacl *ACL, bDaclDefaulted BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetSecurityDescriptorDacl, libAdvapi32, "SetSecurityDescriptorDacl")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(bDaclPresent), uintptr(unsafe.Pointer(pDacl)), uintptr(bDaclDefaulted))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetSecurityDescriptorGroup(pSecurityDescriptor PSECURITY_DESCRIPTOR, pGroup PSID, bGroupDefaulted BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetSecurityDescriptorGroup, libAdvapi32, "SetSecurityDescriptorGroup")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(unsafe.Pointer(pGroup)), uintptr(bGroupDefaulted))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetSecurityDescriptorOwner(pSecurityDescriptor PSECURITY_DESCRIPTOR, pOwner PSID, bOwnerDefaulted BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetSecurityDescriptorOwner, libAdvapi32, "SetSecurityDescriptorOwner")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(unsafe.Pointer(pOwner)), uintptr(bOwnerDefaulted))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetSecurityDescriptorRMControl(SecurityDescriptor PSECURITY_DESCRIPTOR, RMControl *byte) uint32 {
	addr := lazyAddr(&pSetSecurityDescriptorRMControl, libAdvapi32, "SetSecurityDescriptorRMControl")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SecurityDescriptor)), uintptr(unsafe.Pointer(RMControl)))
	return uint32(ret)
}

func SetSecurityDescriptorSacl(pSecurityDescriptor PSECURITY_DESCRIPTOR, bSaclPresent BOOL, pSacl *ACL, bSaclDefaulted BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetSecurityDescriptorSacl, libAdvapi32, "SetSecurityDescriptorSacl")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(bSaclPresent), uintptr(unsafe.Pointer(pSacl)), uintptr(bSaclDefaulted))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetTokenInformation(TokenHandle HANDLE, TokenInformationClass TOKEN_INFORMATION_CLASS, TokenInformation unsafe.Pointer, TokenInformationLength uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetTokenInformation, libAdvapi32, "SetTokenInformation")
	ret, _, err := syscall.SyscallN(addr, TokenHandle, uintptr(TokenInformationClass), uintptr(TokenInformation), uintptr(TokenInformationLength))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetCachedSigningLevel(SourceFiles *HANDLE, SourceFileCount uint32, Flags uint32, TargetFile HANDLE) BOOL {
	addr := lazyAddr(&pSetCachedSigningLevel, libKernel32, "SetCachedSigningLevel")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SourceFiles)), uintptr(SourceFileCount), uintptr(Flags), TargetFile)
	return BOOL(ret)
}

func GetCachedSigningLevel(File HANDLE, Flags *uint32, SigningLevel *uint32, Thumbprint *byte, ThumbprintSize *uint32, ThumbprintAlgorithm *uint32) BOOL {
	addr := lazyAddr(&pGetCachedSigningLevel, libKernel32, "GetCachedSigningLevel")
	ret, _, _ := syscall.SyscallN(addr, File, uintptr(unsafe.Pointer(Flags)), uintptr(unsafe.Pointer(SigningLevel)), uintptr(unsafe.Pointer(Thumbprint)), uintptr(unsafe.Pointer(ThumbprintSize)), uintptr(unsafe.Pointer(ThumbprintAlgorithm)))
	return BOOL(ret)
}

func SetUserObjectSecurity(hObj HANDLE, pSIRequested *OBJECT_SECURITY_INFORMATION, pSID PSECURITY_DESCRIPTOR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetUserObjectSecurity, libUser32, "SetUserObjectSecurity")
	ret, _, err := syscall.SyscallN(addr, hObj, uintptr(unsafe.Pointer(pSIRequested)), uintptr(unsafe.Pointer(pSID)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetUserObjectSecurity(hObj HANDLE, pSIRequested *uint32, pSID PSECURITY_DESCRIPTOR, nLength uint32, lpnLengthNeeded *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetUserObjectSecurity, libUser32, "GetUserObjectSecurity")
	ret, _, err := syscall.SyscallN(addr, hObj, uintptr(unsafe.Pointer(pSIRequested)), uintptr(unsafe.Pointer(pSID)), uintptr(nLength), uintptr(unsafe.Pointer(lpnLengthNeeded)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AccessCheckAndAuditAlarmA(SubsystemName PSTR, HandleId unsafe.Pointer, ObjectTypeName PSTR, ObjectName PSTR, SecurityDescriptor PSECURITY_DESCRIPTOR, DesiredAccess uint32, GenericMapping *GENERIC_MAPPING, ObjectCreation BOOL, GrantedAccess *uint32, AccessStatus *int32, pfGenerateOnClose *int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAccessCheckAndAuditAlarmA, libAdvapi32, "AccessCheckAndAuditAlarmA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), uintptr(unsafe.Pointer(ObjectTypeName)), uintptr(unsafe.Pointer(ObjectName)), uintptr(unsafe.Pointer(SecurityDescriptor)), uintptr(DesiredAccess), uintptr(unsafe.Pointer(GenericMapping)), uintptr(ObjectCreation), uintptr(unsafe.Pointer(GrantedAccess)), uintptr(unsafe.Pointer(AccessStatus)), uintptr(unsafe.Pointer(pfGenerateOnClose)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AccessCheckByTypeAndAuditAlarmA(SubsystemName PSTR, HandleId unsafe.Pointer, ObjectTypeName PSTR, ObjectName PSTR, SecurityDescriptor PSECURITY_DESCRIPTOR, PrincipalSelfSid PSID, DesiredAccess uint32, AuditType AUDIT_EVENT_TYPE, Flags uint32, ObjectTypeList *OBJECT_TYPE_LIST, ObjectTypeListLength uint32, GenericMapping *GENERIC_MAPPING, ObjectCreation BOOL, GrantedAccess *uint32, AccessStatus *int32, pfGenerateOnClose *int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAccessCheckByTypeAndAuditAlarmA, libAdvapi32, "AccessCheckByTypeAndAuditAlarmA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), uintptr(unsafe.Pointer(ObjectTypeName)), uintptr(unsafe.Pointer(ObjectName)), uintptr(unsafe.Pointer(SecurityDescriptor)), uintptr(unsafe.Pointer(PrincipalSelfSid)), uintptr(DesiredAccess), uintptr(AuditType), uintptr(Flags), uintptr(unsafe.Pointer(ObjectTypeList)), uintptr(ObjectTypeListLength), uintptr(unsafe.Pointer(GenericMapping)), uintptr(ObjectCreation), uintptr(unsafe.Pointer(GrantedAccess)), uintptr(unsafe.Pointer(AccessStatus)), uintptr(unsafe.Pointer(pfGenerateOnClose)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AccessCheckByTypeResultListAndAuditAlarmA(SubsystemName PSTR, HandleId unsafe.Pointer, ObjectTypeName PSTR, ObjectName PSTR, SecurityDescriptor PSECURITY_DESCRIPTOR, PrincipalSelfSid PSID, DesiredAccess uint32, AuditType AUDIT_EVENT_TYPE, Flags uint32, ObjectTypeList *OBJECT_TYPE_LIST, ObjectTypeListLength uint32, GenericMapping *GENERIC_MAPPING, ObjectCreation BOOL, GrantedAccess *uint32, AccessStatusList *uint32, pfGenerateOnClose *int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAccessCheckByTypeResultListAndAuditAlarmA, libAdvapi32, "AccessCheckByTypeResultListAndAuditAlarmA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), uintptr(unsafe.Pointer(ObjectTypeName)), uintptr(unsafe.Pointer(ObjectName)), uintptr(unsafe.Pointer(SecurityDescriptor)), uintptr(unsafe.Pointer(PrincipalSelfSid)), uintptr(DesiredAccess), uintptr(AuditType), uintptr(Flags), uintptr(unsafe.Pointer(ObjectTypeList)), uintptr(ObjectTypeListLength), uintptr(unsafe.Pointer(GenericMapping)), uintptr(ObjectCreation), uintptr(unsafe.Pointer(GrantedAccess)), uintptr(unsafe.Pointer(AccessStatusList)), uintptr(unsafe.Pointer(pfGenerateOnClose)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AccessCheckByTypeResultListAndAuditAlarmByHandleA(SubsystemName PSTR, HandleId unsafe.Pointer, ClientToken HANDLE, ObjectTypeName PSTR, ObjectName PSTR, SecurityDescriptor PSECURITY_DESCRIPTOR, PrincipalSelfSid PSID, DesiredAccess uint32, AuditType AUDIT_EVENT_TYPE, Flags uint32, ObjectTypeList *OBJECT_TYPE_LIST, ObjectTypeListLength uint32, GenericMapping *GENERIC_MAPPING, ObjectCreation BOOL, GrantedAccess *uint32, AccessStatusList *uint32, pfGenerateOnClose *int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAccessCheckByTypeResultListAndAuditAlarmByHandleA, libAdvapi32, "AccessCheckByTypeResultListAndAuditAlarmByHandleA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), ClientToken, uintptr(unsafe.Pointer(ObjectTypeName)), uintptr(unsafe.Pointer(ObjectName)), uintptr(unsafe.Pointer(SecurityDescriptor)), uintptr(unsafe.Pointer(PrincipalSelfSid)), uintptr(DesiredAccess), uintptr(AuditType), uintptr(Flags), uintptr(unsafe.Pointer(ObjectTypeList)), uintptr(ObjectTypeListLength), uintptr(unsafe.Pointer(GenericMapping)), uintptr(ObjectCreation), uintptr(unsafe.Pointer(GrantedAccess)), uintptr(unsafe.Pointer(AccessStatusList)), uintptr(unsafe.Pointer(pfGenerateOnClose)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ObjectOpenAuditAlarmA(SubsystemName PSTR, HandleId unsafe.Pointer, ObjectTypeName PSTR, ObjectName PSTR, pSecurityDescriptor PSECURITY_DESCRIPTOR, ClientToken HANDLE, DesiredAccess uint32, GrantedAccess uint32, Privileges *PRIVILEGE_SET, ObjectCreation BOOL, AccessGranted BOOL, GenerateOnClose *int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pObjectOpenAuditAlarmA, libAdvapi32, "ObjectOpenAuditAlarmA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), uintptr(unsafe.Pointer(ObjectTypeName)), uintptr(unsafe.Pointer(ObjectName)), uintptr(unsafe.Pointer(pSecurityDescriptor)), ClientToken, uintptr(DesiredAccess), uintptr(GrantedAccess), uintptr(unsafe.Pointer(Privileges)), uintptr(ObjectCreation), uintptr(AccessGranted), uintptr(unsafe.Pointer(GenerateOnClose)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ObjectPrivilegeAuditAlarmA(SubsystemName PSTR, HandleId unsafe.Pointer, ClientToken HANDLE, DesiredAccess uint32, Privileges *PRIVILEGE_SET, AccessGranted BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pObjectPrivilegeAuditAlarmA, libAdvapi32, "ObjectPrivilegeAuditAlarmA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), ClientToken, uintptr(DesiredAccess), uintptr(unsafe.Pointer(Privileges)), uintptr(AccessGranted))
	return BOOL(ret), WIN32_ERROR(err)
}

func ObjectCloseAuditAlarmA(SubsystemName PSTR, HandleId unsafe.Pointer, GenerateOnClose BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pObjectCloseAuditAlarmA, libAdvapi32, "ObjectCloseAuditAlarmA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), uintptr(GenerateOnClose))
	return BOOL(ret), WIN32_ERROR(err)
}

func ObjectDeleteAuditAlarmA(SubsystemName PSTR, HandleId unsafe.Pointer, GenerateOnClose BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pObjectDeleteAuditAlarmA, libAdvapi32, "ObjectDeleteAuditAlarmA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(HandleId), uintptr(GenerateOnClose))
	return BOOL(ret), WIN32_ERROR(err)
}

func PrivilegedServiceAuditAlarmA(SubsystemName PSTR, ServiceName PSTR, ClientToken HANDLE, Privileges *PRIVILEGE_SET, AccessGranted BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pPrivilegedServiceAuditAlarmA, libAdvapi32, "PrivilegedServiceAuditAlarmA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SubsystemName)), uintptr(unsafe.Pointer(ServiceName)), ClientToken, uintptr(unsafe.Pointer(Privileges)), uintptr(AccessGranted))
	return BOOL(ret), WIN32_ERROR(err)
}

func AddConditionalAce(pAcl *ACL, dwAceRevision uint32, AceFlags ACE_FLAGS, AceType byte, AccessMask uint32, pSid PSID, ConditionStr PWSTR, ReturnLength *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddConditionalAce, libAdvapi32, "AddConditionalAce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAcl)), uintptr(dwAceRevision), uintptr(AceFlags), uintptr(AceType), uintptr(AccessMask), uintptr(unsafe.Pointer(pSid)), uintptr(unsafe.Pointer(ConditionStr)), uintptr(unsafe.Pointer(ReturnLength)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetFileSecurityA(lpFileName PSTR, SecurityInformation uint32, pSecurityDescriptor PSECURITY_DESCRIPTOR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetFileSecurityA, libAdvapi32, "SetFileSecurityA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFileName)), uintptr(SecurityInformation), uintptr(unsafe.Pointer(pSecurityDescriptor)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetFileSecurityA(lpFileName PSTR, RequestedInformation uint32, pSecurityDescriptor PSECURITY_DESCRIPTOR, nLength uint32, lpnLengthNeeded *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetFileSecurityA, libAdvapi32, "GetFileSecurityA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFileName)), uintptr(RequestedInformation), uintptr(unsafe.Pointer(pSecurityDescriptor)), uintptr(nLength), uintptr(unsafe.Pointer(lpnLengthNeeded)))
	return BOOL(ret), WIN32_ERROR(err)
}

func LookupAccountSidA(lpSystemName PSTR, Sid PSID, Name PSTR, cchName *uint32, ReferencedDomainName PSTR, cchReferencedDomainName *uint32, peUse *SID_NAME_USE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLookupAccountSidA, libAdvapi32, "LookupAccountSidA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSystemName)), uintptr(unsafe.Pointer(Sid)), uintptr(unsafe.Pointer(Name)), uintptr(unsafe.Pointer(cchName)), uintptr(unsafe.Pointer(ReferencedDomainName)), uintptr(unsafe.Pointer(cchReferencedDomainName)), uintptr(unsafe.Pointer(peUse)))
	return BOOL(ret), WIN32_ERROR(err)
}

var LookupAccountSid = LookupAccountSidW

func LookupAccountSidW(lpSystemName PWSTR, Sid PSID, Name PWSTR, cchName *uint32, ReferencedDomainName PWSTR, cchReferencedDomainName *uint32, peUse *SID_NAME_USE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLookupAccountSidW, libAdvapi32, "LookupAccountSidW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSystemName)), uintptr(unsafe.Pointer(Sid)), uintptr(unsafe.Pointer(Name)), uintptr(unsafe.Pointer(cchName)), uintptr(unsafe.Pointer(ReferencedDomainName)), uintptr(unsafe.Pointer(cchReferencedDomainName)), uintptr(unsafe.Pointer(peUse)))
	return BOOL(ret), WIN32_ERROR(err)
}

func LookupAccountNameA(lpSystemName PSTR, lpAccountName PSTR, Sid PSID, cbSid *uint32, ReferencedDomainName PSTR, cchReferencedDomainName *uint32, peUse *SID_NAME_USE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLookupAccountNameA, libAdvapi32, "LookupAccountNameA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSystemName)), uintptr(unsafe.Pointer(lpAccountName)), uintptr(unsafe.Pointer(Sid)), uintptr(unsafe.Pointer(cbSid)), uintptr(unsafe.Pointer(ReferencedDomainName)), uintptr(unsafe.Pointer(cchReferencedDomainName)), uintptr(unsafe.Pointer(peUse)))
	return BOOL(ret), WIN32_ERROR(err)
}

var LookupAccountName = LookupAccountNameW

func LookupAccountNameW(lpSystemName PWSTR, lpAccountName PWSTR, Sid PSID, cbSid *uint32, ReferencedDomainName PWSTR, cchReferencedDomainName *uint32, peUse *SID_NAME_USE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLookupAccountNameW, libAdvapi32, "LookupAccountNameW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSystemName)), uintptr(unsafe.Pointer(lpAccountName)), uintptr(unsafe.Pointer(Sid)), uintptr(unsafe.Pointer(cbSid)), uintptr(unsafe.Pointer(ReferencedDomainName)), uintptr(unsafe.Pointer(cchReferencedDomainName)), uintptr(unsafe.Pointer(peUse)))
	return BOOL(ret), WIN32_ERROR(err)
}

func LookupPrivilegeValueA(lpSystemName PSTR, lpName PSTR, lpLuid *LUID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLookupPrivilegeValueA, libAdvapi32, "LookupPrivilegeValueA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSystemName)), uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(lpLuid)))
	return BOOL(ret), WIN32_ERROR(err)
}

var LookupPrivilegeValue = LookupPrivilegeValueW

func LookupPrivilegeValueW(lpSystemName PWSTR, lpName PWSTR, lpLuid *LUID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLookupPrivilegeValueW, libAdvapi32, "LookupPrivilegeValueW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSystemName)), uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(lpLuid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func LookupPrivilegeNameA(lpSystemName PSTR, lpLuid *LUID, lpName PSTR, cchName *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLookupPrivilegeNameA, libAdvapi32, "LookupPrivilegeNameA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSystemName)), uintptr(unsafe.Pointer(lpLuid)), uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(cchName)))
	return BOOL(ret), WIN32_ERROR(err)
}

var LookupPrivilegeName = LookupPrivilegeNameW

func LookupPrivilegeNameW(lpSystemName PWSTR, lpLuid *LUID, lpName PWSTR, cchName *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLookupPrivilegeNameW, libAdvapi32, "LookupPrivilegeNameW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSystemName)), uintptr(unsafe.Pointer(lpLuid)), uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(cchName)))
	return BOOL(ret), WIN32_ERROR(err)
}

func LookupPrivilegeDisplayNameA(lpSystemName PSTR, lpName PSTR, lpDisplayName PSTR, cchDisplayName *uint32, lpLanguageId *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLookupPrivilegeDisplayNameA, libAdvapi32, "LookupPrivilegeDisplayNameA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSystemName)), uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(lpDisplayName)), uintptr(unsafe.Pointer(cchDisplayName)), uintptr(unsafe.Pointer(lpLanguageId)))
	return BOOL(ret), WIN32_ERROR(err)
}

var LookupPrivilegeDisplayName = LookupPrivilegeDisplayNameW

func LookupPrivilegeDisplayNameW(lpSystemName PWSTR, lpName PWSTR, lpDisplayName PWSTR, cchDisplayName *uint32, lpLanguageId *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLookupPrivilegeDisplayNameW, libAdvapi32, "LookupPrivilegeDisplayNameW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSystemName)), uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(lpDisplayName)), uintptr(unsafe.Pointer(cchDisplayName)), uintptr(unsafe.Pointer(lpLanguageId)))
	return BOOL(ret), WIN32_ERROR(err)
}

func LogonUserA(lpszUsername PSTR, lpszDomain PSTR, lpszPassword PSTR, dwLogonType LOGON32_LOGON, dwLogonProvider LOGON32_PROVIDER, phToken *HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLogonUserA, libAdvapi32, "LogonUserA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszUsername)), uintptr(unsafe.Pointer(lpszDomain)), uintptr(unsafe.Pointer(lpszPassword)), uintptr(dwLogonType), uintptr(dwLogonProvider), uintptr(unsafe.Pointer(phToken)))
	return BOOL(ret), WIN32_ERROR(err)
}

var LogonUser = LogonUserW

func LogonUserW(lpszUsername PWSTR, lpszDomain PWSTR, lpszPassword PWSTR, dwLogonType LOGON32_LOGON, dwLogonProvider LOGON32_PROVIDER, phToken *HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLogonUserW, libAdvapi32, "LogonUserW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszUsername)), uintptr(unsafe.Pointer(lpszDomain)), uintptr(unsafe.Pointer(lpszPassword)), uintptr(dwLogonType), uintptr(dwLogonProvider), uintptr(unsafe.Pointer(phToken)))
	return BOOL(ret), WIN32_ERROR(err)
}

func LogonUserExA(lpszUsername PSTR, lpszDomain PSTR, lpszPassword PSTR, dwLogonType LOGON32_LOGON, dwLogonProvider LOGON32_PROVIDER, phToken *HANDLE, ppLogonSid *PSID, ppProfileBuffer unsafe.Pointer, pdwProfileLength *uint32, pQuotaLimits *QUOTA_LIMITS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLogonUserExA, libAdvapi32, "LogonUserExA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszUsername)), uintptr(unsafe.Pointer(lpszDomain)), uintptr(unsafe.Pointer(lpszPassword)), uintptr(dwLogonType), uintptr(dwLogonProvider), uintptr(unsafe.Pointer(phToken)), uintptr(unsafe.Pointer(ppLogonSid)), uintptr(ppProfileBuffer), uintptr(unsafe.Pointer(pdwProfileLength)), uintptr(unsafe.Pointer(pQuotaLimits)))
	return BOOL(ret), WIN32_ERROR(err)
}

var LogonUserEx = LogonUserExW

func LogonUserExW(lpszUsername PWSTR, lpszDomain PWSTR, lpszPassword PWSTR, dwLogonType LOGON32_LOGON, dwLogonProvider LOGON32_PROVIDER, phToken *HANDLE, ppLogonSid *PSID, ppProfileBuffer unsafe.Pointer, pdwProfileLength *uint32, pQuotaLimits *QUOTA_LIMITS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLogonUserExW, libAdvapi32, "LogonUserExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszUsername)), uintptr(unsafe.Pointer(lpszDomain)), uintptr(unsafe.Pointer(lpszPassword)), uintptr(dwLogonType), uintptr(dwLogonProvider), uintptr(unsafe.Pointer(phToken)), uintptr(unsafe.Pointer(ppLogonSid)), uintptr(ppProfileBuffer), uintptr(unsafe.Pointer(pdwProfileLength)), uintptr(unsafe.Pointer(pQuotaLimits)))
	return BOOL(ret), WIN32_ERROR(err)
}
