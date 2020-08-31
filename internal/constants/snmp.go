package constants

const (
	DocsIfDocsisBaseCapability       = ".1.3.6.1.2.1.10.127.1.1.5.0"
	SysDescr                         = ".1.3.6.1.2.1.1.1.0"
	SysUpTimeInstance                = ".1.3.6.1.2.1.1.3.0"
	DocsIf31CmDocsisBaseCapability   = ".1.3.6.1.4.1.4491.2.1.28.1.1.0"
	DocsVers10                       = 1
	DocsVer11                        = 2
	DocsVer20                        = 3
	DocsVer30                        = 4
	DocsVer31                        = 5
	IfType                           = ".1.3.6.1.2.1.2.2.1.3"
	IfInOctets                       = ".1.3.6.1.2.1.2.2.1.10"
	IfOutOctets                      = ".1.3.6.1.2.1.2.2.1.16"
	IfHCInOctets                     = ".1.3.6.1.2.1.31.1.1.1.6"
	IfHCOutOctets                    = ".1.3.6.1.2.1.31.1.1.1.10"
	DocsIf3CmStatusUsTxPower         = ".1.3.6.1.4.1.4491.2.1.20.1.2.1.1"
	DocsIfUpChannelId                = ".1.3.6.1.2.1.10.127.1.1.2.1.1"
	DocsIfUpChannelFrequency         = ".1.3.6.1.2.1.10.127.1.1.2.1.2"
	DocsIfUpChannelWidth             = ".1.3.6.1.2.1.10.127.1.1.2.1.3"
	DocsIfUpChannelModulationProfile = ".1.3.6.1.2.1.10.127.1.1.2.1.4"
	DocsIfUpChannelTxTimingOffset    = ".1.3.6.1.2.1.10.127.1.1.2.1.6"
	DocsIfCmStatusTxPower            = ".1.3.6.1.2.1.10.127.1.2.2.1.3"

	DocsIf31CmDsOfdmChannelPowerCenterFrequency        = ".1.3.6.1.4.1.4491.2.1.28.1.11.1.2"
	DocsIf31CmDsOfdmChannelPowerRxPower                = ".1.3.6.1.4.1.4491.2.1.28.1.11.1.3"
	DocsIf31CmDsOfdmProfileStatsConfigChangeCt         = ".1.3.6.1.4.1.4491.2.1.28.1.10.1.2"
	DocsIf31CmDsOfdmProfileStatsTotalCodewords         = ".1.3.6.1.4.1.4491.2.1.28.1.10.1.3"
	DocsIf31CmDsOfdmProfileStatsCorrectedCodewords     = ".1.3.6.1.4.1.4491.2.1.28.1.10.1.4"
	DocsIf31CmDsOfdmProfileStatsUncorrectableCodewords = ".1.3.6.1.4.1.4491.2.1.28.1.10.1.5"
	DocsIf3CmtsCmUsStatusRxPower                       = ".1.3.6.1.4.1.4491.2.1.20.1.4.1.3"
	DocsIf3CmtsCmUsStatusSignalNoise                   = ".1.3.6.1.4.1.4491.2.1.20.1.4.1.4"
	DocsIf3CmtsCmUsStatusMicroreflections              = ".1.3.6.1.4.1.4491.2.1.20.1.4.1.5"
	DocsIf3CmtsCmUsStatusUnerroreds                    = ".1.3.6.1.4.1.4491.2.1.20.1.4.1.7"
	DocsIf3CmtsCmUsStatusCorrecteds                    = ".1.3.6.1.4.1.4491.2.1.20.1.4.1.8"
	DocsIf3CmtsCmUsStatusUncorrectables                = ".1.3.6.1.4.1.4491.2.1.20.1.4.1.9"

	DocsPnmBulkDestIpAddr = ".1.3.6.1.4.1.4491.2.1.27.1.1.1.2"
	DocsPnmBulkDestIpAddrType = ".1.3.6.1.4.1.4491.2.1.27.1.1.1.1"
	// add OFDM downstream interface index to the end of the oid
	DocsPnmCmDsOfdmRxMerFileName = ".1.3.6.1.4.1.4491.2.1.27.1.2.5.1.8"
	// add OFDM downstream interface index to the end of the oid
	DocsPnmCmDsOfdmRxMerFileEnable = ".1.3.6.1.4.1.4491.2.1.27.1.2.5.1.1"
	AddrTypeIpv4 = 1
	AddrTypeIpv6 = 2

	IntfTypeOfdmDownstream = 277
)

const (
	CmStatusOther                     = 1
	CmStatusRanging                   = 2
	CmStatusRangingAborted            = 3
	CmStatusRangingComplete           = 4
	CmStatusIpComplete                = 5
	CmStatusRegistrationComplete      = 6
	CmStatusAccessDenied              = 7
	CmStatusOperational               = 8
	CmStatusRegisteredBPIInitializing = 9
)
