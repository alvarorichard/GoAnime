package win32

import (
	"syscall"
	"unsafe"
)

type (
	HIMC              = uintptr
	HIMCC             = uintptr
	HSAVEDUILANGUAGES = uintptr
)

const (
	ALL_SERVICE_TYPES                         uint32  = 0x0
	HIGHLEVEL_SERVICE_TYPES                   uint32  = 0x1
	LOWLEVEL_SERVICE_TYPES                    uint32  = 0x2
	ALL_SERVICES                              uint32  = 0x0
	ONLINE_SERVICES                           uint32  = 0x1
	OFFLINE_SERVICES                          uint32  = 0x2
	MAX_LEADBYTES                             uint32  = 0xc
	MAX_DEFAULTCHAR                           uint32  = 0x2
	HIGH_SURROGATE_START                      uint32  = 0xd800
	HIGH_SURROGATE_END                        uint32  = 0xdbff
	LOW_SURROGATE_START                       uint32  = 0xdc00
	LOW_SURROGATE_END                         uint32  = 0xdfff
	WC_COMPOSITECHECK                         uint32  = 0x200
	WC_DISCARDNS                              uint32  = 0x10
	WC_SEPCHARS                               uint32  = 0x20
	WC_DEFAULTCHAR                            uint32  = 0x40
	WC_ERR_INVALID_CHARS                      uint32  = 0x80
	WC_NO_BEST_FIT_CHARS                      uint32  = 0x400
	CT_CTYPE1                                 uint32  = 0x1
	CT_CTYPE2                                 uint32  = 0x2
	CT_CTYPE3                                 uint32  = 0x4
	C1_UPPER                                  uint32  = 0x1
	C1_LOWER                                  uint32  = 0x2
	C1_DIGIT                                  uint32  = 0x4
	C1_SPACE                                  uint32  = 0x8
	C1_PUNCT                                  uint32  = 0x10
	C1_CNTRL                                  uint32  = 0x20
	C1_BLANK                                  uint32  = 0x40
	C1_XDIGIT                                 uint32  = 0x80
	C1_ALPHA                                  uint32  = 0x100
	C1_DEFINED                                uint32  = 0x200
	C2_LEFTTORIGHT                            uint32  = 0x1
	C2_RIGHTTOLEFT                            uint32  = 0x2
	C2_EUROPENUMBER                           uint32  = 0x3
	C2_EUROPESEPARATOR                        uint32  = 0x4
	C2_EUROPETERMINATOR                       uint32  = 0x5
	C2_ARABICNUMBER                           uint32  = 0x6
	C2_COMMONSEPARATOR                        uint32  = 0x7
	C2_BLOCKSEPARATOR                         uint32  = 0x8
	C2_SEGMENTSEPARATOR                       uint32  = 0x9
	C2_WHITESPACE                             uint32  = 0xa
	C2_OTHERNEUTRAL                           uint32  = 0xb
	C2_NOTAPPLICABLE                          uint32  = 0x0
	C3_NONSPACING                             uint32  = 0x1
	C3_DIACRITIC                              uint32  = 0x2
	C3_VOWELMARK                              uint32  = 0x4
	C3_SYMBOL                                 uint32  = 0x8
	C3_KATAKANA                               uint32  = 0x10
	C3_HIRAGANA                               uint32  = 0x20
	C3_HALFWIDTH                              uint32  = 0x40
	C3_FULLWIDTH                              uint32  = 0x80
	C3_IDEOGRAPH                              uint32  = 0x100
	C3_KASHIDA                                uint32  = 0x200
	C3_LEXICAL                                uint32  = 0x400
	C3_HIGHSURROGATE                          uint32  = 0x800
	C3_LOWSURROGATE                           uint32  = 0x1000
	C3_ALPHA                                  uint32  = 0x8000
	C3_NOTAPPLICABLE                          uint32  = 0x0
	LCMAP_LOWERCASE                           uint32  = 0x100
	LCMAP_UPPERCASE                           uint32  = 0x200
	LCMAP_TITLECASE                           uint32  = 0x300
	LCMAP_SORTKEY                             uint32  = 0x400
	LCMAP_BYTEREV                             uint32  = 0x800
	LCMAP_HIRAGANA                            uint32  = 0x100000
	LCMAP_KATAKANA                            uint32  = 0x200000
	LCMAP_HALFWIDTH                           uint32  = 0x400000
	LCMAP_FULLWIDTH                           uint32  = 0x800000
	LCMAP_LINGUISTIC_CASING                   uint32  = 0x1000000
	LCMAP_SIMPLIFIED_CHINESE                  uint32  = 0x2000000
	LCMAP_TRADITIONAL_CHINESE                 uint32  = 0x4000000
	LCMAP_SORTHANDLE                          uint32  = 0x20000000
	LCMAP_HASH                                uint32  = 0x40000
	FIND_STARTSWITH                           uint32  = 0x100000
	FIND_ENDSWITH                             uint32  = 0x200000
	FIND_FROMSTART                            uint32  = 0x400000
	FIND_FROMEND                              uint32  = 0x800000
	LCID_ALTERNATE_SORTS                      uint32  = 0x4
	LOCALE_ALL                                uint32  = 0x0
	LOCALE_WINDOWS                            uint32  = 0x1
	LOCALE_SUPPLEMENTAL                       uint32  = 0x2
	LOCALE_ALTERNATE_SORTS                    uint32  = 0x4
	LOCALE_REPLACEMENT                        uint32  = 0x8
	LOCALE_NEUTRALDATA                        uint32  = 0x10
	LOCALE_SPECIFICDATA                       uint32  = 0x20
	CSTR_LESS_THAN                            uint32  = 0x1
	CSTR_EQUAL                                uint32  = 0x2
	CSTR_GREATER_THAN                         uint32  = 0x3
	CP_ACP                                    uint32  = 0x0
	CP_OEMCP                                  uint32  = 0x1
	CP_MACCP                                  uint32  = 0x2
	CP_THREAD_ACP                             uint32  = 0x3
	CP_SYMBOL                                 uint32  = 0x2a
	CP_UTF7                                   uint32  = 0xfde8
	CP_UTF8                                   uint32  = 0xfde9
	CTRY_DEFAULT                              uint32  = 0x0
	CTRY_ALBANIA                              uint32  = 0x163
	CTRY_ALGERIA                              uint32  = 0xd5
	CTRY_ARGENTINA                            uint32  = 0x36
	CTRY_ARMENIA                              uint32  = 0x176
	CTRY_AUSTRALIA                            uint32  = 0x3d
	CTRY_AUSTRIA                              uint32  = 0x2b
	CTRY_AZERBAIJAN                           uint32  = 0x3e2
	CTRY_BAHRAIN                              uint32  = 0x3cd
	CTRY_BELARUS                              uint32  = 0x177
	CTRY_BELGIUM                              uint32  = 0x20
	CTRY_BELIZE                               uint32  = 0x1f5
	CTRY_BOLIVIA                              uint32  = 0x24f
	CTRY_BRAZIL                               uint32  = 0x37
	CTRY_BRUNEI_DARUSSALAM                    uint32  = 0x2a1
	CTRY_BULGARIA                             uint32  = 0x167
	CTRY_CANADA                               uint32  = 0x2
	CTRY_CARIBBEAN                            uint32  = 0x1
	CTRY_CHILE                                uint32  = 0x38
	CTRY_COLOMBIA                             uint32  = 0x39
	CTRY_COSTA_RICA                           uint32  = 0x1fa
	CTRY_CROATIA                              uint32  = 0x181
	CTRY_CZECH                                uint32  = 0x1a4
	CTRY_DENMARK                              uint32  = 0x2d
	CTRY_DOMINICAN_REPUBLIC                   uint32  = 0x1
	CTRY_ECUADOR                              uint32  = 0x251
	CTRY_EGYPT                                uint32  = 0x14
	CTRY_EL_SALVADOR                          uint32  = 0x1f7
	CTRY_ESTONIA                              uint32  = 0x174
	CTRY_FAEROE_ISLANDS                       uint32  = 0x12a
	CTRY_FINLAND                              uint32  = 0x166
	CTRY_FRANCE                               uint32  = 0x21
	CTRY_GEORGIA                              uint32  = 0x3e3
	CTRY_GERMANY                              uint32  = 0x31
	CTRY_GREECE                               uint32  = 0x1e
	CTRY_GUATEMALA                            uint32  = 0x1f6
	CTRY_HONDURAS                             uint32  = 0x1f8
	CTRY_HONG_KONG                            uint32  = 0x354
	CTRY_HUNGARY                              uint32  = 0x24
	CTRY_ICELAND                              uint32  = 0x162
	CTRY_INDIA                                uint32  = 0x5b
	CTRY_INDONESIA                            uint32  = 0x3e
	CTRY_IRAN                                 uint32  = 0x3d5
	CTRY_IRAQ                                 uint32  = 0x3c4
	CTRY_IRELAND                              uint32  = 0x161
	CTRY_ISRAEL                               uint32  = 0x3cc
	CTRY_ITALY                                uint32  = 0x27
	CTRY_JAMAICA                              uint32  = 0x1
	CTRY_JAPAN                                uint32  = 0x51
	CTRY_JORDAN                               uint32  = 0x3c2
	CTRY_KAZAKSTAN                            uint32  = 0x7
	CTRY_KENYA                                uint32  = 0xfe
	CTRY_KUWAIT                               uint32  = 0x3c5
	CTRY_KYRGYZSTAN                           uint32  = 0x3e4
	CTRY_LATVIA                               uint32  = 0x173
	CTRY_LEBANON                              uint32  = 0x3c1
	CTRY_LIBYA                                uint32  = 0xda
	CTRY_LIECHTENSTEIN                        uint32  = 0x29
	CTRY_LITHUANIA                            uint32  = 0x172
	CTRY_LUXEMBOURG                           uint32  = 0x160
	CTRY_MACAU                                uint32  = 0x355
	CTRY_MACEDONIA                            uint32  = 0x185
	CTRY_MALAYSIA                             uint32  = 0x3c
	CTRY_MALDIVES                             uint32  = 0x3c0
	CTRY_MEXICO                               uint32  = 0x34
	CTRY_MONACO                               uint32  = 0x21
	CTRY_MONGOLIA                             uint32  = 0x3d0
	CTRY_MOROCCO                              uint32  = 0xd4
	CTRY_NETHERLANDS                          uint32  = 0x1f
	CTRY_NEW_ZEALAND                          uint32  = 0x40
	CTRY_NICARAGUA                            uint32  = 0x1f9
	CTRY_NORWAY                               uint32  = 0x2f
	CTRY_OMAN                                 uint32  = 0x3c8
	CTRY_PAKISTAN                             uint32  = 0x5c
	CTRY_PANAMA                               uint32  = 0x1fb
	CTRY_PARAGUAY                             uint32  = 0x253
	CTRY_PERU                                 uint32  = 0x33
	CTRY_PHILIPPINES                          uint32  = 0x3f
	CTRY_POLAND                               uint32  = 0x30
	CTRY_PORTUGAL                             uint32  = 0x15f
	CTRY_PRCHINA                              uint32  = 0x56
	CTRY_PUERTO_RICO                          uint32  = 0x1
	CTRY_QATAR                                uint32  = 0x3ce
	CTRY_ROMANIA                              uint32  = 0x28
	CTRY_RUSSIA                               uint32  = 0x7
	CTRY_SAUDI_ARABIA                         uint32  = 0x3c6
	CTRY_SERBIA                               uint32  = 0x17d
	CTRY_SINGAPORE                            uint32  = 0x41
	CTRY_SLOVAK                               uint32  = 0x1a5
	CTRY_SLOVENIA                             uint32  = 0x182
	CTRY_SOUTH_AFRICA                         uint32  = 0x1b
	CTRY_SOUTH_KOREA                          uint32  = 0x52
	CTRY_SPAIN                                uint32  = 0x22
	CTRY_SWEDEN                               uint32  = 0x2e
	CTRY_SWITZERLAND                          uint32  = 0x29
	CTRY_SYRIA                                uint32  = 0x3c3
	CTRY_TAIWAN                               uint32  = 0x376
	CTRY_TATARSTAN                            uint32  = 0x7
	CTRY_THAILAND                             uint32  = 0x42
	CTRY_TRINIDAD_Y_TOBAGO                    uint32  = 0x1
	CTRY_TUNISIA                              uint32  = 0xd8
	CTRY_TURKEY                               uint32  = 0x5a
	CTRY_UAE                                  uint32  = 0x3cb
	CTRY_UKRAINE                              uint32  = 0x17c
	CTRY_UNITED_KINGDOM                       uint32  = 0x2c
	CTRY_UNITED_STATES                        uint32  = 0x1
	CTRY_URUGUAY                              uint32  = 0x256
	CTRY_UZBEKISTAN                           uint32  = 0x7
	CTRY_VENEZUELA                            uint32  = 0x3a
	CTRY_VIET_NAM                             uint32  = 0x54
	CTRY_YEMEN                                uint32  = 0x3c7
	CTRY_ZIMBABWE                             uint32  = 0x107
	LOCALE_NOUSEROVERRIDE                     uint32  = 0x80000000
	LOCALE_USE_CP_ACP                         uint32  = 0x40000000
	LOCALE_RETURN_NUMBER                      uint32  = 0x20000000
	LOCALE_RETURN_GENITIVE_NAMES              uint32  = 0x10000000
	LOCALE_ALLOW_NEUTRAL_NAMES                uint32  = 0x8000000
	LOCALE_SLOCALIZEDDISPLAYNAME              uint32  = 0x2
	LOCALE_SENGLISHDISPLAYNAME                uint32  = 0x72
	LOCALE_SNATIVEDISPLAYNAME                 uint32  = 0x73
	LOCALE_SLOCALIZEDLANGUAGENAME             uint32  = 0x6f
	LOCALE_SENGLISHLANGUAGENAME               uint32  = 0x1001
	LOCALE_SNATIVELANGUAGENAME                uint32  = 0x4
	LOCALE_SLOCALIZEDCOUNTRYNAME              uint32  = 0x6
	LOCALE_SENGLISHCOUNTRYNAME                uint32  = 0x1002
	LOCALE_SNATIVECOUNTRYNAME                 uint32  = 0x8
	LOCALE_IDIALINGCODE                       uint32  = 0x5
	LOCALE_SLIST                              uint32  = 0xc
	LOCALE_IMEASURE                           uint32  = 0xd
	LOCALE_SDECIMAL                           uint32  = 0xe
	LOCALE_STHOUSAND                          uint32  = 0xf
	LOCALE_SGROUPING                          uint32  = 0x10
	LOCALE_IDIGITS                            uint32  = 0x11
	LOCALE_ILZERO                             uint32  = 0x12
	LOCALE_INEGNUMBER                         uint32  = 0x1010
	LOCALE_SNATIVEDIGITS                      uint32  = 0x13
	LOCALE_SCURRENCY                          uint32  = 0x14
	LOCALE_SINTLSYMBOL                        uint32  = 0x15
	LOCALE_SMONDECIMALSEP                     uint32  = 0x16
	LOCALE_SMONTHOUSANDSEP                    uint32  = 0x17
	LOCALE_SMONGROUPING                       uint32  = 0x18
	LOCALE_ICURRDIGITS                        uint32  = 0x19
	LOCALE_ICURRENCY                          uint32  = 0x1b
	LOCALE_INEGCURR                           uint32  = 0x1c
	LOCALE_SSHORTDATE                         uint32  = 0x1f
	LOCALE_SLONGDATE                          uint32  = 0x20
	LOCALE_STIMEFORMAT                        uint32  = 0x1003
	LOCALE_SAM                                uint32  = 0x28
	LOCALE_SPM                                uint32  = 0x29
	LOCALE_ICALENDARTYPE                      uint32  = 0x1009
	LOCALE_IOPTIONALCALENDAR                  uint32  = 0x100b
	LOCALE_IFIRSTDAYOFWEEK                    uint32  = 0x100c
	LOCALE_IFIRSTWEEKOFYEAR                   uint32  = 0x100d
	LOCALE_SDAYNAME1                          uint32  = 0x2a
	LOCALE_SDAYNAME2                          uint32  = 0x2b
	LOCALE_SDAYNAME3                          uint32  = 0x2c
	LOCALE_SDAYNAME4                          uint32  = 0x2d
	LOCALE_SDAYNAME5                          uint32  = 0x2e
	LOCALE_SDAYNAME6                          uint32  = 0x2f
	LOCALE_SDAYNAME7                          uint32  = 0x30
	LOCALE_SABBREVDAYNAME1                    uint32  = 0x31
	LOCALE_SABBREVDAYNAME2                    uint32  = 0x32
	LOCALE_SABBREVDAYNAME3                    uint32  = 0x33
	LOCALE_SABBREVDAYNAME4                    uint32  = 0x34
	LOCALE_SABBREVDAYNAME5                    uint32  = 0x35
	LOCALE_SABBREVDAYNAME6                    uint32  = 0x36
	LOCALE_SABBREVDAYNAME7                    uint32  = 0x37
	LOCALE_SMONTHNAME1                        uint32  = 0x38
	LOCALE_SMONTHNAME2                        uint32  = 0x39
	LOCALE_SMONTHNAME3                        uint32  = 0x3a
	LOCALE_SMONTHNAME4                        uint32  = 0x3b
	LOCALE_SMONTHNAME5                        uint32  = 0x3c
	LOCALE_SMONTHNAME6                        uint32  = 0x3d
	LOCALE_SMONTHNAME7                        uint32  = 0x3e
	LOCALE_SMONTHNAME8                        uint32  = 0x3f
	LOCALE_SMONTHNAME9                        uint32  = 0x40
	LOCALE_SMONTHNAME10                       uint32  = 0x41
	LOCALE_SMONTHNAME11                       uint32  = 0x42
	LOCALE_SMONTHNAME12                       uint32  = 0x43
	LOCALE_SMONTHNAME13                       uint32  = 0x100e
	LOCALE_SABBREVMONTHNAME1                  uint32  = 0x44
	LOCALE_SABBREVMONTHNAME2                  uint32  = 0x45
	LOCALE_SABBREVMONTHNAME3                  uint32  = 0x46
	LOCALE_SABBREVMONTHNAME4                  uint32  = 0x47
	LOCALE_SABBREVMONTHNAME5                  uint32  = 0x48
	LOCALE_SABBREVMONTHNAME6                  uint32  = 0x49
	LOCALE_SABBREVMONTHNAME7                  uint32  = 0x4a
	LOCALE_SABBREVMONTHNAME8                  uint32  = 0x4b
	LOCALE_SABBREVMONTHNAME9                  uint32  = 0x4c
	LOCALE_SABBREVMONTHNAME10                 uint32  = 0x4d
	LOCALE_SABBREVMONTHNAME11                 uint32  = 0x4e
	LOCALE_SABBREVMONTHNAME12                 uint32  = 0x4f
	LOCALE_SABBREVMONTHNAME13                 uint32  = 0x100f
	LOCALE_SPOSITIVESIGN                      uint32  = 0x50
	LOCALE_SNEGATIVESIGN                      uint32  = 0x51
	LOCALE_IPOSSIGNPOSN                       uint32  = 0x52
	LOCALE_INEGSIGNPOSN                       uint32  = 0x53
	LOCALE_IPOSSYMPRECEDES                    uint32  = 0x54
	LOCALE_IPOSSEPBYSPACE                     uint32  = 0x55
	LOCALE_INEGSYMPRECEDES                    uint32  = 0x56
	LOCALE_INEGSEPBYSPACE                     uint32  = 0x57
	LOCALE_FONTSIGNATURE                      uint32  = 0x58
	LOCALE_SISO639LANGNAME                    uint32  = 0x59
	LOCALE_SISO3166CTRYNAME                   uint32  = 0x5a
	LOCALE_IPAPERSIZE                         uint32  = 0x100a
	LOCALE_SENGCURRNAME                       uint32  = 0x1007
	LOCALE_SNATIVECURRNAME                    uint32  = 0x1008
	LOCALE_SYEARMONTH                         uint32  = 0x1006
	LOCALE_SSORTNAME                          uint32  = 0x1013
	LOCALE_IDIGITSUBSTITUTION                 uint32  = 0x1014
	LOCALE_SNAME                              uint32  = 0x5c
	LOCALE_SDURATION                          uint32  = 0x5d
	LOCALE_SSHORTESTDAYNAME1                  uint32  = 0x60
	LOCALE_SSHORTESTDAYNAME2                  uint32  = 0x61
	LOCALE_SSHORTESTDAYNAME3                  uint32  = 0x62
	LOCALE_SSHORTESTDAYNAME4                  uint32  = 0x63
	LOCALE_SSHORTESTDAYNAME5                  uint32  = 0x64
	LOCALE_SSHORTESTDAYNAME6                  uint32  = 0x65
	LOCALE_SSHORTESTDAYNAME7                  uint32  = 0x66
	LOCALE_SISO639LANGNAME2                   uint32  = 0x67
	LOCALE_SISO3166CTRYNAME2                  uint32  = 0x68
	LOCALE_SNAN                               uint32  = 0x69
	LOCALE_SPOSINFINITY                       uint32  = 0x6a
	LOCALE_SNEGINFINITY                       uint32  = 0x6b
	LOCALE_SSCRIPTS                           uint32  = 0x6c
	LOCALE_SPARENT                            uint32  = 0x6d
	LOCALE_SCONSOLEFALLBACKNAME               uint32  = 0x6e
	LOCALE_IREADINGLAYOUT                     uint32  = 0x70
	LOCALE_INEUTRAL                           uint32  = 0x71
	LOCALE_INEGATIVEPERCENT                   uint32  = 0x74
	LOCALE_IPOSITIVEPERCENT                   uint32  = 0x75
	LOCALE_SPERCENT                           uint32  = 0x76
	LOCALE_SPERMILLE                          uint32  = 0x77
	LOCALE_SMONTHDAY                          uint32  = 0x78
	LOCALE_SSHORTTIME                         uint32  = 0x79
	LOCALE_SOPENTYPELANGUAGETAG               uint32  = 0x7a
	LOCALE_SSORTLOCALE                        uint32  = 0x7b
	LOCALE_SRELATIVELONGDATE                  uint32  = 0x7c
	LOCALE_ICONSTRUCTEDLOCALE                 uint32  = 0x7d
	LOCALE_SSHORTESTAM                        uint32  = 0x7e
	LOCALE_SSHORTESTPM                        uint32  = 0x7f
	LOCALE_IUSEUTF8LEGACYACP                  uint32  = 0x666
	LOCALE_IUSEUTF8LEGACYOEMCP                uint32  = 0x999
	LOCALE_IDEFAULTCODEPAGE                   uint32  = 0xb
	LOCALE_IDEFAULTANSICODEPAGE               uint32  = 0x1004
	LOCALE_IDEFAULTMACCODEPAGE                uint32  = 0x1011
	LOCALE_IDEFAULTEBCDICCODEPAGE             uint32  = 0x1012
	LOCALE_ILANGUAGE                          uint32  = 0x1
	LOCALE_SABBREVLANGNAME                    uint32  = 0x3
	LOCALE_SABBREVCTRYNAME                    uint32  = 0x7
	LOCALE_IGEOID                             uint32  = 0x5b
	LOCALE_IDEFAULTLANGUAGE                   uint32  = 0x9
	LOCALE_IDEFAULTCOUNTRY                    uint32  = 0xa
	LOCALE_IINTLCURRDIGITS                    uint32  = 0x1a
	LOCALE_SDATE                              uint32  = 0x1d
	LOCALE_STIME                              uint32  = 0x1e
	LOCALE_IDATE                              uint32  = 0x21
	LOCALE_ILDATE                             uint32  = 0x22
	LOCALE_ITIME                              uint32  = 0x23
	LOCALE_ITIMEMARKPOSN                      uint32  = 0x1005
	LOCALE_ICENTURY                           uint32  = 0x24
	LOCALE_ITLZERO                            uint32  = 0x25
	LOCALE_IDAYLZERO                          uint32  = 0x26
	LOCALE_IMONLZERO                          uint32  = 0x27
	LOCALE_SKEYBOARDSTOINSTALL                uint32  = 0x5e
	LOCALE_SLANGUAGE                          uint32  = 0x2
	LOCALE_SLANGDISPLAYNAME                   uint32  = 0x6f
	LOCALE_SENGLANGUAGE                       uint32  = 0x1001
	LOCALE_SNATIVELANGNAME                    uint32  = 0x4
	LOCALE_SCOUNTRY                           uint32  = 0x6
	LOCALE_SENGCOUNTRY                        uint32  = 0x1002
	LOCALE_SNATIVECTRYNAME                    uint32  = 0x8
	LOCALE_ICOUNTRY                           uint32  = 0x5
	LOCALE_S1159                              uint32  = 0x28
	LOCALE_S2359                              uint32  = 0x29
	CAL_NOUSEROVERRIDE                        uint32  = 0x80000000
	CAL_USE_CP_ACP                            uint32  = 0x40000000
	CAL_RETURN_NUMBER                         uint32  = 0x20000000
	CAL_RETURN_GENITIVE_NAMES                 uint32  = 0x10000000
	CAL_ICALINTVALUE                          uint32  = 0x1
	CAL_SCALNAME                              uint32  = 0x2
	CAL_IYEAROFFSETRANGE                      uint32  = 0x3
	CAL_SERASTRING                            uint32  = 0x4
	CAL_SSHORTDATE                            uint32  = 0x5
	CAL_SLONGDATE                             uint32  = 0x6
	CAL_SDAYNAME1                             uint32  = 0x7
	CAL_SDAYNAME2                             uint32  = 0x8
	CAL_SDAYNAME3                             uint32  = 0x9
	CAL_SDAYNAME4                             uint32  = 0xa
	CAL_SDAYNAME5                             uint32  = 0xb
	CAL_SDAYNAME6                             uint32  = 0xc
	CAL_SDAYNAME7                             uint32  = 0xd
	CAL_SABBREVDAYNAME1                       uint32  = 0xe
	CAL_SABBREVDAYNAME2                       uint32  = 0xf
	CAL_SABBREVDAYNAME3                       uint32  = 0x10
	CAL_SABBREVDAYNAME4                       uint32  = 0x11
	CAL_SABBREVDAYNAME5                       uint32  = 0x12
	CAL_SABBREVDAYNAME6                       uint32  = 0x13
	CAL_SABBREVDAYNAME7                       uint32  = 0x14
	CAL_SMONTHNAME1                           uint32  = 0x15
	CAL_SMONTHNAME2                           uint32  = 0x16
	CAL_SMONTHNAME3                           uint32  = 0x17
	CAL_SMONTHNAME4                           uint32  = 0x18
	CAL_SMONTHNAME5                           uint32  = 0x19
	CAL_SMONTHNAME6                           uint32  = 0x1a
	CAL_SMONTHNAME7                           uint32  = 0x1b
	CAL_SMONTHNAME8                           uint32  = 0x1c
	CAL_SMONTHNAME9                           uint32  = 0x1d
	CAL_SMONTHNAME10                          uint32  = 0x1e
	CAL_SMONTHNAME11                          uint32  = 0x1f
	CAL_SMONTHNAME12                          uint32  = 0x20
	CAL_SMONTHNAME13                          uint32  = 0x21
	CAL_SABBREVMONTHNAME1                     uint32  = 0x22
	CAL_SABBREVMONTHNAME2                     uint32  = 0x23
	CAL_SABBREVMONTHNAME3                     uint32  = 0x24
	CAL_SABBREVMONTHNAME4                     uint32  = 0x25
	CAL_SABBREVMONTHNAME5                     uint32  = 0x26
	CAL_SABBREVMONTHNAME6                     uint32  = 0x27
	CAL_SABBREVMONTHNAME7                     uint32  = 0x28
	CAL_SABBREVMONTHNAME8                     uint32  = 0x29
	CAL_SABBREVMONTHNAME9                     uint32  = 0x2a
	CAL_SABBREVMONTHNAME10                    uint32  = 0x2b
	CAL_SABBREVMONTHNAME11                    uint32  = 0x2c
	CAL_SABBREVMONTHNAME12                    uint32  = 0x2d
	CAL_SABBREVMONTHNAME13                    uint32  = 0x2e
	CAL_SYEARMONTH                            uint32  = 0x2f
	CAL_ITWODIGITYEARMAX                      uint32  = 0x30
	CAL_SSHORTESTDAYNAME1                     uint32  = 0x31
	CAL_SSHORTESTDAYNAME2                     uint32  = 0x32
	CAL_SSHORTESTDAYNAME3                     uint32  = 0x33
	CAL_SSHORTESTDAYNAME4                     uint32  = 0x34
	CAL_SSHORTESTDAYNAME5                     uint32  = 0x35
	CAL_SSHORTESTDAYNAME6                     uint32  = 0x36
	CAL_SSHORTESTDAYNAME7                     uint32  = 0x37
	CAL_SMONTHDAY                             uint32  = 0x38
	CAL_SABBREVERASTRING                      uint32  = 0x39
	CAL_SRELATIVELONGDATE                     uint32  = 0x3a
	CAL_SENGLISHERANAME                       uint32  = 0x3b
	CAL_SENGLISHABBREVERANAME                 uint32  = 0x3c
	CAL_SJAPANESEERAFIRSTYEAR                 uint32  = 0x3d
	ENUM_ALL_CALENDARS                        uint32  = 0xffffffff
	CAL_GREGORIAN                             uint32  = 0x1
	CAL_GREGORIAN_US                          uint32  = 0x2
	CAL_JAPAN                                 uint32  = 0x3
	CAL_TAIWAN                                uint32  = 0x4
	CAL_KOREA                                 uint32  = 0x5
	CAL_HIJRI                                 uint32  = 0x6
	CAL_THAI                                  uint32  = 0x7
	CAL_HEBREW                                uint32  = 0x8
	CAL_GREGORIAN_ME_FRENCH                   uint32  = 0x9
	CAL_GREGORIAN_ARABIC                      uint32  = 0xa
	CAL_GREGORIAN_XLIT_ENGLISH                uint32  = 0xb
	CAL_GREGORIAN_XLIT_FRENCH                 uint32  = 0xc
	CAL_PERSIAN                               uint32  = 0x16
	CAL_UMALQURA                              uint32  = 0x17
	LGRPID_WESTERN_EUROPE                     uint32  = 0x1
	LGRPID_CENTRAL_EUROPE                     uint32  = 0x2
	LGRPID_BALTIC                             uint32  = 0x3
	LGRPID_GREEK                              uint32  = 0x4
	LGRPID_CYRILLIC                           uint32  = 0x5
	LGRPID_TURKIC                             uint32  = 0x6
	LGRPID_TURKISH                            uint32  = 0x6
	LGRPID_JAPANESE                           uint32  = 0x7
	LGRPID_KOREAN                             uint32  = 0x8
	LGRPID_TRADITIONAL_CHINESE                uint32  = 0x9
	LGRPID_SIMPLIFIED_CHINESE                 uint32  = 0xa
	LGRPID_THAI                               uint32  = 0xb
	LGRPID_HEBREW                             uint32  = 0xc
	LGRPID_ARABIC                             uint32  = 0xd
	LGRPID_VIETNAMESE                         uint32  = 0xe
	LGRPID_INDIC                              uint32  = 0xf
	LGRPID_GEORGIAN                           uint32  = 0x10
	LGRPID_ARMENIAN                           uint32  = 0x11
	MUI_LANGUAGE_ID                           uint32  = 0x4
	MUI_LANGUAGE_NAME                         uint32  = 0x8
	MUI_MERGE_SYSTEM_FALLBACK                 uint32  = 0x10
	MUI_MERGE_USER_FALLBACK                   uint32  = 0x20
	MUI_THREAD_LANGUAGES                      uint32  = 0x40
	MUI_CONSOLE_FILTER                        uint32  = 0x100
	MUI_COMPLEX_SCRIPT_FILTER                 uint32  = 0x200
	MUI_RESET_FILTERS                         uint32  = 0x1
	MUI_USER_PREFERRED_UI_LANGUAGES           uint32  = 0x10
	MUI_USE_INSTALLED_LANGUAGES               uint32  = 0x20
	MUI_USE_SEARCH_ALL_LANGUAGES              uint32  = 0x40
	MUI_LANG_NEUTRAL_PE_FILE                  uint32  = 0x100
	MUI_NON_LANG_NEUTRAL_FILE                 uint32  = 0x200
	MUI_MACHINE_LANGUAGE_SETTINGS             uint32  = 0x400
	MUI_FILETYPE_NOT_LANGUAGE_NEUTRAL         uint32  = 0x1
	MUI_FILETYPE_LANGUAGE_NEUTRAL_MAIN        uint32  = 0x2
	MUI_FILETYPE_LANGUAGE_NEUTRAL_MUI         uint32  = 0x4
	MUI_QUERY_TYPE                            uint32  = 0x1
	MUI_QUERY_CHECKSUM                        uint32  = 0x2
	MUI_QUERY_LANGUAGE_NAME                   uint32  = 0x4
	MUI_QUERY_RESOURCE_TYPES                  uint32  = 0x8
	MUI_FILEINFO_VERSION                      uint32  = 0x1
	MUI_FULL_LANGUAGE                         uint32  = 0x1
	MUI_PARTIAL_LANGUAGE                      uint32  = 0x2
	MUI_LIP_LANGUAGE                          uint32  = 0x4
	MUI_LANGUAGE_INSTALLED                    uint32  = 0x20
	MUI_LANGUAGE_LICENSED                     uint32  = 0x40
	GEOID_NOT_AVAILABLE                       int32   = -1
	SORTING_PARADIGM_NLS                      uint32  = 0x0
	SORTING_PARADIGM_ICU                      uint32  = 0x1000000
	IDN_ALLOW_UNASSIGNED                      uint32  = 0x1
	IDN_USE_STD3_ASCII_RULES                  uint32  = 0x2
	IDN_EMAIL_ADDRESS                         uint32  = 0x4
	IDN_RAW_PUNYCODE                          uint32  = 0x8
	VS_ALLOW_LATIN                            uint32  = 0x1
	GSS_ALLOW_INHERITED_COMMON                uint32  = 0x1
	MUI_FORMAT_REG_COMPAT                     uint32  = 0x1
	MUI_FORMAT_INF_COMPAT                     uint32  = 0x2
	MUI_VERIFY_FILE_EXISTS                    uint32  = 0x4
	MUI_SKIP_STRING_CACHE                     uint32  = 0x8
	MUI_IMMUTABLE_LOOKUP                      uint32  = 0x10
	LOCALE_NAME_INVARIANT                     string  = ""
	LOCALE_NAME_SYSTEM_DEFAULT                string  = "!x-sys-default-locale"
	MIN_SPELLING_NTDDI                        uint32  = 0x6020000
	SCRIPT_UNDEFINED                          uint32  = 0x0
	USP_E_SCRIPT_NOT_IN_FONT                  HRESULT = -2147220992
	SGCM_RTL                                  uint32  = 0x1
	SSA_PASSWORD                              uint32  = 0x1
	SSA_TAB                                   uint32  = 0x2
	SSA_CLIP                                  uint32  = 0x4
	SSA_FIT                                   uint32  = 0x8
	SSA_DZWG                                  uint32  = 0x10
	SSA_FALLBACK                              uint32  = 0x20
	SSA_BREAK                                 uint32  = 0x40
	SSA_GLYPHS                                uint32  = 0x80
	SSA_RTL                                   uint32  = 0x100
	SSA_GCP                                   uint32  = 0x200
	SSA_HOTKEY                                uint32  = 0x400
	SSA_METAFILE                              uint32  = 0x800
	SSA_LINK                                  uint32  = 0x1000
	SSA_HIDEHOTKEY                            uint32  = 0x2000
	SSA_HOTKEYONLY                            uint32  = 0x2400
	SSA_FULLMEASURE                           uint32  = 0x4000000
	SSA_LPKANSIFALLBACK                       uint32  = 0x8000000
	SSA_PIDX                                  uint32  = 0x10000000
	SSA_LAYOUTRTL                             uint32  = 0x20000000
	SSA_DONTGLYPH                             uint32  = 0x40000000
	SSA_NOKASHIDA                             uint32  = 0x80000000
	SCRIPT_DIGITSUBSTITUTE_CONTEXT            uint32  = 0x0
	SCRIPT_DIGITSUBSTITUTE_NONE               uint32  = 0x1
	SCRIPT_DIGITSUBSTITUTE_NATIONAL           uint32  = 0x2
	SCRIPT_DIGITSUBSTITUTE_TRADITIONAL        uint32  = 0x3
	UNISCRIBE_OPENTYPE                        uint32  = 0x100
	SCRIPT_TAG_UNKNOWN                        uint32  = 0x0
	MUI_LANGUAGE_EXACT                        uint32  = 0x10
	NLS_CP_CPINFO                             uint32  = 0x10000000
	NLS_CP_MBTOWC                             uint32  = 0x40000000
	NLS_CP_WCTOMB                             uint32  = 0x80000000
	U_DISABLE_RENAMING                        uint32  = 0x1
	U_SHOW_CPLUSPLUS_API                      uint32  = 0x0
	U_DEFAULT_SHOW_DRAFT                      uint32  = 0x0
	U_HIDE_DRAFT_API                          uint32  = 0x1
	U_HIDE_DEPRECATED_API                     uint32  = 0x1
	U_HIDE_OBSOLETE_API                       uint32  = 0x1
	U_HIDE_INTERNAL_API                       uint32  = 0x1
	U_NO_DEFAULT_INCLUDE_UTF_HEADERS          uint32  = 0x1
	U_DEBUG                                   uint32  = 0x1
	UCLN_NO_AUTO_CLEANUP                      uint32  = 0x1
	U_OVERRIDE_CXX_ALLOCATION                 uint32  = 0x1
	U_ENABLE_TRACING                          uint32  = 0x0
	UCONFIG_ENABLE_PLUGINS                    uint32  = 0x0
	U_ENABLE_DYLOAD                           uint32  = 0x1
	U_CHECK_DYLOAD                            uint32  = 0x1
	U_HAVE_LIB_SUFFIX                         uint32  = 0x1
	U_LIB_SUFFIX_C_NAME_STRING                string  = ""
	UCONFIG_ONLY_COLLATION                    uint32  = 0x0
	UCONFIG_NO_BREAK_ITERATION                uint32  = 0x1
	UCONFIG_NO_IDNA                           uint32  = 0x1
	UCONFIG_NO_FORMATTING                     uint32  = 0x1
	UCONFIG_NO_TRANSLITERATION                uint32  = 0x1
	UCONFIG_NO_REGULAR_EXPRESSIONS            uint32  = 0x1
	UCONFIG_NO_FILE_IO                        uint32  = 0x0
	UCONFIG_NO_CONVERSION                     uint32  = 0x0
	UCONFIG_NO_LEGACY_CONVERSION              uint32  = 0x1
	UCONFIG_ONLY_HTML_CONVERSION              uint32  = 0x0
	UCONFIG_NO_NORMALIZATION                  uint32  = 0x0
	UCONFIG_NO_COLLATION                      uint32  = 0x1
	UCONFIG_NO_SERVICE                        uint32  = 0x0
	UCONFIG_HAVE_PARSEALLINPUT                uint32  = 0x1
	UCONFIG_NO_FILTERED_BREAK_ITERATION       uint32  = 0x0
	U_PF_UNKNOWN                              uint32  = 0x0
	U_PF_WINDOWS                              uint32  = 0x3e8
	U_PF_MINGW                                uint32  = 0x708
	U_PF_CYGWIN                               uint32  = 0x76c
	U_PF_HPUX                                 uint32  = 0x834
	U_PF_SOLARIS                              uint32  = 0xa28
	U_PF_BSD                                  uint32  = 0xbb8
	U_PF_AIX                                  uint32  = 0xc1c
	U_PF_IRIX                                 uint32  = 0xc80
	U_PF_DARWIN                               uint32  = 0xdac
	U_PF_IPHONE                               uint32  = 0xdde
	U_PF_QNX                                  uint32  = 0xe74
	U_PF_LINUX                                uint32  = 0xfa0
	U_PF_BROWSER_NATIVE_CLIENT                uint32  = 0xfb4
	U_PF_ANDROID                              uint32  = 0xfd2
	U_PF_FUCHSIA                              uint32  = 0x1004
	U_PF_EMSCRIPTEN                           uint32  = 0x1392
	U_PF_OS390                                uint32  = 0x2328
	U_PF_OS400                                uint32  = 0x24b8
	U_PLATFORM                                uint32  = 0x708
	U_PLATFORM_USES_ONLY_WIN32_API            uint32  = 0x1
	U_PLATFORM_HAS_WIN32_API                  uint32  = 0x1
	U_PLATFORM_IMPLEMENTS_POSIX               uint32  = 0x0
	U_PLATFORM_IS_LINUX_BASED                 uint32  = 0x1
	U_PLATFORM_IS_DARWIN_BASED                uint32  = 0x1
	U_HAVE_STDINT_H                           uint32  = 0x1
	U_HAVE_INTTYPES_H                         uint32  = 0x1
	U_GCC_MAJOR_MINOR                         uint32  = 0x0
	U_IS_BIG_ENDIAN                           uint32  = 0x0
	U_HAVE_PLACEMENT_NEW                      uint32  = 0x0
	U_HAVE_DEBUG_LOCATION_NEW                 uint32  = 0x1
	U_CPLUSPLUS_VERSION                       uint32  = 0x0
	U_ASCII_FAMILY                            uint32  = 0x0
	U_EBCDIC_FAMILY                           uint32  = 0x1
	U_CHARSET_FAMILY                          uint32  = 0x1
	U_CHARSET_IS_UTF8                         uint32  = 0x1
	U_HAVE_WCHAR_H                            uint32  = 0x0
	U_SIZEOF_WCHAR_T                          uint32  = 0x1
	U_HAVE_WCSCPY                             uint32  = 0x0
	U_HAVE_CHAR16_T                           uint32  = 0x1
	U_DEFINE_FALSE_AND_TRUE                   uint32  = 0x1
	U_SIZEOF_UCHAR                            uint32  = 0x2
	U_CHAR16_IS_TYPEDEF                       uint32  = 0x1
	U_SENTINEL                                int32   = -1
	U8_LEAD3_T1_BITS                          string  = " 000000000000\x1000"
	U8_LEAD4_T1_BITS                          string  = ""
	U8_MAX_LENGTH                             uint32  = 0x4
	U16_MAX_LENGTH                            uint32  = 0x2
	U_HIDE_OBSOLETE_UTF_OLD_H                 uint32  = 0x0
	UTF_SIZE                                  uint32  = 0x10
	UTF8_ERROR_VALUE_1                        uint32  = 0x15
	UTF8_ERROR_VALUE_2                        uint32  = 0x9f
	UTF_ERROR_VALUE                           uint32  = 0xffff
	UTF8_MAX_CHAR_LENGTH                      uint32  = 0x4
	UTF16_MAX_CHAR_LENGTH                     uint32  = 0x2
	UTF32_MAX_CHAR_LENGTH                     uint32  = 0x1
	UTF_MAX_CHAR_LENGTH                       uint32  = 0x2
	U_COPYRIGHT_STRING_LENGTH                 uint32  = 0x80
	U_MAX_VERSION_LENGTH                      uint32  = 0x4
	U_MAX_VERSION_STRING_LENGTH               uint32  = 0x14
	U_MILLIS_PER_SECOND                       uint32  = 0x3e8
	U_MILLIS_PER_MINUTE                       uint32  = 0xea60
	U_MILLIS_PER_HOUR                         uint32  = 0x36ee80
	U_MILLIS_PER_DAY                          uint32  = 0x5265c00
	U_COMBINED_IMPLEMENTATION                 uint32  = 0x1
	U_SHAPE_LENGTH_GROW_SHRINK                uint32  = 0x0
	U_SHAPE_LAMALEF_RESIZE                    uint32  = 0x0
	U_SHAPE_LENGTH_FIXED_SPACES_NEAR          uint32  = 0x1
	U_SHAPE_LAMALEF_NEAR                      uint32  = 0x1
	U_SHAPE_LENGTH_FIXED_SPACES_AT_END        uint32  = 0x2
	U_SHAPE_LAMALEF_END                       uint32  = 0x2
	U_SHAPE_LENGTH_FIXED_SPACES_AT_BEGINNING  uint32  = 0x3
	U_SHAPE_LAMALEF_BEGIN                     uint32  = 0x3
	U_SHAPE_LAMALEF_AUTO                      uint32  = 0x10000
	U_SHAPE_LENGTH_MASK                       uint32  = 0x10003
	U_SHAPE_LAMALEF_MASK                      uint32  = 0x10003
	U_SHAPE_TEXT_DIRECTION_LOGICAL            uint32  = 0x0
	U_SHAPE_TEXT_DIRECTION_VISUAL_RTL         uint32  = 0x0
	U_SHAPE_TEXT_DIRECTION_VISUAL_LTR         uint32  = 0x4
	U_SHAPE_TEXT_DIRECTION_MASK               uint32  = 0x4
	U_SHAPE_LETTERS_NOOP                      uint32  = 0x0
	U_SHAPE_LETTERS_SHAPE                     uint32  = 0x8
	U_SHAPE_LETTERS_UNSHAPE                   uint32  = 0x10
	U_SHAPE_LETTERS_SHAPE_TASHKEEL_ISOLATED   uint32  = 0x18
	U_SHAPE_LETTERS_MASK                      uint32  = 0x18
	U_SHAPE_DIGITS_NOOP                       uint32  = 0x0
	U_SHAPE_DIGITS_EN2AN                      uint32  = 0x20
	U_SHAPE_DIGITS_AN2EN                      uint32  = 0x40
	U_SHAPE_DIGITS_ALEN2AN_INIT_LR            uint32  = 0x60
	U_SHAPE_DIGITS_ALEN2AN_INIT_AL            uint32  = 0x80
	U_SHAPE_DIGITS_RESERVED                   uint32  = 0xa0
	U_SHAPE_DIGITS_MASK                       uint32  = 0xe0
	U_SHAPE_DIGIT_TYPE_AN                     uint32  = 0x0
	U_SHAPE_DIGIT_TYPE_AN_EXTENDED            uint32  = 0x100
	U_SHAPE_DIGIT_TYPE_RESERVED               uint32  = 0x200
	U_SHAPE_DIGIT_TYPE_MASK                   uint32  = 0x300
	U_SHAPE_AGGREGATE_TASHKEEL                uint32  = 0x4000
	U_SHAPE_AGGREGATE_TASHKEEL_NOOP           uint32  = 0x0
	U_SHAPE_AGGREGATE_TASHKEEL_MASK           uint32  = 0x4000
	U_SHAPE_PRESERVE_PRESENTATION             uint32  = 0x8000
	U_SHAPE_PRESERVE_PRESENTATION_NOOP        uint32  = 0x0
	U_SHAPE_PRESERVE_PRESENTATION_MASK        uint32  = 0x8000
	U_SHAPE_SEEN_TWOCELL_NEAR                 uint32  = 0x200000
	U_SHAPE_SEEN_MASK                         uint32  = 0x700000
	U_SHAPE_YEHHAMZA_TWOCELL_NEAR             uint32  = 0x1000000
	U_SHAPE_YEHHAMZA_MASK                     uint32  = 0x3800000
	U_SHAPE_TASHKEEL_BEGIN                    uint32  = 0x40000
	U_SHAPE_TASHKEEL_END                      uint32  = 0x60000
	U_SHAPE_TASHKEEL_RESIZE                   uint32  = 0x80000
	U_SHAPE_TASHKEEL_REPLACE_BY_TATWEEL       uint32  = 0xc0000
	U_SHAPE_TASHKEEL_MASK                     uint32  = 0xe0000
	U_SHAPE_SPACES_RELATIVE_TO_TEXT_BEGIN_END uint32  = 0x4000000
	U_SHAPE_SPACES_RELATIVE_TO_TEXT_MASK      uint32  = 0x4000000
	U_SHAPE_TAIL_NEW_UNICODE                  uint32  = 0x8000000
	U_SHAPE_TAIL_TYPE_MASK                    uint32  = 0x8000000
	ULOC_CHINESE                              string  = "zh"
	ULOC_ENGLISH                              string  = "en"
	ULOC_FRENCH                               string  = "fr"
	ULOC_GERMAN                               string  = "de"
	ULOC_ITALIAN                              string  = "it"
	ULOC_JAPANESE                             string  = "ja"
	ULOC_KOREAN                               string  = "ko"
	ULOC_SIMPLIFIED_CHINESE                   string  = "zh_CN"
	ULOC_TRADITIONAL_CHINESE                  string  = "zh_TW"
	ULOC_CANADA                               string  = "en_CA"
	ULOC_CANADA_FRENCH                        string  = "fr_CA"
	ULOC_CHINA                                string  = "zh_CN"
	ULOC_PRC                                  string  = "zh_CN"
	ULOC_FRANCE                               string  = "fr_FR"
	ULOC_GERMANY                              string  = "de_DE"
	ULOC_ITALY                                string  = "it_IT"
	ULOC_JAPAN                                string  = "ja_JP"
	ULOC_KOREA                                string  = "ko_KR"
	ULOC_TAIWAN                               string  = "zh_TW"
	ULOC_UK                                   string  = "en_GB"
	ULOC_US                                   string  = "en_US"
	ULOC_LANG_CAPACITY                        uint32  = 0xc
	ULOC_COUNTRY_CAPACITY                     uint32  = 0x4
	ULOC_FULLNAME_CAPACITY                    uint32  = 0x9d
	ULOC_SCRIPT_CAPACITY                      uint32  = 0x6
	ULOC_KEYWORDS_CAPACITY                    uint32  = 0x60
	ULOC_KEYWORD_AND_VALUES_CAPACITY          uint32  = 0x64
	ULOC_KEYWORD_SEPARATOR_UNICODE            uint32  = 0x40
	ULOC_KEYWORD_ASSIGN_UNICODE               uint32  = 0x3d
	ULOC_KEYWORD_ITEM_SEPARATOR_UNICODE       uint32  = 0x3b
	UCNV_SUB_STOP_ON_ILLEGAL                  string  = "i"
	UCNV_SKIP_STOP_ON_ILLEGAL                 string  = "i"
	UCNV_ESCAPE_JAVA                          string  = "J"
	UCNV_ESCAPE_C                             string  = "C"
	UCNV_ESCAPE_XML_DEC                       string  = "D"
	UCNV_ESCAPE_XML_HEX                       string  = "X"
	UCNV_ESCAPE_UNICODE                       string  = "U"
	UCNV_ESCAPE_CSS2                          string  = "S"
	UCNV_MAX_CONVERTER_NAME_LENGTH            uint32  = 0x3c
	UCNV_SI                                   uint32  = 0xf
	UCNV_SO                                   uint32  = 0xe
	UCNV_OPTION_SEP_STRING                    string  = ","
	UCNV_VALUE_SEP_STRING                     string  = "="
	UCNV_LOCALE_OPTION_STRING                 string  = ",locale="
	UCNV_VERSION_OPTION_STRING                string  = ",version="
	UCNV_SWAP_LFNL_OPTION_STRING              string  = ",swaplfnl"
	U_FOLD_CASE_DEFAULT                       uint32  = 0x0
	U_FOLD_CASE_EXCLUDE_SPECIAL_I             uint32  = 0x1
	U_TITLECASE_WHOLE_STRING                  uint32  = 0x20
	U_TITLECASE_SENTENCES                     uint32  = 0x40
	U_TITLECASE_NO_LOWERCASE                  uint32  = 0x100
	U_TITLECASE_NO_BREAK_ADJUSTMENT           uint32  = 0x200
	U_TITLECASE_ADJUST_TO_CASED               uint32  = 0x400
	U_EDITS_NO_RESET                          uint32  = 0x2000
	U_OMIT_UNCHANGED_TEXT                     uint32  = 0x4000
	U_COMPARE_CODE_POINT_ORDER                uint32  = 0x8000
	U_COMPARE_IGNORE_CASE                     uint32  = 0x10000
	UNORM_INPUT_IS_FCD                        uint32  = 0x20000
	UCHAR_MIN_VALUE                           uint32  = 0x0
	UCHAR_MAX_VALUE                           uint32  = 0x10ffff
	UBIDI_DEFAULT_LTR                         uint32  = 0xfe
	UBIDI_DEFAULT_RTL                         uint32  = 0xff
	UBIDI_MAX_EXPLICIT_LEVEL                  uint32  = 0x7d
	UBIDI_LEVEL_OVERRIDE                      uint32  = 0x80
	UBIDI_MAP_NOWHERE                         int32   = -1
	UBIDI_KEEP_BASE_COMBINING                 uint32  = 0x1
	UBIDI_DO_MIRRORING                        uint32  = 0x2
	UBIDI_INSERT_LRM_FOR_NUMERIC              uint32  = 0x4
	UBIDI_REMOVE_BIDI_CONTROLS                uint32  = 0x8
	UBIDI_OUTPUT_REVERSE                      uint32  = 0x10
	USPREP_DEFAULT                            uint32  = 0x0
	USPREP_ALLOW_UNASSIGNED                   uint32  = 0x1
	U_ICU_VERSION_BUNDLE                      string  = "icuver"
	U_ICU_DATA_KEY                            string  = "DataVersion"
	UCAL_UNKNOWN_ZONE_ID                      string  = "Etc/Unknown"
	UDAT_YEAR                                 string  = "y"
	UDAT_QUARTER                              string  = "QQQQ"
	UDAT_ABBR_QUARTER                         string  = "QQQ"
	UDAT_YEAR_QUARTER                         string  = "yQQQQ"
	UDAT_YEAR_ABBR_QUARTER                    string  = "yQQQ"
	UDAT_MONTH                                string  = "MMMM"
	UDAT_ABBR_MONTH                           string  = "MMM"
	UDAT_NUM_MONTH                            string  = "M"
	UDAT_YEAR_MONTH                           string  = "yMMMM"
	UDAT_YEAR_ABBR_MONTH                      string  = "yMMM"
	UDAT_YEAR_NUM_MONTH                       string  = "yM"
	UDAT_DAY                                  string  = "d"
	UDAT_YEAR_MONTH_DAY                       string  = "yMMMMd"
	UDAT_YEAR_ABBR_MONTH_DAY                  string  = "yMMMd"
	UDAT_YEAR_NUM_MONTH_DAY                   string  = "yMd"
	UDAT_WEEKDAY                              string  = "EEEE"
	UDAT_ABBR_WEEKDAY                         string  = "E"
	UDAT_YEAR_MONTH_WEEKDAY_DAY               string  = "yMMMMEEEEd"
	UDAT_YEAR_ABBR_MONTH_WEEKDAY_DAY          string  = "yMMMEd"
	UDAT_YEAR_NUM_MONTH_WEEKDAY_DAY           string  = "yMEd"
	UDAT_MONTH_DAY                            string  = "MMMMd"
	UDAT_ABBR_MONTH_DAY                       string  = "MMMd"
	UDAT_NUM_MONTH_DAY                        string  = "Md"
	UDAT_MONTH_WEEKDAY_DAY                    string  = "MMMMEEEEd"
	UDAT_ABBR_MONTH_WEEKDAY_DAY               string  = "MMMEd"
	UDAT_NUM_MONTH_WEEKDAY_DAY                string  = "MEd"
	UDAT_HOUR                                 string  = "j"
	UDAT_HOUR24                               string  = "H"
	UDAT_MINUTE                               string  = "m"
	UDAT_HOUR_MINUTE                          string  = "jm"
	UDAT_HOUR24_MINUTE                        string  = "Hm"
	UDAT_SECOND                               string  = "s"
	UDAT_HOUR_MINUTE_SECOND                   string  = "jms"
	UDAT_HOUR24_MINUTE_SECOND                 string  = "Hms"
	UDAT_MINUTE_SECOND                        string  = "ms"
	UDAT_LOCATION_TZ                          string  = "VVVV"
	UDAT_GENERIC_TZ                           string  = "vvvv"
	UDAT_ABBR_GENERIC_TZ                      string  = "v"
	UDAT_SPECIFIC_TZ                          string  = "zzzz"
	UDAT_ABBR_SPECIFIC_TZ                     string  = "z"
	UDAT_ABBR_UTC_TZ                          string  = "ZZZZ"
	USEARCH_DONE                              int32   = -1
	U_HAVE_STD_STRING                         uint32  = 0x0
	UCONFIG_FORMAT_FASTPATHS_49               uint32  = 0x1
	U_PLATFORM_HAS_WINUWP_API                 uint32  = 0x0
	U_IOSTREAM_SOURCE                         uint32  = 0x30c1f
	U_HAVE_RVALUE_REFERENCES                  uint32  = 0x1
	U_USING_ICU_NAMESPACE                     uint32  = 0x1
	U_ICUDATA_TYPE_LETTER                     string  = "e"
	U_UNICODE_VERSION                         string  = "8.0"
	CANITER_SKIP_ZEROES                       uint32  = 0x1
	NUMSYS_NAME_CAPACITY                      uint32  = 0x8
	U_HAVE_RBNF                               uint32  = 0x0
	IME_CMODE_ALPHANUMERIC                    uint32  = 0x0
	IME_CMODE_NATIVE                          uint32  = 0x1
	IME_CMODE_CHINESE                         uint32  = 0x1
	IME_CMODE_HANGUL                          uint32  = 0x1
	IME_CMODE_JAPANESE                        uint32  = 0x1
	IME_CMODE_KATAKANA                        uint32  = 0x2
	IME_CMODE_LANGUAGE                        uint32  = 0x3
	IME_CMODE_FULLSHAPE                       uint32  = 0x8
	IME_CMODE_ROMAN                           uint32  = 0x10
	IME_CMODE_CHARCODE                        uint32  = 0x20
	IME_CMODE_HANJACONVERT                    uint32  = 0x40
	IME_CMODE_NATIVESYMBOL                    uint32  = 0x80
	MAX_MIMECP_NAME                           uint32  = 0x40
	MAX_MIMECSET_NAME                         uint32  = 0x32
	MAX_MIMEFACE_NAME                         uint32  = 0x20
	MAX_RFC1766_NAME                          uint32  = 0x6
	MAX_LOCALE_NAME                           uint32  = 0x20
	MAX_SCRIPT_NAME                           uint32  = 0x30
	CPIOD_PEEK                                int32   = 1073741824
	CPIOD_FORCE_PROMPT                        int32   = -2147483648
	UITER_UNKNOWN_INDEX                       int32   = -2
	UCPTRIE_FAST_SHIFT                        int32   = 6
	UCPTRIE_FAST_DATA_BLOCK_LENGTH            int32   = 64
	UCPTRIE_FAST_DATA_MASK                    int32   = 63
	UCPTRIE_SMALL_MAX                         int32   = 4095
	UCPTRIE_ERROR_VALUE_NEG_DATA_OFFSET       int32   = 1
	UCPTRIE_HIGH_VALUE_NEG_DATA_OFFSET        int32   = 2
	UTEXT_PROVIDER_LENGTH_IS_EXPENSIVE        int32   = 1
	UTEXT_PROVIDER_STABLE_CHUNKS              int32   = 2
	UTEXT_PROVIDER_WRITABLE                   int32   = 3
	UTEXT_PROVIDER_HAS_META_DATA              int32   = 4
	UTEXT_PROVIDER_OWNS_TEXT                  int32   = 5
	UTEXT_MAGIC                               int32   = 878368812
	USET_IGNORE_SPACE                         int32   = 1
	USET_CASE_INSENSITIVE                     int32   = 2
	USET_ADD_CASE_MAPPINGS                    int32   = 4
	USET_SERIALIZED_STATIC_ARRAY_CAPACITY     int32   = 8
	U_PARSE_CONTEXT_LEN                       int32   = 16
	UIDNA_DEFAULT                             int32   = 0
	UIDNA_USE_STD3_RULES                      int32   = 2
	UIDNA_CHECK_BIDI                          int32   = 4
	UIDNA_CHECK_CONTEXTJ                      int32   = 8
	UIDNA_NONTRANSITIONAL_TO_ASCII            int32   = 16
	UIDNA_NONTRANSITIONAL_TO_UNICODE          int32   = 32
	UIDNA_CHECK_CONTEXTO                      int32   = 64
	UIDNA_ERROR_EMPTY_LABEL                   int32   = 1
	UIDNA_ERROR_LABEL_TOO_LONG                int32   = 2
	UIDNA_ERROR_DOMAIN_NAME_TOO_LONG          int32   = 4
	UIDNA_ERROR_LEADING_HYPHEN                int32   = 8
	UIDNA_ERROR_TRAILING_HYPHEN               int32   = 16
	UIDNA_ERROR_HYPHEN_3_4                    int32   = 32
	UIDNA_ERROR_LEADING_COMBINING_MARK        int32   = 64
	UIDNA_ERROR_DISALLOWED                    int32   = 128
	UIDNA_ERROR_PUNYCODE                      int32   = 256
	UIDNA_ERROR_LABEL_HAS_DOT                 int32   = 512
	UIDNA_ERROR_INVALID_ACE_LABEL             int32   = 1024
	UIDNA_ERROR_BIDI                          int32   = 2048
	UIDNA_ERROR_CONTEXTJ                      int32   = 4096
	UIDNA_ERROR_CONTEXTO_PUNCTUATION          int32   = 8192
	UIDNA_ERROR_CONTEXTO_DIGITS               int32   = 16384
	UMSGPAT_ARG_NAME_NOT_NUMBER               int32   = -1
	UMSGPAT_ARG_NAME_NOT_VALID                int32   = -2
)

var (
	ELS_GUID_LANGUAGE_DETECTION = syscall.GUID{0xCF7E00B1, 0x909B, 0x4D95,
		[8]byte{0xA8, 0xF4, 0x61, 0x1F, 0x7C, 0x37, 0x77, 0x02}}

	ELS_GUID_SCRIPT_DETECTION = syscall.GUID{0x2D64B439, 0x6CAF, 0x4F6B,
		[8]byte{0xB6, 0x88, 0xE5, 0xD0, 0xF4, 0xFA, 0xA7, 0xD7}}

	ELS_GUID_TRANSLITERATION_HANT_TO_HANS = syscall.GUID{0xA3A8333B, 0xF4FC, 0x42F6,
		[8]byte{0xA0, 0xC4, 0x04, 0x62, 0xFE, 0x73, 0x17, 0xCB}}

	ELS_GUID_TRANSLITERATION_HANS_TO_HANT = syscall.GUID{0x3CACCDC8, 0x5590, 0x42DC,
		[8]byte{0x9A, 0x7B, 0xB5, 0xA6, 0xB5, 0xB3, 0xB6, 0x3B}}

	ELS_GUID_TRANSLITERATION_MALAYALAM_TO_LATIN = syscall.GUID{0xD8B983B1, 0xF8BF, 0x4A2B,
		[8]byte{0xBC, 0xD5, 0x5B, 0x5E, 0xA2, 0x06, 0x13, 0xE1}}

	ELS_GUID_TRANSLITERATION_DEVANAGARI_TO_LATIN = syscall.GUID{0xC4A4DCFE, 0x2661, 0x4D02,
		[8]byte{0x98, 0x35, 0xF4, 0x81, 0x87, 0x10, 0x98, 0x03}}

	ELS_GUID_TRANSLITERATION_CYRILLIC_TO_LATIN = syscall.GUID{0x3DD12A98, 0x5AFD, 0x4903,
		[8]byte{0xA1, 0x3F, 0xE1, 0x7E, 0x6C, 0x0B, 0xFE, 0x01}}

	ELS_GUID_TRANSLITERATION_BENGALI_TO_LATIN = syscall.GUID{0xF4DFD825, 0x91A4, 0x489F,
		[8]byte{0x85, 0x5E, 0x9A, 0xD9, 0xBE, 0xE5, 0x57, 0x27}}

	ELS_GUID_TRANSLITERATION_HANGUL_DECOMPOSITION = syscall.GUID{0x4BA2A721, 0xE43D, 0x41B7,
		[8]byte{0xB3, 0x30, 0x53, 0x6A, 0xE1, 0xE4, 0x88, 0x63}}
)

// enums

// enum
// flags
type FOLD_STRING_MAP_FLAGS uint32

const (
	MAP_COMPOSITE        FOLD_STRING_MAP_FLAGS = 64
	MAP_EXPAND_LIGATURES FOLD_STRING_MAP_FLAGS = 8192
	MAP_FOLDCZONE        FOLD_STRING_MAP_FLAGS = 16
	MAP_FOLDDIGITS       FOLD_STRING_MAP_FLAGS = 128
	MAP_PRECOMPOSED      FOLD_STRING_MAP_FLAGS = 32
)

// enum
type ENUM_DATE_FORMATS_FLAGS uint32

const (
	DATE_SHORTDATE        ENUM_DATE_FORMATS_FLAGS = 1
	DATE_LONGDATE         ENUM_DATE_FORMATS_FLAGS = 2
	DATE_YEARMONTH        ENUM_DATE_FORMATS_FLAGS = 8
	DATE_MONTHDAY         ENUM_DATE_FORMATS_FLAGS = 128
	DATE_AUTOLAYOUT       ENUM_DATE_FORMATS_FLAGS = 64
	DATE_LTRREADING       ENUM_DATE_FORMATS_FLAGS = 16
	DATE_RTLREADING       ENUM_DATE_FORMATS_FLAGS = 32
	DATE_USE_ALT_CALENDAR ENUM_DATE_FORMATS_FLAGS = 4
)

// enum
type TRANSLATE_CHARSET_INFO_FLAGS uint32

const (
	TCI_SRCCHARSET  TRANSLATE_CHARSET_INFO_FLAGS = 1
	TCI_SRCCODEPAGE TRANSLATE_CHARSET_INFO_FLAGS = 2
	TCI_SRCFONTSIG  TRANSLATE_CHARSET_INFO_FLAGS = 3
	TCI_SRCLOCALE   TRANSLATE_CHARSET_INFO_FLAGS = 4096
)

// enum
// flags
type TIME_FORMAT_FLAGS uint32

const (
	TIME_NOMINUTESORSECONDS TIME_FORMAT_FLAGS = 1
	TIME_NOSECONDS          TIME_FORMAT_FLAGS = 2
	TIME_NOTIMEMARKER       TIME_FORMAT_FLAGS = 4
	TIME_FORCE24HOURFORMAT  TIME_FORMAT_FLAGS = 8
)

// enum
type ENUM_SYSTEM_LANGUAGE_GROUPS_FLAGS uint32

const (
	LGRPID_INSTALLED ENUM_SYSTEM_LANGUAGE_GROUPS_FLAGS = 1
	LGRPID_SUPPORTED ENUM_SYSTEM_LANGUAGE_GROUPS_FLAGS = 2
)

// enum
// flags
type MULTI_BYTE_TO_WIDE_CHAR_FLAGS uint32

const (
	MB_COMPOSITE         MULTI_BYTE_TO_WIDE_CHAR_FLAGS = 2
	MB_ERR_INVALID_CHARS MULTI_BYTE_TO_WIDE_CHAR_FLAGS = 8
	MB_PRECOMPOSED       MULTI_BYTE_TO_WIDE_CHAR_FLAGS = 1
	MB_USEGLYPHCHARS     MULTI_BYTE_TO_WIDE_CHAR_FLAGS = 4
)

// enum
// flags
type COMPARE_STRING_FLAGS uint32

const (
	LINGUISTIC_IGNORECASE      COMPARE_STRING_FLAGS = 16
	LINGUISTIC_IGNOREDIACRITIC COMPARE_STRING_FLAGS = 32
	NORM_IGNORECASE            COMPARE_STRING_FLAGS = 1
	NORM_IGNOREKANATYPE        COMPARE_STRING_FLAGS = 65536
	NORM_IGNORENONSPACE        COMPARE_STRING_FLAGS = 2
	NORM_IGNORESYMBOLS         COMPARE_STRING_FLAGS = 4
	NORM_IGNOREWIDTH           COMPARE_STRING_FLAGS = 131072
	NORM_LINGUISTIC_CASING     COMPARE_STRING_FLAGS = 134217728
	SORT_DIGITSASNUMBERS       COMPARE_STRING_FLAGS = 8
	SORT_STRINGSORT            COMPARE_STRING_FLAGS = 4096
)

// enum
type IS_VALID_LOCALE_FLAGS uint32

const (
	LCID_INSTALLED IS_VALID_LOCALE_FLAGS = 1
	LCID_SUPPORTED IS_VALID_LOCALE_FLAGS = 2
)

// enum
type ENUM_SYSTEM_CODE_PAGES_FLAGS uint32

const (
	CP_INSTALLED ENUM_SYSTEM_CODE_PAGES_FLAGS = 1
	CP_SUPPORTED ENUM_SYSTEM_CODE_PAGES_FLAGS = 2
)

// enum
type SCRIPT_IS_COMPLEX_FLAGS uint32

const (
	SIC_ASCIIDIGIT SCRIPT_IS_COMPLEX_FLAGS = 2
	SIC_COMPLEX    SCRIPT_IS_COMPLEX_FLAGS = 1
	SIC_NEUTRAL    SCRIPT_IS_COMPLEX_FLAGS = 4
)

// enum
// flags
type IS_TEXT_UNICODE_RESULT uint32

const (
	IS_TEXT_UNICODE_ASCII16            IS_TEXT_UNICODE_RESULT = 1
	IS_TEXT_UNICODE_REVERSE_ASCII16    IS_TEXT_UNICODE_RESULT = 16
	IS_TEXT_UNICODE_STATISTICS         IS_TEXT_UNICODE_RESULT = 2
	IS_TEXT_UNICODE_REVERSE_STATISTICS IS_TEXT_UNICODE_RESULT = 32
	IS_TEXT_UNICODE_CONTROLS           IS_TEXT_UNICODE_RESULT = 4
	IS_TEXT_UNICODE_REVERSE_CONTROLS   IS_TEXT_UNICODE_RESULT = 64
	IS_TEXT_UNICODE_SIGNATURE          IS_TEXT_UNICODE_RESULT = 8
	IS_TEXT_UNICODE_REVERSE_SIGNATURE  IS_TEXT_UNICODE_RESULT = 128
	IS_TEXT_UNICODE_ILLEGAL_CHARS      IS_TEXT_UNICODE_RESULT = 256
	IS_TEXT_UNICODE_ODD_LENGTH         IS_TEXT_UNICODE_RESULT = 512
	IS_TEXT_UNICODE_NULL_BYTES         IS_TEXT_UNICODE_RESULT = 4096
	IS_TEXT_UNICODE_UNICODE_MASK       IS_TEXT_UNICODE_RESULT = 15
	IS_TEXT_UNICODE_REVERSE_MASK       IS_TEXT_UNICODE_RESULT = 240
	IS_TEXT_UNICODE_NOT_UNICODE_MASK   IS_TEXT_UNICODE_RESULT = 3840
	IS_TEXT_UNICODE_NOT_ASCII_MASK     IS_TEXT_UNICODE_RESULT = 61440
)

// enum
type SYSNLS_FUNCTION int32

const (
	COMPARE_STRING SYSNLS_FUNCTION = 1
)

// enum
type SYSGEOTYPE int32

const (
	GEO_NATION            SYSGEOTYPE = 1
	GEO_LATITUDE          SYSGEOTYPE = 2
	GEO_LONGITUDE         SYSGEOTYPE = 3
	GEO_ISO2              SYSGEOTYPE = 4
	GEO_ISO3              SYSGEOTYPE = 5
	GEO_RFC1766           SYSGEOTYPE = 6
	GEO_LCID              SYSGEOTYPE = 7
	GEO_FRIENDLYNAME      SYSGEOTYPE = 8
	GEO_OFFICIALNAME      SYSGEOTYPE = 9
	GEO_TIMEZONES         SYSGEOTYPE = 10
	GEO_OFFICIALLANGUAGES SYSGEOTYPE = 11
	GEO_ISO_UN_NUMBER     SYSGEOTYPE = 12
	GEO_PARENT            SYSGEOTYPE = 13
	GEO_DIALINGCODE       SYSGEOTYPE = 14
	GEO_CURRENCYCODE      SYSGEOTYPE = 15
	GEO_CURRENCYSYMBOL    SYSGEOTYPE = 16
	GEO_NAME              SYSGEOTYPE = 17
	GEO_ID                SYSGEOTYPE = 18
)

// enum
type SYSGEOCLASS int32

const (
	GEOCLASS_NATION SYSGEOCLASS = 16
	GEOCLASS_REGION SYSGEOCLASS = 14
	GEOCLASS_ALL    SYSGEOCLASS = 0
)

// enum
type NORM_FORM int32

const (
	NormalizationOther NORM_FORM = 0
	NormalizationC     NORM_FORM = 1
	NormalizationD     NORM_FORM = 2
	NormalizationKC    NORM_FORM = 5
	NormalizationKD    NORM_FORM = 6
)

// enum
type WORDLIST_TYPE int32

const (
	WORDLIST_TYPE_IGNORE      WORDLIST_TYPE = 0
	WORDLIST_TYPE_ADD         WORDLIST_TYPE = 1
	WORDLIST_TYPE_EXCLUDE     WORDLIST_TYPE = 2
	WORDLIST_TYPE_AUTOCORRECT WORDLIST_TYPE = 3
)

// enum
type CORRECTIVE_ACTION int32

const (
	CORRECTIVE_ACTION_NONE            CORRECTIVE_ACTION = 0
	CORRECTIVE_ACTION_GET_SUGGESTIONS CORRECTIVE_ACTION = 1
	CORRECTIVE_ACTION_REPLACE         CORRECTIVE_ACTION = 2
	CORRECTIVE_ACTION_DELETE          CORRECTIVE_ACTION = 3
)

// enum
type SCRIPT_JUSTIFY int32

const (
	SCRIPT_JUSTIFY_NONE           SCRIPT_JUSTIFY = 0
	SCRIPT_JUSTIFY_ARABIC_BLANK   SCRIPT_JUSTIFY = 1
	SCRIPT_JUSTIFY_CHARACTER      SCRIPT_JUSTIFY = 2
	SCRIPT_JUSTIFY_RESERVED1      SCRIPT_JUSTIFY = 3
	SCRIPT_JUSTIFY_BLANK          SCRIPT_JUSTIFY = 4
	SCRIPT_JUSTIFY_RESERVED2      SCRIPT_JUSTIFY = 5
	SCRIPT_JUSTIFY_RESERVED3      SCRIPT_JUSTIFY = 6
	SCRIPT_JUSTIFY_ARABIC_NORMAL  SCRIPT_JUSTIFY = 7
	SCRIPT_JUSTIFY_ARABIC_KASHIDA SCRIPT_JUSTIFY = 8
	SCRIPT_JUSTIFY_ARABIC_ALEF    SCRIPT_JUSTIFY = 9
	SCRIPT_JUSTIFY_ARABIC_HA      SCRIPT_JUSTIFY = 10
	SCRIPT_JUSTIFY_ARABIC_RA      SCRIPT_JUSTIFY = 11
	SCRIPT_JUSTIFY_ARABIC_BA      SCRIPT_JUSTIFY = 12
	SCRIPT_JUSTIFY_ARABIC_BARA    SCRIPT_JUSTIFY = 13
	SCRIPT_JUSTIFY_ARABIC_SEEN    SCRIPT_JUSTIFY = 14
	SCRIPT_JUSTIFY_ARABIC_SEEN_M  SCRIPT_JUSTIFY = 15
)

// enum
type UErrorCode int32

const (
	U_USING_FALLBACK_WARNING           UErrorCode = -128
	U_ERROR_WARNING_START              UErrorCode = -128
	U_USING_DEFAULT_WARNING            UErrorCode = -127
	U_SAFECLONE_ALLOCATED_WARNING      UErrorCode = -126
	U_STATE_OLD_WARNING                UErrorCode = -125
	U_STRING_NOT_TERMINATED_WARNING    UErrorCode = -124
	U_SORT_KEY_TOO_SHORT_WARNING       UErrorCode = -123
	U_AMBIGUOUS_ALIAS_WARNING          UErrorCode = -122
	U_DIFFERENT_UCA_VERSION            UErrorCode = -121
	U_PLUGIN_CHANGED_LEVEL_WARNING     UErrorCode = -120
	U_ZERO_ERROR                       UErrorCode = 0
	U_ILLEGAL_ARGUMENT_ERROR           UErrorCode = 1
	U_MISSING_RESOURCE_ERROR           UErrorCode = 2
	U_INVALID_FORMAT_ERROR             UErrorCode = 3
	U_FILE_ACCESS_ERROR                UErrorCode = 4
	U_INTERNAL_PROGRAM_ERROR           UErrorCode = 5
	U_MESSAGE_PARSE_ERROR              UErrorCode = 6
	U_MEMORY_ALLOCATION_ERROR          UErrorCode = 7
	U_INDEX_OUTOFBOUNDS_ERROR          UErrorCode = 8
	U_PARSE_ERROR                      UErrorCode = 9
	U_INVALID_CHAR_FOUND               UErrorCode = 10
	U_TRUNCATED_CHAR_FOUND             UErrorCode = 11
	U_ILLEGAL_CHAR_FOUND               UErrorCode = 12
	U_INVALID_TABLE_FORMAT             UErrorCode = 13
	U_INVALID_TABLE_FILE               UErrorCode = 14
	U_BUFFER_OVERFLOW_ERROR            UErrorCode = 15
	U_UNSUPPORTED_ERROR                UErrorCode = 16
	U_RESOURCE_TYPE_MISMATCH           UErrorCode = 17
	U_ILLEGAL_ESCAPE_SEQUENCE          UErrorCode = 18
	U_UNSUPPORTED_ESCAPE_SEQUENCE      UErrorCode = 19
	U_NO_SPACE_AVAILABLE               UErrorCode = 20
	U_CE_NOT_FOUND_ERROR               UErrorCode = 21
	U_PRIMARY_TOO_LONG_ERROR           UErrorCode = 22
	U_STATE_TOO_OLD_ERROR              UErrorCode = 23
	U_TOO_MANY_ALIASES_ERROR           UErrorCode = 24
	U_ENUM_OUT_OF_SYNC_ERROR           UErrorCode = 25
	U_INVARIANT_CONVERSION_ERROR       UErrorCode = 26
	U_INVALID_STATE_ERROR              UErrorCode = 27
	U_COLLATOR_VERSION_MISMATCH        UErrorCode = 28
	U_USELESS_COLLATOR_ERROR           UErrorCode = 29
	U_NO_WRITE_PERMISSION              UErrorCode = 30
	U_BAD_VARIABLE_DEFINITION          UErrorCode = 65536
	U_PARSE_ERROR_START                UErrorCode = 65536
	U_MALFORMED_RULE                   UErrorCode = 65537
	U_MALFORMED_SET                    UErrorCode = 65538
	U_MALFORMED_SYMBOL_REFERENCE       UErrorCode = 65539
	U_MALFORMED_UNICODE_ESCAPE         UErrorCode = 65540
	U_MALFORMED_VARIABLE_DEFINITION    UErrorCode = 65541
	U_MALFORMED_VARIABLE_REFERENCE     UErrorCode = 65542
	U_MISMATCHED_SEGMENT_DELIMITERS    UErrorCode = 65543
	U_MISPLACED_ANCHOR_START           UErrorCode = 65544
	U_MISPLACED_CURSOR_OFFSET          UErrorCode = 65545
	U_MISPLACED_QUANTIFIER             UErrorCode = 65546
	U_MISSING_OPERATOR                 UErrorCode = 65547
	U_MISSING_SEGMENT_CLOSE            UErrorCode = 65548
	U_MULTIPLE_ANTE_CONTEXTS           UErrorCode = 65549
	U_MULTIPLE_CURSORS                 UErrorCode = 65550
	U_MULTIPLE_POST_CONTEXTS           UErrorCode = 65551
	U_TRAILING_BACKSLASH               UErrorCode = 65552
	U_UNDEFINED_SEGMENT_REFERENCE      UErrorCode = 65553
	U_UNDEFINED_VARIABLE               UErrorCode = 65554
	U_UNQUOTED_SPECIAL                 UErrorCode = 65555
	U_UNTERMINATED_QUOTE               UErrorCode = 65556
	U_RULE_MASK_ERROR                  UErrorCode = 65557
	U_MISPLACED_COMPOUND_FILTER        UErrorCode = 65558
	U_MULTIPLE_COMPOUND_FILTERS        UErrorCode = 65559
	U_INVALID_RBT_SYNTAX               UErrorCode = 65560
	U_INVALID_PROPERTY_PATTERN         UErrorCode = 65561
	U_MALFORMED_PRAGMA                 UErrorCode = 65562
	U_UNCLOSED_SEGMENT                 UErrorCode = 65563
	U_ILLEGAL_CHAR_IN_SEGMENT          UErrorCode = 65564
	U_VARIABLE_RANGE_EXHAUSTED         UErrorCode = 65565
	U_VARIABLE_RANGE_OVERLAP           UErrorCode = 65566
	U_ILLEGAL_CHARACTER                UErrorCode = 65567
	U_INTERNAL_TRANSLITERATOR_ERROR    UErrorCode = 65568
	U_INVALID_ID                       UErrorCode = 65569
	U_INVALID_FUNCTION                 UErrorCode = 65570
	U_UNEXPECTED_TOKEN                 UErrorCode = 65792
	U_FMT_PARSE_ERROR_START            UErrorCode = 65792
	U_MULTIPLE_DECIMAL_SEPARATORS      UErrorCode = 65793
	U_MULTIPLE_DECIMAL_SEPERATORS      UErrorCode = 65793
	U_MULTIPLE_EXPONENTIAL_SYMBOLS     UErrorCode = 65794
	U_MALFORMED_EXPONENTIAL_PATTERN    UErrorCode = 65795
	U_MULTIPLE_PERCENT_SYMBOLS         UErrorCode = 65796
	U_MULTIPLE_PERMILL_SYMBOLS         UErrorCode = 65797
	U_MULTIPLE_PAD_SPECIFIERS          UErrorCode = 65798
	U_PATTERN_SYNTAX_ERROR             UErrorCode = 65799
	U_ILLEGAL_PAD_POSITION             UErrorCode = 65800
	U_UNMATCHED_BRACES                 UErrorCode = 65801
	U_UNSUPPORTED_PROPERTY             UErrorCode = 65802
	U_UNSUPPORTED_ATTRIBUTE            UErrorCode = 65803
	U_ARGUMENT_TYPE_MISMATCH           UErrorCode = 65804
	U_DUPLICATE_KEYWORD                UErrorCode = 65805
	U_UNDEFINED_KEYWORD                UErrorCode = 65806
	U_DEFAULT_KEYWORD_MISSING          UErrorCode = 65807
	U_DECIMAL_NUMBER_SYNTAX_ERROR      UErrorCode = 65808
	U_FORMAT_INEXACT_ERROR             UErrorCode = 65809
	U_NUMBER_ARG_OUTOFBOUNDS_ERROR     UErrorCode = 65810
	U_NUMBER_SKELETON_SYNTAX_ERROR     UErrorCode = 65811
	U_BRK_INTERNAL_ERROR               UErrorCode = 66048
	U_BRK_ERROR_START                  UErrorCode = 66048
	U_BRK_HEX_DIGITS_EXPECTED          UErrorCode = 66049
	U_BRK_SEMICOLON_EXPECTED           UErrorCode = 66050
	U_BRK_RULE_SYNTAX                  UErrorCode = 66051
	U_BRK_UNCLOSED_SET                 UErrorCode = 66052
	U_BRK_ASSIGN_ERROR                 UErrorCode = 66053
	U_BRK_VARIABLE_REDFINITION         UErrorCode = 66054
	U_BRK_MISMATCHED_PAREN             UErrorCode = 66055
	U_BRK_NEW_LINE_IN_QUOTED_STRING    UErrorCode = 66056
	U_BRK_UNDEFINED_VARIABLE           UErrorCode = 66057
	U_BRK_INIT_ERROR                   UErrorCode = 66058
	U_BRK_RULE_EMPTY_SET               UErrorCode = 66059
	U_BRK_UNRECOGNIZED_OPTION          UErrorCode = 66060
	U_BRK_MALFORMED_RULE_TAG           UErrorCode = 66061
	U_REGEX_INTERNAL_ERROR             UErrorCode = 66304
	U_REGEX_ERROR_START                UErrorCode = 66304
	U_REGEX_RULE_SYNTAX                UErrorCode = 66305
	U_REGEX_INVALID_STATE              UErrorCode = 66306
	U_REGEX_BAD_ESCAPE_SEQUENCE        UErrorCode = 66307
	U_REGEX_PROPERTY_SYNTAX            UErrorCode = 66308
	U_REGEX_UNIMPLEMENTED              UErrorCode = 66309
	U_REGEX_MISMATCHED_PAREN           UErrorCode = 66310
	U_REGEX_NUMBER_TOO_BIG             UErrorCode = 66311
	U_REGEX_BAD_INTERVAL               UErrorCode = 66312
	U_REGEX_MAX_LT_MIN                 UErrorCode = 66313
	U_REGEX_INVALID_BACK_REF           UErrorCode = 66314
	U_REGEX_INVALID_FLAG               UErrorCode = 66315
	U_REGEX_LOOK_BEHIND_LIMIT          UErrorCode = 66316
	U_REGEX_SET_CONTAINS_STRING        UErrorCode = 66317
	U_REGEX_MISSING_CLOSE_BRACKET      UErrorCode = 66319
	U_REGEX_INVALID_RANGE              UErrorCode = 66320
	U_REGEX_STACK_OVERFLOW             UErrorCode = 66321
	U_REGEX_TIME_OUT                   UErrorCode = 66322
	U_REGEX_STOPPED_BY_CALLER          UErrorCode = 66323
	U_REGEX_PATTERN_TOO_BIG            UErrorCode = 66324
	U_REGEX_INVALID_CAPTURE_GROUP_NAME UErrorCode = 66325
	U_IDNA_PROHIBITED_ERROR            UErrorCode = 66560
	U_IDNA_ERROR_START                 UErrorCode = 66560
	U_IDNA_UNASSIGNED_ERROR            UErrorCode = 66561
	U_IDNA_CHECK_BIDI_ERROR            UErrorCode = 66562
	U_IDNA_STD3_ASCII_RULES_ERROR      UErrorCode = 66563
	U_IDNA_ACE_PREFIX_ERROR            UErrorCode = 66564
	U_IDNA_VERIFICATION_ERROR          UErrorCode = 66565
	U_IDNA_LABEL_TOO_LONG_ERROR        UErrorCode = 66566
	U_IDNA_ZERO_LENGTH_LABEL_ERROR     UErrorCode = 66567
	U_IDNA_DOMAIN_NAME_TOO_LONG_ERROR  UErrorCode = 66568
	U_STRINGPREP_PROHIBITED_ERROR      UErrorCode = 66560
	U_STRINGPREP_UNASSIGNED_ERROR      UErrorCode = 66561
	U_STRINGPREP_CHECK_BIDI_ERROR      UErrorCode = 66562
	U_PLUGIN_ERROR_START               UErrorCode = 66816
	U_PLUGIN_TOO_HIGH                  UErrorCode = 66816
	U_PLUGIN_DIDNT_SET_LEVEL           UErrorCode = 66817
)

// enum
type UTraceLevel int32

const (
	UTRACE_OFF        UTraceLevel = -1
	UTRACE_ERROR      UTraceLevel = 0
	UTRACE_WARNING    UTraceLevel = 3
	UTRACE_OPEN_CLOSE UTraceLevel = 5
	UTRACE_INFO       UTraceLevel = 7
	UTRACE_VERBOSE    UTraceLevel = 9
)

// enum
type UTraceFunctionNumber int32

const (
	UTRACE_FUNCTION_START              UTraceFunctionNumber = 0
	UTRACE_U_INIT                      UTraceFunctionNumber = 0
	UTRACE_U_CLEANUP                   UTraceFunctionNumber = 1
	UTRACE_CONVERSION_START            UTraceFunctionNumber = 4096
	UTRACE_UCNV_OPEN                   UTraceFunctionNumber = 4096
	UTRACE_UCNV_OPEN_PACKAGE           UTraceFunctionNumber = 4097
	UTRACE_UCNV_OPEN_ALGORITHMIC       UTraceFunctionNumber = 4098
	UTRACE_UCNV_CLONE                  UTraceFunctionNumber = 4099
	UTRACE_UCNV_CLOSE                  UTraceFunctionNumber = 4100
	UTRACE_UCNV_FLUSH_CACHE            UTraceFunctionNumber = 4101
	UTRACE_UCNV_LOAD                   UTraceFunctionNumber = 4102
	UTRACE_UCNV_UNLOAD                 UTraceFunctionNumber = 4103
	UTRACE_COLLATION_START             UTraceFunctionNumber = 8192
	UTRACE_UCOL_OPEN                   UTraceFunctionNumber = 8192
	UTRACE_UCOL_CLOSE                  UTraceFunctionNumber = 8193
	UTRACE_UCOL_STRCOLL                UTraceFunctionNumber = 8194
	UTRACE_UCOL_GET_SORTKEY            UTraceFunctionNumber = 8195
	UTRACE_UCOL_GETLOCALE              UTraceFunctionNumber = 8196
	UTRACE_UCOL_NEXTSORTKEYPART        UTraceFunctionNumber = 8197
	UTRACE_UCOL_STRCOLLITER            UTraceFunctionNumber = 8198
	UTRACE_UCOL_OPEN_FROM_SHORT_STRING UTraceFunctionNumber = 8199
	UTRACE_UCOL_STRCOLLUTF8            UTraceFunctionNumber = 8200
	UTRACE_UDATA_START                 UTraceFunctionNumber = 12288
	UTRACE_UDATA_RESOURCE              UTraceFunctionNumber = 12288
	UTRACE_UDATA_BUNDLE                UTraceFunctionNumber = 12289
	UTRACE_UDATA_DATA_FILE             UTraceFunctionNumber = 12290
	UTRACE_UDATA_RES_FILE              UTraceFunctionNumber = 12291
)

// enum
type UStringTrieResult int32

const (
	USTRINGTRIE_NO_MATCH           UStringTrieResult = 0
	USTRINGTRIE_NO_VALUE           UStringTrieResult = 1
	USTRINGTRIE_FINAL_VALUE        UStringTrieResult = 2
	USTRINGTRIE_INTERMEDIATE_VALUE UStringTrieResult = 3
)

// enum
type UScriptCode int32

const (
	USCRIPT_INVALID_CODE                 UScriptCode = -1
	USCRIPT_COMMON                       UScriptCode = 0
	USCRIPT_INHERITED                    UScriptCode = 1
	USCRIPT_ARABIC                       UScriptCode = 2
	USCRIPT_ARMENIAN                     UScriptCode = 3
	USCRIPT_BENGALI                      UScriptCode = 4
	USCRIPT_BOPOMOFO                     UScriptCode = 5
	USCRIPT_CHEROKEE                     UScriptCode = 6
	USCRIPT_COPTIC                       UScriptCode = 7
	USCRIPT_CYRILLIC                     UScriptCode = 8
	USCRIPT_DESERET                      UScriptCode = 9
	USCRIPT_DEVANAGARI                   UScriptCode = 10
	USCRIPT_ETHIOPIC                     UScriptCode = 11
	USCRIPT_GEORGIAN                     UScriptCode = 12
	USCRIPT_GOTHIC                       UScriptCode = 13
	USCRIPT_GREEK                        UScriptCode = 14
	USCRIPT_GUJARATI                     UScriptCode = 15
	USCRIPT_GURMUKHI                     UScriptCode = 16
	USCRIPT_HAN                          UScriptCode = 17
	USCRIPT_HANGUL                       UScriptCode = 18
	USCRIPT_HEBREW                       UScriptCode = 19
	USCRIPT_HIRAGANA                     UScriptCode = 20
	USCRIPT_KANNADA                      UScriptCode = 21
	USCRIPT_KATAKANA                     UScriptCode = 22
	USCRIPT_KHMER                        UScriptCode = 23
	USCRIPT_LAO                          UScriptCode = 24
	USCRIPT_LATIN                        UScriptCode = 25
	USCRIPT_MALAYALAM                    UScriptCode = 26
	USCRIPT_MONGOLIAN                    UScriptCode = 27
	USCRIPT_MYANMAR                      UScriptCode = 28
	USCRIPT_OGHAM                        UScriptCode = 29
	USCRIPT_OLD_ITALIC                   UScriptCode = 30
	USCRIPT_ORIYA                        UScriptCode = 31
	USCRIPT_RUNIC                        UScriptCode = 32
	USCRIPT_SINHALA                      UScriptCode = 33
	USCRIPT_SYRIAC                       UScriptCode = 34
	USCRIPT_TAMIL                        UScriptCode = 35
	USCRIPT_TELUGU                       UScriptCode = 36
	USCRIPT_THAANA                       UScriptCode = 37
	USCRIPT_THAI                         UScriptCode = 38
	USCRIPT_TIBETAN                      UScriptCode = 39
	USCRIPT_CANADIAN_ABORIGINAL          UScriptCode = 40
	USCRIPT_UCAS                         UScriptCode = 40
	USCRIPT_YI                           UScriptCode = 41
	USCRIPT_TAGALOG                      UScriptCode = 42
	USCRIPT_HANUNOO                      UScriptCode = 43
	USCRIPT_BUHID                        UScriptCode = 44
	USCRIPT_TAGBANWA                     UScriptCode = 45
	USCRIPT_BRAILLE                      UScriptCode = 46
	USCRIPT_CYPRIOT                      UScriptCode = 47
	USCRIPT_LIMBU                        UScriptCode = 48
	USCRIPT_LINEAR_B                     UScriptCode = 49
	USCRIPT_OSMANYA                      UScriptCode = 50
	USCRIPT_SHAVIAN                      UScriptCode = 51
	USCRIPT_TAI_LE                       UScriptCode = 52
	USCRIPT_UGARITIC                     UScriptCode = 53
	USCRIPT_KATAKANA_OR_HIRAGANA         UScriptCode = 54
	USCRIPT_BUGINESE                     UScriptCode = 55
	USCRIPT_GLAGOLITIC                   UScriptCode = 56
	USCRIPT_KHAROSHTHI                   UScriptCode = 57
	USCRIPT_SYLOTI_NAGRI                 UScriptCode = 58
	USCRIPT_NEW_TAI_LUE                  UScriptCode = 59
	USCRIPT_TIFINAGH                     UScriptCode = 60
	USCRIPT_OLD_PERSIAN                  UScriptCode = 61
	USCRIPT_BALINESE                     UScriptCode = 62
	USCRIPT_BATAK                        UScriptCode = 63
	USCRIPT_BLISSYMBOLS                  UScriptCode = 64
	USCRIPT_BRAHMI                       UScriptCode = 65
	USCRIPT_CHAM                         UScriptCode = 66
	USCRIPT_CIRTH                        UScriptCode = 67
	USCRIPT_OLD_CHURCH_SLAVONIC_CYRILLIC UScriptCode = 68
	USCRIPT_DEMOTIC_EGYPTIAN             UScriptCode = 69
	USCRIPT_HIERATIC_EGYPTIAN            UScriptCode = 70
	USCRIPT_EGYPTIAN_HIEROGLYPHS         UScriptCode = 71
	USCRIPT_KHUTSURI                     UScriptCode = 72
	USCRIPT_SIMPLIFIED_HAN               UScriptCode = 73
	USCRIPT_TRADITIONAL_HAN              UScriptCode = 74
	USCRIPT_PAHAWH_HMONG                 UScriptCode = 75
	USCRIPT_OLD_HUNGARIAN                UScriptCode = 76
	USCRIPT_HARAPPAN_INDUS               UScriptCode = 77
	USCRIPT_JAVANESE                     UScriptCode = 78
	USCRIPT_KAYAH_LI                     UScriptCode = 79
	USCRIPT_LATIN_FRAKTUR                UScriptCode = 80
	USCRIPT_LATIN_GAELIC                 UScriptCode = 81
	USCRIPT_LEPCHA                       UScriptCode = 82
	USCRIPT_LINEAR_A                     UScriptCode = 83
	USCRIPT_MANDAIC                      UScriptCode = 84
	USCRIPT_MANDAEAN                     UScriptCode = 84
	USCRIPT_MAYAN_HIEROGLYPHS            UScriptCode = 85
	USCRIPT_MEROITIC_HIEROGLYPHS         UScriptCode = 86
	USCRIPT_MEROITIC                     UScriptCode = 86
	USCRIPT_NKO                          UScriptCode = 87
	USCRIPT_ORKHON                       UScriptCode = 88
	USCRIPT_OLD_PERMIC                   UScriptCode = 89
	USCRIPT_PHAGS_PA                     UScriptCode = 90
	USCRIPT_PHOENICIAN                   UScriptCode = 91
	USCRIPT_MIAO                         UScriptCode = 92
	USCRIPT_PHONETIC_POLLARD             UScriptCode = 92
	USCRIPT_RONGORONGO                   UScriptCode = 93
	USCRIPT_SARATI                       UScriptCode = 94
	USCRIPT_ESTRANGELO_SYRIAC            UScriptCode = 95
	USCRIPT_WESTERN_SYRIAC               UScriptCode = 96
	USCRIPT_EASTERN_SYRIAC               UScriptCode = 97
	USCRIPT_TENGWAR                      UScriptCode = 98
	USCRIPT_VAI                          UScriptCode = 99
	USCRIPT_VISIBLE_SPEECH               UScriptCode = 100
	USCRIPT_CUNEIFORM                    UScriptCode = 101
	USCRIPT_UNWRITTEN_LANGUAGES          UScriptCode = 102
	USCRIPT_UNKNOWN                      UScriptCode = 103
	USCRIPT_CARIAN                       UScriptCode = 104
	USCRIPT_JAPANESE                     UScriptCode = 105
	USCRIPT_LANNA                        UScriptCode = 106
	USCRIPT_LYCIAN                       UScriptCode = 107
	USCRIPT_LYDIAN                       UScriptCode = 108
	USCRIPT_OL_CHIKI                     UScriptCode = 109
	USCRIPT_REJANG                       UScriptCode = 110
	USCRIPT_SAURASHTRA                   UScriptCode = 111
	USCRIPT_SIGN_WRITING                 UScriptCode = 112
	USCRIPT_SUNDANESE                    UScriptCode = 113
	USCRIPT_MOON                         UScriptCode = 114
	USCRIPT_MEITEI_MAYEK                 UScriptCode = 115
	USCRIPT_IMPERIAL_ARAMAIC             UScriptCode = 116
	USCRIPT_AVESTAN                      UScriptCode = 117
	USCRIPT_CHAKMA                       UScriptCode = 118
	USCRIPT_KOREAN                       UScriptCode = 119
	USCRIPT_KAITHI                       UScriptCode = 120
	USCRIPT_MANICHAEAN                   UScriptCode = 121
	USCRIPT_INSCRIPTIONAL_PAHLAVI        UScriptCode = 122
	USCRIPT_PSALTER_PAHLAVI              UScriptCode = 123
	USCRIPT_BOOK_PAHLAVI                 UScriptCode = 124
	USCRIPT_INSCRIPTIONAL_PARTHIAN       UScriptCode = 125
	USCRIPT_SAMARITAN                    UScriptCode = 126
	USCRIPT_TAI_VIET                     UScriptCode = 127
	USCRIPT_MATHEMATICAL_NOTATION        UScriptCode = 128
	USCRIPT_SYMBOLS                      UScriptCode = 129
	USCRIPT_BAMUM                        UScriptCode = 130
	USCRIPT_LISU                         UScriptCode = 131
	USCRIPT_NAKHI_GEBA                   UScriptCode = 132
	USCRIPT_OLD_SOUTH_ARABIAN            UScriptCode = 133
	USCRIPT_BASSA_VAH                    UScriptCode = 134
	USCRIPT_DUPLOYAN                     UScriptCode = 135
	USCRIPT_ELBASAN                      UScriptCode = 136
	USCRIPT_GRANTHA                      UScriptCode = 137
	USCRIPT_KPELLE                       UScriptCode = 138
	USCRIPT_LOMA                         UScriptCode = 139
	USCRIPT_MENDE                        UScriptCode = 140
	USCRIPT_MEROITIC_CURSIVE             UScriptCode = 141
	USCRIPT_OLD_NORTH_ARABIAN            UScriptCode = 142
	USCRIPT_NABATAEAN                    UScriptCode = 143
	USCRIPT_PALMYRENE                    UScriptCode = 144
	USCRIPT_KHUDAWADI                    UScriptCode = 145
	USCRIPT_SINDHI                       UScriptCode = 145
	USCRIPT_WARANG_CITI                  UScriptCode = 146
	USCRIPT_AFAKA                        UScriptCode = 147
	USCRIPT_JURCHEN                      UScriptCode = 148
	USCRIPT_MRO                          UScriptCode = 149
	USCRIPT_NUSHU                        UScriptCode = 150
	USCRIPT_SHARADA                      UScriptCode = 151
	USCRIPT_SORA_SOMPENG                 UScriptCode = 152
	USCRIPT_TAKRI                        UScriptCode = 153
	USCRIPT_TANGUT                       UScriptCode = 154
	USCRIPT_WOLEAI                       UScriptCode = 155
	USCRIPT_ANATOLIAN_HIEROGLYPHS        UScriptCode = 156
	USCRIPT_KHOJKI                       UScriptCode = 157
	USCRIPT_TIRHUTA                      UScriptCode = 158
	USCRIPT_CAUCASIAN_ALBANIAN           UScriptCode = 159
	USCRIPT_MAHAJANI                     UScriptCode = 160
	USCRIPT_AHOM                         UScriptCode = 161
	USCRIPT_HATRAN                       UScriptCode = 162
	USCRIPT_MODI                         UScriptCode = 163
	USCRIPT_MULTANI                      UScriptCode = 164
	USCRIPT_PAU_CIN_HAU                  UScriptCode = 165
	USCRIPT_SIDDHAM                      UScriptCode = 166
	USCRIPT_ADLAM                        UScriptCode = 167
	USCRIPT_BHAIKSUKI                    UScriptCode = 168
	USCRIPT_MARCHEN                      UScriptCode = 169
	USCRIPT_NEWA                         UScriptCode = 170
	USCRIPT_OSAGE                        UScriptCode = 171
	USCRIPT_HAN_WITH_BOPOMOFO            UScriptCode = 172
	USCRIPT_JAMO                         UScriptCode = 173
	USCRIPT_SYMBOLS_EMOJI                UScriptCode = 174
	USCRIPT_MASARAM_GONDI                UScriptCode = 175
	USCRIPT_SOYOMBO                      UScriptCode = 176
	USCRIPT_ZANABAZAR_SQUARE             UScriptCode = 177
	USCRIPT_DOGRA                        UScriptCode = 178
	USCRIPT_GUNJALA_GONDI                UScriptCode = 179
	USCRIPT_MAKASAR                      UScriptCode = 180
	USCRIPT_MEDEFAIDRIN                  UScriptCode = 181
	USCRIPT_HANIFI_ROHINGYA              UScriptCode = 182
	USCRIPT_SOGDIAN                      UScriptCode = 183
	USCRIPT_OLD_SOGDIAN                  UScriptCode = 184
	USCRIPT_ELYMAIC                      UScriptCode = 185
	USCRIPT_NYIAKENG_PUACHUE_HMONG       UScriptCode = 186
	USCRIPT_NANDINAGARI                  UScriptCode = 187
	USCRIPT_WANCHO                       UScriptCode = 188
	USCRIPT_CHORASMIAN                   UScriptCode = 189
	USCRIPT_DIVES_AKURU                  UScriptCode = 190
	USCRIPT_KHITAN_SMALL_SCRIPT          UScriptCode = 191
	USCRIPT_YEZIDI                       UScriptCode = 192
)

// enum
type UScriptUsage int32

const (
	USCRIPT_USAGE_NOT_ENCODED  UScriptUsage = 0
	USCRIPT_USAGE_UNKNOWN      UScriptUsage = 1
	USCRIPT_USAGE_EXCLUDED     UScriptUsage = 2
	USCRIPT_USAGE_LIMITED_USE  UScriptUsage = 3
	USCRIPT_USAGE_ASPIRATIONAL UScriptUsage = 4
	USCRIPT_USAGE_RECOMMENDED  UScriptUsage = 5
)

// enum
type UCharIteratorOrigin int32

const (
	UITER_START   UCharIteratorOrigin = 0
	UITER_CURRENT UCharIteratorOrigin = 1
	UITER_LIMIT   UCharIteratorOrigin = 2
	UITER_ZERO    UCharIteratorOrigin = 3
	UITER_LENGTH  UCharIteratorOrigin = 4
)

// enum
type ULocDataLocaleType int32

const (
	ULOC_ACTUAL_LOCALE ULocDataLocaleType = 0
	ULOC_VALID_LOCALE  ULocDataLocaleType = 1
)

// enum
type ULocAvailableType int32

const (
	ULOC_AVAILABLE_DEFAULT             ULocAvailableType = 0
	ULOC_AVAILABLE_ONLY_LEGACY_ALIASES ULocAvailableType = 1
	ULOC_AVAILABLE_WITH_LEGACY_ALIASES ULocAvailableType = 2
)

// enum
type ULayoutType int32

const (
	ULOC_LAYOUT_LTR     ULayoutType = 0
	ULOC_LAYOUT_RTL     ULayoutType = 1
	ULOC_LAYOUT_TTB     ULayoutType = 2
	ULOC_LAYOUT_BTT     ULayoutType = 3
	ULOC_LAYOUT_UNKNOWN ULayoutType = 4
)

// enum
type UAcceptResult int32

const (
	ULOC_ACCEPT_FAILED   UAcceptResult = 0
	ULOC_ACCEPT_VALID    UAcceptResult = 1
	ULOC_ACCEPT_FALLBACK UAcceptResult = 2
)

// enum
type UResType int32

const (
	URES_NONE       UResType = -1
	URES_STRING     UResType = 0
	URES_BINARY     UResType = 1
	URES_TABLE      UResType = 2
	URES_ALIAS      UResType = 3
	URES_INT        UResType = 7
	URES_ARRAY      UResType = 8
	URES_INT_VECTOR UResType = 14
)

// enum
type UDisplayContextType int32

const (
	UDISPCTX_TYPE_DIALECT_HANDLING    UDisplayContextType = 0
	UDISPCTX_TYPE_CAPITALIZATION      UDisplayContextType = 1
	UDISPCTX_TYPE_DISPLAY_LENGTH      UDisplayContextType = 2
	UDISPCTX_TYPE_SUBSTITUTE_HANDLING UDisplayContextType = 3
)

// enum
type UDisplayContext int32

const (
	UDISPCTX_STANDARD_NAMES                           UDisplayContext = 0
	UDISPCTX_DIALECT_NAMES                            UDisplayContext = 1
	UDISPCTX_CAPITALIZATION_NONE                      UDisplayContext = 256
	UDISPCTX_CAPITALIZATION_FOR_MIDDLE_OF_SENTENCE    UDisplayContext = 257
	UDISPCTX_CAPITALIZATION_FOR_BEGINNING_OF_SENTENCE UDisplayContext = 258
	UDISPCTX_CAPITALIZATION_FOR_UI_LIST_OR_MENU       UDisplayContext = 259
	UDISPCTX_CAPITALIZATION_FOR_STANDALONE            UDisplayContext = 260
	UDISPCTX_LENGTH_FULL                              UDisplayContext = 512
	UDISPCTX_LENGTH_SHORT                             UDisplayContext = 513
	UDISPCTX_SUBSTITUTE                               UDisplayContext = 768
	UDISPCTX_NO_SUBSTITUTE                            UDisplayContext = 769
)

// enum
type UDialectHandling int32

const (
	ULDN_STANDARD_NAMES UDialectHandling = 0
	ULDN_DIALECT_NAMES  UDialectHandling = 1
)

// enum
type UCurrencyUsage int32

const (
	UCURR_USAGE_STANDARD UCurrencyUsage = 0
	UCURR_USAGE_CASH     UCurrencyUsage = 1
)

// enum
type UCurrNameStyle int32

const (
	UCURR_SYMBOL_NAME        UCurrNameStyle = 0
	UCURR_LONG_NAME          UCurrNameStyle = 1
	UCURR_NARROW_SYMBOL_NAME UCurrNameStyle = 2
)

// enum
type UCurrCurrencyType int32

const (
	UCURR_ALL            UCurrCurrencyType = 2147483647
	UCURR_COMMON         UCurrCurrencyType = 1
	UCURR_UNCOMMON       UCurrCurrencyType = 2
	UCURR_DEPRECATED     UCurrCurrencyType = 4
	UCURR_NON_DEPRECATED UCurrCurrencyType = 8
)

// enum
type UCPMapRangeOption int32

const (
	UCPMAP_RANGE_NORMAL                UCPMapRangeOption = 0
	UCPMAP_RANGE_FIXED_LEAD_SURROGATES UCPMapRangeOption = 1
	UCPMAP_RANGE_FIXED_ALL_SURROGATES  UCPMapRangeOption = 2
)

// enum
type UCPTrieType int32

const (
	UCPTRIE_TYPE_ANY   UCPTrieType = -1
	UCPTRIE_TYPE_FAST  UCPTrieType = 0
	UCPTRIE_TYPE_SMALL UCPTrieType = 1
)

// enum
type UCPTrieValueWidth int32

const (
	UCPTRIE_VALUE_BITS_ANY UCPTrieValueWidth = -1
	UCPTRIE_VALUE_BITS_16  UCPTrieValueWidth = 0
	UCPTRIE_VALUE_BITS_32  UCPTrieValueWidth = 1
	UCPTRIE_VALUE_BITS_8   UCPTrieValueWidth = 2
)

// enum
type UConverterCallbackReason int32

const (
	UCNV_UNASSIGNED UConverterCallbackReason = 0
	UCNV_ILLEGAL    UConverterCallbackReason = 1
	UCNV_IRREGULAR  UConverterCallbackReason = 2
	UCNV_RESET      UConverterCallbackReason = 3
	UCNV_CLOSE      UConverterCallbackReason = 4
	UCNV_CLONE      UConverterCallbackReason = 5
)

// enum
type UConverterType int32

const (
	UCNV_UNSUPPORTED_CONVERTER               UConverterType = -1
	UCNV_SBCS                                UConverterType = 0
	UCNV_DBCS                                UConverterType = 1
	UCNV_MBCS                                UConverterType = 2
	UCNV_LATIN_1                             UConverterType = 3
	UCNV_UTF8                                UConverterType = 4
	UCNV_UTF16_BigEndian                     UConverterType = 5
	UCNV_UTF16_LittleEndian                  UConverterType = 6
	UCNV_UTF32_BigEndian                     UConverterType = 7
	UCNV_UTF32_LittleEndian                  UConverterType = 8
	UCNV_EBCDIC_STATEFUL                     UConverterType = 9
	UCNV_ISO_2022                            UConverterType = 10
	UCNV_LMBCS_1                             UConverterType = 11
	UCNV_LMBCS_2                             UConverterType = 12
	UCNV_LMBCS_3                             UConverterType = 13
	UCNV_LMBCS_4                             UConverterType = 14
	UCNV_LMBCS_5                             UConverterType = 15
	UCNV_LMBCS_6                             UConverterType = 16
	UCNV_LMBCS_8                             UConverterType = 17
	UCNV_LMBCS_11                            UConverterType = 18
	UCNV_LMBCS_16                            UConverterType = 19
	UCNV_LMBCS_17                            UConverterType = 20
	UCNV_LMBCS_18                            UConverterType = 21
	UCNV_LMBCS_19                            UConverterType = 22
	UCNV_LMBCS_LAST                          UConverterType = 22
	UCNV_HZ                                  UConverterType = 23
	UCNV_SCSU                                UConverterType = 24
	UCNV_ISCII                               UConverterType = 25
	UCNV_US_ASCII                            UConverterType = 26
	UCNV_UTF7                                UConverterType = 27
	UCNV_BOCU1                               UConverterType = 28
	UCNV_UTF16                               UConverterType = 29
	UCNV_UTF32                               UConverterType = 30
	UCNV_CESU8                               UConverterType = 31
	UCNV_IMAP_MAILBOX                        UConverterType = 32
	UCNV_COMPOUND_TEXT                       UConverterType = 33
	UCNV_NUMBER_OF_SUPPORTED_CONVERTER_TYPES UConverterType = 34
)

// enum
type UConverterPlatform int32

const (
	UCNV_UNKNOWN UConverterPlatform = -1
	UCNV_IBM     UConverterPlatform = 0
)

// enum
type UConverterUnicodeSet int32

const (
	UCNV_ROUNDTRIP_SET              UConverterUnicodeSet = 0
	UCNV_ROUNDTRIP_AND_FALLBACK_SET UConverterUnicodeSet = 1
)

// enum
type UProperty int32

const (
	UCHAR_ALPHABETIC                      UProperty = 0
	UCHAR_BINARY_START                    UProperty = 0
	UCHAR_ASCII_HEX_DIGIT                 UProperty = 1
	UCHAR_BIDI_CONTROL                    UProperty = 2
	UCHAR_BIDI_MIRRORED                   UProperty = 3
	UCHAR_DASH                            UProperty = 4
	UCHAR_DEFAULT_IGNORABLE_CODE_POINT    UProperty = 5
	UCHAR_DEPRECATED                      UProperty = 6
	UCHAR_DIACRITIC                       UProperty = 7
	UCHAR_EXTENDER                        UProperty = 8
	UCHAR_FULL_COMPOSITION_EXCLUSION      UProperty = 9
	UCHAR_GRAPHEME_BASE                   UProperty = 10
	UCHAR_GRAPHEME_EXTEND                 UProperty = 11
	UCHAR_GRAPHEME_LINK                   UProperty = 12
	UCHAR_HEX_DIGIT                       UProperty = 13
	UCHAR_HYPHEN                          UProperty = 14
	UCHAR_ID_CONTINUE                     UProperty = 15
	UCHAR_ID_START                        UProperty = 16
	UCHAR_IDEOGRAPHIC                     UProperty = 17
	UCHAR_IDS_BINARY_OPERATOR             UProperty = 18
	UCHAR_IDS_TRINARY_OPERATOR            UProperty = 19
	UCHAR_JOIN_CONTROL                    UProperty = 20
	UCHAR_LOGICAL_ORDER_EXCEPTION         UProperty = 21
	UCHAR_LOWERCASE                       UProperty = 22
	UCHAR_MATH                            UProperty = 23
	UCHAR_NONCHARACTER_CODE_POINT         UProperty = 24
	UCHAR_QUOTATION_MARK                  UProperty = 25
	UCHAR_RADICAL                         UProperty = 26
	UCHAR_SOFT_DOTTED                     UProperty = 27
	UCHAR_TERMINAL_PUNCTUATION            UProperty = 28
	UCHAR_UNIFIED_IDEOGRAPH               UProperty = 29
	UCHAR_UPPERCASE                       UProperty = 30
	UCHAR_WHITE_SPACE                     UProperty = 31
	UCHAR_XID_CONTINUE                    UProperty = 32
	UCHAR_XID_START                       UProperty = 33
	UCHAR_CASE_SENSITIVE                  UProperty = 34
	UCHAR_S_TERM                          UProperty = 35
	UCHAR_VARIATION_SELECTOR              UProperty = 36
	UCHAR_NFD_INERT                       UProperty = 37
	UCHAR_NFKD_INERT                      UProperty = 38
	UCHAR_NFC_INERT                       UProperty = 39
	UCHAR_NFKC_INERT                      UProperty = 40
	UCHAR_SEGMENT_STARTER                 UProperty = 41
	UCHAR_PATTERN_SYNTAX                  UProperty = 42
	UCHAR_PATTERN_WHITE_SPACE             UProperty = 43
	UCHAR_POSIX_ALNUM                     UProperty = 44
	UCHAR_POSIX_BLANK                     UProperty = 45
	UCHAR_POSIX_GRAPH                     UProperty = 46
	UCHAR_POSIX_PRINT                     UProperty = 47
	UCHAR_POSIX_XDIGIT                    UProperty = 48
	UCHAR_CASED                           UProperty = 49
	UCHAR_CASE_IGNORABLE                  UProperty = 50
	UCHAR_CHANGES_WHEN_LOWERCASED         UProperty = 51
	UCHAR_CHANGES_WHEN_UPPERCASED         UProperty = 52
	UCHAR_CHANGES_WHEN_TITLECASED         UProperty = 53
	UCHAR_CHANGES_WHEN_CASEFOLDED         UProperty = 54
	UCHAR_CHANGES_WHEN_CASEMAPPED         UProperty = 55
	UCHAR_CHANGES_WHEN_NFKC_CASEFOLDED    UProperty = 56
	UCHAR_EMOJI                           UProperty = 57
	UCHAR_EMOJI_PRESENTATION              UProperty = 58
	UCHAR_EMOJI_MODIFIER                  UProperty = 59
	UCHAR_EMOJI_MODIFIER_BASE             UProperty = 60
	UCHAR_EMOJI_COMPONENT                 UProperty = 61
	UCHAR_REGIONAL_INDICATOR              UProperty = 62
	UCHAR_PREPENDED_CONCATENATION_MARK    UProperty = 63
	UCHAR_EXTENDED_PICTOGRAPHIC           UProperty = 64
	UCHAR_BIDI_CLASS                      UProperty = 4096
	UCHAR_INT_START                       UProperty = 4096
	UCHAR_BLOCK                           UProperty = 4097
	UCHAR_CANONICAL_COMBINING_CLASS       UProperty = 4098
	UCHAR_DECOMPOSITION_TYPE              UProperty = 4099
	UCHAR_EAST_ASIAN_WIDTH                UProperty = 4100
	UCHAR_GENERAL_CATEGORY                UProperty = 4101
	UCHAR_JOINING_GROUP                   UProperty = 4102
	UCHAR_JOINING_TYPE                    UProperty = 4103
	UCHAR_LINE_BREAK                      UProperty = 4104
	UCHAR_NUMERIC_TYPE                    UProperty = 4105
	UCHAR_SCRIPT                          UProperty = 4106
	UCHAR_HANGUL_SYLLABLE_TYPE            UProperty = 4107
	UCHAR_NFD_QUICK_CHECK                 UProperty = 4108
	UCHAR_NFKD_QUICK_CHECK                UProperty = 4109
	UCHAR_NFC_QUICK_CHECK                 UProperty = 4110
	UCHAR_NFKC_QUICK_CHECK                UProperty = 4111
	UCHAR_LEAD_CANONICAL_COMBINING_CLASS  UProperty = 4112
	UCHAR_TRAIL_CANONICAL_COMBINING_CLASS UProperty = 4113
	UCHAR_GRAPHEME_CLUSTER_BREAK          UProperty = 4114
	UCHAR_SENTENCE_BREAK                  UProperty = 4115
	UCHAR_WORD_BREAK                      UProperty = 4116
	UCHAR_BIDI_PAIRED_BRACKET_TYPE        UProperty = 4117
	UCHAR_INDIC_POSITIONAL_CATEGORY       UProperty = 4118
	UCHAR_INDIC_SYLLABIC_CATEGORY         UProperty = 4119
	UCHAR_VERTICAL_ORIENTATION            UProperty = 4120
	UCHAR_GENERAL_CATEGORY_MASK           UProperty = 8192
	UCHAR_MASK_START                      UProperty = 8192
	UCHAR_NUMERIC_VALUE                   UProperty = 12288
	UCHAR_DOUBLE_START                    UProperty = 12288
	UCHAR_AGE                             UProperty = 16384
	UCHAR_STRING_START                    UProperty = 16384
	UCHAR_BIDI_MIRRORING_GLYPH            UProperty = 16385
	UCHAR_CASE_FOLDING                    UProperty = 16386
	UCHAR_LOWERCASE_MAPPING               UProperty = 16388
	UCHAR_NAME                            UProperty = 16389
	UCHAR_SIMPLE_CASE_FOLDING             UProperty = 16390
	UCHAR_SIMPLE_LOWERCASE_MAPPING        UProperty = 16391
	UCHAR_SIMPLE_TITLECASE_MAPPING        UProperty = 16392
	UCHAR_SIMPLE_UPPERCASE_MAPPING        UProperty = 16393
	UCHAR_TITLECASE_MAPPING               UProperty = 16394
	UCHAR_UPPERCASE_MAPPING               UProperty = 16396
	UCHAR_BIDI_PAIRED_BRACKET             UProperty = 16397
	UCHAR_SCRIPT_EXTENSIONS               UProperty = 28672
	UCHAR_OTHER_PROPERTY_START            UProperty = 28672
	UCHAR_INVALID_CODE                    UProperty = -1
)

// enum
type UCharCategory int32

const (
	U_UNASSIGNED             UCharCategory = 0
	U_GENERAL_OTHER_TYPES    UCharCategory = 0
	U_UPPERCASE_LETTER       UCharCategory = 1
	U_LOWERCASE_LETTER       UCharCategory = 2
	U_TITLECASE_LETTER       UCharCategory = 3
	U_MODIFIER_LETTER        UCharCategory = 4
	U_OTHER_LETTER           UCharCategory = 5
	U_NON_SPACING_MARK       UCharCategory = 6
	U_ENCLOSING_MARK         UCharCategory = 7
	U_COMBINING_SPACING_MARK UCharCategory = 8
	U_DECIMAL_DIGIT_NUMBER   UCharCategory = 9
	U_LETTER_NUMBER          UCharCategory = 10
	U_OTHER_NUMBER           UCharCategory = 11
	U_SPACE_SEPARATOR        UCharCategory = 12
	U_LINE_SEPARATOR         UCharCategory = 13
	U_PARAGRAPH_SEPARATOR    UCharCategory = 14
	U_CONTROL_CHAR           UCharCategory = 15
	U_FORMAT_CHAR            UCharCategory = 16
	U_PRIVATE_USE_CHAR       UCharCategory = 17
	U_SURROGATE              UCharCategory = 18
	U_DASH_PUNCTUATION       UCharCategory = 19
	U_START_PUNCTUATION      UCharCategory = 20
	U_END_PUNCTUATION        UCharCategory = 21
	U_CONNECTOR_PUNCTUATION  UCharCategory = 22
	U_OTHER_PUNCTUATION      UCharCategory = 23
	U_MATH_SYMBOL            UCharCategory = 24
	U_CURRENCY_SYMBOL        UCharCategory = 25
	U_MODIFIER_SYMBOL        UCharCategory = 26
	U_OTHER_SYMBOL           UCharCategory = 27
	U_INITIAL_PUNCTUATION    UCharCategory = 28
	U_FINAL_PUNCTUATION      UCharCategory = 29
	U_CHAR_CATEGORY_COUNT    UCharCategory = 30
)

// enum
type UCharDirection int32

const (
	U_LEFT_TO_RIGHT              UCharDirection = 0
	U_RIGHT_TO_LEFT              UCharDirection = 1
	U_EUROPEAN_NUMBER            UCharDirection = 2
	U_EUROPEAN_NUMBER_SEPARATOR  UCharDirection = 3
	U_EUROPEAN_NUMBER_TERMINATOR UCharDirection = 4
	U_ARABIC_NUMBER              UCharDirection = 5
	U_COMMON_NUMBER_SEPARATOR    UCharDirection = 6
	U_BLOCK_SEPARATOR            UCharDirection = 7
	U_SEGMENT_SEPARATOR          UCharDirection = 8
	U_WHITE_SPACE_NEUTRAL        UCharDirection = 9
	U_OTHER_NEUTRAL              UCharDirection = 10
	U_LEFT_TO_RIGHT_EMBEDDING    UCharDirection = 11
	U_LEFT_TO_RIGHT_OVERRIDE     UCharDirection = 12
	U_RIGHT_TO_LEFT_ARABIC       UCharDirection = 13
	U_RIGHT_TO_LEFT_EMBEDDING    UCharDirection = 14
	U_RIGHT_TO_LEFT_OVERRIDE     UCharDirection = 15
	U_POP_DIRECTIONAL_FORMAT     UCharDirection = 16
	U_DIR_NON_SPACING_MARK       UCharDirection = 17
	U_BOUNDARY_NEUTRAL           UCharDirection = 18
	U_FIRST_STRONG_ISOLATE       UCharDirection = 19
	U_LEFT_TO_RIGHT_ISOLATE      UCharDirection = 20
	U_RIGHT_TO_LEFT_ISOLATE      UCharDirection = 21
	U_POP_DIRECTIONAL_ISOLATE    UCharDirection = 22
)

// enum
type UBidiPairedBracketType int32

const (
	U_BPT_NONE  UBidiPairedBracketType = 0
	U_BPT_OPEN  UBidiPairedBracketType = 1
	U_BPT_CLOSE UBidiPairedBracketType = 2
)

// enum
type UBlockCode int32

const (
	UBLOCK_NO_BLOCK                                       UBlockCode = 0
	UBLOCK_BASIC_LATIN                                    UBlockCode = 1
	UBLOCK_LATIN_1_SUPPLEMENT                             UBlockCode = 2
	UBLOCK_LATIN_EXTENDED_A                               UBlockCode = 3
	UBLOCK_LATIN_EXTENDED_B                               UBlockCode = 4
	UBLOCK_IPA_EXTENSIONS                                 UBlockCode = 5
	UBLOCK_SPACING_MODIFIER_LETTERS                       UBlockCode = 6
	UBLOCK_COMBINING_DIACRITICAL_MARKS                    UBlockCode = 7
	UBLOCK_GREEK                                          UBlockCode = 8
	UBLOCK_CYRILLIC                                       UBlockCode = 9
	UBLOCK_ARMENIAN                                       UBlockCode = 10
	UBLOCK_HEBREW                                         UBlockCode = 11
	UBLOCK_ARABIC                                         UBlockCode = 12
	UBLOCK_SYRIAC                                         UBlockCode = 13
	UBLOCK_THAANA                                         UBlockCode = 14
	UBLOCK_DEVANAGARI                                     UBlockCode = 15
	UBLOCK_BENGALI                                        UBlockCode = 16
	UBLOCK_GURMUKHI                                       UBlockCode = 17
	UBLOCK_GUJARATI                                       UBlockCode = 18
	UBLOCK_ORIYA                                          UBlockCode = 19
	UBLOCK_TAMIL                                          UBlockCode = 20
	UBLOCK_TELUGU                                         UBlockCode = 21
	UBLOCK_KANNADA                                        UBlockCode = 22
	UBLOCK_MALAYALAM                                      UBlockCode = 23
	UBLOCK_SINHALA                                        UBlockCode = 24
	UBLOCK_THAI                                           UBlockCode = 25
	UBLOCK_LAO                                            UBlockCode = 26
	UBLOCK_TIBETAN                                        UBlockCode = 27
	UBLOCK_MYANMAR                                        UBlockCode = 28
	UBLOCK_GEORGIAN                                       UBlockCode = 29
	UBLOCK_HANGUL_JAMO                                    UBlockCode = 30
	UBLOCK_ETHIOPIC                                       UBlockCode = 31
	UBLOCK_CHEROKEE                                       UBlockCode = 32
	UBLOCK_UNIFIED_CANADIAN_ABORIGINAL_SYLLABICS          UBlockCode = 33
	UBLOCK_OGHAM                                          UBlockCode = 34
	UBLOCK_RUNIC                                          UBlockCode = 35
	UBLOCK_KHMER                                          UBlockCode = 36
	UBLOCK_MONGOLIAN                                      UBlockCode = 37
	UBLOCK_LATIN_EXTENDED_ADDITIONAL                      UBlockCode = 38
	UBLOCK_GREEK_EXTENDED                                 UBlockCode = 39
	UBLOCK_GENERAL_PUNCTUATION                            UBlockCode = 40
	UBLOCK_SUPERSCRIPTS_AND_SUBSCRIPTS                    UBlockCode = 41
	UBLOCK_CURRENCY_SYMBOLS                               UBlockCode = 42
	UBLOCK_COMBINING_MARKS_FOR_SYMBOLS                    UBlockCode = 43
	UBLOCK_LETTERLIKE_SYMBOLS                             UBlockCode = 44
	UBLOCK_NUMBER_FORMS                                   UBlockCode = 45
	UBLOCK_ARROWS                                         UBlockCode = 46
	UBLOCK_MATHEMATICAL_OPERATORS                         UBlockCode = 47
	UBLOCK_MISCELLANEOUS_TECHNICAL                        UBlockCode = 48
	UBLOCK_CONTROL_PICTURES                               UBlockCode = 49
	UBLOCK_OPTICAL_CHARACTER_RECOGNITION                  UBlockCode = 50
	UBLOCK_ENCLOSED_ALPHANUMERICS                         UBlockCode = 51
	UBLOCK_BOX_DRAWING                                    UBlockCode = 52
	UBLOCK_BLOCK_ELEMENTS                                 UBlockCode = 53
	UBLOCK_GEOMETRIC_SHAPES                               UBlockCode = 54
	UBLOCK_MISCELLANEOUS_SYMBOLS                          UBlockCode = 55
	UBLOCK_DINGBATS                                       UBlockCode = 56
	UBLOCK_BRAILLE_PATTERNS                               UBlockCode = 57
	UBLOCK_CJK_RADICALS_SUPPLEMENT                        UBlockCode = 58
	UBLOCK_KANGXI_RADICALS                                UBlockCode = 59
	UBLOCK_IDEOGRAPHIC_DESCRIPTION_CHARACTERS             UBlockCode = 60
	UBLOCK_CJK_SYMBOLS_AND_PUNCTUATION                    UBlockCode = 61
	UBLOCK_HIRAGANA                                       UBlockCode = 62
	UBLOCK_KATAKANA                                       UBlockCode = 63
	UBLOCK_BOPOMOFO                                       UBlockCode = 64
	UBLOCK_HANGUL_COMPATIBILITY_JAMO                      UBlockCode = 65
	UBLOCK_KANBUN                                         UBlockCode = 66
	UBLOCK_BOPOMOFO_EXTENDED                              UBlockCode = 67
	UBLOCK_ENCLOSED_CJK_LETTERS_AND_MONTHS                UBlockCode = 68
	UBLOCK_CJK_COMPATIBILITY                              UBlockCode = 69
	UBLOCK_CJK_UNIFIED_IDEOGRAPHS_EXTENSION_A             UBlockCode = 70
	UBLOCK_CJK_UNIFIED_IDEOGRAPHS                         UBlockCode = 71
	UBLOCK_YI_SYLLABLES                                   UBlockCode = 72
	UBLOCK_YI_RADICALS                                    UBlockCode = 73
	UBLOCK_HANGUL_SYLLABLES                               UBlockCode = 74
	UBLOCK_HIGH_SURROGATES                                UBlockCode = 75
	UBLOCK_HIGH_PRIVATE_USE_SURROGATES                    UBlockCode = 76
	UBLOCK_LOW_SURROGATES                                 UBlockCode = 77
	UBLOCK_PRIVATE_USE_AREA                               UBlockCode = 78
	UBLOCK_PRIVATE_USE                                    UBlockCode = 78
	UBLOCK_CJK_COMPATIBILITY_IDEOGRAPHS                   UBlockCode = 79
	UBLOCK_ALPHABETIC_PRESENTATION_FORMS                  UBlockCode = 80
	UBLOCK_ARABIC_PRESENTATION_FORMS_A                    UBlockCode = 81
	UBLOCK_COMBINING_HALF_MARKS                           UBlockCode = 82
	UBLOCK_CJK_COMPATIBILITY_FORMS                        UBlockCode = 83
	UBLOCK_SMALL_FORM_VARIANTS                            UBlockCode = 84
	UBLOCK_ARABIC_PRESENTATION_FORMS_B                    UBlockCode = 85
	UBLOCK_SPECIALS                                       UBlockCode = 86
	UBLOCK_HALFWIDTH_AND_FULLWIDTH_FORMS                  UBlockCode = 87
	UBLOCK_OLD_ITALIC                                     UBlockCode = 88
	UBLOCK_GOTHIC                                         UBlockCode = 89
	UBLOCK_DESERET                                        UBlockCode = 90
	UBLOCK_BYZANTINE_MUSICAL_SYMBOLS                      UBlockCode = 91
	UBLOCK_MUSICAL_SYMBOLS                                UBlockCode = 92
	UBLOCK_MATHEMATICAL_ALPHANUMERIC_SYMBOLS              UBlockCode = 93
	UBLOCK_CJK_UNIFIED_IDEOGRAPHS_EXTENSION_B             UBlockCode = 94
	UBLOCK_CJK_COMPATIBILITY_IDEOGRAPHS_SUPPLEMENT        UBlockCode = 95
	UBLOCK_TAGS                                           UBlockCode = 96
	UBLOCK_CYRILLIC_SUPPLEMENT                            UBlockCode = 97
	UBLOCK_CYRILLIC_SUPPLEMENTARY                         UBlockCode = 97
	UBLOCK_TAGALOG                                        UBlockCode = 98
	UBLOCK_HANUNOO                                        UBlockCode = 99
	UBLOCK_BUHID                                          UBlockCode = 100
	UBLOCK_TAGBANWA                                       UBlockCode = 101
	UBLOCK_MISCELLANEOUS_MATHEMATICAL_SYMBOLS_A           UBlockCode = 102
	UBLOCK_SUPPLEMENTAL_ARROWS_A                          UBlockCode = 103
	UBLOCK_SUPPLEMENTAL_ARROWS_B                          UBlockCode = 104
	UBLOCK_MISCELLANEOUS_MATHEMATICAL_SYMBOLS_B           UBlockCode = 105
	UBLOCK_SUPPLEMENTAL_MATHEMATICAL_OPERATORS            UBlockCode = 106
	UBLOCK_KATAKANA_PHONETIC_EXTENSIONS                   UBlockCode = 107
	UBLOCK_VARIATION_SELECTORS                            UBlockCode = 108
	UBLOCK_SUPPLEMENTARY_PRIVATE_USE_AREA_A               UBlockCode = 109
	UBLOCK_SUPPLEMENTARY_PRIVATE_USE_AREA_B               UBlockCode = 110
	UBLOCK_LIMBU                                          UBlockCode = 111
	UBLOCK_TAI_LE                                         UBlockCode = 112
	UBLOCK_KHMER_SYMBOLS                                  UBlockCode = 113
	UBLOCK_PHONETIC_EXTENSIONS                            UBlockCode = 114
	UBLOCK_MISCELLANEOUS_SYMBOLS_AND_ARROWS               UBlockCode = 115
	UBLOCK_YIJING_HEXAGRAM_SYMBOLS                        UBlockCode = 116
	UBLOCK_LINEAR_B_SYLLABARY                             UBlockCode = 117
	UBLOCK_LINEAR_B_IDEOGRAMS                             UBlockCode = 118
	UBLOCK_AEGEAN_NUMBERS                                 UBlockCode = 119
	UBLOCK_UGARITIC                                       UBlockCode = 120
	UBLOCK_SHAVIAN                                        UBlockCode = 121
	UBLOCK_OSMANYA                                        UBlockCode = 122
	UBLOCK_CYPRIOT_SYLLABARY                              UBlockCode = 123
	UBLOCK_TAI_XUAN_JING_SYMBOLS                          UBlockCode = 124
	UBLOCK_VARIATION_SELECTORS_SUPPLEMENT                 UBlockCode = 125
	UBLOCK_ANCIENT_GREEK_MUSICAL_NOTATION                 UBlockCode = 126
	UBLOCK_ANCIENT_GREEK_NUMBERS                          UBlockCode = 127
	UBLOCK_ARABIC_SUPPLEMENT                              UBlockCode = 128
	UBLOCK_BUGINESE                                       UBlockCode = 129
	UBLOCK_CJK_STROKES                                    UBlockCode = 130
	UBLOCK_COMBINING_DIACRITICAL_MARKS_SUPPLEMENT         UBlockCode = 131
	UBLOCK_COPTIC                                         UBlockCode = 132
	UBLOCK_ETHIOPIC_EXTENDED                              UBlockCode = 133
	UBLOCK_ETHIOPIC_SUPPLEMENT                            UBlockCode = 134
	UBLOCK_GEORGIAN_SUPPLEMENT                            UBlockCode = 135
	UBLOCK_GLAGOLITIC                                     UBlockCode = 136
	UBLOCK_KHAROSHTHI                                     UBlockCode = 137
	UBLOCK_MODIFIER_TONE_LETTERS                          UBlockCode = 138
	UBLOCK_NEW_TAI_LUE                                    UBlockCode = 139
	UBLOCK_OLD_PERSIAN                                    UBlockCode = 140
	UBLOCK_PHONETIC_EXTENSIONS_SUPPLEMENT                 UBlockCode = 141
	UBLOCK_SUPPLEMENTAL_PUNCTUATION                       UBlockCode = 142
	UBLOCK_SYLOTI_NAGRI                                   UBlockCode = 143
	UBLOCK_TIFINAGH                                       UBlockCode = 144
	UBLOCK_VERTICAL_FORMS                                 UBlockCode = 145
	UBLOCK_NKO                                            UBlockCode = 146
	UBLOCK_BALINESE                                       UBlockCode = 147
	UBLOCK_LATIN_EXTENDED_C                               UBlockCode = 148
	UBLOCK_LATIN_EXTENDED_D                               UBlockCode = 149
	UBLOCK_PHAGS_PA                                       UBlockCode = 150
	UBLOCK_PHOENICIAN                                     UBlockCode = 151
	UBLOCK_CUNEIFORM                                      UBlockCode = 152
	UBLOCK_CUNEIFORM_NUMBERS_AND_PUNCTUATION              UBlockCode = 153
	UBLOCK_COUNTING_ROD_NUMERALS                          UBlockCode = 154
	UBLOCK_SUNDANESE                                      UBlockCode = 155
	UBLOCK_LEPCHA                                         UBlockCode = 156
	UBLOCK_OL_CHIKI                                       UBlockCode = 157
	UBLOCK_CYRILLIC_EXTENDED_A                            UBlockCode = 158
	UBLOCK_VAI                                            UBlockCode = 159
	UBLOCK_CYRILLIC_EXTENDED_B                            UBlockCode = 160
	UBLOCK_SAURASHTRA                                     UBlockCode = 161
	UBLOCK_KAYAH_LI                                       UBlockCode = 162
	UBLOCK_REJANG                                         UBlockCode = 163
	UBLOCK_CHAM                                           UBlockCode = 164
	UBLOCK_ANCIENT_SYMBOLS                                UBlockCode = 165
	UBLOCK_PHAISTOS_DISC                                  UBlockCode = 166
	UBLOCK_LYCIAN                                         UBlockCode = 167
	UBLOCK_CARIAN                                         UBlockCode = 168
	UBLOCK_LYDIAN                                         UBlockCode = 169
	UBLOCK_MAHJONG_TILES                                  UBlockCode = 170
	UBLOCK_DOMINO_TILES                                   UBlockCode = 171
	UBLOCK_SAMARITAN                                      UBlockCode = 172
	UBLOCK_UNIFIED_CANADIAN_ABORIGINAL_SYLLABICS_EXTENDED UBlockCode = 173
	UBLOCK_TAI_THAM                                       UBlockCode = 174
	UBLOCK_VEDIC_EXTENSIONS                               UBlockCode = 175
	UBLOCK_LISU                                           UBlockCode = 176
	UBLOCK_BAMUM                                          UBlockCode = 177
	UBLOCK_COMMON_INDIC_NUMBER_FORMS                      UBlockCode = 178
	UBLOCK_DEVANAGARI_EXTENDED                            UBlockCode = 179
	UBLOCK_HANGUL_JAMO_EXTENDED_A                         UBlockCode = 180
	UBLOCK_JAVANESE                                       UBlockCode = 181
	UBLOCK_MYANMAR_EXTENDED_A                             UBlockCode = 182
	UBLOCK_TAI_VIET                                       UBlockCode = 183
	UBLOCK_MEETEI_MAYEK                                   UBlockCode = 184
	UBLOCK_HANGUL_JAMO_EXTENDED_B                         UBlockCode = 185
	UBLOCK_IMPERIAL_ARAMAIC                               UBlockCode = 186
	UBLOCK_OLD_SOUTH_ARABIAN                              UBlockCode = 187
	UBLOCK_AVESTAN                                        UBlockCode = 188
	UBLOCK_INSCRIPTIONAL_PARTHIAN                         UBlockCode = 189
	UBLOCK_INSCRIPTIONAL_PAHLAVI                          UBlockCode = 190
	UBLOCK_OLD_TURKIC                                     UBlockCode = 191
	UBLOCK_RUMI_NUMERAL_SYMBOLS                           UBlockCode = 192
	UBLOCK_KAITHI                                         UBlockCode = 193
	UBLOCK_EGYPTIAN_HIEROGLYPHS                           UBlockCode = 194
	UBLOCK_ENCLOSED_ALPHANUMERIC_SUPPLEMENT               UBlockCode = 195
	UBLOCK_ENCLOSED_IDEOGRAPHIC_SUPPLEMENT                UBlockCode = 196
	UBLOCK_CJK_UNIFIED_IDEOGRAPHS_EXTENSION_C             UBlockCode = 197
	UBLOCK_MANDAIC                                        UBlockCode = 198
	UBLOCK_BATAK                                          UBlockCode = 199
	UBLOCK_ETHIOPIC_EXTENDED_A                            UBlockCode = 200
	UBLOCK_BRAHMI                                         UBlockCode = 201
	UBLOCK_BAMUM_SUPPLEMENT                               UBlockCode = 202
	UBLOCK_KANA_SUPPLEMENT                                UBlockCode = 203
	UBLOCK_PLAYING_CARDS                                  UBlockCode = 204
	UBLOCK_MISCELLANEOUS_SYMBOLS_AND_PICTOGRAPHS          UBlockCode = 205
	UBLOCK_EMOTICONS                                      UBlockCode = 206
	UBLOCK_TRANSPORT_AND_MAP_SYMBOLS                      UBlockCode = 207
	UBLOCK_ALCHEMICAL_SYMBOLS                             UBlockCode = 208
	UBLOCK_CJK_UNIFIED_IDEOGRAPHS_EXTENSION_D             UBlockCode = 209
	UBLOCK_ARABIC_EXTENDED_A                              UBlockCode = 210
	UBLOCK_ARABIC_MATHEMATICAL_ALPHABETIC_SYMBOLS         UBlockCode = 211
	UBLOCK_CHAKMA                                         UBlockCode = 212
	UBLOCK_MEETEI_MAYEK_EXTENSIONS                        UBlockCode = 213
	UBLOCK_MEROITIC_CURSIVE                               UBlockCode = 214
	UBLOCK_MEROITIC_HIEROGLYPHS                           UBlockCode = 215
	UBLOCK_MIAO                                           UBlockCode = 216
	UBLOCK_SHARADA                                        UBlockCode = 217
	UBLOCK_SORA_SOMPENG                                   UBlockCode = 218
	UBLOCK_SUNDANESE_SUPPLEMENT                           UBlockCode = 219
	UBLOCK_TAKRI                                          UBlockCode = 220
	UBLOCK_BASSA_VAH                                      UBlockCode = 221
	UBLOCK_CAUCASIAN_ALBANIAN                             UBlockCode = 222
	UBLOCK_COPTIC_EPACT_NUMBERS                           UBlockCode = 223
	UBLOCK_COMBINING_DIACRITICAL_MARKS_EXTENDED           UBlockCode = 224
	UBLOCK_DUPLOYAN                                       UBlockCode = 225
	UBLOCK_ELBASAN                                        UBlockCode = 226
	UBLOCK_GEOMETRIC_SHAPES_EXTENDED                      UBlockCode = 227
	UBLOCK_GRANTHA                                        UBlockCode = 228
	UBLOCK_KHOJKI                                         UBlockCode = 229
	UBLOCK_KHUDAWADI                                      UBlockCode = 230
	UBLOCK_LATIN_EXTENDED_E                               UBlockCode = 231
	UBLOCK_LINEAR_A                                       UBlockCode = 232
	UBLOCK_MAHAJANI                                       UBlockCode = 233
	UBLOCK_MANICHAEAN                                     UBlockCode = 234
	UBLOCK_MENDE_KIKAKUI                                  UBlockCode = 235
	UBLOCK_MODI                                           UBlockCode = 236
	UBLOCK_MRO                                            UBlockCode = 237
	UBLOCK_MYANMAR_EXTENDED_B                             UBlockCode = 238
	UBLOCK_NABATAEAN                                      UBlockCode = 239
	UBLOCK_OLD_NORTH_ARABIAN                              UBlockCode = 240
	UBLOCK_OLD_PERMIC                                     UBlockCode = 241
	UBLOCK_ORNAMENTAL_DINGBATS                            UBlockCode = 242
	UBLOCK_PAHAWH_HMONG                                   UBlockCode = 243
	UBLOCK_PALMYRENE                                      UBlockCode = 244
	UBLOCK_PAU_CIN_HAU                                    UBlockCode = 245
	UBLOCK_PSALTER_PAHLAVI                                UBlockCode = 246
	UBLOCK_SHORTHAND_FORMAT_CONTROLS                      UBlockCode = 247
	UBLOCK_SIDDHAM                                        UBlockCode = 248
	UBLOCK_SINHALA_ARCHAIC_NUMBERS                        UBlockCode = 249
	UBLOCK_SUPPLEMENTAL_ARROWS_C                          UBlockCode = 250
	UBLOCK_TIRHUTA                                        UBlockCode = 251
	UBLOCK_WARANG_CITI                                    UBlockCode = 252
	UBLOCK_AHOM                                           UBlockCode = 253
	UBLOCK_ANATOLIAN_HIEROGLYPHS                          UBlockCode = 254
	UBLOCK_CHEROKEE_SUPPLEMENT                            UBlockCode = 255
	UBLOCK_CJK_UNIFIED_IDEOGRAPHS_EXTENSION_E             UBlockCode = 256
	UBLOCK_EARLY_DYNASTIC_CUNEIFORM                       UBlockCode = 257
	UBLOCK_HATRAN                                         UBlockCode = 258
	UBLOCK_MULTANI                                        UBlockCode = 259
	UBLOCK_OLD_HUNGARIAN                                  UBlockCode = 260
	UBLOCK_SUPPLEMENTAL_SYMBOLS_AND_PICTOGRAPHS           UBlockCode = 261
	UBLOCK_SUTTON_SIGNWRITING                             UBlockCode = 262
	UBLOCK_ADLAM                                          UBlockCode = 263
	UBLOCK_BHAIKSUKI                                      UBlockCode = 264
	UBLOCK_CYRILLIC_EXTENDED_C                            UBlockCode = 265
	UBLOCK_GLAGOLITIC_SUPPLEMENT                          UBlockCode = 266
	UBLOCK_IDEOGRAPHIC_SYMBOLS_AND_PUNCTUATION            UBlockCode = 267
	UBLOCK_MARCHEN                                        UBlockCode = 268
	UBLOCK_MONGOLIAN_SUPPLEMENT                           UBlockCode = 269
	UBLOCK_NEWA                                           UBlockCode = 270
	UBLOCK_OSAGE                                          UBlockCode = 271
	UBLOCK_TANGUT                                         UBlockCode = 272
	UBLOCK_TANGUT_COMPONENTS                              UBlockCode = 273
	UBLOCK_CJK_UNIFIED_IDEOGRAPHS_EXTENSION_F             UBlockCode = 274
	UBLOCK_KANA_EXTENDED_A                                UBlockCode = 275
	UBLOCK_MASARAM_GONDI                                  UBlockCode = 276
	UBLOCK_NUSHU                                          UBlockCode = 277
	UBLOCK_SOYOMBO                                        UBlockCode = 278
	UBLOCK_SYRIAC_SUPPLEMENT                              UBlockCode = 279
	UBLOCK_ZANABAZAR_SQUARE                               UBlockCode = 280
	UBLOCK_CHESS_SYMBOLS                                  UBlockCode = 281
	UBLOCK_DOGRA                                          UBlockCode = 282
	UBLOCK_GEORGIAN_EXTENDED                              UBlockCode = 283
	UBLOCK_GUNJALA_GONDI                                  UBlockCode = 284
	UBLOCK_HANIFI_ROHINGYA                                UBlockCode = 285
	UBLOCK_INDIC_SIYAQ_NUMBERS                            UBlockCode = 286
	UBLOCK_MAKASAR                                        UBlockCode = 287
	UBLOCK_MAYAN_NUMERALS                                 UBlockCode = 288
	UBLOCK_MEDEFAIDRIN                                    UBlockCode = 289
	UBLOCK_OLD_SOGDIAN                                    UBlockCode = 290
	UBLOCK_SOGDIAN                                        UBlockCode = 291
	UBLOCK_EGYPTIAN_HIEROGLYPH_FORMAT_CONTROLS            UBlockCode = 292
	UBLOCK_ELYMAIC                                        UBlockCode = 293
	UBLOCK_NANDINAGARI                                    UBlockCode = 294
	UBLOCK_NYIAKENG_PUACHUE_HMONG                         UBlockCode = 295
	UBLOCK_OTTOMAN_SIYAQ_NUMBERS                          UBlockCode = 296
	UBLOCK_SMALL_KANA_EXTENSION                           UBlockCode = 297
	UBLOCK_SYMBOLS_AND_PICTOGRAPHS_EXTENDED_A             UBlockCode = 298
	UBLOCK_TAMIL_SUPPLEMENT                               UBlockCode = 299
	UBLOCK_WANCHO                                         UBlockCode = 300
	UBLOCK_CHORASMIAN                                     UBlockCode = 301
	UBLOCK_CJK_UNIFIED_IDEOGRAPHS_EXTENSION_G             UBlockCode = 302
	UBLOCK_DIVES_AKURU                                    UBlockCode = 303
	UBLOCK_KHITAN_SMALL_SCRIPT                            UBlockCode = 304
	UBLOCK_LISU_SUPPLEMENT                                UBlockCode = 305
	UBLOCK_SYMBOLS_FOR_LEGACY_COMPUTING                   UBlockCode = 306
	UBLOCK_TANGUT_SUPPLEMENT                              UBlockCode = 307
	UBLOCK_YEZIDI                                         UBlockCode = 308
	UBLOCK_INVALID_CODE                                   UBlockCode = -1
)

// enum
type UEastAsianWidth int32

const (
	U_EA_NEUTRAL   UEastAsianWidth = 0
	U_EA_AMBIGUOUS UEastAsianWidth = 1
	U_EA_HALFWIDTH UEastAsianWidth = 2
	U_EA_FULLWIDTH UEastAsianWidth = 3
	U_EA_NARROW    UEastAsianWidth = 4
	U_EA_WIDE      UEastAsianWidth = 5
)

// enum
type UCharNameChoice int32

const (
	U_UNICODE_CHAR_NAME  UCharNameChoice = 0
	U_EXTENDED_CHAR_NAME UCharNameChoice = 2
	U_CHAR_NAME_ALIAS    UCharNameChoice = 3
)

// enum
type UPropertyNameChoice int32

const (
	U_SHORT_PROPERTY_NAME UPropertyNameChoice = 0
	U_LONG_PROPERTY_NAME  UPropertyNameChoice = 1
)

// enum
type UDecompositionType int32

const (
	U_DT_NONE      UDecompositionType = 0
	U_DT_CANONICAL UDecompositionType = 1
	U_DT_COMPAT    UDecompositionType = 2
	U_DT_CIRCLE    UDecompositionType = 3
	U_DT_FINAL     UDecompositionType = 4
	U_DT_FONT      UDecompositionType = 5
	U_DT_FRACTION  UDecompositionType = 6
	U_DT_INITIAL   UDecompositionType = 7
	U_DT_ISOLATED  UDecompositionType = 8
	U_DT_MEDIAL    UDecompositionType = 9
	U_DT_NARROW    UDecompositionType = 10
	U_DT_NOBREAK   UDecompositionType = 11
	U_DT_SMALL     UDecompositionType = 12
	U_DT_SQUARE    UDecompositionType = 13
	U_DT_SUB       UDecompositionType = 14
	U_DT_SUPER     UDecompositionType = 15
	U_DT_VERTICAL  UDecompositionType = 16
	U_DT_WIDE      UDecompositionType = 17
)

// enum
type UJoiningType int32

const (
	U_JT_NON_JOINING   UJoiningType = 0
	U_JT_JOIN_CAUSING  UJoiningType = 1
	U_JT_DUAL_JOINING  UJoiningType = 2
	U_JT_LEFT_JOINING  UJoiningType = 3
	U_JT_RIGHT_JOINING UJoiningType = 4
	U_JT_TRANSPARENT   UJoiningType = 5
)

// enum
type UJoiningGroup int32

const (
	U_JG_NO_JOINING_GROUP         UJoiningGroup = 0
	U_JG_AIN                      UJoiningGroup = 1
	U_JG_ALAPH                    UJoiningGroup = 2
	U_JG_ALEF                     UJoiningGroup = 3
	U_JG_BEH                      UJoiningGroup = 4
	U_JG_BETH                     UJoiningGroup = 5
	U_JG_DAL                      UJoiningGroup = 6
	U_JG_DALATH_RISH              UJoiningGroup = 7
	U_JG_E                        UJoiningGroup = 8
	U_JG_FEH                      UJoiningGroup = 9
	U_JG_FINAL_SEMKATH            UJoiningGroup = 10
	U_JG_GAF                      UJoiningGroup = 11
	U_JG_GAMAL                    UJoiningGroup = 12
	U_JG_HAH                      UJoiningGroup = 13
	U_JG_TEH_MARBUTA_GOAL         UJoiningGroup = 14
	U_JG_HAMZA_ON_HEH_GOAL        UJoiningGroup = 14
	U_JG_HE                       UJoiningGroup = 15
	U_JG_HEH                      UJoiningGroup = 16
	U_JG_HEH_GOAL                 UJoiningGroup = 17
	U_JG_HETH                     UJoiningGroup = 18
	U_JG_KAF                      UJoiningGroup = 19
	U_JG_KAPH                     UJoiningGroup = 20
	U_JG_KNOTTED_HEH              UJoiningGroup = 21
	U_JG_LAM                      UJoiningGroup = 22
	U_JG_LAMADH                   UJoiningGroup = 23
	U_JG_MEEM                     UJoiningGroup = 24
	U_JG_MIM                      UJoiningGroup = 25
	U_JG_NOON                     UJoiningGroup = 26
	U_JG_NUN                      UJoiningGroup = 27
	U_JG_PE                       UJoiningGroup = 28
	U_JG_QAF                      UJoiningGroup = 29
	U_JG_QAPH                     UJoiningGroup = 30
	U_JG_REH                      UJoiningGroup = 31
	U_JG_REVERSED_PE              UJoiningGroup = 32
	U_JG_SAD                      UJoiningGroup = 33
	U_JG_SADHE                    UJoiningGroup = 34
	U_JG_SEEN                     UJoiningGroup = 35
	U_JG_SEMKATH                  UJoiningGroup = 36
	U_JG_SHIN                     UJoiningGroup = 37
	U_JG_SWASH_KAF                UJoiningGroup = 38
	U_JG_SYRIAC_WAW               UJoiningGroup = 39
	U_JG_TAH                      UJoiningGroup = 40
	U_JG_TAW                      UJoiningGroup = 41
	U_JG_TEH_MARBUTA              UJoiningGroup = 42
	U_JG_TETH                     UJoiningGroup = 43
	U_JG_WAW                      UJoiningGroup = 44
	U_JG_YEH                      UJoiningGroup = 45
	U_JG_YEH_BARREE               UJoiningGroup = 46
	U_JG_YEH_WITH_TAIL            UJoiningGroup = 47
	U_JG_YUDH                     UJoiningGroup = 48
	U_JG_YUDH_HE                  UJoiningGroup = 49
	U_JG_ZAIN                     UJoiningGroup = 50
	U_JG_FE                       UJoiningGroup = 51
	U_JG_KHAPH                    UJoiningGroup = 52
	U_JG_ZHAIN                    UJoiningGroup = 53
	U_JG_BURUSHASKI_YEH_BARREE    UJoiningGroup = 54
	U_JG_FARSI_YEH                UJoiningGroup = 55
	U_JG_NYA                      UJoiningGroup = 56
	U_JG_ROHINGYA_YEH             UJoiningGroup = 57
	U_JG_MANICHAEAN_ALEPH         UJoiningGroup = 58
	U_JG_MANICHAEAN_AYIN          UJoiningGroup = 59
	U_JG_MANICHAEAN_BETH          UJoiningGroup = 60
	U_JG_MANICHAEAN_DALETH        UJoiningGroup = 61
	U_JG_MANICHAEAN_DHAMEDH       UJoiningGroup = 62
	U_JG_MANICHAEAN_FIVE          UJoiningGroup = 63
	U_JG_MANICHAEAN_GIMEL         UJoiningGroup = 64
	U_JG_MANICHAEAN_HETH          UJoiningGroup = 65
	U_JG_MANICHAEAN_HUNDRED       UJoiningGroup = 66
	U_JG_MANICHAEAN_KAPH          UJoiningGroup = 67
	U_JG_MANICHAEAN_LAMEDH        UJoiningGroup = 68
	U_JG_MANICHAEAN_MEM           UJoiningGroup = 69
	U_JG_MANICHAEAN_NUN           UJoiningGroup = 70
	U_JG_MANICHAEAN_ONE           UJoiningGroup = 71
	U_JG_MANICHAEAN_PE            UJoiningGroup = 72
	U_JG_MANICHAEAN_QOPH          UJoiningGroup = 73
	U_JG_MANICHAEAN_RESH          UJoiningGroup = 74
	U_JG_MANICHAEAN_SADHE         UJoiningGroup = 75
	U_JG_MANICHAEAN_SAMEKH        UJoiningGroup = 76
	U_JG_MANICHAEAN_TAW           UJoiningGroup = 77
	U_JG_MANICHAEAN_TEN           UJoiningGroup = 78
	U_JG_MANICHAEAN_TETH          UJoiningGroup = 79
	U_JG_MANICHAEAN_THAMEDH       UJoiningGroup = 80
	U_JG_MANICHAEAN_TWENTY        UJoiningGroup = 81
	U_JG_MANICHAEAN_WAW           UJoiningGroup = 82
	U_JG_MANICHAEAN_YODH          UJoiningGroup = 83
	U_JG_MANICHAEAN_ZAYIN         UJoiningGroup = 84
	U_JG_STRAIGHT_WAW             UJoiningGroup = 85
	U_JG_AFRICAN_FEH              UJoiningGroup = 86
	U_JG_AFRICAN_NOON             UJoiningGroup = 87
	U_JG_AFRICAN_QAF              UJoiningGroup = 88
	U_JG_MALAYALAM_BHA            UJoiningGroup = 89
	U_JG_MALAYALAM_JA             UJoiningGroup = 90
	U_JG_MALAYALAM_LLA            UJoiningGroup = 91
	U_JG_MALAYALAM_LLLA           UJoiningGroup = 92
	U_JG_MALAYALAM_NGA            UJoiningGroup = 93
	U_JG_MALAYALAM_NNA            UJoiningGroup = 94
	U_JG_MALAYALAM_NNNA           UJoiningGroup = 95
	U_JG_MALAYALAM_NYA            UJoiningGroup = 96
	U_JG_MALAYALAM_RA             UJoiningGroup = 97
	U_JG_MALAYALAM_SSA            UJoiningGroup = 98
	U_JG_MALAYALAM_TTA            UJoiningGroup = 99
	U_JG_HANIFI_ROHINGYA_KINNA_YA UJoiningGroup = 100
	U_JG_HANIFI_ROHINGYA_PA       UJoiningGroup = 101
)

// enum
type UGraphemeClusterBreak int32

const (
	U_GCB_OTHER              UGraphemeClusterBreak = 0
	U_GCB_CONTROL            UGraphemeClusterBreak = 1
	U_GCB_CR                 UGraphemeClusterBreak = 2
	U_GCB_EXTEND             UGraphemeClusterBreak = 3
	U_GCB_L                  UGraphemeClusterBreak = 4
	U_GCB_LF                 UGraphemeClusterBreak = 5
	U_GCB_LV                 UGraphemeClusterBreak = 6
	U_GCB_LVT                UGraphemeClusterBreak = 7
	U_GCB_T                  UGraphemeClusterBreak = 8
	U_GCB_V                  UGraphemeClusterBreak = 9
	U_GCB_SPACING_MARK       UGraphemeClusterBreak = 10
	U_GCB_PREPEND            UGraphemeClusterBreak = 11
	U_GCB_REGIONAL_INDICATOR UGraphemeClusterBreak = 12
	U_GCB_E_BASE             UGraphemeClusterBreak = 13
	U_GCB_E_BASE_GAZ         UGraphemeClusterBreak = 14
	U_GCB_E_MODIFIER         UGraphemeClusterBreak = 15
	U_GCB_GLUE_AFTER_ZWJ     UGraphemeClusterBreak = 16
	U_GCB_ZWJ                UGraphemeClusterBreak = 17
)

// enum
type UWordBreakValues int32

const (
	U_WB_OTHER              UWordBreakValues = 0
	U_WB_ALETTER            UWordBreakValues = 1
	U_WB_FORMAT             UWordBreakValues = 2
	U_WB_KATAKANA           UWordBreakValues = 3
	U_WB_MIDLETTER          UWordBreakValues = 4
	U_WB_MIDNUM             UWordBreakValues = 5
	U_WB_NUMERIC            UWordBreakValues = 6
	U_WB_EXTENDNUMLET       UWordBreakValues = 7
	U_WB_CR                 UWordBreakValues = 8
	U_WB_EXTEND             UWordBreakValues = 9
	U_WB_LF                 UWordBreakValues = 10
	U_WB_MIDNUMLET          UWordBreakValues = 11
	U_WB_NEWLINE            UWordBreakValues = 12
	U_WB_REGIONAL_INDICATOR UWordBreakValues = 13
	U_WB_HEBREW_LETTER      UWordBreakValues = 14
	U_WB_SINGLE_QUOTE       UWordBreakValues = 15
	U_WB_DOUBLE_QUOTE       UWordBreakValues = 16
	U_WB_E_BASE             UWordBreakValues = 17
	U_WB_E_BASE_GAZ         UWordBreakValues = 18
	U_WB_E_MODIFIER         UWordBreakValues = 19
	U_WB_GLUE_AFTER_ZWJ     UWordBreakValues = 20
	U_WB_ZWJ                UWordBreakValues = 21
	U_WB_WSEGSPACE          UWordBreakValues = 22
)

// enum
type USentenceBreak int32

const (
	U_SB_OTHER     USentenceBreak = 0
	U_SB_ATERM     USentenceBreak = 1
	U_SB_CLOSE     USentenceBreak = 2
	U_SB_FORMAT    USentenceBreak = 3
	U_SB_LOWER     USentenceBreak = 4
	U_SB_NUMERIC   USentenceBreak = 5
	U_SB_OLETTER   USentenceBreak = 6
	U_SB_SEP       USentenceBreak = 7
	U_SB_SP        USentenceBreak = 8
	U_SB_STERM     USentenceBreak = 9
	U_SB_UPPER     USentenceBreak = 10
	U_SB_CR        USentenceBreak = 11
	U_SB_EXTEND    USentenceBreak = 12
	U_SB_LF        USentenceBreak = 13
	U_SB_SCONTINUE USentenceBreak = 14
)

// enum
type ULineBreak int32

const (
	U_LB_UNKNOWN                      ULineBreak = 0
	U_LB_AMBIGUOUS                    ULineBreak = 1
	U_LB_ALPHABETIC                   ULineBreak = 2
	U_LB_BREAK_BOTH                   ULineBreak = 3
	U_LB_BREAK_AFTER                  ULineBreak = 4
	U_LB_BREAK_BEFORE                 ULineBreak = 5
	U_LB_MANDATORY_BREAK              ULineBreak = 6
	U_LB_CONTINGENT_BREAK             ULineBreak = 7
	U_LB_CLOSE_PUNCTUATION            ULineBreak = 8
	U_LB_COMBINING_MARK               ULineBreak = 9
	U_LB_CARRIAGE_RETURN              ULineBreak = 10
	U_LB_EXCLAMATION                  ULineBreak = 11
	U_LB_GLUE                         ULineBreak = 12
	U_LB_HYPHEN                       ULineBreak = 13
	U_LB_IDEOGRAPHIC                  ULineBreak = 14
	U_LB_INSEPARABLE                  ULineBreak = 15
	U_LB_INSEPERABLE                  ULineBreak = 15
	U_LB_INFIX_NUMERIC                ULineBreak = 16
	U_LB_LINE_FEED                    ULineBreak = 17
	U_LB_NONSTARTER                   ULineBreak = 18
	U_LB_NUMERIC                      ULineBreak = 19
	U_LB_OPEN_PUNCTUATION             ULineBreak = 20
	U_LB_POSTFIX_NUMERIC              ULineBreak = 21
	U_LB_PREFIX_NUMERIC               ULineBreak = 22
	U_LB_QUOTATION                    ULineBreak = 23
	U_LB_COMPLEX_CONTEXT              ULineBreak = 24
	U_LB_SURROGATE                    ULineBreak = 25
	U_LB_SPACE                        ULineBreak = 26
	U_LB_BREAK_SYMBOLS                ULineBreak = 27
	U_LB_ZWSPACE                      ULineBreak = 28
	U_LB_NEXT_LINE                    ULineBreak = 29
	U_LB_WORD_JOINER                  ULineBreak = 30
	U_LB_H2                           ULineBreak = 31
	U_LB_H3                           ULineBreak = 32
	U_LB_JL                           ULineBreak = 33
	U_LB_JT                           ULineBreak = 34
	U_LB_JV                           ULineBreak = 35
	U_LB_CLOSE_PARENTHESIS            ULineBreak = 36
	U_LB_CONDITIONAL_JAPANESE_STARTER ULineBreak = 37
	U_LB_HEBREW_LETTER                ULineBreak = 38
	U_LB_REGIONAL_INDICATOR           ULineBreak = 39
	U_LB_E_BASE                       ULineBreak = 40
	U_LB_E_MODIFIER                   ULineBreak = 41
	U_LB_ZWJ                          ULineBreak = 42
)

// enum
type UNumericType int32

const (
	U_NT_NONE    UNumericType = 0
	U_NT_DECIMAL UNumericType = 1
	U_NT_DIGIT   UNumericType = 2
	U_NT_NUMERIC UNumericType = 3
)

// enum
type UHangulSyllableType int32

const (
	U_HST_NOT_APPLICABLE UHangulSyllableType = 0
	U_HST_LEADING_JAMO   UHangulSyllableType = 1
	U_HST_VOWEL_JAMO     UHangulSyllableType = 2
	U_HST_TRAILING_JAMO  UHangulSyllableType = 3
	U_HST_LV_SYLLABLE    UHangulSyllableType = 4
	U_HST_LVT_SYLLABLE   UHangulSyllableType = 5
)

// enum
type UIndicPositionalCategory int32

const (
	U_INPC_NA                       UIndicPositionalCategory = 0
	U_INPC_BOTTOM                   UIndicPositionalCategory = 1
	U_INPC_BOTTOM_AND_LEFT          UIndicPositionalCategory = 2
	U_INPC_BOTTOM_AND_RIGHT         UIndicPositionalCategory = 3
	U_INPC_LEFT                     UIndicPositionalCategory = 4
	U_INPC_LEFT_AND_RIGHT           UIndicPositionalCategory = 5
	U_INPC_OVERSTRUCK               UIndicPositionalCategory = 6
	U_INPC_RIGHT                    UIndicPositionalCategory = 7
	U_INPC_TOP                      UIndicPositionalCategory = 8
	U_INPC_TOP_AND_BOTTOM           UIndicPositionalCategory = 9
	U_INPC_TOP_AND_BOTTOM_AND_RIGHT UIndicPositionalCategory = 10
	U_INPC_TOP_AND_LEFT             UIndicPositionalCategory = 11
	U_INPC_TOP_AND_LEFT_AND_RIGHT   UIndicPositionalCategory = 12
	U_INPC_TOP_AND_RIGHT            UIndicPositionalCategory = 13
	U_INPC_VISUAL_ORDER_LEFT        UIndicPositionalCategory = 14
	U_INPC_TOP_AND_BOTTOM_AND_LEFT  UIndicPositionalCategory = 15
)

// enum
type UIndicSyllabicCategory int32

const (
	U_INSC_OTHER                       UIndicSyllabicCategory = 0
	U_INSC_AVAGRAHA                    UIndicSyllabicCategory = 1
	U_INSC_BINDU                       UIndicSyllabicCategory = 2
	U_INSC_BRAHMI_JOINING_NUMBER       UIndicSyllabicCategory = 3
	U_INSC_CANTILLATION_MARK           UIndicSyllabicCategory = 4
	U_INSC_CONSONANT                   UIndicSyllabicCategory = 5
	U_INSC_CONSONANT_DEAD              UIndicSyllabicCategory = 6
	U_INSC_CONSONANT_FINAL             UIndicSyllabicCategory = 7
	U_INSC_CONSONANT_HEAD_LETTER       UIndicSyllabicCategory = 8
	U_INSC_CONSONANT_INITIAL_POSTFIXED UIndicSyllabicCategory = 9
	U_INSC_CONSONANT_KILLER            UIndicSyllabicCategory = 10
	U_INSC_CONSONANT_MEDIAL            UIndicSyllabicCategory = 11
	U_INSC_CONSONANT_PLACEHOLDER       UIndicSyllabicCategory = 12
	U_INSC_CONSONANT_PRECEDING_REPHA   UIndicSyllabicCategory = 13
	U_INSC_CONSONANT_PREFIXED          UIndicSyllabicCategory = 14
	U_INSC_CONSONANT_SUBJOINED         UIndicSyllabicCategory = 15
	U_INSC_CONSONANT_SUCCEEDING_REPHA  UIndicSyllabicCategory = 16
	U_INSC_CONSONANT_WITH_STACKER      UIndicSyllabicCategory = 17
	U_INSC_GEMINATION_MARK             UIndicSyllabicCategory = 18
	U_INSC_INVISIBLE_STACKER           UIndicSyllabicCategory = 19
	U_INSC_JOINER                      UIndicSyllabicCategory = 20
	U_INSC_MODIFYING_LETTER            UIndicSyllabicCategory = 21
	U_INSC_NON_JOINER                  UIndicSyllabicCategory = 22
	U_INSC_NUKTA                       UIndicSyllabicCategory = 23
	U_INSC_NUMBER                      UIndicSyllabicCategory = 24
	U_INSC_NUMBER_JOINER               UIndicSyllabicCategory = 25
	U_INSC_PURE_KILLER                 UIndicSyllabicCategory = 26
	U_INSC_REGISTER_SHIFTER            UIndicSyllabicCategory = 27
	U_INSC_SYLLABLE_MODIFIER           UIndicSyllabicCategory = 28
	U_INSC_TONE_LETTER                 UIndicSyllabicCategory = 29
	U_INSC_TONE_MARK                   UIndicSyllabicCategory = 30
	U_INSC_VIRAMA                      UIndicSyllabicCategory = 31
	U_INSC_VISARGA                     UIndicSyllabicCategory = 32
	U_INSC_VOWEL                       UIndicSyllabicCategory = 33
	U_INSC_VOWEL_DEPENDENT             UIndicSyllabicCategory = 34
	U_INSC_VOWEL_INDEPENDENT           UIndicSyllabicCategory = 35
)

// enum
type UVerticalOrientation int32

const (
	U_VO_ROTATED             UVerticalOrientation = 0
	U_VO_TRANSFORMED_ROTATED UVerticalOrientation = 1
	U_VO_TRANSFORMED_UPRIGHT UVerticalOrientation = 2
	U_VO_UPRIGHT             UVerticalOrientation = 3
)

// enum
type UBiDiDirection int32

const (
	UBIDI_LTR     UBiDiDirection = 0
	UBIDI_RTL     UBiDiDirection = 1
	UBIDI_MIXED   UBiDiDirection = 2
	UBIDI_NEUTRAL UBiDiDirection = 3
)

// enum
type UBiDiReorderingMode int32

const (
	UBIDI_REORDER_DEFAULT                     UBiDiReorderingMode = 0
	UBIDI_REORDER_NUMBERS_SPECIAL             UBiDiReorderingMode = 1
	UBIDI_REORDER_GROUP_NUMBERS_WITH_R        UBiDiReorderingMode = 2
	UBIDI_REORDER_RUNS_ONLY                   UBiDiReorderingMode = 3
	UBIDI_REORDER_INVERSE_NUMBERS_AS_L        UBiDiReorderingMode = 4
	UBIDI_REORDER_INVERSE_LIKE_DIRECT         UBiDiReorderingMode = 5
	UBIDI_REORDER_INVERSE_FOR_NUMBERS_SPECIAL UBiDiReorderingMode = 6
)

// enum
type UBiDiReorderingOption int32

const (
	UBIDI_OPTION_DEFAULT         UBiDiReorderingOption = 0
	UBIDI_OPTION_INSERT_MARKS    UBiDiReorderingOption = 1
	UBIDI_OPTION_REMOVE_CONTROLS UBiDiReorderingOption = 2
	UBIDI_OPTION_STREAMING       UBiDiReorderingOption = 4
)

// enum
type UBiDiOrder int32

const (
	UBIDI_LOGICAL UBiDiOrder = 0
	UBIDI_VISUAL  UBiDiOrder = 1
)

// enum
type UBiDiMirroring int32

const (
	UBIDI_MIRRORING_OFF UBiDiMirroring = 0
	UBIDI_MIRRORING_ON  UBiDiMirroring = 1
)

// enum
type USetSpanCondition int32

const (
	USET_SPAN_NOT_CONTAINED USetSpanCondition = 0
	USET_SPAN_CONTAINED     USetSpanCondition = 1
	USET_SPAN_SIMPLE        USetSpanCondition = 2
)

// enum
type UNormalization2Mode int32

const (
	UNORM2_COMPOSE            UNormalization2Mode = 0
	UNORM2_DECOMPOSE          UNormalization2Mode = 1
	UNORM2_FCD                UNormalization2Mode = 2
	UNORM2_COMPOSE_CONTIGUOUS UNormalization2Mode = 3
)

// enum
type UNormalizationCheckResult int32

const (
	UNORM_NO    UNormalizationCheckResult = 0
	UNORM_YES   UNormalizationCheckResult = 1
	UNORM_MAYBE UNormalizationCheckResult = 2
)

// enum
type UNormalizationMode int32

const (
	UNORM_NONE       UNormalizationMode = 1
	UNORM_NFD        UNormalizationMode = 2
	UNORM_NFKD       UNormalizationMode = 3
	UNORM_NFC        UNormalizationMode = 4
	UNORM_DEFAULT    UNormalizationMode = 4
	UNORM_NFKC       UNormalizationMode = 5
	UNORM_FCD        UNormalizationMode = 6
	UNORM_MODE_COUNT UNormalizationMode = 7
)

// enum
type UStringPrepProfileType int32

const (
	USPREP_RFC3491_NAMEPREP               UStringPrepProfileType = 0
	USPREP_RFC3530_NFS4_CS_PREP           UStringPrepProfileType = 1
	USPREP_RFC3530_NFS4_CS_PREP_CI        UStringPrepProfileType = 2
	USPREP_RFC3530_NFS4_CIS_PREP          UStringPrepProfileType = 3
	USPREP_RFC3530_NFS4_MIXED_PREP_PREFIX UStringPrepProfileType = 4
	USPREP_RFC3530_NFS4_MIXED_PREP_SUFFIX UStringPrepProfileType = 5
	USPREP_RFC3722_ISCSI                  UStringPrepProfileType = 6
	USPREP_RFC3920_NODEPREP               UStringPrepProfileType = 7
	USPREP_RFC3920_RESOURCEPREP           UStringPrepProfileType = 8
	USPREP_RFC4011_MIB                    UStringPrepProfileType = 9
	USPREP_RFC4013_SASLPREP               UStringPrepProfileType = 10
	USPREP_RFC4505_TRACE                  UStringPrepProfileType = 11
	USPREP_RFC4518_LDAP                   UStringPrepProfileType = 12
	USPREP_RFC4518_LDAP_CI                UStringPrepProfileType = 13
)

// enum
type UBreakIteratorType int32

const (
	UBRK_CHARACTER UBreakIteratorType = 0
	UBRK_WORD      UBreakIteratorType = 1
	UBRK_LINE      UBreakIteratorType = 2
	UBRK_SENTENCE  UBreakIteratorType = 3
)

// enum
type UWordBreak int32

const (
	UBRK_WORD_NONE         UWordBreak = 0
	UBRK_WORD_NONE_LIMIT   UWordBreak = 100
	UBRK_WORD_NUMBER       UWordBreak = 100
	UBRK_WORD_NUMBER_LIMIT UWordBreak = 200
	UBRK_WORD_LETTER       UWordBreak = 200
	UBRK_WORD_LETTER_LIMIT UWordBreak = 300
	UBRK_WORD_KANA         UWordBreak = 300
	UBRK_WORD_KANA_LIMIT   UWordBreak = 400
	UBRK_WORD_IDEO         UWordBreak = 400
	UBRK_WORD_IDEO_LIMIT   UWordBreak = 500
)

// enum
type ULineBreakTag int32

const (
	UBRK_LINE_SOFT       ULineBreakTag = 0
	UBRK_LINE_SOFT_LIMIT ULineBreakTag = 100
	UBRK_LINE_HARD       ULineBreakTag = 100
	UBRK_LINE_HARD_LIMIT ULineBreakTag = 200
)

// enum
type USentenceBreakTag int32

const (
	UBRK_SENTENCE_TERM       USentenceBreakTag = 0
	UBRK_SENTENCE_TERM_LIMIT USentenceBreakTag = 100
	UBRK_SENTENCE_SEP        USentenceBreakTag = 100
	UBRK_SENTENCE_SEP_LIMIT  USentenceBreakTag = 200
)

// enum
type UCalendarType int32

const (
	UCAL_TRADITIONAL UCalendarType = 0
	UCAL_DEFAULT     UCalendarType = 0
	UCAL_GREGORIAN   UCalendarType = 1
)

// enum
type UCalendarDateFields int32

const (
	UCAL_ERA                  UCalendarDateFields = 0
	UCAL_YEAR                 UCalendarDateFields = 1
	UCAL_MONTH                UCalendarDateFields = 2
	UCAL_WEEK_OF_YEAR         UCalendarDateFields = 3
	UCAL_WEEK_OF_MONTH        UCalendarDateFields = 4
	UCAL_DATE                 UCalendarDateFields = 5
	UCAL_DAY_OF_YEAR          UCalendarDateFields = 6
	UCAL_DAY_OF_WEEK          UCalendarDateFields = 7
	UCAL_DAY_OF_WEEK_IN_MONTH UCalendarDateFields = 8
	UCAL_AM_PM                UCalendarDateFields = 9
	UCAL_HOUR                 UCalendarDateFields = 10
	UCAL_HOUR_OF_DAY          UCalendarDateFields = 11
	UCAL_MINUTE               UCalendarDateFields = 12
	UCAL_SECOND               UCalendarDateFields = 13
	UCAL_MILLISECOND          UCalendarDateFields = 14
	UCAL_ZONE_OFFSET          UCalendarDateFields = 15
	UCAL_DST_OFFSET           UCalendarDateFields = 16
	UCAL_YEAR_WOY             UCalendarDateFields = 17
	UCAL_DOW_LOCAL            UCalendarDateFields = 18
	UCAL_EXTENDED_YEAR        UCalendarDateFields = 19
	UCAL_JULIAN_DAY           UCalendarDateFields = 20
	UCAL_MILLISECONDS_IN_DAY  UCalendarDateFields = 21
	UCAL_IS_LEAP_MONTH        UCalendarDateFields = 22
	UCAL_FIELD_COUNT          UCalendarDateFields = 23
	UCAL_DAY_OF_MONTH         UCalendarDateFields = 5
)

// enum
type UCalendarDaysOfWeek int32

const (
	UCAL_SUNDAY    UCalendarDaysOfWeek = 1
	UCAL_MONDAY    UCalendarDaysOfWeek = 2
	UCAL_TUESDAY   UCalendarDaysOfWeek = 3
	UCAL_WEDNESDAY UCalendarDaysOfWeek = 4
	UCAL_THURSDAY  UCalendarDaysOfWeek = 5
	UCAL_FRIDAY    UCalendarDaysOfWeek = 6
	UCAL_SATURDAY  UCalendarDaysOfWeek = 7
)

// enum
type UCalendarMonths int32

const (
	UCAL_JANUARY    UCalendarMonths = 0
	UCAL_FEBRUARY   UCalendarMonths = 1
	UCAL_MARCH      UCalendarMonths = 2
	UCAL_APRIL      UCalendarMonths = 3
	UCAL_MAY        UCalendarMonths = 4
	UCAL_JUNE       UCalendarMonths = 5
	UCAL_JULY       UCalendarMonths = 6
	UCAL_AUGUST     UCalendarMonths = 7
	UCAL_SEPTEMBER  UCalendarMonths = 8
	UCAL_OCTOBER    UCalendarMonths = 9
	UCAL_NOVEMBER   UCalendarMonths = 10
	UCAL_DECEMBER   UCalendarMonths = 11
	UCAL_UNDECIMBER UCalendarMonths = 12
)

// enum
type UCalendarAMPMs int32

const (
	UCAL_AM UCalendarAMPMs = 0
	UCAL_PM UCalendarAMPMs = 1
)

// enum
type USystemTimeZoneType int32

const (
	UCAL_ZONE_TYPE_ANY                USystemTimeZoneType = 0
	UCAL_ZONE_TYPE_CANONICAL          USystemTimeZoneType = 1
	UCAL_ZONE_TYPE_CANONICAL_LOCATION USystemTimeZoneType = 2
)

// enum
type UCalendarDisplayNameType int32

const (
	UCAL_STANDARD       UCalendarDisplayNameType = 0
	UCAL_SHORT_STANDARD UCalendarDisplayNameType = 1
	UCAL_DST            UCalendarDisplayNameType = 2
	UCAL_SHORT_DST      UCalendarDisplayNameType = 3
)

// enum
type UCalendarAttribute int32

const (
	UCAL_LENIENT                    UCalendarAttribute = 0
	UCAL_FIRST_DAY_OF_WEEK          UCalendarAttribute = 1
	UCAL_MINIMAL_DAYS_IN_FIRST_WEEK UCalendarAttribute = 2
	UCAL_REPEATED_WALL_TIME         UCalendarAttribute = 3
	UCAL_SKIPPED_WALL_TIME          UCalendarAttribute = 4
)

// enum
type UCalendarWallTimeOption int32

const (
	UCAL_WALLTIME_LAST       UCalendarWallTimeOption = 0
	UCAL_WALLTIME_FIRST      UCalendarWallTimeOption = 1
	UCAL_WALLTIME_NEXT_VALID UCalendarWallTimeOption = 2
)

// enum
type UCalendarLimitType int32

const (
	UCAL_MINIMUM          UCalendarLimitType = 0
	UCAL_MAXIMUM          UCalendarLimitType = 1
	UCAL_GREATEST_MINIMUM UCalendarLimitType = 2
	UCAL_LEAST_MAXIMUM    UCalendarLimitType = 3
	UCAL_ACTUAL_MINIMUM   UCalendarLimitType = 4
	UCAL_ACTUAL_MAXIMUM   UCalendarLimitType = 5
)

// enum
type UCalendarWeekdayType int32

const (
	UCAL_WEEKDAY       UCalendarWeekdayType = 0
	UCAL_WEEKEND       UCalendarWeekdayType = 1
	UCAL_WEEKEND_ONSET UCalendarWeekdayType = 2
	UCAL_WEEKEND_CEASE UCalendarWeekdayType = 3
)

// enum
type UTimeZoneTransitionType int32

const (
	UCAL_TZ_TRANSITION_NEXT               UTimeZoneTransitionType = 0
	UCAL_TZ_TRANSITION_NEXT_INCLUSIVE     UTimeZoneTransitionType = 1
	UCAL_TZ_TRANSITION_PREVIOUS           UTimeZoneTransitionType = 2
	UCAL_TZ_TRANSITION_PREVIOUS_INCLUSIVE UTimeZoneTransitionType = 3
)

// enum
type UCollationResult int32

const (
	UCOL_EQUAL   UCollationResult = 0
	UCOL_GREATER UCollationResult = 1
	UCOL_LESS    UCollationResult = -1
)

// enum
type UColAttributeValue int32

const (
	UCOL_DEFAULT           UColAttributeValue = -1
	UCOL_PRIMARY           UColAttributeValue = 0
	UCOL_SECONDARY         UColAttributeValue = 1
	UCOL_TERTIARY          UColAttributeValue = 2
	UCOL_DEFAULT_STRENGTH  UColAttributeValue = 2
	UCOL_CE_STRENGTH_LIMIT UColAttributeValue = 3
	UCOL_QUATERNARY        UColAttributeValue = 3
	UCOL_IDENTICAL         UColAttributeValue = 15
	UCOL_STRENGTH_LIMIT    UColAttributeValue = 16
	UCOL_OFF               UColAttributeValue = 16
	UCOL_ON                UColAttributeValue = 17
	UCOL_SHIFTED           UColAttributeValue = 20
	UCOL_NON_IGNORABLE     UColAttributeValue = 21
	UCOL_LOWER_FIRST       UColAttributeValue = 24
	UCOL_UPPER_FIRST       UColAttributeValue = 25
)

// enum
type UColReorderCode int32

const (
	UCOL_REORDER_CODE_DEFAULT     UColReorderCode = -1
	UCOL_REORDER_CODE_NONE        UColReorderCode = 103
	UCOL_REORDER_CODE_OTHERS      UColReorderCode = 103
	UCOL_REORDER_CODE_SPACE       UColReorderCode = 4096
	UCOL_REORDER_CODE_FIRST       UColReorderCode = 4096
	UCOL_REORDER_CODE_PUNCTUATION UColReorderCode = 4097
	UCOL_REORDER_CODE_SYMBOL      UColReorderCode = 4098
	UCOL_REORDER_CODE_CURRENCY    UColReorderCode = 4099
	UCOL_REORDER_CODE_DIGIT       UColReorderCode = 4100
)

// enum
type UColAttribute int32

const (
	UCOL_FRENCH_COLLATION   UColAttribute = 0
	UCOL_ALTERNATE_HANDLING UColAttribute = 1
	UCOL_CASE_FIRST         UColAttribute = 2
	UCOL_CASE_LEVEL         UColAttribute = 3
	UCOL_NORMALIZATION_MODE UColAttribute = 4
	UCOL_DECOMPOSITION_MODE UColAttribute = 4
	UCOL_STRENGTH           UColAttribute = 5
	UCOL_NUMERIC_COLLATION  UColAttribute = 7
	UCOL_ATTRIBUTE_COUNT    UColAttribute = 8
)

// enum
type UColRuleOption int32

const (
	UCOL_TAILORING_ONLY UColRuleOption = 0
	UCOL_FULL_RULES     UColRuleOption = 1
)

// enum
type UColBoundMode int32

const (
	UCOL_BOUND_LOWER      UColBoundMode = 0
	UCOL_BOUND_UPPER      UColBoundMode = 1
	UCOL_BOUND_UPPER_LONG UColBoundMode = 2
)

// enum
type UFormattableType int32

const (
	UFMT_DATE   UFormattableType = 0
	UFMT_DOUBLE UFormattableType = 1
	UFMT_LONG   UFormattableType = 2
	UFMT_STRING UFormattableType = 3
	UFMT_ARRAY  UFormattableType = 4
	UFMT_INT64  UFormattableType = 5
	UFMT_OBJECT UFormattableType = 6
)

// enum
type UFieldCategory int32

const (
	UFIELD_CATEGORY_UNDEFINED          UFieldCategory = 0
	UFIELD_CATEGORY_DATE               UFieldCategory = 1
	UFIELD_CATEGORY_NUMBER             UFieldCategory = 2
	UFIELD_CATEGORY_LIST               UFieldCategory = 3
	UFIELD_CATEGORY_RELATIVE_DATETIME  UFieldCategory = 4
	UFIELD_CATEGORY_DATE_INTERVAL      UFieldCategory = 5
	UFIELD_CATEGORY_LIST_SPAN          UFieldCategory = 4099
	UFIELD_CATEGORY_DATE_INTERVAL_SPAN UFieldCategory = 4101
)

// enum
type UGender int32

const (
	UGENDER_MALE   UGender = 0
	UGENDER_FEMALE UGender = 1
	UGENDER_OTHER  UGender = 2
)

// enum
type UListFormatterField int32

const (
	ULISTFMT_LITERAL_FIELD UListFormatterField = 0
	ULISTFMT_ELEMENT_FIELD UListFormatterField = 1
)

// enum
type UListFormatterType int32

const (
	ULISTFMT_TYPE_AND   UListFormatterType = 0
	ULISTFMT_TYPE_OR    UListFormatterType = 1
	ULISTFMT_TYPE_UNITS UListFormatterType = 2
)

// enum
type UListFormatterWidth int32

const (
	ULISTFMT_WIDTH_WIDE   UListFormatterWidth = 0
	ULISTFMT_WIDTH_SHORT  UListFormatterWidth = 1
	ULISTFMT_WIDTH_NARROW UListFormatterWidth = 2
)

// enum
type ULocaleDataExemplarSetType int32

const (
	ULOCDATA_ES_STANDARD    ULocaleDataExemplarSetType = 0
	ULOCDATA_ES_AUXILIARY   ULocaleDataExemplarSetType = 1
	ULOCDATA_ES_INDEX       ULocaleDataExemplarSetType = 2
	ULOCDATA_ES_PUNCTUATION ULocaleDataExemplarSetType = 3
)

// enum
type ULocaleDataDelimiterType int32

const (
	ULOCDATA_QUOTATION_START     ULocaleDataDelimiterType = 0
	ULOCDATA_QUOTATION_END       ULocaleDataDelimiterType = 1
	ULOCDATA_ALT_QUOTATION_START ULocaleDataDelimiterType = 2
	ULOCDATA_ALT_QUOTATION_END   ULocaleDataDelimiterType = 3
)

// enum
type UMeasurementSystem int32

const (
	UMS_SI UMeasurementSystem = 0
	UMS_US UMeasurementSystem = 1
	UMS_UK UMeasurementSystem = 2
)

// enum
type UNumberFormatStyle int32

const (
	UNUM_PATTERN_DECIMAL       UNumberFormatStyle = 0
	UNUM_DECIMAL               UNumberFormatStyle = 1
	UNUM_CURRENCY              UNumberFormatStyle = 2
	UNUM_PERCENT               UNumberFormatStyle = 3
	UNUM_SCIENTIFIC            UNumberFormatStyle = 4
	UNUM_SPELLOUT              UNumberFormatStyle = 5
	UNUM_ORDINAL               UNumberFormatStyle = 6
	UNUM_DURATION              UNumberFormatStyle = 7
	UNUM_NUMBERING_SYSTEM      UNumberFormatStyle = 8
	UNUM_PATTERN_RULEBASED     UNumberFormatStyle = 9
	UNUM_CURRENCY_ISO          UNumberFormatStyle = 10
	UNUM_CURRENCY_PLURAL       UNumberFormatStyle = 11
	UNUM_CURRENCY_ACCOUNTING   UNumberFormatStyle = 12
	UNUM_CASH_CURRENCY         UNumberFormatStyle = 13
	UNUM_DECIMAL_COMPACT_SHORT UNumberFormatStyle = 14
	UNUM_DECIMAL_COMPACT_LONG  UNumberFormatStyle = 15
	UNUM_CURRENCY_STANDARD     UNumberFormatStyle = 16
	UNUM_DEFAULT               UNumberFormatStyle = 1
	UNUM_IGNORE                UNumberFormatStyle = 0
)

// enum
type UNumberFormatRoundingMode int32

const (
	UNUM_ROUND_CEILING     UNumberFormatRoundingMode = 0
	UNUM_ROUND_FLOOR       UNumberFormatRoundingMode = 1
	UNUM_ROUND_DOWN        UNumberFormatRoundingMode = 2
	UNUM_ROUND_UP          UNumberFormatRoundingMode = 3
	UNUM_ROUND_HALFEVEN    UNumberFormatRoundingMode = 4
	UNUM_ROUND_HALFDOWN    UNumberFormatRoundingMode = 5
	UNUM_ROUND_HALFUP      UNumberFormatRoundingMode = 6
	UNUM_ROUND_UNNECESSARY UNumberFormatRoundingMode = 7
)

// enum
type UNumberFormatPadPosition int32

const (
	UNUM_PAD_BEFORE_PREFIX UNumberFormatPadPosition = 0
	UNUM_PAD_AFTER_PREFIX  UNumberFormatPadPosition = 1
	UNUM_PAD_BEFORE_SUFFIX UNumberFormatPadPosition = 2
	UNUM_PAD_AFTER_SUFFIX  UNumberFormatPadPosition = 3
)

// enum
type UNumberCompactStyle int32

const (
	UNUM_SHORT UNumberCompactStyle = 0
	UNUM_LONG  UNumberCompactStyle = 1
)

// enum
type UCurrencySpacing int32

const (
	UNUM_CURRENCY_MATCH             UCurrencySpacing = 0
	UNUM_CURRENCY_SURROUNDING_MATCH UCurrencySpacing = 1
	UNUM_CURRENCY_INSERT            UCurrencySpacing = 2
	UNUM_CURRENCY_SPACING_COUNT     UCurrencySpacing = 3
)

// enum
type UNumberFormatFields int32

const (
	UNUM_INTEGER_FIELD            UNumberFormatFields = 0
	UNUM_FRACTION_FIELD           UNumberFormatFields = 1
	UNUM_DECIMAL_SEPARATOR_FIELD  UNumberFormatFields = 2
	UNUM_EXPONENT_SYMBOL_FIELD    UNumberFormatFields = 3
	UNUM_EXPONENT_SIGN_FIELD      UNumberFormatFields = 4
	UNUM_EXPONENT_FIELD           UNumberFormatFields = 5
	UNUM_GROUPING_SEPARATOR_FIELD UNumberFormatFields = 6
	UNUM_CURRENCY_FIELD           UNumberFormatFields = 7
	UNUM_PERCENT_FIELD            UNumberFormatFields = 8
	UNUM_PERMILL_FIELD            UNumberFormatFields = 9
	UNUM_SIGN_FIELD               UNumberFormatFields = 10
	UNUM_MEASURE_UNIT_FIELD       UNumberFormatFields = 11
	UNUM_COMPACT_FIELD            UNumberFormatFields = 12
)

// enum
type UNumberFormatAttributeValue int32

const (
	UNUM_FORMAT_ATTRIBUTE_VALUE_HIDDEN UNumberFormatAttributeValue = 0
)

// enum
type UNumberFormatAttribute int32

const (
	UNUM_PARSE_INT_ONLY                      UNumberFormatAttribute = 0
	UNUM_GROUPING_USED                       UNumberFormatAttribute = 1
	UNUM_DECIMAL_ALWAYS_SHOWN                UNumberFormatAttribute = 2
	UNUM_MAX_INTEGER_DIGITS                  UNumberFormatAttribute = 3
	UNUM_MIN_INTEGER_DIGITS                  UNumberFormatAttribute = 4
	UNUM_INTEGER_DIGITS                      UNumberFormatAttribute = 5
	UNUM_MAX_FRACTION_DIGITS                 UNumberFormatAttribute = 6
	UNUM_MIN_FRACTION_DIGITS                 UNumberFormatAttribute = 7
	UNUM_FRACTION_DIGITS                     UNumberFormatAttribute = 8
	UNUM_MULTIPLIER                          UNumberFormatAttribute = 9
	UNUM_GROUPING_SIZE                       UNumberFormatAttribute = 10
	UNUM_ROUNDING_MODE                       UNumberFormatAttribute = 11
	UNUM_ROUNDING_INCREMENT                  UNumberFormatAttribute = 12
	UNUM_FORMAT_WIDTH                        UNumberFormatAttribute = 13
	UNUM_PADDING_POSITION                    UNumberFormatAttribute = 14
	UNUM_SECONDARY_GROUPING_SIZE             UNumberFormatAttribute = 15
	UNUM_SIGNIFICANT_DIGITS_USED             UNumberFormatAttribute = 16
	UNUM_MIN_SIGNIFICANT_DIGITS              UNumberFormatAttribute = 17
	UNUM_MAX_SIGNIFICANT_DIGITS              UNumberFormatAttribute = 18
	UNUM_LENIENT_PARSE                       UNumberFormatAttribute = 19
	UNUM_PARSE_ALL_INPUT                     UNumberFormatAttribute = 20
	UNUM_SCALE                               UNumberFormatAttribute = 21
	UNUM_MINIMUM_GROUPING_DIGITS             UNumberFormatAttribute = 22
	UNUM_CURRENCY_USAGE                      UNumberFormatAttribute = 23
	UNUM_FORMAT_FAIL_IF_MORE_THAN_MAX_DIGITS UNumberFormatAttribute = 4096
	UNUM_PARSE_NO_EXPONENT                   UNumberFormatAttribute = 4097
	UNUM_PARSE_DECIMAL_MARK_REQUIRED         UNumberFormatAttribute = 4098
	UNUM_PARSE_CASE_SENSITIVE                UNumberFormatAttribute = 4099
	UNUM_SIGN_ALWAYS_SHOWN                   UNumberFormatAttribute = 4100
)

// enum
type UNumberFormatTextAttribute int32

const (
	UNUM_POSITIVE_PREFIX   UNumberFormatTextAttribute = 0
	UNUM_POSITIVE_SUFFIX   UNumberFormatTextAttribute = 1
	UNUM_NEGATIVE_PREFIX   UNumberFormatTextAttribute = 2
	UNUM_NEGATIVE_SUFFIX   UNumberFormatTextAttribute = 3
	UNUM_PADDING_CHARACTER UNumberFormatTextAttribute = 4
	UNUM_CURRENCY_CODE     UNumberFormatTextAttribute = 5
	UNUM_DEFAULT_RULESET   UNumberFormatTextAttribute = 6
	UNUM_PUBLIC_RULESETS   UNumberFormatTextAttribute = 7
)

// enum
type UNumberFormatSymbol int32

const (
	UNUM_DECIMAL_SEPARATOR_SYMBOL           UNumberFormatSymbol = 0
	UNUM_GROUPING_SEPARATOR_SYMBOL          UNumberFormatSymbol = 1
	UNUM_PATTERN_SEPARATOR_SYMBOL           UNumberFormatSymbol = 2
	UNUM_PERCENT_SYMBOL                     UNumberFormatSymbol = 3
	UNUM_ZERO_DIGIT_SYMBOL                  UNumberFormatSymbol = 4
	UNUM_DIGIT_SYMBOL                       UNumberFormatSymbol = 5
	UNUM_MINUS_SIGN_SYMBOL                  UNumberFormatSymbol = 6
	UNUM_PLUS_SIGN_SYMBOL                   UNumberFormatSymbol = 7
	UNUM_CURRENCY_SYMBOL                    UNumberFormatSymbol = 8
	UNUM_INTL_CURRENCY_SYMBOL               UNumberFormatSymbol = 9
	UNUM_MONETARY_SEPARATOR_SYMBOL          UNumberFormatSymbol = 10
	UNUM_EXPONENTIAL_SYMBOL                 UNumberFormatSymbol = 11
	UNUM_PERMILL_SYMBOL                     UNumberFormatSymbol = 12
	UNUM_PAD_ESCAPE_SYMBOL                  UNumberFormatSymbol = 13
	UNUM_INFINITY_SYMBOL                    UNumberFormatSymbol = 14
	UNUM_NAN_SYMBOL                         UNumberFormatSymbol = 15
	UNUM_SIGNIFICANT_DIGIT_SYMBOL           UNumberFormatSymbol = 16
	UNUM_MONETARY_GROUPING_SEPARATOR_SYMBOL UNumberFormatSymbol = 17
	UNUM_ONE_DIGIT_SYMBOL                   UNumberFormatSymbol = 18
	UNUM_TWO_DIGIT_SYMBOL                   UNumberFormatSymbol = 19
	UNUM_THREE_DIGIT_SYMBOL                 UNumberFormatSymbol = 20
	UNUM_FOUR_DIGIT_SYMBOL                  UNumberFormatSymbol = 21
	UNUM_FIVE_DIGIT_SYMBOL                  UNumberFormatSymbol = 22
	UNUM_SIX_DIGIT_SYMBOL                   UNumberFormatSymbol = 23
	UNUM_SEVEN_DIGIT_SYMBOL                 UNumberFormatSymbol = 24
	UNUM_EIGHT_DIGIT_SYMBOL                 UNumberFormatSymbol = 25
	UNUM_NINE_DIGIT_SYMBOL                  UNumberFormatSymbol = 26
	UNUM_EXPONENT_MULTIPLICATION_SYMBOL     UNumberFormatSymbol = 27
)

// enum
type UDateFormatStyle int32

const (
	UDAT_FULL            UDateFormatStyle = 0
	UDAT_LONG            UDateFormatStyle = 1
	UDAT_MEDIUM          UDateFormatStyle = 2
	UDAT_SHORT           UDateFormatStyle = 3
	UDAT_DEFAULT         UDateFormatStyle = 2
	UDAT_RELATIVE        UDateFormatStyle = 128
	UDAT_FULL_RELATIVE   UDateFormatStyle = 128
	UDAT_LONG_RELATIVE   UDateFormatStyle = 129
	UDAT_MEDIUM_RELATIVE UDateFormatStyle = 130
	UDAT_SHORT_RELATIVE  UDateFormatStyle = 131
	UDAT_NONE            UDateFormatStyle = -1
	UDAT_PATTERN         UDateFormatStyle = -2
)

// enum
type UDateFormatField int32

const (
	UDAT_ERA_FIELD                           UDateFormatField = 0
	UDAT_YEAR_FIELD                          UDateFormatField = 1
	UDAT_MONTH_FIELD                         UDateFormatField = 2
	UDAT_DATE_FIELD                          UDateFormatField = 3
	UDAT_HOUR_OF_DAY1_FIELD                  UDateFormatField = 4
	UDAT_HOUR_OF_DAY0_FIELD                  UDateFormatField = 5
	UDAT_MINUTE_FIELD                        UDateFormatField = 6
	UDAT_SECOND_FIELD                        UDateFormatField = 7
	UDAT_FRACTIONAL_SECOND_FIELD             UDateFormatField = 8
	UDAT_DAY_OF_WEEK_FIELD                   UDateFormatField = 9
	UDAT_DAY_OF_YEAR_FIELD                   UDateFormatField = 10
	UDAT_DAY_OF_WEEK_IN_MONTH_FIELD          UDateFormatField = 11
	UDAT_WEEK_OF_YEAR_FIELD                  UDateFormatField = 12
	UDAT_WEEK_OF_MONTH_FIELD                 UDateFormatField = 13
	UDAT_AM_PM_FIELD                         UDateFormatField = 14
	UDAT_HOUR1_FIELD                         UDateFormatField = 15
	UDAT_HOUR0_FIELD                         UDateFormatField = 16
	UDAT_TIMEZONE_FIELD                      UDateFormatField = 17
	UDAT_YEAR_WOY_FIELD                      UDateFormatField = 18
	UDAT_DOW_LOCAL_FIELD                     UDateFormatField = 19
	UDAT_EXTENDED_YEAR_FIELD                 UDateFormatField = 20
	UDAT_JULIAN_DAY_FIELD                    UDateFormatField = 21
	UDAT_MILLISECONDS_IN_DAY_FIELD           UDateFormatField = 22
	UDAT_TIMEZONE_RFC_FIELD                  UDateFormatField = 23
	UDAT_TIMEZONE_GENERIC_FIELD              UDateFormatField = 24
	UDAT_STANDALONE_DAY_FIELD                UDateFormatField = 25
	UDAT_STANDALONE_MONTH_FIELD              UDateFormatField = 26
	UDAT_QUARTER_FIELD                       UDateFormatField = 27
	UDAT_STANDALONE_QUARTER_FIELD            UDateFormatField = 28
	UDAT_TIMEZONE_SPECIAL_FIELD              UDateFormatField = 29
	UDAT_YEAR_NAME_FIELD                     UDateFormatField = 30
	UDAT_TIMEZONE_LOCALIZED_GMT_OFFSET_FIELD UDateFormatField = 31
	UDAT_TIMEZONE_ISO_FIELD                  UDateFormatField = 32
	UDAT_TIMEZONE_ISO_LOCAL_FIELD            UDateFormatField = 33
	UDAT_AM_PM_MIDNIGHT_NOON_FIELD           UDateFormatField = 35
	UDAT_FLEXIBLE_DAY_PERIOD_FIELD           UDateFormatField = 36
)

// enum
type UDateFormatBooleanAttribute int32

const (
	UDAT_PARSE_ALLOW_WHITESPACE            UDateFormatBooleanAttribute = 0
	UDAT_PARSE_ALLOW_NUMERIC               UDateFormatBooleanAttribute = 1
	UDAT_PARSE_PARTIAL_LITERAL_MATCH       UDateFormatBooleanAttribute = 2
	UDAT_PARSE_MULTIPLE_PATTERNS_FOR_MATCH UDateFormatBooleanAttribute = 3
	UDAT_BOOLEAN_ATTRIBUTE_COUNT           UDateFormatBooleanAttribute = 4
)

// enum
type UDateFormatSymbolType int32

const (
	UDAT_ERAS                        UDateFormatSymbolType = 0
	UDAT_MONTHS                      UDateFormatSymbolType = 1
	UDAT_SHORT_MONTHS                UDateFormatSymbolType = 2
	UDAT_WEEKDAYS                    UDateFormatSymbolType = 3
	UDAT_SHORT_WEEKDAYS              UDateFormatSymbolType = 4
	UDAT_AM_PMS                      UDateFormatSymbolType = 5
	UDAT_LOCALIZED_CHARS             UDateFormatSymbolType = 6
	UDAT_ERA_NAMES                   UDateFormatSymbolType = 7
	UDAT_NARROW_MONTHS               UDateFormatSymbolType = 8
	UDAT_NARROW_WEEKDAYS             UDateFormatSymbolType = 9
	UDAT_STANDALONE_MONTHS           UDateFormatSymbolType = 10
	UDAT_STANDALONE_SHORT_MONTHS     UDateFormatSymbolType = 11
	UDAT_STANDALONE_NARROW_MONTHS    UDateFormatSymbolType = 12
	UDAT_STANDALONE_WEEKDAYS         UDateFormatSymbolType = 13
	UDAT_STANDALONE_SHORT_WEEKDAYS   UDateFormatSymbolType = 14
	UDAT_STANDALONE_NARROW_WEEKDAYS  UDateFormatSymbolType = 15
	UDAT_QUARTERS                    UDateFormatSymbolType = 16
	UDAT_SHORT_QUARTERS              UDateFormatSymbolType = 17
	UDAT_STANDALONE_QUARTERS         UDateFormatSymbolType = 18
	UDAT_STANDALONE_SHORT_QUARTERS   UDateFormatSymbolType = 19
	UDAT_SHORTER_WEEKDAYS            UDateFormatSymbolType = 20
	UDAT_STANDALONE_SHORTER_WEEKDAYS UDateFormatSymbolType = 21
	UDAT_CYCLIC_YEARS_WIDE           UDateFormatSymbolType = 22
	UDAT_CYCLIC_YEARS_ABBREVIATED    UDateFormatSymbolType = 23
	UDAT_CYCLIC_YEARS_NARROW         UDateFormatSymbolType = 24
	UDAT_ZODIAC_NAMES_WIDE           UDateFormatSymbolType = 25
	UDAT_ZODIAC_NAMES_ABBREVIATED    UDateFormatSymbolType = 26
	UDAT_ZODIAC_NAMES_NARROW         UDateFormatSymbolType = 27
)

// enum
type UDateTimePatternField int32

const (
	UDATPG_ERA_FIELD                  UDateTimePatternField = 0
	UDATPG_YEAR_FIELD                 UDateTimePatternField = 1
	UDATPG_QUARTER_FIELD              UDateTimePatternField = 2
	UDATPG_MONTH_FIELD                UDateTimePatternField = 3
	UDATPG_WEEK_OF_YEAR_FIELD         UDateTimePatternField = 4
	UDATPG_WEEK_OF_MONTH_FIELD        UDateTimePatternField = 5
	UDATPG_WEEKDAY_FIELD              UDateTimePatternField = 6
	UDATPG_DAY_OF_YEAR_FIELD          UDateTimePatternField = 7
	UDATPG_DAY_OF_WEEK_IN_MONTH_FIELD UDateTimePatternField = 8
	UDATPG_DAY_FIELD                  UDateTimePatternField = 9
	UDATPG_DAYPERIOD_FIELD            UDateTimePatternField = 10
	UDATPG_HOUR_FIELD                 UDateTimePatternField = 11
	UDATPG_MINUTE_FIELD               UDateTimePatternField = 12
	UDATPG_SECOND_FIELD               UDateTimePatternField = 13
	UDATPG_FRACTIONAL_SECOND_FIELD    UDateTimePatternField = 14
	UDATPG_ZONE_FIELD                 UDateTimePatternField = 15
	UDATPG_FIELD_COUNT                UDateTimePatternField = 16
)

// enum
type UDateTimePGDisplayWidth int32

const (
	UDATPG_WIDE        UDateTimePGDisplayWidth = 0
	UDATPG_ABBREVIATED UDateTimePGDisplayWidth = 1
	UDATPG_NARROW      UDateTimePGDisplayWidth = 2
)

// enum
type UDateTimePatternMatchOptions int32

const (
	UDATPG_MATCH_NO_OPTIONS        UDateTimePatternMatchOptions = 0
	UDATPG_MATCH_HOUR_FIELD_LENGTH UDateTimePatternMatchOptions = 2048
	UDATPG_MATCH_ALL_FIELDS_LENGTH UDateTimePatternMatchOptions = 65535
)

// enum
type UDateTimePatternConflict int32

const (
	UDATPG_NO_CONFLICT   UDateTimePatternConflict = 0
	UDATPG_BASE_CONFLICT UDateTimePatternConflict = 1
	UDATPG_CONFLICT      UDateTimePatternConflict = 2
)

// enum
type UNumberUnitWidth int32

const (
	UNUM_UNIT_WIDTH_NARROW    UNumberUnitWidth = 0
	UNUM_UNIT_WIDTH_SHORT     UNumberUnitWidth = 1
	UNUM_UNIT_WIDTH_FULL_NAME UNumberUnitWidth = 2
	UNUM_UNIT_WIDTH_ISO_CODE  UNumberUnitWidth = 3
	UNUM_UNIT_WIDTH_HIDDEN    UNumberUnitWidth = 4
	UNUM_UNIT_WIDTH_COUNT     UNumberUnitWidth = 5
)

// enum
type UNumberGroupingStrategy int32

const (
	UNUM_GROUPING_OFF        UNumberGroupingStrategy = 0
	UNUM_GROUPING_MIN2       UNumberGroupingStrategy = 1
	UNUM_GROUPING_AUTO       UNumberGroupingStrategy = 2
	UNUM_GROUPING_ON_ALIGNED UNumberGroupingStrategy = 3
	UNUM_GROUPING_THOUSANDS  UNumberGroupingStrategy = 4
)

// enum
type UNumberSignDisplay int32

const (
	UNUM_SIGN_AUTO                   UNumberSignDisplay = 0
	UNUM_SIGN_ALWAYS                 UNumberSignDisplay = 1
	UNUM_SIGN_NEVER                  UNumberSignDisplay = 2
	UNUM_SIGN_ACCOUNTING             UNumberSignDisplay = 3
	UNUM_SIGN_ACCOUNTING_ALWAYS      UNumberSignDisplay = 4
	UNUM_SIGN_EXCEPT_ZERO            UNumberSignDisplay = 5
	UNUM_SIGN_ACCOUNTING_EXCEPT_ZERO UNumberSignDisplay = 6
	UNUM_SIGN_COUNT                  UNumberSignDisplay = 7
)

// enum
type UNumberDecimalSeparatorDisplay int32

const (
	UNUM_DECIMAL_SEPARATOR_AUTO   UNumberDecimalSeparatorDisplay = 0
	UNUM_DECIMAL_SEPARATOR_ALWAYS UNumberDecimalSeparatorDisplay = 1
	UNUM_DECIMAL_SEPARATOR_COUNT  UNumberDecimalSeparatorDisplay = 2
)

// enum
type UNumberRangeCollapse int32

const (
	UNUM_RANGE_COLLAPSE_AUTO UNumberRangeCollapse = 0
	UNUM_RANGE_COLLAPSE_NONE UNumberRangeCollapse = 1
	UNUM_RANGE_COLLAPSE_UNIT UNumberRangeCollapse = 2
	UNUM_RANGE_COLLAPSE_ALL  UNumberRangeCollapse = 3
)

// enum
type UNumberRangeIdentityFallback int32

const (
	UNUM_IDENTITY_FALLBACK_SINGLE_VALUE                  UNumberRangeIdentityFallback = 0
	UNUM_IDENTITY_FALLBACK_APPROXIMATELY_OR_SINGLE_VALUE UNumberRangeIdentityFallback = 1
	UNUM_IDENTITY_FALLBACK_APPROXIMATELY                 UNumberRangeIdentityFallback = 2
	UNUM_IDENTITY_FALLBACK_RANGE                         UNumberRangeIdentityFallback = 3
)

// enum
type UNumberRangeIdentityResult int32

const (
	UNUM_IDENTITY_RESULT_EQUAL_BEFORE_ROUNDING UNumberRangeIdentityResult = 0
	UNUM_IDENTITY_RESULT_EQUAL_AFTER_ROUNDING  UNumberRangeIdentityResult = 1
	UNUM_IDENTITY_RESULT_NOT_EQUAL             UNumberRangeIdentityResult = 2
)

// enum
type UPluralType int32

const (
	UPLURAL_TYPE_CARDINAL UPluralType = 0
	UPLURAL_TYPE_ORDINAL  UPluralType = 1
)

// enum
type URegexpFlag int32

const (
	UREGEX_CASE_INSENSITIVE         URegexpFlag = 2
	UREGEX_COMMENTS                 URegexpFlag = 4
	UREGEX_DOTALL                   URegexpFlag = 32
	UREGEX_LITERAL                  URegexpFlag = 16
	UREGEX_MULTILINE                URegexpFlag = 8
	UREGEX_UNIX_LINES               URegexpFlag = 1
	UREGEX_UWORD                    URegexpFlag = 256
	UREGEX_ERROR_ON_UNKNOWN_ESCAPES URegexpFlag = 512
)

// enum
type URegionType int32

const (
	URGN_UNKNOWN      URegionType = 0
	URGN_TERRITORY    URegionType = 1
	URGN_WORLD        URegionType = 2
	URGN_CONTINENT    URegionType = 3
	URGN_SUBCONTINENT URegionType = 4
	URGN_GROUPING     URegionType = 5
	URGN_DEPRECATED   URegionType = 6
)

// enum
type UDateRelativeDateTimeFormatterStyle int32

const (
	UDAT_STYLE_LONG   UDateRelativeDateTimeFormatterStyle = 0
	UDAT_STYLE_SHORT  UDateRelativeDateTimeFormatterStyle = 1
	UDAT_STYLE_NARROW UDateRelativeDateTimeFormatterStyle = 2
)

// enum
type URelativeDateTimeUnit int32

const (
	UDAT_REL_UNIT_YEAR      URelativeDateTimeUnit = 0
	UDAT_REL_UNIT_QUARTER   URelativeDateTimeUnit = 1
	UDAT_REL_UNIT_MONTH     URelativeDateTimeUnit = 2
	UDAT_REL_UNIT_WEEK      URelativeDateTimeUnit = 3
	UDAT_REL_UNIT_DAY       URelativeDateTimeUnit = 4
	UDAT_REL_UNIT_HOUR      URelativeDateTimeUnit = 5
	UDAT_REL_UNIT_MINUTE    URelativeDateTimeUnit = 6
	UDAT_REL_UNIT_SECOND    URelativeDateTimeUnit = 7
	UDAT_REL_UNIT_SUNDAY    URelativeDateTimeUnit = 8
	UDAT_REL_UNIT_MONDAY    URelativeDateTimeUnit = 9
	UDAT_REL_UNIT_TUESDAY   URelativeDateTimeUnit = 10
	UDAT_REL_UNIT_WEDNESDAY URelativeDateTimeUnit = 11
	UDAT_REL_UNIT_THURSDAY  URelativeDateTimeUnit = 12
	UDAT_REL_UNIT_FRIDAY    URelativeDateTimeUnit = 13
	UDAT_REL_UNIT_SATURDAY  URelativeDateTimeUnit = 14
)

// enum
type URelativeDateTimeFormatterField int32

const (
	UDAT_REL_LITERAL_FIELD URelativeDateTimeFormatterField = 0
	UDAT_REL_NUMERIC_FIELD URelativeDateTimeFormatterField = 1
)

// enum
type USearchAttribute int32

const (
	USEARCH_OVERLAP            USearchAttribute = 0
	USEARCH_ELEMENT_COMPARISON USearchAttribute = 2
)

// enum
type USearchAttributeValue int32

const (
	USEARCH_DEFAULT                         USearchAttributeValue = -1
	USEARCH_OFF                             USearchAttributeValue = 0
	USEARCH_ON                              USearchAttributeValue = 1
	USEARCH_STANDARD_ELEMENT_COMPARISON     USearchAttributeValue = 2
	USEARCH_PATTERN_BASE_WEIGHT_IS_WILDCARD USearchAttributeValue = 3
	USEARCH_ANY_BASE_WEIGHT_IS_WILDCARD     USearchAttributeValue = 4
)

// enum
type USpoofChecks int32

const (
	USPOOF_SINGLE_SCRIPT_CONFUSABLE USpoofChecks = 1
	USPOOF_MIXED_SCRIPT_CONFUSABLE  USpoofChecks = 2
	USPOOF_WHOLE_SCRIPT_CONFUSABLE  USpoofChecks = 4
	USPOOF_CONFUSABLE               USpoofChecks = 7
	USPOOF_RESTRICTION_LEVEL        USpoofChecks = 16
	USPOOF_INVISIBLE                USpoofChecks = 32
	USPOOF_CHAR_LIMIT               USpoofChecks = 64
	USPOOF_MIXED_NUMBERS            USpoofChecks = 128
	USPOOF_HIDDEN_OVERLAY           USpoofChecks = 256
	USPOOF_ALL_CHECKS               USpoofChecks = 65535
	USPOOF_AUX_INFO                 USpoofChecks = 1073741824
)

// enum
type URestrictionLevel int32

const (
	USPOOF_ASCII                     URestrictionLevel = 268435456
	USPOOF_SINGLE_SCRIPT_RESTRICTIVE URestrictionLevel = 536870912
	USPOOF_HIGHLY_RESTRICTIVE        URestrictionLevel = 805306368
	USPOOF_MODERATELY_RESTRICTIVE    URestrictionLevel = 1073741824
	USPOOF_MINIMALLY_RESTRICTIVE     URestrictionLevel = 1342177280
	USPOOF_UNRESTRICTIVE             URestrictionLevel = 1610612736
	USPOOF_RESTRICTION_LEVEL_MASK    URestrictionLevel = 2130706432
)

// enum
type UDateTimeScale int32

const (
	UDTS_JAVA_TIME              UDateTimeScale = 0
	UDTS_UNIX_TIME              UDateTimeScale = 1
	UDTS_ICU4C_TIME             UDateTimeScale = 2
	UDTS_WINDOWS_FILE_TIME      UDateTimeScale = 3
	UDTS_DOTNET_DATE_TIME       UDateTimeScale = 4
	UDTS_MAC_OLD_TIME           UDateTimeScale = 5
	UDTS_MAC_TIME               UDateTimeScale = 6
	UDTS_EXCEL_TIME             UDateTimeScale = 7
	UDTS_DB2_TIME               UDateTimeScale = 8
	UDTS_UNIX_MICROSECONDS_TIME UDateTimeScale = 9
)

// enum
type UTimeScaleValue int32

const (
	UTSV_UNITS_VALUE        UTimeScaleValue = 0
	UTSV_EPOCH_OFFSET_VALUE UTimeScaleValue = 1
	UTSV_FROM_MIN_VALUE     UTimeScaleValue = 2
	UTSV_FROM_MAX_VALUE     UTimeScaleValue = 3
	UTSV_TO_MIN_VALUE       UTimeScaleValue = 4
	UTSV_TO_MAX_VALUE       UTimeScaleValue = 5
)

// enum
type UTransDirection int32

const (
	UTRANS_FORWARD UTransDirection = 0
	UTRANS_REVERSE UTransDirection = 1
)

// enum
type UStringTrieBuildOption int32

const (
	USTRINGTRIE_BUILD_FAST  UStringTrieBuildOption = 0
	USTRINGTRIE_BUILD_SMALL UStringTrieBuildOption = 1
)

// enum
type UMessagePatternApostropheMode int32

const (
	UMSGPAT_APOS_DOUBLE_OPTIONAL UMessagePatternApostropheMode = 0
	UMSGPAT_APOS_DOUBLE_REQUIRED UMessagePatternApostropheMode = 1
)

// enum
type UMessagePatternPartType int32

const (
	UMSGPAT_PART_TYPE_MSG_START      UMessagePatternPartType = 0
	UMSGPAT_PART_TYPE_MSG_LIMIT      UMessagePatternPartType = 1
	UMSGPAT_PART_TYPE_SKIP_SYNTAX    UMessagePatternPartType = 2
	UMSGPAT_PART_TYPE_INSERT_CHAR    UMessagePatternPartType = 3
	UMSGPAT_PART_TYPE_REPLACE_NUMBER UMessagePatternPartType = 4
	UMSGPAT_PART_TYPE_ARG_START      UMessagePatternPartType = 5
	UMSGPAT_PART_TYPE_ARG_LIMIT      UMessagePatternPartType = 6
	UMSGPAT_PART_TYPE_ARG_NUMBER     UMessagePatternPartType = 7
	UMSGPAT_PART_TYPE_ARG_NAME       UMessagePatternPartType = 8
	UMSGPAT_PART_TYPE_ARG_TYPE       UMessagePatternPartType = 9
	UMSGPAT_PART_TYPE_ARG_STYLE      UMessagePatternPartType = 10
	UMSGPAT_PART_TYPE_ARG_SELECTOR   UMessagePatternPartType = 11
	UMSGPAT_PART_TYPE_ARG_INT        UMessagePatternPartType = 12
	UMSGPAT_PART_TYPE_ARG_DOUBLE     UMessagePatternPartType = 13
)

// enum
type UMessagePatternArgType int32

const (
	UMSGPAT_ARG_TYPE_NONE          UMessagePatternArgType = 0
	UMSGPAT_ARG_TYPE_SIMPLE        UMessagePatternArgType = 1
	UMSGPAT_ARG_TYPE_CHOICE        UMessagePatternArgType = 2
	UMSGPAT_ARG_TYPE_PLURAL        UMessagePatternArgType = 3
	UMSGPAT_ARG_TYPE_SELECT        UMessagePatternArgType = 4
	UMSGPAT_ARG_TYPE_SELECTORDINAL UMessagePatternArgType = 5
)

// enum
type UAlphabeticIndexLabelType int32

const (
	U_ALPHAINDEX_NORMAL    UAlphabeticIndexLabelType = 0
	U_ALPHAINDEX_UNDERFLOW UAlphabeticIndexLabelType = 1
	U_ALPHAINDEX_INFLOW    UAlphabeticIndexLabelType = 2
	U_ALPHAINDEX_OVERFLOW  UAlphabeticIndexLabelType = 3
)

// enum
type UTimeZoneNameType int32

const (
	UTZNM_UNKNOWN           UTimeZoneNameType = 0
	UTZNM_LONG_GENERIC      UTimeZoneNameType = 1
	UTZNM_LONG_STANDARD     UTimeZoneNameType = 2
	UTZNM_LONG_DAYLIGHT     UTimeZoneNameType = 4
	UTZNM_SHORT_GENERIC     UTimeZoneNameType = 8
	UTZNM_SHORT_STANDARD    UTimeZoneNameType = 16
	UTZNM_SHORT_DAYLIGHT    UTimeZoneNameType = 32
	UTZNM_EXEMPLAR_LOCATION UTimeZoneNameType = 64
)

// enum
type UTimeZoneFormatStyle int32

const (
	UTZFMT_STYLE_GENERIC_LOCATION         UTimeZoneFormatStyle = 0
	UTZFMT_STYLE_GENERIC_LONG             UTimeZoneFormatStyle = 1
	UTZFMT_STYLE_GENERIC_SHORT            UTimeZoneFormatStyle = 2
	UTZFMT_STYLE_SPECIFIC_LONG            UTimeZoneFormatStyle = 3
	UTZFMT_STYLE_SPECIFIC_SHORT           UTimeZoneFormatStyle = 4
	UTZFMT_STYLE_LOCALIZED_GMT            UTimeZoneFormatStyle = 5
	UTZFMT_STYLE_LOCALIZED_GMT_SHORT      UTimeZoneFormatStyle = 6
	UTZFMT_STYLE_ISO_BASIC_SHORT          UTimeZoneFormatStyle = 7
	UTZFMT_STYLE_ISO_BASIC_LOCAL_SHORT    UTimeZoneFormatStyle = 8
	UTZFMT_STYLE_ISO_BASIC_FIXED          UTimeZoneFormatStyle = 9
	UTZFMT_STYLE_ISO_BASIC_LOCAL_FIXED    UTimeZoneFormatStyle = 10
	UTZFMT_STYLE_ISO_BASIC_FULL           UTimeZoneFormatStyle = 11
	UTZFMT_STYLE_ISO_BASIC_LOCAL_FULL     UTimeZoneFormatStyle = 12
	UTZFMT_STYLE_ISO_EXTENDED_FIXED       UTimeZoneFormatStyle = 13
	UTZFMT_STYLE_ISO_EXTENDED_LOCAL_FIXED UTimeZoneFormatStyle = 14
	UTZFMT_STYLE_ISO_EXTENDED_FULL        UTimeZoneFormatStyle = 15
	UTZFMT_STYLE_ISO_EXTENDED_LOCAL_FULL  UTimeZoneFormatStyle = 16
	UTZFMT_STYLE_ZONE_ID                  UTimeZoneFormatStyle = 17
	UTZFMT_STYLE_ZONE_ID_SHORT            UTimeZoneFormatStyle = 18
	UTZFMT_STYLE_EXEMPLAR_LOCATION        UTimeZoneFormatStyle = 19
)

// enum
type UTimeZoneFormatGMTOffsetPatternType int32

const (
	UTZFMT_PAT_POSITIVE_HM  UTimeZoneFormatGMTOffsetPatternType = 0
	UTZFMT_PAT_POSITIVE_HMS UTimeZoneFormatGMTOffsetPatternType = 1
	UTZFMT_PAT_NEGATIVE_HM  UTimeZoneFormatGMTOffsetPatternType = 2
	UTZFMT_PAT_NEGATIVE_HMS UTimeZoneFormatGMTOffsetPatternType = 3
	UTZFMT_PAT_POSITIVE_H   UTimeZoneFormatGMTOffsetPatternType = 4
	UTZFMT_PAT_NEGATIVE_H   UTimeZoneFormatGMTOffsetPatternType = 5
	UTZFMT_PAT_COUNT        UTimeZoneFormatGMTOffsetPatternType = 6
)

// enum
type UTimeZoneFormatTimeType int32

const (
	UTZFMT_TIME_TYPE_UNKNOWN  UTimeZoneFormatTimeType = 0
	UTZFMT_TIME_TYPE_STANDARD UTimeZoneFormatTimeType = 1
	UTZFMT_TIME_TYPE_DAYLIGHT UTimeZoneFormatTimeType = 2
)

// enum
type UTimeZoneFormatParseOption int32

const (
	UTZFMT_PARSE_OPTION_NONE                      UTimeZoneFormatParseOption = 0
	UTZFMT_PARSE_OPTION_ALL_STYLES                UTimeZoneFormatParseOption = 1
	UTZFMT_PARSE_OPTION_TZ_DATABASE_ABBREVIATIONS UTimeZoneFormatParseOption = 2
)

// enum
type UMeasureFormatWidth int32

const (
	UMEASFMT_WIDTH_WIDE    UMeasureFormatWidth = 0
	UMEASFMT_WIDTH_SHORT   UMeasureFormatWidth = 1
	UMEASFMT_WIDTH_NARROW  UMeasureFormatWidth = 2
	UMEASFMT_WIDTH_NUMERIC UMeasureFormatWidth = 3
	UMEASFMT_WIDTH_COUNT   UMeasureFormatWidth = 4
)

// enum
type UDateRelativeUnit int32

const (
	UDAT_RELATIVE_SECONDS    UDateRelativeUnit = 0
	UDAT_RELATIVE_MINUTES    UDateRelativeUnit = 1
	UDAT_RELATIVE_HOURS      UDateRelativeUnit = 2
	UDAT_RELATIVE_DAYS       UDateRelativeUnit = 3
	UDAT_RELATIVE_WEEKS      UDateRelativeUnit = 4
	UDAT_RELATIVE_MONTHS     UDateRelativeUnit = 5
	UDAT_RELATIVE_YEARS      UDateRelativeUnit = 6
	UDAT_RELATIVE_UNIT_COUNT UDateRelativeUnit = 7
)

// enum
type UDateAbsoluteUnit int32

const (
	UDAT_ABSOLUTE_SUNDAY     UDateAbsoluteUnit = 0
	UDAT_ABSOLUTE_MONDAY     UDateAbsoluteUnit = 1
	UDAT_ABSOLUTE_TUESDAY    UDateAbsoluteUnit = 2
	UDAT_ABSOLUTE_WEDNESDAY  UDateAbsoluteUnit = 3
	UDAT_ABSOLUTE_THURSDAY   UDateAbsoluteUnit = 4
	UDAT_ABSOLUTE_FRIDAY     UDateAbsoluteUnit = 5
	UDAT_ABSOLUTE_SATURDAY   UDateAbsoluteUnit = 6
	UDAT_ABSOLUTE_DAY        UDateAbsoluteUnit = 7
	UDAT_ABSOLUTE_WEEK       UDateAbsoluteUnit = 8
	UDAT_ABSOLUTE_MONTH      UDateAbsoluteUnit = 9
	UDAT_ABSOLUTE_YEAR       UDateAbsoluteUnit = 10
	UDAT_ABSOLUTE_NOW        UDateAbsoluteUnit = 11
	UDAT_ABSOLUTE_UNIT_COUNT UDateAbsoluteUnit = 12
)

// enum
type UDateDirection int32

const (
	UDAT_DIRECTION_LAST_2 UDateDirection = 0
	UDAT_DIRECTION_LAST   UDateDirection = 1
	UDAT_DIRECTION_THIS   UDateDirection = 2
	UDAT_DIRECTION_NEXT   UDateDirection = 3
	UDAT_DIRECTION_NEXT_2 UDateDirection = 4
	UDAT_DIRECTION_PLAIN  UDateDirection = 5
	UDAT_DIRECTION_COUNT  UDateDirection = 6
)

// enum
type MIMECONTF int32

const (
	MIMECONTF_MAILNEWS         MIMECONTF = 1
	MIMECONTF_BROWSER          MIMECONTF = 2
	MIMECONTF_MINIMAL          MIMECONTF = 4
	MIMECONTF_IMPORT           MIMECONTF = 8
	MIMECONTF_SAVABLE_MAILNEWS MIMECONTF = 256
	MIMECONTF_SAVABLE_BROWSER  MIMECONTF = 512
	MIMECONTF_EXPORT           MIMECONTF = 1024
	MIMECONTF_PRIVCONVERTER    MIMECONTF = 65536
	MIMECONTF_VALID            MIMECONTF = 131072
	MIMECONTF_VALID_NLS        MIMECONTF = 262144
	MIMECONTF_MIME_IE4         MIMECONTF = 268435456
	MIMECONTF_MIME_LATEST      MIMECONTF = 536870912
	MIMECONTF_MIME_REGISTRY    MIMECONTF = 1073741824
)

// enum
type SCRIPTCONTF int32

const (
	SidDefault     SCRIPTCONTF = 0
	SidMerge       SCRIPTCONTF = 1
	SidAsciiSym    SCRIPTCONTF = 2
	SidAsciiLatin  SCRIPTCONTF = 3
	SidLatin       SCRIPTCONTF = 4
	SidGreek       SCRIPTCONTF = 5
	SidCyrillic    SCRIPTCONTF = 6
	SidArmenian    SCRIPTCONTF = 7
	SidHebrew      SCRIPTCONTF = 8
	SidArabic      SCRIPTCONTF = 9
	SidDevanagari  SCRIPTCONTF = 10
	SidBengali     SCRIPTCONTF = 11
	SidGurmukhi    SCRIPTCONTF = 12
	SidGujarati    SCRIPTCONTF = 13
	SidOriya       SCRIPTCONTF = 14
	SidTamil       SCRIPTCONTF = 15
	SidTelugu      SCRIPTCONTF = 16
	SidKannada     SCRIPTCONTF = 17
	SidMalayalam   SCRIPTCONTF = 18
	SidThai        SCRIPTCONTF = 19
	SidLao         SCRIPTCONTF = 20
	SidTibetan     SCRIPTCONTF = 21
	SidGeorgian    SCRIPTCONTF = 22
	SidHangul      SCRIPTCONTF = 23
	SidKana        SCRIPTCONTF = 24
	SidBopomofo    SCRIPTCONTF = 25
	SidHan         SCRIPTCONTF = 26
	SidEthiopic    SCRIPTCONTF = 27
	SidCanSyllabic SCRIPTCONTF = 28
	SidCherokee    SCRIPTCONTF = 29
	SidYi          SCRIPTCONTF = 30
	SidBraille     SCRIPTCONTF = 31
	SidRunic       SCRIPTCONTF = 32
	SidOgham       SCRIPTCONTF = 33
	SidSinhala     SCRIPTCONTF = 34
	SidSyriac      SCRIPTCONTF = 35
	SidBurmese     SCRIPTCONTF = 36
	SidKhmer       SCRIPTCONTF = 37
	SidThaana      SCRIPTCONTF = 38
	SidMongolian   SCRIPTCONTF = 39
	SidUserDefined SCRIPTCONTF = 40
	SidLim         SCRIPTCONTF = 41
	SidFEFirst     SCRIPTCONTF = 23
	SidFELast      SCRIPTCONTF = 26
)

// enum
type TagMLCONVCHARF int32

const (
	MLCONVCHARF_AUTODETECT     TagMLCONVCHARF = 1
	MLCONVCHARF_ENTITIZE       TagMLCONVCHARF = 2
	MLCONVCHARF_NCR_ENTITIZE   TagMLCONVCHARF = 2
	MLCONVCHARF_NAME_ENTITIZE  TagMLCONVCHARF = 4
	MLCONVCHARF_USEDEFCHAR     TagMLCONVCHARF = 8
	MLCONVCHARF_NOBESTFITCHARS TagMLCONVCHARF = 16
	MLCONVCHARF_DETECTJPN      TagMLCONVCHARF = 32
)

// enum
type TagMLCPF int32

const (
	MLDETECTF_MAILNEWS           TagMLCPF = 1
	MLDETECTF_BROWSER            TagMLCPF = 2
	MLDETECTF_VALID              TagMLCPF = 4
	MLDETECTF_VALID_NLS          TagMLCPF = 8
	MLDETECTF_PRESERVE_ORDER     TagMLCPF = 16
	MLDETECTF_PREFERRED_ONLY     TagMLCPF = 32
	MLDETECTF_FILTER_SPECIALCHAR TagMLCPF = 64
	MLDETECTF_EURO_UTF8          TagMLCPF = 128
)

// enum
type MLDETECTCP int32

const (
	MLDETECTCP_NONE   MLDETECTCP = 0
	MLDETECTCP_7BIT   MLDETECTCP = 1
	MLDETECTCP_8BIT   MLDETECTCP = 2
	MLDETECTCP_DBCS   MLDETECTCP = 4
	MLDETECTCP_HTML   MLDETECTCP = 8
	MLDETECTCP_NUMBER MLDETECTCP = 16
)

// enum
type SCRIPTFONTCONTF int32

const (
	SCRIPTCONTF_FIXED_FONT        SCRIPTFONTCONTF = 1
	SCRIPTCONTF_PROPORTIONAL_FONT SCRIPTFONTCONTF = 2
	SCRIPTCONTF_SCRIPT_USER       SCRIPTFONTCONTF = 65536
	SCRIPTCONTF_SCRIPT_HIDE       SCRIPTFONTCONTF = 131072
	SCRIPTCONTF_SCRIPT_SYSTEM     SCRIPTFONTCONTF = 262144
)

// enum
type MLSTR_FLAGS int32

const (
	MLSTR_READ  MLSTR_FLAGS = 1
	MLSTR_WRITE MLSTR_FLAGS = 2
)

// structs

type FONTSIGNATURE struct {
	FsUsb [4]uint32
	FsCsb [2]uint32
}

type CHARSETINFO struct {
	CiCharset uint32
	CiACP     uint32
	Fs        FONTSIGNATURE
}

type LOCALESIGNATURE struct {
	LsUsb          [4]uint32
	LsCsbDefault   [2]uint32
	LsCsbSupported [2]uint32
}

type NEWTEXTMETRICEXA struct {
	NtmTm      NEWTEXTMETRICA
	NtmFontSig FONTSIGNATURE
}

type NEWTEXTMETRICEX = NEWTEXTMETRICEXW
type NEWTEXTMETRICEXW struct {
	NtmTm      NEWTEXTMETRICW
	NtmFontSig FONTSIGNATURE
}

type ENUMTEXTMETRICA struct {
	EtmNewTextMetricEx NEWTEXTMETRICEXA
	EtmAxesList        AXESLISTA
}

type ENUMTEXTMETRIC = ENUMTEXTMETRICW
type ENUMTEXTMETRICW struct {
	EtmNewTextMetricEx NEWTEXTMETRICEXW
	EtmAxesList        AXESLISTW
}

type CPINFO struct {
	MaxCharSize uint32
	DefaultChar [2]byte
	LeadByte    [12]byte
}

type CPINFOEXA struct {
	MaxCharSize        uint32
	DefaultChar        [2]byte
	LeadByte           [12]byte
	UnicodeDefaultChar uint16
	CodePage           uint32
	CodePageName       [260]CHAR
}

type CPINFOEX = CPINFOEXW
type CPINFOEXW struct {
	MaxCharSize        uint32
	DefaultChar        [2]byte
	LeadByte           [12]byte
	UnicodeDefaultChar uint16
	CodePage           uint32
	CodePageName       [260]uint16
}

type NUMBERFMTA struct {
	NumDigits     uint32
	LeadingZero   uint32
	Grouping      uint32
	LpDecimalSep  PSTR
	LpThousandSep PSTR
	NegativeOrder uint32
}

type NUMBERFMT = NUMBERFMTW
type NUMBERFMTW struct {
	NumDigits     uint32
	LeadingZero   uint32
	Grouping      uint32
	LpDecimalSep  PWSTR
	LpThousandSep PWSTR
	NegativeOrder uint32
}

type CURRENCYFMTA struct {
	NumDigits        uint32
	LeadingZero      uint32
	Grouping         uint32
	LpDecimalSep     PSTR
	LpThousandSep    PSTR
	NegativeOrder    uint32
	PositiveOrder    uint32
	LpCurrencySymbol PSTR
}

type CURRENCYFMT = CURRENCYFMTW
type CURRENCYFMTW struct {
	NumDigits        uint32
	LeadingZero      uint32
	Grouping         uint32
	LpDecimalSep     PWSTR
	LpThousandSep    PWSTR
	NegativeOrder    uint32
	PositiveOrder    uint32
	LpCurrencySymbol PWSTR
}

type NLSVERSIONINFO struct {
	DwNLSVersionInfoSize uint32
	DwNLSVersion         uint32
	DwDefinedVersion     uint32
	DwEffectiveId        uint32
	GuidCustomVersion    syscall.GUID
}

type NLSVERSIONINFOEX struct {
	DwNLSVersionInfoSize uint32
	DwNLSVersion         uint32
	DwDefinedVersion     uint32
	DwEffectiveId        uint32
	GuidCustomVersion    syscall.GUID
}

type FILEMUIINFO struct {
	DwSize               uint32
	DwVersion            uint32
	DwFileType           uint32
	PChecksum            [16]byte
	PServiceChecksum     [16]byte
	DwLanguageNameOffset uint32
	DwTypeIDMainSize     uint32
	DwTypeIDMainOffset   uint32
	DwTypeNameMainOffset uint32
	DwTypeIDMUISize      uint32
	DwTypeIDMUIOffset    uint32
	DwTypeNameMUIOffset  uint32
	AbBuffer             [8]byte
}

type MAPPING_SERVICE_INFO struct {
	Size                      uintptr
	PszCopyright              PWSTR
	WMajorVersion             uint16
	WMinorVersion             uint16
	WBuildVersion             uint16
	WStepVersion              uint16
	DwInputContentTypesCount  uint32
	PrgInputContentTypes      *PWSTR
	DwOutputContentTypesCount uint32
	PrgOutputContentTypes     *PWSTR
	DwInputLanguagesCount     uint32
	PrgInputLanguages         *PWSTR
	DwOutputLanguagesCount    uint32
	PrgOutputLanguages        *PWSTR
	DwInputScriptsCount       uint32
	PrgInputScripts           *PWSTR
	DwOutputScriptsCount      uint32
	PrgOutputScripts          *PWSTR
	Guid                      syscall.GUID
	PszCategory               PWSTR
	PszDescription            PWSTR
	DwPrivateDataSize         uint32
	PPrivateData              unsafe.Pointer
	PContext                  unsafe.Pointer
	Bitfield_                 uint32
}

type MAPPING_ENUM_OPTIONS struct {
	Size                 uintptr
	PszCategory          PWSTR
	PszInputLanguage     PWSTR
	PszOutputLanguage    PWSTR
	PszInputScript       PWSTR
	PszOutputScript      PWSTR
	PszInputContentType  PWSTR
	PszOutputContentType PWSTR
	PGuid                *syscall.GUID
	Bitfield_            uint32
}

type MAPPING_OPTIONS struct {
	Size                      uintptr
	PszInputLanguage          PWSTR
	PszOutputLanguage         PWSTR
	PszInputScript            PWSTR
	PszOutputScript           PWSTR
	PszInputContentType       PWSTR
	PszOutputContentType      PWSTR
	PszUILanguage             PWSTR
	PfnRecognizeCallback      PFN_MAPPINGCALLBACKPROC
	PRecognizeCallerData      unsafe.Pointer
	DwRecognizeCallerDataSize uint32
	PfnActionCallback         PFN_MAPPINGCALLBACKPROC
	PActionCallerData         unsafe.Pointer
	DwActionCallerDataSize    uint32
	DwServiceFlag             uint32
	Bitfield_                 uint32
}

type MAPPING_DATA_RANGE struct {
	DwStartIndex          uint32
	DwEndIndex            uint32
	PszDescription        PWSTR
	DwDescriptionLength   uint32
	PData                 unsafe.Pointer
	DwDataSize            uint32
	PszContentType        PWSTR
	PrgActionIds          *PWSTR
	DwActionsCount        uint32
	PrgActionDisplayNames *PWSTR
}

type MAPPING_PROPERTY_BAG struct {
	Size              uintptr
	PrgResultRanges   *MAPPING_DATA_RANGE
	DwRangesCount     uint32
	PServiceData      unsafe.Pointer
	DwServiceDataSize uint32
	PCallerData       unsafe.Pointer
	DwCallerDataSize  uint32
	PContext          unsafe.Pointer
}

type SpellCheckerFactory struct {
}

type SCRIPT_CONTROL struct {
	Bitfield_ uint32
}

type SCRIPT_STATE struct {
	Bitfield_ uint16
}

type SCRIPT_ANALYSIS struct {
	Bitfield_ uint16
	S         SCRIPT_STATE
}

type SCRIPT_ITEM struct {
	ICharPos int32
	A        SCRIPT_ANALYSIS
}

type SCRIPT_VISATTR struct {
	Bitfield_ uint16
}

type GOFFSET struct {
	Du int32
	Dv int32
}

type SCRIPT_LOGATTR struct {
	Bitfield_ byte
}

type SCRIPT_PROPERTIES struct {
	Bitfield1_ uint32
	Bitfield2_ uint32
}

type SCRIPT_FONTPROPERTIES struct {
	CBytes        int32
	WgBlank       uint16
	WgDefault     uint16
	WgInvalid     uint16
	WgKashida     uint16
	IKashidaWidth int32
}

type SCRIPT_TABDEF struct {
	CTabStops  int32
	IScale     int32
	PTabStops  *int32
	ITabOrigin int32
}

type SCRIPT_DIGITSUBSTITUTE struct {
	Bitfield1_ uint32
	Bitfield2_ uint32
	DwReserved uint32
}

type Opentype_feature_record struct {
	TagFeature uint32
	LParameter int32
}

type Textrange_properties struct {
	PotfRecords *Opentype_feature_record
	CotfRecords int32
}

type Script_charprop struct {
	Bitfield_ uint16
}

type Script_glyphprop struct {
	Sva      SCRIPT_VISATTR
	Reserved uint16
}

type UReplaceableCallbacks struct {
	Length   uintptr
	CharAt   uintptr
	Char32At uintptr
	Replace  uintptr
	Extract  uintptr
	Copy     uintptr
}

type UFieldPosition struct {
	Field      int32
	BeginIndex int32
	EndIndex   int32
}

type UCharIterator struct {
	Context       unsafe.Pointer
	Length        int32
	Start         int32
	Index         int32
	Limit         int32
	ReservedField int32
	GetIndex      UCharIteratorGetIndex
	Move          UCharIteratorMove
	HasNext       UCharIteratorHasNext
	HasPrevious   UCharIteratorHasPrevious
	Current       UCharIteratorCurrent
	Next          UCharIteratorNext
	Previous      UCharIteratorPrevious
	ReservedFn    UCharIteratorReserved
	GetState      UCharIteratorGetState
	SetState      UCharIteratorSetState
}

type UEnumeration struct {
}

type UResourceBundle struct {
}

type ULocaleDisplayNames struct {
}

type UCPMap struct {
}

type UCPTrieData struct {
	Data [1]uint64
}

func (this *UCPTrieData) Ptr0() *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *UCPTrieData) Ptr0Val() unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *UCPTrieData) Ptr16() **uint16 {
	return (**uint16)(unsafe.Pointer(this))
}

func (this *UCPTrieData) Ptr16Val() *uint16 {
	return *(**uint16)(unsafe.Pointer(this))
}

func (this *UCPTrieData) Ptr32() **uint32 {
	return (**uint32)(unsafe.Pointer(this))
}

func (this *UCPTrieData) Ptr32Val() *uint32 {
	return *(**uint32)(unsafe.Pointer(this))
}

func (this *UCPTrieData) Ptr8() **byte {
	return (**byte)(unsafe.Pointer(this))
}

func (this *UCPTrieData) Ptr8Val() *byte {
	return *(**byte)(unsafe.Pointer(this))
}

type UCPTrie struct {
	Index              *uint16
	Data               UCPTrieData
	IndexLength        int32
	DataLength         int32
	HighStart          int32
	Shifted12HighStart uint16
	Type_              int8
	ValueWidth         int8
	Reserved32         uint32
	Reserved16         uint16
	Index3NullOffset   uint16
	DataNullOffset     int32
	NullValue          uint32
}

type UMutableCPTrie struct {
}

type UConverter struct {
}

type UConverterFromUnicodeArgs struct {
	Size        uint16
	Flush       int8
	Converter   *UConverter
	Source      *uint16
	SourceLimit *uint16
	Target      PSTR
	TargetLimit PSTR
	Offsets     *int32
}

type UConverterToUnicodeArgs struct {
	Size        uint16
	Flush       int8
	Converter   *UConverter
	Source      PSTR
	SourceLimit PSTR
	Target      *uint16
	TargetLimit *uint16
	Offsets     *int32
}

type USet struct {
}

type UBiDi struct {
}

type UBiDiTransform struct {
}

type UTextFuncs struct {
	TableSize             int32
	Reserved1             int32
	Reserved2             int32
	Reserved3             int32
	Clone                 UTextClone
	NativeLength          UTextNativeLength
	Access                UTextAccess
	Extract               UTextExtract
	Replace               UTextReplace
	Copy                  UTextCopy
	MapOffsetToNative     UTextMapOffsetToNative
	MapNativeIndexToUTF16 UTextMapNativeIndexToUTF16
	Close                 UTextClose
	Spare1                UTextClose
	Spare2                UTextClose
	Spare3                UTextClose
}

type UText struct {
	Magic               uint32
	Flags               int32
	ProviderProperties  int32
	SizeOfStruct        int32
	ChunkNativeLimit    int64
	ExtraSize           int32
	NativeIndexingLimit int32
	ChunkNativeStart    int64
	ChunkOffset         int32
	ChunkLength         int32
	ChunkContents       *uint16
	PFuncs              *UTextFuncs
	PExtra              unsafe.Pointer
	Context             unsafe.Pointer
	P                   unsafe.Pointer
	Q                   unsafe.Pointer
	R                   unsafe.Pointer
	PrivP               unsafe.Pointer
	A                   int64
	B                   int32
	C                   int32
	PrivA               int64
	PrivB               int32
	PrivC               int32
}

type USerializedSet struct {
	Array       *uint16
	BmpLength   int32
	Length      int32
	StaticArray [8]uint16
}

type UNormalizer2 struct {
}

type UConverterSelector struct {
}

type UBreakIterator struct {
}

type UCaseMap struct {
}

type UParseError struct {
	Line        int32
	Offset      int32
	PreContext  [16]uint16
	PostContext [16]uint16
}

type UStringPrepProfile struct {
}

type UIDNA struct {
}

type UIDNAInfo struct {
	Size                    int16
	IsTransitionalDifferent int8
	ReservedB3              int8
	Errors                  uint32
	ReservedI2              int32
	ReservedI3              int32
}

type UCollator struct {
}

type UCollationElements struct {
}

type UCharsetDetector struct {
}

type UCharsetMatch struct {
}

type UFieldPositionIterator struct {
}

type UConstrainedFieldPosition struct {
}

type UFormattedValue struct {
}

type UDateIntervalFormat struct {
}

type UFormattedDateInterval struct {
}

type UGenderInfo struct {
}

type UListFormatter struct {
}

type UFormattedList struct {
}

type ULocaleData struct {
}

type UDateFormatSymbols struct {
}

type UNumberFormatter struct {
}

type UFormattedNumber struct {
}

type UNumberingSystem struct {
}

type UFormattedNumberRange struct {
}

type UPluralRules struct {
}

type URegularExpression struct {
}

type URegion struct {
}

type URelativeDateTimeFormatter struct {
}

type UFormattedRelativeDateTime struct {
}

type UStringSearch struct {
}

type USpoofChecker struct {
}

type USpoofCheckResult struct {
}

type UTransPosition struct {
	ContextStart int32
	ContextLimit int32
	Start        int32
	Limit        int32
}

type UHashtable struct {
}

type USearch struct {
}

type CMLangString struct {
}

type CMLangConvertCharset struct {
}

type CMultiLanguage struct {
}

type MIMECPINFO struct {
	DwFlags             uint32
	UiCodePage          uint32
	UiFamilyCodePage    uint32
	WszDescription      [64]uint16
	WszWebCharset       [50]uint16
	WszHeaderCharset    [50]uint16
	WszBodyCharset      [50]uint16
	WszFixedWidthFont   [32]uint16
	WszProportionalFont [32]uint16
	BGDICharset         byte
}

type MIMECSETINFO struct {
	UiCodePage         uint32
	UiInternetEncoding uint32
	WszCharset         [50]uint16
}

type RFC1766INFO struct {
	Lcid          uint32
	WszRfc1766    [6]uint16
	WszLocaleName [32]uint16
}

type SCRIPTINFO struct {
	ScriptId            byte
	UiCodePage          uint32
	WszDescription      [48]uint16
	WszFixedWidthFont   [32]uint16
	WszProportionalFont [32]uint16
}

type DetectEncodingInfo struct {
	NLangID     uint32
	NCodePage   uint32
	NDocPercent int32
	NConfidence int32
}

type TagSCRIPFONTINFO struct {
	Scripts int64
	WszFont [32]uint16
}

type UNICODERANGE struct {
	WcFrom uint16
	WcTo   uint16
}

// func types

type LOCALE_ENUMPROCA = uintptr
type LOCALE_ENUMPROCA_func = func(param0 PSTR) BOOL

type LOCALE_ENUMPROCW = uintptr
type LOCALE_ENUMPROCW_func = func(param0 PWSTR) BOOL

type LANGUAGEGROUP_ENUMPROCA = uintptr
type LANGUAGEGROUP_ENUMPROCA_func = func(param0 uint32, param1 PSTR, param2 PSTR, param3 uint32, param4 uintptr) BOOL

type LANGGROUPLOCALE_ENUMPROCA = uintptr
type LANGGROUPLOCALE_ENUMPROCA_func = func(param0 uint32, param1 uint32, param2 PSTR, param3 uintptr) BOOL

type UILANGUAGE_ENUMPROCA = uintptr
type UILANGUAGE_ENUMPROCA_func = func(param0 PSTR, param1 uintptr) BOOL

type CODEPAGE_ENUMPROCA = uintptr
type CODEPAGE_ENUMPROCA_func = func(param0 PSTR) BOOL

type DATEFMT_ENUMPROCA = uintptr
type DATEFMT_ENUMPROCA_func = func(param0 PSTR) BOOL

type DATEFMT_ENUMPROCEXA = uintptr
type DATEFMT_ENUMPROCEXA_func = func(param0 PSTR, param1 uint32) BOOL

type TIMEFMT_ENUMPROCA = uintptr
type TIMEFMT_ENUMPROCA_func = func(param0 PSTR) BOOL

type CALINFO_ENUMPROCA = uintptr
type CALINFO_ENUMPROCA_func = func(param0 PSTR) BOOL

type CALINFO_ENUMPROCEXA = uintptr
type CALINFO_ENUMPROCEXA_func = func(param0 PSTR, param1 uint32) BOOL

type LANGUAGEGROUP_ENUMPROCW = uintptr
type LANGUAGEGROUP_ENUMPROCW_func = func(param0 uint32, param1 PWSTR, param2 PWSTR, param3 uint32, param4 uintptr) BOOL

type LANGGROUPLOCALE_ENUMPROCW = uintptr
type LANGGROUPLOCALE_ENUMPROCW_func = func(param0 uint32, param1 uint32, param2 PWSTR, param3 uintptr) BOOL

type UILANGUAGE_ENUMPROCW = uintptr
type UILANGUAGE_ENUMPROCW_func = func(param0 PWSTR, param1 uintptr) BOOL

type CODEPAGE_ENUMPROCW = uintptr
type CODEPAGE_ENUMPROCW_func = func(param0 PWSTR) BOOL

type DATEFMT_ENUMPROCW = uintptr
type DATEFMT_ENUMPROCW_func = func(param0 PWSTR) BOOL

type DATEFMT_ENUMPROCEXW = uintptr
type DATEFMT_ENUMPROCEXW_func = func(param0 PWSTR, param1 uint32) BOOL

type TIMEFMT_ENUMPROCW = uintptr
type TIMEFMT_ENUMPROCW_func = func(param0 PWSTR) BOOL

type CALINFO_ENUMPROCW = uintptr
type CALINFO_ENUMPROCW_func = func(param0 PWSTR) BOOL

type CALINFO_ENUMPROCEXW = uintptr
type CALINFO_ENUMPROCEXW_func = func(param0 PWSTR, param1 uint32) BOOL

type GEO_ENUMPROC = uintptr
type GEO_ENUMPROC_func = func(param0 int32) BOOL

type GEO_ENUMNAMEPROC = uintptr
type GEO_ENUMNAMEPROC_func = func(param0 PWSTR, param1 LPARAM) BOOL

type CALINFO_ENUMPROCEXEX = uintptr
type CALINFO_ENUMPROCEXEX_func = func(param0 PWSTR, param1 uint32, param2 PWSTR, param3 LPARAM) BOOL

type DATEFMT_ENUMPROCEXEX = uintptr
type DATEFMT_ENUMPROCEXEX_func = func(param0 PWSTR, param1 uint32, param2 LPARAM) BOOL

type TIMEFMT_ENUMPROCEX = uintptr
type TIMEFMT_ENUMPROCEX_func = func(param0 PWSTR, param1 LPARAM) BOOL

type LOCALE_ENUMPROCEX = uintptr
type LOCALE_ENUMPROCEX_func = func(param0 PWSTR, param1 uint32, param2 LPARAM) BOOL

type PFN_MAPPINGCALLBACKPROC = uintptr
type PFN_MAPPINGCALLBACKPROC_func = func(pBag *MAPPING_PROPERTY_BAG, data unsafe.Pointer, dwDataSize uint32, Result HRESULT)

type UTraceEntry = uintptr
type UTraceEntry_func = func(context unsafe.Pointer, fnNumber int32)

type UTraceExit = uintptr
type UTraceExit_func = func(context unsafe.Pointer, fnNumber int32, fmt PSTR, args *int8)

type UTraceData = uintptr
type UTraceData_func = func(context unsafe.Pointer, fnNumber int32, level int32, fmt PSTR, args *int8)

type UCharIteratorGetIndex = uintptr
type UCharIteratorGetIndex_func = func(iter *UCharIterator, origin UCharIteratorOrigin) int32

type UCharIteratorMove = uintptr
type UCharIteratorMove_func = func(iter *UCharIterator, delta int32, origin UCharIteratorOrigin) int32

type UCharIteratorHasNext = uintptr
type UCharIteratorHasNext_func = func(iter *UCharIterator) int8

type UCharIteratorHasPrevious = uintptr
type UCharIteratorHasPrevious_func = func(iter *UCharIterator) int8

type UCharIteratorCurrent = uintptr
type UCharIteratorCurrent_func = func(iter *UCharIterator) int32

type UCharIteratorNext = uintptr
type UCharIteratorNext_func = func(iter *UCharIterator) int32

type UCharIteratorPrevious = uintptr
type UCharIteratorPrevious_func = func(iter *UCharIterator) int32

type UCharIteratorReserved = uintptr
type UCharIteratorReserved_func = func(iter *UCharIterator, something int32) int32

type UCharIteratorGetState = uintptr
type UCharIteratorGetState_func = func(iter *UCharIterator) uint32

type UCharIteratorSetState = uintptr
type UCharIteratorSetState_func = func(iter *UCharIterator, state uint32, pErrorCode *UErrorCode)

type UCPMapValueFilter = uintptr
type UCPMapValueFilter_func = func(context unsafe.Pointer, value uint32) uint32

type UConverterToUCallback = uintptr
type UConverterToUCallback_func = func(context unsafe.Pointer, args *UConverterToUnicodeArgs, codeUnits PSTR, length int32, reason UConverterCallbackReason, pErrorCode *UErrorCode)

type UConverterFromUCallback = uintptr
type UConverterFromUCallback_func = func(context unsafe.Pointer, args *UConverterFromUnicodeArgs, codeUnits *uint16, length int32, codePoint int32, reason UConverterCallbackReason, pErrorCode *UErrorCode)

type UMemAllocFn = uintptr
type UMemAllocFn_func = func(context unsafe.Pointer, size uintptr) unsafe.Pointer

type UMemReallocFn = uintptr
type UMemReallocFn_func = func(context unsafe.Pointer, mem unsafe.Pointer, size uintptr) unsafe.Pointer

type UMemFreeFn = uintptr
type UMemFreeFn_func = func(context unsafe.Pointer, mem unsafe.Pointer)

type UCharEnumTypeRange = uintptr
type UCharEnumTypeRange_func = func(context unsafe.Pointer, start int32, limit int32, type_ UCharCategory) int8

type UEnumCharNamesFn = uintptr
type UEnumCharNamesFn_func = func(context unsafe.Pointer, code int32, nameChoice UCharNameChoice, name PSTR, length int32) int8

type UBiDiClassCallback = uintptr
type UBiDiClassCallback_func = func(context unsafe.Pointer, c int32) UCharDirection

type UTextClone = uintptr
type UTextClone_func = func(dest *UText, src *UText, deep int8, status *UErrorCode) *UText

type UTextNativeLength = uintptr
type UTextNativeLength_func = func(ut *UText) int64

type UTextAccess = uintptr
type UTextAccess_func = func(ut *UText, nativeIndex int64, forward int8) int8

type UTextExtract = uintptr
type UTextExtract_func = func(ut *UText, nativeStart int64, nativeLimit int64, dest *uint16, destCapacity int32, status *UErrorCode) int32

type UTextReplace = uintptr
type UTextReplace_func = func(ut *UText, nativeStart int64, nativeLimit int64, replacementText *uint16, replacmentLength int32, status *UErrorCode) int32

type UTextCopy = uintptr
type UTextCopy_func = func(ut *UText, nativeStart int64, nativeLimit int64, nativeDest int64, move int8, status *UErrorCode)

type UTextMapOffsetToNative = uintptr
type UTextMapOffsetToNative_func = func(ut *UText) int64

type UTextMapNativeIndexToUTF16 = uintptr
type UTextMapNativeIndexToUTF16_func = func(ut *UText, nativeIndex int64) int32

type UTextClose = uintptr
type UTextClose_func = func(ut *UText)

type UNESCAPE_CHAR_AT = uintptr
type UNESCAPE_CHAR_AT_func = func(offset int32, context unsafe.Pointer) uint16

type URegexMatchCallback = uintptr
type URegexMatchCallback_func = func(context unsafe.Pointer, steps int32) int8

type URegexFindProgressCallback = uintptr
type URegexFindProgressCallback_func = func(context unsafe.Pointer, matchIndex int64) int8

type UStringCaseMapper = uintptr
type UStringCaseMapper_func = func(csm *UCaseMap, dest *uint16, destCapacity int32, src *uint16, srcLength int32, pErrorCode *UErrorCode) int32

// interfaces

// B7C82D61-FBE8-4B47-9B27-6C0D2E0DE0A3
var IID_ISpellingError = syscall.GUID{0xB7C82D61, 0xFBE8, 0x4B47,
	[8]byte{0x9B, 0x27, 0x6C, 0x0D, 0x2E, 0x0D, 0xE0, 0xA3}}

type ISpellingErrorInterface interface {
	IUnknownInterface
	Get_StartIndex(value *uint32) HRESULT
	Get_Length(value *uint32) HRESULT
	Get_CorrectiveAction(value *CORRECTIVE_ACTION) HRESULT
	Get_Replacement(value *PWSTR) HRESULT
}

type ISpellingErrorVtbl struct {
	IUnknownVtbl
	Get_StartIndex       uintptr
	Get_Length           uintptr
	Get_CorrectiveAction uintptr
	Get_Replacement      uintptr
}

type ISpellingError struct {
	IUnknown
}

func (this *ISpellingError) Vtbl() *ISpellingErrorVtbl {
	return (*ISpellingErrorVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISpellingError) Get_StartIndex(value *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_StartIndex, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellingError) Get_Length(value *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Length, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellingError) Get_CorrectiveAction(value *CORRECTIVE_ACTION) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CorrectiveAction, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellingError) Get_Replacement(value *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Replacement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

// 803E3BD4-2828-4410-8290-418D1D73C762
var IID_IEnumSpellingError = syscall.GUID{0x803E3BD4, 0x2828, 0x4410,
	[8]byte{0x82, 0x90, 0x41, 0x8D, 0x1D, 0x73, 0xC7, 0x62}}

type IEnumSpellingErrorInterface interface {
	IUnknownInterface
	Next(value **ISpellingError) HRESULT
}

type IEnumSpellingErrorVtbl struct {
	IUnknownVtbl
	Next uintptr
}

type IEnumSpellingError struct {
	IUnknown
}

func (this *IEnumSpellingError) Vtbl() *IEnumSpellingErrorVtbl {
	return (*IEnumSpellingErrorVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumSpellingError) Next(value **ISpellingError) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

// 432E5F85-35CF-4606-A801-6F70277E1D7A
var IID_IOptionDescription = syscall.GUID{0x432E5F85, 0x35CF, 0x4606,
	[8]byte{0xA8, 0x01, 0x6F, 0x70, 0x27, 0x7E, 0x1D, 0x7A}}

type IOptionDescriptionInterface interface {
	IUnknownInterface
	Get_Id(value *PWSTR) HRESULT
	Get_Heading(value *PWSTR) HRESULT
	Get_Description(value *PWSTR) HRESULT
	Get_Labels(value **IEnumString) HRESULT
}

type IOptionDescriptionVtbl struct {
	IUnknownVtbl
	Get_Id          uintptr
	Get_Heading     uintptr
	Get_Description uintptr
	Get_Labels      uintptr
}

type IOptionDescription struct {
	IUnknown
}

func (this *IOptionDescription) Vtbl() *IOptionDescriptionVtbl {
	return (*IOptionDescriptionVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOptionDescription) Get_Id(value *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Id, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *IOptionDescription) Get_Heading(value *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Heading, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *IOptionDescription) Get_Description(value *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Description, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *IOptionDescription) Get_Labels(value **IEnumString) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Labels, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

// 0B83A5B0-792F-4EAB-9799-ACF52C5ED08A
var IID_ISpellCheckerChangedEventHandler = syscall.GUID{0x0B83A5B0, 0x792F, 0x4EAB,
	[8]byte{0x97, 0x99, 0xAC, 0xF5, 0x2C, 0x5E, 0xD0, 0x8A}}

type ISpellCheckerChangedEventHandlerInterface interface {
	IUnknownInterface
	Invoke(sender *ISpellChecker) HRESULT
}

type ISpellCheckerChangedEventHandlerVtbl struct {
	IUnknownVtbl
	Invoke uintptr
}

type ISpellCheckerChangedEventHandler struct {
	IUnknown
}

func (this *ISpellCheckerChangedEventHandler) Vtbl() *ISpellCheckerChangedEventHandlerVtbl {
	return (*ISpellCheckerChangedEventHandlerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISpellCheckerChangedEventHandler) Invoke(sender *ISpellChecker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Invoke, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(sender)))
	return HRESULT(ret)
}

// B6FD0B71-E2BC-4653-8D05-F197E412770B
var IID_ISpellChecker = syscall.GUID{0xB6FD0B71, 0xE2BC, 0x4653,
	[8]byte{0x8D, 0x05, 0xF1, 0x97, 0xE4, 0x12, 0x77, 0x0B}}

type ISpellCheckerInterface interface {
	IUnknownInterface
	Get_LanguageTag(value *PWSTR) HRESULT
	Check(text PWSTR, value **IEnumSpellingError) HRESULT
	Suggest(word PWSTR, value **IEnumString) HRESULT
	Add(word PWSTR) HRESULT
	Ignore(word PWSTR) HRESULT
	AutoCorrect(from PWSTR, to PWSTR) HRESULT
	GetOptionValue(optionId PWSTR, value *byte) HRESULT
	Get_OptionIds(value **IEnumString) HRESULT
	Get_Id(value *PWSTR) HRESULT
	Get_LocalizedName(value *PWSTR) HRESULT
	Add_SpellCheckerChanged(handler *ISpellCheckerChangedEventHandler, eventCookie *uint32) HRESULT
	Remove_SpellCheckerChanged(eventCookie uint32) HRESULT
	GetOptionDescription(optionId PWSTR, value **IOptionDescription) HRESULT
	ComprehensiveCheck(text PWSTR, value **IEnumSpellingError) HRESULT
}

type ISpellCheckerVtbl struct {
	IUnknownVtbl
	Get_LanguageTag            uintptr
	Check                      uintptr
	Suggest                    uintptr
	Add                        uintptr
	Ignore                     uintptr
	AutoCorrect                uintptr
	GetOptionValue             uintptr
	Get_OptionIds              uintptr
	Get_Id                     uintptr
	Get_LocalizedName          uintptr
	Add_SpellCheckerChanged    uintptr
	Remove_SpellCheckerChanged uintptr
	GetOptionDescription       uintptr
	ComprehensiveCheck         uintptr
}

type ISpellChecker struct {
	IUnknown
}

func (this *ISpellChecker) Vtbl() *ISpellCheckerVtbl {
	return (*ISpellCheckerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISpellChecker) Get_LanguageTag(value *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_LanguageTag, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellChecker) Check(text PWSTR, value **IEnumSpellingError) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Check, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellChecker) Suggest(word PWSTR, value **IEnumString) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Suggest, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(word)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellChecker) Add(word PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Add, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(word)))
	return HRESULT(ret)
}

func (this *ISpellChecker) Ignore(word PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Ignore, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(word)))
	return HRESULT(ret)
}

func (this *ISpellChecker) AutoCorrect(from PWSTR, to PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AutoCorrect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(from)), uintptr(unsafe.Pointer(to)))
	return HRESULT(ret)
}

func (this *ISpellChecker) GetOptionValue(optionId PWSTR, value *byte) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetOptionValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(optionId)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellChecker) Get_OptionIds(value **IEnumString) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_OptionIds, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellChecker) Get_Id(value *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Id, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellChecker) Get_LocalizedName(value *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_LocalizedName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellChecker) Add_SpellCheckerChanged(handler *ISpellCheckerChangedEventHandler, eventCookie *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Add_SpellCheckerChanged, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(handler)), uintptr(unsafe.Pointer(eventCookie)))
	return HRESULT(ret)
}

func (this *ISpellChecker) Remove_SpellCheckerChanged(eventCookie uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Remove_SpellCheckerChanged, uintptr(unsafe.Pointer(this)), uintptr(eventCookie))
	return HRESULT(ret)
}

func (this *ISpellChecker) GetOptionDescription(optionId PWSTR, value **IOptionDescription) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetOptionDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(optionId)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellChecker) ComprehensiveCheck(text PWSTR, value **IEnumSpellingError) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ComprehensiveCheck, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

// E7ED1C71-87F7-4378-A840-C9200DACEE47
var IID_ISpellChecker2 = syscall.GUID{0xE7ED1C71, 0x87F7, 0x4378,
	[8]byte{0xA8, 0x40, 0xC9, 0x20, 0x0D, 0xAC, 0xEE, 0x47}}

type ISpellChecker2Interface interface {
	ISpellCheckerInterface
	Remove(word PWSTR) HRESULT
}

type ISpellChecker2Vtbl struct {
	ISpellCheckerVtbl
	Remove uintptr
}

type ISpellChecker2 struct {
	ISpellChecker
}

func (this *ISpellChecker2) Vtbl() *ISpellChecker2Vtbl {
	return (*ISpellChecker2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISpellChecker2) Remove(word PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Remove, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(word)))
	return HRESULT(ret)
}

// 8E018A9D-2415-4677-BF08-794EA61F94BB
var IID_ISpellCheckerFactory = syscall.GUID{0x8E018A9D, 0x2415, 0x4677,
	[8]byte{0xBF, 0x08, 0x79, 0x4E, 0xA6, 0x1F, 0x94, 0xBB}}

type ISpellCheckerFactoryInterface interface {
	IUnknownInterface
	Get_SupportedLanguages(value **IEnumString) HRESULT
	IsSupported(languageTag PWSTR, value *BOOL) HRESULT
	CreateSpellChecker(languageTag PWSTR, value **ISpellChecker) HRESULT
}

type ISpellCheckerFactoryVtbl struct {
	IUnknownVtbl
	Get_SupportedLanguages uintptr
	IsSupported            uintptr
	CreateSpellChecker     uintptr
}

type ISpellCheckerFactory struct {
	IUnknown
}

func (this *ISpellCheckerFactory) Vtbl() *ISpellCheckerFactoryVtbl {
	return (*ISpellCheckerFactoryVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISpellCheckerFactory) Get_SupportedLanguages(value **IEnumString) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_SupportedLanguages, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellCheckerFactory) IsSupported(languageTag PWSTR, value *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsSupported, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(languageTag)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellCheckerFactory) CreateSpellChecker(languageTag PWSTR, value **ISpellChecker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateSpellChecker, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(languageTag)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

// AA176B85-0E12-4844-8E1A-EEF1DA77F586
var IID_IUserDictionariesRegistrar = syscall.GUID{0xAA176B85, 0x0E12, 0x4844,
	[8]byte{0x8E, 0x1A, 0xEE, 0xF1, 0xDA, 0x77, 0xF5, 0x86}}

type IUserDictionariesRegistrarInterface interface {
	IUnknownInterface
	RegisterUserDictionary(dictionaryPath PWSTR, languageTag PWSTR) HRESULT
	UnregisterUserDictionary(dictionaryPath PWSTR, languageTag PWSTR) HRESULT
}

type IUserDictionariesRegistrarVtbl struct {
	IUnknownVtbl
	RegisterUserDictionary   uintptr
	UnregisterUserDictionary uintptr
}

type IUserDictionariesRegistrar struct {
	IUnknown
}

func (this *IUserDictionariesRegistrar) Vtbl() *IUserDictionariesRegistrarVtbl {
	return (*IUserDictionariesRegistrarVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUserDictionariesRegistrar) RegisterUserDictionary(dictionaryPath PWSTR, languageTag PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RegisterUserDictionary, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(dictionaryPath)), uintptr(unsafe.Pointer(languageTag)))
	return HRESULT(ret)
}

func (this *IUserDictionariesRegistrar) UnregisterUserDictionary(dictionaryPath PWSTR, languageTag PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UnregisterUserDictionary, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(dictionaryPath)), uintptr(unsafe.Pointer(languageTag)))
	return HRESULT(ret)
}

// 73E976E0-8ED4-4EB1-80D7-1BE0A16B0C38
var IID_ISpellCheckProvider = syscall.GUID{0x73E976E0, 0x8ED4, 0x4EB1,
	[8]byte{0x80, 0xD7, 0x1B, 0xE0, 0xA1, 0x6B, 0x0C, 0x38}}

type ISpellCheckProviderInterface interface {
	IUnknownInterface
	Get_LanguageTag(value *PWSTR) HRESULT
	Check(text PWSTR, value **IEnumSpellingError) HRESULT
	Suggest(word PWSTR, value **IEnumString) HRESULT
	GetOptionValue(optionId PWSTR, value *byte) HRESULT
	SetOptionValue(optionId PWSTR, value byte) HRESULT
	Get_OptionIds(value **IEnumString) HRESULT
	Get_Id(value *PWSTR) HRESULT
	Get_LocalizedName(value *PWSTR) HRESULT
	GetOptionDescription(optionId PWSTR, value **IOptionDescription) HRESULT
	InitializeWordlist(wordlistType WORDLIST_TYPE, words *IEnumString) HRESULT
}

type ISpellCheckProviderVtbl struct {
	IUnknownVtbl
	Get_LanguageTag      uintptr
	Check                uintptr
	Suggest              uintptr
	GetOptionValue       uintptr
	SetOptionValue       uintptr
	Get_OptionIds        uintptr
	Get_Id               uintptr
	Get_LocalizedName    uintptr
	GetOptionDescription uintptr
	InitializeWordlist   uintptr
}

type ISpellCheckProvider struct {
	IUnknown
}

func (this *ISpellCheckProvider) Vtbl() *ISpellCheckProviderVtbl {
	return (*ISpellCheckProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISpellCheckProvider) Get_LanguageTag(value *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_LanguageTag, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellCheckProvider) Check(text PWSTR, value **IEnumSpellingError) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Check, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellCheckProvider) Suggest(word PWSTR, value **IEnumString) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Suggest, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(word)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellCheckProvider) GetOptionValue(optionId PWSTR, value *byte) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetOptionValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(optionId)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellCheckProvider) SetOptionValue(optionId PWSTR, value byte) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetOptionValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(optionId)), uintptr(value))
	return HRESULT(ret)
}

func (this *ISpellCheckProvider) Get_OptionIds(value **IEnumString) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_OptionIds, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellCheckProvider) Get_Id(value *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Id, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellCheckProvider) Get_LocalizedName(value *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_LocalizedName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellCheckProvider) GetOptionDescription(optionId PWSTR, value **IOptionDescription) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetOptionDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(optionId)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellCheckProvider) InitializeWordlist(wordlistType WORDLIST_TYPE, words *IEnumString) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InitializeWordlist, uintptr(unsafe.Pointer(this)), uintptr(wordlistType), uintptr(unsafe.Pointer(words)))
	return HRESULT(ret)
}

// 0C58F8DE-8E94-479E-9717-70C42C4AD2C3
var IID_IComprehensiveSpellCheckProvider = syscall.GUID{0x0C58F8DE, 0x8E94, 0x479E,
	[8]byte{0x97, 0x17, 0x70, 0xC4, 0x2C, 0x4A, 0xD2, 0xC3}}

type IComprehensiveSpellCheckProviderInterface interface {
	IUnknownInterface
	ComprehensiveCheck(text PWSTR, value **IEnumSpellingError) HRESULT
}

type IComprehensiveSpellCheckProviderVtbl struct {
	IUnknownVtbl
	ComprehensiveCheck uintptr
}

type IComprehensiveSpellCheckProvider struct {
	IUnknown
}

func (this *IComprehensiveSpellCheckProvider) Vtbl() *IComprehensiveSpellCheckProviderVtbl {
	return (*IComprehensiveSpellCheckProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IComprehensiveSpellCheckProvider) ComprehensiveCheck(text PWSTR, value **IEnumSpellingError) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ComprehensiveCheck, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

// 9F671E11-77D6-4C92-AEFB-615215E3A4BE
var IID_ISpellCheckProviderFactory = syscall.GUID{0x9F671E11, 0x77D6, 0x4C92,
	[8]byte{0xAE, 0xFB, 0x61, 0x52, 0x15, 0xE3, 0xA4, 0xBE}}

type ISpellCheckProviderFactoryInterface interface {
	IUnknownInterface
	Get_SupportedLanguages(value **IEnumString) HRESULT
	IsSupported(languageTag PWSTR, value *BOOL) HRESULT
	CreateSpellCheckProvider(languageTag PWSTR, value **ISpellCheckProvider) HRESULT
}

type ISpellCheckProviderFactoryVtbl struct {
	IUnknownVtbl
	Get_SupportedLanguages   uintptr
	IsSupported              uintptr
	CreateSpellCheckProvider uintptr
}

type ISpellCheckProviderFactory struct {
	IUnknown
}

func (this *ISpellCheckProviderFactory) Vtbl() *ISpellCheckProviderFactoryVtbl {
	return (*ISpellCheckProviderFactoryVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISpellCheckProviderFactory) Get_SupportedLanguages(value **IEnumString) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_SupportedLanguages, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellCheckProviderFactory) IsSupported(languageTag PWSTR, value *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsSupported, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(languageTag)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ISpellCheckProviderFactory) CreateSpellCheckProvider(languageTag PWSTR, value **ISpellCheckProvider) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateSpellCheckProvider, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(languageTag)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

// D24ACD21-BA72-11D0-B188-00AA0038C969
var IID_IMLangStringBufW = syscall.GUID{0xD24ACD21, 0xBA72, 0x11D0,
	[8]byte{0xB1, 0x88, 0x00, 0xAA, 0x00, 0x38, 0xC9, 0x69}}

type IMLangStringBufWInterface interface {
	IUnknownInterface
	GetStatus(plFlags *int32, pcchBuf *int32) HRESULT
	LockBuf(cchOffset int32, cchMaxLock int32, ppszBuf **uint16, pcchBuf *int32) HRESULT
	UnlockBuf(pszBuf PWSTR, cchOffset int32, cchWrite int32) HRESULT
	Insert(cchOffset int32, cchMaxInsert int32, pcchActual *int32) HRESULT
	Delete(cchOffset int32, cchDelete int32) HRESULT
}

type IMLangStringBufWVtbl struct {
	IUnknownVtbl
	GetStatus uintptr
	LockBuf   uintptr
	UnlockBuf uintptr
	Insert    uintptr
	Delete    uintptr
}

type IMLangStringBufW struct {
	IUnknown
}

func (this *IMLangStringBufW) Vtbl() *IMLangStringBufWVtbl {
	return (*IMLangStringBufWVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMLangStringBufW) GetStatus(plFlags *int32, pcchBuf *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStatus, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(plFlags)), uintptr(unsafe.Pointer(pcchBuf)))
	return HRESULT(ret)
}

func (this *IMLangStringBufW) LockBuf(cchOffset int32, cchMaxLock int32, ppszBuf **uint16, pcchBuf *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().LockBuf, uintptr(unsafe.Pointer(this)), uintptr(cchOffset), uintptr(cchMaxLock), uintptr(unsafe.Pointer(ppszBuf)), uintptr(unsafe.Pointer(pcchBuf)))
	return HRESULT(ret)
}

func (this *IMLangStringBufW) UnlockBuf(pszBuf PWSTR, cchOffset int32, cchWrite int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UnlockBuf, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszBuf)), uintptr(cchOffset), uintptr(cchWrite))
	return HRESULT(ret)
}

func (this *IMLangStringBufW) Insert(cchOffset int32, cchMaxInsert int32, pcchActual *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Insert, uintptr(unsafe.Pointer(this)), uintptr(cchOffset), uintptr(cchMaxInsert), uintptr(unsafe.Pointer(pcchActual)))
	return HRESULT(ret)
}

func (this *IMLangStringBufW) Delete(cchOffset int32, cchDelete int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Delete, uintptr(unsafe.Pointer(this)), uintptr(cchOffset), uintptr(cchDelete))
	return HRESULT(ret)
}

// D24ACD23-BA72-11D0-B188-00AA0038C969
var IID_IMLangStringBufA = syscall.GUID{0xD24ACD23, 0xBA72, 0x11D0,
	[8]byte{0xB1, 0x88, 0x00, 0xAA, 0x00, 0x38, 0xC9, 0x69}}

type IMLangStringBufAInterface interface {
	IUnknownInterface
	GetStatus(plFlags *int32, pcchBuf *int32) HRESULT
	LockBuf(cchOffset int32, cchMaxLock int32, ppszBuf **CHAR, pcchBuf *int32) HRESULT
	UnlockBuf(pszBuf PSTR, cchOffset int32, cchWrite int32) HRESULT
	Insert(cchOffset int32, cchMaxInsert int32, pcchActual *int32) HRESULT
	Delete(cchOffset int32, cchDelete int32) HRESULT
}

type IMLangStringBufAVtbl struct {
	IUnknownVtbl
	GetStatus uintptr
	LockBuf   uintptr
	UnlockBuf uintptr
	Insert    uintptr
	Delete    uintptr
}

type IMLangStringBufA struct {
	IUnknown
}

func (this *IMLangStringBufA) Vtbl() *IMLangStringBufAVtbl {
	return (*IMLangStringBufAVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMLangStringBufA) GetStatus(plFlags *int32, pcchBuf *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStatus, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(plFlags)), uintptr(unsafe.Pointer(pcchBuf)))
	return HRESULT(ret)
}

func (this *IMLangStringBufA) LockBuf(cchOffset int32, cchMaxLock int32, ppszBuf **CHAR, pcchBuf *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().LockBuf, uintptr(unsafe.Pointer(this)), uintptr(cchOffset), uintptr(cchMaxLock), uintptr(unsafe.Pointer(ppszBuf)), uintptr(unsafe.Pointer(pcchBuf)))
	return HRESULT(ret)
}

func (this *IMLangStringBufA) UnlockBuf(pszBuf PSTR, cchOffset int32, cchWrite int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UnlockBuf, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszBuf)), uintptr(cchOffset), uintptr(cchWrite))
	return HRESULT(ret)
}

func (this *IMLangStringBufA) Insert(cchOffset int32, cchMaxInsert int32, pcchActual *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Insert, uintptr(unsafe.Pointer(this)), uintptr(cchOffset), uintptr(cchMaxInsert), uintptr(unsafe.Pointer(pcchActual)))
	return HRESULT(ret)
}

func (this *IMLangStringBufA) Delete(cchOffset int32, cchDelete int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Delete, uintptr(unsafe.Pointer(this)), uintptr(cchOffset), uintptr(cchDelete))
	return HRESULT(ret)
}

// C04D65CE-B70D-11D0-B188-00AA0038C969
var IID_IMLangString = syscall.GUID{0xC04D65CE, 0xB70D, 0x11D0,
	[8]byte{0xB1, 0x88, 0x00, 0xAA, 0x00, 0x38, 0xC9, 0x69}}

type IMLangStringInterface interface {
	IUnknownInterface
	Sync(fNoAccess BOOL) HRESULT
	GetLength(plLen *int32) HRESULT
	SetMLStr(lDestPos int32, lDestLen int32, pSrcMLStr *IUnknown, lSrcPos int32, lSrcLen int32) HRESULT
	GetMLStr(lSrcPos int32, lSrcLen int32, pUnkOuter *IUnknown, dwClsContext uint32, piid *syscall.GUID, ppDestMLStr **IUnknown, plDestPos *int32, plDestLen *int32) HRESULT
}

type IMLangStringVtbl struct {
	IUnknownVtbl
	Sync      uintptr
	GetLength uintptr
	SetMLStr  uintptr
	GetMLStr  uintptr
}

type IMLangString struct {
	IUnknown
}

func (this *IMLangString) Vtbl() *IMLangStringVtbl {
	return (*IMLangStringVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMLangString) Sync(fNoAccess BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Sync, uintptr(unsafe.Pointer(this)), uintptr(fNoAccess))
	return HRESULT(ret)
}

func (this *IMLangString) GetLength(plLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLength, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(plLen)))
	return HRESULT(ret)
}

func (this *IMLangString) SetMLStr(lDestPos int32, lDestLen int32, pSrcMLStr *IUnknown, lSrcPos int32, lSrcLen int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetMLStr, uintptr(unsafe.Pointer(this)), uintptr(lDestPos), uintptr(lDestLen), uintptr(unsafe.Pointer(pSrcMLStr)), uintptr(lSrcPos), uintptr(lSrcLen))
	return HRESULT(ret)
}

func (this *IMLangString) GetMLStr(lSrcPos int32, lSrcLen int32, pUnkOuter *IUnknown, dwClsContext uint32, piid *syscall.GUID, ppDestMLStr **IUnknown, plDestPos *int32, plDestLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetMLStr, uintptr(unsafe.Pointer(this)), uintptr(lSrcPos), uintptr(lSrcLen), uintptr(unsafe.Pointer(pUnkOuter)), uintptr(dwClsContext), uintptr(unsafe.Pointer(piid)), uintptr(unsafe.Pointer(ppDestMLStr)), uintptr(unsafe.Pointer(plDestPos)), uintptr(unsafe.Pointer(plDestLen)))
	return HRESULT(ret)
}

// C04D65D0-B70D-11D0-B188-00AA0038C969
var IID_IMLangStringWStr = syscall.GUID{0xC04D65D0, 0xB70D, 0x11D0,
	[8]byte{0xB1, 0x88, 0x00, 0xAA, 0x00, 0x38, 0xC9, 0x69}}

type IMLangStringWStrInterface interface {
	IMLangStringInterface
	SetWStr(lDestPos int32, lDestLen int32, pszSrc PWSTR, cchSrc int32, pcchActual *int32, plActualLen *int32) HRESULT
	SetStrBufW(lDestPos int32, lDestLen int32, pSrcBuf *IMLangStringBufW, pcchActual *int32, plActualLen *int32) HRESULT
	GetWStr(lSrcPos int32, lSrcLen int32, pszDest PWSTR, cchDest int32, pcchActual *int32, plActualLen *int32) HRESULT
	GetStrBufW(lSrcPos int32, lSrcMaxLen int32, ppDestBuf **IMLangStringBufW, plDestLen *int32) HRESULT
	LockWStr(lSrcPos int32, lSrcLen int32, lFlags int32, cchRequest int32, ppszDest *PWSTR, pcchDest *int32, plDestLen *int32) HRESULT
	UnlockWStr(pszSrc PWSTR, cchSrc int32, pcchActual *int32, plActualLen *int32) HRESULT
	SetLocale(lDestPos int32, lDestLen int32, locale uint32) HRESULT
	GetLocale(lSrcPos int32, lSrcMaxLen int32, plocale *uint32, plLocalePos *int32, plLocaleLen *int32) HRESULT
}

type IMLangStringWStrVtbl struct {
	IMLangStringVtbl
	SetWStr    uintptr
	SetStrBufW uintptr
	GetWStr    uintptr
	GetStrBufW uintptr
	LockWStr   uintptr
	UnlockWStr uintptr
	SetLocale  uintptr
	GetLocale  uintptr
}

type IMLangStringWStr struct {
	IMLangString
}

func (this *IMLangStringWStr) Vtbl() *IMLangStringWStrVtbl {
	return (*IMLangStringWStrVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMLangStringWStr) SetWStr(lDestPos int32, lDestLen int32, pszSrc PWSTR, cchSrc int32, pcchActual *int32, plActualLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetWStr, uintptr(unsafe.Pointer(this)), uintptr(lDestPos), uintptr(lDestLen), uintptr(unsafe.Pointer(pszSrc)), uintptr(cchSrc), uintptr(unsafe.Pointer(pcchActual)), uintptr(unsafe.Pointer(plActualLen)))
	return HRESULT(ret)
}

func (this *IMLangStringWStr) SetStrBufW(lDestPos int32, lDestLen int32, pSrcBuf *IMLangStringBufW, pcchActual *int32, plActualLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetStrBufW, uintptr(unsafe.Pointer(this)), uintptr(lDestPos), uintptr(lDestLen), uintptr(unsafe.Pointer(pSrcBuf)), uintptr(unsafe.Pointer(pcchActual)), uintptr(unsafe.Pointer(plActualLen)))
	return HRESULT(ret)
}

func (this *IMLangStringWStr) GetWStr(lSrcPos int32, lSrcLen int32, pszDest PWSTR, cchDest int32, pcchActual *int32, plActualLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetWStr, uintptr(unsafe.Pointer(this)), uintptr(lSrcPos), uintptr(lSrcLen), uintptr(unsafe.Pointer(pszDest)), uintptr(cchDest), uintptr(unsafe.Pointer(pcchActual)), uintptr(unsafe.Pointer(plActualLen)))
	return HRESULT(ret)
}

func (this *IMLangStringWStr) GetStrBufW(lSrcPos int32, lSrcMaxLen int32, ppDestBuf **IMLangStringBufW, plDestLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStrBufW, uintptr(unsafe.Pointer(this)), uintptr(lSrcPos), uintptr(lSrcMaxLen), uintptr(unsafe.Pointer(ppDestBuf)), uintptr(unsafe.Pointer(plDestLen)))
	return HRESULT(ret)
}

func (this *IMLangStringWStr) LockWStr(lSrcPos int32, lSrcLen int32, lFlags int32, cchRequest int32, ppszDest *PWSTR, pcchDest *int32, plDestLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().LockWStr, uintptr(unsafe.Pointer(this)), uintptr(lSrcPos), uintptr(lSrcLen), uintptr(lFlags), uintptr(cchRequest), uintptr(unsafe.Pointer(ppszDest)), uintptr(unsafe.Pointer(pcchDest)), uintptr(unsafe.Pointer(plDestLen)))
	return HRESULT(ret)
}

func (this *IMLangStringWStr) UnlockWStr(pszSrc PWSTR, cchSrc int32, pcchActual *int32, plActualLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UnlockWStr, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszSrc)), uintptr(cchSrc), uintptr(unsafe.Pointer(pcchActual)), uintptr(unsafe.Pointer(plActualLen)))
	return HRESULT(ret)
}

func (this *IMLangStringWStr) SetLocale(lDestPos int32, lDestLen int32, locale uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetLocale, uintptr(unsafe.Pointer(this)), uintptr(lDestPos), uintptr(lDestLen), uintptr(locale))
	return HRESULT(ret)
}

func (this *IMLangStringWStr) GetLocale(lSrcPos int32, lSrcMaxLen int32, plocale *uint32, plLocalePos *int32, plLocaleLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLocale, uintptr(unsafe.Pointer(this)), uintptr(lSrcPos), uintptr(lSrcMaxLen), uintptr(unsafe.Pointer(plocale)), uintptr(unsafe.Pointer(plLocalePos)), uintptr(unsafe.Pointer(plLocaleLen)))
	return HRESULT(ret)
}

// C04D65D2-B70D-11D0-B188-00AA0038C969
var IID_IMLangStringAStr = syscall.GUID{0xC04D65D2, 0xB70D, 0x11D0,
	[8]byte{0xB1, 0x88, 0x00, 0xAA, 0x00, 0x38, 0xC9, 0x69}}

type IMLangStringAStrInterface interface {
	IMLangStringInterface
	SetAStr(lDestPos int32, lDestLen int32, uCodePage uint32, pszSrc PSTR, cchSrc int32, pcchActual *int32, plActualLen *int32) HRESULT
	SetStrBufA(lDestPos int32, lDestLen int32, uCodePage uint32, pSrcBuf *IMLangStringBufA, pcchActual *int32, plActualLen *int32) HRESULT
	GetAStr(lSrcPos int32, lSrcLen int32, uCodePageIn uint32, puCodePageOut *uint32, pszDest PSTR, cchDest int32, pcchActual *int32, plActualLen *int32) HRESULT
	GetStrBufA(lSrcPos int32, lSrcMaxLen int32, puDestCodePage *uint32, ppDestBuf **IMLangStringBufA, plDestLen *int32) HRESULT
	LockAStr(lSrcPos int32, lSrcLen int32, lFlags int32, uCodePageIn uint32, cchRequest int32, puCodePageOut *uint32, ppszDest *PSTR, pcchDest *int32, plDestLen *int32) HRESULT
	UnlockAStr(pszSrc PSTR, cchSrc int32, pcchActual *int32, plActualLen *int32) HRESULT
	SetLocale(lDestPos int32, lDestLen int32, locale uint32) HRESULT
	GetLocale(lSrcPos int32, lSrcMaxLen int32, plocale *uint32, plLocalePos *int32, plLocaleLen *int32) HRESULT
}

type IMLangStringAStrVtbl struct {
	IMLangStringVtbl
	SetAStr    uintptr
	SetStrBufA uintptr
	GetAStr    uintptr
	GetStrBufA uintptr
	LockAStr   uintptr
	UnlockAStr uintptr
	SetLocale  uintptr
	GetLocale  uintptr
}

type IMLangStringAStr struct {
	IMLangString
}

func (this *IMLangStringAStr) Vtbl() *IMLangStringAStrVtbl {
	return (*IMLangStringAStrVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMLangStringAStr) SetAStr(lDestPos int32, lDestLen int32, uCodePage uint32, pszSrc PSTR, cchSrc int32, pcchActual *int32, plActualLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetAStr, uintptr(unsafe.Pointer(this)), uintptr(lDestPos), uintptr(lDestLen), uintptr(uCodePage), uintptr(unsafe.Pointer(pszSrc)), uintptr(cchSrc), uintptr(unsafe.Pointer(pcchActual)), uintptr(unsafe.Pointer(plActualLen)))
	return HRESULT(ret)
}

func (this *IMLangStringAStr) SetStrBufA(lDestPos int32, lDestLen int32, uCodePage uint32, pSrcBuf *IMLangStringBufA, pcchActual *int32, plActualLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetStrBufA, uintptr(unsafe.Pointer(this)), uintptr(lDestPos), uintptr(lDestLen), uintptr(uCodePage), uintptr(unsafe.Pointer(pSrcBuf)), uintptr(unsafe.Pointer(pcchActual)), uintptr(unsafe.Pointer(plActualLen)))
	return HRESULT(ret)
}

func (this *IMLangStringAStr) GetAStr(lSrcPos int32, lSrcLen int32, uCodePageIn uint32, puCodePageOut *uint32, pszDest PSTR, cchDest int32, pcchActual *int32, plActualLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAStr, uintptr(unsafe.Pointer(this)), uintptr(lSrcPos), uintptr(lSrcLen), uintptr(uCodePageIn), uintptr(unsafe.Pointer(puCodePageOut)), uintptr(unsafe.Pointer(pszDest)), uintptr(cchDest), uintptr(unsafe.Pointer(pcchActual)), uintptr(unsafe.Pointer(plActualLen)))
	return HRESULT(ret)
}

func (this *IMLangStringAStr) GetStrBufA(lSrcPos int32, lSrcMaxLen int32, puDestCodePage *uint32, ppDestBuf **IMLangStringBufA, plDestLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStrBufA, uintptr(unsafe.Pointer(this)), uintptr(lSrcPos), uintptr(lSrcMaxLen), uintptr(unsafe.Pointer(puDestCodePage)), uintptr(unsafe.Pointer(ppDestBuf)), uintptr(unsafe.Pointer(plDestLen)))
	return HRESULT(ret)
}

func (this *IMLangStringAStr) LockAStr(lSrcPos int32, lSrcLen int32, lFlags int32, uCodePageIn uint32, cchRequest int32, puCodePageOut *uint32, ppszDest *PSTR, pcchDest *int32, plDestLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().LockAStr, uintptr(unsafe.Pointer(this)), uintptr(lSrcPos), uintptr(lSrcLen), uintptr(lFlags), uintptr(uCodePageIn), uintptr(cchRequest), uintptr(unsafe.Pointer(puCodePageOut)), uintptr(unsafe.Pointer(ppszDest)), uintptr(unsafe.Pointer(pcchDest)), uintptr(unsafe.Pointer(plDestLen)))
	return HRESULT(ret)
}

func (this *IMLangStringAStr) UnlockAStr(pszSrc PSTR, cchSrc int32, pcchActual *int32, plActualLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UnlockAStr, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszSrc)), uintptr(cchSrc), uintptr(unsafe.Pointer(pcchActual)), uintptr(unsafe.Pointer(plActualLen)))
	return HRESULT(ret)
}

func (this *IMLangStringAStr) SetLocale(lDestPos int32, lDestLen int32, locale uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetLocale, uintptr(unsafe.Pointer(this)), uintptr(lDestPos), uintptr(lDestLen), uintptr(locale))
	return HRESULT(ret)
}

func (this *IMLangStringAStr) GetLocale(lSrcPos int32, lSrcMaxLen int32, plocale *uint32, plLocalePos *int32, plLocaleLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLocale, uintptr(unsafe.Pointer(this)), uintptr(lSrcPos), uintptr(lSrcMaxLen), uintptr(unsafe.Pointer(plocale)), uintptr(unsafe.Pointer(plLocalePos)), uintptr(unsafe.Pointer(plLocaleLen)))
	return HRESULT(ret)
}

// F5BE2EE1-BFD7-11D0-B188-00AA0038C969
var IID_IMLangLineBreakConsole = syscall.GUID{0xF5BE2EE1, 0xBFD7, 0x11D0,
	[8]byte{0xB1, 0x88, 0x00, 0xAA, 0x00, 0x38, 0xC9, 0x69}}

type IMLangLineBreakConsoleInterface interface {
	IUnknownInterface
	BreakLineML(pSrcMLStr *IMLangString, lSrcPos int32, lSrcLen int32, cMinColumns int32, cMaxColumns int32, plLineLen *int32, plSkipLen *int32) HRESULT
	BreakLineW(locale uint32, pszSrc PWSTR, cchSrc int32, cMaxColumns int32, pcchLine *int32, pcchSkip *int32) HRESULT
	BreakLineA(locale uint32, uCodePage uint32, pszSrc PSTR, cchSrc int32, cMaxColumns int32, pcchLine *int32, pcchSkip *int32) HRESULT
}

type IMLangLineBreakConsoleVtbl struct {
	IUnknownVtbl
	BreakLineML uintptr
	BreakLineW  uintptr
	BreakLineA  uintptr
}

type IMLangLineBreakConsole struct {
	IUnknown
}

func (this *IMLangLineBreakConsole) Vtbl() *IMLangLineBreakConsoleVtbl {
	return (*IMLangLineBreakConsoleVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMLangLineBreakConsole) BreakLineML(pSrcMLStr *IMLangString, lSrcPos int32, lSrcLen int32, cMinColumns int32, cMaxColumns int32, plLineLen *int32, plSkipLen *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().BreakLineML, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pSrcMLStr)), uintptr(lSrcPos), uintptr(lSrcLen), uintptr(cMinColumns), uintptr(cMaxColumns), uintptr(unsafe.Pointer(plLineLen)), uintptr(unsafe.Pointer(plSkipLen)))
	return HRESULT(ret)
}

func (this *IMLangLineBreakConsole) BreakLineW(locale uint32, pszSrc PWSTR, cchSrc int32, cMaxColumns int32, pcchLine *int32, pcchSkip *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().BreakLineW, uintptr(unsafe.Pointer(this)), uintptr(locale), uintptr(unsafe.Pointer(pszSrc)), uintptr(cchSrc), uintptr(cMaxColumns), uintptr(unsafe.Pointer(pcchLine)), uintptr(unsafe.Pointer(pcchSkip)))
	return HRESULT(ret)
}

func (this *IMLangLineBreakConsole) BreakLineA(locale uint32, uCodePage uint32, pszSrc PSTR, cchSrc int32, cMaxColumns int32, pcchLine *int32, pcchSkip *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().BreakLineA, uintptr(unsafe.Pointer(this)), uintptr(locale), uintptr(uCodePage), uintptr(unsafe.Pointer(pszSrc)), uintptr(cchSrc), uintptr(cMaxColumns), uintptr(unsafe.Pointer(pcchLine)), uintptr(unsafe.Pointer(pcchSkip)))
	return HRESULT(ret)
}

// 275C23E3-3747-11D0-9FEA-00AA003F8646
var IID_IEnumCodePage = syscall.GUID{0x275C23E3, 0x3747, 0x11D0,
	[8]byte{0x9F, 0xEA, 0x00, 0xAA, 0x00, 0x3F, 0x86, 0x46}}

type IEnumCodePageInterface interface {
	IUnknownInterface
	Clone(ppEnum **IEnumCodePage) HRESULT
	Next(celt uint32, rgelt *MIMECPINFO, pceltFetched *uint32) HRESULT
	Reset() HRESULT
	Skip(celt uint32) HRESULT
}

type IEnumCodePageVtbl struct {
	IUnknownVtbl
	Clone uintptr
	Next  uintptr
	Reset uintptr
	Skip  uintptr
}

type IEnumCodePage struct {
	IUnknown
}

func (this *IEnumCodePage) Vtbl() *IEnumCodePageVtbl {
	return (*IEnumCodePageVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumCodePage) Clone(ppEnum **IEnumCodePage) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnum)))
	return HRESULT(ret)
}

func (this *IEnumCodePage) Next(celt uint32, rgelt *MIMECPINFO, pceltFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(celt), uintptr(unsafe.Pointer(rgelt)), uintptr(unsafe.Pointer(pceltFetched)))
	return HRESULT(ret)
}

func (this *IEnumCodePage) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumCodePage) Skip(celt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(celt))
	return HRESULT(ret)
}

// 3DC39D1D-C030-11D0-B81B-00C04FC9B31F
var IID_IEnumRfc1766 = syscall.GUID{0x3DC39D1D, 0xC030, 0x11D0,
	[8]byte{0xB8, 0x1B, 0x00, 0xC0, 0x4F, 0xC9, 0xB3, 0x1F}}

type IEnumRfc1766Interface interface {
	IUnknownInterface
	Clone(ppEnum **IEnumRfc1766) HRESULT
	Next(celt uint32, rgelt *RFC1766INFO, pceltFetched *uint32) HRESULT
	Reset() HRESULT
	Skip(celt uint32) HRESULT
}

type IEnumRfc1766Vtbl struct {
	IUnknownVtbl
	Clone uintptr
	Next  uintptr
	Reset uintptr
	Skip  uintptr
}

type IEnumRfc1766 struct {
	IUnknown
}

func (this *IEnumRfc1766) Vtbl() *IEnumRfc1766Vtbl {
	return (*IEnumRfc1766Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumRfc1766) Clone(ppEnum **IEnumRfc1766) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnum)))
	return HRESULT(ret)
}

func (this *IEnumRfc1766) Next(celt uint32, rgelt *RFC1766INFO, pceltFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(celt), uintptr(unsafe.Pointer(rgelt)), uintptr(unsafe.Pointer(pceltFetched)))
	return HRESULT(ret)
}

func (this *IEnumRfc1766) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumRfc1766) Skip(celt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(celt))
	return HRESULT(ret)
}

// AE5F1430-388B-11D2-8380-00C04F8F5DA1
var IID_IEnumScript = syscall.GUID{0xAE5F1430, 0x388B, 0x11D2,
	[8]byte{0x83, 0x80, 0x00, 0xC0, 0x4F, 0x8F, 0x5D, 0xA1}}

type IEnumScriptInterface interface {
	IUnknownInterface
	Clone(ppEnum **IEnumScript) HRESULT
	Next(celt uint32, rgelt *SCRIPTINFO, pceltFetched *uint32) HRESULT
	Reset() HRESULT
	Skip(celt uint32) HRESULT
}

type IEnumScriptVtbl struct {
	IUnknownVtbl
	Clone uintptr
	Next  uintptr
	Reset uintptr
	Skip  uintptr
}

type IEnumScript struct {
	IUnknown
}

func (this *IEnumScript) Vtbl() *IEnumScriptVtbl {
	return (*IEnumScriptVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumScript) Clone(ppEnum **IEnumScript) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnum)))
	return HRESULT(ret)
}

func (this *IEnumScript) Next(celt uint32, rgelt *SCRIPTINFO, pceltFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(celt), uintptr(unsafe.Pointer(rgelt)), uintptr(unsafe.Pointer(pceltFetched)))
	return HRESULT(ret)
}

func (this *IEnumScript) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumScript) Skip(celt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(celt))
	return HRESULT(ret)
}

// D66D6F98-CDAA-11D0-B822-00C04FC9B31F
var IID_IMLangConvertCharset = syscall.GUID{0xD66D6F98, 0xCDAA, 0x11D0,
	[8]byte{0xB8, 0x22, 0x00, 0xC0, 0x4F, 0xC9, 0xB3, 0x1F}}

type IMLangConvertCharsetInterface interface {
	IUnknownInterface
	Initialize(uiSrcCodePage uint32, uiDstCodePage uint32, dwProperty uint32) HRESULT
	GetSourceCodePage(puiSrcCodePage *uint32) HRESULT
	GetDestinationCodePage(puiDstCodePage *uint32) HRESULT
	GetProperty(pdwProperty *uint32) HRESULT
	DoConversion(pSrcStr *byte, pcSrcSize *uint32, pDstStr *byte, pcDstSize *uint32) HRESULT
	DoConversionToUnicode(pSrcStr PSTR, pcSrcSize *uint32, pDstStr PWSTR, pcDstSize *uint32) HRESULT
	DoConversionFromUnicode(pSrcStr PWSTR, pcSrcSize *uint32, pDstStr PSTR, pcDstSize *uint32) HRESULT
}

type IMLangConvertCharsetVtbl struct {
	IUnknownVtbl
	Initialize              uintptr
	GetSourceCodePage       uintptr
	GetDestinationCodePage  uintptr
	GetProperty             uintptr
	DoConversion            uintptr
	DoConversionToUnicode   uintptr
	DoConversionFromUnicode uintptr
}

type IMLangConvertCharset struct {
	IUnknown
}

func (this *IMLangConvertCharset) Vtbl() *IMLangConvertCharsetVtbl {
	return (*IMLangConvertCharsetVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMLangConvertCharset) Initialize(uiSrcCodePage uint32, uiDstCodePage uint32, dwProperty uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Initialize, uintptr(unsafe.Pointer(this)), uintptr(uiSrcCodePage), uintptr(uiDstCodePage), uintptr(dwProperty))
	return HRESULT(ret)
}

func (this *IMLangConvertCharset) GetSourceCodePage(puiSrcCodePage *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSourceCodePage, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(puiSrcCodePage)))
	return HRESULT(ret)
}

func (this *IMLangConvertCharset) GetDestinationCodePage(puiDstCodePage *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDestinationCodePage, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(puiDstCodePage)))
	return HRESULT(ret)
}

func (this *IMLangConvertCharset) GetProperty(pdwProperty *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetProperty, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwProperty)))
	return HRESULT(ret)
}

func (this *IMLangConvertCharset) DoConversion(pSrcStr *byte, pcSrcSize *uint32, pDstStr *byte, pcDstSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DoConversion, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pSrcStr)), uintptr(unsafe.Pointer(pcSrcSize)), uintptr(unsafe.Pointer(pDstStr)), uintptr(unsafe.Pointer(pcDstSize)))
	return HRESULT(ret)
}

func (this *IMLangConvertCharset) DoConversionToUnicode(pSrcStr PSTR, pcSrcSize *uint32, pDstStr PWSTR, pcDstSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DoConversionToUnicode, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pSrcStr)), uintptr(unsafe.Pointer(pcSrcSize)), uintptr(unsafe.Pointer(pDstStr)), uintptr(unsafe.Pointer(pcDstSize)))
	return HRESULT(ret)
}

func (this *IMLangConvertCharset) DoConversionFromUnicode(pSrcStr PWSTR, pcSrcSize *uint32, pDstStr PSTR, pcDstSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DoConversionFromUnicode, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pSrcStr)), uintptr(unsafe.Pointer(pcSrcSize)), uintptr(unsafe.Pointer(pDstStr)), uintptr(unsafe.Pointer(pcDstSize)))
	return HRESULT(ret)
}

// 275C23E1-3747-11D0-9FEA-00AA003F8646
var IID_IMultiLanguage = syscall.GUID{0x275C23E1, 0x3747, 0x11D0,
	[8]byte{0x9F, 0xEA, 0x00, 0xAA, 0x00, 0x3F, 0x86, 0x46}}

type IMultiLanguageInterface interface {
	IUnknownInterface
	GetNumberOfCodePageInfo(pcCodePage *uint32) HRESULT
	GetCodePageInfo(uiCodePage uint32, pCodePageInfo *MIMECPINFO) HRESULT
	GetFamilyCodePage(uiCodePage uint32, puiFamilyCodePage *uint32) HRESULT
	EnumCodePages(grfFlags uint32, ppEnumCodePage **IEnumCodePage) HRESULT
	GetCharsetInfo(Charset BSTR, pCharsetInfo *MIMECSETINFO) HRESULT
	IsConvertible(dwSrcEncoding uint32, dwDstEncoding uint32) HRESULT
	ConvertString(pdwMode *uint32, dwSrcEncoding uint32, dwDstEncoding uint32, pSrcStr *byte, pcSrcSize *uint32, pDstStr *byte, pcDstSize *uint32) HRESULT
	ConvertStringToUnicode(pdwMode *uint32, dwEncoding uint32, pSrcStr PSTR, pcSrcSize *uint32, pDstStr PWSTR, pcDstSize *uint32) HRESULT
	ConvertStringFromUnicode(pdwMode *uint32, dwEncoding uint32, pSrcStr PWSTR, pcSrcSize *uint32, pDstStr PSTR, pcDstSize *uint32) HRESULT
	ConvertStringReset() HRESULT
	GetRfc1766FromLcid(Locale uint32, pbstrRfc1766 *BSTR) HRESULT
	GetLcidFromRfc1766(pLocale *uint32, bstrRfc1766 BSTR) HRESULT
	EnumRfc1766(ppEnumRfc1766 **IEnumRfc1766) HRESULT
	GetRfc1766Info(Locale uint32, pRfc1766Info *RFC1766INFO) HRESULT
	CreateConvertCharset(uiSrcCodePage uint32, uiDstCodePage uint32, dwProperty uint32, ppMLangConvertCharset **IMLangConvertCharset) HRESULT
}

type IMultiLanguageVtbl struct {
	IUnknownVtbl
	GetNumberOfCodePageInfo  uintptr
	GetCodePageInfo          uintptr
	GetFamilyCodePage        uintptr
	EnumCodePages            uintptr
	GetCharsetInfo           uintptr
	IsConvertible            uintptr
	ConvertString            uintptr
	ConvertStringToUnicode   uintptr
	ConvertStringFromUnicode uintptr
	ConvertStringReset       uintptr
	GetRfc1766FromLcid       uintptr
	GetLcidFromRfc1766       uintptr
	EnumRfc1766              uintptr
	GetRfc1766Info           uintptr
	CreateConvertCharset     uintptr
}

type IMultiLanguage struct {
	IUnknown
}

func (this *IMultiLanguage) Vtbl() *IMultiLanguageVtbl {
	return (*IMultiLanguageVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMultiLanguage) GetNumberOfCodePageInfo(pcCodePage *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNumberOfCodePageInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcCodePage)))
	return HRESULT(ret)
}

func (this *IMultiLanguage) GetCodePageInfo(uiCodePage uint32, pCodePageInfo *MIMECPINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCodePageInfo, uintptr(unsafe.Pointer(this)), uintptr(uiCodePage), uintptr(unsafe.Pointer(pCodePageInfo)))
	return HRESULT(ret)
}

func (this *IMultiLanguage) GetFamilyCodePage(uiCodePage uint32, puiFamilyCodePage *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFamilyCodePage, uintptr(unsafe.Pointer(this)), uintptr(uiCodePage), uintptr(unsafe.Pointer(puiFamilyCodePage)))
	return HRESULT(ret)
}

func (this *IMultiLanguage) EnumCodePages(grfFlags uint32, ppEnumCodePage **IEnumCodePage) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumCodePages, uintptr(unsafe.Pointer(this)), uintptr(grfFlags), uintptr(unsafe.Pointer(ppEnumCodePage)))
	return HRESULT(ret)
}

func (this *IMultiLanguage) GetCharsetInfo(Charset BSTR, pCharsetInfo *MIMECSETINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCharsetInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(Charset)), uintptr(unsafe.Pointer(pCharsetInfo)))
	return HRESULT(ret)
}

func (this *IMultiLanguage) IsConvertible(dwSrcEncoding uint32, dwDstEncoding uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsConvertible, uintptr(unsafe.Pointer(this)), uintptr(dwSrcEncoding), uintptr(dwDstEncoding))
	return HRESULT(ret)
}

func (this *IMultiLanguage) ConvertString(pdwMode *uint32, dwSrcEncoding uint32, dwDstEncoding uint32, pSrcStr *byte, pcSrcSize *uint32, pDstStr *byte, pcDstSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertString, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwMode)), uintptr(dwSrcEncoding), uintptr(dwDstEncoding), uintptr(unsafe.Pointer(pSrcStr)), uintptr(unsafe.Pointer(pcSrcSize)), uintptr(unsafe.Pointer(pDstStr)), uintptr(unsafe.Pointer(pcDstSize)))
	return HRESULT(ret)
}

func (this *IMultiLanguage) ConvertStringToUnicode(pdwMode *uint32, dwEncoding uint32, pSrcStr PSTR, pcSrcSize *uint32, pDstStr PWSTR, pcDstSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertStringToUnicode, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwMode)), uintptr(dwEncoding), uintptr(unsafe.Pointer(pSrcStr)), uintptr(unsafe.Pointer(pcSrcSize)), uintptr(unsafe.Pointer(pDstStr)), uintptr(unsafe.Pointer(pcDstSize)))
	return HRESULT(ret)
}

func (this *IMultiLanguage) ConvertStringFromUnicode(pdwMode *uint32, dwEncoding uint32, pSrcStr PWSTR, pcSrcSize *uint32, pDstStr PSTR, pcDstSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertStringFromUnicode, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwMode)), uintptr(dwEncoding), uintptr(unsafe.Pointer(pSrcStr)), uintptr(unsafe.Pointer(pcSrcSize)), uintptr(unsafe.Pointer(pDstStr)), uintptr(unsafe.Pointer(pcDstSize)))
	return HRESULT(ret)
}

func (this *IMultiLanguage) ConvertStringReset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertStringReset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IMultiLanguage) GetRfc1766FromLcid(Locale uint32, pbstrRfc1766 *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRfc1766FromLcid, uintptr(unsafe.Pointer(this)), uintptr(Locale), uintptr(unsafe.Pointer(pbstrRfc1766)))
	return HRESULT(ret)
}

func (this *IMultiLanguage) GetLcidFromRfc1766(pLocale *uint32, bstrRfc1766 BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLcidFromRfc1766, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pLocale)), uintptr(unsafe.Pointer(bstrRfc1766)))
	return HRESULT(ret)
}

func (this *IMultiLanguage) EnumRfc1766(ppEnumRfc1766 **IEnumRfc1766) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumRfc1766, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnumRfc1766)))
	return HRESULT(ret)
}

func (this *IMultiLanguage) GetRfc1766Info(Locale uint32, pRfc1766Info *RFC1766INFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRfc1766Info, uintptr(unsafe.Pointer(this)), uintptr(Locale), uintptr(unsafe.Pointer(pRfc1766Info)))
	return HRESULT(ret)
}

func (this *IMultiLanguage) CreateConvertCharset(uiSrcCodePage uint32, uiDstCodePage uint32, dwProperty uint32, ppMLangConvertCharset **IMLangConvertCharset) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateConvertCharset, uintptr(unsafe.Pointer(this)), uintptr(uiSrcCodePage), uintptr(uiDstCodePage), uintptr(dwProperty), uintptr(unsafe.Pointer(ppMLangConvertCharset)))
	return HRESULT(ret)
}

// DCCFC164-2B38-11D2-B7EC-00C04F8F5D9A
var IID_IMultiLanguage2 = syscall.GUID{0xDCCFC164, 0x2B38, 0x11D2,
	[8]byte{0xB7, 0xEC, 0x00, 0xC0, 0x4F, 0x8F, 0x5D, 0x9A}}

type IMultiLanguage2Interface interface {
	IUnknownInterface
	GetNumberOfCodePageInfo(pcCodePage *uint32) HRESULT
	GetCodePageInfo(uiCodePage uint32, LangId uint16, pCodePageInfo *MIMECPINFO) HRESULT
	GetFamilyCodePage(uiCodePage uint32, puiFamilyCodePage *uint32) HRESULT
	EnumCodePages(grfFlags uint32, LangId uint16, ppEnumCodePage **IEnumCodePage) HRESULT
	GetCharsetInfo(Charset BSTR, pCharsetInfo *MIMECSETINFO) HRESULT
	IsConvertible(dwSrcEncoding uint32, dwDstEncoding uint32) HRESULT
	ConvertString(pdwMode *uint32, dwSrcEncoding uint32, dwDstEncoding uint32, pSrcStr *byte, pcSrcSize *uint32, pDstStr *byte, pcDstSize *uint32) HRESULT
	ConvertStringToUnicode(pdwMode *uint32, dwEncoding uint32, pSrcStr PSTR, pcSrcSize *uint32, pDstStr PWSTR, pcDstSize *uint32) HRESULT
	ConvertStringFromUnicode(pdwMode *uint32, dwEncoding uint32, pSrcStr PWSTR, pcSrcSize *uint32, pDstStr PSTR, pcDstSize *uint32) HRESULT
	ConvertStringReset() HRESULT
	GetRfc1766FromLcid(Locale uint32, pbstrRfc1766 *BSTR) HRESULT
	GetLcidFromRfc1766(pLocale *uint32, bstrRfc1766 BSTR) HRESULT
	EnumRfc1766(LangId uint16, ppEnumRfc1766 **IEnumRfc1766) HRESULT
	GetRfc1766Info(Locale uint32, LangId uint16, pRfc1766Info *RFC1766INFO) HRESULT
	CreateConvertCharset(uiSrcCodePage uint32, uiDstCodePage uint32, dwProperty uint32, ppMLangConvertCharset **IMLangConvertCharset) HRESULT
	ConvertStringInIStream(pdwMode *uint32, dwFlag uint32, lpFallBack PWSTR, dwSrcEncoding uint32, dwDstEncoding uint32, pstmIn *IStream, pstmOut *IStream) HRESULT
	ConvertStringToUnicodeEx(pdwMode *uint32, dwEncoding uint32, pSrcStr PSTR, pcSrcSize *uint32, pDstStr PWSTR, pcDstSize *uint32, dwFlag uint32, lpFallBack PWSTR) HRESULT
	ConvertStringFromUnicodeEx(pdwMode *uint32, dwEncoding uint32, pSrcStr PWSTR, pcSrcSize *uint32, pDstStr PSTR, pcDstSize *uint32, dwFlag uint32, lpFallBack PWSTR) HRESULT
	DetectCodepageInIStream(dwFlag uint32, dwPrefWinCodePage uint32, pstmIn *IStream, lpEncoding *DetectEncodingInfo, pnScores *int32) HRESULT
	DetectInputCodepage(dwFlag uint32, dwPrefWinCodePage uint32, pSrcStr PSTR, pcSrcSize *int32, lpEncoding *DetectEncodingInfo, pnScores *int32) HRESULT
	ValidateCodePage(uiCodePage uint32, hwnd HWND) HRESULT
	GetCodePageDescription(uiCodePage uint32, lcid uint32, lpWideCharStr PWSTR, cchWideChar int32) HRESULT
	IsCodePageInstallable(uiCodePage uint32) HRESULT
	SetMimeDBSource(dwSource MIMECONTF) HRESULT
	GetNumberOfScripts(pnScripts *uint32) HRESULT
	EnumScripts(dwFlags uint32, LangId uint16, ppEnumScript **IEnumScript) HRESULT
	ValidateCodePageEx(uiCodePage uint32, hwnd HWND, dwfIODControl uint32) HRESULT
}

type IMultiLanguage2Vtbl struct {
	IUnknownVtbl
	GetNumberOfCodePageInfo    uintptr
	GetCodePageInfo            uintptr
	GetFamilyCodePage          uintptr
	EnumCodePages              uintptr
	GetCharsetInfo             uintptr
	IsConvertible              uintptr
	ConvertString              uintptr
	ConvertStringToUnicode     uintptr
	ConvertStringFromUnicode   uintptr
	ConvertStringReset         uintptr
	GetRfc1766FromLcid         uintptr
	GetLcidFromRfc1766         uintptr
	EnumRfc1766                uintptr
	GetRfc1766Info             uintptr
	CreateConvertCharset       uintptr
	ConvertStringInIStream     uintptr
	ConvertStringToUnicodeEx   uintptr
	ConvertStringFromUnicodeEx uintptr
	DetectCodepageInIStream    uintptr
	DetectInputCodepage        uintptr
	ValidateCodePage           uintptr
	GetCodePageDescription     uintptr
	IsCodePageInstallable      uintptr
	SetMimeDBSource            uintptr
	GetNumberOfScripts         uintptr
	EnumScripts                uintptr
	ValidateCodePageEx         uintptr
}

type IMultiLanguage2 struct {
	IUnknown
}

func (this *IMultiLanguage2) Vtbl() *IMultiLanguage2Vtbl {
	return (*IMultiLanguage2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMultiLanguage2) GetNumberOfCodePageInfo(pcCodePage *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNumberOfCodePageInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcCodePage)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) GetCodePageInfo(uiCodePage uint32, LangId uint16, pCodePageInfo *MIMECPINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCodePageInfo, uintptr(unsafe.Pointer(this)), uintptr(uiCodePage), uintptr(LangId), uintptr(unsafe.Pointer(pCodePageInfo)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) GetFamilyCodePage(uiCodePage uint32, puiFamilyCodePage *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFamilyCodePage, uintptr(unsafe.Pointer(this)), uintptr(uiCodePage), uintptr(unsafe.Pointer(puiFamilyCodePage)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) EnumCodePages(grfFlags uint32, LangId uint16, ppEnumCodePage **IEnumCodePage) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumCodePages, uintptr(unsafe.Pointer(this)), uintptr(grfFlags), uintptr(LangId), uintptr(unsafe.Pointer(ppEnumCodePage)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) GetCharsetInfo(Charset BSTR, pCharsetInfo *MIMECSETINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCharsetInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(Charset)), uintptr(unsafe.Pointer(pCharsetInfo)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) IsConvertible(dwSrcEncoding uint32, dwDstEncoding uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsConvertible, uintptr(unsafe.Pointer(this)), uintptr(dwSrcEncoding), uintptr(dwDstEncoding))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) ConvertString(pdwMode *uint32, dwSrcEncoding uint32, dwDstEncoding uint32, pSrcStr *byte, pcSrcSize *uint32, pDstStr *byte, pcDstSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertString, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwMode)), uintptr(dwSrcEncoding), uintptr(dwDstEncoding), uintptr(unsafe.Pointer(pSrcStr)), uintptr(unsafe.Pointer(pcSrcSize)), uintptr(unsafe.Pointer(pDstStr)), uintptr(unsafe.Pointer(pcDstSize)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) ConvertStringToUnicode(pdwMode *uint32, dwEncoding uint32, pSrcStr PSTR, pcSrcSize *uint32, pDstStr PWSTR, pcDstSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertStringToUnicode, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwMode)), uintptr(dwEncoding), uintptr(unsafe.Pointer(pSrcStr)), uintptr(unsafe.Pointer(pcSrcSize)), uintptr(unsafe.Pointer(pDstStr)), uintptr(unsafe.Pointer(pcDstSize)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) ConvertStringFromUnicode(pdwMode *uint32, dwEncoding uint32, pSrcStr PWSTR, pcSrcSize *uint32, pDstStr PSTR, pcDstSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertStringFromUnicode, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwMode)), uintptr(dwEncoding), uintptr(unsafe.Pointer(pSrcStr)), uintptr(unsafe.Pointer(pcSrcSize)), uintptr(unsafe.Pointer(pDstStr)), uintptr(unsafe.Pointer(pcDstSize)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) ConvertStringReset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertStringReset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) GetRfc1766FromLcid(Locale uint32, pbstrRfc1766 *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRfc1766FromLcid, uintptr(unsafe.Pointer(this)), uintptr(Locale), uintptr(unsafe.Pointer(pbstrRfc1766)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) GetLcidFromRfc1766(pLocale *uint32, bstrRfc1766 BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLcidFromRfc1766, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pLocale)), uintptr(unsafe.Pointer(bstrRfc1766)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) EnumRfc1766(LangId uint16, ppEnumRfc1766 **IEnumRfc1766) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumRfc1766, uintptr(unsafe.Pointer(this)), uintptr(LangId), uintptr(unsafe.Pointer(ppEnumRfc1766)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) GetRfc1766Info(Locale uint32, LangId uint16, pRfc1766Info *RFC1766INFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRfc1766Info, uintptr(unsafe.Pointer(this)), uintptr(Locale), uintptr(LangId), uintptr(unsafe.Pointer(pRfc1766Info)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) CreateConvertCharset(uiSrcCodePage uint32, uiDstCodePage uint32, dwProperty uint32, ppMLangConvertCharset **IMLangConvertCharset) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateConvertCharset, uintptr(unsafe.Pointer(this)), uintptr(uiSrcCodePage), uintptr(uiDstCodePage), uintptr(dwProperty), uintptr(unsafe.Pointer(ppMLangConvertCharset)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) ConvertStringInIStream(pdwMode *uint32, dwFlag uint32, lpFallBack PWSTR, dwSrcEncoding uint32, dwDstEncoding uint32, pstmIn *IStream, pstmOut *IStream) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertStringInIStream, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwMode)), uintptr(dwFlag), uintptr(unsafe.Pointer(lpFallBack)), uintptr(dwSrcEncoding), uintptr(dwDstEncoding), uintptr(unsafe.Pointer(pstmIn)), uintptr(unsafe.Pointer(pstmOut)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) ConvertStringToUnicodeEx(pdwMode *uint32, dwEncoding uint32, pSrcStr PSTR, pcSrcSize *uint32, pDstStr PWSTR, pcDstSize *uint32, dwFlag uint32, lpFallBack PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertStringToUnicodeEx, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwMode)), uintptr(dwEncoding), uintptr(unsafe.Pointer(pSrcStr)), uintptr(unsafe.Pointer(pcSrcSize)), uintptr(unsafe.Pointer(pDstStr)), uintptr(unsafe.Pointer(pcDstSize)), uintptr(dwFlag), uintptr(unsafe.Pointer(lpFallBack)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) ConvertStringFromUnicodeEx(pdwMode *uint32, dwEncoding uint32, pSrcStr PWSTR, pcSrcSize *uint32, pDstStr PSTR, pcDstSize *uint32, dwFlag uint32, lpFallBack PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertStringFromUnicodeEx, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwMode)), uintptr(dwEncoding), uintptr(unsafe.Pointer(pSrcStr)), uintptr(unsafe.Pointer(pcSrcSize)), uintptr(unsafe.Pointer(pDstStr)), uintptr(unsafe.Pointer(pcDstSize)), uintptr(dwFlag), uintptr(unsafe.Pointer(lpFallBack)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) DetectCodepageInIStream(dwFlag uint32, dwPrefWinCodePage uint32, pstmIn *IStream, lpEncoding *DetectEncodingInfo, pnScores *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DetectCodepageInIStream, uintptr(unsafe.Pointer(this)), uintptr(dwFlag), uintptr(dwPrefWinCodePage), uintptr(unsafe.Pointer(pstmIn)), uintptr(unsafe.Pointer(lpEncoding)), uintptr(unsafe.Pointer(pnScores)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) DetectInputCodepage(dwFlag uint32, dwPrefWinCodePage uint32, pSrcStr PSTR, pcSrcSize *int32, lpEncoding *DetectEncodingInfo, pnScores *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DetectInputCodepage, uintptr(unsafe.Pointer(this)), uintptr(dwFlag), uintptr(dwPrefWinCodePage), uintptr(unsafe.Pointer(pSrcStr)), uintptr(unsafe.Pointer(pcSrcSize)), uintptr(unsafe.Pointer(lpEncoding)), uintptr(unsafe.Pointer(pnScores)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) ValidateCodePage(uiCodePage uint32, hwnd HWND) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ValidateCodePage, uintptr(unsafe.Pointer(this)), uintptr(uiCodePage), hwnd)
	return HRESULT(ret)
}

func (this *IMultiLanguage2) GetCodePageDescription(uiCodePage uint32, lcid uint32, lpWideCharStr PWSTR, cchWideChar int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCodePageDescription, uintptr(unsafe.Pointer(this)), uintptr(uiCodePage), uintptr(lcid), uintptr(unsafe.Pointer(lpWideCharStr)), uintptr(cchWideChar))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) IsCodePageInstallable(uiCodePage uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsCodePageInstallable, uintptr(unsafe.Pointer(this)), uintptr(uiCodePage))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) SetMimeDBSource(dwSource MIMECONTF) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetMimeDBSource, uintptr(unsafe.Pointer(this)), uintptr(dwSource))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) GetNumberOfScripts(pnScripts *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNumberOfScripts, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pnScripts)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) EnumScripts(dwFlags uint32, LangId uint16, ppEnumScript **IEnumScript) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumScripts, uintptr(unsafe.Pointer(this)), uintptr(dwFlags), uintptr(LangId), uintptr(unsafe.Pointer(ppEnumScript)))
	return HRESULT(ret)
}

func (this *IMultiLanguage2) ValidateCodePageEx(uiCodePage uint32, hwnd HWND, dwfIODControl uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ValidateCodePageEx, uintptr(unsafe.Pointer(this)), uintptr(uiCodePage), hwnd, uintptr(dwfIODControl))
	return HRESULT(ret)
}

// 359F3443-BD4A-11D0-B188-00AA0038C969
var IID_IMLangCodePages = syscall.GUID{0x359F3443, 0xBD4A, 0x11D0,
	[8]byte{0xB1, 0x88, 0x00, 0xAA, 0x00, 0x38, 0xC9, 0x69}}

type IMLangCodePagesInterface interface {
	IUnknownInterface
	GetCharCodePages(chSrc uint16, pdwCodePages *uint32) HRESULT
	GetStrCodePages(pszSrc PWSTR, cchSrc int32, dwPriorityCodePages uint32, pdwCodePages *uint32, pcchCodePages *int32) HRESULT
	CodePageToCodePages(uCodePage uint32, pdwCodePages *uint32) HRESULT
	CodePagesToCodePage(dwCodePages uint32, uDefaultCodePage uint32, puCodePage *uint32) HRESULT
}

type IMLangCodePagesVtbl struct {
	IUnknownVtbl
	GetCharCodePages    uintptr
	GetStrCodePages     uintptr
	CodePageToCodePages uintptr
	CodePagesToCodePage uintptr
}

type IMLangCodePages struct {
	IUnknown
}

func (this *IMLangCodePages) Vtbl() *IMLangCodePagesVtbl {
	return (*IMLangCodePagesVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMLangCodePages) GetCharCodePages(chSrc uint16, pdwCodePages *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCharCodePages, uintptr(unsafe.Pointer(this)), uintptr(chSrc), uintptr(unsafe.Pointer(pdwCodePages)))
	return HRESULT(ret)
}

func (this *IMLangCodePages) GetStrCodePages(pszSrc PWSTR, cchSrc int32, dwPriorityCodePages uint32, pdwCodePages *uint32, pcchCodePages *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStrCodePages, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszSrc)), uintptr(cchSrc), uintptr(dwPriorityCodePages), uintptr(unsafe.Pointer(pdwCodePages)), uintptr(unsafe.Pointer(pcchCodePages)))
	return HRESULT(ret)
}

func (this *IMLangCodePages) CodePageToCodePages(uCodePage uint32, pdwCodePages *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CodePageToCodePages, uintptr(unsafe.Pointer(this)), uintptr(uCodePage), uintptr(unsafe.Pointer(pdwCodePages)))
	return HRESULT(ret)
}

func (this *IMLangCodePages) CodePagesToCodePage(dwCodePages uint32, uDefaultCodePage uint32, puCodePage *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CodePagesToCodePage, uintptr(unsafe.Pointer(this)), uintptr(dwCodePages), uintptr(uDefaultCodePage), uintptr(unsafe.Pointer(puCodePage)))
	return HRESULT(ret)
}

// 359F3441-BD4A-11D0-B188-00AA0038C969
var IID_IMLangFontLink = syscall.GUID{0x359F3441, 0xBD4A, 0x11D0,
	[8]byte{0xB1, 0x88, 0x00, 0xAA, 0x00, 0x38, 0xC9, 0x69}}

type IMLangFontLinkInterface interface {
	IMLangCodePagesInterface
	GetFontCodePages(hDC HDC, hFont HFONT, pdwCodePages *uint32) HRESULT
	MapFont(hDC HDC, dwCodePages uint32, hSrcFont HFONT, phDestFont *HFONT) HRESULT
	ReleaseFont(hFont HFONT) HRESULT
	ResetFontMapping() HRESULT
}

type IMLangFontLinkVtbl struct {
	IMLangCodePagesVtbl
	GetFontCodePages uintptr
	MapFont          uintptr
	ReleaseFont      uintptr
	ResetFontMapping uintptr
}

type IMLangFontLink struct {
	IMLangCodePages
}

func (this *IMLangFontLink) Vtbl() *IMLangFontLinkVtbl {
	return (*IMLangFontLinkVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMLangFontLink) GetFontCodePages(hDC HDC, hFont HFONT, pdwCodePages *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFontCodePages, uintptr(unsafe.Pointer(this)), hDC, hFont, uintptr(unsafe.Pointer(pdwCodePages)))
	return HRESULT(ret)
}

func (this *IMLangFontLink) MapFont(hDC HDC, dwCodePages uint32, hSrcFont HFONT, phDestFont *HFONT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MapFont, uintptr(unsafe.Pointer(this)), hDC, uintptr(dwCodePages), hSrcFont, uintptr(unsafe.Pointer(phDestFont)))
	return HRESULT(ret)
}

func (this *IMLangFontLink) ReleaseFont(hFont HFONT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReleaseFont, uintptr(unsafe.Pointer(this)), hFont)
	return HRESULT(ret)
}

func (this *IMLangFontLink) ResetFontMapping() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ResetFontMapping, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// DCCFC162-2B38-11D2-B7EC-00C04F8F5D9A
var IID_IMLangFontLink2 = syscall.GUID{0xDCCFC162, 0x2B38, 0x11D2,
	[8]byte{0xB7, 0xEC, 0x00, 0xC0, 0x4F, 0x8F, 0x5D, 0x9A}}

type IMLangFontLink2Interface interface {
	IMLangCodePagesInterface
	GetFontCodePages(hDC HDC, hFont HFONT, pdwCodePages *uint32) HRESULT
	ReleaseFont(hFont HFONT) HRESULT
	ResetFontMapping() HRESULT
	MapFont(hDC HDC, dwCodePages uint32, chSrc uint16, pFont *HFONT) HRESULT
	GetFontUnicodeRanges(hDC HDC, puiRanges *uint32, pUranges *UNICODERANGE) HRESULT
	GetScriptFontInfo(sid byte, dwFlags uint32, puiFonts *uint32, pScriptFont *TagSCRIPFONTINFO) HRESULT
	CodePageToScriptID(uiCodePage uint32, pSid *byte) HRESULT
}

type IMLangFontLink2Vtbl struct {
	IMLangCodePagesVtbl
	GetFontCodePages     uintptr
	ReleaseFont          uintptr
	ResetFontMapping     uintptr
	MapFont              uintptr
	GetFontUnicodeRanges uintptr
	GetScriptFontInfo    uintptr
	CodePageToScriptID   uintptr
}

type IMLangFontLink2 struct {
	IMLangCodePages
}

func (this *IMLangFontLink2) Vtbl() *IMLangFontLink2Vtbl {
	return (*IMLangFontLink2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMLangFontLink2) GetFontCodePages(hDC HDC, hFont HFONT, pdwCodePages *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFontCodePages, uintptr(unsafe.Pointer(this)), hDC, hFont, uintptr(unsafe.Pointer(pdwCodePages)))
	return HRESULT(ret)
}

func (this *IMLangFontLink2) ReleaseFont(hFont HFONT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReleaseFont, uintptr(unsafe.Pointer(this)), hFont)
	return HRESULT(ret)
}

func (this *IMLangFontLink2) ResetFontMapping() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ResetFontMapping, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IMLangFontLink2) MapFont(hDC HDC, dwCodePages uint32, chSrc uint16, pFont *HFONT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MapFont, uintptr(unsafe.Pointer(this)), hDC, uintptr(dwCodePages), uintptr(chSrc), uintptr(unsafe.Pointer(pFont)))
	return HRESULT(ret)
}

func (this *IMLangFontLink2) GetFontUnicodeRanges(hDC HDC, puiRanges *uint32, pUranges *UNICODERANGE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFontUnicodeRanges, uintptr(unsafe.Pointer(this)), hDC, uintptr(unsafe.Pointer(puiRanges)), uintptr(unsafe.Pointer(pUranges)))
	return HRESULT(ret)
}

func (this *IMLangFontLink2) GetScriptFontInfo(sid byte, dwFlags uint32, puiFonts *uint32, pScriptFont *TagSCRIPFONTINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetScriptFontInfo, uintptr(unsafe.Pointer(this)), uintptr(sid), uintptr(dwFlags), uintptr(unsafe.Pointer(puiFonts)), uintptr(unsafe.Pointer(pScriptFont)))
	return HRESULT(ret)
}

func (this *IMLangFontLink2) CodePageToScriptID(uiCodePage uint32, pSid *byte) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CodePageToScriptID, uintptr(unsafe.Pointer(this)), uintptr(uiCodePage), uintptr(unsafe.Pointer(pSid)))
	return HRESULT(ret)
}

// 4E5868AB-B157-4623-9ACC-6A1D9CAEBE04
var IID_IMultiLanguage3 = syscall.GUID{0x4E5868AB, 0xB157, 0x4623,
	[8]byte{0x9A, 0xCC, 0x6A, 0x1D, 0x9C, 0xAE, 0xBE, 0x04}}

type IMultiLanguage3Interface interface {
	IMultiLanguage2Interface
	DetectOutboundCodePage(dwFlags uint32, lpWideCharStr PWSTR, cchWideChar uint32, puiPreferredCodePages *uint32, nPreferredCodePages uint32, puiDetectedCodePages *uint32, pnDetectedCodePages *uint32, lpSpecialChar PWSTR) HRESULT
	DetectOutboundCodePageInIStream(dwFlags uint32, pStrIn *IStream, puiPreferredCodePages *uint32, nPreferredCodePages uint32, puiDetectedCodePages *uint32, pnDetectedCodePages *uint32, lpSpecialChar PWSTR) HRESULT
}

type IMultiLanguage3Vtbl struct {
	IMultiLanguage2Vtbl
	DetectOutboundCodePage          uintptr
	DetectOutboundCodePageInIStream uintptr
}

type IMultiLanguage3 struct {
	IMultiLanguage2
}

func (this *IMultiLanguage3) Vtbl() *IMultiLanguage3Vtbl {
	return (*IMultiLanguage3Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMultiLanguage3) DetectOutboundCodePage(dwFlags uint32, lpWideCharStr PWSTR, cchWideChar uint32, puiPreferredCodePages *uint32, nPreferredCodePages uint32, puiDetectedCodePages *uint32, pnDetectedCodePages *uint32, lpSpecialChar PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DetectOutboundCodePage, uintptr(unsafe.Pointer(this)), uintptr(dwFlags), uintptr(unsafe.Pointer(lpWideCharStr)), uintptr(cchWideChar), uintptr(unsafe.Pointer(puiPreferredCodePages)), uintptr(nPreferredCodePages), uintptr(unsafe.Pointer(puiDetectedCodePages)), uintptr(unsafe.Pointer(pnDetectedCodePages)), uintptr(unsafe.Pointer(lpSpecialChar)))
	return HRESULT(ret)
}

func (this *IMultiLanguage3) DetectOutboundCodePageInIStream(dwFlags uint32, pStrIn *IStream, puiPreferredCodePages *uint32, nPreferredCodePages uint32, puiDetectedCodePages *uint32, pnDetectedCodePages *uint32, lpSpecialChar PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DetectOutboundCodePageInIStream, uintptr(unsafe.Pointer(this)), uintptr(dwFlags), uintptr(unsafe.Pointer(pStrIn)), uintptr(unsafe.Pointer(puiPreferredCodePages)), uintptr(nPreferredCodePages), uintptr(unsafe.Pointer(puiDetectedCodePages)), uintptr(unsafe.Pointer(pnDetectedCodePages)), uintptr(unsafe.Pointer(lpSpecialChar)))
	return HRESULT(ret)
}

var (
	pGetTextCharset                    uintptr
	pGetTextCharsetInfo                uintptr
	pTranslateCharsetInfo              uintptr
	pGetDateFormatA                    uintptr
	pGetDateFormatW                    uintptr
	pGetTimeFormatA                    uintptr
	pGetTimeFormatW                    uintptr
	pGetTimeFormatEx                   uintptr
	pGetDateFormatEx                   uintptr
	pGetDurationFormatEx               uintptr
	pCompareStringEx                   uintptr
	pCompareStringOrdinal              uintptr
	pCompareStringW                    uintptr
	pFoldStringW                       uintptr
	pGetStringTypeExW                  uintptr
	pGetStringTypeW                    uintptr
	pMultiByteToWideChar               uintptr
	pWideCharToMultiByte               uintptr
	pIsValidCodePage                   uintptr
	pGetACP                            uintptr
	pGetOEMCP                          uintptr
	pGetCPInfo                         uintptr
	pGetCPInfoExA                      uintptr
	pGetCPInfoExW                      uintptr
	pCompareStringA                    uintptr
	pFindNLSString                     uintptr
	pLCMapStringW                      uintptr
	pLCMapStringA                      uintptr
	pGetLocaleInfoW                    uintptr
	pGetLocaleInfoA                    uintptr
	pSetLocaleInfoA                    uintptr
	pSetLocaleInfoW                    uintptr
	pGetCalendarInfoA                  uintptr
	pGetCalendarInfoW                  uintptr
	pSetCalendarInfoA                  uintptr
	pSetCalendarInfoW                  uintptr
	pIsDBCSLeadByte                    uintptr
	pIsDBCSLeadByteEx                  uintptr
	pLocaleNameToLCID                  uintptr
	pLCIDToLocaleName                  uintptr
	pGetDurationFormat                 uintptr
	pGetNumberFormatA                  uintptr
	pGetNumberFormatW                  uintptr
	pGetCurrencyFormatA                uintptr
	pGetCurrencyFormatW                uintptr
	pEnumCalendarInfoA                 uintptr
	pEnumCalendarInfoW                 uintptr
	pEnumCalendarInfoExA               uintptr
	pEnumCalendarInfoExW               uintptr
	pEnumTimeFormatsA                  uintptr
	pEnumTimeFormatsW                  uintptr
	pEnumDateFormatsA                  uintptr
	pEnumDateFormatsW                  uintptr
	pEnumDateFormatsExA                uintptr
	pEnumDateFormatsExW                uintptr
	pIsValidLanguageGroup              uintptr
	pGetNLSVersion                     uintptr
	pIsValidLocale                     uintptr
	pGetGeoInfoA                       uintptr
	pGetGeoInfoW                       uintptr
	pGetGeoInfoEx                      uintptr
	pEnumSystemGeoID                   uintptr
	pEnumSystemGeoNames                uintptr
	pGetUserGeoID                      uintptr
	pGetUserDefaultGeoName             uintptr
	pSetUserGeoID                      uintptr
	pSetUserGeoName                    uintptr
	pConvertDefaultLocale              uintptr
	pGetSystemDefaultUILanguage        uintptr
	pGetThreadLocale                   uintptr
	pSetThreadLocale                   uintptr
	pGetUserDefaultUILanguage          uintptr
	pGetUserDefaultLangID              uintptr
	pGetSystemDefaultLangID            uintptr
	pGetSystemDefaultLCID              uintptr
	pGetUserDefaultLCID                uintptr
	pSetThreadUILanguage               uintptr
	pGetThreadUILanguage               uintptr
	pGetProcessPreferredUILanguages    uintptr
	pSetProcessPreferredUILanguages    uintptr
	pGetUserPreferredUILanguages       uintptr
	pGetSystemPreferredUILanguages     uintptr
	pGetThreadPreferredUILanguages     uintptr
	pSetThreadPreferredUILanguages     uintptr
	pGetFileMUIInfo                    uintptr
	pGetFileMUIPath                    uintptr
	pGetUILanguageInfo                 uintptr
	pSetThreadPreferredUILanguages2    uintptr
	pRestoreThreadPreferredUILanguages uintptr
	pNotifyUILanguageChange            uintptr
	pGetStringTypeExA                  uintptr
	pGetStringTypeA                    uintptr
	pFoldStringA                       uintptr
	pEnumSystemLocalesA                uintptr
	pEnumSystemLocalesW                uintptr
	pEnumSystemLanguageGroupsA         uintptr
	pEnumSystemLanguageGroupsW         uintptr
	pEnumLanguageGroupLocalesA         uintptr
	pEnumLanguageGroupLocalesW         uintptr
	pEnumUILanguagesA                  uintptr
	pEnumUILanguagesW                  uintptr
	pEnumSystemCodePagesA              uintptr
	pEnumSystemCodePagesW              uintptr
	pIdnToNameprepUnicode              uintptr
	pNormalizeString                   uintptr
	pIsNormalizedString                uintptr
	pVerifyScripts                     uintptr
	pGetStringScripts                  uintptr
	pGetLocaleInfoEx                   uintptr
	pGetCalendarInfoEx                 uintptr
	pGetNumberFormatEx                 uintptr
	pGetCurrencyFormatEx               uintptr
	pGetUserDefaultLocaleName          uintptr
	pGetSystemDefaultLocaleName        uintptr
	pIsNLSDefinedString                uintptr
	pGetNLSVersionEx                   uintptr
	pIsValidNLSVersion                 uintptr
	pFindNLSStringEx                   uintptr
	pLCMapStringEx                     uintptr
	pIsValidLocaleName                 uintptr
	pEnumCalendarInfoExEx              uintptr
	pEnumDateFormatsExEx               uintptr
	pEnumTimeFormatsEx                 uintptr
	pEnumSystemLocalesEx               uintptr
	pResolveLocaleName                 uintptr
	pFindStringOrdinal                 uintptr
	pLstrcmpA                          uintptr
	pLstrcmpW                          uintptr
	pLstrcmpiA                         uintptr
	pLstrcmpiW                         uintptr
	pLstrcpynA                         uintptr
	pLstrcpynW                         uintptr
	pLstrcpyA                          uintptr
	pLstrcpyW                          uintptr
	pLstrcatA                          uintptr
	pLstrcatW                          uintptr
	pLstrlenA                          uintptr
	pLstrlenW                          uintptr
	pIsTextUnicode                     uintptr
)

func GetTextCharset(hdc HDC) int32 {
	addr := lazyAddr(&pGetTextCharset, libGdi32, "GetTextCharset")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return int32(ret)
}

func GetTextCharsetInfo(hdc HDC, lpSig *FONTSIGNATURE, dwFlags uint32) int32 {
	addr := lazyAddr(&pGetTextCharsetInfo, libGdi32, "GetTextCharsetInfo")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpSig)), uintptr(dwFlags))
	return int32(ret)
}

func TranslateCharsetInfo(lpSrc *uint32, lpCs *CHARSETINFO, dwFlags TRANSLATE_CHARSET_INFO_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pTranslateCharsetInfo, libGdi32, "TranslateCharsetInfo")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSrc)), uintptr(unsafe.Pointer(lpCs)), uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetDateFormatA(Locale uint32, dwFlags uint32, lpDate *SYSTEMTIME, lpFormat PSTR, lpDateStr PSTR, cchDate int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetDateFormatA, libKernel32, "GetDateFormatA")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwFlags), uintptr(unsafe.Pointer(lpDate)), uintptr(unsafe.Pointer(lpFormat)), uintptr(unsafe.Pointer(lpDateStr)), uintptr(cchDate))
	return int32(ret), WIN32_ERROR(err)
}

var GetDateFormat = GetDateFormatW

func GetDateFormatW(Locale uint32, dwFlags uint32, lpDate *SYSTEMTIME, lpFormat PWSTR, lpDateStr PWSTR, cchDate int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetDateFormatW, libKernel32, "GetDateFormatW")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwFlags), uintptr(unsafe.Pointer(lpDate)), uintptr(unsafe.Pointer(lpFormat)), uintptr(unsafe.Pointer(lpDateStr)), uintptr(cchDate))
	return int32(ret), WIN32_ERROR(err)
}

func GetTimeFormatA(Locale uint32, dwFlags uint32, lpTime *SYSTEMTIME, lpFormat PSTR, lpTimeStr PSTR, cchTime int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetTimeFormatA, libKernel32, "GetTimeFormatA")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwFlags), uintptr(unsafe.Pointer(lpTime)), uintptr(unsafe.Pointer(lpFormat)), uintptr(unsafe.Pointer(lpTimeStr)), uintptr(cchTime))
	return int32(ret), WIN32_ERROR(err)
}

var GetTimeFormat = GetTimeFormatW

func GetTimeFormatW(Locale uint32, dwFlags uint32, lpTime *SYSTEMTIME, lpFormat PWSTR, lpTimeStr PWSTR, cchTime int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetTimeFormatW, libKernel32, "GetTimeFormatW")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwFlags), uintptr(unsafe.Pointer(lpTime)), uintptr(unsafe.Pointer(lpFormat)), uintptr(unsafe.Pointer(lpTimeStr)), uintptr(cchTime))
	return int32(ret), WIN32_ERROR(err)
}

func GetTimeFormatEx(lpLocaleName PWSTR, dwFlags TIME_FORMAT_FLAGS, lpTime *SYSTEMTIME, lpFormat PWSTR, lpTimeStr PWSTR, cchTime int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetTimeFormatEx, libKernel32, "GetTimeFormatEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(dwFlags), uintptr(unsafe.Pointer(lpTime)), uintptr(unsafe.Pointer(lpFormat)), uintptr(unsafe.Pointer(lpTimeStr)), uintptr(cchTime))
	return int32(ret), WIN32_ERROR(err)
}

func GetDateFormatEx(lpLocaleName PWSTR, dwFlags ENUM_DATE_FORMATS_FLAGS, lpDate *SYSTEMTIME, lpFormat PWSTR, lpDateStr PWSTR, cchDate int32, lpCalendar PWSTR) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetDateFormatEx, libKernel32, "GetDateFormatEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(dwFlags), uintptr(unsafe.Pointer(lpDate)), uintptr(unsafe.Pointer(lpFormat)), uintptr(unsafe.Pointer(lpDateStr)), uintptr(cchDate), uintptr(unsafe.Pointer(lpCalendar)))
	return int32(ret), WIN32_ERROR(err)
}

func GetDurationFormatEx(lpLocaleName PWSTR, dwFlags uint32, lpDuration *SYSTEMTIME, ullDuration uint64, lpFormat PWSTR, lpDurationStr PWSTR, cchDuration int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetDurationFormatEx, libKernel32, "GetDurationFormatEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(dwFlags), uintptr(unsafe.Pointer(lpDuration)), uintptr(ullDuration), uintptr(unsafe.Pointer(lpFormat)), uintptr(unsafe.Pointer(lpDurationStr)), uintptr(cchDuration))
	return int32(ret), WIN32_ERROR(err)
}

func CompareStringEx(lpLocaleName PWSTR, dwCmpFlags COMPARE_STRING_FLAGS, lpString1 PWSTR, cchCount1 int32, lpString2 PWSTR, cchCount2 int32, lpVersionInformation *NLSVERSIONINFO, lpReserved unsafe.Pointer, lParam LPARAM) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pCompareStringEx, libKernel32, "CompareStringEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(dwCmpFlags), uintptr(unsafe.Pointer(lpString1)), uintptr(cchCount1), uintptr(unsafe.Pointer(lpString2)), uintptr(cchCount2), uintptr(unsafe.Pointer(lpVersionInformation)), uintptr(lpReserved), lParam)
	return int32(ret), WIN32_ERROR(err)
}

func CompareStringOrdinal(lpString1 PWSTR, cchCount1 int32, lpString2 PWSTR, cchCount2 int32, bIgnoreCase BOOL) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pCompareStringOrdinal, libKernel32, "CompareStringOrdinal")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString1)), uintptr(cchCount1), uintptr(unsafe.Pointer(lpString2)), uintptr(cchCount2), uintptr(bIgnoreCase))
	return int32(ret), WIN32_ERROR(err)
}

var CompareString = CompareStringW

func CompareStringW(Locale uint32, dwCmpFlags uint32, lpString1 PWSTR, cchCount1 int32, lpString2 PWSTR, cchCount2 int32) int32 {
	addr := lazyAddr(&pCompareStringW, libKernel32, "CompareStringW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwCmpFlags), uintptr(unsafe.Pointer(lpString1)), uintptr(cchCount1), uintptr(unsafe.Pointer(lpString2)), uintptr(cchCount2))
	return int32(ret)
}

var FoldString = FoldStringW

func FoldStringW(dwMapFlags FOLD_STRING_MAP_FLAGS, lpSrcStr PWSTR, cchSrc int32, lpDestStr PWSTR, cchDest int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pFoldStringW, libKernel32, "FoldStringW")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwMapFlags), uintptr(unsafe.Pointer(lpSrcStr)), uintptr(cchSrc), uintptr(unsafe.Pointer(lpDestStr)), uintptr(cchDest))
	return int32(ret), WIN32_ERROR(err)
}

var GetStringTypeEx = GetStringTypeExW

func GetStringTypeExW(Locale uint32, dwInfoType uint32, lpSrcStr PWSTR, cchSrc int32, lpCharType *uint16) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetStringTypeExW, libKernel32, "GetStringTypeExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwInfoType), uintptr(unsafe.Pointer(lpSrcStr)), uintptr(cchSrc), uintptr(unsafe.Pointer(lpCharType)))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetStringType = GetStringTypeW

func GetStringTypeW(dwInfoType uint32, lpSrcStr PWSTR, cchSrc int32, lpCharType *uint16) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetStringTypeW, libKernel32, "GetStringTypeW")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwInfoType), uintptr(unsafe.Pointer(lpSrcStr)), uintptr(cchSrc), uintptr(unsafe.Pointer(lpCharType)))
	return BOOL(ret), WIN32_ERROR(err)
}

func MultiByteToWideChar(CodePage uint32, dwFlags MULTI_BYTE_TO_WIDE_CHAR_FLAGS, lpMultiByteStr PSTR, cbMultiByte int32, lpWideCharStr PWSTR, cchWideChar int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pMultiByteToWideChar, libKernel32, "MultiByteToWideChar")
	ret, _, err := syscall.SyscallN(addr, uintptr(CodePage), uintptr(dwFlags), uintptr(unsafe.Pointer(lpMultiByteStr)), uintptr(cbMultiByte), uintptr(unsafe.Pointer(lpWideCharStr)), uintptr(cchWideChar))
	return int32(ret), WIN32_ERROR(err)
}

func WideCharToMultiByte(CodePage uint32, dwFlags uint32, lpWideCharStr PWSTR, cchWideChar int32, lpMultiByteStr PSTR, cbMultiByte int32, lpDefaultChar PSTR, lpUsedDefaultChar *int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pWideCharToMultiByte, libKernel32, "WideCharToMultiByte")
	ret, _, err := syscall.SyscallN(addr, uintptr(CodePage), uintptr(dwFlags), uintptr(unsafe.Pointer(lpWideCharStr)), uintptr(cchWideChar), uintptr(unsafe.Pointer(lpMultiByteStr)), uintptr(cbMultiByte), uintptr(unsafe.Pointer(lpDefaultChar)), uintptr(unsafe.Pointer(lpUsedDefaultChar)))
	return int32(ret), WIN32_ERROR(err)
}

func IsValidCodePage(CodePage uint32) BOOL {
	addr := lazyAddr(&pIsValidCodePage, libKernel32, "IsValidCodePage")
	ret, _, _ := syscall.SyscallN(addr, uintptr(CodePage))
	return BOOL(ret)
}

func GetACP() uint32 {
	addr := lazyAddr(&pGetACP, libKernel32, "GetACP")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func GetOEMCP() uint32 {
	addr := lazyAddr(&pGetOEMCP, libKernel32, "GetOEMCP")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func GetCPInfo(CodePage uint32, lpCPInfo *CPINFO) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetCPInfo, libKernel32, "GetCPInfo")
	ret, _, err := syscall.SyscallN(addr, uintptr(CodePage), uintptr(unsafe.Pointer(lpCPInfo)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetCPInfoExA(CodePage uint32, dwFlags uint32, lpCPInfoEx *CPINFOEXA) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetCPInfoExA, libKernel32, "GetCPInfoExA")
	ret, _, err := syscall.SyscallN(addr, uintptr(CodePage), uintptr(dwFlags), uintptr(unsafe.Pointer(lpCPInfoEx)))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetCPInfoEx = GetCPInfoExW

func GetCPInfoExW(CodePage uint32, dwFlags uint32, lpCPInfoEx *CPINFOEXW) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetCPInfoExW, libKernel32, "GetCPInfoExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(CodePage), uintptr(dwFlags), uintptr(unsafe.Pointer(lpCPInfoEx)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CompareStringA(Locale uint32, dwCmpFlags uint32, lpString1 *int8, cchCount1 int32, lpString2 *int8, cchCount2 int32) int32 {
	addr := lazyAddr(&pCompareStringA, libKernel32, "CompareStringA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwCmpFlags), uintptr(unsafe.Pointer(lpString1)), uintptr(cchCount1), uintptr(unsafe.Pointer(lpString2)), uintptr(cchCount2))
	return int32(ret)
}

func FindNLSString(Locale uint32, dwFindNLSStringFlags uint32, lpStringSource PWSTR, cchSource int32, lpStringValue PWSTR, cchValue int32, pcchFound *int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pFindNLSString, libKernel32, "FindNLSString")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwFindNLSStringFlags), uintptr(unsafe.Pointer(lpStringSource)), uintptr(cchSource), uintptr(unsafe.Pointer(lpStringValue)), uintptr(cchValue), uintptr(unsafe.Pointer(pcchFound)))
	return int32(ret), WIN32_ERROR(err)
}

var LCMapString = LCMapStringW

func LCMapStringW(Locale uint32, dwMapFlags uint32, lpSrcStr PWSTR, cchSrc int32, lpDestStr PWSTR, cchDest int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pLCMapStringW, libKernel32, "LCMapStringW")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwMapFlags), uintptr(unsafe.Pointer(lpSrcStr)), uintptr(cchSrc), uintptr(unsafe.Pointer(lpDestStr)), uintptr(cchDest))
	return int32(ret), WIN32_ERROR(err)
}

func LCMapStringA(Locale uint32, dwMapFlags uint32, lpSrcStr PSTR, cchSrc int32, lpDestStr PSTR, cchDest int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pLCMapStringA, libKernel32, "LCMapStringA")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwMapFlags), uintptr(unsafe.Pointer(lpSrcStr)), uintptr(cchSrc), uintptr(unsafe.Pointer(lpDestStr)), uintptr(cchDest))
	return int32(ret), WIN32_ERROR(err)
}

var GetLocaleInfo = GetLocaleInfoW

func GetLocaleInfoW(Locale uint32, LCType uint32, lpLCData PWSTR, cchData int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetLocaleInfoW, libKernel32, "GetLocaleInfoW")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(LCType), uintptr(unsafe.Pointer(lpLCData)), uintptr(cchData))
	return int32(ret), WIN32_ERROR(err)
}

func GetLocaleInfoA(Locale uint32, LCType uint32, lpLCData PSTR, cchData int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetLocaleInfoA, libKernel32, "GetLocaleInfoA")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(LCType), uintptr(unsafe.Pointer(lpLCData)), uintptr(cchData))
	return int32(ret), WIN32_ERROR(err)
}

func SetLocaleInfoA(Locale uint32, LCType uint32, lpLCData PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetLocaleInfoA, libKernel32, "SetLocaleInfoA")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(LCType), uintptr(unsafe.Pointer(lpLCData)))
	return BOOL(ret), WIN32_ERROR(err)
}

var SetLocaleInfo = SetLocaleInfoW

func SetLocaleInfoW(Locale uint32, LCType uint32, lpLCData PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetLocaleInfoW, libKernel32, "SetLocaleInfoW")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(LCType), uintptr(unsafe.Pointer(lpLCData)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetCalendarInfoA(Locale uint32, Calendar uint32, CalType uint32, lpCalData PSTR, cchData int32, lpValue *uint32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetCalendarInfoA, libKernel32, "GetCalendarInfoA")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(Calendar), uintptr(CalType), uintptr(unsafe.Pointer(lpCalData)), uintptr(cchData), uintptr(unsafe.Pointer(lpValue)))
	return int32(ret), WIN32_ERROR(err)
}

var GetCalendarInfo = GetCalendarInfoW

func GetCalendarInfoW(Locale uint32, Calendar uint32, CalType uint32, lpCalData PWSTR, cchData int32, lpValue *uint32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetCalendarInfoW, libKernel32, "GetCalendarInfoW")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(Calendar), uintptr(CalType), uintptr(unsafe.Pointer(lpCalData)), uintptr(cchData), uintptr(unsafe.Pointer(lpValue)))
	return int32(ret), WIN32_ERROR(err)
}

func SetCalendarInfoA(Locale uint32, Calendar uint32, CalType uint32, lpCalData PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetCalendarInfoA, libKernel32, "SetCalendarInfoA")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(Calendar), uintptr(CalType), uintptr(unsafe.Pointer(lpCalData)))
	return BOOL(ret), WIN32_ERROR(err)
}

var SetCalendarInfo = SetCalendarInfoW

func SetCalendarInfoW(Locale uint32, Calendar uint32, CalType uint32, lpCalData PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetCalendarInfoW, libKernel32, "SetCalendarInfoW")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(Calendar), uintptr(CalType), uintptr(unsafe.Pointer(lpCalData)))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsDBCSLeadByte(TestChar byte) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsDBCSLeadByte, libKernel32, "IsDBCSLeadByte")
	ret, _, err := syscall.SyscallN(addr, uintptr(TestChar))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsDBCSLeadByteEx(CodePage uint32, TestChar byte) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsDBCSLeadByteEx, libKernel32, "IsDBCSLeadByteEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(CodePage), uintptr(TestChar))
	return BOOL(ret), WIN32_ERROR(err)
}

func LocaleNameToLCID(lpName PWSTR, dwFlags uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pLocaleNameToLCID, libKernel32, "LocaleNameToLCID")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)), uintptr(dwFlags))
	return uint32(ret), WIN32_ERROR(err)
}

func LCIDToLocaleName(Locale uint32, lpName PWSTR, cchName int32, dwFlags uint32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pLCIDToLocaleName, libKernel32, "LCIDToLocaleName")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(unsafe.Pointer(lpName)), uintptr(cchName), uintptr(dwFlags))
	return int32(ret), WIN32_ERROR(err)
}

func GetDurationFormat(Locale uint32, dwFlags uint32, lpDuration *SYSTEMTIME, ullDuration uint64, lpFormat PWSTR, lpDurationStr PWSTR, cchDuration int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetDurationFormat, libKernel32, "GetDurationFormat")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwFlags), uintptr(unsafe.Pointer(lpDuration)), uintptr(ullDuration), uintptr(unsafe.Pointer(lpFormat)), uintptr(unsafe.Pointer(lpDurationStr)), uintptr(cchDuration))
	return int32(ret), WIN32_ERROR(err)
}

func GetNumberFormatA(Locale uint32, dwFlags uint32, lpValue PSTR, lpFormat *NUMBERFMTA, lpNumberStr PSTR, cchNumber int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetNumberFormatA, libKernel32, "GetNumberFormatA")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwFlags), uintptr(unsafe.Pointer(lpValue)), uintptr(unsafe.Pointer(lpFormat)), uintptr(unsafe.Pointer(lpNumberStr)), uintptr(cchNumber))
	return int32(ret), WIN32_ERROR(err)
}

var GetNumberFormat = GetNumberFormatW

func GetNumberFormatW(Locale uint32, dwFlags uint32, lpValue PWSTR, lpFormat *NUMBERFMTW, lpNumberStr PWSTR, cchNumber int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetNumberFormatW, libKernel32, "GetNumberFormatW")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwFlags), uintptr(unsafe.Pointer(lpValue)), uintptr(unsafe.Pointer(lpFormat)), uintptr(unsafe.Pointer(lpNumberStr)), uintptr(cchNumber))
	return int32(ret), WIN32_ERROR(err)
}

func GetCurrencyFormatA(Locale uint32, dwFlags uint32, lpValue PSTR, lpFormat *CURRENCYFMTA, lpCurrencyStr PSTR, cchCurrency int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetCurrencyFormatA, libKernel32, "GetCurrencyFormatA")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwFlags), uintptr(unsafe.Pointer(lpValue)), uintptr(unsafe.Pointer(lpFormat)), uintptr(unsafe.Pointer(lpCurrencyStr)), uintptr(cchCurrency))
	return int32(ret), WIN32_ERROR(err)
}

var GetCurrencyFormat = GetCurrencyFormatW

func GetCurrencyFormatW(Locale uint32, dwFlags uint32, lpValue PWSTR, lpFormat *CURRENCYFMTW, lpCurrencyStr PWSTR, cchCurrency int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetCurrencyFormatW, libKernel32, "GetCurrencyFormatW")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwFlags), uintptr(unsafe.Pointer(lpValue)), uintptr(unsafe.Pointer(lpFormat)), uintptr(unsafe.Pointer(lpCurrencyStr)), uintptr(cchCurrency))
	return int32(ret), WIN32_ERROR(err)
}

func EnumCalendarInfoA(lpCalInfoEnumProc CALINFO_ENUMPROCA, Locale uint32, Calendar uint32, CalType uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumCalendarInfoA, libKernel32, "EnumCalendarInfoA")
	ret, _, err := syscall.SyscallN(addr, lpCalInfoEnumProc, uintptr(Locale), uintptr(Calendar), uintptr(CalType))
	return BOOL(ret), WIN32_ERROR(err)
}

var EnumCalendarInfo = EnumCalendarInfoW

func EnumCalendarInfoW(lpCalInfoEnumProc CALINFO_ENUMPROCW, Locale uint32, Calendar uint32, CalType uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumCalendarInfoW, libKernel32, "EnumCalendarInfoW")
	ret, _, err := syscall.SyscallN(addr, lpCalInfoEnumProc, uintptr(Locale), uintptr(Calendar), uintptr(CalType))
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumCalendarInfoExA(lpCalInfoEnumProcEx CALINFO_ENUMPROCEXA, Locale uint32, Calendar uint32, CalType uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumCalendarInfoExA, libKernel32, "EnumCalendarInfoExA")
	ret, _, err := syscall.SyscallN(addr, lpCalInfoEnumProcEx, uintptr(Locale), uintptr(Calendar), uintptr(CalType))
	return BOOL(ret), WIN32_ERROR(err)
}

var EnumCalendarInfoEx = EnumCalendarInfoExW

func EnumCalendarInfoExW(lpCalInfoEnumProcEx CALINFO_ENUMPROCEXW, Locale uint32, Calendar uint32, CalType uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumCalendarInfoExW, libKernel32, "EnumCalendarInfoExW")
	ret, _, err := syscall.SyscallN(addr, lpCalInfoEnumProcEx, uintptr(Locale), uintptr(Calendar), uintptr(CalType))
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumTimeFormatsA(lpTimeFmtEnumProc TIMEFMT_ENUMPROCA, Locale uint32, dwFlags TIME_FORMAT_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumTimeFormatsA, libKernel32, "EnumTimeFormatsA")
	ret, _, err := syscall.SyscallN(addr, lpTimeFmtEnumProc, uintptr(Locale), uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

var EnumTimeFormats = EnumTimeFormatsW

func EnumTimeFormatsW(lpTimeFmtEnumProc TIMEFMT_ENUMPROCW, Locale uint32, dwFlags TIME_FORMAT_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumTimeFormatsW, libKernel32, "EnumTimeFormatsW")
	ret, _, err := syscall.SyscallN(addr, lpTimeFmtEnumProc, uintptr(Locale), uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumDateFormatsA(lpDateFmtEnumProc DATEFMT_ENUMPROCA, Locale uint32, dwFlags uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumDateFormatsA, libKernel32, "EnumDateFormatsA")
	ret, _, err := syscall.SyscallN(addr, lpDateFmtEnumProc, uintptr(Locale), uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

var EnumDateFormats = EnumDateFormatsW

func EnumDateFormatsW(lpDateFmtEnumProc DATEFMT_ENUMPROCW, Locale uint32, dwFlags uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumDateFormatsW, libKernel32, "EnumDateFormatsW")
	ret, _, err := syscall.SyscallN(addr, lpDateFmtEnumProc, uintptr(Locale), uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumDateFormatsExA(lpDateFmtEnumProcEx DATEFMT_ENUMPROCEXA, Locale uint32, dwFlags uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumDateFormatsExA, libKernel32, "EnumDateFormatsExA")
	ret, _, err := syscall.SyscallN(addr, lpDateFmtEnumProcEx, uintptr(Locale), uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

var EnumDateFormatsEx = EnumDateFormatsExW

func EnumDateFormatsExW(lpDateFmtEnumProcEx DATEFMT_ENUMPROCEXW, Locale uint32, dwFlags uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumDateFormatsExW, libKernel32, "EnumDateFormatsExW")
	ret, _, err := syscall.SyscallN(addr, lpDateFmtEnumProcEx, uintptr(Locale), uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsValidLanguageGroup(LanguageGroup uint32, dwFlags ENUM_SYSTEM_LANGUAGE_GROUPS_FLAGS) BOOL {
	addr := lazyAddr(&pIsValidLanguageGroup, libKernel32, "IsValidLanguageGroup")
	ret, _, _ := syscall.SyscallN(addr, uintptr(LanguageGroup), uintptr(dwFlags))
	return BOOL(ret)
}

func GetNLSVersion(Function uint32, Locale uint32, lpVersionInformation *NLSVERSIONINFO) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNLSVersion, libKernel32, "GetNLSVersion")
	ret, _, err := syscall.SyscallN(addr, uintptr(Function), uintptr(Locale), uintptr(unsafe.Pointer(lpVersionInformation)))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsValidLocale(Locale uint32, dwFlags IS_VALID_LOCALE_FLAGS) BOOL {
	addr := lazyAddr(&pIsValidLocale, libKernel32, "IsValidLocale")
	ret, _, _ := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwFlags))
	return BOOL(ret)
}

func GetGeoInfoA(Location int32, GeoType uint32, lpGeoData PSTR, cchData int32, LangId uint16) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetGeoInfoA, libKernel32, "GetGeoInfoA")
	ret, _, err := syscall.SyscallN(addr, uintptr(Location), uintptr(GeoType), uintptr(unsafe.Pointer(lpGeoData)), uintptr(cchData), uintptr(LangId))
	return int32(ret), WIN32_ERROR(err)
}

var GetGeoInfo = GetGeoInfoW

func GetGeoInfoW(Location int32, GeoType uint32, lpGeoData PWSTR, cchData int32, LangId uint16) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetGeoInfoW, libKernel32, "GetGeoInfoW")
	ret, _, err := syscall.SyscallN(addr, uintptr(Location), uintptr(GeoType), uintptr(unsafe.Pointer(lpGeoData)), uintptr(cchData), uintptr(LangId))
	return int32(ret), WIN32_ERROR(err)
}

func GetGeoInfoEx(location PWSTR, geoType uint32, geoData PWSTR, geoDataCount int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetGeoInfoEx, libKernel32, "GetGeoInfoEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(location)), uintptr(geoType), uintptr(unsafe.Pointer(geoData)), uintptr(geoDataCount))
	return int32(ret), WIN32_ERROR(err)
}

func EnumSystemGeoID(GeoClass uint32, ParentGeoId int32, lpGeoEnumProc GEO_ENUMPROC) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumSystemGeoID, libKernel32, "EnumSystemGeoID")
	ret, _, err := syscall.SyscallN(addr, uintptr(GeoClass), uintptr(ParentGeoId), lpGeoEnumProc)
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumSystemGeoNames(geoClass uint32, geoEnumProc GEO_ENUMNAMEPROC, data LPARAM) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumSystemGeoNames, libKernel32, "EnumSystemGeoNames")
	ret, _, err := syscall.SyscallN(addr, uintptr(geoClass), geoEnumProc, data)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetUserGeoID(GeoClass uint32) int32 {
	addr := lazyAddr(&pGetUserGeoID, libKernel32, "GetUserGeoID")
	ret, _, _ := syscall.SyscallN(addr, uintptr(GeoClass))
	return int32(ret)
}

func GetUserDefaultGeoName(geoName PWSTR, geoNameCount int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetUserDefaultGeoName, libKernel32, "GetUserDefaultGeoName")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(geoName)), uintptr(geoNameCount))
	return int32(ret), WIN32_ERROR(err)
}

func SetUserGeoID(GeoId int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetUserGeoID, libKernel32, "SetUserGeoID")
	ret, _, err := syscall.SyscallN(addr, uintptr(GeoId))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetUserGeoName(geoName PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetUserGeoName, libKernel32, "SetUserGeoName")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(geoName)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ConvertDefaultLocale(Locale uint32) uint32 {
	addr := lazyAddr(&pConvertDefaultLocale, libKernel32, "ConvertDefaultLocale")
	ret, _, _ := syscall.SyscallN(addr, uintptr(Locale))
	return uint32(ret)
}

func GetSystemDefaultUILanguage() uint16 {
	addr := lazyAddr(&pGetSystemDefaultUILanguage, libKernel32, "GetSystemDefaultUILanguage")
	ret, _, _ := syscall.SyscallN(addr)
	return uint16(ret)
}

func GetThreadLocale() uint32 {
	addr := lazyAddr(&pGetThreadLocale, libKernel32, "GetThreadLocale")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func SetThreadLocale(Locale uint32) BOOL {
	addr := lazyAddr(&pSetThreadLocale, libKernel32, "SetThreadLocale")
	ret, _, _ := syscall.SyscallN(addr, uintptr(Locale))
	return BOOL(ret)
}

func GetUserDefaultUILanguage() uint16 {
	addr := lazyAddr(&pGetUserDefaultUILanguage, libKernel32, "GetUserDefaultUILanguage")
	ret, _, _ := syscall.SyscallN(addr)
	return uint16(ret)
}

func GetUserDefaultLangID() uint16 {
	addr := lazyAddr(&pGetUserDefaultLangID, libKernel32, "GetUserDefaultLangID")
	ret, _, _ := syscall.SyscallN(addr)
	return uint16(ret)
}

func GetSystemDefaultLangID() uint16 {
	addr := lazyAddr(&pGetSystemDefaultLangID, libKernel32, "GetSystemDefaultLangID")
	ret, _, _ := syscall.SyscallN(addr)
	return uint16(ret)
}

func GetSystemDefaultLCID() uint32 {
	addr := lazyAddr(&pGetSystemDefaultLCID, libKernel32, "GetSystemDefaultLCID")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func GetUserDefaultLCID() uint32 {
	addr := lazyAddr(&pGetUserDefaultLCID, libKernel32, "GetUserDefaultLCID")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func SetThreadUILanguage(LangId uint16) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pSetThreadUILanguage, libKernel32, "SetThreadUILanguage")
	ret, _, err := syscall.SyscallN(addr, uintptr(LangId))
	return uint16(ret), WIN32_ERROR(err)
}

func GetThreadUILanguage() uint16 {
	addr := lazyAddr(&pGetThreadUILanguage, libKernel32, "GetThreadUILanguage")
	ret, _, _ := syscall.SyscallN(addr)
	return uint16(ret)
}

func GetProcessPreferredUILanguages(dwFlags uint32, pulNumLanguages *uint32, pwszLanguagesBuffer PWSTR, pcchLanguagesBuffer *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessPreferredUILanguages, libKernel32, "GetProcessPreferredUILanguages")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(unsafe.Pointer(pulNumLanguages)), uintptr(unsafe.Pointer(pwszLanguagesBuffer)), uintptr(unsafe.Pointer(pcchLanguagesBuffer)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetProcessPreferredUILanguages(dwFlags uint32, pwszLanguagesBuffer PWSTR, pulNumLanguages *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProcessPreferredUILanguages, libKernel32, "SetProcessPreferredUILanguages")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(unsafe.Pointer(pwszLanguagesBuffer)), uintptr(unsafe.Pointer(pulNumLanguages)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetUserPreferredUILanguages(dwFlags uint32, pulNumLanguages *uint32, pwszLanguagesBuffer PWSTR, pcchLanguagesBuffer *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetUserPreferredUILanguages, libKernel32, "GetUserPreferredUILanguages")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(unsafe.Pointer(pulNumLanguages)), uintptr(unsafe.Pointer(pwszLanguagesBuffer)), uintptr(unsafe.Pointer(pcchLanguagesBuffer)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetSystemPreferredUILanguages(dwFlags uint32, pulNumLanguages *uint32, pwszLanguagesBuffer PWSTR, pcchLanguagesBuffer *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetSystemPreferredUILanguages, libKernel32, "GetSystemPreferredUILanguages")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(unsafe.Pointer(pulNumLanguages)), uintptr(unsafe.Pointer(pwszLanguagesBuffer)), uintptr(unsafe.Pointer(pcchLanguagesBuffer)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetThreadPreferredUILanguages(dwFlags uint32, pulNumLanguages *uint32, pwszLanguagesBuffer PWSTR, pcchLanguagesBuffer *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetThreadPreferredUILanguages, libKernel32, "GetThreadPreferredUILanguages")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(unsafe.Pointer(pulNumLanguages)), uintptr(unsafe.Pointer(pwszLanguagesBuffer)), uintptr(unsafe.Pointer(pcchLanguagesBuffer)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetThreadPreferredUILanguages(dwFlags uint32, pwszLanguagesBuffer PWSTR, pulNumLanguages *uint32) BOOL {
	addr := lazyAddr(&pSetThreadPreferredUILanguages, libKernel32, "SetThreadPreferredUILanguages")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(unsafe.Pointer(pwszLanguagesBuffer)), uintptr(unsafe.Pointer(pulNumLanguages)))
	return BOOL(ret)
}

func GetFileMUIInfo(dwFlags uint32, pcwszFilePath PWSTR, pFileMUIInfo *FILEMUIINFO, pcbFileMUIInfo *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetFileMUIInfo, libKernel32, "GetFileMUIInfo")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(unsafe.Pointer(pcwszFilePath)), uintptr(unsafe.Pointer(pFileMUIInfo)), uintptr(unsafe.Pointer(pcbFileMUIInfo)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetFileMUIPath(dwFlags uint32, pcwszFilePath PWSTR, pwszLanguage PWSTR, pcchLanguage *uint32, pwszFileMUIPath PWSTR, pcchFileMUIPath *uint32, pululEnumerator *uint64) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetFileMUIPath, libKernel32, "GetFileMUIPath")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(unsafe.Pointer(pcwszFilePath)), uintptr(unsafe.Pointer(pwszLanguage)), uintptr(unsafe.Pointer(pcchLanguage)), uintptr(unsafe.Pointer(pwszFileMUIPath)), uintptr(unsafe.Pointer(pcchFileMUIPath)), uintptr(unsafe.Pointer(pululEnumerator)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetUILanguageInfo(dwFlags uint32, pwmszLanguage PWSTR, pwszFallbackLanguages PWSTR, pcchFallbackLanguages *uint32, pAttributes *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetUILanguageInfo, libKernel32, "GetUILanguageInfo")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(unsafe.Pointer(pwmszLanguage)), uintptr(unsafe.Pointer(pwszFallbackLanguages)), uintptr(unsafe.Pointer(pcchFallbackLanguages)), uintptr(unsafe.Pointer(pAttributes)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetThreadPreferredUILanguages2(flags uint32, languages PWSTR, numLanguagesSet *uint32, snapshot *HSAVEDUILANGUAGES) BOOL {
	addr := lazyAddr(&pSetThreadPreferredUILanguages2, libKernel32, "SetThreadPreferredUILanguages2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(flags), uintptr(unsafe.Pointer(languages)), uintptr(unsafe.Pointer(numLanguagesSet)), uintptr(unsafe.Pointer(snapshot)))
	return BOOL(ret)
}

func RestoreThreadPreferredUILanguages(snapshot HSAVEDUILANGUAGES) {
	addr := lazyAddr(&pRestoreThreadPreferredUILanguages, libKernel32, "RestoreThreadPreferredUILanguages")
	syscall.SyscallN(addr, snapshot)
}

func NotifyUILanguageChange(dwFlags uint32, pcwstrNewLanguage PWSTR, pcwstrPreviousLanguage PWSTR, dwReserved uint32, pdwStatusRtrn *uint32) BOOL {
	addr := lazyAddr(&pNotifyUILanguageChange, libKernel32, "NotifyUILanguageChange")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(unsafe.Pointer(pcwstrNewLanguage)), uintptr(unsafe.Pointer(pcwstrPreviousLanguage)), uintptr(dwReserved), uintptr(unsafe.Pointer(pdwStatusRtrn)))
	return BOOL(ret)
}

func GetStringTypeExA(Locale uint32, dwInfoType uint32, lpSrcStr PSTR, cchSrc int32, lpCharType *uint16) BOOL {
	addr := lazyAddr(&pGetStringTypeExA, libKernel32, "GetStringTypeExA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwInfoType), uintptr(unsafe.Pointer(lpSrcStr)), uintptr(cchSrc), uintptr(unsafe.Pointer(lpCharType)))
	return BOOL(ret)
}

func GetStringTypeA(Locale uint32, dwInfoType uint32, lpSrcStr PSTR, cchSrc int32, lpCharType *uint16) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetStringTypeA, libKernel32, "GetStringTypeA")
	ret, _, err := syscall.SyscallN(addr, uintptr(Locale), uintptr(dwInfoType), uintptr(unsafe.Pointer(lpSrcStr)), uintptr(cchSrc), uintptr(unsafe.Pointer(lpCharType)))
	return BOOL(ret), WIN32_ERROR(err)
}

func FoldStringA(dwMapFlags FOLD_STRING_MAP_FLAGS, lpSrcStr PSTR, cchSrc int32, lpDestStr PSTR, cchDest int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pFoldStringA, libKernel32, "FoldStringA")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwMapFlags), uintptr(unsafe.Pointer(lpSrcStr)), uintptr(cchSrc), uintptr(unsafe.Pointer(lpDestStr)), uintptr(cchDest))
	return int32(ret), WIN32_ERROR(err)
}

func EnumSystemLocalesA(lpLocaleEnumProc LOCALE_ENUMPROCA, dwFlags uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumSystemLocalesA, libKernel32, "EnumSystemLocalesA")
	ret, _, err := syscall.SyscallN(addr, lpLocaleEnumProc, uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

var EnumSystemLocales = EnumSystemLocalesW

func EnumSystemLocalesW(lpLocaleEnumProc LOCALE_ENUMPROCW, dwFlags uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumSystemLocalesW, libKernel32, "EnumSystemLocalesW")
	ret, _, err := syscall.SyscallN(addr, lpLocaleEnumProc, uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumSystemLanguageGroupsA(lpLanguageGroupEnumProc LANGUAGEGROUP_ENUMPROCA, dwFlags ENUM_SYSTEM_LANGUAGE_GROUPS_FLAGS, lParam uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumSystemLanguageGroupsA, libKernel32, "EnumSystemLanguageGroupsA")
	ret, _, err := syscall.SyscallN(addr, lpLanguageGroupEnumProc, uintptr(dwFlags), lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

var EnumSystemLanguageGroups = EnumSystemLanguageGroupsW

func EnumSystemLanguageGroupsW(lpLanguageGroupEnumProc LANGUAGEGROUP_ENUMPROCW, dwFlags ENUM_SYSTEM_LANGUAGE_GROUPS_FLAGS, lParam uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumSystemLanguageGroupsW, libKernel32, "EnumSystemLanguageGroupsW")
	ret, _, err := syscall.SyscallN(addr, lpLanguageGroupEnumProc, uintptr(dwFlags), lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumLanguageGroupLocalesA(lpLangGroupLocaleEnumProc LANGGROUPLOCALE_ENUMPROCA, LanguageGroup uint32, dwFlags uint32, lParam uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumLanguageGroupLocalesA, libKernel32, "EnumLanguageGroupLocalesA")
	ret, _, err := syscall.SyscallN(addr, lpLangGroupLocaleEnumProc, uintptr(LanguageGroup), uintptr(dwFlags), lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

var EnumLanguageGroupLocales = EnumLanguageGroupLocalesW

func EnumLanguageGroupLocalesW(lpLangGroupLocaleEnumProc LANGGROUPLOCALE_ENUMPROCW, LanguageGroup uint32, dwFlags uint32, lParam uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumLanguageGroupLocalesW, libKernel32, "EnumLanguageGroupLocalesW")
	ret, _, err := syscall.SyscallN(addr, lpLangGroupLocaleEnumProc, uintptr(LanguageGroup), uintptr(dwFlags), lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumUILanguagesA(lpUILanguageEnumProc UILANGUAGE_ENUMPROCA, dwFlags uint32, lParam uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumUILanguagesA, libKernel32, "EnumUILanguagesA")
	ret, _, err := syscall.SyscallN(addr, lpUILanguageEnumProc, uintptr(dwFlags), lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

var EnumUILanguages = EnumUILanguagesW

func EnumUILanguagesW(lpUILanguageEnumProc UILANGUAGE_ENUMPROCW, dwFlags uint32, lParam uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumUILanguagesW, libKernel32, "EnumUILanguagesW")
	ret, _, err := syscall.SyscallN(addr, lpUILanguageEnumProc, uintptr(dwFlags), lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumSystemCodePagesA(lpCodePageEnumProc CODEPAGE_ENUMPROCA, dwFlags ENUM_SYSTEM_CODE_PAGES_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumSystemCodePagesA, libKernel32, "EnumSystemCodePagesA")
	ret, _, err := syscall.SyscallN(addr, lpCodePageEnumProc, uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

var EnumSystemCodePages = EnumSystemCodePagesW

func EnumSystemCodePagesW(lpCodePageEnumProc CODEPAGE_ENUMPROCW, dwFlags ENUM_SYSTEM_CODE_PAGES_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumSystemCodePagesW, libKernel32, "EnumSystemCodePagesW")
	ret, _, err := syscall.SyscallN(addr, lpCodePageEnumProc, uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func IdnToNameprepUnicode(dwFlags uint32, lpUnicodeCharStr PWSTR, cchUnicodeChar int32, lpNameprepCharStr PWSTR, cchNameprepChar int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pIdnToNameprepUnicode, libKernel32, "IdnToNameprepUnicode")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(unsafe.Pointer(lpUnicodeCharStr)), uintptr(cchUnicodeChar), uintptr(unsafe.Pointer(lpNameprepCharStr)), uintptr(cchNameprepChar))
	return int32(ret), WIN32_ERROR(err)
}

func NormalizeString(NormForm NORM_FORM, lpSrcString PWSTR, cwSrcLength int32, lpDstString PWSTR, cwDstLength int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pNormalizeString, libKernel32, "NormalizeString")
	ret, _, err := syscall.SyscallN(addr, uintptr(NormForm), uintptr(unsafe.Pointer(lpSrcString)), uintptr(cwSrcLength), uintptr(unsafe.Pointer(lpDstString)), uintptr(cwDstLength))
	return int32(ret), WIN32_ERROR(err)
}

func IsNormalizedString(NormForm NORM_FORM, lpString PWSTR, cwLength int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsNormalizedString, libKernel32, "IsNormalizedString")
	ret, _, err := syscall.SyscallN(addr, uintptr(NormForm), uintptr(unsafe.Pointer(lpString)), uintptr(cwLength))
	return BOOL(ret), WIN32_ERROR(err)
}

func VerifyScripts(dwFlags uint32, lpLocaleScripts PWSTR, cchLocaleScripts int32, lpTestScripts PWSTR, cchTestScripts int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pVerifyScripts, libKernel32, "VerifyScripts")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(unsafe.Pointer(lpLocaleScripts)), uintptr(cchLocaleScripts), uintptr(unsafe.Pointer(lpTestScripts)), uintptr(cchTestScripts))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetStringScripts(dwFlags uint32, lpString PWSTR, cchString int32, lpScripts PWSTR, cchScripts int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetStringScripts, libKernel32, "GetStringScripts")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(unsafe.Pointer(lpString)), uintptr(cchString), uintptr(unsafe.Pointer(lpScripts)), uintptr(cchScripts))
	return int32(ret), WIN32_ERROR(err)
}

func GetLocaleInfoEx(lpLocaleName PWSTR, LCType uint32, lpLCData PWSTR, cchData int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetLocaleInfoEx, libKernel32, "GetLocaleInfoEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(LCType), uintptr(unsafe.Pointer(lpLCData)), uintptr(cchData))
	return int32(ret), WIN32_ERROR(err)
}

func GetCalendarInfoEx(lpLocaleName PWSTR, Calendar uint32, lpReserved PWSTR, CalType uint32, lpCalData PWSTR, cchData int32, lpValue *uint32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetCalendarInfoEx, libKernel32, "GetCalendarInfoEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(Calendar), uintptr(unsafe.Pointer(lpReserved)), uintptr(CalType), uintptr(unsafe.Pointer(lpCalData)), uintptr(cchData), uintptr(unsafe.Pointer(lpValue)))
	return int32(ret), WIN32_ERROR(err)
}

func GetNumberFormatEx(lpLocaleName PWSTR, dwFlags uint32, lpValue PWSTR, lpFormat *NUMBERFMTW, lpNumberStr PWSTR, cchNumber int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetNumberFormatEx, libKernel32, "GetNumberFormatEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(dwFlags), uintptr(unsafe.Pointer(lpValue)), uintptr(unsafe.Pointer(lpFormat)), uintptr(unsafe.Pointer(lpNumberStr)), uintptr(cchNumber))
	return int32(ret), WIN32_ERROR(err)
}

func GetCurrencyFormatEx(lpLocaleName PWSTR, dwFlags uint32, lpValue PWSTR, lpFormat *CURRENCYFMTW, lpCurrencyStr PWSTR, cchCurrency int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetCurrencyFormatEx, libKernel32, "GetCurrencyFormatEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(dwFlags), uintptr(unsafe.Pointer(lpValue)), uintptr(unsafe.Pointer(lpFormat)), uintptr(unsafe.Pointer(lpCurrencyStr)), uintptr(cchCurrency))
	return int32(ret), WIN32_ERROR(err)
}

func GetUserDefaultLocaleName(lpLocaleName PWSTR, cchLocaleName int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetUserDefaultLocaleName, libKernel32, "GetUserDefaultLocaleName")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(cchLocaleName))
	return int32(ret), WIN32_ERROR(err)
}

func GetSystemDefaultLocaleName(lpLocaleName PWSTR, cchLocaleName int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetSystemDefaultLocaleName, libKernel32, "GetSystemDefaultLocaleName")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(cchLocaleName))
	return int32(ret), WIN32_ERROR(err)
}

func IsNLSDefinedString(Function uint32, dwFlags uint32, lpVersionInformation *NLSVERSIONINFO, lpString PWSTR, cchStr int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsNLSDefinedString, libKernel32, "IsNLSDefinedString")
	ret, _, err := syscall.SyscallN(addr, uintptr(Function), uintptr(dwFlags), uintptr(unsafe.Pointer(lpVersionInformation)), uintptr(unsafe.Pointer(lpString)), uintptr(cchStr))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNLSVersionEx(function uint32, lpLocaleName PWSTR, lpVersionInformation *NLSVERSIONINFOEX) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNLSVersionEx, libKernel32, "GetNLSVersionEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(function), uintptr(unsafe.Pointer(lpLocaleName)), uintptr(unsafe.Pointer(lpVersionInformation)))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsValidNLSVersion(function uint32, lpLocaleName PWSTR, lpVersionInformation *NLSVERSIONINFOEX) uint32 {
	addr := lazyAddr(&pIsValidNLSVersion, libKernel32, "IsValidNLSVersion")
	ret, _, _ := syscall.SyscallN(addr, uintptr(function), uintptr(unsafe.Pointer(lpLocaleName)), uintptr(unsafe.Pointer(lpVersionInformation)))
	return uint32(ret)
}

func FindNLSStringEx(lpLocaleName PWSTR, dwFindNLSStringFlags uint32, lpStringSource PWSTR, cchSource int32, lpStringValue PWSTR, cchValue int32, pcchFound *int32, lpVersionInformation *NLSVERSIONINFO, lpReserved unsafe.Pointer, sortHandle LPARAM) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pFindNLSStringEx, libKernel32, "FindNLSStringEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(dwFindNLSStringFlags), uintptr(unsafe.Pointer(lpStringSource)), uintptr(cchSource), uintptr(unsafe.Pointer(lpStringValue)), uintptr(cchValue), uintptr(unsafe.Pointer(pcchFound)), uintptr(unsafe.Pointer(lpVersionInformation)), uintptr(lpReserved), sortHandle)
	return int32(ret), WIN32_ERROR(err)
}

func LCMapStringEx(lpLocaleName PWSTR, dwMapFlags uint32, lpSrcStr PWSTR, cchSrc int32, lpDestStr PWSTR, cchDest int32, lpVersionInformation *NLSVERSIONINFO, lpReserved unsafe.Pointer, sortHandle LPARAM) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pLCMapStringEx, libKernel32, "LCMapStringEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(dwMapFlags), uintptr(unsafe.Pointer(lpSrcStr)), uintptr(cchSrc), uintptr(unsafe.Pointer(lpDestStr)), uintptr(cchDest), uintptr(unsafe.Pointer(lpVersionInformation)), uintptr(lpReserved), sortHandle)
	return int32(ret), WIN32_ERROR(err)
}

func IsValidLocaleName(lpLocaleName PWSTR) BOOL {
	addr := lazyAddr(&pIsValidLocaleName, libKernel32, "IsValidLocaleName")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpLocaleName)))
	return BOOL(ret)
}

func EnumCalendarInfoExEx(pCalInfoEnumProcExEx CALINFO_ENUMPROCEXEX, lpLocaleName PWSTR, Calendar uint32, lpReserved PWSTR, CalType uint32, lParam LPARAM) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumCalendarInfoExEx, libKernel32, "EnumCalendarInfoExEx")
	ret, _, err := syscall.SyscallN(addr, pCalInfoEnumProcExEx, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(Calendar), uintptr(unsafe.Pointer(lpReserved)), uintptr(CalType), lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumDateFormatsExEx(lpDateFmtEnumProcExEx DATEFMT_ENUMPROCEXEX, lpLocaleName PWSTR, dwFlags ENUM_DATE_FORMATS_FLAGS, lParam LPARAM) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumDateFormatsExEx, libKernel32, "EnumDateFormatsExEx")
	ret, _, err := syscall.SyscallN(addr, lpDateFmtEnumProcExEx, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(dwFlags), lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumTimeFormatsEx(lpTimeFmtEnumProcEx TIMEFMT_ENUMPROCEX, lpLocaleName PWSTR, dwFlags uint32, lParam LPARAM) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumTimeFormatsEx, libKernel32, "EnumTimeFormatsEx")
	ret, _, err := syscall.SyscallN(addr, lpTimeFmtEnumProcEx, uintptr(unsafe.Pointer(lpLocaleName)), uintptr(dwFlags), lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumSystemLocalesEx(lpLocaleEnumProcEx LOCALE_ENUMPROCEX, dwFlags uint32, lParam LPARAM, lpReserved unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumSystemLocalesEx, libKernel32, "EnumSystemLocalesEx")
	ret, _, err := syscall.SyscallN(addr, lpLocaleEnumProcEx, uintptr(dwFlags), lParam, uintptr(lpReserved))
	return BOOL(ret), WIN32_ERROR(err)
}

func ResolveLocaleName(lpNameToResolve PWSTR, lpLocaleName PWSTR, cchLocaleName int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pResolveLocaleName, libKernel32, "ResolveLocaleName")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpNameToResolve)), uintptr(unsafe.Pointer(lpLocaleName)), uintptr(cchLocaleName))
	return int32(ret), WIN32_ERROR(err)
}

func FindStringOrdinal(dwFindStringOrdinalFlags uint32, lpStringSource PWSTR, cchSource int32, lpStringValue PWSTR, cchValue int32, bIgnoreCase BOOL) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pFindStringOrdinal, libKernel32, "FindStringOrdinal")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFindStringOrdinalFlags), uintptr(unsafe.Pointer(lpStringSource)), uintptr(cchSource), uintptr(unsafe.Pointer(lpStringValue)), uintptr(cchValue), uintptr(bIgnoreCase))
	return int32(ret), WIN32_ERROR(err)
}

func LstrcmpA(lpString1 PSTR, lpString2 PSTR) int32 {
	addr := lazyAddr(&pLstrcmpA, libKernel32, "lstrcmpA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString1)), uintptr(unsafe.Pointer(lpString2)))
	return int32(ret)
}

var Lstrcmp = LstrcmpW

func LstrcmpW(lpString1 PWSTR, lpString2 PWSTR) int32 {
	addr := lazyAddr(&pLstrcmpW, libKernel32, "lstrcmpW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString1)), uintptr(unsafe.Pointer(lpString2)))
	return int32(ret)
}

func LstrcmpiA(lpString1 PSTR, lpString2 PSTR) int32 {
	addr := lazyAddr(&pLstrcmpiA, libKernel32, "lstrcmpiA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString1)), uintptr(unsafe.Pointer(lpString2)))
	return int32(ret)
}

var Lstrcmpi = LstrcmpiW

func LstrcmpiW(lpString1 PWSTR, lpString2 PWSTR) int32 {
	addr := lazyAddr(&pLstrcmpiW, libKernel32, "lstrcmpiW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString1)), uintptr(unsafe.Pointer(lpString2)))
	return int32(ret)
}

func LstrcpynA(lpString1 PSTR, lpString2 PSTR, iMaxLength int32) PSTR {
	addr := lazyAddr(&pLstrcpynA, libKernel32, "lstrcpynA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString1)), uintptr(unsafe.Pointer(lpString2)), uintptr(iMaxLength))
	return (PSTR)(unsafe.Pointer(ret))
}

var Lstrcpyn = LstrcpynW

func LstrcpynW(lpString1 PWSTR, lpString2 PWSTR, iMaxLength int32) PWSTR {
	addr := lazyAddr(&pLstrcpynW, libKernel32, "lstrcpynW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString1)), uintptr(unsafe.Pointer(lpString2)), uintptr(iMaxLength))
	return (PWSTR)(unsafe.Pointer(ret))
}

func LstrcpyA(lpString1 PSTR, lpString2 PSTR) PSTR {
	addr := lazyAddr(&pLstrcpyA, libKernel32, "lstrcpyA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString1)), uintptr(unsafe.Pointer(lpString2)))
	return (PSTR)(unsafe.Pointer(ret))
}

var Lstrcpy = LstrcpyW

func LstrcpyW(lpString1 PWSTR, lpString2 PWSTR) PWSTR {
	addr := lazyAddr(&pLstrcpyW, libKernel32, "lstrcpyW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString1)), uintptr(unsafe.Pointer(lpString2)))
	return (PWSTR)(unsafe.Pointer(ret))
}

func LstrcatA(lpString1 PSTR, lpString2 PSTR) PSTR {
	addr := lazyAddr(&pLstrcatA, libKernel32, "lstrcatA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString1)), uintptr(unsafe.Pointer(lpString2)))
	return (PSTR)(unsafe.Pointer(ret))
}

var Lstrcat = LstrcatW

func LstrcatW(lpString1 PWSTR, lpString2 PWSTR) PWSTR {
	addr := lazyAddr(&pLstrcatW, libKernel32, "lstrcatW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString1)), uintptr(unsafe.Pointer(lpString2)))
	return (PWSTR)(unsafe.Pointer(ret))
}

func LstrlenA(lpString PSTR) int32 {
	addr := lazyAddr(&pLstrlenA, libKernel32, "lstrlenA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString)))
	return int32(ret)
}

var Lstrlen = LstrlenW

func LstrlenW(lpString PWSTR) int32 {
	addr := lazyAddr(&pLstrlenW, libKernel32, "lstrlenW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString)))
	return int32(ret)
}

func IsTextUnicode(lpv unsafe.Pointer, iSize int32, lpiResult *IS_TEXT_UNICODE_RESULT) BOOL {
	addr := lazyAddr(&pIsTextUnicode, libAdvapi32, "IsTextUnicode")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lpv), uintptr(iSize), uintptr(unsafe.Pointer(lpiResult)))
	return BOOL(ret)
}
