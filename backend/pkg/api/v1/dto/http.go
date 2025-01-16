package dto

const (
	StatusOk = 0

	StatusInvalidJson   = 1500
	StatusParamInvalid  = 1501
	StatusParamTooLarge = 1502
	StatusUnauthorized  = 1503
	StatusDataNotFound  = 1510
	StatusInternalError = 1999

	StatusErrJson                  = 1505
	StatusSaveFileErr              = 1506
	StatusErrDb                    = 1507
	StatusQueryParamNotExist       = 1508
	StatusSha256File               = 1513
	StatusErrParseFormData         = 1514
	StatusNotMatchedVersion        = 1515
	StatusTaskIdInvalid            = 1520
	StatusTaskNotMatchImageProject = 1521
	StatusErrDecryptFile           = 1522

	StatusRequestBodyLarge         = 1600
	StatusErrContentRangeHeader    = 1601
	StatusErrMissingFilenameHeader = 1602
	StatusErrMissingFile           = 1603
	StatusErrWrongStartHeader      = 1604
	StatusErrWrongFilesize         = 1605
	StatusErrFileNotExist          = 1606

	StatusErrQueryIncremental = 1700

	StatusErrQueryABI = 1711

	StatusSastRuleNameEmpty = 2020
	StatusTaskRepositoryErr = 2021
	StatusCreateTaskErr     = 2022

	StatusUserPasswordErr       = 2030
	StatusUserNameUsedErr       = 2031
	StatusUserNotExistErr       = 2032
	StatusForbiddenEditAdminErr = 2033
	StatusLoginErr              = 2034
	StatusLdapUpdateErr         = 2035
	StatusLdapDeleteErr         = 2036

	StatusKeyNotExistErr = 2040
	StatusKeyUsedErr     = 2041
	StatusKeyNameUsedErr = 2042

	StatusRegistryNotExistErr      = 2050
	StatusRegistryNameUsedErr      = 2051
	StatusRegistrySwapOrderErr     = 2052
	StatusRegistryKeyTypeErr       = 2053
	StatusRegistryInternalExistErr = 2054

	StatusDeptNotExistErr                = 2080
	StatusDeptNameUsedErr                = 2081
	StatusDeptForbiddenDeleteErr         = 2082
	StatusDeptForbiddenDeleteByMemberErr = 2083

	StatusMemberNotExistErr = 2090
	StatusMemberForbidden   = 2091
	StatusMemberExistErr    = 2092

	StatusAssetNotExistErr    = 2010
	StatusAssetNotPermission  = 2011
	StatusAssetNameVersionErr = 2012
	StatusAssetNameUsedErr    = 2013
	StatusAssetTypeErr        = 2014

	StatusLicenseUnAuthorize          = 2100
	StatusLicenseInvalid              = 2101
	StatusLicenseExpired              = 2102
	StatusLicenseParseErr             = 2105
	StatusLicenseMaintenanceExpired   = 2106
	StatusLicenseAuthorizeErr         = 2107
	StatusLicenseNotMatchMachineIdErr = 2108
	StatusLicenseNotInitErr           = 2109

	StatusNetworkGetErr = 2110
	StatusNetworkSetErr = 2111

	StatusRuleNotExistErr   = 2200
	StatusRuleNameUsedErr   = 2201
	StatusRuleNotPermission = 2203
	StatusNotOwnerErr       = 2204
	StatusNotOwnerDeleteErr = 2205

	StatusAssetGroupNotExistErr   = 2210
	StatusAssetGroupNameUsedErr   = 2211
	StatusAssetGroupNotPermission = 2212

	StatusOrgNotExistErr = 2220
	StatusOrgNameUsedErr = 2221

	StatusExprNotExistErr      = 2230
	StatusExprNotPermission    = 2231
	StatusExprNoRunnable       = 2232
	StatusExprAssetDisabled    = 2233
	StatusExprSubGroupDisabled = 2234
	StatusExprGroupDisabled    = 2235
	StatusExprGlobalDisabled   = 2236

	StatusDetectTmplNotExistErr = 2240
	StatusDetectTmplNameUsedErr = 2241
)

