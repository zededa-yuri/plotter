// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    sysStatLog, err := UnmarshalSysStatLog(bytes)
//    bytes, err = sysStatLog.Marshal()

package main

import "encoding/json"

type SysStatLog []SysStatLogElement

func UnmarshalSysStatLog(data []byte) (SysStatLog, error) {
	var r SysStatLog
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *SysStatLog) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type SysStatLogElement struct {
	Date   string `json:"date"`
	Memory Memory `json:"memory"`
	CPU    CPU    `json:"cpu"`
}
