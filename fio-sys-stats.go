// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    welcome, err := UnmarshalWelcome(bytes)
//    bytes, err = welcome.Marshal()

package main

import "encoding/json"

func UnmarshalSysStat(data []byte) (SysStat, error) {
	var r SysStat
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *SysStat) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type SysStat struct {
	Zfs        Zfs  `json:"zfs"`
	BeforeTest Test `json:"before_test"`
	AfterTest  Test `json:"after_test"`
}

type Test struct {
	Date          string `json:"date"`
	Memory        Memory `json:"memory"`
	CPU           CPU    `json:"cpu"`
	Fragmentation string `json:"fragmentation"`
}

type CPU struct {
	User    int64 `json:"user"`
	Nice    int64 `json:"nice"`
	System  int64 `json:"system"`
	Idle    int64 `json:"idle"`
	Iowait  int64 `json:"iowait"`
	IRQ     int64 `json:"irq"`
	Softirq int64 `json:"softirq"`
	Steal   int64 `json:"steal"`
	Guest   int64 `json:"guest"`
}

type Memory struct {
	MemTotal  int64 `json:"MemTotal"`
	MemFree   int64 `json:"MemFree"`
	Cached    int64 `json:"Cached"`
	SwapTotal int64 `json:"swapTotal"`
	SwapFree  int64 `json:"swapFree"`
}

type Zfs struct {
	Version                                string `json:"version"`
	ZfsCompressedArcEnabled                int64  `json:"zfs_compressed_arc_enabled"`
	ZfsVdevMinAutoAshift                   int64  `json:"zfs_vdev_min_auto_ashift"`
	ZvolRequestSync                        int64  `json:"zvol_request_sync"`
	ZfsArcMin                              int64  `json:"zfs_arc_min"`
	ZfsArcMax                              int64  `json:"zfs_arc_max"`
	ZfsVdevAggregationLimitNonRotating     int64  `json:"zfs_vdev_aggregation_limit_non_rotating"`
	ZfsVdevAsyncWriteActiveMinDirtyPercent int64  `json:"zfs_vdev_async_write_active_min_dirty_percent"`
	ZfsVdevAsyncWriteActiveMaxDirtyPercent int64  `json:"zfs_vdev_async_write_active_max_dirty_percent"`
	ZfsDelayMinDirtyPercent                int64  `json:"zfs_delay_min_dirty_percent"`
	ZfsDelayScale                          int64  `json:"zfs_delay_scale"`
	ZfsDirtyDataMax                        int64  `json:"zfs_dirty_data_max"`
	ZfsDirtyDataSyncPercent                int64  `json:"zfs_dirty_data_sync_percent"`
	ZfsPrefetchDisable                     int64  `json:"zfs_prefetch_disable"`
	ZfsVdevSyncReadMinActive               int64  `json:"zfs_vdev_sync_read_min_active"`
	ZfsVdevSyncReadMaxActive               int64  `json:"zfs_vdev_sync_read_max_active"`
	ZfsVdevSyncWriteMinActive              int64  `json:"zfs_vdev_sync_write_min_active"`
	ZfsVdevSyncWriteMaxActive              int64  `json:"zfs_vdev_sync_write_max_active"`
	ZfsVdevAsyncReadMinActive              int64  `json:"zfs_vdev_async_read_min_active"`
	ZfsVdevAsyncReadMaxActive              int64  `json:"zfs_vdev_async_read_max_active"`
	ZfsVdevAsyncWriteMinActive             int64  `json:"zfs_vdev_async_write_min_active"`
	ZfsVdevAsyncWriteMaxActive             int64  `json:"zfs_vdev_async_write_max_active"`
}
