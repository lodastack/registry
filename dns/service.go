package dns

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/config"
	"github.com/lodastack/registry/httpd"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/tree"

	dnslib "github.com/miekg/dns"
)

const (
	resType       = "machine"
	domainSuffix  = "."
	matchIPPrefix = "10."
	purgeInterval = 2
)

func (s *Service) parseQuery(m *dnslib.Msg) {
	machines := func(s *Service, ns string, domain string) []dnslib.RR {
		var res []dnslib.RR
		resList, err := s.tree.GetResourceList(ns, resType)
		if err != nil {
			s.logger.Errorf("DNS search failed: %s", err)
			return res
		}
		if resList == nil {
			return res
		}

		var iparray []string
		for _, r := range *resList {
			if ips, ok := r[model.IpProp]; ok {
				iparray = append(iparray, strings.Split(ips, ",")...)
			}
		}

		for _, ip := range removeRepByMap(iparray) {
			// only return prefix matched IPs
			if ip != "" && strings.HasPrefix(ip, matchIPPrefix) {
				rr, err := dnslib.NewRR(fmt.Sprintf("%s A %s", domain, ip))
				if err == nil {
					rr.Header().Ttl = 60
					res = append(res, rr)
				}
			}
		}

		s.mu.Lock()
		s.cache[domain] = res
		s.mu.Unlock()
		return res
	}

	for _, q := range m.Question {
		switch q.Qtype {
		case dnslib.TypeA:
			s.logger.Infof("Query for %s", q.Name)
			ns := strings.TrimSuffix(q.Name, domainSuffix)
			s.mu.RLock()
			answer, ok := s.cache[q.Name]
			s.mu.RUnlock()
			if ok {
				m.Answer = answer
				return
			}
			m.Answer = machines(s, ns, q.Name)
			return
		}
	}
}

func (s *Service) handleDNSRequest(w dnslib.ResponseWriter, r *dnslib.Msg) {
	m := new(dnslib.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dnslib.OpcodeQuery:
		s.parseQuery(m)
	}

	w.WriteMsg(m)
}

// Service provides DNS service.
type Service struct {
	enable bool
	root   string
	port   int
	conf   config.DNSConfig
	server *dnslib.Server

	mu    sync.RWMutex
	cache map[string][]dnslib.RR

	tree tree.TreeMethod

	logger *log.Logger
}

// New DNS service
func New(rootName string, c config.DNSConfig, cluster httpd.Cluster) (*Service, error) {
	tree, err := tree.NewTree(rootName, cluster)
	if err != nil {
		log.Errorf("init tree fail: %s", err.Error())
		return nil, err
	}
	return &Service{
		enable: c.Enable,
		root:   rootName,
		port:   c.Port,
		conf:   c,
		server: &dnslib.Server{Addr: ":" + strconv.Itoa(c.Port), Net: "udp"},
		cache:  make(map[string][]dnslib.RR),
		tree:   tree,

		logger: log.New("INFO", "dns", model.LogBackend),
	}, nil
}

// Start DNS service
func (s *Service) Start() error {
	if !s.enable {
		s.logger.Info("DNS module not enable")
		return nil
	}
	// attach request handler func
	dnslib.HandleFunc(s.root+".", s.handleDNSRequest)

	// start server
	s.logger.Infof("Starting DNS module at %d", s.port)
	go func() {
		err := s.server.ListenAndServe()
		if err != nil {
			s.logger.Errorf("Failed to start DNS service: %s", err.Error())
		}
	}()
	go s.purgeCache()
	return nil
}

// Close DNS service
func (s *Service) Close() error {
	if !s.enable {
		return nil
	}
	return s.server.Shutdown()
}

func (s *Service) purgeCache() {
	ticker := time.NewTicker(time.Duration(purgeInterval) * time.Minute)
	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			s.cache = make(map[string][]dnslib.RR)
			s.mu.Unlock()
		}
	}
}

func removeRepByMap(slc []string) []string {
	var result []string
	tempMap := map[string]byte{}
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l {
			result = append(result, e)
		}
	}
	return result
}
