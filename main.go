// enum-dns project main.go
//
// A simple ENUM dns server
//
// This DNS server replies to ENUM query using a customizable
// backend implementation. The difference with other general
// purpose DNS systems is that it allows to answer enum queries
// using custom implementations that can for instance answer
// queries based on number ranges.
//
// Author: Hadrien Kohl <hadrien.kohl@gmail.com>
//
//

package main

import (
	"enum-dns/enum"
	"enum-dns/enum/backend/memory"
	"enum-dns/enum/rest"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/miekg/dns"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	Info    *log.Logger
	Error   *log.Logger
	Warning *log.Logger
	Trace   *log.Logger

	backend enum.Backend
	domain  string
	Config  *Configuration
)

type Configuration struct {
	suffixes          []string
	address           string
	defaultService    string
	defaultPreference uint16
	domain            string
}

func main() {

	// TODO: Create configuration from file.
	Config = &Configuration{
		suffixes:          nil,
		defaultService:    "E2U+sip",
		defaultPreference: 100,
		address:           "127.0.0.1:5354",
		domain:            "e164.arpa.",
	}

	// Initialize the loggers.
	Info = log.New(os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(os.Stdout,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(os.Stderr,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Trace = log.New(os.Stderr,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	// Make sure domain is FQDN
	domain = dns.Fqdn(Config.domain)

	var err error
	// Static implementation for now.
	//backend = enum.NewStaticBackend()

	// Mysql implementation
	//backend, err = enum.NewMysqlBackend("mysql", "enum:j8v6xkaK@tcp(127.0.0.1:3307)/enum")

	// BoltDB implementation
	//backend, err = bolt.NewBoltDBBackend("/home/hadrien/enum.bolt")
	//defer backend.Close()
	if err != nil {
		Error.Fatalf("dns: could not connect to the database: %v", err)
	}

	// Memory
	backend, err = memory.NewMemoryBackend()
	if err != nil {
		Error.Fatalf("dns: could not connect to the database: %v", err)
	}
	defer backend.Close()

	backend.PushRange(enum.NumberRange{
		Lower: 4740000000,
		Upper: 4749999999,
		Records: []enum.Record{
			{Regexp: "!^(.*)$!sip:\\@mobile!"},
		},
	})
	backend.PushRange(enum.NumberRange{
		Lower: 4790000000,
		Upper: 4799999999,
		Records: []enum.Record{
			{Regexp: "!^(.*)$!sip:\\@mobile!"},
		},
	})
	backend.PushRange(enum.NumberRange{
		Lower: 47580000000000,
		Upper: 47589999999999,
		Records: []enum.Record{
			{Regexp: "!^(.*)$!sip:\\@mobile!"},
		},
	})
	backend.PushRange(enum.NumberRange{
		Lower: 4759000000,
		Upper: 4759999999,
		Records: []enum.Record{
			{Regexp: "!^(.*)$!sip:\\@mobile!"},
		},
	})

	Info.Printf("Starting enum dns on %v", Config.address)
	server := &dns.Server{Addr: Config.address, Net: "udp"}

	go func() {
		handler := rest.CreateHttpHandlerFor(&backend,
			http.FileServer(
				http.Dir("/home/hadrien/Projects/Go/src/enum-dns/ui/dist/"),
			),
		)

		if err := http.ListenAndServe(":8080", handler); err != nil {
			Error.Fatalf("dns: error starting http server: %s", err)
		}
	}()

	dns.HandleFunc(dns.Fqdn(Config.domain), handleRequest)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			Error.Fatalf("dns: error starting udp server: %v", err)
		}
	}()

	// Wait for signal or error from the dns server.
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-sig:
			// Fatalf calls os.Exit(1)
			Error.Fatalf("Signal (%d) received, stopping\n", s)
		}
	}
}

// Create a new answer for the message.
func NewAnswerForRequest(message *dns.Msg) *dns.Msg {
	answer := new(dns.Msg)
	answer.SetReply(message)
	answer.Authoritative = true
	answer.RecursionAvailable = false
	return answer
}

// Extract the E164 part of an ENUM query.
func extractE164FromName(name string, domain string) (number uint64, err error) {
	numberprefix := strings.TrimSuffix(name, domain)
	if len(numberprefix) == len(name) {
		err = errors.New("The domain '" + domain + "' was not present in the name '" + name + "'")
	} else {
		number, err = enum.ConvertEnumToInt(numberprefix)
	}
	return
}

// Create a dns answer with
func createAnswer(request *dns.Msg) (answer *dns.Msg, err error) {

	if len(request.Question) != 1 {
		err = errors.New("Received more than on question")
		return
	}

	question := request.Question[0]
	if question.Qtype != dns.TypeNAPTR {
		err = errors.New("Received an unsupported query type '" + dns.Type(question.Qtype).String() + "'")
		return
	}

	var number uint64
	if number, err = extractE164FromName(question.Name, domain); err != nil {
		return
	}

	var numberrange enum.NumberRange
	ranges, err := backend.RangesBetween(number, number, 1)
	if err != nil || len(ranges) != 1 {
		return
	}
	numberrange = ranges[0]

	answer = NewAnswerForRequest(request)

	// Create and populate the naptr answer.
	for _, record := range numberrange.Records {
		naptr := new(dns.NAPTR)
		naptr.Hdr = dns.RR_Header{Name: question.Name, Rrtype: question.Qtype, Class: question.Qclass, Ttl: 0}
		naptr.Regexp = record.Regexp

		naptr.Preference = Config.defaultPreference // 1
		naptr.Service = Config.defaultService       //"E2U+sip"

		// Always terminal rule.
		naptr.Flags = "u"
		naptr.Order = 1
		naptr.Replacement = "."

		answer.Answer = append(answer.Answer, naptr)
	}

	return

}

func handleRequest(writer dns.ResponseWriter, request *dns.Msg) {

	defer func(s time.Time) {
		Trace.Printf("dns request for %v (%s) (%v) from client %s (%s)",
			request.Question[0], "udp", time.Now().Sub(s), writer.RemoteAddr().String(),
			writer.RemoteAddr().Network())
	}(time.Now())

	if answer, err := createAnswer(request); err == nil {

		if answer != nil {
			if err := writer.WriteMsg(answer); err != nil {
				Error.Printf("error sending answer: %v", err)
			}
			return
		} else {
			Trace.Printf("no result found for %s", request.Question[0])
			notfound := &dns.Msg{}
			notfound.SetReply(request)
			notfound.SetRcode(request, dns.RcodeSuccess)
			writer.WriteMsg(notfound)
		}
	} else {
		Error.Printf("[ERR] Error getting the answer: %v", err)
		error := &dns.Msg{}
		error.SetReply(request)
		error.SetRcode(request, dns.RcodeServerFailure)
		writer.WriteMsg(error)
	}

}
