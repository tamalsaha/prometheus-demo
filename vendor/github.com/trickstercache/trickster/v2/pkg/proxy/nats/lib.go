package nats

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	no "github.com/trickstercache/trickster/v2/pkg/proxy/nats/options"

	"github.com/nats-io/nats.go"
	"k8s.io/klog/v2"
)

const (
	natsConnectionTimeout       = 350 * time.Millisecond
	natsConnectionRetryInterval = 100 * time.Millisecond
)

var (
	nc *nats.Conn
	m  sync.RWMutex
)

func InitNATS(o *no.Options) error {
	m.Lock()
	defer m.Unlock()
	if nc != nil {
		old := nc
		defer old.Drain()
	}
	if o == nil {
		return nil
	}
	var err error
	nc, err = NewConnection(o)
	return err
}

func Connection() (*nats.Conn, error) {
	m.RLock()
	defer m.RUnlock()

	if nc == nil {
		return nil, errors.New("NATS connection not initialized")
	}

	return nc, nil
}

// NewConnection creates a new NATS connection
func NewConnection(o *no.Options) (nc *nats.Conn, err error) {
	hostname, _ := os.Hostname()
	opts := []nats.Option{
		nats.Name(fmt.Sprintf("trickster.%s", hostname)),
		nats.MaxReconnects(-1),
		nats.ErrorHandler(errorHandler),
		nats.ReconnectHandler(reconnectHandler),
		nats.DisconnectErrHandler(disconnectHandler),
		// nats.UseOldRequestStyle(),
	}

	if _, err := os.Stat(o.CredPath); os.IsNotExist(err) {
		if _, e2 := os.Stat(o.PasswordPath); e2 == nil {
			password, err := os.ReadFile(o.PasswordPath)
			if err != nil {
				return nil, err
			}
			opts = append(opts, nats.UserInfo(o.Username, string(password)))
		}
	} else {
		opts = append(opts, nats.UserCredentials(o.CredPath))
	}

	//if os.Getenv("NATS_CERTIFICATE") != "" && os.Getenv("NATS_KEY") != "" {
	//	opts = append(opts, nats.ClientCert(os.Getenv("NATS_CERTIFICATE"), os.Getenv("NATS_KEY")))
	//}
	//
	//if os.Getenv("NATS_CA") != "" {
	//	opts = append(opts, nats.RootCAs(os.Getenv("NATS_CA")))
	//}

	// initial connections can error due to DNS lookups etc, just retry, eventually with backoff
	ctx, cancel := context.WithTimeout(context.Background(), natsConnectionTimeout)
	defer cancel()

	ticker := time.NewTicker(natsConnectionRetryInterval)
	for {
		select {
		case <-ticker.C:
			nc, err := nats.Connect(o.Address, opts...)
			if err == nil {
				return nc, nil
			}
			klog.V(5).InfoS("failed to connect to event receiver", "error", err)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// called during errors subscriptions etc
func errorHandler(nc *nats.Conn, s *nats.Subscription, err error) {
	if s != nil {
		klog.V(5).Infof("error in event receiver connection: %s: subscription: %s: %s", nc.ConnectedUrl(), s.Subject, err)
		return
	}
	klog.V(5).Infof("Error in event receiver connection: %s: %s", nc.ConnectedUrl(), err)
}

// called after reconnection
func reconnectHandler(nc *nats.Conn) {
	klog.V(5).Infof("Reconnected to %s", nc.ConnectedUrl())
}

// called after disconnection
func disconnectHandler(nc *nats.Conn, err error) {
	if err != nil {
		klog.V(5).Infof("Disconnected from event receiver due to error: %v", err)
	} else {
		klog.V(5).Infof("Disconnected from event receiver")
	}
}
