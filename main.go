package main

import (
    "log"
    "net"
    "net/http"
    "runtime"
    "time"

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
    // Use all available CPU cores
    runtime.GOMAXPROCS(runtime.NumCPU())

    // Create a cache sized for optimal memory usage on t3.small (2GB RAM)
    // Assuming average key+value size of 512 bytes, reserve 1.5GB for cache
    maxItems := 3_000_000 // Approximately 1.5GB of data

    c := cache.NewCache(maxItems)
    handler := api.NewHandler(c)

    // Optimize server settings
    server := &http.Server{
        Addr:         ":7171",
        Handler:      handler.Router(), // Use custom router for better performance
        ReadTimeout:  2 * time.Second,  // Reduced timeouts
        WriteTimeout: 2 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Enable TCP keep-alives
    ln, err := net.Listen("tcp", ":7171")
    if err != nil {
        log.Fatal(err)
    }
    tcpListener := ln.(*net.TCPListener)
    
    log.Println("Starting optimized key-value cache server on port 7171...")
    if err := server.Serve(tcpKeepAliveListener{tcpListener}); err != nil {
        log.Fatal(err)
    }
}
