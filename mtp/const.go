package mtp

// DO NOT EDIT : generated automatically

const AC_ReadWrite = 0x0000
const AC_ReadOnly = 0x0001
const AC_ReadOnly_with_Object_Deletion = 0x0002

var AC_names = map[int]string{0x0000: "ReadWrite",
	0x0001: "ReadOnly",
	0x0002: "ReadOnly_with_Object_Deletion",
}

const AT_Undefined = 0x0000
const AT_GenericFolder = 0x0001
const AT_Album = 0x0002
const AT_TimeSequence = 0x0003
const AT_HorizontalPanoramic = 0x0004
const AT_VerticalPanoramic = 0x0005
const AT_2DPanoramic = 0x0006
const AT_AncillaryData = 0x0007

var AT_names = map[int]string{0x0000: "Undefined",
	0x0001: "GenericFolder",
	0x0002: "Album",
	0x0003: "TimeSequence",
	0x0004: "HorizontalPanoramic",
	0x0005: "VerticalPanoramic",
	0x0006: "2DPanoramic",
	0x0007: "AncillaryData",
}

const DL_LE = 0x0F
const DL_BE = 0xF0

var DL_names = map[int]string{0x0F: "LE",
	0xF0: "BE",
}

// device property code
const DPC_Undefined = 0x5000
const DPC_BatteryLevel = 0x5001
const DPC_FunctionalMode = 0x5002
const DPC_ImageSize = 0x5003
const DPC_CompressionSetting = 0x5004
const DPC_WhiteBalance = 0x5005
const DPC_RGBGain = 0x5006
const DPC_FNumber = 0x5007
const DPC_FocalLength = 0x5008
const DPC_FocusDistance = 0x5009
const DPC_FocusMode = 0x500A
const DPC_ExposureMeteringMode = 0x500B
const DPC_FlashMode = 0x500C
const DPC_ExposureTime = 0x500D
const DPC_ExposureProgramMode = 0x500E
const DPC_ExposureIndex = 0x500F
const DPC_ExposureBiasCompensation = 0x5010
const DPC_DateTime = 0x5011
const DPC_CaptureDelay = 0x5012
const DPC_StillCaptureMode = 0x5013
const DPC_Contrast = 0x5014
const DPC_Sharpness = 0x5015
const DPC_DigitalZoom = 0x5016
const DPC_EffectMode = 0x5017
const DPC_BurstNumber = 0x5018
const DPC_BurstInterval = 0x5019
const DPC_TimelapseNumber = 0x501A
const DPC_TimelapseInterval = 0x501B
const DPC_FocusMeteringMode = 0x501C
const DPC_UploadURL = 0x501D
const DPC_Artist = 0x501E
const DPC_CopyrightInfo = 0x501F
const DPC_SupportedStreams = 0x5020
const DPC_EnabledStreams = 0x5021
const DPC_VideoFormat = 0x5022
const DPC_VideoResolution = 0x5023
const DPC_VideoQuality = 0x5024
const DPC_VideoFrameRate = 0x5025
const DPC_VideoContrast = 0x5026
const DPC_VideoBrightness = 0x5027
const DPC_AudioFormat = 0x5028
const DPC_AudioBitrate = 0x5029
const DPC_AudioSamplingRate = 0x502A
const DPC_AudioBitPerSample = 0x502B
const DPC_AudioVolume = 0x502C
const DPC_EXTENSION = 0xD000
const DPC_CANON_BeepMode = 0xD001
const DPC_EK_ColorTemperature = 0xD001
const DPC_CANON_BatteryKind = 0xD002
const DPC_EK_DateTimeStampFormat = 0xD002
const DPC_CANON_BatteryStatus = 0xD003
const DPC_EK_BeepMode = 0xD003
const DPC_CANON_UILockType = 0xD004
const DPC_CASIO_UNKNOWN_1 = 0xD004
const DPC_EK_VideoOut = 0xD004
const DPC_CANON_CameraMode = 0xD005
const DPC_CASIO_UNKNOWN_2 = 0xD005
const DPC_EK_PowerSaving = 0xD005
const DPC_CANON_ImageQuality = 0xD006
const DPC_EK_UI_Language = 0xD006
const DPC_CANON_FullViewFileFormat = 0xD007
const DPC_CASIO_UNKNOWN_3 = 0xD007
const DPC_CANON_ImageSize = 0xD008
const DPC_CASIO_RECORD_LIGHT = 0xD008
const DPC_CANON_SelfTime = 0xD009
const DPC_CASIO_UNKNOWN_4 = 0xD009
const DPC_CANON_FlashMode = 0xD00A
const DPC_CASIO_UNKNOWN_5 = 0xD00A
const DPC_CANON_Beep = 0xD00B
const DPC_CASIO_MOVIE_MODE = 0xD00B
const DPC_CANON_ShootingMode = 0xD00C
const DPC_CASIO_HD_SETTING = 0xD00C
const DPC_CANON_ImageMode = 0xD00D
const DPC_CASIO_HS_SETTING = 0xD00D
const DPC_CANON_DriveMode = 0xD00E
const DPC_CANON_EZoom = 0xD00F
const DPC_CASIO_CS_HIGH_SPEED = 0xD00F
const DPC_CANON_MeteringMode = 0xD010
const DPC_CASIO_CS_UPPER_LIMIT = 0xD010
const DPC_NIKON_ShootingBank = 0xD010
const DPC_CANON_AFDistance = 0xD011
const DPC_CASIO_CS_SHOT = 0xD011
const DPC_NIKON_ShootingBankNameA = 0xD011
const DPC_CANON_FocusingPoint = 0xD012
const DPC_CASIO_UNKNOWN_6 = 0xD012
const DPC_NIKON_ShootingBankNameB = 0xD012
const DPC_CANON_WhiteBalance = 0xD013
const DPC_CASIO_UNKNOWN_7 = 0xD013
const DPC_NIKON_ShootingBankNameC = 0xD013
const DPC_CANON_SlowShutterSetting = 0xD014
const DPC_NIKON_ShootingBankNameD = 0xD014
const DPC_CANON_AFMode = 0xD015
const DPC_CASIO_UNKNOWN_8 = 0xD015
const DPC_NIKON_ResetBank0 = 0xD015
const DPC_CANON_ImageStabilization = 0xD016
const DPC_NIKON_RawCompression = 0xD016
const DPC_CANON_Contrast = 0xD017
const DPC_CASIO_UNKNOWN_9 = 0xD017
const DPC_FUJI_ColorTemperature = 0xD017
const DPC_NIKON_WhiteBalanceAutoBias = 0xD017
const DPC_CANON_ColorGain = 0xD018
const DPC_CASIO_UNKNOWN_10 = 0xD018
const DPC_FUJI_Quality = 0xD018
const DPC_NIKON_WhiteBalanceTungstenBias = 0xD018
const DPC_CANON_Sharpness = 0xD019
const DPC_CASIO_UNKNOWN_11 = 0xD019
const DPC_NIKON_WhiteBalanceFluorescentBias = 0xD019
const DPC_CANON_Sensitivity = 0xD01A
const DPC_CASIO_UNKNOWN_12 = 0xD01A
const DPC_NIKON_WhiteBalanceDaylightBias = 0xD01A
const DPC_CANON_ParameterSet = 0xD01B
const DPC_CASIO_UNKNOWN_13 = 0xD01B
const DPC_NIKON_WhiteBalanceFlashBias = 0xD01B
const DPC_CANON_ISOSpeed = 0xD01C
const DPC_CASIO_UNKNOWN_14 = 0xD01C
const DPC_NIKON_WhiteBalanceCloudyBias = 0xD01C
const DPC_CANON_Aperture = 0xD01D
const DPC_CASIO_UNKNOWN_15 = 0xD01D
const DPC_NIKON_WhiteBalanceShadeBias = 0xD01D
const DPC_CANON_ShutterSpeed = 0xD01E
const DPC_NIKON_WhiteBalanceColorTemperature = 0xD01E
const DPC_CANON_ExpCompensation = 0xD01F
const DPC_NIKON_WhiteBalancePresetNo = 0xD01F
const DPC_CANON_FlashCompensation = 0xD020
const DPC_CASIO_UNKNOWN_16 = 0xD020
const DPC_NIKON_WhiteBalancePresetName0 = 0xD020
const DPC_CANON_AEBExposureCompensation = 0xD021
const DPC_NIKON_WhiteBalancePresetName1 = 0xD021
const DPC_NIKON_WhiteBalancePresetName2 = 0xD022
const DPC_CANON_AvOpen = 0xD023
const DPC_NIKON_WhiteBalancePresetName3 = 0xD023
const DPC_CANON_AvMax = 0xD024
const DPC_NIKON_WhiteBalancePresetName4 = 0xD024
const DPC_CANON_FocalLength = 0xD025
const DPC_NIKON_WhiteBalancePresetVal0 = 0xD025
const DPC_CANON_FocalLengthTele = 0xD026
const DPC_NIKON_WhiteBalancePresetVal1 = 0xD026
const DPC_CANON_FocalLengthWide = 0xD027
const DPC_NIKON_WhiteBalancePresetVal2 = 0xD027
const DPC_CANON_FocalLengthDenominator = 0xD028
const DPC_NIKON_WhiteBalancePresetVal3 = 0xD028
const DPC_CANON_CaptureTransferMode = 0xD029
const DPC_NIKON_WhiteBalancePresetVal4 = 0xD029
const DPC_CANON_Zoom = 0xD02A
const DPC_NIKON_ImageSharpening = 0xD02A
const DPC_CANON_NamePrefix = 0xD02B
const DPC_NIKON_ToneCompensation = 0xD02B
const DPC_CANON_SizeQualityMode = 0xD02C
const DPC_NIKON_ColorModel = 0xD02C
const DPC_CANON_SupportedThumbSize = 0xD02D
const DPC_NIKON_HueAdjustment = 0xD02D
const DPC_CANON_SizeOfOutputDataFromCamera = 0xD02E
const DPC_CANON_SizeOfInputDataToCamera = 0xD02F
const DPC_CANON_RemoteAPIVersion = 0xD030
const DPC_CASIO_UNKNOWN_17 = 0xD030
const DPC_NIKON_ShootingMode = 0xD030
const DPC_CANON_FirmwareVersion = 0xD031
const DPC_NIKON_JPEG_Compression_Policy = 0xD031
const DPC_CANON_CameraModel = 0xD032
const DPC_NIKON_ColorSpace = 0xD032
const DPC_CANON_CameraOwner = 0xD033
const DPC_NIKON_AutoDXCrop = 0xD033
const DPC_CANON_UnixTime = 0xD034
const DPC_CANON_CameraBodyID = 0xD035
const DPC_CANON_CameraOutput = 0xD036
const DPC_NIKON_VideoMode = 0xD036
const DPC_CANON_DispAv = 0xD037
const DPC_NIKON_EffectMode = 0xD037
const DPC_CANON_AvOpenApex = 0xD038
const DPC_CANON_DZoomMagnification = 0xD039
const DPC_CANON_MlSpotPos = 0xD03A
const DPC_CANON_DispAvMax = 0xD03B
const DPC_CANON_AvMaxApex = 0xD03C
const DPC_CANON_EZoomStartPosition = 0xD03D
const DPC_CANON_FocalLengthOfTele = 0xD03E
const DPC_CANON_EZoomSizeOfTele = 0xD03F
const DPC_CANON_PhotoEffect = 0xD040
const DPC_NIKON_CSMMenuBankSelect = 0xD040
const DPC_CANON_AssistLight = 0xD041
const DPC_NIKON_MenuBankNameA = 0xD041
const DPC_CANON_FlashQuantityCount = 0xD042
const DPC_NIKON_MenuBankNameB = 0xD042
const DPC_CANON_RotationAngle = 0xD043
const DPC_NIKON_MenuBankNameC = 0xD043
const DPC_CANON_RotationScene = 0xD044
const DPC_NIKON_MenuBankNameD = 0xD044
const DPC_CANON_EventEmulateMode = 0xD045
const DPC_NIKON_ResetBank = 0xD045
const DPC_CANON_DPOFVersion = 0xD046
const DPC_CANON_TypeOfSupportedSlideShow = 0xD047
const DPC_CANON_AverageFilesizes = 0xD048
const DPC_NIKON_A1AFCModePriority = 0xD048
const DPC_CANON_ModelID = 0xD049
const DPC_NIKON_A2AFSModePriority = 0xD049
const DPC_NIKON_A3GroupDynamicAF = 0xD04A
const DPC_NIKON_A4AFActivation = 0xD04B
const DPC_NIKON_FocusAreaIllumManualFocus = 0xD04C
const DPC_NIKON_FocusAreaIllumContinuous = 0xD04D
const DPC_NIKON_FocusAreaIllumWhenSelected = 0xD04E
const DPC_NIKON_VerticalAFON = 0xD050
const DPC_NIKON_AFLockOn = 0xD051
const DPC_NIKON_FocusAreaZone = 0xD052
const DPC_NIKON_EnableCopyright = 0xD053
const DPC_NIKON_ISOAuto = 0xD054
const DPC_NIKON_EVISOStep = 0xD055
const DPC_NIKON_EVStepExposureComp = 0xD057
const DPC_NIKON_ExposureCompensation = 0xD058
const DPC_NIKON_CenterWeightArea = 0xD059
const DPC_NIKON_ExposureBaseMatrix = 0xD05A
const DPC_NIKON_ExposureBaseCenter = 0xD05B
const DPC_NIKON_ExposureBaseSpot = 0xD05C
const DPC_NIKON_LiveViewAFArea = 0xD05D
const DPC_NIKON_AELockMode = 0xD05E
const DPC_NIKON_AELAFLMode = 0xD05F
const DPC_NIKON_LiveViewAFFocus = 0xD061
const DPC_NIKON_MeterOff = 0xD062
const DPC_NIKON_SelfTimer = 0xD063
const DPC_NIKON_MonitorOff = 0xD064
const DPC_NIKON_ImgConfTime = 0xD065
const DPC_NIKON_AutoOffTimers = 0xD066
const DPC_NIKON_AngleLevel = 0xD067
const DPC_NIKON_D2MaximumShots = 0xD069
const DPC_NIKON_ExposureDelayMode = 0xD06A
const DPC_NIKON_LongExposureNoiseReduction = 0xD06B
const DPC_NIKON_FileNumberSequence = 0xD06C
const DPC_NIKON_ControlPanelFinderRearControl = 0xD06D
const DPC_NIKON_ControlPanelFinderViewfinder = 0xD06E
const DPC_NIKON_D7Illumination = 0xD06F
const DPC_NIKON_NrHighISO = 0xD070
const DPC_NIKON_SHSET_CH_GUID_DISP = 0xD071
const DPC_NIKON_ArtistName = 0xD072
const DPC_NIKON_CopyrightInfo = 0xD073
const DPC_NIKON_FlashSyncSpeed = 0xD074
const DPC_NIKON_E3AAFlashMode = 0xD076
const DPC_NIKON_E4ModelingFlash = 0xD077
const DPC_NIKON_BracketOrder = 0xD07A
const DPC_NIKON_BracketingSet = 0xD07C
const DPC_CASIO_UNKNOWN_18 = 0xD080
const DPC_NIKON_F1CenterButtonShootingMode = 0xD080
const DPC_NIKON_CenterButtonPlaybackMode = 0xD081
const DPC_NIKON_F2Multiselector = 0xD082
const DPC_NIKON_CenterButtonZoomRatio = 0xD08B
const DPC_NIKON_FunctionButton2 = 0xD08C
const DPC_NIKON_AFAreaPoint = 0xD08D
const DPC_NIKON_NormalAFOn = 0xD08E
const DPC_NIKON_CleanImageSensor = 0xD08F
const DPC_NIKON_ImageCommentString = 0xD090
const DPC_NIKON_ImageCommentEnable = 0xD091
const DPC_NIKON_ImageRotation = 0xD092
const DPC_NIKON_ManualSetLensNo = 0xD093
const DPC_NIKON_MovScreenSize = 0xD0A0
const DPC_NIKON_MovVoice = 0xD0A1
const DPC_NIKON_MovMicrophone = 0xD0A2
const DPC_NIKON_Bracketing = 0xD0C0
const DPC_NIKON_AutoExposureBracketStep = 0xD0C1
const DPC_NIKON_AutoExposureBracketProgram = 0xD0C2
const DPC_NIKON_AutoExposureBracketCount = 0xD0C3
const DPC_NIKON_WhiteBalanceBracketStep = 0xD0C4
const DPC_NIKON_WhiteBalanceBracketProgram = 0xD0C5
const DPC_NIKON_LensID = 0xD0E0
const DPC_NIKON_LensSort = 0xD0E1
const DPC_NIKON_LensType = 0xD0E2
const DPC_NIKON_FocalLengthMin = 0xD0E3
const DPC_NIKON_FocalLengthMax = 0xD0E4
const DPC_NIKON_MaxApAtMinFocalLength = 0xD0E5
const DPC_NIKON_MaxApAtMaxFocalLength = 0xD0E6
const DPC_NIKON_FinderISODisp = 0xD0F0
const DPC_NIKON_AutoOffPhoto = 0xD0F2
const DPC_NIKON_AutoOffMenu = 0xD0F3
const DPC_NIKON_AutoOffInfo = 0xD0F4
const DPC_NIKON_SelfTimerShootNum = 0xD0F5
const DPC_NIKON_VignetteCtrl = 0xD0F7
const DPC_NIKON_AutoDistortionControl = 0xD0F8
const DPC_NIKON_SceneMode = 0xD0F9
const DPC_CANON_EOS_Aperture = 0xD101
const DPC_MTP_SecureTime = 0xD101
const DPC_NIKON_ACPower = 0xD101
const DPC_CANON_EOS_ShutterSpeed = 0xD102
const DPC_MTP_DeviceCertificate = 0xD102
const DPC_NIKON_WarningStatus = 0xD102
const DPC_OLYMPUS_ResolutionMode = 0xD102
const DPC_CANON_EOS_ISOSpeed = 0xD103
const DPC_MTP_RevocationInfo = 0xD103
const DPC_OLYMPUS_FocusPriority = 0xD103
const DPC_CANON_EOS_ExpCompensation = 0xD104
const DPC_NIKON_AFLockStatus = 0xD104
const DPC_OLYMPUS_DriveMode = 0xD104
const DPC_CANON_EOS_AutoExposureMode = 0xD105
const DPC_NIKON_AELockStatus = 0xD105
const DPC_OLYMPUS_DateTimeFormat = 0xD105
const DPC_CANON_EOS_DriveMode = 0xD106
const DPC_NIKON_FVLockStatus = 0xD106
const DPC_OLYMPUS_ExposureBiasStep = 0xD106
const DPC_NIKON_AutofocusLCDTopMode2 = 0xD107
const DPC_OLYMPUS_WBMode = 0xD107
const DPC_CANON_EOS_FocusMode = 0xD108
const DPC_NIKON_AutofocusArea = 0xD108
const DPC_OLYMPUS_OneTouchWB = 0xD108
const DPC_CANON_EOS_WhiteBalance = 0xD109
const DPC_NIKON_FlexibleProgram = 0xD109
const DPC_OLYMPUS_ManualWB = 0xD109
const DPC_CANON_EOS_ColorTemperature = 0xD10A
const DPC_OLYMPUS_ManualWBRBBias = 0xD10A
const DPC_CANON_EOS_WhiteBalanceAdjustA = 0xD10B
const DPC_OLYMPUS_CustomWB = 0xD10B
const DPC_CANON_EOS_WhiteBalanceAdjustB = 0xD10C
const DPC_NIKON_USBSpeed = 0xD10C
const DPC_OLYMPUS_CustomWBValue = 0xD10C
const DPC_CANON_EOS_WhiteBalanceXA = 0xD10D
const DPC_NIKON_CCDNumber = 0xD10D
const DPC_OLYMPUS_ExposureTimeEx = 0xD10D
const DPC_CANON_EOS_WhiteBalanceXB = 0xD10E
const DPC_NIKON_CameraOrientation = 0xD10E
const DPC_OLYMPUS_BulbMode = 0xD10E
const DPC_CANON_EOS_ColorSpace = 0xD10F
const DPC_NIKON_GroupPtnType = 0xD10F
const DPC_OLYMPUS_AntiMirrorMode = 0xD10F
const DPC_CANON_EOS_PictureStyle = 0xD110
const DPC_NIKON_FNumberLock = 0xD110
const DPC_OLYMPUS_AEBracketingFrame = 0xD110
const DPC_CANON_EOS_BatteryPower = 0xD111
const DPC_OLYMPUS_AEBracketingStep = 0xD111
const DPC_CANON_EOS_BatterySelect = 0xD112
const DPC_NIKON_TVLockSetting = 0xD112
const DPC_OLYMPUS_WBBracketingFrame = 0xD112
const DPC_CANON_EOS_CameraTime = 0xD113
const DPC_NIKON_AVLockSetting = 0xD113
const DPC_OLYMPUS_WBBracketingRBRange = 0xD113
const DPC_NIKON_IllumSetting = 0xD114
const DPC_OLYMPUS_WBBracketingGMFrame = 0xD114
const DPC_CANON_EOS_Owner = 0xD115
const DPC_NIKON_FocusPointBright = 0xD115
const DPC_OLYMPUS_WBBracketingGMRange = 0xD115
const DPC_CANON_EOS_ModelID = 0xD116
const DPC_OLYMPUS_FLBracketingFrame = 0xD118
const DPC_CANON_EOS_PTPExtensionVersion = 0xD119
const DPC_OLYMPUS_FLBracketingStep = 0xD119
const DPC_CANON_EOS_DPOFVersion = 0xD11A
const DPC_OLYMPUS_FlashBiasCompensation = 0xD11A
const DPC_CANON_EOS_AvailableShots = 0xD11B
const DPC_OLYMPUS_ManualFocusMode = 0xD11B
const DPC_CANON_EOS_CaptureDestination = 0xD11C
const DPC_CANON_EOS_BracketMode = 0xD11D
const DPC_OLYMPUS_RawSaveMode = 0xD11D
const DPC_CANON_EOS_CurrentStorage = 0xD11E
const DPC_OLYMPUS_AUXLightMode = 0xD11E
const DPC_CANON_EOS_CurrentFolder = 0xD11F
const DPC_OLYMPUS_LensSinkMode = 0xD11F
const DPC_NIKON_ExternalFlashAttached = 0xD120
const DPC_OLYMPUS_BeepStatus = 0xD120
const DPC_NIKON_ExternalFlashStatus = 0xD121
const DPC_NIKON_ExternalFlashSort = 0xD122
const DPC_OLYMPUS_ColorSpace = 0xD122
const DPC_NIKON_ExternalFlashMode = 0xD123
const DPC_OLYMPUS_ColorMatching = 0xD123
const DPC_NIKON_ExternalFlashCompensation = 0xD124
const DPC_OLYMPUS_Saturation = 0xD124
const DPC_NIKON_NewExternalFlashMode = 0xD125
const DPC_NIKON_FlashExposureCompensation = 0xD126
const DPC_OLYMPUS_NoiseReductionPattern = 0xD126
const DPC_OLYMPUS_NoiseReductionRandom = 0xD127
const DPC_OLYMPUS_ShadingMode = 0xD129
const DPC_OLYMPUS_ISOBoostMode = 0xD12A
const DPC_OLYMPUS_ExposureIndexBiasStep = 0xD12B
const DPC_OLYMPUS_FilterEffect = 0xD12C
const DPC_OLYMPUS_ColorTune = 0xD12D
const DPC_OLYMPUS_Language = 0xD12E
const DPC_OLYMPUS_LanguageCode = 0xD12F
const DPC_CANON_EOS_CompressionS = 0xD130
const DPC_NIKON_HDRMode = 0xD130
const DPC_OLYMPUS_RecviewMode = 0xD130
const DPC_CANON_EOS_CompressionM1 = 0xD131
const DPC_MTP_PlaysForSureID = 0xD131
const DPC_NIKON_HDRHighDynamic = 0xD131
const DPC_OLYMPUS_SleepTime = 0xD131
const DPC_CANON_EOS_CompressionM2 = 0xD132
const DPC_MTP_ZUNE_UNKNOWN2 = 0xD132
const DPC_NIKON_HDRSmoothing = 0xD132
const DPC_OLYMPUS_ManualWBGMBias = 0xD132
const DPC_CANON_EOS_CompressionL = 0xD133
const DPC_OLYMPUS_AELAFLMode = 0xD135
const DPC_OLYMPUS_AELButtonStatus = 0xD136
const DPC_OLYMPUS_CompressionSettingEx = 0xD137
const DPC_OLYMPUS_ToneMode = 0xD139
const DPC_OLYMPUS_GradationMode = 0xD13A
const DPC_OLYMPUS_DevelopMode = 0xD13B
const DPC_OLYMPUS_ExtendInnerFlashMode = 0xD13C
const DPC_OLYMPUS_OutputDeviceMode = 0xD13D
const DPC_OLYMPUS_LiveViewMode = 0xD13E
const DPC_CANON_EOS_PCWhiteBalance1 = 0xD140
const DPC_NIKON_OptimizeImage = 0xD140
const DPC_OLYMPUS_LCDBacklight = 0xD140
const DPC_CANON_EOS_PCWhiteBalance2 = 0xD141
const DPC_OLYMPUS_CustomDevelop = 0xD141
const DPC_CANON_EOS_PCWhiteBalance3 = 0xD142
const DPC_NIKON_Saturation = 0xD142
const DPC_OLYMPUS_GradationAutoBias = 0xD142
const DPC_CANON_EOS_PCWhiteBalance4 = 0xD143
const DPC_NIKON_BW_FillerEffect = 0xD143
const DPC_OLYMPUS_FlashRCMode = 0xD143
const DPC_CANON_EOS_PCWhiteBalance5 = 0xD144
const DPC_NIKON_BW_Sharpness = 0xD144
const DPC_OLYMPUS_FlashRCGroupValue = 0xD144
const DPC_CANON_EOS_MWhiteBalance = 0xD145
const DPC_NIKON_BW_Contrast = 0xD145
const DPC_OLYMPUS_FlashRCChannelValue = 0xD145
const DPC_NIKON_BW_Setting_Type = 0xD146
const DPC_OLYMPUS_FlashRCFPMode = 0xD146
const DPC_OLYMPUS_FlashRCPhotoChromicMode = 0xD147
const DPC_NIKON_Slot2SaveMode = 0xD148
const DPC_OLYMPUS_FlashRCPhotoChromicBias = 0xD148
const DPC_NIKON_RawBitMode = 0xD149
const DPC_OLYMPUS_FlashRCPhotoChromicManualBias = 0xD149
const DPC_OLYMPUS_FlashRCQuantityLightLevel = 0xD14A
const DPC_OLYMPUS_FocusMeteringValue = 0xD14B
const DPC_OLYMPUS_ISOBracketingFrame = 0xD14C
const DPC_OLYMPUS_ISOBracketingStep = 0xD14D
const DPC_NIKON_ISOAutoTime = 0xD14E
const DPC_OLYMPUS_BulbMFMode = 0xD14E
const DPC_NIKON_FlourescentType = 0xD14F
const DPC_OLYMPUS_BurstFPSValue = 0xD14F
const DPC_CANON_EOS_PictureStyleStandard = 0xD150
const DPC_NIKON_TuneColourTemperature = 0xD150
const DPC_OLYMPUS_ISOAutoBaseValue = 0xD150
const DPC_CANON_EOS_PictureStylePortrait = 0xD151
const DPC_NIKON_TunePreset0 = 0xD151
const DPC_OLYMPUS_ISOAutoMaxValue = 0xD151
const DPC_CANON_EOS_PictureStyleLandscape = 0xD152
const DPC_NIKON_TunePreset1 = 0xD152
const DPC_OLYMPUS_BulbLimiterValue = 0xD152
const DPC_CANON_EOS_PictureStyleNeutral = 0xD153
const DPC_NIKON_TunePreset2 = 0xD153
const DPC_OLYMPUS_DPIMode = 0xD153
const DPC_CANON_EOS_PictureStyleFaithful = 0xD154
const DPC_NIKON_TunePreset3 = 0xD154
const DPC_OLYMPUS_DPICustomValue = 0xD154
const DPC_CANON_EOS_PictureStyleBlackWhite = 0xD155
const DPC_NIKON_TunePreset4 = 0xD155
const DPC_OLYMPUS_ResolutionValueSetting = 0xD155
const DPC_OLYMPUS_AFTargetSize = 0xD157
const DPC_OLYMPUS_LightSensorMode = 0xD158
const DPC_OLYMPUS_AEBracket = 0xD159
const DPC_OLYMPUS_WBRBBracket = 0xD15A
const DPC_OLYMPUS_WBGMBracket = 0xD15B
const DPC_OLYMPUS_FlashBracket = 0xD15C
const DPC_OLYMPUS_ISOBracket = 0xD15D
const DPC_OLYMPUS_MyModeStatus = 0xD15E
const DPC_CANON_EOS_PictureStyleUserSet1 = 0xD160
const DPC_NIKON_BeepOff = 0xD160
const DPC_CANON_EOS_PictureStyleUserSet2 = 0xD161
const DPC_NIKON_AutofocusMode = 0xD161
const DPC_CANON_EOS_PictureStyleUserSet3 = 0xD162
const DPC_NIKON_AFAssist = 0xD163
const DPC_NIKON_ImageReview = 0xD165
const DPC_NIKON_AFAreaIllumination = 0xD166
const DPC_NIKON_FlashMode = 0xD167
const DPC_NIKON_FlashCommanderMode = 0xD168
const DPC_NIKON_FlashSign = 0xD169
const DPC_NIKON_ISO_Auto = 0xD16A
const DPC_NIKON_RemoteTimeout = 0xD16B
const DPC_NIKON_GridDisplay = 0xD16C
const DPC_NIKON_FlashModeManualPower = 0xD16D
const DPC_NIKON_FlashModeCommanderPower = 0xD16E
const DPC_NIKON_AutoFP = 0xD16F
const DPC_CANON_EOS_PictureStyleParam1 = 0xD170
const DPC_CANON_EOS_PictureStyleParam2 = 0xD171
const DPC_CANON_EOS_PictureStyleParam3 = 0xD172
const DPC_CANON_EOS_FlavorLUTParams = 0xD17f
const DPC_CANON_EOS_CustomFunc1 = 0xD180
const DPC_NIKON_CSMMenu = 0xD180
const DPC_CANON_EOS_CustomFunc2 = 0xD181
const DPC_MTP_ZUNE_UNKNOWN1 = 0xD181
const DPC_MTP_Zune_UnknownVersion = 0xD181
const DPC_NIKON_WarningDisplay = 0xD181
const DPC_CANON_EOS_CustomFunc3 = 0xD182
const DPC_NIKON_BatteryCellKind = 0xD182
const DPC_CANON_EOS_CustomFunc4 = 0xD183
const DPC_NIKON_ISOAutoHiLimit = 0xD183
const DPC_CANON_EOS_CustomFunc5 = 0xD184
const DPC_NIKON_DynamicAFArea = 0xD184
const DPC_CANON_EOS_CustomFunc6 = 0xD185
const DPC_CANON_EOS_CustomFunc7 = 0xD186
const DPC_NIKON_ContinuousSpeedHigh = 0xD186
const DPC_CANON_EOS_CustomFunc8 = 0xD187
const DPC_NIKON_InfoDispSetting = 0xD187
const DPC_CANON_EOS_CustomFunc9 = 0xD188
const DPC_CANON_EOS_CustomFunc10 = 0xD189
const DPC_NIKON_PreviewButton = 0xD189
const DPC_NIKON_PreviewButton2 = 0xD18A
const DPC_NIKON_AEAFLockButton2 = 0xD18B
const DPC_NIKON_IndicatorDisp = 0xD18D
const DPC_NIKON_CellKindPriority = 0xD18E
const DPC_CANON_EOS_CustomFunc11 = 0xD18a
const DPC_CANON_EOS_CustomFunc12 = 0xD18b
const DPC_CANON_EOS_CustomFunc13 = 0xD18c
const DPC_CANON_EOS_CustomFunc14 = 0xD18d
const DPC_CANON_EOS_CustomFunc15 = 0xD18e
const DPC_CANON_EOS_CustomFunc16 = 0xD18f
const DPC_CANON_EOS_CustomFunc17 = 0xD190
const DPC_NIKON_BracketingFramesAndSteps = 0xD190
const DPC_CANON_EOS_CustomFunc18 = 0xD191
const DPC_CANON_EOS_CustomFunc19 = 0xD192
const DPC_NIKON_LiveViewMode = 0xD1A0
const DPC_NIKON_LiveViewDriveMode = 0xD1A1
const DPC_NIKON_LiveViewStatus = 0xD1A2
const DPC_NIKON_LiveViewImageZoomRatio = 0xD1A3
const DPC_NIKON_LiveViewProhibitCondition = 0xD1A4
const DPC_NIKON_ExposureDisplayStatus = 0xD1B0
const DPC_NIKON_ExposureIndicateStatus = 0xD1B1
const DPC_NIKON_InfoDispErrStatus = 0xD1B2
const DPC_NIKON_ExposureIndicateLightup = 0xD1B3
const DPC_NIKON_FlashOpen = 0xD1C0
const DPC_NIKON_FlashCharged = 0xD1C1
const DPC_NIKON_FlashMRepeatValue = 0xD1D0
const DPC_NIKON_FlashMRepeatCount = 0xD1D1
const DPC_NIKON_FlashMRepeatInterval = 0xD1D2
const DPC_NIKON_FlashCommandChannel = 0xD1D3
const DPC_NIKON_FlashCommandSelfMode = 0xD1D4
const DPC_NIKON_FlashCommandSelfCompensation = 0xD1D5
const DPC_NIKON_FlashCommandSelfValue = 0xD1D6
const DPC_NIKON_FlashCommandAMode = 0xD1D7
const DPC_NIKON_FlashCommandACompensation = 0xD1D8
const DPC_NIKON_FlashCommandAValue = 0xD1D9
const DPC_NIKON_FlashCommandBMode = 0xD1DA
const DPC_NIKON_FlashCommandBCompensation = 0xD1DB
const DPC_NIKON_FlashCommandBValue = 0xD1DC
const DPC_CANON_EOS_CustomFuncEx = 0xD1a0
const DPC_CANON_EOS_MyMenu = 0xD1a1
const DPC_CANON_EOS_MyMenuList = 0xD1a2
const DPC_CANON_EOS_WftStatus = 0xD1a3
const DPC_CANON_EOS_WftInputTransmission = 0xD1a4
const DPC_CANON_EOS_HDDirectoryStructure = 0xD1a5
const DPC_CANON_EOS_BatteryInfo = 0xD1a6
const DPC_CANON_EOS_AdapterInfo = 0xD1a7
const DPC_CANON_EOS_LensStatus = 0xD1a8
const DPC_CANON_EOS_QuickReviewTime = 0xD1a9
const DPC_CANON_EOS_CardExtension = 0xD1aa
const DPC_CANON_EOS_TempStatus = 0xD1ab
const DPC_CANON_EOS_ShutterCounter = 0xD1ac
const DPC_CANON_EOS_SpecialOption = 0xD1ad
const DPC_CANON_EOS_PhotoStudioMode = 0xD1ae
const DPC_CANON_EOS_SerialNumber = 0xD1af
const DPC_CANON_EOS_EVFOutputDevice = 0xD1b0
const DPC_CANON_EOS_EVFMode = 0xD1b1
const DPC_CANON_EOS_DepthOfFieldPreview = 0xD1b2
const DPC_CANON_EOS_EVFSharpness = 0xD1b3
const DPC_CANON_EOS_EVFWBMode = 0xD1b4
const DPC_CANON_EOS_EVFClickWBCoeffs = 0xD1b5
const DPC_CANON_EOS_EVFColorTemp = 0xD1b6
const DPC_CANON_EOS_ExposureSimMode = 0xD1b7
const DPC_CANON_EOS_EVFRecordStatus = 0xD1b8
const DPC_CANON_EOS_LvAfSystem = 0xD1ba
const DPC_CANON_EOS_MovSize = 0xD1bb
const DPC_CANON_EOS_LvViewTypeSelect = 0xD1bc
const DPC_CANON_EOS_Artist = 0xD1d0
const DPC_CANON_EOS_Copyright = 0xD1d1
const DPC_CANON_EOS_BracketValue = 0xD1d2
const DPC_CANON_EOS_FocusInfoEx = 0xD1d3
const DPC_CANON_EOS_DepthOfField = 0xD1d4
const DPC_CANON_EOS_Brightness = 0xD1d5
const DPC_CANON_EOS_LensAdjustParams = 0xD1d6
const DPC_CANON_EOS_EFComp = 0xD1d7
const DPC_CANON_EOS_LensName = 0xD1d8
const DPC_CANON_EOS_AEB = 0xD1d9
const DPC_CANON_EOS_StroboSetting = 0xD1da
const DPC_CANON_EOS_StroboWirelessSetting = 0xD1db
const DPC_CANON_EOS_StroboFiring = 0xD1dc
const DPC_CANON_EOS_LensID = 0xD1dd
const DPC_NIKON_ActivePicCtrlItem = 0xD200
const DPC_FUJI_ReleaseMode = 0xD201
const DPC_NIKON_ChangePicCtrlItem = 0xD201
const DPC_FUJI_FocusAreas = 0xD206
const DPC_FUJI_AELock = 0xD213
const DPC_MTP_ZUNE_UNKNOWN3 = 0xD215
const DPC_MTP_ZUNE_UNKNOWN4 = 0xD216
const DPC_FUJI_Aperture = 0xD218
const DPC_FUJI_ShutterSpeed = 0xD219
const DPC_MTP_SynchronizationPartner = 0xD401
const DPC_MTP_DeviceFriendlyName = 0xD402
const DPC_MTP_VolumeLevel = 0xD403
const DPC_MTP_DeviceIcon = 0xD405
const DPC_MTP_SessionInitiatorInfo = 0xD406
const DPC_MTP_PerceivedDeviceType = 0xD407
const DPC_MTP_PlaybackRate = 0xD410
const DPC_MTP_PlaybackObject = 0xD411
const DPC_MTP_PlaybackContainerIndex = 0xD412
const DPC_MTP_PlaybackPosition = 0xD413
const DPC_EXTENSION_MASK = 0xF000