var statusText = map[int]string{
	StatusInvalidJson:   "json数据格式错误",
	StatusParamInvalid:  "无效参数",
	StatusParamTooLarge: "上传文件大小超过限制",
	StatusUnauthorized:  "没有该企业的操作权限",
	StatusDataNotFound:  "请求数据不存在",
	StatusInternalError: "内部错误,请稍后重试",

	StatusErrJson:            "请求数据json格式错误",
	StatusSaveFileErr:        "保存文件时出错",
	StatusErrDb:              "数据处理失败",
	StatusQueryParamNotExist: "路由参数缺失",
	StatusSha256File:         "计算文件SHA-256哈希失败",
	StatusErrParseFormData:   "请求form-data数据解析失败",

	StatusTaskIdInvalid:            "参数taskId不能为空",
	StatusTaskNotMatchImageProject: "不能扫描非镜像项目",
	StatusErrDecryptFile:           "打开加密文件错误",

	StatusRequestBodyLarge:         "请求数据太大",
	StatusErrContentRangeHeader:    "Content-Range请求头格式错误",
	StatusErrMissingFilenameHeader: "请求头X-Filename解析失败",
	StatusErrMissingFile:           "获取file表单数据文件错误",
	StatusErrWrongStartHeader:      "Content-Range请求头首次上传文件时start不为0",
	StatusErrWrongFilesize:         "上传片断请求与已上传文件大小不一致",
	StatusErrFileNotExist:          "文件不存在",

	StatusErrQueryIncremental: "不支持增量查询",

	StatusErrQueryABI: "查询组件ABI信息失败",

	StatusSastRuleNameEmpty: "规则名称错误",
	StatusTaskRepositoryErr: "仓库中地址配置不正确或者与密钥不匹配",
	StatusCreateTaskErr:     "批量创建任务失败",

	StatusUserPasswordErr:       "用户密码错误",
	StatusUserNameUsedErr:       "用户账号已被使用",
	StatusUserNotExistErr:       "用户不存在",
	StatusForbiddenEditAdminErr: "禁止编辑管理用户",
	StatusLoginErr:              "用户名或密码错误",
	StatusLdapUpdateErr:         "ladp用户不能更改密码",
	StatusLdapDeleteErr:         "ladp用户不能删除",

	StatusKeyNotExistErr: "密钥不存在",
	StatusKeyUsedErr:     "密钥已被镜像引用,请解除引用后删除",
	StatusKeyNameUsedErr: "密钥名称已被使用",

	StatusRegistryNotExistErr:      "镜像源不存在",
	StatusRegistryNameUsedErr:      "镜像名称已被使用",
	StatusRegistrySwapOrderErr:     "镜像源交换顺序语言不一致",
	StatusRegistryKeyTypeErr:       "镜像源选择密钥类型错误",
	StatusRegistryInternalExistErr: "不能重复创建内置镜像源",

	StatusDeptNotExistErr:                "部门不存在",
	StatusDeptNameUsedErr:                "部门名称已被使用",
	StatusDeptForbiddenDeleteErr:         "部门下具有用户禁止删除",
	StatusDeptForbiddenDeleteByMemberErr: "部门已被项目组或项目分配禁止删除",

	StatusMemberNotExistErr: "关系不存在",
	StatusMemberForbidden:   "禁止删除",
	StatusMemberExistErr:    "关系已经存在",

	StatusAssetNotExistErr:    "资产不存在",
	StatusAssetNotPermission:  "无该资产权限",
	StatusAssetNameVersionErr: "资产名称版本重复",
	StatusAssetNameUsedErr:    "资产名称重复",
	StatusAssetTypeErr:        "资产类型错误",

	StatusLicenseUnAuthorize:          "服务端未授权，请联系售后",
	StatusLicenseInvalid:              "证书类型无效，请联系售后",
	StatusLicenseExpired:              "授权已到期，功能受到限制。请联系销售人员续期授权许可",
	StatusLicenseParseErr:             "证书解析失败",
	StatusLicenseMaintenanceExpired:   "维保时间已经过期",
	StatusLicenseAuthorizeErr:         "授权失败",
	StatusLicenseNotMatchMachineIdErr: "证书机器码不匹配",
	StatusLicenseNotInitErr:           "服务端证书未初始化",

	StatusNetworkGetErr: "获取网络配置失败，请稍后重试",
	StatusNetworkSetErr: "配置网络失败",

	StatusRuleNotExistErr:   "规则不存在",
	StatusRuleNameUsedErr:   "规则名称已经被占用",
	StatusRuleNotPermission: "无该规则权限",
	StatusNotOwnerErr:       "非拥有者不能更改规则权限",
	StatusNotOwnerDeleteErr: "非拥有者不能删除规则",

	StatusAssetGroupNotExistErr:   "资产组不存在",
	StatusAssetGroupNameUsedErr:   "资产组名称已被使用",
	StatusAssetGroupNotPermission: "无该资产组权限",

	StatusOrgNotExistErr: "组织不存在",
	StatusOrgNameUsedErr: "组织名称已被使用",

	StatusExprNotExistErr:      "表达式不存在",
	StatusExprNotPermission:    "无表达式权限",
	StatusExprNoRunnable:       "无可运行的表达式",
	StatusExprAssetDisabled:    "应用表达式功能已禁用",
	StatusExprSubGroupDisabled: "项目表达式功能已禁用",
	StatusExprGroupDisabled:    "资产组表达式功能已禁用",
	StatusExprGlobalDisabled:   "表达式功能已禁用",

	StatusDetectTmplNotExistErr: "检测模板不存在",
	StatusDetectTmplNameUsedErr: "检测模板名称已被使用",
}

// StatusText returns a text for the HTTP status code. It returns the empty
// string if the code is unknown.
func StatusText(code int) string {
	return statusText[code]
}
