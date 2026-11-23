package win32

import (
	"syscall"
	"unsafe"
)

const (
	MM_HINT_T0_                                                  uint32   = 0x1
	MM_HINT_T1_                                                  uint32   = 0x2
	MM_HINT_T2_                                                  uint32   = 0x3
	MM_HINT_NTA_                                                 uint32   = 0x0
	ANYSIZE_ARRAY                                                uint32   = 0x1
	MEMORY_ALLOCATION_ALIGNMENT                                  uint32   = 0x10
	X86_CACHE_ALIGNMENT_SIZE                                     uint32   = 0x40
	ARM_CACHE_ALIGNMENT_SIZE                                     uint32   = 0x80
	SYSTEM_CACHE_ALIGNMENT_SIZE                                  uint32   = 0x40
	PRAGMA_DEPRECATED_DDK                                        uint32   = 0x1
	UCSCHAR_INVALID_CHARACTER                                    uint32   = 0xffffffff
	MIN_UCSCHAR                                                  uint32   = 0x0
	MAX_UCSCHAR                                                  uint32   = 0x10ffff
	ALL_PROCESSOR_GROUPS                                         uint32   = 0xffff
	MAXIMUM_PROC_PER_GROUP                                       uint32   = 0x40
	MAXIMUM_PROCESSORS                                           uint32   = 0x40
	APPLICATION_ERROR_MASK                                       uint32   = 0x20000000
	ERROR_SEVERITY_SUCCESS                                       uint32   = 0x0
	ERROR_SEVERITY_INFORMATIONAL                                 uint32   = 0x40000000
	ERROR_SEVERITY_WARNING                                       uint32   = 0x80000000
	ERROR_SEVERITY_ERROR                                         uint32   = 0xc0000000
	MAXLONGLONG                                                  uint64   = 0x7fffffffffffffff
	UNICODE_STRING_MAX_CHARS                                     uint32   = 0x7fff
	MINCHAR                                                      uint32   = 0x80
	MAXCHAR                                                      uint32   = 0x7f
	MINSHORT                                                     uint32   = 0x8000
	MAXSHORT                                                     uint32   = 0x7fff
	MINLONG                                                      uint32   = 0x80000000
	MAXLONG                                                      uint32   = 0x7fffffff
	MAXBYTE                                                      uint32   = 0xff
	MAXWORD                                                      uint32   = 0xffff
	MAXDWORD                                                     uint32   = 0xffffffff
	ENCLAVE_SHORT_ID_LENGTH                                      uint32   = 0x10
	ENCLAVE_LONG_ID_LENGTH                                       uint32   = 0x20
	VER_SERVER_NT                                                uint32   = 0x80000000
	VER_WORKSTATION_NT                                           uint32   = 0x40000000
	VER_SUITE_SMALLBUSINESS                                      uint32   = 0x1
	VER_SUITE_ENTERPRISE                                         uint32   = 0x2
	VER_SUITE_BACKOFFICE                                         uint32   = 0x4
	VER_SUITE_COMMUNICATIONS                                     uint32   = 0x8
	VER_SUITE_TERMINAL                                           uint32   = 0x10
	VER_SUITE_SMALLBUSINESS_RESTRICTED                           uint32   = 0x20
	VER_SUITE_EMBEDDEDNT                                         uint32   = 0x40
	VER_SUITE_DATACENTER                                         uint32   = 0x80
	VER_SUITE_SINGLEUSERTS                                       uint32   = 0x100
	VER_SUITE_PERSONAL                                           uint32   = 0x200
	VER_SUITE_BLADE                                              uint32   = 0x400
	VER_SUITE_EMBEDDED_RESTRICTED                                uint32   = 0x800
	VER_SUITE_SECURITY_APPLIANCE                                 uint32   = 0x1000
	VER_SUITE_STORAGE_SERVER                                     uint32   = 0x2000
	VER_SUITE_COMPUTE_SERVER                                     uint32   = 0x4000
	VER_SUITE_WH_SERVER                                          uint32   = 0x8000
	VER_SUITE_MULTIUSERTS                                        uint32   = 0x20000
	PRODUCT_STANDARD_SERVER_CORE                                 uint32   = 0xd
	PRODUCT_SOLUTION_EMBEDDEDSERVER_CORE                         uint32   = 0x39
	PRODUCT_PROFESSIONAL_EMBEDDED                                uint32   = 0x3a
	PRODUCT_EMBEDDED                                             uint32   = 0x41
	PRODUCT_EMBEDDED_AUTOMOTIVE                                  uint32   = 0x55
	PRODUCT_EMBEDDED_INDUSTRY_A                                  uint32   = 0x56
	PRODUCT_THINPC                                               uint32   = 0x57
	PRODUCT_EMBEDDED_A                                           uint32   = 0x58
	PRODUCT_EMBEDDED_INDUSTRY                                    uint32   = 0x59
	PRODUCT_EMBEDDED_E                                           uint32   = 0x5a
	PRODUCT_EMBEDDED_INDUSTRY_E                                  uint32   = 0x5b
	PRODUCT_EMBEDDED_INDUSTRY_A_E                                uint32   = 0x5c
	PRODUCT_CORE_ARM                                             uint32   = 0x61
	PRODUCT_EMBEDDED_INDUSTRY_EVAL                               uint32   = 0x69
	PRODUCT_EMBEDDED_INDUSTRY_E_EVAL                             uint32   = 0x6a
	PRODUCT_EMBEDDED_EVAL                                        uint32   = 0x6b
	PRODUCT_EMBEDDED_E_EVAL                                      uint32   = 0x6c
	PRODUCT_NANO_SERVER                                          uint32   = 0x6d
	PRODUCT_CLOUD_STORAGE_SERVER                                 uint32   = 0x6e
	PRODUCT_CORE_CONNECTED                                       uint32   = 0x6f
	PRODUCT_PROFESSIONAL_STUDENT                                 uint32   = 0x70
	PRODUCT_CORE_CONNECTED_N                                     uint32   = 0x71
	PRODUCT_PROFESSIONAL_STUDENT_N                               uint32   = 0x72
	PRODUCT_CORE_CONNECTED_SINGLELANGUAGE                        uint32   = 0x73
	PRODUCT_CORE_CONNECTED_COUNTRYSPECIFIC                       uint32   = 0x74
	PRODUCT_CONNECTED_CAR                                        uint32   = 0x75
	PRODUCT_INDUSTRY_HANDHELD                                    uint32   = 0x76
	PRODUCT_PPI_PRO                                              uint32   = 0x77
	PRODUCT_ARM64_SERVER                                         uint32   = 0x78
	PRODUCT_CLOUD_HOST_INFRASTRUCTURE_SERVER                     uint32   = 0x7c
	PRODUCT_PROFESSIONAL_S                                       uint32   = 0x7f
	PRODUCT_PROFESSIONAL_S_N                                     uint32   = 0x80
	PRODUCT_HOLOGRAPHIC                                          uint32   = 0x87
	PRODUCT_HOLOGRAPHIC_BUSINESS                                 uint32   = 0x88
	PRODUCT_PRO_SINGLE_LANGUAGE                                  uint32   = 0x8a
	PRODUCT_PRO_CHINA                                            uint32   = 0x8b
	PRODUCT_ENTERPRISE_SUBSCRIPTION                              uint32   = 0x8c
	PRODUCT_ENTERPRISE_SUBSCRIPTION_N                            uint32   = 0x8d
	PRODUCT_DATACENTER_NANO_SERVER                               uint32   = 0x8f
	PRODUCT_STANDARD_NANO_SERVER                                 uint32   = 0x90
	PRODUCT_DATACENTER_WS_SERVER_CORE                            uint32   = 0x93
	PRODUCT_STANDARD_WS_SERVER_CORE                              uint32   = 0x94
	PRODUCT_UTILITY_VM                                           uint32   = 0x95
	PRODUCT_DATACENTER_EVALUATION_SERVER_CORE                    uint32   = 0x9f
	PRODUCT_STANDARD_EVALUATION_SERVER_CORE                      uint32   = 0xa0
	PRODUCT_PRO_FOR_EDUCATION                                    uint32   = 0xa4
	PRODUCT_PRO_FOR_EDUCATION_N                                  uint32   = 0xa5
	PRODUCT_AZURE_SERVER_CORE                                    uint32   = 0xa8
	PRODUCT_AZURE_NANO_SERVER                                    uint32   = 0xa9
	PRODUCT_ENTERPRISEG                                          uint32   = 0xab
	PRODUCT_ENTERPRISEGN                                         uint32   = 0xac
	PRODUCT_SERVERRDSH                                           uint32   = 0xaf
	PRODUCT_CLOUD                                                uint32   = 0xb2
	PRODUCT_CLOUDN                                               uint32   = 0xb3
	PRODUCT_HUBOS                                                uint32   = 0xb4
	PRODUCT_ONECOREUPDATEOS                                      uint32   = 0xb6
	PRODUCT_CLOUDE                                               uint32   = 0xb7
	PRODUCT_IOTOS                                                uint32   = 0xb9
	PRODUCT_CLOUDEN                                              uint32   = 0xba
	PRODUCT_IOTEDGEOS                                            uint32   = 0xbb
	PRODUCT_IOTENTERPRISE                                        uint32   = 0xbc
	PRODUCT_LITE                                                 uint32   = 0xbd
	PRODUCT_IOTENTERPRISES                                       uint32   = 0xbf
	PRODUCT_XBOX_SYSTEMOS                                        uint32   = 0xc0
	PRODUCT_XBOX_NATIVEOS                                        uint32   = 0xc1
	PRODUCT_XBOX_GAMEOS                                          uint32   = 0xc2
	PRODUCT_XBOX_ERAOS                                           uint32   = 0xc3
	PRODUCT_XBOX_DURANGOHOSTOS                                   uint32   = 0xc4
	PRODUCT_XBOX_SCARLETTHOSTOS                                  uint32   = 0xc5
	PRODUCT_AZURE_SERVER_CLOUDHOST                               uint32   = 0xc7
	PRODUCT_AZURE_SERVER_CLOUDMOS                                uint32   = 0xc8
	PRODUCT_CLOUDEDITIONN                                        uint32   = 0xca
	PRODUCT_CLOUDEDITION                                         uint32   = 0xcb
	PRODUCT_AZURESTACKHCI_SERVER_CORE                            uint32   = 0x196
	PRODUCT_DATACENTER_SERVER_AZURE_EDITION                      uint32   = 0x197
	PRODUCT_DATACENTER_SERVER_CORE_AZURE_EDITION                 uint32   = 0x198
	PRODUCT_UNLICENSED                                           uint32   = 0xabcdabcd
	LANG_NEUTRAL                                                 uint32   = 0x0
	LANG_INVARIANT                                               uint32   = 0x7f
	LANG_AFRIKAANS                                               uint32   = 0x36
	LANG_ALBANIAN                                                uint32   = 0x1c
	LANG_ALSATIAN                                                uint32   = 0x84
	LANG_AMHARIC                                                 uint32   = 0x5e
	LANG_ARABIC                                                  uint32   = 0x1
	LANG_ARMENIAN                                                uint32   = 0x2b
	LANG_ASSAMESE                                                uint32   = 0x4d
	LANG_AZERI                                                   uint32   = 0x2c
	LANG_AZERBAIJANI                                             uint32   = 0x2c
	LANG_BANGLA                                                  uint32   = 0x45
	LANG_BASHKIR                                                 uint32   = 0x6d
	LANG_BASQUE                                                  uint32   = 0x2d
	LANG_BELARUSIAN                                              uint32   = 0x23
	LANG_BENGALI                                                 uint32   = 0x45
	LANG_BRETON                                                  uint32   = 0x7e
	LANG_BOSNIAN                                                 uint32   = 0x1a
	LANG_BOSNIAN_NEUTRAL                                         uint32   = 0x781a
	LANG_BULGARIAN                                               uint32   = 0x2
	LANG_CATALAN                                                 uint32   = 0x3
	LANG_CENTRAL_KURDISH                                         uint32   = 0x92
	LANG_CHEROKEE                                                uint32   = 0x5c
	LANG_CHINESE                                                 uint32   = 0x4
	LANG_CHINESE_SIMPLIFIED                                      uint32   = 0x4
	LANG_CHINESE_TRADITIONAL                                     uint32   = 0x7c04
	LANG_CORSICAN                                                uint32   = 0x83
	LANG_CROATIAN                                                uint32   = 0x1a
	LANG_CZECH                                                   uint32   = 0x5
	LANG_DANISH                                                  uint32   = 0x6
	LANG_DARI                                                    uint32   = 0x8c
	LANG_DIVEHI                                                  uint32   = 0x65
	LANG_DUTCH                                                   uint32   = 0x13
	LANG_ENGLISH                                                 uint32   = 0x9
	LANG_ESTONIAN                                                uint32   = 0x25
	LANG_FAEROESE                                                uint32   = 0x38
	LANG_FARSI                                                   uint32   = 0x29
	LANG_FILIPINO                                                uint32   = 0x64
	LANG_FINNISH                                                 uint32   = 0xb
	LANG_FRENCH                                                  uint32   = 0xc
	LANG_FRISIAN                                                 uint32   = 0x62
	LANG_FULAH                                                   uint32   = 0x67
	LANG_GALICIAN                                                uint32   = 0x56
	LANG_GEORGIAN                                                uint32   = 0x37
	LANG_GERMAN                                                  uint32   = 0x7
	LANG_GREEK                                                   uint32   = 0x8
	LANG_GREENLANDIC                                             uint32   = 0x6f
	LANG_GUJARATI                                                uint32   = 0x47
	LANG_HAUSA                                                   uint32   = 0x68
	LANG_HAWAIIAN                                                uint32   = 0x75
	LANG_HEBREW                                                  uint32   = 0xd
	LANG_HINDI                                                   uint32   = 0x39
	LANG_HUNGARIAN                                               uint32   = 0xe
	LANG_ICELANDIC                                               uint32   = 0xf
	LANG_IGBO                                                    uint32   = 0x70
	LANG_INDONESIAN                                              uint32   = 0x21
	LANG_INUKTITUT                                               uint32   = 0x5d
	LANG_IRISH                                                   uint32   = 0x3c
	LANG_ITALIAN                                                 uint32   = 0x10
	LANG_JAPANESE                                                uint32   = 0x11
	LANG_KANNADA                                                 uint32   = 0x4b
	LANG_KASHMIRI                                                uint32   = 0x60
	LANG_KAZAK                                                   uint32   = 0x3f
	LANG_KHMER                                                   uint32   = 0x53
	LANG_KICHE                                                   uint32   = 0x86
	LANG_KINYARWANDA                                             uint32   = 0x87
	LANG_KONKANI                                                 uint32   = 0x57
	LANG_KOREAN                                                  uint32   = 0x12
	LANG_KYRGYZ                                                  uint32   = 0x40
	LANG_LAO                                                     uint32   = 0x54
	LANG_LATVIAN                                                 uint32   = 0x26
	LANG_LITHUANIAN                                              uint32   = 0x27
	LANG_LOWER_SORBIAN                                           uint32   = 0x2e
	LANG_LUXEMBOURGISH                                           uint32   = 0x6e
	LANG_MACEDONIAN                                              uint32   = 0x2f
	LANG_MALAY                                                   uint32   = 0x3e
	LANG_MALAYALAM                                               uint32   = 0x4c
	LANG_MALTESE                                                 uint32   = 0x3a
	LANG_MANIPURI                                                uint32   = 0x58
	LANG_MAORI                                                   uint32   = 0x81
	LANG_MAPUDUNGUN                                              uint32   = 0x7a
	LANG_MARATHI                                                 uint32   = 0x4e
	LANG_MOHAWK                                                  uint32   = 0x7c
	LANG_MONGOLIAN                                               uint32   = 0x50
	LANG_NEPALI                                                  uint32   = 0x61
	LANG_NORWEGIAN                                               uint32   = 0x14
	LANG_OCCITAN                                                 uint32   = 0x82
	LANG_ODIA                                                    uint32   = 0x48
	LANG_ORIYA                                                   uint32   = 0x48
	LANG_PASHTO                                                  uint32   = 0x63
	LANG_PERSIAN                                                 uint32   = 0x29
	LANG_POLISH                                                  uint32   = 0x15
	LANG_PORTUGUESE                                              uint32   = 0x16
	LANG_PULAR                                                   uint32   = 0x67
	LANG_PUNJABI                                                 uint32   = 0x46
	LANG_QUECHUA                                                 uint32   = 0x6b
	LANG_ROMANIAN                                                uint32   = 0x18
	LANG_ROMANSH                                                 uint32   = 0x17
	LANG_RUSSIAN                                                 uint32   = 0x19
	LANG_SAKHA                                                   uint32   = 0x85
	LANG_SAMI                                                    uint32   = 0x3b
	LANG_SANSKRIT                                                uint32   = 0x4f
	LANG_SCOTTISH_GAELIC                                         uint32   = 0x91
	LANG_SERBIAN                                                 uint32   = 0x1a
	LANG_SERBIAN_NEUTRAL                                         uint32   = 0x7c1a
	LANG_SINDHI                                                  uint32   = 0x59
	LANG_SINHALESE                                               uint32   = 0x5b
	LANG_SLOVAK                                                  uint32   = 0x1b
	LANG_SLOVENIAN                                               uint32   = 0x24
	LANG_SOTHO                                                   uint32   = 0x6c
	LANG_SPANISH                                                 uint32   = 0xa
	LANG_SWAHILI                                                 uint32   = 0x41
	LANG_SWEDISH                                                 uint32   = 0x1d
	LANG_SYRIAC                                                  uint32   = 0x5a
	LANG_TAJIK                                                   uint32   = 0x28
	LANG_TAMAZIGHT                                               uint32   = 0x5f
	LANG_TAMIL                                                   uint32   = 0x49
	LANG_TATAR                                                   uint32   = 0x44
	LANG_TELUGU                                                  uint32   = 0x4a
	LANG_THAI                                                    uint32   = 0x1e
	LANG_TIBETAN                                                 uint32   = 0x51
	LANG_TIGRIGNA                                                uint32   = 0x73
	LANG_TIGRINYA                                                uint32   = 0x73
	LANG_TSWANA                                                  uint32   = 0x32
	LANG_TURKISH                                                 uint32   = 0x1f
	LANG_TURKMEN                                                 uint32   = 0x42
	LANG_UIGHUR                                                  uint32   = 0x80
	LANG_UKRAINIAN                                               uint32   = 0x22
	LANG_UPPER_SORBIAN                                           uint32   = 0x2e
	LANG_URDU                                                    uint32   = 0x20
	LANG_UZBEK                                                   uint32   = 0x43
	LANG_VALENCIAN                                               uint32   = 0x3
	LANG_VIETNAMESE                                              uint32   = 0x2a
	LANG_WELSH                                                   uint32   = 0x52
	LANG_WOLOF                                                   uint32   = 0x88
	LANG_XHOSA                                                   uint32   = 0x34
	LANG_YAKUT                                                   uint32   = 0x85
	LANG_YI                                                      uint32   = 0x78
	LANG_YORUBA                                                  uint32   = 0x6a
	LANG_ZULU                                                    uint32   = 0x35
	SUBLANG_NEUTRAL                                              uint32   = 0x0
	SUBLANG_DEFAULT                                              uint32   = 0x1
	SUBLANG_SYS_DEFAULT                                          uint32   = 0x2
	SUBLANG_CUSTOM_DEFAULT                                       uint32   = 0x3
	SUBLANG_CUSTOM_UNSPECIFIED                                   uint32   = 0x4
	SUBLANG_UI_CUSTOM_DEFAULT                                    uint32   = 0x5
	SUBLANG_AFRIKAANS_SOUTH_AFRICA                               uint32   = 0x1
	SUBLANG_ALBANIAN_ALBANIA                                     uint32   = 0x1
	SUBLANG_ALSATIAN_FRANCE                                      uint32   = 0x1
	SUBLANG_AMHARIC_ETHIOPIA                                     uint32   = 0x1
	SUBLANG_ARABIC_SAUDI_ARABIA                                  uint32   = 0x1
	SUBLANG_ARABIC_IRAQ                                          uint32   = 0x2
	SUBLANG_ARABIC_EGYPT                                         uint32   = 0x3
	SUBLANG_ARABIC_LIBYA                                         uint32   = 0x4
	SUBLANG_ARABIC_ALGERIA                                       uint32   = 0x5
	SUBLANG_ARABIC_MOROCCO                                       uint32   = 0x6
	SUBLANG_ARABIC_TUNISIA                                       uint32   = 0x7
	SUBLANG_ARABIC_OMAN                                          uint32   = 0x8
	SUBLANG_ARABIC_YEMEN                                         uint32   = 0x9
	SUBLANG_ARABIC_SYRIA                                         uint32   = 0xa
	SUBLANG_ARABIC_JORDAN                                        uint32   = 0xb
	SUBLANG_ARABIC_LEBANON                                       uint32   = 0xc
	SUBLANG_ARABIC_KUWAIT                                        uint32   = 0xd
	SUBLANG_ARABIC_UAE                                           uint32   = 0xe
	SUBLANG_ARABIC_BAHRAIN                                       uint32   = 0xf
	SUBLANG_ARABIC_QATAR                                         uint32   = 0x10
	SUBLANG_ARMENIAN_ARMENIA                                     uint32   = 0x1
	SUBLANG_ASSAMESE_INDIA                                       uint32   = 0x1
	SUBLANG_AZERI_LATIN                                          uint32   = 0x1
	SUBLANG_AZERI_CYRILLIC                                       uint32   = 0x2
	SUBLANG_AZERBAIJANI_AZERBAIJAN_LATIN                         uint32   = 0x1
	SUBLANG_AZERBAIJANI_AZERBAIJAN_CYRILLIC                      uint32   = 0x2
	SUBLANG_BANGLA_INDIA                                         uint32   = 0x1
	SUBLANG_BANGLA_BANGLADESH                                    uint32   = 0x2
	SUBLANG_BASHKIR_RUSSIA                                       uint32   = 0x1
	SUBLANG_BASQUE_BASQUE                                        uint32   = 0x1
	SUBLANG_BELARUSIAN_BELARUS                                   uint32   = 0x1
	SUBLANG_BENGALI_INDIA                                        uint32   = 0x1
	SUBLANG_BENGALI_BANGLADESH                                   uint32   = 0x2
	SUBLANG_BOSNIAN_BOSNIA_HERZEGOVINA_LATIN                     uint32   = 0x5
	SUBLANG_BOSNIAN_BOSNIA_HERZEGOVINA_CYRILLIC                  uint32   = 0x8
	SUBLANG_BRETON_FRANCE                                        uint32   = 0x1
	SUBLANG_BULGARIAN_BULGARIA                                   uint32   = 0x1
	SUBLANG_CATALAN_CATALAN                                      uint32   = 0x1
	SUBLANG_CENTRAL_KURDISH_IRAQ                                 uint32   = 0x1
	SUBLANG_CHEROKEE_CHEROKEE                                    uint32   = 0x1
	SUBLANG_CHINESE_TRADITIONAL                                  uint32   = 0x1
	SUBLANG_CHINESE_SIMPLIFIED                                   uint32   = 0x2
	SUBLANG_CHINESE_HONGKONG                                     uint32   = 0x3
	SUBLANG_CHINESE_SINGAPORE                                    uint32   = 0x4
	SUBLANG_CHINESE_MACAU                                        uint32   = 0x5
	SUBLANG_CORSICAN_FRANCE                                      uint32   = 0x1
	SUBLANG_CZECH_CZECH_REPUBLIC                                 uint32   = 0x1
	SUBLANG_CROATIAN_CROATIA                                     uint32   = 0x1
	SUBLANG_CROATIAN_BOSNIA_HERZEGOVINA_LATIN                    uint32   = 0x4
	SUBLANG_DANISH_DENMARK                                       uint32   = 0x1
	SUBLANG_DARI_AFGHANISTAN                                     uint32   = 0x1
	SUBLANG_DIVEHI_MALDIVES                                      uint32   = 0x1
	SUBLANG_DUTCH                                                uint32   = 0x1
	SUBLANG_DUTCH_BELGIAN                                        uint32   = 0x2
	SUBLANG_ENGLISH_US                                           uint32   = 0x1
	SUBLANG_ENGLISH_UK                                           uint32   = 0x2
	SUBLANG_ENGLISH_AUS                                          uint32   = 0x3
	SUBLANG_ENGLISH_CAN                                          uint32   = 0x4
	SUBLANG_ENGLISH_NZ                                           uint32   = 0x5
	SUBLANG_ENGLISH_EIRE                                         uint32   = 0x6
	SUBLANG_ENGLISH_SOUTH_AFRICA                                 uint32   = 0x7
	SUBLANG_ENGLISH_JAMAICA                                      uint32   = 0x8
	SUBLANG_ENGLISH_CARIBBEAN                                    uint32   = 0x9
	SUBLANG_ENGLISH_BELIZE                                       uint32   = 0xa
	SUBLANG_ENGLISH_TRINIDAD                                     uint32   = 0xb
	SUBLANG_ENGLISH_ZIMBABWE                                     uint32   = 0xc
	SUBLANG_ENGLISH_PHILIPPINES                                  uint32   = 0xd
	SUBLANG_ENGLISH_INDIA                                        uint32   = 0x10
	SUBLANG_ENGLISH_MALAYSIA                                     uint32   = 0x11
	SUBLANG_ENGLISH_SINGAPORE                                    uint32   = 0x12
	SUBLANG_ESTONIAN_ESTONIA                                     uint32   = 0x1
	SUBLANG_FAEROESE_FAROE_ISLANDS                               uint32   = 0x1
	SUBLANG_FILIPINO_PHILIPPINES                                 uint32   = 0x1
	SUBLANG_FINNISH_FINLAND                                      uint32   = 0x1
	SUBLANG_FRENCH                                               uint32   = 0x1
	SUBLANG_FRENCH_BELGIAN                                       uint32   = 0x2
	SUBLANG_FRENCH_CANADIAN                                      uint32   = 0x3
	SUBLANG_FRENCH_SWISS                                         uint32   = 0x4
	SUBLANG_FRENCH_LUXEMBOURG                                    uint32   = 0x5
	SUBLANG_FRENCH_MONACO                                        uint32   = 0x6
	SUBLANG_FRISIAN_NETHERLANDS                                  uint32   = 0x1
	SUBLANG_FULAH_SENEGAL                                        uint32   = 0x2
	SUBLANG_GALICIAN_GALICIAN                                    uint32   = 0x1
	SUBLANG_GEORGIAN_GEORGIA                                     uint32   = 0x1
	SUBLANG_GERMAN                                               uint32   = 0x1
	SUBLANG_GERMAN_SWISS                                         uint32   = 0x2
	SUBLANG_GERMAN_AUSTRIAN                                      uint32   = 0x3
	SUBLANG_GERMAN_LUXEMBOURG                                    uint32   = 0x4
	SUBLANG_GERMAN_LIECHTENSTEIN                                 uint32   = 0x5
	SUBLANG_GREEK_GREECE                                         uint32   = 0x1
	SUBLANG_GREENLANDIC_GREENLAND                                uint32   = 0x1
	SUBLANG_GUJARATI_INDIA                                       uint32   = 0x1
	SUBLANG_HAUSA_NIGERIA_LATIN                                  uint32   = 0x1
	SUBLANG_HAWAIIAN_US                                          uint32   = 0x1
	SUBLANG_HEBREW_ISRAEL                                        uint32   = 0x1
	SUBLANG_HINDI_INDIA                                          uint32   = 0x1
	SUBLANG_HUNGARIAN_HUNGARY                                    uint32   = 0x1
	SUBLANG_ICELANDIC_ICELAND                                    uint32   = 0x1
	SUBLANG_IGBO_NIGERIA                                         uint32   = 0x1
	SUBLANG_INDONESIAN_INDONESIA                                 uint32   = 0x1
	SUBLANG_INUKTITUT_CANADA                                     uint32   = 0x1
	SUBLANG_INUKTITUT_CANADA_LATIN                               uint32   = 0x2
	SUBLANG_IRISH_IRELAND                                        uint32   = 0x2
	SUBLANG_ITALIAN                                              uint32   = 0x1
	SUBLANG_ITALIAN_SWISS                                        uint32   = 0x2
	SUBLANG_JAPANESE_JAPAN                                       uint32   = 0x1
	SUBLANG_KANNADA_INDIA                                        uint32   = 0x1
	SUBLANG_KASHMIRI_SASIA                                       uint32   = 0x2
	SUBLANG_KASHMIRI_INDIA                                       uint32   = 0x2
	SUBLANG_KAZAK_KAZAKHSTAN                                     uint32   = 0x1
	SUBLANG_KHMER_CAMBODIA                                       uint32   = 0x1
	SUBLANG_KICHE_GUATEMALA                                      uint32   = 0x1
	SUBLANG_KINYARWANDA_RWANDA                                   uint32   = 0x1
	SUBLANG_KONKANI_INDIA                                        uint32   = 0x1
	SUBLANG_KOREAN                                               uint32   = 0x1
	SUBLANG_KYRGYZ_KYRGYZSTAN                                    uint32   = 0x1
	SUBLANG_LAO_LAO                                              uint32   = 0x1
	SUBLANG_LATVIAN_LATVIA                                       uint32   = 0x1
	SUBLANG_LITHUANIAN                                           uint32   = 0x1
	SUBLANG_LOWER_SORBIAN_GERMANY                                uint32   = 0x2
	SUBLANG_LUXEMBOURGISH_LUXEMBOURG                             uint32   = 0x1
	SUBLANG_MACEDONIAN_MACEDONIA                                 uint32   = 0x1
	SUBLANG_MALAY_MALAYSIA                                       uint32   = 0x1
	SUBLANG_MALAY_BRUNEI_DARUSSALAM                              uint32   = 0x2
	SUBLANG_MALAYALAM_INDIA                                      uint32   = 0x1
	SUBLANG_MALTESE_MALTA                                        uint32   = 0x1
	SUBLANG_MAORI_NEW_ZEALAND                                    uint32   = 0x1
	SUBLANG_MAPUDUNGUN_CHILE                                     uint32   = 0x1
	SUBLANG_MARATHI_INDIA                                        uint32   = 0x1
	SUBLANG_MOHAWK_MOHAWK                                        uint32   = 0x1
	SUBLANG_MONGOLIAN_CYRILLIC_MONGOLIA                          uint32   = 0x1
	SUBLANG_MONGOLIAN_PRC                                        uint32   = 0x2
	SUBLANG_NEPALI_INDIA                                         uint32   = 0x2
	SUBLANG_NEPALI_NEPAL                                         uint32   = 0x1
	SUBLANG_NORWEGIAN_BOKMAL                                     uint32   = 0x1
	SUBLANG_NORWEGIAN_NYNORSK                                    uint32   = 0x2
	SUBLANG_OCCITAN_FRANCE                                       uint32   = 0x1
	SUBLANG_ODIA_INDIA                                           uint32   = 0x1
	SUBLANG_ORIYA_INDIA                                          uint32   = 0x1
	SUBLANG_PASHTO_AFGHANISTAN                                   uint32   = 0x1
	SUBLANG_PERSIAN_IRAN                                         uint32   = 0x1
	SUBLANG_POLISH_POLAND                                        uint32   = 0x1
	SUBLANG_PORTUGUESE                                           uint32   = 0x2
	SUBLANG_PORTUGUESE_BRAZILIAN                                 uint32   = 0x1
	SUBLANG_PULAR_SENEGAL                                        uint32   = 0x2
	SUBLANG_PUNJABI_INDIA                                        uint32   = 0x1
	SUBLANG_PUNJABI_PAKISTAN                                     uint32   = 0x2
	SUBLANG_QUECHUA_BOLIVIA                                      uint32   = 0x1
	SUBLANG_QUECHUA_ECUADOR                                      uint32   = 0x2
	SUBLANG_QUECHUA_PERU                                         uint32   = 0x3
	SUBLANG_ROMANIAN_ROMANIA                                     uint32   = 0x1
	SUBLANG_ROMANSH_SWITZERLAND                                  uint32   = 0x1
	SUBLANG_RUSSIAN_RUSSIA                                       uint32   = 0x1
	SUBLANG_SAKHA_RUSSIA                                         uint32   = 0x1
	SUBLANG_SAMI_NORTHERN_NORWAY                                 uint32   = 0x1
	SUBLANG_SAMI_NORTHERN_SWEDEN                                 uint32   = 0x2
	SUBLANG_SAMI_NORTHERN_FINLAND                                uint32   = 0x3
	SUBLANG_SAMI_LULE_NORWAY                                     uint32   = 0x4
	SUBLANG_SAMI_LULE_SWEDEN                                     uint32   = 0x5
	SUBLANG_SAMI_SOUTHERN_NORWAY                                 uint32   = 0x6
	SUBLANG_SAMI_SOUTHERN_SWEDEN                                 uint32   = 0x7
	SUBLANG_SAMI_SKOLT_FINLAND                                   uint32   = 0x8
	SUBLANG_SAMI_INARI_FINLAND                                   uint32   = 0x9
	SUBLANG_SANSKRIT_INDIA                                       uint32   = 0x1
	SUBLANG_SCOTTISH_GAELIC                                      uint32   = 0x1
	SUBLANG_SERBIAN_BOSNIA_HERZEGOVINA_LATIN                     uint32   = 0x6
	SUBLANG_SERBIAN_BOSNIA_HERZEGOVINA_CYRILLIC                  uint32   = 0x7
	SUBLANG_SERBIAN_MONTENEGRO_LATIN                             uint32   = 0xb
	SUBLANG_SERBIAN_MONTENEGRO_CYRILLIC                          uint32   = 0xc
	SUBLANG_SERBIAN_SERBIA_LATIN                                 uint32   = 0x9
	SUBLANG_SERBIAN_SERBIA_CYRILLIC                              uint32   = 0xa
	SUBLANG_SERBIAN_CROATIA                                      uint32   = 0x1
	SUBLANG_SERBIAN_LATIN                                        uint32   = 0x2
	SUBLANG_SERBIAN_CYRILLIC                                     uint32   = 0x3
	SUBLANG_SINDHI_INDIA                                         uint32   = 0x1
	SUBLANG_SINDHI_PAKISTAN                                      uint32   = 0x2
	SUBLANG_SINDHI_AFGHANISTAN                                   uint32   = 0x2
	SUBLANG_SINHALESE_SRI_LANKA                                  uint32   = 0x1
	SUBLANG_SOTHO_NORTHERN_SOUTH_AFRICA                          uint32   = 0x1
	SUBLANG_SLOVAK_SLOVAKIA                                      uint32   = 0x1
	SUBLANG_SLOVENIAN_SLOVENIA                                   uint32   = 0x1
	SUBLANG_SPANISH                                              uint32   = 0x1
	SUBLANG_SPANISH_MEXICAN                                      uint32   = 0x2
	SUBLANG_SPANISH_MODERN                                       uint32   = 0x3
	SUBLANG_SPANISH_GUATEMALA                                    uint32   = 0x4
	SUBLANG_SPANISH_COSTA_RICA                                   uint32   = 0x5
	SUBLANG_SPANISH_PANAMA                                       uint32   = 0x6
	SUBLANG_SPANISH_DOMINICAN_REPUBLIC                           uint32   = 0x7
	SUBLANG_SPANISH_VENEZUELA                                    uint32   = 0x8
	SUBLANG_SPANISH_COLOMBIA                                     uint32   = 0x9
	SUBLANG_SPANISH_PERU                                         uint32   = 0xa
	SUBLANG_SPANISH_ARGENTINA                                    uint32   = 0xb
	SUBLANG_SPANISH_ECUADOR                                      uint32   = 0xc
	SUBLANG_SPANISH_CHILE                                        uint32   = 0xd
	SUBLANG_SPANISH_URUGUAY                                      uint32   = 0xe
	SUBLANG_SPANISH_PARAGUAY                                     uint32   = 0xf
	SUBLANG_SPANISH_BOLIVIA                                      uint32   = 0x10
	SUBLANG_SPANISH_EL_SALVADOR                                  uint32   = 0x11
	SUBLANG_SPANISH_HONDURAS                                     uint32   = 0x12
	SUBLANG_SPANISH_NICARAGUA                                    uint32   = 0x13
	SUBLANG_SPANISH_PUERTO_RICO                                  uint32   = 0x14
	SUBLANG_SPANISH_US                                           uint32   = 0x15
	SUBLANG_SWAHILI_KENYA                                        uint32   = 0x1
	SUBLANG_SWEDISH                                              uint32   = 0x1
	SUBLANG_SWEDISH_FINLAND                                      uint32   = 0x2
	SUBLANG_SYRIAC_SYRIA                                         uint32   = 0x1
	SUBLANG_TAJIK_TAJIKISTAN                                     uint32   = 0x1
	SUBLANG_TAMAZIGHT_ALGERIA_LATIN                              uint32   = 0x2
	SUBLANG_TAMAZIGHT_MOROCCO_TIFINAGH                           uint32   = 0x4
	SUBLANG_TAMIL_INDIA                                          uint32   = 0x1
	SUBLANG_TAMIL_SRI_LANKA                                      uint32   = 0x2
	SUBLANG_TATAR_RUSSIA                                         uint32   = 0x1
	SUBLANG_TELUGU_INDIA                                         uint32   = 0x1
	SUBLANG_THAI_THAILAND                                        uint32   = 0x1
	SUBLANG_TIBETAN_PRC                                          uint32   = 0x1
	SUBLANG_TIGRIGNA_ERITREA                                     uint32   = 0x2
	SUBLANG_TIGRINYA_ERITREA                                     uint32   = 0x2
	SUBLANG_TIGRINYA_ETHIOPIA                                    uint32   = 0x1
	SUBLANG_TSWANA_BOTSWANA                                      uint32   = 0x2
	SUBLANG_TSWANA_SOUTH_AFRICA                                  uint32   = 0x1
	SUBLANG_TURKISH_TURKEY                                       uint32   = 0x1
	SUBLANG_TURKMEN_TURKMENISTAN                                 uint32   = 0x1
	SUBLANG_UIGHUR_PRC                                           uint32   = 0x1
	SUBLANG_UKRAINIAN_UKRAINE                                    uint32   = 0x1
	SUBLANG_UPPER_SORBIAN_GERMANY                                uint32   = 0x1
	SUBLANG_URDU_PAKISTAN                                        uint32   = 0x1
	SUBLANG_URDU_INDIA                                           uint32   = 0x2
	SUBLANG_UZBEK_LATIN                                          uint32   = 0x1
	SUBLANG_UZBEK_CYRILLIC                                       uint32   = 0x2
	SUBLANG_VALENCIAN_VALENCIA                                   uint32   = 0x2
	SUBLANG_VIETNAMESE_VIETNAM                                   uint32   = 0x1
	SUBLANG_WELSH_UNITED_KINGDOM                                 uint32   = 0x1
	SUBLANG_WOLOF_SENEGAL                                        uint32   = 0x1
	SUBLANG_XHOSA_SOUTH_AFRICA                                   uint32   = 0x1
	SUBLANG_YAKUT_RUSSIA                                         uint32   = 0x1
	SUBLANG_YI_PRC                                               uint32   = 0x1
	SUBLANG_YORUBA_NIGERIA                                       uint32   = 0x1
	SUBLANG_ZULU_SOUTH_AFRICA                                    uint32   = 0x1
	SORT_DEFAULT                                                 uint32   = 0x0
	SORT_INVARIANT_MATH                                          uint32   = 0x1
	SORT_JAPANESE_XJIS                                           uint32   = 0x0
	SORT_JAPANESE_UNICODE                                        uint32   = 0x1
	SORT_JAPANESE_RADICALSTROKE                                  uint32   = 0x4
	SORT_CHINESE_BIG5                                            uint32   = 0x0
	SORT_CHINESE_PRCP                                            uint32   = 0x0
	SORT_CHINESE_UNICODE                                         uint32   = 0x1
	SORT_CHINESE_PRC                                             uint32   = 0x2
	SORT_CHINESE_BOPOMOFO                                        uint32   = 0x3
	SORT_CHINESE_RADICALSTROKE                                   uint32   = 0x4
	SORT_KOREAN_KSC                                              uint32   = 0x0
	SORT_KOREAN_UNICODE                                          uint32   = 0x1
	SORT_GERMAN_PHONE_BOOK                                       uint32   = 0x1
	SORT_HUNGARIAN_DEFAULT                                       uint32   = 0x0
	SORT_HUNGARIAN_TECHNICAL                                     uint32   = 0x1
	SORT_GEORGIAN_TRADITIONAL                                    uint32   = 0x0
	SORT_GEORGIAN_MODERN                                         uint32   = 0x1
	NLS_VALID_LOCALE_MASK                                        uint32   = 0xfffff
	LOCALE_NAME_MAX_LENGTH                                       uint32   = 0x55
	LOCALE_TRANSIENT_KEYBOARD1                                   uint32   = 0x2000
	LOCALE_TRANSIENT_KEYBOARD2                                   uint32   = 0x2400
	LOCALE_TRANSIENT_KEYBOARD3                                   uint32   = 0x2800
	LOCALE_TRANSIENT_KEYBOARD4                                   uint32   = 0x2c00
	MAXIMUM_WAIT_OBJECTS                                         uint32   = 0x40
	MAXIMUM_SUSPEND_COUNT                                        uint32   = 0x7f
	PF_TEMPORAL_LEVEL_1                                          uint32   = 0x1
	PF_TEMPORAL_LEVEL_2                                          uint32   = 0x2
	PF_TEMPORAL_LEVEL_3                                          uint32   = 0x3
	PF_NON_TEMPORAL_LEVEL_ALL                                    uint32   = 0x0
	EXCEPTION_READ_FAULT                                         uint32   = 0x0
	EXCEPTION_WRITE_FAULT                                        uint32   = 0x1
	EXCEPTION_EXECUTE_FAULT                                      uint32   = 0x8
	CONTEXT_AMD64                                                int32    = 1048576
	CONTEXT_KERNEL_DEBUGGER                                      int32    = 67108864
	CONTEXT_EXCEPTION_ACTIVE                                     int32    = 134217728
	CONTEXT_SERVICE_ACTIVE                                       int32    = 268435456
	CONTEXT_EXCEPTION_REQUEST                                    int32    = 1073741824
	CONTEXT_EXCEPTION_REPORTING                                  int32    = -2147483648
	CONTEXT_UNWOUND_TO_CALL                                      uint32   = 0x20000000
	INITIAL_MXCSR                                                uint32   = 0x1f80
	INITIAL_FPCSR                                                uint32   = 0x27f
	RUNTIME_FUNCTION_INDIRECT                                    uint32   = 0x1
	UNW_FLAG_NO_EPILOGUE                                         uint32   = 0x80000000
	UNWIND_CHAIN_LIMIT                                           uint32   = 0x20
	OUT_OF_PROCESS_FUNCTION_TABLE_CALLBACK_EXPORT_NAME           string   = "OutOfProcessFunctionTableCallback"
	CONTEXT_ARM                                                  int32    = 2097152
	INITIAL_CPSR                                                 uint32   = 0x10
	INITIAL_FPSCR                                                uint32   = 0x0
	ARM_MAX_BREAKPOINTS                                          uint32   = 0x8
	ARM_MAX_WATCHPOINTS                                          uint32   = 0x1
	ARM64_PREFETCH_PLD                                           uint32   = 0x0
	ARM64_PREFETCH_PLI                                           uint32   = 0x8
	ARM64_PREFETCH_PST                                           uint32   = 0x10
	ARM64_PREFETCH_L1                                            uint32   = 0x0
	ARM64_PREFETCH_L2                                            uint32   = 0x2
	ARM64_PREFETCH_L3                                            uint32   = 0x4
	ARM64_PREFETCH_KEEP                                          uint32   = 0x0
	ARM64_PREFETCH_STRM                                          uint32   = 0x1
	ARM64_MULT_INTRINSICS_SUPPORTED                              uint32   = 0x1
	CONTEXT_ARM64                                                int32    = 4194304
	CONTEXT_ARM64_UNWOUND_TO_CALL                                uint32   = 0x20000000
	CONTEXT_ARM64_RET_TO_GUEST                                   uint32   = 0x4000000
	CONTEXT_RET_TO_GUEST                                         uint32   = 0x4000000
	ARM64_MAX_BREAKPOINTS                                        uint32   = 0x8
	ARM64_MAX_WATCHPOINTS                                        uint32   = 0x2
	NONVOL_INT_NUMREG_ARM64                                      uint32   = 0xb
	NONVOL_FP_NUMREG_ARM64                                       uint32   = 0x8
	BREAK_DEBUG_BASE                                             uint32   = 0x80000
	ASSERT_BREAKPOINT                                            uint32   = 0x80003
	SIZE_OF_80387_REGISTERS                                      uint32   = 0x50
	CONTEXT_i386                                                 int32    = 65536
	CONTEXT_i486                                                 int32    = 65536
	MAXIMUM_SUPPORTED_EXTENSION                                  uint32   = 0x200
	EXCEPTION_NONCONTINUABLE                                     uint32   = 0x1
	EXCEPTION_UNWINDING                                          uint32   = 0x2
	EXCEPTION_EXIT_UNWIND                                        uint32   = 0x4
	EXCEPTION_STACK_INVALID                                      uint32   = 0x8
	EXCEPTION_NESTED_CALL                                        uint32   = 0x10
	EXCEPTION_TARGET_UNWIND                                      uint32   = 0x20
	EXCEPTION_COLLIDED_UNWIND                                    uint32   = 0x40
	EXCEPTION_SOFTWARE_ORIGINATE                                 uint32   = 0x80
	EXCEPTION_MAXIMUM_PARAMETERS                                 uint32   = 0xf
	DELETE                                                       uint32   = 0x10000
	WRITE_DAC                                                    uint32   = 0x40000
	WRITE_OWNER                                                  uint32   = 0x80000
	ACCESS_SYSTEM_SECURITY                                       uint32   = 0x1000000
	MAXIMUM_ALLOWED                                              uint32   = 0x2000000
	GENERIC_READ                                                 uint32   = 0x80000000
	GENERIC_WRITE                                                uint32   = 0x40000000
	GENERIC_EXECUTE                                              uint32   = 0x20000000
	GENERIC_ALL                                                  uint32   = 0x10000000
	SID_REVISION                                                 uint32   = 0x1
	SID_MAX_SUB_AUTHORITIES                                      uint32   = 0xf
	SID_RECOMMENDED_SUB_AUTHORITIES                              uint32   = 0x1
	SID_HASH_SIZE                                                uint32   = 0x20
	SECURITY_NULL_RID                                            int32    = 0
	SECURITY_WORLD_RID                                           int32    = 0
	SECURITY_LOCAL_RID                                           int32    = 0
	SECURITY_LOCAL_LOGON_RID                                     int32    = 1
	SECURITY_CREATOR_OWNER_RID                                   int32    = 0
	SECURITY_CREATOR_GROUP_RID                                   int32    = 1
	SECURITY_CREATOR_OWNER_SERVER_RID                            int32    = 2
	SECURITY_CREATOR_GROUP_SERVER_RID                            int32    = 3
	SECURITY_CREATOR_OWNER_RIGHTS_RID                            int32    = 4
	SECURITY_DIALUP_RID                                          int32    = 1
	SECURITY_NETWORK_RID                                         int32    = 2
	SECURITY_BATCH_RID                                           int32    = 3
	SECURITY_INTERACTIVE_RID                                     int32    = 4
	SECURITY_LOGON_IDS_RID                                       int32    = 5
	SECURITY_LOGON_IDS_RID_COUNT                                 int32    = 3
	SECURITY_SERVICE_RID                                         int32    = 6
	SECURITY_ANONYMOUS_LOGON_RID                                 int32    = 7
	SECURITY_PROXY_RID                                           int32    = 8
	SECURITY_ENTERPRISE_CONTROLLERS_RID                          int32    = 9
	SECURITY_SERVER_LOGON_RID                                    int32    = 9
	SECURITY_PRINCIPAL_SELF_RID                                  int32    = 10
	SECURITY_AUTHENTICATED_USER_RID                              int32    = 11
	SECURITY_RESTRICTED_CODE_RID                                 int32    = 12
	SECURITY_TERMINAL_SERVER_RID                                 int32    = 13
	SECURITY_REMOTE_LOGON_RID                                    int32    = 14
	SECURITY_THIS_ORGANIZATION_RID                               int32    = 15
	SECURITY_IUSER_RID                                           int32    = 17
	SECURITY_LOCAL_SYSTEM_RID                                    int32    = 18
	SECURITY_LOCAL_SERVICE_RID                                   int32    = 19
	SECURITY_NETWORK_SERVICE_RID                                 int32    = 20
	SECURITY_NT_NON_UNIQUE                                       int32    = 21
	SECURITY_NT_NON_UNIQUE_SUB_AUTH_COUNT                        int32    = 3
	SECURITY_ENTERPRISE_READONLY_CONTROLLERS_RID                 int32    = 22
	SECURITY_BUILTIN_DOMAIN_RID                                  int32    = 32
	SECURITY_WRITE_RESTRICTED_CODE_RID                           int32    = 33
	SECURITY_PACKAGE_BASE_RID                                    int32    = 64
	SECURITY_PACKAGE_RID_COUNT                                   int32    = 2
	SECURITY_PACKAGE_NTLM_RID                                    int32    = 10
	SECURITY_PACKAGE_SCHANNEL_RID                                int32    = 14
	SECURITY_PACKAGE_DIGEST_RID                                  int32    = 21
	SECURITY_CRED_TYPE_BASE_RID                                  int32    = 65
	SECURITY_CRED_TYPE_RID_COUNT                                 int32    = 2
	SECURITY_CRED_TYPE_THIS_ORG_CERT_RID                         int32    = 1
	SECURITY_MIN_BASE_RID                                        int32    = 80
	SECURITY_SERVICE_ID_BASE_RID                                 int32    = 80
	SECURITY_SERVICE_ID_RID_COUNT                                int32    = 6
	SECURITY_RESERVED_ID_BASE_RID                                int32    = 81
	SECURITY_APPPOOL_ID_BASE_RID                                 int32    = 82
	SECURITY_APPPOOL_ID_RID_COUNT                                int32    = 6
	SECURITY_VIRTUALSERVER_ID_BASE_RID                           int32    = 83
	SECURITY_VIRTUALSERVER_ID_RID_COUNT                          int32    = 6
	SECURITY_USERMODEDRIVERHOST_ID_BASE_RID                      int32    = 84
	SECURITY_USERMODEDRIVERHOST_ID_RID_COUNT                     int32    = 6
	SECURITY_CLOUD_INFRASTRUCTURE_SERVICES_ID_BASE_RID           int32    = 85
	SECURITY_CLOUD_INFRASTRUCTURE_SERVICES_ID_RID_COUNT          int32    = 6
	SECURITY_WMIHOST_ID_BASE_RID                                 int32    = 86
	SECURITY_WMIHOST_ID_RID_COUNT                                int32    = 6
	SECURITY_TASK_ID_BASE_RID                                    int32    = 87
	SECURITY_NFS_ID_BASE_RID                                     int32    = 88
	SECURITY_COM_ID_BASE_RID                                     int32    = 89
	SECURITY_WINDOW_MANAGER_BASE_RID                             int32    = 90
	SECURITY_RDV_GFX_BASE_RID                                    int32    = 91
	SECURITY_DASHOST_ID_BASE_RID                                 int32    = 92
	SECURITY_DASHOST_ID_RID_COUNT                                int32    = 6
	SECURITY_USERMANAGER_ID_BASE_RID                             int32    = 93
	SECURITY_USERMANAGER_ID_RID_COUNT                            int32    = 6
	SECURITY_WINRM_ID_BASE_RID                                   int32    = 94
	SECURITY_WINRM_ID_RID_COUNT                                  int32    = 6
	SECURITY_CCG_ID_BASE_RID                                     int32    = 95
	SECURITY_UMFD_BASE_RID                                       int32    = 96
	SECURITY_VIRTUALACCOUNT_ID_RID_COUNT                         int32    = 6
	SECURITY_MAX_BASE_RID                                        int32    = 111
	SECURITY_MAX_ALWAYS_FILTERED                                 int32    = 999
	SECURITY_MIN_NEVER_FILTERED                                  int32    = 1000
	SECURITY_OTHER_ORGANIZATION_RID                              int32    = 1000
	SECURITY_WINDOWSMOBILE_ID_BASE_RID                           int32    = 112
	SECURITY_INSTALLER_GROUP_CAPABILITY_BASE                     uint32   = 0x20
	SECURITY_INSTALLER_GROUP_CAPABILITY_RID_COUNT                uint32   = 0x9
	SECURITY_INSTALLER_CAPABILITY_RID_COUNT                      uint32   = 0xa
	SECURITY_LOCAL_ACCOUNT_RID                                   int32    = 113
	SECURITY_LOCAL_ACCOUNT_AND_ADMIN_RID                         int32    = 114
	DOMAIN_GROUP_RID_AUTHORIZATION_DATA_IS_COMPOUNDED            int32    = 496
	DOMAIN_GROUP_RID_AUTHORIZATION_DATA_CONTAINS_CLAIMS          int32    = 497
	DOMAIN_GROUP_RID_ENTERPRISE_READONLY_DOMAIN_CONTROLLERS      int32    = 498
	FOREST_USER_RID_MAX                                          int32    = 499
	DOMAIN_USER_RID_ADMIN                                        int32    = 500
	DOMAIN_USER_RID_GUEST                                        int32    = 501
	DOMAIN_USER_RID_KRBTGT                                       int32    = 502
	DOMAIN_USER_RID_DEFAULT_ACCOUNT                              int32    = 503
	DOMAIN_USER_RID_WDAG_ACCOUNT                                 int32    = 504
	DOMAIN_USER_RID_MAX                                          int32    = 999
	DOMAIN_GROUP_RID_ADMINS                                      int32    = 512
	DOMAIN_GROUP_RID_USERS                                       int32    = 513
	DOMAIN_GROUP_RID_GUESTS                                      int32    = 514
	DOMAIN_GROUP_RID_COMPUTERS                                   int32    = 515
	DOMAIN_GROUP_RID_CONTROLLERS                                 int32    = 516
	DOMAIN_GROUP_RID_CERT_ADMINS                                 int32    = 517
	DOMAIN_GROUP_RID_SCHEMA_ADMINS                               int32    = 518
	DOMAIN_GROUP_RID_ENTERPRISE_ADMINS                           int32    = 519
	DOMAIN_GROUP_RID_POLICY_ADMINS                               int32    = 520
	DOMAIN_GROUP_RID_READONLY_CONTROLLERS                        int32    = 521
	DOMAIN_GROUP_RID_CLONEABLE_CONTROLLERS                       int32    = 522
	DOMAIN_GROUP_RID_CDC_RESERVED                                int32    = 524
	DOMAIN_GROUP_RID_PROTECTED_USERS                             int32    = 525
	DOMAIN_GROUP_RID_KEY_ADMINS                                  int32    = 526
	DOMAIN_GROUP_RID_ENTERPRISE_KEY_ADMINS                       int32    = 527
	DOMAIN_ALIAS_RID_ADMINS                                      int32    = 544
	DOMAIN_ALIAS_RID_USERS                                       int32    = 545
	DOMAIN_ALIAS_RID_GUESTS                                      int32    = 546
	DOMAIN_ALIAS_RID_POWER_USERS                                 int32    = 547
	DOMAIN_ALIAS_RID_ACCOUNT_OPS                                 int32    = 548
	DOMAIN_ALIAS_RID_SYSTEM_OPS                                  int32    = 549
	DOMAIN_ALIAS_RID_PRINT_OPS                                   int32    = 550
	DOMAIN_ALIAS_RID_BACKUP_OPS                                  int32    = 551
	DOMAIN_ALIAS_RID_REPLICATOR                                  int32    = 552
	DOMAIN_ALIAS_RID_RAS_SERVERS                                 int32    = 553
	DOMAIN_ALIAS_RID_PREW2KCOMPACCESS                            int32    = 554
	DOMAIN_ALIAS_RID_REMOTE_DESKTOP_USERS                        int32    = 555
	DOMAIN_ALIAS_RID_NETWORK_CONFIGURATION_OPS                   int32    = 556
	DOMAIN_ALIAS_RID_INCOMING_FOREST_TRUST_BUILDERS              int32    = 557
	DOMAIN_ALIAS_RID_MONITORING_USERS                            int32    = 558
	DOMAIN_ALIAS_RID_LOGGING_USERS                               int32    = 559
	DOMAIN_ALIAS_RID_AUTHORIZATIONACCESS                         int32    = 560
	DOMAIN_ALIAS_RID_TS_LICENSE_SERVERS                          int32    = 561
	DOMAIN_ALIAS_RID_DCOM_USERS                                  int32    = 562
	DOMAIN_ALIAS_RID_IUSERS                                      int32    = 568
	DOMAIN_ALIAS_RID_CRYPTO_OPERATORS                            int32    = 569
	DOMAIN_ALIAS_RID_CACHEABLE_PRINCIPALS_GROUP                  int32    = 571
	DOMAIN_ALIAS_RID_NON_CACHEABLE_PRINCIPALS_GROUP              int32    = 572
	DOMAIN_ALIAS_RID_EVENT_LOG_READERS_GROUP                     int32    = 573
	DOMAIN_ALIAS_RID_CERTSVC_DCOM_ACCESS_GROUP                   int32    = 574
	DOMAIN_ALIAS_RID_RDS_REMOTE_ACCESS_SERVERS                   int32    = 575
	DOMAIN_ALIAS_RID_RDS_ENDPOINT_SERVERS                        int32    = 576
	DOMAIN_ALIAS_RID_RDS_MANAGEMENT_SERVERS                      int32    = 577
	DOMAIN_ALIAS_RID_HYPER_V_ADMINS                              int32    = 578
	DOMAIN_ALIAS_RID_ACCESS_CONTROL_ASSISTANCE_OPS               int32    = 579
	DOMAIN_ALIAS_RID_REMOTE_MANAGEMENT_USERS                     int32    = 580
	DOMAIN_ALIAS_RID_DEFAULT_ACCOUNT                             int32    = 581
	DOMAIN_ALIAS_RID_STORAGE_REPLICA_ADMINS                      int32    = 582
	DOMAIN_ALIAS_RID_DEVICE_OWNERS                               int32    = 583
	SECURITY_APP_PACKAGE_BASE_RID                                int32    = 2
	SECURITY_BUILTIN_APP_PACKAGE_RID_COUNT                       int32    = 2
	SECURITY_APP_PACKAGE_RID_COUNT                               int32    = 8
	SECURITY_CAPABILITY_BASE_RID                                 int32    = 3
	SECURITY_CAPABILITY_APP_RID                                  uint64   = 0x400
	SECURITY_BUILTIN_CAPABILITY_RID_COUNT                        int32    = 2
	SECURITY_CAPABILITY_RID_COUNT                                int32    = 5
	SECURITY_PARENT_PACKAGE_RID_COUNT                            int32    = 8
	SECURITY_CHILD_PACKAGE_RID_COUNT                             int32    = 12
	SECURITY_BUILTIN_PACKAGE_ANY_PACKAGE                         int32    = 1
	SECURITY_BUILTIN_PACKAGE_ANY_RESTRICTED_PACKAGE              int32    = 2
	SECURITY_CAPABILITY_INTERNET_CLIENT                          int32    = 1
	SECURITY_CAPABILITY_INTERNET_CLIENT_SERVER                   int32    = 2
	SECURITY_CAPABILITY_PRIVATE_NETWORK_CLIENT_SERVER            int32    = 3
	SECURITY_CAPABILITY_PICTURES_LIBRARY                         int32    = 4
	SECURITY_CAPABILITY_VIDEOS_LIBRARY                           int32    = 5
	SECURITY_CAPABILITY_MUSIC_LIBRARY                            int32    = 6
	SECURITY_CAPABILITY_DOCUMENTS_LIBRARY                        int32    = 7
	SECURITY_CAPABILITY_ENTERPRISE_AUTHENTICATION                int32    = 8
	SECURITY_CAPABILITY_SHARED_USER_CERTIFICATES                 int32    = 9
	SECURITY_CAPABILITY_REMOVABLE_STORAGE                        int32    = 10
	SECURITY_CAPABILITY_APPOINTMENTS                             int32    = 11
	SECURITY_CAPABILITY_CONTACTS                                 int32    = 12
	SECURITY_CAPABILITY_INTERNET_EXPLORER                        int32    = 4096
	SECURITY_MANDATORY_UNTRUSTED_RID                             int32    = 0
	SECURITY_MANDATORY_LOW_RID                                   int32    = 4096
	SECURITY_MANDATORY_MEDIUM_RID                                int32    = 8192
	SECURITY_MANDATORY_MEDIUM_PLUS_RID                           uint32   = 0x2100
	SECURITY_MANDATORY_HIGH_RID                                  int32    = 12288
	SECURITY_MANDATORY_SYSTEM_RID                                int32    = 16384
	SECURITY_MANDATORY_PROTECTED_PROCESS_RID                     int32    = 20480
	SECURITY_MANDATORY_MAXIMUM_USER_RID                          int32    = 16384
	SECURITY_AUTHENTICATION_AUTHORITY_RID_COUNT                  int32    = 1
	SECURITY_AUTHENTICATION_AUTHORITY_ASSERTED_RID               int32    = 1
	SECURITY_AUTHENTICATION_SERVICE_ASSERTED_RID                 int32    = 2
	SECURITY_AUTHENTICATION_FRESH_KEY_AUTH_RID                   int32    = 3
	SECURITY_AUTHENTICATION_KEY_TRUST_RID                        int32    = 4
	SECURITY_AUTHENTICATION_KEY_PROPERTY_MFA_RID                 int32    = 5
	SECURITY_AUTHENTICATION_KEY_PROPERTY_ATTESTATION_RID         int32    = 6
	SECURITY_PROCESS_TRUST_AUTHORITY_RID_COUNT                   int32    = 2
	SECURITY_PROCESS_PROTECTION_TYPE_FULL_RID                    int32    = 1024
	SECURITY_PROCESS_PROTECTION_TYPE_LITE_RID                    int32    = 512
	SECURITY_PROCESS_PROTECTION_TYPE_NONE_RID                    int32    = 0
	SECURITY_PROCESS_PROTECTION_LEVEL_WINTCB_RID                 int32    = 8192
	SECURITY_PROCESS_PROTECTION_LEVEL_WINDOWS_RID                int32    = 4096
	SECURITY_PROCESS_PROTECTION_LEVEL_APP_RID                    int32    = 2048
	SECURITY_PROCESS_PROTECTION_LEVEL_ANTIMALWARE_RID            int32    = 1536
	SECURITY_PROCESS_PROTECTION_LEVEL_AUTHENTICODE_RID           int32    = 1024
	SECURITY_PROCESS_PROTECTION_LEVEL_NONE_RID                   int32    = 0
	SECURITY_TRUSTED_INSTALLER_RID1                              uint32   = 0x38fb89b5
	SECURITY_TRUSTED_INSTALLER_RID2                              uint32   = 0xcbc28419
	SECURITY_TRUSTED_INSTALLER_RID3                              uint32   = 0x6d236c5c
	SECURITY_TRUSTED_INSTALLER_RID4                              uint32   = 0x6e770057
	SECURITY_TRUSTED_INSTALLER_RID5                              uint32   = 0x876402c0
	SE_GROUP_MANDATORY                                           int32    = 1
	SE_GROUP_ENABLED_BY_DEFAULT                                  int32    = 2
	SE_GROUP_ENABLED                                             int32    = 4
	SE_GROUP_OWNER                                               int32    = 8
	SE_GROUP_USE_FOR_DENY_ONLY                                   int32    = 16
	SE_GROUP_INTEGRITY                                           int32    = 32
	SE_GROUP_INTEGRITY_ENABLED                                   int32    = 64
	SE_GROUP_LOGON_ID                                            int32    = -1073741824
	SE_GROUP_RESOURCE                                            int32    = 536870912
	ACL_REVISION1                                                uint32   = 0x1
	ACL_REVISION2                                                uint32   = 0x2
	ACL_REVISION3                                                uint32   = 0x3
	ACL_REVISION4                                                uint32   = 0x4
	MAX_ACL_REVISION                                             uint32   = 0x4
	ACCESS_MIN_MS_ACE_TYPE                                       uint32   = 0x0
	ACCESS_ALLOWED_ACE_TYPE                                      uint32   = 0x0
	ACCESS_DENIED_ACE_TYPE                                       uint32   = 0x1
	SYSTEM_AUDIT_ACE_TYPE                                        uint32   = 0x2
	SYSTEM_ALARM_ACE_TYPE                                        uint32   = 0x3
	ACCESS_MAX_MS_V2_ACE_TYPE                                    uint32   = 0x3
	ACCESS_ALLOWED_COMPOUND_ACE_TYPE                             uint32   = 0x4
	ACCESS_MAX_MS_V3_ACE_TYPE                                    uint32   = 0x4
	ACCESS_MIN_MS_OBJECT_ACE_TYPE                                uint32   = 0x5
	ACCESS_ALLOWED_OBJECT_ACE_TYPE                               uint32   = 0x5
	ACCESS_DENIED_OBJECT_ACE_TYPE                                uint32   = 0x6
	SYSTEM_AUDIT_OBJECT_ACE_TYPE                                 uint32   = 0x7
	SYSTEM_ALARM_OBJECT_ACE_TYPE                                 uint32   = 0x8
	ACCESS_MAX_MS_OBJECT_ACE_TYPE                                uint32   = 0x8
	ACCESS_MAX_MS_V4_ACE_TYPE                                    uint32   = 0x8
	ACCESS_MAX_MS_ACE_TYPE                                       uint32   = 0x8
	ACCESS_ALLOWED_CALLBACK_ACE_TYPE                             uint32   = 0x9
	ACCESS_DENIED_CALLBACK_ACE_TYPE                              uint32   = 0xa
	ACCESS_ALLOWED_CALLBACK_OBJECT_ACE_TYPE                      uint32   = 0xb
	ACCESS_DENIED_CALLBACK_OBJECT_ACE_TYPE                       uint32   = 0xc
	SYSTEM_AUDIT_CALLBACK_ACE_TYPE                               uint32   = 0xd
	SYSTEM_ALARM_CALLBACK_ACE_TYPE                               uint32   = 0xe
	SYSTEM_AUDIT_CALLBACK_OBJECT_ACE_TYPE                        uint32   = 0xf
	SYSTEM_ALARM_CALLBACK_OBJECT_ACE_TYPE                        uint32   = 0x10
	SYSTEM_MANDATORY_LABEL_ACE_TYPE                              uint32   = 0x11
	SYSTEM_RESOURCE_ATTRIBUTE_ACE_TYPE                           uint32   = 0x12
	SYSTEM_SCOPED_POLICY_ID_ACE_TYPE                             uint32   = 0x13
	SYSTEM_PROCESS_TRUST_LABEL_ACE_TYPE                          uint32   = 0x14
	SYSTEM_ACCESS_FILTER_ACE_TYPE                                uint32   = 0x15
	ACCESS_MAX_MS_V5_ACE_TYPE                                    uint32   = 0x15
	VALID_INHERIT_FLAGS                                          uint32   = 0x1f
	CRITICAL_ACE_FLAG                                            uint32   = 0x20
	TRUST_PROTECTED_FILTER_ACE_FLAG                              uint32   = 0x40
	SYSTEM_MANDATORY_LABEL_NO_WRITE_UP                           uint32   = 0x1
	SYSTEM_MANDATORY_LABEL_NO_READ_UP                            uint32   = 0x2
	SYSTEM_MANDATORY_LABEL_NO_EXECUTE_UP                         uint32   = 0x4
	SYSTEM_PROCESS_TRUST_LABEL_VALID_MASK                        uint32   = 0xffffff
	SYSTEM_PROCESS_TRUST_NOCONSTRAINT_MASK                       uint32   = 0xffffffff
	SYSTEM_ACCESS_FILTER_VALID_MASK                              uint32   = 0xffffff
	SYSTEM_ACCESS_FILTER_NOCONSTRAINT_MASK                       uint32   = 0xffffffff
	SECURITY_DESCRIPTOR_REVISION                                 uint32   = 0x1
	SECURITY_DESCRIPTOR_REVISION1                                uint32   = 0x1
	SE_OWNER_DEFAULTED                                           uint32   = 0x1
	SE_GROUP_DEFAULTED                                           uint32   = 0x2
	SE_DACL_PRESENT                                              uint32   = 0x4
	SE_DACL_DEFAULTED                                            uint32   = 0x8
	SE_SACL_PRESENT                                              uint32   = 0x10
	SE_SACL_DEFAULTED                                            uint32   = 0x20
	SE_DACL_AUTO_INHERIT_REQ                                     uint32   = 0x100
	SE_SACL_AUTO_INHERIT_REQ                                     uint32   = 0x200
	SE_DACL_AUTO_INHERITED                                       uint32   = 0x400
	SE_SACL_AUTO_INHERITED                                       uint32   = 0x800
	SE_DACL_PROTECTED                                            uint32   = 0x1000
	SE_SACL_PROTECTED                                            uint32   = 0x2000
	SE_RM_CONTROL_VALID                                          uint32   = 0x4000
	SE_SELF_RELATIVE                                             uint32   = 0x8000
	ACCESS_OBJECT_GUID                                           uint32   = 0x0
	ACCESS_PROPERTY_SET_GUID                                     uint32   = 0x1
	ACCESS_PROPERTY_GUID                                         uint32   = 0x2
	ACCESS_MAX_LEVEL                                             uint32   = 0x4
	AUDIT_ALLOW_NO_PRIVILEGE                                     uint32   = 0x1
	ACCESS_DS_SOURCE_A                                           string   = "DS"
	ACCESS_DS_SOURCE_W                                           string   = "DS"
	ACCESS_DS_OBJECT_TYPE_NAME_A                                 string   = "Directory Service Object"
	ACCESS_DS_OBJECT_TYPE_NAME_W                                 string   = "Directory Service Object"
	PRIVILEGE_SET_ALL_NECESSARY                                  uint32   = 0x1
	ACCESS_REASON_TYPE_MASK                                      uint32   = 0xff0000
	ACCESS_REASON_DATA_MASK                                      uint32   = 0xffff
	ACCESS_REASON_STAGING_MASK                                   uint32   = 0x80000000
	ACCESS_REASON_EXDATA_MASK                                    uint32   = 0x7f000000
	SE_SECURITY_DESCRIPTOR_FLAG_NO_OWNER_ACE                     uint32   = 0x1
	SE_SECURITY_DESCRIPTOR_FLAG_NO_LABEL_ACE                     uint32   = 0x2
	SE_SECURITY_DESCRIPTOR_FLAG_NO_ACCESS_FILTER_ACE             uint32   = 0x4
	SE_SECURITY_DESCRIPTOR_VALID_FLAGS                           uint32   = 0x7
	SE_ACCESS_CHECK_FLAG_NO_LEARNING_MODE_LOGGING                uint32   = 0x8
	SE_ACCESS_CHECK_VALID_FLAGS                                  uint32   = 0x8
	SE_CREATE_TOKEN_NAME                                         string   = "SeCreateTokenPrivilege"
	SE_ASSIGNPRIMARYTOKEN_NAME                                   string   = "SeAssignPrimaryTokenPrivilege"
	SE_LOCK_MEMORY_NAME                                          string   = "SeLockMemoryPrivilege"
	SE_INCREASE_QUOTA_NAME                                       string   = "SeIncreaseQuotaPrivilege"
	SE_UNSOLICITED_INPUT_NAME                                    string   = "SeUnsolicitedInputPrivilege"
	SE_MACHINE_ACCOUNT_NAME                                      string   = "SeMachineAccountPrivilege"
	SE_TCB_NAME                                                  string   = "SeTcbPrivilege"
	SE_SECURITY_NAME                                             string   = "SeSecurityPrivilege"
	SE_TAKE_OWNERSHIP_NAME                                       string   = "SeTakeOwnershipPrivilege"
	SE_LOAD_DRIVER_NAME                                          string   = "SeLoadDriverPrivilege"
	SE_SYSTEM_PROFILE_NAME                                       string   = "SeSystemProfilePrivilege"
	SE_SYSTEMTIME_NAME                                           string   = "SeSystemtimePrivilege"
	SE_PROF_SINGLE_PROCESS_NAME                                  string   = "SeProfileSingleProcessPrivilege"
	SE_INC_BASE_PRIORITY_NAME                                    string   = "SeIncreaseBasePriorityPrivilege"
	SE_CREATE_PAGEFILE_NAME                                      string   = "SeCreatePagefilePrivilege"
	SE_CREATE_PERMANENT_NAME                                     string   = "SeCreatePermanentPrivilege"
	SE_BACKUP_NAME                                               string   = "SeBackupPrivilege"
	SE_RESTORE_NAME                                              string   = "SeRestorePrivilege"
	SE_SHUTDOWN_NAME                                             string   = "SeShutdownPrivilege"
	SE_DEBUG_NAME                                                string   = "SeDebugPrivilege"
	SE_AUDIT_NAME                                                string   = "SeAuditPrivilege"
	SE_SYSTEM_ENVIRONMENT_NAME                                   string   = "SeSystemEnvironmentPrivilege"
	SE_CHANGE_NOTIFY_NAME                                        string   = "SeChangeNotifyPrivilege"
	SE_REMOTE_SHUTDOWN_NAME                                      string   = "SeRemoteShutdownPrivilege"
	SE_UNDOCK_NAME                                               string   = "SeUndockPrivilege"
	SE_SYNC_AGENT_NAME                                           string   = "SeSyncAgentPrivilege"
	SE_ENABLE_DELEGATION_NAME                                    string   = "SeEnableDelegationPrivilege"
	SE_MANAGE_VOLUME_NAME                                        string   = "SeManageVolumePrivilege"
	SE_IMPERSONATE_NAME                                          string   = "SeImpersonatePrivilege"
	SE_CREATE_GLOBAL_NAME                                        string   = "SeCreateGlobalPrivilege"
	SE_TRUSTED_CREDMAN_ACCESS_NAME                               string   = "SeTrustedCredManAccessPrivilege"
	SE_RELABEL_NAME                                              string   = "SeRelabelPrivilege"
	SE_INC_WORKING_SET_NAME                                      string   = "SeIncreaseWorkingSetPrivilege"
	SE_TIME_ZONE_NAME                                            string   = "SeTimeZonePrivilege"
	SE_CREATE_SYMBOLIC_LINK_NAME                                 string   = "SeCreateSymbolicLinkPrivilege"
	SE_DELEGATE_SESSION_USER_IMPERSONATE_NAME                    string   = "SeDelegateSessionUserImpersonatePrivilege"
	SE_ACTIVATE_AS_USER_CAPABILITY                               string   = "activateAsUser"
	SE_CONSTRAINED_IMPERSONATION_CAPABILITY                      string   = "constrainedImpersonation"
	SE_SESSION_IMPERSONATION_CAPABILITY                          string   = "sessionImpersonation"
	SE_MUMA_CAPABILITY                                           string   = "muma"
	SE_DEVELOPMENT_MODE_NETWORK_CAPABILITY                       string   = "developmentModeNetwork"
	SE_PERMISSIVE_LEARNING_MODE_CAPABILITY                       string   = "permissiveLearningMode"
	POLICY_AUDIT_SUBCATEGORY_COUNT                               uint32   = 0x3b
	TOKEN_SOURCE_LENGTH                                          uint32   = 0x8
	CLAIM_SECURITY_ATTRIBUTE_TYPE_INVALID                        uint32   = 0x0
	CLAIM_SECURITY_ATTRIBUTE_CUSTOM_FLAGS                        uint32   = 0xffff0000
	CLAIM_SECURITY_ATTRIBUTES_INFORMATION_VERSION_V1             uint32   = 0x1
	CLAIM_SECURITY_ATTRIBUTES_INFORMATION_VERSION                uint32   = 0x1
	PROCESS_TRUST_LABEL_SECURITY_INFORMATION                     int32    = 128
	ACCESS_FILTER_SECURITY_INFORMATION                           int32    = 256
	SE_SIGNING_LEVEL_UNCHECKED                                   uint32   = 0x0
	SE_SIGNING_LEVEL_UNSIGNED                                    uint32   = 0x1
	SE_SIGNING_LEVEL_ENTERPRISE                                  uint32   = 0x2
	SE_SIGNING_LEVEL_CUSTOM_1                                    uint32   = 0x3
	SE_SIGNING_LEVEL_DEVELOPER                                   uint32   = 0x3
	SE_SIGNING_LEVEL_AUTHENTICODE                                uint32   = 0x4
	SE_SIGNING_LEVEL_CUSTOM_2                                    uint32   = 0x5
	SE_SIGNING_LEVEL_STORE                                       uint32   = 0x6
	SE_SIGNING_LEVEL_CUSTOM_3                                    uint32   = 0x7
	SE_SIGNING_LEVEL_ANTIMALWARE                                 uint32   = 0x7
	SE_SIGNING_LEVEL_MICROSOFT                                   uint32   = 0x8
	SE_SIGNING_LEVEL_CUSTOM_4                                    uint32   = 0x9
	SE_SIGNING_LEVEL_CUSTOM_5                                    uint32   = 0xa
	SE_SIGNING_LEVEL_DYNAMIC_CODEGEN                             uint32   = 0xb
	SE_SIGNING_LEVEL_WINDOWS                                     uint32   = 0xc
	SE_SIGNING_LEVEL_CUSTOM_7                                    uint32   = 0xd
	SE_SIGNING_LEVEL_WINDOWS_TCB                                 uint32   = 0xe
	SE_SIGNING_LEVEL_CUSTOM_6                                    uint32   = 0xf
	SE_LEARNING_MODE_FLAG_PERMISSIVE                             uint32   = 0x1
	JOB_OBJECT_ASSIGN_PROCESS                                    uint32   = 0x1
	JOB_OBJECT_SET_ATTRIBUTES                                    uint32   = 0x2
	JOB_OBJECT_QUERY                                             uint32   = 0x4
	JOB_OBJECT_TERMINATE                                         uint32   = 0x8
	JOB_OBJECT_SET_SECURITY_ATTRIBUTES                           uint32   = 0x10
	JOB_OBJECT_IMPERSONATE                                       uint32   = 0x20
	FLS_MAXIMUM_AVAILABLE                                        uint32   = 0xff0
	TLS_MINIMUM_AVAILABLE                                        uint32   = 0x40
	THREAD_DYNAMIC_CODE_ALLOW                                    uint32   = 0x1
	THREAD_BASE_PRIORITY_LOWRT                                   uint32   = 0xf
	THREAD_BASE_PRIORITY_MAX                                     uint32   = 0x2
	THREAD_BASE_PRIORITY_MIN                                     int32    = -2
	THREAD_BASE_PRIORITY_IDLE                                    int32    = -15
	COMPONENT_KTM                                                uint32   = 0x1
	COMPONENT_VALID_FLAGS                                        uint32   = 0x1
	MEMORY_PRIORITY_LOWEST                                       uint32   = 0x0
	DYNAMIC_EH_CONTINUATION_TARGET_ADD                           uint32   = 0x1
	DYNAMIC_EH_CONTINUATION_TARGET_PROCESSED                     uint32   = 0x2
	DYNAMIC_ENFORCED_ADDRESS_RANGE_ADD                           uint32   = 0x1
	DYNAMIC_ENFORCED_ADDRESS_RANGE_PROCESSED                     uint32   = 0x2
	QUOTA_LIMITS_HARDWS_MIN_ENABLE                               uint32   = 0x1
	QUOTA_LIMITS_HARDWS_MIN_DISABLE                              uint32   = 0x2
	QUOTA_LIMITS_HARDWS_MAX_ENABLE                               uint32   = 0x4
	QUOTA_LIMITS_HARDWS_MAX_DISABLE                              uint32   = 0x8
	QUOTA_LIMITS_USE_DEFAULT_LIMITS                              uint32   = 0x10
	MAX_HW_COUNTERS                                              uint32   = 0x10
	THREAD_PROFILING_FLAG_DISPATCH                               uint32   = 0x1
	JOB_OBJECT_NET_RATE_CONTROL_MAX_DSCP_TAG                     uint32   = 0x40
	JOB_OBJECT_MSG_END_OF_JOB_TIME                               uint32   = 0x1
	JOB_OBJECT_MSG_END_OF_PROCESS_TIME                           uint32   = 0x2
	JOB_OBJECT_MSG_ACTIVE_PROCESS_LIMIT                          uint32   = 0x3
	JOB_OBJECT_MSG_ACTIVE_PROCESS_ZERO                           uint32   = 0x4
	JOB_OBJECT_MSG_NEW_PROCESS                                   uint32   = 0x6
	JOB_OBJECT_MSG_EXIT_PROCESS                                  uint32   = 0x7
	JOB_OBJECT_MSG_ABNORMAL_EXIT_PROCESS                         uint32   = 0x8
	JOB_OBJECT_MSG_PROCESS_MEMORY_LIMIT                          uint32   = 0x9
	JOB_OBJECT_MSG_JOB_MEMORY_LIMIT                              uint32   = 0xa
	JOB_OBJECT_MSG_NOTIFICATION_LIMIT                            uint32   = 0xb
	JOB_OBJECT_MSG_JOB_CYCLE_TIME_LIMIT                          uint32   = 0xc
	JOB_OBJECT_MSG_SILO_TERMINATED                               uint32   = 0xd
	JOB_OBJECT_MSG_MINIMUM                                       uint32   = 0x1
	JOB_OBJECT_MSG_MAXIMUM                                       uint32   = 0xd
	JOB_OBJECT_UILIMIT_ALL                                       uint32   = 0xff
	JOB_OBJECT_UI_VALID_FLAGS                                    uint32   = 0xff
	MEMORY_PARTITION_QUERY_ACCESS                                uint32   = 0x1
	MEMORY_PARTITION_MODIFY_ACCESS                               uint32   = 0x2
	EVENT_MODIFY_STATE                                           uint32   = 0x2
	MUTANT_QUERY_STATE                                           uint32   = 0x1
	SEMAPHORE_MODIFY_STATE                                       uint32   = 0x2
	TIMER_QUERY_STATE                                            uint32   = 0x1
	TIMER_MODIFY_STATE                                           uint32   = 0x2
	TIME_ZONE_ID_UNKNOWN                                         uint32   = 0x0
	TIME_ZONE_ID_STANDARD                                        uint32   = 0x1
	TIME_ZONE_ID_DAYLIGHT                                        uint32   = 0x2
	LTP_PC_SMT                                                   uint32   = 0x1
	CACHE_FULLY_ASSOCIATIVE                                      uint32   = 0xff
	PROCESSOR_INTEL_386                                          uint32   = 0x182
	PROCESSOR_INTEL_486                                          uint32   = 0x1e6
	PROCESSOR_INTEL_PENTIUM                                      uint32   = 0x24a
	PROCESSOR_INTEL_IA64                                         uint32   = 0x898
	PROCESSOR_AMD_X8664                                          uint32   = 0x21d8
	PROCESSOR_MIPS_R4000                                         uint32   = 0xfa0
	PROCESSOR_ALPHA_21064                                        uint32   = 0x5248
	PROCESSOR_PPC_601                                            uint32   = 0x259
	PROCESSOR_PPC_603                                            uint32   = 0x25b
	PROCESSOR_PPC_604                                            uint32   = 0x25c
	PROCESSOR_PPC_620                                            uint32   = 0x26c
	PROCESSOR_HITACHI_SH3                                        uint32   = 0x2713
	PROCESSOR_HITACHI_SH3E                                       uint32   = 0x2714
	PROCESSOR_HITACHI_SH4                                        uint32   = 0x2715
	PROCESSOR_MOTOROLA_821                                       uint32   = 0x335
	PROCESSOR_SHx_SH3                                            uint32   = 0x67
	PROCESSOR_SHx_SH4                                            uint32   = 0x68
	PROCESSOR_STRONGARM                                          uint32   = 0xa11
	PROCESSOR_ARM720                                             uint32   = 0x720
	PROCESSOR_ARM820                                             uint32   = 0x820
	PROCESSOR_ARM920                                             uint32   = 0x920
	PROCESSOR_ARM_7TDMI                                          uint32   = 0x11171
	PROCESSOR_OPTIL                                              uint32   = 0x494f
	PROCESSOR_ARCHITECTURE_MIPS                                  uint32   = 0x1
	PROCESSOR_ARCHITECTURE_ALPHA                                 uint32   = 0x2
	PROCESSOR_ARCHITECTURE_PPC                                   uint32   = 0x3
	PROCESSOR_ARCHITECTURE_SHX                                   uint32   = 0x4
	PROCESSOR_ARCHITECTURE_ALPHA64                               uint32   = 0x7
	PROCESSOR_ARCHITECTURE_MSIL                                  uint32   = 0x8
	PROCESSOR_ARCHITECTURE_IA32_ON_WIN64                         uint32   = 0xa
	PROCESSOR_ARCHITECTURE_NEUTRAL                               uint32   = 0xb
	PROCESSOR_ARCHITECTURE_ARM64                                 uint32   = 0xc
	PROCESSOR_ARCHITECTURE_ARM32_ON_WIN64                        uint32   = 0xd
	PROCESSOR_ARCHITECTURE_IA32_ON_ARM64                         uint32   = 0xe
	PF_PPC_MOVEMEM_64BIT_OK                                      uint32   = 0x4
	PF_ALPHA_BYTE_INSTRUCTIONS                                   uint32   = 0x5
	PF_SSE_DAZ_MODE_AVAILABLE                                    uint32   = 0xb
	PF_ARM_NEON_INSTRUCTIONS_AVAILABLE                           uint32   = 0x13
	PF_RDRAND_INSTRUCTION_AVAILABLE                              uint32   = 0x1c
	PF_RDTSCP_INSTRUCTION_AVAILABLE                              uint32   = 0x20
	PF_RDPID_INSTRUCTION_AVAILABLE                               uint32   = 0x21
	PF_MONITORX_INSTRUCTION_AVAILABLE                            uint32   = 0x23
	PF_SSSE3_INSTRUCTIONS_AVAILABLE                              uint32   = 0x24
	PF_SSE4_1_INSTRUCTIONS_AVAILABLE                             uint32   = 0x25
	PF_SSE4_2_INSTRUCTIONS_AVAILABLE                             uint32   = 0x26
	PF_AVX_INSTRUCTIONS_AVAILABLE                                uint32   = 0x27
	PF_AVX2_INSTRUCTIONS_AVAILABLE                               uint32   = 0x28
	PF_AVX512F_INSTRUCTIONS_AVAILABLE                            uint32   = 0x29
	PF_ERMS_AVAILABLE                                            uint32   = 0x2a
	PF_ARM_V82_DP_INSTRUCTIONS_AVAILABLE                         uint32   = 0x2b
	PF_ARM_V83_JSCVT_INSTRUCTIONS_AVAILABLE                      uint32   = 0x2c
	XSTATE_LEGACY_FLOATING_POINT                                 uint32   = 0x0
	XSTATE_LEGACY_SSE                                            uint32   = 0x1
	XSTATE_GSSE                                                  uint32   = 0x2
	XSTATE_AVX                                                   uint32   = 0x2
	XSTATE_MPX_BNDREGS                                           uint32   = 0x3
	XSTATE_MPX_BNDCSR                                            uint32   = 0x4
	XSTATE_AVX512_KMASK                                          uint32   = 0x5
	XSTATE_AVX512_ZMM_H                                          uint32   = 0x6
	XSTATE_AVX512_ZMM                                            uint32   = 0x7
	XSTATE_IPT                                                   uint32   = 0x8
	XSTATE_PASID                                                 uint32   = 0xa
	XSTATE_CET_U                                                 uint32   = 0xb
	XSTATE_CET_S                                                 uint32   = 0xc
	XSTATE_AMX_TILE_CONFIG                                       uint32   = 0x11
	XSTATE_AMX_TILE_DATA                                         uint32   = 0x12
	XSTATE_LWP                                                   uint32   = 0x3e
	MAXIMUM_XSTATE_FEATURES                                      uint32   = 0x40
	XSTATE_COMPACTION_ENABLE                                     uint32   = 0x3f
	XSTATE_ALIGN_BIT                                             uint32   = 0x1
	XSTATE_XFD_BIT                                               uint32   = 0x2
	XSTATE_CONTROLFLAG_XSAVEOPT_MASK                             uint32   = 0x1
	XSTATE_CONTROLFLAG_XSAVEC_MASK                               uint32   = 0x2
	XSTATE_CONTROLFLAG_XFD_MASK                                  uint32   = 0x4
	CFG_CALL_TARGET_VALID                                        uint32   = 0x1
	CFG_CALL_TARGET_PROCESSED                                    uint32   = 0x2
	CFG_CALL_TARGET_CONVERT_EXPORT_SUPPRESSED_TO_VALID           uint32   = 0x4
	CFG_CALL_TARGET_VALID_XFG                                    uint32   = 0x8
	CFG_CALL_TARGET_CONVERT_XFG_TO_CFG                           uint32   = 0x10
	SESSION_QUERY_ACCESS                                         uint32   = 0x1
	SESSION_MODIFY_ACCESS                                        uint32   = 0x2
	MEM_TOP_DOWN                                                 uint32   = 0x100000
	MEM_WRITE_WATCH                                              uint32   = 0x200000
	MEM_PHYSICAL                                                 uint32   = 0x400000
	MEM_ROTATE                                                   uint32   = 0x800000
	MEM_DIFFERENT_IMAGE_BASE_OK                                  uint32   = 0x800000
	MEM_4MB_PAGES                                                uint32   = 0x80000000
	MEM_COALESCE_PLACEHOLDERS                                    uint32   = 0x1
	MEM_EXTENDED_PARAMETER_GRAPHICS                              uint32   = 0x1
	MEM_EXTENDED_PARAMETER_NONPAGED                              uint32   = 0x2
	MEM_EXTENDED_PARAMETER_ZERO_PAGES_OPTIONAL                   uint32   = 0x4
	MEM_EXTENDED_PARAMETER_NONPAGED_LARGE                        uint32   = 0x8
	MEM_EXTENDED_PARAMETER_NONPAGED_HUGE                         uint32   = 0x10
	MEM_EXTENDED_PARAMETER_SOFT_FAULT_PAGES                      uint32   = 0x20
	MEM_EXTENDED_PARAMETER_EC_CODE                               uint32   = 0x40
	MEM_EXTENDED_PARAMETER_TYPE_BITS                             uint32   = 0x8
	SEC_HUGE_PAGES                                               uint32   = 0x20000
	WRITE_WATCH_FLAG_RESET                                       uint32   = 0x1
	ENCLAVE_TYPE_SGX                                             uint32   = 0x1
	ENCLAVE_TYPE_SGX2                                            uint32   = 0x2
	ENCLAVE_TYPE_VBS                                             uint32   = 0x10
	ENCLAVE_VBS_FLAG_DEBUG                                       uint32   = 0x1
	ENCLAVE_TYPE_VBS_BASIC                                       uint32   = 0x11
	VBS_BASIC_PAGE_MEASURED_DATA                                 uint32   = 0x1
	VBS_BASIC_PAGE_UNMEASURED_DATA                               uint32   = 0x2
	VBS_BASIC_PAGE_ZERO_FILL                                     uint32   = 0x3
	VBS_BASIC_PAGE_THREAD_DESCRIPTOR                             uint32   = 0x4
	VBS_BASIC_PAGE_SYSTEM_CALL                                   uint32   = 0x5
	DEDICATED_MEMORY_CACHE_ELIGIBLE                              uint32   = 0x1
	TREE_CONNECT_ATTRIBUTE_PRIVACY                               uint32   = 0x4000
	TREE_CONNECT_ATTRIBUTE_INTEGRITY                             uint32   = 0x8000
	TREE_CONNECT_ATTRIBUTE_GLOBAL                                uint32   = 0x4
	TREE_CONNECT_ATTRIBUTE_PINNED                                uint32   = 0x2
	FILE_ATTRIBUTE_STRICTLY_SEQUENTIAL                           uint32   = 0x20000000
	MAILSLOT_NO_MESSAGE                                          uint32   = 0xffffffff
	MAILSLOT_WAIT_FOREVER                                        uint32   = 0xffffffff
	FILE_CASE_SENSITIVE_SEARCH                                   uint32   = 0x1
	FILE_CASE_PRESERVED_NAMES                                    uint32   = 0x2
	FILE_UNICODE_ON_DISK                                         uint32   = 0x4
	FILE_PERSISTENT_ACLS                                         uint32   = 0x8
	FILE_FILE_COMPRESSION                                        uint32   = 0x10
	FILE_VOLUME_QUOTAS                                           uint32   = 0x20
	FILE_SUPPORTS_SPARSE_FILES                                   uint32   = 0x40
	FILE_SUPPORTS_REPARSE_POINTS                                 uint32   = 0x80
	FILE_SUPPORTS_REMOTE_STORAGE                                 uint32   = 0x100
	FILE_RETURNS_CLEANUP_RESULT_INFO                             uint32   = 0x200
	FILE_SUPPORTS_POSIX_UNLINK_RENAME                            uint32   = 0x400
	FILE_SUPPORTS_BYPASS_IO                                      uint32   = 0x800
	FILE_VOLUME_IS_COMPRESSED                                    uint32   = 0x8000
	FILE_SUPPORTS_OBJECT_IDS                                     uint32   = 0x10000
	FILE_SUPPORTS_ENCRYPTION                                     uint32   = 0x20000
	FILE_NAMED_STREAMS                                           uint32   = 0x40000
	FILE_READ_ONLY_VOLUME                                        uint32   = 0x80000
	FILE_SEQUENTIAL_WRITE_ONCE                                   uint32   = 0x100000
	FILE_SUPPORTS_TRANSACTIONS                                   uint32   = 0x200000
	FILE_SUPPORTS_HARD_LINKS                                     uint32   = 0x400000
	FILE_SUPPORTS_EXTENDED_ATTRIBUTES                            uint32   = 0x800000
	FILE_SUPPORTS_OPEN_BY_FILE_ID                                uint32   = 0x1000000
	FILE_SUPPORTS_USN_JOURNAL                                    uint32   = 0x2000000
	FILE_SUPPORTS_INTEGRITY_STREAMS                              uint32   = 0x4000000
	FILE_SUPPORTS_BLOCK_REFCOUNTING                              uint32   = 0x8000000
	FILE_SUPPORTS_SPARSE_VDL                                     uint32   = 0x10000000
	FILE_DAX_VOLUME                                              uint32   = 0x20000000
	FILE_SUPPORTS_GHOSTING                                       uint32   = 0x40000000
	FILE_CS_FLAG_CASE_SENSITIVE_DIR                              uint32   = 0x1
	FLUSH_FLAGS_FILE_DATA_ONLY                                   uint32   = 0x1
	FLUSH_FLAGS_NO_SYNC                                          uint32   = 0x2
	FLUSH_FLAGS_FILE_DATA_SYNC_ONLY                              uint32   = 0x4
	IO_REPARSE_TAG_RESERVED_ZERO                                 uint32   = 0x0
	IO_REPARSE_TAG_RESERVED_ONE                                  uint32   = 0x1
	IO_REPARSE_TAG_RESERVED_TWO                                  uint32   = 0x2
	IO_REPARSE_TAG_RESERVED_RANGE                                uint32   = 0x2
	IO_REPARSE_TAG_MOUNT_POINT                                   uint32   = 0xa0000003
	IO_REPARSE_TAG_HSM                                           uint32   = 0xc0000004
	IO_REPARSE_TAG_HSM2                                          uint32   = 0x80000006
	IO_REPARSE_TAG_SIS                                           uint32   = 0x80000007
	IO_REPARSE_TAG_WIM                                           uint32   = 0x80000008
	IO_REPARSE_TAG_CSV                                           uint32   = 0x80000009
	IO_REPARSE_TAG_DFS                                           uint32   = 0x8000000a
	IO_REPARSE_TAG_SYMLINK                                       uint32   = 0xa000000c
	IO_REPARSE_TAG_DFSR                                          uint32   = 0x80000012
	IO_REPARSE_TAG_DEDUP                                         uint32   = 0x80000013
	IO_REPARSE_TAG_NFS                                           uint32   = 0x80000014
	IO_REPARSE_TAG_FILE_PLACEHOLDER                              uint32   = 0x80000015
	IO_REPARSE_TAG_WOF                                           uint32   = 0x80000017
	IO_REPARSE_TAG_WCI                                           uint32   = 0x80000018
	IO_REPARSE_TAG_WCI_1                                         uint32   = 0x90001018
	IO_REPARSE_TAG_GLOBAL_REPARSE                                uint32   = 0xa0000019
	IO_REPARSE_TAG_CLOUD                                         uint32   = 0x9000001a
	IO_REPARSE_TAG_CLOUD_1                                       uint32   = 0x9000101a
	IO_REPARSE_TAG_CLOUD_2                                       uint32   = 0x9000201a
	IO_REPARSE_TAG_CLOUD_3                                       uint32   = 0x9000301a
	IO_REPARSE_TAG_CLOUD_4                                       uint32   = 0x9000401a
	IO_REPARSE_TAG_CLOUD_5                                       uint32   = 0x9000501a
	IO_REPARSE_TAG_CLOUD_6                                       uint32   = 0x9000601a
	IO_REPARSE_TAG_CLOUD_7                                       uint32   = 0x9000701a
	IO_REPARSE_TAG_CLOUD_8                                       uint32   = 0x9000801a
	IO_REPARSE_TAG_CLOUD_9                                       uint32   = 0x9000901a
	IO_REPARSE_TAG_CLOUD_A                                       uint32   = 0x9000a01a
	IO_REPARSE_TAG_CLOUD_B                                       uint32   = 0x9000b01a
	IO_REPARSE_TAG_CLOUD_C                                       uint32   = 0x9000c01a
	IO_REPARSE_TAG_CLOUD_D                                       uint32   = 0x9000d01a
	IO_REPARSE_TAG_CLOUD_E                                       uint32   = 0x9000e01a
	IO_REPARSE_TAG_CLOUD_F                                       uint32   = 0x9000f01a
	IO_REPARSE_TAG_CLOUD_MASK                                    uint32   = 0xf000
	IO_REPARSE_TAG_APPEXECLINK                                   uint32   = 0x8000001b
	IO_REPARSE_TAG_PROJFS                                        uint32   = 0x9000001c
	IO_REPARSE_TAG_STORAGE_SYNC                                  uint32   = 0x8000001e
	IO_REPARSE_TAG_WCI_TOMBSTONE                                 uint32   = 0xa000001f
	IO_REPARSE_TAG_UNHANDLED                                     uint32   = 0x80000020
	IO_REPARSE_TAG_ONEDRIVE                                      uint32   = 0x80000021
	IO_REPARSE_TAG_PROJFS_TOMBSTONE                              uint32   = 0xa0000022
	IO_REPARSE_TAG_AF_UNIX                                       uint32   = 0x80000023
	IO_REPARSE_TAG_WCI_LINK                                      uint32   = 0xa0000027
	IO_REPARSE_TAG_WCI_LINK_1                                    uint32   = 0xa0001027
	IO_REPARSE_TAG_DATALESS_CIM                                  uint32   = 0xa0000028
	SCRUB_DATA_INPUT_FLAG_RESUME                                 uint32   = 0x1
	SCRUB_DATA_INPUT_FLAG_SKIP_IN_SYNC                           uint32   = 0x2
	SCRUB_DATA_INPUT_FLAG_SKIP_NON_INTEGRITY_DATA                uint32   = 0x4
	SCRUB_DATA_INPUT_FLAG_IGNORE_REDUNDANCY                      uint32   = 0x8
	SCRUB_DATA_INPUT_FLAG_SKIP_DATA                              uint32   = 0x10
	SCRUB_DATA_INPUT_FLAG_SCRUB_BY_OBJECT_ID                     uint32   = 0x20
	SCRUB_DATA_INPUT_FLAG_OPLOCK_NOT_ACQUIRED                    uint32   = 0x40
	SCRUB_DATA_OUTPUT_FLAG_INCOMPLETE                            uint32   = 0x1
	SCRUB_DATA_OUTPUT_FLAG_NON_USER_DATA_RANGE                   uint32   = 0x10000
	SCRUB_DATA_OUTPUT_FLAG_PARITY_EXTENT_DATA_RETURNED           uint32   = 0x20000
	SCRUB_DATA_OUTPUT_FLAG_RESUME_CONTEXT_LENGTH_SPECIFIED       uint32   = 0x40000
	SHUFFLE_FILE_FLAG_SKIP_INITIALIZING_NEW_CLUSTERS             uint32   = 0x1
	IO_COMPLETION_MODIFY_STATE                                   uint32   = 0x2
	SMB_CCF_APP_INSTANCE_EA_NAME                                 string   = "ClusteredApplicationInstance"
	NETWORK_APP_INSTANCE_CSV_FLAGS_VALID_ONLY_IF_CSV_COORDINATOR uint32   = 0x1
	POWERBUTTON_ACTION_INDEX_NOTHING                             uint32   = 0x0
	POWERBUTTON_ACTION_INDEX_SLEEP                               uint32   = 0x1
	POWERBUTTON_ACTION_INDEX_HIBERNATE                           uint32   = 0x2
	POWERBUTTON_ACTION_INDEX_SHUTDOWN                            uint32   = 0x3
	POWERBUTTON_ACTION_INDEX_TURN_OFF_THE_DISPLAY                uint32   = 0x4
	POWERBUTTON_ACTION_VALUE_NOTHING                             uint32   = 0x0
	POWERBUTTON_ACTION_VALUE_SLEEP                               uint32   = 0x2
	POWERBUTTON_ACTION_VALUE_HIBERNATE                           uint32   = 0x3
	POWERBUTTON_ACTION_VALUE_SHUTDOWN                            uint32   = 0x6
	POWERBUTTON_ACTION_VALUE_TURN_OFF_THE_DISPLAY                uint32   = 0x8
	PERFSTATE_POLICY_CHANGE_IDEAL                                uint32   = 0x0
	PERFSTATE_POLICY_CHANGE_SINGLE                               uint32   = 0x1
	PERFSTATE_POLICY_CHANGE_ROCKET                               uint32   = 0x2
	PERFSTATE_POLICY_CHANGE_IDEAL_AGGRESSIVE                     uint32   = 0x3
	PERFSTATE_POLICY_CHANGE_DECREASE_MAX                         uint32   = 0x2
	PERFSTATE_POLICY_CHANGE_INCREASE_MAX                         uint32   = 0x3
	PROCESSOR_THROTTLE_DISABLED                                  uint32   = 0x0
	PROCESSOR_THROTTLE_ENABLED                                   uint32   = 0x1
	PROCESSOR_THROTTLE_AUTOMATIC                                 uint32   = 0x2
	PROCESSOR_PERF_BOOST_POLICY_DISABLED                         uint32   = 0x0
	PROCESSOR_PERF_BOOST_POLICY_MAX                              uint32   = 0x64
	PROCESSOR_PERF_BOOST_MODE_DISABLED                           uint32   = 0x0
	PROCESSOR_PERF_BOOST_MODE_ENABLED                            uint32   = 0x1
	PROCESSOR_PERF_BOOST_MODE_AGGRESSIVE                         uint32   = 0x2
	PROCESSOR_PERF_BOOST_MODE_EFFICIENT_ENABLED                  uint32   = 0x3
	PROCESSOR_PERF_BOOST_MODE_EFFICIENT_AGGRESSIVE               uint32   = 0x4
	PROCESSOR_PERF_BOOST_MODE_AGGRESSIVE_AT_GUARANTEED           uint32   = 0x5
	PROCESSOR_PERF_BOOST_MODE_EFFICIENT_AGGRESSIVE_AT_GUARANTEED uint32   = 0x6
	PROCESSOR_PERF_BOOST_MODE_MAX                                uint32   = 0x6
	PROCESSOR_PERF_AUTONOMOUS_MODE_DISABLED                      uint32   = 0x0
	PROCESSOR_PERF_AUTONOMOUS_MODE_ENABLED                       uint32   = 0x1
	PROCESSOR_PERF_PERFORMANCE_PREFERENCE                        uint32   = 0xff
	PROCESSOR_PERF_ENERGY_PREFERENCE                             uint32   = 0x0
	PROCESSOR_PERF_MINIMUM_ACTIVITY_WINDOW                       uint32   = 0x0
	PROCESSOR_PERF_MAXIMUM_ACTIVITY_WINDOW                       uint32   = 0x4bb2a980
	PROCESSOR_DUTY_CYCLING_DISABLED                              uint32   = 0x0
	PROCESSOR_DUTY_CYCLING_ENABLED                               uint32   = 0x1
	CORE_PARKING_POLICY_CHANGE_IDEAL                             uint32   = 0x0
	CORE_PARKING_POLICY_CHANGE_SINGLE                            uint32   = 0x1
	CORE_PARKING_POLICY_CHANGE_ROCKET                            uint32   = 0x2
	CORE_PARKING_POLICY_CHANGE_MULTISTEP                         uint32   = 0x3
	CORE_PARKING_POLICY_CHANGE_MAX                               uint32   = 0x3
	POWER_DEVICE_IDLE_POLICY_PERFORMANCE                         uint32   = 0x0
	POWER_DEVICE_IDLE_POLICY_CONSERVATIVE                        uint32   = 0x1
	POWER_CONNECTIVITY_IN_STANDBY_DISABLED                       uint32   = 0x0
	POWER_CONNECTIVITY_IN_STANDBY_ENABLED                        uint32   = 0x1
	POWER_CONNECTIVITY_IN_STANDBY_SYSTEM_MANAGED                 uint32   = 0x2
	POWER_DISCONNECTED_STANDBY_MODE_NORMAL                       uint32   = 0x0
	POWER_DISCONNECTED_STANDBY_MODE_AGGRESSIVE                   uint32   = 0x1
	POWER_SYSTEM_MAXIMUM                                         uint32   = 0x7
	DIAGNOSTIC_REASON_VERSION                                    uint32   = 0x0
	DIAGNOSTIC_REASON_SIMPLE_STRING                              uint32   = 0x1
	DIAGNOSTIC_REASON_DETAILED_STRING                            uint32   = 0x2
	DIAGNOSTIC_REASON_NOT_SPECIFIED                              uint32   = 0x80000000
	POWER_REQUEST_CONTEXT_VERSION                                uint32   = 0x0
	PDCAP_D0_SUPPORTED                                           uint32   = 0x1
	PDCAP_D1_SUPPORTED                                           uint32   = 0x2
	PDCAP_D2_SUPPORTED                                           uint32   = 0x4
	PDCAP_D3_SUPPORTED                                           uint32   = 0x8
	PDCAP_WAKE_FROM_D0_SUPPORTED                                 uint32   = 0x10
	PDCAP_WAKE_FROM_D1_SUPPORTED                                 uint32   = 0x20
	PDCAP_WAKE_FROM_D2_SUPPORTED                                 uint32   = 0x40
	PDCAP_WAKE_FROM_D3_SUPPORTED                                 uint32   = 0x80
	PDCAP_WARM_EJECT_SUPPORTED                                   uint32   = 0x100
	POWER_SETTING_VALUE_VERSION                                  uint32   = 0x1
	PROC_IDLE_BUCKET_COUNT                                       uint32   = 0x6
	PROC_IDLE_BUCKET_COUNT_EX                                    uint32   = 0x10
	ACPI_PPM_SOFTWARE_ALL                                        uint32   = 0xfc
	ACPI_PPM_SOFTWARE_ANY                                        uint32   = 0xfd
	ACPI_PPM_HARDWARE_ALL                                        uint32   = 0xfe
	MS_PPM_SOFTWARE_ALL                                          uint32   = 0x1
	PPM_FIRMWARE_ACPI1C2                                         uint32   = 0x1
	PPM_FIRMWARE_ACPI1C3                                         uint32   = 0x2
	PPM_FIRMWARE_ACPI1TSTATES                                    uint32   = 0x4
	PPM_FIRMWARE_CST                                             uint32   = 0x8
	PPM_FIRMWARE_CSD                                             uint32   = 0x10
	PPM_FIRMWARE_PCT                                             uint32   = 0x20
	PPM_FIRMWARE_PSS                                             uint32   = 0x40
	PPM_FIRMWARE_XPSS                                            uint32   = 0x80
	PPM_FIRMWARE_PPC                                             uint32   = 0x100
	PPM_FIRMWARE_PSD                                             uint32   = 0x200
	PPM_FIRMWARE_PTC                                             uint32   = 0x400
	PPM_FIRMWARE_TSS                                             uint32   = 0x800
	PPM_FIRMWARE_TPC                                             uint32   = 0x1000
	PPM_FIRMWARE_TSD                                             uint32   = 0x2000
	PPM_FIRMWARE_PCCH                                            uint32   = 0x4000
	PPM_FIRMWARE_PCCP                                            uint32   = 0x8000
	PPM_FIRMWARE_OSC                                             uint32   = 0x10000
	PPM_FIRMWARE_PDC                                             uint32   = 0x20000
	PPM_FIRMWARE_CPC                                             uint32   = 0x40000
	PPM_FIRMWARE_LPI                                             uint32   = 0x80000
	PPM_PERFORMANCE_IMPLEMENTATION_NONE                          uint32   = 0x0
	PPM_PERFORMANCE_IMPLEMENTATION_PSTATES                       uint32   = 0x1
	PPM_PERFORMANCE_IMPLEMENTATION_PCCV1                         uint32   = 0x2
	PPM_PERFORMANCE_IMPLEMENTATION_CPPC                          uint32   = 0x3
	PPM_PERFORMANCE_IMPLEMENTATION_PEP                           uint32   = 0x4
	PPM_IDLE_IMPLEMENTATION_NONE                                 uint32   = 0x0
	PPM_IDLE_IMPLEMENTATION_CSTATES                              uint32   = 0x1
	PPM_IDLE_IMPLEMENTATION_PEP                                  uint32   = 0x2
	PPM_IDLE_IMPLEMENTATION_MICROPEP                             uint32   = 0x3
	PPM_IDLE_IMPLEMENTATION_LPISTATES                            uint32   = 0x4
	POWER_ACTION_QUERY_ALLOWED                                   uint32   = 0x1
	POWER_ACTION_UI_ALLOWED                                      uint32   = 0x2
	POWER_ACTION_OVERRIDE_APPS                                   uint32   = 0x4
	POWER_ACTION_HIBERBOOT                                       uint32   = 0x8
	POWER_ACTION_USER_NOTIFY                                     uint32   = 0x10
	POWER_ACTION_DOZE_TO_HIBERNATE                               uint32   = 0x20
	POWER_ACTION_ACPI_CRITICAL                                   uint32   = 0x1000000
	POWER_ACTION_ACPI_USER_NOTIFY                                uint32   = 0x2000000
	POWER_ACTION_DIRECTED_DRIPS                                  uint32   = 0x4000000
	POWER_ACTION_PSEUDO_TRANSITION                               uint32   = 0x8000000
	POWER_ACTION_LIGHTEST_FIRST                                  uint32   = 0x10000000
	POWER_ACTION_LOCK_CONSOLE                                    uint32   = 0x20000000
	POWER_ACTION_DISABLE_WAKES                                   uint32   = 0x40000000
	POWER_ACTION_CRITICAL                                        uint32   = 0x80000000
	POWER_USER_NOTIFY_FORCED_SHUTDOWN                            uint32   = 0x20
	BATTERY_DISCHARGE_FLAGS_EVENTCODE_MASK                       uint32   = 0x7
	BATTERY_DISCHARGE_FLAGS_ENABLE                               uint32   = 0x80000000
	NUM_DISCHARGE_POLICIES                                       uint32   = 0x4
	DISCHARGE_POLICY_CRITICAL                                    uint32   = 0x0
	DISCHARGE_POLICY_LOW                                         uint32   = 0x1
	PROCESSOR_IDLESTATE_POLICY_COUNT                             uint32   = 0x3
	PO_THROTTLE_NONE                                             uint32   = 0x0
	PO_THROTTLE_CONSTANT                                         uint32   = 0x1
	PO_THROTTLE_DEGRADE                                          uint32   = 0x2
	PO_THROTTLE_ADAPTIVE                                         uint32   = 0x3
	PO_THROTTLE_MAXIMUM                                          uint32   = 0x4
	HIBERFILE_TYPE_NONE                                          uint32   = 0x0
	HIBERFILE_TYPE_REDUCED                                       uint32   = 0x1
	HIBERFILE_TYPE_FULL                                          uint32   = 0x2
	HIBERFILE_TYPE_MAX                                           uint32   = 0x3
	IMAGE_DOS_SIGNATURE                                          uint16   = 0x5a4d
	IMAGE_OS2_SIGNATURE                                          uint16   = 0x454e
	IMAGE_OS2_SIGNATURE_LE                                       uint16   = 0x454c
	IMAGE_VXD_SIGNATURE                                          uint16   = 0x454c
	IMAGE_NT_SIGNATURE                                           uint16   = 0x4550
	IMAGE_SIZEOF_FILE_HEADER                                     uint32   = 0x14
	IMAGE_NUMBEROF_DIRECTORY_ENTRIES                             uint32   = 0x10
	IMAGE_SIZEOF_SHORT_NAME                                      uint32   = 0x8
	IMAGE_SIZEOF_SECTION_HEADER                                  uint32   = 0x28
	IMAGE_SIZEOF_SYMBOL                                          uint32   = 0x12
	IMAGE_SYM_SECTION_MAX                                        uint32   = 0xfeff
	IMAGE_SYM_SECTION_MAX_EX                                     uint32   = 0x7fffffff
	IMAGE_SYM_TYPE_NULL                                          uint32   = 0x0
	IMAGE_SYM_TYPE_VOID                                          uint32   = 0x1
	IMAGE_SYM_TYPE_CHAR                                          uint32   = 0x2
	IMAGE_SYM_TYPE_SHORT                                         uint32   = 0x3
	IMAGE_SYM_TYPE_INT                                           uint32   = 0x4
	IMAGE_SYM_TYPE_LONG                                          uint32   = 0x5
	IMAGE_SYM_TYPE_FLOAT                                         uint32   = 0x6
	IMAGE_SYM_TYPE_DOUBLE                                        uint32   = 0x7
	IMAGE_SYM_TYPE_STRUCT                                        uint32   = 0x8
	IMAGE_SYM_TYPE_UNION                                         uint32   = 0x9
	IMAGE_SYM_TYPE_ENUM                                          uint32   = 0xa
	IMAGE_SYM_TYPE_MOE                                           uint32   = 0xb
	IMAGE_SYM_TYPE_BYTE                                          uint32   = 0xc
	IMAGE_SYM_TYPE_WORD                                          uint32   = 0xd
	IMAGE_SYM_TYPE_UINT                                          uint32   = 0xe
	IMAGE_SYM_TYPE_DWORD                                         uint32   = 0xf
	IMAGE_SYM_TYPE_PCODE                                         uint32   = 0x8000
	IMAGE_SYM_DTYPE_NULL                                         uint32   = 0x0
	IMAGE_SYM_DTYPE_POINTER                                      uint32   = 0x1
	IMAGE_SYM_DTYPE_FUNCTION                                     uint32   = 0x2
	IMAGE_SYM_DTYPE_ARRAY                                        uint32   = 0x3
	IMAGE_SYM_CLASS_NULL                                         uint32   = 0x0
	IMAGE_SYM_CLASS_AUTOMATIC                                    uint32   = 0x1
	IMAGE_SYM_CLASS_EXTERNAL                                     uint32   = 0x2
	IMAGE_SYM_CLASS_STATIC                                       uint32   = 0x3
	IMAGE_SYM_CLASS_REGISTER                                     uint32   = 0x4
	IMAGE_SYM_CLASS_EXTERNAL_DEF                                 uint32   = 0x5
	IMAGE_SYM_CLASS_LABEL                                        uint32   = 0x6
	IMAGE_SYM_CLASS_UNDEFINED_LABEL                              uint32   = 0x7
	IMAGE_SYM_CLASS_MEMBER_OF_STRUCT                             uint32   = 0x8
	IMAGE_SYM_CLASS_ARGUMENT                                     uint32   = 0x9
	IMAGE_SYM_CLASS_STRUCT_TAG                                   uint32   = 0xa
	IMAGE_SYM_CLASS_MEMBER_OF_UNION                              uint32   = 0xb
	IMAGE_SYM_CLASS_UNION_TAG                                    uint32   = 0xc
	IMAGE_SYM_CLASS_TYPE_DEFINITION                              uint32   = 0xd
	IMAGE_SYM_CLASS_UNDEFINED_STATIC                             uint32   = 0xe
	IMAGE_SYM_CLASS_ENUM_TAG                                     uint32   = 0xf
	IMAGE_SYM_CLASS_MEMBER_OF_ENUM                               uint32   = 0x10
	IMAGE_SYM_CLASS_REGISTER_PARAM                               uint32   = 0x11
	IMAGE_SYM_CLASS_BIT_FIELD                                    uint32   = 0x12
	IMAGE_SYM_CLASS_FAR_EXTERNAL                                 uint32   = 0x44
	IMAGE_SYM_CLASS_BLOCK                                        uint32   = 0x64
	IMAGE_SYM_CLASS_FUNCTION                                     uint32   = 0x65
	IMAGE_SYM_CLASS_END_OF_STRUCT                                uint32   = 0x66
	IMAGE_SYM_CLASS_FILE                                         uint32   = 0x67
	IMAGE_SYM_CLASS_SECTION                                      uint32   = 0x68
	IMAGE_SYM_CLASS_WEAK_EXTERNAL                                uint32   = 0x69
	IMAGE_SYM_CLASS_CLR_TOKEN                                    uint32   = 0x6b
	N_BTMASK                                                     uint32   = 0xf
	N_TMASK                                                      uint32   = 0x30
	N_TMASK1                                                     uint32   = 0xc0
	N_TMASK2                                                     uint32   = 0xf0
	N_BTSHFT                                                     uint32   = 0x4
	N_TSHIFT                                                     uint32   = 0x2
	IMAGE_COMDAT_SELECT_NODUPLICATES                             uint32   = 0x1
	IMAGE_COMDAT_SELECT_ANY                                      uint32   = 0x2
	IMAGE_COMDAT_SELECT_SAME_SIZE                                uint32   = 0x3
	IMAGE_COMDAT_SELECT_EXACT_MATCH                              uint32   = 0x4
	IMAGE_COMDAT_SELECT_ASSOCIATIVE                              uint32   = 0x5
	IMAGE_COMDAT_SELECT_LARGEST                                  uint32   = 0x6
	IMAGE_COMDAT_SELECT_NEWEST                                   uint32   = 0x7
	IMAGE_WEAK_EXTERN_SEARCH_NOLIBRARY                           uint32   = 0x1
	IMAGE_WEAK_EXTERN_SEARCH_LIBRARY                             uint32   = 0x2
	IMAGE_WEAK_EXTERN_SEARCH_ALIAS                               uint32   = 0x3
	IMAGE_WEAK_EXTERN_ANTI_DEPENDENCY                            uint32   = 0x4
	IMAGE_REL_I386_ABSOLUTE                                      uint32   = 0x0
	IMAGE_REL_I386_DIR16                                         uint32   = 0x1
	IMAGE_REL_I386_REL16                                         uint32   = 0x2
	IMAGE_REL_I386_DIR32                                         uint32   = 0x6
	IMAGE_REL_I386_DIR32NB                                       uint32   = 0x7
	IMAGE_REL_I386_SEG12                                         uint32   = 0x9
	IMAGE_REL_I386_SECTION                                       uint32   = 0xa
	IMAGE_REL_I386_SECREL                                        uint32   = 0xb
	IMAGE_REL_I386_TOKEN                                         uint32   = 0xc
	IMAGE_REL_I386_SECREL7                                       uint32   = 0xd
	IMAGE_REL_I386_REL32                                         uint32   = 0x14
	IMAGE_REL_MIPS_ABSOLUTE                                      uint32   = 0x0
	IMAGE_REL_MIPS_REFHALF                                       uint32   = 0x1
	IMAGE_REL_MIPS_REFWORD                                       uint32   = 0x2
	IMAGE_REL_MIPS_JMPADDR                                       uint32   = 0x3
	IMAGE_REL_MIPS_REFHI                                         uint32   = 0x4
	IMAGE_REL_MIPS_REFLO                                         uint32   = 0x5
	IMAGE_REL_MIPS_GPREL                                         uint32   = 0x6
	IMAGE_REL_MIPS_LITERAL                                       uint32   = 0x7
	IMAGE_REL_MIPS_SECTION                                       uint32   = 0xa
	IMAGE_REL_MIPS_SECREL                                        uint32   = 0xb
	IMAGE_REL_MIPS_SECRELLO                                      uint32   = 0xc
	IMAGE_REL_MIPS_SECRELHI                                      uint32   = 0xd
	IMAGE_REL_MIPS_TOKEN                                         uint32   = 0xe
	IMAGE_REL_MIPS_JMPADDR16                                     uint32   = 0x10
	IMAGE_REL_MIPS_REFWORDNB                                     uint32   = 0x22
	IMAGE_REL_MIPS_PAIR                                          uint32   = 0x25
	IMAGE_REL_ALPHA_ABSOLUTE                                     uint32   = 0x0
	IMAGE_REL_ALPHA_REFLONG                                      uint32   = 0x1
	IMAGE_REL_ALPHA_REFQUAD                                      uint32   = 0x2
	IMAGE_REL_ALPHA_GPREL32                                      uint32   = 0x3
	IMAGE_REL_ALPHA_LITERAL                                      uint32   = 0x4
	IMAGE_REL_ALPHA_LITUSE                                       uint32   = 0x5
	IMAGE_REL_ALPHA_GPDISP                                       uint32   = 0x6
	IMAGE_REL_ALPHA_BRADDR                                       uint32   = 0x7
	IMAGE_REL_ALPHA_HINT                                         uint32   = 0x8
	IMAGE_REL_ALPHA_INLINE_REFLONG                               uint32   = 0x9
	IMAGE_REL_ALPHA_REFHI                                        uint32   = 0xa
	IMAGE_REL_ALPHA_REFLO                                        uint32   = 0xb
	IMAGE_REL_ALPHA_PAIR                                         uint32   = 0xc
	IMAGE_REL_ALPHA_MATCH                                        uint32   = 0xd
	IMAGE_REL_ALPHA_SECTION                                      uint32   = 0xe
	IMAGE_REL_ALPHA_SECREL                                       uint32   = 0xf
	IMAGE_REL_ALPHA_REFLONGNB                                    uint32   = 0x10
	IMAGE_REL_ALPHA_SECRELLO                                     uint32   = 0x11
	IMAGE_REL_ALPHA_SECRELHI                                     uint32   = 0x12
	IMAGE_REL_ALPHA_REFQ3                                        uint32   = 0x13
	IMAGE_REL_ALPHA_REFQ2                                        uint32   = 0x14
	IMAGE_REL_ALPHA_REFQ1                                        uint32   = 0x15
	IMAGE_REL_ALPHA_GPRELLO                                      uint32   = 0x16
	IMAGE_REL_ALPHA_GPRELHI                                      uint32   = 0x17
	IMAGE_REL_PPC_ABSOLUTE                                       uint32   = 0x0
	IMAGE_REL_PPC_ADDR64                                         uint32   = 0x1
	IMAGE_REL_PPC_ADDR32                                         uint32   = 0x2
	IMAGE_REL_PPC_ADDR24                                         uint32   = 0x3
	IMAGE_REL_PPC_ADDR16                                         uint32   = 0x4
	IMAGE_REL_PPC_ADDR14                                         uint32   = 0x5
	IMAGE_REL_PPC_REL24                                          uint32   = 0x6
	IMAGE_REL_PPC_REL14                                          uint32   = 0x7
	IMAGE_REL_PPC_TOCREL16                                       uint32   = 0x8
	IMAGE_REL_PPC_TOCREL14                                       uint32   = 0x9
	IMAGE_REL_PPC_ADDR32NB                                       uint32   = 0xa
	IMAGE_REL_PPC_SECREL                                         uint32   = 0xb
	IMAGE_REL_PPC_SECTION                                        uint32   = 0xc
	IMAGE_REL_PPC_IFGLUE                                         uint32   = 0xd
	IMAGE_REL_PPC_IMGLUE                                         uint32   = 0xe
	IMAGE_REL_PPC_SECREL16                                       uint32   = 0xf
	IMAGE_REL_PPC_REFHI                                          uint32   = 0x10
	IMAGE_REL_PPC_REFLO                                          uint32   = 0x11
	IMAGE_REL_PPC_PAIR                                           uint32   = 0x12
	IMAGE_REL_PPC_SECRELLO                                       uint32   = 0x13
	IMAGE_REL_PPC_SECRELHI                                       uint32   = 0x14
	IMAGE_REL_PPC_GPREL                                          uint32   = 0x15
	IMAGE_REL_PPC_TOKEN                                          uint32   = 0x16
	IMAGE_REL_PPC_TYPEMASK                                       uint32   = 0xff
	IMAGE_REL_PPC_NEG                                            uint32   = 0x100
	IMAGE_REL_PPC_BRTAKEN                                        uint32   = 0x200
	IMAGE_REL_PPC_BRNTAKEN                                       uint32   = 0x400
	IMAGE_REL_PPC_TOCDEFN                                        uint32   = 0x800
	IMAGE_REL_SH3_ABSOLUTE                                       uint32   = 0x0
	IMAGE_REL_SH3_DIRECT16                                       uint32   = 0x1
	IMAGE_REL_SH3_DIRECT32                                       uint32   = 0x2
	IMAGE_REL_SH3_DIRECT8                                        uint32   = 0x3
	IMAGE_REL_SH3_DIRECT8_WORD                                   uint32   = 0x4
	IMAGE_REL_SH3_DIRECT8_LONG                                   uint32   = 0x5
	IMAGE_REL_SH3_DIRECT4                                        uint32   = 0x6
	IMAGE_REL_SH3_DIRECT4_WORD                                   uint32   = 0x7
	IMAGE_REL_SH3_DIRECT4_LONG                                   uint32   = 0x8
	IMAGE_REL_SH3_PCREL8_WORD                                    uint32   = 0x9
	IMAGE_REL_SH3_PCREL8_LONG                                    uint32   = 0xa
	IMAGE_REL_SH3_PCREL12_WORD                                   uint32   = 0xb
	IMAGE_REL_SH3_STARTOF_SECTION                                uint32   = 0xc
	IMAGE_REL_SH3_SIZEOF_SECTION                                 uint32   = 0xd
	IMAGE_REL_SH3_SECTION                                        uint32   = 0xe
	IMAGE_REL_SH3_SECREL                                         uint32   = 0xf
	IMAGE_REL_SH3_DIRECT32_NB                                    uint32   = 0x10
	IMAGE_REL_SH3_GPREL4_LONG                                    uint32   = 0x11
	IMAGE_REL_SH3_TOKEN                                          uint32   = 0x12
	IMAGE_REL_SHM_PCRELPT                                        uint32   = 0x13
	IMAGE_REL_SHM_REFLO                                          uint32   = 0x14
	IMAGE_REL_SHM_REFHALF                                        uint32   = 0x15
	IMAGE_REL_SHM_RELLO                                          uint32   = 0x16
	IMAGE_REL_SHM_RELHALF                                        uint32   = 0x17
	IMAGE_REL_SHM_PAIR                                           uint32   = 0x18
	IMAGE_REL_SH_NOMODE                                          uint32   = 0x8000
	IMAGE_REL_ARM_ABSOLUTE                                       uint32   = 0x0
	IMAGE_REL_ARM_ADDR32                                         uint32   = 0x1
	IMAGE_REL_ARM_ADDR32NB                                       uint32   = 0x2
	IMAGE_REL_ARM_BRANCH24                                       uint32   = 0x3
	IMAGE_REL_ARM_BRANCH11                                       uint32   = 0x4
	IMAGE_REL_ARM_TOKEN                                          uint32   = 0x5
	IMAGE_REL_ARM_GPREL12                                        uint32   = 0x6
	IMAGE_REL_ARM_GPREL7                                         uint32   = 0x7
	IMAGE_REL_ARM_BLX24                                          uint32   = 0x8
	IMAGE_REL_ARM_BLX11                                          uint32   = 0x9
	IMAGE_REL_ARM_SECTION                                        uint32   = 0xe
	IMAGE_REL_ARM_SECREL                                         uint32   = 0xf
	IMAGE_REL_ARM_MOV32A                                         uint32   = 0x10
	IMAGE_REL_ARM_MOV32                                          uint32   = 0x10
	IMAGE_REL_ARM_MOV32T                                         uint32   = 0x11
	IMAGE_REL_THUMB_MOV32                                        uint32   = 0x11
	IMAGE_REL_ARM_BRANCH20T                                      uint32   = 0x12
	IMAGE_REL_THUMB_BRANCH20                                     uint32   = 0x12
	IMAGE_REL_ARM_BRANCH24T                                      uint32   = 0x14
	IMAGE_REL_THUMB_BRANCH24                                     uint32   = 0x14
	IMAGE_REL_ARM_BLX23T                                         uint32   = 0x15
	IMAGE_REL_THUMB_BLX23                                        uint32   = 0x15
	IMAGE_REL_AM_ABSOLUTE                                        uint32   = 0x0
	IMAGE_REL_AM_ADDR32                                          uint32   = 0x1
	IMAGE_REL_AM_ADDR32NB                                        uint32   = 0x2
	IMAGE_REL_AM_CALL32                                          uint32   = 0x3
	IMAGE_REL_AM_FUNCINFO                                        uint32   = 0x4
	IMAGE_REL_AM_REL32_1                                         uint32   = 0x5
	IMAGE_REL_AM_REL32_2                                         uint32   = 0x6
	IMAGE_REL_AM_SECREL                                          uint32   = 0x7
	IMAGE_REL_AM_SECTION                                         uint32   = 0x8
	IMAGE_REL_AM_TOKEN                                           uint32   = 0x9
	IMAGE_REL_ARM64_ABSOLUTE                                     uint32   = 0x0
	IMAGE_REL_ARM64_ADDR32                                       uint32   = 0x1
	IMAGE_REL_ARM64_ADDR32NB                                     uint32   = 0x2
	IMAGE_REL_ARM64_BRANCH26                                     uint32   = 0x3
	IMAGE_REL_ARM64_PAGEBASE_REL21                               uint32   = 0x4
	IMAGE_REL_ARM64_REL21                                        uint32   = 0x5
	IMAGE_REL_ARM64_PAGEOFFSET_12A                               uint32   = 0x6
	IMAGE_REL_ARM64_PAGEOFFSET_12L                               uint32   = 0x7
	IMAGE_REL_ARM64_SECREL                                       uint32   = 0x8
	IMAGE_REL_ARM64_SECREL_LOW12A                                uint32   = 0x9
	IMAGE_REL_ARM64_SECREL_HIGH12A                               uint32   = 0xa
	IMAGE_REL_ARM64_SECREL_LOW12L                                uint32   = 0xb
	IMAGE_REL_ARM64_TOKEN                                        uint32   = 0xc
	IMAGE_REL_ARM64_SECTION                                      uint32   = 0xd
	IMAGE_REL_ARM64_ADDR64                                       uint32   = 0xe
	IMAGE_REL_ARM64_BRANCH19                                     uint32   = 0xf
	IMAGE_REL_AMD64_ABSOLUTE                                     uint32   = 0x0
	IMAGE_REL_AMD64_ADDR64                                       uint32   = 0x1
	IMAGE_REL_AMD64_ADDR32                                       uint32   = 0x2
	IMAGE_REL_AMD64_ADDR32NB                                     uint32   = 0x3
	IMAGE_REL_AMD64_REL32                                        uint32   = 0x4
	IMAGE_REL_AMD64_REL32_1                                      uint32   = 0x5
	IMAGE_REL_AMD64_REL32_2                                      uint32   = 0x6
	IMAGE_REL_AMD64_REL32_3                                      uint32   = 0x7
	IMAGE_REL_AMD64_REL32_4                                      uint32   = 0x8
	IMAGE_REL_AMD64_REL32_5                                      uint32   = 0x9
	IMAGE_REL_AMD64_SECTION                                      uint32   = 0xa
	IMAGE_REL_AMD64_SECREL                                       uint32   = 0xb
	IMAGE_REL_AMD64_SECREL7                                      uint32   = 0xc
	IMAGE_REL_AMD64_TOKEN                                        uint32   = 0xd
	IMAGE_REL_AMD64_SREL32                                       uint32   = 0xe
	IMAGE_REL_AMD64_PAIR                                         uint32   = 0xf
	IMAGE_REL_AMD64_SSPAN32                                      uint32   = 0x10
	IMAGE_REL_AMD64_EHANDLER                                     uint32   = 0x11
	IMAGE_REL_AMD64_IMPORT_BR                                    uint32   = 0x12
	IMAGE_REL_AMD64_IMPORT_CALL                                  uint32   = 0x13
	IMAGE_REL_AMD64_CFG_BR                                       uint32   = 0x14
	IMAGE_REL_AMD64_CFG_BR_REX                                   uint32   = 0x15
	IMAGE_REL_AMD64_CFG_CALL                                     uint32   = 0x16
	IMAGE_REL_AMD64_INDIR_BR                                     uint32   = 0x17
	IMAGE_REL_AMD64_INDIR_BR_REX                                 uint32   = 0x18
	IMAGE_REL_AMD64_INDIR_CALL                                   uint32   = 0x19
	IMAGE_REL_AMD64_INDIR_BR_SWITCHTABLE_FIRST                   uint32   = 0x20
	IMAGE_REL_AMD64_INDIR_BR_SWITCHTABLE_LAST                    uint32   = 0x2f
	IMAGE_REL_IA64_ABSOLUTE                                      uint32   = 0x0
	IMAGE_REL_IA64_IMM14                                         uint32   = 0x1
	IMAGE_REL_IA64_IMM22                                         uint32   = 0x2
	IMAGE_REL_IA64_IMM64                                         uint32   = 0x3
	IMAGE_REL_IA64_DIR32                                         uint32   = 0x4
	IMAGE_REL_IA64_DIR64                                         uint32   = 0x5
	IMAGE_REL_IA64_PCREL21B                                      uint32   = 0x6
	IMAGE_REL_IA64_PCREL21M                                      uint32   = 0x7
	IMAGE_REL_IA64_PCREL21F                                      uint32   = 0x8
	IMAGE_REL_IA64_GPREL22                                       uint32   = 0x9
	IMAGE_REL_IA64_LTOFF22                                       uint32   = 0xa
	IMAGE_REL_IA64_SECTION                                       uint32   = 0xb
	IMAGE_REL_IA64_SECREL22                                      uint32   = 0xc
	IMAGE_REL_IA64_SECREL64I                                     uint32   = 0xd
	IMAGE_REL_IA64_SECREL32                                      uint32   = 0xe
	IMAGE_REL_IA64_DIR32NB                                       uint32   = 0x10
	IMAGE_REL_IA64_SREL14                                        uint32   = 0x11
	IMAGE_REL_IA64_SREL22                                        uint32   = 0x12
	IMAGE_REL_IA64_SREL32                                        uint32   = 0x13
	IMAGE_REL_IA64_UREL32                                        uint32   = 0x14
	IMAGE_REL_IA64_PCREL60X                                      uint32   = 0x15
	IMAGE_REL_IA64_PCREL60B                                      uint32   = 0x16
	IMAGE_REL_IA64_PCREL60F                                      uint32   = 0x17
	IMAGE_REL_IA64_PCREL60I                                      uint32   = 0x18
	IMAGE_REL_IA64_PCREL60M                                      uint32   = 0x19
	IMAGE_REL_IA64_IMMGPREL64                                    uint32   = 0x1a
	IMAGE_REL_IA64_TOKEN                                         uint32   = 0x1b
	IMAGE_REL_IA64_GPREL32                                       uint32   = 0x1c
	IMAGE_REL_IA64_ADDEND                                        uint32   = 0x1f
	IMAGE_REL_CEF_ABSOLUTE                                       uint32   = 0x0
	IMAGE_REL_CEF_ADDR32                                         uint32   = 0x1
	IMAGE_REL_CEF_ADDR64                                         uint32   = 0x2
	IMAGE_REL_CEF_ADDR32NB                                       uint32   = 0x3
	IMAGE_REL_CEF_SECTION                                        uint32   = 0x4
	IMAGE_REL_CEF_SECREL                                         uint32   = 0x5
	IMAGE_REL_CEF_TOKEN                                          uint32   = 0x6
	IMAGE_REL_CEE_ABSOLUTE                                       uint32   = 0x0
	IMAGE_REL_CEE_ADDR32                                         uint32   = 0x1
	IMAGE_REL_CEE_ADDR64                                         uint32   = 0x2
	IMAGE_REL_CEE_ADDR32NB                                       uint32   = 0x3
	IMAGE_REL_CEE_SECTION                                        uint32   = 0x4
	IMAGE_REL_CEE_SECREL                                         uint32   = 0x5
	IMAGE_REL_CEE_TOKEN                                          uint32   = 0x6
	IMAGE_REL_M32R_ABSOLUTE                                      uint32   = 0x0
	IMAGE_REL_M32R_ADDR32                                        uint32   = 0x1
	IMAGE_REL_M32R_ADDR32NB                                      uint32   = 0x2
	IMAGE_REL_M32R_ADDR24                                        uint32   = 0x3
	IMAGE_REL_M32R_GPREL16                                       uint32   = 0x4
	IMAGE_REL_M32R_PCREL24                                       uint32   = 0x5
	IMAGE_REL_M32R_PCREL16                                       uint32   = 0x6
	IMAGE_REL_M32R_PCREL8                                        uint32   = 0x7
	IMAGE_REL_M32R_REFHALF                                       uint32   = 0x8
	IMAGE_REL_M32R_REFHI                                         uint32   = 0x9
	IMAGE_REL_M32R_REFLO                                         uint32   = 0xa
	IMAGE_REL_M32R_PAIR                                          uint32   = 0xb
	IMAGE_REL_M32R_SECTION                                       uint32   = 0xc
	IMAGE_REL_M32R_SECREL32                                      uint32   = 0xd
	IMAGE_REL_M32R_TOKEN                                         uint32   = 0xe
	IMAGE_REL_EBC_ABSOLUTE                                       uint32   = 0x0
	IMAGE_REL_EBC_ADDR32NB                                       uint32   = 0x1
	IMAGE_REL_EBC_REL32                                          uint32   = 0x2
	IMAGE_REL_EBC_SECTION                                        uint32   = 0x3
	IMAGE_REL_EBC_SECREL                                         uint32   = 0x4
	EMARCH_ENC_I17_IMM7B_INST_WORD_X                             uint32   = 0x3
	EMARCH_ENC_I17_IMM7B_SIZE_X                                  uint32   = 0x7
	EMARCH_ENC_I17_IMM7B_INST_WORD_POS_X                         uint32   = 0x4
	EMARCH_ENC_I17_IMM7B_VAL_POS_X                               uint32   = 0x0
	EMARCH_ENC_I17_IMM9D_INST_WORD_X                             uint32   = 0x3
	EMARCH_ENC_I17_IMM9D_SIZE_X                                  uint32   = 0x9
	EMARCH_ENC_I17_IMM9D_INST_WORD_POS_X                         uint32   = 0x12
	EMARCH_ENC_I17_IMM9D_VAL_POS_X                               uint32   = 0x7
	EMARCH_ENC_I17_IMM5C_INST_WORD_X                             uint32   = 0x3
	EMARCH_ENC_I17_IMM5C_SIZE_X                                  uint32   = 0x5
	EMARCH_ENC_I17_IMM5C_INST_WORD_POS_X                         uint32   = 0xd
	EMARCH_ENC_I17_IMM5C_VAL_POS_X                               uint32   = 0x10
	EMARCH_ENC_I17_IC_INST_WORD_X                                uint32   = 0x3
	EMARCH_ENC_I17_IC_SIZE_X                                     uint32   = 0x1
	EMARCH_ENC_I17_IC_INST_WORD_POS_X                            uint32   = 0xc
	EMARCH_ENC_I17_IC_VAL_POS_X                                  uint32   = 0x15
	EMARCH_ENC_I17_IMM41a_INST_WORD_X                            uint32   = 0x1
	EMARCH_ENC_I17_IMM41a_SIZE_X                                 uint32   = 0xa
	EMARCH_ENC_I17_IMM41a_INST_WORD_POS_X                        uint32   = 0xe
	EMARCH_ENC_I17_IMM41a_VAL_POS_X                              uint32   = 0x16
	EMARCH_ENC_I17_IMM41b_INST_WORD_X                            uint32   = 0x1
	EMARCH_ENC_I17_IMM41b_SIZE_X                                 uint32   = 0x8
	EMARCH_ENC_I17_IMM41b_INST_WORD_POS_X                        uint32   = 0x18
	EMARCH_ENC_I17_IMM41b_VAL_POS_X                              uint32   = 0x20
	EMARCH_ENC_I17_IMM41c_INST_WORD_X                            uint32   = 0x2
	EMARCH_ENC_I17_IMM41c_SIZE_X                                 uint32   = 0x17
	EMARCH_ENC_I17_IMM41c_INST_WORD_POS_X                        uint32   = 0x0
	EMARCH_ENC_I17_IMM41c_VAL_POS_X                              uint32   = 0x28
	EMARCH_ENC_I17_SIGN_INST_WORD_X                              uint32   = 0x3
	EMARCH_ENC_I17_SIGN_SIZE_X                                   uint32   = 0x1
	EMARCH_ENC_I17_SIGN_INST_WORD_POS_X                          uint32   = 0x1b
	EMARCH_ENC_I17_SIGN_VAL_POS_X                                uint32   = 0x3f
	X3_OPCODE_INST_WORD_X                                        uint32   = 0x3
	X3_OPCODE_SIZE_X                                             uint32   = 0x4
	X3_OPCODE_INST_WORD_POS_X                                    uint32   = 0x1c
	X3_OPCODE_SIGN_VAL_POS_X                                     uint32   = 0x0
	X3_I_INST_WORD_X                                             uint32   = 0x3
	X3_I_SIZE_X                                                  uint32   = 0x1
	X3_I_INST_WORD_POS_X                                         uint32   = 0x1b
	X3_I_SIGN_VAL_POS_X                                          uint32   = 0x3b
	X3_D_WH_INST_WORD_X                                          uint32   = 0x3
	X3_D_WH_SIZE_X                                               uint32   = 0x3
	X3_D_WH_INST_WORD_POS_X                                      uint32   = 0x18
	X3_D_WH_SIGN_VAL_POS_X                                       uint32   = 0x0
	X3_IMM20_INST_WORD_X                                         uint32   = 0x3
	X3_IMM20_SIZE_X                                              uint32   = 0x14
	X3_IMM20_INST_WORD_POS_X                                     uint32   = 0x4
	X3_IMM20_SIGN_VAL_POS_X                                      uint32   = 0x0
	X3_IMM39_1_INST_WORD_X                                       uint32   = 0x2
	X3_IMM39_1_SIZE_X                                            uint32   = 0x17
	X3_IMM39_1_INST_WORD_POS_X                                   uint32   = 0x0
	X3_IMM39_1_SIGN_VAL_POS_X                                    uint32   = 0x24
	X3_IMM39_2_INST_WORD_X                                       uint32   = 0x1
	X3_IMM39_2_SIZE_X                                            uint32   = 0x10
	X3_IMM39_2_INST_WORD_POS_X                                   uint32   = 0x10
	X3_IMM39_2_SIGN_VAL_POS_X                                    uint32   = 0x14
	X3_P_INST_WORD_X                                             uint32   = 0x3
	X3_P_SIZE_X                                                  uint32   = 0x4
	X3_P_INST_WORD_POS_X                                         uint32   = 0x0
	X3_P_SIGN_VAL_POS_X                                          uint32   = 0x0
	X3_TMPLT_INST_WORD_X                                         uint32   = 0x0
	X3_TMPLT_SIZE_X                                              uint32   = 0x4
	X3_TMPLT_INST_WORD_POS_X                                     uint32   = 0x0
	X3_TMPLT_SIGN_VAL_POS_X                                      uint32   = 0x0
	X3_BTYPE_QP_INST_WORD_X                                      uint32   = 0x2
	X3_BTYPE_QP_SIZE_X                                           uint32   = 0x9
	X3_BTYPE_QP_INST_WORD_POS_X                                  uint32   = 0x17
	X3_BTYPE_QP_INST_VAL_POS_X                                   uint32   = 0x0
	X3_EMPTY_INST_WORD_X                                         uint32   = 0x1
	X3_EMPTY_SIZE_X                                              uint32   = 0x2
	X3_EMPTY_INST_WORD_POS_X                                     uint32   = 0xe
	X3_EMPTY_INST_VAL_POS_X                                      uint32   = 0x0
	IMAGE_REL_BASED_ABSOLUTE                                     uint32   = 0x0
	IMAGE_REL_BASED_HIGH                                         uint32   = 0x1
	IMAGE_REL_BASED_LOW                                          uint32   = 0x2
	IMAGE_REL_BASED_HIGHLOW                                      uint32   = 0x3
	IMAGE_REL_BASED_HIGHADJ                                      uint32   = 0x4
	IMAGE_REL_BASED_MACHINE_SPECIFIC_5                           uint32   = 0x5
	IMAGE_REL_BASED_RESERVED                                     uint32   = 0x6
	IMAGE_REL_BASED_MACHINE_SPECIFIC_7                           uint32   = 0x7
	IMAGE_REL_BASED_MACHINE_SPECIFIC_8                           uint32   = 0x8
	IMAGE_REL_BASED_MACHINE_SPECIFIC_9                           uint32   = 0x9
	IMAGE_REL_BASED_DIR64                                        uint32   = 0xa
	IMAGE_REL_BASED_IA64_IMM64                                   uint32   = 0x9
	IMAGE_REL_BASED_MIPS_JMPADDR                                 uint32   = 0x5
	IMAGE_REL_BASED_MIPS_JMPADDR16                               uint32   = 0x9
	IMAGE_REL_BASED_ARM_MOV32                                    uint32   = 0x5
	IMAGE_REL_BASED_THUMB_MOV32                                  uint32   = 0x7
	IMAGE_ARCHIVE_START_SIZE                                     uint32   = 0x8
	IMAGE_ARCHIVE_START                                          string   = "!<arch>\n"
	IMAGE_ARCHIVE_END                                            string   = "`\n"
	IMAGE_ARCHIVE_PAD                                            string   = "\n"
	IMAGE_ARCHIVE_LINKER_MEMBER                                  string   = "/               "
	IMAGE_ARCHIVE_LONGNAMES_MEMBER                               string   = "//              "
	IMAGE_ARCHIVE_HYBRIDMAP_MEMBER                               string   = "/<HYBRIDMAP>/   "
	IMAGE_SIZEOF_ARCHIVE_MEMBER_HDR                              uint32   = 0x3c
	IMAGE_ORDINAL_FLAG64                                         uint64   = 0x8000000000000000
	IMAGE_ORDINAL_FLAG32                                         uint32   = 0x80000000
	IMAGE_ORDINAL_FLAG                                           uint64   = 0x8000000000000000
	IMAGE_RESOURCE_NAME_IS_STRING                                uint32   = 0x80000000
	IMAGE_RESOURCE_DATA_IS_DIRECTORY                             uint32   = 0x80000000
	IMAGE_DYNAMIC_RELOCATION_GUARD_RF_PROLOGUE                   uint32   = 0x1
	IMAGE_DYNAMIC_RELOCATION_GUARD_RF_EPILOGUE                   uint32   = 0x2
	IMAGE_DYNAMIC_RELOCATION_GUARD_IMPORT_CONTROL_TRANSFER       uint32   = 0x3
	IMAGE_DYNAMIC_RELOCATION_GUARD_INDIR_CONTROL_TRANSFER        uint32   = 0x4
	IMAGE_DYNAMIC_RELOCATION_GUARD_SWITCHTABLE_BRANCH            uint32   = 0x5
	IMAGE_HOT_PATCH_BASE_OBLIGATORY                              uint32   = 0x1
	IMAGE_HOT_PATCH_BASE_CAN_ROLL_BACK                           uint32   = 0x2
	IMAGE_HOT_PATCH_CHUNK_INVERSE                                uint32   = 0x80000000
	IMAGE_HOT_PATCH_CHUNK_OBLIGATORY                             uint32   = 0x40000000
	IMAGE_HOT_PATCH_CHUNK_RESERVED                               uint32   = 0x3ff03000
	IMAGE_HOT_PATCH_CHUNK_TYPE                                   uint32   = 0xfc000
	IMAGE_HOT_PATCH_CHUNK_SOURCE_RVA                             uint32   = 0x8000
	IMAGE_HOT_PATCH_CHUNK_TARGET_RVA                             uint32   = 0x4000
	IMAGE_HOT_PATCH_CHUNK_SIZE                                   uint32   = 0xfff
	IMAGE_HOT_PATCH_NONE                                         uint32   = 0x0
	IMAGE_HOT_PATCH_FUNCTION                                     uint32   = 0x1c000
	IMAGE_HOT_PATCH_ABSOLUTE                                     uint32   = 0x2c000
	IMAGE_HOT_PATCH_REL32                                        uint32   = 0x3c000
	IMAGE_HOT_PATCH_CALL_TARGET                                  uint32   = 0x44000
	IMAGE_HOT_PATCH_INDIRECT                                     uint32   = 0x5c000
	IMAGE_HOT_PATCH_NO_CALL_TARGET                               uint32   = 0x64000
	IMAGE_HOT_PATCH_DYNAMIC_VALUE                                uint32   = 0x78000
	IMAGE_GUARD_CF_INSTRUMENTED                                  uint32   = 0x100
	IMAGE_GUARD_CFW_INSTRUMENTED                                 uint32   = 0x200
	IMAGE_GUARD_CF_FUNCTION_TABLE_PRESENT                        uint32   = 0x400
	IMAGE_GUARD_SECURITY_COOKIE_UNUSED                           uint32   = 0x800
	IMAGE_GUARD_PROTECT_DELAYLOAD_IAT                            uint32   = 0x1000
	IMAGE_GUARD_DELAYLOAD_IAT_IN_ITS_OWN_SECTION                 uint32   = 0x2000
	IMAGE_GUARD_CF_EXPORT_SUPPRESSION_INFO_PRESENT               uint32   = 0x4000
	IMAGE_GUARD_CF_ENABLE_EXPORT_SUPPRESSION                     uint32   = 0x8000
	IMAGE_GUARD_CF_LONGJUMP_TABLE_PRESENT                        uint32   = 0x10000
	IMAGE_GUARD_RF_INSTRUMENTED                                  uint32   = 0x20000
	IMAGE_GUARD_RF_ENABLE                                        uint32   = 0x40000
	IMAGE_GUARD_RF_STRICT                                        uint32   = 0x80000
	IMAGE_GUARD_RETPOLINE_PRESENT                                uint32   = 0x100000
	IMAGE_GUARD_EH_CONTINUATION_TABLE_PRESENT                    uint32   = 0x400000
	IMAGE_GUARD_XFG_ENABLED                                      uint32   = 0x800000
	IMAGE_GUARD_CF_FUNCTION_TABLE_SIZE_MASK                      uint32   = 0xf0000000
	IMAGE_GUARD_CF_FUNCTION_TABLE_SIZE_SHIFT                     uint32   = 0x1c
	IMAGE_GUARD_FLAG_FID_SUPPRESSED                              uint32   = 0x1
	IMAGE_GUARD_FLAG_EXPORT_SUPPRESSED                           uint32   = 0x2
	IMAGE_GUARD_FLAG_FID_LANGEXCPTHANDLER                        uint32   = 0x4
	IMAGE_GUARD_FLAG_FID_XFG                                     uint32   = 0x8
	IMAGE_ENCLAVE_LONG_ID_LENGTH                                 uint32   = 0x20
	IMAGE_ENCLAVE_SHORT_ID_LENGTH                                uint32   = 0x10
	IMAGE_ENCLAVE_POLICY_DEBUGGABLE                              uint32   = 0x1
	IMAGE_ENCLAVE_FLAG_PRIMARY_IMAGE                             uint32   = 0x1
	IMAGE_ENCLAVE_IMPORT_MATCH_NONE                              uint32   = 0x0
	IMAGE_ENCLAVE_IMPORT_MATCH_UNIQUE_ID                         uint32   = 0x1
	IMAGE_ENCLAVE_IMPORT_MATCH_AUTHOR_ID                         uint32   = 0x2
	IMAGE_ENCLAVE_IMPORT_MATCH_FAMILY_ID                         uint32   = 0x3
	IMAGE_ENCLAVE_IMPORT_MATCH_IMAGE_ID                          uint32   = 0x4
	IMAGE_DEBUG_TYPE_OMAP_TO_SRC                                 uint32   = 0x7
	IMAGE_DEBUG_TYPE_OMAP_FROM_SRC                               uint32   = 0x8
	IMAGE_DEBUG_TYPE_RESERVED10                                  uint32   = 0xa
	IMAGE_DEBUG_TYPE_CLSID                                       uint32   = 0xb
	IMAGE_DEBUG_TYPE_VC_FEATURE                                  uint32   = 0xc
	IMAGE_DEBUG_TYPE_POGO                                        uint32   = 0xd
	IMAGE_DEBUG_TYPE_ILTCG                                       uint32   = 0xe
	IMAGE_DEBUG_TYPE_MPX                                         uint32   = 0xf
	IMAGE_DEBUG_TYPE_REPRO                                       uint32   = 0x10
	IMAGE_DEBUG_TYPE_EX_DLLCHARACTERISTICS                       uint32   = 0x14
	FRAME_FPO                                                    uint32   = 0x0
	FRAME_TRAP                                                   uint32   = 0x1
	FRAME_TSS                                                    uint32   = 0x2
	FRAME_NONFPO                                                 uint32   = 0x3
	SIZEOF_RFPO_DATA                                             uint32   = 0x10
	IMAGE_DEBUG_MISC_EXENAME                                     uint32   = 0x1
	IMAGE_SEPARATE_DEBUG_SIGNATURE                               uint32   = 0x4944
	NON_PAGED_DEBUG_SIGNATURE                                    uint32   = 0x494e
	IMAGE_SEPARATE_DEBUG_FLAGS_MASK                              uint32   = 0x8000
	IMAGE_SEPARATE_DEBUG_MISMATCH                                uint32   = 0x8000
	IMPORT_OBJECT_HDR_SIG2                                       uint32   = 0xffff
	UNWIND_HISTORY_TABLE_SIZE                                    uint32   = 0xc
	RTL_RUN_ONCE_CHECK_ONLY                                      uint32   = 0x1
	RTL_RUN_ONCE_ASYNC                                           uint32   = 0x2
	RTL_RUN_ONCE_INIT_FAILED                                     uint32   = 0x4
	RTL_RUN_ONCE_CTX_RESERVED_BITS                               uint32   = 0x2
	FAST_FAIL_LEGACY_GS_VIOLATION                                uint32   = 0x0
	FAST_FAIL_VTGUARD_CHECK_FAILURE                              uint32   = 0x1
	FAST_FAIL_STACK_COOKIE_CHECK_FAILURE                         uint32   = 0x2
	FAST_FAIL_CORRUPT_LIST_ENTRY                                 uint32   = 0x3
	FAST_FAIL_INCORRECT_STACK                                    uint32   = 0x4
	FAST_FAIL_INVALID_ARG                                        uint32   = 0x5
	FAST_FAIL_GS_COOKIE_INIT                                     uint32   = 0x6
	FAST_FAIL_FATAL_APP_EXIT                                     uint32   = 0x7
	FAST_FAIL_RANGE_CHECK_FAILURE                                uint32   = 0x8
	FAST_FAIL_UNSAFE_REGISTRY_ACCESS                             uint32   = 0x9
	FAST_FAIL_GUARD_ICALL_CHECK_FAILURE                          uint32   = 0xa
	FAST_FAIL_GUARD_WRITE_CHECK_FAILURE                          uint32   = 0xb
	FAST_FAIL_INVALID_FIBER_SWITCH                               uint32   = 0xc
	FAST_FAIL_INVALID_SET_OF_CONTEXT                             uint32   = 0xd
	FAST_FAIL_INVALID_REFERENCE_COUNT                            uint32   = 0xe
	FAST_FAIL_INVALID_JUMP_BUFFER                                uint32   = 0x12
	FAST_FAIL_MRDATA_MODIFIED                                    uint32   = 0x13
	FAST_FAIL_CERTIFICATION_FAILURE                              uint32   = 0x14
	FAST_FAIL_INVALID_EXCEPTION_CHAIN                            uint32   = 0x15
	FAST_FAIL_CRYPTO_LIBRARY                                     uint32   = 0x16
	FAST_FAIL_INVALID_CALL_IN_DLL_CALLOUT                        uint32   = 0x17
	FAST_FAIL_INVALID_IMAGE_BASE                                 uint32   = 0x18
	FAST_FAIL_DLOAD_PROTECTION_FAILURE                           uint32   = 0x19
	FAST_FAIL_UNSAFE_EXTENSION_CALL                              uint32   = 0x1a
	FAST_FAIL_DEPRECATED_SERVICE_INVOKED                         uint32   = 0x1b
	FAST_FAIL_INVALID_BUFFER_ACCESS                              uint32   = 0x1c
	FAST_FAIL_INVALID_BALANCED_TREE                              uint32   = 0x1d
	FAST_FAIL_INVALID_NEXT_THREAD                                uint32   = 0x1e
	FAST_FAIL_GUARD_ICALL_CHECK_SUPPRESSED                       uint32   = 0x1f
	FAST_FAIL_APCS_DISABLED                                      uint32   = 0x20
	FAST_FAIL_INVALID_IDLE_STATE                                 uint32   = 0x21
	FAST_FAIL_MRDATA_PROTECTION_FAILURE                          uint32   = 0x22
	FAST_FAIL_UNEXPECTED_HEAP_EXCEPTION                          uint32   = 0x23
	FAST_FAIL_INVALID_LOCK_STATE                                 uint32   = 0x24
	FAST_FAIL_GUARD_JUMPTABLE                                    uint32   = 0x25
	FAST_FAIL_INVALID_LONGJUMP_TARGET                            uint32   = 0x26
	FAST_FAIL_INVALID_DISPATCH_CONTEXT                           uint32   = 0x27
	FAST_FAIL_INVALID_THREAD                                     uint32   = 0x28
	FAST_FAIL_INVALID_SYSCALL_NUMBER                             uint32   = 0x29
	FAST_FAIL_INVALID_FILE_OPERATION                             uint32   = 0x2a
	FAST_FAIL_LPAC_ACCESS_DENIED                                 uint32   = 0x2b
	FAST_FAIL_GUARD_SS_FAILURE                                   uint32   = 0x2c
	FAST_FAIL_LOADER_CONTINUITY_FAILURE                          uint32   = 0x2d
	FAST_FAIL_GUARD_EXPORT_SUPPRESSION_FAILURE                   uint32   = 0x2e
	FAST_FAIL_INVALID_CONTROL_STACK                              uint32   = 0x2f
	FAST_FAIL_SET_CONTEXT_DENIED                                 uint32   = 0x30
	FAST_FAIL_INVALID_IAT                                        uint32   = 0x31
	FAST_FAIL_HEAP_METADATA_CORRUPTION                           uint32   = 0x32
	FAST_FAIL_PAYLOAD_RESTRICTION_VIOLATION                      uint32   = 0x33
	FAST_FAIL_LOW_LABEL_ACCESS_DENIED                            uint32   = 0x34
	FAST_FAIL_ENCLAVE_CALL_FAILURE                               uint32   = 0x35
	FAST_FAIL_UNHANDLED_LSS_EXCEPTON                             uint32   = 0x36
	FAST_FAIL_ADMINLESS_ACCESS_DENIED                            uint32   = 0x37
	FAST_FAIL_UNEXPECTED_CALL                                    uint32   = 0x38
	FAST_FAIL_CONTROL_INVALID_RETURN_ADDRESS                     uint32   = 0x39
	FAST_FAIL_UNEXPECTED_HOST_BEHAVIOR                           uint32   = 0x3a
	FAST_FAIL_FLAGS_CORRUPTION                                   uint32   = 0x3b
	FAST_FAIL_VEH_CORRUPTION                                     uint32   = 0x3c
	FAST_FAIL_ETW_CORRUPTION                                     uint32   = 0x3d
	FAST_FAIL_RIO_ABORT                                          uint32   = 0x3e
	FAST_FAIL_INVALID_PFN                                        uint32   = 0x3f
	FAST_FAIL_GUARD_ICALL_CHECK_FAILURE_XFG                      uint32   = 0x40
	FAST_FAIL_CAST_GUARD                                         uint32   = 0x41
	FAST_FAIL_HOST_VISIBILITY_CHANGE                             uint32   = 0x42
	FAST_FAIL_KERNEL_CET_SHADOW_STACK_ASSIST                     uint32   = 0x43
	FAST_FAIL_PATCH_CALLBACK_FAILED                              uint32   = 0x44
	FAST_FAIL_NTDLL_PATCH_FAILED                                 uint32   = 0x45
	FAST_FAIL_INVALID_FLS_DATA                                   uint32   = 0x46
	FAST_FAIL_INVALID_FAST_FAIL_CODE                             uint32   = 0xffffffff
	IS_TEXT_UNICODE_DBCS_LEADBYTE                                uint32   = 0x400
	IS_TEXT_UNICODE_UTF8                                         uint32   = 0x800
	COMPRESSION_FORMAT_NONE                                      uint32   = 0x0
	COMPRESSION_FORMAT_DEFAULT                                   uint32   = 0x1
	COMPRESSION_FORMAT_LZNT1                                     uint32   = 0x2
	COMPRESSION_FORMAT_XPRESS                                    uint32   = 0x3
	COMPRESSION_FORMAT_XPRESS_HUFF                               uint32   = 0x4
	COMPRESSION_FORMAT_XP10                                      uint32   = 0x5
	COMPRESSION_ENGINE_STANDARD                                  uint32   = 0x0
	COMPRESSION_ENGINE_MAXIMUM                                   uint32   = 0x100
	COMPRESSION_ENGINE_HIBER                                     uint32   = 0x200
	SEF_AI_USE_EXTRA_PARAMS                                      uint32   = 0x800
	SEF_FORCE_USER_MODE                                          uint32   = 0x2000
	MESSAGE_RESOURCE_UNICODE                                     uint32   = 0x1
	MESSAGE_RESOURCE_UTF8                                        uint32   = 0x2
	VER_EQUAL                                                    uint32   = 0x1
	VER_GREATER                                                  uint32   = 0x2
	VER_GREATER_EQUAL                                            uint32   = 0x3
	VER_LESS                                                     uint32   = 0x4
	VER_LESS_EQUAL                                               uint32   = 0x5
	VER_AND                                                      uint32   = 0x6
	VER_OR                                                       uint32   = 0x7
	VER_CONDITION_MASK                                           uint32   = 0x7
	VER_NUM_BITS_PER_CONDITION_MASK                              uint32   = 0x3
	VER_NT_WORKSTATION                                           uint32   = 0x1
	VER_NT_DOMAIN_CONTROLLER                                     uint32   = 0x2
	VER_NT_SERVER                                                uint32   = 0x3
	RTL_UMS_VERSION                                              uint32   = 0x100
	VRL_PREDEFINED_CLASS_BEGIN                                   uint32   = 0x1
	VRL_CUSTOM_CLASS_BEGIN                                       uint32   = 0x100
	VRL_ENABLE_KERNEL_BREAKS                                     uint32   = 0x80000000
	CTMF_INCLUDE_APPCONTAINER                                    uint32   = 0x1
	CTMF_INCLUDE_LPAC                                            uint32   = 0x2
	FLUSH_NV_MEMORY_IN_FLAG_NO_DRAIN                             uint32   = 0x1
	WRITE_NV_MEMORY_FLAG_FLUSH                                   uint32   = 0x1
	WRITE_NV_MEMORY_FLAG_NON_TEMPORAL                            uint32   = 0x2
	WRITE_NV_MEMORY_FLAG_NO_DRAIN                                uint32   = 0x100
	FILL_NV_MEMORY_FLAG_FLUSH                                    uint32   = 0x1
	FILL_NV_MEMORY_FLAG_NON_TEMPORAL                             uint32   = 0x2
	FILL_NV_MEMORY_FLAG_NO_DRAIN                                 uint32   = 0x100
	IMAGE_POLICY_METADATA_VERSION                                uint32   = 0x1
	IMAGE_POLICY_SECTION_NAME                                    string   = ".tPolicy"
	RTL_VIRTUAL_UNWIND2_VALIDATE_PAC                             uint32   = 0x1
	RTL_CRITICAL_SECTION_FLAG_NO_DEBUG_INFO                      uint32   = 0x1000000
	RTL_CRITICAL_SECTION_FLAG_DYNAMIC_SPIN                       uint32   = 0x2000000
	RTL_CRITICAL_SECTION_FLAG_STATIC_INIT                        uint32   = 0x4000000
	RTL_CRITICAL_SECTION_FLAG_RESOURCE_TYPE                      uint32   = 0x8000000
	RTL_CRITICAL_SECTION_FLAG_FORCE_DEBUG_INFO                   uint32   = 0x10000000
	RTL_CRITICAL_SECTION_ALL_FLAG_BITS                           uint32   = 0xff000000
	RTL_CRITICAL_SECTION_DEBUG_FLAG_STATIC_INIT                  uint32   = 0x1
	RTL_CONDITION_VARIABLE_LOCKMODE_SHARED                       uint32   = 0x1
	HEAP_OPTIMIZE_RESOURCES_CURRENT_VERSION                      uint32   = 0x1
	WT_EXECUTEINUITHREAD                                         uint32   = 0x2
	WT_EXECUTEINPERSISTENTIOTHREAD                               uint32   = 0x40
	WT_EXECUTEINLONGTHREAD                                       uint32   = 0x10
	WT_EXECUTEDELETEWAIT                                         uint32   = 0x8
	ACTIVATION_CONTEXT_PATH_TYPE_NONE                            uint32   = 0x1
	ACTIVATION_CONTEXT_PATH_TYPE_WIN32_FILE                      uint32   = 0x2
	ACTIVATION_CONTEXT_PATH_TYPE_URL                             uint32   = 0x3
	ACTIVATION_CONTEXT_PATH_TYPE_ASSEMBLYREF                     uint32   = 0x4
	CREATE_BOUNDARY_DESCRIPTOR_ADD_APPCONTAINER_SID              uint32   = 0x1
	PERFORMANCE_DATA_VERSION                                     uint32   = 0x1
	READ_THREAD_PROFILING_FLAG_DISPATCHING                       uint32   = 0x1
	READ_THREAD_PROFILING_FLAG_HARDWARE_COUNTERS                 uint32   = 0x2
	UNIFIEDBUILDREVISION_KEY                                     string   = "\\Registry\\Machine\\Software\\Microsoft\\Windows NT\\CurrentVersion"
	UNIFIEDBUILDREVISION_VALUE                                   string   = "UBR"
	UNIFIEDBUILDREVISION_MIN                                     uint32   = 0x0
	DEVICEFAMILYDEVICEFORM_KEY                                   string   = "\\Registry\\Machine\\Software\\Microsoft\\Windows NT\\CurrentVersion\\OEM"
	DEVICEFAMILYDEVICEFORM_VALUE                                 string   = "DeviceForm"
	DLL_PROCESS_ATTACH                                           uint32   = 0x1
	DLL_THREAD_ATTACH                                            uint32   = 0x2
	DLL_THREAD_DETACH                                            uint32   = 0x3
	DLL_PROCESS_DETACH                                           uint32   = 0x0
	EVENTLOG_FORWARDS_READ                                       uint32   = 0x4
	EVENTLOG_BACKWARDS_READ                                      uint32   = 0x8
	EVENTLOG_START_PAIRED_EVENT                                  uint32   = 0x1
	EVENTLOG_END_PAIRED_EVENT                                    uint32   = 0x2
	EVENTLOG_END_ALL_PAIRED_EVENTS                               uint32   = 0x4
	EVENTLOG_PAIRED_EVENT_ACTIVE                                 uint32   = 0x8
	EVENTLOG_PAIRED_EVENT_INACTIVE                               uint32   = 0x10
	MAXLOGICALLOGNAMESIZE                                        uint32   = 0x100
	REG_REFRESH_HIVE                                             int32    = 2
	REG_NO_LAZY_FLUSH                                            int32    = 4
	REG_APP_HIVE                                                 int32    = 16
	REG_PROCESS_PRIVATE                                          int32    = 32
	REG_START_JOURNAL                                            int32    = 64
	REG_HIVE_EXACT_FILE_GROWTH                                   int32    = 128
	REG_HIVE_NO_RM                                               int32    = 256
	REG_HIVE_SINGLE_LOG                                          int32    = 512
	REG_BOOT_HIVE                                                int32    = 1024
	REG_LOAD_HIVE_OPEN_HANDLE                                    int32    = 2048
	REG_FLUSH_HIVE_FILE_GROWTH                                   int32    = 4096
	REG_OPEN_READ_ONLY                                           int32    = 8192
	REG_IMMUTABLE                                                int32    = 16384
	REG_NO_IMPERSONATION_FALLBACK                                int32    = 32768
	REG_APP_HIVE_OPEN_READ_ONLY                                  int32    = 8192
	REG_FORCE_UNLOAD                                             uint32   = 0x1
	REG_UNLOAD_LEGAL_FLAGS                                       uint32   = 0x1
	SERVICE_USER_SERVICE                                         uint32   = 0x40
	SERVICE_USERSERVICE_INSTANCE                                 uint32   = 0x80
	SERVICE_INTERACTIVE_PROCESS                                  uint32   = 0x100
	SERVICE_PKG_SERVICE                                          uint32   = 0x200
	CM_SERVICE_NETWORK_BOOT_LOAD                                 uint32   = 0x1
	CM_SERVICE_VIRTUAL_DISK_BOOT_LOAD                            uint32   = 0x2
	CM_SERVICE_USB_DISK_BOOT_LOAD                                uint32   = 0x4
	CM_SERVICE_SD_DISK_BOOT_LOAD                                 uint32   = 0x8
	CM_SERVICE_USB3_DISK_BOOT_LOAD                               uint32   = 0x10
	CM_SERVICE_MEASURED_BOOT_LOAD                                uint32   = 0x20
	CM_SERVICE_VERIFIER_BOOT_LOAD                                uint32   = 0x40
	CM_SERVICE_WINPE_BOOT_LOAD                                   uint32   = 0x80
	CM_SERVICE_RAM_DISK_BOOT_LOAD                                uint32   = 0x100
	TAPE_PSEUDO_LOGICAL_POSITION                                 int32    = 2
	TAPE_PSEUDO_LOGICAL_BLOCK                                    int32    = 3
	TAPE_DRIVE_FIXED                                             uint32   = 0x1
	TAPE_DRIVE_SELECT                                            uint32   = 0x2
	TAPE_DRIVE_INITIATOR                                         uint32   = 0x4
	TAPE_DRIVE_ERASE_SHORT                                       uint32   = 0x10
	TAPE_DRIVE_ERASE_LONG                                        uint32   = 0x20
	TAPE_DRIVE_ERASE_BOP_ONLY                                    uint32   = 0x40
	TAPE_DRIVE_ERASE_IMMEDIATE                                   uint32   = 0x80
	TAPE_DRIVE_TAPE_CAPACITY                                     uint32   = 0x100
	TAPE_DRIVE_TAPE_REMAINING                                    uint32   = 0x200
	TAPE_DRIVE_FIXED_BLOCK                                       uint32   = 0x400
	TAPE_DRIVE_VARIABLE_BLOCK                                    uint32   = 0x800
	TAPE_DRIVE_WRITE_PROTECT                                     uint32   = 0x1000
	TAPE_DRIVE_EOT_WZ_SIZE                                       uint32   = 0x2000
	TAPE_DRIVE_ECC                                               uint32   = 0x10000
	TAPE_DRIVE_COMPRESSION                                       uint32   = 0x20000
	TAPE_DRIVE_PADDING                                           uint32   = 0x40000
	TAPE_DRIVE_REPORT_SMKS                                       uint32   = 0x80000
	TAPE_DRIVE_GET_ABSOLUTE_BLK                                  uint32   = 0x100000
	TAPE_DRIVE_GET_LOGICAL_BLK                                   uint32   = 0x200000
	TAPE_DRIVE_SET_EOT_WZ_SIZE                                   uint32   = 0x400000
	TAPE_DRIVE_EJECT_MEDIA                                       uint32   = 0x1000000
	TAPE_DRIVE_CLEAN_REQUESTS                                    uint32   = 0x2000000
	TAPE_DRIVE_SET_CMP_BOP_ONLY                                  uint32   = 0x4000000
	TAPE_DRIVE_RESERVED_BIT                                      uint32   = 0x80000000
	TAPE_DRIVE_FORMAT                                            uint32   = 0xa0000000
	TAPE_DRIVE_FORMAT_IMMEDIATE                                  uint32   = 0xc0000000
	TAPE_DRIVE_HIGH_FEATURES                                     uint32   = 0x80000000
	TAPE_QUERY_DRIVE_PARAMETERS                                  int32    = 0
	TAPE_QUERY_MEDIA_CAPACITY                                    int32    = 1
	TAPE_CHECK_FOR_DRIVE_PROBLEM                                 int32    = 2
	TAPE_QUERY_IO_ERROR_DATA                                     int32    = 3
	TAPE_QUERY_DEVICE_ERROR_DATA                                 int32    = 4
	TRANSACTIONMANAGER_QUERY_INFORMATION                         uint32   = 0x1
	TRANSACTIONMANAGER_SET_INFORMATION                           uint32   = 0x2
	TRANSACTIONMANAGER_RECOVER                                   uint32   = 0x4
	TRANSACTIONMANAGER_RENAME                                    uint32   = 0x8
	TRANSACTIONMANAGER_CREATE_RM                                 uint32   = 0x10
	TRANSACTIONMANAGER_BIND_TRANSACTION                          uint32   = 0x20
	TRANSACTION_QUERY_INFORMATION                                uint32   = 0x1
	TRANSACTION_SET_INFORMATION                                  uint32   = 0x2
	TRANSACTION_ENLIST                                           uint32   = 0x4
	TRANSACTION_COMMIT                                           uint32   = 0x8
	TRANSACTION_ROLLBACK                                         uint32   = 0x10
	TRANSACTION_PROPAGATE                                        uint32   = 0x20
	TRANSACTION_RIGHT_RESERVED1                                  uint32   = 0x40
	RESOURCEMANAGER_QUERY_INFORMATION                            uint32   = 0x1
	RESOURCEMANAGER_SET_INFORMATION                              uint32   = 0x2
	RESOURCEMANAGER_RECOVER                                      uint32   = 0x4
	RESOURCEMANAGER_ENLIST                                       uint32   = 0x8
	RESOURCEMANAGER_GET_NOTIFICATION                             uint32   = 0x10
	RESOURCEMANAGER_REGISTER_PROTOCOL                            uint32   = 0x20
	RESOURCEMANAGER_COMPLETE_PROPAGATION                         uint32   = 0x40
	ENLISTMENT_QUERY_INFORMATION                                 uint32   = 0x1
	ENLISTMENT_SET_INFORMATION                                   uint32   = 0x2
	ENLISTMENT_RECOVER                                           uint32   = 0x4
	ENLISTMENT_SUBORDINATE_RIGHTS                                uint32   = 0x8
	ENLISTMENT_SUPERIOR_RIGHTS                                   uint32   = 0x10
	PcTeb                                                        uint32   = 0x18
	ACTIVATION_CONTEXT_SECTION_ASSEMBLY_INFORMATION              uint32   = 0x1
	ACTIVATION_CONTEXT_SECTION_DLL_REDIRECTION                   uint32   = 0x2
	ACTIVATION_CONTEXT_SECTION_WINDOW_CLASS_REDIRECTION          uint32   = 0x3
	ACTIVATION_CONTEXT_SECTION_COM_SERVER_REDIRECTION            uint32   = 0x4
	ACTIVATION_CONTEXT_SECTION_COM_INTERFACE_REDIRECTION         uint32   = 0x5
	ACTIVATION_CONTEXT_SECTION_COM_TYPE_LIBRARY_REDIRECTION      uint32   = 0x6
	ACTIVATION_CONTEXT_SECTION_COM_PROGID_REDIRECTION            uint32   = 0x7
	ACTIVATION_CONTEXT_SECTION_GLOBAL_OBJECT_RENAME_TABLE        uint32   = 0x8
	ACTIVATION_CONTEXT_SECTION_CLR_SURROGATES                    uint32   = 0x9
	ACTIVATION_CONTEXT_SECTION_APPLICATION_SETTINGS              uint32   = 0xa
	ACTIVATION_CONTEXT_SECTION_COMPATIBILITY_INFO                uint32   = 0xb
	ACTIVATION_CONTEXT_SECTION_WINRT_ACTIVATABLE_CLASSES         uint32   = 0xc
	BSF_MSGSRV32ISOK                                             uint32   = 0x80000000
	BSF_MSGSRV32ISOK_BIT                                         uint32   = 0x1f
	DBT_APPYBEGIN                                                uint32   = 0x0
	DBT_APPYEND                                                  uint32   = 0x1
	DBT_DEVNODES_CHANGED                                         uint32   = 0x7
	DBT_QUERYCHANGECONFIG                                        uint32   = 0x17
	DBT_CONFIGCHANGED                                            uint32   = 0x18
	DBT_CONFIGCHANGECANCELED                                     uint32   = 0x19
	DBT_MONITORCHANGE                                            uint32   = 0x1b
	DBT_SHELLLOGGEDON                                            uint32   = 0x20
	DBT_CONFIGMGAPI32                                            uint32   = 0x22
	DBT_VXDINITCOMPLETE                                          uint32   = 0x23
	DBT_VOLLOCKQUERYLOCK                                         uint32   = 0x8041
	DBT_VOLLOCKLOCKTAKEN                                         uint32   = 0x8042
	DBT_VOLLOCKLOCKFAILED                                        uint32   = 0x8043
	DBT_VOLLOCKQUERYUNLOCK                                       uint32   = 0x8044
	DBT_VOLLOCKLOCKRELEASED                                      uint32   = 0x8045
	DBT_VOLLOCKUNLOCKFAILED                                      uint32   = 0x8046
	LOCKP_ALLOW_WRITES                                           uint32   = 0x1
	LOCKP_FAIL_WRITES                                            uint32   = 0x0
	LOCKP_FAIL_MEM_MAPPING                                       uint32   = 0x2
	LOCKP_ALLOW_MEM_MAPPING                                      uint32   = 0x0
	LOCKP_USER_MASK                                              uint32   = 0x3
	LOCKP_LOCK_FOR_FORMAT                                        uint32   = 0x4
	LOCKF_LOGICAL_LOCK                                           uint32   = 0x0
	LOCKF_PHYSICAL_LOCK                                          uint32   = 0x1
	DBT_NO_DISK_SPACE                                            uint32   = 0x47
	DBT_LOW_DISK_SPACE                                           uint32   = 0x48
	DBT_CONFIGMGPRIVATE                                          uint32   = 0x7fff
	DBT_DEVICEARRIVAL                                            uint32   = 0x8000
	DBT_DEVICEQUERYREMOVE                                        uint32   = 0x8001
	DBT_DEVICEQUERYREMOVEFAILED                                  uint32   = 0x8002
	DBT_DEVICEREMOVEPENDING                                      uint32   = 0x8003
	DBT_DEVICEREMOVECOMPLETE                                     uint32   = 0x8004
	DBT_DEVICETYPESPECIFIC                                       uint32   = 0x8005
	DBT_CUSTOMEVENT                                              uint32   = 0x8006
	DBT_DEVTYP_DEVNODE                                           uint32   = 0x1
	DBT_DEVTYP_NET                                               uint32   = 0x4
	DBTF_RESOURCE                                                uint32   = 0x1
	DBTF_XPORT                                                   uint32   = 0x2
	DBTF_SLOWNET                                                 uint32   = 0x4
	DBT_VPOWERDAPI                                               uint32   = 0x8100
	DBT_USERDEFINED                                              uint32   = 0xffff
	D3DNTHAL_NUMCLIPVERTICES                                     uint32   = 0x14
	D3DNTHAL_SCENE_CAPTURE_START                                 int32    = 0
	D3DNTHAL_SCENE_CAPTURE_END                                   int32    = 1
	D3DNTHAL_CONTEXT_BAD                                         int64    = 512
	D3DNTHAL_OUTOFCONTEXTS                                       int64    = 513
	D3DNTHAL2_CB32_SETRENDERTARGET                               int32    = 1
	D3DHAL_STATESETBEGIN                                         uint32   = 0x0
	D3DHAL_STATESETEND                                           uint32   = 0x1
	D3DHAL_STATESETDELETE                                        uint32   = 0x2
	D3DHAL_STATESETEXECUTE                                       uint32   = 0x3
	D3DHAL_STATESETCAPTURE                                       uint32   = 0x4
	D3DNTHALDP2_USERMEMVERTICES                                  int32    = 1
	D3DNTHALDP2_EXECUTEBUFFER                                    int32    = 2
	D3DNTHALDP2_SWAPVERTEXBUFFER                                 int32    = 4
	D3DNTHALDP2_SWAPCOMMANDBUFFER                                int32    = 8
	D3DNTHALDP2_REQVERTEXBUFSIZE                                 int32    = 16
	D3DNTHALDP2_REQCOMMANDBUFSIZE                                int32    = 32
	D3DNTHALDP2_VIDMEMVERTEXBUF                                  int32    = 64
	D3DNTHALDP2_VIDMEMCOMMANDBUF                                 int32    = 128
	D3DNTHAL3_CB32_CLEAR2                                        int32    = 1
	D3DNTHAL3_CB32_RESERVED                                      int32    = 2
	D3DNTHAL3_CB32_VALIDATETEXTURESTAGESTATE                     int32    = 4
	D3DNTHAL3_CB32_DRAWPRIMITIVES2                               int32    = 8
	D3DNTHAL_TSS_RENDERSTATEBASE                                 uint32   = 0x100
	D3DNTHAL_TSS_MAXSTAGES                                       uint32   = 0x8
	D3DNTHAL_TSS_STATESPERSTAGE                                  uint32   = 0x40
	D3DTSS_TEXTUREMAP                                            uint32   = 0x0
	D3DHAL_SAMPLER_MAXSAMP                                       uint32   = 0x10
	D3DHAL_SAMPLER_MAXVERTEXSAMP                                 uint32   = 0x4
	D3DPMISCCAPS_LINEPATTERNREP                                  int32    = 4
	D3DRS_MAXVERTEXSHADERINST                                    uint32   = 0xc4
	D3DRS_MAXPIXELSHADERINST                                     uint32   = 0xc5
	D3DRENDERSTATE_EVICTMANAGEDTEXTURES                          uint32   = 0x3d
	D3DRENDERSTATE_SCENECAPTURE                                  uint32   = 0x3e
	NT_D3DRS_DELETERTPATCH_                                      uint32   = 0xa9
	D3DINFINITEINSTRUCTIONS                                      uint32   = 0xffffffff
	D3DNTHAL_STATESETCREATE                                      uint32   = 0x5
	D3DNTCLEAR_COMPUTERECTS                                      int32    = 8
	NT_RTPATCHFLAG_HASSEGS_                                      int32    = 1
	NT_RTPATCHFLAG_HASINFO_                                      int32    = 2
	D3DNTHAL_ROW_WEIGHTS                                         uint32   = 0x1
	D3DNTHAL_COL_WEIGHTS                                         uint32   = 0x2
	DP2BLT_POINT                                                 int32    = 1
	DP2BLT_LINEAR                                                int32    = 2
	DDBLT_EXTENDED_PRESENTATION_STRETCHFACTOR                    int32    = 16
	NT_D3DGDI2_MAGIC_                                            uint32   = 0xffffffff
	NT_D3DGDI2_TYPE_GETD3DCAPS8_                                 uint32   = 0x1
	NT_D3DGDI2_TYPE_GETFORMATCOUNT_                              uint32   = 0x2
	NT_D3DGDI2_TYPE_GETFORMAT_                                   uint32   = 0x3
	NT_D3DGDI2_TYPE_DXVERSION_                                   uint32   = 0x4
	NT_D3DGDI2_TYPE_DEFERRED_AGP_AWARE_                          uint32   = 0x18
	NT_D3DGDI2_TYPE_FREE_DEFERRED_AGP_                           uint32   = 0x19
	NT_D3DGDI2_TYPE_DEFER_AGP_FREES_                             uint32   = 0x20
	NT_D3DGDI2_TYPE_GETD3DCAPS9_                                 uint32   = 0x10
	NT_D3DGDI2_TYPE_GETEXTENDEDMODECOUNT_                        uint32   = 0x11
	NT_D3DGDI2_TYPE_GETEXTENDEDMODE_                             uint32   = 0x12
	NT_D3DGDI2_TYPE_GETADAPTERGROUP_                             uint32   = 0x13
	NT_D3DGDI2_TYPE_GETMULTISAMPLEQUALITYLEVELS_                 uint32   = 0x16
	NT_D3DGDI2_TYPE_GETD3DQUERYCOUNT_                            uint32   = 0x21
	NT_D3DGDI2_TYPE_GETD3DQUERY_                                 uint32   = 0x22
	NT_D3DGDI2_TYPE_GETDDIVERSION_                               uint32   = 0x23
	DX9_DDI_VERSION                                              uint32   = 0x4
	NT_D3DDEVCAPS_HWVERTEXBUFFER_                                int32    = 33554432
	NT_D3DDEVCAPS_HWINDEXBUFFER_                                 int32    = 67108864
	NT_D3DDEVCAPS_SUBVOLUMELOCK_                                 int32    = 134217728
	NT_D3DPMISCCAPS_FOGINFVF_                                    int32    = 8192
	NT_D3DFVF_FOG_                                               int32    = 8192
	D3DPRASTERCAPS_STRETCHBLTMULTISAMPLE                         int32    = 8388608
	NT_D3DVS_MAXINSTRUCTIONCOUNT_V1_1_                           uint32   = 0x80
	NT_D3DVS_LABEL_MAX_V3_0_                                     uint32   = 0x800
	NT_D3DVS_TCRDOUTREG_MAX_V1_1_                                uint32   = 0x8
	NT_D3DVS_TCRDOUTREG_MAX_V2_0_                                uint32   = 0x8
	NT_D3DVS_TCRDOUTREG_MAX_V2_1_                                uint32   = 0x8
	NT_D3DVS_OUTPUTREG_MAX_V3_0_                                 uint32   = 0xc
	NT_D3DVS_OUTPUTREG_MAX_SW_DX9_                               uint32   = 0x10
	NT_D3DVS_ATTROUTREG_MAX_V1_1_                                uint32   = 0x2
	NT_D3DVS_ATTROUTREG_MAX_V2_0_                                uint32   = 0x2
	NT_D3DVS_ATTROUTREG_MAX_V2_1_                                uint32   = 0x2
	NT_D3DVS_INPUTREG_MAX_V1_1_                                  uint32   = 0x10
	NT_D3DVS_INPUTREG_MAX_V2_0_                                  uint32   = 0x10
	NT_D3DVS_INPUTREG_MAX_V2_1_                                  uint32   = 0x10
	NT_D3DVS_INPUTREG_MAX_V3_0_                                  uint32   = 0x10
	NT_D3DVS_TEMPREG_MAX_V1_1_                                   uint32   = 0xc
	NT_D3DVS_TEMPREG_MAX_V2_0_                                   uint32   = 0xc
	NT_D3DVS_TEMPREG_MAX_V2_1_                                   uint32   = 0x20
	NT_D3DVS_TEMPREG_MAX_V3_0_                                   uint32   = 0x20
	NT_D3DVS_CONSTREG_MAX_V1_1_                                  uint32   = 0x60
	NT_D3DVS_CONSTREG_MAX_V2_0_                                  uint32   = 0x2000
	NT_D3DVS_CONSTREG_MAX_V2_1_                                  uint32   = 0x2000
	NT_D3DVS_CONSTREG_MAX_V3_0_                                  uint32   = 0x2000
	NT_D3DVS_CONSTINTREG_MAX_SW_DX9_                             uint32   = 0x800
	NT_D3DVS_CONSTINTREG_MAX_V2_0_                               uint32   = 0x10
	NT_D3DVS_CONSTINTREG_MAX_V2_1_                               uint32   = 0x10
	NT_D3DVS_CONSTINTREG_MAX_V3_0_                               uint32   = 0x10
	NT_D3DVS_CONSTBOOLREG_MAX_SW_DX9_                            uint32   = 0x800
	NT_D3DVS_CONSTBOOLREG_MAX_V2_0_                              uint32   = 0x10
	NT_D3DVS_CONSTBOOLREG_MAX_V2_1_                              uint32   = 0x10
	NT_D3DVS_CONSTBOOLREG_MAX_V3_0_                              uint32   = 0x10
	NT_D3DVS_ADDRREG_MAX_V1_1_                                   uint32   = 0x1
	NT_D3DVS_ADDRREG_MAX_V2_0_                                   uint32   = 0x1
	NT_D3DVS_ADDRREG_MAX_V2_1_                                   uint32   = 0x1
	NT_D3DVS_ADDRREG_MAX_V3_0_                                   uint32   = 0x1
	NT_D3DVS_MAXLOOPSTEP_V2_0_                                   uint32   = 0x80
	NT_D3DVS_MAXLOOPSTEP_V2_1_                                   uint32   = 0x80
	NT_D3DVS_MAXLOOPSTEP_V3_0_                                   uint32   = 0x80
	NT_D3DVS_MAXLOOPINITVALUE_V2_0_                              uint32   = 0xff
	NT_D3DVS_MAXLOOPINITVALUE_V2_1_                              uint32   = 0xff
	NT_D3DVS_MAXLOOPINITVALUE_V3_0_                              uint32   = 0xff
	NT_D3DVS_MAXLOOPITERATIONCOUNT_V2_0_                         uint32   = 0xff
	NT_D3DVS_MAXLOOPITERATIONCOUNT_V2_1_                         uint32   = 0xff
	NT_D3DVS_MAXLOOPITERATIONCOUNT_V3_0_                         uint32   = 0xff
	NT_D3DVS_PREDICATE_MAX_V2_1_                                 uint32   = 0x1
	NT_D3DVS_PREDICATE_MAX_V3_0_                                 uint32   = 0x1
	NT_D3DPS_INPUTREG_MAX_V1_1_                                  uint32   = 0x2
	NT_D3DPS_INPUTREG_MAX_V1_2_                                  uint32   = 0x2
	NT_D3DPS_INPUTREG_MAX_V1_3_                                  uint32   = 0x2
	NT_D3DPS_INPUTREG_MAX_V1_4_                                  uint32   = 0x2
	NT_D3DPS_INPUTREG_MAX_V2_0_                                  uint32   = 0x2
	NT_D3DPS_INPUTREG_MAX_V2_1_                                  uint32   = 0x2
	NT_D3DPS_INPUTREG_MAX_V3_0_                                  uint32   = 0xc
	NT_D3DPS_TEMPREG_MAX_V1_1_                                   uint32   = 0x2
	NT_D3DPS_TEMPREG_MAX_V1_2_                                   uint32   = 0x2
	NT_D3DPS_TEMPREG_MAX_V1_3_                                   uint32   = 0x2
	NT_D3DPS_TEMPREG_MAX_V1_4_                                   uint32   = 0x6
	NT_D3DPS_TEMPREG_MAX_V2_0_                                   uint32   = 0xc
	NT_D3DPS_TEMPREG_MAX_V2_1_                                   uint32   = 0x20
	NT_D3DPS_TEMPREG_MAX_V3_0_                                   uint32   = 0x20
	NT_D3DPS_TEXTUREREG_MAX_V1_1_                                uint32   = 0x4
	NT_D3DPS_TEXTUREREG_MAX_V1_2_                                uint32   = 0x4
	NT_D3DPS_TEXTUREREG_MAX_V1_3_                                uint32   = 0x4
	NT_D3DPS_TEXTUREREG_MAX_V1_4_                                uint32   = 0x6
	NT_D3DPS_TEXTUREREG_MAX_V2_0_                                uint32   = 0x8
	NT_D3DPS_TEXTUREREG_MAX_V2_1_                                uint32   = 0x8
	NT_D3DPS_TEXTUREREG_MAX_V3_0_                                uint32   = 0x0
	NT_D3DPS_COLOROUT_MAX_V2_0_                                  uint32   = 0x4
	NT_D3DPS_COLOROUT_MAX_V2_1_                                  uint32   = 0x4
	NT_D3DPS_COLOROUT_MAX_V3_0_                                  uint32   = 0x4
	NT_D3DPS_PREDICATE_MAX_V2_1_                                 uint32   = 0x1
	NT_D3DPS_PREDICATE_MAX_V3_0_                                 uint32   = 0x1
	NT_D3DPS_CONSTREG_MAX_SW_DX9_                                uint32   = 0x2000
	NT_D3DPS_CONSTREG_MAX_V1_1_                                  uint32   = 0x8
	NT_D3DPS_CONSTREG_MAX_V1_2_                                  uint32   = 0x8
	NT_D3DPS_CONSTREG_MAX_V1_3_                                  uint32   = 0x8
	NT_D3DPS_CONSTREG_MAX_V1_4_                                  uint32   = 0x8
	NT_D3DPS_CONSTREG_MAX_V2_0_                                  uint32   = 0x20
	NT_D3DPS_CONSTREG_MAX_V2_1_                                  uint32   = 0x20
	NT_D3DPS_CONSTREG_MAX_V3_0_                                  uint32   = 0xe0
	NT_D3DPS_CONSTBOOLREG_MAX_SW_DX9_                            uint32   = 0x800
	NT_D3DPS_CONSTBOOLREG_MAX_V2_1_                              uint32   = 0x10
	NT_D3DPS_CONSTBOOLREG_MAX_V3_0_                              uint32   = 0x10
	NT_D3DPS_CONSTINTREG_MAX_SW_DX9_                             uint32   = 0x800
	NT_D3DPS_CONSTINTREG_MAX_V2_1_                               uint32   = 0x10
	NT_D3DPS_CONSTINTREG_MAX_V3_0_                               uint32   = 0x10
	NT_D3DPS_MAXLOOPSTEP_V2_1_                                   uint32   = 0x80
	NT_D3DPS_MAXLOOPSTEP_V3_0_                                   uint32   = 0x80
	NT_D3DPS_MAXLOOPINITVALUE_V2_1_                              uint32   = 0xff
	NT_D3DPS_MAXLOOPINITVALUE_V3_0_                              uint32   = 0xff
	NT_D3DPS_MAXLOOPITERATIONCOUNT_V2_1_                         uint32   = 0xff
	NT_D3DPS_MAXLOOPITERATIONCOUNT_V3_0_                         uint32   = 0xff
	NT_D3DPS_INPUTREG_MAX_DX8_                                   uint32   = 0x8
	NT_D3DPS_TEMPREG_MAX_DX8_                                    uint32   = 0x8
	NT_D3DPS_CONSTREG_MAX_DX8_                                   uint32   = 0x8
	NT_D3DPS_TEXTUREREG_MAX_DX8_                                 uint32   = 0x8
	D3DVSDT_FLOAT1                                               uint32   = 0x0
	D3DVSDT_FLOAT2                                               uint32   = 0x1
	D3DVSDT_FLOAT3                                               uint32   = 0x2
	D3DVSDT_FLOAT4                                               uint32   = 0x3
	D3DVSDT_D3DCOLOR                                             uint32   = 0x4
	D3DVSDT_UBYTE4                                               uint32   = 0x5
	D3DVSDT_SHORT2                                               uint32   = 0x6
	D3DVSDT_SHORT4                                               uint32   = 0x7
	D3DVSDE_POSITION                                             uint32   = 0x0
	D3DVSDE_BLENDWEIGHT                                          uint32   = 0x1
	D3DVSDE_BLENDINDICES                                         uint32   = 0x2
	D3DVSDE_NORMAL                                               uint32   = 0x3
	D3DVSDE_PSIZE                                                uint32   = 0x4
	D3DVSDE_DIFFUSE                                              uint32   = 0x5
	D3DVSDE_SPECULAR                                             uint32   = 0x6
	D3DVSDE_TEXCOORD0                                            uint32   = 0x7
	D3DVSDE_TEXCOORD1                                            uint32   = 0x8
	D3DVSDE_TEXCOORD2                                            uint32   = 0x9
	D3DVSDE_TEXCOORD3                                            uint32   = 0xa
	D3DVSDE_TEXCOORD4                                            uint32   = 0xb
	D3DVSDE_TEXCOORD5                                            uint32   = 0xc
	D3DVSDE_TEXCOORD6                                            uint32   = 0xd
	D3DVSDE_TEXCOORD7                                            uint32   = 0xe
	D3DVSDE_POSITION2                                            uint32   = 0xf
	D3DVSDE_NORMAL2                                              uint32   = 0x10
	D3DVSD_TOKENTYPESHIFT                                        uint32   = 0x1d
	D3DVSD_STREAMNUMBERSHIFT                                     uint32   = 0x0
	D3DVSD_DATALOADTYPESHIFT                                     uint32   = 0x1c
	D3DVSD_DATATYPESHIFT                                         uint32   = 0x10
	D3DVSD_SKIPCOUNTSHIFT                                        uint32   = 0x10
	D3DVSD_VERTEXREGSHIFT                                        uint32   = 0x0
	D3DVSD_VERTEXREGINSHIFT                                      uint32   = 0x14
	D3DVSD_CONSTCOUNTSHIFT                                       uint32   = 0x19
	D3DVSD_CONSTADDRESSSHIFT                                     uint32   = 0x0
	D3DVSD_CONSTRSSHIFT                                          uint32   = 0x10
	D3DVSD_EXTCOUNTSHIFT                                         uint32   = 0x18
	D3DVSD_EXTINFOSHIFT                                          uint32   = 0x0
	D3DVSD_STREAMTESSSHIFT                                       uint32   = 0x1c
	DIRECT3D_VERSION                                             uint32   = 0x700
	D3DTRANSFORMCAPS_CLIP                                        int32    = 1
	D3DLIGHTINGMODEL_RGB                                         int32    = 1
	D3DLIGHTINGMODEL_MONO                                        int32    = 2
	D3DLIGHTCAPS_POINT                                           int32    = 1
	D3DLIGHTCAPS_SPOT                                            int32    = 2
	D3DLIGHTCAPS_DIRECTIONAL                                     int32    = 4
	D3DLIGHTCAPS_PARALLELPOINT                                   int32    = 8
	D3DLIGHTCAPS_GLSPOT                                          int32    = 16
	D3DPMISCCAPS_MASKPLANES                                      int32    = 1
	D3DPMISCCAPS_MASKZ                                           int32    = 2
	D3DPMISCCAPS_CONFORMANT                                      int32    = 8
	D3DPMISCCAPS_CULLNONE                                        int32    = 16
	D3DPMISCCAPS_CULLCW                                          int32    = 32
	D3DPMISCCAPS_CULLCCW                                         int32    = 64
	D3DPRASTERCAPS_DITHER                                        int32    = 1
	D3DPRASTERCAPS_ROP2                                          int32    = 2
	D3DPRASTERCAPS_XOR                                           int32    = 4
	D3DPRASTERCAPS_PAT                                           int32    = 8
	D3DPRASTERCAPS_ZTEST                                         int32    = 16
	D3DPRASTERCAPS_SUBPIXEL                                      int32    = 32
	D3DPRASTERCAPS_SUBPIXELX                                     int32    = 64
	D3DPRASTERCAPS_FOGVERTEX                                     int32    = 128
	D3DPRASTERCAPS_FOGTABLE                                      int32    = 256
	D3DPRASTERCAPS_STIPPLE                                       int32    = 512
	D3DPRASTERCAPS_ANTIALIASSORTDEPENDENT                        int32    = 1024
	D3DPRASTERCAPS_ANTIALIASSORTINDEPENDENT                      int32    = 2048
	D3DPRASTERCAPS_ANTIALIASEDGES                                int32    = 4096
	D3DPRASTERCAPS_MIPMAPLODBIAS                                 int32    = 8192
	D3DPRASTERCAPS_ZBIAS                                         int32    = 16384
	D3DPRASTERCAPS_ZBUFFERLESSHSR                                int32    = 32768
	D3DPRASTERCAPS_FOGRANGE                                      int32    = 65536
	D3DPRASTERCAPS_ANISOTROPY                                    int32    = 131072
	D3DPRASTERCAPS_WBUFFER                                       int32    = 262144
	D3DPRASTERCAPS_TRANSLUCENTSORTINDEPENDENT                    int32    = 524288
	D3DPRASTERCAPS_WFOG                                          int32    = 1048576
	D3DPRASTERCAPS_ZFOG                                          int32    = 2097152
	D3DPCMPCAPS_NEVER                                            int32    = 1
	D3DPCMPCAPS_LESS                                             int32    = 2
	D3DPCMPCAPS_EQUAL                                            int32    = 4
	D3DPCMPCAPS_LESSEQUAL                                        int32    = 8
	D3DPCMPCAPS_GREATER                                          int32    = 16
	D3DPCMPCAPS_NOTEQUAL                                         int32    = 32
	D3DPCMPCAPS_GREATEREQUAL                                     int32    = 64
	D3DPCMPCAPS_ALWAYS                                           int32    = 128
	D3DPBLENDCAPS_ZERO                                           int32    = 1
	D3DPBLENDCAPS_ONE                                            int32    = 2
	D3DPBLENDCAPS_SRCCOLOR                                       int32    = 4
	D3DPBLENDCAPS_INVSRCCOLOR                                    int32    = 8
	D3DPBLENDCAPS_SRCALPHA                                       int32    = 16
	D3DPBLENDCAPS_INVSRCALPHA                                    int32    = 32
	D3DPBLENDCAPS_DESTALPHA                                      int32    = 64
	D3DPBLENDCAPS_INVDESTALPHA                                   int32    = 128
	D3DPBLENDCAPS_DESTCOLOR                                      int32    = 256
	D3DPBLENDCAPS_INVDESTCOLOR                                   int32    = 512
	D3DPBLENDCAPS_SRCALPHASAT                                    int32    = 1024
	D3DPBLENDCAPS_BOTHSRCALPHA                                   int32    = 2048
	D3DPBLENDCAPS_BOTHINVSRCALPHA                                int32    = 4096
	D3DPSHADECAPS_COLORFLATMONO                                  int32    = 1
	D3DPSHADECAPS_COLORFLATRGB                                   int32    = 2
	D3DPSHADECAPS_COLORGOURAUDMONO                               int32    = 4
	D3DPSHADECAPS_COLORGOURAUDRGB                                int32    = 8
	D3DPSHADECAPS_COLORPHONGMONO                                 int32    = 16
	D3DPSHADECAPS_COLORPHONGRGB                                  int32    = 32
	D3DPSHADECAPS_SPECULARFLATMONO                               int32    = 64
	D3DPSHADECAPS_SPECULARFLATRGB                                int32    = 128
	D3DPSHADECAPS_SPECULARGOURAUDMONO                            int32    = 256
	D3DPSHADECAPS_SPECULARGOURAUDRGB                             int32    = 512
	D3DPSHADECAPS_SPECULARPHONGMONO                              int32    = 1024
	D3DPSHADECAPS_SPECULARPHONGRGB                               int32    = 2048
	D3DPSHADECAPS_ALPHAFLATBLEND                                 int32    = 4096
	D3DPSHADECAPS_ALPHAFLATSTIPPLED                              int32    = 8192
	D3DPSHADECAPS_ALPHAGOURAUDBLEND                              int32    = 16384
	D3DPSHADECAPS_ALPHAGOURAUDSTIPPLED                           int32    = 32768
	D3DPSHADECAPS_ALPHAPHONGBLEND                                int32    = 65536
	D3DPSHADECAPS_ALPHAPHONGSTIPPLED                             int32    = 131072
	D3DPSHADECAPS_FOGFLAT                                        int32    = 262144
	D3DPSHADECAPS_FOGGOURAUD                                     int32    = 524288
	D3DPSHADECAPS_FOGPHONG                                       int32    = 1048576
	D3DPTEXTURECAPS_PERSPECTIVE                                  int32    = 1
	D3DPTEXTURECAPS_POW2                                         int32    = 2
	D3DPTEXTURECAPS_ALPHA                                        int32    = 4
	D3DPTEXTURECAPS_TRANSPARENCY                                 int32    = 8
	D3DPTEXTURECAPS_BORDER                                       int32    = 16
	D3DPTEXTURECAPS_SQUAREONLY                                   int32    = 32
	D3DPTEXTURECAPS_TEXREPEATNOTSCALEDBYSIZE                     int32    = 64
	D3DPTEXTURECAPS_ALPHAPALETTE                                 int32    = 128
	D3DPTEXTURECAPS_NONPOW2CONDITIONAL                           int32    = 256
	D3DPTEXTURECAPS_PROJECTED                                    int32    = 1024
	D3DPTEXTURECAPS_CUBEMAP                                      int32    = 2048
	D3DPTEXTURECAPS_COLORKEYBLEND                                int32    = 4096
	D3DPTFILTERCAPS_NEAREST                                      int32    = 1
	D3DPTFILTERCAPS_LINEAR                                       int32    = 2
	D3DPTFILTERCAPS_MIPNEAREST                                   int32    = 4
	D3DPTFILTERCAPS_MIPLINEAR                                    int32    = 8
	D3DPTFILTERCAPS_LINEARMIPNEAREST                             int32    = 16
	D3DPTFILTERCAPS_LINEARMIPLINEAR                              int32    = 32
	D3DPTFILTERCAPS_MINFPOINT                                    int32    = 256
	D3DPTFILTERCAPS_MINFLINEAR                                   int32    = 512
	D3DPTFILTERCAPS_MINFANISOTROPIC                              int32    = 1024
	D3DPTFILTERCAPS_MIPFPOINT                                    int32    = 65536
	D3DPTFILTERCAPS_MIPFLINEAR                                   int32    = 131072
	D3DPTFILTERCAPS_MAGFPOINT                                    int32    = 16777216
	D3DPTFILTERCAPS_MAGFLINEAR                                   int32    = 33554432
	D3DPTFILTERCAPS_MAGFANISOTROPIC                              int32    = 67108864
	D3DPTFILTERCAPS_MAGFAFLATCUBIC                               int32    = 134217728
	D3DPTFILTERCAPS_MAGFGAUSSIANCUBIC                            int32    = 268435456
	D3DPTBLENDCAPS_DECAL                                         int32    = 1
	D3DPTBLENDCAPS_MODULATE                                      int32    = 2
	D3DPTBLENDCAPS_DECALALPHA                                    int32    = 4
	D3DPTBLENDCAPS_MODULATEALPHA                                 int32    = 8
	D3DPTBLENDCAPS_DECALMASK                                     int32    = 16
	D3DPTBLENDCAPS_MODULATEMASK                                  int32    = 32
	D3DPTBLENDCAPS_COPY                                          int32    = 64
	D3DPTBLENDCAPS_ADD                                           int32    = 128
	D3DPTADDRESSCAPS_WRAP                                        int32    = 1
	D3DPTADDRESSCAPS_MIRROR                                      int32    = 2
	D3DPTADDRESSCAPS_CLAMP                                       int32    = 4
	D3DPTADDRESSCAPS_BORDER                                      int32    = 8
	D3DPTADDRESSCAPS_INDEPENDENTUV                               int32    = 16
	D3DSTENCILCAPS_KEEP                                          int32    = 1
	D3DSTENCILCAPS_ZERO                                          int32    = 2
	D3DSTENCILCAPS_REPLACE                                       int32    = 4
	D3DSTENCILCAPS_INCRSAT                                       int32    = 8
	D3DSTENCILCAPS_DECRSAT                                       int32    = 16
	D3DSTENCILCAPS_INVERT                                        int32    = 32
	D3DSTENCILCAPS_INCR                                          int32    = 64
	D3DSTENCILCAPS_DECR                                          int32    = 128
	D3DTEXOPCAPS_DISABLE                                         int32    = 1
	D3DTEXOPCAPS_SELECTARG1                                      int32    = 2
	D3DTEXOPCAPS_SELECTARG2                                      int32    = 4
	D3DTEXOPCAPS_MODULATE                                        int32    = 8
	D3DTEXOPCAPS_MODULATE2X                                      int32    = 16
	D3DTEXOPCAPS_MODULATE4X                                      int32    = 32
	D3DTEXOPCAPS_ADD                                             int32    = 64
	D3DTEXOPCAPS_ADDSIGNED                                       int32    = 128
	D3DTEXOPCAPS_ADDSIGNED2X                                     int32    = 256
	D3DTEXOPCAPS_SUBTRACT                                        int32    = 512
	D3DTEXOPCAPS_ADDSMOOTH                                       int32    = 1024
	D3DTEXOPCAPS_BLENDDIFFUSEALPHA                               int32    = 2048
	D3DTEXOPCAPS_BLENDTEXTUREALPHA                               int32    = 4096
	D3DTEXOPCAPS_BLENDFACTORALPHA                                int32    = 8192
	D3DTEXOPCAPS_BLENDTEXTUREALPHAPM                             int32    = 16384
	D3DTEXOPCAPS_BLENDCURRENTALPHA                               int32    = 32768
	D3DTEXOPCAPS_PREMODULATE                                     int32    = 65536
	D3DTEXOPCAPS_MODULATEALPHA_ADDCOLOR                          int32    = 131072
	D3DTEXOPCAPS_MODULATECOLOR_ADDALPHA                          int32    = 262144
	D3DTEXOPCAPS_MODULATEINVALPHA_ADDCOLOR                       int32    = 524288
	D3DTEXOPCAPS_MODULATEINVCOLOR_ADDALPHA                       int32    = 1048576
	D3DTEXOPCAPS_BUMPENVMAP                                      int32    = 2097152
	D3DTEXOPCAPS_BUMPENVMAPLUMINANCE                             int32    = 4194304
	D3DTEXOPCAPS_DOTPRODUCT3                                     int32    = 8388608
	D3DFVFCAPS_TEXCOORDCOUNTMASK                                 int32    = 65535
	D3DFVFCAPS_DONOTSTRIPELEMENTS                                int32    = 524288
	D3DDD_COLORMODEL                                             int32    = 1
	D3DDD_DEVCAPS                                                int32    = 2
	D3DDD_TRANSFORMCAPS                                          int32    = 4
	D3DDD_LIGHTINGCAPS                                           int32    = 8
	D3DDD_BCLIPPING                                              int32    = 16
	D3DDD_LINECAPS                                               int32    = 32
	D3DDD_TRICAPS                                                int32    = 64
	D3DDD_DEVICERENDERBITDEPTH                                   int32    = 128
	D3DDD_DEVICEZBUFFERBITDEPTH                                  int32    = 256
	D3DDD_MAXBUFFERSIZE                                          int32    = 512
	D3DDD_MAXVERTEXCOUNT                                         int32    = 1024
	D3DDEVCAPS_FLOATTLVERTEX                                     int32    = 1
	D3DDEVCAPS_SORTINCREASINGZ                                   int32    = 2
	D3DDEVCAPS_SORTDECREASINGZ                                   int32    = 4
	D3DDEVCAPS_SORTEXACT                                         int32    = 8
	D3DDEVCAPS_EXECUTESYSTEMMEMORY                               int32    = 16
	D3DDEVCAPS_EXECUTEVIDEOMEMORY                                int32    = 32
	D3DDEVCAPS_TLVERTEXSYSTEMMEMORY                              int32    = 64
	D3DDEVCAPS_TLVERTEXVIDEOMEMORY                               int32    = 128
	D3DDEVCAPS_TEXTURESYSTEMMEMORY                               int32    = 256
	D3DDEVCAPS_TEXTUREVIDEOMEMORY                                int32    = 512
	D3DDEVCAPS_DRAWPRIMTLVERTEX                                  int32    = 1024
	D3DDEVCAPS_CANRENDERAFTERFLIP                                int32    = 2048
	D3DDEVCAPS_TEXTURENONLOCALVIDMEM                             int32    = 4096
	D3DDEVCAPS_DRAWPRIMITIVES2                                   int32    = 8192
	D3DDEVCAPS_SEPARATETEXTUREMEMORIES                           int32    = 16384
	D3DDEVCAPS_DRAWPRIMITIVES2EX                                 int32    = 32768
	D3DDEVCAPS_HWTRANSFORMANDLIGHT                               int32    = 65536
	D3DDEVCAPS_CANBLTSYSTONONLOCAL                               int32    = 131072
	D3DDEVCAPS_HWRASTERIZATION                                   int32    = 524288
	D3DVTXPCAPS_TEXGEN                                           int32    = 1
	D3DVTXPCAPS_MATERIALSOURCE7                                  int32    = 2
	D3DVTXPCAPS_VERTEXFOG                                        int32    = 4
	D3DVTXPCAPS_DIRECTIONALLIGHTS                                int32    = 8
	D3DVTXPCAPS_POSITIONALLIGHTS                                 int32    = 16
	D3DVTXPCAPS_LOCALVIEWER                                      int32    = 32
	D3DFDS_COLORMODEL                                            int32    = 1
	D3DFDS_GUID                                                  int32    = 2
	D3DFDS_HARDWARE                                              int32    = 4
	D3DFDS_TRIANGLES                                             int32    = 8
	D3DFDS_LINES                                                 int32    = 16
	D3DFDS_MISCCAPS                                              int32    = 32
	D3DFDS_RASTERCAPS                                            int32    = 64
	D3DFDS_ZCMPCAPS                                              int32    = 128
	D3DFDS_ALPHACMPCAPS                                          int32    = 256
	D3DFDS_SRCBLENDCAPS                                          int32    = 512
	D3DFDS_DSTBLENDCAPS                                          int32    = 1024
	D3DFDS_SHADECAPS                                             int32    = 2048
	D3DFDS_TEXTURECAPS                                           int32    = 4096
	D3DFDS_TEXTUREFILTERCAPS                                     int32    = 8192
	D3DFDS_TEXTUREBLENDCAPS                                      int32    = 16384
	D3DFDS_TEXTUREADDRESSCAPS                                    int32    = 32768
	D3DDEB_BUFSIZE                                               int32    = 1
	D3DDEB_CAPS                                                  int32    = 2
	D3DDEB_LPDATA                                                int32    = 4
	D3DDEBCAPS_SYSTEMMEMORY                                      int32    = 1
	D3DDEBCAPS_VIDEOMEMORY                                       int32    = 2
	D3DMAXUSERCLIPPLANES                                         uint32   = 0x20
	D3DCLIPPLANE0                                                uint32   = 0x1
	D3DCLIPPLANE1                                                uint32   = 0x2
	D3DCLIPPLANE2                                                uint32   = 0x4
	D3DCLIPPLANE3                                                uint32   = 0x8
	D3DCLIPPLANE4                                                uint32   = 0x10
	D3DCLIPPLANE5                                                uint32   = 0x20
	D3DCLIP_LEFT                                                 int32    = 1
	D3DCLIP_RIGHT                                                int32    = 2
	D3DCLIP_TOP                                                  int32    = 4
	D3DCLIP_BOTTOM                                               int32    = 8
	D3DCLIP_FRONT                                                int32    = 16
	D3DCLIP_BACK                                                 int32    = 32
	D3DCLIP_GEN0                                                 int32    = 64
	D3DCLIP_GEN1                                                 int32    = 128
	D3DCLIP_GEN2                                                 int32    = 256
	D3DCLIP_GEN3                                                 int32    = 512
	D3DCLIP_GEN4                                                 int32    = 1024
	D3DCLIP_GEN5                                                 int32    = 2048
	D3DSTATUS_CLIPUNIONLEFT                                      int32    = 1
	D3DSTATUS_CLIPUNIONRIGHT                                     int32    = 2
	D3DSTATUS_CLIPUNIONTOP                                       int32    = 4
	D3DSTATUS_CLIPUNIONBOTTOM                                    int32    = 8
	D3DSTATUS_CLIPUNIONFRONT                                     int32    = 16
	D3DSTATUS_CLIPUNIONBACK                                      int32    = 32
	D3DSTATUS_CLIPUNIONGEN0                                      int32    = 64
	D3DSTATUS_CLIPUNIONGEN1                                      int32    = 128
	D3DSTATUS_CLIPUNIONGEN2                                      int32    = 256
	D3DSTATUS_CLIPUNIONGEN3                                      int32    = 512
	D3DSTATUS_CLIPUNIONGEN4                                      int32    = 1024
	D3DSTATUS_CLIPUNIONGEN5                                      int32    = 2048
	D3DSTATUS_CLIPINTERSECTIONLEFT                               int32    = 4096
	D3DSTATUS_CLIPINTERSECTIONRIGHT                              int32    = 8192
	D3DSTATUS_CLIPINTERSECTIONTOP                                int32    = 16384
	D3DSTATUS_CLIPINTERSECTIONBOTTOM                             int32    = 32768
	D3DSTATUS_CLIPINTERSECTIONFRONT                              int32    = 65536
	D3DSTATUS_CLIPINTERSECTIONBACK                               int32    = 131072
	D3DSTATUS_CLIPINTERSECTIONGEN0                               int32    = 262144
	D3DSTATUS_CLIPINTERSECTIONGEN1                               int32    = 524288
	D3DSTATUS_CLIPINTERSECTIONGEN2                               int32    = 1048576
	D3DSTATUS_CLIPINTERSECTIONGEN3                               int32    = 2097152
	D3DSTATUS_CLIPINTERSECTIONGEN4                               int32    = 4194304
	D3DSTATUS_CLIPINTERSECTIONGEN5                               int32    = 8388608
	D3DSTATUS_ZNOTVISIBLE                                        int32    = 16777216
	D3DTRANSFORM_CLIPPED                                         int32    = 1
	D3DTRANSFORM_UNCLIPPED                                       int32    = 2
	D3DLIGHT_ACTIVE                                              uint32   = 0x1
	D3DLIGHT_NO_SPECULAR                                         uint32   = 0x2
	D3DCOLOR_MONO                                                uint32   = 0x1
	D3DCOLOR_RGB                                                 uint32   = 0x2
	D3DCLEAR_TARGET                                              int32    = 1
	D3DCLEAR_ZBUFFER                                             int32    = 2
	D3DCLEAR_STENCIL                                             int32    = 4
	D3DSTATE_OVERRIDE_BIAS                                       uint32   = 0x100
	D3DRENDERSTATE_WRAPBIAS                                      uint32   = 0x80
	D3DWRAP_U                                                    int32    = 1
	D3DWRAP_V                                                    int32    = 2
	D3DWRAPCOORD_0                                               int32    = 1
	D3DWRAPCOORD_1                                               int32    = 2
	D3DWRAPCOORD_2                                               int32    = 4
	D3DWRAPCOORD_3                                               int32    = 8
	D3DPROCESSVERTICES_TRANSFORMLIGHT                            int32    = 0
	D3DPROCESSVERTICES_TRANSFORM                                 int32    = 1
	D3DPROCESSVERTICES_COPY                                      int32    = 2
	D3DPROCESSVERTICES_OPMASK                                    int32    = 7
	D3DPROCESSVERTICES_UPDATEEXTENTS                             int32    = 8
	D3DPROCESSVERTICES_NOCOLOR                                   int32    = 16
	D3DTSS_TCI_PASSTHRU                                          uint32   = 0x0
	D3DTSS_TCI_CAMERASPACENORMAL                                 uint32   = 0x10000
	D3DTSS_TCI_CAMERASPACEPOSITION                               uint32   = 0x20000
	D3DTSS_TCI_CAMERASPACEREFLECTIONVECTOR                       uint32   = 0x30000
	D3DTA_SELECTMASK                                             uint32   = 0xf
	D3DTA_DIFFUSE                                                uint32   = 0x0
	D3DTA_CURRENT                                                uint32   = 0x1
	D3DTA_TEXTURE                                                uint32   = 0x2
	D3DTA_TFACTOR                                                uint32   = 0x3
	D3DTA_SPECULAR                                               uint32   = 0x4
	D3DTA_COMPLEMENT                                             uint32   = 0x10
	D3DTA_ALPHAREPLICATE                                         uint32   = 0x20
	D3DTRIFLAG_START                                             int32    = 0
	D3DTRIFLAG_ODD                                               int32    = 30
	D3DTRIFLAG_EVEN                                              int32    = 31
	D3DTRIFLAG_EDGEENABLE1                                       int32    = 256
	D3DTRIFLAG_EDGEENABLE2                                       int32    = 512
	D3DTRIFLAG_EDGEENABLE3                                       int32    = 1024
	D3DSETSTATUS_STATUS                                          int32    = 1
	D3DSETSTATUS_EXTENTS                                         int32    = 2
	D3DCLIPSTATUS_STATUS                                         int32    = 1
	D3DCLIPSTATUS_EXTENTS2                                       int32    = 2
	D3DCLIPSTATUS_EXTENTS3                                       int32    = 4
	D3DEXECUTE_CLIPPED                                           int32    = 1
	D3DEXECUTE_UNCLIPPED                                         int32    = 2
	D3DPAL_FREE                                                  uint32   = 0x0
	D3DPAL_READONLY                                              uint32   = 0x40
	D3DPAL_RESERVED                                              uint32   = 0x80
	D3DVBCAPS_SYSTEMMEMORY                                       int32    = 2048
	D3DVBCAPS_WRITEONLY                                          int32    = 65536
	D3DVBCAPS_OPTIMIZED                                          int32    = -2147483648
	D3DVBCAPS_DONOTCLIP                                          int32    = 1
	D3DVOP_LIGHT                                                 uint32   = 0x400
	D3DVOP_TRANSFORM                                             uint32   = 0x1
	D3DVOP_CLIP                                                  uint32   = 0x4
	D3DVOP_EXTENTS                                               uint32   = 0x8
	D3DPV_DONOTCOPYDATA                                          uint32   = 0x1
	D3DFVF_RESERVED0                                             uint32   = 0x1
	D3DFVF_POSITION_MASK                                         uint32   = 0xe
	D3DFVF_XYZ                                                   uint32   = 0x2
	D3DFVF_XYZRHW                                                uint32   = 0x4
	D3DFVF_XYZB1                                                 uint32   = 0x6
	D3DFVF_XYZB2                                                 uint32   = 0x8
	D3DFVF_XYZB3                                                 uint32   = 0xa
	D3DFVF_XYZB4                                                 uint32   = 0xc
	D3DFVF_XYZB5                                                 uint32   = 0xe
	D3DFVF_NORMAL                                                uint32   = 0x10
	D3DFVF_RESERVED1                                             uint32   = 0x20
	D3DFVF_DIFFUSE                                               uint32   = 0x40
	D3DFVF_SPECULAR                                              uint32   = 0x80
	D3DFVF_TEXCOUNT_MASK                                         uint32   = 0xf00
	D3DFVF_TEXCOUNT_SHIFT                                        uint32   = 0x8
	D3DFVF_TEX0                                                  uint32   = 0x0
	D3DFVF_TEX1                                                  uint32   = 0x100
	D3DFVF_TEX2                                                  uint32   = 0x200
	D3DFVF_TEX3                                                  uint32   = 0x300
	D3DFVF_TEX4                                                  uint32   = 0x400
	D3DFVF_TEX5                                                  uint32   = 0x500
	D3DFVF_TEX6                                                  uint32   = 0x600
	D3DFVF_TEX7                                                  uint32   = 0x700
	D3DFVF_TEX8                                                  uint32   = 0x800
	D3DFVF_RESERVED2                                             uint32   = 0xf000
	D3DDP_MAXTEXCOORD                                            uint32   = 0x8
	D3DVIS_INSIDE_FRUSTUM                                        uint32   = 0x0
	D3DVIS_INTERSECT_FRUSTUM                                     uint32   = 0x1
	D3DVIS_OUTSIDE_FRUSTUM                                       uint32   = 0x2
	D3DVIS_INSIDE_LEFT                                           uint32   = 0x0
	D3DVIS_INTERSECT_LEFT                                        uint32   = 0x4
	D3DVIS_OUTSIDE_LEFT                                          uint32   = 0x8
	D3DVIS_INSIDE_RIGHT                                          uint32   = 0x0
	D3DVIS_INTERSECT_RIGHT                                       uint32   = 0x10
	D3DVIS_OUTSIDE_RIGHT                                         uint32   = 0x20
	D3DVIS_INSIDE_TOP                                            uint32   = 0x0
	D3DVIS_INTERSECT_TOP                                         uint32   = 0x40
	D3DVIS_OUTSIDE_TOP                                           uint32   = 0x80
	D3DVIS_INSIDE_BOTTOM                                         uint32   = 0x0
	D3DVIS_INTERSECT_BOTTOM                                      uint32   = 0x100
	D3DVIS_OUTSIDE_BOTTOM                                        uint32   = 0x200
	D3DVIS_INSIDE_NEAR                                           uint32   = 0x0
	D3DVIS_INTERSECT_NEAR                                        uint32   = 0x400
	D3DVIS_OUTSIDE_NEAR                                          uint32   = 0x800
	D3DVIS_INSIDE_FAR                                            uint32   = 0x0
	D3DVIS_INTERSECT_FAR                                         uint32   = 0x1000
	D3DVIS_OUTSIDE_FAR                                           uint32   = 0x2000
	D3DVIS_MASK_FRUSTUM                                          uint32   = 0x3
	D3DVIS_MASK_LEFT                                             uint32   = 0xc
	D3DVIS_MASK_RIGHT                                            uint32   = 0x30
	D3DVIS_MASK_TOP                                              uint32   = 0xc0
	D3DVIS_MASK_BOTTOM                                           uint32   = 0x300
	D3DVIS_MASK_NEAR                                             uint32   = 0xc00
	D3DVIS_MASK_FAR                                              uint32   = 0x3000
	D3DDEVINFOID_TEXTUREMANAGER                                  uint32   = 0x1
	D3DDEVINFOID_D3DTEXTUREMANAGER                               uint32   = 0x2
	D3DDEVINFOID_TEXTURING                                       uint32   = 0x3
	D3DFVF_TEXTUREFORMAT2                                        uint32   = 0x0
	D3DFVF_TEXTUREFORMAT1                                        uint32   = 0x3
	D3DFVF_TEXTUREFORMAT3                                        uint32   = 0x1
	D3DFVF_TEXTUREFORMAT4                                        uint32   = 0x2
	ROTFLAGS_REGISTRATIONKEEPSALIVE                              uint32   = 0x1
	ROTFLAGS_ALLOWANYCLIENT                                      uint32   = 0x2
	ROT_COMPARE_MAX                                              uint32   = 0x800
	WDT_INPROC_CALL                                              uint32   = 0x48746457
	WDT_REMOTE_CALL                                              uint32   = 0x52746457
	WDT_INPROC64_CALL                                            uint32   = 0x50746457
	PROCESS_HEAP_REGION                                          uint32   = 0x1
	PROCESS_HEAP_UNCOMMITTED_RANGE                               uint32   = 0x2
	PROCESS_HEAP_ENTRY_BUSY                                      uint32   = 0x4
	PROCESS_HEAP_SEG_ALLOC                                       uint32   = 0x8
	PROCESS_HEAP_ENTRY_MOVEABLE                                  uint32   = 0x10
	PROCESS_HEAP_ENTRY_DDESHARE                                  uint32   = 0x20
	LMEM_NOCOMPACT                                               uint32   = 0x10
	LMEM_NODISCARD                                               uint32   = 0x20
	LMEM_MODIFY                                                  uint32   = 0x80
	LMEM_DISCARDABLE                                             uint32   = 0xf00
	LMEM_VALID_FLAGS                                             uint32   = 0xf72
	LMEM_INVALID_HANDLE                                          uint32   = 0x8000
	LMEM_DISCARDED                                               uint32   = 0x4000
	LMEM_LOCKCOUNT                                               uint32   = 0xff
	NUMA_NO_PREFERRED_NODE                                       uint32   = 0xffffffff
	REDBOOK_DIGITAL_AUDIO_EXTRACTION_INFO_VERSION                uint32   = 0x1
	FACILITY_MCA_ERROR_CODE                                      uint32   = 0x5
	IO_ERR_INSUFFICIENT_RESOURCES                                NTSTATUS = -1073479678
	IO_ERR_DRIVER_ERROR                                          NTSTATUS = -1073479676
	IO_ERR_SEEK_ERROR                                            NTSTATUS = -1073479674
	IO_ERR_BAD_BLOCK                                             NTSTATUS = -1073479673
	IO_ERR_TIMEOUT                                               NTSTATUS = -1073479671
	IO_ERR_CONTROLLER_ERROR                                      NTSTATUS = -1073479669
	IO_ERR_NOT_READY                                             NTSTATUS = -1073479665
	IO_ERR_INVALID_REQUEST                                       NTSTATUS = -1073479664
	IO_ERR_RESET                                                 NTSTATUS = -1073479661
	IO_ERR_BAD_FIRMWARE                                          NTSTATUS = -1073479655
	IO_WRN_BAD_FIRMWARE                                          NTSTATUS = -2147221478
	IO_WRITE_CACHE_ENABLED                                       NTSTATUS = -2147221472
	IO_RECOVERED_VIA_ECC                                         NTSTATUS = -2147221471
	IO_WRITE_CACHE_DISABLED                                      NTSTATUS = -2147221470
	IO_WARNING_PAGING_FAILURE                                    NTSTATUS = -2147221453
	IO_WRN_FAILURE_PREDICTED                                     NTSTATUS = -2147221452
	IO_WARNING_ALLOCATION_FAILED                                 NTSTATUS = -2147221448
	IO_WARNING_DUPLICATE_SIGNATURE                               NTSTATUS = -2147221446
	IO_WARNING_DUPLICATE_PATH                                    NTSTATUS = -2147221445
	IO_WARNING_WRITE_FUA_PROBLEM                                 NTSTATUS = -2147221372
	IO_WARNING_VOLUME_LOST_DISK_EXTENT                           NTSTATUS = -2147221362
	IO_WARNING_DEVICE_HAS_INTERNAL_DUMP                          NTSTATUS = -2147221361
	IO_WARNING_SOFT_THRESHOLD_REACHED                            NTSTATUS = -2147221360
	IO_WARNING_SOFT_THRESHOLD_REACHED_EX                         NTSTATUS = -2147221359
	IO_WARNING_SOFT_THRESHOLD_REACHED_EX_LUN_LUN                 NTSTATUS = -2147221358
	IO_WARNING_SOFT_THRESHOLD_REACHED_EX_LUN_POOL                NTSTATUS = -2147221357
	IO_WARNING_SOFT_THRESHOLD_REACHED_EX_POOL_LUN                NTSTATUS = -2147221356
	IO_WARNING_SOFT_THRESHOLD_REACHED_EX_POOL_POOL               NTSTATUS = -2147221355
	IO_ERROR_DISK_RESOURCES_EXHAUSTED                            NTSTATUS = -1073479530
	IO_WARNING_DISK_CAPACITY_CHANGED                             NTSTATUS = -2147221353
	IO_WARNING_DISK_PROVISIONING_TYPE_CHANGED                    NTSTATUS = -2147221352
	IO_WARNING_IO_OPERATION_RETRIED                              NTSTATUS = -2147221351
	IO_ERROR_IO_HARDWARE_ERROR                                   NTSTATUS = -1073479526
	IO_WARNING_COMPLETION_TIME                                   NTSTATUS = -2147221349
	IO_WARNING_DISK_SURPRISE_REMOVED                             NTSTATUS = -2147221347
	IO_WARNING_REPEATED_DISK_GUID                                NTSTATUS = -2147221346
	IO_WARNING_DISK_FIRMWARE_UPDATED                             NTSTATUS = 1074004127
	IO_ERR_RETRY_SUCCEEDED                                       NTSTATUS = 262145
	IO_DUMP_CREATION_SUCCESS                                     NTSTATUS = 262306
	IO_FILE_QUOTA_THRESHOLD                                      NTSTATUS = 1074004004
	IO_FILE_QUOTA_LIMIT                                          NTSTATUS = 1074004005
	IO_FILE_QUOTA_STARTED                                        NTSTATUS = 1074004006
	IO_FILE_QUOTA_SUCCEEDED                                      NTSTATUS = 1074004007
	IO_INFO_THROTTLE_COMPLETE                                    NTSTATUS = 1074004087
	IO_CDROM_EXCLUSIVE_LOCK                                      NTSTATUS = 1074004101
	IO_WARNING_ADAPTER_FIRMWARE_UPDATED                          NTSTATUS = 1074004128
	IO_FILE_QUOTA_FAILED                                         NTSTATUS = -2147221464
	IO_LOST_DELAYED_WRITE                                        NTSTATUS = -2147221454
	IO_WARNING_INTERRUPT_STILL_PENDING                           NTSTATUS = -2147221451
	IO_DRIVER_CANCEL_TIMEOUT                                     NTSTATUS = -2147221450
	IO_WARNING_LOG_FLUSH_FAILED                                  NTSTATUS = -2147221447
	IO_WARNING_BUS_RESET                                         NTSTATUS = -2147221386
	IO_WARNING_RESET                                             NTSTATUS = -2147221375
	IO_LOST_DELAYED_WRITE_NETWORK_DISCONNECTED                   NTSTATUS = -2147221365
	IO_LOST_DELAYED_WRITE_NETWORK_SERVER_ERROR                   NTSTATUS = -2147221364
	IO_LOST_DELAYED_WRITE_NETWORK_LOCAL_DISK_ERROR               NTSTATUS = -2147221363
	IO_WARNING_DUMP_DISABLED_DEVICE_GONE                         NTSTATUS = -2147221348
	IO_ERR_CONFIGURATION_ERROR                                   NTSTATUS = -1073479677
	IO_ERR_PARITY                                                NTSTATUS = -1073479675
	IO_ERR_OVERRUN_ERROR                                         NTSTATUS = -1073479672
	IO_ERR_SEQUENCE                                              NTSTATUS = -1073479670
	IO_ERR_INTERNAL_ERROR                                        NTSTATUS = -1073479668
	IO_ERR_INCORRECT_IRQL                                        NTSTATUS = -1073479667
	IO_ERR_INVALID_IOBASE                                        NTSTATUS = -1073479666
	IO_ERR_VERSION                                               NTSTATUS = -1073479663
	IO_ERR_LAYERED_FAILURE                                       NTSTATUS = -1073479662
	IO_ERR_PROTOCOL                                              NTSTATUS = -1073479660
	IO_ERR_MEMORY_CONFLICT_DETECTED                              NTSTATUS = -1073479659
	IO_ERR_PORT_CONFLICT_DETECTED                                NTSTATUS = -1073479658
	IO_ERR_DMA_CONFLICT_DETECTED                                 NTSTATUS = -1073479657
	IO_ERR_IRQ_CONFLICT_DETECTED                                 NTSTATUS = -1073479656
	IO_ERR_DMA_RESOURCE_CONFLICT                                 NTSTATUS = -1073479653
	IO_ERR_INTERRUPT_RESOURCE_CONFLICT                           NTSTATUS = -1073479652
	IO_ERR_MEMORY_RESOURCE_CONFLICT                              NTSTATUS = -1073479651
	IO_ERR_PORT_RESOURCE_CONFLICT                                NTSTATUS = -1073479650
	IO_BAD_BLOCK_WITH_NAME                                       NTSTATUS = -1073479649
	IO_FILE_SYSTEM_CORRUPT                                       NTSTATUS = -1073479639
	IO_FILE_QUOTA_CORRUPT                                        NTSTATUS = -1073479638
	IO_SYSTEM_SLEEP_FAILED                                       NTSTATUS = -1073479637
	IO_DUMP_POINTER_FAILURE                                      NTSTATUS = -1073479636
	IO_DUMP_DRIVER_LOAD_FAILURE                                  NTSTATUS = -1073479635
	IO_DUMP_INITIALIZATION_FAILURE                               NTSTATUS = -1073479634
	IO_DUMP_DUMPFILE_CONFLICT                                    NTSTATUS = -1073479633
	IO_DUMP_DIRECT_CONFIG_FAILED                                 NTSTATUS = -1073479632
	IO_DUMP_PAGE_CONFIG_FAILED                                   NTSTATUS = -1073479631
	IO_FILE_SYSTEM_CORRUPT_WITH_NAME                             NTSTATUS = -1073479625
	IO_ERR_THREAD_STUCK_IN_DEVICE_DRIVER                         NTSTATUS = -1073479572
	IO_ERR_PORT_TIMEOUT                                          NTSTATUS = -1073479563
	IO_ERROR_DUMP_CREATION_ERROR                                 NTSTATUS = -1073479519
	IO_DUMP_CALLBACK_EXCEPTION                                   NTSTATUS = -1073479517
	MCA_INFO_CPU_THERMAL_THROTTLING_REMOVED                      NTSTATUS = 1074069616
	MCA_INFO_NO_MORE_CORRECTED_ERROR_LOGS                        NTSTATUS = 1074069619
	MCA_INFO_MEMORY_PAGE_MARKED_BAD                              NTSTATUS = 1074069620
	MCA_WARNING_CACHE                                            NTSTATUS = -2147155908
	MCA_WARNING_TLB                                              NTSTATUS = -2147155906
	MCA_WARNING_CPU_BUS                                          NTSTATUS = -2147155904
	MCA_WARNING_REGISTER_FILE                                    NTSTATUS = -2147155902
	MCA_WARNING_MAS                                              NTSTATUS = -2147155900
	MCA_WARNING_MEM_UNKNOWN                                      NTSTATUS = -2147155898
	MCA_WARNING_MEM_1_2                                          NTSTATUS = -2147155896
	MCA_WARNING_MEM_1_2_5                                        NTSTATUS = -2147155894
	MCA_WARNING_MEM_1_2_5_4                                      NTSTATUS = -2147155892
	MCA_WARNING_SYSTEM_EVENT                                     NTSTATUS = -2147155890
	MCA_WARNING_PCI_BUS_PARITY                                   NTSTATUS = -2147155888
	MCA_WARNING_PCI_BUS_PARITY_NO_INFO                           NTSTATUS = -2147155886
	MCA_WARNING_PCI_BUS_SERR                                     NTSTATUS = -2147155884
	MCA_WARNING_PCI_BUS_SERR_NO_INFO                             NTSTATUS = -2147155882
	MCA_WARNING_PCI_BUS_MASTER_ABORT                             NTSTATUS = -2147155880
	MCA_WARNING_PCI_BUS_MASTER_ABORT_NO_INFO                     NTSTATUS = -2147155878
	MCA_WARNING_PCI_BUS_TIMEOUT                                  NTSTATUS = -2147155876
	MCA_WARNING_PCI_BUS_TIMEOUT_NO_INFO                          NTSTATUS = -2147155874
	MCA_WARNING_PCI_BUS_UNKNOWN                                  NTSTATUS = -2147155872
	MCA_WARNING_PCI_DEVICE                                       NTSTATUS = -2147155870
	MCA_WARNING_SMBIOS                                           NTSTATUS = -2147155868
	MCA_WARNING_PLATFORM_SPECIFIC                                NTSTATUS = -2147155866
	MCA_WARNING_UNKNOWN                                          NTSTATUS = -2147155864
	MCA_WARNING_UNKNOWN_NO_CPU                                   NTSTATUS = -2147155862
	MCA_WARNING_CMC_THRESHOLD_EXCEEDED                           NTSTATUS = -2147155859
	MCA_WARNING_CPE_THRESHOLD_EXCEEDED                           NTSTATUS = -2147155858
	MCA_WARNING_CPU_THERMAL_THROTTLED                            NTSTATUS = -2147155857
	MCA_WARNING_CPU                                              NTSTATUS = -2147155855
	MCA_ERROR_CACHE                                              NTSTATUS = -1073414083
	MCA_ERROR_TLB                                                NTSTATUS = -1073414081
	MCA_ERROR_CPU_BUS                                            NTSTATUS = -1073414079
	MCA_ERROR_REGISTER_FILE                                      NTSTATUS = -1073414077
	MCA_ERROR_MAS                                                NTSTATUS = -1073414075
	MCA_ERROR_MEM_UNKNOWN                                        NTSTATUS = -1073414073
	MCA_ERROR_MEM_1_2                                            NTSTATUS = -1073414071
	MCA_ERROR_MEM_1_2_5                                          NTSTATUS = -1073414069
	MCA_ERROR_MEM_1_2_5_4                                        NTSTATUS = -1073414067
	MCA_ERROR_SYSTEM_EVENT                                       NTSTATUS = -1073414065
	MCA_ERROR_PCI_BUS_PARITY                                     NTSTATUS = -1073414063
	MCA_ERROR_PCI_BUS_PARITY_NO_INFO                             NTSTATUS = -1073414061
	MCA_ERROR_PCI_BUS_SERR                                       NTSTATUS = -1073414059
	MCA_ERROR_PCI_BUS_SERR_NO_INFO                               NTSTATUS = -1073414057
	MCA_ERROR_PCI_BUS_MASTER_ABORT                               NTSTATUS = -1073414055
	MCA_ERROR_PCI_BUS_MASTER_ABORT_NO_INFO                       NTSTATUS = -1073414053
	MCA_ERROR_PCI_BUS_TIMEOUT                                    NTSTATUS = -1073414051
	MCA_ERROR_PCI_BUS_TIMEOUT_NO_INFO                            NTSTATUS = -1073414049
	MCA_ERROR_PCI_BUS_UNKNOWN                                    NTSTATUS = -1073414047
	MCA_ERROR_PCI_DEVICE                                         NTSTATUS = -1073414045
	MCA_ERROR_SMBIOS                                             NTSTATUS = -1073414043
	MCA_ERROR_PLATFORM_SPECIFIC                                  NTSTATUS = -1073414041
	MCA_ERROR_UNKNOWN                                            NTSTATUS = -1073414039
	MCA_ERROR_UNKNOWN_NO_CPU                                     NTSTATUS = -1073414037
	MCA_ERROR_CPU                                                NTSTATUS = -1073414030
	MCA_MEMORYHIERARCHY_ERROR                                    NTSTATUS = -1073414024
	MCA_TLB_ERROR                                                NTSTATUS = -1073414023
	MCA_BUS_ERROR                                                NTSTATUS = -1073414022
	MCA_BUS_TIMEOUT_ERROR                                        NTSTATUS = -1073414021
	MCA_INTERNALTIMER_ERROR                                      NTSTATUS = -1073414020
	MCA_MICROCODE_ROM_PARITY_ERROR                               NTSTATUS = -1073414018
	MCA_EXTERNAL_ERROR                                           NTSTATUS = -1073414017
	MCA_FRC_ERROR                                                NTSTATUS = -1073414016
	VOLMGR_KSR_ERROR                                             NTSTATUS = -2143813631
	VOLMGR_KSR_READ_ERROR                                        NTSTATUS = -2143813630
	VOLMGR_KSR_BYPASS                                            NTSTATUS = -2143813629
)

var (
	GUID_MAX_POWER_SAVINGS = syscall.GUID{0xA1841308, 0x3541, 0x4FAB,
		[8]byte{0xBC, 0x81, 0xF7, 0x15, 0x56, 0xF2, 0x0B, 0x4A}}

	GUID_MIN_POWER_SAVINGS = syscall.GUID{0x8C5E7FDA, 0xE8BF, 0x4A96,
		[8]byte{0x9A, 0x85, 0xA6, 0xE2, 0x3A, 0x8C, 0x63, 0x5C}}

	GUID_TYPICAL_POWER_SAVINGS = syscall.GUID{0x381B4222, 0xF694, 0x41F0,
		[8]byte{0x96, 0x85, 0xFF, 0x5B, 0xB2, 0x60, 0xDF, 0x2E}}

	NO_SUBGROUP_GUID = syscall.GUID{0xFEA3413E, 0x7E05, 0x4911,
		[8]byte{0x9A, 0x71, 0x70, 0x03, 0x31, 0xF1, 0xC2, 0x94}}

	ALL_POWERSCHEMES_GUID = syscall.GUID{0x68A1E95E, 0x13EA, 0x41E1,
		[8]byte{0x80, 0x11, 0x0C, 0x49, 0x6C, 0xA4, 0x90, 0xB0}}

	GUID_POWERSCHEME_PERSONALITY = syscall.GUID{0x245D8541, 0x3943, 0x4422,
		[8]byte{0xB0, 0x25, 0x13, 0xA7, 0x84, 0xF6, 0x79, 0xB7}}

	GUID_ACTIVE_POWERSCHEME = syscall.GUID{0x31F9F286, 0x5084, 0x42FE,
		[8]byte{0xB7, 0x20, 0x2B, 0x02, 0x64, 0x99, 0x37, 0x63}}

	GUID_IDLE_RESILIENCY_SUBGROUP = syscall.GUID{0x2E601130, 0x5351, 0x4D9D,
		[8]byte{0x8E, 0x04, 0x25, 0x29, 0x66, 0xBA, 0xD0, 0x54}}

	GUID_IDLE_RESILIENCY_PERIOD = syscall.GUID{0xC42B79AA, 0xAA3A, 0x484B,
		[8]byte{0xA9, 0x8F, 0x2C, 0xF3, 0x2A, 0xA9, 0x0A, 0x28}}

	GUID_DEEP_SLEEP_ENABLED = syscall.GUID{0xD502F7EE, 0x1DC7, 0x4EFD,
		[8]byte{0xA5, 0x5D, 0xF0, 0x4B, 0x6F, 0x5C, 0x05, 0x45}}

	GUID_DEEP_SLEEP_PLATFORM_STATE = syscall.GUID{0xD23F2FB8, 0x9536, 0x4038,
		[8]byte{0x9C, 0x94, 0x1C, 0xE0, 0x2E, 0x5C, 0x21, 0x52}}

	GUID_DISK_COALESCING_POWERDOWN_TIMEOUT = syscall.GUID{0xC36F0EB4, 0x2988, 0x4A70,
		[8]byte{0x8E, 0xEE, 0x08, 0x84, 0xFC, 0x2C, 0x24, 0x33}}

	GUID_EXECUTION_REQUIRED_REQUEST_TIMEOUT = syscall.GUID{0x3166BC41, 0x7E98, 0x4E03,
		[8]byte{0xB3, 0x4E, 0xEC, 0x0F, 0x5F, 0x2B, 0x21, 0x8E}}

	GUID_VIDEO_SUBGROUP = syscall.GUID{0x7516B95F, 0xF776, 0x4464,
		[8]byte{0x8C, 0x53, 0x06, 0x16, 0x7F, 0x40, 0xCC, 0x99}}

	GUID_VIDEO_POWERDOWN_TIMEOUT = syscall.GUID{0x3C0BC021, 0xC8A8, 0x4E07,
		[8]byte{0xA9, 0x73, 0x6B, 0x14, 0xCB, 0xCB, 0x2B, 0x7E}}

	GUID_VIDEO_ANNOYANCE_TIMEOUT = syscall.GUID{0x82DBCF2D, 0xCD67, 0x40C5,
		[8]byte{0xBF, 0xDC, 0x9F, 0x1A, 0x5C, 0xCD, 0x46, 0x63}}

	GUID_VIDEO_ADAPTIVE_PERCENT_INCREASE = syscall.GUID{0xEED904DF, 0xB142, 0x4183,
		[8]byte{0xB1, 0x0B, 0x5A, 0x11, 0x97, 0xA3, 0x78, 0x64}}

	GUID_VIDEO_DIM_TIMEOUT = syscall.GUID{0x17AAA29B, 0x8B43, 0x4B94,
		[8]byte{0xAA, 0xFE, 0x35, 0xF6, 0x4D, 0xAA, 0xF1, 0xEE}}

	GUID_VIDEO_ADAPTIVE_POWERDOWN = syscall.GUID{0x90959D22, 0xD6A1, 0x49B9,
		[8]byte{0xAF, 0x93, 0xBC, 0xE8, 0x85, 0xAD, 0x33, 0x5B}}

	GUID_MONITOR_POWER_ON = syscall.GUID{0x02731015, 0x4510, 0x4526,
		[8]byte{0x99, 0xE6, 0xE5, 0xA1, 0x7E, 0xBD, 0x1A, 0xEA}}

	GUID_DEVICE_POWER_POLICY_VIDEO_BRIGHTNESS = syscall.GUID{0xADED5E82, 0xB909, 0x4619,
		[8]byte{0x99, 0x49, 0xF5, 0xD7, 0x1D, 0xAC, 0x0B, 0xCB}}

	GUID_DEVICE_POWER_POLICY_VIDEO_DIM_BRIGHTNESS = syscall.GUID{0xF1FBFDE2, 0xA960, 0x4165,
		[8]byte{0x9F, 0x88, 0x50, 0x66, 0x79, 0x11, 0xCE, 0x96}}

	GUID_VIDEO_CURRENT_MONITOR_BRIGHTNESS = syscall.GUID{0x8FFEE2C6, 0x2D01, 0x46BE,
		[8]byte{0xAD, 0xB9, 0x39, 0x8A, 0xDD, 0xC5, 0xB4, 0xFF}}

	GUID_VIDEO_ADAPTIVE_DISPLAY_BRIGHTNESS = syscall.GUID{0xFBD9AA66, 0x9553, 0x4097,
		[8]byte{0xBA, 0x44, 0xED, 0x6E, 0x9D, 0x65, 0xEA, 0xB8}}

	GUID_CONSOLE_DISPLAY_STATE = syscall.GUID{0x6FE69556, 0x704A, 0x47A0,
		[8]byte{0x8F, 0x24, 0xC2, 0x8D, 0x93, 0x6F, 0xDA, 0x47}}

	GUID_ALLOW_DISPLAY_REQUIRED = syscall.GUID{0xA9CEB8DA, 0xCD46, 0x44FB,
		[8]byte{0xA9, 0x8B, 0x02, 0xAF, 0x69, 0xDE, 0x46, 0x23}}

	GUID_VIDEO_CONSOLE_LOCK_TIMEOUT = syscall.GUID{0x8EC4B3A5, 0x6868, 0x48C2,
		[8]byte{0xBE, 0x75, 0x4F, 0x30, 0x44, 0xBE, 0x88, 0xA7}}

	GUID_ADVANCED_COLOR_QUALITY_BIAS = syscall.GUID{0x684C3E69, 0xA4F7, 0x4014,
		[8]byte{0x87, 0x54, 0xD4, 0x51, 0x79, 0xA5, 0x61, 0x67}}

	GUID_ADAPTIVE_POWER_BEHAVIOR_SUBGROUP = syscall.GUID{0x8619B916, 0xE004, 0x4DD8,
		[8]byte{0x9B, 0x66, 0xDA, 0xE8, 0x6F, 0x80, 0x66, 0x98}}

	GUID_NON_ADAPTIVE_INPUT_TIMEOUT = syscall.GUID{0x5ADBBFBC, 0x074E, 0x4DA1,
		[8]byte{0xBA, 0x38, 0xDB, 0x8B, 0x36, 0xB2, 0xC8, 0xF3}}

	GUID_ADAPTIVE_INPUT_CONTROLLER_STATE = syscall.GUID{0x0E98FAE9, 0xF45A, 0x4DE1,
		[8]byte{0xA7, 0x57, 0x60, 0x31, 0xF1, 0x97, 0xF6, 0xEA}}

	GUID_DISK_SUBGROUP = syscall.GUID{0x0012EE47, 0x9041, 0x4B5D,
		[8]byte{0x9B, 0x77, 0x53, 0x5F, 0xBA, 0x8B, 0x14, 0x42}}

	GUID_DISK_MAX_POWER = syscall.GUID{0x51DEA550, 0xBB38, 0x4BC4,
		[8]byte{0x99, 0x1B, 0xEA, 0xCF, 0x37, 0xBE, 0x5E, 0xC8}}

	GUID_DISK_POWERDOWN_TIMEOUT = syscall.GUID{0x6738E2C4, 0xE8A5, 0x4A42,
		[8]byte{0xB1, 0x6A, 0xE0, 0x40, 0xE7, 0x69, 0x75, 0x6E}}

	GUID_DISK_IDLE_TIMEOUT = syscall.GUID{0x58E39BA8, 0xB8E6, 0x4EF6,
		[8]byte{0x90, 0xD0, 0x89, 0xAE, 0x32, 0xB2, 0x58, 0xD6}}

	GUID_DISK_BURST_IGNORE_THRESHOLD = syscall.GUID{0x80E3C60E, 0xBB94, 0x4AD8,
		[8]byte{0xBB, 0xE0, 0x0D, 0x31, 0x95, 0xEF, 0xC6, 0x63}}

	GUID_DISK_ADAPTIVE_POWERDOWN = syscall.GUID{0x396A32E1, 0x499A, 0x40B2,
		[8]byte{0x91, 0x24, 0xA9, 0x6A, 0xFE, 0x70, 0x76, 0x67}}

	GUID_DISK_NVME_NOPPME = syscall.GUID{0xFC7372B6, 0xAB2D, 0x43EE,
		[8]byte{0x87, 0x97, 0x15, 0xE9, 0x84, 0x1F, 0x2C, 0xCA}}

	GUID_SLEEP_SUBGROUP = syscall.GUID{0x238C9FA8, 0x0AAD, 0x41ED,
		[8]byte{0x83, 0xF4, 0x97, 0xBE, 0x24, 0x2C, 0x8F, 0x20}}

	GUID_SLEEP_IDLE_THRESHOLD = syscall.GUID{0x81CD32E0, 0x7833, 0x44F3,
		[8]byte{0x87, 0x37, 0x70, 0x81, 0xF3, 0x8D, 0x1F, 0x70}}

	GUID_STANDBY_TIMEOUT = syscall.GUID{0x29F6C1DB, 0x86DA, 0x48C5,
		[8]byte{0x9F, 0xDB, 0xF2, 0xB6, 0x7B, 0x1F, 0x44, 0xDA}}

	GUID_UNATTEND_SLEEP_TIMEOUT = syscall.GUID{0x7BC4A2F9, 0xD8FC, 0x4469,
		[8]byte{0xB0, 0x7B, 0x33, 0xEB, 0x78, 0x5A, 0xAC, 0xA0}}

	GUID_HIBERNATE_TIMEOUT = syscall.GUID{0x9D7815A6, 0x7EE4, 0x497E,
		[8]byte{0x88, 0x88, 0x51, 0x5A, 0x05, 0xF0, 0x23, 0x64}}

	GUID_HIBERNATE_FASTS4_POLICY = syscall.GUID{0x94AC6D29, 0x73CE, 0x41A6,
		[8]byte{0x80, 0x9F, 0x63, 0x63, 0xBA, 0x21, 0xB4, 0x7E}}

	GUID_CRITICAL_POWER_TRANSITION = syscall.GUID{0xB7A27025, 0xE569, 0x46C2,
		[8]byte{0xA5, 0x04, 0x2B, 0x96, 0xCA, 0xD2, 0x25, 0xA1}}

	GUID_SYSTEM_AWAYMODE = syscall.GUID{0x98A7F580, 0x01F7, 0x48AA,
		[8]byte{0x9C, 0x0F, 0x44, 0x35, 0x2C, 0x29, 0xE5, 0xC0}}

	GUID_ALLOW_AWAYMODE = syscall.GUID{0x25DFA149, 0x5DD1, 0x4736,
		[8]byte{0xB5, 0xAB, 0xE8, 0xA3, 0x7B, 0x5B, 0x81, 0x87}}

	GUID_USER_PRESENCE_PREDICTION = syscall.GUID{0x82011705, 0xFB95, 0x4D46,
		[8]byte{0x8D, 0x35, 0x40, 0x42, 0xB1, 0xD2, 0x0D, 0xEF}}

	GUID_STANDBY_BUDGET_GRACE_PERIOD = syscall.GUID{0x60C07FE1, 0x0556, 0x45CF,
		[8]byte{0x99, 0x03, 0xD5, 0x6E, 0x32, 0x21, 0x02, 0x42}}

	GUID_STANDBY_BUDGET_PERCENT = syscall.GUID{0x9FE527BE, 0x1B70, 0x48DA,
		[8]byte{0x93, 0x0D, 0x7B, 0xCF, 0x17, 0xB4, 0x49, 0x90}}

	GUID_STANDBY_RESERVE_GRACE_PERIOD = syscall.GUID{0xC763EE92, 0x71E8, 0x4127,
		[8]byte{0x84, 0xEB, 0xF6, 0xED, 0x04, 0x3A, 0x3E, 0x3D}}

	GUID_STANDBY_RESERVE_TIME = syscall.GUID{0x468FE7E5, 0x1158, 0x46EC,
		[8]byte{0x88, 0xBC, 0x5B, 0x96, 0xC9, 0xE4, 0x4F, 0xD0}}

	GUID_STANDBY_RESET_PERCENT = syscall.GUID{0x49CB11A5, 0x56E2, 0x4AFB,
		[8]byte{0x9D, 0x38, 0x3D, 0xF4, 0x78, 0x72, 0xE2, 0x1B}}

	GUID_HUPR_ADAPTIVE_DISPLAY_TIMEOUT = syscall.GUID{0x0A7D6AB6, 0xAC83, 0x4AD1,
		[8]byte{0x82, 0x82, 0xEC, 0xA5, 0xB5, 0x83, 0x08, 0xF3}}

	GUID_ALLOW_STANDBY_STATES = syscall.GUID{0xABFC2519, 0x3608, 0x4C2A,
		[8]byte{0x94, 0xEA, 0x17, 0x1B, 0x0E, 0xD5, 0x46, 0xAB}}

	GUID_ALLOW_RTC_WAKE = syscall.GUID{0xBD3B718A, 0x0680, 0x4D9D,
		[8]byte{0x8A, 0xB2, 0xE1, 0xD2, 0xB4, 0xAC, 0x80, 0x6D}}

	GUID_LEGACY_RTC_MITIGATION = syscall.GUID{0x1A34BDC3, 0x7E6B, 0x442E,
		[8]byte{0xA9, 0xD0, 0x64, 0xB6, 0xEF, 0x37, 0x8E, 0x84}}

	GUID_ALLOW_SYSTEM_REQUIRED = syscall.GUID{0xA4B195F5, 0x8225, 0x47D8,
		[8]byte{0x80, 0x12, 0x9D, 0x41, 0x36, 0x97, 0x86, 0xE2}}

	GUID_POWER_SAVING_STATUS = syscall.GUID{0xE00958C0, 0xC213, 0x4ACE,
		[8]byte{0xAC, 0x77, 0xFE, 0xCC, 0xED, 0x2E, 0xEE, 0xA5}}

	GUID_ENERGY_SAVER_SUBGROUP = syscall.GUID{0xDE830923, 0xA562, 0x41AF,
		[8]byte{0xA0, 0x86, 0xE3, 0xA2, 0xC6, 0xBA, 0xD2, 0xDA}}

	GUID_ENERGY_SAVER_BATTERY_THRESHOLD = syscall.GUID{0xE69653CA, 0xCF7F, 0x4F05,
		[8]byte{0xAA, 0x73, 0xCB, 0x83, 0x3F, 0xA9, 0x0A, 0xD4}}

	GUID_ENERGY_SAVER_BRIGHTNESS = syscall.GUID{0x13D09884, 0xF74E, 0x474A,
		[8]byte{0xA8, 0x52, 0xB6, 0xBD, 0xE8, 0xAD, 0x03, 0xA8}}

	GUID_ENERGY_SAVER_POLICY = syscall.GUID{0x5C5BB349, 0xAD29, 0x4EE2,
		[8]byte{0x9D, 0x0B, 0x2B, 0x25, 0x27, 0x0F, 0x7A, 0x81}}

	GUID_SYSTEM_BUTTON_SUBGROUP = syscall.GUID{0x4F971E89, 0xEEBD, 0x4455,
		[8]byte{0xA8, 0xDE, 0x9E, 0x59, 0x04, 0x0E, 0x73, 0x47}}

	GUID_POWERBUTTON_ACTION = syscall.GUID{0x7648EFA3, 0xDD9C, 0x4E3E,
		[8]byte{0xB5, 0x66, 0x50, 0xF9, 0x29, 0x38, 0x62, 0x80}}

	GUID_SLEEPBUTTON_ACTION = syscall.GUID{0x96996BC0, 0xAD50, 0x47EC,
		[8]byte{0x92, 0x3B, 0x6F, 0x41, 0x87, 0x4D, 0xD9, 0xEB}}

	GUID_USERINTERFACEBUTTON_ACTION = syscall.GUID{0xA7066653, 0x8D6C, 0x40A8,
		[8]byte{0x91, 0x0E, 0xA1, 0xF5, 0x4B, 0x84, 0xC7, 0xE5}}

	GUID_LIDCLOSE_ACTION = syscall.GUID{0x5CA83367, 0x6E45, 0x459F,
		[8]byte{0xA2, 0x7B, 0x47, 0x6B, 0x1D, 0x01, 0xC9, 0x36}}

	GUID_LIDOPEN_POWERSTATE = syscall.GUID{0x99FF10E7, 0x23B1, 0x4C07,
		[8]byte{0xA9, 0xD1, 0x5C, 0x32, 0x06, 0xD7, 0x41, 0xB4}}

	GUID_BATTERY_SUBGROUP = syscall.GUID{0xE73A048D, 0xBF27, 0x4F12,
		[8]byte{0x97, 0x31, 0x8B, 0x20, 0x76, 0xE8, 0x89, 0x1F}}

	GUID_BATTERY_DISCHARGE_ACTION_0 = syscall.GUID{0x637EA02F, 0xBBCB, 0x4015,
		[8]byte{0x8E, 0x2C, 0xA1, 0xC7, 0xB9, 0xC0, 0xB5, 0x46}}

	GUID_BATTERY_DISCHARGE_LEVEL_0 = syscall.GUID{0x9A66D8D7, 0x4FF7, 0x4EF9,
		[8]byte{0xB5, 0xA2, 0x5A, 0x32, 0x6C, 0xA2, 0xA4, 0x69}}

	GUID_BATTERY_DISCHARGE_FLAGS_0 = syscall.GUID{0x5DBB7C9F, 0x38E9, 0x40D2,
		[8]byte{0x97, 0x49, 0x4F, 0x8A, 0x0E, 0x9F, 0x64, 0x0F}}

	GUID_BATTERY_DISCHARGE_ACTION_1 = syscall.GUID{0xD8742DCB, 0x3E6A, 0x4B3C,
		[8]byte{0xB3, 0xFE, 0x37, 0x46, 0x23, 0xCD, 0xCF, 0x06}}

	GUID_BATTERY_DISCHARGE_LEVEL_1 = syscall.GUID{0x8183BA9A, 0xE910, 0x48DA,
		[8]byte{0x87, 0x69, 0x14, 0xAE, 0x6D, 0xC1, 0x17, 0x0A}}

	GUID_BATTERY_DISCHARGE_FLAGS_1 = syscall.GUID{0xBCDED951, 0x187B, 0x4D05,
		[8]byte{0xBC, 0xCC, 0xF7, 0xE5, 0x19, 0x60, 0xC2, 0x58}}

	GUID_BATTERY_DISCHARGE_ACTION_2 = syscall.GUID{0x421CBA38, 0x1A8E, 0x4881,
		[8]byte{0xAC, 0x89, 0xE3, 0x3A, 0x8B, 0x04, 0xEC, 0xE4}}

	GUID_BATTERY_DISCHARGE_LEVEL_2 = syscall.GUID{0x07A07CA2, 0xADAF, 0x40D7,
		[8]byte{0xB0, 0x77, 0x53, 0x3A, 0xAD, 0xED, 0x1B, 0xFA}}

	GUID_BATTERY_DISCHARGE_FLAGS_2 = syscall.GUID{0x7FD2F0C4, 0xFEB7, 0x4DA3,
		[8]byte{0x81, 0x17, 0xE3, 0xFB, 0xED, 0xC4, 0x65, 0x82}}

	GUID_BATTERY_DISCHARGE_ACTION_3 = syscall.GUID{0x80472613, 0x9780, 0x455E,
		[8]byte{0xB3, 0x08, 0x72, 0xD3, 0x00, 0x3C, 0xF2, 0xF8}}

	GUID_BATTERY_DISCHARGE_LEVEL_3 = syscall.GUID{0x58AFD5A6, 0xC2DD, 0x47D2,
		[8]byte{0x9F, 0xBF, 0xEF, 0x70, 0xCC, 0x5C, 0x59, 0x65}}

	GUID_BATTERY_DISCHARGE_FLAGS_3 = syscall.GUID{0x73613CCF, 0xDBFA, 0x4279,
		[8]byte{0x83, 0x56, 0x49, 0x35, 0xF6, 0xBF, 0x62, 0xF3}}

	GUID_PROCESSOR_SETTINGS_SUBGROUP = syscall.GUID{0x54533251, 0x82BE, 0x4824,
		[8]byte{0x96, 0xC1, 0x47, 0xB6, 0x0B, 0x74, 0x0D, 0x00}}

	GUID_PROCESSOR_THROTTLE_POLICY = syscall.GUID{0x57027304, 0x4AF6, 0x4104,
		[8]byte{0x92, 0x60, 0xE3, 0xD9, 0x52, 0x48, 0xFC, 0x36}}

	GUID_PROCESSOR_THROTTLE_MAXIMUM = syscall.GUID{0xBC5038F7, 0x23E0, 0x4960,
		[8]byte{0x96, 0xDA, 0x33, 0xAB, 0xAF, 0x59, 0x35, 0xEC}}

	GUID_PROCESSOR_THROTTLE_MAXIMUM_1 = syscall.GUID{0xBC5038F7, 0x23E0, 0x4960,
		[8]byte{0x96, 0xDA, 0x33, 0xAB, 0xAF, 0x59, 0x35, 0xED}}

	GUID_PROCESSOR_THROTTLE_MINIMUM = syscall.GUID{0x893DEE8E, 0x2BEF, 0x41E0,
		[8]byte{0x89, 0xC6, 0xB5, 0x5D, 0x09, 0x29, 0x96, 0x4C}}

	GUID_PROCESSOR_THROTTLE_MINIMUM_1 = syscall.GUID{0x893DEE8E, 0x2BEF, 0x41E0,
		[8]byte{0x89, 0xC6, 0xB5, 0x5D, 0x09, 0x29, 0x96, 0x4D}}

	GUID_PROCESSOR_FREQUENCY_LIMIT = syscall.GUID{0x75B0AE3F, 0xBCE0, 0x45A7,
		[8]byte{0x8C, 0x89, 0xC9, 0x61, 0x1C, 0x25, 0xE1, 0x00}}

	GUID_PROCESSOR_FREQUENCY_LIMIT_1 = syscall.GUID{0x75B0AE3F, 0xBCE0, 0x45A7,
		[8]byte{0x8C, 0x89, 0xC9, 0x61, 0x1C, 0x25, 0xE1, 0x01}}

	GUID_PROCESSOR_ALLOW_THROTTLING = syscall.GUID{0x3B04D4FD, 0x1CC7, 0x4F23,
		[8]byte{0xAB, 0x1C, 0xD1, 0x33, 0x78, 0x19, 0xC4, 0xBB}}

	GUID_PROCESSOR_IDLESTATE_POLICY = syscall.GUID{0x68F262A7, 0xF621, 0x4069,
		[8]byte{0xB9, 0xA5, 0x48, 0x74, 0x16, 0x9B, 0xE2, 0x3C}}

	GUID_PROCESSOR_PERFSTATE_POLICY = syscall.GUID{0xBBDC3814, 0x18E9, 0x4463,
		[8]byte{0x8A, 0x55, 0xD1, 0x97, 0x32, 0x7C, 0x45, 0xC0}}

	GUID_PROCESSOR_PERF_INCREASE_THRESHOLD = syscall.GUID{0x06CADF0E, 0x64ED, 0x448A,
		[8]byte{0x89, 0x27, 0xCE, 0x7B, 0xF9, 0x0E, 0xB3, 0x5D}}

	GUID_PROCESSOR_PERF_INCREASE_THRESHOLD_1 = syscall.GUID{0x06CADF0E, 0x64ED, 0x448A,
		[8]byte{0x89, 0x27, 0xCE, 0x7B, 0xF9, 0x0E, 0xB3, 0x5E}}

	GUID_PROCESSOR_PERF_DECREASE_THRESHOLD = syscall.GUID{0x12A0AB44, 0xFE28, 0x4FA9,
		[8]byte{0xB3, 0xBD, 0x4B, 0x64, 0xF4, 0x49, 0x60, 0xA6}}

	GUID_PROCESSOR_PERF_DECREASE_THRESHOLD_1 = syscall.GUID{0x12A0AB44, 0xFE28, 0x4FA9,
		[8]byte{0xB3, 0xBD, 0x4B, 0x64, 0xF4, 0x49, 0x60, 0xA7}}

	GUID_PROCESSOR_PERF_INCREASE_POLICY = syscall.GUID{0x465E1F50, 0xB610, 0x473A,
		[8]byte{0xAB, 0x58, 0x00, 0xD1, 0x07, 0x7D, 0xC4, 0x18}}

	GUID_PROCESSOR_PERF_INCREASE_POLICY_1 = syscall.GUID{0x465E1F50, 0xB610, 0x473A,
		[8]byte{0xAB, 0x58, 0x00, 0xD1, 0x07, 0x7D, 0xC4, 0x19}}

	GUID_PROCESSOR_PERF_DECREASE_POLICY = syscall.GUID{0x40FBEFC7, 0x2E9D, 0x4D25,
		[8]byte{0xA1, 0x85, 0x0C, 0xFD, 0x85, 0x74, 0xBA, 0xC6}}

	GUID_PROCESSOR_PERF_DECREASE_POLICY_1 = syscall.GUID{0x40FBEFC7, 0x2E9D, 0x4D25,
		[8]byte{0xA1, 0x85, 0x0C, 0xFD, 0x85, 0x74, 0xBA, 0xC7}}

	GUID_PROCESSOR_PERF_INCREASE_TIME = syscall.GUID{0x984CF492, 0x3BED, 0x4488,
		[8]byte{0xA8, 0xF9, 0x42, 0x86, 0xC9, 0x7B, 0xF5, 0xAA}}

	GUID_PROCESSOR_PERF_INCREASE_TIME_1 = syscall.GUID{0x984CF492, 0x3BED, 0x4488,
		[8]byte{0xA8, 0xF9, 0x42, 0x86, 0xC9, 0x7B, 0xF5, 0xAB}}

	GUID_PROCESSOR_PERF_DECREASE_TIME = syscall.GUID{0xD8EDEB9B, 0x95CF, 0x4F95,
		[8]byte{0xA7, 0x3C, 0xB0, 0x61, 0x97, 0x36, 0x93, 0xC8}}

	GUID_PROCESSOR_PERF_DECREASE_TIME_1 = syscall.GUID{0xD8EDEB9B, 0x95CF, 0x4F95,
		[8]byte{0xA7, 0x3C, 0xB0, 0x61, 0x97, 0x36, 0x93, 0xC9}}

	GUID_PROCESSOR_PERF_TIME_CHECK = syscall.GUID{0x4D2B0152, 0x7D5C, 0x498B,
		[8]byte{0x88, 0xE2, 0x34, 0x34, 0x53, 0x92, 0xA2, 0xC5}}

	GUID_PROCESSOR_PERF_BOOST_POLICY = syscall.GUID{0x45BCC044, 0xD885, 0x43E2,
		[8]byte{0x86, 0x05, 0xEE, 0x0E, 0xC6, 0xE9, 0x6B, 0x59}}

	GUID_PROCESSOR_PERF_BOOST_MODE = syscall.GUID{0xBE337238, 0x0D82, 0x4146,
		[8]byte{0xA9, 0x60, 0x4F, 0x37, 0x49, 0xD4, 0x70, 0xC7}}

	GUID_PROCESSOR_PERF_AUTONOMOUS_MODE = syscall.GUID{0x8BAA4A8A, 0x14C6, 0x4451,
		[8]byte{0x8E, 0x8B, 0x14, 0xBD, 0xBD, 0x19, 0x75, 0x37}}

	GUID_PROCESSOR_PERF_ENERGY_PERFORMANCE_PREFERENCE = syscall.GUID{0x36687F9E, 0xE3A5, 0x4DBF,
		[8]byte{0xB1, 0xDC, 0x15, 0xEB, 0x38, 0x1C, 0x68, 0x63}}

	GUID_PROCESSOR_PERF_ENERGY_PERFORMANCE_PREFERENCE_1 = syscall.GUID{0x36687F9E, 0xE3A5, 0x4DBF,
		[8]byte{0xB1, 0xDC, 0x15, 0xEB, 0x38, 0x1C, 0x68, 0x64}}

	GUID_PROCESSOR_PERF_AUTONOMOUS_ACTIVITY_WINDOW = syscall.GUID{0xCFEDA3D0, 0x7697, 0x4566,
		[8]byte{0xA9, 0x22, 0xA9, 0x08, 0x6C, 0xD4, 0x9D, 0xFA}}

	GUID_PROCESSOR_DUTY_CYCLING = syscall.GUID{0x4E4450B3, 0x6179, 0x4E91,
		[8]byte{0xB8, 0xF1, 0x5B, 0xB9, 0x93, 0x8F, 0x81, 0xA1}}

	GUID_PROCESSOR_IDLE_ALLOW_SCALING = syscall.GUID{0x6C2993B0, 0x8F48, 0x481F,
		[8]byte{0xBC, 0xC6, 0x00, 0xDD, 0x27, 0x42, 0xAA, 0x06}}

	GUID_PROCESSOR_IDLE_DISABLE = syscall.GUID{0x5D76A2CA, 0xE8C0, 0x402F,
		[8]byte{0xA1, 0x33, 0x21, 0x58, 0x49, 0x2D, 0x58, 0xAD}}

	GUID_PROCESSOR_IDLE_STATE_MAXIMUM = syscall.GUID{0x9943E905, 0x9A30, 0x4EC1,
		[8]byte{0x9B, 0x99, 0x44, 0xDD, 0x3B, 0x76, 0xF7, 0xA2}}

	GUID_PROCESSOR_IDLE_TIME_CHECK = syscall.GUID{0xC4581C31, 0x89AB, 0x4597,
		[8]byte{0x8E, 0x2B, 0x9C, 0x9C, 0xAB, 0x44, 0x0E, 0x6B}}

	GUID_PROCESSOR_IDLE_DEMOTE_THRESHOLD = syscall.GUID{0x4B92D758, 0x5A24, 0x4851,
		[8]byte{0xA4, 0x70, 0x81, 0x5D, 0x78, 0xAE, 0xE1, 0x19}}

	GUID_PROCESSOR_IDLE_PROMOTE_THRESHOLD = syscall.GUID{0x7B224883, 0xB3CC, 0x4D79,
		[8]byte{0x81, 0x9F, 0x83, 0x74, 0x15, 0x2C, 0xBE, 0x7C}}

	GUID_PROCESSOR_CORE_PARKING_INCREASE_THRESHOLD = syscall.GUID{0xDF142941, 0x20F3, 0x4EDF,
		[8]byte{0x9A, 0x4A, 0x9C, 0x83, 0xD3, 0xD7, 0x17, 0xD1}}

	GUID_PROCESSOR_CORE_PARKING_DECREASE_THRESHOLD = syscall.GUID{0x68DD2F27, 0xA4CE, 0x4E11,
		[8]byte{0x84, 0x87, 0x37, 0x94, 0xE4, 0x13, 0x5D, 0xFA}}

	GUID_PROCESSOR_CORE_PARKING_INCREASE_POLICY = syscall.GUID{0xC7BE0679, 0x2817, 0x4D69,
		[8]byte{0x9D, 0x02, 0x51, 0x9A, 0x53, 0x7E, 0xD0, 0xC6}}

	GUID_PROCESSOR_CORE_PARKING_DECREASE_POLICY = syscall.GUID{0x71021B41, 0xC749, 0x4D21,
		[8]byte{0xBE, 0x74, 0xA0, 0x0F, 0x33, 0x5D, 0x58, 0x2B}}

	GUID_PROCESSOR_CORE_PARKING_MAX_CORES = syscall.GUID{0xEA062031, 0x0E34, 0x4FF1,
		[8]byte{0x9B, 0x6D, 0xEB, 0x10, 0x59, 0x33, 0x40, 0x28}}

	GUID_PROCESSOR_CORE_PARKING_MAX_CORES_1 = syscall.GUID{0xEA062031, 0x0E34, 0x4FF1,
		[8]byte{0x9B, 0x6D, 0xEB, 0x10, 0x59, 0x33, 0x40, 0x29}}

	GUID_PROCESSOR_CORE_PARKING_MIN_CORES = syscall.GUID{0x0CC5B647, 0xC1DF, 0x4637,
		[8]byte{0x89, 0x1A, 0xDE, 0xC3, 0x5C, 0x31, 0x85, 0x83}}

	GUID_PROCESSOR_CORE_PARKING_MIN_CORES_1 = syscall.GUID{0x0CC5B647, 0xC1DF, 0x4637,
		[8]byte{0x89, 0x1A, 0xDE, 0xC3, 0x5C, 0x31, 0x85, 0x84}}

	GUID_PROCESSOR_CORE_PARKING_INCREASE_TIME = syscall.GUID{0x2DDD5A84, 0x5A71, 0x437E,
		[8]byte{0x91, 0x2A, 0xDB, 0x0B, 0x8C, 0x78, 0x87, 0x32}}

	GUID_PROCESSOR_CORE_PARKING_DECREASE_TIME = syscall.GUID{0xDFD10D17, 0xD5EB, 0x45DD,
		[8]byte{0x87, 0x7A, 0x9A, 0x34, 0xDD, 0xD1, 0x5C, 0x82}}

	GUID_PROCESSOR_CORE_PARKING_AFFINITY_HISTORY_DECREASE_FACTOR = syscall.GUID{0x8F7B45E3, 0xC393, 0x480A,
		[8]byte{0x87, 0x8C, 0xF6, 0x7A, 0xC3, 0xD0, 0x70, 0x82}}

	GUID_PROCESSOR_CORE_PARKING_AFFINITY_HISTORY_THRESHOLD = syscall.GUID{0x5B33697B, 0xE89D, 0x4D38,
		[8]byte{0xAA, 0x46, 0x9E, 0x7D, 0xFB, 0x7C, 0xD2, 0xF9}}

	GUID_PROCESSOR_CORE_PARKING_AFFINITY_WEIGHTING = syscall.GUID{0xE70867F1, 0xFA2F, 0x4F4E,
		[8]byte{0xAE, 0xA1, 0x4D, 0x8A, 0x0B, 0xA2, 0x3B, 0x20}}

	GUID_PROCESSOR_CORE_PARKING_OVER_UTILIZATION_HISTORY_DECREASE_FACTOR = syscall.GUID{0x1299023C, 0xBC28, 0x4F0A,
		[8]byte{0x81, 0xEC, 0xD3, 0x29, 0x5A, 0x8D, 0x81, 0x5D}}

	GUID_PROCESSOR_CORE_PARKING_OVER_UTILIZATION_HISTORY_THRESHOLD = syscall.GUID{0x9AC18E92, 0xAA3C, 0x4E27,
		[8]byte{0xB3, 0x07, 0x01, 0xAE, 0x37, 0x30, 0x71, 0x29}}

	GUID_PROCESSOR_CORE_PARKING_OVER_UTILIZATION_WEIGHTING = syscall.GUID{0x8809C2D8, 0xB155, 0x42D4,
		[8]byte{0xBC, 0xDA, 0x0D, 0x34, 0x56, 0x51, 0xB1, 0xDB}}

	GUID_PROCESSOR_CORE_PARKING_OVER_UTILIZATION_THRESHOLD = syscall.GUID{0x943C8CB6, 0x6F93, 0x4227,
		[8]byte{0xAD, 0x87, 0xE9, 0xA3, 0xFE, 0xEC, 0x08, 0xD1}}

	GUID_PROCESSOR_PARKING_CORE_OVERRIDE = syscall.GUID{0xA55612AA, 0xF624, 0x42C6,
		[8]byte{0xA4, 0x43, 0x73, 0x97, 0xD0, 0x64, 0xC0, 0x4F}}

	GUID_PROCESSOR_PARKING_PERF_STATE = syscall.GUID{0x447235C7, 0x6A8D, 0x4CC0,
		[8]byte{0x8E, 0x24, 0x9E, 0xAF, 0x70, 0xB9, 0x6E, 0x2B}}

	GUID_PROCESSOR_PARKING_PERF_STATE_1 = syscall.GUID{0x447235C7, 0x6A8D, 0x4CC0,
		[8]byte{0x8E, 0x24, 0x9E, 0xAF, 0x70, 0xB9, 0x6E, 0x2C}}

	GUID_PROCESSOR_PARKING_CONCURRENCY_THRESHOLD = syscall.GUID{0x2430AB6F, 0xA520, 0x44A2,
		[8]byte{0x96, 0x01, 0xF7, 0xF2, 0x3B, 0x51, 0x34, 0xB1}}

	GUID_PROCESSOR_PARKING_HEADROOM_THRESHOLD = syscall.GUID{0xF735A673, 0x2066, 0x4F80,
		[8]byte{0xA0, 0xC5, 0xDD, 0xEE, 0x0C, 0xF1, 0xBF, 0x5D}}

	GUID_PROCESSOR_PARKING_DISTRIBUTION_THRESHOLD = syscall.GUID{0x4BDAF4E9, 0xD103, 0x46D7,
		[8]byte{0xA5, 0xF0, 0x62, 0x80, 0x12, 0x16, 0x16, 0xEF}}

	GUID_PROCESSOR_SOFT_PARKING_LATENCY = syscall.GUID{0x97CFAC41, 0x2217, 0x47EB,
		[8]byte{0x99, 0x2D, 0x61, 0x8B, 0x19, 0x77, 0xC9, 0x07}}

	GUID_PROCESSOR_PERF_HISTORY = syscall.GUID{0x7D24BAA7, 0x0B84, 0x480F,
		[8]byte{0x84, 0x0C, 0x1B, 0x07, 0x43, 0xC0, 0x0F, 0x5F}}

	GUID_PROCESSOR_PERF_HISTORY_1 = syscall.GUID{0x7D24BAA7, 0x0B84, 0x480F,
		[8]byte{0x84, 0x0C, 0x1B, 0x07, 0x43, 0xC0, 0x0F, 0x60}}

	GUID_PROCESSOR_PERF_INCREASE_HISTORY = syscall.GUID{0x99B3EF01, 0x752F, 0x46A1,
		[8]byte{0x80, 0xFB, 0x77, 0x30, 0x01, 0x1F, 0x23, 0x54}}

	GUID_PROCESSOR_PERF_DECREASE_HISTORY = syscall.GUID{0x0300F6F8, 0xABD6, 0x45A9,
		[8]byte{0xB7, 0x4F, 0x49, 0x08, 0x69, 0x1A, 0x40, 0xB5}}

	GUID_PROCESSOR_PERF_CORE_PARKING_HISTORY = syscall.GUID{0x77D7F282, 0x8F1A, 0x42CD,
		[8]byte{0x85, 0x37, 0x45, 0x45, 0x0A, 0x83, 0x9B, 0xE8}}

	GUID_PROCESSOR_PERF_LATENCY_HINT = syscall.GUID{0x0822DF31, 0x9C83, 0x441C,
		[8]byte{0xA0, 0x79, 0x0D, 0xE4, 0xCF, 0x00, 0x9C, 0x7B}}

	GUID_PROCESSOR_PERF_LATENCY_HINT_PERF = syscall.GUID{0x619B7505, 0x003B, 0x4E82,
		[8]byte{0xB7, 0xA6, 0x4D, 0xD2, 0x9C, 0x30, 0x09, 0x71}}

	GUID_PROCESSOR_PERF_LATENCY_HINT_PERF_1 = syscall.GUID{0x619B7505, 0x003B, 0x4E82,
		[8]byte{0xB7, 0xA6, 0x4D, 0xD2, 0x9C, 0x30, 0x09, 0x72}}

	GUID_PROCESSOR_LATENCY_HINT_MIN_UNPARK = syscall.GUID{0x616CDAA5, 0x695E, 0x4545,
		[8]byte{0x97, 0xAD, 0x97, 0xDC, 0x2D, 0x1B, 0xDD, 0x88}}

	GUID_PROCESSOR_LATENCY_HINT_MIN_UNPARK_1 = syscall.GUID{0x616CDAA5, 0x695E, 0x4545,
		[8]byte{0x97, 0xAD, 0x97, 0xDC, 0x2D, 0x1B, 0xDD, 0x89}}

	GUID_PROCESSOR_DISTRIBUTE_UTILITY = syscall.GUID{0xE0007330, 0xF589, 0x42ED,
		[8]byte{0xA4, 0x01, 0x5D, 0xDB, 0x10, 0xE7, 0x85, 0xD3}}

	GUID_PROCESSOR_HETEROGENEOUS_POLICY = syscall.GUID{0x7F2F5CFA, 0xF10C, 0x4823,
		[8]byte{0xB5, 0xE1, 0xE9, 0x3A, 0xE8, 0x5F, 0x46, 0xB5}}

	GUID_PROCESSOR_HETERO_DECREASE_TIME = syscall.GUID{0x7F2492B6, 0x60B1, 0x45E5,
		[8]byte{0xAE, 0x55, 0x77, 0x3F, 0x8C, 0xD5, 0xCA, 0xEC}}

	GUID_PROCESSOR_HETERO_INCREASE_TIME = syscall.GUID{0x4009EFA7, 0xE72D, 0x4CBA,
		[8]byte{0x9E, 0xDF, 0x91, 0x08, 0x4E, 0xA8, 0xCB, 0xC3}}

	GUID_PROCESSOR_HETERO_DECREASE_THRESHOLD = syscall.GUID{0xF8861C27, 0x95E7, 0x475C,
		[8]byte{0x86, 0x5B, 0x13, 0xC0, 0xCB, 0x3F, 0x9D, 0x6B}}

	GUID_PROCESSOR_HETERO_INCREASE_THRESHOLD = syscall.GUID{0xB000397D, 0x9B0B, 0x483D,
		[8]byte{0x98, 0xC9, 0x69, 0x2A, 0x60, 0x60, 0xCF, 0xBF}}

	GUID_PROCESSOR_CLASS0_FLOOR_PERF = syscall.GUID{0xFDDC842B, 0x8364, 0x4EDC,
		[8]byte{0x94, 0xCF, 0xC1, 0x7F, 0x60, 0xDE, 0x1C, 0x80}}

	GUID_PROCESSOR_CLASS1_INITIAL_PERF = syscall.GUID{0x1FACFC65, 0xA930, 0x4BC5,
		[8]byte{0x9F, 0x38, 0x50, 0x4E, 0xC0, 0x97, 0xBB, 0xC0}}

	GUID_PROCESSOR_THREAD_SCHEDULING_POLICY = syscall.GUID{0x93B8B6DC, 0x0698, 0x4D1C,
		[8]byte{0x9E, 0xE4, 0x06, 0x44, 0xE9, 0x00, 0xC8, 0x5D}}

	GUID_PROCESSOR_SHORT_THREAD_SCHEDULING_POLICY = syscall.GUID{0xBAE08B81, 0x2D5E, 0x4688,
		[8]byte{0xAD, 0x6A, 0x13, 0x24, 0x33, 0x56, 0x65, 0x4B}}

	GUID_PROCESSOR_SHORT_THREAD_RUNTIME_THRESHOLD = syscall.GUID{0xD92998C2, 0x6A48, 0x49CA,
		[8]byte{0x85, 0xD4, 0x8C, 0xCE, 0xEC, 0x29, 0x45, 0x70}}

	GUID_SYSTEM_COOLING_POLICY = syscall.GUID{0x94D3A615, 0xA899, 0x4AC5,
		[8]byte{0xAE, 0x2B, 0xE4, 0xD8, 0xF6, 0x34, 0x36, 0x7F}}

	GUID_PROCESSOR_RESPONSIVENESS_DISABLE_THRESHOLD = syscall.GUID{0x38B8383D, 0xCCE0, 0x4C79,
		[8]byte{0x9E, 0x3E, 0x56, 0xA4, 0xF1, 0x7C, 0xC4, 0x80}}

	GUID_PROCESSOR_RESPONSIVENESS_DISABLE_THRESHOLD_1 = syscall.GUID{0x38B8383D, 0xCCE0, 0x4C79,
		[8]byte{0x9E, 0x3E, 0x56, 0xA4, 0xF1, 0x7C, 0xC4, 0x81}}

	GUID_PROCESSOR_RESPONSIVENESS_ENABLE_THRESHOLD = syscall.GUID{0x3D44E256, 0x7222, 0x4415,
		[8]byte{0xA9, 0xED, 0x9C, 0x45, 0xFA, 0x3D, 0xD8, 0x30}}

	GUID_PROCESSOR_RESPONSIVENESS_ENABLE_THRESHOLD_1 = syscall.GUID{0x3D44E256, 0x7222, 0x4415,
		[8]byte{0xA9, 0xED, 0x9C, 0x45, 0xFA, 0x3D, 0xD8, 0x31}}

	GUID_PROCESSOR_RESPONSIVENESS_DISABLE_TIME = syscall.GUID{0xF565999F, 0x3FB0, 0x411A,
		[8]byte{0xA2, 0x26, 0x3F, 0x01, 0x98, 0xDE, 0xC1, 0x30}}

	GUID_PROCESSOR_RESPONSIVENESS_DISABLE_TIME_1 = syscall.GUID{0xF565999F, 0x3FB0, 0x411A,
		[8]byte{0xA2, 0x26, 0x3F, 0x01, 0x98, 0xDE, 0xC1, 0x31}}

	GUID_PROCESSOR_RESPONSIVENESS_ENABLE_TIME = syscall.GUID{0x3D915188, 0x7830, 0x49AE,
		[8]byte{0xA7, 0x9A, 0x0F, 0xB0, 0xA1, 0xE5, 0xA2, 0x00}}

	GUID_PROCESSOR_RESPONSIVENESS_ENABLE_TIME_1 = syscall.GUID{0x3D915188, 0x7830, 0x49AE,
		[8]byte{0xA7, 0x9A, 0x0F, 0xB0, 0xA1, 0xE5, 0xA2, 0x01}}

	GUID_PROCESSOR_RESPONSIVENESS_EPP_CEILING = syscall.GUID{0x4427C73B, 0x9756, 0x4A5C,
		[8]byte{0xB8, 0x4B, 0xC7, 0xBD, 0xA7, 0x9C, 0x73, 0x20}}

	GUID_PROCESSOR_RESPONSIVENESS_EPP_CEILING_1 = syscall.GUID{0x4427C73B, 0x9756, 0x4A5C,
		[8]byte{0xB8, 0x4B, 0xC7, 0xBD, 0xA7, 0x9C, 0x73, 0x21}}

	GUID_PROCESSOR_RESPONSIVENESS_PERF_FLOOR = syscall.GUID{0xCE8E92EE, 0x6A86, 0x4572,
		[8]byte{0xBF, 0xE0, 0x20, 0xC2, 0x1D, 0x03, 0xCD, 0x40}}

	GUID_PROCESSOR_RESPONSIVENESS_PERF_FLOOR_1 = syscall.GUID{0xCE8E92EE, 0x6A86, 0x4572,
		[8]byte{0xBF, 0xE0, 0x20, 0xC2, 0x1D, 0x03, 0xCD, 0x41}}

	GUID_LOCK_CONSOLE_ON_WAKE = syscall.GUID{0x0E796BDB, 0x100D, 0x47D6,
		[8]byte{0xA2, 0xD5, 0xF7, 0xD2, 0xDA, 0xA5, 0x1F, 0x51}}

	GUID_DEVICE_IDLE_POLICY = syscall.GUID{0x4FAAB71A, 0x92E5, 0x4726,
		[8]byte{0xB5, 0x31, 0x22, 0x45, 0x59, 0x67, 0x2D, 0x19}}

	GUID_CONNECTIVITY_IN_STANDBY = syscall.GUID{0xF15576E8, 0x98B7, 0x4186,
		[8]byte{0xB9, 0x44, 0xEA, 0xFA, 0x66, 0x44, 0x02, 0xD9}}

	GUID_DISCONNECTED_STANDBY_MODE = syscall.GUID{0x68AFB2D9, 0xEE95, 0x47A8,
		[8]byte{0x8F, 0x50, 0x41, 0x15, 0x08, 0x80, 0x73, 0xB1}}

	GUID_ACDC_POWER_SOURCE = syscall.GUID{0x5D3E9A59, 0xE9D5, 0x4B00,
		[8]byte{0xA6, 0xBD, 0xFF, 0x34, 0xFF, 0x51, 0x65, 0x48}}

	GUID_LIDSWITCH_STATE_CHANGE = syscall.GUID{0xBA3E0F4D, 0xB817, 0x4094,
		[8]byte{0xA2, 0xD1, 0xD5, 0x63, 0x79, 0xE6, 0xA0, 0xF3}}

	GUID_LIDSWITCH_STATE_RELIABILITY = syscall.GUID{0xAE4C4FF1, 0xD361, 0x43F4,
		[8]byte{0x80, 0xAA, 0xBB, 0xB6, 0xEB, 0x03, 0xDE, 0x94}}

	GUID_BATTERY_PERCENTAGE_REMAINING = syscall.GUID{0xA7AD8041, 0xB45A, 0x4CAE,
		[8]byte{0x87, 0xA3, 0xEE, 0xCB, 0xB4, 0x68, 0xA9, 0xE1}}

	GUID_BATTERY_COUNT = syscall.GUID{0x7D263F15, 0xFCA4, 0x49E5,
		[8]byte{0x85, 0x4B, 0xA9, 0xF2, 0xBF, 0xBD, 0x5C, 0x24}}

	GUID_GLOBAL_USER_PRESENCE = syscall.GUID{0x786E8A1D, 0xB427, 0x4344,
		[8]byte{0x92, 0x07, 0x09, 0xE7, 0x0B, 0xDC, 0xBE, 0xA9}}

	GUID_SESSION_DISPLAY_STATUS = syscall.GUID{0x2B84C20E, 0xAD23, 0x4DDF,
		[8]byte{0x93, 0xDB, 0x05, 0xFF, 0xBD, 0x7E, 0xFC, 0xA5}}

	GUID_SESSION_USER_PRESENCE = syscall.GUID{0x3C0F4548, 0xC03F, 0x4C4D,
		[8]byte{0xB9, 0xF2, 0x23, 0x7E, 0xDE, 0x68, 0x63, 0x76}}

	GUID_IDLE_BACKGROUND_TASK = syscall.GUID{0x515C31D8, 0xF734, 0x163D,
		[8]byte{0xA0, 0xFD, 0x11, 0xA0, 0x8C, 0x91, 0xE8, 0xF1}}

	GUID_BACKGROUND_TASK_NOTIFICATION = syscall.GUID{0xCF23F240, 0x2A54, 0x48D8,
		[8]byte{0xB1, 0x14, 0xDE, 0x15, 0x18, 0xFF, 0x05, 0x2E}}

	GUID_APPLAUNCH_BUTTON = syscall.GUID{0x1A689231, 0x7399, 0x4E9A,
		[8]byte{0x8F, 0x99, 0xB7, 0x1F, 0x99, 0x9D, 0xB3, 0xFA}}

	GUID_PCIEXPRESS_SETTINGS_SUBGROUP = syscall.GUID{0x501A4D13, 0x42AF, 0x4429,
		[8]byte{0x9F, 0xD1, 0xA8, 0x21, 0x8C, 0x26, 0x8E, 0x20}}

	GUID_PCIEXPRESS_ASPM_POLICY = syscall.GUID{0xEE12F906, 0xD277, 0x404B,
		[8]byte{0xB6, 0xDA, 0xE5, 0xFA, 0x1A, 0x57, 0x6D, 0xF5}}

	GUID_ENABLE_SWITCH_FORCED_SHUTDOWN = syscall.GUID{0x833A6B62, 0xDFA4, 0x46D1,
		[8]byte{0x82, 0xF8, 0xE0, 0x9E, 0x34, 0xD0, 0x29, 0xD6}}

	GUID_INTSTEER_SUBGROUP = syscall.GUID{0x48672F38, 0x7A9A, 0x4BB2,
		[8]byte{0x8B, 0xF8, 0x3D, 0x85, 0xBE, 0x19, 0xDE, 0x4E}}

	GUID_INTSTEER_MODE = syscall.GUID{0x2BFC24F9, 0x5EA2, 0x4801,
		[8]byte{0x82, 0x13, 0x3D, 0xBA, 0xE0, 0x1A, 0xA3, 0x9D}}

	GUID_INTSTEER_LOAD_PER_PROC_TRIGGER = syscall.GUID{0x73CDE64D, 0xD720, 0x4BB2,
		[8]byte{0xA8, 0x60, 0xC7, 0x55, 0xAF, 0xE7, 0x7E, 0xF2}}

	GUID_INTSTEER_TIME_UNPARK_TRIGGER = syscall.GUID{0xD6BA4903, 0x386F, 0x4C2C,
		[8]byte{0x8A, 0xDB, 0x5C, 0x21, 0xB3, 0x32, 0x8D, 0x25}}

	GUID_GRAPHICS_SUBGROUP = syscall.GUID{0x5FB4938D, 0x1EE8, 0x4B0F,
		[8]byte{0x9A, 0x3C, 0x50, 0x36, 0xB0, 0xAB, 0x99, 0x5C}}

	GUID_GPU_PREFERENCE_POLICY = syscall.GUID{0xDD848B2A, 0x8A5D, 0x4451,
		[8]byte{0x9A, 0xE2, 0x39, 0xCD, 0x41, 0x65, 0x8F, 0x6C}}

	GUID_MIXED_REALITY_MODE = syscall.GUID{0x1E626B4E, 0xCF04, 0x4F8D,
		[8]byte{0x9C, 0xC7, 0xC9, 0x7C, 0x5B, 0x0F, 0x23, 0x91}}

	GUID_SPR_ACTIVE_SESSION_CHANGE = syscall.GUID{0x0E24CE38, 0xC393, 0x4742,
		[8]byte{0xBD, 0xB1, 0x74, 0x4F, 0x4B, 0x9E, 0xE0, 0x8E}}

	PPM_PERFSTATE_CHANGE_GUID = syscall.GUID{0xA5B32DDD, 0x7F39, 0x4ABC,
		[8]byte{0xB8, 0x92, 0x90, 0x0E, 0x43, 0xB5, 0x9E, 0xBB}}

	PPM_PERFSTATE_DOMAIN_CHANGE_GUID = syscall.GUID{0x995E6B7F, 0xD653, 0x497A,
		[8]byte{0xB9, 0x78, 0x36, 0xA3, 0x0C, 0x29, 0xBF, 0x01}}

	PPM_IDLESTATE_CHANGE_GUID = syscall.GUID{0x4838FE4F, 0xF71C, 0x4E51,
		[8]byte{0x9E, 0xCC, 0x84, 0x30, 0xA7, 0xAC, 0x4C, 0x6C}}

	PPM_PERFSTATES_DATA_GUID = syscall.GUID{0x5708CC20, 0x7D40, 0x4BF4,
		[8]byte{0xB4, 0xAA, 0x2B, 0x01, 0x33, 0x8D, 0x01, 0x26}}

	PPM_IDLESTATES_DATA_GUID = syscall.GUID{0xBA138E10, 0xE250, 0x4AD7,
		[8]byte{0x86, 0x16, 0xCF, 0x1A, 0x7A, 0xD4, 0x10, 0xE7}}

	PPM_IDLE_ACCOUNTING_GUID = syscall.GUID{0xE2A26F78, 0xAE07, 0x4EE0,
		[8]byte{0xA3, 0x0F, 0xCE, 0x54, 0xF5, 0x5A, 0x94, 0xCD}}

	PPM_IDLE_ACCOUNTING_EX_GUID = syscall.GUID{0xD67ABD39, 0x81F8, 0x4A5E,
		[8]byte{0x81, 0x52, 0x72, 0xE3, 0x1E, 0xC9, 0x12, 0xEE}}

	PPM_THERMALCONSTRAINT_GUID = syscall.GUID{0xA852C2C8, 0x1A4C, 0x423B,
		[8]byte{0x8C, 0x2C, 0xF3, 0x0D, 0x82, 0x93, 0x1A, 0x88}}

	PPM_PERFMON_PERFSTATE_GUID = syscall.GUID{0x7FD18652, 0x0CFE, 0x40D2,
		[8]byte{0xB0, 0xA1, 0x0B, 0x06, 0x6A, 0x87, 0x75, 0x9E}}

	PPM_THERMAL_POLICY_CHANGE_GUID = syscall.GUID{0x48F377B8, 0x6880, 0x4C7B,
		[8]byte{0x8B, 0xDC, 0x38, 0x01, 0x76, 0xC6, 0x65, 0x4D}}

	GUID_IO_VOLUME_CHANGE = syscall.GUID{0x7373654A, 0x812A, 0x11D0,
		[8]byte{0xBE, 0xC7, 0x08, 0x00, 0x2B, 0xE2, 0x09, 0x2F}}

	GUID_IO_VOLUME_DISMOUNT = syscall.GUID{0xD16A55E8, 0x1059, 0x11D2,
		[8]byte{0x8F, 0xFD, 0x00, 0xA0, 0xC9, 0xA0, 0x6D, 0x32}}

	GUID_IO_VOLUME_DISMOUNT_FAILED = syscall.GUID{0xE3C5B178, 0x105D, 0x11D2,
		[8]byte{0x8F, 0xFD, 0x00, 0xA0, 0xC9, 0xA0, 0x6D, 0x32}}

	GUID_IO_VOLUME_MOUNT = syscall.GUID{0xB5804878, 0x1A96, 0x11D2,
		[8]byte{0x8F, 0xFD, 0x00, 0xA0, 0xC9, 0xA0, 0x6D, 0x32}}

	GUID_IO_VOLUME_LOCK = syscall.GUID{0x50708874, 0xC9AF, 0x11D1,
		[8]byte{0x8F, 0xEF, 0x00, 0xA0, 0xC9, 0xA0, 0x6D, 0x32}}

	GUID_IO_VOLUME_LOCK_FAILED = syscall.GUID{0xAE2EED10, 0x0BA8, 0x11D2,
		[8]byte{0x8F, 0xFB, 0x00, 0xA0, 0xC9, 0xA0, 0x6D, 0x32}}

	GUID_IO_VOLUME_UNLOCK = syscall.GUID{0x9A8C3D68, 0xD0CB, 0x11D1,
		[8]byte{0x8F, 0xEF, 0x00, 0xA0, 0xC9, 0xA0, 0x6D, 0x32}}

	GUID_IO_VOLUME_NAME_CHANGE = syscall.GUID{0x2DE97F83, 0x4C06, 0x11D2,
		[8]byte{0xA5, 0x32, 0x00, 0x60, 0x97, 0x13, 0x05, 0x5A}}

	GUID_IO_VOLUME_NEED_CHKDSK = syscall.GUID{0x799A0960, 0x0A0B, 0x4E03,
		[8]byte{0xAD, 0x88, 0x2F, 0xA7, 0xC6, 0xCE, 0x74, 0x8A}}

	GUID_IO_VOLUME_WORM_NEAR_FULL = syscall.GUID{0xF3BFFF82, 0xF3DE, 0x48D2,
		[8]byte{0xAF, 0x95, 0x45, 0x7F, 0x80, 0xB7, 0x63, 0xF2}}

	GUID_IO_VOLUME_WEARING_OUT = syscall.GUID{0x873113CA, 0x1486, 0x4508,
		[8]byte{0x82, 0xAC, 0xC3, 0xB2, 0xE5, 0x29, 0x7A, 0xAA}}

	GUID_IO_VOLUME_FORCE_CLOSED = syscall.GUID{0x411AD84F, 0x433E, 0x4DC2,
		[8]byte{0xA5, 0xAE, 0x4A, 0x2D, 0x1A, 0x2D, 0xE6, 0x54}}

	GUID_IO_VOLUME_INFO_MAKE_COMPAT = syscall.GUID{0x3AB9A0D2, 0xEF80, 0x45CF,
		[8]byte{0x8C, 0xDC, 0xCB, 0xE0, 0x2A, 0x21, 0x29, 0x06}}

	GUID_IO_VOLUME_PREPARING_EJECT = syscall.GUID{0xC79EB16E, 0x0DAC, 0x4E7A,
		[8]byte{0xA8, 0x6C, 0xB2, 0x5C, 0xEE, 0xAA, 0x88, 0xF6}}

	GUID_IO_VOLUME_BACKGROUND_FORMAT = syscall.GUID{0xA2E5FC86, 0xD5CD, 0x4038,
		[8]byte{0xB2, 0xE3, 0x44, 0x45, 0x06, 0x5C, 0x23, 0x77}}

	GUID_IO_VOLUME_PHYSICAL_CONFIGURATION_CHANGE = syscall.GUID{0x2DE97F84, 0x4C06, 0x11D2,
		[8]byte{0xA5, 0x32, 0x00, 0x60, 0x97, 0x13, 0x05, 0x5A}}

	GUID_IO_VOLUME_UNIQUE_ID_CHANGE = syscall.GUID{0xAF39DA42, 0x6622, 0x41F5,
		[8]byte{0x97, 0x0B, 0x13, 0x9D, 0x09, 0x2F, 0xA3, 0xD9}}

	GUID_IO_VOLUME_FVE_STATUS_CHANGE = syscall.GUID{0x062998B2, 0xEE1F, 0x4B6A,
		[8]byte{0xB8, 0x57, 0xE7, 0x6C, 0xBB, 0xE9, 0xA6, 0xDA}}

	GUID_IO_VOLUME_DEVICE_INTERFACE = syscall.GUID{0x53F5630D, 0xB6BF, 0x11D0,
		[8]byte{0x94, 0xF2, 0x00, 0xA0, 0xC9, 0x1E, 0xFB, 0x8B}}

	GUID_IO_VOLUME_CHANGE_SIZE = syscall.GUID{0x3A1625BE, 0xAD03, 0x49F1,
		[8]byte{0x8E, 0xF8, 0x6B, 0xBA, 0xC1, 0x82, 0xD1, 0xFD}}

	GUID_IO_MEDIA_ARRIVAL = syscall.GUID{0xD07433C0, 0xA98E, 0x11D2,
		[8]byte{0x91, 0x7A, 0x00, 0xA0, 0xC9, 0x06, 0x8F, 0xF3}}

	GUID_IO_MEDIA_REMOVAL = syscall.GUID{0xD07433C1, 0xA98E, 0x11D2,
		[8]byte{0x91, 0x7A, 0x00, 0xA0, 0xC9, 0x06, 0x8F, 0xF3}}

	GUID_IO_CDROM_EXCLUSIVE_LOCK = syscall.GUID{0xBC56C139, 0x7A10, 0x47EE,
		[8]byte{0xA2, 0x94, 0x4C, 0x6A, 0x38, 0xF0, 0x14, 0x9A}}

	GUID_IO_CDROM_EXCLUSIVE_UNLOCK = syscall.GUID{0xA3B6D27D, 0x5E35, 0x4885,
		[8]byte{0x81, 0xE5, 0xEE, 0x18, 0xC0, 0x0E, 0xD7, 0x79}}

	GUID_IO_DEVICE_BECOMING_READY = syscall.GUID{0xD07433F0, 0xA98E, 0x11D2,
		[8]byte{0x91, 0x7A, 0x00, 0xA0, 0xC9, 0x06, 0x8F, 0xF3}}

	GUID_IO_DEVICE_EXTERNAL_REQUEST = syscall.GUID{0xD07433D0, 0xA98E, 0x11D2,
		[8]byte{0x91, 0x7A, 0x00, 0xA0, 0xC9, 0x06, 0x8F, 0xF3}}

	GUID_IO_MEDIA_EJECT_REQUEST = syscall.GUID{0xD07433D1, 0xA98E, 0x11D2,
		[8]byte{0x91, 0x7A, 0x00, 0xA0, 0xC9, 0x06, 0x8F, 0xF3}}

	GUID_IO_DRIVE_REQUIRES_CLEANING = syscall.GUID{0x7207877C, 0x90ED, 0x44E5,
		[8]byte{0xA0, 0x00, 0x81, 0x42, 0x8D, 0x4C, 0x79, 0xBB}}

	GUID_IO_TAPE_ERASE = syscall.GUID{0x852D11EB, 0x4BB8, 0x4507,
		[8]byte{0x9D, 0x9B, 0x41, 0x7C, 0xC2, 0xB1, 0xB4, 0x38}}

	GUID_DEVICE_EVENT_RBC = syscall.GUID{0xD0744792, 0xA98E, 0x11D2,
		[8]byte{0x91, 0x7A, 0x00, 0xA0, 0xC9, 0x06, 0x8F, 0xF3}}

	GUID_IO_DISK_CLONE_ARRIVAL = syscall.GUID{0x6A61885B, 0x7C39, 0x43DD,
		[8]byte{0x9B, 0x56, 0xB8, 0xAC, 0x22, 0xA5, 0x49, 0xAA}}

	GUID_IO_DISK_LAYOUT_CHANGE = syscall.GUID{0x11DFF54C, 0x8469, 0x41F9,
		[8]byte{0xB3, 0xDE, 0xEF, 0x83, 0x64, 0x87, 0xC5, 0x4A}}

	GUID_IO_DISK_HEALTH_NOTIFICATION = syscall.GUID{0x0F1BD644, 0x3916, 0x49C5,
		[8]byte{0xB0, 0x63, 0x99, 0x19, 0x40, 0x11, 0x8F, 0xB2}}

	GUID_DEVINTERFACE_DMR = syscall.GUID{0xD0875FB4, 0x2196, 0x4C7A,
		[8]byte{0xA6, 0x3D, 0xE4, 0x16, 0xAD, 0xDD, 0x60, 0xA1}}

	GUID_DEVINTERFACE_DMP = syscall.GUID{0x25B4E268, 0x2A05, 0x496E,
		[8]byte{0x80, 0x3B, 0x26, 0x68, 0x37, 0xFB, 0xDA, 0x4B}}

	GUID_DEVINTERFACE_DMS = syscall.GUID{0xC96037AE, 0xA558, 0x4470,
		[8]byte{0xB4, 0x32, 0x11, 0x5A, 0x31, 0xB8, 0x55, 0x53}}
)

// enums

// enum
type ALERT_SYSTEM_SEV uint32

const (
	ALERT_SYSTEM_INFORMATIONAL ALERT_SYSTEM_SEV = 1
	ALERT_SYSTEM_WARNING       ALERT_SYSTEM_SEV = 2
	ALERT_SYSTEM_ERROR         ALERT_SYSTEM_SEV = 3
	ALERT_SYSTEM_QUERY         ALERT_SYSTEM_SEV = 4
	ALERT_SYSTEM_CRITICAL      ALERT_SYSTEM_SEV = 5
)

// enum
type APPCOMMAND_ID uint32

const (
	APPCOMMAND_BROWSER_BACKWARD                  APPCOMMAND_ID = 1
	APPCOMMAND_BROWSER_FORWARD                   APPCOMMAND_ID = 2
	APPCOMMAND_BROWSER_REFRESH                   APPCOMMAND_ID = 3
	APPCOMMAND_BROWSER_STOP                      APPCOMMAND_ID = 4
	APPCOMMAND_BROWSER_SEARCH                    APPCOMMAND_ID = 5
	APPCOMMAND_BROWSER_FAVORITES                 APPCOMMAND_ID = 6
	APPCOMMAND_BROWSER_HOME                      APPCOMMAND_ID = 7
	APPCOMMAND_VOLUME_MUTE                       APPCOMMAND_ID = 8
	APPCOMMAND_VOLUME_DOWN                       APPCOMMAND_ID = 9
	APPCOMMAND_VOLUME_UP                         APPCOMMAND_ID = 10
	APPCOMMAND_MEDIA_NEXTTRACK                   APPCOMMAND_ID = 11
	APPCOMMAND_MEDIA_PREVIOUSTRACK               APPCOMMAND_ID = 12
	APPCOMMAND_MEDIA_STOP                        APPCOMMAND_ID = 13
	APPCOMMAND_MEDIA_PLAY_PAUSE                  APPCOMMAND_ID = 14
	APPCOMMAND_LAUNCH_MAIL                       APPCOMMAND_ID = 15
	APPCOMMAND_LAUNCH_MEDIA_SELECT               APPCOMMAND_ID = 16
	APPCOMMAND_LAUNCH_APP1                       APPCOMMAND_ID = 17
	APPCOMMAND_LAUNCH_APP2                       APPCOMMAND_ID = 18
	APPCOMMAND_BASS_DOWN                         APPCOMMAND_ID = 19
	APPCOMMAND_BASS_BOOST                        APPCOMMAND_ID = 20
	APPCOMMAND_BASS_UP                           APPCOMMAND_ID = 21
	APPCOMMAND_TREBLE_DOWN                       APPCOMMAND_ID = 22
	APPCOMMAND_TREBLE_UP                         APPCOMMAND_ID = 23
	APPCOMMAND_MICROPHONE_VOLUME_MUTE            APPCOMMAND_ID = 24
	APPCOMMAND_MICROPHONE_VOLUME_DOWN            APPCOMMAND_ID = 25
	APPCOMMAND_MICROPHONE_VOLUME_UP              APPCOMMAND_ID = 26
	APPCOMMAND_HELP                              APPCOMMAND_ID = 27
	APPCOMMAND_FIND                              APPCOMMAND_ID = 28
	APPCOMMAND_NEW                               APPCOMMAND_ID = 29
	APPCOMMAND_OPEN                              APPCOMMAND_ID = 30
	APPCOMMAND_CLOSE                             APPCOMMAND_ID = 31
	APPCOMMAND_SAVE                              APPCOMMAND_ID = 32
	APPCOMMAND_PRINT                             APPCOMMAND_ID = 33
	APPCOMMAND_UNDO                              APPCOMMAND_ID = 34
	APPCOMMAND_REDO                              APPCOMMAND_ID = 35
	APPCOMMAND_COPY                              APPCOMMAND_ID = 36
	APPCOMMAND_CUT                               APPCOMMAND_ID = 37
	APPCOMMAND_PASTE                             APPCOMMAND_ID = 38
	APPCOMMAND_REPLY_TO_MAIL                     APPCOMMAND_ID = 39
	APPCOMMAND_FORWARD_MAIL                      APPCOMMAND_ID = 40
	APPCOMMAND_SEND_MAIL                         APPCOMMAND_ID = 41
	APPCOMMAND_SPELL_CHECK                       APPCOMMAND_ID = 42
	APPCOMMAND_DICTATE_OR_COMMAND_CONTROL_TOGGLE APPCOMMAND_ID = 43
	APPCOMMAND_MIC_ON_OFF_TOGGLE                 APPCOMMAND_ID = 44
	APPCOMMAND_CORRECTION_LIST                   APPCOMMAND_ID = 45
	APPCOMMAND_MEDIA_PLAY                        APPCOMMAND_ID = 46
	APPCOMMAND_MEDIA_PAUSE                       APPCOMMAND_ID = 47
	APPCOMMAND_MEDIA_RECORD                      APPCOMMAND_ID = 48
	APPCOMMAND_MEDIA_FAST_FORWARD                APPCOMMAND_ID = 49
	APPCOMMAND_MEDIA_REWIND                      APPCOMMAND_ID = 50
	APPCOMMAND_MEDIA_CHANNEL_UP                  APPCOMMAND_ID = 51
	APPCOMMAND_MEDIA_CHANNEL_DOWN                APPCOMMAND_ID = 52
	APPCOMMAND_DELETE                            APPCOMMAND_ID = 53
	APPCOMMAND_DWM_FLIP3D                        APPCOMMAND_ID = 54
)

// enum
// flags
type ATF_FLAGS uint32

const (
	ATF_TIMEOUTON     ATF_FLAGS = 1
	ATF_ONOFFFEEDBACK ATF_FLAGS = 2
)

// enum
// flags
type CHOOSECOLOR_FLAGS uint32

const (
	CC_RGBINIT              CHOOSECOLOR_FLAGS = 1
	CC_FULLOPEN             CHOOSECOLOR_FLAGS = 2
	CC_PREVENTFULLOPEN      CHOOSECOLOR_FLAGS = 4
	CC_SHOWHELP             CHOOSECOLOR_FLAGS = 8
	CC_ENABLEHOOK           CHOOSECOLOR_FLAGS = 16
	CC_ENABLETEMPLATE       CHOOSECOLOR_FLAGS = 32
	CC_ENABLETEMPLATEHANDLE CHOOSECOLOR_FLAGS = 64
	CC_SOLIDCOLOR           CHOOSECOLOR_FLAGS = 128
	CC_ANYCOLOR             CHOOSECOLOR_FLAGS = 256
)

// enum
type CLIPBOARD_FORMATS uint32

const (
	CF_TEXT            CLIPBOARD_FORMATS = 1
	CF_BITMAP          CLIPBOARD_FORMATS = 2
	CF_METAFILEPICT    CLIPBOARD_FORMATS = 3
	CF_SYLK            CLIPBOARD_FORMATS = 4
	CF_DIF             CLIPBOARD_FORMATS = 5
	CF_TIFF            CLIPBOARD_FORMATS = 6
	CF_OEMTEXT         CLIPBOARD_FORMATS = 7
	CF_DIB             CLIPBOARD_FORMATS = 8
	CF_PALETTE         CLIPBOARD_FORMATS = 9
	CF_PENDATA         CLIPBOARD_FORMATS = 10
	CF_RIFF            CLIPBOARD_FORMATS = 11
	CF_WAVE            CLIPBOARD_FORMATS = 12
	CF_UNICODETEXT     CLIPBOARD_FORMATS = 13
	CF_ENHMETAFILE     CLIPBOARD_FORMATS = 14
	CF_HDROP           CLIPBOARD_FORMATS = 15
	CF_LOCALE          CLIPBOARD_FORMATS = 16
	CF_DIBV5           CLIPBOARD_FORMATS = 17
	CF_MAX             CLIPBOARD_FORMATS = 18
	CF_OWNERDISPLAY    CLIPBOARD_FORMATS = 128
	CF_DSPTEXT         CLIPBOARD_FORMATS = 129
	CF_DSPBITMAP       CLIPBOARD_FORMATS = 130
	CF_DSPMETAFILEPICT CLIPBOARD_FORMATS = 131
	CF_DSPENHMETAFILE  CLIPBOARD_FORMATS = 142
	CF_PRIVATEFIRST    CLIPBOARD_FORMATS = 512
	CF_PRIVATELAST     CLIPBOARD_FORMATS = 767
	CF_GDIOBJFIRST     CLIPBOARD_FORMATS = 768
	CF_GDIOBJLAST      CLIPBOARD_FORMATS = 1023
)

// enum
// flags
type GESTURECONFIG_FLAGS uint32

const (
	GC_ALLGESTURES                         GESTURECONFIG_FLAGS = 1
	GC_ZOOM                                GESTURECONFIG_FLAGS = 1
	GC_PAN                                 GESTURECONFIG_FLAGS = 1
	GC_PAN_WITH_SINGLE_FINGER_VERTICALLY   GESTURECONFIG_FLAGS = 2
	GC_PAN_WITH_SINGLE_FINGER_HORIZONTALLY GESTURECONFIG_FLAGS = 4
	GC_PAN_WITH_GUTTER                     GESTURECONFIG_FLAGS = 8
	GC_PAN_WITH_INERTIA                    GESTURECONFIG_FLAGS = 16
	GC_ROTATE                              GESTURECONFIG_FLAGS = 1
	GC_TWOFINGERTAP                        GESTURECONFIG_FLAGS = 1
	GC_PRESSANDTAP                         GESTURECONFIG_FLAGS = 1
	GC_ROLLOVER                            GESTURECONFIG_FLAGS = 1
)

// enum
// flags
type CFE_UNDERLINE uint32

const (
	CFU_CF1UNDERLINE             CFE_UNDERLINE = 255
	CFU_INVERT                   CFE_UNDERLINE = 254
	CFU_UNDERLINETHICKLONGDASH   CFE_UNDERLINE = 18
	CFU_UNDERLINETHICKDOTTED     CFE_UNDERLINE = 17
	CFU_UNDERLINETHICKDASHDOTDOT CFE_UNDERLINE = 16
	CFU_UNDERLINETHICKDASHDOT    CFE_UNDERLINE = 15
	CFU_UNDERLINETHICKDASH       CFE_UNDERLINE = 14
	CFU_UNDERLINELONGDASH        CFE_UNDERLINE = 13
	CFU_UNDERLINEHEAVYWAVE       CFE_UNDERLINE = 12
	CFU_UNDERLINEDOUBLEWAVE      CFE_UNDERLINE = 11
	CFU_UNDERLINEHAIRLINE        CFE_UNDERLINE = 10
	CFU_UNDERLINETHICK           CFE_UNDERLINE = 9
	CFU_UNDERLINEWAVE            CFE_UNDERLINE = 8
	CFU_UNDERLINEDASHDOTDOT      CFE_UNDERLINE = 7
	CFU_UNDERLINEDASHDOT         CFE_UNDERLINE = 6
	CFU_UNDERLINEDASH            CFE_UNDERLINE = 5
	CFU_UNDERLINEDOTTED          CFE_UNDERLINE = 4
	CFU_UNDERLINEDOUBLE          CFE_UNDERLINE = 3
	CFU_UNDERLINEWORD            CFE_UNDERLINE = 2
	CFU_UNDERLINE                CFE_UNDERLINE = 1
	CFU_UNDERLINENONE            CFE_UNDERLINE = 0
)

// enum
type IGP_ID uint32

const (
	IGP_GETIMEVERSION IGP_ID = 4294967292
	IGP_PROPERTY      IGP_ID = 4
	IGP_CONVERSION    IGP_ID = 8
	IGP_SENTENCE      IGP_ID = 12
	IGP_UI            IGP_ID = 16
	IGP_SETCOMPSTR    IGP_ID = 20
	IGP_SELECT        IGP_ID = 24
)

// enum
// flags
type SECTION_FLAGS uint32

const (
	SECTION_ALL_ACCESS           SECTION_FLAGS = 983071
	SECTION_QUERY                SECTION_FLAGS = 1
	SECTION_MAP_WRITE            SECTION_FLAGS = 2
	SECTION_MAP_READ             SECTION_FLAGS = 4
	SECTION_MAP_EXECUTE          SECTION_FLAGS = 8
	SECTION_EXTEND_SIZE          SECTION_FLAGS = 16
	SECTION_MAP_EXECUTE_EXPLICIT SECTION_FLAGS = 32
)

// enum
// flags
type TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH uint32

const (
	TAPE_DRIVE_ABS_BLK_IMMED    TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147491840
	TAPE_DRIVE_ABSOLUTE_BLK     TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147487744
	TAPE_DRIVE_END_OF_DATA      TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147549184
	TAPE_DRIVE_FILEMARKS        TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147745792
	TAPE_DRIVE_LOAD_UNLOAD      TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147483649
	TAPE_DRIVE_LOAD_UNLD_IMMED  TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147483680
	TAPE_DRIVE_LOCK_UNLOCK      TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147483652
	TAPE_DRIVE_LOCK_UNLK_IMMED  TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147483776
	TAPE_DRIVE_LOG_BLK_IMMED    TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147516416
	TAPE_DRIVE_LOGICAL_BLK      TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147500032
	TAPE_DRIVE_RELATIVE_BLKS    TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147614720
	TAPE_DRIVE_REVERSE_POSITION TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2151677952
	TAPE_DRIVE_REWIND_IMMEDIATE TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147483656
	TAPE_DRIVE_SEQUENTIAL_FMKS  TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2148007936
	TAPE_DRIVE_SEQUENTIAL_SMKS  TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2149580800
	TAPE_DRIVE_SET_BLOCK_SIZE   TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147483664
	TAPE_DRIVE_SET_COMPRESSION  TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147484160
	TAPE_DRIVE_SET_ECC          TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147483904
	TAPE_DRIVE_SET_PADDING      TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147484672
	TAPE_DRIVE_SET_REPORT_SMKS  TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147485696
	TAPE_DRIVE_SETMARKS         TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2148532224
	TAPE_DRIVE_SPACE_IMMEDIATE  TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2155872256
	TAPE_DRIVE_TENSION          TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147483650
	TAPE_DRIVE_TENSION_IMMED    TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2147483712
	TAPE_DRIVE_WRITE_FILEMARKS  TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2181038080
	TAPE_DRIVE_WRITE_LONG_FMKS  TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2281701376
	TAPE_DRIVE_WRITE_MARK_IMMED TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2415919104
	TAPE_DRIVE_WRITE_SETMARKS   TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2164260864
	TAPE_DRIVE_WRITE_SHORT_FMKS TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH = 2214592512
)

// enum
type DEV_BROADCAST_HDR_DEVICE_TYPE uint32

const (
	DBT_DEVTYP_DEVICEINTERFACE DEV_BROADCAST_HDR_DEVICE_TYPE = 5
	DBT_DEVTYP_HANDLE          DEV_BROADCAST_HDR_DEVICE_TYPE = 6
	DBT_DEVTYP_OEM             DEV_BROADCAST_HDR_DEVICE_TYPE = 0
	DBT_DEVTYP_PORT            DEV_BROADCAST_HDR_DEVICE_TYPE = 3
	DBT_DEVTYP_VOLUME          DEV_BROADCAST_HDR_DEVICE_TYPE = 2
)

// enum
type DEV_BROADCAST_VOLUME_FLAGS uint16

const (
	DBTF_MEDIA DEV_BROADCAST_VOLUME_FLAGS = 1
	DBTF_NET   DEV_BROADCAST_VOLUME_FLAGS = 2
)

// enum
type ACCESS_REASON_TYPE int32

const (
	AccessReasonNone                     ACCESS_REASON_TYPE = 0
	AccessReasonAllowedAce               ACCESS_REASON_TYPE = 65536
	AccessReasonDeniedAce                ACCESS_REASON_TYPE = 131072
	AccessReasonAllowedParentAce         ACCESS_REASON_TYPE = 196608
	AccessReasonDeniedParentAce          ACCESS_REASON_TYPE = 262144
	AccessReasonNotGrantedByCape         ACCESS_REASON_TYPE = 327680
	AccessReasonNotGrantedByParentCape   ACCESS_REASON_TYPE = 393216
	AccessReasonNotGrantedToAppContainer ACCESS_REASON_TYPE = 458752
	AccessReasonMissingPrivilege         ACCESS_REASON_TYPE = 1048576
	AccessReasonFromPrivilege            ACCESS_REASON_TYPE = 2097152
	AccessReasonIntegrityLevel           ACCESS_REASON_TYPE = 3145728
	AccessReasonOwnership                ACCESS_REASON_TYPE = 4194304
	AccessReasonNullDacl                 ACCESS_REASON_TYPE = 5242880
	AccessReasonEmptyDacl                ACCESS_REASON_TYPE = 6291456
	AccessReasonNoSD                     ACCESS_REASON_TYPE = 7340032
	AccessReasonNoGrant                  ACCESS_REASON_TYPE = 8388608
	AccessReasonTrustLabel               ACCESS_REASON_TYPE = 9437184
	AccessReasonFilterAce                ACCESS_REASON_TYPE = 10485760
)

// enum
type SE_IMAGE_SIGNATURE_TYPE int32

const (
	SeImageSignatureNone             SE_IMAGE_SIGNATURE_TYPE = 0
	SeImageSignatureEmbedded         SE_IMAGE_SIGNATURE_TYPE = 1
	SeImageSignatureCache            SE_IMAGE_SIGNATURE_TYPE = 2
	SeImageSignatureCatalogCached    SE_IMAGE_SIGNATURE_TYPE = 3
	SeImageSignatureCatalogNotCached SE_IMAGE_SIGNATURE_TYPE = 4
	SeImageSignatureCatalogHint      SE_IMAGE_SIGNATURE_TYPE = 5
	SeImageSignaturePackageCatalog   SE_IMAGE_SIGNATURE_TYPE = 6
	SeImageSignaturePplMitigated     SE_IMAGE_SIGNATURE_TYPE = 7
)

// enum
type SE_LEARNING_MODE_DATA_TYPE int32

const (
	SeLearningModeInvalidType SE_LEARNING_MODE_DATA_TYPE = 0
	SeLearningModeSettings    SE_LEARNING_MODE_DATA_TYPE = 1
	SeLearningModeMax         SE_LEARNING_MODE_DATA_TYPE = 2
)

// enum
type SERVERSILO_STATE int32

const (
	SERVERSILO_INITING       SERVERSILO_STATE = 0
	SERVERSILO_STARTED       SERVERSILO_STATE = 1
	SERVERSILO_SHUTTING_DOWN SERVERSILO_STATE = 2
	SERVERSILO_TERMINATING   SERVERSILO_STATE = 3
	SERVERSILO_TERMINATED    SERVERSILO_STATE = 4
)

// enum
type MEM_DEDICATED_ATTRIBUTE_TYPE int32

const (
	MemDedicatedAttributeReadBandwidth  MEM_DEDICATED_ATTRIBUTE_TYPE = 0
	MemDedicatedAttributeReadLatency    MEM_DEDICATED_ATTRIBUTE_TYPE = 1
	MemDedicatedAttributeWriteBandwidth MEM_DEDICATED_ATTRIBUTE_TYPE = 2
	MemDedicatedAttributeWriteLatency   MEM_DEDICATED_ATTRIBUTE_TYPE = 3
	MemDedicatedAttributeMax            MEM_DEDICATED_ATTRIBUTE_TYPE = 4
)

// enum
type MEM_SECTION_EXTENDED_PARAMETER_TYPE int32

const (
	MemSectionExtendedParameterInvalidType       MEM_SECTION_EXTENDED_PARAMETER_TYPE = 0
	MemSectionExtendedParameterUserPhysicalFlags MEM_SECTION_EXTENDED_PARAMETER_TYPE = 1
	MemSectionExtendedParameterNumaNode          MEM_SECTION_EXTENDED_PARAMETER_TYPE = 2
	MemSectionExtendedParameterSigningLevel      MEM_SECTION_EXTENDED_PARAMETER_TYPE = 3
	MemSectionExtendedParameterMax               MEM_SECTION_EXTENDED_PARAMETER_TYPE = 4
)

// enum
type SharedVirtualDiskSupportType int32

const (
	SharedVirtualDisksUnsupported          SharedVirtualDiskSupportType = 0
	SharedVirtualDisksSupported            SharedVirtualDiskSupportType = 1
	SharedVirtualDiskSnapshotsSupported    SharedVirtualDiskSupportType = 3
	SharedVirtualDiskCDPSnapshotsSupported SharedVirtualDiskSupportType = 7
)

// enum
type SharedVirtualDiskHandleState int32

const (
	SharedVirtualDiskHandleStateNone         SharedVirtualDiskHandleState = 0
	SharedVirtualDiskHandleStateFileShared   SharedVirtualDiskHandleState = 1
	SharedVirtualDiskHandleStateHandleShared SharedVirtualDiskHandleState = 3
)

// enum
type MONITOR_DISPLAY_STATE int32

const (
	PowerMonitorOff MONITOR_DISPLAY_STATE = 0
	PowerMonitorOn  MONITOR_DISPLAY_STATE = 1
	PowerMonitorDim MONITOR_DISPLAY_STATE = 2
)

// enum
type USER_ACTIVITY_PRESENCE int32

const (
	PowerUserPresent    USER_ACTIVITY_PRESENCE = 0
	PowerUserNotPresent USER_ACTIVITY_PRESENCE = 1
	PowerUserInactive   USER_ACTIVITY_PRESENCE = 2
	PowerUserMaximum    USER_ACTIVITY_PRESENCE = 3
	PowerUserInvalid    USER_ACTIVITY_PRESENCE = 3
)

// enum
type POWER_USER_PRESENCE_TYPE int32

const (
	UserNotPresent POWER_USER_PRESENCE_TYPE = 0
	UserPresent    POWER_USER_PRESENCE_TYPE = 1
	UserUnknown    POWER_USER_PRESENCE_TYPE = 255
)

// enum
type POWER_MONITOR_REQUEST_REASON int32

const (
	MonitorRequestReasonUnknown                        POWER_MONITOR_REQUEST_REASON = 0
	MonitorRequestReasonPowerButton                    POWER_MONITOR_REQUEST_REASON = 1
	MonitorRequestReasonRemoteConnection               POWER_MONITOR_REQUEST_REASON = 2
	MonitorRequestReasonScMonitorpower                 POWER_MONITOR_REQUEST_REASON = 3
	MonitorRequestReasonUserInput                      POWER_MONITOR_REQUEST_REASON = 4
	MonitorRequestReasonAcDcDisplayBurst               POWER_MONITOR_REQUEST_REASON = 5
	MonitorRequestReasonUserDisplayBurst               POWER_MONITOR_REQUEST_REASON = 6
	MonitorRequestReasonPoSetSystemState               POWER_MONITOR_REQUEST_REASON = 7
	MonitorRequestReasonSetThreadExecutionState        POWER_MONITOR_REQUEST_REASON = 8
	MonitorRequestReasonFullWake                       POWER_MONITOR_REQUEST_REASON = 9
	MonitorRequestReasonSessionUnlock                  POWER_MONITOR_REQUEST_REASON = 10
	MonitorRequestReasonScreenOffRequest               POWER_MONITOR_REQUEST_REASON = 11
	MonitorRequestReasonIdleTimeout                    POWER_MONITOR_REQUEST_REASON = 12
	MonitorRequestReasonPolicyChange                   POWER_MONITOR_REQUEST_REASON = 13
	MonitorRequestReasonSleepButton                    POWER_MONITOR_REQUEST_REASON = 14
	MonitorRequestReasonLid                            POWER_MONITOR_REQUEST_REASON = 15
	MonitorRequestReasonBatteryCountChange             POWER_MONITOR_REQUEST_REASON = 16
	MonitorRequestReasonGracePeriod                    POWER_MONITOR_REQUEST_REASON = 17
	MonitorRequestReasonPnP                            POWER_MONITOR_REQUEST_REASON = 18
	MonitorRequestReasonDP                             POWER_MONITOR_REQUEST_REASON = 19
	MonitorRequestReasonSxTransition                   POWER_MONITOR_REQUEST_REASON = 20
	MonitorRequestReasonSystemIdle                     POWER_MONITOR_REQUEST_REASON = 21
	MonitorRequestReasonNearProximity                  POWER_MONITOR_REQUEST_REASON = 22
	MonitorRequestReasonThermalStandby                 POWER_MONITOR_REQUEST_REASON = 23
	MonitorRequestReasonResumePdc                      POWER_MONITOR_REQUEST_REASON = 24
	MonitorRequestReasonResumeS4                       POWER_MONITOR_REQUEST_REASON = 25
	MonitorRequestReasonTerminal                       POWER_MONITOR_REQUEST_REASON = 26
	MonitorRequestReasonPdcSignal                      POWER_MONITOR_REQUEST_REASON = 27
	MonitorRequestReasonAcDcDisplayBurstSuppressed     POWER_MONITOR_REQUEST_REASON = 28
	MonitorRequestReasonSystemStateEntered             POWER_MONITOR_REQUEST_REASON = 29
	MonitorRequestReasonWinrt                          POWER_MONITOR_REQUEST_REASON = 30
	MonitorRequestReasonUserInputKeyboard              POWER_MONITOR_REQUEST_REASON = 31
	MonitorRequestReasonUserInputMouse                 POWER_MONITOR_REQUEST_REASON = 32
	MonitorRequestReasonUserInputTouchpad              POWER_MONITOR_REQUEST_REASON = 33
	MonitorRequestReasonUserInputPen                   POWER_MONITOR_REQUEST_REASON = 34
	MonitorRequestReasonUserInputAccelerometer         POWER_MONITOR_REQUEST_REASON = 35
	MonitorRequestReasonUserInputHid                   POWER_MONITOR_REQUEST_REASON = 36
	MonitorRequestReasonUserInputPoUserPresent         POWER_MONITOR_REQUEST_REASON = 37
	MonitorRequestReasonUserInputSessionSwitch         POWER_MONITOR_REQUEST_REASON = 38
	MonitorRequestReasonUserInputInitialization        POWER_MONITOR_REQUEST_REASON = 39
	MonitorRequestReasonPdcSignalWindowsMobilePwrNotif POWER_MONITOR_REQUEST_REASON = 40
	MonitorRequestReasonPdcSignalWindowsMobileShell    POWER_MONITOR_REQUEST_REASON = 41
	MonitorRequestReasonPdcSignalHeyCortana            POWER_MONITOR_REQUEST_REASON = 42
	MonitorRequestReasonPdcSignalHolographicShell      POWER_MONITOR_REQUEST_REASON = 43
	MonitorRequestReasonPdcSignalFingerprint           POWER_MONITOR_REQUEST_REASON = 44
	MonitorRequestReasonDirectedDrips                  POWER_MONITOR_REQUEST_REASON = 45
	MonitorRequestReasonDim                            POWER_MONITOR_REQUEST_REASON = 46
	MonitorRequestReasonBuiltinPanel                   POWER_MONITOR_REQUEST_REASON = 47
	MonitorRequestReasonDisplayRequiredUnDim           POWER_MONITOR_REQUEST_REASON = 48
	MonitorRequestReasonBatteryCountChangeSuppressed   POWER_MONITOR_REQUEST_REASON = 49
	MonitorRequestReasonResumeModernStandby            POWER_MONITOR_REQUEST_REASON = 50
	MonitorRequestReasonTerminalInit                   POWER_MONITOR_REQUEST_REASON = 51
	MonitorRequestReasonPdcSignalSensorsHumanPresence  POWER_MONITOR_REQUEST_REASON = 52
	MonitorRequestReasonBatteryPreCritical             POWER_MONITOR_REQUEST_REASON = 53
	MonitorRequestReasonUserInputTouch                 POWER_MONITOR_REQUEST_REASON = 54
	MonitorRequestReasonMax                            POWER_MONITOR_REQUEST_REASON = 55
)

// enum
type POWER_MONITOR_REQUEST_TYPE int32

const (
	MonitorRequestTypeOff          POWER_MONITOR_REQUEST_TYPE = 0
	MonitorRequestTypeOnAndPresent POWER_MONITOR_REQUEST_TYPE = 1
	MonitorRequestTypeToggleOn     POWER_MONITOR_REQUEST_TYPE = 2
)

// enum
type POWER_SETTING_ALTITUDE int32

const (
	ALTITUDE_GROUP_POLICY      POWER_SETTING_ALTITUDE = 0
	ALTITUDE_USER              POWER_SETTING_ALTITUDE = 1
	ALTITUDE_RUNTIME_OVERRIDE  POWER_SETTING_ALTITUDE = 2
	ALTITUDE_PROVISIONING      POWER_SETTING_ALTITUDE = 3
	ALTITUDE_OEM_CUSTOMIZATION POWER_SETTING_ALTITUDE = 4
	ALTITUDE_INTERNAL_OVERRIDE POWER_SETTING_ALTITUDE = 5
	ALTITUDE_OS_DEFAULT        POWER_SETTING_ALTITUDE = 6
)

// enum
type HIBERFILE_BUCKET_SIZE int32

const (
	HiberFileBucket1GB       HIBERFILE_BUCKET_SIZE = 0
	HiberFileBucket2GB       HIBERFILE_BUCKET_SIZE = 1
	HiberFileBucket4GB       HIBERFILE_BUCKET_SIZE = 2
	HiberFileBucket8GB       HIBERFILE_BUCKET_SIZE = 3
	HiberFileBucket16GB      HIBERFILE_BUCKET_SIZE = 4
	HiberFileBucket32GB      HIBERFILE_BUCKET_SIZE = 5
	HiberFileBucketUnlimited HIBERFILE_BUCKET_SIZE = 6
	HiberFileBucketMax       HIBERFILE_BUCKET_SIZE = 7
)

// enum
type IMAGE_AUX_SYMBOL_TYPE int32

const (
	IMAGE_AUX_SYMBOL_TYPE_TOKEN_DEF IMAGE_AUX_SYMBOL_TYPE = 1
)

// enum
type ARM64_FNPDATA_FLAGS int32

const (
	PdataRefToFullXdata       ARM64_FNPDATA_FLAGS = 0
	PdataPackedUnwindFunction ARM64_FNPDATA_FLAGS = 1
	PdataPackedUnwindFragment ARM64_FNPDATA_FLAGS = 2
)

// enum
type ARM64_FNPDATA_CR int32

const (
	PdataCrUnchained        ARM64_FNPDATA_CR = 0
	PdataCrUnchainedSavedLr ARM64_FNPDATA_CR = 1
	PdataCrChainedWithPac   ARM64_FNPDATA_CR = 2
	PdataCrChained          ARM64_FNPDATA_CR = 3
)

// enum
type IMPORT_OBJECT_TYPE int32

const (
	IMPORT_OBJECT_CODE  IMPORT_OBJECT_TYPE = 0
	IMPORT_OBJECT_DATA  IMPORT_OBJECT_TYPE = 1
	IMPORT_OBJECT_CONST IMPORT_OBJECT_TYPE = 2
)

// enum
type IMPORT_OBJECT_NAME_TYPE int32

const (
	IMPORT_OBJECT_ORDINAL         IMPORT_OBJECT_NAME_TYPE = 0
	IMPORT_OBJECT_NAME            IMPORT_OBJECT_NAME_TYPE = 1
	IMPORT_OBJECT_NAME_NO_PREFIX  IMPORT_OBJECT_NAME_TYPE = 2
	IMPORT_OBJECT_NAME_UNDECORATE IMPORT_OBJECT_NAME_TYPE = 3
	IMPORT_OBJECT_NAME_EXPORTAS   IMPORT_OBJECT_NAME_TYPE = 4
)

// enum
type ReplacesCorHdrNumericDefines int32

const (
	COMIMAGE_FLAGS_ILONLY                      ReplacesCorHdrNumericDefines = 1
	COMIMAGE_FLAGS_32BITREQUIRED               ReplacesCorHdrNumericDefines = 2
	COMIMAGE_FLAGS_IL_LIBRARY                  ReplacesCorHdrNumericDefines = 4
	COMIMAGE_FLAGS_STRONGNAMESIGNED            ReplacesCorHdrNumericDefines = 8
	COMIMAGE_FLAGS_NATIVE_ENTRYPOINT           ReplacesCorHdrNumericDefines = 16
	COMIMAGE_FLAGS_TRACKDEBUGDATA              ReplacesCorHdrNumericDefines = 65536
	COMIMAGE_FLAGS_32BITPREFERRED              ReplacesCorHdrNumericDefines = 131072
	COR_VERSION_MAJOR_V2                       ReplacesCorHdrNumericDefines = 2
	COR_VERSION_MAJOR                          ReplacesCorHdrNumericDefines = 2
	COR_VERSION_MINOR                          ReplacesCorHdrNumericDefines = 5
	COR_DELETED_NAME_LENGTH                    ReplacesCorHdrNumericDefines = 8
	COR_VTABLEGAP_NAME_LENGTH                  ReplacesCorHdrNumericDefines = 8
	NATIVE_TYPE_MAX_CB                         ReplacesCorHdrNumericDefines = 1
	COR_ILMETHOD_SECT_SMALL_MAX_DATASIZE       ReplacesCorHdrNumericDefines = 255
	IMAGE_COR_MIH_METHODRVA                    ReplacesCorHdrNumericDefines = 1
	IMAGE_COR_MIH_EHRVA                        ReplacesCorHdrNumericDefines = 2
	IMAGE_COR_MIH_BASICBLOCK                   ReplacesCorHdrNumericDefines = 8
	COR_VTABLE_32BIT                           ReplacesCorHdrNumericDefines = 1
	COR_VTABLE_64BIT                           ReplacesCorHdrNumericDefines = 2
	COR_VTABLE_FROM_UNMANAGED                  ReplacesCorHdrNumericDefines = 4
	COR_VTABLE_FROM_UNMANAGED_RETAIN_APPDOMAIN ReplacesCorHdrNumericDefines = 8
	COR_VTABLE_CALL_MOST_DERIVED               ReplacesCorHdrNumericDefines = 16
	IMAGE_COR_EATJ_THUNK_SIZE                  ReplacesCorHdrNumericDefines = 32
	MAX_CLASS_NAME                             ReplacesCorHdrNumericDefines = 1024
	MAX_PACKAGE_NAME                           ReplacesCorHdrNumericDefines = 1024
)

// enum
type RTL_UMS_SCHEDULER_REASON int32

const (
	UmsSchedulerStartup       RTL_UMS_SCHEDULER_REASON = 0
	UmsSchedulerThreadBlocked RTL_UMS_SCHEDULER_REASON = 1
	UmsSchedulerThreadYield   RTL_UMS_SCHEDULER_REASON = 2
)

// enum
type IMAGE_POLICY_ENTRY_TYPE int32

const (
	ImagePolicyEntryTypeNone          IMAGE_POLICY_ENTRY_TYPE = 0
	ImagePolicyEntryTypeBool          IMAGE_POLICY_ENTRY_TYPE = 1
	ImagePolicyEntryTypeInt8          IMAGE_POLICY_ENTRY_TYPE = 2
	ImagePolicyEntryTypeUInt8         IMAGE_POLICY_ENTRY_TYPE = 3
	ImagePolicyEntryTypeInt16         IMAGE_POLICY_ENTRY_TYPE = 4
	ImagePolicyEntryTypeUInt16        IMAGE_POLICY_ENTRY_TYPE = 5
	ImagePolicyEntryTypeInt32         IMAGE_POLICY_ENTRY_TYPE = 6
	ImagePolicyEntryTypeUInt32        IMAGE_POLICY_ENTRY_TYPE = 7
	ImagePolicyEntryTypeInt64         IMAGE_POLICY_ENTRY_TYPE = 8
	ImagePolicyEntryTypeUInt64        IMAGE_POLICY_ENTRY_TYPE = 9
	ImagePolicyEntryTypeAnsiString    IMAGE_POLICY_ENTRY_TYPE = 10
	ImagePolicyEntryTypeUnicodeString IMAGE_POLICY_ENTRY_TYPE = 11
	ImagePolicyEntryTypeOverride      IMAGE_POLICY_ENTRY_TYPE = 12
	ImagePolicyEntryTypeMaximum       IMAGE_POLICY_ENTRY_TYPE = 13
)

// enum
type IMAGE_POLICY_ID int32

const (
	ImagePolicyIdNone             IMAGE_POLICY_ID = 0
	ImagePolicyIdEtw              IMAGE_POLICY_ID = 1
	ImagePolicyIdDebug            IMAGE_POLICY_ID = 2
	ImagePolicyIdCrashDump        IMAGE_POLICY_ID = 3
	ImagePolicyIdCrashDumpKey     IMAGE_POLICY_ID = 4
	ImagePolicyIdCrashDumpKeyGuid IMAGE_POLICY_ID = 5
	ImagePolicyIdParentSd         IMAGE_POLICY_ID = 6
	ImagePolicyIdParentSdRev      IMAGE_POLICY_ID = 7
	ImagePolicyIdSvn              IMAGE_POLICY_ID = 8
	ImagePolicyIdDeviceId         IMAGE_POLICY_ID = 9
	ImagePolicyIdCapability       IMAGE_POLICY_ID = 10
	ImagePolicyIdScenarioId       IMAGE_POLICY_ID = 11
	ImagePolicyIdMaximum          IMAGE_POLICY_ID = 12
)

// enum
type ACTIVATION_CONTEXT_INFO_CLASS int32

const (
	ActivationContextBasicInformation                      ACTIVATION_CONTEXT_INFO_CLASS = 1
	ActivationContextDetailedInformation                   ACTIVATION_CONTEXT_INFO_CLASS = 2
	AssemblyDetailedInformationInActivationContext         ACTIVATION_CONTEXT_INFO_CLASS = 3
	FileInformationInAssemblyOfAssemblyInActivationContext ACTIVATION_CONTEXT_INFO_CLASS = 4
	RunlevelInformationInActivationContext                 ACTIVATION_CONTEXT_INFO_CLASS = 5
	CompatibilityInformationInActivationContext            ACTIVATION_CONTEXT_INFO_CLASS = 6
	ActivationContextManifestResourceName                  ACTIVATION_CONTEXT_INFO_CLASS = 7
	MaxActivationContextInfoClass                          ACTIVATION_CONTEXT_INFO_CLASS = 8
	AssemblyDetailedInformationInActivationContxt          ACTIVATION_CONTEXT_INFO_CLASS = 3
	FileInformationInAssemblyOfAssemblyInActivationContxt  ACTIVATION_CONTEXT_INFO_CLASS = 4
)

// enum
type CM_SERVICE_NODE_TYPE int32

const (
	DriverType               CM_SERVICE_NODE_TYPE = 1
	FileSystemType           CM_SERVICE_NODE_TYPE = 2
	Win32ServiceOwnProcess   CM_SERVICE_NODE_TYPE = 16
	Win32ServiceShareProcess CM_SERVICE_NODE_TYPE = 32
	AdapterType              CM_SERVICE_NODE_TYPE = 4
	RecognizerType           CM_SERVICE_NODE_TYPE = 8
)

// enum
type CM_SERVICE_LOAD_TYPE int32

const (
	BootLoad    CM_SERVICE_LOAD_TYPE = 0
	SystemLoad  CM_SERVICE_LOAD_TYPE = 1
	AutoLoad    CM_SERVICE_LOAD_TYPE = 2
	DemandLoad  CM_SERVICE_LOAD_TYPE = 3
	DisableLoad CM_SERVICE_LOAD_TYPE = 4
)

// enum
type CM_ERROR_CONTROL_TYPE int32

const (
	IgnoreError   CM_ERROR_CONTROL_TYPE = 0
	NormalError   CM_ERROR_CONTROL_TYPE = 1
	SevereError   CM_ERROR_CONTROL_TYPE = 2
	CriticalError CM_ERROR_CONTROL_TYPE = 3
)

// enum
type TAPE_DRIVE_PROBLEM_TYPE int32

const (
	TapeDriveProblemNone         TAPE_DRIVE_PROBLEM_TYPE = 0
	TapeDriveReadWriteWarning    TAPE_DRIVE_PROBLEM_TYPE = 1
	TapeDriveReadWriteError      TAPE_DRIVE_PROBLEM_TYPE = 2
	TapeDriveReadWarning         TAPE_DRIVE_PROBLEM_TYPE = 3
	TapeDriveWriteWarning        TAPE_DRIVE_PROBLEM_TYPE = 4
	TapeDriveReadError           TAPE_DRIVE_PROBLEM_TYPE = 5
	TapeDriveWriteError          TAPE_DRIVE_PROBLEM_TYPE = 6
	TapeDriveHardwareError       TAPE_DRIVE_PROBLEM_TYPE = 7
	TapeDriveUnsupportedMedia    TAPE_DRIVE_PROBLEM_TYPE = 8
	TapeDriveScsiConnectionError TAPE_DRIVE_PROBLEM_TYPE = 9
	TapeDriveTimetoClean         TAPE_DRIVE_PROBLEM_TYPE = 10
	TapeDriveCleanDriveNow       TAPE_DRIVE_PROBLEM_TYPE = 11
	TapeDriveMediaLifeExpired    TAPE_DRIVE_PROBLEM_TYPE = 12
	TapeDriveSnappedTape         TAPE_DRIVE_PROBLEM_TYPE = 13
)

// enum
type TRANSACTION_STATE int32

const (
	TransactionStateNormal          TRANSACTION_STATE = 1
	TransactionStateIndoubt         TRANSACTION_STATE = 2
	TransactionStateCommittedNotify TRANSACTION_STATE = 3
)

// enum
type TRANSACTION_INFORMATION_CLASS int32

const (
	TransactionBasicInformation              TRANSACTION_INFORMATION_CLASS = 0
	TransactionPropertiesInformation         TRANSACTION_INFORMATION_CLASS = 1
	TransactionEnlistmentInformation         TRANSACTION_INFORMATION_CLASS = 2
	TransactionSuperiorEnlistmentInformation TRANSACTION_INFORMATION_CLASS = 3
	TransactionBindInformation               TRANSACTION_INFORMATION_CLASS = 4
	TransactionDTCPrivateInformation         TRANSACTION_INFORMATION_CLASS = 5
)

// enum
type TRANSACTIONMANAGER_INFORMATION_CLASS int32

const (
	TransactionManagerBasicInformation             TRANSACTIONMANAGER_INFORMATION_CLASS = 0
	TransactionManagerLogInformation               TRANSACTIONMANAGER_INFORMATION_CLASS = 1
	TransactionManagerLogPathInformation           TRANSACTIONMANAGER_INFORMATION_CLASS = 2
	TransactionManagerRecoveryInformation          TRANSACTIONMANAGER_INFORMATION_CLASS = 4
	TransactionManagerOnlineProbeInformation       TRANSACTIONMANAGER_INFORMATION_CLASS = 3
	TransactionManagerOldestTransactionInformation TRANSACTIONMANAGER_INFORMATION_CLASS = 5
)

// enum
type RESOURCEMANAGER_INFORMATION_CLASS int32

const (
	ResourceManagerBasicInformation      RESOURCEMANAGER_INFORMATION_CLASS = 0
	ResourceManagerCompletionInformation RESOURCEMANAGER_INFORMATION_CLASS = 1
)

// enum
type ENLISTMENT_INFORMATION_CLASS int32

const (
	EnlistmentBasicInformation    ENLISTMENT_INFORMATION_CLASS = 0
	EnlistmentRecoveryInformation ENLISTMENT_INFORMATION_CLASS = 1
	EnlistmentCrmInformation      ENLISTMENT_INFORMATION_CLASS = 2
)

// enum
type KTMOBJECT_TYPE int32

const (
	KTMOBJECT_TRANSACTION         KTMOBJECT_TYPE = 0
	KTMOBJECT_TRANSACTION_MANAGER KTMOBJECT_TYPE = 1
	KTMOBJECT_RESOURCE_MANAGER    KTMOBJECT_TYPE = 2
	KTMOBJECT_ENLISTMENT          KTMOBJECT_TYPE = 3
	KTMOBJECT_INVALID             KTMOBJECT_TYPE = 4
)

// structs

type TP_POOL struct {
}

type TP_CLEANUP_GROUP struct {
}

type TEB struct {
}

type RemHGLOBAL struct {
	FNullHGlobal int32
	CbData       uint32
	Data         [1]byte
}

type RemHMETAFILEPICT struct {
	Mm     int32
	XExt   int32
	YExt   int32
	CbData uint32
	Data   [1]byte
}

type RemHENHMETAFILE struct {
	CbData uint32
	Data   [1]byte
}

type RemHBITMAP struct {
	CbData uint32
	Data   [1]byte
}

type RemHPALETTE struct {
	CbData uint32
	Data   [1]byte
}

type RemBRUSH struct {
	CbData uint32
	Data   [1]byte
}

type UserCLIPFORMAT_U struct {
	Data [1]uint64
}

func (this *UserCLIPFORMAT_U) DwValue() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *UserCLIPFORMAT_U) DwValueVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *UserCLIPFORMAT_U) PwszName() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *UserCLIPFORMAT_U) PwszNameVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type UserCLIPFORMAT struct {
	FContext int32
	U        UserCLIPFORMAT_U
}

type GDI_NONREMOTE_U struct {
	Data [1]uint64
}

func (this *GDI_NONREMOTE_U) HInproc() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *GDI_NONREMOTE_U) HInprocVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *GDI_NONREMOTE_U) HRemote() **DWORD_BLOB {
	return (**DWORD_BLOB)(unsafe.Pointer(this))
}

func (this *GDI_NONREMOTE_U) HRemoteVal() *DWORD_BLOB {
	return *(**DWORD_BLOB)(unsafe.Pointer(this))
}

type GDI_NONREMOTE struct {
	FContext int32
	U        GDI_NONREMOTE_U
}

type UserHGLOBAL_U struct {
	Data [1]uint64
}

func (this *UserHGLOBAL_U) HInproc() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *UserHGLOBAL_U) HInprocVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *UserHGLOBAL_U) HRemote() **FLAGGED_BYTE_BLOB {
	return (**FLAGGED_BYTE_BLOB)(unsafe.Pointer(this))
}

func (this *UserHGLOBAL_U) HRemoteVal() *FLAGGED_BYTE_BLOB {
	return *(**FLAGGED_BYTE_BLOB)(unsafe.Pointer(this))
}

func (this *UserHGLOBAL_U) HInproc64() *int64 {
	return (*int64)(unsafe.Pointer(this))
}

func (this *UserHGLOBAL_U) HInproc64Val() int64 {
	return *(*int64)(unsafe.Pointer(this))
}

type UserHGLOBAL struct {
	FContext int32
	U        UserHGLOBAL_U
}

type UserHMETAFILE_U struct {
	Data [1]uint64
}

func (this *UserHMETAFILE_U) HInproc() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *UserHMETAFILE_U) HInprocVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *UserHMETAFILE_U) HRemote() **BYTE_BLOB {
	return (**BYTE_BLOB)(unsafe.Pointer(this))
}

func (this *UserHMETAFILE_U) HRemoteVal() *BYTE_BLOB {
	return *(**BYTE_BLOB)(unsafe.Pointer(this))
}

func (this *UserHMETAFILE_U) HInproc64() *int64 {
	return (*int64)(unsafe.Pointer(this))
}

func (this *UserHMETAFILE_U) HInproc64Val() int64 {
	return *(*int64)(unsafe.Pointer(this))
}

type UserHMETAFILE struct {
	FContext int32
	U        UserHMETAFILE_U
}

type RemoteMETAFILEPICT struct {
	Mm   int32
	XExt int32
	YExt int32
	HMF  *UserHMETAFILE
}

type UserHMETAFILEPICT_U struct {
	Data [1]uint64
}

func (this *UserHMETAFILEPICT_U) HInproc() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *UserHMETAFILEPICT_U) HInprocVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *UserHMETAFILEPICT_U) HRemote() **RemoteMETAFILEPICT {
	return (**RemoteMETAFILEPICT)(unsafe.Pointer(this))
}

func (this *UserHMETAFILEPICT_U) HRemoteVal() *RemoteMETAFILEPICT {
	return *(**RemoteMETAFILEPICT)(unsafe.Pointer(this))
}

func (this *UserHMETAFILEPICT_U) HInproc64() *int64 {
	return (*int64)(unsafe.Pointer(this))
}

func (this *UserHMETAFILEPICT_U) HInproc64Val() int64 {
	return *(*int64)(unsafe.Pointer(this))
}

type UserHMETAFILEPICT struct {
	FContext int32
	U        UserHMETAFILEPICT_U
}

type UserHENHMETAFILE_U struct {
	Data [1]uint64
}

func (this *UserHENHMETAFILE_U) HInproc() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *UserHENHMETAFILE_U) HInprocVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *UserHENHMETAFILE_U) HRemote() **BYTE_BLOB {
	return (**BYTE_BLOB)(unsafe.Pointer(this))
}

func (this *UserHENHMETAFILE_U) HRemoteVal() *BYTE_BLOB {
	return *(**BYTE_BLOB)(unsafe.Pointer(this))
}

func (this *UserHENHMETAFILE_U) HInproc64() *int64 {
	return (*int64)(unsafe.Pointer(this))
}

func (this *UserHENHMETAFILE_U) HInproc64Val() int64 {
	return *(*int64)(unsafe.Pointer(this))
}

type UserHENHMETAFILE struct {
	FContext int32
	U        UserHENHMETAFILE_U
}

type UserBITMAP struct {
	BmType       int32
	BmWidth      int32
	BmHeight     int32
	BmWidthBytes int32
	BmPlanes     uint16
	BmBitsPixel  uint16
	CbSize       uint32
	PBuffer      [1]byte
}

type UserHBITMAP_U struct {
	Data [1]uint64
}

func (this *UserHBITMAP_U) HInproc() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *UserHBITMAP_U) HInprocVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *UserHBITMAP_U) HRemote() **UserBITMAP {
	return (**UserBITMAP)(unsafe.Pointer(this))
}

func (this *UserHBITMAP_U) HRemoteVal() *UserBITMAP {
	return *(**UserBITMAP)(unsafe.Pointer(this))
}

func (this *UserHBITMAP_U) HInproc64() *int64 {
	return (*int64)(unsafe.Pointer(this))
}

func (this *UserHBITMAP_U) HInproc64Val() int64 {
	return *(*int64)(unsafe.Pointer(this))
}

type UserHBITMAP struct {
	FContext int32
	U        UserHBITMAP_U
}

type UserHPALETTE_U struct {
	Data [1]uint64
}

func (this *UserHPALETTE_U) HInproc() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *UserHPALETTE_U) HInprocVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *UserHPALETTE_U) HRemote() **LOGPALETTE {
	return (**LOGPALETTE)(unsafe.Pointer(this))
}

func (this *UserHPALETTE_U) HRemoteVal() *LOGPALETTE {
	return *(**LOGPALETTE)(unsafe.Pointer(this))
}

func (this *UserHPALETTE_U) HInproc64() *int64 {
	return (*int64)(unsafe.Pointer(this))
}

func (this *UserHPALETTE_U) HInproc64Val() int64 {
	return *(*int64)(unsafe.Pointer(this))
}

type UserHPALETTE struct {
	FContext int32
	U        UserHPALETTE_U
}

type RemotableHandle_U struct {
	Data [1]uint32
}

func (this *RemotableHandle_U) HInproc() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *RemotableHandle_U) HInprocVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *RemotableHandle_U) HRemote() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *RemotableHandle_U) HRemoteVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

type RemotableHandle struct {
	FContext int32
	U        RemotableHandle_U
}

type DEVICE_EVENT_MOUNT struct {
	Version              uint32
	Flags                uint32
	FileSystemNameLength uint32
	FileSystemNameOffset uint32
}

type DEVICE_EVENT_BECOMING_READY struct {
	Version               uint32
	Reason                uint32
	Estimated100msToReady uint32
}

type DEVICE_EVENT_EXTERNAL_REQUEST struct {
	Version      uint32
	DeviceClass  uint32
	ButtonStatus uint16
	Request      uint16
	SystemTime   int64
}

type DEVICE_EVENT_GENERIC_DATA struct {
	EventNumber uint32
}

type DEVICE_EVENT_RBC_DATA struct {
	EventNumber    uint32
	SenseQualifier byte
	SenseCode      byte
	SenseKey       byte
	Reserved       byte
	Information    uint32
}

type GUID_IO_DISK_CLONE_ARRIVAL_INFORMATION struct {
	DiskNumber uint32
}

type DISK_HEALTH_NOTIFICATION_DATA struct {
	DeviceGuid syscall.GUID
}

type REDBOOK_DIGITAL_AUDIO_EXTRACTION_INFO struct {
	Version       uint32
	Accurate      uint32
	Supported     uint32
	AccurateMask0 uint32
}

type DEV_BROADCAST_HDR struct {
	Dbch_size       uint32
	Dbch_devicetype DEV_BROADCAST_HDR_DEVICE_TYPE
	Dbch_reserved   uint32
}

type VolLockBroadcast struct {
	Vlb_dbh      DEV_BROADCAST_HDR
	Vlb_owner    uint32
	Vlb_perms    byte
	Vlb_lockType byte
	Vlb_drive    byte
	Vlb_flags    byte
}

type DEV_BROADCAST_HEADER_ struct {
	Dbcd_size       uint32
	Dbcd_devicetype uint32
	Dbcd_reserved   uint32
}

type DEV_BROADCAST_OEM struct {
	Dbco_size       uint32
	Dbco_devicetype uint32
	Dbco_reserved   uint32
	Dbco_identifier uint32
	Dbco_suppfunc   uint32
}

type DEV_BROADCAST_DEVNODE struct {
	Dbcd_size       uint32
	Dbcd_devicetype uint32
	Dbcd_reserved   uint32
	Dbcd_devnode    uint32
}

type DEV_BROADCAST_VOLUME struct {
	Dbcv_size       uint32
	Dbcv_devicetype uint32
	Dbcv_reserved   uint32
	Dbcv_unitmask   uint32
	Dbcv_flags      DEV_BROADCAST_VOLUME_FLAGS
}

type DEV_BROADCAST_PORT_A struct {
	Dbcp_size       uint32
	Dbcp_devicetype uint32
	Dbcp_reserved   uint32
	Dbcp_name       [1]CHAR
}

type DEV_BROADCAST_PORT_ = DEV_BROADCAST_PORT_W
type DEV_BROADCAST_PORT_W struct {
	Dbcp_size       uint32
	Dbcp_devicetype uint32
	Dbcp_reserved   uint32
	Dbcp_name       [1]uint16
}

type DEV_BROADCAST_NET struct {
	Dbcn_size       uint32
	Dbcn_devicetype uint32
	Dbcn_reserved   uint32
	Dbcn_resource   uint32
	Dbcn_flags      uint32
}

type DEV_BROADCAST_DEVICEINTERFACE_A struct {
	Dbcc_size       uint32
	Dbcc_devicetype uint32
	Dbcc_reserved   uint32
	Dbcc_classguid  syscall.GUID
	Dbcc_name       [1]CHAR
}

type DEV_BROADCAST_DEVICEINTERFACE_ = DEV_BROADCAST_DEVICEINTERFACE_W
type DEV_BROADCAST_DEVICEINTERFACE_W struct {
	Dbcc_size       uint32
	Dbcc_devicetype uint32
	Dbcc_reserved   uint32
	Dbcc_classguid  syscall.GUID
	Dbcc_name       [1]uint16
}

type DEV_BROADCAST_HANDLE struct {
	Dbch_size       uint32
	Dbch_devicetype uint32
	Dbch_reserved   uint32
	Dbch_handle     HANDLE
	Dbch_hdevnotify unsafe.Pointer
	Dbch_eventguid  syscall.GUID
	Dbch_nameoffset int32
	Dbch_data       [1]byte
}

type DEV_BROADCAST_HANDLE32 struct {
	Dbch_size       uint32
	Dbch_devicetype uint32
	Dbch_reserved   uint32
	Dbch_handle     uint32
	Dbch_hdevnotify uint32
	Dbch_eventguid  syscall.GUID
	Dbch_nameoffset int32
	Dbch_data       [1]byte
}

type DEV_BROADCAST_HANDLE64 struct {
	Dbch_size       uint32
	Dbch_devicetype uint32
	Dbch_reserved   uint32
	Dbch_handle     uint64
	Dbch_hdevnotify uint64
	Dbch_eventguid  syscall.GUID
	Dbch_nameoffset int32
	Dbch_data       [1]byte
}

type DEV_BROADCAST_USERDEFINED_ struct {
	Dbud_dbh    DEV_BROADCAST_HDR
	Dbud_szName [1]CHAR
}

type AtlThunkData_t struct {
}

type XSAVE_CET_U_FORMAT struct {
	Ia32CetUMsr   uint64
	Ia32Pl3SspMsr uint64
}

type KERNEL_CET_CONTEXT_Anonymous_Anonymous struct {
	Bitfield_ uint16
}

type KERNEL_CET_CONTEXT_Anonymous struct {
	KERNEL_CET_CONTEXT_Anonymous_Anonymous
}

func (this *KERNEL_CET_CONTEXT_Anonymous) AllFlags() *uint16 {
	return (*uint16)(unsafe.Pointer(this))
}

func (this *KERNEL_CET_CONTEXT_Anonymous) AllFlagsVal() uint16 {
	return *(*uint16)(unsafe.Pointer(this))
}

func (this *KERNEL_CET_CONTEXT_Anonymous) Anonymous() *KERNEL_CET_CONTEXT_Anonymous_Anonymous {
	return (*KERNEL_CET_CONTEXT_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *KERNEL_CET_CONTEXT_Anonymous) AnonymousVal() KERNEL_CET_CONTEXT_Anonymous_Anonymous {
	return *(*KERNEL_CET_CONTEXT_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type KERNEL_CET_CONTEXT struct {
	Ssp   uint64
	Rip   uint64
	SegCs uint16
	KERNEL_CET_CONTEXT_Anonymous
	Fill [2]uint16
}

type SCOPE_TABLE_AMD64_Anonymous struct {
	BeginAddress   uint32
	EndAddress     uint32
	HandlerAddress uint32
	JumpTarget     uint32
}

type SCOPE_TABLE_AMD64 struct {
	Count       uint32
	ScopeRecord [1]SCOPE_TABLE_AMD64_Anonymous
}

type SCOPE_TABLE_ARM_Anonymous struct {
	BeginAddress   uint32
	EndAddress     uint32
	HandlerAddress uint32
	JumpTarget     uint32
}

type SCOPE_TABLE_ARM struct {
	Count       uint32
	ScopeRecord [1]SCOPE_TABLE_ARM_Anonymous
}

type SCOPE_TABLE_ARM64_Anonymous struct {
	BeginAddress   uint32
	EndAddress     uint32
	HandlerAddress uint32
	JumpTarget     uint32
}

type SCOPE_TABLE_ARM64 struct {
	Count       uint32
	ScopeRecord [1]SCOPE_TABLE_ARM64_Anonymous
}

type DISPATCHER_CONTEXT_NONVOLREG_ARM64_Anonymous struct {
	GpNvRegs [11]uint64
	FpNvRegs [8]float64
}

type DISPATCHER_CONTEXT_NONVOLREG_ARM64 struct {
	DISPATCHER_CONTEXT_NONVOLREG_ARM64_Anonymous
}

func (this *DISPATCHER_CONTEXT_NONVOLREG_ARM64) Buffer() *[152]byte {
	return (*[152]byte)(unsafe.Pointer(this))
}

func (this *DISPATCHER_CONTEXT_NONVOLREG_ARM64) BufferVal() [152]byte {
	return *(*[152]byte)(unsafe.Pointer(this))
}

func (this *DISPATCHER_CONTEXT_NONVOLREG_ARM64) Anonymous() *DISPATCHER_CONTEXT_NONVOLREG_ARM64_Anonymous {
	return (*DISPATCHER_CONTEXT_NONVOLREG_ARM64_Anonymous)(unsafe.Pointer(this))
}

func (this *DISPATCHER_CONTEXT_NONVOLREG_ARM64) AnonymousVal() DISPATCHER_CONTEXT_NONVOLREG_ARM64_Anonymous {
	return *(*DISPATCHER_CONTEXT_NONVOLREG_ARM64_Anonymous)(unsafe.Pointer(this))
}

type SECURITY_DESCRIPTOR_RELATIVE struct {
	Revision byte
	Sbz1     byte
	Control  uint16
	Owner    uint32
	Group    uint32
	Sacl     uint32
	Dacl     uint32
}

type SECURITY_OBJECT_AI_PARAMS struct {
	Size           uint32
	ConstraintMask uint32
}

type SE_TOKEN_USER_Anonymous1 struct {
	Data [2]uint64
}

func (this *SE_TOKEN_USER_Anonymous1) TokenUser() *TOKEN_USER {
	return (*TOKEN_USER)(unsafe.Pointer(this))
}

func (this *SE_TOKEN_USER_Anonymous1) TokenUserVal() TOKEN_USER {
	return *(*TOKEN_USER)(unsafe.Pointer(this))
}

func (this *SE_TOKEN_USER_Anonymous1) User() *SID_AND_ATTRIBUTES {
	return (*SID_AND_ATTRIBUTES)(unsafe.Pointer(this))
}

func (this *SE_TOKEN_USER_Anonymous1) UserVal() SID_AND_ATTRIBUTES {
	return *(*SID_AND_ATTRIBUTES)(unsafe.Pointer(this))
}

type SE_TOKEN_USER_Anonymous2 struct {
	Data [17]uint32
}

func (this *SE_TOKEN_USER_Anonymous2) Sid() *SID {
	return (*SID)(unsafe.Pointer(this))
}

func (this *SE_TOKEN_USER_Anonymous2) SidVal() SID {
	return *(*SID)(unsafe.Pointer(this))
}

func (this *SE_TOKEN_USER_Anonymous2) Buffer() *[68]byte {
	return (*[68]byte)(unsafe.Pointer(this))
}

func (this *SE_TOKEN_USER_Anonymous2) BufferVal() [68]byte {
	return *(*[68]byte)(unsafe.Pointer(this))
}

type SE_TOKEN_USER struct {
	SE_TOKEN_USER_Anonymous1
	SE_TOKEN_USER_Anonymous2
}

type TOKEN_SID_INFORMATION struct {
	Sid PSID
}

type TOKEN_BNO_ISOLATION_INFORMATION struct {
	IsolationPrefix  PWSTR
	IsolationEnabled BOOLEAN
}

type NT_TIB32_Anonymous struct {
	Data [1]uint32
}

func (this *NT_TIB32_Anonymous) FiberData() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *NT_TIB32_Anonymous) FiberDataVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *NT_TIB32_Anonymous) Version() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *NT_TIB32_Anonymous) VersionVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type NT_TIB32 struct {
	ExceptionList uint32
	StackBase     uint32
	StackLimit    uint32
	SubSystemTib  uint32
	NT_TIB32_Anonymous
	ArbitraryUserPointer uint32
	Self                 uint32
}

type NT_TIB64_Anonymous struct {
	Data [1]uint64
}

func (this *NT_TIB64_Anonymous) FiberData() *uint64 {
	return (*uint64)(unsafe.Pointer(this))
}

func (this *NT_TIB64_Anonymous) FiberDataVal() uint64 {
	return *(*uint64)(unsafe.Pointer(this))
}

func (this *NT_TIB64_Anonymous) Version() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *NT_TIB64_Anonymous) VersionVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type NT_TIB64 struct {
	ExceptionList uint64
	StackBase     uint64
	StackLimit    uint64
	SubSystemTib  uint64
	NT_TIB64_Anonymous
	ArbitraryUserPointer uint64
	Self                 uint64
}

type UMS_CREATE_THREAD_ATTRIBUTES struct {
	UmsVersion        uint32
	UmsContext        unsafe.Pointer
	UmsCompletionList unsafe.Pointer
}

type COMPONENT_FILTER struct {
	ComponentFlags uint32
}

type RATE_QUOTA_LIMIT_Anonymous struct {
	Bitfield_ uint32
}

type RATE_QUOTA_LIMIT struct {
	RATE_QUOTA_LIMIT_Anonymous
}

func (this *RATE_QUOTA_LIMIT) RateData() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *RATE_QUOTA_LIMIT) RateDataVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *RATE_QUOTA_LIMIT) Anonymous() *RATE_QUOTA_LIMIT_Anonymous {
	return (*RATE_QUOTA_LIMIT_Anonymous)(unsafe.Pointer(this))
}

func (this *RATE_QUOTA_LIMIT) AnonymousVal() RATE_QUOTA_LIMIT_Anonymous {
	return *(*RATE_QUOTA_LIMIT_Anonymous)(unsafe.Pointer(this))
}

type QUOTA_LIMITS_EX struct {
	PagedPoolLimit        uintptr
	NonPagedPoolLimit     uintptr
	MinimumWorkingSetSize uintptr
	MaximumWorkingSetSize uintptr
	PagefileLimit         uintptr
	TimeLimit             int64
	WorkingSetLimit       uintptr
	Reserved2             uintptr
	Reserved3             uintptr
	Reserved4             uintptr
	Flags                 uint32
	CpuRateLimit          RATE_QUOTA_LIMIT
}

type PROCESS_MITIGATION_ASLR_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_ASLR_POLICY_Anonymous struct {
	PROCESS_MITIGATION_ASLR_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_ASLR_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_ASLR_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_ASLR_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_ASLR_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_ASLR_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_ASLR_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_ASLR_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_ASLR_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_ASLR_POLICY struct {
	PROCESS_MITIGATION_ASLR_POLICY_Anonymous
}

type PROCESS_MITIGATION_DEP_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_DEP_POLICY_Anonymous struct {
	PROCESS_MITIGATION_DEP_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_DEP_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_DEP_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_DEP_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_DEP_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_DEP_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_DEP_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_DEP_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_DEP_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_DEP_POLICY struct {
	PROCESS_MITIGATION_DEP_POLICY_Anonymous
	Permanent BOOLEAN
}

type PROCESS_MITIGATION_STRICT_HANDLE_CHECK_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_STRICT_HANDLE_CHECK_POLICY_Anonymous struct {
	PROCESS_MITIGATION_STRICT_HANDLE_CHECK_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_STRICT_HANDLE_CHECK_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_STRICT_HANDLE_CHECK_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_STRICT_HANDLE_CHECK_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_STRICT_HANDLE_CHECK_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_STRICT_HANDLE_CHECK_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_STRICT_HANDLE_CHECK_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_STRICT_HANDLE_CHECK_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_STRICT_HANDLE_CHECK_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_STRICT_HANDLE_CHECK_POLICY struct {
	PROCESS_MITIGATION_STRICT_HANDLE_CHECK_POLICY_Anonymous
}

type PROCESS_MITIGATION_SYSTEM_CALL_DISABLE_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_SYSTEM_CALL_DISABLE_POLICY_Anonymous struct {
	PROCESS_MITIGATION_SYSTEM_CALL_DISABLE_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_SYSTEM_CALL_DISABLE_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_SYSTEM_CALL_DISABLE_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_SYSTEM_CALL_DISABLE_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_SYSTEM_CALL_DISABLE_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_SYSTEM_CALL_DISABLE_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_SYSTEM_CALL_DISABLE_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_SYSTEM_CALL_DISABLE_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_SYSTEM_CALL_DISABLE_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_SYSTEM_CALL_DISABLE_POLICY struct {
	PROCESS_MITIGATION_SYSTEM_CALL_DISABLE_POLICY_Anonymous
}

type PROCESS_MITIGATION_EXTENSION_POINT_DISABLE_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_EXTENSION_POINT_DISABLE_POLICY_Anonymous struct {
	PROCESS_MITIGATION_EXTENSION_POINT_DISABLE_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_EXTENSION_POINT_DISABLE_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_EXTENSION_POINT_DISABLE_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_EXTENSION_POINT_DISABLE_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_EXTENSION_POINT_DISABLE_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_EXTENSION_POINT_DISABLE_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_EXTENSION_POINT_DISABLE_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_EXTENSION_POINT_DISABLE_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_EXTENSION_POINT_DISABLE_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_EXTENSION_POINT_DISABLE_POLICY struct {
	PROCESS_MITIGATION_EXTENSION_POINT_DISABLE_POLICY_Anonymous
}

type PROCESS_MITIGATION_DYNAMIC_CODE_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_DYNAMIC_CODE_POLICY_Anonymous struct {
	PROCESS_MITIGATION_DYNAMIC_CODE_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_DYNAMIC_CODE_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_DYNAMIC_CODE_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_DYNAMIC_CODE_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_DYNAMIC_CODE_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_DYNAMIC_CODE_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_DYNAMIC_CODE_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_DYNAMIC_CODE_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_DYNAMIC_CODE_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_DYNAMIC_CODE_POLICY struct {
	PROCESS_MITIGATION_DYNAMIC_CODE_POLICY_Anonymous
}

type PROCESS_MITIGATION_CONTROL_FLOW_GUARD_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_CONTROL_FLOW_GUARD_POLICY_Anonymous struct {
	PROCESS_MITIGATION_CONTROL_FLOW_GUARD_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_CONTROL_FLOW_GUARD_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_CONTROL_FLOW_GUARD_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_CONTROL_FLOW_GUARD_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_CONTROL_FLOW_GUARD_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_CONTROL_FLOW_GUARD_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_CONTROL_FLOW_GUARD_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_CONTROL_FLOW_GUARD_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_CONTROL_FLOW_GUARD_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_CONTROL_FLOW_GUARD_POLICY struct {
	PROCESS_MITIGATION_CONTROL_FLOW_GUARD_POLICY_Anonymous
}

type PROCESS_MITIGATION_BINARY_SIGNATURE_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_BINARY_SIGNATURE_POLICY_Anonymous struct {
	PROCESS_MITIGATION_BINARY_SIGNATURE_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_BINARY_SIGNATURE_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_BINARY_SIGNATURE_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_BINARY_SIGNATURE_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_BINARY_SIGNATURE_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_BINARY_SIGNATURE_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_BINARY_SIGNATURE_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_BINARY_SIGNATURE_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_BINARY_SIGNATURE_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_BINARY_SIGNATURE_POLICY struct {
	PROCESS_MITIGATION_BINARY_SIGNATURE_POLICY_Anonymous
}

type PROCESS_MITIGATION_FONT_DISABLE_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_FONT_DISABLE_POLICY_Anonymous struct {
	PROCESS_MITIGATION_FONT_DISABLE_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_FONT_DISABLE_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_FONT_DISABLE_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_FONT_DISABLE_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_FONT_DISABLE_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_FONT_DISABLE_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_FONT_DISABLE_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_FONT_DISABLE_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_FONT_DISABLE_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_FONT_DISABLE_POLICY struct {
	PROCESS_MITIGATION_FONT_DISABLE_POLICY_Anonymous
}

type PROCESS_MITIGATION_IMAGE_LOAD_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_IMAGE_LOAD_POLICY_Anonymous struct {
	PROCESS_MITIGATION_IMAGE_LOAD_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_IMAGE_LOAD_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_IMAGE_LOAD_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_IMAGE_LOAD_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_IMAGE_LOAD_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_IMAGE_LOAD_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_IMAGE_LOAD_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_IMAGE_LOAD_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_IMAGE_LOAD_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_IMAGE_LOAD_POLICY struct {
	PROCESS_MITIGATION_IMAGE_LOAD_POLICY_Anonymous
}

type PROCESS_MITIGATION_SYSTEM_CALL_FILTER_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_SYSTEM_CALL_FILTER_POLICY_Anonymous struct {
	PROCESS_MITIGATION_SYSTEM_CALL_FILTER_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_SYSTEM_CALL_FILTER_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_SYSTEM_CALL_FILTER_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_SYSTEM_CALL_FILTER_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_SYSTEM_CALL_FILTER_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_SYSTEM_CALL_FILTER_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_SYSTEM_CALL_FILTER_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_SYSTEM_CALL_FILTER_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_SYSTEM_CALL_FILTER_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_SYSTEM_CALL_FILTER_POLICY struct {
	PROCESS_MITIGATION_SYSTEM_CALL_FILTER_POLICY_Anonymous
}

type PROCESS_MITIGATION_PAYLOAD_RESTRICTION_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_PAYLOAD_RESTRICTION_POLICY_Anonymous struct {
	PROCESS_MITIGATION_PAYLOAD_RESTRICTION_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_PAYLOAD_RESTRICTION_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_PAYLOAD_RESTRICTION_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_PAYLOAD_RESTRICTION_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_PAYLOAD_RESTRICTION_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_PAYLOAD_RESTRICTION_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_PAYLOAD_RESTRICTION_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_PAYLOAD_RESTRICTION_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_PAYLOAD_RESTRICTION_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_PAYLOAD_RESTRICTION_POLICY struct {
	PROCESS_MITIGATION_PAYLOAD_RESTRICTION_POLICY_Anonymous
}

type PROCESS_MITIGATION_CHILD_PROCESS_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_CHILD_PROCESS_POLICY_Anonymous struct {
	PROCESS_MITIGATION_CHILD_PROCESS_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_CHILD_PROCESS_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_CHILD_PROCESS_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_CHILD_PROCESS_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_CHILD_PROCESS_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_CHILD_PROCESS_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_CHILD_PROCESS_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_CHILD_PROCESS_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_CHILD_PROCESS_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_CHILD_PROCESS_POLICY struct {
	PROCESS_MITIGATION_CHILD_PROCESS_POLICY_Anonymous
}

type PROCESS_MITIGATION_SIDE_CHANNEL_ISOLATION_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_SIDE_CHANNEL_ISOLATION_POLICY_Anonymous struct {
	PROCESS_MITIGATION_SIDE_CHANNEL_ISOLATION_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_SIDE_CHANNEL_ISOLATION_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_SIDE_CHANNEL_ISOLATION_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_SIDE_CHANNEL_ISOLATION_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_SIDE_CHANNEL_ISOLATION_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_SIDE_CHANNEL_ISOLATION_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_SIDE_CHANNEL_ISOLATION_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_SIDE_CHANNEL_ISOLATION_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_SIDE_CHANNEL_ISOLATION_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_SIDE_CHANNEL_ISOLATION_POLICY struct {
	PROCESS_MITIGATION_SIDE_CHANNEL_ISOLATION_POLICY_Anonymous
}

type PROCESS_MITIGATION_USER_SHADOW_STACK_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_USER_SHADOW_STACK_POLICY_Anonymous struct {
	PROCESS_MITIGATION_USER_SHADOW_STACK_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_USER_SHADOW_STACK_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_USER_SHADOW_STACK_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_USER_SHADOW_STACK_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_USER_SHADOW_STACK_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_USER_SHADOW_STACK_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_USER_SHADOW_STACK_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_USER_SHADOW_STACK_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_USER_SHADOW_STACK_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_USER_SHADOW_STACK_POLICY struct {
	PROCESS_MITIGATION_USER_SHADOW_STACK_POLICY_Anonymous
}

type PROCESS_MITIGATION_REDIRECTION_TRUST_POLICY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type PROCESS_MITIGATION_REDIRECTION_TRUST_POLICY_Anonymous struct {
	PROCESS_MITIGATION_REDIRECTION_TRUST_POLICY_Anonymous_Anonymous
}

func (this *PROCESS_MITIGATION_REDIRECTION_TRUST_POLICY_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_REDIRECTION_TRUST_POLICY_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_REDIRECTION_TRUST_POLICY_Anonymous) Anonymous() *PROCESS_MITIGATION_REDIRECTION_TRUST_POLICY_Anonymous_Anonymous {
	return (*PROCESS_MITIGATION_REDIRECTION_TRUST_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESS_MITIGATION_REDIRECTION_TRUST_POLICY_Anonymous) AnonymousVal() PROCESS_MITIGATION_REDIRECTION_TRUST_POLICY_Anonymous_Anonymous {
	return *(*PROCESS_MITIGATION_REDIRECTION_TRUST_POLICY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type PROCESS_MITIGATION_REDIRECTION_TRUST_POLICY struct {
	PROCESS_MITIGATION_REDIRECTION_TRUST_POLICY_Anonymous
}

type SILOOBJECT_BASIC_INFORMATION struct {
	SiloId            uint32
	SiloParentId      uint32
	NumberOfProcesses uint32
	IsInServerSilo    BOOLEAN
	Reserved          [3]byte
}

type SERVERSILO_BASIC_INFORMATION struct {
	ServiceSessionId     uint32
	State                SERVERSILO_STATE
	ExitStatus           uint32
	IsDownlevelContainer BOOLEAN
	ApiSetSchema         unsafe.Pointer
	HostApiSetSchema     unsafe.Pointer
}

type MEMORY_PARTITION_DEDICATED_MEMORY_ATTRIBUTE struct {
	Type     MEM_DEDICATED_ATTRIBUTE_TYPE
	Reserved uint32
	Value    uint64
}

type MEMORY_PARTITION_DEDICATED_MEMORY_INFORMATION struct {
	NextEntryOffset   uint32
	SizeOfInformation uint32
	Flags             uint32
	AttributesOffset  uint32
	AttributeCount    uint32
	Reserved          uint32
	TypeId            uint64
}

type SCRUB_DATA_INPUT struct {
	Size          uint32
	Flags         uint32
	MaximumIos    uint32
	ObjectId      [4]uint32
	Reserved      [41]uint32
	ResumeContext [1040]byte
}

type SCRUB_PARITY_EXTENT struct {
	Offset int64
	Length uint64
}

type SCRUB_PARITY_EXTENT_DATA struct {
	Size                         uint16
	Flags                        uint16
	NumberOfParityExtents        uint16
	MaximumNumberOfParityExtents uint16
	ParityExtents                [1]SCRUB_PARITY_EXTENT
}

type SCRUB_DATA_OUTPUT struct {
	Size                                               uint32
	Flags                                              uint32
	Status                                             uint32
	ErrorFileOffset                                    uint64
	ErrorLength                                        uint64
	NumberOfBytesRepaired                              uint64
	NumberOfBytesFailed                                uint64
	InternalFileReference                              uint64
	ResumeContextLength                                uint16
	ParityExtentDataOffset                             uint16
	Reserved                                           [9]uint32
	NumberOfMetadataBytesProcessed                     uint64
	NumberOfDataBytesProcessed                         uint64
	TotalNumberOfMetadataBytesInUse                    uint64
	TotalNumberOfDataBytesInUse                        uint64
	DataBytesSkippedDueToNoAllocation                  uint64
	DataBytesSkippedDueToInvalidRun                    uint64
	DataBytesSkippedDueToIntegrityStream               uint64
	DataBytesSkippedDueToRegionBeingClean              uint64
	DataBytesSkippedDueToLockConflict                  uint64
	DataBytesSkippedDueToNoScrubDataFlag               uint64
	DataBytesSkippedDueToNoScrubNonIntegrityStreamFlag uint64
	DataBytesScrubbed                                  uint64
	ResumeContext                                      [1040]byte
}

type SHARED_VIRTUAL_DISK_SUPPORT struct {
	SharedVirtualDiskSupport SharedVirtualDiskSupportType
	HandleState              SharedVirtualDiskHandleState
}

type REARRANGE_FILE_DATA struct {
	SourceStartingOffset uint64
	TargetOffset         uint64
	SourceFileHandle     HANDLE
	Length               uint32
	Flags                uint32
}

type REARRANGE_FILE_DATA32 struct {
	SourceStartingOffset uint64
	TargetOffset         uint64
	SourceFileHandle     uint32
	Length               uint32
	Flags                uint32
}

type SHUFFLE_FILE_DATA struct {
	StartingOffset int64
	Length         int64
	Flags          uint32
}

type NETWORK_APP_INSTANCE_EA struct {
	AppInstanceID syscall.GUID
	CsvFlags      uint32
}

type POWER_USER_PRESENCE struct {
	UserPresence POWER_USER_PRESENCE_TYPE
}

type POWER_SESSION_CONNECT struct {
	Connected BOOLEAN
	Console   BOOLEAN
}

type POWER_SESSION_TIMEOUTS struct {
	InputTimeout   uint32
	DisplayTimeout uint32
}

type POWER_SESSION_RIT_STATE struct {
	Active        BOOLEAN
	LastInputTime uint64
}

type POWER_SESSION_WINLOGON struct {
	SessionId uint32
	Console   BOOLEAN
	Locked    BOOLEAN
}

type POWER_SESSION_ALLOW_EXTERNAL_DMA_DEVICES struct {
	IsAllowed BOOLEAN
}

type POWER_IDLE_RESILIENCY struct {
	CoalescingTimeout    uint32
	IdleResiliencyPeriod uint32
}

type POWER_MONITOR_INVOCATION struct {
	Console       BOOLEAN
	RequestReason POWER_MONITOR_REQUEST_REASON
}

type RESUME_PERFORMANCE struct {
	PostTimeMs              uint32
	TotalResumeTimeMs       uint64
	ResumeCompleteTimestamp uint64
}

type NOTIFY_USER_POWER_SETTING struct {
	Guid syscall.GUID
}

type APPLICATIONLAUNCH_SETTING_VALUE struct {
	ActivationTime   int64
	Flags            uint32
	ButtonInstanceID uint32
}

type POWER_PLATFORM_INFORMATION struct {
	AoAc BOOLEAN
}

type PPM_WMI_LEGACY_PERFSTATE struct {
	Frequency        uint32
	Flags            uint32
	PercentFrequency uint32
}

type PPM_WMI_IDLE_STATE struct {
	Latency        uint32
	Power          uint32
	TimeCheck      uint32
	PromotePercent byte
	DemotePercent  byte
	StateType      byte
	Reserved       byte
	StateFlags     uint32
	Context        uint32
	IdleHandler    uint32
	Reserved1      uint32
}

type PPM_WMI_IDLE_STATES struct {
	Type             uint32
	Count            uint32
	TargetState      uint32
	OldState         uint32
	TargetProcessors uint64
	State            [1]PPM_WMI_IDLE_STATE
}

type PPM_WMI_IDLE_STATES_EX struct {
	Type             uint32
	Count            uint32
	TargetState      uint32
	OldState         uint32
	TargetProcessors unsafe.Pointer
	State            [1]PPM_WMI_IDLE_STATE
}

type PPM_WMI_PERF_STATE struct {
	Frequency        uint32
	Power            uint32
	PercentFrequency byte
	IncreaseLevel    byte
	DecreaseLevel    byte
	Type             byte
	IncreaseTime     uint32
	DecreaseTime     uint32
	Control          uint64
	Status           uint64
	HitCount         uint32
	Reserved1        uint32
	Reserved2        uint64
	Reserved3        uint64
}

type PPM_WMI_PERF_STATES struct {
	Count             uint32
	MaxFrequency      uint32
	CurrentState      uint32
	MaxPerfState      uint32
	MinPerfState      uint32
	LowestPerfState   uint32
	ThermalConstraint uint32
	BusyAdjThreshold  byte
	PolicyType        byte
	Type              byte
	Reserved          byte
	TimerInterval     uint32
	TargetProcessors  uint64
	PStateHandler     uint32
	PStateContext     uint32
	TStateHandler     uint32
	TStateContext     uint32
	FeedbackHandler   uint32
	Reserved1         uint32
	Reserved2         uint64
	State             [1]PPM_WMI_PERF_STATE
}

type PPM_WMI_PERF_STATES_EX struct {
	Count             uint32
	MaxFrequency      uint32
	CurrentState      uint32
	MaxPerfState      uint32
	MinPerfState      uint32
	LowestPerfState   uint32
	ThermalConstraint uint32
	BusyAdjThreshold  byte
	PolicyType        byte
	Type              byte
	Reserved          byte
	TimerInterval     uint32
	TargetProcessors  unsafe.Pointer
	PStateHandler     uint32
	PStateContext     uint32
	TStateHandler     uint32
	TStateContext     uint32
	FeedbackHandler   uint32
	Reserved1         uint32
	Reserved2         uint64
	State             [1]PPM_WMI_PERF_STATE
}

type PPM_IDLE_STATE_ACCOUNTING struct {
	IdleTransitions    uint32
	FailedTransitions  uint32
	InvalidBucketIndex uint32
	TotalTime          uint64
	IdleTimeBuckets    [6]uint32
}

type PPM_IDLE_ACCOUNTING struct {
	StateCount       uint32
	TotalTransitions uint32
	ResetCount       uint32
	StartTime        uint64
	State            [1]PPM_IDLE_STATE_ACCOUNTING
}

type PPM_IDLE_STATE_BUCKET_EX struct {
	TotalTimeUs uint64
	MinTimeUs   uint32
	MaxTimeUs   uint32
	Count       uint32
}

type PPM_IDLE_STATE_ACCOUNTING_EX struct {
	TotalTime            uint64
	IdleTransitions      uint32
	FailedTransitions    uint32
	InvalidBucketIndex   uint32
	MinTimeUs            uint32
	MaxTimeUs            uint32
	CancelledTransitions uint32
	IdleTimeBuckets      [16]PPM_IDLE_STATE_BUCKET_EX
}

type PPM_IDLE_ACCOUNTING_EX struct {
	StateCount       uint32
	TotalTransitions uint32
	ResetCount       uint32
	AbortCount       uint32
	StartTime        uint64
	State            [1]PPM_IDLE_STATE_ACCOUNTING_EX
}

type PPM_PERFSTATE_EVENT struct {
	State     uint32
	Status    uint32
	Latency   uint32
	Speed     uint32
	Processor uint32
}

type PPM_PERFSTATE_DOMAIN_EVENT struct {
	State      uint32
	Latency    uint32
	Speed      uint32
	Processors uint64
}

type PPM_IDLESTATE_EVENT struct {
	NewState   uint32
	OldState   uint32
	Processors uint64
}

type PPM_THERMALCHANGE_EVENT struct {
	ThermalConstraint uint32
	Processors        uint64
}

type PPM_THERMAL_POLICY_EVENT struct {
	Mode       byte
	Processors uint64
}

type PROCESSOR_IDLESTATE_INFO struct {
	TimeCheck      uint32
	DemotePercent  byte
	PromotePercent byte
	Spare          [2]byte
}

type PROCESSOR_IDLESTATE_POLICY_Flags_Anonymous struct {
	Bitfield_ uint16
}

type PROCESSOR_IDLESTATE_POLICY_Flags struct {
	PROCESSOR_IDLESTATE_POLICY_Flags_Anonymous
}

func (this *PROCESSOR_IDLESTATE_POLICY_Flags) AsWORD() *uint16 {
	return (*uint16)(unsafe.Pointer(this))
}

func (this *PROCESSOR_IDLESTATE_POLICY_Flags) AsWORDVal() uint16 {
	return *(*uint16)(unsafe.Pointer(this))
}

func (this *PROCESSOR_IDLESTATE_POLICY_Flags) Anonymous() *PROCESSOR_IDLESTATE_POLICY_Flags_Anonymous {
	return (*PROCESSOR_IDLESTATE_POLICY_Flags_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESSOR_IDLESTATE_POLICY_Flags) AnonymousVal() PROCESSOR_IDLESTATE_POLICY_Flags_Anonymous {
	return *(*PROCESSOR_IDLESTATE_POLICY_Flags_Anonymous)(unsafe.Pointer(this))
}

type PROCESSOR_IDLESTATE_POLICY struct {
	Revision    uint16
	Flags       PROCESSOR_IDLESTATE_POLICY_Flags
	PolicyCount uint32
	Policy      [3]PROCESSOR_IDLESTATE_INFO
}

type PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags_Anonymous struct {
	Bitfield_ byte
}

type PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags struct {
	PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags_Anonymous
}

func (this *PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags) AsBYTE() *byte {
	return (*byte)(unsafe.Pointer(this))
}

func (this *PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags) AsBYTEVal() byte {
	return *(*byte)(unsafe.Pointer(this))
}

func (this *PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags) Anonymous() *PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags_Anonymous {
	return (*PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags_Anonymous)(unsafe.Pointer(this))
}

func (this *PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags) AnonymousVal() PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags_Anonymous {
	return *(*PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags_Anonymous)(unsafe.Pointer(this))
}

type PROCESSOR_PERFSTATE_POLICY_Anonymous struct {
	Data [1]byte
}

func (this *PROCESSOR_PERFSTATE_POLICY_Anonymous) Spare() *byte {
	return (*byte)(unsafe.Pointer(this))
}

func (this *PROCESSOR_PERFSTATE_POLICY_Anonymous) SpareVal() byte {
	return *(*byte)(unsafe.Pointer(this))
}

func (this *PROCESSOR_PERFSTATE_POLICY_Anonymous) Flags() *PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags {
	return (*PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags)(unsafe.Pointer(this))
}

func (this *PROCESSOR_PERFSTATE_POLICY_Anonymous) FlagsVal() PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags {
	return *(*PROCESSOR_PERFSTATE_POLICY_Anonymous_Flags)(unsafe.Pointer(this))
}

type PROCESSOR_PERFSTATE_POLICY struct {
	Revision         uint32
	MaxThrottle      byte
	MinThrottle      byte
	BusyAdjThreshold byte
	PROCESSOR_PERFSTATE_POLICY_Anonymous
	TimeCheck       uint32
	IncreaseTime    uint32
	DecreaseTime    uint32
	IncreasePercent uint32
	DecreasePercent uint32
}

type HIBERFILE_BUCKET struct {
	MaxPhysicalMemory     uint64
	PhysicalMemoryPercent [3]uint32
}

type IMAGE_DOS_HEADER struct {
	E_magic    uint16
	E_cblp     uint16
	E_cp       uint16
	E_crlc     uint16
	E_cparhdr  uint16
	E_minalloc uint16
	E_maxalloc uint16
	E_ss       uint16
	E_sp       uint16
	E_csum     uint16
	E_ip       uint16
	E_cs       uint16
	E_lfarlc   uint16
	E_ovno     uint16
	E_res      [4]uint16
	E_oemid    uint16
	E_oeminfo  uint16
	E_res2     [10]uint16
	E_lfanew   int32
}

type IMAGE_OS2_HEADER struct {
	Ne_magic        uint16
	Ne_ver          CHAR
	Ne_rev          CHAR
	Ne_enttab       uint16
	Ne_cbenttab     uint16
	Ne_crc          int32
	Ne_flags        uint16
	Ne_autodata     uint16
	Ne_heap         uint16
	Ne_stack        uint16
	Ne_csip         int32
	Ne_sssp         int32
	Ne_cseg         uint16
	Ne_cmod         uint16
	Ne_cbnrestab    uint16
	Ne_segtab       uint16
	Ne_rsrctab      uint16
	Ne_restab       uint16
	Ne_modtab       uint16
	Ne_imptab       uint16
	Ne_nrestab      int32
	Ne_cmovent      uint16
	Ne_align        uint16
	Ne_cres         uint16
	Ne_exetyp       byte
	Ne_flagsothers  byte
	Ne_pretthunks   uint16
	Ne_psegrefbytes uint16
	Ne_swaparea     uint16
	Ne_expver       uint16
}

type IMAGE_VXD_HEADER struct {
	E32_magic        uint16
	E32_border       byte
	E32_worder       byte
	E32_level        uint32
	E32_cpu          uint16
	E32_os           uint16
	E32_ver          uint32
	E32_mflags       uint32
	E32_mpages       uint32
	E32_startobj     uint32
	E32_eip          uint32
	E32_stackobj     uint32
	E32_esp          uint32
	E32_pagesize     uint32
	E32_lastpagesize uint32
	E32_fixupsize    uint32
	E32_fixupsum     uint32
	E32_ldrsize      uint32
	E32_ldrsum       uint32
	E32_objtab       uint32
	E32_objcnt       uint32
	E32_objmap       uint32
	E32_itermap      uint32
	E32_rsrctab      uint32
	E32_rsrccnt      uint32
	E32_restab       uint32
	E32_enttab       uint32
	E32_dirtab       uint32
	E32_dircnt       uint32
	E32_fpagetab     uint32
	E32_frectab      uint32
	E32_impmod       uint32
	E32_impmodcnt    uint32
	E32_impproc      uint32
	E32_pagesum      uint32
	E32_datapage     uint32
	E32_preload      uint32
	E32_nrestab      uint32
	E32_cbnrestab    uint32
	E32_nressum      uint32
	E32_autodata     uint32
	E32_debuginfo    uint32
	E32_debuglen     uint32
	E32_instpreload  uint32
	E32_instdemand   uint32
	E32_heapsize     uint32
	E32_res3         [12]byte
	E32_winresoff    uint32
	E32_winreslen    uint32
	E32_devid        uint16
	E32_ddkver       uint16
}

type ANON_OBJECT_HEADER struct {
	Sig1          uint16
	Sig2          uint16
	Version       uint16
	Machine       uint16
	TimeDateStamp uint32
	ClassID       syscall.GUID
	SizeOfData    uint32
}

type ANON_OBJECT_HEADER_V2 struct {
	Sig1           uint16
	Sig2           uint16
	Version        uint16
	Machine        uint16
	TimeDateStamp  uint32
	ClassID        syscall.GUID
	SizeOfData     uint32
	Flags          uint32
	MetaDataSize   uint32
	MetaDataOffset uint32
}

type ANON_OBJECT_HEADER_BIGOBJ struct {
	Sig1                 uint16
	Sig2                 uint16
	Version              uint16
	Machine              uint16
	TimeDateStamp        uint32
	ClassID              syscall.GUID
	SizeOfData           uint32
	Flags                uint32
	MetaDataSize         uint32
	MetaDataOffset       uint32
	NumberOfSections     uint32
	PointerToSymbolTable uint32
	NumberOfSymbols      uint32
}

type IMAGE_SYMBOL_N_Name struct {
	Short uint32
	Long  uint32
}

type IMAGE_SYMBOL_N struct {
	Data [2]uint32
}

func (this *IMAGE_SYMBOL_N) ShortName() *[8]byte {
	return (*[8]byte)(unsafe.Pointer(this))
}

func (this *IMAGE_SYMBOL_N) ShortNameVal() [8]byte {
	return *(*[8]byte)(unsafe.Pointer(this))
}

func (this *IMAGE_SYMBOL_N) Name() *IMAGE_SYMBOL_N_Name {
	return (*IMAGE_SYMBOL_N_Name)(unsafe.Pointer(this))
}

func (this *IMAGE_SYMBOL_N) NameVal() IMAGE_SYMBOL_N_Name {
	return *(*IMAGE_SYMBOL_N_Name)(unsafe.Pointer(this))
}

func (this *IMAGE_SYMBOL_N) LongName() *[2]uint32 {
	return (*[2]uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_SYMBOL_N) LongNameVal() [2]uint32 {
	return *(*[2]uint32)(unsafe.Pointer(this))
}

type IMAGE_SYMBOL struct {
	N                  IMAGE_SYMBOL_N
	Value              uint32
	SectionNumber      int16
	Type               uint16
	StorageClass       byte
	NumberOfAuxSymbols byte
}

type IMAGE_SYMBOL_EX_N_Name struct {
	Short uint32
	Long  uint32
}

type IMAGE_SYMBOL_EX_N struct {
	Data [2]uint32
}

func (this *IMAGE_SYMBOL_EX_N) ShortName() *[8]byte {
	return (*[8]byte)(unsafe.Pointer(this))
}

func (this *IMAGE_SYMBOL_EX_N) ShortNameVal() [8]byte {
	return *(*[8]byte)(unsafe.Pointer(this))
}

func (this *IMAGE_SYMBOL_EX_N) Name() *IMAGE_SYMBOL_EX_N_Name {
	return (*IMAGE_SYMBOL_EX_N_Name)(unsafe.Pointer(this))
}

func (this *IMAGE_SYMBOL_EX_N) NameVal() IMAGE_SYMBOL_EX_N_Name {
	return *(*IMAGE_SYMBOL_EX_N_Name)(unsafe.Pointer(this))
}

func (this *IMAGE_SYMBOL_EX_N) LongName() *[2]uint32 {
	return (*[2]uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_SYMBOL_EX_N) LongNameVal() [2]uint32 {
	return *(*[2]uint32)(unsafe.Pointer(this))
}

type IMAGE_SYMBOL_EX struct {
	N                  IMAGE_SYMBOL_EX_N
	Value              uint32
	SectionNumber      int32
	Type               uint16
	StorageClass       byte
	NumberOfAuxSymbols byte
}

type IMAGE_AUX_SYMBOL_TOKEN_DEF struct {
	BAuxType         byte
	BReserved        byte
	SymbolTableIndex uint32
	RgbReserved      [12]byte
}

type IMAGE_AUX_SYMBOL_Sym_Misc_LnSz struct {
	Linenumber uint16
	Size       uint16
}

type IMAGE_AUX_SYMBOL_Sym_Misc struct {
	Data [1]uint32
}

func (this *IMAGE_AUX_SYMBOL_Sym_Misc) LnSz() *IMAGE_AUX_SYMBOL_Sym_Misc_LnSz {
	return (*IMAGE_AUX_SYMBOL_Sym_Misc_LnSz)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_Sym_Misc) LnSzVal() IMAGE_AUX_SYMBOL_Sym_Misc_LnSz {
	return *(*IMAGE_AUX_SYMBOL_Sym_Misc_LnSz)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_Sym_Misc) TotalSize() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_Sym_Misc) TotalSizeVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type IMAGE_AUX_SYMBOL_Sym_FcnAry_Function struct {
	PointerToLinenumber   uint32
	PointerToNextFunction uint32
}

type IMAGE_AUX_SYMBOL_Sym_FcnAry_Array struct {
	Dimension [4]uint16
}

type IMAGE_AUX_SYMBOL_Sym_FcnAry struct {
	Data [2]uint32
}

func (this *IMAGE_AUX_SYMBOL_Sym_FcnAry) Function() *IMAGE_AUX_SYMBOL_Sym_FcnAry_Function {
	return (*IMAGE_AUX_SYMBOL_Sym_FcnAry_Function)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_Sym_FcnAry) FunctionVal() IMAGE_AUX_SYMBOL_Sym_FcnAry_Function {
	return *(*IMAGE_AUX_SYMBOL_Sym_FcnAry_Function)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_Sym_FcnAry) Array() *IMAGE_AUX_SYMBOL_Sym_FcnAry_Array {
	return (*IMAGE_AUX_SYMBOL_Sym_FcnAry_Array)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_Sym_FcnAry) ArrayVal() IMAGE_AUX_SYMBOL_Sym_FcnAry_Array {
	return *(*IMAGE_AUX_SYMBOL_Sym_FcnAry_Array)(unsafe.Pointer(this))
}

type IMAGE_AUX_SYMBOL_Sym struct {
	TagIndex uint32
	Misc     IMAGE_AUX_SYMBOL_Sym_Misc
	FcnAry   IMAGE_AUX_SYMBOL_Sym_FcnAry
	TvIndex  uint16
}

type IMAGE_AUX_SYMBOL_File struct {
	Name [18]byte
}

type IMAGE_AUX_SYMBOL_Section struct {
	Length              uint32
	NumberOfRelocations uint16
	NumberOfLinenumbers uint16
	CheckSum            uint32
	Number              int16
	Selection           byte
	BReserved           byte
	HighNumber          int16
}

type IMAGE_AUX_SYMBOL_CRC struct {
	Crc         uint32
	RgbReserved [14]byte
}

type IMAGE_AUX_SYMBOL struct {
	Data [5]uint32
}

func (this *IMAGE_AUX_SYMBOL) Sym() *IMAGE_AUX_SYMBOL_Sym {
	return (*IMAGE_AUX_SYMBOL_Sym)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL) SymVal() IMAGE_AUX_SYMBOL_Sym {
	return *(*IMAGE_AUX_SYMBOL_Sym)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL) File() *IMAGE_AUX_SYMBOL_File {
	return (*IMAGE_AUX_SYMBOL_File)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL) FileVal() IMAGE_AUX_SYMBOL_File {
	return *(*IMAGE_AUX_SYMBOL_File)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL) Section() *IMAGE_AUX_SYMBOL_Section {
	return (*IMAGE_AUX_SYMBOL_Section)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL) SectionVal() IMAGE_AUX_SYMBOL_Section {
	return *(*IMAGE_AUX_SYMBOL_Section)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL) TokenDef() *IMAGE_AUX_SYMBOL_TOKEN_DEF {
	return (*IMAGE_AUX_SYMBOL_TOKEN_DEF)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL) TokenDefVal() IMAGE_AUX_SYMBOL_TOKEN_DEF {
	return *(*IMAGE_AUX_SYMBOL_TOKEN_DEF)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL) CRC() *IMAGE_AUX_SYMBOL_CRC {
	return (*IMAGE_AUX_SYMBOL_CRC)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL) CRCVal() IMAGE_AUX_SYMBOL_CRC {
	return *(*IMAGE_AUX_SYMBOL_CRC)(unsafe.Pointer(this))
}

type IMAGE_AUX_SYMBOL_EX_Sym struct {
	WeakDefaultSymIndex uint32
	WeakSearchType      uint32
	RgbReserved         [12]byte
}

type IMAGE_AUX_SYMBOL_EX_File struct {
	Name [20]byte
}

type IMAGE_AUX_SYMBOL_EX_Section struct {
	Length              uint32
	NumberOfRelocations uint16
	NumberOfLinenumbers uint16
	CheckSum            uint32
	Number              int16
	Selection           byte
	BReserved           byte
	HighNumber          int16
	RgbReserved         [2]byte
}

type IMAGE_AUX_SYMBOL_EX_Anonymous struct {
	TokenDef    IMAGE_AUX_SYMBOL_TOKEN_DEF
	RgbReserved [2]byte
}

type IMAGE_AUX_SYMBOL_EX_CRC struct {
	Crc         uint32
	RgbReserved [16]byte
}

type IMAGE_AUX_SYMBOL_EX struct {
	IMAGE_AUX_SYMBOL_EX_Anonymous
}

func (this *IMAGE_AUX_SYMBOL_EX) Sym() *IMAGE_AUX_SYMBOL_EX_Sym {
	return (*IMAGE_AUX_SYMBOL_EX_Sym)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_EX) SymVal() IMAGE_AUX_SYMBOL_EX_Sym {
	return *(*IMAGE_AUX_SYMBOL_EX_Sym)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_EX) File() *IMAGE_AUX_SYMBOL_EX_File {
	return (*IMAGE_AUX_SYMBOL_EX_File)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_EX) FileVal() IMAGE_AUX_SYMBOL_EX_File {
	return *(*IMAGE_AUX_SYMBOL_EX_File)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_EX) Section() *IMAGE_AUX_SYMBOL_EX_Section {
	return (*IMAGE_AUX_SYMBOL_EX_Section)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_EX) SectionVal() IMAGE_AUX_SYMBOL_EX_Section {
	return *(*IMAGE_AUX_SYMBOL_EX_Section)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_EX) Anonymous() *IMAGE_AUX_SYMBOL_EX_Anonymous {
	return (*IMAGE_AUX_SYMBOL_EX_Anonymous)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_EX) AnonymousVal() IMAGE_AUX_SYMBOL_EX_Anonymous {
	return *(*IMAGE_AUX_SYMBOL_EX_Anonymous)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_EX) CRC() *IMAGE_AUX_SYMBOL_EX_CRC {
	return (*IMAGE_AUX_SYMBOL_EX_CRC)(unsafe.Pointer(this))
}

func (this *IMAGE_AUX_SYMBOL_EX) CRCVal() IMAGE_AUX_SYMBOL_EX_CRC {
	return *(*IMAGE_AUX_SYMBOL_EX_CRC)(unsafe.Pointer(this))
}

type IMAGE_RELOCATION_Anonymous struct {
	Data [1]uint32
}

func (this *IMAGE_RELOCATION_Anonymous) VirtualAddress() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_RELOCATION_Anonymous) VirtualAddressVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_RELOCATION_Anonymous) RelocCount() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_RELOCATION_Anonymous) RelocCountVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type IMAGE_RELOCATION struct {
	IMAGE_RELOCATION_Anonymous
	SymbolTableIndex uint32
	Type             uint16
}

type IMAGE_LINENUMBER_Type struct {
	Data [1]uint32
}

func (this *IMAGE_LINENUMBER_Type) SymbolTableIndex() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_LINENUMBER_Type) SymbolTableIndexVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_LINENUMBER_Type) VirtualAddress() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_LINENUMBER_Type) VirtualAddressVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type IMAGE_LINENUMBER struct {
	Type       IMAGE_LINENUMBER_Type
	Linenumber uint16
}

type IMAGE_BASE_RELOCATION struct {
	VirtualAddress uint32
	SizeOfBlock    uint32
}

type IMAGE_ARCHIVE_MEMBER_HEADER struct {
	Name      [16]byte
	Date      [12]byte
	UserID    [6]byte
	GroupID   [6]byte
	Mode      [8]byte
	Size      [10]byte
	EndHeader [2]byte
}

type IMAGE_EXPORT_DIRECTORY struct {
	Characteristics       uint32
	TimeDateStamp         uint32
	MajorVersion          uint16
	MinorVersion          uint16
	Name                  uint32
	Base                  uint32
	NumberOfFunctions     uint32
	NumberOfNames         uint32
	AddressOfFunctions    uint32
	AddressOfNames        uint32
	AddressOfNameOrdinals uint32
}

type IMAGE_IMPORT_BY_NAME struct {
	Hint uint16
	Name [1]CHAR
}

type IMAGE_TLS_DIRECTORY64_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type IMAGE_TLS_DIRECTORY64_Anonymous struct {
	IMAGE_TLS_DIRECTORY64_Anonymous_Anonymous
}

func (this *IMAGE_TLS_DIRECTORY64_Anonymous) Characteristics() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_TLS_DIRECTORY64_Anonymous) CharacteristicsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_TLS_DIRECTORY64_Anonymous) Anonymous() *IMAGE_TLS_DIRECTORY64_Anonymous_Anonymous {
	return (*IMAGE_TLS_DIRECTORY64_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *IMAGE_TLS_DIRECTORY64_Anonymous) AnonymousVal() IMAGE_TLS_DIRECTORY64_Anonymous_Anonymous {
	return *(*IMAGE_TLS_DIRECTORY64_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type IMAGE_TLS_DIRECTORY64 struct {
	StartAddressOfRawData uint64
	EndAddressOfRawData   uint64
	AddressOfIndex        uint64
	AddressOfCallBacks    uint64
	SizeOfZeroFill        uint32
	IMAGE_TLS_DIRECTORY64_Anonymous
}

type IMAGE_TLS_DIRECTORY32_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type IMAGE_TLS_DIRECTORY32_Anonymous struct {
	IMAGE_TLS_DIRECTORY32_Anonymous_Anonymous
}

func (this *IMAGE_TLS_DIRECTORY32_Anonymous) Characteristics() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_TLS_DIRECTORY32_Anonymous) CharacteristicsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_TLS_DIRECTORY32_Anonymous) Anonymous() *IMAGE_TLS_DIRECTORY32_Anonymous_Anonymous {
	return (*IMAGE_TLS_DIRECTORY32_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *IMAGE_TLS_DIRECTORY32_Anonymous) AnonymousVal() IMAGE_TLS_DIRECTORY32_Anonymous_Anonymous {
	return *(*IMAGE_TLS_DIRECTORY32_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type IMAGE_TLS_DIRECTORY32 struct {
	StartAddressOfRawData uint32
	EndAddressOfRawData   uint32
	AddressOfIndex        uint32
	AddressOfCallBacks    uint32
	SizeOfZeroFill        uint32
	IMAGE_TLS_DIRECTORY32_Anonymous
}

type IMAGE_IMPORT_DESCRIPTOR_Anonymous struct {
	Data [1]uint32
}

func (this *IMAGE_IMPORT_DESCRIPTOR_Anonymous) Characteristics() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_IMPORT_DESCRIPTOR_Anonymous) CharacteristicsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_IMPORT_DESCRIPTOR_Anonymous) OriginalFirstThunk() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_IMPORT_DESCRIPTOR_Anonymous) OriginalFirstThunkVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type IMAGE_IMPORT_DESCRIPTOR struct {
	IMAGE_IMPORT_DESCRIPTOR_Anonymous
	TimeDateStamp  uint32
	ForwarderChain uint32
	Name           uint32
	FirstThunk     uint32
}

type IMAGE_BOUND_IMPORT_DESCRIPTOR struct {
	TimeDateStamp               uint32
	OffsetModuleName            uint16
	NumberOfModuleForwarderRefs uint16
}

type IMAGE_BOUND_FORWARDER_REF struct {
	TimeDateStamp    uint32
	OffsetModuleName uint16
	Reserved         uint16
}

type IMAGE_RESOURCE_DIRECTORY struct {
	Characteristics      uint32
	TimeDateStamp        uint32
	MajorVersion         uint16
	MinorVersion         uint16
	NumberOfNamedEntries uint16
	NumberOfIdEntries    uint16
}

type IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous1_Anonymous struct {
	Bitfield_ uint32
}

type IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous1 struct {
	IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous1_Anonymous
}

func (this *IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous1) Anonymous() *IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous1_Anonymous {
	return (*IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous1_Anonymous)(unsafe.Pointer(this))
}

func (this *IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous1) AnonymousVal() IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous1_Anonymous {
	return *(*IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous1_Anonymous)(unsafe.Pointer(this))
}

func (this *IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous1) Name() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous1) NameVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous1) Id() *uint16 {
	return (*uint16)(unsafe.Pointer(this))
}

func (this *IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous1) IdVal() uint16 {
	return *(*uint16)(unsafe.Pointer(this))
}

type IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous2_Anonymous struct {
	Bitfield_ uint32
}

type IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous2 struct {
	IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous2_Anonymous
}

func (this *IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous2) OffsetToData() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous2) OffsetToDataVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous2) Anonymous() *IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous2_Anonymous {
	return (*IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous2_Anonymous)(unsafe.Pointer(this))
}

func (this *IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous2) AnonymousVal() IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous2_Anonymous {
	return *(*IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous2_Anonymous)(unsafe.Pointer(this))
}

type IMAGE_RESOURCE_DIRECTORY_ENTRY struct {
	IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous1
	IMAGE_RESOURCE_DIRECTORY_ENTRY_Anonymous2
}

type IMAGE_RESOURCE_DIRECTORY_STRING struct {
	Length     uint16
	NameString [1]CHAR
}

type IMAGE_RESOURCE_DIR_STRING_U struct {
	Length     uint16
	NameString [1]uint16
}

type IMAGE_RESOURCE_DATA_ENTRY struct {
	OffsetToData uint32
	Size         uint32
	CodePage     uint32
	Reserved     uint32
}

type IMAGE_DYNAMIC_RELOCATION_TABLE struct {
	Version uint32
	Size    uint32
}

type IMAGE_DYNAMIC_RELOCATION32 struct {
	Symbol        uint32
	BaseRelocSize uint32
}

type IMAGE_DYNAMIC_RELOCATION64 struct {
	Symbol        uint64
	BaseRelocSize uint32
}

type IMAGE_DYNAMIC_RELOCATION32_V2 struct {
	HeaderSize    uint32
	FixupInfoSize uint32
	Symbol        uint32
	SymbolGroup   uint32
	Flags         uint32
}

type IMAGE_DYNAMIC_RELOCATION64_V2 struct {
	HeaderSize    uint32
	FixupInfoSize uint32
	Symbol        uint64
	SymbolGroup   uint32
	Flags         uint32
}

type IMAGE_PROLOGUE_DYNAMIC_RELOCATION_HEADER struct {
	PrologueByteCount byte
}

type IMAGE_EPILOGUE_DYNAMIC_RELOCATION_HEADER struct {
	EpilogueCount               uint32
	EpilogueByteCount           byte
	BranchDescriptorElementSize byte
	BranchDescriptorCount       uint16
}

type IMAGE_IMPORT_CONTROL_TRANSFER_DYNAMIC_RELOCATION struct {
	Bitfield_ uint32
}

type IMAGE_INDIR_CONTROL_TRANSFER_DYNAMIC_RELOCATION struct {
	Bitfield_ uint16
}

type IMAGE_SWITCHTABLE_BRANCH_DYNAMIC_RELOCATION struct {
	Bitfield_ uint16
}

type IMAGE_HOT_PATCH_INFO struct {
	Version        uint32
	Size           uint32
	SequenceNumber uint32
	BaseImageList  uint32
	BaseImageCount uint32
	BufferOffset   uint32
	ExtraPatchSize uint32
}

type IMAGE_HOT_PATCH_BASE struct {
	SequenceNumber        uint32
	Flags                 uint32
	OriginalTimeDateStamp uint32
	OriginalCheckSum      uint32
	CodeIntegrityInfo     uint32
	CodeIntegritySize     uint32
	PatchTable            uint32
	BufferOffset          uint32
}

type IMAGE_HOT_PATCH_HASHES struct {
	SHA256 [32]byte
	SHA1   [20]byte
}

type IMAGE_CE_RUNTIME_FUNCTION_ENTRY struct {
	FuncStart uint32
	Bitfield_ uint32
}

type IMAGE_ARM_RUNTIME_FUNCTION_ENTRY_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type IMAGE_ARM_RUNTIME_FUNCTION_ENTRY_Anonymous struct {
	IMAGE_ARM_RUNTIME_FUNCTION_ENTRY_Anonymous_Anonymous
}

func (this *IMAGE_ARM_RUNTIME_FUNCTION_ENTRY_Anonymous) UnwindData() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_ARM_RUNTIME_FUNCTION_ENTRY_Anonymous) UnwindDataVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_ARM_RUNTIME_FUNCTION_ENTRY_Anonymous) Anonymous() *IMAGE_ARM_RUNTIME_FUNCTION_ENTRY_Anonymous_Anonymous {
	return (*IMAGE_ARM_RUNTIME_FUNCTION_ENTRY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *IMAGE_ARM_RUNTIME_FUNCTION_ENTRY_Anonymous) AnonymousVal() IMAGE_ARM_RUNTIME_FUNCTION_ENTRY_Anonymous_Anonymous {
	return *(*IMAGE_ARM_RUNTIME_FUNCTION_ENTRY_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type IMAGE_ARM_RUNTIME_FUNCTION_ENTRY struct {
	BeginAddress uint32
	IMAGE_ARM_RUNTIME_FUNCTION_ENTRY_Anonymous
}

type IMAGE_ARM64_RUNTIME_FUNCTION_ENTRY_XDATA_Anonymous struct {
	Bitfield_ uint32
}

type IMAGE_ARM64_RUNTIME_FUNCTION_ENTRY_XDATA struct {
	IMAGE_ARM64_RUNTIME_FUNCTION_ENTRY_XDATA_Anonymous
}

func (this *IMAGE_ARM64_RUNTIME_FUNCTION_ENTRY_XDATA) HeaderData() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_ARM64_RUNTIME_FUNCTION_ENTRY_XDATA) HeaderDataVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_ARM64_RUNTIME_FUNCTION_ENTRY_XDATA) Anonymous() *IMAGE_ARM64_RUNTIME_FUNCTION_ENTRY_XDATA_Anonymous {
	return (*IMAGE_ARM64_RUNTIME_FUNCTION_ENTRY_XDATA_Anonymous)(unsafe.Pointer(this))
}

func (this *IMAGE_ARM64_RUNTIME_FUNCTION_ENTRY_XDATA) AnonymousVal() IMAGE_ARM64_RUNTIME_FUNCTION_ENTRY_XDATA_Anonymous {
	return *(*IMAGE_ARM64_RUNTIME_FUNCTION_ENTRY_XDATA_Anonymous)(unsafe.Pointer(this))
}

type IMAGE_ALPHA64_RUNTIME_FUNCTION_ENTRY struct {
	BeginAddress     uint64
	EndAddress       uint64
	ExceptionHandler uint64
	HandlerData      uint64
	PrologEndAddress uint64
}

type IMAGE_ALPHA_RUNTIME_FUNCTION_ENTRY struct {
	BeginAddress     uint32
	EndAddress       uint32
	ExceptionHandler uint32
	HandlerData      uint32
	PrologEndAddress uint32
}

type IMAGE_DEBUG_MISC struct {
	DataType uint32
	Length   uint32
	Unicode  BOOLEAN
	Reserved [3]byte
	Data     [1]byte
}

type IMAGE_SEPARATE_DEBUG_HEADER struct {
	Signature          uint16
	Flags              uint16
	Machine            uint16
	Characteristics    uint16
	TimeDateStamp      uint32
	CheckSum           uint32
	ImageBase          uint32
	SizeOfImage        uint32
	NumberOfSections   uint32
	ExportedNamesSize  uint32
	DebugDirectorySize uint32
	SectionAlignment   uint32
	Reserved           [2]uint32
}

type NON_PAGED_DEBUG_INFO struct {
	Signature       uint16
	Flags           uint16
	Size            uint32
	Machine         uint16
	Characteristics uint16
	TimeDateStamp   uint32
	CheckSum        uint32
	SizeOfImage     uint32
	ImageBase       uint64
}

type IMAGE_ARCHITECTURE_HEADER struct {
	Bitfield_     uint32
	FirstEntryRVA uint32
}

type IMAGE_ARCHITECTURE_ENTRY struct {
	FixupInstRVA uint32
	NewInst      uint32
}

type IMPORT_OBJECT_HEADER_Anonymous struct {
	Data [1]uint16
}

func (this *IMPORT_OBJECT_HEADER_Anonymous) Ordinal() *uint16 {
	return (*uint16)(unsafe.Pointer(this))
}

func (this *IMPORT_OBJECT_HEADER_Anonymous) OrdinalVal() uint16 {
	return *(*uint16)(unsafe.Pointer(this))
}

func (this *IMPORT_OBJECT_HEADER_Anonymous) Hint() *uint16 {
	return (*uint16)(unsafe.Pointer(this))
}

func (this *IMPORT_OBJECT_HEADER_Anonymous) HintVal() uint16 {
	return *(*uint16)(unsafe.Pointer(this))
}

type IMPORT_OBJECT_HEADER struct {
	Sig1          uint16
	Sig2          uint16
	Version       uint16
	Machine       uint16
	TimeDateStamp uint32
	SizeOfData    uint32
	IMPORT_OBJECT_HEADER_Anonymous
	Bitfield_ uint16
}

type IMAGE_POLICY_ENTRY_U struct {
	Data [1]uint64
}

func (this *IMAGE_POLICY_ENTRY_U) None() *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) NoneVal() unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) BoolValue() *BOOLEAN {
	return (*BOOLEAN)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) BoolValueVal() BOOLEAN {
	return *(*BOOLEAN)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) Int8Value() *int8 {
	return (*int8)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) Int8ValueVal() int8 {
	return *(*int8)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) UInt8Value() *byte {
	return (*byte)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) UInt8ValueVal() byte {
	return *(*byte)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) Int16Value() *int16 {
	return (*int16)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) Int16ValueVal() int16 {
	return *(*int16)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) UInt16Value() *uint16 {
	return (*uint16)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) UInt16ValueVal() uint16 {
	return *(*uint16)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) Int32Value() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) Int32ValueVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) UInt32Value() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) UInt32ValueVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) Int64Value() *int64 {
	return (*int64)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) Int64ValueVal() int64 {
	return *(*int64)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) UInt64Value() *uint64 {
	return (*uint64)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) UInt64ValueVal() uint64 {
	return *(*uint64)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) AnsiStringValue() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) AnsiStringValueVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) UnicodeStringValue() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *IMAGE_POLICY_ENTRY_U) UnicodeStringValueVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type IMAGE_POLICY_ENTRY struct {
	Type     IMAGE_POLICY_ENTRY_TYPE
	PolicyId IMAGE_POLICY_ID
	U        IMAGE_POLICY_ENTRY_U
}

type IMAGE_POLICY_METADATA struct {
	Version       byte
	Reserved0     [7]byte
	ApplicationId uint64
	Policies      [1]IMAGE_POLICY_ENTRY
}

type HEAP_OPTIMIZE_RESOURCES_INFORMATION struct {
	Version uint32
	Flags   uint32
}

type SUPPORTED_OS_INFO struct {
	MajorVersion uint16
	MinorVersion uint16
}

type MAXVERSIONTESTED_INFO struct {
	MaxVersionTested uint64
}

type PACKEDEVENTINFO struct {
	UlSize                uint32
	UlNumEventsForLogFile uint32
	UlOffsets             [1]uint32
}

type TAPE_GET_DRIVE_PARAMETERS struct {
	ECC                   BOOLEAN
	Compression           BOOLEAN
	DataPadding           BOOLEAN
	ReportSetmarks        BOOLEAN
	DefaultBlockSize      uint32
	MaximumBlockSize      uint32
	MinimumBlockSize      uint32
	MaximumPartitionCount uint32
	FeaturesLow           uint32
	FeaturesHigh          TAPE_GET_DRIVE_PARAMETERS_FEATURES_HIGH
	EOTWarningZoneSize    uint32
}

type TAPE_SET_DRIVE_PARAMETERS struct {
	ECC                BOOLEAN
	Compression        BOOLEAN
	DataPadding        BOOLEAN
	ReportSetmarks     BOOLEAN
	EOTWarningZoneSize uint32
}

type TAPE_GET_MEDIA_PARAMETERS struct {
	Capacity       int64
	Remaining      int64
	BlockSize      uint32
	PartitionCount uint32
	WriteProtected BOOLEAN
}

type TAPE_SET_MEDIA_PARAMETERS struct {
	BlockSize uint32
}

type TAPE_CREATE_PARTITION struct {
	Method uint32
	Count  uint32
	Size   uint32
}

type TAPE_WMI_OPERATIONS struct {
	Method         uint32
	DataBufferSize uint32
	DataBuffer     unsafe.Pointer
}

type TRANSACTION_BASIC_INFORMATION struct {
	TransactionId syscall.GUID
	State         uint32
	Outcome       uint32
}

type TRANSACTIONMANAGER_BASIC_INFORMATION struct {
	TmIdentity   syscall.GUID
	VirtualClock int64
}

type TRANSACTIONMANAGER_LOG_INFORMATION struct {
	LogIdentity syscall.GUID
}

type TRANSACTIONMANAGER_LOGPATH_INFORMATION struct {
	LogPathLength uint32
	LogPath       [1]uint16
}

type TRANSACTIONMANAGER_RECOVERY_INFORMATION struct {
	LastRecoveredLsn uint64
}

type TRANSACTIONMANAGER_OLDEST_INFORMATION struct {
	OldestTransactionGuid syscall.GUID
}

type TRANSACTION_PROPERTIES_INFORMATION struct {
	IsolationLevel    uint32
	IsolationFlags    uint32
	Timeout           int64
	Outcome           uint32
	DescriptionLength uint32
	Description       [1]uint16
}

type TRANSACTION_BIND_INFORMATION struct {
	TmHandle HANDLE
}

type TRANSACTION_ENLISTMENT_PAIR struct {
	EnlistmentId      syscall.GUID
	ResourceManagerId syscall.GUID
}

type TRANSACTION_ENLISTMENTS_INFORMATION struct {
	NumberOfEnlistments uint32
	EnlistmentPair      [1]TRANSACTION_ENLISTMENT_PAIR
}

type TRANSACTION_SUPERIOR_ENLISTMENT_INFORMATION struct {
	SuperiorEnlistmentPair TRANSACTION_ENLISTMENT_PAIR
}

type RESOURCEMANAGER_BASIC_INFORMATION struct {
	ResourceManagerId syscall.GUID
	DescriptionLength uint32
	Description       [1]uint16
}

type RESOURCEMANAGER_COMPLETION_INFORMATION struct {
	IoCompletionPortHandle HANDLE
	CompletionKey          uintptr
}

type ENLISTMENT_BASIC_INFORMATION struct {
	EnlistmentId      syscall.GUID
	TransactionId     syscall.GUID
	ResourceManagerId syscall.GUID
}

type ENLISTMENT_CRM_INFORMATION struct {
	CrmTransactionManagerId syscall.GUID
	CrmResourceManagerId    syscall.GUID
	CrmEnlistmentId         syscall.GUID
}

type TRANSACTION_LIST_ENTRY struct {
	UOW syscall.GUID
}

type TRANSACTION_LIST_INFORMATION struct {
	NumberOfTransactions   uint32
	TransactionInformation [1]TRANSACTION_LIST_ENTRY
}

type KTMOBJECT_CURSOR struct {
	LastQuery     syscall.GUID
	ObjectIdCount uint32
	ObjectIds     [1]syscall.GUID
}

// func types

type PUMS_SCHEDULER_ENTRY_POINT = uintptr
type PUMS_SCHEDULER_ENTRY_POINT_func = func(Reason RTL_UMS_SCHEDULER_REASON, ActivationPayload uintptr, SchedulerParam unsafe.Pointer)

type POUT_OF_PROCESS_FUNCTION_TABLE_CALLBACK = uintptr
type POUT_OF_PROCESS_FUNCTION_TABLE_CALLBACK_func = func(Process HANDLE, TableAddress unsafe.Pointer, Entries *uint32, Functions **IMAGE_RUNTIME_FUNCTION_ENTRY) uint32

type PEXCEPTION_FILTER = uintptr
type PEXCEPTION_FILTER_func = func(ExceptionPointers *EXCEPTION_POINTERS, EstablisherFrame unsafe.Pointer) int32

type PTERMINATION_HANDLER = uintptr
type PTERMINATION_HANDLER_func = func(_abnormal_termination BOOLEAN, EstablisherFrame unsafe.Pointer)

type PIMAGE_TLS_CALLBACK = uintptr
type PIMAGE_TLS_CALLBACK_func = func(DllHandle unsafe.Pointer, Reason uint32, Reserved unsafe.Pointer)

type WORKERCALLBACKFUNC = uintptr
type WORKERCALLBACKFUNC_func = func(param0 unsafe.Pointer)

type APC_CALLBACK_FUNCTION = uintptr
type APC_CALLBACK_FUNCTION_func = func(param0 uint32, param1 unsafe.Pointer, param2 unsafe.Pointer)

var (
	pUnregisterDeviceNotification uintptr
)

func UnregisterDeviceNotification(Handle unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnregisterDeviceNotification, libUser32, "UnregisterDeviceNotification")
	ret, _, err := syscall.SyscallN(addr, uintptr(Handle))
	return BOOL(ret), WIN32_ERROR(err)
}