var DPC_names = map[int]string{0x5000: "Undefined",
	0x5001: "BatteryLevel",
	0x5002: "FunctionalMode",
	0x5003: "ImageSize",
	0x5004: "CompressionSetting",
	0x5005: "WhiteBalance",
	0x5006: "RGBGain",
	0x5007: "FNumber",
	0x5008: "FocalLength",
	0x5009: "FocusDistance",
	0x500A: "FocusMode",
	0x500B: "ExposureMeteringMode",
	0x500C: "FlashMode",
	0x500D: "ExposureTime",
	0x500E: "ExposureProgramMode",
	0x500F: "ExposureIndex",
	0x5010: "ExposureBiasCompensation",
	0x5011: "DateTime",
	0x5012: "CaptureDelay",
	0x5013: "StillCaptureMode",
	0x5014: "Contrast",
	0x5015: "Sharpness",
	0x5016: "DigitalZoom",
	0x5017: "EffectMode",
	0x5018: "BurstNumber",
	0x5019: "BurstInterval",
	0x501A: "TimelapseNumber",
	0x501B: "TimelapseInterval",
	0x501C: "FocusMeteringMode",
	0x501D: "UploadURL",
	0x501E: "Artist",
	0x501F: "CopyrightInfo",
	0x5020: "SupportedStreams",
	0x5021: "EnabledStreams",
	0x5022: "VideoFormat",
	0x5023: "VideoResolution",
	0x5024: "VideoQuality",
	0x5025: "VideoFrameRate",
	0x5026: "VideoContrast",
	0x5027: "VideoBrightness",
	0x5028: "AudioFormat",
	0x5029: "AudioBitrate",
	0x502A: "AudioSamplingRate",
	0x502B: "AudioBitPerSample",
	0x502C: "AudioVolume",
	0xD000: "EXTENSION",
	0xD080: "CASIO_UNKNOWN_18",
	0xD118: "OLYMPUS_FLBracketingFrame",
	0xD127: "OLYMPUS_NoiseReductionRandom",
	0xD129: "OLYMPUS_ShadingMode",
	0xD12A: "OLYMPUS_ISOBoostMode",
	0xD12B: "OLYMPUS_ExposureIndexBiasStep",
	0xD12C: "OLYMPUS_FilterEffect",
	0xD12D: "OLYMPUS_ColorTune",
	0xD12E: "OLYMPUS_Language",
	0xD12F: "OLYMPUS_LanguageCode",
	0xD135: "OLYMPUS_AELAFLMode",
	0xD136: "OLYMPUS_AELButtonStatus",
	0xD137: "OLYMPUS_CompressionSettingEx",
	0xD139: "OLYMPUS_ToneMode",
	0xD13A: "OLYMPUS_GradationMode",
	0xD13B: "OLYMPUS_DevelopMode",
	0xD13C: "OLYMPUS_ExtendInnerFlashMode",
	0xD13D: "OLYMPUS_OutputDeviceMode",
	0xD13E: "OLYMPUS_LiveViewMode",
	0xD147: "OLYMPUS_FlashRCPhotoChromicMode",
	0xD14A: "OLYMPUS_FlashRCQuantityLightLevel",
	0xD14B: "OLYMPUS_FocusMeteringValue",
	0xD14C: "OLYMPUS_ISOBracketingFrame",
	0xD14D: "OLYMPUS_ISOBracketingStep",
	0xD157: "OLYMPUS_AFTargetSize",
	0xD158: "OLYMPUS_LightSensorMode",
	0xD159: "OLYMPUS_AEBracket",
	0xD15A: "OLYMPUS_WBRBBracket",
	0xD15B: "OLYMPUS_WBGMBracket",
	0xD15C: "OLYMPUS_FlashBracket",
	0xD15D: "OLYMPUS_ISOBracket",
	0xD15E: "OLYMPUS_MyModeStatus",
	0xD215: "MTP_ZUNE_UNKNOWN3",
	0xD216: "MTP_ZUNE_UNKNOWN4",
	0xD401: "MTP_SynchronizationPartner",
	0xD402: "MTP_DeviceFriendlyName",
	0xD403: "MTP_VolumeLevel",
	0xD405: "MTP_DeviceIcon",
	0xD406: "MTP_SessionInitiatorInfo",
	0xD407: "MTP_PerceivedDeviceType",
	0xD410: "MTP_PlaybackRate",
	0xD411: "MTP_PlaybackObject",
	0xD412: "MTP_PlaybackContainerIndex",
	0xD413: "MTP_PlaybackPosition",
	0xF000: "EXTENSION_MASK",
}

