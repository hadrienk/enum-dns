package dns

import (
	"enum-dns/enum"
	"errors"
	"github.com/miekg/dns"
	"log"
	"regexp"
	"strings"
	"time"
)

var reg = regexp.MustCompile("^([0-9]|\\.[0-9])+")

type ENUMHandler struct {
	Backend *enum.Backend
	Info    *log.Logger
	Error   *log.Logger
	Warning *log.Logger
	Trace   *log.Logger
}

func (h ENUMHandler) ServeDNS(writer dns.ResponseWriter, request *dns.Msg) {

	defer func(s time.Time) {
		h.Trace.Printf("dns request for %v (%s) (%v) from client %s (%s)",
			request.Question[0], "udp", time.Now().Sub(s), writer.RemoteAddr().String(),
			writer.RemoteAddr().Network())
	}(time.Now())

	if answer, err := h.createAnswer(request); err == nil {

		if answer == nil {
			h.Trace.Printf("no result found for %s", request.Question[0])
			notfound := &dns.Msg{}
			notfound.SetReply(request)
			notfound.SetRcode(request, dns.RcodeSuccess)
			writer.WriteMsg(notfound)
			return
		}

		if err := writer.WriteMsg(answer); err != nil {
			h.Error.Printf("error sending answer: %v", err)
		}

	} else {
		h.Error.Printf("[ERR] Error getting the answer: %v", err)
		error := &dns.Msg{}
		error.SetReply(request)
		error.SetRcode(request, dns.RcodeServerFailure)
		writer.WriteMsg(error)
	}

}

// Instantiate a new answer as a reply of the passed message.
func (h *ENUMHandler) answerForRequest(message *dns.Msg) *dns.Msg {
	answer := new(dns.Msg)
	answer.SetReply(message)
	answer.Authoritative = true
	answer.RecursionAvailable = false
	return answer
}

// Extract the E164 part of an ENUM query. Ex: 1.2.3.4.domain -> 4321.
func extractE164FromName(name string) (number uint64, err error) {
	numberprefix := strings.Join(reg.FindAllString(name, -1), "")
	number, err = enum.ConvertEnumToInt(numberprefix)
	number, err = enum.PrefixToE164(number)
	return
}

// Create a dns answer with
func (h *ENUMHandler) createAnswer(request *dns.Msg) (answer *dns.Msg, err error) {

	if len(request.Question) != 1 {
		err = errors.New("Received more than one question")
		return
	}

	question := request.Question[0]
	if question.Qtype != dns.TypeNAPTR {
		err = errors.New("Received an unsupported query type '" + dns.Type(question.Qtype).String() + "'")
		return
	}

	var number uint64
	if number, err = extractE164FromName(question.Name); err != nil {
		return
	}

	var numberrange enum.NumberRange
	h.Trace.Printf("backend.RangesBetween(%d, %d, 1)", number, number)
	ranges, err := (*h.Backend).RangesBetween(number, number, 1)
	if err != nil || len(ranges) != 1 {
		return
	}
	numberrange = ranges[0]

	answer = h.answerForRequest(request)

	// Create and populate the NAPTR answers.
	for _, record := range numberrange.Records {
		naptr := new(dns.NAPTR)
		naptr.Hdr = dns.RR_Header{Name: question.Name, Rrtype: question.Qtype, Class: question.Qclass, Ttl: 0}
		naptr.Regexp = record.Regexp

		naptr.Preference = record.Preference
		naptr.Service = record.Service
		naptr.Flags = record.Flags
		naptr.Order = record.Order
		naptr.Replacement = record.Replacement

		answer.Answer = append(answer.Answer, naptr)
	}

	return

}
