package tree

import (
	"encoding/json"
	"sync"

	m "github.com/lodastack/models"
	"github.com/lodastack/registry/common"
)

type ReportInfo struct {
	sync.RWMutex
	ReportInfo reportMap
}

type reportMap map[string]m.Report

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

func (t *Tree) AgentReport(info m.Report) error {
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

func (t *Tree) GetReportInfo() map[string]m.Report {
	reportInfo := make(map[string]m.Report, len(t.reports.ReportInfo))
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
	return t.c.Update([]byte(reportBucket), []byte(reportBucket), reportByte)
}

func (t *Tree) readReport() (reportMap, error) {
	reportByte, err := t.c.View([]byte(reportBucket), []byte(reportBucket))
	if err != nil {
		return nil, err
	}
	return newReportMap(reportByte)
}