// device property form field
const DPFF_None = 0x00
const DPFF_Range = 0x01
const DPFF_Enumeration = 0x02

var DPFF_names = map[int]string{0x00: "None",
	0x01: "Range",
	0x02: "Enumeration",
}

// device property get/set
const DPGS_Get = 0x00
const DPGS_GetSet = 0x01

var DPGS_names = map[int]string{0x00: "Get",
	0x01: "GetSet",
}

// data type code
const DTC_UNDEF = 0x0000
const DTC_INT8 = 0x0001
const DTC_UINT8 = 0x0002
const DTC_INT16 = 0x0003
const DTC_UINT16 = 0x0004
const DTC_INT32 = 0x0005
const DTC_UINT32 = 0x0006
const DTC_INT64 = 0x0007
const DTC_UINT64 = 0x0008
const DTC_INT128 = 0x0009
const DTC_UINT128 = 0x000A
const DTC_ARRAY_MASK = 0x4000
const DTC_STR = 0xFFFF

var DTC_names = map[int]string{0x0000: "UNDEF",
	0x0001: "INT8",
	0x0002: "UINT8",
	0x0003: "INT16",
	0x0004: "UINT16",
	0x0005: "INT32",
	0x0006: "UINT32",
	0x0007: "INT64",
	0x0008: "UINT64",
	0x0009: "INT128",
	0x000A: "UINT128",
	0x4000: "ARRAY_MASK",
	0xFFFF: "STR",
}

