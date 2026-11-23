package win32

import (
	"syscall"
	"unsafe"
)

type (
	HWINEVENTHOOK     = uintptr
	HUIANODE          = uintptr
	HUIAPATTERNOBJECT = uintptr
	HUIATEXTRANGE     = uintptr
	HUIAEVENT         = uintptr
)

const (
	ANRUS_PRIORITY_AUDIO_DYNAMIC_DUCK                    uint32  = 0x10
	MSAA_MENU_SIG                                        int32   = -1441927155
	DISPID_ACC_PARENT                                    int32   = -5000
	DISPID_ACC_CHILDCOUNT                                int32   = -5001
	DISPID_ACC_CHILD                                     int32   = -5002
	DISPID_ACC_NAME                                      int32   = -5003
	DISPID_ACC_VALUE                                     int32   = -5004
	DISPID_ACC_DESCRIPTION                               int32   = -5005
	DISPID_ACC_ROLE                                      int32   = -5006
	DISPID_ACC_STATE                                     int32   = -5007
	DISPID_ACC_HELP                                      int32   = -5008
	DISPID_ACC_HELPTOPIC                                 int32   = -5009
	DISPID_ACC_KEYBOARDSHORTCUT                          int32   = -5010
	DISPID_ACC_FOCUS                                     int32   = -5011
	DISPID_ACC_SELECTION                                 int32   = -5012
	DISPID_ACC_DEFAULTACTION                             int32   = -5013
	DISPID_ACC_SELECT                                    int32   = -5014
	DISPID_ACC_LOCATION                                  int32   = -5015
	DISPID_ACC_NAVIGATE                                  int32   = -5016
	DISPID_ACC_HITTEST                                   int32   = -5017
	DISPID_ACC_DODEFAULTACTION                           int32   = -5018
	NAVDIR_MIN                                           uint32  = 0x0
	NAVDIR_UP                                            uint32  = 0x1
	NAVDIR_DOWN                                          uint32  = 0x2
	NAVDIR_LEFT                                          uint32  = 0x3
	NAVDIR_RIGHT                                         uint32  = 0x4
	NAVDIR_NEXT                                          uint32  = 0x5
	NAVDIR_PREVIOUS                                      uint32  = 0x6
	NAVDIR_FIRSTCHILD                                    uint32  = 0x7
	NAVDIR_LASTCHILD                                     uint32  = 0x8
	NAVDIR_MAX                                           uint32  = 0x9
	SELFLAG_NONE                                         uint32  = 0x0
	SELFLAG_TAKEFOCUS                                    uint32  = 0x1
	SELFLAG_TAKESELECTION                                uint32  = 0x2
	SELFLAG_EXTENDSELECTION                              uint32  = 0x4
	SELFLAG_ADDSELECTION                                 uint32  = 0x8
	SELFLAG_REMOVESELECTION                              uint32  = 0x10
	SELFLAG_VALID                                        uint32  = 0x1f
	STATE_SYSTEM_NORMAL                                  uint32  = 0x0
	STATE_SYSTEM_HASPOPUP                                uint32  = 0x40000000
	ROLE_SYSTEM_TITLEBAR                                 uint32  = 0x1
	ROLE_SYSTEM_MENUBAR                                  uint32  = 0x2
	ROLE_SYSTEM_SCROLLBAR                                uint32  = 0x3
	ROLE_SYSTEM_GRIP                                     uint32  = 0x4
	ROLE_SYSTEM_SOUND                                    uint32  = 0x5
	ROLE_SYSTEM_CURSOR                                   uint32  = 0x6
	ROLE_SYSTEM_CARET                                    uint32  = 0x7
	ROLE_SYSTEM_ALERT                                    uint32  = 0x8
	ROLE_SYSTEM_WINDOW                                   uint32  = 0x9
	ROLE_SYSTEM_CLIENT                                   uint32  = 0xa
	ROLE_SYSTEM_MENUPOPUP                                uint32  = 0xb
	ROLE_SYSTEM_MENUITEM                                 uint32  = 0xc
	ROLE_SYSTEM_TOOLTIP                                  uint32  = 0xd
	ROLE_SYSTEM_APPLICATION                              uint32  = 0xe
	ROLE_SYSTEM_DOCUMENT                                 uint32  = 0xf
	ROLE_SYSTEM_PANE                                     uint32  = 0x10
	ROLE_SYSTEM_CHART                                    uint32  = 0x11
	ROLE_SYSTEM_DIALOG                                   uint32  = 0x12
	ROLE_SYSTEM_BORDER                                   uint32  = 0x13
	ROLE_SYSTEM_GROUPING                                 uint32  = 0x14
	ROLE_SYSTEM_SEPARATOR                                uint32  = 0x15
	ROLE_SYSTEM_TOOLBAR                                  uint32  = 0x16
	ROLE_SYSTEM_STATUSBAR                                uint32  = 0x17
	ROLE_SYSTEM_TABLE                                    uint32  = 0x18
	ROLE_SYSTEM_COLUMNHEADER                             uint32  = 0x19
	ROLE_SYSTEM_ROWHEADER                                uint32  = 0x1a
	ROLE_SYSTEM_COLUMN                                   uint32  = 0x1b
	ROLE_SYSTEM_ROW                                      uint32  = 0x1c
	ROLE_SYSTEM_CELL                                     uint32  = 0x1d
	ROLE_SYSTEM_LINK                                     uint32  = 0x1e
	ROLE_SYSTEM_HELPBALLOON                              uint32  = 0x1f
	ROLE_SYSTEM_CHARACTER                                uint32  = 0x20
	ROLE_SYSTEM_LIST                                     uint32  = 0x21
	ROLE_SYSTEM_LISTITEM                                 uint32  = 0x22
	ROLE_SYSTEM_OUTLINE                                  uint32  = 0x23
	ROLE_SYSTEM_OUTLINEITEM                              uint32  = 0x24
	ROLE_SYSTEM_PAGETAB                                  uint32  = 0x25
	ROLE_SYSTEM_PROPERTYPAGE                             uint32  = 0x26
	ROLE_SYSTEM_INDICATOR                                uint32  = 0x27
	ROLE_SYSTEM_GRAPHIC                                  uint32  = 0x28
	ROLE_SYSTEM_STATICTEXT                               uint32  = 0x29
	ROLE_SYSTEM_TEXT                                     uint32  = 0x2a
	ROLE_SYSTEM_PUSHBUTTON                               uint32  = 0x2b
	ROLE_SYSTEM_CHECKBUTTON                              uint32  = 0x2c
	ROLE_SYSTEM_RADIOBUTTON                              uint32  = 0x2d
	ROLE_SYSTEM_COMBOBOX                                 uint32  = 0x2e
	ROLE_SYSTEM_DROPLIST                                 uint32  = 0x2f
	ROLE_SYSTEM_PROGRESSBAR                              uint32  = 0x30
	ROLE_SYSTEM_DIAL                                     uint32  = 0x31
	ROLE_SYSTEM_HOTKEYFIELD                              uint32  = 0x32
	ROLE_SYSTEM_SLIDER                                   uint32  = 0x33
	ROLE_SYSTEM_SPINBUTTON                               uint32  = 0x34
	ROLE_SYSTEM_DIAGRAM                                  uint32  = 0x35
	ROLE_SYSTEM_ANIMATION                                uint32  = 0x36
	ROLE_SYSTEM_EQUATION                                 uint32  = 0x37
	ROLE_SYSTEM_BUTTONDROPDOWN                           uint32  = 0x38
	ROLE_SYSTEM_BUTTONMENU                               uint32  = 0x39
	ROLE_SYSTEM_BUTTONDROPDOWNGRID                       uint32  = 0x3a
	ROLE_SYSTEM_WHITESPACE                               uint32  = 0x3b
	ROLE_SYSTEM_PAGETABLIST                              uint32  = 0x3c
	ROLE_SYSTEM_CLOCK                                    uint32  = 0x3d
	ROLE_SYSTEM_SPLITBUTTON                              uint32  = 0x3e
	ROLE_SYSTEM_IPADDRESS                                uint32  = 0x3f
	ROLE_SYSTEM_OUTLINEBUTTON                            uint32  = 0x40
	UIA_E_ELEMENTNOTENABLED                              uint32  = 0x80040200
	UIA_E_ELEMENTNOTAVAILABLE                            uint32  = 0x80040201
	UIA_E_NOCLICKABLEPOINT                               uint32  = 0x80040202
	UIA_E_PROXYASSEMBLYNOTLOADED                         uint32  = 0x80040203
	UIA_E_NOTSUPPORTED                                   uint32  = 0x80040204
	UIA_E_INVALIDOPERATION                               uint32  = 0x80131509
	UIA_E_TIMEOUT                                        uint32  = 0x80131505
	UiaAppendRuntimeId                                   uint32  = 0x3
	UiaRootObjectId                                      int32   = -25
	UIA_IAFP_DEFAULT                                     uint32  = 0x0
	UIA_IAFP_UNWRAP_BRIDGE                               uint32  = 0x1
	UIA_PFIA_DEFAULT                                     uint32  = 0x0
	UIA_PFIA_UNWRAP_BRIDGE                               uint32  = 0x1
	UIA_ScrollPatternNoScroll                            float64 = -1
	UIA_InvokePatternId                                  int32   = 10000
	UIA_SelectionPatternId                               int32   = 10001
	UIA_ValuePatternId                                   int32   = 10002
	UIA_RangeValuePatternId                              int32   = 10003
	UIA_ScrollPatternId                                  int32   = 10004
	UIA_ExpandCollapsePatternId                          int32   = 10005
	UIA_GridPatternId                                    int32   = 10006
	UIA_GridItemPatternId                                int32   = 10007
	UIA_MultipleViewPatternId                            int32   = 10008
	UIA_WindowPatternId                                  int32   = 10009
	UIA_SelectionItemPatternId                           int32   = 10010
	UIA_DockPatternId                                    int32   = 10011
	UIA_TablePatternId                                   int32   = 10012
	UIA_TableItemPatternId                               int32   = 10013
	UIA_TextPatternId                                    int32   = 10014
	UIA_TogglePatternId                                  int32   = 10015
	UIA_TransformPatternId                               int32   = 10016
	UIA_ScrollItemPatternId                              int32   = 10017
	UIA_LegacyIAccessiblePatternId                       int32   = 10018
	UIA_ItemContainerPatternId                           int32   = 10019
	UIA_VirtualizedItemPatternId                         int32   = 10020
	UIA_SynchronizedInputPatternId                       int32   = 10021
	UIA_ObjectModelPatternId                             int32   = 10022
	UIA_AnnotationPatternId                              int32   = 10023
	UIA_TextPattern2Id                                   int32   = 10024
	UIA_StylesPatternId                                  int32   = 10025
	UIA_SpreadsheetPatternId                             int32   = 10026
	UIA_SpreadsheetItemPatternId                         int32   = 10027
	UIA_TransformPattern2Id                              int32   = 10028
	UIA_TextChildPatternId                               int32   = 10029
	UIA_DragPatternId                                    int32   = 10030
	UIA_DropTargetPatternId                              int32   = 10031
	UIA_TextEditPatternId                                int32   = 10032
	UIA_CustomNavigationPatternId                        int32   = 10033
	UIA_SelectionPattern2Id                              int32   = 10034
	UIA_ToolTipOpenedEventId                             int32   = 20000
	UIA_ToolTipClosedEventId                             int32   = 20001
	UIA_StructureChangedEventId                          int32   = 20002
	UIA_MenuOpenedEventId                                int32   = 20003
	UIA_AutomationPropertyChangedEventId                 int32   = 20004
	UIA_AutomationFocusChangedEventId                    int32   = 20005
	UIA_AsyncContentLoadedEventId                        int32   = 20006
	UIA_MenuClosedEventId                                int32   = 20007
	UIA_LayoutInvalidatedEventId                         int32   = 20008
	UIA_Invoke_InvokedEventId                            int32   = 20009
	UIA_SelectionItem_ElementAddedToSelectionEventId     int32   = 20010
	UIA_SelectionItem_ElementRemovedFromSelectionEventId int32   = 20011
	UIA_SelectionItem_ElementSelectedEventId             int32   = 20012
	UIA_Selection_InvalidatedEventId                     int32   = 20013
	UIA_Text_TextSelectionChangedEventId                 int32   = 20014
	UIA_Text_TextChangedEventId                          int32   = 20015
	UIA_Window_WindowOpenedEventId                       int32   = 20016
	UIA_Window_WindowClosedEventId                       int32   = 20017
	UIA_MenuModeStartEventId                             int32   = 20018
	UIA_MenuModeEndEventId                               int32   = 20019
	UIA_InputReachedTargetEventId                        int32   = 20020
	UIA_InputReachedOtherElementEventId                  int32   = 20021
	UIA_InputDiscardedEventId                            int32   = 20022
	UIA_SystemAlertEventId                               int32   = 20023
	UIA_LiveRegionChangedEventId                         int32   = 20024
	UIA_HostedFragmentRootsInvalidatedEventId            int32   = 20025
	UIA_Drag_DragStartEventId                            int32   = 20026
	UIA_Drag_DragCancelEventId                           int32   = 20027
	UIA_Drag_DragCompleteEventId                         int32   = 20028
	UIA_DropTarget_DragEnterEventId                      int32   = 20029
	UIA_DropTarget_DragLeaveEventId                      int32   = 20030
	UIA_DropTarget_DroppedEventId                        int32   = 20031
	UIA_TextEdit_TextChangedEventId                      int32   = 20032
	UIA_TextEdit_ConversionTargetChangedEventId          int32   = 20033
	UIA_ChangesEventId                                   int32   = 20034
	UIA_NotificationEventId                              int32   = 20035
	UIA_ActiveTextPositionChangedEventId                 int32   = 20036
	UIA_RuntimeIdPropertyId                              int32   = 30000
	UIA_BoundingRectanglePropertyId                      int32   = 30001
	UIA_ProcessIdPropertyId                              int32   = 30002
	UIA_ControlTypePropertyId                            int32   = 30003
	UIA_LocalizedControlTypePropertyId                   int32   = 30004
	UIA_NamePropertyId                                   int32   = 30005
	UIA_AcceleratorKeyPropertyId                         int32   = 30006
	UIA_AccessKeyPropertyId                              int32   = 30007
	UIA_HasKeyboardFocusPropertyId                       int32   = 30008
	UIA_IsKeyboardFocusablePropertyId                    int32   = 30009
	UIA_IsEnabledPropertyId                              int32   = 30010
	UIA_AutomationIdPropertyId                           int32   = 30011
	UIA_ClassNamePropertyId                              int32   = 30012
	UIA_HelpTextPropertyId                               int32   = 30013
	UIA_ClickablePointPropertyId                         int32   = 30014
	UIA_CulturePropertyId                                int32   = 30015
	UIA_IsControlElementPropertyId                       int32   = 30016
	UIA_IsContentElementPropertyId                       int32   = 30017
	UIA_LabeledByPropertyId                              int32   = 30018
	UIA_IsPasswordPropertyId                             int32   = 30019
	UIA_NativeWindowHandlePropertyId                     int32   = 30020
	UIA_ItemTypePropertyId                               int32   = 30021
	UIA_IsOffscreenPropertyId                            int32   = 30022
	UIA_OrientationPropertyId                            int32   = 30023
	UIA_FrameworkIdPropertyId                            int32   = 30024
	UIA_IsRequiredForFormPropertyId                      int32   = 30025
	UIA_ItemStatusPropertyId                             int32   = 30026
	UIA_IsDockPatternAvailablePropertyId                 int32   = 30027
	UIA_IsExpandCollapsePatternAvailablePropertyId       int32   = 30028
	UIA_IsGridItemPatternAvailablePropertyId             int32   = 30029
	UIA_IsGridPatternAvailablePropertyId                 int32   = 30030
	UIA_IsInvokePatternAvailablePropertyId               int32   = 30031
	UIA_IsMultipleViewPatternAvailablePropertyId         int32   = 30032
	UIA_IsRangeValuePatternAvailablePropertyId           int32   = 30033
	UIA_IsScrollPatternAvailablePropertyId               int32   = 30034
	UIA_IsScrollItemPatternAvailablePropertyId           int32   = 30035
	UIA_IsSelectionItemPatternAvailablePropertyId        int32   = 30036
	UIA_IsSelectionPatternAvailablePropertyId            int32   = 30037
	UIA_IsTablePatternAvailablePropertyId                int32   = 30038
	UIA_IsTableItemPatternAvailablePropertyId            int32   = 30039
	UIA_IsTextPatternAvailablePropertyId                 int32   = 30040
	UIA_IsTogglePatternAvailablePropertyId               int32   = 30041
	UIA_IsTransformPatternAvailablePropertyId            int32   = 30042
	UIA_IsValuePatternAvailablePropertyId                int32   = 30043
	UIA_IsWindowPatternAvailablePropertyId               int32   = 30044
	UIA_ValueValuePropertyId                             int32   = 30045
	UIA_ValueIsReadOnlyPropertyId                        int32   = 30046
	UIA_RangeValueValuePropertyId                        int32   = 30047
	UIA_RangeValueIsReadOnlyPropertyId                   int32   = 30048
	UIA_RangeValueMinimumPropertyId                      int32   = 30049
	UIA_RangeValueMaximumPropertyId                      int32   = 30050
	UIA_RangeValueLargeChangePropertyId                  int32   = 30051
	UIA_RangeValueSmallChangePropertyId                  int32   = 30052
	UIA_ScrollHorizontalScrollPercentPropertyId          int32   = 30053
	UIA_ScrollHorizontalViewSizePropertyId               int32   = 30054
	UIA_ScrollVerticalScrollPercentPropertyId            int32   = 30055
	UIA_ScrollVerticalViewSizePropertyId                 int32   = 30056
	UIA_ScrollHorizontallyScrollablePropertyId           int32   = 30057
	UIA_ScrollVerticallyScrollablePropertyId             int32   = 30058
	UIA_SelectionSelectionPropertyId                     int32   = 30059
	UIA_SelectionCanSelectMultiplePropertyId             int32   = 30060
	UIA_SelectionIsSelectionRequiredPropertyId           int32   = 30061
	UIA_GridRowCountPropertyId                           int32   = 30062
	UIA_GridColumnCountPropertyId                        int32   = 30063
	UIA_GridItemRowPropertyId                            int32   = 30064
	UIA_GridItemColumnPropertyId                         int32   = 30065
	UIA_GridItemRowSpanPropertyId                        int32   = 30066
	UIA_GridItemColumnSpanPropertyId                     int32   = 30067
	UIA_GridItemContainingGridPropertyId                 int32   = 30068
	UIA_DockDockPositionPropertyId                       int32   = 30069
	UIA_ExpandCollapseExpandCollapseStatePropertyId      int32   = 30070
	UIA_MultipleViewCurrentViewPropertyId                int32   = 30071
	UIA_MultipleViewSupportedViewsPropertyId             int32   = 30072
	UIA_WindowCanMaximizePropertyId                      int32   = 30073
	UIA_WindowCanMinimizePropertyId                      int32   = 30074
	UIA_WindowWindowVisualStatePropertyId                int32   = 30075
	UIA_WindowWindowInteractionStatePropertyId           int32   = 30076
	UIA_WindowIsModalPropertyId                          int32   = 30077
	UIA_WindowIsTopmostPropertyId                        int32   = 30078
	UIA_SelectionItemIsSelectedPropertyId                int32   = 30079
	UIA_SelectionItemSelectionContainerPropertyId        int32   = 30080
	UIA_TableRowHeadersPropertyId                        int32   = 30081
	UIA_TableColumnHeadersPropertyId                     int32   = 30082
	UIA_TableRowOrColumnMajorPropertyId                  int32   = 30083
	UIA_TableItemRowHeaderItemsPropertyId                int32   = 30084
	UIA_TableItemColumnHeaderItemsPropertyId             int32   = 30085
	UIA_ToggleToggleStatePropertyId                      int32   = 30086
	UIA_TransformCanMovePropertyId                       int32   = 30087
	UIA_TransformCanResizePropertyId                     int32   = 30088
	UIA_TransformCanRotatePropertyId                     int32   = 30089
	UIA_IsLegacyIAccessiblePatternAvailablePropertyId    int32   = 30090
	UIA_LegacyIAccessibleChildIdPropertyId               int32   = 30091
	UIA_LegacyIAccessibleNamePropertyId                  int32   = 30092
	UIA_LegacyIAccessibleValuePropertyId                 int32   = 30093
	UIA_LegacyIAccessibleDescriptionPropertyId           int32   = 30094
	UIA_LegacyIAccessibleRolePropertyId                  int32   = 30095
	UIA_LegacyIAccessibleStatePropertyId                 int32   = 30096
	UIA_LegacyIAccessibleHelpPropertyId                  int32   = 30097
	UIA_LegacyIAccessibleKeyboardShortcutPropertyId      int32   = 30098
	UIA_LegacyIAccessibleSelectionPropertyId             int32   = 30099
	UIA_LegacyIAccessibleDefaultActionPropertyId         int32   = 30100
	UIA_AriaRolePropertyId                               int32   = 30101
	UIA_AriaPropertiesPropertyId                         int32   = 30102
	UIA_IsDataValidForFormPropertyId                     int32   = 30103
	UIA_ControllerForPropertyId                          int32   = 30104
	UIA_DescribedByPropertyId                            int32   = 30105
	UIA_FlowsToPropertyId                                int32   = 30106
	UIA_ProviderDescriptionPropertyId                    int32   = 30107
	UIA_IsItemContainerPatternAvailablePropertyId        int32   = 30108
	UIA_IsVirtualizedItemPatternAvailablePropertyId      int32   = 30109
	UIA_IsSynchronizedInputPatternAvailablePropertyId    int32   = 30110
	UIA_OptimizeForVisualContentPropertyId               int32   = 30111
	UIA_IsObjectModelPatternAvailablePropertyId          int32   = 30112
	UIA_AnnotationAnnotationTypeIdPropertyId             int32   = 30113
	UIA_AnnotationAnnotationTypeNamePropertyId           int32   = 30114
	UIA_AnnotationAuthorPropertyId                       int32   = 30115
	UIA_AnnotationDateTimePropertyId                     int32   = 30116
	UIA_AnnotationTargetPropertyId                       int32   = 30117
	UIA_IsAnnotationPatternAvailablePropertyId           int32   = 30118
	UIA_IsTextPattern2AvailablePropertyId                int32   = 30119
	UIA_StylesStyleIdPropertyId                          int32   = 30120
	UIA_StylesStyleNamePropertyId                        int32   = 30121
	UIA_StylesFillColorPropertyId                        int32   = 30122
	UIA_StylesFillPatternStylePropertyId                 int32   = 30123
	UIA_StylesShapePropertyId                            int32   = 30124
	UIA_StylesFillPatternColorPropertyId                 int32   = 30125
	UIA_StylesExtendedPropertiesPropertyId               int32   = 30126
	UIA_IsStylesPatternAvailablePropertyId               int32   = 30127
	UIA_IsSpreadsheetPatternAvailablePropertyId          int32   = 30128
	UIA_SpreadsheetItemFormulaPropertyId                 int32   = 30129
	UIA_SpreadsheetItemAnnotationObjectsPropertyId       int32   = 30130
	UIA_SpreadsheetItemAnnotationTypesPropertyId         int32   = 30131
	UIA_IsSpreadsheetItemPatternAvailablePropertyId      int32   = 30132
	UIA_Transform2CanZoomPropertyId                      int32   = 30133
	UIA_IsTransformPattern2AvailablePropertyId           int32   = 30134
	UIA_LiveSettingPropertyId                            int32   = 30135
	UIA_IsTextChildPatternAvailablePropertyId            int32   = 30136
	UIA_IsDragPatternAvailablePropertyId                 int32   = 30137
	UIA_DragIsGrabbedPropertyId                          int32   = 30138
	UIA_DragDropEffectPropertyId                         int32   = 30139
	UIA_DragDropEffectsPropertyId                        int32   = 30140
	UIA_IsDropTargetPatternAvailablePropertyId           int32   = 30141
	UIA_DropTargetDropTargetEffectPropertyId             int32   = 30142
	UIA_DropTargetDropTargetEffectsPropertyId            int32   = 30143
	UIA_DragGrabbedItemsPropertyId                       int32   = 30144
	UIA_Transform2ZoomLevelPropertyId                    int32   = 30145
	UIA_Transform2ZoomMinimumPropertyId                  int32   = 30146
	UIA_Transform2ZoomMaximumPropertyId                  int32   = 30147
	UIA_FlowsFromPropertyId                              int32   = 30148
	UIA_IsTextEditPatternAvailablePropertyId             int32   = 30149
	UIA_IsPeripheralPropertyId                           int32   = 30150
	UIA_IsCustomNavigationPatternAvailablePropertyId     int32   = 30151
	UIA_PositionInSetPropertyId                          int32   = 30152
	UIA_SizeOfSetPropertyId                              int32   = 30153
	UIA_LevelPropertyId                                  int32   = 30154
	UIA_AnnotationTypesPropertyId                        int32   = 30155
	UIA_AnnotationObjectsPropertyId                      int32   = 30156
	UIA_LandmarkTypePropertyId                           int32   = 30157
	UIA_LocalizedLandmarkTypePropertyId                  int32   = 30158
	UIA_FullDescriptionPropertyId                        int32   = 30159
	UIA_FillColorPropertyId                              int32   = 30160
	UIA_OutlineColorPropertyId                           int32   = 30161
	UIA_FillTypePropertyId                               int32   = 30162
	UIA_VisualEffectsPropertyId                          int32   = 30163
	UIA_OutlineThicknessPropertyId                       int32   = 30164
	UIA_CenterPointPropertyId                            int32   = 30165
	UIA_RotationPropertyId                               int32   = 30166
	UIA_SizePropertyId                                   int32   = 30167
	UIA_IsSelectionPattern2AvailablePropertyId           int32   = 30168
	UIA_Selection2FirstSelectedItemPropertyId            int32   = 30169
	UIA_Selection2LastSelectedItemPropertyId             int32   = 30170
	UIA_Selection2CurrentSelectedItemPropertyId          int32   = 30171
	UIA_Selection2ItemCountPropertyId                    int32   = 30172
	UIA_HeadingLevelPropertyId                           int32   = 30173
	UIA_IsDialogPropertyId                               int32   = 30174
	UIA_AnimationStyleAttributeId                        int32   = 40000
	UIA_BackgroundColorAttributeId                       int32   = 40001
	UIA_BulletStyleAttributeId                           int32   = 40002
	UIA_CapStyleAttributeId                              int32   = 40003
	UIA_CultureAttributeId                               int32   = 40004
	UIA_FontNameAttributeId                              int32   = 40005
	UIA_FontSizeAttributeId                              int32   = 40006
	UIA_FontWeightAttributeId                            int32   = 40007
	UIA_ForegroundColorAttributeId                       int32   = 40008
	UIA_HorizontalTextAlignmentAttributeId               int32   = 40009
	UIA_IndentationFirstLineAttributeId                  int32   = 40010
	UIA_IndentationLeadingAttributeId                    int32   = 40011
	UIA_IndentationTrailingAttributeId                   int32   = 40012
	UIA_IsHiddenAttributeId                              int32   = 40013
	UIA_IsItalicAttributeId                              int32   = 40014
	UIA_IsReadOnlyAttributeId                            int32   = 40015
	UIA_IsSubscriptAttributeId                           int32   = 40016
	UIA_IsSuperscriptAttributeId                         int32   = 40017
	UIA_MarginBottomAttributeId                          int32   = 40018
	UIA_MarginLeadingAttributeId                         int32   = 40019
	UIA_MarginTopAttributeId                             int32   = 40020
	UIA_MarginTrailingAttributeId                        int32   = 40021
	UIA_OutlineStylesAttributeId                         int32   = 40022
	UIA_OverlineColorAttributeId                         int32   = 40023
	UIA_OverlineStyleAttributeId                         int32   = 40024
	UIA_StrikethroughColorAttributeId                    int32   = 40025
	UIA_StrikethroughStyleAttributeId                    int32   = 40026
	UIA_TabsAttributeId                                  int32   = 40027
	UIA_TextFlowDirectionsAttributeId                    int32   = 40028
	UIA_UnderlineColorAttributeId                        int32   = 40029
	UIA_UnderlineStyleAttributeId                        int32   = 40030
	UIA_AnnotationTypesAttributeId                       int32   = 40031
	UIA_AnnotationObjectsAttributeId                     int32   = 40032
	UIA_StyleNameAttributeId                             int32   = 40033
	UIA_StyleIdAttributeId                               int32   = 40034
	UIA_LinkAttributeId                                  int32   = 40035
	UIA_IsActiveAttributeId                              int32   = 40036
	UIA_SelectionActiveEndAttributeId                    int32   = 40037
	UIA_CaretPositionAttributeId                         int32   = 40038
	UIA_CaretBidiModeAttributeId                         int32   = 40039
	UIA_LineSpacingAttributeId                           int32   = 40040
	UIA_BeforeParagraphSpacingAttributeId                int32   = 40041
	UIA_AfterParagraphSpacingAttributeId                 int32   = 40042
	UIA_SayAsInterpretAsAttributeId                      int32   = 40043
	UIA_ButtonControlTypeId                              int32   = 50000
	UIA_CalendarControlTypeId                            int32   = 50001
	UIA_CheckBoxControlTypeId                            int32   = 50002
	UIA_ComboBoxControlTypeId                            int32   = 50003
	UIA_EditControlTypeId                                int32   = 50004
	UIA_HyperlinkControlTypeId                           int32   = 50005
	UIA_ImageControlTypeId                               int32   = 50006
	UIA_ListItemControlTypeId                            int32   = 50007
	UIA_ListControlTypeId                                int32   = 50008
	UIA_MenuControlTypeId                                int32   = 50009
	UIA_MenuBarControlTypeId                             int32   = 50010
	UIA_MenuItemControlTypeId                            int32   = 50011
	UIA_ProgressBarControlTypeId                         int32   = 50012
	UIA_RadioButtonControlTypeId                         int32   = 50013
	UIA_ScrollBarControlTypeId                           int32   = 50014
	UIA_SliderControlTypeId                              int32   = 50015
	UIA_SpinnerControlTypeId                             int32   = 50016
	UIA_StatusBarControlTypeId                           int32   = 50017
	UIA_TabControlTypeId                                 int32   = 50018
	UIA_TabItemControlTypeId                             int32   = 50019
	UIA_TextControlTypeId                                int32   = 50020
	UIA_ToolBarControlTypeId                             int32   = 50021
	UIA_ToolTipControlTypeId                             int32   = 50022
	UIA_TreeControlTypeId                                int32   = 50023
	UIA_TreeItemControlTypeId                            int32   = 50024
	UIA_CustomControlTypeId                              int32   = 50025
	UIA_GroupControlTypeId                               int32   = 50026
	UIA_ThumbControlTypeId                               int32   = 50027
	UIA_DataGridControlTypeId                            int32   = 50028
	UIA_DataItemControlTypeId                            int32   = 50029
	UIA_DocumentControlTypeId                            int32   = 50030
	UIA_SplitButtonControlTypeId                         int32   = 50031
	UIA_WindowControlTypeId                              int32   = 50032
	UIA_PaneControlTypeId                                int32   = 50033
	UIA_HeaderControlTypeId                              int32   = 50034
	UIA_HeaderItemControlTypeId                          int32   = 50035
	UIA_TableControlTypeId                               int32   = 50036
	UIA_TitleBarControlTypeId                            int32   = 50037
	UIA_SeparatorControlTypeId                           int32   = 50038
	UIA_SemanticZoomControlTypeId                        int32   = 50039
	UIA_AppBarControlTypeId                              int32   = 50040
	AnnotationType_Unknown                               int32   = 60000
	AnnotationType_SpellingError                         int32   = 60001
	AnnotationType_GrammarError                          int32   = 60002
	AnnotationType_Comment                               int32   = 60003
	AnnotationType_FormulaError                          int32   = 60004
	AnnotationType_TrackChanges                          int32   = 60005
	AnnotationType_Header                                int32   = 60006
	AnnotationType_Footer                                int32   = 60007
	AnnotationType_Highlighted                           int32   = 60008
	AnnotationType_Endnote                               int32   = 60009
	AnnotationType_Footnote                              int32   = 60010
	AnnotationType_InsertionChange                       int32   = 60011
	AnnotationType_DeletionChange                        int32   = 60012
	AnnotationType_MoveChange                            int32   = 60013
	AnnotationType_FormatChange                          int32   = 60014
	AnnotationType_UnsyncedChange                        int32   = 60015
	AnnotationType_EditingLockedChange                   int32   = 60016
	AnnotationType_ExternalChange                        int32   = 60017
	AnnotationType_ConflictingChange                     int32   = 60018
	AnnotationType_Author                                int32   = 60019
	AnnotationType_AdvancedProofingIssue                 int32   = 60020
	AnnotationType_DataValidationError                   int32   = 60021
	AnnotationType_CircularReferenceError                int32   = 60022
	AnnotationType_Mathematics                           int32   = 60023
	AnnotationType_Sensitive                             int32   = 60024
	StyleId_Custom                                       int32   = 70000
	StyleId_Heading1                                     int32   = 70001
	StyleId_Heading2                                     int32   = 70002
	StyleId_Heading3                                     int32   = 70003
	StyleId_Heading4                                     int32   = 70004
	StyleId_Heading5                                     int32   = 70005
	StyleId_Heading6                                     int32   = 70006
	StyleId_Heading7                                     int32   = 70007
	StyleId_Heading8                                     int32   = 70008
	StyleId_Heading9                                     int32   = 70009
	StyleId_Title                                        int32   = 70010
	StyleId_Subtitle                                     int32   = 70011
	StyleId_Normal                                       int32   = 70012
	StyleId_Emphasis                                     int32   = 70013
	StyleId_Quote                                        int32   = 70014
	StyleId_BulletedList                                 int32   = 70015
	StyleId_NumberedList                                 int32   = 70016
	UIA_CustomLandmarkTypeId                             int32   = 80000
	UIA_FormLandmarkTypeId                               int32   = 80001
	UIA_MainLandmarkTypeId                               int32   = 80002
	UIA_NavigationLandmarkTypeId                         int32   = 80003
	UIA_SearchLandmarkTypeId                             int32   = 80004
	HeadingLevel_None                                    int32   = 80050
	HeadingLevel1                                        int32   = 80051
	HeadingLevel2                                        int32   = 80052
	HeadingLevel3                                        int32   = 80053
	HeadingLevel4                                        int32   = 80054
	HeadingLevel5                                        int32   = 80055
	HeadingLevel6                                        int32   = 80056
	HeadingLevel7                                        int32   = 80057
	HeadingLevel8                                        int32   = 80058
	HeadingLevel9                                        int32   = 80059
	UIA_SummaryChangeId                                  int32   = 90000
	UIA_SayAsInterpretAsMetadataId                       int32   = 100000
)

var (
	LIBID_Accessibility = syscall.GUID{0x1EA4DBF0, 0x3C3B, 0x11CF,
		[8]byte{0x81, 0x0C, 0x00, 0xAA, 0x00, 0x38, 0x9B, 0x71}}

	CLSID_AccPropServices = syscall.GUID{0xB5F8350B, 0x0548, 0x48B1,
		[8]byte{0xA6, 0xEE, 0x88, 0xBD, 0x00, 0xB4, 0xA5, 0xE7}}

	IIS_IsOleaccProxy = syscall.GUID{0x902697FA, 0x80E4, 0x4560,
		[8]byte{0x80, 0x2A, 0xA1, 0x3F, 0x22, 0xA6, 0x47, 0x09}}

	IIS_ControlAccessible = syscall.GUID{0x38C682A6, 0x9731, 0x43F2,
		[8]byte{0x9F, 0xAE, 0xE9, 0x01, 0xE6, 0x41, 0xB1, 0x01}}

	PROPID_ACC_NAME = syscall.GUID{0x608D3DF8, 0x8128, 0x4AA7,
		[8]byte{0xA4, 0x28, 0xF5, 0x5E, 0x49, 0x26, 0x72, 0x91}}

	PROPID_ACC_VALUE = syscall.GUID{0x123FE443, 0x211A, 0x4615,
		[8]byte{0x95, 0x27, 0xC4, 0x5A, 0x7E, 0x93, 0x71, 0x7A}}

	PROPID_ACC_DESCRIPTION = syscall.GUID{0x4D48DFE4, 0xBD3F, 0x491F,
		[8]byte{0xA6, 0x48, 0x49, 0x2D, 0x6F, 0x20, 0xC5, 0x88}}

	PROPID_ACC_ROLE = syscall.GUID{0xCB905FF2, 0x7BD1, 0x4C05,
		[8]byte{0xB3, 0xC8, 0xE6, 0xC2, 0x41, 0x36, 0x4D, 0x70}}

	PROPID_ACC_STATE = syscall.GUID{0xA8D4D5B0, 0x0A21, 0x42D0,
		[8]byte{0xA5, 0xC0, 0x51, 0x4E, 0x98, 0x4F, 0x45, 0x7B}}

	PROPID_ACC_HELP = syscall.GUID{0xC831E11F, 0x44DB, 0x4A99,
		[8]byte{0x97, 0x68, 0xCB, 0x8F, 0x97, 0x8B, 0x72, 0x31}}

	PROPID_ACC_KEYBOARDSHORTCUT = syscall.GUID{0x7D9BCEEE, 0x7D1E, 0x4979,
		[8]byte{0x93, 0x82, 0x51, 0x80, 0xF4, 0x17, 0x2C, 0x34}}

	PROPID_ACC_DEFAULTACTION = syscall.GUID{0x180C072B, 0xC27F, 0x43C7,
		[8]byte{0x99, 0x22, 0xF6, 0x35, 0x62, 0xA4, 0x63, 0x2B}}

	PROPID_ACC_HELPTOPIC = syscall.GUID{0x787D1379, 0x8EDE, 0x440B,
		[8]byte{0x8A, 0xEC, 0x11, 0xF7, 0xBF, 0x90, 0x30, 0xB3}}

	PROPID_ACC_FOCUS = syscall.GUID{0x6EB335DF, 0x1C29, 0x4127,
		[8]byte{0xB1, 0x2C, 0xDE, 0xE9, 0xFD, 0x15, 0x7F, 0x2B}}

	PROPID_ACC_SELECTION = syscall.GUID{0xB99D073C, 0xD731, 0x405B,
		[8]byte{0x90, 0x61, 0xD9, 0x5E, 0x8F, 0x84, 0x29, 0x84}}

	PROPID_ACC_PARENT = syscall.GUID{0x474C22B6, 0xFFC2, 0x467A,
		[8]byte{0xB1, 0xB5, 0xE9, 0x58, 0xB4, 0x65, 0x73, 0x30}}

	PROPID_ACC_NAV_UP = syscall.GUID{0x016E1A2B, 0x1A4E, 0x4767,
		[8]byte{0x86, 0x12, 0x33, 0x86, 0xF6, 0x69, 0x35, 0xEC}}

	PROPID_ACC_NAV_DOWN = syscall.GUID{0x031670ED, 0x3CDF, 0x48D2,
		[8]byte{0x96, 0x13, 0x13, 0x8F, 0x2D, 0xD8, 0xA6, 0x68}}

	PROPID_ACC_NAV_LEFT = syscall.GUID{0x228086CB, 0x82F1, 0x4A39,
		[8]byte{0x87, 0x05, 0xDC, 0xDC, 0x0F, 0xFF, 0x92, 0xF5}}

	PROPID_ACC_NAV_RIGHT = syscall.GUID{0xCD211D9F, 0xE1CB, 0x4FE5,
		[8]byte{0xA7, 0x7C, 0x92, 0x0B, 0x88, 0x4D, 0x09, 0x5B}}

	PROPID_ACC_NAV_PREV = syscall.GUID{0x776D3891, 0xC73B, 0x4480,
		[8]byte{0xB3, 0xF6, 0x07, 0x6A, 0x16, 0xA1, 0x5A, 0xF6}}

	PROPID_ACC_NAV_NEXT = syscall.GUID{0x1CDC5455, 0x8CD9, 0x4C92,
		[8]byte{0xA3, 0x71, 0x39, 0x39, 0xA2, 0xFE, 0x3E, 0xEE}}

	PROPID_ACC_NAV_FIRSTCHILD = syscall.GUID{0xCFD02558, 0x557B, 0x4C67,
		[8]byte{0x84, 0xF9, 0x2A, 0x09, 0xFC, 0xE4, 0x07, 0x49}}

	PROPID_ACC_NAV_LASTCHILD = syscall.GUID{0x302ECAA5, 0x48D5, 0x4F8D,
		[8]byte{0xB6, 0x71, 0x1A, 0x8D, 0x20, 0xA7, 0x78, 0x32}}

	PROPID_ACC_ROLEMAP = syscall.GUID{0xF79ACDA2, 0x140D, 0x4FE6,
		[8]byte{0x89, 0x14, 0x20, 0x84, 0x76, 0x32, 0x82, 0x69}}

	PROPID_ACC_VALUEMAP = syscall.GUID{0xDA1C3D79, 0xFC5C, 0x420E,
		[8]byte{0xB3, 0x99, 0x9D, 0x15, 0x33, 0x54, 0x9E, 0x75}}

	PROPID_ACC_STATEMAP = syscall.GUID{0x43946C5E, 0x0AC0, 0x4042,
		[8]byte{0xB5, 0x25, 0x07, 0xBB, 0xDB, 0xE1, 0x7F, 0xA7}}

	PROPID_ACC_DESCRIPTIONMAP = syscall.GUID{0x1FF1435F, 0x8A14, 0x477B,
		[8]byte{0xB2, 0x26, 0xA0, 0xAB, 0xE2, 0x79, 0x97, 0x5D}}

	PROPID_ACC_DODEFAULTACTION = syscall.GUID{0x1BA09523, 0x2E3B, 0x49A6,
		[8]byte{0xA0, 0x59, 0x59, 0x68, 0x2A, 0x3C, 0x48, 0xFD}}

	RuntimeId_Property_GUID = syscall.GUID{0xA39EEBFA, 0x7FBA, 0x4C89,
		[8]byte{0xB4, 0xD4, 0xB9, 0x9E, 0x2D, 0xE7, 0xD1, 0x60}}

	BoundingRectangle_Property_GUID = syscall.GUID{0x7BBFE8B2, 0x3BFC, 0x48DD,
		[8]byte{0xB7, 0x29, 0xC7, 0x94, 0xB8, 0x46, 0xE9, 0xA1}}

	ProcessId_Property_GUID = syscall.GUID{0x40499998, 0x9C31, 0x4245,
		[8]byte{0xA4, 0x03, 0x87, 0x32, 0x0E, 0x59, 0xEA, 0xF6}}

	ControlType_Property_GUID = syscall.GUID{0xCA774FEA, 0x28AC, 0x4BC2,
		[8]byte{0x94, 0xCA, 0xAC, 0xEC, 0x6D, 0x6C, 0x10, 0xA3}}

	LocalizedControlType_Property_GUID = syscall.GUID{0x8763404F, 0xA1BD, 0x452A,
		[8]byte{0x89, 0xC4, 0x3F, 0x01, 0xD3, 0x83, 0x38, 0x06}}

	Name_Property_GUID = syscall.GUID{0xC3A6921B, 0x4A99, 0x44F1,
		[8]byte{0xBC, 0xA6, 0x61, 0x18, 0x70, 0x52, 0xC4, 0x31}}

	AcceleratorKey_Property_GUID = syscall.GUID{0x514865DF, 0x2557, 0x4CB9,
		[8]byte{0xAE, 0xED, 0x6C, 0xED, 0x08, 0x4C, 0xE5, 0x2C}}

	AccessKey_Property_GUID = syscall.GUID{0x06827B12, 0xA7F9, 0x4A15,
		[8]byte{0x91, 0x7C, 0xFF, 0xA5, 0xAD, 0x3E, 0xB0, 0xA7}}

	HasKeyboardFocus_Property_GUID = syscall.GUID{0xCF8AFD39, 0x3F46, 0x4800,
		[8]byte{0x96, 0x56, 0xB2, 0xBF, 0x12, 0x52, 0x99, 0x05}}

	IsKeyboardFocusable_Property_GUID = syscall.GUID{0xF7B8552A, 0x0859, 0x4B37,
		[8]byte{0xB9, 0xCB, 0x51, 0xE7, 0x20, 0x92, 0xF2, 0x9F}}

	IsEnabled_Property_GUID = syscall.GUID{0x2109427F, 0xDA60, 0x4FED,
		[8]byte{0xBF, 0x1B, 0x26, 0x4B, 0xDC, 0xE6, 0xEB, 0x3A}}

	AutomationId_Property_GUID = syscall.GUID{0xC82C0500, 0xB60E, 0x4310,
		[8]byte{0xA2, 0x67, 0x30, 0x3C, 0x53, 0x1F, 0x8E, 0xE5}}

	ClassName_Property_GUID = syscall.GUID{0x157B7215, 0x894F, 0x4B65,
		[8]byte{0x84, 0xE2, 0xAA, 0xC0, 0xDA, 0x08, 0xB1, 0x6B}}

	HelpText_Property_GUID = syscall.GUID{0x08555685, 0x0977, 0x45C7,
		[8]byte{0xA7, 0xA6, 0xAB, 0xAF, 0x56, 0x84, 0x12, 0x1A}}

	ClickablePoint_Property_GUID = syscall.GUID{0x0196903B, 0xB203, 0x4818,
		[8]byte{0xA9, 0xF3, 0xF0, 0x8E, 0x67, 0x5F, 0x23, 0x41}}

	Culture_Property_GUID = syscall.GUID{0xE2D74F27, 0x3D79, 0x4DC2,
		[8]byte{0xB8, 0x8B, 0x30, 0x44, 0x96, 0x3A, 0x8A, 0xFB}}

	IsControlElement_Property_GUID = syscall.GUID{0x95F35085, 0xABCC, 0x4AFD,
		[8]byte{0xA5, 0xF4, 0xDB, 0xB4, 0x6C, 0x23, 0x0F, 0xDB}}

	IsContentElement_Property_GUID = syscall.GUID{0x4BDA64A8, 0xF5D8, 0x480B,
		[8]byte{0x81, 0x55, 0xEF, 0x2E, 0x89, 0xAD, 0xB6, 0x72}}

	LabeledBy_Property_GUID = syscall.GUID{0xE5B8924B, 0xFC8A, 0x4A35,
		[8]byte{0x80, 0x31, 0xCF, 0x78, 0xAC, 0x43, 0xE5, 0x5E}}

	IsPassword_Property_GUID = syscall.GUID{0xE8482EB1, 0x687C, 0x497B,
		[8]byte{0xBE, 0xBC, 0x03, 0xBE, 0x53, 0xEC, 0x14, 0x54}}

	NewNativeWindowHandle_Property_GUID = syscall.GUID{0x5196B33B, 0x380A, 0x4982,
		[8]byte{0x95, 0xE1, 0x91, 0xF3, 0xEF, 0x60, 0xE0, 0x24}}

	ItemType_Property_GUID = syscall.GUID{0xCDDA434D, 0x6222, 0x413B,
		[8]byte{0xA6, 0x8A, 0x32, 0x5D, 0xD1, 0xD4, 0x0F, 0x39}}

	IsOffscreen_Property_GUID = syscall.GUID{0x03C3D160, 0xDB79, 0x42DB,
		[8]byte{0xA2, 0xEF, 0x1C, 0x23, 0x1E, 0xED, 0xE5, 0x07}}

	Orientation_Property_GUID = syscall.GUID{0xA01EEE62, 0x3884, 0x4415,
		[8]byte{0x88, 0x7E, 0x67, 0x8E, 0xC2, 0x1E, 0x39, 0xBA}}

	FrameworkId_Property_GUID = syscall.GUID{0xDBFD9900, 0x7E1A, 0x4F58,
		[8]byte{0xB6, 0x1B, 0x70, 0x63, 0x12, 0x0F, 0x77, 0x3B}}

	IsRequiredForForm_Property_GUID = syscall.GUID{0x4F5F43CF, 0x59FB, 0x4BDE,
		[8]byte{0xA2, 0x70, 0x60, 0x2E, 0x5E, 0x11, 0x41, 0xE9}}

	ItemStatus_Property_GUID = syscall.GUID{0x51DE0321, 0x3973, 0x43E7,
		[8]byte{0x89, 0x13, 0x0B, 0x08, 0xE8, 0x13, 0xC3, 0x7F}}

	AriaRole_Property_GUID = syscall.GUID{0xDD207B95, 0xBE4A, 0x4E0D,
		[8]byte{0xB7, 0x27, 0x63, 0xAC, 0xE9, 0x4B, 0x69, 0x16}}

	AriaProperties_Property_GUID = syscall.GUID{0x4213678C, 0xE025, 0x4922,
		[8]byte{0xBE, 0xB5, 0xE4, 0x3B, 0xA0, 0x8E, 0x62, 0x21}}

	IsDataValidForForm_Property_GUID = syscall.GUID{0x445AC684, 0xC3FC, 0x4DD9,
		[8]byte{0xAC, 0xF8, 0x84, 0x5A, 0x57, 0x92, 0x96, 0xBA}}

	ControllerFor_Property_GUID = syscall.GUID{0x51124C8A, 0xA5D2, 0x4F13,
		[8]byte{0x9B, 0xE6, 0x7F, 0xA8, 0xBA, 0x9D, 0x3A, 0x90}}

	DescribedBy_Property_GUID = syscall.GUID{0x7C5865B8, 0x9992, 0x40FD,
		[8]byte{0x8D, 0xB0, 0x6B, 0xF1, 0xD3, 0x17, 0xF9, 0x98}}

	FlowsTo_Property_GUID = syscall.GUID{0xE4F33D20, 0x559A, 0x47FB,
		[8]byte{0xA8, 0x30, 0xF9, 0xCB, 0x4F, 0xF1, 0xA7, 0x0A}}

	ProviderDescription_Property_GUID = syscall.GUID{0xDCA5708A, 0xC16B, 0x4CD9,
		[8]byte{0xB8, 0x89, 0xBE, 0xB1, 0x6A, 0x80, 0x49, 0x04}}

	OptimizeForVisualContent_Property_GUID = syscall.GUID{0x6A852250, 0xC75A, 0x4E5D,
		[8]byte{0xB8, 0x58, 0xE3, 0x81, 0xB0, 0xF7, 0x88, 0x61}}

	IsDockPatternAvailable_Property_GUID = syscall.GUID{0x2600A4C4, 0x2FF8, 0x4C96,
		[8]byte{0xAE, 0x31, 0x8F, 0xE6, 0x19, 0xA1, 0x3C, 0x6C}}

	IsExpandCollapsePatternAvailable_Property_GUID = syscall.GUID{0x929D3806, 0x5287, 0x4725,
		[8]byte{0xAA, 0x16, 0x22, 0x2A, 0xFC, 0x63, 0xD5, 0x95}}

	IsGridItemPatternAvailable_Property_GUID = syscall.GUID{0x5A43E524, 0xF9A2, 0x4B12,
		[8]byte{0x84, 0xC8, 0xB4, 0x8A, 0x3E, 0xFE, 0xDD, 0x34}}

	IsGridPatternAvailable_Property_GUID = syscall.GUID{0x5622C26C, 0xF0EF, 0x4F3B,
		[8]byte{0x97, 0xCB, 0x71, 0x4C, 0x08, 0x68, 0x58, 0x8B}}

	IsInvokePatternAvailable_Property_GUID = syscall.GUID{0x4E725738, 0x8364, 0x4679,
		[8]byte{0xAA, 0x6C, 0xF3, 0xF4, 0x19, 0x31, 0xF7, 0x50}}

	IsMultipleViewPatternAvailable_Property_GUID = syscall.GUID{0xFF0A31EB, 0x8E25, 0x469D,
		[8]byte{0x8D, 0x6E, 0xE7, 0x71, 0xA2, 0x7C, 0x1B, 0x90}}

	IsRangeValuePatternAvailable_Property_GUID = syscall.GUID{0xFDA4244A, 0xEB4D, 0x43FF,
		[8]byte{0xB5, 0xAD, 0xED, 0x36, 0xD3, 0x73, 0xEC, 0x4C}}

	IsScrollPatternAvailable_Property_GUID = syscall.GUID{0x3EBB7B4A, 0x828A, 0x4B57,
		[8]byte{0x9D, 0x22, 0x2F, 0xEA, 0x16, 0x32, 0xED, 0x0D}}

	IsScrollItemPatternAvailable_Property_GUID = syscall.GUID{0x1CAD1A05, 0x0927, 0x4B76,
		[8]byte{0x97, 0xE1, 0x0F, 0xCD, 0xB2, 0x09, 0xB9, 0x8A}}

	IsSelectionItemPatternAvailable_Property_GUID = syscall.GUID{0x8BECD62D, 0x0BC3, 0x4109,
		[8]byte{0xBE, 0xE2, 0x8E, 0x67, 0x15, 0x29, 0x0E, 0x68}}

	IsSelectionPatternAvailable_Property_GUID = syscall.GUID{0xF588ACBE, 0xC769, 0x4838,
		[8]byte{0x9A, 0x60, 0x26, 0x86, 0xDC, 0x11, 0x88, 0xC4}}

	IsTablePatternAvailable_Property_GUID = syscall.GUID{0xCB83575F, 0x45C2, 0x4048,
		[8]byte{0x9C, 0x76, 0x15, 0x97, 0x15, 0xA1, 0x39, 0xDF}}

	IsTableItemPatternAvailable_Property_GUID = syscall.GUID{0xEB36B40D, 0x8EA4, 0x489B,
		[8]byte{0xA0, 0x13, 0xE6, 0x0D, 0x59, 0x51, 0xFE, 0x34}}

	IsTextPatternAvailable_Property_GUID = syscall.GUID{0xFBE2D69D, 0xAFF6, 0x4A45,
		[8]byte{0x82, 0xE2, 0xFC, 0x92, 0xA8, 0x2F, 0x59, 0x17}}

	IsTogglePatternAvailable_Property_GUID = syscall.GUID{0x78686D53, 0xFCD0, 0x4B83,
		[8]byte{0x9B, 0x78, 0x58, 0x32, 0xCE, 0x63, 0xBB, 0x5B}}

	IsTransformPatternAvailable_Property_GUID = syscall.GUID{0xA7F78804, 0xD68B, 0x4077,
		[8]byte{0xA5, 0xC6, 0x7A, 0x5E, 0xA1, 0xAC, 0x31, 0xC5}}

	IsValuePatternAvailable_Property_GUID = syscall.GUID{0x0B5020A7, 0x2119, 0x473B,
		[8]byte{0xBE, 0x37, 0x5C, 0xEB, 0x98, 0xBB, 0xFB, 0x22}}

	IsWindowPatternAvailable_Property_GUID = syscall.GUID{0xE7A57BB1, 0x5888, 0x4155,
		[8]byte{0x98, 0xDC, 0xB4, 0x22, 0xFD, 0x57, 0xF2, 0xBC}}

	IsLegacyIAccessiblePatternAvailable_Property_GUID = syscall.GUID{0xD8EBD0C7, 0x929A, 0x4EE7,
		[8]byte{0x8D, 0x3A, 0xD3, 0xD9, 0x44, 0x13, 0x02, 0x7B}}

	IsItemContainerPatternAvailable_Property_GUID = syscall.GUID{0x624B5CA7, 0xFE40, 0x4957,
		[8]byte{0xA0, 0x19, 0x20, 0xC4, 0xCF, 0x11, 0x92, 0x0F}}

	IsVirtualizedItemPatternAvailable_Property_GUID = syscall.GUID{0x302CB151, 0x2AC8, 0x45D6,
		[8]byte{0x97, 0x7B, 0xD2, 0xB3, 0xA5, 0xA5, 0x3F, 0x20}}

	IsSynchronizedInputPatternAvailable_Property_GUID = syscall.GUID{0x75D69CC5, 0xD2BF, 0x4943,
		[8]byte{0x87, 0x6E, 0xB4, 0x5B, 0x62, 0xA6, 0xCC, 0x66}}

	IsObjectModelPatternAvailable_Property_GUID = syscall.GUID{0x6B21D89B, 0x2841, 0x412F,
		[8]byte{0x8E, 0xF2, 0x15, 0xCA, 0x95, 0x23, 0x18, 0xBA}}

	IsAnnotationPatternAvailable_Property_GUID = syscall.GUID{0x0B5B3238, 0x6D5C, 0x41B6,
		[8]byte{0xBC, 0xC4, 0x5E, 0x80, 0x7F, 0x65, 0x51, 0xC4}}

	IsTextPattern2Available_Property_GUID = syscall.GUID{0x41CF921D, 0xE3F1, 0x4B22,
		[8]byte{0x9C, 0x81, 0xE1, 0xC3, 0xED, 0x33, 0x1C, 0x22}}

	IsTextEditPatternAvailable_Property_GUID = syscall.GUID{0x7843425C, 0x8B32, 0x484C,
		[8]byte{0x9A, 0xB5, 0xE3, 0x20, 0x05, 0x71, 0xFF, 0xDA}}

	IsCustomNavigationPatternAvailable_Property_GUID = syscall.GUID{0x8F8E80D4, 0x2351, 0x48E0,
		[8]byte{0x87, 0x4A, 0x54, 0xAA, 0x73, 0x13, 0x88, 0x9A}}

	IsStylesPatternAvailable_Property_GUID = syscall.GUID{0x27F353D3, 0x459C, 0x4B59,
		[8]byte{0xA4, 0x90, 0x50, 0x61, 0x1D, 0xAC, 0xAF, 0xB5}}

	IsSpreadsheetPatternAvailable_Property_GUID = syscall.GUID{0x6FF43732, 0xE4B4, 0x4555,
		[8]byte{0x97, 0xBC, 0xEC, 0xDB, 0xBC, 0x4D, 0x18, 0x88}}

	IsSpreadsheetItemPatternAvailable_Property_GUID = syscall.GUID{0x9FE79B2A, 0x2F94, 0x43FD,
		[8]byte{0x99, 0x6B, 0x54, 0x9E, 0x31, 0x6F, 0x4A, 0xCD}}

	IsTransformPattern2Available_Property_GUID = syscall.GUID{0x25980B4B, 0xBE04, 0x4710,
		[8]byte{0xAB, 0x4A, 0xFD, 0xA3, 0x1D, 0xBD, 0x28, 0x95}}

	IsTextChildPatternAvailable_Property_GUID = syscall.GUID{0x559E65DF, 0x30FF, 0x43B5,
		[8]byte{0xB5, 0xED, 0x5B, 0x28, 0x3B, 0x80, 0xC7, 0xE9}}

	IsDragPatternAvailable_Property_GUID = syscall.GUID{0xE997A7B7, 0x1D39, 0x4CA7,
		[8]byte{0xBE, 0x0F, 0x27, 0x7F, 0xCF, 0x56, 0x05, 0xCC}}

	IsDropTargetPatternAvailable_Property_GUID = syscall.GUID{0x0686B62E, 0x8E19, 0x4AAF,
		[8]byte{0x87, 0x3D, 0x38, 0x4F, 0x6D, 0x3B, 0x92, 0xBE}}

	IsStructuredMarkupPatternAvailable_Property_GUID = syscall.GUID{0xB0D4C196, 0x2C0B, 0x489C,
		[8]byte{0xB1, 0x65, 0xA4, 0x05, 0x92, 0x8C, 0x6F, 0x3D}}

	IsPeripheral_Property_GUID = syscall.GUID{0xDA758276, 0x7ED5, 0x49D4,
		[8]byte{0x8E, 0x68, 0xEC, 0xC9, 0xA2, 0xD3, 0x00, 0xDD}}

	PositionInSet_Property_GUID = syscall.GUID{0x33D1DC54, 0x641E, 0x4D76,
		[8]byte{0xA6, 0xB1, 0x13, 0xF3, 0x41, 0xC1, 0xF8, 0x96}}

	SizeOfSet_Property_GUID = syscall.GUID{0x1600D33C, 0x3B9F, 0x4369,
		[8]byte{0x94, 0x31, 0xAA, 0x29, 0x3F, 0x34, 0x4C, 0xF1}}

	Level_Property_GUID = syscall.GUID{0x242AC529, 0xCD36, 0x400F,
		[8]byte{0xAA, 0xD9, 0x78, 0x76, 0xEF, 0x3A, 0xF6, 0x27}}

	AnnotationTypes_Property_GUID = syscall.GUID{0x64B71F76, 0x53C4, 0x4696,
		[8]byte{0xA2, 0x19, 0x20, 0xE9, 0x40, 0xC9, 0xA1, 0x76}}

	AnnotationObjects_Property_GUID = syscall.GUID{0x310910C8, 0x7C6E, 0x4F20,
		[8]byte{0xBE, 0xCD, 0x4A, 0xAF, 0x6D, 0x19, 0x11, 0x56}}

	LandmarkType_Property_GUID = syscall.GUID{0x454045F2, 0x6F61, 0x49F7,
		[8]byte{0xA4, 0xF8, 0xB5, 0xF0, 0xCF, 0x82, 0xDA, 0x1E}}

	LocalizedLandmarkType_Property_GUID = syscall.GUID{0x7AC81980, 0xEAFB, 0x4FB2,
		[8]byte{0xBF, 0x91, 0xF4, 0x85, 0xBE, 0xF5, 0xE8, 0xE1}}

	FullDescription_Property_GUID = syscall.GUID{0x0D4450FF, 0x6AEF, 0x4F33,
		[8]byte{0x95, 0xDD, 0x7B, 0xEF, 0xA7, 0x2A, 0x43, 0x91}}

	Value_Value_Property_GUID = syscall.GUID{0xE95F5E64, 0x269F, 0x4A85,
		[8]byte{0xBA, 0x99, 0x40, 0x92, 0xC3, 0xEA, 0x29, 0x86}}

	Value_IsReadOnly_Property_GUID = syscall.GUID{0xEB090F30, 0xE24C, 0x4799,
		[8]byte{0xA7, 0x05, 0x0D, 0x24, 0x7B, 0xC0, 0x37, 0xF8}}

	RangeValue_Value_Property_GUID = syscall.GUID{0x131F5D98, 0xC50C, 0x489D,
		[8]byte{0xAB, 0xE5, 0xAE, 0x22, 0x08, 0x98, 0xC5, 0xF7}}

	RangeValue_IsReadOnly_Property_GUID = syscall.GUID{0x25FA1055, 0xDEBF, 0x4373,
		[8]byte{0xA7, 0x9E, 0x1F, 0x1A, 0x19, 0x08, 0xD3, 0xC4}}

	RangeValue_Minimum_Property_GUID = syscall.GUID{0x78CBD3B2, 0x684D, 0x4860,
		[8]byte{0xAF, 0x93, 0xD1, 0xF9, 0x5C, 0xB0, 0x22, 0xFD}}

	RangeValue_Maximum_Property_GUID = syscall.GUID{0x19319914, 0xF979, 0x4B35,
		[8]byte{0xA1, 0xA6, 0xD3, 0x7E, 0x05, 0x43, 0x34, 0x73}}

	RangeValue_LargeChange_Property_GUID = syscall.GUID{0xA1F96325, 0x3A3D, 0x4B44,
		[8]byte{0x8E, 0x1F, 0x4A, 0x46, 0xD9, 0x84, 0x40, 0x19}}

	RangeValue_SmallChange_Property_GUID = syscall.GUID{0x81C2C457, 0x3941, 0x4107,
		[8]byte{0x99, 0x75, 0x13, 0x97, 0x60, 0xF7, 0xC0, 0x72}}

	Scroll_HorizontalScrollPercent_Property_GUID = syscall.GUID{0xC7C13C0E, 0xEB21, 0x47FF,
		[8]byte{0xAC, 0xC4, 0xB5, 0xA3, 0x35, 0x0F, 0x51, 0x91}}

	Scroll_HorizontalViewSize_Property_GUID = syscall.GUID{0x70C2E5D4, 0xFCB0, 0x4713,
		[8]byte{0xA9, 0xAA, 0xAF, 0x92, 0xFF, 0x79, 0xE4, 0xCD}}

	Scroll_VerticalScrollPercent_Property_GUID = syscall.GUID{0x6C8D7099, 0xB2A8, 0x4948,
		[8]byte{0xBF, 0xF7, 0x3C, 0xF9, 0x05, 0x8B, 0xFE, 0xFB}}

	Scroll_VerticalViewSize_Property_GUID = syscall.GUID{0xDE6A2E22, 0xD8C7, 0x40C5,
		[8]byte{0x83, 0xBA, 0xE5, 0xF6, 0x81, 0xD5, 0x31, 0x08}}

	Scroll_HorizontallyScrollable_Property_GUID = syscall.GUID{0x8B925147, 0x28CD, 0x49AE,
		[8]byte{0xBD, 0x63, 0xF4, 0x41, 0x18, 0xD2, 0xE7, 0x19}}

	Scroll_VerticallyScrollable_Property_GUID = syscall.GUID{0x89164798, 0x0068, 0x4315,
		[8]byte{0xB8, 0x9A, 0x1E, 0x7C, 0xFB, 0xBC, 0x3D, 0xFC}}

	Selection_Selection_Property_GUID = syscall.GUID{0xAA6DC2A2, 0x0E2B, 0x4D38,
		[8]byte{0x96, 0xD5, 0x34, 0xE4, 0x70, 0xB8, 0x18, 0x53}}

	Selection_CanSelectMultiple_Property_GUID = syscall.GUID{0x49D73DA5, 0xC883, 0x4500,
		[8]byte{0x88, 0x3D, 0x8F, 0xCF, 0x8D, 0xAF, 0x6C, 0xBE}}

	Selection_IsSelectionRequired_Property_GUID = syscall.GUID{0xB1AE4422, 0x63FE, 0x44E7,
		[8]byte{0xA5, 0xA5, 0xA7, 0x38, 0xC8, 0x29, 0xB1, 0x9A}}

	Grid_RowCount_Property_GUID = syscall.GUID{0x2A9505BF, 0xC2EB, 0x4FB6,
		[8]byte{0xB3, 0x56, 0x82, 0x45, 0xAE, 0x53, 0x70, 0x3E}}

	Grid_ColumnCount_Property_GUID = syscall.GUID{0xFE96F375, 0x44AA, 0x4536,
		[8]byte{0xAC, 0x7A, 0x2A, 0x75, 0xD7, 0x1A, 0x3E, 0xFC}}

	GridItem_Row_Property_GUID = syscall.GUID{0x6223972A, 0xC945, 0x4563,
		[8]byte{0x93, 0x29, 0xFD, 0xC9, 0x74, 0xAF, 0x25, 0x53}}

	GridItem_Column_Property_GUID = syscall.GUID{0xC774C15C, 0x62C0, 0x4519,
		[8]byte{0x8B, 0xDC, 0x47, 0xBE, 0x57, 0x3C, 0x8A, 0xD5}}

	GridItem_RowSpan_Property_GUID = syscall.GUID{0x4582291C, 0x466B, 0x4E93,
		[8]byte{0x8E, 0x83, 0x3D, 0x17, 0x15, 0xEC, 0x0C, 0x5E}}

	GridItem_ColumnSpan_Property_GUID = syscall.GUID{0x583EA3F5, 0x86D0, 0x4B08,
		[8]byte{0xA6, 0xEC, 0x2C, 0x54, 0x63, 0xFF, 0xC1, 0x09}}

	GridItem_Parent_Property_GUID = syscall.GUID{0x9D912252, 0xB97F, 0x4ECC,
		[8]byte{0x85, 0x10, 0xEA, 0x0E, 0x33, 0x42, 0x7C, 0x72}}

	Dock_DockPosition_Property_GUID = syscall.GUID{0x6D67F02E, 0xC0B0, 0x4B10,
		[8]byte{0xB5, 0xB9, 0x18, 0xD6, 0xEC, 0xF9, 0x87, 0x60}}

	ExpandCollapse_ExpandCollapseState_Property_GUID = syscall.GUID{0x275A4C48, 0x85A7, 0x4F69,
		[8]byte{0xAB, 0xA0, 0xAF, 0x15, 0x76, 0x10, 0x00, 0x2B}}

	MultipleView_CurrentView_Property_GUID = syscall.GUID{0x7A81A67A, 0xB94F, 0x4875,
		[8]byte{0x91, 0x8B, 0x65, 0xC8, 0xD2, 0xF9, 0x98, 0xE5}}

	MultipleView_SupportedViews_Property_GUID = syscall.GUID{0x8D5DB9FD, 0xCE3C, 0x4AE7,
		[8]byte{0xB7, 0x88, 0x40, 0x0A, 0x3C, 0x64, 0x55, 0x47}}

	Window_CanMaximize_Property_GUID = syscall.GUID{0x64FFF53F, 0x635D, 0x41C1,
		[8]byte{0x95, 0x0C, 0xCB, 0x5A, 0xDF, 0xBE, 0x28, 0xE3}}

	Window_CanMinimize_Property_GUID = syscall.GUID{0xB73B4625, 0x5988, 0x4B97,
		[8]byte{0xB4, 0xC2, 0xA6, 0xFE, 0x6E, 0x78, 0xC8, 0xC6}}

	Window_WindowVisualState_Property_GUID = syscall.GUID{0x4AB7905F, 0xE860, 0x453E,
		[8]byte{0xA3, 0x0A, 0xF6, 0x43, 0x1E, 0x5D, 0xAA, 0xD5}}

	Window_WindowInteractionState_Property_GUID = syscall.GUID{0x4FED26A4, 0x0455, 0x4FA2,
		[8]byte{0xB2, 0x1C, 0xC4, 0xDA, 0x2D, 0xB1, 0xFF, 0x9C}}

	Window_IsModal_Property_GUID = syscall.GUID{0xFF4E6892, 0x37B9, 0x4FCA,
		[8]byte{0x85, 0x32, 0xFF, 0xE6, 0x74, 0xEC, 0xFE, 0xED}}

	Window_IsTopmost_Property_GUID = syscall.GUID{0xEF7D85D3, 0x0937, 0x4962,
		[8]byte{0x92, 0x41, 0xB6, 0x23, 0x45, 0xF2, 0x40, 0x41}}

	SelectionItem_IsSelected_Property_GUID = syscall.GUID{0xF122835F, 0xCD5F, 0x43DF,
		[8]byte{0xB7, 0x9D, 0x4B, 0x84, 0x9E, 0x9E, 0x60, 0x20}}

	SelectionItem_SelectionContainer_Property_GUID = syscall.GUID{0xA4365B6E, 0x9C1E, 0x4B63,
		[8]byte{0x8B, 0x53, 0xC2, 0x42, 0x1D, 0xD1, 0xE8, 0xFB}}

	Table_RowHeaders_Property_GUID = syscall.GUID{0xD9E35B87, 0x6EB8, 0x4562,
		[8]byte{0xAA, 0xC6, 0xA8, 0xA9, 0x07, 0x52, 0x36, 0xA8}}

	Table_ColumnHeaders_Property_GUID = syscall.GUID{0xAFF1D72B, 0x968D, 0x42B1,
		[8]byte{0xB4, 0x59, 0x15, 0x0B, 0x29, 0x9D, 0xA6, 0x64}}

	Table_RowOrColumnMajor_Property_GUID = syscall.GUID{0x83BE75C3, 0x29FE, 0x4A30,
		[8]byte{0x85, 0xE1, 0x2A, 0x62, 0x77, 0xFD, 0x10, 0x6E}}

	TableItem_RowHeaderItems_Property_GUID = syscall.GUID{0xB3F853A0, 0x0574, 0x4CD8,
		[8]byte{0xBC, 0xD7, 0xED, 0x59, 0x23, 0x57, 0x2D, 0x97}}

	TableItem_ColumnHeaderItems_Property_GUID = syscall.GUID{0x967A56A3, 0x74B6, 0x431E,
		[8]byte{0x8D, 0xE6, 0x99, 0xC4, 0x11, 0x03, 0x1C, 0x58}}

	Toggle_ToggleState_Property_GUID = syscall.GUID{0xB23CDC52, 0x22C2, 0x4C6C,
		[8]byte{0x9D, 0xED, 0xF5, 0xC4, 0x22, 0x47, 0x9E, 0xDE}}

	Transform_CanMove_Property_GUID = syscall.GUID{0x1B75824D, 0x208B, 0x4FDF,
		[8]byte{0xBC, 0xCD, 0xF1, 0xF4, 0xE5, 0x74, 0x1F, 0x4F}}

	Transform_CanResize_Property_GUID = syscall.GUID{0xBB98DCA5, 0x4C1A, 0x41D4,
		[8]byte{0xA4, 0xF6, 0xEB, 0xC1, 0x28, 0x64, 0x41, 0x80}}

	Transform_CanRotate_Property_GUID = syscall.GUID{0x10079B48, 0x3849, 0x476F,
		[8]byte{0xAC, 0x96, 0x44, 0xA9, 0x5C, 0x84, 0x40, 0xD9}}

	LegacyIAccessible_ChildId_Property_GUID = syscall.GUID{0x9A191B5D, 0x9EF2, 0x4787,
		[8]byte{0xA4, 0x59, 0xDC, 0xDE, 0x88, 0x5D, 0xD4, 0xE8}}

	LegacyIAccessible_Name_Property_GUID = syscall.GUID{0xCAEB063D, 0x40AE, 0x4869,
		[8]byte{0xAA, 0x5A, 0x1B, 0x8E, 0x5D, 0x66, 0x67, 0x39}}

	LegacyIAccessible_Value_Property_GUID = syscall.GUID{0xB5C5B0B6, 0x8217, 0x4A77,
		[8]byte{0x97, 0xA5, 0x19, 0x0A, 0x85, 0xED, 0x01, 0x56}}

	LegacyIAccessible_Description_Property_GUID = syscall.GUID{0x46448418, 0x7D70, 0x4EA9,
		[8]byte{0x9D, 0x27, 0xB7, 0xE7, 0x75, 0xCF, 0x2A, 0xD7}}

	LegacyIAccessible_Role_Property_GUID = syscall.GUID{0x6856E59F, 0xCBAF, 0x4E31,
		[8]byte{0x93, 0xE8, 0xBC, 0xBF, 0x6F, 0x7E, 0x49, 0x1C}}

	LegacyIAccessible_State_Property_GUID = syscall.GUID{0xDF985854, 0x2281, 0x4340,
		[8]byte{0xAB, 0x9C, 0xC6, 0x0E, 0x2C, 0x58, 0x03, 0xF6}}

	LegacyIAccessible_Help_Property_GUID = syscall.GUID{0x94402352, 0x161C, 0x4B77,
		[8]byte{0xA9, 0x8D, 0xA8, 0x72, 0xCC, 0x33, 0x94, 0x7A}}

	LegacyIAccessible_KeyboardShortcut_Property_GUID = syscall.GUID{0x8F6909AC, 0x00B8, 0x4259,
		[8]byte{0xA4, 0x1C, 0x96, 0x62, 0x66, 0xD4, 0x3A, 0x8A}}

	LegacyIAccessible_Selection_Property_GUID = syscall.GUID{0x8AA8B1E0, 0x0891, 0x40CC,
		[8]byte{0x8B, 0x06, 0x90, 0xD7, 0xD4, 0x16, 0x62, 0x19}}

	LegacyIAccessible_DefaultAction_Property_GUID = syscall.GUID{0x3B331729, 0xEAAD, 0x4502,
		[8]byte{0xB8, 0x5F, 0x92, 0x61, 0x56, 0x22, 0x91, 0x3C}}

	Annotation_AnnotationTypeId_Property_GUID = syscall.GUID{0x20AE484F, 0x69EF, 0x4C48,
		[8]byte{0x8F, 0x5B, 0xC4, 0x93, 0x8B, 0x20, 0x6A, 0xC7}}

	Annotation_AnnotationTypeName_Property_GUID = syscall.GUID{0x9B818892, 0x5AC9, 0x4AF9,
		[8]byte{0xAA, 0x96, 0xF5, 0x8A, 0x77, 0xB0, 0x58, 0xE3}}

	Annotation_Author_Property_GUID = syscall.GUID{0x7A528462, 0x9C5C, 0x4A03,
		[8]byte{0xA9, 0x74, 0x8B, 0x30, 0x7A, 0x99, 0x37, 0xF2}}

	Annotation_DateTime_Property_GUID = syscall.GUID{0x99B5CA5D, 0x1ACF, 0x414B,
		[8]byte{0xA4, 0xD0, 0x6B, 0x35, 0x0B, 0x04, 0x75, 0x78}}

	Annotation_Target_Property_GUID = syscall.GUID{0xB71B302D, 0x2104, 0x44AD,
		[8]byte{0x9C, 0x5C, 0x09, 0x2B, 0x49, 0x07, 0xD7, 0x0F}}

	Styles_StyleId_Property_GUID = syscall.GUID{0xDA82852F, 0x3817, 0x4233,
		[8]byte{0x82, 0xAF, 0x02, 0x27, 0x9E, 0x72, 0xCC, 0x77}}

	Styles_StyleName_Property_GUID = syscall.GUID{0x1C12B035, 0x05D1, 0x4F55,
		[8]byte{0x9E, 0x8E, 0x14, 0x89, 0xF3, 0xFF, 0x55, 0x0D}}

	Styles_FillColor_Property_GUID = syscall.GUID{0x63EFF97A, 0xA1C5, 0x4B1D,
		[8]byte{0x84, 0xEB, 0xB7, 0x65, 0xF2, 0xED, 0xD6, 0x32}}

	Styles_FillPatternStyle_Property_GUID = syscall.GUID{0x81CF651F, 0x482B, 0x4451,
		[8]byte{0xA3, 0x0A, 0xE1, 0x54, 0x5E, 0x55, 0x4F, 0xB8}}

	Styles_Shape_Property_GUID = syscall.GUID{0xC71A23F8, 0x778C, 0x400D,
		[8]byte{0x84, 0x58, 0x3B, 0x54, 0x3E, 0x52, 0x69, 0x84}}

	Styles_FillPatternColor_Property_GUID = syscall.GUID{0x939A59FE, 0x8FBD, 0x4E75,
		[8]byte{0xA2, 0x71, 0xAC, 0x45, 0x95, 0x19, 0x51, 0x63}}

	Styles_ExtendedProperties_Property_GUID = syscall.GUID{0xF451CDA0, 0xBA0A, 0x4681,
		[8]byte{0xB0, 0xB0, 0x0D, 0xBD, 0xB5, 0x3E, 0x58, 0xF3}}

	SpreadsheetItem_Formula_Property_GUID = syscall.GUID{0xE602E47D, 0x1B47, 0x4BEA,
		[8]byte{0x87, 0xCF, 0x3B, 0x0B, 0x0B, 0x5C, 0x15, 0xB6}}

	SpreadsheetItem_AnnotationObjects_Property_GUID = syscall.GUID{0xA3194C38, 0xC9BC, 0x4604,
		[8]byte{0x93, 0x96, 0xAE, 0x3F, 0x9F, 0x45, 0x7F, 0x7B}}

	SpreadsheetItem_AnnotationTypes_Property_GUID = syscall.GUID{0xC70C51D0, 0xD602, 0x4B45,
		[8]byte{0xAF, 0xBC, 0xB4, 0x71, 0x2B, 0x96, 0xD7, 0x2B}}

	Transform2_CanZoom_Property_GUID = syscall.GUID{0xF357E890, 0xA756, 0x4359,
		[8]byte{0x9C, 0xA6, 0x86, 0x70, 0x2B, 0xF8, 0xF3, 0x81}}

	LiveSetting_Property_GUID = syscall.GUID{0xC12BCD8E, 0x2A8E, 0x4950,
		[8]byte{0x8A, 0xE7, 0x36, 0x25, 0x11, 0x1D, 0x58, 0xEB}}

	Drag_IsGrabbed_Property_GUID = syscall.GUID{0x45F206F3, 0x75CC, 0x4CCA,
		[8]byte{0xA9, 0xB9, 0xFC, 0xDF, 0xB9, 0x82, 0xD8, 0xA2}}

	Drag_GrabbedItems_Property_GUID = syscall.GUID{0x77C1562C, 0x7B86, 0x4B21,
		[8]byte{0x9E, 0xD7, 0x3C, 0xEF, 0xDA, 0x6F, 0x4C, 0x43}}

	Drag_DropEffect_Property_GUID = syscall.GUID{0x646F2779, 0x48D3, 0x4B23,
		[8]byte{0x89, 0x02, 0x4B, 0xF1, 0x00, 0x00, 0x5D, 0xF3}}

	Drag_DropEffects_Property_GUID = syscall.GUID{0xF5D61156, 0x7CE6, 0x49BE,
		[8]byte{0xA8, 0x36, 0x92, 0x69, 0xDC, 0xEC, 0x92, 0x0F}}

	DropTarget_DropTargetEffect_Property_GUID = syscall.GUID{0x8BB75975, 0xA0CA, 0x4981,
		[8]byte{0xB8, 0x18, 0x87, 0xFC, 0x66, 0xE9, 0x50, 0x9D}}

	DropTarget_DropTargetEffects_Property_GUID = syscall.GUID{0xBC1DD4ED, 0xCB89, 0x45F1,
		[8]byte{0xA5, 0x92, 0xE0, 0x3B, 0x08, 0xAE, 0x79, 0x0F}}

	Transform2_ZoomLevel_Property_GUID = syscall.GUID{0xEEE29F1A, 0xF4A2, 0x4B5B,
		[8]byte{0xAC, 0x65, 0x95, 0xCF, 0x93, 0x28, 0x33, 0x87}}

	Transform2_ZoomMinimum_Property_GUID = syscall.GUID{0x742CCC16, 0x4AD1, 0x4E07,
		[8]byte{0x96, 0xFE, 0xB1, 0x22, 0xC6, 0xE6, 0xB2, 0x2B}}

	Transform2_ZoomMaximum_Property_GUID = syscall.GUID{0x42AB6B77, 0xCEB0, 0x4ECA,
		[8]byte{0xB8, 0x2A, 0x6C, 0xFA, 0x5F, 0xA1, 0xFC, 0x08}}

	FlowsFrom_Property_GUID = syscall.GUID{0x05C6844F, 0x19DE, 0x48F8,
		[8]byte{0x95, 0xFA, 0x88, 0x0D, 0x5B, 0x0F, 0xD6, 0x15}}

	FillColor_Property_GUID = syscall.GUID{0x6E0EC4D0, 0xE2A8, 0x4A56,
		[8]byte{0x9D, 0xE7, 0x95, 0x33, 0x89, 0x93, 0x3B, 0x39}}

	OutlineColor_Property_GUID = syscall.GUID{0xC395D6C0, 0x4B55, 0x4762,
		[8]byte{0xA0, 0x73, 0xFD, 0x30, 0x3A, 0x63, 0x4F, 0x52}}

	FillType_Property_GUID = syscall.GUID{0xC6FC74E4, 0x8CB9, 0x429C,
		[8]byte{0xA9, 0xE1, 0x9B, 0xC4, 0xAC, 0x37, 0x2B, 0x62}}

	VisualEffects_Property_GUID = syscall.GUID{0xE61A8565, 0xAAD9, 0x46D7,
		[8]byte{0x9E, 0x70, 0x4E, 0x8A, 0x84, 0x20, 0xD4, 0x20}}

	OutlineThickness_Property_GUID = syscall.GUID{0x13E67CC7, 0xDAC2, 0x4888,
		[8]byte{0xBD, 0xD3, 0x37, 0x5C, 0x62, 0xFA, 0x96, 0x18}}

	CenterPoint_Property_GUID = syscall.GUID{0x0CB00C08, 0x540C, 0x4EDB,
		[8]byte{0x94, 0x45, 0x26, 0x35, 0x9E, 0xA6, 0x97, 0x85}}

	Rotation_Property_GUID = syscall.GUID{0x767CDC7D, 0xAEC0, 0x4110,
		[8]byte{0xAD, 0x32, 0x30, 0xED, 0xD4, 0x03, 0x49, 0x2E}}

	Size_Property_GUID = syscall.GUID{0x2B5F761D, 0xF885, 0x4404,
		[8]byte{0x97, 0x3F, 0x9B, 0x1D, 0x98, 0xE3, 0x6D, 0x8F}}

	ToolTipOpened_Event_GUID = syscall.GUID{0x3F4B97FF, 0x2EDC, 0x451D,
		[8]byte{0xBC, 0xA4, 0x95, 0xA3, 0x18, 0x8D, 0x5B, 0x03}}

	ToolTipClosed_Event_GUID = syscall.GUID{0x276D71EF, 0x24A9, 0x49B6,
		[8]byte{0x8E, 0x97, 0xDA, 0x98, 0xB4, 0x01, 0xBB, 0xCD}}

	StructureChanged_Event_GUID = syscall.GUID{0x59977961, 0x3EDD, 0x4B11,
		[8]byte{0xB1, 0x3B, 0x67, 0x6B, 0x2A, 0x2A, 0x6C, 0xA9}}

	MenuOpened_Event_GUID = syscall.GUID{0xEBE2E945, 0x66CA, 0x4ED1,
		[8]byte{0x9F, 0xF8, 0x2A, 0xD7, 0xDF, 0x0A, 0x1B, 0x08}}

	AutomationPropertyChanged_Event_GUID = syscall.GUID{0x2527FBA1, 0x8D7A, 0x4630,
		[8]byte{0xA4, 0xCC, 0xE6, 0x63, 0x15, 0x94, 0x2F, 0x52}}

	AutomationFocusChanged_Event_GUID = syscall.GUID{0xB68A1F17, 0xF60D, 0x41A7,
		[8]byte{0xA3, 0xCC, 0xB0, 0x52, 0x92, 0x15, 0x5F, 0xE0}}

	ActiveTextPositionChanged_Event_GUID = syscall.GUID{0xA5C09E9C, 0xC77D, 0x4F25,
		[8]byte{0xB4, 0x91, 0xE5, 0xBB, 0x70, 0x17, 0xCB, 0xD4}}

	AsyncContentLoaded_Event_GUID = syscall.GUID{0x5FDEE11C, 0xD2FA, 0x4FB9,
		[8]byte{0x90, 0x4E, 0x5C, 0xBE, 0xE8, 0x94, 0xD5, 0xEF}}

	MenuClosed_Event_GUID = syscall.GUID{0x3CF1266E, 0x1582, 0x4041,
		[8]byte{0xAC, 0xD7, 0x88, 0xA3, 0x5A, 0x96, 0x52, 0x97}}

	LayoutInvalidated_Event_GUID = syscall.GUID{0xED7D6544, 0xA6BD, 0x4595,
		[8]byte{0x9B, 0xAE, 0x3D, 0x28, 0x94, 0x6C, 0xC7, 0x15}}

	Invoke_Invoked_Event_GUID = syscall.GUID{0xDFD699F0, 0xC915, 0x49DD,
		[8]byte{0xB4, 0x22, 0xDD, 0xE7, 0x85, 0xC3, 0xD2, 0x4B}}

	SelectionItem_ElementAddedToSelectionEvent_Event_GUID = syscall.GUID{0x3C822DD1, 0xC407, 0x4DBA,
		[8]byte{0x91, 0xDD, 0x79, 0xD4, 0xAE, 0xD0, 0xAE, 0xC6}}

	SelectionItem_ElementRemovedFromSelectionEvent_Event_GUID = syscall.GUID{0x097FA8A9, 0x7079, 0x41AF,
		[8]byte{0x8B, 0x9C, 0x09, 0x34, 0xD8, 0x30, 0x5E, 0x5C}}

	SelectionItem_ElementSelectedEvent_Event_GUID = syscall.GUID{0xB9C7DBFB, 0x4EBE, 0x4532,
		[8]byte{0xAA, 0xF4, 0x00, 0x8C, 0xF6, 0x47, 0x23, 0x3C}}

	Selection_InvalidatedEvent_Event_GUID = syscall.GUID{0xCAC14904, 0x16B4, 0x4B53,
		[8]byte{0x8E, 0x47, 0x4C, 0xB1, 0xDF, 0x26, 0x7B, 0xB7}}

	Text_TextSelectionChangedEvent_Event_GUID = syscall.GUID{0x918EDAA1, 0x71B3, 0x49AE,
		[8]byte{0x97, 0x41, 0x79, 0xBE, 0xB8, 0xD3, 0x58, 0xF3}}

	Text_TextChangedEvent_Event_GUID = syscall.GUID{0x4A342082, 0xF483, 0x48C4,
		[8]byte{0xAC, 0x11, 0xA8, 0x4B, 0x43, 0x5E, 0x2A, 0x84}}

	Window_WindowOpened_Event_GUID = syscall.GUID{0xD3E81D06, 0xDE45, 0x4F2F,
		[8]byte{0x96, 0x33, 0xDE, 0x9E, 0x02, 0xFB, 0x65, 0xAF}}

	Window_WindowClosed_Event_GUID = syscall.GUID{0xEDF141F8, 0xFA67, 0x4E22,
		[8]byte{0xBB, 0xF7, 0x94, 0x4E, 0x05, 0x73, 0x5E, 0xE2}}

	MenuModeStart_Event_GUID = syscall.GUID{0x18D7C631, 0x166A, 0x4AC9,
		[8]byte{0xAE, 0x3B, 0xEF, 0x4B, 0x54, 0x20, 0xE6, 0x81}}

	MenuModeEnd_Event_GUID = syscall.GUID{0x9ECD4C9F, 0x80DD, 0x47B8,
		[8]byte{0x82, 0x67, 0x5A, 0xEC, 0x06, 0xBB, 0x2C, 0xFF}}

	InputReachedTarget_Event_GUID = syscall.GUID{0x93ED549A, 0x0549, 0x40F0,
		[8]byte{0xBE, 0xDB, 0x28, 0xE4, 0x4F, 0x7D, 0xE2, 0xA3}}

	InputReachedOtherElement_Event_GUID = syscall.GUID{0xED201D8A, 0x4E6C, 0x415E,
		[8]byte{0xA8, 0x74, 0x24, 0x60, 0xC9, 0xB6, 0x6B, 0xA8}}

	InputDiscarded_Event_GUID = syscall.GUID{0x7F36C367, 0x7B18, 0x417C,
		[8]byte{0x97, 0xE3, 0x9D, 0x58, 0xDD, 0xC9, 0x44, 0xAB}}

	SystemAlert_Event_GUID = syscall.GUID{0xD271545D, 0x7A3A, 0x47A7,
		[8]byte{0x84, 0x74, 0x81, 0xD2, 0x9A, 0x24, 0x51, 0xC9}}

	LiveRegionChanged_Event_GUID = syscall.GUID{0x102D5E90, 0xE6A9, 0x41B6,
		[8]byte{0xB1, 0xC5, 0xA9, 0xB1, 0x92, 0x9D, 0x95, 0x10}}

	HostedFragmentRootsInvalidated_Event_GUID = syscall.GUID{0xE6BDB03E, 0x0921, 0x4EC5,
		[8]byte{0x8D, 0xCF, 0xEA, 0xE8, 0x77, 0xB0, 0x42, 0x6B}}

	Drag_DragStart_Event_GUID = syscall.GUID{0x883A480B, 0x3AA9, 0x429D,
		[8]byte{0x95, 0xE4, 0xD9, 0xC8, 0xD0, 0x11, 0xF0, 0xDD}}

	Drag_DragCancel_Event_GUID = syscall.GUID{0xC3EDE6FA, 0x3451, 0x4E0F,
		[8]byte{0x9E, 0x71, 0xDF, 0x9C, 0x28, 0x0A, 0x46, 0x57}}

	Drag_DragComplete_Event_GUID = syscall.GUID{0x38E96188, 0xEF1F, 0x463E,
		[8]byte{0x91, 0xCA, 0x3A, 0x77, 0x92, 0xC2, 0x9C, 0xAF}}

	DropTarget_DragEnter_Event_GUID = syscall.GUID{0xAAD9319B, 0x032C, 0x4A88,
		[8]byte{0x96, 0x1D, 0x1C, 0xF5, 0x79, 0x58, 0x1E, 0x34}}

	DropTarget_DragLeave_Event_GUID = syscall.GUID{0x0F82EB15, 0x24A2, 0x4988,
		[8]byte{0x92, 0x17, 0xDE, 0x16, 0x2A, 0xEE, 0x27, 0x2B}}

	DropTarget_Dropped_Event_GUID = syscall.GUID{0x622CEAD8, 0x1EDB, 0x4A3D,
		[8]byte{0xAB, 0xBC, 0xBE, 0x22, 0x11, 0xFF, 0x68, 0xB5}}

	StructuredMarkup_CompositionComplete_Event_GUID = syscall.GUID{0xC48A3C17, 0x677A, 0x4047,
		[8]byte{0xA6, 0x8D, 0xFC, 0x12, 0x57, 0x52, 0x8A, 0xEF}}

	StructuredMarkup_Deleted_Event_GUID = syscall.GUID{0xF9D0A020, 0xE1C1, 0x4ECF,
		[8]byte{0xB9, 0xAA, 0x52, 0xEF, 0xDE, 0x7E, 0x41, 0xE1}}

	StructuredMarkup_SelectionChanged_Event_GUID = syscall.GUID{0xA7C815F7, 0xFF9F, 0x41C7,
		[8]byte{0xA3, 0xA7, 0xAB, 0x6C, 0xBF, 0xDB, 0x49, 0x03}}

	Invoke_Pattern_GUID = syscall.GUID{0xD976C2FC, 0x66EA, 0x4A6E,
		[8]byte{0xB2, 0x8F, 0xC2, 0x4C, 0x75, 0x46, 0xAD, 0x37}}

	Selection_Pattern_GUID = syscall.GUID{0x66E3B7E8, 0xD821, 0x4D25,
		[8]byte{0x87, 0x61, 0x43, 0x5D, 0x2C, 0x8B, 0x25, 0x3F}}

	Value_Pattern_GUID = syscall.GUID{0x17FAAD9E, 0xC877, 0x475B,
		[8]byte{0xB9, 0x33, 0x77, 0x33, 0x27, 0x79, 0xB6, 0x37}}

	RangeValue_Pattern_GUID = syscall.GUID{0x18B00D87, 0xB1C9, 0x476A,
		[8]byte{0xBF, 0xBD, 0x5F, 0x0B, 0xDB, 0x92, 0x6F, 0x63}}

	Scroll_Pattern_GUID = syscall.GUID{0x895FA4B4, 0x759D, 0x4C50,
		[8]byte{0x8E, 0x15, 0x03, 0x46, 0x06, 0x72, 0x00, 0x3C}}

	ExpandCollapse_Pattern_GUID = syscall.GUID{0xAE05EFA2, 0xF9D1, 0x428A,
		[8]byte{0x83, 0x4C, 0x53, 0xA5, 0xC5, 0x2F, 0x9B, 0x8B}}

	Grid_Pattern_GUID = syscall.GUID{0x260A2CCB, 0x93A8, 0x4E44,
		[8]byte{0xA4, 0xC1, 0x3D, 0xF3, 0x97, 0xF2, 0xB0, 0x2B}}

	GridItem_Pattern_GUID = syscall.GUID{0xF2D5C877, 0xA462, 0x4957,
		[8]byte{0xA2, 0xA5, 0x2C, 0x96, 0xB3, 0x03, 0xBC, 0x63}}

	MultipleView_Pattern_GUID = syscall.GUID{0x547A6AE4, 0x113F, 0x47C4,
		[8]byte{0x85, 0x0F, 0xDB, 0x4D, 0xFA, 0x46, 0x6B, 0x1D}}

	Window_Pattern_GUID = syscall.GUID{0x27901735, 0xC760, 0x4994,
		[8]byte{0xAD, 0x11, 0x59, 0x19, 0xE6, 0x06, 0xB1, 0x10}}

	SelectionItem_Pattern_GUID = syscall.GUID{0x9BC64EEB, 0x87C7, 0x4B28,
		[8]byte{0x94, 0xBB, 0x4D, 0x9F, 0xA4, 0x37, 0xB6, 0xEF}}

	Dock_Pattern_GUID = syscall.GUID{0x9CBAA846, 0x83C8, 0x428D,
		[8]byte{0x82, 0x7F, 0x7E, 0x60, 0x63, 0xFE, 0x06, 0x20}}

	Table_Pattern_GUID = syscall.GUID{0xC415218E, 0xA028, 0x461E,
		[8]byte{0xAA, 0x92, 0x8F, 0x92, 0x5C, 0xF7, 0x93, 0x51}}

	TableItem_Pattern_GUID = syscall.GUID{0xDF1343BD, 0x1888, 0x4A29,
		[8]byte{0xA5, 0x0C, 0xB9, 0x2E, 0x6D, 0xE3, 0x7F, 0x6F}}

	Text_Pattern_GUID = syscall.GUID{0x8615F05D, 0x7DE5, 0x44FD,
		[8]byte{0xA6, 0x79, 0x2C, 0xA4, 0xB4, 0x60, 0x33, 0xA8}}

	Toggle_Pattern_GUID = syscall.GUID{0x0B419760, 0xE2F4, 0x43FF,
		[8]byte{0x8C, 0x5F, 0x94, 0x57, 0xC8, 0x2B, 0x56, 0xE9}}

	Transform_Pattern_GUID = syscall.GUID{0x24B46FDB, 0x587E, 0x49F1,
		[8]byte{0x9C, 0x4A, 0xD8, 0xE9, 0x8B, 0x66, 0x4B, 0x7B}}

	ScrollItem_Pattern_GUID = syscall.GUID{0x4591D005, 0xA803, 0x4D5C,
		[8]byte{0xB4, 0xD5, 0x8D, 0x28, 0x00, 0xF9, 0x06, 0xA7}}

	LegacyIAccessible_Pattern_GUID = syscall.GUID{0x54CC0A9F, 0x3395, 0x48AF,
		[8]byte{0xBA, 0x8D, 0x73, 0xF8, 0x56, 0x90, 0xF3, 0xE0}}

	ItemContainer_Pattern_GUID = syscall.GUID{0x3D13DA0F, 0x8B9A, 0x4A99,
		[8]byte{0x85, 0xFA, 0xC5, 0xC9, 0xA6, 0x9F, 0x1E, 0xD4}}

	VirtualizedItem_Pattern_GUID = syscall.GUID{0xF510173E, 0x2E71, 0x45E9,
		[8]byte{0xA6, 0xE5, 0x62, 0xF6, 0xED, 0x82, 0x89, 0xD5}}

	SynchronizedInput_Pattern_GUID = syscall.GUID{0x05C288A6, 0xC47B, 0x488B,
		[8]byte{0xB6, 0x53, 0x33, 0x97, 0x7A, 0x55, 0x1B, 0x8B}}

	ObjectModel_Pattern_GUID = syscall.GUID{0x3E04ACFE, 0x08FC, 0x47EC,
		[8]byte{0x96, 0xBC, 0x35, 0x3F, 0xA3, 0xB3, 0x4A, 0xA7}}

	Annotation_Pattern_GUID = syscall.GUID{0xF6C72AD7, 0x356C, 0x4850,
		[8]byte{0x92, 0x91, 0x31, 0x6F, 0x60, 0x8A, 0x8C, 0x84}}

	Text_Pattern2_GUID = syscall.GUID{0x498479A2, 0x5B22, 0x448D,
		[8]byte{0xB6, 0xE4, 0x64, 0x74, 0x90, 0x86, 0x06, 0x98}}

	TextEdit_Pattern_GUID = syscall.GUID{0x69F3FF89, 0x5AF9, 0x4C75,
		[8]byte{0x93, 0x40, 0xF2, 0xDE, 0x29, 0x2E, 0x45, 0x91}}

	CustomNavigation_Pattern_GUID = syscall.GUID{0xAFEA938A, 0x621E, 0x4054,
		[8]byte{0xBB, 0x2C, 0x2F, 0x46, 0x11, 0x4D, 0xAC, 0x3F}}

	Styles_Pattern_GUID = syscall.GUID{0x1AE62655, 0xDA72, 0x4D60,
		[8]byte{0xA1, 0x53, 0xE5, 0xAA, 0x69, 0x88, 0xE3, 0xBF}}

	Spreadsheet_Pattern_GUID = syscall.GUID{0x6A5B24C9, 0x9D1E, 0x4B85,
		[8]byte{0x9E, 0x44, 0xC0, 0x2E, 0x31, 0x69, 0xB1, 0x0B}}

	SpreadsheetItem_Pattern_GUID = syscall.GUID{0x32CF83FF, 0xF1A8, 0x4A8C,
		[8]byte{0x86, 0x58, 0xD4, 0x7B, 0xA7, 0x4E, 0x20, 0xBA}}

	Tranform_Pattern2_GUID = syscall.GUID{0x8AFCFD07, 0xA369, 0x44DE,
		[8]byte{0x98, 0x8B, 0x2F, 0x7F, 0xF4, 0x9F, 0xB8, 0xA8}}

	TextChild_Pattern_GUID = syscall.GUID{0x7533CAB7, 0x3BFE, 0x41EF,
		[8]byte{0x9E, 0x85, 0xE2, 0x63, 0x8C, 0xBE, 0x16, 0x9E}}

	Drag_Pattern_GUID = syscall.GUID{0xC0BEE21F, 0xCCB3, 0x4FED,
		[8]byte{0x99, 0x5B, 0x11, 0x4F, 0x6E, 0x3D, 0x27, 0x28}}

	DropTarget_Pattern_GUID = syscall.GUID{0x0BCBEC56, 0xBD34, 0x4B7B,
		[8]byte{0x9F, 0xD5, 0x26, 0x59, 0x90, 0x5E, 0xA3, 0xDC}}

	StructuredMarkup_Pattern_GUID = syscall.GUID{0xABBD0878, 0x8665, 0x4F5C,
		[8]byte{0x94, 0xFC, 0x36, 0xE7, 0xD8, 0xBB, 0x70, 0x6B}}

	Button_Control_GUID = syscall.GUID{0x5A78E369, 0xC6A1, 0x4F33,
		[8]byte{0xA9, 0xD7, 0x79, 0xF2, 0x0D, 0x0C, 0x78, 0x8E}}

	Calendar_Control_GUID = syscall.GUID{0x8913EB88, 0x00E5, 0x46BC,
		[8]byte{0x8E, 0x4E, 0x14, 0xA7, 0x86, 0xE1, 0x65, 0xA1}}

	CheckBox_Control_GUID = syscall.GUID{0xFB50F922, 0xA3DB, 0x49C0,
		[8]byte{0x8B, 0xC3, 0x06, 0xDA, 0xD5, 0x57, 0x78, 0xE2}}

	ComboBox_Control_GUID = syscall.GUID{0x54CB426C, 0x2F33, 0x4FFF,
		[8]byte{0xAA, 0xA1, 0xAE, 0xF6, 0x0D, 0xAC, 0x5D, 0xEB}}

	Edit_Control_GUID = syscall.GUID{0x6504A5C8, 0x2C86, 0x4F87,
		[8]byte{0xAE, 0x7B, 0x1A, 0xBD, 0xDC, 0x81, 0x0C, 0xF9}}

	Hyperlink_Control_GUID = syscall.GUID{0x8A56022C, 0xB00D, 0x4D15,
		[8]byte{0x8F, 0xF0, 0x5B, 0x6B, 0x26, 0x6E, 0x5E, 0x02}}

	Image_Control_GUID = syscall.GUID{0x2D3736E4, 0x6B16, 0x4C57,
		[8]byte{0xA9, 0x62, 0xF9, 0x32, 0x60, 0xA7, 0x52, 0x43}}

	ListItem_Control_GUID = syscall.GUID{0x7B3717F2, 0x44D1, 0x4A58,
		[8]byte{0x98, 0xA8, 0xF1, 0x2A, 0x9B, 0x8F, 0x78, 0xE2}}

	List_Control_GUID = syscall.GUID{0x9B149EE1, 0x7CCA, 0x4CFC,
		[8]byte{0x9A, 0xF1, 0xCA, 0xC7, 0xBD, 0xDD, 0x30, 0x31}}

	Menu_Control_GUID = syscall.GUID{0x2E9B1440, 0x0EA8, 0x41FD,
		[8]byte{0xB3, 0x74, 0xC1, 0xEA, 0x6F, 0x50, 0x3C, 0xD1}}

	MenuBar_Control_GUID = syscall.GUID{0xCC384250, 0x0E7B, 0x4AE8,
		[8]byte{0x95, 0xAE, 0xA0, 0x8F, 0x26, 0x1B, 0x52, 0xEE}}

	MenuItem_Control_GUID = syscall.GUID{0xF45225D3, 0xD0A0, 0x49D8,
		[8]byte{0x98, 0x34, 0x9A, 0x00, 0x0D, 0x2A, 0xED, 0xDC}}

	ProgressBar_Control_GUID = syscall.GUID{0x228C9F86, 0xC36C, 0x47BB,
		[8]byte{0x9F, 0xB6, 0xA5, 0x83, 0x4B, 0xFC, 0x53, 0xA4}}

	RadioButton_Control_GUID = syscall.GUID{0x3BDB49DB, 0xFE2C, 0x4483,
		[8]byte{0xB3, 0xE1, 0xE5, 0x7F, 0x21, 0x94, 0x40, 0xC6}}

	ScrollBar_Control_GUID = syscall.GUID{0xDAF34B36, 0x5065, 0x4946,
		[8]byte{0xB2, 0x2F, 0x92, 0x59, 0x5F, 0xC0, 0x75, 0x1A}}

	Slider_Control_GUID = syscall.GUID{0xB033C24B, 0x3B35, 0x4CEA,
		[8]byte{0xB6, 0x09, 0x76, 0x36, 0x82, 0xFA, 0x66, 0x0B}}

	Spinner_Control_GUID = syscall.GUID{0x60CC4B38, 0x3CB1, 0x4161,
		[8]byte{0xB4, 0x42, 0xC6, 0xB7, 0x26, 0xC1, 0x78, 0x25}}

	StatusBar_Control_GUID = syscall.GUID{0xD45E7D1B, 0x5873, 0x475F,
		[8]byte{0x95, 0xA4, 0x04, 0x33, 0xE1, 0xF1, 0xB0, 0x0A}}

	Tab_Control_GUID = syscall.GUID{0x38CD1F2D, 0x337A, 0x4BD2,
		[8]byte{0xA5, 0xE3, 0xAD, 0xB4, 0x69, 0xE3, 0x0B, 0xD3}}

	TabItem_Control_GUID = syscall.GUID{0x2C6A634F, 0x921B, 0x4E6E,
		[8]byte{0xB2, 0x6E, 0x08, 0xFC, 0xB0, 0x79, 0x8F, 0x4C}}

	Text_Control_GUID = syscall.GUID{0xAE9772DC, 0xD331, 0x4F09,
		[8]byte{0xBE, 0x20, 0x7E, 0x6D, 0xFA, 0xF0, 0x7B, 0x0A}}

	ToolBar_Control_GUID = syscall.GUID{0x8F06B751, 0xE182, 0x4E98,
		[8]byte{0x88, 0x93, 0x22, 0x84, 0x54, 0x3A, 0x7D, 0xCE}}

	ToolTip_Control_GUID = syscall.GUID{0x05DDC6D1, 0x2137, 0x4768,
		[8]byte{0x98, 0xEA, 0x73, 0xF5, 0x2F, 0x71, 0x34, 0xF3}}

	Tree_Control_GUID = syscall.GUID{0x7561349C, 0xD241, 0x43F4,
		[8]byte{0x99, 0x08, 0xB5, 0xF0, 0x91, 0xBE, 0xE6, 0x11}}

	TreeItem_Control_GUID = syscall.GUID{0x62C9FEB9, 0x8FFC, 0x4878,
		[8]byte{0xA3, 0xA4, 0x96, 0xB0, 0x30, 0x31, 0x5C, 0x18}}

	Custom_Control_GUID = syscall.GUID{0xF29EA0C3, 0xADB7, 0x430A,
		[8]byte{0xBA, 0x90, 0xE5, 0x2C, 0x73, 0x13, 0xE6, 0xED}}

	Group_Control_GUID = syscall.GUID{0xAD50AA1C, 0xE8C8, 0x4774,
		[8]byte{0xAE, 0x1B, 0xDD, 0x86, 0xDF, 0x0B, 0x3B, 0xDC}}

	Thumb_Control_GUID = syscall.GUID{0x701CA877, 0xE310, 0x4DD6,
		[8]byte{0xB6, 0x44, 0x79, 0x7E, 0x4F, 0xAE, 0xA2, 0x13}}

	DataGrid_Control_GUID = syscall.GUID{0x84B783AF, 0xD103, 0x4B0A,
		[8]byte{0x84, 0x15, 0xE7, 0x39, 0x42, 0x41, 0x0F, 0x4B}}

	DataItem_Control_GUID = syscall.GUID{0xA0177842, 0xD94F, 0x42A5,
		[8]byte{0x81, 0x4B, 0x60, 0x68, 0xAD, 0xDC, 0x8D, 0xA5}}

	Document_Control_GUID = syscall.GUID{0x3CD6BB6F, 0x6F08, 0x4562,
		[8]byte{0xB2, 0x29, 0xE4, 0xE2, 0xFC, 0x7A, 0x9E, 0xB4}}

	SplitButton_Control_GUID = syscall.GUID{0x7011F01F, 0x4ACE, 0x4901,
		[8]byte{0xB4, 0x61, 0x92, 0x0A, 0x6F, 0x1C, 0xA6, 0x50}}

	Window_Control_GUID = syscall.GUID{0xE13A7242, 0xF462, 0x4F4D,
		[8]byte{0xAE, 0xC1, 0x53, 0xB2, 0x8D, 0x6C, 0x32, 0x90}}

	Pane_Control_GUID = syscall.GUID{0x5C2B3F5B, 0x9182, 0x42A3,
		[8]byte{0x8D, 0xEC, 0x8C, 0x04, 0xC1, 0xEE, 0x63, 0x4D}}

	Header_Control_GUID = syscall.GUID{0x5B90CBCE, 0x78FB, 0x4614,
		[8]byte{0x82, 0xB6, 0x55, 0x4D, 0x74, 0x71, 0x8E, 0x67}}

	HeaderItem_Control_GUID = syscall.GUID{0xE6BC12CB, 0x7C8E, 0x49CF,
		[8]byte{0xB1, 0x68, 0x4A, 0x93, 0xA3, 0x2B, 0xEB, 0xB0}}

	Table_Control_GUID = syscall.GUID{0x773BFA0E, 0x5BC4, 0x4DEB,
		[8]byte{0x92, 0x1B, 0xDE, 0x7B, 0x32, 0x06, 0x22, 0x9E}}

	TitleBar_Control_GUID = syscall.GUID{0x98AA55BF, 0x3BB0, 0x4B65,
		[8]byte{0x83, 0x6E, 0x2E, 0xA3, 0x0D, 0xBC, 0x17, 0x1F}}

	Separator_Control_GUID = syscall.GUID{0x8767EBA3, 0x2A63, 0x4AB0,
		[8]byte{0xAC, 0x8D, 0xAA, 0x50, 0xE2, 0x3D, 0xE9, 0x78}}

	SemanticZoom_Control_GUID = syscall.GUID{0x5FD34A43, 0x061E, 0x42C8,
		[8]byte{0xB5, 0x89, 0x9D, 0xCC, 0xF7, 0x4B, 0xC4, 0x3A}}

	AppBar_Control_GUID = syscall.GUID{0x6114908D, 0xCC02, 0x4D37,
		[8]byte{0x87, 0x5B, 0xB5, 0x30, 0xC7, 0x13, 0x95, 0x54}}

	Text_AnimationStyle_Attribute_GUID = syscall.GUID{0x628209F0, 0x7C9A, 0x4D57,
		[8]byte{0xBE, 0x64, 0x1F, 0x18, 0x36, 0x57, 0x1F, 0xF5}}

	Text_BackgroundColor_Attribute_GUID = syscall.GUID{0xFDC49A07, 0x583D, 0x4F17,
		[8]byte{0xAD, 0x27, 0x77, 0xFC, 0x83, 0x2A, 0x3C, 0x0B}}

	Text_BulletStyle_Attribute_GUID = syscall.GUID{0xC1097C90, 0xD5C4, 0x4237,
		[8]byte{0x97, 0x81, 0x3B, 0xEC, 0x8B, 0xA5, 0x4E, 0x48}}

	Text_CapStyle_Attribute_GUID = syscall.GUID{0xFB059C50, 0x92CC, 0x49A5,
		[8]byte{0xBA, 0x8F, 0x0A, 0xA8, 0x72, 0xBB, 0xA2, 0xF3}}

	Text_Culture_Attribute_GUID = syscall.GUID{0xC2025AF9, 0xA42D, 0x4CED,
		[8]byte{0xA1, 0xFB, 0xC6, 0x74, 0x63, 0x15, 0x22, 0x2E}}

	Text_FontName_Attribute_GUID = syscall.GUID{0x64E63BA8, 0xF2E5, 0x476E,
		[8]byte{0xA4, 0x77, 0x17, 0x34, 0xFE, 0xAA, 0xF7, 0x26}}

	Text_FontSize_Attribute_GUID = syscall.GUID{0xDC5EEEFF, 0x0506, 0x4673,
		[8]byte{0x93, 0xF2, 0x37, 0x7E, 0x4A, 0x8E, 0x01, 0xF1}}

	Text_FontWeight_Attribute_GUID = syscall.GUID{0x6FC02359, 0xB316, 0x4F5F,
		[8]byte{0xB4, 0x01, 0xF1, 0xCE, 0x55, 0x74, 0x18, 0x53}}

	Text_ForegroundColor_Attribute_GUID = syscall.GUID{0x72D1C95D, 0x5E60, 0x471A,
		[8]byte{0x96, 0xB1, 0x6C, 0x1B, 0x3B, 0x77, 0xA4, 0x36}}

	Text_HorizontalTextAlignment_Attribute_GUID = syscall.GUID{0x04EA6161, 0xFBA3, 0x477A,
		[8]byte{0x95, 0x2A, 0xBB, 0x32, 0x6D, 0x02, 0x6A, 0x5B}}

	Text_IndentationFirstLine_Attribute_GUID = syscall.GUID{0x206F9AD5, 0xC1D3, 0x424A,
		[8]byte{0x81, 0x82, 0x6D, 0xA9, 0xA7, 0xF3, 0xD6, 0x32}}

	Text_IndentationLeading_Attribute_GUID = syscall.GUID{0x5CF66BAC, 0x2D45, 0x4A4B,
		[8]byte{0xB6, 0xC9, 0xF7, 0x22, 0x1D, 0x28, 0x15, 0xB0}}

	Text_IndentationTrailing_Attribute_GUID = syscall.GUID{0x97FF6C0F, 0x1CE4, 0x408A,
		[8]byte{0xB6, 0x7B, 0x94, 0xD8, 0x3E, 0xB6, 0x9B, 0xF2}}

	Text_IsHidden_Attribute_GUID = syscall.GUID{0x360182FB, 0xBDD7, 0x47F6,
		[8]byte{0xAB, 0x69, 0x19, 0xE3, 0x3F, 0x8A, 0x33, 0x44}}

	Text_IsItalic_Attribute_GUID = syscall.GUID{0xFCE12A56, 0x1336, 0x4A34,
		[8]byte{0x96, 0x63, 0x1B, 0xAB, 0x47, 0x23, 0x93, 0x20}}

	Text_IsReadOnly_Attribute_GUID = syscall.GUID{0xA738156B, 0xCA3E, 0x495E,
		[8]byte{0x95, 0x14, 0x83, 0x3C, 0x44, 0x0F, 0xEB, 0x11}}

	Text_IsSubscript_Attribute_GUID = syscall.GUID{0xF0EAD858, 0x8F53, 0x413C,
		[8]byte{0x87, 0x3F, 0x1A, 0x7D, 0x7F, 0x5E, 0x0D, 0xE4}}

	Text_IsSuperscript_Attribute_GUID = syscall.GUID{0xDA706EE4, 0xB3AA, 0x4645,
		[8]byte{0xA4, 0x1F, 0xCD, 0x25, 0x15, 0x7D, 0xEA, 0x76}}

	Text_MarginBottom_Attribute_GUID = syscall.GUID{0x7EE593C4, 0x72B4, 0x4CAC,
		[8]byte{0x92, 0x71, 0x3E, 0xD2, 0x4B, 0x0E, 0x4D, 0x42}}

	Text_MarginLeading_Attribute_GUID = syscall.GUID{0x9E9242D0, 0x5ED0, 0x4900,
		[8]byte{0x8E, 0x8A, 0xEE, 0xCC, 0x03, 0x83, 0x5A, 0xFC}}

	Text_MarginTop_Attribute_GUID = syscall.GUID{0x683D936F, 0xC9B9, 0x4A9A,
		[8]byte{0xB3, 0xD9, 0xD2, 0x0D, 0x33, 0x31, 0x1E, 0x2A}}

	Text_MarginTrailing_Attribute_GUID = syscall.GUID{0xAF522F98, 0x999D, 0x40AF,
		[8]byte{0xA5, 0xB2, 0x01, 0x69, 0xD0, 0x34, 0x20, 0x02}}

	Text_OutlineStyles_Attribute_GUID = syscall.GUID{0x5B675B27, 0xDB89, 0x46FE,
		[8]byte{0x97, 0x0C, 0x61, 0x4D, 0x52, 0x3B, 0xB9, 0x7D}}

	Text_OverlineColor_Attribute_GUID = syscall.GUID{0x83AB383A, 0xFD43, 0x40DA,
		[8]byte{0xAB, 0x3E, 0xEC, 0xF8, 0x16, 0x5C, 0xBB, 0x6D}}

	Text_OverlineStyle_Attribute_GUID = syscall.GUID{0x0A234D66, 0x617E, 0x427F,
		[8]byte{0x87, 0x1D, 0xE1, 0xFF, 0x1E, 0x0C, 0x21, 0x3F}}

	Text_StrikethroughColor_Attribute_GUID = syscall.GUID{0xBFE15A18, 0x8C41, 0x4C5A,
		[8]byte{0x9A, 0x0B, 0x04, 0xAF, 0x0E, 0x07, 0xF4, 0x87}}

	Text_StrikethroughStyle_Attribute_GUID = syscall.GUID{0x72913EF1, 0xDA00, 0x4F01,
		[8]byte{0x89, 0x9C, 0xAC, 0x5A, 0x85, 0x77, 0xA3, 0x07}}

	Text_Tabs_Attribute_GUID = syscall.GUID{0x2E68D00B, 0x92FE, 0x42D8,
		[8]byte{0x89, 0x9A, 0xA7, 0x84, 0xAA, 0x44, 0x54, 0xA1}}

	Text_TextFlowDirections_Attribute_GUID = syscall.GUID{0x8BDF8739, 0xF420, 0x423E,
		[8]byte{0xAF, 0x77, 0x20, 0xA5, 0xD9, 0x73, 0xA9, 0x07}}

	Text_UnderlineColor_Attribute_GUID = syscall.GUID{0xBFA12C73, 0xFDE2, 0x4473,
		[8]byte{0xBF, 0x64, 0x10, 0x36, 0xD6, 0xAA, 0x0F, 0x45}}

	Text_UnderlineStyle_Attribute_GUID = syscall.GUID{0x5F3B21C0, 0xEDE4, 0x44BD,
		[8]byte{0x9C, 0x36, 0x38, 0x53, 0x03, 0x8C, 0xBF, 0xEB}}

	Text_AnnotationTypes_Attribute_GUID = syscall.GUID{0xAD2EB431, 0xEE4E, 0x4BE1,
		[8]byte{0xA7, 0xBA, 0x55, 0x59, 0x15, 0x5A, 0x73, 0xEF}}

	Text_AnnotationObjects_Attribute_GUID = syscall.GUID{0xFF41CF68, 0xE7AB, 0x40B9,
		[8]byte{0x8C, 0x72, 0x72, 0xA8, 0xED, 0x94, 0x01, 0x7D}}

	Text_StyleName_Attribute_GUID = syscall.GUID{0x22C9E091, 0x4D66, 0x45D8,
		[8]byte{0xA8, 0x28, 0x73, 0x7B, 0xAB, 0x4C, 0x98, 0xA7}}

	Text_StyleId_Attribute_GUID = syscall.GUID{0x14C300DE, 0xC32B, 0x449B,
		[8]byte{0xAB, 0x7C, 0xB0, 0xE0, 0x78, 0x9A, 0xEA, 0x5D}}

	Text_Link_Attribute_GUID = syscall.GUID{0xB38EF51D, 0x9E8D, 0x4E46,
		[8]byte{0x91, 0x44, 0x56, 0xEB, 0xE1, 0x77, 0x32, 0x9B}}

	Text_IsActive_Attribute_GUID = syscall.GUID{0xF5A4E533, 0xE1B8, 0x436B,
		[8]byte{0x93, 0x5D, 0xB5, 0x7A, 0xA3, 0xF5, 0x58, 0xC4}}

	Text_SelectionActiveEnd_Attribute_GUID = syscall.GUID{0x1F668CC3, 0x9BBF, 0x416B,
		[8]byte{0xB0, 0xA2, 0xF8, 0x9F, 0x86, 0xF6, 0x61, 0x2C}}

	Text_CaretPosition_Attribute_GUID = syscall.GUID{0xB227B131, 0x9889, 0x4752,
		[8]byte{0xA9, 0x1B, 0x73, 0x3E, 0xFD, 0xC5, 0xC5, 0xA0}}

	Text_CaretBidiMode_Attribute_GUID = syscall.GUID{0x929EE7A6, 0x51D3, 0x4715,
		[8]byte{0x96, 0xDC, 0xB6, 0x94, 0xFA, 0x24, 0xA1, 0x68}}

	Text_BeforeParagraphSpacing_Attribute_GUID = syscall.GUID{0xBE7B0AB1, 0xC822, 0x4A24,
		[8]byte{0x85, 0xE9, 0xC8, 0xF2, 0x65, 0x0F, 0xC7, 0x9C}}

	Text_AfterParagraphSpacing_Attribute_GUID = syscall.GUID{0x588CBB38, 0xE62F, 0x497C,
		[8]byte{0xB5, 0xD1, 0xCC, 0xDF, 0x0E, 0xE8, 0x23, 0xD8}}

	Text_LineSpacing_Attribute_GUID = syscall.GUID{0x63FF70AE, 0xD943, 0x4B47,
		[8]byte{0x8A, 0xB7, 0xA7, 0xA0, 0x33, 0xD3, 0x21, 0x4B}}

	Text_BeforeSpacing_Attribute_GUID = syscall.GUID{0xBE7B0AB1, 0xC822, 0x4A24,
		[8]byte{0x85, 0xE9, 0xC8, 0xF2, 0x65, 0x0F, 0xC7, 0x9C}}

	Text_AfterSpacing_Attribute_GUID = syscall.GUID{0x588CBB38, 0xE62F, 0x497C,
		[8]byte{0xB5, 0xD1, 0xCC, 0xDF, 0x0E, 0xE8, 0x23, 0xD8}}

	Text_SayAsInterpretAs_Attribute_GUID = syscall.GUID{0xB38AD6AC, 0xEEE1, 0x4B6E,
		[8]byte{0x88, 0xCC, 0x01, 0x4C, 0xEF, 0xA9, 0x3F, 0xCB}}

	TextEdit_TextChanged_Event_GUID = syscall.GUID{0x120B0308, 0xEC22, 0x4EB8,
		[8]byte{0x9C, 0x98, 0x98, 0x67, 0xCD, 0xA1, 0xB1, 0x65}}

	TextEdit_ConversionTargetChanged_Event_GUID = syscall.GUID{0x3388C183, 0xED4F, 0x4C8B,
		[8]byte{0x9B, 0xAA, 0x36, 0x4D, 0x51, 0xD8, 0x84, 0x7F}}

	Changes_Event_GUID = syscall.GUID{0x7DF26714, 0x614F, 0x4E05,
		[8]byte{0x94, 0x88, 0x71, 0x6C, 0x5B, 0xA1, 0x94, 0x36}}

	Annotation_Custom_GUID = syscall.GUID{0x9EC82750, 0x3931, 0x4952,
		[8]byte{0x85, 0xBC, 0x1D, 0xBF, 0xF7, 0x8A, 0x43, 0xE3}}

	Annotation_SpellingError_GUID = syscall.GUID{0xAE85567E, 0x9ECE, 0x423F,
		[8]byte{0x81, 0xB7, 0x96, 0xC4, 0x3D, 0x53, 0xE5, 0x0E}}

	Annotation_GrammarError_GUID = syscall.GUID{0x757A048D, 0x4518, 0x41C6,
		[8]byte{0x85, 0x4C, 0xDC, 0x00, 0x9B, 0x7C, 0xFB, 0x53}}

	Annotation_Comment_GUID = syscall.GUID{0xFD2FDA30, 0x26B3, 0x4C06,
		[8]byte{0x8B, 0xC7, 0x98, 0xF1, 0x53, 0x2E, 0x46, 0xFD}}

	Annotation_FormulaError_GUID = syscall.GUID{0x95611982, 0x0CAB, 0x46D5,
		[8]byte{0xA2, 0xF0, 0xE3, 0x0D, 0x19, 0x05, 0xF8, 0xBF}}

	Annotation_TrackChanges_GUID = syscall.GUID{0x21E6E888, 0xDC14, 0x4016,
		[8]byte{0xAC, 0x27, 0x19, 0x05, 0x53, 0xC8, 0xC4, 0x70}}

	Annotation_Header_GUID = syscall.GUID{0x867B409B, 0xB216, 0x4472,
		[8]byte{0xA2, 0x19, 0x52, 0x5E, 0x31, 0x06, 0x81, 0xF8}}

	Annotation_Footer_GUID = syscall.GUID{0xCCEAB046, 0x1833, 0x47AA,
		[8]byte{0x80, 0x80, 0x70, 0x1E, 0xD0, 0xB0, 0xC8, 0x32}}

	Annotation_Highlighted_GUID = syscall.GUID{0x757C884E, 0x8083, 0x4081,
		[8]byte{0x8B, 0x9C, 0xE8, 0x7F, 0x50, 0x72, 0xF0, 0xE4}}

	Annotation_Endnote_GUID = syscall.GUID{0x7565725C, 0x2D99, 0x4839,
		[8]byte{0x96, 0x0D, 0x33, 0xD3, 0xB8, 0x66, 0xAB, 0xA5}}

	Annotation_Footnote_GUID = syscall.GUID{0x3DE10E21, 0x4125, 0x42DB,
		[8]byte{0x86, 0x20, 0xBE, 0x80, 0x83, 0x08, 0x06, 0x24}}

	Annotation_InsertionChange_GUID = syscall.GUID{0x0DBEB3A6, 0xDF15, 0x4164,
		[8]byte{0xA3, 0xC0, 0xE2, 0x1A, 0x8C, 0xE9, 0x31, 0xC4}}

	Annotation_DeletionChange_GUID = syscall.GUID{0xBE3D5B05, 0x951D, 0x42E7,
		[8]byte{0x90, 0x1D, 0xAD, 0xC8, 0xC2, 0xCF, 0x34, 0xD0}}

	Annotation_MoveChange_GUID = syscall.GUID{0x9DA587EB, 0x23E5, 0x4490,
		[8]byte{0xB3, 0x85, 0x1A, 0x22, 0xDD, 0xC8, 0xB1, 0x87}}

	Annotation_FormatChange_GUID = syscall.GUID{0xEB247345, 0xD4F1, 0x41CE,
		[8]byte{0x8E, 0x52, 0xF7, 0x9B, 0x69, 0x63, 0x5E, 0x48}}

	Annotation_UnsyncedChange_GUID = syscall.GUID{0x1851116A, 0x0E47, 0x4B30,
		[8]byte{0x8C, 0xB5, 0xD7, 0xDA, 0xE4, 0xFB, 0xCD, 0x1B}}

	Annotation_EditingLockedChange_GUID = syscall.GUID{0xC31F3E1C, 0x7423, 0x4DAC,
		[8]byte{0x83, 0x48, 0x41, 0xF0, 0x99, 0xFF, 0x6F, 0x64}}

	Annotation_ExternalChange_GUID = syscall.GUID{0x75A05B31, 0x5F11, 0x42FD,
		[8]byte{0x88, 0x7D, 0xDF, 0xA0, 0x10, 0xDB, 0x23, 0x92}}

	Annotation_ConflictingChange_GUID = syscall.GUID{0x98AF8802, 0x517C, 0x459F,
		[8]byte{0xAF, 0x13, 0x01, 0x6D, 0x3F, 0xAB, 0x87, 0x7E}}

	Annotation_Author_GUID = syscall.GUID{0xF161D3A7, 0xF81B, 0x4128,
		[8]byte{0xB1, 0x7F, 0x71, 0xF6, 0x90, 0x91, 0x45, 0x20}}

	Annotation_AdvancedProofingIssue_GUID = syscall.GUID{0xDAC7B72C, 0xC0F2, 0x4B84,
		[8]byte{0xB9, 0x0D, 0x5F, 0xAF, 0xC0, 0xF0, 0xEF, 0x1C}}

	Annotation_DataValidationError_GUID = syscall.GUID{0xC8649FA8, 0x9775, 0x437E,
		[8]byte{0xAD, 0x46, 0xE7, 0x09, 0xD9, 0x3C, 0x23, 0x43}}

	Annotation_CircularReferenceError_GUID = syscall.GUID{0x25BD9CF4, 0x1745, 0x4659,
		[8]byte{0xBA, 0x67, 0x72, 0x7F, 0x03, 0x18, 0xC6, 0x16}}

	Annotation_Mathematics_GUID = syscall.GUID{0xEAAB634B, 0x26D0, 0x40C1,
		[8]byte{0x80, 0x73, 0x57, 0xCA, 0x1C, 0x63, 0x3C, 0x9B}}

	Annotation_Sensitive_GUID = syscall.GUID{0x37F4C04F, 0x0F12, 0x4464,
		[8]byte{0x92, 0x9C, 0x82, 0x8F, 0xD1, 0x52, 0x92, 0xE3}}

	Changes_Summary_GUID = syscall.GUID{0x313D65A6, 0xE60F, 0x4D62,
		[8]byte{0x98, 0x61, 0x55, 0xAF, 0xD7, 0x28, 0xD2, 0x07}}

	StyleId_Custom_GUID = syscall.GUID{0xEF2EDD3E, 0xA999, 0x4B7C,
		[8]byte{0xA3, 0x78, 0x09, 0xBB, 0xD5, 0x2A, 0x35, 0x16}}

	StyleId_Heading1_GUID = syscall.GUID{0x7F7E8F69, 0x6866, 0x4621,
		[8]byte{0x93, 0x0C, 0x9A, 0x5D, 0x0C, 0xA5, 0x96, 0x1C}}

	StyleId_Heading2_GUID = syscall.GUID{0xBAA9B241, 0x5C69, 0x469D,
		[8]byte{0x85, 0xAD, 0x47, 0x47, 0x37, 0xB5, 0x2B, 0x14}}

	StyleId_Heading3_GUID = syscall.GUID{0xBF8BE9D2, 0xD8B8, 0x4EC5,
		[8]byte{0x8C, 0x52, 0x9C, 0xFB, 0x0D, 0x03, 0x59, 0x70}}

	StyleId_Heading4_GUID = syscall.GUID{0x8436FFC0, 0x9578, 0x45FC,
		[8]byte{0x83, 0xA4, 0xFF, 0x40, 0x05, 0x33, 0x15, 0xDD}}

	StyleId_Heading5_GUID = syscall.GUID{0x909F424D, 0x0DBF, 0x406E,
		[8]byte{0x97, 0xBB, 0x4E, 0x77, 0x3D, 0x97, 0x98, 0xF7}}

	StyleId_Heading6_GUID = syscall.GUID{0x89D23459, 0x5D5B, 0x4824,
		[8]byte{0xA4, 0x20, 0x11, 0xD3, 0xED, 0x82, 0xE4, 0x0F}}

	StyleId_Heading7_GUID = syscall.GUID{0xA3790473, 0xE9AE, 0x422D,
		[8]byte{0xB8, 0xE3, 0x3B, 0x67, 0x5C, 0x61, 0x81, 0xA4}}

	StyleId_Heading8_GUID = syscall.GUID{0x2BC14145, 0xA40C, 0x4881,
		[8]byte{0x84, 0xAE, 0xF2, 0x23, 0x56, 0x85, 0x38, 0x0C}}

	StyleId_Heading9_GUID = syscall.GUID{0xC70D9133, 0xBB2A, 0x43D3,
		[8]byte{0x8A, 0xC6, 0x33, 0x65, 0x78, 0x84, 0xB0, 0xF0}}

	StyleId_Title_GUID = syscall.GUID{0x15D8201A, 0xFFCF, 0x481F,
		[8]byte{0xB0, 0xA1, 0x30, 0xB6, 0x3B, 0xE9, 0x8F, 0x07}}

	StyleId_Subtitle_GUID = syscall.GUID{0xB5D9FC17, 0x5D6F, 0x4420,
		[8]byte{0xB4, 0x39, 0x7C, 0xB1, 0x9A, 0xD4, 0x34, 0xE2}}

	StyleId_Normal_GUID = syscall.GUID{0xCD14D429, 0xE45E, 0x4475,
		[8]byte{0xA1, 0xC5, 0x7F, 0x9E, 0x6B, 0xE9, 0x6E, 0xBA}}

	StyleId_Emphasis_GUID = syscall.GUID{0xCA6E7DBE, 0x355E, 0x4820,
		[8]byte{0x95, 0xA0, 0x92, 0x5F, 0x04, 0x1D, 0x34, 0x70}}

	StyleId_Quote_GUID = syscall.GUID{0x5D1C21EA, 0x8195, 0x4F6C,
		[8]byte{0x87, 0xEA, 0x5D, 0xAB, 0xEC, 0xE6, 0x4C, 0x1D}}

	StyleId_BulletedList_GUID = syscall.GUID{0x5963ED64, 0x6426, 0x4632,
		[8]byte{0x8C, 0xAF, 0xA3, 0x2A, 0xD4, 0x02, 0xD9, 0x1A}}

	StyleId_NumberedList_GUID = syscall.GUID{0x1E96DBD5, 0x64C3, 0x43D0,
		[8]byte{0xB1, 0xEE, 0xB5, 0x3B, 0x06, 0xE3, 0xED, 0xDF}}

	Notification_Event_GUID = syscall.GUID{0x72C5A2F7, 0x9788, 0x480F,
		[8]byte{0xB8, 0xEB, 0x4D, 0xEE, 0x00, 0xF6, 0x18, 0x6F}}

	SID_IsUIAutomationObject = syscall.GUID{0xB96FDB85, 0x7204, 0x4724,
		[8]byte{0x84, 0x2B, 0xC7, 0x05, 0x9D, 0xED, 0xB9, 0xD0}}

	SID_ControlElementProvider = syscall.GUID{0xF4791D68, 0xE254, 0x4BA3,
		[8]byte{0x9A, 0x53, 0x26, 0xA5, 0xC5, 0x49, 0x79, 0x46}}

	IsSelectionPattern2Available_Property_GUID = syscall.GUID{0x490806FB, 0x6E89, 0x4A47,
		[8]byte{0x83, 0x19, 0xD2, 0x66, 0xE5, 0x11, 0xF0, 0x21}}

	Selection2_FirstSelectedItem_Property_GUID = syscall.GUID{0xCC24EA67, 0x369C, 0x4E55,
		[8]byte{0x9F, 0xF7, 0x38, 0xDA, 0x69, 0x54, 0x0C, 0x29}}

	Selection2_LastSelectedItem_Property_GUID = syscall.GUID{0xCF7BDA90, 0x2D83, 0x49F8,
		[8]byte{0x86, 0x0C, 0x9C, 0xE3, 0x94, 0xCF, 0x89, 0xB4}}

	Selection2_CurrentSelectedItem_Property_GUID = syscall.GUID{0x34257C26, 0x83B5, 0x41A6,
		[8]byte{0x93, 0x9C, 0xAE, 0x84, 0x1C, 0x13, 0x62, 0x36}}

	Selection2_ItemCount_Property_GUID = syscall.GUID{0xBB49EB9F, 0x456D, 0x4048,
		[8]byte{0xB5, 0x91, 0x9C, 0x20, 0x26, 0xB8, 0x46, 0x36}}

	Selection_Pattern2_GUID = syscall.GUID{0xFBA25CAB, 0xAB98, 0x49F7,
		[8]byte{0xA7, 0xDC, 0xFE, 0x53, 0x9D, 0xC1, 0x5B, 0xE7}}

	HeadingLevel_Property_GUID = syscall.GUID{0x29084272, 0xAAAF, 0x4A30,
		[8]byte{0x87, 0x96, 0x3C, 0x12, 0xF6, 0x2B, 0x6B, 0xBB}}

	IsDialog_Property_GUID = syscall.GUID{0x9D0DFB9B, 0x8436, 0x4501,
		[8]byte{0xBB, 0xBB, 0xE5, 0x34, 0xA4, 0xFB, 0x3B, 0x3F}}
)

// enums

// enum
// flags
type STICKYKEYS_FLAGS uint32

const (
	SKF_STICKYKEYSON    STICKYKEYS_FLAGS = 1
	SKF_AVAILABLE       STICKYKEYS_FLAGS = 2
	SKF_HOTKEYACTIVE    STICKYKEYS_FLAGS = 4
	SKF_CONFIRMHOTKEY   STICKYKEYS_FLAGS = 8
	SKF_HOTKEYSOUND     STICKYKEYS_FLAGS = 16
	SKF_INDICATOR       STICKYKEYS_FLAGS = 32
	SKF_AUDIBLEFEEDBACK STICKYKEYS_FLAGS = 64
	SKF_TRISTATE        STICKYKEYS_FLAGS = 128
	SKF_TWOKEYSOFF      STICKYKEYS_FLAGS = 256
	SKF_LALTLATCHED     STICKYKEYS_FLAGS = 268435456
	SKF_LCTLLATCHED     STICKYKEYS_FLAGS = 67108864
	SKF_LSHIFTLATCHED   STICKYKEYS_FLAGS = 16777216
	SKF_RALTLATCHED     STICKYKEYS_FLAGS = 536870912
	SKF_RCTLLATCHED     STICKYKEYS_FLAGS = 134217728
	SKF_RSHIFTLATCHED   STICKYKEYS_FLAGS = 33554432
	SKF_LWINLATCHED     STICKYKEYS_FLAGS = 1073741824
	SKF_RWINLATCHED     STICKYKEYS_FLAGS = 2147483648
	SKF_LALTLOCKED      STICKYKEYS_FLAGS = 1048576
	SKF_LCTLLOCKED      STICKYKEYS_FLAGS = 262144
	SKF_LSHIFTLOCKED    STICKYKEYS_FLAGS = 65536
	SKF_RALTLOCKED      STICKYKEYS_FLAGS = 2097152
	SKF_RCTLLOCKED      STICKYKEYS_FLAGS = 524288
	SKF_RSHIFTLOCKED    STICKYKEYS_FLAGS = 131072
	SKF_LWINLOCKED      STICKYKEYS_FLAGS = 4194304
	SKF_RWINLOCKED      STICKYKEYS_FLAGS = 8388608
)

// enum
// flags
type SOUNDSENTRY_FLAGS uint32

const (
	SSF_SOUNDSENTRYON SOUNDSENTRY_FLAGS = 1
	SSF_AVAILABLE     SOUNDSENTRY_FLAGS = 2
	SSF_INDICATOR     SOUNDSENTRY_FLAGS = 4
)

// enum
// flags
type ACC_UTILITY_STATE_FLAGS uint32

const (
	ANRUS_ON_SCREEN_KEYBOARD_ACTIVE    ACC_UTILITY_STATE_FLAGS = 1
	ANRUS_TOUCH_MODIFICATION_ACTIVE    ACC_UTILITY_STATE_FLAGS = 2
	ANRUS_PRIORITY_AUDIO_ACTIVE        ACC_UTILITY_STATE_FLAGS = 4
	ANRUS_PRIORITY_AUDIO_ACTIVE_NODUCK ACC_UTILITY_STATE_FLAGS = 8
)

// enum
type SOUND_SENTRY_GRAPHICS_EFFECT uint32

const (
	SSGF_DISPLAY SOUND_SENTRY_GRAPHICS_EFFECT = 3
	SSGF_NONE    SOUND_SENTRY_GRAPHICS_EFFECT = 0
)

// enum
// flags
type SERIALKEYS_FLAGS uint32

const (
	SERKF_AVAILABLE    SERIALKEYS_FLAGS = 2
	SERKF_INDICATOR    SERIALKEYS_FLAGS = 4
	SERKF_SERIALKEYSON SERIALKEYS_FLAGS = 1
)

// enum
// flags
type HIGHCONTRASTW_FLAGS uint32

const (
	HCF_HIGHCONTRASTON       HIGHCONTRASTW_FLAGS = 1
	HCF_AVAILABLE            HIGHCONTRASTW_FLAGS = 2
	HCF_HOTKEYACTIVE         HIGHCONTRASTW_FLAGS = 4
	HCF_CONFIRMHOTKEY        HIGHCONTRASTW_FLAGS = 8
	HCF_HOTKEYSOUND          HIGHCONTRASTW_FLAGS = 16
	HCF_INDICATOR            HIGHCONTRASTW_FLAGS = 32
	HCF_HOTKEYAVAILABLE      HIGHCONTRASTW_FLAGS = 64
	HCF_OPTION_NOTHEMECHANGE HIGHCONTRASTW_FLAGS = 4096
)

// enum
type SOUNDSENTRY_TEXT_EFFECT uint32

const (
	SSTF_BORDER  SOUNDSENTRY_TEXT_EFFECT = 2
	SSTF_CHARS   SOUNDSENTRY_TEXT_EFFECT = 1
	SSTF_DISPLAY SOUNDSENTRY_TEXT_EFFECT = 3
	SSTF_NONE    SOUNDSENTRY_TEXT_EFFECT = 0
)

// enum
type SOUNDSENTRY_WINDOWS_EFFECT uint32

const (
	SSWF_CUSTOM  SOUNDSENTRY_WINDOWS_EFFECT = 4
	SSWF_DISPLAY SOUNDSENTRY_WINDOWS_EFFECT = 3
	SSWF_NONE    SOUNDSENTRY_WINDOWS_EFFECT = 0
	SSWF_TITLE   SOUNDSENTRY_WINDOWS_EFFECT = 1
	SSWF_WINDOW  SOUNDSENTRY_WINDOWS_EFFECT = 2
)

// enum
type AnnoScope int32

const (
	ANNO_THIS      AnnoScope = 0
	ANNO_CONTAINER AnnoScope = 1
)

// enum
type NavigateDirection int32

const (
	NavigateDirection_Parent          NavigateDirection = 0
	NavigateDirection_NextSibling     NavigateDirection = 1
	NavigateDirection_PreviousSibling NavigateDirection = 2
	NavigateDirection_FirstChild      NavigateDirection = 3
	NavigateDirection_LastChild       NavigateDirection = 4
)

// enum
type ProviderOptions int32

const (
	ProviderOptions_ClientSideProvider     ProviderOptions = 1
	ProviderOptions_ServerSideProvider     ProviderOptions = 2
	ProviderOptions_NonClientAreaProvider  ProviderOptions = 4
	ProviderOptions_OverrideProvider       ProviderOptions = 8
	ProviderOptions_ProviderOwnsSetFocus   ProviderOptions = 16
	ProviderOptions_UseComThreading        ProviderOptions = 32
	ProviderOptions_RefuseNonClientSupport ProviderOptions = 64
	ProviderOptions_HasNativeIAccessible   ProviderOptions = 128
	ProviderOptions_UseClientCoordinates   ProviderOptions = 256
)

// enum
type StructureChangeType int32

const (
	StructureChangeType_ChildAdded          StructureChangeType = 0
	StructureChangeType_ChildRemoved        StructureChangeType = 1
	StructureChangeType_ChildrenInvalidated StructureChangeType = 2
	StructureChangeType_ChildrenBulkAdded   StructureChangeType = 3
	StructureChangeType_ChildrenBulkRemoved StructureChangeType = 4
	StructureChangeType_ChildrenReordered   StructureChangeType = 5
)

// enum
type TextEditChangeType int32

const (
	TextEditChangeType_None                 TextEditChangeType = 0
	TextEditChangeType_AutoCorrect          TextEditChangeType = 1
	TextEditChangeType_Composition          TextEditChangeType = 2
	TextEditChangeType_CompositionFinalized TextEditChangeType = 3
	TextEditChangeType_AutoComplete         TextEditChangeType = 4
)

// enum
type OrientationType int32

const (
	OrientationType_None       OrientationType = 0
	OrientationType_Horizontal OrientationType = 1
	OrientationType_Vertical   OrientationType = 2
)

// enum
type DockPosition int32

const (
	DockPosition_Top    DockPosition = 0
	DockPosition_Left   DockPosition = 1
	DockPosition_Bottom DockPosition = 2
	DockPosition_Right  DockPosition = 3
	DockPosition_Fill   DockPosition = 4
	DockPosition_None   DockPosition = 5
)

// enum
type ExpandCollapseState int32

const (
	ExpandCollapseState_Collapsed         ExpandCollapseState = 0
	ExpandCollapseState_Expanded          ExpandCollapseState = 1
	ExpandCollapseState_PartiallyExpanded ExpandCollapseState = 2
	ExpandCollapseState_LeafNode          ExpandCollapseState = 3
)

// enum
type ScrollAmount int32

const (
	ScrollAmount_LargeDecrement ScrollAmount = 0
	ScrollAmount_SmallDecrement ScrollAmount = 1
	ScrollAmount_NoAmount       ScrollAmount = 2
	ScrollAmount_LargeIncrement ScrollAmount = 3
	ScrollAmount_SmallIncrement ScrollAmount = 4
)

// enum
type RowOrColumnMajor int32

const (
	RowOrColumnMajor_RowMajor      RowOrColumnMajor = 0
	RowOrColumnMajor_ColumnMajor   RowOrColumnMajor = 1
	RowOrColumnMajor_Indeterminate RowOrColumnMajor = 2
)

// enum
type ToggleState int32

const (
	ToggleState_Off           ToggleState = 0
	ToggleState_On            ToggleState = 1
	ToggleState_Indeterminate ToggleState = 2
)

// enum
type WindowVisualState int32

const (
	WindowVisualState_Normal    WindowVisualState = 0
	WindowVisualState_Maximized WindowVisualState = 1
	WindowVisualState_Minimized WindowVisualState = 2
)

// enum
type SynchronizedInputType int32

const (
	SynchronizedInputType_KeyUp          SynchronizedInputType = 1
	SynchronizedInputType_KeyDown        SynchronizedInputType = 2
	SynchronizedInputType_LeftMouseUp    SynchronizedInputType = 4
	SynchronizedInputType_LeftMouseDown  SynchronizedInputType = 8
	SynchronizedInputType_RightMouseUp   SynchronizedInputType = 16
	SynchronizedInputType_RightMouseDown SynchronizedInputType = 32
)

// enum
type WindowInteractionState int32

const (
	WindowInteractionState_Running                 WindowInteractionState = 0
	WindowInteractionState_Closing                 WindowInteractionState = 1
	WindowInteractionState_ReadyForUserInteraction WindowInteractionState = 2
	WindowInteractionState_BlockedByModalWindow    WindowInteractionState = 3
	WindowInteractionState_NotResponding           WindowInteractionState = 4
)

// enum
type SayAsInterpretAs int32

const (
	SayAsInterpretAs_None                       SayAsInterpretAs = 0
	SayAsInterpretAs_Spell                      SayAsInterpretAs = 1
	SayAsInterpretAs_Cardinal                   SayAsInterpretAs = 2
	SayAsInterpretAs_Ordinal                    SayAsInterpretAs = 3
	SayAsInterpretAs_Number                     SayAsInterpretAs = 4
	SayAsInterpretAs_Date                       SayAsInterpretAs = 5
	SayAsInterpretAs_Time                       SayAsInterpretAs = 6
	SayAsInterpretAs_Telephone                  SayAsInterpretAs = 7
	SayAsInterpretAs_Currency                   SayAsInterpretAs = 8
	SayAsInterpretAs_Net                        SayAsInterpretAs = 9
	SayAsInterpretAs_Url                        SayAsInterpretAs = 10
	SayAsInterpretAs_Address                    SayAsInterpretAs = 11
	SayAsInterpretAs_Alphanumeric               SayAsInterpretAs = 12
	SayAsInterpretAs_Name                       SayAsInterpretAs = 13
	SayAsInterpretAs_Media                      SayAsInterpretAs = 14
	SayAsInterpretAs_Date_MonthDayYear          SayAsInterpretAs = 15
	SayAsInterpretAs_Date_DayMonthYear          SayAsInterpretAs = 16
	SayAsInterpretAs_Date_YearMonthDay          SayAsInterpretAs = 17
	SayAsInterpretAs_Date_YearMonth             SayAsInterpretAs = 18
	SayAsInterpretAs_Date_MonthYear             SayAsInterpretAs = 19
	SayAsInterpretAs_Date_DayMonth              SayAsInterpretAs = 20
	SayAsInterpretAs_Date_MonthDay              SayAsInterpretAs = 21
	SayAsInterpretAs_Date_Year                  SayAsInterpretAs = 22
	SayAsInterpretAs_Time_HoursMinutesSeconds12 SayAsInterpretAs = 23
	SayAsInterpretAs_Time_HoursMinutes12        SayAsInterpretAs = 24
	SayAsInterpretAs_Time_HoursMinutesSeconds24 SayAsInterpretAs = 25
	SayAsInterpretAs_Time_HoursMinutes24        SayAsInterpretAs = 26
)

// enum
type TextUnit int32

const (
	TextUnit_Character TextUnit = 0
	TextUnit_Format    TextUnit = 1
	TextUnit_Word      TextUnit = 2
	TextUnit_Line      TextUnit = 3
	TextUnit_Paragraph TextUnit = 4
	TextUnit_Page      TextUnit = 5
	TextUnit_Document  TextUnit = 6
)

// enum
type TextPatternRangeEndpoint int32

const (
	TextPatternRangeEndpoint_Start TextPatternRangeEndpoint = 0
	TextPatternRangeEndpoint_End   TextPatternRangeEndpoint = 1
)

// enum
type SupportedTextSelection int32

const (
	SupportedTextSelection_None     SupportedTextSelection = 0
	SupportedTextSelection_Single   SupportedTextSelection = 1
	SupportedTextSelection_Multiple SupportedTextSelection = 2
)

// enum
type LiveSetting int32

const (
	Off       LiveSetting = 0
	Polite    LiveSetting = 1
	Assertive LiveSetting = 2
)

// enum
type ActiveEnd int32

const (
	ActiveEnd_None  ActiveEnd = 0
	ActiveEnd_Start ActiveEnd = 1
	ActiveEnd_End   ActiveEnd = 2
)

// enum
type CaretPosition int32

const (
	CaretPosition_Unknown         CaretPosition = 0
	CaretPosition_EndOfLine       CaretPosition = 1
	CaretPosition_BeginningOfLine CaretPosition = 2
)

// enum
type CaretBidiMode int32

const (
	CaretBidiMode_LTR CaretBidiMode = 0
	CaretBidiMode_RTL CaretBidiMode = 1
)

// enum
type ZoomUnit int32

const (
	ZoomUnit_NoAmount       ZoomUnit = 0
	ZoomUnit_LargeDecrement ZoomUnit = 1
	ZoomUnit_SmallDecrement ZoomUnit = 2
	ZoomUnit_LargeIncrement ZoomUnit = 3
	ZoomUnit_SmallIncrement ZoomUnit = 4
)

// enum
type AnimationStyle int32

const (
	AnimationStyle_None               AnimationStyle = 0
	AnimationStyle_LasVegasLights     AnimationStyle = 1
	AnimationStyle_BlinkingBackground AnimationStyle = 2
	AnimationStyle_SparkleText        AnimationStyle = 3
	AnimationStyle_MarchingBlackAnts  AnimationStyle = 4
	AnimationStyle_MarchingRedAnts    AnimationStyle = 5
	AnimationStyle_Shimmer            AnimationStyle = 6
	AnimationStyle_Other              AnimationStyle = -1
)

// enum
type BulletStyle int32

const (
	BulletStyle_None               BulletStyle = 0
	BulletStyle_HollowRoundBullet  BulletStyle = 1
	BulletStyle_FilledRoundBullet  BulletStyle = 2
	BulletStyle_HollowSquareBullet BulletStyle = 3
	BulletStyle_FilledSquareBullet BulletStyle = 4
	BulletStyle_DashBullet         BulletStyle = 5
	BulletStyle_Other              BulletStyle = -1
)

// enum
type CapStyle int32

const (
	CapStyle_None          CapStyle = 0
	CapStyle_SmallCap      CapStyle = 1
	CapStyle_AllCap        CapStyle = 2
	CapStyle_AllPetiteCaps CapStyle = 3
	CapStyle_PetiteCaps    CapStyle = 4
	CapStyle_Unicase       CapStyle = 5
	CapStyle_Titling       CapStyle = 6
	CapStyle_Other         CapStyle = -1
)

// enum
type FillType int32

const (
	FillType_None     FillType = 0
	FillType_Color    FillType = 1
	FillType_Gradient FillType = 2
	FillType_Picture  FillType = 3
	FillType_Pattern  FillType = 4
)

// enum
type FlowDirections int32

const (
	FlowDirections_Default     FlowDirections = 0
	FlowDirections_RightToLeft FlowDirections = 1
	FlowDirections_BottomToTop FlowDirections = 2
	FlowDirections_Vertical    FlowDirections = 4
)

// enum
type HorizontalTextAlignment int32

const (
	HorizontalTextAlignment_Left      HorizontalTextAlignment = 0
	HorizontalTextAlignment_Centered  HorizontalTextAlignment = 1
	HorizontalTextAlignment_Right     HorizontalTextAlignment = 2
	HorizontalTextAlignment_Justified HorizontalTextAlignment = 3
)

// enum
type OutlineStyles int32

const (
	OutlineStyles_None     OutlineStyles = 0
	OutlineStyles_Outline  OutlineStyles = 1
	OutlineStyles_Shadow   OutlineStyles = 2
	OutlineStyles_Engraved OutlineStyles = 4
	OutlineStyles_Embossed OutlineStyles = 8
)

// enum
type TextDecorationLineStyle int32

const (
	TextDecorationLineStyle_None            TextDecorationLineStyle = 0
	TextDecorationLineStyle_Single          TextDecorationLineStyle = 1
	TextDecorationLineStyle_WordsOnly       TextDecorationLineStyle = 2
	TextDecorationLineStyle_Double          TextDecorationLineStyle = 3
	TextDecorationLineStyle_Dot             TextDecorationLineStyle = 4
	TextDecorationLineStyle_Dash            TextDecorationLineStyle = 5
	TextDecorationLineStyle_DashDot         TextDecorationLineStyle = 6
	TextDecorationLineStyle_DashDotDot      TextDecorationLineStyle = 7
	TextDecorationLineStyle_Wavy            TextDecorationLineStyle = 8
	TextDecorationLineStyle_ThickSingle     TextDecorationLineStyle = 9
	TextDecorationLineStyle_DoubleWavy      TextDecorationLineStyle = 11
	TextDecorationLineStyle_ThickWavy       TextDecorationLineStyle = 12
	TextDecorationLineStyle_LongDash        TextDecorationLineStyle = 13
	TextDecorationLineStyle_ThickDash       TextDecorationLineStyle = 14
	TextDecorationLineStyle_ThickDashDot    TextDecorationLineStyle = 15
	TextDecorationLineStyle_ThickDashDotDot TextDecorationLineStyle = 16
	TextDecorationLineStyle_ThickDot        TextDecorationLineStyle = 17
	TextDecorationLineStyle_ThickLongDash   TextDecorationLineStyle = 18
	TextDecorationLineStyle_Other           TextDecorationLineStyle = -1
)

// enum
type VisualEffects int32

const (
	VisualEffects_None       VisualEffects = 0
	VisualEffects_Shadow     VisualEffects = 1
	VisualEffects_Reflection VisualEffects = 2
	VisualEffects_Glow       VisualEffects = 4
	VisualEffects_SoftEdges  VisualEffects = 8
	VisualEffects_Bevel      VisualEffects = 16
)

// enum
type NotificationProcessing int32

const (
	NotificationProcessing_ImportantAll          NotificationProcessing = 0
	NotificationProcessing_ImportantMostRecent   NotificationProcessing = 1
	NotificationProcessing_All                   NotificationProcessing = 2
	NotificationProcessing_MostRecent            NotificationProcessing = 3
	NotificationProcessing_CurrentThenMostRecent NotificationProcessing = 4
)

// enum
type NotificationKind int32

const (
	NotificationKind_ItemAdded       NotificationKind = 0
	NotificationKind_ItemRemoved     NotificationKind = 1
	NotificationKind_ActionCompleted NotificationKind = 2
	NotificationKind_ActionAborted   NotificationKind = 3
	NotificationKind_Other           NotificationKind = 4
)

// enum
type UIAutomationType int32

const (
	UIAutomationType_Int             UIAutomationType = 1
	UIAutomationType_Bool            UIAutomationType = 2
	UIAutomationType_String          UIAutomationType = 3
	UIAutomationType_Double          UIAutomationType = 4
	UIAutomationType_Point           UIAutomationType = 5
	UIAutomationType_Rect            UIAutomationType = 6
	UIAutomationType_Element         UIAutomationType = 7
	UIAutomationType_Array           UIAutomationType = 65536
	UIAutomationType_Out             UIAutomationType = 131072
	UIAutomationType_IntArray        UIAutomationType = 65537
	UIAutomationType_BoolArray       UIAutomationType = 65538
	UIAutomationType_StringArray     UIAutomationType = 65539
	UIAutomationType_DoubleArray     UIAutomationType = 65540
	UIAutomationType_PointArray      UIAutomationType = 65541
	UIAutomationType_RectArray       UIAutomationType = 65542
	UIAutomationType_ElementArray    UIAutomationType = 65543
	UIAutomationType_OutInt          UIAutomationType = 131073
	UIAutomationType_OutBool         UIAutomationType = 131074
	UIAutomationType_OutString       UIAutomationType = 131075
	UIAutomationType_OutDouble       UIAutomationType = 131076
	UIAutomationType_OutPoint        UIAutomationType = 131077
	UIAutomationType_OutRect         UIAutomationType = 131078
	UIAutomationType_OutElement      UIAutomationType = 131079
	UIAutomationType_OutIntArray     UIAutomationType = 196609
	UIAutomationType_OutBoolArray    UIAutomationType = 196610
	UIAutomationType_OutStringArray  UIAutomationType = 196611
	UIAutomationType_OutDoubleArray  UIAutomationType = 196612
	UIAutomationType_OutPointArray   UIAutomationType = 196613
	UIAutomationType_OutRectArray    UIAutomationType = 196614
	UIAutomationType_OutElementArray UIAutomationType = 196615
)

// enum
type TreeScope int32

const (
	TreeScope_None        TreeScope = 0
	TreeScope_Element     TreeScope = 1
	TreeScope_Children    TreeScope = 2
	TreeScope_Descendants TreeScope = 4
	TreeScope_Parent      TreeScope = 8
	TreeScope_Ancestors   TreeScope = 16
	TreeScope_Subtree     TreeScope = 7
)

// enum
type PropertyConditionFlags int32

const (
	PropertyConditionFlags_None           PropertyConditionFlags = 0
	PropertyConditionFlags_IgnoreCase     PropertyConditionFlags = 1
	PropertyConditionFlags_MatchSubstring PropertyConditionFlags = 2
)

// enum
type AutomationElementMode int32

const (
	AutomationElementMode_None AutomationElementMode = 0
	AutomationElementMode_Full AutomationElementMode = 1
)

// enum
type TreeTraversalOptions int32

const (
	TreeTraversalOptions_Default          TreeTraversalOptions = 0
	TreeTraversalOptions_PostOrder        TreeTraversalOptions = 1
	TreeTraversalOptions_LastToFirstOrder TreeTraversalOptions = 2
)

// enum
type ConnectionRecoveryBehaviorOptions int32

const (
	ConnectionRecoveryBehaviorOptions_Disabled ConnectionRecoveryBehaviorOptions = 0
	ConnectionRecoveryBehaviorOptions_Enabled  ConnectionRecoveryBehaviorOptions = 1
)

// enum
type CoalesceEventsOptions int32

const (
	CoalesceEventsOptions_Disabled CoalesceEventsOptions = 0
	CoalesceEventsOptions_Enabled  CoalesceEventsOptions = 1
)

// enum
type ConditionType int32

const (
	ConditionType_True     ConditionType = 0
	ConditionType_False    ConditionType = 1
	ConditionType_Property ConditionType = 2
	ConditionType_And      ConditionType = 3
	ConditionType_Or       ConditionType = 4
	ConditionType_Not      ConditionType = 5
)

// enum
type NormalizeState int32

const (
	NormalizeState_None   NormalizeState = 0
	NormalizeState_View   NormalizeState = 1
	NormalizeState_Custom NormalizeState = 2
)

// enum
type ProviderType int32

const (
	ProviderType_BaseHwnd      ProviderType = 0
	ProviderType_Proxy         ProviderType = 1
	ProviderType_NonClientArea ProviderType = 2
)

// enum
type AutomationIdentifierType int32

const (
	AutomationIdentifierType_Property      AutomationIdentifierType = 0
	AutomationIdentifierType_Pattern       AutomationIdentifierType = 1
	AutomationIdentifierType_Event         AutomationIdentifierType = 2
	AutomationIdentifierType_ControlType   AutomationIdentifierType = 3
	AutomationIdentifierType_TextAttribute AutomationIdentifierType = 4
	AutomationIdentifierType_LandmarkType  AutomationIdentifierType = 5
	AutomationIdentifierType_Annotation    AutomationIdentifierType = 6
	AutomationIdentifierType_Changes       AutomationIdentifierType = 7
	AutomationIdentifierType_Style         AutomationIdentifierType = 8
)

// enum
type EventArgsType int32

const (
	EventArgsType_Simple                    EventArgsType = 0
	EventArgsType_PropertyChanged           EventArgsType = 1
	EventArgsType_StructureChanged          EventArgsType = 2
	EventArgsType_AsyncContentLoaded        EventArgsType = 3
	EventArgsType_WindowClosed              EventArgsType = 4
	EventArgsType_TextEditTextChanged       EventArgsType = 5
	EventArgsType_Changes                   EventArgsType = 6
	EventArgsType_Notification              EventArgsType = 7
	EventArgsType_ActiveTextPositionChanged EventArgsType = 8
	EventArgsType_StructuredMarkup          EventArgsType = 9
)

// enum
type AsyncContentLoadedState int32

const (
	AsyncContentLoadedState_Beginning AsyncContentLoadedState = 0
	AsyncContentLoadedState_Progress  AsyncContentLoadedState = 1
	AsyncContentLoadedState_Completed AsyncContentLoadedState = 2
)

// structs

type CAccPropServices struct {
}

type MSAAMENUINFO struct {
	DwMSAASignature uint32
	CchWText        uint32
	PszWText        PWSTR
}

type CUIAutomation struct {
}

type CUIAutomation8 struct {
}

type CUIAutomationRegistrar struct {
}

type UiaRect struct {
	Left   float64
	Top    float64
	Width  float64
	Height float64
}

type UiaPoint struct {
	X float64
	Y float64
}

type UiaChangeInfo struct {
	UiaId     int32
	Payload   VARIANT
	ExtraInfo VARIANT
}

type UIAutomationParameter struct {
	Type_ UIAutomationType
	PData unsafe.Pointer
}

type UIAutomationPropertyInfo struct {
	Guid              syscall.GUID
	PProgrammaticName PWSTR
	Type_             UIAutomationType
}

type UIAutomationEventInfo struct {
	Guid              syscall.GUID
	PProgrammaticName PWSTR
}

type UIAutomationMethodInfo struct {
	PProgrammaticName PWSTR
	DoSetFocus        BOOL
	CInParameters     uint32
	COutParameters    uint32
	PParameterTypes   *UIAutomationType
	PParameterNames   *PWSTR
}

type UIAutomationPatternInfo struct {
	Guid                syscall.GUID
	PProgrammaticName   PWSTR
	ProviderInterfaceId syscall.GUID
	ClientInterfaceId   syscall.GUID
	CProperties         uint32
	PProperties         *UIAutomationPropertyInfo
	CMethods            uint32
	PMethods            *UIAutomationMethodInfo
	CEvents             uint32
	PEvents             *UIAutomationEventInfo
	PPatternHandler     *IUIAutomationPatternHandler
}

type ExtendedProperty struct {
	PropertyName  BSTR
	PropertyValue BSTR
}

type UiaCondition struct {
	ConditionType ConditionType
}

type UiaPropertyCondition struct {
	ConditionType ConditionType
	PropertyId    int32
	Value         VARIANT
	Flags         PropertyConditionFlags
}

type UiaAndOrCondition struct {
	ConditionType ConditionType
	PpConditions  **UiaCondition
	CConditions   int32
}

type UiaNotCondition struct {
	ConditionType ConditionType
	PCondition    *UiaCondition
}

type UiaCacheRequest struct {
	PViewCondition        *UiaCondition
	Scope                 TreeScope
	PProperties           *int32
	CProperties           int32
	PPatterns             *int32
	CPatterns             int32
	AutomationElementMode AutomationElementMode
}

type UiaFindParams struct {
	MaxDepth       int32
	FindFirst      BOOL
	ExcludeRoot    BOOL
	PFindCondition *UiaCondition
}

type UiaEventArgs struct {
	Type    EventArgsType
	EventId int32
}

type UiaPropertyChangedEventArgs struct {
	Type       EventArgsType
	EventId    int32
	PropertyId int32
	OldValue   VARIANT
	NewValue   VARIANT
}

type UiaStructureChangedEventArgs struct {
	Type                EventArgsType
	EventId             int32
	StructureChangeType StructureChangeType
	PRuntimeId          *int32
	CRuntimeIdLen       int32
}

type UiaTextEditTextChangedEventArgs struct {
	Type               EventArgsType
	EventId            int32
	TextEditChangeType TextEditChangeType
	PTextChange        *SAFEARRAY
}

type UiaChangesEventArgs struct {
	Type         EventArgsType
	EventId      int32
	EventIdCount int32
	PUiaChanges  *UiaChangeInfo
}

type UiaAsyncContentLoadedEventArgs struct {
	Type                    EventArgsType
	EventId                 int32
	AsyncContentLoadedState AsyncContentLoadedState
	PercentComplete         float64
}

type UiaWindowClosedEventArgs struct {
	Type          EventArgsType
	EventId       int32
	PRuntimeId    *int32
	CRuntimeIdLen int32
}

type SERIALKEYSA struct {
	CbSize         uint32
	DwFlags        SERIALKEYS_FLAGS
	LpszActivePort PSTR
	LpszPort       PSTR
	IBaudRate      uint32
	IPortState     uint32
	IActive        uint32
}

type SERIALKEYS = SERIALKEYSW
type SERIALKEYSW struct {
	CbSize         uint32
	DwFlags        SERIALKEYS_FLAGS
	LpszActivePort PWSTR
	LpszPort       PWSTR
	IBaudRate      uint32
	IPortState     uint32
	IActive        uint32
}

type HIGHCONTRASTA struct {
	CbSize            uint32
	DwFlags           HIGHCONTRASTW_FLAGS
	LpszDefaultScheme PSTR
}

type HIGHCONTRAST = HIGHCONTRASTW
type HIGHCONTRASTW struct {
	CbSize            uint32
	DwFlags           HIGHCONTRASTW_FLAGS
	LpszDefaultScheme PWSTR
}

type FILTERKEYS struct {
	CbSize      uint32
	DwFlags     uint32
	IWaitMSec   uint32
	IDelayMSec  uint32
	IRepeatMSec uint32
	IBounceMSec uint32
}

type STICKYKEYS struct {
	CbSize  uint32
	DwFlags STICKYKEYS_FLAGS
}

type MOUSEKEYS struct {
	CbSize          uint32
	DwFlags         uint32
	IMaxSpeed       uint32
	ITimeToMaxSpeed uint32
	ICtrlSpeed      uint32
	DwReserved1     uint32
	DwReserved2     uint32
}

type ACCESSTIMEOUT struct {
	CbSize       uint32
	DwFlags      uint32
	ITimeOutMSec uint32
}

type SOUNDSENTRYA struct {
	CbSize                 uint32
	DwFlags                SOUNDSENTRY_FLAGS
	IFSTextEffect          SOUNDSENTRY_TEXT_EFFECT
	IFSTextEffectMSec      uint32
	IFSTextEffectColorBits uint32
	IFSGrafEffect          SOUND_SENTRY_GRAPHICS_EFFECT
	IFSGrafEffectMSec      uint32
	IFSGrafEffectColor     uint32
	IWindowsEffect         SOUNDSENTRY_WINDOWS_EFFECT
	IWindowsEffectMSec     uint32
	LpszWindowsEffectDLL   PSTR
	IWindowsEffectOrdinal  uint32
}

type SOUNDSENTRY = SOUNDSENTRYW
type SOUNDSENTRYW struct {
	CbSize                 uint32
	DwFlags                SOUNDSENTRY_FLAGS
	IFSTextEffect          SOUNDSENTRY_TEXT_EFFECT
	IFSTextEffectMSec      uint32
	IFSTextEffectColorBits uint32
	IFSGrafEffect          SOUND_SENTRY_GRAPHICS_EFFECT
	IFSGrafEffectMSec      uint32
	IFSGrafEffectColor     uint32
	IWindowsEffect         SOUNDSENTRY_WINDOWS_EFFECT
	IWindowsEffectMSec     uint32
	LpszWindowsEffectDLL   PWSTR
	IWindowsEffectOrdinal  uint32
}

type TOGGLEKEYS struct {
	CbSize  uint32
	DwFlags uint32
}

// func types

type LPFNLRESULTFROMOBJECT = uintptr
type LPFNLRESULTFROMOBJECT_func = func(riid *syscall.GUID, wParam WPARAM, punk *IUnknown) LRESULT

type LPFNOBJECTFROMLRESULT = uintptr
type LPFNOBJECTFROMLRESULT_func = func(lResult LRESULT, riid *syscall.GUID, wParam WPARAM, ppvObject unsafe.Pointer) HRESULT

type LPFNACCESSIBLEOBJECTFROMWINDOW = uintptr
type LPFNACCESSIBLEOBJECTFROMWINDOW_func = func(hwnd HWND, dwId uint32, riid *syscall.GUID, ppvObject unsafe.Pointer) HRESULT

type LPFNACCESSIBLEOBJECTFROMPOINT = uintptr
type LPFNACCESSIBLEOBJECTFROMPOINT_func = func(ptScreen POINT, ppacc **IAccessible, pvarChild *VARIANT) HRESULT

type LPFNCREATESTDACCESSIBLEOBJECT = uintptr
type LPFNCREATESTDACCESSIBLEOBJECT_func = func(hwnd HWND, idObject int32, riid *syscall.GUID, ppvObject unsafe.Pointer) HRESULT

type LPFNACCESSIBLECHILDREN = uintptr
type LPFNACCESSIBLECHILDREN_func = func(paccContainer *IAccessible, iChildStart int32, cChildren int32, rgvarChildren *VARIANT, pcObtained *int32) HRESULT

type UiaProviderCallback = uintptr
type UiaProviderCallback_func = func(hwnd HWND, providerType ProviderType) *SAFEARRAY

type UiaEventCallback = uintptr
type UiaEventCallback_func = func(pArgs *UiaEventArgs, pRequestedData *SAFEARRAY, pTreeStructure BSTR)

type WINEVENTPROC = uintptr
type WINEVENTPROC_func = func(hWinEventHook HWINEVENTHOOK, event uint32, hwnd HWND, idObject int32, idChild int32, idEventThread uint32, dwmsEventTime uint32)

// interfaces

// 00000000-0000-0000-0000-000000000000
var IID_IRicheditWindowlessAccessibility = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type IRicheditWindowlessAccessibilityInterface interface {
	IUnknownInterface
	CreateProvider(pSite *IRawElementProviderWindowlessSite, ppProvider **IRawElementProviderSimple) HRESULT
}

type IRicheditWindowlessAccessibilityVtbl struct {
	IUnknownVtbl
	CreateProvider uintptr
}

type IRicheditWindowlessAccessibility struct {
	IUnknown
}

func (this *IRicheditWindowlessAccessibility) Vtbl() *IRicheditWindowlessAccessibilityVtbl {
	return (*IRicheditWindowlessAccessibilityVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRicheditWindowlessAccessibility) CreateProvider(pSite *IRawElementProviderWindowlessSite, ppProvider **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateProvider, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pSite)), uintptr(unsafe.Pointer(ppProvider)))
	return HRESULT(ret)
}

// 00000000-0000-0000-0000-000000000000
var IID_IRichEditUiaInformation = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type IRichEditUiaInformationInterface interface {
	IUnknownInterface
	GetBoundaryRectangle(pUiaRect *UiaRect) HRESULT
	IsVisible() HRESULT
}

type IRichEditUiaInformationVtbl struct {
	IUnknownVtbl
	GetBoundaryRectangle uintptr
	IsVisible            uintptr
}

type IRichEditUiaInformation struct {
	IUnknown
}

func (this *IRichEditUiaInformation) Vtbl() *IRichEditUiaInformationVtbl {
	return (*IRichEditUiaInformationVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRichEditUiaInformation) GetBoundaryRectangle(pUiaRect *UiaRect) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBoundaryRectangle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUiaRect)))
	return HRESULT(ret)
}

func (this *IRichEditUiaInformation) IsVisible() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsVisible, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 618736E0-3C3D-11CF-810C-00AA00389B71
var IID_IAccessible = syscall.GUID{0x618736E0, 0x3C3D, 0x11CF,
	[8]byte{0x81, 0x0C, 0x00, 0xAA, 0x00, 0x38, 0x9B, 0x71}}

type IAccessibleInterface interface {
	IDispatchInterface
	Get_accParent(ppdispParent **IDispatch) HRESULT
	Get_accChildCount(pcountChildren *int32) HRESULT
	Get_accChild(varChild VARIANT, ppdispChild **IDispatch) HRESULT
	Get_accName(varChild VARIANT, pszName *BSTR) HRESULT
	Get_accValue(varChild VARIANT, pszValue *BSTR) HRESULT
	Get_accDescription(varChild VARIANT, pszDescription *BSTR) HRESULT
	Get_accRole(varChild VARIANT, pvarRole *VARIANT) HRESULT
	Get_accState(varChild VARIANT, pvarState *VARIANT) HRESULT
	Get_accHelp(varChild VARIANT, pszHelp *BSTR) HRESULT
	Get_accHelpTopic(pszHelpFile *BSTR, varChild VARIANT, pidTopic *int32) HRESULT
	Get_accKeyboardShortcut(varChild VARIANT, pszKeyboardShortcut *BSTR) HRESULT
	Get_accFocus(pvarChild *VARIANT) HRESULT
	Get_accSelection(pvarChildren *VARIANT) HRESULT
	Get_accDefaultAction(varChild VARIANT, pszDefaultAction *BSTR) HRESULT
	AccSelect(flagsSelect int32, varChild VARIANT) HRESULT
	AccLocation(pxLeft *int32, pyTop *int32, pcxWidth *int32, pcyHeight *int32, varChild VARIANT) HRESULT
	AccNavigate(navDir int32, varStart VARIANT, pvarEndUpAt *VARIANT) HRESULT
	AccHitTest(xLeft int32, yTop int32, pvarChild *VARIANT) HRESULT
	AccDoDefaultAction(varChild VARIANT) HRESULT
	Put_accName(varChild VARIANT, szName BSTR) HRESULT
	Put_accValue(varChild VARIANT, szValue BSTR) HRESULT
}

type IAccessibleVtbl struct {
	IDispatchVtbl
	Get_accParent           uintptr
	Get_accChildCount       uintptr
	Get_accChild            uintptr
	Get_accName             uintptr
	Get_accValue            uintptr
	Get_accDescription      uintptr
	Get_accRole             uintptr
	Get_accState            uintptr
	Get_accHelp             uintptr
	Get_accHelpTopic        uintptr
	Get_accKeyboardShortcut uintptr
	Get_accFocus            uintptr
	Get_accSelection        uintptr
	Get_accDefaultAction    uintptr
	AccSelect               uintptr
	AccLocation             uintptr
	AccNavigate             uintptr
	AccHitTest              uintptr
	AccDoDefaultAction      uintptr
	Put_accName             uintptr
	Put_accValue            uintptr
}

type IAccessible struct {
	IDispatch
}

func (this *IAccessible) Vtbl() *IAccessibleVtbl {
	return (*IAccessibleVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAccessible) Get_accParent(ppdispParent **IDispatch) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_accParent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppdispParent)))
	return HRESULT(ret)
}

func (this *IAccessible) Get_accChildCount(pcountChildren *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_accChildCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcountChildren)))
	return HRESULT(ret)
}

func (this *IAccessible) Get_accChild(varChild VARIANT, ppdispChild **IDispatch) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_accChild, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&varChild)), uintptr(unsafe.Pointer(ppdispChild)))
	return HRESULT(ret)
}

func (this *IAccessible) Get_accName(varChild VARIANT, pszName *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_accName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&varChild)), uintptr(unsafe.Pointer(pszName)))
	return HRESULT(ret)
}

func (this *IAccessible) Get_accValue(varChild VARIANT, pszValue *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_accValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&varChild)), uintptr(unsafe.Pointer(pszValue)))
	return HRESULT(ret)
}

func (this *IAccessible) Get_accDescription(varChild VARIANT, pszDescription *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_accDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&varChild)), uintptr(unsafe.Pointer(pszDescription)))
	return HRESULT(ret)
}

func (this *IAccessible) Get_accRole(varChild VARIANT, pvarRole *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_accRole, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&varChild)), uintptr(unsafe.Pointer(pvarRole)))
	return HRESULT(ret)
}

func (this *IAccessible) Get_accState(varChild VARIANT, pvarState *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_accState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&varChild)), uintptr(unsafe.Pointer(pvarState)))
	return HRESULT(ret)
}

func (this *IAccessible) Get_accHelp(varChild VARIANT, pszHelp *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_accHelp, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&varChild)), uintptr(unsafe.Pointer(pszHelp)))
	return HRESULT(ret)
}

func (this *IAccessible) Get_accHelpTopic(pszHelpFile *BSTR, varChild VARIANT, pidTopic *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_accHelpTopic, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszHelpFile)), uintptr(unsafe.Pointer(&varChild)), uintptr(unsafe.Pointer(pidTopic)))
	return HRESULT(ret)
}

func (this *IAccessible) Get_accKeyboardShortcut(varChild VARIANT, pszKeyboardShortcut *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_accKeyboardShortcut, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&varChild)), uintptr(unsafe.Pointer(pszKeyboardShortcut)))
	return HRESULT(ret)
}

func (this *IAccessible) Get_accFocus(pvarChild *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_accFocus, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pvarChild)))
	return HRESULT(ret)
}

func (this *IAccessible) Get_accSelection(pvarChildren *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_accSelection, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pvarChildren)))
	return HRESULT(ret)
}

func (this *IAccessible) Get_accDefaultAction(varChild VARIANT, pszDefaultAction *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_accDefaultAction, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&varChild)), uintptr(unsafe.Pointer(pszDefaultAction)))
	return HRESULT(ret)
}

func (this *IAccessible) AccSelect(flagsSelect int32, varChild VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AccSelect, uintptr(unsafe.Pointer(this)), uintptr(flagsSelect), uintptr(unsafe.Pointer(&varChild)))
	return HRESULT(ret)
}

func (this *IAccessible) AccLocation(pxLeft *int32, pyTop *int32, pcxWidth *int32, pcyHeight *int32, varChild VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AccLocation, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pxLeft)), uintptr(unsafe.Pointer(pyTop)), uintptr(unsafe.Pointer(pcxWidth)), uintptr(unsafe.Pointer(pcyHeight)), uintptr(unsafe.Pointer(&varChild)))
	return HRESULT(ret)
}

func (this *IAccessible) AccNavigate(navDir int32, varStart VARIANT, pvarEndUpAt *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AccNavigate, uintptr(unsafe.Pointer(this)), uintptr(navDir), uintptr(unsafe.Pointer(&varStart)), uintptr(unsafe.Pointer(pvarEndUpAt)))
	return HRESULT(ret)
}

func (this *IAccessible) AccHitTest(xLeft int32, yTop int32, pvarChild *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AccHitTest, uintptr(unsafe.Pointer(this)), uintptr(xLeft), uintptr(yTop), uintptr(unsafe.Pointer(pvarChild)))
	return HRESULT(ret)
}

func (this *IAccessible) AccDoDefaultAction(varChild VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AccDoDefaultAction, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&varChild)))
	return HRESULT(ret)
}

func (this *IAccessible) Put_accName(varChild VARIANT, szName BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_accName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&varChild)), uintptr(unsafe.Pointer(szName)))
	return HRESULT(ret)
}

func (this *IAccessible) Put_accValue(varChild VARIANT, szValue BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_accValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&varChild)), uintptr(unsafe.Pointer(szValue)))
	return HRESULT(ret)
}

// 03022430-ABC4-11D0-BDE2-00AA001A1953
var IID_IAccessibleHandler = syscall.GUID{0x03022430, 0xABC4, 0x11D0,
	[8]byte{0xBD, 0xE2, 0x00, 0xAA, 0x00, 0x1A, 0x19, 0x53}}

type IAccessibleHandlerInterface interface {
	IUnknownInterface
	AccessibleObjectFromID(hwnd int32, lObjectID int32, pIAccessible **IAccessible) HRESULT
}

type IAccessibleHandlerVtbl struct {
	IUnknownVtbl
	AccessibleObjectFromID uintptr
}

type IAccessibleHandler struct {
	IUnknown
}

func (this *IAccessibleHandler) Vtbl() *IAccessibleHandlerVtbl {
	return (*IAccessibleHandlerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAccessibleHandler) AccessibleObjectFromID(hwnd int32, lObjectID int32, pIAccessible **IAccessible) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AccessibleObjectFromID, uintptr(unsafe.Pointer(this)), uintptr(hwnd), uintptr(lObjectID), uintptr(unsafe.Pointer(pIAccessible)))
	return HRESULT(ret)
}

// BF3ABD9C-76DA-4389-9EB6-1427D25ABAB7
var IID_IAccessibleWindowlessSite = syscall.GUID{0xBF3ABD9C, 0x76DA, 0x4389,
	[8]byte{0x9E, 0xB6, 0x14, 0x27, 0xD2, 0x5A, 0xBA, 0xB7}}

type IAccessibleWindowlessSiteInterface interface {
	IUnknownInterface
	AcquireObjectIdRange(rangeSize int32, pRangeOwner *IAccessibleHandler, pRangeBase *int32) HRESULT
	ReleaseObjectIdRange(rangeBase int32, pRangeOwner *IAccessibleHandler) HRESULT
	QueryObjectIdRanges(pRangesOwner *IAccessibleHandler, psaRanges **SAFEARRAY) HRESULT
	GetParentAccessible(ppParent **IAccessible) HRESULT
}

type IAccessibleWindowlessSiteVtbl struct {
	IUnknownVtbl
	AcquireObjectIdRange uintptr
	ReleaseObjectIdRange uintptr
	QueryObjectIdRanges  uintptr
	GetParentAccessible  uintptr
}

type IAccessibleWindowlessSite struct {
	IUnknown
}

func (this *IAccessibleWindowlessSite) Vtbl() *IAccessibleWindowlessSiteVtbl {
	return (*IAccessibleWindowlessSiteVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAccessibleWindowlessSite) AcquireObjectIdRange(rangeSize int32, pRangeOwner *IAccessibleHandler, pRangeBase *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AcquireObjectIdRange, uintptr(unsafe.Pointer(this)), uintptr(rangeSize), uintptr(unsafe.Pointer(pRangeOwner)), uintptr(unsafe.Pointer(pRangeBase)))
	return HRESULT(ret)
}

func (this *IAccessibleWindowlessSite) ReleaseObjectIdRange(rangeBase int32, pRangeOwner *IAccessibleHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReleaseObjectIdRange, uintptr(unsafe.Pointer(this)), uintptr(rangeBase), uintptr(unsafe.Pointer(pRangeOwner)))
	return HRESULT(ret)
}

func (this *IAccessibleWindowlessSite) QueryObjectIdRanges(pRangesOwner *IAccessibleHandler, psaRanges **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryObjectIdRanges, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRangesOwner)), uintptr(unsafe.Pointer(psaRanges)))
	return HRESULT(ret)
}

func (this *IAccessibleWindowlessSite) GetParentAccessible(ppParent **IAccessible) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetParentAccessible, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppParent)))
	return HRESULT(ret)
}

// 7852B78D-1CFD-41C1-A615-9C0C85960B5F
var IID_IAccIdentity = syscall.GUID{0x7852B78D, 0x1CFD, 0x41C1,
	[8]byte{0xA6, 0x15, 0x9C, 0x0C, 0x85, 0x96, 0x0B, 0x5F}}

type IAccIdentityInterface interface {
	IUnknownInterface
	GetIdentityString(dwIDChild uint32, ppIDString **byte, pdwIDStringLen *uint32) HRESULT
}

type IAccIdentityVtbl struct {
	IUnknownVtbl
	GetIdentityString uintptr
}

type IAccIdentity struct {
	IUnknown
}

func (this *IAccIdentity) Vtbl() *IAccIdentityVtbl {
	return (*IAccIdentityVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAccIdentity) GetIdentityString(dwIDChild uint32, ppIDString **byte, pdwIDStringLen *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetIdentityString, uintptr(unsafe.Pointer(this)), uintptr(dwIDChild), uintptr(unsafe.Pointer(ppIDString)), uintptr(unsafe.Pointer(pdwIDStringLen)))
	return HRESULT(ret)
}

// 76C0DBBB-15E0-4E7B-B61B-20EEEA2001E0
var IID_IAccPropServer = syscall.GUID{0x76C0DBBB, 0x15E0, 0x4E7B,
	[8]byte{0xB6, 0x1B, 0x20, 0xEE, 0xEA, 0x20, 0x01, 0xE0}}

type IAccPropServerInterface interface {
	IUnknownInterface
	GetPropValue(pIDString *byte, dwIDStringLen uint32, idProp syscall.GUID, pvarValue *VARIANT, pfHasProp *BOOL) HRESULT
}

type IAccPropServerVtbl struct {
	IUnknownVtbl
	GetPropValue uintptr
}

type IAccPropServer struct {
	IUnknown
}

func (this *IAccPropServer) Vtbl() *IAccPropServerVtbl {
	return (*IAccPropServerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAccPropServer) GetPropValue(pIDString *byte, dwIDStringLen uint32, idProp syscall.GUID, pvarValue *VARIANT, pfHasProp *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pIDString)), uintptr(dwIDStringLen), uintptr(unsafe.Pointer(&idProp)), uintptr(unsafe.Pointer(pvarValue)), uintptr(unsafe.Pointer(pfHasProp)))
	return HRESULT(ret)
}

// 6E26E776-04F0-495D-80E4-3330352E3169
var IID_IAccPropServices = syscall.GUID{0x6E26E776, 0x04F0, 0x495D,
	[8]byte{0x80, 0xE4, 0x33, 0x30, 0x35, 0x2E, 0x31, 0x69}}

type IAccPropServicesInterface interface {
	IUnknownInterface
	SetPropValue(pIDString *byte, dwIDStringLen uint32, idProp syscall.GUID, var_ VARIANT) HRESULT
	SetPropServer(pIDString *byte, dwIDStringLen uint32, paProps *syscall.GUID, cProps int32, pServer *IAccPropServer, annoScope AnnoScope) HRESULT
	ClearProps(pIDString *byte, dwIDStringLen uint32, paProps *syscall.GUID, cProps int32) HRESULT
	SetHwndProp(hwnd HWND, idObject uint32, idChild uint32, idProp syscall.GUID, var_ VARIANT) HRESULT
	SetHwndPropStr(hwnd HWND, idObject uint32, idChild uint32, idProp syscall.GUID, str PWSTR) HRESULT
	SetHwndPropServer(hwnd HWND, idObject uint32, idChild uint32, paProps *syscall.GUID, cProps int32, pServer *IAccPropServer, annoScope AnnoScope) HRESULT
	ClearHwndProps(hwnd HWND, idObject uint32, idChild uint32, paProps *syscall.GUID, cProps int32) HRESULT
	ComposeHwndIdentityString(hwnd HWND, idObject uint32, idChild uint32, ppIDString **byte, pdwIDStringLen *uint32) HRESULT
	DecomposeHwndIdentityString(pIDString *byte, dwIDStringLen uint32, phwnd *HWND, pidObject *uint32, pidChild *uint32) HRESULT
	SetHmenuProp(hmenu HMENU, idChild uint32, idProp syscall.GUID, var_ VARIANT) HRESULT
	SetHmenuPropStr(hmenu HMENU, idChild uint32, idProp syscall.GUID, str PWSTR) HRESULT
	SetHmenuPropServer(hmenu HMENU, idChild uint32, paProps *syscall.GUID, cProps int32, pServer *IAccPropServer, annoScope AnnoScope) HRESULT
	ClearHmenuProps(hmenu HMENU, idChild uint32, paProps *syscall.GUID, cProps int32) HRESULT
	ComposeHmenuIdentityString(hmenu HMENU, idChild uint32, ppIDString **byte, pdwIDStringLen *uint32) HRESULT
	DecomposeHmenuIdentityString(pIDString *byte, dwIDStringLen uint32, phmenu *HMENU, pidChild *uint32) HRESULT
}

type IAccPropServicesVtbl struct {
	IUnknownVtbl
	SetPropValue                 uintptr
	SetPropServer                uintptr
	ClearProps                   uintptr
	SetHwndProp                  uintptr
	SetHwndPropStr               uintptr
	SetHwndPropServer            uintptr
	ClearHwndProps               uintptr
	ComposeHwndIdentityString    uintptr
	DecomposeHwndIdentityString  uintptr
	SetHmenuProp                 uintptr
	SetHmenuPropStr              uintptr
	SetHmenuPropServer           uintptr
	ClearHmenuProps              uintptr
	ComposeHmenuIdentityString   uintptr
	DecomposeHmenuIdentityString uintptr
}

type IAccPropServices struct {
	IUnknown
}

func (this *IAccPropServices) Vtbl() *IAccPropServicesVtbl {
	return (*IAccPropServicesVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAccPropServices) SetPropValue(pIDString *byte, dwIDStringLen uint32, idProp syscall.GUID, var_ VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPropValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pIDString)), uintptr(dwIDStringLen), uintptr(unsafe.Pointer(&idProp)), uintptr(unsafe.Pointer(&var_)))
	return HRESULT(ret)
}

func (this *IAccPropServices) SetPropServer(pIDString *byte, dwIDStringLen uint32, paProps *syscall.GUID, cProps int32, pServer *IAccPropServer, annoScope AnnoScope) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPropServer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pIDString)), uintptr(dwIDStringLen), uintptr(unsafe.Pointer(paProps)), uintptr(cProps), uintptr(unsafe.Pointer(pServer)), uintptr(annoScope))
	return HRESULT(ret)
}

func (this *IAccPropServices) ClearProps(pIDString *byte, dwIDStringLen uint32, paProps *syscall.GUID, cProps int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ClearProps, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pIDString)), uintptr(dwIDStringLen), uintptr(unsafe.Pointer(paProps)), uintptr(cProps))
	return HRESULT(ret)
}

func (this *IAccPropServices) SetHwndProp(hwnd HWND, idObject uint32, idChild uint32, idProp syscall.GUID, var_ VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHwndProp, uintptr(unsafe.Pointer(this)), hwnd, uintptr(idObject), uintptr(idChild), uintptr(unsafe.Pointer(&idProp)), uintptr(unsafe.Pointer(&var_)))
	return HRESULT(ret)
}

func (this *IAccPropServices) SetHwndPropStr(hwnd HWND, idObject uint32, idChild uint32, idProp syscall.GUID, str PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHwndPropStr, uintptr(unsafe.Pointer(this)), hwnd, uintptr(idObject), uintptr(idChild), uintptr(unsafe.Pointer(&idProp)), uintptr(unsafe.Pointer(str)))
	return HRESULT(ret)
}

func (this *IAccPropServices) SetHwndPropServer(hwnd HWND, idObject uint32, idChild uint32, paProps *syscall.GUID, cProps int32, pServer *IAccPropServer, annoScope AnnoScope) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHwndPropServer, uintptr(unsafe.Pointer(this)), hwnd, uintptr(idObject), uintptr(idChild), uintptr(unsafe.Pointer(paProps)), uintptr(cProps), uintptr(unsafe.Pointer(pServer)), uintptr(annoScope))
	return HRESULT(ret)
}

func (this *IAccPropServices) ClearHwndProps(hwnd HWND, idObject uint32, idChild uint32, paProps *syscall.GUID, cProps int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ClearHwndProps, uintptr(unsafe.Pointer(this)), hwnd, uintptr(idObject), uintptr(idChild), uintptr(unsafe.Pointer(paProps)), uintptr(cProps))
	return HRESULT(ret)
}

func (this *IAccPropServices) ComposeHwndIdentityString(hwnd HWND, idObject uint32, idChild uint32, ppIDString **byte, pdwIDStringLen *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ComposeHwndIdentityString, uintptr(unsafe.Pointer(this)), hwnd, uintptr(idObject), uintptr(idChild), uintptr(unsafe.Pointer(ppIDString)), uintptr(unsafe.Pointer(pdwIDStringLen)))
	return HRESULT(ret)
}

func (this *IAccPropServices) DecomposeHwndIdentityString(pIDString *byte, dwIDStringLen uint32, phwnd *HWND, pidObject *uint32, pidChild *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DecomposeHwndIdentityString, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pIDString)), uintptr(dwIDStringLen), uintptr(unsafe.Pointer(phwnd)), uintptr(unsafe.Pointer(pidObject)), uintptr(unsafe.Pointer(pidChild)))
	return HRESULT(ret)
}

func (this *IAccPropServices) SetHmenuProp(hmenu HMENU, idChild uint32, idProp syscall.GUID, var_ VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHmenuProp, uintptr(unsafe.Pointer(this)), hmenu, uintptr(idChild), uintptr(unsafe.Pointer(&idProp)), uintptr(unsafe.Pointer(&var_)))
	return HRESULT(ret)
}

func (this *IAccPropServices) SetHmenuPropStr(hmenu HMENU, idChild uint32, idProp syscall.GUID, str PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHmenuPropStr, uintptr(unsafe.Pointer(this)), hmenu, uintptr(idChild), uintptr(unsafe.Pointer(&idProp)), uintptr(unsafe.Pointer(str)))
	return HRESULT(ret)
}

func (this *IAccPropServices) SetHmenuPropServer(hmenu HMENU, idChild uint32, paProps *syscall.GUID, cProps int32, pServer *IAccPropServer, annoScope AnnoScope) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHmenuPropServer, uintptr(unsafe.Pointer(this)), hmenu, uintptr(idChild), uintptr(unsafe.Pointer(paProps)), uintptr(cProps), uintptr(unsafe.Pointer(pServer)), uintptr(annoScope))
	return HRESULT(ret)
}

func (this *IAccPropServices) ClearHmenuProps(hmenu HMENU, idChild uint32, paProps *syscall.GUID, cProps int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ClearHmenuProps, uintptr(unsafe.Pointer(this)), hmenu, uintptr(idChild), uintptr(unsafe.Pointer(paProps)), uintptr(cProps))
	return HRESULT(ret)
}

func (this *IAccPropServices) ComposeHmenuIdentityString(hmenu HMENU, idChild uint32, ppIDString **byte, pdwIDStringLen *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ComposeHmenuIdentityString, uintptr(unsafe.Pointer(this)), hmenu, uintptr(idChild), uintptr(unsafe.Pointer(ppIDString)), uintptr(unsafe.Pointer(pdwIDStringLen)))
	return HRESULT(ret)
}

func (this *IAccPropServices) DecomposeHmenuIdentityString(pIDString *byte, dwIDStringLen uint32, phmenu *HMENU, pidChild *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DecomposeHmenuIdentityString, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pIDString)), uintptr(dwIDStringLen), uintptr(unsafe.Pointer(phmenu)), uintptr(unsafe.Pointer(pidChild)))
	return HRESULT(ret)
}

// D6DD68D1-86FD-4332-8666-9ABEDEA2D24C
var IID_IRawElementProviderSimple = syscall.GUID{0xD6DD68D1, 0x86FD, 0x4332,
	[8]byte{0x86, 0x66, 0x9A, 0xBE, 0xDE, 0xA2, 0xD2, 0x4C}}

type IRawElementProviderSimpleInterface interface {
	IUnknownInterface
	Get_ProviderOptions(pRetVal *ProviderOptions) HRESULT
	GetPatternProvider(patternId int32, pRetVal **IUnknown) HRESULT
	GetPropertyValue(propertyId int32, pRetVal *VARIANT) HRESULT
	Get_HostRawElementProvider(pRetVal **IRawElementProviderSimple) HRESULT
}

type IRawElementProviderSimpleVtbl struct {
	IUnknownVtbl
	Get_ProviderOptions        uintptr
	GetPatternProvider         uintptr
	GetPropertyValue           uintptr
	Get_HostRawElementProvider uintptr
}

type IRawElementProviderSimple struct {
	IUnknown
}

func (this *IRawElementProviderSimple) Vtbl() *IRawElementProviderSimpleVtbl {
	return (*IRawElementProviderSimpleVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRawElementProviderSimple) Get_ProviderOptions(pRetVal *ProviderOptions) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ProviderOptions, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IRawElementProviderSimple) GetPatternProvider(patternId int32, pRetVal **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPatternProvider, uintptr(unsafe.Pointer(this)), uintptr(patternId), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IRawElementProviderSimple) GetPropertyValue(propertyId int32, pRetVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyValue, uintptr(unsafe.Pointer(this)), uintptr(propertyId), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IRawElementProviderSimple) Get_HostRawElementProvider(pRetVal **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_HostRawElementProvider, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// F8B80ADA-2C44-48D0-89BE-5FF23C9CD875
var IID_IAccessibleEx = syscall.GUID{0xF8B80ADA, 0x2C44, 0x48D0,
	[8]byte{0x89, 0xBE, 0x5F, 0xF2, 0x3C, 0x9C, 0xD8, 0x75}}

type IAccessibleExInterface interface {
	IUnknownInterface
	GetObjectForChild(idChild int32, pRetVal **IAccessibleEx) HRESULT
	GetIAccessiblePair(ppAcc **IAccessible, pidChild *int32) HRESULT
	GetRuntimeId(pRetVal **SAFEARRAY) HRESULT
	ConvertReturnedElement(pIn *IRawElementProviderSimple, ppRetValOut **IAccessibleEx) HRESULT
}

type IAccessibleExVtbl struct {
	IUnknownVtbl
	GetObjectForChild      uintptr
	GetIAccessiblePair     uintptr
	GetRuntimeId           uintptr
	ConvertReturnedElement uintptr
}

type IAccessibleEx struct {
	IUnknown
}

func (this *IAccessibleEx) Vtbl() *IAccessibleExVtbl {
	return (*IAccessibleExVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAccessibleEx) GetObjectForChild(idChild int32, pRetVal **IAccessibleEx) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetObjectForChild, uintptr(unsafe.Pointer(this)), uintptr(idChild), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IAccessibleEx) GetIAccessiblePair(ppAcc **IAccessible, pidChild *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetIAccessiblePair, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppAcc)), uintptr(unsafe.Pointer(pidChild)))
	return HRESULT(ret)
}

func (this *IAccessibleEx) GetRuntimeId(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRuntimeId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IAccessibleEx) ConvertReturnedElement(pIn *IRawElementProviderSimple, ppRetValOut **IAccessibleEx) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertReturnedElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pIn)), uintptr(unsafe.Pointer(ppRetValOut)))
	return HRESULT(ret)
}

// A0A839A9-8DA1-4A82-806A-8E0D44E79F56
var IID_IRawElementProviderSimple2 = syscall.GUID{0xA0A839A9, 0x8DA1, 0x4A82,
	[8]byte{0x80, 0x6A, 0x8E, 0x0D, 0x44, 0xE7, 0x9F, 0x56}}

type IRawElementProviderSimple2Interface interface {
	IRawElementProviderSimpleInterface
	ShowContextMenu() HRESULT
}

type IRawElementProviderSimple2Vtbl struct {
	IRawElementProviderSimpleVtbl
	ShowContextMenu uintptr
}

type IRawElementProviderSimple2 struct {
	IRawElementProviderSimple
}

func (this *IRawElementProviderSimple2) Vtbl() *IRawElementProviderSimple2Vtbl {
	return (*IRawElementProviderSimple2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRawElementProviderSimple2) ShowContextMenu() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ShowContextMenu, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// FCF5D820-D7EC-4613-BDF6-42A84CE7DAAF
var IID_IRawElementProviderSimple3 = syscall.GUID{0xFCF5D820, 0xD7EC, 0x4613,
	[8]byte{0xBD, 0xF6, 0x42, 0xA8, 0x4C, 0xE7, 0xDA, 0xAF}}

type IRawElementProviderSimple3Interface interface {
	IRawElementProviderSimple2Interface
	GetMetadataValue(targetId int32, metadataId int32, returnVal *VARIANT) HRESULT
}

type IRawElementProviderSimple3Vtbl struct {
	IRawElementProviderSimple2Vtbl
	GetMetadataValue uintptr
}

type IRawElementProviderSimple3 struct {
	IRawElementProviderSimple2
}

func (this *IRawElementProviderSimple3) Vtbl() *IRawElementProviderSimple3Vtbl {
	return (*IRawElementProviderSimple3Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRawElementProviderSimple3) GetMetadataValue(targetId int32, metadataId int32, returnVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetMetadataValue, uintptr(unsafe.Pointer(this)), uintptr(targetId), uintptr(metadataId), uintptr(unsafe.Pointer(returnVal)))
	return HRESULT(ret)
}

// 620CE2A5-AB8F-40A9-86CB-DE3C75599B58
var IID_IRawElementProviderFragmentRoot = syscall.GUID{0x620CE2A5, 0xAB8F, 0x40A9,
	[8]byte{0x86, 0xCB, 0xDE, 0x3C, 0x75, 0x59, 0x9B, 0x58}}

type IRawElementProviderFragmentRootInterface interface {
	IUnknownInterface
	ElementProviderFromPoint(x float64, y float64, pRetVal **IRawElementProviderFragment) HRESULT
	GetFocus(pRetVal **IRawElementProviderFragment) HRESULT
}

type IRawElementProviderFragmentRootVtbl struct {
	IUnknownVtbl
	ElementProviderFromPoint uintptr
	GetFocus                 uintptr
}

type IRawElementProviderFragmentRoot struct {
	IUnknown
}

func (this *IRawElementProviderFragmentRoot) Vtbl() *IRawElementProviderFragmentRootVtbl {
	return (*IRawElementProviderFragmentRootVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRawElementProviderFragmentRoot) ElementProviderFromPoint(x float64, y float64, pRetVal **IRawElementProviderFragment) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ElementProviderFromPoint, uintptr(unsafe.Pointer(this)), uintptr(x), uintptr(y), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IRawElementProviderFragmentRoot) GetFocus(pRetVal **IRawElementProviderFragment) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFocus, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// F7063DA8-8359-439C-9297-BBC5299A7D87
var IID_IRawElementProviderFragment = syscall.GUID{0xF7063DA8, 0x8359, 0x439C,
	[8]byte{0x92, 0x97, 0xBB, 0xC5, 0x29, 0x9A, 0x7D, 0x87}}

type IRawElementProviderFragmentInterface interface {
	IUnknownInterface
	Navigate(direction NavigateDirection, pRetVal **IRawElementProviderFragment) HRESULT
	GetRuntimeId(pRetVal **SAFEARRAY) HRESULT
	Get_BoundingRectangle(pRetVal *UiaRect) HRESULT
	GetEmbeddedFragmentRoots(pRetVal **SAFEARRAY) HRESULT
	SetFocus() HRESULT
	Get_FragmentRoot(pRetVal **IRawElementProviderFragmentRoot) HRESULT
}

type IRawElementProviderFragmentVtbl struct {
	IUnknownVtbl
	Navigate                 uintptr
	GetRuntimeId             uintptr
	Get_BoundingRectangle    uintptr
	GetEmbeddedFragmentRoots uintptr
	SetFocus                 uintptr
	Get_FragmentRoot         uintptr
}

type IRawElementProviderFragment struct {
	IUnknown
}

func (this *IRawElementProviderFragment) Vtbl() *IRawElementProviderFragmentVtbl {
	return (*IRawElementProviderFragmentVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRawElementProviderFragment) Navigate(direction NavigateDirection, pRetVal **IRawElementProviderFragment) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Navigate, uintptr(unsafe.Pointer(this)), uintptr(direction), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IRawElementProviderFragment) GetRuntimeId(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRuntimeId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IRawElementProviderFragment) Get_BoundingRectangle(pRetVal *UiaRect) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_BoundingRectangle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IRawElementProviderFragment) GetEmbeddedFragmentRoots(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEmbeddedFragmentRoots, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IRawElementProviderFragment) SetFocus() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFocus, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IRawElementProviderFragment) Get_FragmentRoot(pRetVal **IRawElementProviderFragmentRoot) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_FragmentRoot, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// A407B27B-0F6D-4427-9292-473C7BF93258
var IID_IRawElementProviderAdviseEvents = syscall.GUID{0xA407B27B, 0x0F6D, 0x4427,
	[8]byte{0x92, 0x92, 0x47, 0x3C, 0x7B, 0xF9, 0x32, 0x58}}

type IRawElementProviderAdviseEventsInterface interface {
	IUnknownInterface
	AdviseEventAdded(eventId int32, propertyIDs *SAFEARRAY) HRESULT
	AdviseEventRemoved(eventId int32, propertyIDs *SAFEARRAY) HRESULT
}

type IRawElementProviderAdviseEventsVtbl struct {
	IUnknownVtbl
	AdviseEventAdded   uintptr
	AdviseEventRemoved uintptr
}

type IRawElementProviderAdviseEvents struct {
	IUnknown
}

func (this *IRawElementProviderAdviseEvents) Vtbl() *IRawElementProviderAdviseEventsVtbl {
	return (*IRawElementProviderAdviseEventsVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRawElementProviderAdviseEvents) AdviseEventAdded(eventId int32, propertyIDs *SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AdviseEventAdded, uintptr(unsafe.Pointer(this)), uintptr(eventId), uintptr(unsafe.Pointer(propertyIDs)))
	return HRESULT(ret)
}

func (this *IRawElementProviderAdviseEvents) AdviseEventRemoved(eventId int32, propertyIDs *SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AdviseEventRemoved, uintptr(unsafe.Pointer(this)), uintptr(eventId), uintptr(unsafe.Pointer(propertyIDs)))
	return HRESULT(ret)
}

// 1D5DF27C-8947-4425-B8D9-79787BB460B8
var IID_IRawElementProviderHwndOverride = syscall.GUID{0x1D5DF27C, 0x8947, 0x4425,
	[8]byte{0xB8, 0xD9, 0x79, 0x78, 0x7B, 0xB4, 0x60, 0xB8}}

type IRawElementProviderHwndOverrideInterface interface {
	IUnknownInterface
	GetOverrideProviderForHwnd(hwnd HWND, pRetVal **IRawElementProviderSimple) HRESULT
}

type IRawElementProviderHwndOverrideVtbl struct {
	IUnknownVtbl
	GetOverrideProviderForHwnd uintptr
}

type IRawElementProviderHwndOverride struct {
	IUnknown
}

func (this *IRawElementProviderHwndOverride) Vtbl() *IRawElementProviderHwndOverrideVtbl {
	return (*IRawElementProviderHwndOverrideVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRawElementProviderHwndOverride) GetOverrideProviderForHwnd(hwnd HWND, pRetVal **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetOverrideProviderForHwnd, uintptr(unsafe.Pointer(this)), hwnd, uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 4FD82B78-A43E-46AC-9803-0A6969C7C183
var IID_IProxyProviderWinEventSink = syscall.GUID{0x4FD82B78, 0xA43E, 0x46AC,
	[8]byte{0x98, 0x03, 0x0A, 0x69, 0x69, 0xC7, 0xC1, 0x83}}

type IProxyProviderWinEventSinkInterface interface {
	IUnknownInterface
	AddAutomationPropertyChangedEvent(pProvider *IRawElementProviderSimple, id int32, newValue VARIANT) HRESULT
	AddAutomationEvent(pProvider *IRawElementProviderSimple, id int32) HRESULT
	AddStructureChangedEvent(pProvider *IRawElementProviderSimple, structureChangeType StructureChangeType, runtimeId *SAFEARRAY) HRESULT
}

type IProxyProviderWinEventSinkVtbl struct {
	IUnknownVtbl
	AddAutomationPropertyChangedEvent uintptr
	AddAutomationEvent                uintptr
	AddStructureChangedEvent          uintptr
}

type IProxyProviderWinEventSink struct {
	IUnknown
}

func (this *IProxyProviderWinEventSink) Vtbl() *IProxyProviderWinEventSinkVtbl {
	return (*IProxyProviderWinEventSinkVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IProxyProviderWinEventSink) AddAutomationPropertyChangedEvent(pProvider *IRawElementProviderSimple, id int32, newValue VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddAutomationPropertyChangedEvent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pProvider)), uintptr(id), uintptr(unsafe.Pointer(&newValue)))
	return HRESULT(ret)
}

func (this *IProxyProviderWinEventSink) AddAutomationEvent(pProvider *IRawElementProviderSimple, id int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddAutomationEvent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pProvider)), uintptr(id))
	return HRESULT(ret)
}

func (this *IProxyProviderWinEventSink) AddStructureChangedEvent(pProvider *IRawElementProviderSimple, structureChangeType StructureChangeType, runtimeId *SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddStructureChangedEvent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pProvider)), uintptr(structureChangeType), uintptr(unsafe.Pointer(runtimeId)))
	return HRESULT(ret)
}

// 89592AD4-F4E0-43D5-A3B6-BAD7E111B435
var IID_IProxyProviderWinEventHandler = syscall.GUID{0x89592AD4, 0xF4E0, 0x43D5,
	[8]byte{0xA3, 0xB6, 0xBA, 0xD7, 0xE1, 0x11, 0xB4, 0x35}}

type IProxyProviderWinEventHandlerInterface interface {
	IUnknownInterface
	RespondToWinEvent(idWinEvent uint32, hwnd HWND, idObject int32, idChild int32, pSink *IProxyProviderWinEventSink) HRESULT
}

type IProxyProviderWinEventHandlerVtbl struct {
	IUnknownVtbl
	RespondToWinEvent uintptr
}

type IProxyProviderWinEventHandler struct {
	IUnknown
}

func (this *IProxyProviderWinEventHandler) Vtbl() *IProxyProviderWinEventHandlerVtbl {
	return (*IProxyProviderWinEventHandlerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IProxyProviderWinEventHandler) RespondToWinEvent(idWinEvent uint32, hwnd HWND, idObject int32, idChild int32, pSink *IProxyProviderWinEventSink) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RespondToWinEvent, uintptr(unsafe.Pointer(this)), uintptr(idWinEvent), hwnd, uintptr(idObject), uintptr(idChild), uintptr(unsafe.Pointer(pSink)))
	return HRESULT(ret)
}

// 0A2A93CC-BFAD-42AC-9B2E-0991FB0D3EA0
var IID_IRawElementProviderWindowlessSite = syscall.GUID{0x0A2A93CC, 0xBFAD, 0x42AC,
	[8]byte{0x9B, 0x2E, 0x09, 0x91, 0xFB, 0x0D, 0x3E, 0xA0}}

type IRawElementProviderWindowlessSiteInterface interface {
	IUnknownInterface
	GetAdjacentFragment(direction NavigateDirection, ppParent **IRawElementProviderFragment) HRESULT
	GetRuntimeIdPrefix(pRetVal **SAFEARRAY) HRESULT
}

type IRawElementProviderWindowlessSiteVtbl struct {
	IUnknownVtbl
	GetAdjacentFragment uintptr
	GetRuntimeIdPrefix  uintptr
}

type IRawElementProviderWindowlessSite struct {
	IUnknown
}

func (this *IRawElementProviderWindowlessSite) Vtbl() *IRawElementProviderWindowlessSiteVtbl {
	return (*IRawElementProviderWindowlessSiteVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRawElementProviderWindowlessSite) GetAdjacentFragment(direction NavigateDirection, ppParent **IRawElementProviderFragment) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAdjacentFragment, uintptr(unsafe.Pointer(this)), uintptr(direction), uintptr(unsafe.Pointer(ppParent)))
	return HRESULT(ret)
}

func (this *IRawElementProviderWindowlessSite) GetRuntimeIdPrefix(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRuntimeIdPrefix, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 33AC331B-943E-4020-B295-DB37784974A3
var IID_IAccessibleHostingElementProviders = syscall.GUID{0x33AC331B, 0x943E, 0x4020,
	[8]byte{0xB2, 0x95, 0xDB, 0x37, 0x78, 0x49, 0x74, 0xA3}}

type IAccessibleHostingElementProvidersInterface interface {
	IUnknownInterface
	GetEmbeddedFragmentRoots(pRetVal **SAFEARRAY) HRESULT
	GetObjectIdForProvider(pProvider *IRawElementProviderSimple, pidObject *int32) HRESULT
}

type IAccessibleHostingElementProvidersVtbl struct {
	IUnknownVtbl
	GetEmbeddedFragmentRoots uintptr
	GetObjectIdForProvider   uintptr
}

type IAccessibleHostingElementProviders struct {
	IUnknown
}

func (this *IAccessibleHostingElementProviders) Vtbl() *IAccessibleHostingElementProvidersVtbl {
	return (*IAccessibleHostingElementProvidersVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAccessibleHostingElementProviders) GetEmbeddedFragmentRoots(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEmbeddedFragmentRoots, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IAccessibleHostingElementProviders) GetObjectIdForProvider(pProvider *IRawElementProviderSimple, pidObject *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetObjectIdForProvider, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pProvider)), uintptr(unsafe.Pointer(pidObject)))
	return HRESULT(ret)
}

// 24BE0B07-D37D-487A-98CF-A13ED465E9B3
var IID_IRawElementProviderHostingAccessibles = syscall.GUID{0x24BE0B07, 0xD37D, 0x487A,
	[8]byte{0x98, 0xCF, 0xA1, 0x3E, 0xD4, 0x65, 0xE9, 0xB3}}

type IRawElementProviderHostingAccessiblesInterface interface {
	IUnknownInterface
	GetEmbeddedAccessibles(pRetVal **SAFEARRAY) HRESULT
}

type IRawElementProviderHostingAccessiblesVtbl struct {
	IUnknownVtbl
	GetEmbeddedAccessibles uintptr
}

type IRawElementProviderHostingAccessibles struct {
	IUnknown
}

func (this *IRawElementProviderHostingAccessibles) Vtbl() *IRawElementProviderHostingAccessiblesVtbl {
	return (*IRawElementProviderHostingAccessiblesVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRawElementProviderHostingAccessibles) GetEmbeddedAccessibles(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEmbeddedAccessibles, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 159BC72C-4AD3-485E-9637-D7052EDF0146
var IID_IDockProvider = syscall.GUID{0x159BC72C, 0x4AD3, 0x485E,
	[8]byte{0x96, 0x37, 0xD7, 0x05, 0x2E, 0xDF, 0x01, 0x46}}

type IDockProviderInterface interface {
	IUnknownInterface
	SetDockPosition(dockPosition DockPosition) HRESULT
	Get_DockPosition(pRetVal *DockPosition) HRESULT
}

type IDockProviderVtbl struct {
	IUnknownVtbl
	SetDockPosition  uintptr
	Get_DockPosition uintptr
}

type IDockProvider struct {
	IUnknown
}

func (this *IDockProvider) Vtbl() *IDockProviderVtbl {
	return (*IDockProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDockProvider) SetDockPosition(dockPosition DockPosition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDockPosition, uintptr(unsafe.Pointer(this)), uintptr(dockPosition))
	return HRESULT(ret)
}

func (this *IDockProvider) Get_DockPosition(pRetVal *DockPosition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_DockPosition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// D847D3A5-CAB0-4A98-8C32-ECB45C59AD24
var IID_IExpandCollapseProvider = syscall.GUID{0xD847D3A5, 0xCAB0, 0x4A98,
	[8]byte{0x8C, 0x32, 0xEC, 0xB4, 0x5C, 0x59, 0xAD, 0x24}}

type IExpandCollapseProviderInterface interface {
	IUnknownInterface
	Expand() HRESULT
	Collapse() HRESULT
	Get_ExpandCollapseState(pRetVal *ExpandCollapseState) HRESULT
}

type IExpandCollapseProviderVtbl struct {
	IUnknownVtbl
	Expand                  uintptr
	Collapse                uintptr
	Get_ExpandCollapseState uintptr
}

type IExpandCollapseProvider struct {
	IUnknown
}

func (this *IExpandCollapseProvider) Vtbl() *IExpandCollapseProviderVtbl {
	return (*IExpandCollapseProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IExpandCollapseProvider) Expand() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Expand, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IExpandCollapseProvider) Collapse() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Collapse, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IExpandCollapseProvider) Get_ExpandCollapseState(pRetVal *ExpandCollapseState) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ExpandCollapseState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// B17D6187-0907-464B-A168-0EF17A1572B1
var IID_IGridProvider = syscall.GUID{0xB17D6187, 0x0907, 0x464B,
	[8]byte{0xA1, 0x68, 0x0E, 0xF1, 0x7A, 0x15, 0x72, 0xB1}}

type IGridProviderInterface interface {
	IUnknownInterface
	GetItem(row int32, column int32, pRetVal **IRawElementProviderSimple) HRESULT
	Get_RowCount(pRetVal *int32) HRESULT
	Get_ColumnCount(pRetVal *int32) HRESULT
}

type IGridProviderVtbl struct {
	IUnknownVtbl
	GetItem         uintptr
	Get_RowCount    uintptr
	Get_ColumnCount uintptr
}

type IGridProvider struct {
	IUnknown
}

func (this *IGridProvider) Vtbl() *IGridProviderVtbl {
	return (*IGridProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IGridProvider) GetItem(row int32, column int32, pRetVal **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetItem, uintptr(unsafe.Pointer(this)), uintptr(row), uintptr(column), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IGridProvider) Get_RowCount(pRetVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_RowCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IGridProvider) Get_ColumnCount(pRetVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ColumnCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// D02541F1-FB81-4D64-AE32-F520F8A6DBD1
var IID_IGridItemProvider = syscall.GUID{0xD02541F1, 0xFB81, 0x4D64,
	[8]byte{0xAE, 0x32, 0xF5, 0x20, 0xF8, 0xA6, 0xDB, 0xD1}}

type IGridItemProviderInterface interface {
	IUnknownInterface
	Get_Row(pRetVal *int32) HRESULT
	Get_Column(pRetVal *int32) HRESULT
	Get_RowSpan(pRetVal *int32) HRESULT
	Get_ColumnSpan(pRetVal *int32) HRESULT
	Get_ContainingGrid(pRetVal **IRawElementProviderSimple) HRESULT
}

type IGridItemProviderVtbl struct {
	IUnknownVtbl
	Get_Row            uintptr
	Get_Column         uintptr
	Get_RowSpan        uintptr
	Get_ColumnSpan     uintptr
	Get_ContainingGrid uintptr
}

type IGridItemProvider struct {
	IUnknown
}

func (this *IGridItemProvider) Vtbl() *IGridItemProviderVtbl {
	return (*IGridItemProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IGridItemProvider) Get_Row(pRetVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Row, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IGridItemProvider) Get_Column(pRetVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Column, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IGridItemProvider) Get_RowSpan(pRetVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_RowSpan, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IGridItemProvider) Get_ColumnSpan(pRetVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ColumnSpan, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IGridItemProvider) Get_ContainingGrid(pRetVal **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ContainingGrid, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 54FCB24B-E18E-47A2-B4D3-ECCBE77599A2
var IID_IInvokeProvider = syscall.GUID{0x54FCB24B, 0xE18E, 0x47A2,
	[8]byte{0xB4, 0xD3, 0xEC, 0xCB, 0xE7, 0x75, 0x99, 0xA2}}

type IInvokeProviderInterface interface {
	IUnknownInterface
	Invoke() HRESULT
}

type IInvokeProviderVtbl struct {
	IUnknownVtbl
	Invoke uintptr
}

type IInvokeProvider struct {
	IUnknown
}

func (this *IInvokeProvider) Vtbl() *IInvokeProviderVtbl {
	return (*IInvokeProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IInvokeProvider) Invoke() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Invoke, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 6278CAB1-B556-4A1A-B4E0-418ACC523201
var IID_IMultipleViewProvider = syscall.GUID{0x6278CAB1, 0xB556, 0x4A1A,
	[8]byte{0xB4, 0xE0, 0x41, 0x8A, 0xCC, 0x52, 0x32, 0x01}}

type IMultipleViewProviderInterface interface {
	IUnknownInterface
	GetViewName(viewId int32, pRetVal *BSTR) HRESULT
	SetCurrentView(viewId int32) HRESULT
	Get_CurrentView(pRetVal *int32) HRESULT
	GetSupportedViews(pRetVal **SAFEARRAY) HRESULT
}

type IMultipleViewProviderVtbl struct {
	IUnknownVtbl
	GetViewName       uintptr
	SetCurrentView    uintptr
	Get_CurrentView   uintptr
	GetSupportedViews uintptr
}

type IMultipleViewProvider struct {
	IUnknown
}

func (this *IMultipleViewProvider) Vtbl() *IMultipleViewProviderVtbl {
	return (*IMultipleViewProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMultipleViewProvider) GetViewName(viewId int32, pRetVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetViewName, uintptr(unsafe.Pointer(this)), uintptr(viewId), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IMultipleViewProvider) SetCurrentView(viewId int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCurrentView, uintptr(unsafe.Pointer(this)), uintptr(viewId))
	return HRESULT(ret)
}

func (this *IMultipleViewProvider) Get_CurrentView(pRetVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentView, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IMultipleViewProvider) GetSupportedViews(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSupportedViews, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 36DC7AEF-33E6-4691-AFE1-2BE7274B3D33
var IID_IRangeValueProvider = syscall.GUID{0x36DC7AEF, 0x33E6, 0x4691,
	[8]byte{0xAF, 0xE1, 0x2B, 0xE7, 0x27, 0x4B, 0x3D, 0x33}}

type IRangeValueProviderInterface interface {
	IUnknownInterface
	SetValue(val float64) HRESULT
	Get_Value(pRetVal *float64) HRESULT
	Get_IsReadOnly(pRetVal *BOOL) HRESULT
	Get_Maximum(pRetVal *float64) HRESULT
	Get_Minimum(pRetVal *float64) HRESULT
	Get_LargeChange(pRetVal *float64) HRESULT
	Get_SmallChange(pRetVal *float64) HRESULT
}

type IRangeValueProviderVtbl struct {
	IUnknownVtbl
	SetValue        uintptr
	Get_Value       uintptr
	Get_IsReadOnly  uintptr
	Get_Maximum     uintptr
	Get_Minimum     uintptr
	Get_LargeChange uintptr
	Get_SmallChange uintptr
}

type IRangeValueProvider struct {
	IUnknown
}

func (this *IRangeValueProvider) Vtbl() *IRangeValueProviderVtbl {
	return (*IRangeValueProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRangeValueProvider) SetValue(val float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetValue, uintptr(unsafe.Pointer(this)), uintptr(val))
	return HRESULT(ret)
}

func (this *IRangeValueProvider) Get_Value(pRetVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Value, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IRangeValueProvider) Get_IsReadOnly(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_IsReadOnly, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IRangeValueProvider) Get_Maximum(pRetVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Maximum, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IRangeValueProvider) Get_Minimum(pRetVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Minimum, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IRangeValueProvider) Get_LargeChange(pRetVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_LargeChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IRangeValueProvider) Get_SmallChange(pRetVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_SmallChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 2360C714-4BF1-4B26-BA65-9B21316127EB
var IID_IScrollItemProvider = syscall.GUID{0x2360C714, 0x4BF1, 0x4B26,
	[8]byte{0xBA, 0x65, 0x9B, 0x21, 0x31, 0x61, 0x27, 0xEB}}

type IScrollItemProviderInterface interface {
	IUnknownInterface
	ScrollIntoView() HRESULT
}

type IScrollItemProviderVtbl struct {
	IUnknownVtbl
	ScrollIntoView uintptr
}

type IScrollItemProvider struct {
	IUnknown
}

func (this *IScrollItemProvider) Vtbl() *IScrollItemProviderVtbl {
	return (*IScrollItemProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IScrollItemProvider) ScrollIntoView() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ScrollIntoView, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// FB8B03AF-3BDF-48D4-BD36-1A65793BE168
var IID_ISelectionProvider = syscall.GUID{0xFB8B03AF, 0x3BDF, 0x48D4,
	[8]byte{0xBD, 0x36, 0x1A, 0x65, 0x79, 0x3B, 0xE1, 0x68}}

type ISelectionProviderInterface interface {
	IUnknownInterface
	GetSelection(pRetVal **SAFEARRAY) HRESULT
	Get_CanSelectMultiple(pRetVal *BOOL) HRESULT
	Get_IsSelectionRequired(pRetVal *BOOL) HRESULT
}

type ISelectionProviderVtbl struct {
	IUnknownVtbl
	GetSelection            uintptr
	Get_CanSelectMultiple   uintptr
	Get_IsSelectionRequired uintptr
}

type ISelectionProvider struct {
	IUnknown
}

func (this *ISelectionProvider) Vtbl() *ISelectionProviderVtbl {
	return (*ISelectionProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISelectionProvider) GetSelection(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSelection, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ISelectionProvider) Get_CanSelectMultiple(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CanSelectMultiple, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ISelectionProvider) Get_IsSelectionRequired(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_IsSelectionRequired, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 14F68475-EE1C-44F6-A869-D239381F0FE7
var IID_ISelectionProvider2 = syscall.GUID{0x14F68475, 0xEE1C, 0x44F6,
	[8]byte{0xA8, 0x69, 0xD2, 0x39, 0x38, 0x1F, 0x0F, 0xE7}}

type ISelectionProvider2Interface interface {
	ISelectionProviderInterface
	Get_FirstSelectedItem(retVal **IRawElementProviderSimple) HRESULT
	Get_LastSelectedItem(retVal **IRawElementProviderSimple) HRESULT
	Get_CurrentSelectedItem(retVal **IRawElementProviderSimple) HRESULT
	Get_ItemCount(retVal *int32) HRESULT
}

type ISelectionProvider2Vtbl struct {
	ISelectionProviderVtbl
	Get_FirstSelectedItem   uintptr
	Get_LastSelectedItem    uintptr
	Get_CurrentSelectedItem uintptr
	Get_ItemCount           uintptr
}

type ISelectionProvider2 struct {
	ISelectionProvider
}

func (this *ISelectionProvider2) Vtbl() *ISelectionProvider2Vtbl {
	return (*ISelectionProvider2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISelectionProvider2) Get_FirstSelectedItem(retVal **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_FirstSelectedItem, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *ISelectionProvider2) Get_LastSelectedItem(retVal **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_LastSelectedItem, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *ISelectionProvider2) Get_CurrentSelectedItem(retVal **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentSelectedItem, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *ISelectionProvider2) Get_ItemCount(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ItemCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// B38B8077-1FC3-42A5-8CAE-D40C2215055A
var IID_IScrollProvider = syscall.GUID{0xB38B8077, 0x1FC3, 0x42A5,
	[8]byte{0x8C, 0xAE, 0xD4, 0x0C, 0x22, 0x15, 0x05, 0x5A}}

type IScrollProviderInterface interface {
	IUnknownInterface
	Scroll(horizontalAmount ScrollAmount, verticalAmount ScrollAmount) HRESULT
	SetScrollPercent(horizontalPercent float64, verticalPercent float64) HRESULT
	Get_HorizontalScrollPercent(pRetVal *float64) HRESULT
	Get_VerticalScrollPercent(pRetVal *float64) HRESULT
	Get_HorizontalViewSize(pRetVal *float64) HRESULT
	Get_VerticalViewSize(pRetVal *float64) HRESULT
	Get_HorizontallyScrollable(pRetVal *BOOL) HRESULT
	Get_VerticallyScrollable(pRetVal *BOOL) HRESULT
}

type IScrollProviderVtbl struct {
	IUnknownVtbl
	Scroll                      uintptr
	SetScrollPercent            uintptr
	Get_HorizontalScrollPercent uintptr
	Get_VerticalScrollPercent   uintptr
	Get_HorizontalViewSize      uintptr
	Get_VerticalViewSize        uintptr
	Get_HorizontallyScrollable  uintptr
	Get_VerticallyScrollable    uintptr
}

type IScrollProvider struct {
	IUnknown
}

func (this *IScrollProvider) Vtbl() *IScrollProviderVtbl {
	return (*IScrollProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IScrollProvider) Scroll(horizontalAmount ScrollAmount, verticalAmount ScrollAmount) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Scroll, uintptr(unsafe.Pointer(this)), uintptr(horizontalAmount), uintptr(verticalAmount))
	return HRESULT(ret)
}

func (this *IScrollProvider) SetScrollPercent(horizontalPercent float64, verticalPercent float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetScrollPercent, uintptr(unsafe.Pointer(this)), uintptr(horizontalPercent), uintptr(verticalPercent))
	return HRESULT(ret)
}

func (this *IScrollProvider) Get_HorizontalScrollPercent(pRetVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_HorizontalScrollPercent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IScrollProvider) Get_VerticalScrollPercent(pRetVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_VerticalScrollPercent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IScrollProvider) Get_HorizontalViewSize(pRetVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_HorizontalViewSize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IScrollProvider) Get_VerticalViewSize(pRetVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_VerticalViewSize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IScrollProvider) Get_HorizontallyScrollable(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_HorizontallyScrollable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IScrollProvider) Get_VerticallyScrollable(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_VerticallyScrollable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 2ACAD808-B2D4-452D-A407-91FF1AD167B2
var IID_ISelectionItemProvider = syscall.GUID{0x2ACAD808, 0xB2D4, 0x452D,
	[8]byte{0xA4, 0x07, 0x91, 0xFF, 0x1A, 0xD1, 0x67, 0xB2}}

type ISelectionItemProviderInterface interface {
	IUnknownInterface
	Select() HRESULT
	AddToSelection() HRESULT
	RemoveFromSelection() HRESULT
	Get_IsSelected(pRetVal *BOOL) HRESULT
	Get_SelectionContainer(pRetVal **IRawElementProviderSimple) HRESULT
}

type ISelectionItemProviderVtbl struct {
	IUnknownVtbl
	Select                 uintptr
	AddToSelection         uintptr
	RemoveFromSelection    uintptr
	Get_IsSelected         uintptr
	Get_SelectionContainer uintptr
}

type ISelectionItemProvider struct {
	IUnknown
}

func (this *ISelectionItemProvider) Vtbl() *ISelectionItemProviderVtbl {
	return (*ISelectionItemProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISelectionItemProvider) Select() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Select, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ISelectionItemProvider) AddToSelection() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddToSelection, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ISelectionItemProvider) RemoveFromSelection() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveFromSelection, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ISelectionItemProvider) Get_IsSelected(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_IsSelected, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ISelectionItemProvider) Get_SelectionContainer(pRetVal **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_SelectionContainer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 29DB1A06-02CE-4CF7-9B42-565D4FAB20EE
var IID_ISynchronizedInputProvider = syscall.GUID{0x29DB1A06, 0x02CE, 0x4CF7,
	[8]byte{0x9B, 0x42, 0x56, 0x5D, 0x4F, 0xAB, 0x20, 0xEE}}

type ISynchronizedInputProviderInterface interface {
	IUnknownInterface
	StartListening(inputType SynchronizedInputType) HRESULT
	Cancel() HRESULT
}

type ISynchronizedInputProviderVtbl struct {
	IUnknownVtbl
	StartListening uintptr
	Cancel         uintptr
}

type ISynchronizedInputProvider struct {
	IUnknown
}

func (this *ISynchronizedInputProvider) Vtbl() *ISynchronizedInputProviderVtbl {
	return (*ISynchronizedInputProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISynchronizedInputProvider) StartListening(inputType SynchronizedInputType) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().StartListening, uintptr(unsafe.Pointer(this)), uintptr(inputType))
	return HRESULT(ret)
}

func (this *ISynchronizedInputProvider) Cancel() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Cancel, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 9C860395-97B3-490A-B52A-858CC22AF166
var IID_ITableProvider = syscall.GUID{0x9C860395, 0x97B3, 0x490A,
	[8]byte{0xB5, 0x2A, 0x85, 0x8C, 0xC2, 0x2A, 0xF1, 0x66}}

type ITableProviderInterface interface {
	IUnknownInterface
	GetRowHeaders(pRetVal **SAFEARRAY) HRESULT
	GetColumnHeaders(pRetVal **SAFEARRAY) HRESULT
	Get_RowOrColumnMajor(pRetVal *RowOrColumnMajor) HRESULT
}

type ITableProviderVtbl struct {
	IUnknownVtbl
	GetRowHeaders        uintptr
	GetColumnHeaders     uintptr
	Get_RowOrColumnMajor uintptr
}

type ITableProvider struct {
	IUnknown
}

func (this *ITableProvider) Vtbl() *ITableProviderVtbl {
	return (*ITableProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITableProvider) GetRowHeaders(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRowHeaders, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITableProvider) GetColumnHeaders(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetColumnHeaders, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITableProvider) Get_RowOrColumnMajor(pRetVal *RowOrColumnMajor) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_RowOrColumnMajor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// B9734FA6-771F-4D78-9C90-2517999349CD
var IID_ITableItemProvider = syscall.GUID{0xB9734FA6, 0x771F, 0x4D78,
	[8]byte{0x9C, 0x90, 0x25, 0x17, 0x99, 0x93, 0x49, 0xCD}}

type ITableItemProviderInterface interface {
	IUnknownInterface
	GetRowHeaderItems(pRetVal **SAFEARRAY) HRESULT
	GetColumnHeaderItems(pRetVal **SAFEARRAY) HRESULT
}

type ITableItemProviderVtbl struct {
	IUnknownVtbl
	GetRowHeaderItems    uintptr
	GetColumnHeaderItems uintptr
}

type ITableItemProvider struct {
	IUnknown
}

func (this *ITableItemProvider) Vtbl() *ITableItemProviderVtbl {
	return (*ITableItemProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITableItemProvider) GetRowHeaderItems(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRowHeaderItems, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITableItemProvider) GetColumnHeaderItems(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetColumnHeaderItems, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 56D00BD0-C4F4-433C-A836-1A52A57E0892
var IID_IToggleProvider = syscall.GUID{0x56D00BD0, 0xC4F4, 0x433C,
	[8]byte{0xA8, 0x36, 0x1A, 0x52, 0xA5, 0x7E, 0x08, 0x92}}

type IToggleProviderInterface interface {
	IUnknownInterface
	Toggle() HRESULT
	Get_ToggleState(pRetVal *ToggleState) HRESULT
}

type IToggleProviderVtbl struct {
	IUnknownVtbl
	Toggle          uintptr
	Get_ToggleState uintptr
}

type IToggleProvider struct {
	IUnknown
}

func (this *IToggleProvider) Vtbl() *IToggleProviderVtbl {
	return (*IToggleProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IToggleProvider) Toggle() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Toggle, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IToggleProvider) Get_ToggleState(pRetVal *ToggleState) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ToggleState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 6829DDC4-4F91-4FFA-B86F-BD3E2987CB4C
var IID_ITransformProvider = syscall.GUID{0x6829DDC4, 0x4F91, 0x4FFA,
	[8]byte{0xB8, 0x6F, 0xBD, 0x3E, 0x29, 0x87, 0xCB, 0x4C}}

type ITransformProviderInterface interface {
	IUnknownInterface
	Move(x float64, y float64) HRESULT
	Resize(width float64, height float64) HRESULT
	Rotate(degrees float64) HRESULT
	Get_CanMove(pRetVal *BOOL) HRESULT
	Get_CanResize(pRetVal *BOOL) HRESULT
	Get_CanRotate(pRetVal *BOOL) HRESULT
}

type ITransformProviderVtbl struct {
	IUnknownVtbl
	Move          uintptr
	Resize        uintptr
	Rotate        uintptr
	Get_CanMove   uintptr
	Get_CanResize uintptr
	Get_CanRotate uintptr
}

type ITransformProvider struct {
	IUnknown
}

func (this *ITransformProvider) Vtbl() *ITransformProviderVtbl {
	return (*ITransformProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITransformProvider) Move(x float64, y float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Move, uintptr(unsafe.Pointer(this)), uintptr(x), uintptr(y))
	return HRESULT(ret)
}

func (this *ITransformProvider) Resize(width float64, height float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Resize, uintptr(unsafe.Pointer(this)), uintptr(width), uintptr(height))
	return HRESULT(ret)
}

func (this *ITransformProvider) Rotate(degrees float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Rotate, uintptr(unsafe.Pointer(this)), uintptr(degrees))
	return HRESULT(ret)
}

func (this *ITransformProvider) Get_CanMove(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CanMove, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITransformProvider) Get_CanResize(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CanResize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITransformProvider) Get_CanRotate(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CanRotate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// C7935180-6FB3-4201-B174-7DF73ADBF64A
var IID_IValueProvider = syscall.GUID{0xC7935180, 0x6FB3, 0x4201,
	[8]byte{0xB1, 0x74, 0x7D, 0xF7, 0x3A, 0xDB, 0xF6, 0x4A}}

type IValueProviderInterface interface {
	IUnknownInterface
	SetValue(val PWSTR) HRESULT
	Get_Value(pRetVal *BSTR) HRESULT
	Get_IsReadOnly(pRetVal *BOOL) HRESULT
}

type IValueProviderVtbl struct {
	IUnknownVtbl
	SetValue       uintptr
	Get_Value      uintptr
	Get_IsReadOnly uintptr
}

type IValueProvider struct {
	IUnknown
}

func (this *IValueProvider) Vtbl() *IValueProviderVtbl {
	return (*IValueProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IValueProvider) SetValue(val PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(val)))
	return HRESULT(ret)
}

func (this *IValueProvider) Get_Value(pRetVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Value, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IValueProvider) Get_IsReadOnly(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_IsReadOnly, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 987DF77B-DB06-4D77-8F8A-86A9C3BB90B9
var IID_IWindowProvider = syscall.GUID{0x987DF77B, 0xDB06, 0x4D77,
	[8]byte{0x8F, 0x8A, 0x86, 0xA9, 0xC3, 0xBB, 0x90, 0xB9}}

type IWindowProviderInterface interface {
	IUnknownInterface
	SetVisualState(state WindowVisualState) HRESULT
	Close() HRESULT
	WaitForInputIdle(milliseconds int32, pRetVal *BOOL) HRESULT
	Get_CanMaximize(pRetVal *BOOL) HRESULT
	Get_CanMinimize(pRetVal *BOOL) HRESULT
	Get_IsModal(pRetVal *BOOL) HRESULT
	Get_WindowVisualState(pRetVal *WindowVisualState) HRESULT
	Get_WindowInteractionState(pRetVal *WindowInteractionState) HRESULT
	Get_IsTopmost(pRetVal *BOOL) HRESULT
}

type IWindowProviderVtbl struct {
	IUnknownVtbl
	SetVisualState             uintptr
	Close                      uintptr
	WaitForInputIdle           uintptr
	Get_CanMaximize            uintptr
	Get_CanMinimize            uintptr
	Get_IsModal                uintptr
	Get_WindowVisualState      uintptr
	Get_WindowInteractionState uintptr
	Get_IsTopmost              uintptr
}

type IWindowProvider struct {
	IUnknown
}

func (this *IWindowProvider) Vtbl() *IWindowProviderVtbl {
	return (*IWindowProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IWindowProvider) SetVisualState(state WindowVisualState) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetVisualState, uintptr(unsafe.Pointer(this)), uintptr(state))
	return HRESULT(ret)
}

func (this *IWindowProvider) Close() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Close, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IWindowProvider) WaitForInputIdle(milliseconds int32, pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().WaitForInputIdle, uintptr(unsafe.Pointer(this)), uintptr(milliseconds), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IWindowProvider) Get_CanMaximize(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CanMaximize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IWindowProvider) Get_CanMinimize(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CanMinimize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IWindowProvider) Get_IsModal(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_IsModal, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IWindowProvider) Get_WindowVisualState(pRetVal *WindowVisualState) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_WindowVisualState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IWindowProvider) Get_WindowInteractionState(pRetVal *WindowInteractionState) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_WindowInteractionState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IWindowProvider) Get_IsTopmost(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_IsTopmost, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// E44C3566-915D-4070-99C6-047BFF5A08F5
var IID_ILegacyIAccessibleProvider = syscall.GUID{0xE44C3566, 0x915D, 0x4070,
	[8]byte{0x99, 0xC6, 0x04, 0x7B, 0xFF, 0x5A, 0x08, 0xF5}}

type ILegacyIAccessibleProviderInterface interface {
	IUnknownInterface
	Select(flagsSelect int32) HRESULT
	DoDefaultAction() HRESULT
	SetValue(szValue PWSTR) HRESULT
	GetIAccessible(ppAccessible **IAccessible) HRESULT
	Get_ChildId(pRetVal *int32) HRESULT
	Get_Name(pszName *BSTR) HRESULT
	Get_Value(pszValue *BSTR) HRESULT
	Get_Description(pszDescription *BSTR) HRESULT
	Get_Role(pdwRole *uint32) HRESULT
	Get_State(pdwState *uint32) HRESULT
	Get_Help(pszHelp *BSTR) HRESULT
	Get_KeyboardShortcut(pszKeyboardShortcut *BSTR) HRESULT
	GetSelection(pvarSelectedChildren **SAFEARRAY) HRESULT
	Get_DefaultAction(pszDefaultAction *BSTR) HRESULT
}

type ILegacyIAccessibleProviderVtbl struct {
	IUnknownVtbl
	Select               uintptr
	DoDefaultAction      uintptr
	SetValue             uintptr
	GetIAccessible       uintptr
	Get_ChildId          uintptr
	Get_Name             uintptr
	Get_Value            uintptr
	Get_Description      uintptr
	Get_Role             uintptr
	Get_State            uintptr
	Get_Help             uintptr
	Get_KeyboardShortcut uintptr
	GetSelection         uintptr
	Get_DefaultAction    uintptr
}

type ILegacyIAccessibleProvider struct {
	IUnknown
}

func (this *ILegacyIAccessibleProvider) Vtbl() *ILegacyIAccessibleProviderVtbl {
	return (*ILegacyIAccessibleProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ILegacyIAccessibleProvider) Select(flagsSelect int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Select, uintptr(unsafe.Pointer(this)), uintptr(flagsSelect))
	return HRESULT(ret)
}

func (this *ILegacyIAccessibleProvider) DoDefaultAction() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DoDefaultAction, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ILegacyIAccessibleProvider) SetValue(szValue PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szValue)))
	return HRESULT(ret)
}

func (this *ILegacyIAccessibleProvider) GetIAccessible(ppAccessible **IAccessible) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetIAccessible, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppAccessible)))
	return HRESULT(ret)
}

func (this *ILegacyIAccessibleProvider) Get_ChildId(pRetVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ChildId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ILegacyIAccessibleProvider) Get_Name(pszName *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Name, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszName)))
	return HRESULT(ret)
}

func (this *ILegacyIAccessibleProvider) Get_Value(pszValue *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Value, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszValue)))
	return HRESULT(ret)
}

func (this *ILegacyIAccessibleProvider) Get_Description(pszDescription *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Description, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszDescription)))
	return HRESULT(ret)
}

func (this *ILegacyIAccessibleProvider) Get_Role(pdwRole *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Role, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwRole)))
	return HRESULT(ret)
}

func (this *ILegacyIAccessibleProvider) Get_State(pdwState *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_State, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwState)))
	return HRESULT(ret)
}

func (this *ILegacyIAccessibleProvider) Get_Help(pszHelp *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Help, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszHelp)))
	return HRESULT(ret)
}

func (this *ILegacyIAccessibleProvider) Get_KeyboardShortcut(pszKeyboardShortcut *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_KeyboardShortcut, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszKeyboardShortcut)))
	return HRESULT(ret)
}

func (this *ILegacyIAccessibleProvider) GetSelection(pvarSelectedChildren **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSelection, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pvarSelectedChildren)))
	return HRESULT(ret)
}

func (this *ILegacyIAccessibleProvider) Get_DefaultAction(pszDefaultAction *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_DefaultAction, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszDefaultAction)))
	return HRESULT(ret)
}

// E747770B-39CE-4382-AB30-D8FB3F336F24
var IID_IItemContainerProvider = syscall.GUID{0xE747770B, 0x39CE, 0x4382,
	[8]byte{0xAB, 0x30, 0xD8, 0xFB, 0x3F, 0x33, 0x6F, 0x24}}

type IItemContainerProviderInterface interface {
	IUnknownInterface
	FindItemByProperty(pStartAfter *IRawElementProviderSimple, propertyId int32, value VARIANT, pFound **IRawElementProviderSimple) HRESULT
}

type IItemContainerProviderVtbl struct {
	IUnknownVtbl
	FindItemByProperty uintptr
}

type IItemContainerProvider struct {
	IUnknown
}

func (this *IItemContainerProvider) Vtbl() *IItemContainerProviderVtbl {
	return (*IItemContainerProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IItemContainerProvider) FindItemByProperty(pStartAfter *IRawElementProviderSimple, propertyId int32, value VARIANT, pFound **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindItemByProperty, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pStartAfter)), uintptr(propertyId), uintptr(unsafe.Pointer(&value)), uintptr(unsafe.Pointer(pFound)))
	return HRESULT(ret)
}

// CB98B665-2D35-4FAC-AD35-F3C60D0C0B8B
var IID_IVirtualizedItemProvider = syscall.GUID{0xCB98B665, 0x2D35, 0x4FAC,
	[8]byte{0xAD, 0x35, 0xF3, 0xC6, 0x0D, 0x0C, 0x0B, 0x8B}}

type IVirtualizedItemProviderInterface interface {
	IUnknownInterface
	Realize() HRESULT
}

type IVirtualizedItemProviderVtbl struct {
	IUnknownVtbl
	Realize uintptr
}

type IVirtualizedItemProvider struct {
	IUnknown
}

func (this *IVirtualizedItemProvider) Vtbl() *IVirtualizedItemProviderVtbl {
	return (*IVirtualizedItemProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IVirtualizedItemProvider) Realize() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Realize, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 3AD86EBD-F5EF-483D-BB18-B1042A475D64
var IID_IObjectModelProvider = syscall.GUID{0x3AD86EBD, 0xF5EF, 0x483D,
	[8]byte{0xBB, 0x18, 0xB1, 0x04, 0x2A, 0x47, 0x5D, 0x64}}

type IObjectModelProviderInterface interface {
	IUnknownInterface
	GetUnderlyingObjectModel(ppUnknown **IUnknown) HRESULT
}

type IObjectModelProviderVtbl struct {
	IUnknownVtbl
	GetUnderlyingObjectModel uintptr
}

type IObjectModelProvider struct {
	IUnknown
}

func (this *IObjectModelProvider) Vtbl() *IObjectModelProviderVtbl {
	return (*IObjectModelProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IObjectModelProvider) GetUnderlyingObjectModel(ppUnknown **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetUnderlyingObjectModel, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppUnknown)))
	return HRESULT(ret)
}

// F95C7E80-BD63-4601-9782-445EBFF011FC
var IID_IAnnotationProvider = syscall.GUID{0xF95C7E80, 0xBD63, 0x4601,
	[8]byte{0x97, 0x82, 0x44, 0x5E, 0xBF, 0xF0, 0x11, 0xFC}}

type IAnnotationProviderInterface interface {
	IUnknownInterface
	Get_AnnotationTypeId(retVal *int32) HRESULT
	Get_AnnotationTypeName(retVal *BSTR) HRESULT
	Get_Author(retVal *BSTR) HRESULT
	Get_DateTime(retVal *BSTR) HRESULT
	Get_Target(retVal **IRawElementProviderSimple) HRESULT
}

type IAnnotationProviderVtbl struct {
	IUnknownVtbl
	Get_AnnotationTypeId   uintptr
	Get_AnnotationTypeName uintptr
	Get_Author             uintptr
	Get_DateTime           uintptr
	Get_Target             uintptr
}

type IAnnotationProvider struct {
	IUnknown
}

func (this *IAnnotationProvider) Vtbl() *IAnnotationProviderVtbl {
	return (*IAnnotationProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAnnotationProvider) Get_AnnotationTypeId(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_AnnotationTypeId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IAnnotationProvider) Get_AnnotationTypeName(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_AnnotationTypeName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IAnnotationProvider) Get_Author(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Author, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IAnnotationProvider) Get_DateTime(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_DateTime, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IAnnotationProvider) Get_Target(retVal **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Target, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 19B6B649-F5D7-4A6D-BDCB-129252BE588A
var IID_IStylesProvider = syscall.GUID{0x19B6B649, 0xF5D7, 0x4A6D,
	[8]byte{0xBD, 0xCB, 0x12, 0x92, 0x52, 0xBE, 0x58, 0x8A}}

type IStylesProviderInterface interface {
	IUnknownInterface
	Get_StyleId(retVal *int32) HRESULT
	Get_StyleName(retVal *BSTR) HRESULT
	Get_FillColor(retVal *int32) HRESULT
	Get_FillPatternStyle(retVal *BSTR) HRESULT
	Get_Shape(retVal *BSTR) HRESULT
	Get_FillPatternColor(retVal *int32) HRESULT
	Get_ExtendedProperties(retVal *BSTR) HRESULT
}

type IStylesProviderVtbl struct {
	IUnknownVtbl
	Get_StyleId            uintptr
	Get_StyleName          uintptr
	Get_FillColor          uintptr
	Get_FillPatternStyle   uintptr
	Get_Shape              uintptr
	Get_FillPatternColor   uintptr
	Get_ExtendedProperties uintptr
}

type IStylesProvider struct {
	IUnknown
}

func (this *IStylesProvider) Vtbl() *IStylesProviderVtbl {
	return (*IStylesProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IStylesProvider) Get_StyleId(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_StyleId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IStylesProvider) Get_StyleName(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_StyleName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IStylesProvider) Get_FillColor(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_FillColor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IStylesProvider) Get_FillPatternStyle(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_FillPatternStyle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IStylesProvider) Get_Shape(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Shape, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IStylesProvider) Get_FillPatternColor(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_FillPatternColor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IStylesProvider) Get_ExtendedProperties(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ExtendedProperties, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 6F6B5D35-5525-4F80-B758-85473832FFC7
var IID_ISpreadsheetProvider = syscall.GUID{0x6F6B5D35, 0x5525, 0x4F80,
	[8]byte{0xB7, 0x58, 0x85, 0x47, 0x38, 0x32, 0xFF, 0xC7}}

type ISpreadsheetProviderInterface interface {
	IUnknownInterface
	GetItemByName(name PWSTR, pRetVal **IRawElementProviderSimple) HRESULT
}

type ISpreadsheetProviderVtbl struct {
	IUnknownVtbl
	GetItemByName uintptr
}

type ISpreadsheetProvider struct {
	IUnknown
}

func (this *ISpreadsheetProvider) Vtbl() *ISpreadsheetProviderVtbl {
	return (*ISpreadsheetProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISpreadsheetProvider) GetItemByName(name PWSTR, pRetVal **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetItemByName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// EAED4660-7B3D-4879-A2E6-365CE603F3D0
var IID_ISpreadsheetItemProvider = syscall.GUID{0xEAED4660, 0x7B3D, 0x4879,
	[8]byte{0xA2, 0xE6, 0x36, 0x5C, 0xE6, 0x03, 0xF3, 0xD0}}

type ISpreadsheetItemProviderInterface interface {
	IUnknownInterface
	Get_Formula(pRetVal *BSTR) HRESULT
	GetAnnotationObjects(pRetVal **SAFEARRAY) HRESULT
	GetAnnotationTypes(pRetVal **SAFEARRAY) HRESULT
}

type ISpreadsheetItemProviderVtbl struct {
	IUnknownVtbl
	Get_Formula          uintptr
	GetAnnotationObjects uintptr
	GetAnnotationTypes   uintptr
}

type ISpreadsheetItemProvider struct {
	IUnknown
}

func (this *ISpreadsheetItemProvider) Vtbl() *ISpreadsheetItemProviderVtbl {
	return (*ISpreadsheetItemProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISpreadsheetItemProvider) Get_Formula(pRetVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Formula, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ISpreadsheetItemProvider) GetAnnotationObjects(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAnnotationObjects, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ISpreadsheetItemProvider) GetAnnotationTypes(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAnnotationTypes, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 4758742F-7AC2-460C-BC48-09FC09308A93
var IID_ITransformProvider2 = syscall.GUID{0x4758742F, 0x7AC2, 0x460C,
	[8]byte{0xBC, 0x48, 0x09, 0xFC, 0x09, 0x30, 0x8A, 0x93}}

type ITransformProvider2Interface interface {
	ITransformProviderInterface
	Zoom(zoom float64) HRESULT
	Get_CanZoom(pRetVal *BOOL) HRESULT
	Get_ZoomLevel(pRetVal *float64) HRESULT
	Get_ZoomMinimum(pRetVal *float64) HRESULT
	Get_ZoomMaximum(pRetVal *float64) HRESULT
	ZoomByUnit(zoomUnit ZoomUnit) HRESULT
}

type ITransformProvider2Vtbl struct {
	ITransformProviderVtbl
	Zoom            uintptr
	Get_CanZoom     uintptr
	Get_ZoomLevel   uintptr
	Get_ZoomMinimum uintptr
	Get_ZoomMaximum uintptr
	ZoomByUnit      uintptr
}

type ITransformProvider2 struct {
	ITransformProvider
}

func (this *ITransformProvider2) Vtbl() *ITransformProvider2Vtbl {
	return (*ITransformProvider2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITransformProvider2) Zoom(zoom float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Zoom, uintptr(unsafe.Pointer(this)), uintptr(zoom))
	return HRESULT(ret)
}

func (this *ITransformProvider2) Get_CanZoom(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CanZoom, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITransformProvider2) Get_ZoomLevel(pRetVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ZoomLevel, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITransformProvider2) Get_ZoomMinimum(pRetVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ZoomMinimum, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITransformProvider2) Get_ZoomMaximum(pRetVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ZoomMaximum, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITransformProvider2) ZoomByUnit(zoomUnit ZoomUnit) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ZoomByUnit, uintptr(unsafe.Pointer(this)), uintptr(zoomUnit))
	return HRESULT(ret)
}

// 6AA7BBBB-7FF9-497D-904F-D20B897929D8
var IID_IDragProvider = syscall.GUID{0x6AA7BBBB, 0x7FF9, 0x497D,
	[8]byte{0x90, 0x4F, 0xD2, 0x0B, 0x89, 0x79, 0x29, 0xD8}}

type IDragProviderInterface interface {
	IUnknownInterface
	Get_IsGrabbed(pRetVal *BOOL) HRESULT
	Get_DropEffect(pRetVal *BSTR) HRESULT
	Get_DropEffects(pRetVal **SAFEARRAY) HRESULT
	GetGrabbedItems(pRetVal **SAFEARRAY) HRESULT
}

type IDragProviderVtbl struct {
	IUnknownVtbl
	Get_IsGrabbed   uintptr
	Get_DropEffect  uintptr
	Get_DropEffects uintptr
	GetGrabbedItems uintptr
}

type IDragProvider struct {
	IUnknown
}

func (this *IDragProvider) Vtbl() *IDragProviderVtbl {
	return (*IDragProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDragProvider) Get_IsGrabbed(pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_IsGrabbed, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IDragProvider) Get_DropEffect(pRetVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_DropEffect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IDragProvider) Get_DropEffects(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_DropEffects, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IDragProvider) GetGrabbedItems(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetGrabbedItems, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// BAE82BFD-358A-481C-85A0-D8B4D90A5D61
var IID_IDropTargetProvider = syscall.GUID{0xBAE82BFD, 0x358A, 0x481C,
	[8]byte{0x85, 0xA0, 0xD8, 0xB4, 0xD9, 0x0A, 0x5D, 0x61}}

type IDropTargetProviderInterface interface {
	IUnknownInterface
	Get_DropTargetEffect(pRetVal *BSTR) HRESULT
	Get_DropTargetEffects(pRetVal **SAFEARRAY) HRESULT
}

type IDropTargetProviderVtbl struct {
	IUnknownVtbl
	Get_DropTargetEffect  uintptr
	Get_DropTargetEffects uintptr
}

type IDropTargetProvider struct {
	IUnknown
}

func (this *IDropTargetProvider) Vtbl() *IDropTargetProviderVtbl {
	return (*IDropTargetProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDropTargetProvider) Get_DropTargetEffect(pRetVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_DropTargetEffect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IDropTargetProvider) Get_DropTargetEffects(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_DropTargetEffects, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 5347AD7B-C355-46F8-AFF5-909033582F63
var IID_ITextRangeProvider = syscall.GUID{0x5347AD7B, 0xC355, 0x46F8,
	[8]byte{0xAF, 0xF5, 0x90, 0x90, 0x33, 0x58, 0x2F, 0x63}}

type ITextRangeProviderInterface interface {
	IUnknownInterface
	Clone(pRetVal **ITextRangeProvider) HRESULT
	Compare(range_ *ITextRangeProvider, pRetVal *BOOL) HRESULT
	CompareEndpoints(endpoint TextPatternRangeEndpoint, targetRange *ITextRangeProvider, targetEndpoint TextPatternRangeEndpoint, pRetVal *int32) HRESULT
	ExpandToEnclosingUnit(unit TextUnit) HRESULT
	FindAttribute(attributeId int32, val VARIANT, backward BOOL, pRetVal **ITextRangeProvider) HRESULT
	FindText(text BSTR, backward BOOL, ignoreCase BOOL, pRetVal **ITextRangeProvider) HRESULT
	GetAttributeValue(attributeId int32, pRetVal *VARIANT) HRESULT
	GetBoundingRectangles(pRetVal **SAFEARRAY) HRESULT
	GetEnclosingElement(pRetVal **IRawElementProviderSimple) HRESULT
	GetText(maxLength int32, pRetVal *BSTR) HRESULT
	Move(unit TextUnit, count int32, pRetVal *int32) HRESULT
	MoveEndpointByUnit(endpoint TextPatternRangeEndpoint, unit TextUnit, count int32, pRetVal *int32) HRESULT
	MoveEndpointByRange(endpoint TextPatternRangeEndpoint, targetRange *ITextRangeProvider, targetEndpoint TextPatternRangeEndpoint) HRESULT
	Select() HRESULT
	AddToSelection() HRESULT
	RemoveFromSelection() HRESULT
	ScrollIntoView(alignToTop BOOL) HRESULT
	GetChildren(pRetVal **SAFEARRAY) HRESULT
}

type ITextRangeProviderVtbl struct {
	IUnknownVtbl
	Clone                 uintptr
	Compare               uintptr
	CompareEndpoints      uintptr
	ExpandToEnclosingUnit uintptr
	FindAttribute         uintptr
	FindText              uintptr
	GetAttributeValue     uintptr
	GetBoundingRectangles uintptr
	GetEnclosingElement   uintptr
	GetText               uintptr
	Move                  uintptr
	MoveEndpointByUnit    uintptr
	MoveEndpointByRange   uintptr
	Select                uintptr
	AddToSelection        uintptr
	RemoveFromSelection   uintptr
	ScrollIntoView        uintptr
	GetChildren           uintptr
}

type ITextRangeProvider struct {
	IUnknown
}

func (this *ITextRangeProvider) Vtbl() *ITextRangeProviderVtbl {
	return (*ITextRangeProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextRangeProvider) Clone(pRetVal **ITextRangeProvider) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) Compare(range_ *ITextRangeProvider, pRetVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Compare, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(range_)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) CompareEndpoints(endpoint TextPatternRangeEndpoint, targetRange *ITextRangeProvider, targetEndpoint TextPatternRangeEndpoint, pRetVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CompareEndpoints, uintptr(unsafe.Pointer(this)), uintptr(endpoint), uintptr(unsafe.Pointer(targetRange)), uintptr(targetEndpoint), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) ExpandToEnclosingUnit(unit TextUnit) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ExpandToEnclosingUnit, uintptr(unsafe.Pointer(this)), uintptr(unit))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) FindAttribute(attributeId int32, val VARIANT, backward BOOL, pRetVal **ITextRangeProvider) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindAttribute, uintptr(unsafe.Pointer(this)), uintptr(attributeId), uintptr(unsafe.Pointer(&val)), uintptr(backward), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) FindText(text BSTR, backward BOOL, ignoreCase BOOL, pRetVal **ITextRangeProvider) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(text)), uintptr(backward), uintptr(ignoreCase), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) GetAttributeValue(attributeId int32, pRetVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAttributeValue, uintptr(unsafe.Pointer(this)), uintptr(attributeId), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) GetBoundingRectangles(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBoundingRectangles, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) GetEnclosingElement(pRetVal **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEnclosingElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) GetText(maxLength int32, pRetVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetText, uintptr(unsafe.Pointer(this)), uintptr(maxLength), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) Move(unit TextUnit, count int32, pRetVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Move, uintptr(unsafe.Pointer(this)), uintptr(unit), uintptr(count), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) MoveEndpointByUnit(endpoint TextPatternRangeEndpoint, unit TextUnit, count int32, pRetVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveEndpointByUnit, uintptr(unsafe.Pointer(this)), uintptr(endpoint), uintptr(unit), uintptr(count), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) MoveEndpointByRange(endpoint TextPatternRangeEndpoint, targetRange *ITextRangeProvider, targetEndpoint TextPatternRangeEndpoint) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveEndpointByRange, uintptr(unsafe.Pointer(this)), uintptr(endpoint), uintptr(unsafe.Pointer(targetRange)), uintptr(targetEndpoint))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) Select() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Select, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) AddToSelection() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddToSelection, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) RemoveFromSelection() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveFromSelection, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) ScrollIntoView(alignToTop BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ScrollIntoView, uintptr(unsafe.Pointer(this)), uintptr(alignToTop))
	return HRESULT(ret)
}

func (this *ITextRangeProvider) GetChildren(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetChildren, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 3589C92C-63F3-4367-99BB-ADA653B77CF2
var IID_ITextProvider = syscall.GUID{0x3589C92C, 0x63F3, 0x4367,
	[8]byte{0x99, 0xBB, 0xAD, 0xA6, 0x53, 0xB7, 0x7C, 0xF2}}

type ITextProviderInterface interface {
	IUnknownInterface
	GetSelection(pRetVal **SAFEARRAY) HRESULT
	GetVisibleRanges(pRetVal **SAFEARRAY) HRESULT
	RangeFromChild(childElement *IRawElementProviderSimple, pRetVal **ITextRangeProvider) HRESULT
	RangeFromPoint(point UiaPoint, pRetVal **ITextRangeProvider) HRESULT
	Get_DocumentRange(pRetVal **ITextRangeProvider) HRESULT
	Get_SupportedTextSelection(pRetVal *SupportedTextSelection) HRESULT
}

type ITextProviderVtbl struct {
	IUnknownVtbl
	GetSelection               uintptr
	GetVisibleRanges           uintptr
	RangeFromChild             uintptr
	RangeFromPoint             uintptr
	Get_DocumentRange          uintptr
	Get_SupportedTextSelection uintptr
}

type ITextProvider struct {
	IUnknown
}

func (this *ITextProvider) Vtbl() *ITextProviderVtbl {
	return (*ITextProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextProvider) GetSelection(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSelection, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextProvider) GetVisibleRanges(pRetVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetVisibleRanges, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextProvider) RangeFromChild(childElement *IRawElementProviderSimple, pRetVal **ITextRangeProvider) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RangeFromChild, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(childElement)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextProvider) RangeFromPoint(point UiaPoint, pRetVal **ITextRangeProvider) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RangeFromPoint, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&point)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextProvider) Get_DocumentRange(pRetVal **ITextRangeProvider) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_DocumentRange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextProvider) Get_SupportedTextSelection(pRetVal *SupportedTextSelection) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_SupportedTextSelection, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 0DC5E6ED-3E16-4BF1-8F9A-A979878BC195
var IID_ITextProvider2 = syscall.GUID{0x0DC5E6ED, 0x3E16, 0x4BF1,
	[8]byte{0x8F, 0x9A, 0xA9, 0x79, 0x87, 0x8B, 0xC1, 0x95}}

type ITextProvider2Interface interface {
	ITextProviderInterface
	RangeFromAnnotation(annotationElement *IRawElementProviderSimple, pRetVal **ITextRangeProvider) HRESULT
	GetCaretRange(isActive *BOOL, pRetVal **ITextRangeProvider) HRESULT
}

type ITextProvider2Vtbl struct {
	ITextProviderVtbl
	RangeFromAnnotation uintptr
	GetCaretRange       uintptr
}

type ITextProvider2 struct {
	ITextProvider
}

func (this *ITextProvider2) Vtbl() *ITextProvider2Vtbl {
	return (*ITextProvider2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextProvider2) RangeFromAnnotation(annotationElement *IRawElementProviderSimple, pRetVal **ITextRangeProvider) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RangeFromAnnotation, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(annotationElement)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextProvider2) GetCaretRange(isActive *BOOL, pRetVal **ITextRangeProvider) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCaretRange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(isActive)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// EA3605B4-3A05-400E-B5F9-4E91B40F6176
var IID_ITextEditProvider = syscall.GUID{0xEA3605B4, 0x3A05, 0x400E,
	[8]byte{0xB5, 0xF9, 0x4E, 0x91, 0xB4, 0x0F, 0x61, 0x76}}

type ITextEditProviderInterface interface {
	ITextProviderInterface
	GetActiveComposition(pRetVal **ITextRangeProvider) HRESULT
	GetConversionTarget(pRetVal **ITextRangeProvider) HRESULT
}

type ITextEditProviderVtbl struct {
	ITextProviderVtbl
	GetActiveComposition uintptr
	GetConversionTarget  uintptr
}

type ITextEditProvider struct {
	ITextProvider
}

func (this *ITextEditProvider) Vtbl() *ITextEditProviderVtbl {
	return (*ITextEditProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextEditProvider) GetActiveComposition(pRetVal **ITextRangeProvider) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetActiveComposition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextEditProvider) GetConversionTarget(pRetVal **ITextRangeProvider) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetConversionTarget, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 9BBCE42C-1921-4F18-89CA-DBA1910A0386
var IID_ITextRangeProvider2 = syscall.GUID{0x9BBCE42C, 0x1921, 0x4F18,
	[8]byte{0x89, 0xCA, 0xDB, 0xA1, 0x91, 0x0A, 0x03, 0x86}}

type ITextRangeProvider2Interface interface {
	ITextRangeProviderInterface
	ShowContextMenu() HRESULT
}

type ITextRangeProvider2Vtbl struct {
	ITextRangeProviderVtbl
	ShowContextMenu uintptr
}

type ITextRangeProvider2 struct {
	ITextRangeProvider
}

func (this *ITextRangeProvider2) Vtbl() *ITextRangeProvider2Vtbl {
	return (*ITextRangeProvider2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextRangeProvider2) ShowContextMenu() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ShowContextMenu, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 4C2DE2B9-C88F-4F88-A111-F1D336B7D1A9
var IID_ITextChildProvider = syscall.GUID{0x4C2DE2B9, 0xC88F, 0x4F88,
	[8]byte{0xA1, 0x11, 0xF1, 0xD3, 0x36, 0xB7, 0xD1, 0xA9}}

type ITextChildProviderInterface interface {
	IUnknownInterface
	Get_TextContainer(pRetVal **IRawElementProviderSimple) HRESULT
	Get_TextRange(pRetVal **ITextRangeProvider) HRESULT
}

type ITextChildProviderVtbl struct {
	IUnknownVtbl
	Get_TextContainer uintptr
	Get_TextRange     uintptr
}

type ITextChildProvider struct {
	IUnknown
}

func (this *ITextChildProvider) Vtbl() *ITextChildProviderVtbl {
	return (*ITextChildProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextChildProvider) Get_TextContainer(pRetVal **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_TextContainer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *ITextChildProvider) Get_TextRange(pRetVal **ITextRangeProvider) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_TextRange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// 2062A28A-8C07-4B94-8E12-7037C622AEB8
var IID_ICustomNavigationProvider = syscall.GUID{0x2062A28A, 0x8C07, 0x4B94,
	[8]byte{0x8E, 0x12, 0x70, 0x37, 0xC6, 0x22, 0xAE, 0xB8}}

type ICustomNavigationProviderInterface interface {
	IUnknownInterface
	Navigate(direction NavigateDirection, pRetVal **IRawElementProviderSimple) HRESULT
}

type ICustomNavigationProviderVtbl struct {
	IUnknownVtbl
	Navigate uintptr
}

type ICustomNavigationProvider struct {
	IUnknown
}

func (this *ICustomNavigationProvider) Vtbl() *ICustomNavigationProviderVtbl {
	return (*ICustomNavigationProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICustomNavigationProvider) Navigate(direction NavigateDirection, pRetVal **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Navigate, uintptr(unsafe.Pointer(this)), uintptr(direction), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// C03A7FE4-9431-409F-BED8-AE7C2299BC8D
var IID_IUIAutomationPatternInstance = syscall.GUID{0xC03A7FE4, 0x9431, 0x409F,
	[8]byte{0xBE, 0xD8, 0xAE, 0x7C, 0x22, 0x99, 0xBC, 0x8D}}

type IUIAutomationPatternInstanceInterface interface {
	IUnknownInterface
	GetProperty(index uint32, cached BOOL, type_ UIAutomationType, pPtr unsafe.Pointer) HRESULT
	CallMethod(index uint32, pParams *UIAutomationParameter, cParams uint32) HRESULT
}

type IUIAutomationPatternInstanceVtbl struct {
	IUnknownVtbl
	GetProperty uintptr
	CallMethod  uintptr
}

type IUIAutomationPatternInstance struct {
	IUnknown
}

func (this *IUIAutomationPatternInstance) Vtbl() *IUIAutomationPatternInstanceVtbl {
	return (*IUIAutomationPatternInstanceVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationPatternInstance) GetProperty(index uint32, cached BOOL, type_ UIAutomationType, pPtr unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetProperty, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(cached), uintptr(type_), uintptr(pPtr))
	return HRESULT(ret)
}

func (this *IUIAutomationPatternInstance) CallMethod(index uint32, pParams *UIAutomationParameter, cParams uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CallMethod, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(pParams)), uintptr(cParams))
	return HRESULT(ret)
}

// D97022F3-A947-465E-8B2A-AC4315FA54E8
var IID_IUIAutomationPatternHandler = syscall.GUID{0xD97022F3, 0xA947, 0x465E,
	[8]byte{0x8B, 0x2A, 0xAC, 0x43, 0x15, 0xFA, 0x54, 0xE8}}

type IUIAutomationPatternHandlerInterface interface {
	IUnknownInterface
	CreateClientWrapper(pPatternInstance *IUIAutomationPatternInstance, pClientWrapper **IUnknown) HRESULT
	Dispatch(pTarget *IUnknown, index uint32, pParams *UIAutomationParameter, cParams uint32) HRESULT
}

type IUIAutomationPatternHandlerVtbl struct {
	IUnknownVtbl
	CreateClientWrapper uintptr
	Dispatch            uintptr
}

type IUIAutomationPatternHandler struct {
	IUnknown
}

func (this *IUIAutomationPatternHandler) Vtbl() *IUIAutomationPatternHandlerVtbl {
	return (*IUIAutomationPatternHandlerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationPatternHandler) CreateClientWrapper(pPatternInstance *IUIAutomationPatternInstance, pClientWrapper **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateClientWrapper, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPatternInstance)), uintptr(unsafe.Pointer(pClientWrapper)))
	return HRESULT(ret)
}

func (this *IUIAutomationPatternHandler) Dispatch(pTarget *IUnknown, index uint32, pParams *UIAutomationParameter, cParams uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Dispatch, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pTarget)), uintptr(index), uintptr(unsafe.Pointer(pParams)), uintptr(cParams))
	return HRESULT(ret)
}

// 8609C4EC-4A1A-4D88-A357-5A66E060E1CF
var IID_IUIAutomationRegistrar = syscall.GUID{0x8609C4EC, 0x4A1A, 0x4D88,
	[8]byte{0xA3, 0x57, 0x5A, 0x66, 0xE0, 0x60, 0xE1, 0xCF}}

type IUIAutomationRegistrarInterface interface {
	IUnknownInterface
	RegisterProperty(property *UIAutomationPropertyInfo, propertyId *int32) HRESULT
	RegisterEvent(event *UIAutomationEventInfo, eventId *int32) HRESULT
	RegisterPattern(pattern *UIAutomationPatternInfo, pPatternId *int32, pPatternAvailablePropertyId *int32, propertyIdCount uint32, pPropertyIds *int32, eventIdCount uint32, pEventIds *int32) HRESULT
}

type IUIAutomationRegistrarVtbl struct {
	IUnknownVtbl
	RegisterProperty uintptr
	RegisterEvent    uintptr
	RegisterPattern  uintptr
}

type IUIAutomationRegistrar struct {
	IUnknown
}

func (this *IUIAutomationRegistrar) Vtbl() *IUIAutomationRegistrarVtbl {
	return (*IUIAutomationRegistrarVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationRegistrar) RegisterProperty(property *UIAutomationPropertyInfo, propertyId *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RegisterProperty, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(property)), uintptr(unsafe.Pointer(propertyId)))
	return HRESULT(ret)
}

func (this *IUIAutomationRegistrar) RegisterEvent(event *UIAutomationEventInfo, eventId *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RegisterEvent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(event)), uintptr(unsafe.Pointer(eventId)))
	return HRESULT(ret)
}

func (this *IUIAutomationRegistrar) RegisterPattern(pattern *UIAutomationPatternInfo, pPatternId *int32, pPatternAvailablePropertyId *int32, propertyIdCount uint32, pPropertyIds *int32, eventIdCount uint32, pEventIds *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RegisterPattern, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pattern)), uintptr(unsafe.Pointer(pPatternId)), uintptr(unsafe.Pointer(pPatternAvailablePropertyId)), uintptr(propertyIdCount), uintptr(unsafe.Pointer(pPropertyIds)), uintptr(eventIdCount), uintptr(unsafe.Pointer(pEventIds)))
	return HRESULT(ret)
}

// D22108AA-8AC5-49A5-837B-37BBB3D7591E
var IID_IUIAutomationElement = syscall.GUID{0xD22108AA, 0x8AC5, 0x49A5,
	[8]byte{0x83, 0x7B, 0x37, 0xBB, 0xB3, 0xD7, 0x59, 0x1E}}

type IUIAutomationElementInterface interface {
	IUnknownInterface
	SetFocus() HRESULT
	GetRuntimeId(runtimeId **SAFEARRAY) HRESULT
	FindFirst(scope TreeScope, condition *IUIAutomationCondition, found **IUIAutomationElement) HRESULT
	FindAll(scope TreeScope, condition *IUIAutomationCondition, found **IUIAutomationElementArray) HRESULT
	FindFirstBuildCache(scope TreeScope, condition *IUIAutomationCondition, cacheRequest *IUIAutomationCacheRequest, found **IUIAutomationElement) HRESULT
	FindAllBuildCache(scope TreeScope, condition *IUIAutomationCondition, cacheRequest *IUIAutomationCacheRequest, found **IUIAutomationElementArray) HRESULT
	BuildUpdatedCache(cacheRequest *IUIAutomationCacheRequest, updatedElement **IUIAutomationElement) HRESULT
	GetCurrentPropertyValue(propertyId int32, retVal *VARIANT) HRESULT
	GetCurrentPropertyValueEx(propertyId int32, ignoreDefaultValue BOOL, retVal *VARIANT) HRESULT
	GetCachedPropertyValue(propertyId int32, retVal *VARIANT) HRESULT
	GetCachedPropertyValueEx(propertyId int32, ignoreDefaultValue BOOL, retVal *VARIANT) HRESULT
	GetCurrentPatternAs(patternId int32, riid *syscall.GUID, patternObject unsafe.Pointer) HRESULT
	GetCachedPatternAs(patternId int32, riid *syscall.GUID, patternObject unsafe.Pointer) HRESULT
	GetCurrentPattern(patternId int32, patternObject **IUnknown) HRESULT
	GetCachedPattern(patternId int32, patternObject **IUnknown) HRESULT
	GetCachedParent(parent **IUIAutomationElement) HRESULT
	GetCachedChildren(children **IUIAutomationElementArray) HRESULT
	Get_CurrentProcessId(retVal *int32) HRESULT
	Get_CurrentControlType(retVal *int32) HRESULT
	Get_CurrentLocalizedControlType(retVal *BSTR) HRESULT
	Get_CurrentName(retVal *BSTR) HRESULT
	Get_CurrentAcceleratorKey(retVal *BSTR) HRESULT
	Get_CurrentAccessKey(retVal *BSTR) HRESULT
	Get_CurrentHasKeyboardFocus(retVal *BOOL) HRESULT
	Get_CurrentIsKeyboardFocusable(retVal *BOOL) HRESULT
	Get_CurrentIsEnabled(retVal *BOOL) HRESULT
	Get_CurrentAutomationId(retVal *BSTR) HRESULT
	Get_CurrentClassName(retVal *BSTR) HRESULT
	Get_CurrentHelpText(retVal *BSTR) HRESULT
	Get_CurrentCulture(retVal *int32) HRESULT
	Get_CurrentIsControlElement(retVal *BOOL) HRESULT
	Get_CurrentIsContentElement(retVal *BOOL) HRESULT
	Get_CurrentIsPassword(retVal *BOOL) HRESULT
	Get_CurrentNativeWindowHandle(retVal *HWND) HRESULT
	Get_CurrentItemType(retVal *BSTR) HRESULT
	Get_CurrentIsOffscreen(retVal *BOOL) HRESULT
	Get_CurrentOrientation(retVal *OrientationType) HRESULT
	Get_CurrentFrameworkId(retVal *BSTR) HRESULT
	Get_CurrentIsRequiredForForm(retVal *BOOL) HRESULT
	Get_CurrentItemStatus(retVal *BSTR) HRESULT
	Get_CurrentBoundingRectangle(retVal *RECT) HRESULT
	Get_CurrentLabeledBy(retVal **IUIAutomationElement) HRESULT
	Get_CurrentAriaRole(retVal *BSTR) HRESULT
	Get_CurrentAriaProperties(retVal *BSTR) HRESULT
	Get_CurrentIsDataValidForForm(retVal *BOOL) HRESULT
	Get_CurrentControllerFor(retVal **IUIAutomationElementArray) HRESULT
	Get_CurrentDescribedBy(retVal **IUIAutomationElementArray) HRESULT
	Get_CurrentFlowsTo(retVal **IUIAutomationElementArray) HRESULT
	Get_CurrentProviderDescription(retVal *BSTR) HRESULT
	Get_CachedProcessId(retVal *int32) HRESULT
	Get_CachedControlType(retVal *int32) HRESULT
	Get_CachedLocalizedControlType(retVal *BSTR) HRESULT
	Get_CachedName(retVal *BSTR) HRESULT
	Get_CachedAcceleratorKey(retVal *BSTR) HRESULT
	Get_CachedAccessKey(retVal *BSTR) HRESULT
	Get_CachedHasKeyboardFocus(retVal *BOOL) HRESULT
	Get_CachedIsKeyboardFocusable(retVal *BOOL) HRESULT
	Get_CachedIsEnabled(retVal *BOOL) HRESULT
	Get_CachedAutomationId(retVal *BSTR) HRESULT
	Get_CachedClassName(retVal *BSTR) HRESULT
	Get_CachedHelpText(retVal *BSTR) HRESULT
	Get_CachedCulture(retVal *int32) HRESULT
	Get_CachedIsControlElement(retVal *BOOL) HRESULT
	Get_CachedIsContentElement(retVal *BOOL) HRESULT
	Get_CachedIsPassword(retVal *BOOL) HRESULT
	Get_CachedNativeWindowHandle(retVal *HWND) HRESULT
	Get_CachedItemType(retVal *BSTR) HRESULT
	Get_CachedIsOffscreen(retVal *BOOL) HRESULT
	Get_CachedOrientation(retVal *OrientationType) HRESULT
	Get_CachedFrameworkId(retVal *BSTR) HRESULT
	Get_CachedIsRequiredForForm(retVal *BOOL) HRESULT
	Get_CachedItemStatus(retVal *BSTR) HRESULT
	Get_CachedBoundingRectangle(retVal *RECT) HRESULT
	Get_CachedLabeledBy(retVal **IUIAutomationElement) HRESULT
	Get_CachedAriaRole(retVal *BSTR) HRESULT
	Get_CachedAriaProperties(retVal *BSTR) HRESULT
	Get_CachedIsDataValidForForm(retVal *BOOL) HRESULT
	Get_CachedControllerFor(retVal **IUIAutomationElementArray) HRESULT
	Get_CachedDescribedBy(retVal **IUIAutomationElementArray) HRESULT
	Get_CachedFlowsTo(retVal **IUIAutomationElementArray) HRESULT
	Get_CachedProviderDescription(retVal *BSTR) HRESULT
	GetClickablePoint(clickable *POINT, gotClickable *BOOL) HRESULT
}

type IUIAutomationElementVtbl struct {
	IUnknownVtbl
	SetFocus                        uintptr
	GetRuntimeId                    uintptr
	FindFirst                       uintptr
	FindAll                         uintptr
	FindFirstBuildCache             uintptr
	FindAllBuildCache               uintptr
	BuildUpdatedCache               uintptr
	GetCurrentPropertyValue         uintptr
	GetCurrentPropertyValueEx       uintptr
	GetCachedPropertyValue          uintptr
	GetCachedPropertyValueEx        uintptr
	GetCurrentPatternAs             uintptr
	GetCachedPatternAs              uintptr
	GetCurrentPattern               uintptr
	GetCachedPattern                uintptr
	GetCachedParent                 uintptr
	GetCachedChildren               uintptr
	Get_CurrentProcessId            uintptr
	Get_CurrentControlType          uintptr
	Get_CurrentLocalizedControlType uintptr
	Get_CurrentName                 uintptr
	Get_CurrentAcceleratorKey       uintptr
	Get_CurrentAccessKey            uintptr
	Get_CurrentHasKeyboardFocus     uintptr
	Get_CurrentIsKeyboardFocusable  uintptr
	Get_CurrentIsEnabled            uintptr
	Get_CurrentAutomationId         uintptr
	Get_CurrentClassName            uintptr
	Get_CurrentHelpText             uintptr
	Get_CurrentCulture              uintptr
	Get_CurrentIsControlElement     uintptr
	Get_CurrentIsContentElement     uintptr
	Get_CurrentIsPassword           uintptr
	Get_CurrentNativeWindowHandle   uintptr
	Get_CurrentItemType             uintptr
	Get_CurrentIsOffscreen          uintptr
	Get_CurrentOrientation          uintptr
	Get_CurrentFrameworkId          uintptr
	Get_CurrentIsRequiredForForm    uintptr
	Get_CurrentItemStatus           uintptr
	Get_CurrentBoundingRectangle    uintptr
	Get_CurrentLabeledBy            uintptr
	Get_CurrentAriaRole             uintptr
	Get_CurrentAriaProperties       uintptr
	Get_CurrentIsDataValidForForm   uintptr
	Get_CurrentControllerFor        uintptr
	Get_CurrentDescribedBy          uintptr
	Get_CurrentFlowsTo              uintptr
	Get_CurrentProviderDescription  uintptr
	Get_CachedProcessId             uintptr
	Get_CachedControlType           uintptr
	Get_CachedLocalizedControlType  uintptr
	Get_CachedName                  uintptr
	Get_CachedAcceleratorKey        uintptr
	Get_CachedAccessKey             uintptr
	Get_CachedHasKeyboardFocus      uintptr
	Get_CachedIsKeyboardFocusable   uintptr
	Get_CachedIsEnabled             uintptr
	Get_CachedAutomationId          uintptr
	Get_CachedClassName             uintptr
	Get_CachedHelpText              uintptr
	Get_CachedCulture               uintptr
	Get_CachedIsControlElement      uintptr
	Get_CachedIsContentElement      uintptr
	Get_CachedIsPassword            uintptr
	Get_CachedNativeWindowHandle    uintptr
	Get_CachedItemType              uintptr
	Get_CachedIsOffscreen           uintptr
	Get_CachedOrientation           uintptr
	Get_CachedFrameworkId           uintptr
	Get_CachedIsRequiredForForm     uintptr
	Get_CachedItemStatus            uintptr
	Get_CachedBoundingRectangle     uintptr
	Get_CachedLabeledBy             uintptr
	Get_CachedAriaRole              uintptr
	Get_CachedAriaProperties        uintptr
	Get_CachedIsDataValidForForm    uintptr
	Get_CachedControllerFor         uintptr
	Get_CachedDescribedBy           uintptr
	Get_CachedFlowsTo               uintptr
	Get_CachedProviderDescription   uintptr
	GetClickablePoint               uintptr
}

type IUIAutomationElement struct {
	IUnknown
}

func (this *IUIAutomationElement) Vtbl() *IUIAutomationElementVtbl {
	return (*IUIAutomationElementVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationElement) SetFocus() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFocus, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) GetRuntimeId(runtimeId **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRuntimeId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(runtimeId)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) FindFirst(scope TreeScope, condition *IUIAutomationCondition, found **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindFirst, uintptr(unsafe.Pointer(this)), uintptr(scope), uintptr(unsafe.Pointer(condition)), uintptr(unsafe.Pointer(found)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) FindAll(scope TreeScope, condition *IUIAutomationCondition, found **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindAll, uintptr(unsafe.Pointer(this)), uintptr(scope), uintptr(unsafe.Pointer(condition)), uintptr(unsafe.Pointer(found)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) FindFirstBuildCache(scope TreeScope, condition *IUIAutomationCondition, cacheRequest *IUIAutomationCacheRequest, found **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindFirstBuildCache, uintptr(unsafe.Pointer(this)), uintptr(scope), uintptr(unsafe.Pointer(condition)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(found)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) FindAllBuildCache(scope TreeScope, condition *IUIAutomationCondition, cacheRequest *IUIAutomationCacheRequest, found **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindAllBuildCache, uintptr(unsafe.Pointer(this)), uintptr(scope), uintptr(unsafe.Pointer(condition)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(found)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) BuildUpdatedCache(cacheRequest *IUIAutomationCacheRequest, updatedElement **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().BuildUpdatedCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(updatedElement)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) GetCurrentPropertyValue(propertyId int32, retVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentPropertyValue, uintptr(unsafe.Pointer(this)), uintptr(propertyId), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) GetCurrentPropertyValueEx(propertyId int32, ignoreDefaultValue BOOL, retVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentPropertyValueEx, uintptr(unsafe.Pointer(this)), uintptr(propertyId), uintptr(ignoreDefaultValue), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) GetCachedPropertyValue(propertyId int32, retVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedPropertyValue, uintptr(unsafe.Pointer(this)), uintptr(propertyId), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) GetCachedPropertyValueEx(propertyId int32, ignoreDefaultValue BOOL, retVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedPropertyValueEx, uintptr(unsafe.Pointer(this)), uintptr(propertyId), uintptr(ignoreDefaultValue), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) GetCurrentPatternAs(patternId int32, riid *syscall.GUID, patternObject unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentPatternAs, uintptr(unsafe.Pointer(this)), uintptr(patternId), uintptr(unsafe.Pointer(riid)), uintptr(patternObject))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) GetCachedPatternAs(patternId int32, riid *syscall.GUID, patternObject unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedPatternAs, uintptr(unsafe.Pointer(this)), uintptr(patternId), uintptr(unsafe.Pointer(riid)), uintptr(patternObject))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) GetCurrentPattern(patternId int32, patternObject **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentPattern, uintptr(unsafe.Pointer(this)), uintptr(patternId), uintptr(unsafe.Pointer(patternObject)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) GetCachedPattern(patternId int32, patternObject **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedPattern, uintptr(unsafe.Pointer(this)), uintptr(patternId), uintptr(unsafe.Pointer(patternObject)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) GetCachedParent(parent **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedParent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(parent)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) GetCachedChildren(children **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedChildren, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(children)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentProcessId(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentProcessId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentControlType(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentControlType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentLocalizedControlType(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentLocalizedControlType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentName(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentAcceleratorKey(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentAcceleratorKey, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentAccessKey(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentAccessKey, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentHasKeyboardFocus(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentHasKeyboardFocus, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentIsKeyboardFocusable(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsKeyboardFocusable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentIsEnabled(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsEnabled, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentAutomationId(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentAutomationId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentClassName(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentClassName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentHelpText(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentHelpText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentCulture(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentCulture, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentIsControlElement(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsControlElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentIsContentElement(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsContentElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentIsPassword(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsPassword, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentNativeWindowHandle(retVal *HWND) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentNativeWindowHandle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentItemType(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentItemType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentIsOffscreen(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsOffscreen, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentOrientation(retVal *OrientationType) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentOrientation, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentFrameworkId(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentFrameworkId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentIsRequiredForForm(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsRequiredForForm, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentItemStatus(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentItemStatus, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentBoundingRectangle(retVal *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentBoundingRectangle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentLabeledBy(retVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentLabeledBy, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentAriaRole(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentAriaRole, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentAriaProperties(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentAriaProperties, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentIsDataValidForForm(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsDataValidForForm, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentControllerFor(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentControllerFor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentDescribedBy(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentDescribedBy, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentFlowsTo(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentFlowsTo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CurrentProviderDescription(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentProviderDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedProcessId(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedProcessId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedControlType(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedControlType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedLocalizedControlType(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedLocalizedControlType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedName(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedAcceleratorKey(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedAcceleratorKey, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedAccessKey(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedAccessKey, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedHasKeyboardFocus(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedHasKeyboardFocus, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedIsKeyboardFocusable(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsKeyboardFocusable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedIsEnabled(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsEnabled, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedAutomationId(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedAutomationId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedClassName(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedClassName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedHelpText(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedHelpText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedCulture(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedCulture, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedIsControlElement(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsControlElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedIsContentElement(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsContentElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedIsPassword(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsPassword, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedNativeWindowHandle(retVal *HWND) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedNativeWindowHandle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedItemType(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedItemType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedIsOffscreen(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsOffscreen, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedOrientation(retVal *OrientationType) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedOrientation, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedFrameworkId(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedFrameworkId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedIsRequiredForForm(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsRequiredForForm, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedItemStatus(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedItemStatus, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedBoundingRectangle(retVal *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedBoundingRectangle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedLabeledBy(retVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedLabeledBy, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedAriaRole(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedAriaRole, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedAriaProperties(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedAriaProperties, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedIsDataValidForForm(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsDataValidForForm, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedControllerFor(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedControllerFor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedDescribedBy(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedDescribedBy, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedFlowsTo(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedFlowsTo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) Get_CachedProviderDescription(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedProviderDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement) GetClickablePoint(clickable *POINT, gotClickable *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetClickablePoint, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(clickable)), uintptr(unsafe.Pointer(gotClickable)))
	return HRESULT(ret)
}

// 14314595-B4BC-4055-95F2-58F2E42C9855
var IID_IUIAutomationElementArray = syscall.GUID{0x14314595, 0xB4BC, 0x4055,
	[8]byte{0x95, 0xF2, 0x58, 0xF2, 0xE4, 0x2C, 0x98, 0x55}}

type IUIAutomationElementArrayInterface interface {
	IUnknownInterface
	Get_Length(length *int32) HRESULT
	GetElement(index int32, element **IUIAutomationElement) HRESULT
}

type IUIAutomationElementArrayVtbl struct {
	IUnknownVtbl
	Get_Length uintptr
	GetElement uintptr
}

type IUIAutomationElementArray struct {
	IUnknown
}

func (this *IUIAutomationElementArray) Vtbl() *IUIAutomationElementArrayVtbl {
	return (*IUIAutomationElementArrayVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationElementArray) Get_Length(length *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Length, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(length)))
	return HRESULT(ret)
}

func (this *IUIAutomationElementArray) GetElement(index int32, element **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetElement, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(element)))
	return HRESULT(ret)
}

// 352FFBA8-0973-437C-A61F-F64CAFD81DF9
var IID_IUIAutomationCondition = syscall.GUID{0x352FFBA8, 0x0973, 0x437C,
	[8]byte{0xA6, 0x1F, 0xF6, 0x4C, 0xAF, 0xD8, 0x1D, 0xF9}}

type IUIAutomationConditionInterface interface {
	IUnknownInterface
}

type IUIAutomationConditionVtbl struct {
	IUnknownVtbl
}

type IUIAutomationCondition struct {
	IUnknown
}

func (this *IUIAutomationCondition) Vtbl() *IUIAutomationConditionVtbl {
	return (*IUIAutomationConditionVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

// 1B4E1F2E-75EB-4D0B-8952-5A69988E2307
var IID_IUIAutomationBoolCondition = syscall.GUID{0x1B4E1F2E, 0x75EB, 0x4D0B,
	[8]byte{0x89, 0x52, 0x5A, 0x69, 0x98, 0x8E, 0x23, 0x07}}

type IUIAutomationBoolConditionInterface interface {
	IUIAutomationConditionInterface
	Get_BooleanValue(boolVal *BOOL) HRESULT
}

type IUIAutomationBoolConditionVtbl struct {
	IUIAutomationConditionVtbl
	Get_BooleanValue uintptr
}

type IUIAutomationBoolCondition struct {
	IUIAutomationCondition
}

func (this *IUIAutomationBoolCondition) Vtbl() *IUIAutomationBoolConditionVtbl {
	return (*IUIAutomationBoolConditionVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationBoolCondition) Get_BooleanValue(boolVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_BooleanValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(boolVal)))
	return HRESULT(ret)
}

// 99EBF2CB-5578-4267-9AD4-AFD6EA77E94B
var IID_IUIAutomationPropertyCondition = syscall.GUID{0x99EBF2CB, 0x5578, 0x4267,
	[8]byte{0x9A, 0xD4, 0xAF, 0xD6, 0xEA, 0x77, 0xE9, 0x4B}}

type IUIAutomationPropertyConditionInterface interface {
	IUIAutomationConditionInterface
	Get_PropertyId(propertyId *int32) HRESULT
	Get_PropertyValue(propertyValue *VARIANT) HRESULT
	Get_PropertyConditionFlags(flags *PropertyConditionFlags) HRESULT
}

type IUIAutomationPropertyConditionVtbl struct {
	IUIAutomationConditionVtbl
	Get_PropertyId             uintptr
	Get_PropertyValue          uintptr
	Get_PropertyConditionFlags uintptr
}

type IUIAutomationPropertyCondition struct {
	IUIAutomationCondition
}

func (this *IUIAutomationPropertyCondition) Vtbl() *IUIAutomationPropertyConditionVtbl {
	return (*IUIAutomationPropertyConditionVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationPropertyCondition) Get_PropertyId(propertyId *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_PropertyId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(propertyId)))
	return HRESULT(ret)
}

func (this *IUIAutomationPropertyCondition) Get_PropertyValue(propertyValue *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_PropertyValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(propertyValue)))
	return HRESULT(ret)
}

func (this *IUIAutomationPropertyCondition) Get_PropertyConditionFlags(flags *PropertyConditionFlags) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_PropertyConditionFlags, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(flags)))
	return HRESULT(ret)
}

// A7D0AF36-B912-45FE-9855-091DDC174AEC
var IID_IUIAutomationAndCondition = syscall.GUID{0xA7D0AF36, 0xB912, 0x45FE,
	[8]byte{0x98, 0x55, 0x09, 0x1D, 0xDC, 0x17, 0x4A, 0xEC}}

type IUIAutomationAndConditionInterface interface {
	IUIAutomationConditionInterface
	Get_ChildCount(childCount *int32) HRESULT
	GetChildrenAsNativeArray(childArray ***IUIAutomationCondition, childArrayCount *int32) HRESULT
	GetChildren(childArray **SAFEARRAY) HRESULT
}

type IUIAutomationAndConditionVtbl struct {
	IUIAutomationConditionVtbl
	Get_ChildCount           uintptr
	GetChildrenAsNativeArray uintptr
	GetChildren              uintptr
}

type IUIAutomationAndCondition struct {
	IUIAutomationCondition
}

func (this *IUIAutomationAndCondition) Vtbl() *IUIAutomationAndConditionVtbl {
	return (*IUIAutomationAndConditionVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationAndCondition) Get_ChildCount(childCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ChildCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(childCount)))
	return HRESULT(ret)
}

func (this *IUIAutomationAndCondition) GetChildrenAsNativeArray(childArray ***IUIAutomationCondition, childArrayCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetChildrenAsNativeArray, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(childArray)), uintptr(unsafe.Pointer(childArrayCount)))
	return HRESULT(ret)
}

func (this *IUIAutomationAndCondition) GetChildren(childArray **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetChildren, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(childArray)))
	return HRESULT(ret)
}

// 8753F032-3DB1-47B5-A1FC-6E34A266C712
var IID_IUIAutomationOrCondition = syscall.GUID{0x8753F032, 0x3DB1, 0x47B5,
	[8]byte{0xA1, 0xFC, 0x6E, 0x34, 0xA2, 0x66, 0xC7, 0x12}}

type IUIAutomationOrConditionInterface interface {
	IUIAutomationConditionInterface
	Get_ChildCount(childCount *int32) HRESULT
	GetChildrenAsNativeArray(childArray ***IUIAutomationCondition, childArrayCount *int32) HRESULT
	GetChildren(childArray **SAFEARRAY) HRESULT
}

type IUIAutomationOrConditionVtbl struct {
	IUIAutomationConditionVtbl
	Get_ChildCount           uintptr
	GetChildrenAsNativeArray uintptr
	GetChildren              uintptr
}

type IUIAutomationOrCondition struct {
	IUIAutomationCondition
}

func (this *IUIAutomationOrCondition) Vtbl() *IUIAutomationOrConditionVtbl {
	return (*IUIAutomationOrConditionVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationOrCondition) Get_ChildCount(childCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ChildCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(childCount)))
	return HRESULT(ret)
}

func (this *IUIAutomationOrCondition) GetChildrenAsNativeArray(childArray ***IUIAutomationCondition, childArrayCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetChildrenAsNativeArray, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(childArray)), uintptr(unsafe.Pointer(childArrayCount)))
	return HRESULT(ret)
}

func (this *IUIAutomationOrCondition) GetChildren(childArray **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetChildren, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(childArray)))
	return HRESULT(ret)
}

// F528B657-847B-498C-8896-D52B565407A1
var IID_IUIAutomationNotCondition = syscall.GUID{0xF528B657, 0x847B, 0x498C,
	[8]byte{0x88, 0x96, 0xD5, 0x2B, 0x56, 0x54, 0x07, 0xA1}}

type IUIAutomationNotConditionInterface interface {
	IUIAutomationConditionInterface
	GetChild(condition **IUIAutomationCondition) HRESULT
}

type IUIAutomationNotConditionVtbl struct {
	IUIAutomationConditionVtbl
	GetChild uintptr
}

type IUIAutomationNotCondition struct {
	IUIAutomationCondition
}

func (this *IUIAutomationNotCondition) Vtbl() *IUIAutomationNotConditionVtbl {
	return (*IUIAutomationNotConditionVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationNotCondition) GetChild(condition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetChild, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(condition)))
	return HRESULT(ret)
}

// B32A92B5-BC25-4078-9C08-D7EE95C48E03
var IID_IUIAutomationCacheRequest = syscall.GUID{0xB32A92B5, 0xBC25, 0x4078,
	[8]byte{0x9C, 0x08, 0xD7, 0xEE, 0x95, 0xC4, 0x8E, 0x03}}

type IUIAutomationCacheRequestInterface interface {
	IUnknownInterface
	AddProperty(propertyId int32) HRESULT
	AddPattern(patternId int32) HRESULT
	Clone(clonedRequest **IUIAutomationCacheRequest) HRESULT
	Get_TreeScope(scope *TreeScope) HRESULT
	Put_TreeScope(scope TreeScope) HRESULT
	Get_TreeFilter(filter **IUIAutomationCondition) HRESULT
	Put_TreeFilter(filter *IUIAutomationCondition) HRESULT
	Get_AutomationElementMode(mode *AutomationElementMode) HRESULT
	Put_AutomationElementMode(mode AutomationElementMode) HRESULT
}

type IUIAutomationCacheRequestVtbl struct {
	IUnknownVtbl
	AddProperty               uintptr
	AddPattern                uintptr
	Clone                     uintptr
	Get_TreeScope             uintptr
	Put_TreeScope             uintptr
	Get_TreeFilter            uintptr
	Put_TreeFilter            uintptr
	Get_AutomationElementMode uintptr
	Put_AutomationElementMode uintptr
}

type IUIAutomationCacheRequest struct {
	IUnknown
}

func (this *IUIAutomationCacheRequest) Vtbl() *IUIAutomationCacheRequestVtbl {
	return (*IUIAutomationCacheRequestVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationCacheRequest) AddProperty(propertyId int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddProperty, uintptr(unsafe.Pointer(this)), uintptr(propertyId))
	return HRESULT(ret)
}

func (this *IUIAutomationCacheRequest) AddPattern(patternId int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddPattern, uintptr(unsafe.Pointer(this)), uintptr(patternId))
	return HRESULT(ret)
}

func (this *IUIAutomationCacheRequest) Clone(clonedRequest **IUIAutomationCacheRequest) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(clonedRequest)))
	return HRESULT(ret)
}

func (this *IUIAutomationCacheRequest) Get_TreeScope(scope *TreeScope) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_TreeScope, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(scope)))
	return HRESULT(ret)
}

func (this *IUIAutomationCacheRequest) Put_TreeScope(scope TreeScope) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_TreeScope, uintptr(unsafe.Pointer(this)), uintptr(scope))
	return HRESULT(ret)
}

func (this *IUIAutomationCacheRequest) Get_TreeFilter(filter **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_TreeFilter, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(filter)))
	return HRESULT(ret)
}

func (this *IUIAutomationCacheRequest) Put_TreeFilter(filter *IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_TreeFilter, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(filter)))
	return HRESULT(ret)
}

func (this *IUIAutomationCacheRequest) Get_AutomationElementMode(mode *AutomationElementMode) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_AutomationElementMode, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(mode)))
	return HRESULT(ret)
}

func (this *IUIAutomationCacheRequest) Put_AutomationElementMode(mode AutomationElementMode) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_AutomationElementMode, uintptr(unsafe.Pointer(this)), uintptr(mode))
	return HRESULT(ret)
}

// 4042C624-389C-4AFC-A630-9DF854A541FC
var IID_IUIAutomationTreeWalker = syscall.GUID{0x4042C624, 0x389C, 0x4AFC,
	[8]byte{0xA6, 0x30, 0x9D, 0xF8, 0x54, 0xA5, 0x41, 0xFC}}

type IUIAutomationTreeWalkerInterface interface {
	IUnknownInterface
	GetParentElement(element *IUIAutomationElement, parent **IUIAutomationElement) HRESULT
	GetFirstChildElement(element *IUIAutomationElement, first **IUIAutomationElement) HRESULT
	GetLastChildElement(element *IUIAutomationElement, last **IUIAutomationElement) HRESULT
	GetNextSiblingElement(element *IUIAutomationElement, next **IUIAutomationElement) HRESULT
	GetPreviousSiblingElement(element *IUIAutomationElement, previous **IUIAutomationElement) HRESULT
	NormalizeElement(element *IUIAutomationElement, normalized **IUIAutomationElement) HRESULT
	GetParentElementBuildCache(element *IUIAutomationElement, cacheRequest *IUIAutomationCacheRequest, parent **IUIAutomationElement) HRESULT
	GetFirstChildElementBuildCache(element *IUIAutomationElement, cacheRequest *IUIAutomationCacheRequest, first **IUIAutomationElement) HRESULT
	GetLastChildElementBuildCache(element *IUIAutomationElement, cacheRequest *IUIAutomationCacheRequest, last **IUIAutomationElement) HRESULT
	GetNextSiblingElementBuildCache(element *IUIAutomationElement, cacheRequest *IUIAutomationCacheRequest, next **IUIAutomationElement) HRESULT
	GetPreviousSiblingElementBuildCache(element *IUIAutomationElement, cacheRequest *IUIAutomationCacheRequest, previous **IUIAutomationElement) HRESULT
	NormalizeElementBuildCache(element *IUIAutomationElement, cacheRequest *IUIAutomationCacheRequest, normalized **IUIAutomationElement) HRESULT
	Get_Condition(condition **IUIAutomationCondition) HRESULT
}

type IUIAutomationTreeWalkerVtbl struct {
	IUnknownVtbl
	GetParentElement                    uintptr
	GetFirstChildElement                uintptr
	GetLastChildElement                 uintptr
	GetNextSiblingElement               uintptr
	GetPreviousSiblingElement           uintptr
	NormalizeElement                    uintptr
	GetParentElementBuildCache          uintptr
	GetFirstChildElementBuildCache      uintptr
	GetLastChildElementBuildCache       uintptr
	GetNextSiblingElementBuildCache     uintptr
	GetPreviousSiblingElementBuildCache uintptr
	NormalizeElementBuildCache          uintptr
	Get_Condition                       uintptr
}

type IUIAutomationTreeWalker struct {
	IUnknown
}

func (this *IUIAutomationTreeWalker) Vtbl() *IUIAutomationTreeWalkerVtbl {
	return (*IUIAutomationTreeWalkerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTreeWalker) GetParentElement(element *IUIAutomationElement, parent **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetParentElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(parent)))
	return HRESULT(ret)
}

func (this *IUIAutomationTreeWalker) GetFirstChildElement(element *IUIAutomationElement, first **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFirstChildElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(first)))
	return HRESULT(ret)
}

func (this *IUIAutomationTreeWalker) GetLastChildElement(element *IUIAutomationElement, last **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLastChildElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(last)))
	return HRESULT(ret)
}

func (this *IUIAutomationTreeWalker) GetNextSiblingElement(element *IUIAutomationElement, next **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNextSiblingElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(next)))
	return HRESULT(ret)
}

func (this *IUIAutomationTreeWalker) GetPreviousSiblingElement(element *IUIAutomationElement, previous **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPreviousSiblingElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(previous)))
	return HRESULT(ret)
}

func (this *IUIAutomationTreeWalker) NormalizeElement(element *IUIAutomationElement, normalized **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().NormalizeElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(normalized)))
	return HRESULT(ret)
}

func (this *IUIAutomationTreeWalker) GetParentElementBuildCache(element *IUIAutomationElement, cacheRequest *IUIAutomationCacheRequest, parent **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetParentElementBuildCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(parent)))
	return HRESULT(ret)
}

func (this *IUIAutomationTreeWalker) GetFirstChildElementBuildCache(element *IUIAutomationElement, cacheRequest *IUIAutomationCacheRequest, first **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFirstChildElementBuildCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(first)))
	return HRESULT(ret)
}

func (this *IUIAutomationTreeWalker) GetLastChildElementBuildCache(element *IUIAutomationElement, cacheRequest *IUIAutomationCacheRequest, last **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLastChildElementBuildCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(last)))
	return HRESULT(ret)
}

func (this *IUIAutomationTreeWalker) GetNextSiblingElementBuildCache(element *IUIAutomationElement, cacheRequest *IUIAutomationCacheRequest, next **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNextSiblingElementBuildCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(next)))
	return HRESULT(ret)
}

func (this *IUIAutomationTreeWalker) GetPreviousSiblingElementBuildCache(element *IUIAutomationElement, cacheRequest *IUIAutomationCacheRequest, previous **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPreviousSiblingElementBuildCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(previous)))
	return HRESULT(ret)
}

func (this *IUIAutomationTreeWalker) NormalizeElementBuildCache(element *IUIAutomationElement, cacheRequest *IUIAutomationCacheRequest, normalized **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().NormalizeElementBuildCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(normalized)))
	return HRESULT(ret)
}

func (this *IUIAutomationTreeWalker) Get_Condition(condition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Condition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(condition)))
	return HRESULT(ret)
}

// 146C3C17-F12E-4E22-8C27-F894B9B79C69
var IID_IUIAutomationEventHandler = syscall.GUID{0x146C3C17, 0xF12E, 0x4E22,
	[8]byte{0x8C, 0x27, 0xF8, 0x94, 0xB9, 0xB7, 0x9C, 0x69}}

type IUIAutomationEventHandlerInterface interface {
	IUnknownInterface
	HandleAutomationEvent(sender *IUIAutomationElement, eventId int32) HRESULT
}

type IUIAutomationEventHandlerVtbl struct {
	IUnknownVtbl
	HandleAutomationEvent uintptr
}

type IUIAutomationEventHandler struct {
	IUnknown
}

func (this *IUIAutomationEventHandler) Vtbl() *IUIAutomationEventHandlerVtbl {
	return (*IUIAutomationEventHandlerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationEventHandler) HandleAutomationEvent(sender *IUIAutomationElement, eventId int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HandleAutomationEvent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(sender)), uintptr(eventId))
	return HRESULT(ret)
}

// 40CD37D4-C756-4B0C-8C6F-BDDFEEB13B50
var IID_IUIAutomationPropertyChangedEventHandler = syscall.GUID{0x40CD37D4, 0xC756, 0x4B0C,
	[8]byte{0x8C, 0x6F, 0xBD, 0xDF, 0xEE, 0xB1, 0x3B, 0x50}}

type IUIAutomationPropertyChangedEventHandlerInterface interface {
	IUnknownInterface
	HandlePropertyChangedEvent(sender *IUIAutomationElement, propertyId int32, newValue VARIANT) HRESULT
}

type IUIAutomationPropertyChangedEventHandlerVtbl struct {
	IUnknownVtbl
	HandlePropertyChangedEvent uintptr
}

type IUIAutomationPropertyChangedEventHandler struct {
	IUnknown
}

func (this *IUIAutomationPropertyChangedEventHandler) Vtbl() *IUIAutomationPropertyChangedEventHandlerVtbl {
	return (*IUIAutomationPropertyChangedEventHandlerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationPropertyChangedEventHandler) HandlePropertyChangedEvent(sender *IUIAutomationElement, propertyId int32, newValue VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HandlePropertyChangedEvent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(sender)), uintptr(propertyId), uintptr(unsafe.Pointer(&newValue)))
	return HRESULT(ret)
}

// E81D1B4E-11C5-42F8-9754-E7036C79F054
var IID_IUIAutomationStructureChangedEventHandler = syscall.GUID{0xE81D1B4E, 0x11C5, 0x42F8,
	[8]byte{0x97, 0x54, 0xE7, 0x03, 0x6C, 0x79, 0xF0, 0x54}}

type IUIAutomationStructureChangedEventHandlerInterface interface {
	IUnknownInterface
	HandleStructureChangedEvent(sender *IUIAutomationElement, changeType StructureChangeType, runtimeId *SAFEARRAY) HRESULT
}

type IUIAutomationStructureChangedEventHandlerVtbl struct {
	IUnknownVtbl
	HandleStructureChangedEvent uintptr
}

type IUIAutomationStructureChangedEventHandler struct {
	IUnknown
}

func (this *IUIAutomationStructureChangedEventHandler) Vtbl() *IUIAutomationStructureChangedEventHandlerVtbl {
	return (*IUIAutomationStructureChangedEventHandlerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationStructureChangedEventHandler) HandleStructureChangedEvent(sender *IUIAutomationElement, changeType StructureChangeType, runtimeId *SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HandleStructureChangedEvent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(sender)), uintptr(changeType), uintptr(unsafe.Pointer(runtimeId)))
	return HRESULT(ret)
}

// C270F6B5-5C69-4290-9745-7A7F97169468
var IID_IUIAutomationFocusChangedEventHandler = syscall.GUID{0xC270F6B5, 0x5C69, 0x4290,
	[8]byte{0x97, 0x45, 0x7A, 0x7F, 0x97, 0x16, 0x94, 0x68}}

type IUIAutomationFocusChangedEventHandlerInterface interface {
	IUnknownInterface
	HandleFocusChangedEvent(sender *IUIAutomationElement) HRESULT
}

type IUIAutomationFocusChangedEventHandlerVtbl struct {
	IUnknownVtbl
	HandleFocusChangedEvent uintptr
}

type IUIAutomationFocusChangedEventHandler struct {
	IUnknown
}

func (this *IUIAutomationFocusChangedEventHandler) Vtbl() *IUIAutomationFocusChangedEventHandlerVtbl {
	return (*IUIAutomationFocusChangedEventHandlerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationFocusChangedEventHandler) HandleFocusChangedEvent(sender *IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HandleFocusChangedEvent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(sender)))
	return HRESULT(ret)
}

// 92FAA680-E704-4156-931A-E32D5BB38F3F
var IID_IUIAutomationTextEditTextChangedEventHandler = syscall.GUID{0x92FAA680, 0xE704, 0x4156,
	[8]byte{0x93, 0x1A, 0xE3, 0x2D, 0x5B, 0xB3, 0x8F, 0x3F}}

type IUIAutomationTextEditTextChangedEventHandlerInterface interface {
	IUnknownInterface
	HandleTextEditTextChangedEvent(sender *IUIAutomationElement, textEditChangeType TextEditChangeType, eventStrings *SAFEARRAY) HRESULT
}

type IUIAutomationTextEditTextChangedEventHandlerVtbl struct {
	IUnknownVtbl
	HandleTextEditTextChangedEvent uintptr
}

type IUIAutomationTextEditTextChangedEventHandler struct {
	IUnknown
}

func (this *IUIAutomationTextEditTextChangedEventHandler) Vtbl() *IUIAutomationTextEditTextChangedEventHandlerVtbl {
	return (*IUIAutomationTextEditTextChangedEventHandlerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTextEditTextChangedEventHandler) HandleTextEditTextChangedEvent(sender *IUIAutomationElement, textEditChangeType TextEditChangeType, eventStrings *SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HandleTextEditTextChangedEvent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(sender)), uintptr(textEditChangeType), uintptr(unsafe.Pointer(eventStrings)))
	return HRESULT(ret)
}

// 58EDCA55-2C3E-4980-B1B9-56C17F27A2A0
var IID_IUIAutomationChangesEventHandler = syscall.GUID{0x58EDCA55, 0x2C3E, 0x4980,
	[8]byte{0xB1, 0xB9, 0x56, 0xC1, 0x7F, 0x27, 0xA2, 0xA0}}

type IUIAutomationChangesEventHandlerInterface interface {
	IUnknownInterface
	HandleChangesEvent(sender *IUIAutomationElement, uiaChanges *UiaChangeInfo, changesCount int32) HRESULT
}

type IUIAutomationChangesEventHandlerVtbl struct {
	IUnknownVtbl
	HandleChangesEvent uintptr
}

type IUIAutomationChangesEventHandler struct {
	IUnknown
}

func (this *IUIAutomationChangesEventHandler) Vtbl() *IUIAutomationChangesEventHandlerVtbl {
	return (*IUIAutomationChangesEventHandlerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationChangesEventHandler) HandleChangesEvent(sender *IUIAutomationElement, uiaChanges *UiaChangeInfo, changesCount int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HandleChangesEvent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(sender)), uintptr(unsafe.Pointer(uiaChanges)), uintptr(changesCount))
	return HRESULT(ret)
}

// C7CB2637-E6C2-4D0C-85DE-4948C02175C7
var IID_IUIAutomationNotificationEventHandler = syscall.GUID{0xC7CB2637, 0xE6C2, 0x4D0C,
	[8]byte{0x85, 0xDE, 0x49, 0x48, 0xC0, 0x21, 0x75, 0xC7}}

type IUIAutomationNotificationEventHandlerInterface interface {
	IUnknownInterface
	HandleNotificationEvent(sender *IUIAutomationElement, notificationKind NotificationKind, notificationProcessing NotificationProcessing, displayString BSTR, activityId BSTR) HRESULT
}

type IUIAutomationNotificationEventHandlerVtbl struct {
	IUnknownVtbl
	HandleNotificationEvent uintptr
}

type IUIAutomationNotificationEventHandler struct {
	IUnknown
}

func (this *IUIAutomationNotificationEventHandler) Vtbl() *IUIAutomationNotificationEventHandlerVtbl {
	return (*IUIAutomationNotificationEventHandlerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationNotificationEventHandler) HandleNotificationEvent(sender *IUIAutomationElement, notificationKind NotificationKind, notificationProcessing NotificationProcessing, displayString BSTR, activityId BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HandleNotificationEvent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(sender)), uintptr(notificationKind), uintptr(notificationProcessing), uintptr(unsafe.Pointer(displayString)), uintptr(unsafe.Pointer(activityId)))
	return HRESULT(ret)
}

// FB377FBE-8EA6-46D5-9C73-6499642D3059
var IID_IUIAutomationInvokePattern = syscall.GUID{0xFB377FBE, 0x8EA6, 0x46D5,
	[8]byte{0x9C, 0x73, 0x64, 0x99, 0x64, 0x2D, 0x30, 0x59}}

type IUIAutomationInvokePatternInterface interface {
	IUnknownInterface
	Invoke() HRESULT
}

type IUIAutomationInvokePatternVtbl struct {
	IUnknownVtbl
	Invoke uintptr
}

type IUIAutomationInvokePattern struct {
	IUnknown
}

func (this *IUIAutomationInvokePattern) Vtbl() *IUIAutomationInvokePatternVtbl {
	return (*IUIAutomationInvokePatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationInvokePattern) Invoke() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Invoke, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// FDE5EF97-1464-48F6-90BF-43D0948E86EC
var IID_IUIAutomationDockPattern = syscall.GUID{0xFDE5EF97, 0x1464, 0x48F6,
	[8]byte{0x90, 0xBF, 0x43, 0xD0, 0x94, 0x8E, 0x86, 0xEC}}

type IUIAutomationDockPatternInterface interface {
	IUnknownInterface
	SetDockPosition(dockPos DockPosition) HRESULT
	Get_CurrentDockPosition(retVal *DockPosition) HRESULT
	Get_CachedDockPosition(retVal *DockPosition) HRESULT
}

type IUIAutomationDockPatternVtbl struct {
	IUnknownVtbl
	SetDockPosition         uintptr
	Get_CurrentDockPosition uintptr
	Get_CachedDockPosition  uintptr
}

type IUIAutomationDockPattern struct {
	IUnknown
}

func (this *IUIAutomationDockPattern) Vtbl() *IUIAutomationDockPatternVtbl {
	return (*IUIAutomationDockPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationDockPattern) SetDockPosition(dockPos DockPosition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDockPosition, uintptr(unsafe.Pointer(this)), uintptr(dockPos))
	return HRESULT(ret)
}

func (this *IUIAutomationDockPattern) Get_CurrentDockPosition(retVal *DockPosition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentDockPosition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationDockPattern) Get_CachedDockPosition(retVal *DockPosition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedDockPosition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 619BE086-1F4E-4EE4-BAFA-210128738730
var IID_IUIAutomationExpandCollapsePattern = syscall.GUID{0x619BE086, 0x1F4E, 0x4EE4,
	[8]byte{0xBA, 0xFA, 0x21, 0x01, 0x28, 0x73, 0x87, 0x30}}

type IUIAutomationExpandCollapsePatternInterface interface {
	IUnknownInterface
	Expand() HRESULT
	Collapse() HRESULT
	Get_CurrentExpandCollapseState(retVal *ExpandCollapseState) HRESULT
	Get_CachedExpandCollapseState(retVal *ExpandCollapseState) HRESULT
}

type IUIAutomationExpandCollapsePatternVtbl struct {
	IUnknownVtbl
	Expand                         uintptr
	Collapse                       uintptr
	Get_CurrentExpandCollapseState uintptr
	Get_CachedExpandCollapseState  uintptr
}

type IUIAutomationExpandCollapsePattern struct {
	IUnknown
}

func (this *IUIAutomationExpandCollapsePattern) Vtbl() *IUIAutomationExpandCollapsePatternVtbl {
	return (*IUIAutomationExpandCollapsePatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationExpandCollapsePattern) Expand() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Expand, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomationExpandCollapsePattern) Collapse() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Collapse, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomationExpandCollapsePattern) Get_CurrentExpandCollapseState(retVal *ExpandCollapseState) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentExpandCollapseState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationExpandCollapsePattern) Get_CachedExpandCollapseState(retVal *ExpandCollapseState) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedExpandCollapseState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 414C3CDC-856B-4F5B-8538-3131C6302550
var IID_IUIAutomationGridPattern = syscall.GUID{0x414C3CDC, 0x856B, 0x4F5B,
	[8]byte{0x85, 0x38, 0x31, 0x31, 0xC6, 0x30, 0x25, 0x50}}

type IUIAutomationGridPatternInterface interface {
	IUnknownInterface
	GetItem(row int32, column int32, element **IUIAutomationElement) HRESULT
	Get_CurrentRowCount(retVal *int32) HRESULT
	Get_CurrentColumnCount(retVal *int32) HRESULT
	Get_CachedRowCount(retVal *int32) HRESULT
	Get_CachedColumnCount(retVal *int32) HRESULT
}

type IUIAutomationGridPatternVtbl struct {
	IUnknownVtbl
	GetItem                uintptr
	Get_CurrentRowCount    uintptr
	Get_CurrentColumnCount uintptr
	Get_CachedRowCount     uintptr
	Get_CachedColumnCount  uintptr
}

type IUIAutomationGridPattern struct {
	IUnknown
}

func (this *IUIAutomationGridPattern) Vtbl() *IUIAutomationGridPatternVtbl {
	return (*IUIAutomationGridPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationGridPattern) GetItem(row int32, column int32, element **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetItem, uintptr(unsafe.Pointer(this)), uintptr(row), uintptr(column), uintptr(unsafe.Pointer(element)))
	return HRESULT(ret)
}

func (this *IUIAutomationGridPattern) Get_CurrentRowCount(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentRowCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationGridPattern) Get_CurrentColumnCount(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentColumnCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationGridPattern) Get_CachedRowCount(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedRowCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationGridPattern) Get_CachedColumnCount(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedColumnCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 78F8EF57-66C3-4E09-BD7C-E79B2004894D
var IID_IUIAutomationGridItemPattern = syscall.GUID{0x78F8EF57, 0x66C3, 0x4E09,
	[8]byte{0xBD, 0x7C, 0xE7, 0x9B, 0x20, 0x04, 0x89, 0x4D}}

type IUIAutomationGridItemPatternInterface interface {
	IUnknownInterface
	Get_CurrentContainingGrid(retVal **IUIAutomationElement) HRESULT
	Get_CurrentRow(retVal *int32) HRESULT
	Get_CurrentColumn(retVal *int32) HRESULT
	Get_CurrentRowSpan(retVal *int32) HRESULT
	Get_CurrentColumnSpan(retVal *int32) HRESULT
	Get_CachedContainingGrid(retVal **IUIAutomationElement) HRESULT
	Get_CachedRow(retVal *int32) HRESULT
	Get_CachedColumn(retVal *int32) HRESULT
	Get_CachedRowSpan(retVal *int32) HRESULT
	Get_CachedColumnSpan(retVal *int32) HRESULT
}

type IUIAutomationGridItemPatternVtbl struct {
	IUnknownVtbl
	Get_CurrentContainingGrid uintptr
	Get_CurrentRow            uintptr
	Get_CurrentColumn         uintptr
	Get_CurrentRowSpan        uintptr
	Get_CurrentColumnSpan     uintptr
	Get_CachedContainingGrid  uintptr
	Get_CachedRow             uintptr
	Get_CachedColumn          uintptr
	Get_CachedRowSpan         uintptr
	Get_CachedColumnSpan      uintptr
}

type IUIAutomationGridItemPattern struct {
	IUnknown
}

func (this *IUIAutomationGridItemPattern) Vtbl() *IUIAutomationGridItemPatternVtbl {
	return (*IUIAutomationGridItemPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationGridItemPattern) Get_CurrentContainingGrid(retVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentContainingGrid, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationGridItemPattern) Get_CurrentRow(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentRow, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationGridItemPattern) Get_CurrentColumn(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentColumn, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationGridItemPattern) Get_CurrentRowSpan(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentRowSpan, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationGridItemPattern) Get_CurrentColumnSpan(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentColumnSpan, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationGridItemPattern) Get_CachedContainingGrid(retVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedContainingGrid, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationGridItemPattern) Get_CachedRow(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedRow, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationGridItemPattern) Get_CachedColumn(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedColumn, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationGridItemPattern) Get_CachedRowSpan(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedRowSpan, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationGridItemPattern) Get_CachedColumnSpan(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedColumnSpan, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 8D253C91-1DC5-4BB5-B18F-ADE16FA495E8
var IID_IUIAutomationMultipleViewPattern = syscall.GUID{0x8D253C91, 0x1DC5, 0x4BB5,
	[8]byte{0xB1, 0x8F, 0xAD, 0xE1, 0x6F, 0xA4, 0x95, 0xE8}}

type IUIAutomationMultipleViewPatternInterface interface {
	IUnknownInterface
	GetViewName(view int32, name *BSTR) HRESULT
	SetCurrentView(view int32) HRESULT
	Get_CurrentCurrentView(retVal *int32) HRESULT
	GetCurrentSupportedViews(retVal **SAFEARRAY) HRESULT
	Get_CachedCurrentView(retVal *int32) HRESULT
	GetCachedSupportedViews(retVal **SAFEARRAY) HRESULT
}

type IUIAutomationMultipleViewPatternVtbl struct {
	IUnknownVtbl
	GetViewName              uintptr
	SetCurrentView           uintptr
	Get_CurrentCurrentView   uintptr
	GetCurrentSupportedViews uintptr
	Get_CachedCurrentView    uintptr
	GetCachedSupportedViews  uintptr
}

type IUIAutomationMultipleViewPattern struct {
	IUnknown
}

func (this *IUIAutomationMultipleViewPattern) Vtbl() *IUIAutomationMultipleViewPatternVtbl {
	return (*IUIAutomationMultipleViewPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationMultipleViewPattern) GetViewName(view int32, name *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetViewName, uintptr(unsafe.Pointer(this)), uintptr(view), uintptr(unsafe.Pointer(name)))
	return HRESULT(ret)
}

func (this *IUIAutomationMultipleViewPattern) SetCurrentView(view int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCurrentView, uintptr(unsafe.Pointer(this)), uintptr(view))
	return HRESULT(ret)
}

func (this *IUIAutomationMultipleViewPattern) Get_CurrentCurrentView(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentCurrentView, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationMultipleViewPattern) GetCurrentSupportedViews(retVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentSupportedViews, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationMultipleViewPattern) Get_CachedCurrentView(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedCurrentView, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationMultipleViewPattern) GetCachedSupportedViews(retVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedSupportedViews, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 71C284B3-C14D-4D14-981E-19751B0D756D
var IID_IUIAutomationObjectModelPattern = syscall.GUID{0x71C284B3, 0xC14D, 0x4D14,
	[8]byte{0x98, 0x1E, 0x19, 0x75, 0x1B, 0x0D, 0x75, 0x6D}}

type IUIAutomationObjectModelPatternInterface interface {
	IUnknownInterface
	GetUnderlyingObjectModel(retVal **IUnknown) HRESULT
}

type IUIAutomationObjectModelPatternVtbl struct {
	IUnknownVtbl
	GetUnderlyingObjectModel uintptr
}

type IUIAutomationObjectModelPattern struct {
	IUnknown
}

func (this *IUIAutomationObjectModelPattern) Vtbl() *IUIAutomationObjectModelPatternVtbl {
	return (*IUIAutomationObjectModelPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationObjectModelPattern) GetUnderlyingObjectModel(retVal **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetUnderlyingObjectModel, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 59213F4F-7346-49E5-B120-80555987A148
var IID_IUIAutomationRangeValuePattern = syscall.GUID{0x59213F4F, 0x7346, 0x49E5,
	[8]byte{0xB1, 0x20, 0x80, 0x55, 0x59, 0x87, 0xA1, 0x48}}

type IUIAutomationRangeValuePatternInterface interface {
	IUnknownInterface
	SetValue(val float64) HRESULT
	Get_CurrentValue(retVal *float64) HRESULT
	Get_CurrentIsReadOnly(retVal *BOOL) HRESULT
	Get_CurrentMaximum(retVal *float64) HRESULT
	Get_CurrentMinimum(retVal *float64) HRESULT
	Get_CurrentLargeChange(retVal *float64) HRESULT
	Get_CurrentSmallChange(retVal *float64) HRESULT
	Get_CachedValue(retVal *float64) HRESULT
	Get_CachedIsReadOnly(retVal *BOOL) HRESULT
	Get_CachedMaximum(retVal *float64) HRESULT
	Get_CachedMinimum(retVal *float64) HRESULT
	Get_CachedLargeChange(retVal *float64) HRESULT
	Get_CachedSmallChange(retVal *float64) HRESULT
}

type IUIAutomationRangeValuePatternVtbl struct {
	IUnknownVtbl
	SetValue               uintptr
	Get_CurrentValue       uintptr
	Get_CurrentIsReadOnly  uintptr
	Get_CurrentMaximum     uintptr
	Get_CurrentMinimum     uintptr
	Get_CurrentLargeChange uintptr
	Get_CurrentSmallChange uintptr
	Get_CachedValue        uintptr
	Get_CachedIsReadOnly   uintptr
	Get_CachedMaximum      uintptr
	Get_CachedMinimum      uintptr
	Get_CachedLargeChange  uintptr
	Get_CachedSmallChange  uintptr
}

type IUIAutomationRangeValuePattern struct {
	IUnknown
}

func (this *IUIAutomationRangeValuePattern) Vtbl() *IUIAutomationRangeValuePatternVtbl {
	return (*IUIAutomationRangeValuePatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationRangeValuePattern) SetValue(val float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetValue, uintptr(unsafe.Pointer(this)), uintptr(val))
	return HRESULT(ret)
}

func (this *IUIAutomationRangeValuePattern) Get_CurrentValue(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationRangeValuePattern) Get_CurrentIsReadOnly(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsReadOnly, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationRangeValuePattern) Get_CurrentMaximum(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentMaximum, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationRangeValuePattern) Get_CurrentMinimum(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentMinimum, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationRangeValuePattern) Get_CurrentLargeChange(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentLargeChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationRangeValuePattern) Get_CurrentSmallChange(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentSmallChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationRangeValuePattern) Get_CachedValue(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationRangeValuePattern) Get_CachedIsReadOnly(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsReadOnly, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationRangeValuePattern) Get_CachedMaximum(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedMaximum, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationRangeValuePattern) Get_CachedMinimum(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedMinimum, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationRangeValuePattern) Get_CachedLargeChange(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedLargeChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationRangeValuePattern) Get_CachedSmallChange(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedSmallChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 88F4D42A-E881-459D-A77C-73BBBB7E02DC
var IID_IUIAutomationScrollPattern = syscall.GUID{0x88F4D42A, 0xE881, 0x459D,
	[8]byte{0xA7, 0x7C, 0x73, 0xBB, 0xBB, 0x7E, 0x02, 0xDC}}

type IUIAutomationScrollPatternInterface interface {
	IUnknownInterface
	Scroll(horizontalAmount ScrollAmount, verticalAmount ScrollAmount) HRESULT
	SetScrollPercent(horizontalPercent float64, verticalPercent float64) HRESULT
	Get_CurrentHorizontalScrollPercent(retVal *float64) HRESULT
	Get_CurrentVerticalScrollPercent(retVal *float64) HRESULT
	Get_CurrentHorizontalViewSize(retVal *float64) HRESULT
	Get_CurrentVerticalViewSize(retVal *float64) HRESULT
	Get_CurrentHorizontallyScrollable(retVal *BOOL) HRESULT
	Get_CurrentVerticallyScrollable(retVal *BOOL) HRESULT
	Get_CachedHorizontalScrollPercent(retVal *float64) HRESULT
	Get_CachedVerticalScrollPercent(retVal *float64) HRESULT
	Get_CachedHorizontalViewSize(retVal *float64) HRESULT
	Get_CachedVerticalViewSize(retVal *float64) HRESULT
	Get_CachedHorizontallyScrollable(retVal *BOOL) HRESULT
	Get_CachedVerticallyScrollable(retVal *BOOL) HRESULT
}

type IUIAutomationScrollPatternVtbl struct {
	IUnknownVtbl
	Scroll                             uintptr
	SetScrollPercent                   uintptr
	Get_CurrentHorizontalScrollPercent uintptr
	Get_CurrentVerticalScrollPercent   uintptr
	Get_CurrentHorizontalViewSize      uintptr
	Get_CurrentVerticalViewSize        uintptr
	Get_CurrentHorizontallyScrollable  uintptr
	Get_CurrentVerticallyScrollable    uintptr
	Get_CachedHorizontalScrollPercent  uintptr
	Get_CachedVerticalScrollPercent    uintptr
	Get_CachedHorizontalViewSize       uintptr
	Get_CachedVerticalViewSize         uintptr
	Get_CachedHorizontallyScrollable   uintptr
	Get_CachedVerticallyScrollable     uintptr
}

type IUIAutomationScrollPattern struct {
	IUnknown
}

func (this *IUIAutomationScrollPattern) Vtbl() *IUIAutomationScrollPatternVtbl {
	return (*IUIAutomationScrollPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationScrollPattern) Scroll(horizontalAmount ScrollAmount, verticalAmount ScrollAmount) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Scroll, uintptr(unsafe.Pointer(this)), uintptr(horizontalAmount), uintptr(verticalAmount))
	return HRESULT(ret)
}

func (this *IUIAutomationScrollPattern) SetScrollPercent(horizontalPercent float64, verticalPercent float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetScrollPercent, uintptr(unsafe.Pointer(this)), uintptr(horizontalPercent), uintptr(verticalPercent))
	return HRESULT(ret)
}

func (this *IUIAutomationScrollPattern) Get_CurrentHorizontalScrollPercent(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentHorizontalScrollPercent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationScrollPattern) Get_CurrentVerticalScrollPercent(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentVerticalScrollPercent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationScrollPattern) Get_CurrentHorizontalViewSize(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentHorizontalViewSize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationScrollPattern) Get_CurrentVerticalViewSize(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentVerticalViewSize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationScrollPattern) Get_CurrentHorizontallyScrollable(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentHorizontallyScrollable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationScrollPattern) Get_CurrentVerticallyScrollable(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentVerticallyScrollable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationScrollPattern) Get_CachedHorizontalScrollPercent(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedHorizontalScrollPercent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationScrollPattern) Get_CachedVerticalScrollPercent(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedVerticalScrollPercent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationScrollPattern) Get_CachedHorizontalViewSize(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedHorizontalViewSize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationScrollPattern) Get_CachedVerticalViewSize(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedVerticalViewSize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationScrollPattern) Get_CachedHorizontallyScrollable(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedHorizontallyScrollable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationScrollPattern) Get_CachedVerticallyScrollable(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedVerticallyScrollable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// B488300F-D015-4F19-9C29-BB595E3645EF
var IID_IUIAutomationScrollItemPattern = syscall.GUID{0xB488300F, 0xD015, 0x4F19,
	[8]byte{0x9C, 0x29, 0xBB, 0x59, 0x5E, 0x36, 0x45, 0xEF}}

type IUIAutomationScrollItemPatternInterface interface {
	IUnknownInterface
	ScrollIntoView() HRESULT
}

type IUIAutomationScrollItemPatternVtbl struct {
	IUnknownVtbl
	ScrollIntoView uintptr
}

type IUIAutomationScrollItemPattern struct {
	IUnknown
}

func (this *IUIAutomationScrollItemPattern) Vtbl() *IUIAutomationScrollItemPatternVtbl {
	return (*IUIAutomationScrollItemPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationScrollItemPattern) ScrollIntoView() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ScrollIntoView, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 5ED5202E-B2AC-47A6-B638-4B0BF140D78E
var IID_IUIAutomationSelectionPattern = syscall.GUID{0x5ED5202E, 0xB2AC, 0x47A6,
	[8]byte{0xB6, 0x38, 0x4B, 0x0B, 0xF1, 0x40, 0xD7, 0x8E}}

type IUIAutomationSelectionPatternInterface interface {
	IUnknownInterface
	GetCurrentSelection(retVal **IUIAutomationElementArray) HRESULT
	Get_CurrentCanSelectMultiple(retVal *BOOL) HRESULT
	Get_CurrentIsSelectionRequired(retVal *BOOL) HRESULT
	GetCachedSelection(retVal **IUIAutomationElementArray) HRESULT
	Get_CachedCanSelectMultiple(retVal *BOOL) HRESULT
	Get_CachedIsSelectionRequired(retVal *BOOL) HRESULT
}

type IUIAutomationSelectionPatternVtbl struct {
	IUnknownVtbl
	GetCurrentSelection            uintptr
	Get_CurrentCanSelectMultiple   uintptr
	Get_CurrentIsSelectionRequired uintptr
	GetCachedSelection             uintptr
	Get_CachedCanSelectMultiple    uintptr
	Get_CachedIsSelectionRequired  uintptr
}

type IUIAutomationSelectionPattern struct {
	IUnknown
}

func (this *IUIAutomationSelectionPattern) Vtbl() *IUIAutomationSelectionPatternVtbl {
	return (*IUIAutomationSelectionPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationSelectionPattern) GetCurrentSelection(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentSelection, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionPattern) Get_CurrentCanSelectMultiple(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentCanSelectMultiple, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionPattern) Get_CurrentIsSelectionRequired(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsSelectionRequired, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionPattern) GetCachedSelection(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedSelection, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionPattern) Get_CachedCanSelectMultiple(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedCanSelectMultiple, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionPattern) Get_CachedIsSelectionRequired(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsSelectionRequired, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 0532BFAE-C011-4E32-A343-6D642D798555
var IID_IUIAutomationSelectionPattern2 = syscall.GUID{0x0532BFAE, 0xC011, 0x4E32,
	[8]byte{0xA3, 0x43, 0x6D, 0x64, 0x2D, 0x79, 0x85, 0x55}}

type IUIAutomationSelectionPattern2Interface interface {
	IUIAutomationSelectionPatternInterface
	Get_CurrentFirstSelectedItem(retVal **IUIAutomationElement) HRESULT
	Get_CurrentLastSelectedItem(retVal **IUIAutomationElement) HRESULT
	Get_CurrentCurrentSelectedItem(retVal **IUIAutomationElement) HRESULT
	Get_CurrentItemCount(retVal *int32) HRESULT
	Get_CachedFirstSelectedItem(retVal **IUIAutomationElement) HRESULT
	Get_CachedLastSelectedItem(retVal **IUIAutomationElement) HRESULT
	Get_CachedCurrentSelectedItem(retVal **IUIAutomationElement) HRESULT
	Get_CachedItemCount(retVal *int32) HRESULT
}

type IUIAutomationSelectionPattern2Vtbl struct {
	IUIAutomationSelectionPatternVtbl
	Get_CurrentFirstSelectedItem   uintptr
	Get_CurrentLastSelectedItem    uintptr
	Get_CurrentCurrentSelectedItem uintptr
	Get_CurrentItemCount           uintptr
	Get_CachedFirstSelectedItem    uintptr
	Get_CachedLastSelectedItem     uintptr
	Get_CachedCurrentSelectedItem  uintptr
	Get_CachedItemCount            uintptr
}

type IUIAutomationSelectionPattern2 struct {
	IUIAutomationSelectionPattern
}

func (this *IUIAutomationSelectionPattern2) Vtbl() *IUIAutomationSelectionPattern2Vtbl {
	return (*IUIAutomationSelectionPattern2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationSelectionPattern2) Get_CurrentFirstSelectedItem(retVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentFirstSelectedItem, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionPattern2) Get_CurrentLastSelectedItem(retVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentLastSelectedItem, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionPattern2) Get_CurrentCurrentSelectedItem(retVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentCurrentSelectedItem, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionPattern2) Get_CurrentItemCount(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentItemCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionPattern2) Get_CachedFirstSelectedItem(retVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedFirstSelectedItem, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionPattern2) Get_CachedLastSelectedItem(retVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedLastSelectedItem, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionPattern2) Get_CachedCurrentSelectedItem(retVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedCurrentSelectedItem, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionPattern2) Get_CachedItemCount(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedItemCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// A8EFA66A-0FDA-421A-9194-38021F3578EA
var IID_IUIAutomationSelectionItemPattern = syscall.GUID{0xA8EFA66A, 0x0FDA, 0x421A,
	[8]byte{0x91, 0x94, 0x38, 0x02, 0x1F, 0x35, 0x78, 0xEA}}

type IUIAutomationSelectionItemPatternInterface interface {
	IUnknownInterface
	Select() HRESULT
	AddToSelection() HRESULT
	RemoveFromSelection() HRESULT
	Get_CurrentIsSelected(retVal *BOOL) HRESULT
	Get_CurrentSelectionContainer(retVal **IUIAutomationElement) HRESULT
	Get_CachedIsSelected(retVal *BOOL) HRESULT
	Get_CachedSelectionContainer(retVal **IUIAutomationElement) HRESULT
}

type IUIAutomationSelectionItemPatternVtbl struct {
	IUnknownVtbl
	Select                        uintptr
	AddToSelection                uintptr
	RemoveFromSelection           uintptr
	Get_CurrentIsSelected         uintptr
	Get_CurrentSelectionContainer uintptr
	Get_CachedIsSelected          uintptr
	Get_CachedSelectionContainer  uintptr
}

type IUIAutomationSelectionItemPattern struct {
	IUnknown
}

func (this *IUIAutomationSelectionItemPattern) Vtbl() *IUIAutomationSelectionItemPatternVtbl {
	return (*IUIAutomationSelectionItemPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationSelectionItemPattern) Select() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Select, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionItemPattern) AddToSelection() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddToSelection, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionItemPattern) RemoveFromSelection() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveFromSelection, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionItemPattern) Get_CurrentIsSelected(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsSelected, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionItemPattern) Get_CurrentSelectionContainer(retVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentSelectionContainer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionItemPattern) Get_CachedIsSelected(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsSelected, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSelectionItemPattern) Get_CachedSelectionContainer(retVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedSelectionContainer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 2233BE0B-AFB7-448B-9FDA-3B378AA5EAE1
var IID_IUIAutomationSynchronizedInputPattern = syscall.GUID{0x2233BE0B, 0xAFB7, 0x448B,
	[8]byte{0x9F, 0xDA, 0x3B, 0x37, 0x8A, 0xA5, 0xEA, 0xE1}}

type IUIAutomationSynchronizedInputPatternInterface interface {
	IUnknownInterface
	StartListening(inputType SynchronizedInputType) HRESULT
	Cancel() HRESULT
}

type IUIAutomationSynchronizedInputPatternVtbl struct {
	IUnknownVtbl
	StartListening uintptr
	Cancel         uintptr
}

type IUIAutomationSynchronizedInputPattern struct {
	IUnknown
}

func (this *IUIAutomationSynchronizedInputPattern) Vtbl() *IUIAutomationSynchronizedInputPatternVtbl {
	return (*IUIAutomationSynchronizedInputPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationSynchronizedInputPattern) StartListening(inputType SynchronizedInputType) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().StartListening, uintptr(unsafe.Pointer(this)), uintptr(inputType))
	return HRESULT(ret)
}

func (this *IUIAutomationSynchronizedInputPattern) Cancel() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Cancel, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 620E691C-EA96-4710-A850-754B24CE2417
var IID_IUIAutomationTablePattern = syscall.GUID{0x620E691C, 0xEA96, 0x4710,
	[8]byte{0xA8, 0x50, 0x75, 0x4B, 0x24, 0xCE, 0x24, 0x17}}

type IUIAutomationTablePatternInterface interface {
	IUnknownInterface
	GetCurrentRowHeaders(retVal **IUIAutomationElementArray) HRESULT
	GetCurrentColumnHeaders(retVal **IUIAutomationElementArray) HRESULT
	Get_CurrentRowOrColumnMajor(retVal *RowOrColumnMajor) HRESULT
	GetCachedRowHeaders(retVal **IUIAutomationElementArray) HRESULT
	GetCachedColumnHeaders(retVal **IUIAutomationElementArray) HRESULT
	Get_CachedRowOrColumnMajor(retVal *RowOrColumnMajor) HRESULT
}

type IUIAutomationTablePatternVtbl struct {
	IUnknownVtbl
	GetCurrentRowHeaders        uintptr
	GetCurrentColumnHeaders     uintptr
	Get_CurrentRowOrColumnMajor uintptr
	GetCachedRowHeaders         uintptr
	GetCachedColumnHeaders      uintptr
	Get_CachedRowOrColumnMajor  uintptr
}

type IUIAutomationTablePattern struct {
	IUnknown
}

func (this *IUIAutomationTablePattern) Vtbl() *IUIAutomationTablePatternVtbl {
	return (*IUIAutomationTablePatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTablePattern) GetCurrentRowHeaders(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentRowHeaders, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTablePattern) GetCurrentColumnHeaders(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentColumnHeaders, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTablePattern) Get_CurrentRowOrColumnMajor(retVal *RowOrColumnMajor) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentRowOrColumnMajor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTablePattern) GetCachedRowHeaders(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedRowHeaders, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTablePattern) GetCachedColumnHeaders(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedColumnHeaders, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTablePattern) Get_CachedRowOrColumnMajor(retVal *RowOrColumnMajor) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedRowOrColumnMajor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 0B964EB3-EF2E-4464-9C79-61D61737A27E
var IID_IUIAutomationTableItemPattern = syscall.GUID{0x0B964EB3, 0xEF2E, 0x4464,
	[8]byte{0x9C, 0x79, 0x61, 0xD6, 0x17, 0x37, 0xA2, 0x7E}}

type IUIAutomationTableItemPatternInterface interface {
	IUnknownInterface
	GetCurrentRowHeaderItems(retVal **IUIAutomationElementArray) HRESULT
	GetCurrentColumnHeaderItems(retVal **IUIAutomationElementArray) HRESULT
	GetCachedRowHeaderItems(retVal **IUIAutomationElementArray) HRESULT
	GetCachedColumnHeaderItems(retVal **IUIAutomationElementArray) HRESULT
}

type IUIAutomationTableItemPatternVtbl struct {
	IUnknownVtbl
	GetCurrentRowHeaderItems    uintptr
	GetCurrentColumnHeaderItems uintptr
	GetCachedRowHeaderItems     uintptr
	GetCachedColumnHeaderItems  uintptr
}

type IUIAutomationTableItemPattern struct {
	IUnknown
}

func (this *IUIAutomationTableItemPattern) Vtbl() *IUIAutomationTableItemPatternVtbl {
	return (*IUIAutomationTableItemPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTableItemPattern) GetCurrentRowHeaderItems(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentRowHeaderItems, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTableItemPattern) GetCurrentColumnHeaderItems(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentColumnHeaderItems, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTableItemPattern) GetCachedRowHeaderItems(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedRowHeaderItems, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTableItemPattern) GetCachedColumnHeaderItems(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedColumnHeaderItems, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 94CF8058-9B8D-4AB9-8BFD-4CD0A33C8C70
var IID_IUIAutomationTogglePattern = syscall.GUID{0x94CF8058, 0x9B8D, 0x4AB9,
	[8]byte{0x8B, 0xFD, 0x4C, 0xD0, 0xA3, 0x3C, 0x8C, 0x70}}

type IUIAutomationTogglePatternInterface interface {
	IUnknownInterface
	Toggle() HRESULT
	Get_CurrentToggleState(retVal *ToggleState) HRESULT
	Get_CachedToggleState(retVal *ToggleState) HRESULT
}

type IUIAutomationTogglePatternVtbl struct {
	IUnknownVtbl
	Toggle                 uintptr
	Get_CurrentToggleState uintptr
	Get_CachedToggleState  uintptr
}

type IUIAutomationTogglePattern struct {
	IUnknown
}

func (this *IUIAutomationTogglePattern) Vtbl() *IUIAutomationTogglePatternVtbl {
	return (*IUIAutomationTogglePatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTogglePattern) Toggle() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Toggle, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomationTogglePattern) Get_CurrentToggleState(retVal *ToggleState) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentToggleState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTogglePattern) Get_CachedToggleState(retVal *ToggleState) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedToggleState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// A9B55844-A55D-4EF0-926D-569C16FF89BB
var IID_IUIAutomationTransformPattern = syscall.GUID{0xA9B55844, 0xA55D, 0x4EF0,
	[8]byte{0x92, 0x6D, 0x56, 0x9C, 0x16, 0xFF, 0x89, 0xBB}}

type IUIAutomationTransformPatternInterface interface {
	IUnknownInterface
	Move(x float64, y float64) HRESULT
	Resize(width float64, height float64) HRESULT
	Rotate(degrees float64) HRESULT
	Get_CurrentCanMove(retVal *BOOL) HRESULT
	Get_CurrentCanResize(retVal *BOOL) HRESULT
	Get_CurrentCanRotate(retVal *BOOL) HRESULT
	Get_CachedCanMove(retVal *BOOL) HRESULT
	Get_CachedCanResize(retVal *BOOL) HRESULT
	Get_CachedCanRotate(retVal *BOOL) HRESULT
}

type IUIAutomationTransformPatternVtbl struct {
	IUnknownVtbl
	Move                 uintptr
	Resize               uintptr
	Rotate               uintptr
	Get_CurrentCanMove   uintptr
	Get_CurrentCanResize uintptr
	Get_CurrentCanRotate uintptr
	Get_CachedCanMove    uintptr
	Get_CachedCanResize  uintptr
	Get_CachedCanRotate  uintptr
}

type IUIAutomationTransformPattern struct {
	IUnknown
}

func (this *IUIAutomationTransformPattern) Vtbl() *IUIAutomationTransformPatternVtbl {
	return (*IUIAutomationTransformPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTransformPattern) Move(x float64, y float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Move, uintptr(unsafe.Pointer(this)), uintptr(x), uintptr(y))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern) Resize(width float64, height float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Resize, uintptr(unsafe.Pointer(this)), uintptr(width), uintptr(height))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern) Rotate(degrees float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Rotate, uintptr(unsafe.Pointer(this)), uintptr(degrees))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern) Get_CurrentCanMove(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentCanMove, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern) Get_CurrentCanResize(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentCanResize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern) Get_CurrentCanRotate(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentCanRotate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern) Get_CachedCanMove(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedCanMove, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern) Get_CachedCanResize(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedCanResize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern) Get_CachedCanRotate(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedCanRotate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// A94CD8B1-0844-4CD6-9D2D-640537AB39E9
var IID_IUIAutomationValuePattern = syscall.GUID{0xA94CD8B1, 0x0844, 0x4CD6,
	[8]byte{0x9D, 0x2D, 0x64, 0x05, 0x37, 0xAB, 0x39, 0xE9}}

type IUIAutomationValuePatternInterface interface {
	IUnknownInterface
	SetValue(val BSTR) HRESULT
	Get_CurrentValue(retVal *BSTR) HRESULT
	Get_CurrentIsReadOnly(retVal *BOOL) HRESULT
	Get_CachedValue(retVal *BSTR) HRESULT
	Get_CachedIsReadOnly(retVal *BOOL) HRESULT
}

type IUIAutomationValuePatternVtbl struct {
	IUnknownVtbl
	SetValue              uintptr
	Get_CurrentValue      uintptr
	Get_CurrentIsReadOnly uintptr
	Get_CachedValue       uintptr
	Get_CachedIsReadOnly  uintptr
}

type IUIAutomationValuePattern struct {
	IUnknown
}

func (this *IUIAutomationValuePattern) Vtbl() *IUIAutomationValuePatternVtbl {
	return (*IUIAutomationValuePatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationValuePattern) SetValue(val BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(val)))
	return HRESULT(ret)
}

func (this *IUIAutomationValuePattern) Get_CurrentValue(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationValuePattern) Get_CurrentIsReadOnly(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsReadOnly, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationValuePattern) Get_CachedValue(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationValuePattern) Get_CachedIsReadOnly(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsReadOnly, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 0FAEF453-9208-43EF-BBB2-3B485177864F
var IID_IUIAutomationWindowPattern = syscall.GUID{0x0FAEF453, 0x9208, 0x43EF,
	[8]byte{0xBB, 0xB2, 0x3B, 0x48, 0x51, 0x77, 0x86, 0x4F}}

type IUIAutomationWindowPatternInterface interface {
	IUnknownInterface
	Close() HRESULT
	WaitForInputIdle(milliseconds int32, success *BOOL) HRESULT
	SetWindowVisualState(state WindowVisualState) HRESULT
	Get_CurrentCanMaximize(retVal *BOOL) HRESULT
	Get_CurrentCanMinimize(retVal *BOOL) HRESULT
	Get_CurrentIsModal(retVal *BOOL) HRESULT
	Get_CurrentIsTopmost(retVal *BOOL) HRESULT
	Get_CurrentWindowVisualState(retVal *WindowVisualState) HRESULT
	Get_CurrentWindowInteractionState(retVal *WindowInteractionState) HRESULT
	Get_CachedCanMaximize(retVal *BOOL) HRESULT
	Get_CachedCanMinimize(retVal *BOOL) HRESULT
	Get_CachedIsModal(retVal *BOOL) HRESULT
	Get_CachedIsTopmost(retVal *BOOL) HRESULT
	Get_CachedWindowVisualState(retVal *WindowVisualState) HRESULT
	Get_CachedWindowInteractionState(retVal *WindowInteractionState) HRESULT
}

type IUIAutomationWindowPatternVtbl struct {
	IUnknownVtbl
	Close                             uintptr
	WaitForInputIdle                  uintptr
	SetWindowVisualState              uintptr
	Get_CurrentCanMaximize            uintptr
	Get_CurrentCanMinimize            uintptr
	Get_CurrentIsModal                uintptr
	Get_CurrentIsTopmost              uintptr
	Get_CurrentWindowVisualState      uintptr
	Get_CurrentWindowInteractionState uintptr
	Get_CachedCanMaximize             uintptr
	Get_CachedCanMinimize             uintptr
	Get_CachedIsModal                 uintptr
	Get_CachedIsTopmost               uintptr
	Get_CachedWindowVisualState       uintptr
	Get_CachedWindowInteractionState  uintptr
}

type IUIAutomationWindowPattern struct {
	IUnknown
}

func (this *IUIAutomationWindowPattern) Vtbl() *IUIAutomationWindowPatternVtbl {
	return (*IUIAutomationWindowPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationWindowPattern) Close() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Close, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomationWindowPattern) WaitForInputIdle(milliseconds int32, success *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().WaitForInputIdle, uintptr(unsafe.Pointer(this)), uintptr(milliseconds), uintptr(unsafe.Pointer(success)))
	return HRESULT(ret)
}

func (this *IUIAutomationWindowPattern) SetWindowVisualState(state WindowVisualState) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetWindowVisualState, uintptr(unsafe.Pointer(this)), uintptr(state))
	return HRESULT(ret)
}

func (this *IUIAutomationWindowPattern) Get_CurrentCanMaximize(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentCanMaximize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationWindowPattern) Get_CurrentCanMinimize(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentCanMinimize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationWindowPattern) Get_CurrentIsModal(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsModal, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationWindowPattern) Get_CurrentIsTopmost(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsTopmost, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationWindowPattern) Get_CurrentWindowVisualState(retVal *WindowVisualState) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentWindowVisualState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationWindowPattern) Get_CurrentWindowInteractionState(retVal *WindowInteractionState) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentWindowInteractionState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationWindowPattern) Get_CachedCanMaximize(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedCanMaximize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationWindowPattern) Get_CachedCanMinimize(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedCanMinimize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationWindowPattern) Get_CachedIsModal(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsModal, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationWindowPattern) Get_CachedIsTopmost(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsTopmost, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationWindowPattern) Get_CachedWindowVisualState(retVal *WindowVisualState) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedWindowVisualState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationWindowPattern) Get_CachedWindowInteractionState(retVal *WindowInteractionState) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedWindowInteractionState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// A543CC6A-F4AE-494B-8239-C814481187A8
var IID_IUIAutomationTextRange = syscall.GUID{0xA543CC6A, 0xF4AE, 0x494B,
	[8]byte{0x82, 0x39, 0xC8, 0x14, 0x48, 0x11, 0x87, 0xA8}}

type IUIAutomationTextRangeInterface interface {
	IUnknownInterface
	Clone(clonedRange **IUIAutomationTextRange) HRESULT
	Compare(range_ *IUIAutomationTextRange, areSame *BOOL) HRESULT
	CompareEndpoints(srcEndPoint TextPatternRangeEndpoint, range_ *IUIAutomationTextRange, targetEndPoint TextPatternRangeEndpoint, compValue *int32) HRESULT
	ExpandToEnclosingUnit(textUnit TextUnit) HRESULT
	FindAttribute(attr int32, val VARIANT, backward BOOL, found **IUIAutomationTextRange) HRESULT
	FindText(text BSTR, backward BOOL, ignoreCase BOOL, found **IUIAutomationTextRange) HRESULT
	GetAttributeValue(attr int32, value *VARIANT) HRESULT
	GetBoundingRectangles(boundingRects **SAFEARRAY) HRESULT
	GetEnclosingElement(enclosingElement **IUIAutomationElement) HRESULT
	GetText(maxLength int32, text *BSTR) HRESULT
	Move(unit TextUnit, count int32, moved *int32) HRESULT
	MoveEndpointByUnit(endpoint TextPatternRangeEndpoint, unit TextUnit, count int32, moved *int32) HRESULT
	MoveEndpointByRange(srcEndPoint TextPatternRangeEndpoint, range_ *IUIAutomationTextRange, targetEndPoint TextPatternRangeEndpoint) HRESULT
	Select() HRESULT
	AddToSelection() HRESULT
	RemoveFromSelection() HRESULT
	ScrollIntoView(alignToTop BOOL) HRESULT
	GetChildren(children **IUIAutomationElementArray) HRESULT
}

type IUIAutomationTextRangeVtbl struct {
	IUnknownVtbl
	Clone                 uintptr
	Compare               uintptr
	CompareEndpoints      uintptr
	ExpandToEnclosingUnit uintptr
	FindAttribute         uintptr
	FindText              uintptr
	GetAttributeValue     uintptr
	GetBoundingRectangles uintptr
	GetEnclosingElement   uintptr
	GetText               uintptr
	Move                  uintptr
	MoveEndpointByUnit    uintptr
	MoveEndpointByRange   uintptr
	Select                uintptr
	AddToSelection        uintptr
	RemoveFromSelection   uintptr
	ScrollIntoView        uintptr
	GetChildren           uintptr
}

type IUIAutomationTextRange struct {
	IUnknown
}

func (this *IUIAutomationTextRange) Vtbl() *IUIAutomationTextRangeVtbl {
	return (*IUIAutomationTextRangeVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTextRange) Clone(clonedRange **IUIAutomationTextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(clonedRange)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) Compare(range_ *IUIAutomationTextRange, areSame *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Compare, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(range_)), uintptr(unsafe.Pointer(areSame)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) CompareEndpoints(srcEndPoint TextPatternRangeEndpoint, range_ *IUIAutomationTextRange, targetEndPoint TextPatternRangeEndpoint, compValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CompareEndpoints, uintptr(unsafe.Pointer(this)), uintptr(srcEndPoint), uintptr(unsafe.Pointer(range_)), uintptr(targetEndPoint), uintptr(unsafe.Pointer(compValue)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) ExpandToEnclosingUnit(textUnit TextUnit) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ExpandToEnclosingUnit, uintptr(unsafe.Pointer(this)), uintptr(textUnit))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) FindAttribute(attr int32, val VARIANT, backward BOOL, found **IUIAutomationTextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindAttribute, uintptr(unsafe.Pointer(this)), uintptr(attr), uintptr(unsafe.Pointer(&val)), uintptr(backward), uintptr(unsafe.Pointer(found)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) FindText(text BSTR, backward BOOL, ignoreCase BOOL, found **IUIAutomationTextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(text)), uintptr(backward), uintptr(ignoreCase), uintptr(unsafe.Pointer(found)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) GetAttributeValue(attr int32, value *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAttributeValue, uintptr(unsafe.Pointer(this)), uintptr(attr), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) GetBoundingRectangles(boundingRects **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBoundingRectangles, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(boundingRects)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) GetEnclosingElement(enclosingElement **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEnclosingElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(enclosingElement)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) GetText(maxLength int32, text *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetText, uintptr(unsafe.Pointer(this)), uintptr(maxLength), uintptr(unsafe.Pointer(text)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) Move(unit TextUnit, count int32, moved *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Move, uintptr(unsafe.Pointer(this)), uintptr(unit), uintptr(count), uintptr(unsafe.Pointer(moved)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) MoveEndpointByUnit(endpoint TextPatternRangeEndpoint, unit TextUnit, count int32, moved *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveEndpointByUnit, uintptr(unsafe.Pointer(this)), uintptr(endpoint), uintptr(unit), uintptr(count), uintptr(unsafe.Pointer(moved)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) MoveEndpointByRange(srcEndPoint TextPatternRangeEndpoint, range_ *IUIAutomationTextRange, targetEndPoint TextPatternRangeEndpoint) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveEndpointByRange, uintptr(unsafe.Pointer(this)), uintptr(srcEndPoint), uintptr(unsafe.Pointer(range_)), uintptr(targetEndPoint))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) Select() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Select, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) AddToSelection() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddToSelection, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) RemoveFromSelection() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveFromSelection, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) ScrollIntoView(alignToTop BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ScrollIntoView, uintptr(unsafe.Pointer(this)), uintptr(alignToTop))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange) GetChildren(children **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetChildren, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(children)))
	return HRESULT(ret)
}

// BB9B40E0-5E04-46BD-9BE0-4B601B9AFAD4
var IID_IUIAutomationTextRange2 = syscall.GUID{0xBB9B40E0, 0x5E04, 0x46BD,
	[8]byte{0x9B, 0xE0, 0x4B, 0x60, 0x1B, 0x9A, 0xFA, 0xD4}}

type IUIAutomationTextRange2Interface interface {
	IUIAutomationTextRangeInterface
	ShowContextMenu() HRESULT
}

type IUIAutomationTextRange2Vtbl struct {
	IUIAutomationTextRangeVtbl
	ShowContextMenu uintptr
}

type IUIAutomationTextRange2 struct {
	IUIAutomationTextRange
}

func (this *IUIAutomationTextRange2) Vtbl() *IUIAutomationTextRange2Vtbl {
	return (*IUIAutomationTextRange2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTextRange2) ShowContextMenu() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ShowContextMenu, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 6A315D69-5512-4C2E-85F0-53FCE6DD4BC2
var IID_IUIAutomationTextRange3 = syscall.GUID{0x6A315D69, 0x5512, 0x4C2E,
	[8]byte{0x85, 0xF0, 0x53, 0xFC, 0xE6, 0xDD, 0x4B, 0xC2}}

type IUIAutomationTextRange3Interface interface {
	IUIAutomationTextRange2Interface
	GetEnclosingElementBuildCache(cacheRequest *IUIAutomationCacheRequest, enclosingElement **IUIAutomationElement) HRESULT
	GetChildrenBuildCache(cacheRequest *IUIAutomationCacheRequest, children **IUIAutomationElementArray) HRESULT
	GetAttributeValues(attributeIds *int32, attributeIdCount int32, attributeValues **SAFEARRAY) HRESULT
}

type IUIAutomationTextRange3Vtbl struct {
	IUIAutomationTextRange2Vtbl
	GetEnclosingElementBuildCache uintptr
	GetChildrenBuildCache         uintptr
	GetAttributeValues            uintptr
}

type IUIAutomationTextRange3 struct {
	IUIAutomationTextRange2
}

func (this *IUIAutomationTextRange3) Vtbl() *IUIAutomationTextRange3Vtbl {
	return (*IUIAutomationTextRange3Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTextRange3) GetEnclosingElementBuildCache(cacheRequest *IUIAutomationCacheRequest, enclosingElement **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEnclosingElementBuildCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(enclosingElement)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange3) GetChildrenBuildCache(cacheRequest *IUIAutomationCacheRequest, children **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetChildrenBuildCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(children)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRange3) GetAttributeValues(attributeIds *int32, attributeIdCount int32, attributeValues **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAttributeValues, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(attributeIds)), uintptr(attributeIdCount), uintptr(unsafe.Pointer(attributeValues)))
	return HRESULT(ret)
}

// CE4AE76A-E717-4C98-81EA-47371D028EB6
var IID_IUIAutomationTextRangeArray = syscall.GUID{0xCE4AE76A, 0xE717, 0x4C98,
	[8]byte{0x81, 0xEA, 0x47, 0x37, 0x1D, 0x02, 0x8E, 0xB6}}

type IUIAutomationTextRangeArrayInterface interface {
	IUnknownInterface
	Get_Length(length *int32) HRESULT
	GetElement(index int32, element **IUIAutomationTextRange) HRESULT
}

type IUIAutomationTextRangeArrayVtbl struct {
	IUnknownVtbl
	Get_Length uintptr
	GetElement uintptr
}

type IUIAutomationTextRangeArray struct {
	IUnknown
}

func (this *IUIAutomationTextRangeArray) Vtbl() *IUIAutomationTextRangeArrayVtbl {
	return (*IUIAutomationTextRangeArrayVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTextRangeArray) Get_Length(length *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Length, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(length)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextRangeArray) GetElement(index int32, element **IUIAutomationTextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetElement, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(element)))
	return HRESULT(ret)
}

// 32EBA289-3583-42C9-9C59-3B6D9A1E9B6A
var IID_IUIAutomationTextPattern = syscall.GUID{0x32EBA289, 0x3583, 0x42C9,
	[8]byte{0x9C, 0x59, 0x3B, 0x6D, 0x9A, 0x1E, 0x9B, 0x6A}}

type IUIAutomationTextPatternInterface interface {
	IUnknownInterface
	RangeFromPoint(pt POINT, range_ **IUIAutomationTextRange) HRESULT
	RangeFromChild(child *IUIAutomationElement, range_ **IUIAutomationTextRange) HRESULT
	GetSelection(ranges **IUIAutomationTextRangeArray) HRESULT
	GetVisibleRanges(ranges **IUIAutomationTextRangeArray) HRESULT
	Get_DocumentRange(range_ **IUIAutomationTextRange) HRESULT
	Get_SupportedTextSelection(supportedTextSelection *SupportedTextSelection) HRESULT
}

type IUIAutomationTextPatternVtbl struct {
	IUnknownVtbl
	RangeFromPoint             uintptr
	RangeFromChild             uintptr
	GetSelection               uintptr
	GetVisibleRanges           uintptr
	Get_DocumentRange          uintptr
	Get_SupportedTextSelection uintptr
}

type IUIAutomationTextPattern struct {
	IUnknown
}

func (this *IUIAutomationTextPattern) Vtbl() *IUIAutomationTextPatternVtbl {
	return (*IUIAutomationTextPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTextPattern) RangeFromPoint(pt POINT, range_ **IUIAutomationTextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RangeFromPoint, uintptr(unsafe.Pointer(this)), *(*uintptr)(unsafe.Pointer(&pt)), uintptr(unsafe.Pointer(range_)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextPattern) RangeFromChild(child *IUIAutomationElement, range_ **IUIAutomationTextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RangeFromChild, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(child)), uintptr(unsafe.Pointer(range_)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextPattern) GetSelection(ranges **IUIAutomationTextRangeArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSelection, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ranges)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextPattern) GetVisibleRanges(ranges **IUIAutomationTextRangeArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetVisibleRanges, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ranges)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextPattern) Get_DocumentRange(range_ **IUIAutomationTextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_DocumentRange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(range_)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextPattern) Get_SupportedTextSelection(supportedTextSelection *SupportedTextSelection) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_SupportedTextSelection, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(supportedTextSelection)))
	return HRESULT(ret)
}

// 506A921A-FCC9-409F-B23B-37EB74106872
var IID_IUIAutomationTextPattern2 = syscall.GUID{0x506A921A, 0xFCC9, 0x409F,
	[8]byte{0xB2, 0x3B, 0x37, 0xEB, 0x74, 0x10, 0x68, 0x72}}

type IUIAutomationTextPattern2Interface interface {
	IUIAutomationTextPatternInterface
	RangeFromAnnotation(annotation *IUIAutomationElement, range_ **IUIAutomationTextRange) HRESULT
	GetCaretRange(isActive *BOOL, range_ **IUIAutomationTextRange) HRESULT
}

type IUIAutomationTextPattern2Vtbl struct {
	IUIAutomationTextPatternVtbl
	RangeFromAnnotation uintptr
	GetCaretRange       uintptr
}

type IUIAutomationTextPattern2 struct {
	IUIAutomationTextPattern
}

func (this *IUIAutomationTextPattern2) Vtbl() *IUIAutomationTextPattern2Vtbl {
	return (*IUIAutomationTextPattern2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTextPattern2) RangeFromAnnotation(annotation *IUIAutomationElement, range_ **IUIAutomationTextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RangeFromAnnotation, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(annotation)), uintptr(unsafe.Pointer(range_)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextPattern2) GetCaretRange(isActive *BOOL, range_ **IUIAutomationTextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCaretRange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(isActive)), uintptr(unsafe.Pointer(range_)))
	return HRESULT(ret)
}

// 17E21576-996C-4870-99D9-BFF323380C06
var IID_IUIAutomationTextEditPattern = syscall.GUID{0x17E21576, 0x996C, 0x4870,
	[8]byte{0x99, 0xD9, 0xBF, 0xF3, 0x23, 0x38, 0x0C, 0x06}}

type IUIAutomationTextEditPatternInterface interface {
	IUIAutomationTextPatternInterface
	GetActiveComposition(range_ **IUIAutomationTextRange) HRESULT
	GetConversionTarget(range_ **IUIAutomationTextRange) HRESULT
}

type IUIAutomationTextEditPatternVtbl struct {
	IUIAutomationTextPatternVtbl
	GetActiveComposition uintptr
	GetConversionTarget  uintptr
}

type IUIAutomationTextEditPattern struct {
	IUIAutomationTextPattern
}

func (this *IUIAutomationTextEditPattern) Vtbl() *IUIAutomationTextEditPatternVtbl {
	return (*IUIAutomationTextEditPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTextEditPattern) GetActiveComposition(range_ **IUIAutomationTextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetActiveComposition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(range_)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextEditPattern) GetConversionTarget(range_ **IUIAutomationTextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetConversionTarget, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(range_)))
	return HRESULT(ret)
}

// 01EA217A-1766-47ED-A6CC-ACF492854B1F
var IID_IUIAutomationCustomNavigationPattern = syscall.GUID{0x01EA217A, 0x1766, 0x47ED,
	[8]byte{0xA6, 0xCC, 0xAC, 0xF4, 0x92, 0x85, 0x4B, 0x1F}}

type IUIAutomationCustomNavigationPatternInterface interface {
	IUnknownInterface
	Navigate(direction NavigateDirection, pRetVal **IUIAutomationElement) HRESULT
}

type IUIAutomationCustomNavigationPatternVtbl struct {
	IUnknownVtbl
	Navigate uintptr
}

type IUIAutomationCustomNavigationPattern struct {
	IUnknown
}

func (this *IUIAutomationCustomNavigationPattern) Vtbl() *IUIAutomationCustomNavigationPatternVtbl {
	return (*IUIAutomationCustomNavigationPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationCustomNavigationPattern) Navigate(direction NavigateDirection, pRetVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Navigate, uintptr(unsafe.Pointer(this)), uintptr(direction), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

// F97933B0-8DAE-4496-8997-5BA015FE0D82
var IID_IUIAutomationActiveTextPositionChangedEventHandler = syscall.GUID{0xF97933B0, 0x8DAE, 0x4496,
	[8]byte{0x89, 0x97, 0x5B, 0xA0, 0x15, 0xFE, 0x0D, 0x82}}

type IUIAutomationActiveTextPositionChangedEventHandlerInterface interface {
	IUnknownInterface
	HandleActiveTextPositionChangedEvent(sender *IUIAutomationElement, range_ *IUIAutomationTextRange) HRESULT
}

type IUIAutomationActiveTextPositionChangedEventHandlerVtbl struct {
	IUnknownVtbl
	HandleActiveTextPositionChangedEvent uintptr
}

type IUIAutomationActiveTextPositionChangedEventHandler struct {
	IUnknown
}

func (this *IUIAutomationActiveTextPositionChangedEventHandler) Vtbl() *IUIAutomationActiveTextPositionChangedEventHandlerVtbl {
	return (*IUIAutomationActiveTextPositionChangedEventHandlerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationActiveTextPositionChangedEventHandler) HandleActiveTextPositionChangedEvent(sender *IUIAutomationElement, range_ *IUIAutomationTextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HandleActiveTextPositionChangedEvent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(sender)), uintptr(unsafe.Pointer(range_)))
	return HRESULT(ret)
}

// 828055AD-355B-4435-86D5-3B51C14A9B1B
var IID_IUIAutomationLegacyIAccessiblePattern = syscall.GUID{0x828055AD, 0x355B, 0x4435,
	[8]byte{0x86, 0xD5, 0x3B, 0x51, 0xC1, 0x4A, 0x9B, 0x1B}}

type IUIAutomationLegacyIAccessiblePatternInterface interface {
	IUnknownInterface
	Select(flagsSelect int32) HRESULT
	DoDefaultAction() HRESULT
	SetValue(szValue PWSTR) HRESULT
	Get_CurrentChildId(pRetVal *int32) HRESULT
	Get_CurrentName(pszName *BSTR) HRESULT
	Get_CurrentValue(pszValue *BSTR) HRESULT
	Get_CurrentDescription(pszDescription *BSTR) HRESULT
	Get_CurrentRole(pdwRole *uint32) HRESULT
	Get_CurrentState(pdwState *uint32) HRESULT
	Get_CurrentHelp(pszHelp *BSTR) HRESULT
	Get_CurrentKeyboardShortcut(pszKeyboardShortcut *BSTR) HRESULT
	GetCurrentSelection(pvarSelectedChildren **IUIAutomationElementArray) HRESULT
	Get_CurrentDefaultAction(pszDefaultAction *BSTR) HRESULT
	Get_CachedChildId(pRetVal *int32) HRESULT
	Get_CachedName(pszName *BSTR) HRESULT
	Get_CachedValue(pszValue *BSTR) HRESULT
	Get_CachedDescription(pszDescription *BSTR) HRESULT
	Get_CachedRole(pdwRole *uint32) HRESULT
	Get_CachedState(pdwState *uint32) HRESULT
	Get_CachedHelp(pszHelp *BSTR) HRESULT
	Get_CachedKeyboardShortcut(pszKeyboardShortcut *BSTR) HRESULT
	GetCachedSelection(pvarSelectedChildren **IUIAutomationElementArray) HRESULT
	Get_CachedDefaultAction(pszDefaultAction *BSTR) HRESULT
	GetIAccessible(ppAccessible **IAccessible) HRESULT
}

type IUIAutomationLegacyIAccessiblePatternVtbl struct {
	IUnknownVtbl
	Select                      uintptr
	DoDefaultAction             uintptr
	SetValue                    uintptr
	Get_CurrentChildId          uintptr
	Get_CurrentName             uintptr
	Get_CurrentValue            uintptr
	Get_CurrentDescription      uintptr
	Get_CurrentRole             uintptr
	Get_CurrentState            uintptr
	Get_CurrentHelp             uintptr
	Get_CurrentKeyboardShortcut uintptr
	GetCurrentSelection         uintptr
	Get_CurrentDefaultAction    uintptr
	Get_CachedChildId           uintptr
	Get_CachedName              uintptr
	Get_CachedValue             uintptr
	Get_CachedDescription       uintptr
	Get_CachedRole              uintptr
	Get_CachedState             uintptr
	Get_CachedHelp              uintptr
	Get_CachedKeyboardShortcut  uintptr
	GetCachedSelection          uintptr
	Get_CachedDefaultAction     uintptr
	GetIAccessible              uintptr
}

type IUIAutomationLegacyIAccessiblePattern struct {
	IUnknown
}

func (this *IUIAutomationLegacyIAccessiblePattern) Vtbl() *IUIAutomationLegacyIAccessiblePatternVtbl {
	return (*IUIAutomationLegacyIAccessiblePatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationLegacyIAccessiblePattern) Select(flagsSelect int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Select, uintptr(unsafe.Pointer(this)), uintptr(flagsSelect))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) DoDefaultAction() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DoDefaultAction, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) SetValue(szValue PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szValue)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CurrentChildId(pRetVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentChildId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CurrentName(pszName *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszName)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CurrentValue(pszValue *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszValue)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CurrentDescription(pszDescription *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszDescription)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CurrentRole(pdwRole *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentRole, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwRole)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CurrentState(pdwState *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwState)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CurrentHelp(pszHelp *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentHelp, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszHelp)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CurrentKeyboardShortcut(pszKeyboardShortcut *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentKeyboardShortcut, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszKeyboardShortcut)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) GetCurrentSelection(pvarSelectedChildren **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentSelection, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pvarSelectedChildren)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CurrentDefaultAction(pszDefaultAction *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentDefaultAction, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszDefaultAction)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CachedChildId(pRetVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedChildId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRetVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CachedName(pszName *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszName)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CachedValue(pszValue *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszValue)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CachedDescription(pszDescription *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszDescription)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CachedRole(pdwRole *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedRole, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwRole)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CachedState(pdwState *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwState)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CachedHelp(pszHelp *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedHelp, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszHelp)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CachedKeyboardShortcut(pszKeyboardShortcut *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedKeyboardShortcut, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszKeyboardShortcut)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) GetCachedSelection(pvarSelectedChildren **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedSelection, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pvarSelectedChildren)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) Get_CachedDefaultAction(pszDefaultAction *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedDefaultAction, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszDefaultAction)))
	return HRESULT(ret)
}

func (this *IUIAutomationLegacyIAccessiblePattern) GetIAccessible(ppAccessible **IAccessible) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetIAccessible, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppAccessible)))
	return HRESULT(ret)
}

// C690FDB2-27A8-423C-812D-429773C9084E
var IID_IUIAutomationItemContainerPattern = syscall.GUID{0xC690FDB2, 0x27A8, 0x423C,
	[8]byte{0x81, 0x2D, 0x42, 0x97, 0x73, 0xC9, 0x08, 0x4E}}

type IUIAutomationItemContainerPatternInterface interface {
	IUnknownInterface
	FindItemByProperty(pStartAfter *IUIAutomationElement, propertyId int32, value VARIANT, pFound **IUIAutomationElement) HRESULT
}

type IUIAutomationItemContainerPatternVtbl struct {
	IUnknownVtbl
	FindItemByProperty uintptr
}

type IUIAutomationItemContainerPattern struct {
	IUnknown
}

func (this *IUIAutomationItemContainerPattern) Vtbl() *IUIAutomationItemContainerPatternVtbl {
	return (*IUIAutomationItemContainerPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationItemContainerPattern) FindItemByProperty(pStartAfter *IUIAutomationElement, propertyId int32, value VARIANT, pFound **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindItemByProperty, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pStartAfter)), uintptr(propertyId), uintptr(unsafe.Pointer(&value)), uintptr(unsafe.Pointer(pFound)))
	return HRESULT(ret)
}

// 6BA3D7A6-04CF-4F11-8793-A8D1CDE9969F
var IID_IUIAutomationVirtualizedItemPattern = syscall.GUID{0x6BA3D7A6, 0x04CF, 0x4F11,
	[8]byte{0x87, 0x93, 0xA8, 0xD1, 0xCD, 0xE9, 0x96, 0x9F}}

type IUIAutomationVirtualizedItemPatternInterface interface {
	IUnknownInterface
	Realize() HRESULT
}

type IUIAutomationVirtualizedItemPatternVtbl struct {
	IUnknownVtbl
	Realize uintptr
}

type IUIAutomationVirtualizedItemPattern struct {
	IUnknown
}

func (this *IUIAutomationVirtualizedItemPattern) Vtbl() *IUIAutomationVirtualizedItemPatternVtbl {
	return (*IUIAutomationVirtualizedItemPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationVirtualizedItemPattern) Realize() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Realize, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 9A175B21-339E-41B1-8E8B-623F6B681098
var IID_IUIAutomationAnnotationPattern = syscall.GUID{0x9A175B21, 0x339E, 0x41B1,
	[8]byte{0x8E, 0x8B, 0x62, 0x3F, 0x6B, 0x68, 0x10, 0x98}}

type IUIAutomationAnnotationPatternInterface interface {
	IUnknownInterface
	Get_CurrentAnnotationTypeId(retVal *int32) HRESULT
	Get_CurrentAnnotationTypeName(retVal *BSTR) HRESULT
	Get_CurrentAuthor(retVal *BSTR) HRESULT
	Get_CurrentDateTime(retVal *BSTR) HRESULT
	Get_CurrentTarget(retVal **IUIAutomationElement) HRESULT
	Get_CachedAnnotationTypeId(retVal *int32) HRESULT
	Get_CachedAnnotationTypeName(retVal *BSTR) HRESULT
	Get_CachedAuthor(retVal *BSTR) HRESULT
	Get_CachedDateTime(retVal *BSTR) HRESULT
	Get_CachedTarget(retVal **IUIAutomationElement) HRESULT
}

type IUIAutomationAnnotationPatternVtbl struct {
	IUnknownVtbl
	Get_CurrentAnnotationTypeId   uintptr
	Get_CurrentAnnotationTypeName uintptr
	Get_CurrentAuthor             uintptr
	Get_CurrentDateTime           uintptr
	Get_CurrentTarget             uintptr
	Get_CachedAnnotationTypeId    uintptr
	Get_CachedAnnotationTypeName  uintptr
	Get_CachedAuthor              uintptr
	Get_CachedDateTime            uintptr
	Get_CachedTarget              uintptr
}

type IUIAutomationAnnotationPattern struct {
	IUnknown
}

func (this *IUIAutomationAnnotationPattern) Vtbl() *IUIAutomationAnnotationPatternVtbl {
	return (*IUIAutomationAnnotationPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationAnnotationPattern) Get_CurrentAnnotationTypeId(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentAnnotationTypeId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationAnnotationPattern) Get_CurrentAnnotationTypeName(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentAnnotationTypeName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationAnnotationPattern) Get_CurrentAuthor(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentAuthor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationAnnotationPattern) Get_CurrentDateTime(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentDateTime, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationAnnotationPattern) Get_CurrentTarget(retVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentTarget, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationAnnotationPattern) Get_CachedAnnotationTypeId(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedAnnotationTypeId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationAnnotationPattern) Get_CachedAnnotationTypeName(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedAnnotationTypeName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationAnnotationPattern) Get_CachedAuthor(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedAuthor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationAnnotationPattern) Get_CachedDateTime(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedDateTime, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationAnnotationPattern) Get_CachedTarget(retVal **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedTarget, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 85B5F0A2-BD79-484A-AD2B-388C9838D5FB
var IID_IUIAutomationStylesPattern = syscall.GUID{0x85B5F0A2, 0xBD79, 0x484A,
	[8]byte{0xAD, 0x2B, 0x38, 0x8C, 0x98, 0x38, 0xD5, 0xFB}}

type IUIAutomationStylesPatternInterface interface {
	IUnknownInterface
	Get_CurrentStyleId(retVal *int32) HRESULT
	Get_CurrentStyleName(retVal *BSTR) HRESULT
	Get_CurrentFillColor(retVal *int32) HRESULT
	Get_CurrentFillPatternStyle(retVal *BSTR) HRESULT
	Get_CurrentShape(retVal *BSTR) HRESULT
	Get_CurrentFillPatternColor(retVal *int32) HRESULT
	Get_CurrentExtendedProperties(retVal *BSTR) HRESULT
	GetCurrentExtendedPropertiesAsArray(propertyArray **ExtendedProperty, propertyCount *int32) HRESULT
	Get_CachedStyleId(retVal *int32) HRESULT
	Get_CachedStyleName(retVal *BSTR) HRESULT
	Get_CachedFillColor(retVal *int32) HRESULT
	Get_CachedFillPatternStyle(retVal *BSTR) HRESULT
	Get_CachedShape(retVal *BSTR) HRESULT
	Get_CachedFillPatternColor(retVal *int32) HRESULT
	Get_CachedExtendedProperties(retVal *BSTR) HRESULT
	GetCachedExtendedPropertiesAsArray(propertyArray **ExtendedProperty, propertyCount *int32) HRESULT
}

type IUIAutomationStylesPatternVtbl struct {
	IUnknownVtbl
	Get_CurrentStyleId                  uintptr
	Get_CurrentStyleName                uintptr
	Get_CurrentFillColor                uintptr
	Get_CurrentFillPatternStyle         uintptr
	Get_CurrentShape                    uintptr
	Get_CurrentFillPatternColor         uintptr
	Get_CurrentExtendedProperties       uintptr
	GetCurrentExtendedPropertiesAsArray uintptr
	Get_CachedStyleId                   uintptr
	Get_CachedStyleName                 uintptr
	Get_CachedFillColor                 uintptr
	Get_CachedFillPatternStyle          uintptr
	Get_CachedShape                     uintptr
	Get_CachedFillPatternColor          uintptr
	Get_CachedExtendedProperties        uintptr
	GetCachedExtendedPropertiesAsArray  uintptr
}

type IUIAutomationStylesPattern struct {
	IUnknown
}

func (this *IUIAutomationStylesPattern) Vtbl() *IUIAutomationStylesPatternVtbl {
	return (*IUIAutomationStylesPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationStylesPattern) Get_CurrentStyleId(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentStyleId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) Get_CurrentStyleName(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentStyleName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) Get_CurrentFillColor(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentFillColor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) Get_CurrentFillPatternStyle(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentFillPatternStyle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) Get_CurrentShape(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentShape, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) Get_CurrentFillPatternColor(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentFillPatternColor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) Get_CurrentExtendedProperties(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentExtendedProperties, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) GetCurrentExtendedPropertiesAsArray(propertyArray **ExtendedProperty, propertyCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentExtendedPropertiesAsArray, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(propertyArray)), uintptr(unsafe.Pointer(propertyCount)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) Get_CachedStyleId(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedStyleId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) Get_CachedStyleName(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedStyleName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) Get_CachedFillColor(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedFillColor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) Get_CachedFillPatternStyle(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedFillPatternStyle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) Get_CachedShape(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedShape, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) Get_CachedFillPatternColor(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedFillPatternColor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) Get_CachedExtendedProperties(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedExtendedProperties, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationStylesPattern) GetCachedExtendedPropertiesAsArray(propertyArray **ExtendedProperty, propertyCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedExtendedPropertiesAsArray, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(propertyArray)), uintptr(unsafe.Pointer(propertyCount)))
	return HRESULT(ret)
}

// 7517A7C8-FAAE-4DE9-9F08-29B91E8595C1
var IID_IUIAutomationSpreadsheetPattern = syscall.GUID{0x7517A7C8, 0xFAAE, 0x4DE9,
	[8]byte{0x9F, 0x08, 0x29, 0xB9, 0x1E, 0x85, 0x95, 0xC1}}

type IUIAutomationSpreadsheetPatternInterface interface {
	IUnknownInterface
	GetItemByName(name BSTR, element **IUIAutomationElement) HRESULT
}

type IUIAutomationSpreadsheetPatternVtbl struct {
	IUnknownVtbl
	GetItemByName uintptr
}

type IUIAutomationSpreadsheetPattern struct {
	IUnknown
}

func (this *IUIAutomationSpreadsheetPattern) Vtbl() *IUIAutomationSpreadsheetPatternVtbl {
	return (*IUIAutomationSpreadsheetPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationSpreadsheetPattern) GetItemByName(name BSTR, element **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetItemByName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(element)))
	return HRESULT(ret)
}

// 7D4FB86C-8D34-40E1-8E83-62C15204E335
var IID_IUIAutomationSpreadsheetItemPattern = syscall.GUID{0x7D4FB86C, 0x8D34, 0x40E1,
	[8]byte{0x8E, 0x83, 0x62, 0xC1, 0x52, 0x04, 0xE3, 0x35}}

type IUIAutomationSpreadsheetItemPatternInterface interface {
	IUnknownInterface
	Get_CurrentFormula(retVal *BSTR) HRESULT
	GetCurrentAnnotationObjects(retVal **IUIAutomationElementArray) HRESULT
	GetCurrentAnnotationTypes(retVal **SAFEARRAY) HRESULT
	Get_CachedFormula(retVal *BSTR) HRESULT
	GetCachedAnnotationObjects(retVal **IUIAutomationElementArray) HRESULT
	GetCachedAnnotationTypes(retVal **SAFEARRAY) HRESULT
}

type IUIAutomationSpreadsheetItemPatternVtbl struct {
	IUnknownVtbl
	Get_CurrentFormula          uintptr
	GetCurrentAnnotationObjects uintptr
	GetCurrentAnnotationTypes   uintptr
	Get_CachedFormula           uintptr
	GetCachedAnnotationObjects  uintptr
	GetCachedAnnotationTypes    uintptr
}

type IUIAutomationSpreadsheetItemPattern struct {
	IUnknown
}

func (this *IUIAutomationSpreadsheetItemPattern) Vtbl() *IUIAutomationSpreadsheetItemPatternVtbl {
	return (*IUIAutomationSpreadsheetItemPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationSpreadsheetItemPattern) Get_CurrentFormula(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentFormula, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSpreadsheetItemPattern) GetCurrentAnnotationObjects(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentAnnotationObjects, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSpreadsheetItemPattern) GetCurrentAnnotationTypes(retVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentAnnotationTypes, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSpreadsheetItemPattern) Get_CachedFormula(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedFormula, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSpreadsheetItemPattern) GetCachedAnnotationObjects(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedAnnotationObjects, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationSpreadsheetItemPattern) GetCachedAnnotationTypes(retVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedAnnotationTypes, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 6D74D017-6ECB-4381-B38B-3C17A48FF1C2
var IID_IUIAutomationTransformPattern2 = syscall.GUID{0x6D74D017, 0x6ECB, 0x4381,
	[8]byte{0xB3, 0x8B, 0x3C, 0x17, 0xA4, 0x8F, 0xF1, 0xC2}}

type IUIAutomationTransformPattern2Interface interface {
	IUIAutomationTransformPatternInterface
	Zoom(zoomValue float64) HRESULT
	ZoomByUnit(zoomUnit ZoomUnit) HRESULT
	Get_CurrentCanZoom(retVal *BOOL) HRESULT
	Get_CachedCanZoom(retVal *BOOL) HRESULT
	Get_CurrentZoomLevel(retVal *float64) HRESULT
	Get_CachedZoomLevel(retVal *float64) HRESULT
	Get_CurrentZoomMinimum(retVal *float64) HRESULT
	Get_CachedZoomMinimum(retVal *float64) HRESULT
	Get_CurrentZoomMaximum(retVal *float64) HRESULT
	Get_CachedZoomMaximum(retVal *float64) HRESULT
}

type IUIAutomationTransformPattern2Vtbl struct {
	IUIAutomationTransformPatternVtbl
	Zoom                   uintptr
	ZoomByUnit             uintptr
	Get_CurrentCanZoom     uintptr
	Get_CachedCanZoom      uintptr
	Get_CurrentZoomLevel   uintptr
	Get_CachedZoomLevel    uintptr
	Get_CurrentZoomMinimum uintptr
	Get_CachedZoomMinimum  uintptr
	Get_CurrentZoomMaximum uintptr
	Get_CachedZoomMaximum  uintptr
}

type IUIAutomationTransformPattern2 struct {
	IUIAutomationTransformPattern
}

func (this *IUIAutomationTransformPattern2) Vtbl() *IUIAutomationTransformPattern2Vtbl {
	return (*IUIAutomationTransformPattern2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTransformPattern2) Zoom(zoomValue float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Zoom, uintptr(unsafe.Pointer(this)), uintptr(zoomValue))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern2) ZoomByUnit(zoomUnit ZoomUnit) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ZoomByUnit, uintptr(unsafe.Pointer(this)), uintptr(zoomUnit))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern2) Get_CurrentCanZoom(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentCanZoom, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern2) Get_CachedCanZoom(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedCanZoom, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern2) Get_CurrentZoomLevel(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentZoomLevel, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern2) Get_CachedZoomLevel(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedZoomLevel, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern2) Get_CurrentZoomMinimum(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentZoomMinimum, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern2) Get_CachedZoomMinimum(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedZoomMinimum, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern2) Get_CurrentZoomMaximum(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentZoomMaximum, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationTransformPattern2) Get_CachedZoomMaximum(retVal *float64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedZoomMaximum, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 6552B038-AE05-40C8-ABFD-AA08352AAB86
var IID_IUIAutomationTextChildPattern = syscall.GUID{0x6552B038, 0xAE05, 0x40C8,
	[8]byte{0xAB, 0xFD, 0xAA, 0x08, 0x35, 0x2A, 0xAB, 0x86}}

type IUIAutomationTextChildPatternInterface interface {
	IUnknownInterface
	Get_TextContainer(container **IUIAutomationElement) HRESULT
	Get_TextRange(range_ **IUIAutomationTextRange) HRESULT
}

type IUIAutomationTextChildPatternVtbl struct {
	IUnknownVtbl
	Get_TextContainer uintptr
	Get_TextRange     uintptr
}

type IUIAutomationTextChildPattern struct {
	IUnknown
}

func (this *IUIAutomationTextChildPattern) Vtbl() *IUIAutomationTextChildPatternVtbl {
	return (*IUIAutomationTextChildPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationTextChildPattern) Get_TextContainer(container **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_TextContainer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(container)))
	return HRESULT(ret)
}

func (this *IUIAutomationTextChildPattern) Get_TextRange(range_ **IUIAutomationTextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_TextRange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(range_)))
	return HRESULT(ret)
}

// 1DC7B570-1F54-4BAD-BCDA-D36A722FB7BD
var IID_IUIAutomationDragPattern = syscall.GUID{0x1DC7B570, 0x1F54, 0x4BAD,
	[8]byte{0xBC, 0xDA, 0xD3, 0x6A, 0x72, 0x2F, 0xB7, 0xBD}}

type IUIAutomationDragPatternInterface interface {
	IUnknownInterface
	Get_CurrentIsGrabbed(retVal *BOOL) HRESULT
	Get_CachedIsGrabbed(retVal *BOOL) HRESULT
	Get_CurrentDropEffect(retVal *BSTR) HRESULT
	Get_CachedDropEffect(retVal *BSTR) HRESULT
	Get_CurrentDropEffects(retVal **SAFEARRAY) HRESULT
	Get_CachedDropEffects(retVal **SAFEARRAY) HRESULT
	GetCurrentGrabbedItems(retVal **IUIAutomationElementArray) HRESULT
	GetCachedGrabbedItems(retVal **IUIAutomationElementArray) HRESULT
}

type IUIAutomationDragPatternVtbl struct {
	IUnknownVtbl
	Get_CurrentIsGrabbed   uintptr
	Get_CachedIsGrabbed    uintptr
	Get_CurrentDropEffect  uintptr
	Get_CachedDropEffect   uintptr
	Get_CurrentDropEffects uintptr
	Get_CachedDropEffects  uintptr
	GetCurrentGrabbedItems uintptr
	GetCachedGrabbedItems  uintptr
}

type IUIAutomationDragPattern struct {
	IUnknown
}

func (this *IUIAutomationDragPattern) Vtbl() *IUIAutomationDragPatternVtbl {
	return (*IUIAutomationDragPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationDragPattern) Get_CurrentIsGrabbed(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsGrabbed, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationDragPattern) Get_CachedIsGrabbed(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsGrabbed, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationDragPattern) Get_CurrentDropEffect(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentDropEffect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationDragPattern) Get_CachedDropEffect(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedDropEffect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationDragPattern) Get_CurrentDropEffects(retVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentDropEffects, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationDragPattern) Get_CachedDropEffects(retVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedDropEffects, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationDragPattern) GetCurrentGrabbedItems(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentGrabbedItems, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationDragPattern) GetCachedGrabbedItems(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCachedGrabbedItems, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 69A095F7-EEE4-430E-A46B-FB73B1AE39A5
var IID_IUIAutomationDropTargetPattern = syscall.GUID{0x69A095F7, 0xEEE4, 0x430E,
	[8]byte{0xA4, 0x6B, 0xFB, 0x73, 0xB1, 0xAE, 0x39, 0xA5}}

type IUIAutomationDropTargetPatternInterface interface {
	IUnknownInterface
	Get_CurrentDropTargetEffect(retVal *BSTR) HRESULT
	Get_CachedDropTargetEffect(retVal *BSTR) HRESULT
	Get_CurrentDropTargetEffects(retVal **SAFEARRAY) HRESULT
	Get_CachedDropTargetEffects(retVal **SAFEARRAY) HRESULT
}

type IUIAutomationDropTargetPatternVtbl struct {
	IUnknownVtbl
	Get_CurrentDropTargetEffect  uintptr
	Get_CachedDropTargetEffect   uintptr
	Get_CurrentDropTargetEffects uintptr
	Get_CachedDropTargetEffects  uintptr
}

type IUIAutomationDropTargetPattern struct {
	IUnknown
}

func (this *IUIAutomationDropTargetPattern) Vtbl() *IUIAutomationDropTargetPatternVtbl {
	return (*IUIAutomationDropTargetPatternVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationDropTargetPattern) Get_CurrentDropTargetEffect(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentDropTargetEffect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationDropTargetPattern) Get_CachedDropTargetEffect(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedDropTargetEffect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationDropTargetPattern) Get_CurrentDropTargetEffects(retVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentDropTargetEffects, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationDropTargetPattern) Get_CachedDropTargetEffects(retVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedDropTargetEffects, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 6749C683-F70D-4487-A698-5F79D55290D6
var IID_IUIAutomationElement2 = syscall.GUID{0x6749C683, 0xF70D, 0x4487,
	[8]byte{0xA6, 0x98, 0x5F, 0x79, 0xD5, 0x52, 0x90, 0xD6}}

type IUIAutomationElement2Interface interface {
	IUIAutomationElementInterface
	Get_CurrentOptimizeForVisualContent(retVal *BOOL) HRESULT
	Get_CachedOptimizeForVisualContent(retVal *BOOL) HRESULT
	Get_CurrentLiveSetting(retVal *LiveSetting) HRESULT
	Get_CachedLiveSetting(retVal *LiveSetting) HRESULT
	Get_CurrentFlowsFrom(retVal **IUIAutomationElementArray) HRESULT
	Get_CachedFlowsFrom(retVal **IUIAutomationElementArray) HRESULT
}

type IUIAutomationElement2Vtbl struct {
	IUIAutomationElementVtbl
	Get_CurrentOptimizeForVisualContent uintptr
	Get_CachedOptimizeForVisualContent  uintptr
	Get_CurrentLiveSetting              uintptr
	Get_CachedLiveSetting               uintptr
	Get_CurrentFlowsFrom                uintptr
	Get_CachedFlowsFrom                 uintptr
}

type IUIAutomationElement2 struct {
	IUIAutomationElement
}

func (this *IUIAutomationElement2) Vtbl() *IUIAutomationElement2Vtbl {
	return (*IUIAutomationElement2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationElement2) Get_CurrentOptimizeForVisualContent(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentOptimizeForVisualContent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement2) Get_CachedOptimizeForVisualContent(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedOptimizeForVisualContent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement2) Get_CurrentLiveSetting(retVal *LiveSetting) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentLiveSetting, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement2) Get_CachedLiveSetting(retVal *LiveSetting) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedLiveSetting, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement2) Get_CurrentFlowsFrom(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentFlowsFrom, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement2) Get_CachedFlowsFrom(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedFlowsFrom, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 8471DF34-AEE0-4A01-A7DE-7DB9AF12C296
var IID_IUIAutomationElement3 = syscall.GUID{0x8471DF34, 0xAEE0, 0x4A01,
	[8]byte{0xA7, 0xDE, 0x7D, 0xB9, 0xAF, 0x12, 0xC2, 0x96}}

type IUIAutomationElement3Interface interface {
	IUIAutomationElement2Interface
	ShowContextMenu() HRESULT
	Get_CurrentIsPeripheral(retVal *BOOL) HRESULT
	Get_CachedIsPeripheral(retVal *BOOL) HRESULT
}

type IUIAutomationElement3Vtbl struct {
	IUIAutomationElement2Vtbl
	ShowContextMenu         uintptr
	Get_CurrentIsPeripheral uintptr
	Get_CachedIsPeripheral  uintptr
}

type IUIAutomationElement3 struct {
	IUIAutomationElement2
}

func (this *IUIAutomationElement3) Vtbl() *IUIAutomationElement3Vtbl {
	return (*IUIAutomationElement3Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationElement3) ShowContextMenu() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ShowContextMenu, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement3) Get_CurrentIsPeripheral(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsPeripheral, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement3) Get_CachedIsPeripheral(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsPeripheral, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 3B6E233C-52FB-4063-A4C9-77C075C2A06B
var IID_IUIAutomationElement4 = syscall.GUID{0x3B6E233C, 0x52FB, 0x4063,
	[8]byte{0xA4, 0xC9, 0x77, 0xC0, 0x75, 0xC2, 0xA0, 0x6B}}

type IUIAutomationElement4Interface interface {
	IUIAutomationElement3Interface
	Get_CurrentPositionInSet(retVal *int32) HRESULT
	Get_CurrentSizeOfSet(retVal *int32) HRESULT
	Get_CurrentLevel(retVal *int32) HRESULT
	Get_CurrentAnnotationTypes(retVal **SAFEARRAY) HRESULT
	Get_CurrentAnnotationObjects(retVal **IUIAutomationElementArray) HRESULT
	Get_CachedPositionInSet(retVal *int32) HRESULT
	Get_CachedSizeOfSet(retVal *int32) HRESULT
	Get_CachedLevel(retVal *int32) HRESULT
	Get_CachedAnnotationTypes(retVal **SAFEARRAY) HRESULT
	Get_CachedAnnotationObjects(retVal **IUIAutomationElementArray) HRESULT
}

type IUIAutomationElement4Vtbl struct {
	IUIAutomationElement3Vtbl
	Get_CurrentPositionInSet     uintptr
	Get_CurrentSizeOfSet         uintptr
	Get_CurrentLevel             uintptr
	Get_CurrentAnnotationTypes   uintptr
	Get_CurrentAnnotationObjects uintptr
	Get_CachedPositionInSet      uintptr
	Get_CachedSizeOfSet          uintptr
	Get_CachedLevel              uintptr
	Get_CachedAnnotationTypes    uintptr
	Get_CachedAnnotationObjects  uintptr
}

type IUIAutomationElement4 struct {
	IUIAutomationElement3
}

func (this *IUIAutomationElement4) Vtbl() *IUIAutomationElement4Vtbl {
	return (*IUIAutomationElement4Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationElement4) Get_CurrentPositionInSet(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentPositionInSet, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement4) Get_CurrentSizeOfSet(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentSizeOfSet, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement4) Get_CurrentLevel(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentLevel, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement4) Get_CurrentAnnotationTypes(retVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentAnnotationTypes, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement4) Get_CurrentAnnotationObjects(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentAnnotationObjects, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement4) Get_CachedPositionInSet(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedPositionInSet, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement4) Get_CachedSizeOfSet(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedSizeOfSet, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement4) Get_CachedLevel(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedLevel, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement4) Get_CachedAnnotationTypes(retVal **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedAnnotationTypes, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement4) Get_CachedAnnotationObjects(retVal **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedAnnotationObjects, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 98141C1D-0D0E-4175-BBE2-6BFF455842A7
var IID_IUIAutomationElement5 = syscall.GUID{0x98141C1D, 0x0D0E, 0x4175,
	[8]byte{0xBB, 0xE2, 0x6B, 0xFF, 0x45, 0x58, 0x42, 0xA7}}

type IUIAutomationElement5Interface interface {
	IUIAutomationElement4Interface
	Get_CurrentLandmarkType(retVal *int32) HRESULT
	Get_CurrentLocalizedLandmarkType(retVal *BSTR) HRESULT
	Get_CachedLandmarkType(retVal *int32) HRESULT
	Get_CachedLocalizedLandmarkType(retVal *BSTR) HRESULT
}

type IUIAutomationElement5Vtbl struct {
	IUIAutomationElement4Vtbl
	Get_CurrentLandmarkType          uintptr
	Get_CurrentLocalizedLandmarkType uintptr
	Get_CachedLandmarkType           uintptr
	Get_CachedLocalizedLandmarkType  uintptr
}

type IUIAutomationElement5 struct {
	IUIAutomationElement4
}

func (this *IUIAutomationElement5) Vtbl() *IUIAutomationElement5Vtbl {
	return (*IUIAutomationElement5Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationElement5) Get_CurrentLandmarkType(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentLandmarkType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement5) Get_CurrentLocalizedLandmarkType(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentLocalizedLandmarkType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement5) Get_CachedLandmarkType(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedLandmarkType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement5) Get_CachedLocalizedLandmarkType(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedLocalizedLandmarkType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 4780D450-8BCA-4977-AFA5-A4A517F555E3
var IID_IUIAutomationElement6 = syscall.GUID{0x4780D450, 0x8BCA, 0x4977,
	[8]byte{0xAF, 0xA5, 0xA4, 0xA5, 0x17, 0xF5, 0x55, 0xE3}}

type IUIAutomationElement6Interface interface {
	IUIAutomationElement5Interface
	Get_CurrentFullDescription(retVal *BSTR) HRESULT
	Get_CachedFullDescription(retVal *BSTR) HRESULT
}

type IUIAutomationElement6Vtbl struct {
	IUIAutomationElement5Vtbl
	Get_CurrentFullDescription uintptr
	Get_CachedFullDescription  uintptr
}

type IUIAutomationElement6 struct {
	IUIAutomationElement5
}

func (this *IUIAutomationElement6) Vtbl() *IUIAutomationElement6Vtbl {
	return (*IUIAutomationElement6Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationElement6) Get_CurrentFullDescription(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentFullDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement6) Get_CachedFullDescription(retVal *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedFullDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 204E8572-CFC3-4C11-B0C8-7DA7420750B7
var IID_IUIAutomationElement7 = syscall.GUID{0x204E8572, 0xCFC3, 0x4C11,
	[8]byte{0xB0, 0xC8, 0x7D, 0xA7, 0x42, 0x07, 0x50, 0xB7}}

type IUIAutomationElement7Interface interface {
	IUIAutomationElement6Interface
	FindFirstWithOptions(scope TreeScope, condition *IUIAutomationCondition, traversalOptions TreeTraversalOptions, root *IUIAutomationElement, found **IUIAutomationElement) HRESULT
	FindAllWithOptions(scope TreeScope, condition *IUIAutomationCondition, traversalOptions TreeTraversalOptions, root *IUIAutomationElement, found **IUIAutomationElementArray) HRESULT
	FindFirstWithOptionsBuildCache(scope TreeScope, condition *IUIAutomationCondition, cacheRequest *IUIAutomationCacheRequest, traversalOptions TreeTraversalOptions, root *IUIAutomationElement, found **IUIAutomationElement) HRESULT
	FindAllWithOptionsBuildCache(scope TreeScope, condition *IUIAutomationCondition, cacheRequest *IUIAutomationCacheRequest, traversalOptions TreeTraversalOptions, root *IUIAutomationElement, found **IUIAutomationElementArray) HRESULT
	GetCurrentMetadataValue(targetId int32, metadataId int32, returnVal *VARIANT) HRESULT
}

type IUIAutomationElement7Vtbl struct {
	IUIAutomationElement6Vtbl
	FindFirstWithOptions           uintptr
	FindAllWithOptions             uintptr
	FindFirstWithOptionsBuildCache uintptr
	FindAllWithOptionsBuildCache   uintptr
	GetCurrentMetadataValue        uintptr
}

type IUIAutomationElement7 struct {
	IUIAutomationElement6
}

func (this *IUIAutomationElement7) Vtbl() *IUIAutomationElement7Vtbl {
	return (*IUIAutomationElement7Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationElement7) FindFirstWithOptions(scope TreeScope, condition *IUIAutomationCondition, traversalOptions TreeTraversalOptions, root *IUIAutomationElement, found **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindFirstWithOptions, uintptr(unsafe.Pointer(this)), uintptr(scope), uintptr(unsafe.Pointer(condition)), uintptr(traversalOptions), uintptr(unsafe.Pointer(root)), uintptr(unsafe.Pointer(found)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement7) FindAllWithOptions(scope TreeScope, condition *IUIAutomationCondition, traversalOptions TreeTraversalOptions, root *IUIAutomationElement, found **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindAllWithOptions, uintptr(unsafe.Pointer(this)), uintptr(scope), uintptr(unsafe.Pointer(condition)), uintptr(traversalOptions), uintptr(unsafe.Pointer(root)), uintptr(unsafe.Pointer(found)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement7) FindFirstWithOptionsBuildCache(scope TreeScope, condition *IUIAutomationCondition, cacheRequest *IUIAutomationCacheRequest, traversalOptions TreeTraversalOptions, root *IUIAutomationElement, found **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindFirstWithOptionsBuildCache, uintptr(unsafe.Pointer(this)), uintptr(scope), uintptr(unsafe.Pointer(condition)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(traversalOptions), uintptr(unsafe.Pointer(root)), uintptr(unsafe.Pointer(found)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement7) FindAllWithOptionsBuildCache(scope TreeScope, condition *IUIAutomationCondition, cacheRequest *IUIAutomationCacheRequest, traversalOptions TreeTraversalOptions, root *IUIAutomationElement, found **IUIAutomationElementArray) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindAllWithOptionsBuildCache, uintptr(unsafe.Pointer(this)), uintptr(scope), uintptr(unsafe.Pointer(condition)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(traversalOptions), uintptr(unsafe.Pointer(root)), uintptr(unsafe.Pointer(found)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement7) GetCurrentMetadataValue(targetId int32, metadataId int32, returnVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentMetadataValue, uintptr(unsafe.Pointer(this)), uintptr(targetId), uintptr(metadataId), uintptr(unsafe.Pointer(returnVal)))
	return HRESULT(ret)
}

// 8C60217D-5411-4CDE-BCC0-1CEDA223830C
var IID_IUIAutomationElement8 = syscall.GUID{0x8C60217D, 0x5411, 0x4CDE,
	[8]byte{0xBC, 0xC0, 0x1C, 0xED, 0xA2, 0x23, 0x83, 0x0C}}

type IUIAutomationElement8Interface interface {
	IUIAutomationElement7Interface
	Get_CurrentHeadingLevel(retVal *int32) HRESULT
	Get_CachedHeadingLevel(retVal *int32) HRESULT
}

type IUIAutomationElement8Vtbl struct {
	IUIAutomationElement7Vtbl
	Get_CurrentHeadingLevel uintptr
	Get_CachedHeadingLevel  uintptr
}

type IUIAutomationElement8 struct {
	IUIAutomationElement7
}

func (this *IUIAutomationElement8) Vtbl() *IUIAutomationElement8Vtbl {
	return (*IUIAutomationElement8Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationElement8) Get_CurrentHeadingLevel(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentHeadingLevel, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement8) Get_CachedHeadingLevel(retVal *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedHeadingLevel, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 39325FAC-039D-440E-A3A3-5EB81A5CECC3
var IID_IUIAutomationElement9 = syscall.GUID{0x39325FAC, 0x039D, 0x440E,
	[8]byte{0xA3, 0xA3, 0x5E, 0xB8, 0x1A, 0x5C, 0xEC, 0xC3}}

type IUIAutomationElement9Interface interface {
	IUIAutomationElement8Interface
	Get_CurrentIsDialog(retVal *BOOL) HRESULT
	Get_CachedIsDialog(retVal *BOOL) HRESULT
}

type IUIAutomationElement9Vtbl struct {
	IUIAutomationElement8Vtbl
	Get_CurrentIsDialog uintptr
	Get_CachedIsDialog  uintptr
}

type IUIAutomationElement9 struct {
	IUIAutomationElement8
}

func (this *IUIAutomationElement9) Vtbl() *IUIAutomationElement9Vtbl {
	return (*IUIAutomationElement9Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationElement9) Get_CurrentIsDialog(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurrentIsDialog, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

func (this *IUIAutomationElement9) Get_CachedIsDialog(retVal *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CachedIsDialog, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(retVal)))
	return HRESULT(ret)
}

// 85B94ECD-849D-42B6-B94D-D6DB23FDF5A4
var IID_IUIAutomationProxyFactory = syscall.GUID{0x85B94ECD, 0x849D, 0x42B6,
	[8]byte{0xB9, 0x4D, 0xD6, 0xDB, 0x23, 0xFD, 0xF5, 0xA4}}

type IUIAutomationProxyFactoryInterface interface {
	IUnknownInterface
	CreateProvider(hwnd HWND, idObject int32, idChild int32, provider **IRawElementProviderSimple) HRESULT
	Get_ProxyFactoryId(factoryId *BSTR) HRESULT
}

type IUIAutomationProxyFactoryVtbl struct {
	IUnknownVtbl
	CreateProvider     uintptr
	Get_ProxyFactoryId uintptr
}

type IUIAutomationProxyFactory struct {
	IUnknown
}

func (this *IUIAutomationProxyFactory) Vtbl() *IUIAutomationProxyFactoryVtbl {
	return (*IUIAutomationProxyFactoryVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationProxyFactory) CreateProvider(hwnd HWND, idObject int32, idChild int32, provider **IRawElementProviderSimple) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateProvider, uintptr(unsafe.Pointer(this)), hwnd, uintptr(idObject), uintptr(idChild), uintptr(unsafe.Pointer(provider)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactory) Get_ProxyFactoryId(factoryId *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ProxyFactoryId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(factoryId)))
	return HRESULT(ret)
}

// D50E472E-B64B-490C-BCA1-D30696F9F289
var IID_IUIAutomationProxyFactoryEntry = syscall.GUID{0xD50E472E, 0xB64B, 0x490C,
	[8]byte{0xBC, 0xA1, 0xD3, 0x06, 0x96, 0xF9, 0xF2, 0x89}}

type IUIAutomationProxyFactoryEntryInterface interface {
	IUnknownInterface
	Get_ProxyFactory(factory **IUIAutomationProxyFactory) HRESULT
	Get_ClassName(className *BSTR) HRESULT
	Get_ImageName(imageName *BSTR) HRESULT
	Get_AllowSubstringMatch(allowSubstringMatch *BOOL) HRESULT
	Get_CanCheckBaseClass(canCheckBaseClass *BOOL) HRESULT
	Get_NeedsAdviseEvents(adviseEvents *BOOL) HRESULT
	Put_ClassName(className PWSTR) HRESULT
	Put_ImageName(imageName PWSTR) HRESULT
	Put_AllowSubstringMatch(allowSubstringMatch BOOL) HRESULT
	Put_CanCheckBaseClass(canCheckBaseClass BOOL) HRESULT
	Put_NeedsAdviseEvents(adviseEvents BOOL) HRESULT
	SetWinEventsForAutomationEvent(eventId int32, propertyId int32, winEvents *SAFEARRAY) HRESULT
	GetWinEventsForAutomationEvent(eventId int32, propertyId int32, winEvents **SAFEARRAY) HRESULT
}

type IUIAutomationProxyFactoryEntryVtbl struct {
	IUnknownVtbl
	Get_ProxyFactory               uintptr
	Get_ClassName                  uintptr
	Get_ImageName                  uintptr
	Get_AllowSubstringMatch        uintptr
	Get_CanCheckBaseClass          uintptr
	Get_NeedsAdviseEvents          uintptr
	Put_ClassName                  uintptr
	Put_ImageName                  uintptr
	Put_AllowSubstringMatch        uintptr
	Put_CanCheckBaseClass          uintptr
	Put_NeedsAdviseEvents          uintptr
	SetWinEventsForAutomationEvent uintptr
	GetWinEventsForAutomationEvent uintptr
}

type IUIAutomationProxyFactoryEntry struct {
	IUnknown
}

func (this *IUIAutomationProxyFactoryEntry) Vtbl() *IUIAutomationProxyFactoryEntryVtbl {
	return (*IUIAutomationProxyFactoryEntryVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationProxyFactoryEntry) Get_ProxyFactory(factory **IUIAutomationProxyFactory) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ProxyFactory, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(factory)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryEntry) Get_ClassName(className *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ClassName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(className)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryEntry) Get_ImageName(imageName *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ImageName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(imageName)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryEntry) Get_AllowSubstringMatch(allowSubstringMatch *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_AllowSubstringMatch, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(allowSubstringMatch)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryEntry) Get_CanCheckBaseClass(canCheckBaseClass *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CanCheckBaseClass, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(canCheckBaseClass)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryEntry) Get_NeedsAdviseEvents(adviseEvents *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_NeedsAdviseEvents, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(adviseEvents)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryEntry) Put_ClassName(className PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_ClassName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(className)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryEntry) Put_ImageName(imageName PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_ImageName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(imageName)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryEntry) Put_AllowSubstringMatch(allowSubstringMatch BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_AllowSubstringMatch, uintptr(unsafe.Pointer(this)), uintptr(allowSubstringMatch))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryEntry) Put_CanCheckBaseClass(canCheckBaseClass BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_CanCheckBaseClass, uintptr(unsafe.Pointer(this)), uintptr(canCheckBaseClass))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryEntry) Put_NeedsAdviseEvents(adviseEvents BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_NeedsAdviseEvents, uintptr(unsafe.Pointer(this)), uintptr(adviseEvents))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryEntry) SetWinEventsForAutomationEvent(eventId int32, propertyId int32, winEvents *SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetWinEventsForAutomationEvent, uintptr(unsafe.Pointer(this)), uintptr(eventId), uintptr(propertyId), uintptr(unsafe.Pointer(winEvents)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryEntry) GetWinEventsForAutomationEvent(eventId int32, propertyId int32, winEvents **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetWinEventsForAutomationEvent, uintptr(unsafe.Pointer(this)), uintptr(eventId), uintptr(propertyId), uintptr(unsafe.Pointer(winEvents)))
	return HRESULT(ret)
}

// 09E31E18-872D-4873-93D1-1E541EC133FD
var IID_IUIAutomationProxyFactoryMapping = syscall.GUID{0x09E31E18, 0x872D, 0x4873,
	[8]byte{0x93, 0xD1, 0x1E, 0x54, 0x1E, 0xC1, 0x33, 0xFD}}

type IUIAutomationProxyFactoryMappingInterface interface {
	IUnknownInterface
	Get_Count(count *uint32) HRESULT
	GetTable(table **SAFEARRAY) HRESULT
	GetEntry(index uint32, entry **IUIAutomationProxyFactoryEntry) HRESULT
	SetTable(factoryList *SAFEARRAY) HRESULT
	InsertEntries(before uint32, factoryList *SAFEARRAY) HRESULT
	InsertEntry(before uint32, factory *IUIAutomationProxyFactoryEntry) HRESULT
	RemoveEntry(index uint32) HRESULT
	ClearTable() HRESULT
	RestoreDefaultTable() HRESULT
}

type IUIAutomationProxyFactoryMappingVtbl struct {
	IUnknownVtbl
	Get_Count           uintptr
	GetTable            uintptr
	GetEntry            uintptr
	SetTable            uintptr
	InsertEntries       uintptr
	InsertEntry         uintptr
	RemoveEntry         uintptr
	ClearTable          uintptr
	RestoreDefaultTable uintptr
}

type IUIAutomationProxyFactoryMapping struct {
	IUnknown
}

func (this *IUIAutomationProxyFactoryMapping) Vtbl() *IUIAutomationProxyFactoryMappingVtbl {
	return (*IUIAutomationProxyFactoryMappingVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationProxyFactoryMapping) Get_Count(count *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Count, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(count)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryMapping) GetTable(table **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(table)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryMapping) GetEntry(index uint32, entry **IUIAutomationProxyFactoryEntry) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEntry, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(entry)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryMapping) SetTable(factoryList *SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetTable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(factoryList)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryMapping) InsertEntries(before uint32, factoryList *SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InsertEntries, uintptr(unsafe.Pointer(this)), uintptr(before), uintptr(unsafe.Pointer(factoryList)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryMapping) InsertEntry(before uint32, factory *IUIAutomationProxyFactoryEntry) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InsertEntry, uintptr(unsafe.Pointer(this)), uintptr(before), uintptr(unsafe.Pointer(factory)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryMapping) RemoveEntry(index uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveEntry, uintptr(unsafe.Pointer(this)), uintptr(index))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryMapping) ClearTable() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ClearTable, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomationProxyFactoryMapping) RestoreDefaultTable() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RestoreDefaultTable, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// C9EE12F2-C13B-4408-997C-639914377F4E
var IID_IUIAutomationEventHandlerGroup = syscall.GUID{0xC9EE12F2, 0xC13B, 0x4408,
	[8]byte{0x99, 0x7C, 0x63, 0x99, 0x14, 0x37, 0x7F, 0x4E}}

type IUIAutomationEventHandlerGroupInterface interface {
	IUnknownInterface
	AddActiveTextPositionChangedEventHandler(scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationActiveTextPositionChangedEventHandler) HRESULT
	AddAutomationEventHandler(eventId int32, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationEventHandler) HRESULT
	AddChangesEventHandler(scope TreeScope, changeTypes *int32, changesCount int32, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationChangesEventHandler) HRESULT
	AddNotificationEventHandler(scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationNotificationEventHandler) HRESULT
	AddPropertyChangedEventHandler(scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationPropertyChangedEventHandler, propertyArray *int32, propertyCount int32) HRESULT
	AddStructureChangedEventHandler(scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationStructureChangedEventHandler) HRESULT
	AddTextEditTextChangedEventHandler(scope TreeScope, textEditChangeType TextEditChangeType, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationTextEditTextChangedEventHandler) HRESULT
}

type IUIAutomationEventHandlerGroupVtbl struct {
	IUnknownVtbl
	AddActiveTextPositionChangedEventHandler uintptr
	AddAutomationEventHandler                uintptr
	AddChangesEventHandler                   uintptr
	AddNotificationEventHandler              uintptr
	AddPropertyChangedEventHandler           uintptr
	AddStructureChangedEventHandler          uintptr
	AddTextEditTextChangedEventHandler       uintptr
}

type IUIAutomationEventHandlerGroup struct {
	IUnknown
}

func (this *IUIAutomationEventHandlerGroup) Vtbl() *IUIAutomationEventHandlerGroupVtbl {
	return (*IUIAutomationEventHandlerGroupVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomationEventHandlerGroup) AddActiveTextPositionChangedEventHandler(scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationActiveTextPositionChangedEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddActiveTextPositionChangedEventHandler, uintptr(unsafe.Pointer(this)), uintptr(scope), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomationEventHandlerGroup) AddAutomationEventHandler(eventId int32, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddAutomationEventHandler, uintptr(unsafe.Pointer(this)), uintptr(eventId), uintptr(scope), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomationEventHandlerGroup) AddChangesEventHandler(scope TreeScope, changeTypes *int32, changesCount int32, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationChangesEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddChangesEventHandler, uintptr(unsafe.Pointer(this)), uintptr(scope), uintptr(unsafe.Pointer(changeTypes)), uintptr(changesCount), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomationEventHandlerGroup) AddNotificationEventHandler(scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationNotificationEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddNotificationEventHandler, uintptr(unsafe.Pointer(this)), uintptr(scope), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomationEventHandlerGroup) AddPropertyChangedEventHandler(scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationPropertyChangedEventHandler, propertyArray *int32, propertyCount int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddPropertyChangedEventHandler, uintptr(unsafe.Pointer(this)), uintptr(scope), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)), uintptr(unsafe.Pointer(propertyArray)), uintptr(propertyCount))
	return HRESULT(ret)
}

func (this *IUIAutomationEventHandlerGroup) AddStructureChangedEventHandler(scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationStructureChangedEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddStructureChangedEventHandler, uintptr(unsafe.Pointer(this)), uintptr(scope), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomationEventHandlerGroup) AddTextEditTextChangedEventHandler(scope TreeScope, textEditChangeType TextEditChangeType, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationTextEditTextChangedEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddTextEditTextChangedEventHandler, uintptr(unsafe.Pointer(this)), uintptr(scope), uintptr(textEditChangeType), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

// 30CBE57D-D9D0-452A-AB13-7AC5AC4825EE
var IID_IUIAutomation = syscall.GUID{0x30CBE57D, 0xD9D0, 0x452A,
	[8]byte{0xAB, 0x13, 0x7A, 0xC5, 0xAC, 0x48, 0x25, 0xEE}}

type IUIAutomationInterface interface {
	IUnknownInterface
	CompareElements(el1 *IUIAutomationElement, el2 *IUIAutomationElement, areSame *BOOL) HRESULT
	CompareRuntimeIds(runtimeId1 *SAFEARRAY, runtimeId2 *SAFEARRAY, areSame *BOOL) HRESULT
	GetRootElement(root **IUIAutomationElement) HRESULT
	ElementFromHandle(hwnd HWND, element **IUIAutomationElement) HRESULT
	ElementFromPoint(pt POINT, element **IUIAutomationElement) HRESULT
	GetFocusedElement(element **IUIAutomationElement) HRESULT
	GetRootElementBuildCache(cacheRequest *IUIAutomationCacheRequest, root **IUIAutomationElement) HRESULT
	ElementFromHandleBuildCache(hwnd HWND, cacheRequest *IUIAutomationCacheRequest, element **IUIAutomationElement) HRESULT
	ElementFromPointBuildCache(pt POINT, cacheRequest *IUIAutomationCacheRequest, element **IUIAutomationElement) HRESULT
	GetFocusedElementBuildCache(cacheRequest *IUIAutomationCacheRequest, element **IUIAutomationElement) HRESULT
	CreateTreeWalker(pCondition *IUIAutomationCondition, walker **IUIAutomationTreeWalker) HRESULT
	Get_ControlViewWalker(walker **IUIAutomationTreeWalker) HRESULT
	Get_ContentViewWalker(walker **IUIAutomationTreeWalker) HRESULT
	Get_RawViewWalker(walker **IUIAutomationTreeWalker) HRESULT
	Get_RawViewCondition(condition **IUIAutomationCondition) HRESULT
	Get_ControlViewCondition(condition **IUIAutomationCondition) HRESULT
	Get_ContentViewCondition(condition **IUIAutomationCondition) HRESULT
	CreateCacheRequest(cacheRequest **IUIAutomationCacheRequest) HRESULT
	CreateTrueCondition(newCondition **IUIAutomationCondition) HRESULT
	CreateFalseCondition(newCondition **IUIAutomationCondition) HRESULT
	CreatePropertyCondition(propertyId int32, value VARIANT, newCondition **IUIAutomationCondition) HRESULT
	CreatePropertyConditionEx(propertyId int32, value VARIANT, flags PropertyConditionFlags, newCondition **IUIAutomationCondition) HRESULT
	CreateAndCondition(condition1 *IUIAutomationCondition, condition2 *IUIAutomationCondition, newCondition **IUIAutomationCondition) HRESULT
	CreateAndConditionFromArray(conditions *SAFEARRAY, newCondition **IUIAutomationCondition) HRESULT
	CreateAndConditionFromNativeArray(conditions **IUIAutomationCondition, conditionCount int32, newCondition **IUIAutomationCondition) HRESULT
	CreateOrCondition(condition1 *IUIAutomationCondition, condition2 *IUIAutomationCondition, newCondition **IUIAutomationCondition) HRESULT
	CreateOrConditionFromArray(conditions *SAFEARRAY, newCondition **IUIAutomationCondition) HRESULT
	CreateOrConditionFromNativeArray(conditions **IUIAutomationCondition, conditionCount int32, newCondition **IUIAutomationCondition) HRESULT
	CreateNotCondition(condition *IUIAutomationCondition, newCondition **IUIAutomationCondition) HRESULT
	AddAutomationEventHandler(eventId int32, element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationEventHandler) HRESULT
	RemoveAutomationEventHandler(eventId int32, element *IUIAutomationElement, handler *IUIAutomationEventHandler) HRESULT
	AddPropertyChangedEventHandlerNativeArray(element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationPropertyChangedEventHandler, propertyArray *int32, propertyCount int32) HRESULT
	AddPropertyChangedEventHandler(element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationPropertyChangedEventHandler, propertyArray *SAFEARRAY) HRESULT
	RemovePropertyChangedEventHandler(element *IUIAutomationElement, handler *IUIAutomationPropertyChangedEventHandler) HRESULT
	AddStructureChangedEventHandler(element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationStructureChangedEventHandler) HRESULT
	RemoveStructureChangedEventHandler(element *IUIAutomationElement, handler *IUIAutomationStructureChangedEventHandler) HRESULT
	AddFocusChangedEventHandler(cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationFocusChangedEventHandler) HRESULT
	RemoveFocusChangedEventHandler(handler *IUIAutomationFocusChangedEventHandler) HRESULT
	RemoveAllEventHandlers() HRESULT
	IntNativeArrayToSafeArray(array *int32, arrayCount int32, safeArray **SAFEARRAY) HRESULT
	IntSafeArrayToNativeArray(intArray *SAFEARRAY, array **int32, arrayCount *int32) HRESULT
	RectToVariant(rc RECT, var_ *VARIANT) HRESULT
	VariantToRect(var_ VARIANT, rc *RECT) HRESULT
	SafeArrayToRectNativeArray(rects *SAFEARRAY, rectArray **RECT, rectArrayCount *int32) HRESULT
	CreateProxyFactoryEntry(factory *IUIAutomationProxyFactory, factoryEntry **IUIAutomationProxyFactoryEntry) HRESULT
	Get_ProxyFactoryMapping(factoryMapping **IUIAutomationProxyFactoryMapping) HRESULT
	GetPropertyProgrammaticName(property int32, name *BSTR) HRESULT
	GetPatternProgrammaticName(pattern int32, name *BSTR) HRESULT
	PollForPotentialSupportedPatterns(pElement *IUIAutomationElement, patternIds **SAFEARRAY, patternNames **SAFEARRAY) HRESULT
	PollForPotentialSupportedProperties(pElement *IUIAutomationElement, propertyIds **SAFEARRAY, propertyNames **SAFEARRAY) HRESULT
	CheckNotSupported(value VARIANT, isNotSupported *BOOL) HRESULT
	Get_ReservedNotSupportedValue(notSupportedValue **IUnknown) HRESULT
	Get_ReservedMixedAttributeValue(mixedAttributeValue **IUnknown) HRESULT
	ElementFromIAccessible(accessible *IAccessible, childId int32, element **IUIAutomationElement) HRESULT
	ElementFromIAccessibleBuildCache(accessible *IAccessible, childId int32, cacheRequest *IUIAutomationCacheRequest, element **IUIAutomationElement) HRESULT
}

type IUIAutomationVtbl struct {
	IUnknownVtbl
	CompareElements                           uintptr
	CompareRuntimeIds                         uintptr
	GetRootElement                            uintptr
	ElementFromHandle                         uintptr
	ElementFromPoint                          uintptr
	GetFocusedElement                         uintptr
	GetRootElementBuildCache                  uintptr
	ElementFromHandleBuildCache               uintptr
	ElementFromPointBuildCache                uintptr
	GetFocusedElementBuildCache               uintptr
	CreateTreeWalker                          uintptr
	Get_ControlViewWalker                     uintptr
	Get_ContentViewWalker                     uintptr
	Get_RawViewWalker                         uintptr
	Get_RawViewCondition                      uintptr
	Get_ControlViewCondition                  uintptr
	Get_ContentViewCondition                  uintptr
	CreateCacheRequest                        uintptr
	CreateTrueCondition                       uintptr
	CreateFalseCondition                      uintptr
	CreatePropertyCondition                   uintptr
	CreatePropertyConditionEx                 uintptr
	CreateAndCondition                        uintptr
	CreateAndConditionFromArray               uintptr
	CreateAndConditionFromNativeArray         uintptr
	CreateOrCondition                         uintptr
	CreateOrConditionFromArray                uintptr
	CreateOrConditionFromNativeArray          uintptr
	CreateNotCondition                        uintptr
	AddAutomationEventHandler                 uintptr
	RemoveAutomationEventHandler              uintptr
	AddPropertyChangedEventHandlerNativeArray uintptr
	AddPropertyChangedEventHandler            uintptr
	RemovePropertyChangedEventHandler         uintptr
	AddStructureChangedEventHandler           uintptr
	RemoveStructureChangedEventHandler        uintptr
	AddFocusChangedEventHandler               uintptr
	RemoveFocusChangedEventHandler            uintptr
	RemoveAllEventHandlers                    uintptr
	IntNativeArrayToSafeArray                 uintptr
	IntSafeArrayToNativeArray                 uintptr
	RectToVariant                             uintptr
	VariantToRect                             uintptr
	SafeArrayToRectNativeArray                uintptr
	CreateProxyFactoryEntry                   uintptr
	Get_ProxyFactoryMapping                   uintptr
	GetPropertyProgrammaticName               uintptr
	GetPatternProgrammaticName                uintptr
	PollForPotentialSupportedPatterns         uintptr
	PollForPotentialSupportedProperties       uintptr
	CheckNotSupported                         uintptr
	Get_ReservedNotSupportedValue             uintptr
	Get_ReservedMixedAttributeValue           uintptr
	ElementFromIAccessible                    uintptr
	ElementFromIAccessibleBuildCache          uintptr
}

type IUIAutomation struct {
	IUnknown
}

func (this *IUIAutomation) Vtbl() *IUIAutomationVtbl {
	return (*IUIAutomationVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomation) CompareElements(el1 *IUIAutomationElement, el2 *IUIAutomationElement, areSame *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CompareElements, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(el1)), uintptr(unsafe.Pointer(el2)), uintptr(unsafe.Pointer(areSame)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CompareRuntimeIds(runtimeId1 *SAFEARRAY, runtimeId2 *SAFEARRAY, areSame *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CompareRuntimeIds, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(runtimeId1)), uintptr(unsafe.Pointer(runtimeId2)), uintptr(unsafe.Pointer(areSame)))
	return HRESULT(ret)
}

func (this *IUIAutomation) GetRootElement(root **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRootElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(root)))
	return HRESULT(ret)
}

func (this *IUIAutomation) ElementFromHandle(hwnd HWND, element **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ElementFromHandle, uintptr(unsafe.Pointer(this)), hwnd, uintptr(unsafe.Pointer(element)))
	return HRESULT(ret)
}

func (this *IUIAutomation) ElementFromPoint(pt POINT, element **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ElementFromPoint, uintptr(unsafe.Pointer(this)), *(*uintptr)(unsafe.Pointer(&pt)), uintptr(unsafe.Pointer(element)))
	return HRESULT(ret)
}

func (this *IUIAutomation) GetFocusedElement(element **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFocusedElement, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)))
	return HRESULT(ret)
}

func (this *IUIAutomation) GetRootElementBuildCache(cacheRequest *IUIAutomationCacheRequest, root **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRootElementBuildCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(root)))
	return HRESULT(ret)
}

func (this *IUIAutomation) ElementFromHandleBuildCache(hwnd HWND, cacheRequest *IUIAutomationCacheRequest, element **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ElementFromHandleBuildCache, uintptr(unsafe.Pointer(this)), hwnd, uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(element)))
	return HRESULT(ret)
}

func (this *IUIAutomation) ElementFromPointBuildCache(pt POINT, cacheRequest *IUIAutomationCacheRequest, element **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ElementFromPointBuildCache, uintptr(unsafe.Pointer(this)), *(*uintptr)(unsafe.Pointer(&pt)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(element)))
	return HRESULT(ret)
}

func (this *IUIAutomation) GetFocusedElementBuildCache(cacheRequest *IUIAutomationCacheRequest, element **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFocusedElementBuildCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(element)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CreateTreeWalker(pCondition *IUIAutomationCondition, walker **IUIAutomationTreeWalker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateTreeWalker, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCondition)), uintptr(unsafe.Pointer(walker)))
	return HRESULT(ret)
}

func (this *IUIAutomation) Get_ControlViewWalker(walker **IUIAutomationTreeWalker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ControlViewWalker, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(walker)))
	return HRESULT(ret)
}

func (this *IUIAutomation) Get_ContentViewWalker(walker **IUIAutomationTreeWalker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ContentViewWalker, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(walker)))
	return HRESULT(ret)
}

func (this *IUIAutomation) Get_RawViewWalker(walker **IUIAutomationTreeWalker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_RawViewWalker, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(walker)))
	return HRESULT(ret)
}

func (this *IUIAutomation) Get_RawViewCondition(condition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_RawViewCondition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(condition)))
	return HRESULT(ret)
}

func (this *IUIAutomation) Get_ControlViewCondition(condition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ControlViewCondition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(condition)))
	return HRESULT(ret)
}

func (this *IUIAutomation) Get_ContentViewCondition(condition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ContentViewCondition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(condition)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CreateCacheRequest(cacheRequest **IUIAutomationCacheRequest) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateCacheRequest, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(cacheRequest)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CreateTrueCondition(newCondition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateTrueCondition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(newCondition)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CreateFalseCondition(newCondition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateFalseCondition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(newCondition)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CreatePropertyCondition(propertyId int32, value VARIANT, newCondition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreatePropertyCondition, uintptr(unsafe.Pointer(this)), uintptr(propertyId), uintptr(unsafe.Pointer(&value)), uintptr(unsafe.Pointer(newCondition)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CreatePropertyConditionEx(propertyId int32, value VARIANT, flags PropertyConditionFlags, newCondition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreatePropertyConditionEx, uintptr(unsafe.Pointer(this)), uintptr(propertyId), uintptr(unsafe.Pointer(&value)), uintptr(flags), uintptr(unsafe.Pointer(newCondition)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CreateAndCondition(condition1 *IUIAutomationCondition, condition2 *IUIAutomationCondition, newCondition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateAndCondition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(condition1)), uintptr(unsafe.Pointer(condition2)), uintptr(unsafe.Pointer(newCondition)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CreateAndConditionFromArray(conditions *SAFEARRAY, newCondition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateAndConditionFromArray, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(conditions)), uintptr(unsafe.Pointer(newCondition)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CreateAndConditionFromNativeArray(conditions **IUIAutomationCondition, conditionCount int32, newCondition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateAndConditionFromNativeArray, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(conditions)), uintptr(conditionCount), uintptr(unsafe.Pointer(newCondition)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CreateOrCondition(condition1 *IUIAutomationCondition, condition2 *IUIAutomationCondition, newCondition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateOrCondition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(condition1)), uintptr(unsafe.Pointer(condition2)), uintptr(unsafe.Pointer(newCondition)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CreateOrConditionFromArray(conditions *SAFEARRAY, newCondition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateOrConditionFromArray, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(conditions)), uintptr(unsafe.Pointer(newCondition)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CreateOrConditionFromNativeArray(conditions **IUIAutomationCondition, conditionCount int32, newCondition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateOrConditionFromNativeArray, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(conditions)), uintptr(conditionCount), uintptr(unsafe.Pointer(newCondition)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CreateNotCondition(condition *IUIAutomationCondition, newCondition **IUIAutomationCondition) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateNotCondition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(condition)), uintptr(unsafe.Pointer(newCondition)))
	return HRESULT(ret)
}

func (this *IUIAutomation) AddAutomationEventHandler(eventId int32, element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddAutomationEventHandler, uintptr(unsafe.Pointer(this)), uintptr(eventId), uintptr(unsafe.Pointer(element)), uintptr(scope), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomation) RemoveAutomationEventHandler(eventId int32, element *IUIAutomationElement, handler *IUIAutomationEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveAutomationEventHandler, uintptr(unsafe.Pointer(this)), uintptr(eventId), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomation) AddPropertyChangedEventHandlerNativeArray(element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationPropertyChangedEventHandler, propertyArray *int32, propertyCount int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddPropertyChangedEventHandlerNativeArray, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(scope), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)), uintptr(unsafe.Pointer(propertyArray)), uintptr(propertyCount))
	return HRESULT(ret)
}

func (this *IUIAutomation) AddPropertyChangedEventHandler(element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationPropertyChangedEventHandler, propertyArray *SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddPropertyChangedEventHandler, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(scope), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)), uintptr(unsafe.Pointer(propertyArray)))
	return HRESULT(ret)
}

func (this *IUIAutomation) RemovePropertyChangedEventHandler(element *IUIAutomationElement, handler *IUIAutomationPropertyChangedEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemovePropertyChangedEventHandler, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomation) AddStructureChangedEventHandler(element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationStructureChangedEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddStructureChangedEventHandler, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(scope), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomation) RemoveStructureChangedEventHandler(element *IUIAutomationElement, handler *IUIAutomationStructureChangedEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveStructureChangedEventHandler, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomation) AddFocusChangedEventHandler(cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationFocusChangedEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddFocusChangedEventHandler, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomation) RemoveFocusChangedEventHandler(handler *IUIAutomationFocusChangedEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveFocusChangedEventHandler, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomation) RemoveAllEventHandlers() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveAllEventHandlers, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IUIAutomation) IntNativeArrayToSafeArray(array *int32, arrayCount int32, safeArray **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IntNativeArrayToSafeArray, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(array)), uintptr(arrayCount), uintptr(unsafe.Pointer(safeArray)))
	return HRESULT(ret)
}

func (this *IUIAutomation) IntSafeArrayToNativeArray(intArray *SAFEARRAY, array **int32, arrayCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IntSafeArrayToNativeArray, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(intArray)), uintptr(unsafe.Pointer(array)), uintptr(unsafe.Pointer(arrayCount)))
	return HRESULT(ret)
}

func (this *IUIAutomation) RectToVariant(rc RECT, var_ *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RectToVariant, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&rc)), uintptr(unsafe.Pointer(var_)))
	return HRESULT(ret)
}

func (this *IUIAutomation) VariantToRect(var_ VARIANT, rc *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().VariantToRect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&var_)), uintptr(unsafe.Pointer(rc)))
	return HRESULT(ret)
}

func (this *IUIAutomation) SafeArrayToRectNativeArray(rects *SAFEARRAY, rectArray **RECT, rectArrayCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SafeArrayToRectNativeArray, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rects)), uintptr(unsafe.Pointer(rectArray)), uintptr(unsafe.Pointer(rectArrayCount)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CreateProxyFactoryEntry(factory *IUIAutomationProxyFactory, factoryEntry **IUIAutomationProxyFactoryEntry) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateProxyFactoryEntry, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(factory)), uintptr(unsafe.Pointer(factoryEntry)))
	return HRESULT(ret)
}

func (this *IUIAutomation) Get_ProxyFactoryMapping(factoryMapping **IUIAutomationProxyFactoryMapping) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ProxyFactoryMapping, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(factoryMapping)))
	return HRESULT(ret)
}

func (this *IUIAutomation) GetPropertyProgrammaticName(property int32, name *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyProgrammaticName, uintptr(unsafe.Pointer(this)), uintptr(property), uintptr(unsafe.Pointer(name)))
	return HRESULT(ret)
}

func (this *IUIAutomation) GetPatternProgrammaticName(pattern int32, name *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPatternProgrammaticName, uintptr(unsafe.Pointer(this)), uintptr(pattern), uintptr(unsafe.Pointer(name)))
	return HRESULT(ret)
}

func (this *IUIAutomation) PollForPotentialSupportedPatterns(pElement *IUIAutomationElement, patternIds **SAFEARRAY, patternNames **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PollForPotentialSupportedPatterns, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pElement)), uintptr(unsafe.Pointer(patternIds)), uintptr(unsafe.Pointer(patternNames)))
	return HRESULT(ret)
}

func (this *IUIAutomation) PollForPotentialSupportedProperties(pElement *IUIAutomationElement, propertyIds **SAFEARRAY, propertyNames **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PollForPotentialSupportedProperties, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pElement)), uintptr(unsafe.Pointer(propertyIds)), uintptr(unsafe.Pointer(propertyNames)))
	return HRESULT(ret)
}

func (this *IUIAutomation) CheckNotSupported(value VARIANT, isNotSupported *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CheckNotSupported, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&value)), uintptr(unsafe.Pointer(isNotSupported)))
	return HRESULT(ret)
}

func (this *IUIAutomation) Get_ReservedNotSupportedValue(notSupportedValue **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ReservedNotSupportedValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(notSupportedValue)))
	return HRESULT(ret)
}

func (this *IUIAutomation) Get_ReservedMixedAttributeValue(mixedAttributeValue **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ReservedMixedAttributeValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(mixedAttributeValue)))
	return HRESULT(ret)
}

func (this *IUIAutomation) ElementFromIAccessible(accessible *IAccessible, childId int32, element **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ElementFromIAccessible, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(accessible)), uintptr(childId), uintptr(unsafe.Pointer(element)))
	return HRESULT(ret)
}

func (this *IUIAutomation) ElementFromIAccessibleBuildCache(accessible *IAccessible, childId int32, cacheRequest *IUIAutomationCacheRequest, element **IUIAutomationElement) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ElementFromIAccessibleBuildCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(accessible)), uintptr(childId), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(element)))
	return HRESULT(ret)
}

// 34723AFF-0C9D-49D0-9896-7AB52DF8CD8A
var IID_IUIAutomation2 = syscall.GUID{0x34723AFF, 0x0C9D, 0x49D0,
	[8]byte{0x98, 0x96, 0x7A, 0xB5, 0x2D, 0xF8, 0xCD, 0x8A}}

type IUIAutomation2Interface interface {
	IUIAutomationInterface
	Get_AutoSetFocus(autoSetFocus *BOOL) HRESULT
	Put_AutoSetFocus(autoSetFocus BOOL) HRESULT
	Get_ConnectionTimeout(timeout *uint32) HRESULT
	Put_ConnectionTimeout(timeout uint32) HRESULT
	Get_TransactionTimeout(timeout *uint32) HRESULT
	Put_TransactionTimeout(timeout uint32) HRESULT
}

type IUIAutomation2Vtbl struct {
	IUIAutomationVtbl
	Get_AutoSetFocus       uintptr
	Put_AutoSetFocus       uintptr
	Get_ConnectionTimeout  uintptr
	Put_ConnectionTimeout  uintptr
	Get_TransactionTimeout uintptr
	Put_TransactionTimeout uintptr
}

type IUIAutomation2 struct {
	IUIAutomation
}

func (this *IUIAutomation2) Vtbl() *IUIAutomation2Vtbl {
	return (*IUIAutomation2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomation2) Get_AutoSetFocus(autoSetFocus *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_AutoSetFocus, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(autoSetFocus)))
	return HRESULT(ret)
}

func (this *IUIAutomation2) Put_AutoSetFocus(autoSetFocus BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_AutoSetFocus, uintptr(unsafe.Pointer(this)), uintptr(autoSetFocus))
	return HRESULT(ret)
}

func (this *IUIAutomation2) Get_ConnectionTimeout(timeout *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ConnectionTimeout, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(timeout)))
	return HRESULT(ret)
}

func (this *IUIAutomation2) Put_ConnectionTimeout(timeout uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_ConnectionTimeout, uintptr(unsafe.Pointer(this)), uintptr(timeout))
	return HRESULT(ret)
}

func (this *IUIAutomation2) Get_TransactionTimeout(timeout *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_TransactionTimeout, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(timeout)))
	return HRESULT(ret)
}

func (this *IUIAutomation2) Put_TransactionTimeout(timeout uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_TransactionTimeout, uintptr(unsafe.Pointer(this)), uintptr(timeout))
	return HRESULT(ret)
}

// 73D768DA-9B51-4B89-936E-C209290973E7
var IID_IUIAutomation3 = syscall.GUID{0x73D768DA, 0x9B51, 0x4B89,
	[8]byte{0x93, 0x6E, 0xC2, 0x09, 0x29, 0x09, 0x73, 0xE7}}

type IUIAutomation3Interface interface {
	IUIAutomation2Interface
	AddTextEditTextChangedEventHandler(element *IUIAutomationElement, scope TreeScope, textEditChangeType TextEditChangeType, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationTextEditTextChangedEventHandler) HRESULT
	RemoveTextEditTextChangedEventHandler(element *IUIAutomationElement, handler *IUIAutomationTextEditTextChangedEventHandler) HRESULT
}

type IUIAutomation3Vtbl struct {
	IUIAutomation2Vtbl
	AddTextEditTextChangedEventHandler    uintptr
	RemoveTextEditTextChangedEventHandler uintptr
}

type IUIAutomation3 struct {
	IUIAutomation2
}

func (this *IUIAutomation3) Vtbl() *IUIAutomation3Vtbl {
	return (*IUIAutomation3Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomation3) AddTextEditTextChangedEventHandler(element *IUIAutomationElement, scope TreeScope, textEditChangeType TextEditChangeType, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationTextEditTextChangedEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddTextEditTextChangedEventHandler, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(scope), uintptr(textEditChangeType), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomation3) RemoveTextEditTextChangedEventHandler(element *IUIAutomationElement, handler *IUIAutomationTextEditTextChangedEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveTextEditTextChangedEventHandler, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

// 1189C02A-05F8-4319-8E21-E817E3DB2860
var IID_IUIAutomation4 = syscall.GUID{0x1189C02A, 0x05F8, 0x4319,
	[8]byte{0x8E, 0x21, 0xE8, 0x17, 0xE3, 0xDB, 0x28, 0x60}}

type IUIAutomation4Interface interface {
	IUIAutomation3Interface
	AddChangesEventHandler(element *IUIAutomationElement, scope TreeScope, changeTypes *int32, changesCount int32, pCacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationChangesEventHandler) HRESULT
	RemoveChangesEventHandler(element *IUIAutomationElement, handler *IUIAutomationChangesEventHandler) HRESULT
}

type IUIAutomation4Vtbl struct {
	IUIAutomation3Vtbl
	AddChangesEventHandler    uintptr
	RemoveChangesEventHandler uintptr
}

type IUIAutomation4 struct {
	IUIAutomation3
}

func (this *IUIAutomation4) Vtbl() *IUIAutomation4Vtbl {
	return (*IUIAutomation4Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomation4) AddChangesEventHandler(element *IUIAutomationElement, scope TreeScope, changeTypes *int32, changesCount int32, pCacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationChangesEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddChangesEventHandler, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(scope), uintptr(unsafe.Pointer(changeTypes)), uintptr(changesCount), uintptr(unsafe.Pointer(pCacheRequest)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomation4) RemoveChangesEventHandler(element *IUIAutomationElement, handler *IUIAutomationChangesEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveChangesEventHandler, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

// 25F700C8-D816-4057-A9DC-3CBDEE77E256
var IID_IUIAutomation5 = syscall.GUID{0x25F700C8, 0xD816, 0x4057,
	[8]byte{0xA9, 0xDC, 0x3C, 0xBD, 0xEE, 0x77, 0xE2, 0x56}}

type IUIAutomation5Interface interface {
	IUIAutomation4Interface
	AddNotificationEventHandler(element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationNotificationEventHandler) HRESULT
	RemoveNotificationEventHandler(element *IUIAutomationElement, handler *IUIAutomationNotificationEventHandler) HRESULT
}

type IUIAutomation5Vtbl struct {
	IUIAutomation4Vtbl
	AddNotificationEventHandler    uintptr
	RemoveNotificationEventHandler uintptr
}

type IUIAutomation5 struct {
	IUIAutomation4
}

func (this *IUIAutomation5) Vtbl() *IUIAutomation5Vtbl {
	return (*IUIAutomation5Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomation5) AddNotificationEventHandler(element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationNotificationEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddNotificationEventHandler, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(scope), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomation5) RemoveNotificationEventHandler(element *IUIAutomationElement, handler *IUIAutomationNotificationEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveNotificationEventHandler, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

// AAE072DA-29E3-413D-87A7-192DBF81ED10
var IID_IUIAutomation6 = syscall.GUID{0xAAE072DA, 0x29E3, 0x413D,
	[8]byte{0x87, 0xA7, 0x19, 0x2D, 0xBF, 0x81, 0xED, 0x10}}

type IUIAutomation6Interface interface {
	IUIAutomation5Interface
	CreateEventHandlerGroup(handlerGroup **IUIAutomationEventHandlerGroup) HRESULT
	AddEventHandlerGroup(element *IUIAutomationElement, handlerGroup *IUIAutomationEventHandlerGroup) HRESULT
	RemoveEventHandlerGroup(element *IUIAutomationElement, handlerGroup *IUIAutomationEventHandlerGroup) HRESULT
	Get_ConnectionRecoveryBehavior(connectionRecoveryBehaviorOptions *ConnectionRecoveryBehaviorOptions) HRESULT
	Put_ConnectionRecoveryBehavior(connectionRecoveryBehaviorOptions ConnectionRecoveryBehaviorOptions) HRESULT
	Get_CoalesceEvents(coalesceEventsOptions *CoalesceEventsOptions) HRESULT
	Put_CoalesceEvents(coalesceEventsOptions CoalesceEventsOptions) HRESULT
	AddActiveTextPositionChangedEventHandler(element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationActiveTextPositionChangedEventHandler) HRESULT
	RemoveActiveTextPositionChangedEventHandler(element *IUIAutomationElement, handler *IUIAutomationActiveTextPositionChangedEventHandler) HRESULT
}

type IUIAutomation6Vtbl struct {
	IUIAutomation5Vtbl
	CreateEventHandlerGroup                     uintptr
	AddEventHandlerGroup                        uintptr
	RemoveEventHandlerGroup                     uintptr
	Get_ConnectionRecoveryBehavior              uintptr
	Put_ConnectionRecoveryBehavior              uintptr
	Get_CoalesceEvents                          uintptr
	Put_CoalesceEvents                          uintptr
	AddActiveTextPositionChangedEventHandler    uintptr
	RemoveActiveTextPositionChangedEventHandler uintptr
}

type IUIAutomation6 struct {
	IUIAutomation5
}

func (this *IUIAutomation6) Vtbl() *IUIAutomation6Vtbl {
	return (*IUIAutomation6Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIAutomation6) CreateEventHandlerGroup(handlerGroup **IUIAutomationEventHandlerGroup) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateEventHandlerGroup, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(handlerGroup)))
	return HRESULT(ret)
}

func (this *IUIAutomation6) AddEventHandlerGroup(element *IUIAutomationElement, handlerGroup *IUIAutomationEventHandlerGroup) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddEventHandlerGroup, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(handlerGroup)))
	return HRESULT(ret)
}

func (this *IUIAutomation6) RemoveEventHandlerGroup(element *IUIAutomationElement, handlerGroup *IUIAutomationEventHandlerGroup) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveEventHandlerGroup, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(handlerGroup)))
	return HRESULT(ret)
}

func (this *IUIAutomation6) Get_ConnectionRecoveryBehavior(connectionRecoveryBehaviorOptions *ConnectionRecoveryBehaviorOptions) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ConnectionRecoveryBehavior, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(connectionRecoveryBehaviorOptions)))
	return HRESULT(ret)
}

func (this *IUIAutomation6) Put_ConnectionRecoveryBehavior(connectionRecoveryBehaviorOptions ConnectionRecoveryBehaviorOptions) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_ConnectionRecoveryBehavior, uintptr(unsafe.Pointer(this)), uintptr(connectionRecoveryBehaviorOptions))
	return HRESULT(ret)
}

func (this *IUIAutomation6) Get_CoalesceEvents(coalesceEventsOptions *CoalesceEventsOptions) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CoalesceEvents, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(coalesceEventsOptions)))
	return HRESULT(ret)
}

func (this *IUIAutomation6) Put_CoalesceEvents(coalesceEventsOptions CoalesceEventsOptions) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_CoalesceEvents, uintptr(unsafe.Pointer(this)), uintptr(coalesceEventsOptions))
	return HRESULT(ret)
}

func (this *IUIAutomation6) AddActiveTextPositionChangedEventHandler(element *IUIAutomationElement, scope TreeScope, cacheRequest *IUIAutomationCacheRequest, handler *IUIAutomationActiveTextPositionChangedEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddActiveTextPositionChangedEventHandler, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(scope), uintptr(unsafe.Pointer(cacheRequest)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

func (this *IUIAutomation6) RemoveActiveTextPositionChangedEventHandler(element *IUIAutomationElement, handler *IUIAutomationActiveTextPositionChangedEventHandler) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveActiveTextPositionChangedEventHandler, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(element)), uintptr(unsafe.Pointer(handler)))
	return HRESULT(ret)
}

var (
	pRegisterPointerInputTarget     uintptr
	pUnregisterPointerInputTarget   uintptr
	pRegisterPointerInputTargetEx   uintptr
	pUnregisterPointerInputTargetEx uintptr
	pNotifyWinEvent                 uintptr
	pSetWinEventHook                uintptr
	pIsWinEventHookInstalled        uintptr
	pUnhookWinEvent                 uintptr
)

func RegisterPointerInputTarget(hwnd HWND, pointerType POINTER_INPUT_TYPE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterPointerInputTarget, libUser32, "RegisterPointerInputTarget")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(pointerType))
	return BOOL(ret), WIN32_ERROR(err)
}

func UnregisterPointerInputTarget(hwnd HWND, pointerType POINTER_INPUT_TYPE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnregisterPointerInputTarget, libUser32, "UnregisterPointerInputTarget")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(pointerType))
	return BOOL(ret), WIN32_ERROR(err)
}

func RegisterPointerInputTargetEx(hwnd HWND, pointerType POINTER_INPUT_TYPE, fObserve BOOL) BOOL {
	addr := lazyAddr(&pRegisterPointerInputTargetEx, libUser32, "RegisterPointerInputTargetEx")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(pointerType), uintptr(fObserve))
	return BOOL(ret)
}

func UnregisterPointerInputTargetEx(hwnd HWND, pointerType POINTER_INPUT_TYPE) BOOL {
	addr := lazyAddr(&pUnregisterPointerInputTargetEx, libUser32, "UnregisterPointerInputTargetEx")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(pointerType))
	return BOOL(ret)
}

func NotifyWinEvent(event uint32, hwnd HWND, idObject int32, idChild int32) {
	addr := lazyAddr(&pNotifyWinEvent, libUser32, "NotifyWinEvent")
	syscall.SyscallN(addr, uintptr(event), hwnd, uintptr(idObject), uintptr(idChild))
}

func SetWinEventHook(eventMin uint32, eventMax uint32, hmodWinEventProc HINSTANCE, pfnWinEventProc WINEVENTPROC, idProcess uint32, idThread uint32, dwFlags uint32) HWINEVENTHOOK {
	addr := lazyAddr(&pSetWinEventHook, libUser32, "SetWinEventHook")
	ret, _, _ := syscall.SyscallN(addr, uintptr(eventMin), uintptr(eventMax), hmodWinEventProc, pfnWinEventProc, uintptr(idProcess), uintptr(idThread), uintptr(dwFlags))
	return ret
}

func IsWinEventHookInstalled(event uint32) BOOL {
	addr := lazyAddr(&pIsWinEventHookInstalled, libUser32, "IsWinEventHookInstalled")
	ret, _, _ := syscall.SyscallN(addr, uintptr(event))
	return BOOL(ret)
}

func UnhookWinEvent(hWinEventHook HWINEVENTHOOK) BOOL {
	addr := lazyAddr(&pUnhookWinEvent, libUser32, "UnhookWinEvent")
	ret, _, _ := syscall.SyscallN(addr, hWinEventHook)
	return BOOL(ret)
}
