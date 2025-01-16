package bhaserver

import (
	"bin-vul-inspector/pkg/client"
)

type ScanReq struct {
	Type       string  `json:"type"`        // type of the scan, e.g., SFS/SSFS/BSD
	OssBucket  string  `json:"oss_bucket"`  // name of the oss bucket
	InputPath  string  `json:"input_path"`  // path of the scan target
	OutputDir  string  `json:"output_dir"`  // directory where the scan results will be stored
	ModelPath  string  `json:"model_path"`  // path of the model file
	ModelMD5   string  `json:"model_md5"`   // md5	of the model file, for integrity verification
	TopN       uint    `json:"top_n"`       // number of top results to return from the scan(default is 100)
	MinimumSim float32 `json:"minimum_sim"` // minimum_sim
	Timeout    uint    `json:"timeout"`     // time in minutes to wait before the scan times out (default is 60 minutes)
}

type ScanResp struct {
	client.Response
	Data struct {
		Id string `json:"id"` // scan id
	}
}
