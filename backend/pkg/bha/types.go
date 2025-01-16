package bha

// Result
// bha server result
type Result struct {
	FileId   string `json:"file_id"`   // 文件id
	FilePath string `json:"file_path"` // 文件路径
	FileArch string `json:"file_arch"` // 二进制架构
	FuncS    []Func `json:"funcs"`     // 函数集
}

// Func
// function
type Func struct {
	Addr    string       `json:"addr"`    // 函数地址
	FName   string       `json:"fname"`   // 检测文件函数名称
	Results []FuncResult `json:"results"` // 结果集
}

// FuncResult
// Function result
type FuncResult struct {
	Purl     string   `json:"purl,omitempty"`     // purl
	Version  string   `json:"version,omitempty"`  // 版本
	Refs     []string `json:"refs"`               // 引用
	FName    string   `json:"fname"`              // 匹配文件函数名称
	CVE      string   `json:"cve,omitempty"`      // CVE编号
	Arch     string   `json:"arch,omitempty"`     // 架构
	OptLevel string   `json:"optlevel,omitempty"` // 优化等级
	Sim      float64  `json:"sim"`                // 相似分数
}

// 检测方式
const (
	FastDetectMethod        = "fast"        // 快速
	IntelligentDetectMethod = "intelligent" // 智能
)

// 检测算法
const (
	SFSAlgorithm  = "sfs"
	SSFSAlgorithm = "ssfs"
	BSDAlgorithm  = "bsd"
)

func DetectMethods() []string {
	return []string{FastDetectMethod, IntelligentDetectMethod}
}

func Algorithms() []string {
	return append(FastDMAlgorithms(), IntelligentDMAlgorithms()...)
}

func FastDMAlgorithms() []string {
	return []string{SFSAlgorithm}
}

func IntelligentDMAlgorithms() []string {
	return []string{SSFSAlgorithm, BSDAlgorithm}
}

const (
	StatusSuccessful = "successful"
	StatusFailed     = "failed"
	StatusTerminated = "terminated"
	StatusTimeout    = "timeout"
)

func Status() []string {
	return []string{StatusSuccessful, StatusFailed, StatusTerminated, StatusTimeout}
}