// event code
const EC_Undefined = 0x4000
const EC_CancelTransaction = 0x4001
const EC_ObjectAdded = 0x4002
const EC_ObjectRemoved = 0x4003
const EC_StoreAdded = 0x4004
const EC_StoreRemoved = 0x4005
const EC_DevicePropChanged = 0x4006
const EC_ObjectInfoChanged = 0x4007
const EC_DeviceInfoChanged = 0x4008
const EC_RequestObjectTransfer = 0x4009
const EC_StoreFull = 0x400A
const EC_DeviceReset = 0x400B
const EC_StorageInfoChanged = 0x400C
const EC_CaptureComplete = 0x400D
const EC_UnreportedStatus = 0x400E
const EC_CANON_ObjectInfoChanged = 0xC008
const EC_CANON_RequestObjectTransfer = 0xC009
const EC_CANON_CameraModeChanged = 0xC00C
const EC_CANON_ShutterButtonPressed = 0xC00E
const EC_CANON_StartDirectTransfer = 0xC011
const EC_CANON_StopDirectTransfer = 0xC013
const EC_Nikon_ObjectAddedInSDRAM = 0xC101
const EC_Nikon_CaptureCompleteRecInSdram = 0xC102
const EC_Nikon_AdvancedTransfer = 0xC103
const EC_Nikon_PreviewImageAdded = 0xC104
const EC_CANON_EOS_RequestObjectTransferTS = 0xC1a2
const EC_MTP_ObjectPropChanged = 0xC801
const EC_MTP_ObjectPropDescChanged = 0xC802
const EC_MTP_ObjectReferencesChanged = 0xC803
const EC_CANON_EOS_RequestGetEvent = 0xc101
const EC_CANON_EOS_ObjectAddedEx = 0xc181
const EC_CANON_EOS_ObjectRemoved = 0xc182
const EC_CANON_EOS_RequestGetObjectInfoEx = 0xc183
const EC_CANON_EOS_StorageStatusChanged = 0xc184
const EC_CANON_EOS_StorageInfoChanged = 0xc185
const EC_CANON_EOS_RequestObjectTransfer = 0xc186
const EC_CANON_EOS_ObjectInfoChangedEx = 0xc187
const EC_CANON_EOS_ObjectContentChanged = 0xc188
const EC_CANON_EOS_PropValueChanged = 0xc189
const EC_CANON_EOS_AvailListChanged = 0xc18a
const EC_CANON_EOS_CameraStatusChanged = 0xc18b
const EC_CANON_EOS_WillSoonShutdown = 0xc18d
const EC_CANON_EOS_ShutdownTimerUpdated = 0xc18e
const EC_CANON_EOS_RequestCancelTransfer = 0xc18f
const EC_CANON_EOS_RequestObjectTransferDT = 0xc190
const EC_CANON_EOS_RequestCancelTransferDT = 0xc191
const EC_CANON_EOS_StoreAdded = 0xc192
const EC_CANON_EOS_StoreRemoved = 0xc193
const EC_CANON_EOS_BulbExposureTime = 0xc194
const EC_CANON_EOS_RecordingTime = 0xc195
const EC_CANON_EOS_AfResult = 0xc1a3

var EC_names = map[int]string{0x4000: "Undefined",
	0x4001: "CancelTransaction",
	0x4002: "ObjectAdded",
	0x4003: "ObjectRemoved",
	0x4004: "StoreAdded",
	0x4005: "StoreRemoved",
	0x4006: "DevicePropChanged",
	0x4007: "ObjectInfoChanged",
	0x4008: "DeviceInfoChanged",
	0x4009: "RequestObjectTransfer",
	0x400A: "StoreFull",
	0x400B: "DeviceReset",
	0x400C: "StorageInfoChanged",
	0x400D: "CaptureComplete",
	0x400E: "UnreportedStatus",
	0xC101: "Nikon_ObjectAddedInSDRAM",
	0xC102: "Nikon_CaptureCompleteRecInSdram",
	0xC103: "Nikon_AdvancedTransfer",
	0xC104: "Nikon_PreviewImageAdded",
	0xC801: "MTP_ObjectPropChanged",
	0xC802: "MTP_ObjectPropDescChanged",
	0xC803: "MTP_ObjectReferencesChanged",
}

const ERROR_TIMEOUT = 0x02FA
const ERROR_CANCEL = 0x02FB
const ERROR_BADPARAM = 0x02FC
const ERROR_RESP_EXPECTED = 0x02FD
const ERROR_DATA_EXPECTED = 0x02FE
const ERROR_IO = 0x02FF

var ERROR_names = map[int]string{0x02FA: "TIMEOUT",
	0x02FB: "CANCEL",
	0x02FC: "BADPARAM",
	0x02FD: "RESP_EXPECTED",
	0x02FE: "DATA_EXPECTED",
	0x02FF: "IO",
}

const FST_Undefined = 0x0000
const FST_GenericFlat = 0x0001
const FST_GenericHierarchical = 0x0002
const FST_DCF = 0x0003

var FST_names = map[int]string{0x0000: "Undefined",
	0x0001: "GenericFlat",
	0x0002: "GenericHierarchical",
	0x0003: "DCF",
}

// get object handles
const GOH_ALL_ASSOCS = 0x00000000
const GOH_ALL_FORMATS = 0x00000000
const GOH_ALL_STORAGE = 0xffffffff
const GOH_ROOT_PARENT = 0xffffffff

var GOH_names = map[int64]string{0x00000000: "ALL_ASSOCS",
	0xffffffff: "ALL_STORAGE",
}

const HANDLER_ROOT = 0x00000000
const HANDLER_SPECIAL = 0xffffffff

var HANDLER_names = map[int64]string{0x00000000: "ROOT",
	0xffffffff: "SPECIAL",
}

const NIKON_MaxCurvePoints = 19

var NIKON_names = map[int]string{19: "MaxCurvePoints"}

