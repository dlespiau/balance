package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dlespiau/balance"
	k8s "github.com/dlespiau/balance/pkg/kubernetes"
)

func proxyDirector(req *http.Request) {}

func proxyTransport(keepAlive bool) http.RoundTripper {
	return &http.Transport{
		DisableKeepAlives: !keepAlive,

		// Rest are from http.DefaultTransport
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

type proxyStats struct {
	sync.Mutex
	// Maps key:endpoint -> number of requests
	requests map[string]int
}

type proxy struct {
	balancer  balance.LoadBalancer
	header    string
	reverse   httputil.ReverseProxy
	noForward bool
	stats     proxyStats
}

func (p *proxy) printStats() {
	type stat struct {
		key   string
		value int
	}
	var results []stat

	p.stats.Lock()
	for key, requests := range p.stats.requests {
		results = append(results, stat{key, requests})
	}
	p.stats.requests = make(map[string]int)
	p.stats.Unlock()

	sort.Slice(results, func(i, j int) bool {
		return results[i].key < results[j].key
	})
	for i := range results {
		fmt.Printf("%s: %d\n", results[i].key, results[i].value)
	}
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get(p.header)
	if key == "" {
		http.Error(w, fmt.Sprintf("unable to find %s header", p.header), http.StatusBadRequest)
		return
	}

	endpoint := p.balancer.Get(key)

	// Update stats. XXX make it faster!
	p.stats.Lock()
	p.stats.requests[fmt.Sprintf("%s-%s", key, endpoint.Key())]++
	p.stats.Unlock()

	// Forward request to endpoint.
	if p.noForward {
		p.balancer.Put(endpoint)
		return
	}

	r.Host = endpoint.Key()
	r.URL.Host = endpoint.Key()
	r.URL.Scheme = "http"

	p.reverse.ServeHTTP(w, r)

	p.balancer.Put(endpoint)

}

type options struct {
	kubeconfig  string
	namespace   string
	service     string
	listen      string
	header      string
	keepAlive   bool
	noForward   bool
	method      string
	boundedLoad struct {
		loadFactor float64
	}
}

func makeLoadBalancer(opts *options) balance.LoadBalancer {
	switch opts.method {
	case "consistent":
		return balance.NewConsistent(balance.ConsistentConfig{})
	case "bounded-load":
		return balance.NewConsistent(balance.ConsistentConfig{
			LoadFactor: opts.boundedLoad.loadFactor,
		})
	default:
		return nil
	}
}

func main() {
	opts := options{}
	flag.StringVar(&opts.kubeconfig, "k8s.kubeconfig", "", "(optional) absolute path to the kubeconfig file")
	flag.StringVar(&opts.namespace, "k8s.namespace", "default", "namespace of the service to load balance")
	flag.StringVar(&opts.service, "k8s.service", "", "name of the service to load balance")
	flag.StringVar(&opts.listen, "proxy.listen", ":8081", "address the proxy should listen on")
	flag.StringVar(&opts.header, "proxy.header", "X-Affinity", "name of the HTTP header taken as input")
	flag.BoolVar(&opts.keepAlive, "proxy.keep-alive", true, "whether the proxy should keep its connections to endpoints alive")
	flag.BoolVar(&opts.noForward, "proxy.no-forward", false, "don't forward request downstream (debug)")
	flag.StringVar(&opts.method, "proxy.method", "bounded-load", "which load balancing method should be used (one of consistent, bounded-load)")
	flag.Float64Var(&opts.boundedLoad.loadFactor, "proxy.bounded-load.load-factor", 1.25, "spread of the maximum load from the average load")
	flag.Parse()

	log.SetLevel(log.DebugLevel)

	if opts.kubeconfig == "" {
		opts.kubeconfig = os.Getenv("KUBECONFIG")
	}

	client, err := k8s.NewClientWithConfig(&k8s.ClientConfig{
		Kubeconfig: opts.kubeconfig,
	})
	if err != nil {
		log.Fatal(err)
	}

	balancer := makeLoadBalancer(&opts)
	if balancer == nil {
		log.Fatal("unknown load balancing method: %s", opts.method)
	}

	watcher := balance.EndpointWatcher{
		Client: client,
		Service: balance.Service{
			Namespace: opts.namespace,
			Name:      opts.service,
			Port:      "8080",
		},
		Receiver: balancer,
	}

	watcher.Start(make(<-chan interface{}))

	proxy := &proxy{
		noForward: opts.noForward,
		balancer:  balancer.(balance.LoadBalancer),
		header:    opts.header,
		reverse: httputil.ReverseProxy{
			Director:  proxyDirector,
			Transport: proxyTransport(opts.keepAlive),
		},
		stats: proxyStats{
			requests: make(map[string]int),
		},
	}

	// Print stats on SIGUSR1
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGUSR1)
	go func() {
		for range signals {
			proxy.printStats()
		}
	}()

	log.Fatal(http.ListenAndServe(opts.listen, proxy))
}
