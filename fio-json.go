package main

import "encoding/json"

func UnmarshalFio(data []byte) (Fio, error) {
	var r Fio
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Fio) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Fio struct {
	FioVersion    string        `json:"fio version"`
	Timestamp     int64         `json:"timestamp"`
	TimestampMS   int64         `json:"timestamp_ms"`
	Time          string        `json:"time"`
	GlobalOptions GlobalOptions `json:"global options"`
	Jobs          []Job         `json:"jobs"`
	DiskUtil      []DiskUtil    `json:"disk_util"`
}

type DiskUtil struct {
	Name        string  `json:"name"`
	ReadIos     int64   `json:"read_ios"`
	WriteIos    int64   `json:"write_ios"`
	ReadMerges  int64   `json:"read_merges"`
	WriteMerges int64   `json:"write_merges"`
	ReadTicks   int64   `json:"read_ticks"`
	WriteTicks  int64   `json:"write_ticks"`
	InQueue     int64   `json:"in_queue"`
	Util        float64 `json:"util"`
}

type GlobalOptions struct {
	Size     string `json:"size"`
	Runtime  string `json:"runtime"`
	Filename string `json:"filename"`
}

type Job struct {
	Jobname           string             `json:"jobname"`
	Groupid           int64              `json:"groupid"`
	Error             int64              `json:"error"`
	Eta               int64              `json:"eta"`
	Elapsed           int64              `json:"elapsed"`
	JobOptions        JobOptions         `json:"job options"`
	Read              Read               `json:"read"`
	Write             Read               `json:"write"`
	Trim              Read               `json:"trim"`
	USRCPU            float64            `json:"usr_cpu"`
	SysCPU            float64            `json:"sys_cpu"`
	Ctx               int64              `json:"ctx"`
	Majf              int64              `json:"majf"`
	Minf              int64              `json:"minf"`
	IodepthLevel      IodepthLevel       `json:"iodepth_level"`
	LatencyNS         map[string]float64 `json:"latency_ns"`
	LatencyUs         map[string]float64 `json:"latency_us"`
	LatencyMS         map[string]float64 `json:"latency_ms"`
	LatencyDepth      int64              `json:"latency_depth"`
	LatencyTarget     int64              `json:"latency_target"`
	LatencyPercentile float64            `json:"latency_percentile"`
	LatencyWindow     int64              `json:"latency_window"`
}

type IodepthLevel struct {
	The1  float64 `json:"1"`
	The2  float64 `json:"2"`
	The4  float64 `json:"4"`
	The8  float64 `json:"8"`
	The16 float64 `json:"16"`
	The32 float64 `json:"32"`
	The64 float64 `json:">=64"`
}

type JobOptions struct {
	Rw             string `json:"rw"`
	Bs             string `json:"bs"`
	Ioengine       string `json:"ioengine"`
	Iodepth        string `json:"iodepth"`
	Numjobs        string `json:"numjobs"`
	Direct         string `json:"direct"`
	GroupReporting string `json:"group_reporting"`
	Invalidate     string `json:"invalidate"`
	Loops          string `json:"loops"`
	WriteBWLog     string `json:"write_bw_log"`
	WriteLatLog    string `json:"write_lat_log"`
	WriteIopsLog   string `json:"write_iops_log"`
	LogAvgMsec     string `json:"log_avg_msec"`
}

type Read struct {
	IoBytes     int64   `json:"io_bytes"`
	IoKbytes    int64   `json:"io_kbytes"`
	BW          int64   `json:"bw"`
	Iops        float64 `json:"iops"`
	Runtime     int64   `json:"runtime"`
	TotalIos    int64   `json:"total_ios"`
	ShortIos    int64   `json:"short_ios"`
	DropIos     int64   `json:"drop_ios"`
	SlatNS      LatNS   `json:"slat_ns"`
	ClatNS      LatNS   `json:"clat_ns"`
	LatNS       LatNS   `json:"lat_ns"`
	BWMin       int64   `json:"bw_min"`
	BWMax       int64   `json:"bw_max"`
	BWAgg       float64 `json:"bw_agg"`
	BWMean      float64 `json:"bw_mean"`
	BWDev       float64 `json:"bw_dev"`
	BWSamples   int64   `json:"bw_samples"`
	IopsMin     int64   `json:"iops_min"`
	IopsMax     int64   `json:"iops_max"`
	IopsMean    float64 `json:"iops_mean"`
	IopsStddev  float64 `json:"iops_stddev"`
	IopsSamples int64   `json:"iops_samples"`
}

type LatNS struct {
	Min        int64            `json:"min"`
	Max        int64            `json:"max"`
	Mean       float64          `json:"mean"`
	Stddev     float64          `json:"stddev"`
	Percentile map[string]int64 `json:"percentile,omitempty"`
}