// operation code
const OC_Undefined = 0x1000
const OC_GetDeviceInfo = 0x1001
const OC_OpenSession = 0x1002
const OC_CloseSession = 0x1003
const OC_GetStorageIDs = 0x1004
const OC_GetStorageInfo = 0x1005
const OC_GetNumObjects = 0x1006
const OC_GetObjectHandles = 0x1007
const OC_GetObjectInfo = 0x1008
const OC_GetObject = 0x1009
const OC_GetThumb = 0x100A
const OC_DeleteObject = 0x100B
const OC_SendObjectInfo = 0x100C
const OC_SendObject = 0x100D
const OC_InitiateCapture = 0x100E
const OC_FormatStore = 0x100F
const OC_ResetDevice = 0x1010
const OC_SelfTest = 0x1011
const OC_SetObjectProtection = 0x1012
const OC_PowerDown = 0x1013
const OC_GetDevicePropDesc = 0x1014
const OC_GetDevicePropValue = 0x1015
const OC_SetDevicePropValue = 0x1016
const OC_ResetDevicePropValue = 0x1017
const OC_TerminateOpenCapture = 0x1018
const OC_MoveObject = 0x1019
const OC_CopyObject = 0x101A
const OC_GetPartialObject = 0x101B
const OC_InitiateOpenCapture = 0x101C
const OC_StartEnumHandles = 0x101D
const OC_EnumHandles = 0x101E
const OC_StopEnumHandles = 0x101F
const OC_GetVendorExtensionMaps = 0x1020
const OC_GetVendorDeviceInfo = 0x1021
const OC_GetResizedImageObject = 0x1022
const OC_GetFilesystemManifest = 0x1023
const OC_GetStreamInfo = 0x1024
const OC_GetStream = 0x1025
const OC_EXTENSION = 0x9000
const OC_CANON_GetPartialObjectInfo = 0x9001
const OC_CASIO_STILL_START = 0x9001
const OC_CANON_SetObjectArchive = 0x9002
const OC_CASIO_STILL_STOP = 0x9002
const OC_CANON_KeepDeviceOn = 0x9003
const OC_EK_GetSerial = 0x9003
const OC_CANON_LockDeviceUI = 0x9004
const OC_EK_SetSerial = 0x9004
const OC_CANON_UnlockDeviceUI = 0x9005
const OC_EK_SendFileObjectInfo = 0x9005
const OC_CANON_GetObjectHandleByName = 0x9006
const OC_EK_SendFileObject = 0x9006
const OC_NIKON_GetProfileAllData = 0x9006
const OC_CASIO_FOCUS = 0x9007
const OC_NIKON_SendProfileData = 0x9007
const OC_CANON_InitiateReleaseControl = 0x9008
const OC_EK_SetText = 0x9008
const OC_NIKON_DeleteProfile = 0x9008
const OC_CANON_TerminateReleaseControl = 0x9009
const OC_CASIO_CF_PRESS = 0x9009
const OC_NIKON_SetProfileData = 0x9009
const OC_CANON_TerminatePlaybackMode = 0x900A
const OC_CASIO_CF_RELEASE = 0x900A
const OC_CANON_ViewfinderOn = 0x900B
const OC_CANON_ViewfinderOff = 0x900C
const OC_CASIO_GET_OBJECT_INFO = 0x900C
const OC_CANON_DoAeAfAwb = 0x900D
const OC_CANON_GetCustomizeSpec = 0x900E
const OC_CANON_GetCustomizeItemInfo = 0x900F
const OC_CANON_GetCustomizeData = 0x9010
const OC_NIKON_AdvancedTransfer = 0x9010
const OC_CANON_SetCustomizeData = 0x9011
const OC_NIKON_GetFileInfoInBlock = 0x9011
const OC_CANON_GetCaptureStatus = 0x9012
const OC_CANON_CheckEvent = 0x9013
const OC_CANON_FocusLock = 0x9014
const OC_CANON_FocusUnlock = 0x9015
const OC_CANON_GetLocalReleaseParam = 0x9016
const OC_CANON_SetLocalReleaseParam = 0x9017
const OC_CANON_AskAboutPcEvf = 0x9018
const OC_CANON_SendPartialObject = 0x9019
const OC_CANON_InitiateCaptureInMemory = 0x901A
const OC_CANON_GetPartialObjectEx = 0x901B
const OC_CANON_SetObjectTime = 0x901C
const OC_CANON_GetViewfinderImage = 0x901D
const OC_CANON_GetObjectAttributes = 0x901E
const OC_CANON_ChangeUSBProtocol = 0x901F
const OC_CANON_GetChanges = 0x9020
const OC_CANON_GetObjectInfoEx = 0x9021
const OC_CANON_InitiateDirectTransfer = 0x9022
const OC_CANON_TerminateDirectTransfer = 0x9023
const OC_CANON_SendObjectInfoByPath = 0x9024
const OC_CASIO_SHUTTER = 0x9024
const OC_CANON_SendObjectByPath = 0x9025
const OC_CASIO_GET_OBJECT = 0x9025
const OC_CANON_InitiateDirectTansferEx = 0x9026
const OC_CASIO_GET_THUMBNAIL = 0x9026
const OC_CANON_GetAncillaryObjectHandles = 0x9027
const OC_CASIO_GET_STILL_HANDLES = 0x9027
const OC_CANON_GetTreeInfo = 0x9028
const OC_CASIO_STILL_RESET = 0x9028
const OC_CANON_GetTreeSize = 0x9029
const OC_CASIO_HALF_PRESS = 0x9029
const OC_CANON_NotifyProgress = 0x902A
const OC_CASIO_HALF_RELEASE = 0x902A
const OC_CANON_NotifyCancelAccepted = 0x902B
const OC_CASIO_CS_PRESS = 0x902B
const OC_CANON_902C = 0x902C
const OC_CASIO_CS_RELEASE = 0x902C
const OC_CANON_GetDirectory = 0x902D
const OC_CASIO_ZOOM = 0x902D
const OC_CASIO_CZ_PRESS = 0x902E
const OC_CASIO_CZ_RELEASE = 0x902F
const OC_CANON_SetPairingInfo = 0x9030
const OC_CANON_GetPairingInfo = 0x9031
const OC_CANON_DeletePairingInfo = 0x9032
const OC_CANON_GetMACAddress = 0x9033
const OC_CANON_SetDisplayMonitor = 0x9034
const OC_CANON_PairingComplete = 0x9035
const OC_CANON_GetWirelessMAXChannel = 0x9036
const OC_CASIO_MOVIE_START = 0x9041
const OC_CASIO_MOVIE_STOP = 0x9042
const OC_CASIO_MOVIE_PRESS = 0x9043
const OC_CASIO_MOVIE_RELEASE = 0x9044
const OC_CASIO_GET_MOVIE_HANDLES = 0x9045
const OC_CASIO_MOVIE_RESET = 0x9046
const OC_NIKON_GetLargeThumb = 0x90C4
const OC_NIKON_GetPictCtrlData = 0x90CC
const OC_NIKON_SetPictCtrlData = 0x90CD
const OC_NIKON_DelCstPicCtrl = 0x90CE
const OC_NIKON_GetPicCtrlCapability = 0x90CF
const OC_NIKON_GetDevicePTPIPInfo = 0x90E0
const OC_CANON_EOS_GetStorageIDs = 0x9101
const OC_MTP_WMDRMPD_GetSecureTimeChallenge = 0x9101
const OC_OLYMPUS_Capture = 0x9101
const OC_CANON_EOS_GetStorageInfo = 0x9102
const OC_MTP_WMDRMPD_GetSecureTimeResponse = 0x9102
const OC_CANON_EOS_GetObjectInfo = 0x9103
const OC_MTP_WMDRMPD_SetLicenseResponse = 0x9103
const OC_OLYMPUS_SelfCleaning = 0x9103
const OC_CANON_EOS_GetObject = 0x9104
const OC_MTP_WMDRMPD_GetSyncList = 0x9104
const OC_CANON_EOS_DeleteObject = 0x9105
const OC_MTP_WMDRMPD_SendMeterChallengeQuery = 0x9105
const OC_CANON_EOS_FormatStore = 0x9106
const OC_MTP_WMDRMPD_GetMeterChallenge = 0x9106
const OC_OLYMPUS_SetRGBGain = 0x9106
const OC_CANON_EOS_GetPartialObject = 0x9107
const OC_MTP_WMDRMPD_SetMeterResponse = 0x9107
const OC_OLYMPUS_SetPresetMode = 0x9107
const OC_CANON_EOS_GetDeviceInfoEx = 0x9108
const OC_MTP_WMDRMPD_CleanDataStore = 0x9108
const OC_OLYMPUS_SetWBBiasAll = 0x9108
const OC_CANON_EOS_GetObjectInfoEx = 0x9109
const OC_MTP_WMDRMPD_GetLicenseState = 0x9109
const OC_CANON_EOS_GetThumbEx = 0x910A
const OC_MTP_WMDRMPD_SendWMDRMPDCommand = 0x910A
const OC_CANON_EOS_SendPartialObject = 0x910B
const OC_MTP_WMDRMPD_SendWMDRMPDRequest = 0x910B
const OC_CANON_EOS_SetObjectAttributes = 0x910C
const OC_CANON_EOS_GetObjectTime = 0x910D
const OC_CANON_EOS_SetObjectTime = 0x910E
const OC_CANON_EOS_RemoteRelease = 0x910F
const OC_OLYMPUS_GetCameraControlMode = 0x910a
const OC_OLYMPUS_SetCameraControlMode = 0x910b
const OC_OLYMPUS_SetWBRGBGain = 0x910c
const OC_CANON_EOS_SetDevicePropValueEx = 0x9110
const OC_CANON_EOS_GetRemoteMode = 0x9113
const OC_CANON_EOS_SetRemoteMode = 0x9114
const OC_CANON_EOS_SetEventMode = 0x9115
const OC_CANON_EOS_GetEvent = 0x9116
const OC_CANON_EOS_TransferComplete = 0x9117
const OC_CANON_EOS_CancelTransfer = 0x9118
const OC_CANON_EOS_ResetTransfer = 0x9119
const OC_CANON_EOS_PCHDDCapacity = 0x911A
const OC_CANON_EOS_SetUILock = 0x911B
const OC_CANON_EOS_ResetUILock = 0x911C
const OC_CANON_EOS_KeepDeviceOn = 0x911D
const OC_CANON_EOS_SetNullPacketMode = 0x911E
const OC_CANON_EOS_UpdateFirmware = 0x911F
const OC_CANON_EOS_TransferCompleteDT = 0x9120
const OC_CANON_EOS_CancelTransferDT = 0x9121
const OC_CANON_EOS_GetWftProfile = 0x9122
const OC_CANON_EOS_SetWftProfile = 0x9122
const OC_MTP_WPDWCN_ProcessWFCObject = 0x9122
const OC_CANON_EOS_SetProfileToWft = 0x9124
const OC_CANON_EOS_BulbStart = 0x9125
const OC_CANON_EOS_BulbEnd = 0x9126
const OC_CANON_EOS_RequestDevicePropValue = 0x9127
const OC_CANON_EOS_RemoteReleaseOn = 0x9128
const OC_CANON_EOS_RemoteReleaseOff = 0x9129
const OC_CANON_EOS_InitiateViewfinder = 0x9151
const OC_CANON_EOS_TerminateViewfinder = 0x9152
const OC_CANON_EOS_GetViewFinderData = 0x9153
const OC_CANON_EOS_DoAf = 0x9154
const OC_CANON_EOS_DriveLens = 0x9155
const OC_CANON_EOS_DepthOfFieldPreview = 0x9156
const OC_CANON_EOS_ClickWB = 0x9157
const OC_CANON_EOS_Zoom = 0x9158
const OC_CANON_EOS_ZoomPosition = 0x9159
const OC_CANON_EOS_SetLiveAfFrame = 0x915a
const OC_CANON_EOS_AfCancel = 0x9160
const OC_MTP_AAVT_OpenMediaSession = 0x9170
const OC_MTP_AAVT_CloseMediaSession = 0x9171
const OC_MTP_AAVT_GetNextDataBlock = 0x9172
const OC_MTP_AAVT_SetCurrentTimePosition = 0x9173
const OC_MTP_WMDRMND_SendRegistrationRequest = 0x9180
const OC_MTP_WMDRMND_GetRegistrationResponse = 0x9181
const OC_MTP_WMDRMND_GetProximityChallenge = 0x9182
const OC_MTP_WMDRMND_SendProximityResponse = 0x9183
const OC_MTP_WMDRMND_SendWMDRMNDLicenseRequest = 0x9184
const OC_MTP_WMDRMND_GetWMDRMNDLicenseResponse = 0x9185
const OC_CANON_EOS_FAPIMessageTX = 0x91FE
const OC_CANON_EOS_FAPIMessageRX = 0x91FF
const OC_NIKON_GetPreviewImg = 0x9200
const OC_MTP_WMPPD_ReportAddedDeletedItems = 0x9201
const OC_NIKON_StartLiveView = 0x9201
const OC_MTP_WMPPD_ReportAcquiredItems = 0x9202
const OC_NIKON_EndLiveView = 0x9202
const OC_MTP_WMPPD_PlaylistObjectPref = 0x9203
const OC_NIKON_GetLiveViewImg = 0x9203
const OC_MTP_ZUNE_GETUNDEFINED001 = 0x9204
const OC_NIKON_MfDrive = 0x9204
const OC_NIKON_ChangeAfArea = 0x9205
const OC_NIKON_AfDriveCancel = 0x9206
const OC_MTP_WMDRMPD_SendWMDRMPDAppRequest = 0x9212
const OC_MTP_WMDRMPD_GetWMDRMPDAppResponse = 0x9213
const OC_MTP_WMDRMPD_EnableTrustedFilesOperations = 0x9214
const OC_MTP_WMDRMPD_DisableTrustedFilesOperations = 0x9215
const OC_MTP_WMDRMPD_EndTrustedAppSession = 0x9216
const OC_OLYMPUS_GetDeviceInfo = 0x9301
const OC_OLYMPUS_Init1 = 0x9302
const OC_OLYMPUS_SetDateTime = 0x9402
const OC_OLYMPUS_GetDateTime = 0x9482
const OC_OLYMPUS_SetCameraID = 0x9501
const OC_OLYMPUS_GetCameraID = 0x9581
const OC_MTP_GetObjectPropsSupported = 0x9801
const OC_MTP_GetObjectPropDesc = 0x9802
const OC_MTP_GetObjectPropValue = 0x9803
const OC_MTP_SetObjectPropValue = 0x9804
const OC_MTP_GetObjPropList = 0x9805
const OC_MTP_SetObjPropList = 0x9806
const OC_MTP_GetInterdependendPropdesc = 0x9807
const OC_MTP_SendObjectPropList = 0x9808
const OC_MTP_GetObjectReferences = 0x9810
const OC_MTP_SetObjectReferences = 0x9811
const OC_MTP_UpdateDeviceFirmware = 0x9812
const OC_MTP_Skip = 0x9820
const OC_CHDK = 0x9999
const OC_EXTENSION_MASK = 0xF000

