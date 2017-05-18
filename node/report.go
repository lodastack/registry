package node

import (
	"sync"

	m "github.com/lodastack/models"
)

type ReportInfo struct {
	sync.RWMutex
	ReportInfo map[string]m.Report
}

func (t *Tree) AgentReport(info m.Report) error {
	t.reports.Lock()
	defer t.reports.Unlock()
	if info.NewHostname == "" {
		return ErrInvalidParam
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
