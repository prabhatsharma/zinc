/* Copyright 2022 Zinc Labs Inc. and Contributors
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package core

import (
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"gopkg.in/segmentio/analytics-go.v3"

	"github.com/zinclabs/zinc/pkg/config"
	"github.com/zinclabs/zinc/pkg/ider"
	"github.com/zinclabs/zinc/pkg/meta"
	"github.com/zinclabs/zinc/pkg/metadata"
)

// Telemetry instance
var Telemetry = newTelemetry()

type telemetry struct {
	instanceID   string
	events       chan analytics.Track
	baseInfo     map[string]interface{}
	baseInfoOnce sync.Once
}

func newTelemetry() *telemetry {
	t := new(telemetry)
	t.events = make(chan analytics.Track, 100)
	t.initBaseInfo()

	go t.runEvents()

	return t
}

func (t *telemetry) createInstanceID() string {
	instanceID := ider.Generate()
	_ = metadata.KV.Set("instance_id", []byte(instanceID))
	return instanceID
}

func (t *telemetry) getInstanceID() string {
	if t.instanceID != "" {
		return t.instanceID
	}

	val, err := metadata.KV.Get("instance_id")
	if err != nil {
		log.Error().Err(err).Msg("core.Telemetry.GetInstanceID: error accessing stored fields")
	}
	if val != nil {
		t.instanceID = string(val)
	}
	if t.instanceID == "" {
		t.instanceID = t.createInstanceID()
	}
	return t.instanceID
}

func (t *telemetry) initBaseInfo() {
	t.baseInfoOnce.Do(func() {
		m, _ := mem.VirtualMemory()
		cpuCount, _ := cpu.Counts(true)
		zone, _ := time.Now().Local().Zone()

		t.baseInfo = map[string]interface{}{
			"os":           runtime.GOOS,
			"arch":         runtime.GOARCH,
			"zinc_version": meta.Version,
			"time_zone":    zone,
			"cpu_count":    cpuCount,
			"total_memory": m.Total / 1024 / 1024,
		}
	})
}

func (t *telemetry) Instance() {
	if !config.Global.TelemetryEnable {
		return
	}

	traits := analytics.NewTraits().
		Set("index_count", len(ZINC_INDEX_LIST)).
		Set("total_index_size_mb", t.TotalIndexSize())

	for k, v := range t.baseInfo {
		traits.Set(k, v)
	}

	_ = meta.SEGMENT_CLIENT.Enqueue(analytics.Identify{
		UserId: t.getInstanceID(),
		Traits: traits,
	})
}

func (t *telemetry) Event(event string, data map[string]interface{}) {
	if !config.Global.TelemetryEnable {
		return
	}

	props := analytics.NewProperties()
	for k, v := range t.baseInfo {
		props.Set(k, v)
	}
	for k, v := range data {
		props.Set(k, v)
	}

	t.events <- analytics.Track{
		UserId:     t.getInstanceID(),
		Event:      event,
		Properties: props,
	}
}

func (t *telemetry) runEvents() {
	for event := range t.events {
		_ = meta.SEGMENT_CLIENT.Enqueue(event)
	}
}

func (t *telemetry) TotalIndexSize() float64 {
	TotalIndexSize := 0.0
	for k := range ZINC_INDEX_LIST {
		TotalIndexSize += t.GetIndexSize(k)
	}
	return math.Round(TotalIndexSize)
}

func (t *telemetry) GetIndexSize(indexName string) float64 {
	if index, ok := ZINC_INDEX_LIST[indexName]; ok {
		return index.LoadStorageSize()
	}
	return 0.0
}

func (t *telemetry) HeartBeat() {
	m, _ := mem.VirtualMemory()
	data := make(map[string]interface{})
	data["index_count"] = len(ZINC_INDEX_LIST)
	data["total_index_size_mb"] = t.TotalIndexSize()
	data["memory_used_percent"] = m.UsedPercent
	t.Event("heartbeat", data)
}

func (t *telemetry) Cron() {
	c := cron.New()
	_, _ = c.AddFunc("@every 30m", t.HeartBeat)
	c.Start()
}