var OC_names = map[int]string{0x1000: "Undefined",
	0x1001: "GetDeviceInfo",
	0x1002: "OpenSession",
	0x1003: "CloseSession",
	0x1004: "GetStorageIDs",
	0x1005: "GetStorageInfo",
	0x1006: "GetNumObjects",
	0x1007: "GetObjectHandles",
	0x1008: "GetObjectInfo",
	0x1009: "GetObject",
	0x100A: "GetThumb",
	0x100B: "DeleteObject",
	0x100C: "SendObjectInfo",
	0x100D: "SendObject",
	0x100E: "InitiateCapture",
	0x100F: "FormatStore",
	0x1010: "ResetDevice",
	0x1011: "SelfTest",
	0x1012: "SetObjectProtection",
	0x1013: "PowerDown",
	0x1014: "GetDevicePropDesc",
	0x1015: "GetDevicePropValue",
	0x1016: "SetDevicePropValue",
	0x1017: "ResetDevicePropValue",
	0x1018: "TerminateOpenCapture",
	0x1019: "MoveObject",
	0x101A: "CopyObject",
	0x101B: "GetPartialObject",
	0x101C: "InitiateOpenCapture",
	0x101D: "StartEnumHandles",
	0x101E: "EnumHandles",
	0x101F: "StopEnumHandles",
	0x1020: "GetVendorExtensionMaps",
	0x1021: "GetVendorDeviceInfo",
	0x1022: "GetResizedImageObject",
	0x1023: "GetFilesystemManifest",
	0x1024: "GetStreamInfo",
	0x1025: "GetStream",
	0x9000: "EXTENSION",
	0x9007: "CASIO_FOCUS",
	0x902E: "CASIO_CZ_PRESS",
	0x902F: "CASIO_CZ_RELEASE",
	0x9041: "CASIO_MOVIE_START",
	0x9042: "CASIO_MOVIE_STOP",
	0x9043: "CASIO_MOVIE_PRESS",
	0x9044: "CASIO_MOVIE_RELEASE",
	0x9045: "CASIO_GET_MOVIE_HANDLES",
	0x9046: "CASIO_MOVIE_RESET",
	0x9170: "MTP_AAVT_OpenMediaSession",
	0x9171: "MTP_AAVT_CloseMediaSession",
	0x9172: "MTP_AAVT_GetNextDataBlock",
	0x9173: "MTP_AAVT_SetCurrentTimePosition",
	0x9180: "MTP_WMDRMND_SendRegistrationRequest",
	0x9181: "MTP_WMDRMND_GetRegistrationResponse",
	0x9182: "MTP_WMDRMND_GetProximityChallenge",
	0x9183: "MTP_WMDRMND_SendProximityResponse",
	0x9184: "MTP_WMDRMND_SendWMDRMNDLicenseRequest",
	0x9185: "MTP_WMDRMND_GetWMDRMNDLicenseResponse",
	0x9201: "MTP_WMPPD_ReportAddedDeletedItems",
	0x9202: "MTP_WMPPD_ReportAcquiredItems",
	0x9203: "MTP_WMPPD_PlaylistObjectPref",
	0x9204: "MTP_ZUNE_GETUNDEFINED001",
	0x9212: "MTP_WMDRMPD_SendWMDRMPDAppRequest",
	0x9213: "MTP_WMDRMPD_GetWMDRMPDAppResponse",
	0x9214: "MTP_WMDRMPD_EnableTrustedFilesOperations",
	0x9215: "MTP_WMDRMPD_DisableTrustedFilesOperations",
	0x9216: "MTP_WMDRMPD_EndTrustedAppSession",
	0x9301: "OLYMPUS_GetDeviceInfo",
	0x9302: "OLYMPUS_Init1",
	0x9402: "OLYMPUS_SetDateTime",
	0x9482: "OLYMPUS_GetDateTime",
	0x9501: "OLYMPUS_SetCameraID",
	0x9581: "OLYMPUS_GetCameraID",
	0x9801: "MTP_GetObjectPropsSupported",
	0x9802: "MTP_GetObjectPropDesc",
	0x9803: "MTP_GetObjectPropValue",
	0x9804: "MTP_SetObjectPropValue",
	0x9805: "MTP_GetObjPropList",
	0x9806: "MTP_SetObjPropList",
	0x9807: "MTP_GetInterdependendPropdesc",
	0x9808: "MTP_SendObjectPropList",
	0x9810: "MTP_GetObjectReferences",
	0x9811: "MTP_SetObjectReferences",
	0x9812: "MTP_UpdateDeviceFirmware",
	0x9820: "MTP_Skip",
	0x9999: "CHDK",
	0xF000: "EXTENSION_MASK",
}

// object format code
const OFC_Undefined = 0x3000
const OFC_Association = 0x3001
const OFC_Script = 0x3002
const OFC_Executable = 0x3003
const OFC_Text = 0x3004
const OFC_HTML = 0x3005
const OFC_DPOF = 0x3006
const OFC_AIFF = 0x3007
const OFC_WAV = 0x3008
const OFC_MP3 = 0x3009
const OFC_AVI = 0x300A
const OFC_MPEG = 0x300B
const OFC_ASF = 0x300C
const OFC_Defined = 0x3800
const OFC_EXIF_JPEG = 0x3801
const OFC_TIFF_EP = 0x3802
const OFC_FlashPix = 0x3803
const OFC_BMP = 0x3804
const OFC_CIFF = 0x3805
const OFC_Undefined_0x3806 = 0x3806
const OFC_GIF = 0x3807
const OFC_JFIF = 0x3808
const OFC_PCD = 0x3809
const OFC_PICT = 0x380A
const OFC_PNG = 0x380B
const OFC_Undefined_0x380C = 0x380C
const OFC_TIFF = 0x380D
const OFC_TIFF_IT = 0x380E
const OFC_JP2 = 0x380F
const OFC_JPX = 0x3810
const OFC_DNG = 0x3811
const OFC_EK_M3U = 0xb002
const OFC_CANON_CRW = 0xb101
const OFC_CANON_CRW3 = 0xb103
const OFC_CANON_MOV = 0xb104
const OFC_CANON_CHDK_CRW = 0xb1ff
const OFC_MTP_MediaCard = 0xb211
const OFC_MTP_MediaCardGroup = 0xb212
const OFC_MTP_Encounter = 0xb213
const OFC_MTP_EncounterBox = 0xb214
const OFC_MTP_M4A = 0xb215
const OFC_MTP_Firmware = 0xb802
const OFC_MTP_WindowsImageFormat = 0xb881
const OFC_MTP_UndefinedAudio = 0xb900
const OFC_MTP_WMA = 0xb901
const OFC_MTP_OGG = 0xb902
const OFC_MTP_AAC = 0xb903
const OFC_MTP_AudibleCodec = 0xb904
const OFC_MTP_FLAC = 0xb906
const OFC_MTP_SamsungPlaylist = 0xb909
const OFC_MTP_UndefinedVideo = 0xb980
const OFC_MTP_WMV = 0xb981
const OFC_MTP_MP4 = 0xb982
const OFC_MTP_MP2 = 0xb983
const OFC_MTP_3GP = 0xb984
const OFC_MTP_UndefinedCollection = 0xba00
const OFC_MTP_AbstractMultimediaAlbum = 0xba01
const OFC_MTP_AbstractImageAlbum = 0xba02
const OFC_MTP_AbstractAudioAlbum = 0xba03
const OFC_MTP_AbstractVideoAlbum = 0xba04
const OFC_MTP_AbstractAudioVideoPlaylist = 0xba05
const OFC_MTP_AbstractContactGroup = 0xba06
const OFC_MTP_AbstractMessageFolder = 0xba07
const OFC_MTP_AbstractChapteredProduction = 0xba08
const OFC_MTP_AbstractAudioPlaylist = 0xba09
const OFC_MTP_AbstractVideoPlaylist = 0xba0a
const OFC_MTP_AbstractMediacast = 0xba0b
const OFC_MTP_WPLPlaylist = 0xba10
const OFC_MTP_M3UPlaylist = 0xba11
const OFC_MTP_MPLPlaylist = 0xba12
const OFC_MTP_ASXPlaylist = 0xba13
const OFC_MTP_PLSPlaylist = 0xba14
const OFC_MTP_UndefinedDocument = 0xba80
const OFC_MTP_AbstractDocument = 0xba81
const OFC_MTP_XMLDocument = 0xba82
const OFC_MTP_MSWordDocument = 0xba83
const OFC_MTP_MHTCompiledHTMLDocument = 0xba84
const OFC_MTP_MSExcelSpreadsheetXLS = 0xba85
const OFC_MTP_MSPowerpointPresentationPPT = 0xba86
const OFC_MTP_UndefinedMessage = 0xbb00
const OFC_MTP_AbstractMessage = 0xbb01
const OFC_MTP_UndefinedContact = 0xbb80
const OFC_MTP_AbstractContact = 0xbb81
const OFC_MTP_vCard2 = 0xbb82
const OFC_MTP_vCard3 = 0xbb83
const OFC_MTP_UndefinedCalendarItem = 0xbe00
const OFC_MTP_AbstractCalendarItem = 0xbe01
const OFC_MTP_vCalendar1 = 0xbe02
const OFC_MTP_vCalendar2 = 0xbe03
const OFC_MTP_UndefinedWindowsExecutable = 0xbe80
const OFC_MTP_MediaCast = 0xbe81
const OFC_MTP_Section = 0xbe82

var OFC_names = map[int]string{0x3000: "Undefined",
	0x3001: "Association",
	0x3002: "Script",
	0x3003: "Executable",
	0x3004: "Text",
	0x3005: "HTML",
	0x3006: "DPOF",
	0x3007: "AIFF",
	0x3008: "WAV",
	0x3009: "MP3",
	0x300A: "AVI",
	0x300B: "MPEG",
	0x300C: "ASF",
	0x3800: "Defined",
	0x3801: "EXIF_JPEG",
	0x3802: "TIFF_EP",
	0x3803: "FlashPix",
	0x3804: "BMP",
	0x3805: "CIFF",
	0x3806: "Undefined_0x3806",
	0x3807: "GIF",
	0x3808: "JFIF",
	0x3809: "PCD",
	0x380A: "PICT",
	0x380B: "PNG",
	0x380C: "Undefined_0x380C",
	0x380D: "TIFF",
	0x380E: "TIFF_IT",
	0x380F: "JP2",
	0x3810: "JPX",
	0x3811: "DNG",
	0xb211: "MTP_MediaCard",
	0xb212: "MTP_MediaCardGroup",
	0xb213: "MTP_Encounter",
	0xb214: "MTP_EncounterBox",
	0xb215: "MTP_M4A",
	0xb802: "MTP_Firmware",
	0xb881: "MTP_WindowsImageFormat",
	0xb900: "MTP_UndefinedAudio",
	0xb901: "MTP_WMA",
	0xb902: "MTP_OGG",
	0xb903: "MTP_AAC",
	0xb904: "MTP_AudibleCodec",
	0xb906: "MTP_FLAC",
	0xb909: "MTP_SamsungPlaylist",
	0xb980: "MTP_UndefinedVideo",
	0xb981: "MTP_WMV",
	0xb982: "MTP_MP4",
	0xb983: "MTP_MP2",
	0xb984: "MTP_3GP",
	0xba00: "MTP_UndefinedCollection",
	0xba01: "MTP_AbstractMultimediaAlbum",
	0xba02: "MTP_AbstractImageAlbum",
	0xba03: "MTP_AbstractAudioAlbum",
	0xba04: "MTP_AbstractVideoAlbum",
	0xba05: "MTP_AbstractAudioVideoPlaylist",
	0xba06: "MTP_AbstractContactGroup",
	0xba07: "MTP_AbstractMessageFolder",
	0xba08: "MTP_AbstractChapteredProduction",
	0xba09: "MTP_AbstractAudioPlaylist",
	0xba0a: "MTP_AbstractVideoPlaylist",
	0xba0b: "MTP_AbstractMediacast",
	0xba10: "MTP_WPLPlaylist",
	0xba11: "MTP_M3UPlaylist",
	0xba12: "MTP_MPLPlaylist",
	0xba13: "MTP_ASXPlaylist",
	0xba14: "MTP_PLSPlaylist",
	0xba80: "MTP_UndefinedDocument",
	0xba81: "MTP_AbstractDocument",
	0xba82: "MTP_XMLDocument",
	0xba83: "MTP_MSWordDocument",
	0xba84: "MTP_MHTCompiledHTMLDocument",
	0xba85: "MTP_MSExcelSpreadsheetXLS",
	0xba86: "MTP_MSPowerpointPresentationPPT",
	0xbb00: "MTP_UndefinedMessage",
	0xbb01: "MTP_AbstractMessage",
	0xbb80: "MTP_UndefinedContact",
	0xbb81: "MTP_AbstractContact",
	0xbb82: "MTP_vCard2",
	0xbb83: "MTP_vCard3",
	0xbe00: "MTP_UndefinedCalendarItem",
	0xbe01: "MTP_AbstractCalendarItem",
	0xbe02: "MTP_vCalendar1",
	0xbe03: "MTP_vCalendar2",
	0xbe80: "MTP_UndefinedWindowsExecutable",
	0xbe81: "MTP_MediaCast",
	0xbe82: "MTP_Section",
}

