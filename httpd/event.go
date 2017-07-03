package httpd

import (
	"fmt"
	"github.com/lodastack/log"
	"github.com/lodastack/registry/config"
	"github.com/lodastack/registry/utils"
	"net/http"
)

var clearStatusUrl string = "%s?ns=%s&host=%s"

func clearMachineStatus(hostname string, ns ...string) error {
	if len(ns) == 0 || hostname == "" {
		return ErrInvalidParam
	}

	for _, _ns := range ns {
		q := utils.HttpQuery{
			Timeout:  3,
			Method:   http.MethodPost,
			Url:      fmt.Sprintf(clearStatusUrl, config.C.EventConf.ClearURL, _ns, hostname),
			BodyType: utils.Raw,
		}
		if err := q.DoQuery(); err != nil || q.Result.Status > 299 {
			log.Errorf("clear ns %s host %s status fail", ns, hostname)
		}
	}
	return nil
}
