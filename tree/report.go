package tree

import (
	"encoding/json"
	"sync"

	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/model"
)

// ReportInfo save the agent report infomation.
type ReportInfo struct {
	sync.RWMutex
	ReportInfo reportMap
}

type reportMap map[string]model.Report

func (r *reportMap) Byte() ([]byte, error) {
	return json.Marshal(*r)
}

func (r *reportMap) Unmarshal(data []byte) error {
	return json.Unmarshal(data, r)
}

func newReportMap(data []byte) (reports reportMap, err error) {
	err = reports.Unmarshal(data)
	return
}

// AgentReport handle and save the agent report message.
func (t *Tree) AgentReport(info model.Report) error {
	t.reports.Lock()
	defer t.reports.Unlock()
	if info.NewHostname == "" {
		return common.ErrInvalidParam
	}
	if info.OldHostname != info.NewHostname {
		delete(t.reports.ReportInfo, info.OldHostname)
	}
	t.reports.ReportInfo[info.NewHostname] = info
	return nil
}

// GetReportInfo return all report information.
func (t *Tree) GetReportInfo() map[string]model.Report {
	reportInfo := make(map[string]model.Report, len(t.reports.ReportInfo))
	t.reports.RLock()
	defer t.reports.RUnlock()
	for k, v := range t.reports.ReportInfo {
		reportInfo[k] = v
	}
	return reportInfo
}

func (t *Tree) setReport(reports reportMap) error {
	reportByte, err := reports.Byte()
	if err != nil {
		return err
	}
	return t.cluster.Update([]byte(reportBucket), []byte(reportBucket), reportByte)
}

func (t *Tree) readReport() (reportMap, error) {
	reportByte, err := t.cluster.View([]byte(reportBucket), []byte(reportBucket))
	if err != nil {
		return nil, err
	}
	return newReportMap(reportByte)
}