// object property code
const OPC_WirelessConfigurationFile = 0xB104
const OPC_BuyFlag = 0xD901
const OPC_StorageID = 0xDC01
const OPC_ObjectFormat = 0xDC02
const OPC_ProtectionStatus = 0xDC03
const OPC_ObjectSize = 0xDC04
const OPC_AssociationType = 0xDC05
const OPC_AssociationDesc = 0xDC06
const OPC_ObjectFileName = 0xDC07
const OPC_DateCreated = 0xDC08
const OPC_DateModified = 0xDC09
const OPC_Keywords = 0xDC0A
const OPC_ParentObject = 0xDC0B
const OPC_AllowedFolderContents = 0xDC0C
const OPC_Hidden = 0xDC0D
const OPC_SystemObject = 0xDC0E
const OPC_PersistantUniqueObjectIdentifier = 0xDC41
const OPC_SyncID = 0xDC42
const OPC_PropertyBag = 0xDC43
const OPC_Name = 0xDC44
const OPC_CreatedBy = 0xDC45
const OPC_Artist = 0xDC46
const OPC_DateAuthored = 0xDC47
const OPC_Description = 0xDC48
const OPC_URLReference = 0xDC49
const OPC_LanguageLocale = 0xDC4A
const OPC_CopyrightInformation = 0xDC4B
const OPC_Source = 0xDC4C
const OPC_OriginLocation = 0xDC4D
const OPC_DateAdded = 0xDC4E
const OPC_NonConsumable = 0xDC4F
const OPC_CorruptOrUnplayable = 0xDC50
const OPC_ProducerSerialNumber = 0xDC51
const OPC_RepresentativeSampleFormat = 0xDC81
const OPC_RepresentativeSampleSize = 0xDC82
const OPC_RepresentativeSampleHeight = 0xDC83
const OPC_RepresentativeSampleWidth = 0xDC84
const OPC_RepresentativeSampleDuration = 0xDC85
const OPC_RepresentativeSampleData = 0xDC86
const OPC_Width = 0xDC87
const OPC_Height = 0xDC88
const OPC_Duration = 0xDC89
const OPC_Rating = 0xDC8A
const OPC_Track = 0xDC8B
const OPC_Genre = 0xDC8C
const OPC_Credits = 0xDC8D
const OPC_Lyrics = 0xDC8E
const OPC_SubscriptionContentID = 0xDC8F
const OPC_ProducedBy = 0xDC90
const OPC_UseCount = 0xDC91
const OPC_SkipCount = 0xDC92
const OPC_LastAccessed = 0xDC93
const OPC_ParentalRating = 0xDC94
const OPC_MetaGenre = 0xDC95
const OPC_Composer = 0xDC96
const OPC_EffectiveRating = 0xDC97
const OPC_Subtitle = 0xDC98
const OPC_OriginalReleaseDate = 0xDC99
const OPC_AlbumName = 0xDC9A
const OPC_AlbumArtist = 0xDC9B
const OPC_Mood = 0xDC9C
const OPC_DRMStatus = 0xDC9D
const OPC_SubDescription = 0xDC9E
const OPC_IsCropped = 0xDCD1
const OPC_IsColorCorrected = 0xDCD2
const OPC_ImageBitDepth = 0xDCD3
const OPC_Fnumber = 0xDCD4
const OPC_ExposureTime = 0xDCD5
const OPC_ExposureIndex = 0xDCD6
const OPC_DisplayName = 0xDCE0
const OPC_BodyText = 0xDCE1
const OPC_Subject = 0xDCE2
const OPC_Priority = 0xDCE3
const OPC_GivenName = 0xDD00
const OPC_MiddleNames = 0xDD01
const OPC_FamilyName = 0xDD02
const OPC_Prefix = 0xDD03
const OPC_Suffix = 0xDD04
const OPC_PhoneticGivenName = 0xDD05
const OPC_PhoneticFamilyName = 0xDD06
const OPC_EmailPrimary = 0xDD07
const OPC_EmailPersonal1 = 0xDD08
const OPC_EmailPersonal2 = 0xDD09
const OPC_EmailBusiness1 = 0xDD0A
const OPC_EmailBusiness2 = 0xDD0B
const OPC_EmailOthers = 0xDD0C
const OPC_PhoneNumberPrimary = 0xDD0D
const OPC_PhoneNumberPersonal = 0xDD0E
const OPC_PhoneNumberPersonal2 = 0xDD0F
const OPC_PhoneNumberBusiness = 0xDD10
const OPC_PhoneNumberBusiness2 = 0xDD11
const OPC_PhoneNumberMobile = 0xDD12
const OPC_PhoneNumberMobile2 = 0xDD13
const OPC_FaxNumberPrimary = 0xDD14
const OPC_FaxNumberPersonal = 0xDD15
const OPC_FaxNumberBusiness = 0xDD16
const OPC_PagerNumber = 0xDD17
const OPC_PhoneNumberOthers = 0xDD18
const OPC_PrimaryWebAddress = 0xDD19
const OPC_PersonalWebAddress = 0xDD1A
const OPC_BusinessWebAddress = 0xDD1B
const OPC_InstantMessengerAddress = 0xDD1C
const OPC_InstantMessengerAddress2 = 0xDD1D
const OPC_InstantMessengerAddress3 = 0xDD1E
const OPC_PostalAddressPersonalFull = 0xDD1F
const OPC_PostalAddressPersonalFullLine1 = 0xDD20
const OPC_PostalAddressPersonalFullLine2 = 0xDD21
const OPC_PostalAddressPersonalFullCity = 0xDD22
const OPC_PostalAddressPersonalFullRegion = 0xDD23
const OPC_PostalAddressPersonalFullPostalCode = 0xDD24
const OPC_PostalAddressPersonalFullCountry = 0xDD25
const OPC_PostalAddressBusinessFull = 0xDD26
const OPC_PostalAddressBusinessLine1 = 0xDD27
const OPC_PostalAddressBusinessLine2 = 0xDD28
const OPC_PostalAddressBusinessCity = 0xDD29
const OPC_PostalAddressBusinessRegion = 0xDD2A
const OPC_PostalAddressBusinessPostalCode = 0xDD2B
const OPC_PostalAddressBusinessCountry = 0xDD2C
const OPC_PostalAddressOtherFull = 0xDD2D
const OPC_PostalAddressOtherLine1 = 0xDD2E
const OPC_PostalAddressOtherLine2 = 0xDD2F
const OPC_PostalAddressOtherCity = 0xDD30
const OPC_PostalAddressOtherRegion = 0xDD31
const OPC_PostalAddressOtherPostalCode = 0xDD32
const OPC_PostalAddressOtherCountry = 0xDD33
const OPC_OrganizationName = 0xDD34
const OPC_PhoneticOrganizationName = 0xDD35
const OPC_Role = 0xDD36
const OPC_Birthdate = 0xDD37
const OPC_MessageTo = 0xDD40
const OPC_MessageCC = 0xDD41
const OPC_MessageBCC = 0xDD42
const OPC_MessageRead = 0xDD43
const OPC_MessageReceivedTime = 0xDD44
const OPC_MessageSender = 0xDD45
const OPC_ActivityBeginTime = 0xDD50
const OPC_ActivityEndTime = 0xDD51
const OPC_ActivityLocation = 0xDD52
const OPC_ActivityRequiredAttendees = 0xDD54
const OPC_ActivityOptionalAttendees = 0xDD55
const OPC_ActivityResources = 0xDD56
const OPC_ActivityAccepted = 0xDD57
const OPC_Owner = 0xDD5D
const OPC_Editor = 0xDD5E
const OPC_Webmaster = 0xDD5F
const OPC_URLSource = 0xDD60
const OPC_URLDestination = 0xDD61
const OPC_TimeBookmark = 0xDD62
const OPC_ObjectBookmark = 0xDD63
const OPC_ByteBookmark = 0xDD64
const OPC_LastBuildDate = 0xDD70
const OPC_TimetoLive = 0xDD71
const OPC_MediaGUID = 0xDD72
const OPC_TotalBitRate = 0xDE91
const OPC_BitRateType = 0xDE92
const OPC_SampleRate = 0xDE93
const OPC_NumberOfChannels = 0xDE94
const OPC_AudioBitDepth = 0xDE95
const OPC_ScanDepth = 0xDE97
const OPC_AudioWAVECodec = 0xDE99
const OPC_AudioBitRate = 0xDE9A
const OPC_VideoFourCCCodec = 0xDE9B
const OPC_VideoBitRate = 0xDE9C
const OPC_FramesPerThousandSeconds = 0xDE9D
const OPC_KeyFrameDistance = 0xDE9E
const OPC_BufferSize = 0xDE9F
const OPC_EncodingQuality = 0xDEA0
const OPC_EncodingProfile = 0xDEA1

var OPC_names = map[int]string{0xB104: "WirelessConfigurationFile",
	0xD901: "BuyFlag",
	0xDC01: "StorageID",
	0xDC02: "ObjectFormat",
	0xDC03: "ProtectionStatus",
	0xDC04: "ObjectSize",
	0xDC05: "AssociationType",
	0xDC06: "AssociationDesc",
	0xDC07: "ObjectFileName",
	0xDC08: "DateCreated",
	0xDC09: "DateModified",
	0xDC0A: "Keywords",
	0xDC0B: "ParentObject",
	0xDC0C: "AllowedFolderContents",
	0xDC0D: "Hidden",
	0xDC0E: "SystemObject",
	0xDC41: "PersistantUniqueObjectIdentifier",
	0xDC42: "SyncID",
	0xDC43: "PropertyBag",
	0xDC44: "Name",
	0xDC45: "CreatedBy",
	0xDC46: "Artist",
	0xDC47: "DateAuthored",
	0xDC48: "Description",
	0xDC49: "URLReference",
	0xDC4A: "LanguageLocale",
	0xDC4B: "CopyrightInformation",
	0xDC4C: "Source",
	0xDC4D: "OriginLocation",
	0xDC4E: "DateAdded",
	0xDC4F: "NonConsumable",
	0xDC50: "CorruptOrUnplayable",
	0xDC51: "ProducerSerialNumber",
	0xDC81: "RepresentativeSampleFormat",
	0xDC82: "RepresentativeSampleSize",
	0xDC83: "RepresentativeSampleHeight",
	0xDC84: "RepresentativeSampleWidth",
	0xDC85: "RepresentativeSampleDuration",
	0xDC86: "RepresentativeSampleData",
	0xDC87: "Width",
	0xDC88: "Height",
	0xDC89: "Duration",
	0xDC8A: "Rating",
	0xDC8B: "Track",
	0xDC8C: "Genre",
	0xDC8D: "Credits",
	0xDC8E: "Lyrics",
	0xDC8F: "SubscriptionContentID",
	0xDC90: "ProducedBy",
	0xDC91: "UseCount",
	0xDC92: "SkipCount",
	0xDC93: "LastAccessed",
	0xDC94: "ParentalRating",
	0xDC95: "MetaGenre",
	0xDC96: "Composer",
	0xDC97: "EffectiveRating",
	0xDC98: "Subtitle",
	0xDC99: "OriginalReleaseDate",
	0xDC9A: "AlbumName",
	0xDC9B: "AlbumArtist",
	0xDC9C: "Mood",
	0xDC9D: "DRMStatus",
	0xDC9E: "SubDescription",
	0xDCD1: "IsCropped",
	0xDCD2: "IsColorCorrected",
	0xDCD3: "ImageBitDepth",
	0xDCD4: "Fnumber",
	0xDCD5: "ExposureTime",
	0xDCD6: "ExposureIndex",
	0xDCE0: "DisplayName",
	0xDCE1: "BodyText",
	0xDCE2: "Subject",
	0xDCE3: "Priority",
	0xDD00: "GivenName",
	0xDD01: "MiddleNames",
	0xDD02: "FamilyName",
	0xDD03: "Prefix",
	0xDD04: "Suffix",
	0xDD05: "PhoneticGivenName",
	0xDD06: "PhoneticFamilyName",
	0xDD07: "EmailPrimary",
	0xDD08: "EmailPersonal1",
	0xDD09: "EmailPersonal2",
	0xDD0A: "EmailBusiness1",
	0xDD0B: "EmailBusiness2",
	0xDD0C: "EmailOthers",
	0xDD0D: "PhoneNumberPrimary",
	0xDD0E: "PhoneNumberPersonal",
	0xDD0F: "PhoneNumberPersonal2",
	0xDD10: "PhoneNumberBusiness",
	0xDD11: "PhoneNumberBusiness2",
	0xDD12: "PhoneNumberMobile",
	0xDD13: "PhoneNumberMobile2",
	0xDD14: "FaxNumberPrimary",
	0xDD15: "FaxNumberPersonal",
	0xDD16: "FaxNumberBusiness",
	0xDD17: "PagerNumber",
	0xDD18: "PhoneNumberOthers",
	0xDD19: "PrimaryWebAddress",
	0xDD1A: "PersonalWebAddress",
	0xDD1B: "BusinessWebAddress",
	0xDD1C: "InstantMessengerAddress",
	0xDD1D: "InstantMessengerAddress2",
	0xDD1E: "InstantMessengerAddress3",
	0xDD1F: "PostalAddressPersonalFull",
	0xDD20: "PostalAddressPersonalFullLine1",
	0xDD21: "PostalAddressPersonalFullLine2",
	0xDD22: "PostalAddressPersonalFullCity",
	0xDD23: "PostalAddressPersonalFullRegion",
	0xDD24: "PostalAddressPersonalFullPostalCode",
	0xDD25: "PostalAddressPersonalFullCountry",
	0xDD26: "PostalAddressBusinessFull",
	0xDD27: "PostalAddressBusinessLine1",
	0xDD28: "PostalAddressBusinessLine2",
	0xDD29: "PostalAddressBusinessCity",
	0xDD2A: "PostalAddressBusinessRegion",
	0xDD2B: "PostalAddressBusinessPostalCode",
	0xDD2C: "PostalAddressBusinessCountry",
	0xDD2D: "PostalAddressOtherFull",
	0xDD2E: "PostalAddressOtherLine1",
	0xDD2F: "PostalAddressOtherLine2",
	0xDD30: "PostalAddressOtherCity",
	0xDD31: "PostalAddressOtherRegion",
	0xDD32: "PostalAddressOtherPostalCode",
	0xDD33: "PostalAddressOtherCountry",
	0xDD34: "OrganizationName",
	0xDD35: "PhoneticOrganizationName",
	0xDD36: "Role",
	0xDD37: "Birthdate",
	0xDD40: "MessageTo",
	0xDD41: "MessageCC",
	0xDD42: "MessageBCC",
	0xDD43: "MessageRead",
	0xDD44: "MessageReceivedTime",
	0xDD45: "MessageSender",
	0xDD50: "ActivityBeginTime",
	0xDD51: "ActivityEndTime",
	0xDD52: "ActivityLocation",
	0xDD54: "ActivityRequiredAttendees",
	0xDD55: "ActivityOptionalAttendees",
	0xDD56: "ActivityResources",
	0xDD57: "ActivityAccepted",
	0xDD5D: "Owner",
	0xDD5E: "Editor",
	0xDD5F: "Webmaster",
	0xDD60: "URLSource",
	0xDD61: "URLDestination",
	0xDD62: "TimeBookmark",
	0xDD63: "ObjectBookmark",
	0xDD64: "ByteBookmark",
	0xDD70: "LastBuildDate",
	0xDD71: "TimetoLive",
	0xDD72: "MediaGUID",
	0xDE91: "TotalBitRate",
	0xDE92: "BitRateType",
	0xDE93: "SampleRate",
	0xDE94: "NumberOfChannels",
	0xDE95: "AudioBitDepth",
	0xDE97: "ScanDepth",
	0xDE99: "AudioWAVECodec",
	0xDE9A: "AudioBitRate",
	0xDE9B: "VideoFourCCCodec",
	0xDE9C: "VideoBitRate",
	0xDE9D: "FramesPerThousandSeconds",
	0xDE9E: "KeyFrameDistance",
	0xDE9F: "BufferSize",
	0xDEA0: "EncodingQuality",
	0xDEA1: "EncodingProfile",
}

