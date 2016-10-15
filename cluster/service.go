package cluster

import (
	"net"
	"sync"
	"time"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/model"
)

const (
	connectionTimeout = 10 * time.Second
)

// Transport is the interface the network service must provide.
type Transport interface {
	net.Listener

	// Dial is used to create a new outgoing connection.
	Dial(address string, timeout time.Duration) (net.Conn, error)
}

// Store represents a store of information, managed via consensus.
type Store interface {
	// Leader returns the leader of the consensus system.
	Leader() string

	// Join joins the node, reachable at addr, to the cluster.
	Join(addr string) error

	// UpdateAPIPeers updates the API peers on the store.
	UpdateAPIPeers(peers map[string]string) error

	// Create a bucket, via distributed consensus.
	CreateBucket(name []byte) error

	// Remove a bucket, via distributed consensus.
	RemoveBucket(name []byte) error

	// Get returns the value for the given key.
	View(bucket, key []byte) ([]byte, error)

	// Set sets the value for the given key, via distributed consensus.
	Update(bucket []byte, key []byte, value []byte) error

	// Batch update values for given keys in given buckets, via distributed consensus.
	Batch(rows []model.Row) error

	// Backup database.
	Backup() ([]byte, error)
}

// Service allows access to the cluster and associated meta data,
// via distributed consensus.
type Service struct {
	tn    Transport
	store Store
	addr  net.Addr

	wg sync.WaitGroup

	logger *log.Logger
}

// NewService returns a new instance of the cluster service.
func NewService(tn Transport, store Store) *Service {
	return &Service{
		tn:     tn,
		store:  store,
		addr:   tn.Addr(),
		logger: log.New("INFO", "cluster", model.LogBackend),
	}
}

// Open opens the Service.
func (s *Service) Open() error {
	s.wg.Add(1)
	go s.serve()
	s.logger.Println("service listening on", s.tn.Addr())
	return nil
}

// Close closes the service.
func (s *Service) Close() error {
	s.tn.Close()
	s.wg.Wait()
	return nil
}

// Addr returns the address the service is listening on.
func (s *Service) Addr() string {
	return s.addr.String()
}

func (s *Service) serve() error {
	defer s.wg.Done()

	for {
		conn, err := s.tn.Accept()
		if err != nil {
			return err
		}

		go s.handleConn(conn)
	}
}
