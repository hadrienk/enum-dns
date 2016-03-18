// Copyright 2016 Hadrien Kohl hadrien.kohl@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"enum-dns/enum"
	"enum-dns/enum/backend/memory"
	enumdns "enum-dns/enum/dns"
	"enum-dns/enum/rest"
	"github.com/miekg/dns"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	viper.SetDefault("dns.address", "127.0.0.1:5354")
	viper.SetDefault("dns.domain", "e164.arpa.")

	// Initialize the loggers.
	Info := log.New(os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning := log.New(os.Stdout,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error := log.New(os.Stderr,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Trace := log.New(os.Stderr,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	// Memory
	backend, err := memory.NewMemoryBackend()
	if err != nil {
		Error.Fatalf("backend: could not start the backend: %v", err)
	}
	defer backend.Close()

	backend.PushRange(enum.NumberRange{
		Lower: 100000000000000,
		Upper: 999999999999999,
		Records: []enum.Record{
			{Regexp: "!^(.*)$!sip:\\@default!", Service: "E2U+sip",
				Preference: 100, Replacement: ".", Order: 10},
		},
	})

	domain := dns.Fqdn(viper.GetString("dns.domain"))
	address := viper.GetString("dns.address")
	dnsHandler := enumdns.ENUMHandler{
		Info: Info, Warning: Warning, Trace: Trace, Error: Error,
		Backend: &backend,
	}
	server := &dns.Server{Addr: address, Net: "udp"}
	dns.Handle(domain, dnsHandler)

	go func() {
		Info.Printf("Starting enum dns server on %v", address)
		if err := server.ListenAndServe(); err != nil {
			Error.Fatalf("dns: error starting udp server: %v", err)
		}
	}()

	go func() {

		handler := rest.CreateHttpHandlerFor(&backend,
			// TODO Check that the directory exists.
			http.FileServer(
				http.Dir("./ui/dist/"),
			),
		)

		if err := http.ListenAndServe(":8080", handler); err != nil {
			Error.Fatalf("http: error starting http server: %s", err)
		}
	}()

	// Wait for signal or error from the dns server.
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-sig:
			server.Shutdown()
			// TODO Handle the http server shutdown as well
			// Fatalf calls os.Exit(1)
			Error.Fatalf("Signal (%d) received, stopping\n", s)
		}
	}
}