const OPFF_None = 0x00
const OPFF_Range = 0x01
const OPFF_Enumeration = 0x02
const OPFF_DateTime = 0x03
const OPFF_FixedLengthArray = 0x04
const OPFF_RegularExpression = 0x05
const OPFF_ByteArray = 0x06
const OPFF_LongString = 0xFF

var OPFF_names = map[int]string{0x00: "None",
	0x01: "Range",
	0x02: "Enumeration",
	0x03: "DateTime",
	0x04: "FixedLengthArray",
	0x05: "RegularExpression",
	0x06: "ByteArray",
	0xFF: "LongString",
}

const PS_NoProtection = 0x0000
const PS_ReadOnly = 0x0001
const PS_MTP_ReadOnlyData = 0x8002
const PS_MTP_NonTransferableData = 0x8003

var PS_names = map[int]string{0x0000: "NoProtection",
	0x0001: "ReadOnly",
	0x8002: "MTP_ReadOnlyData",
	0x8003: "MTP_NonTransferableData",
}

// return code
const RC_Undefined = 0x2000
const RC_OK = 0x2001
const RC_GeneralError = 0x2002
const RC_SessionNotOpen = 0x2003
const RC_InvalidTransactionID = 0x2004
const RC_OperationNotSupported = 0x2005
const RC_ParameterNotSupported = 0x2006
const RC_IncompleteTransfer = 0x2007
const RC_InvalidStorageId = 0x2008
const RC_InvalidObjectHandle = 0x2009
const RC_DevicePropNotSupported = 0x200A
const RC_InvalidObjectFormatCode = 0x200B
const RC_StoreFull = 0x200C
const RC_ObjectWriteProtected = 0x200D
const RC_StoreReadOnly = 0x200E
const RC_AccessDenied = 0x200F
const RC_NoThumbnailPresent = 0x2010
const RC_SelfTestFailed = 0x2011
const RC_PartialDeletion = 0x2012
const RC_StoreNotAvailable = 0x2013
const RC_SpecificationByFormatUnsupported = 0x2014
const RC_NoValidObjectInfo = 0x2015
const RC_InvalidCodeFormat = 0x2016
const RC_UnknownVendorCode = 0x2017
const RC_CaptureAlreadyTerminated = 0x2018
const RC_DeviceBusy = 0x2019
const RC_InvalidParentObject = 0x201A
const RC_InvalidDevicePropFormat = 0x201B
const RC_InvalidDevicePropValue = 0x201C
const RC_InvalidParameter = 0x201D
const RC_SessionAlreadyOpened = 0x201E
const RC_TransactionCanceled = 0x201F
const RC_SpecificationOfDestinationUnsupported = 0x2020
const RC_InvalidEnumHandle = 0x2021
const RC_NoStreamEnabled = 0x2022
const RC_InvalidDataSet = 0x2023
const RC_CANON_UNKNOWN_COMMAND = 0xA001
const RC_EK_FilenameRequired = 0xA001
const RC_NIKON_HardwareError = 0xA001
const RC_EK_FilenameConflicts = 0xA002
const RC_NIKON_OutOfFocus = 0xA002
const RC_EK_FilenameInvalid = 0xA003
const RC_NIKON_ChangeCameraModeFailed = 0xA003
const RC_NIKON_InvalidStatus = 0xA004
const RC_CANON_OPERATION_REFUSED = 0xA005
const RC_NIKON_SetPropertyNotSupported = 0xA005
const RC_CANON_LENS_COVER = 0xA006
const RC_NIKON_WbResetError = 0xA006
const RC_NIKON_DustReferenceError = 0xA007
const RC_NIKON_ShutterSpeedBulb = 0xA008
const RC_CANON_A009 = 0xA009
const RC_NIKON_MirrorUpSequence = 0xA009
const RC_NIKON_CameraModeNotAdjustFNumber = 0xA00A
const RC_NIKON_NotLiveView = 0xA00B
const RC_NIKON_MfDriveStepEnd = 0xA00C
const RC_NIKON_MfDriveStepInsufficiency = 0xA00E
const RC_NIKON_AdvancedTransferCancel = 0xA022
const RC_CANON_BATTERY_LOW = 0xA101
const RC_CANON_NOT_READY = 0xA102
const RC_MTP_Invalid_WFC_Syntax = 0xA121
const RC_MTP_WFC_Version_Not_Supported = 0xA122
const RC_MTP_Media_Session_Limit_Reached = 0xA171
const RC_MTP_No_More_Data = 0xA172
const RC_MTP_Undefined = 0xA800
const RC_MTP_Invalid_ObjectPropCode = 0xA801
const RC_MTP_Invalid_ObjectProp_Format = 0xA802
const RC_MTP_Invalid_ObjectProp_Value = 0xA803
const RC_MTP_Invalid_ObjectReference = 0xA804
const RC_MTP_Invalid_Dataset = 0xA806
const RC_MTP_Specification_By_Group_Unsupported = 0xA807
const RC_MTP_Specification_By_Depth_Unsupported = 0xA808
const RC_MTP_Object_Too_Large = 0xA809
const RC_MTP_ObjectProp_Not_Supported = 0xA80A

var RC_names = map[int]string{0x2000: "Undefined",
	0x2001: "OK",
	0x2002: "GeneralError",
	0x2003: "SessionNotOpen",
	0x2004: "InvalidTransactionID",
	0x2005: "OperationNotSupported",
	0x2006: "ParameterNotSupported",
	0x2007: "IncompleteTransfer",
	0x2008: "InvalidStorageId",
	0x2009: "InvalidObjectHandle",
	0x200A: "DevicePropNotSupported",
	0x200B: "InvalidObjectFormatCode",
	0x200C: "StoreFull",
	0x200D: "ObjectWriteProtected",
	0x200E: "StoreReadOnly",
	0x200F: "AccessDenied",
	0x2010: "NoThumbnailPresent",
	0x2011: "SelfTestFailed",
	0x2012: "PartialDeletion",
	0x2013: "StoreNotAvailable",
	0x2014: "SpecificationByFormatUnsupported",
	0x2015: "NoValidObjectInfo",
	0x2016: "InvalidCodeFormat",
	0x2017: "UnknownVendorCode",
	0x2018: "CaptureAlreadyTerminated",
	0x2019: "DeviceBusy",
	0x201A: "InvalidParentObject",
	0x201B: "InvalidDevicePropFormat",
	0x201C: "InvalidDevicePropValue",
	0x201D: "InvalidParameter",
	0x201E: "SessionAlreadyOpened",
	0x201F: "TransactionCanceled",
	0x2020: "SpecificationOfDestinationUnsupported",
	0x2021: "InvalidEnumHandle",
	0x2022: "NoStreamEnabled",
	0x2023: "InvalidDataSet",
	0xA121: "MTP_Invalid_WFC_Syntax",
	0xA122: "MTP_WFC_Version_Not_Supported",
	0xA171: "MTP_Media_Session_Limit_Reached",
	0xA172: "MTP_No_More_Data",
	0xA800: "MTP_Undefined",
	0xA801: "MTP_Invalid_ObjectPropCode",
	0xA802: "MTP_Invalid_ObjectProp_Format",
	0xA803: "MTP_Invalid_ObjectProp_Value",
	0xA804: "MTP_Invalid_ObjectReference",
	0xA806: "MTP_Invalid_Dataset",
	0xA807: "MTP_Specification_By_Group_Unsupported",
	0xA808: "MTP_Specification_By_Depth_Unsupported",
	0xA809: "MTP_Object_Too_Large",
	0xA80A: "MTP_ObjectProp_Not_Supported",
}

// storage
const ST_Undefined = 0x0000
const ST_FixedROM = 0x0001
const ST_RemovableROM = 0x0002
const ST_FixedRAM = 0x0003
const ST_RemovableRAM = 0x0004

var ST_names = map[int]string{0x0000: "Undefined",
	0x0001: "FixedROM",
	0x0002: "RemovableROM",
	0x0003: "FixedRAM",
	0x0004: "RemovableRAM",
}

const USB_CONTAINER_UNDEFINED = 0x0000
const USB_CONTAINER_COMMAND = 0x0001
const USB_CONTAINER_DATA = 0x0002
const USB_CONTAINER_RESPONSE = 0x0003
const USB_CONTAINER_EVENT = 0x0004
const USB_BULK_HS_MAX_PACKET_LEN_READ = 512
const USB_BULK_HS_MAX_PACKET_LEN_WRITE = 512

var USB_names = map[int]string{0x0000: "CONTAINER_UNDEFINED",
	0x0001: "CONTAINER_COMMAND",
	0x0002: "CONTAINER_DATA",
	0x0003: "CONTAINER_RESPONSE",
	0x0004: "CONTAINER_EVENT",
	512:    "BULK_HS_MAX_PACKET_LEN_READ",
}

const VENDOR_EASTMAN_KODAK = 0x00000001
const VENDOR_SEIKO_EPSON = 0x00000002
const VENDOR_AGILENT = 0x00000003
const VENDOR_POLAROID = 0x00000004
const VENDOR_AGFA_GEVAERT = 0x00000005
const VENDOR_MICROSOFT = 0x00000006
const VENDOR_EQUINOX = 0x00000007
const VENDOR_VIEWQUEST = 0x00000008
const VENDOR_STMICROELECTRONICS = 0x00000009
const VENDOR_NIKON = 0x0000000A
const VENDOR_CANON = 0x0000000B
const VENDOR_FOTONATION = 0x0000000C
const VENDOR_PENTAX = 0x0000000D
const VENDOR_FUJI = 0x0000000E

var VENDOR_names = map[int]string{0x00000001: "EASTMAN_KODAK",
	0x00000002: "SEIKO_EPSON",
	0x00000003: "AGILENT",
	0x00000004: "POLAROID",
	0x00000005: "AGFA_GEVAERT",
	0x00000006: "MICROSOFT",
	0x00000007: "EQUINOX",
	0x00000008: "VIEWQUEST",
	0x00000009: "STMICROELECTRONICS",
	0x0000000A: "NIKON",
	0x0000000B: "CANON",
	0x0000000C: "FOTONATION",
	0x0000000D: "PENTAX",
	0x0000000E: "FUJI",
}
