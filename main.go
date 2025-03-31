package main

import (
    "log"
    "net"
    "net/http"
    "runtime"
    "time"
    "runtime/debug"

    "github.com/OmMishra16/key-value-cache/api"
    "github.com/OmMishra16/key-value-cache/cache"
)

// Add this type definition before the main function
type tcpKeepAliveListener struct {
    *net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
    tc, err := ln.AcceptTCP()
    if err != nil {
        return nil, err
    }
    tc.SetKeepAlive(true)
    tc.SetKeepAlivePeriod(3 * time.Minute)
    return tc, nil
}

func main() {
    // Use all available CPU cores (t3.small has 2 cores)
    runtime.GOMAXPROCS(2)

    // Optimize GC for low latency
    debug.SetGCPercent(50)  // More aggressive GC
    
    // Optimize cache size for t3.small
    maxItems := 2_000_000  // Reduced to ensure we stay well under memory limits
    
    c := cache.NewCache(maxItems)
    handler := api.NewHandler(c)

    // Optimize server settings
    server := &http.Server{
        Addr:           ":7171",
        Handler:        handler.Router(),
        ReadTimeout:    2 * time.Second,    // Increased timeout
        WriteTimeout:   2 * time.Second,
        IdleTimeout:    120 * time.Second,
        MaxHeaderBytes: 1 << 16,
    }

    // Enable TCP keep-alives with shorter period
    ln, err := net.Listen("tcp", ":7171")
    if err != nil {
        log.Fatal(err)
    }
    tcpListener := ln.(*net.TCPListener)
    
    log.Printf("Starting cache server with %d max items\n", maxItems)
    if err := server.Serve(tcpKeepAliveListener{tcpListener}); err != nil {
        log.Fatal(err)
    }
}
