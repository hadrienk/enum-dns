package rest

import (
	"encoding/json"
	"enum-dns/enum"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type HttpEndpoint struct {
	backend enum.Backend
	handler http.Handler
}

func CreateHttpHandlerFor(b *enum.Backend) http.Handler {

	router := mux.NewRouter().StrictSlash(true)

	http := HttpEndpoint{
		backend: *b,
	}

	numRe := "[1-9][0-9]{0,14}"
	prefixRe := "[1-9][0-9]{0,13}"
	limitRe := "-?[1-9][0-9]*"

	router.HandleFunc(fmt.Sprintf("/interval/{prefix:%s}", prefixRe), http.IntervalForPrefixHandler)
	router.HandleFunc(fmt.Sprintf("/interval/{from:%s}", numRe), http.IntervalForNumberHandler)
	router.HandleFunc(fmt.Sprintf("/interval/{prefix:%s},{limit:%s}", prefixRe, limitRe), http.IntervalForPrefixHandler)
	router.HandleFunc(fmt.Sprintf("/interval/{from:%s},{to:%s}", numRe, numRe), http.IntervalForNumberHandler)
	router.HandleFunc(fmt.Sprintf("/interval/{from:%s},{to:%s},{limit:%s}", numRe, numRe, limitRe), http.IntervalForNumberHandler)

	http.handler = router

	return &http
}

func (h *HttpEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

func (h *HttpEndpoint) IntervalForPrefixHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	prefix, err := strconv.ParseUint(vars["prefix"], 10, 64)
	limit, err := strconv.ParseInt(vars["limit"], 10, 32)

	if limit == 0 {
		limit = 100
	}

	from, err := enum.PrefixToE164(prefix)
	to, err := enum.PrefixToE164(prefix + 1)

	results, err := h.backend.RangesBetween(from, to, int(limit))

	if err != nil {
		json.NewEncoder(w).Encode(err)
	} else {
		json.NewEncoder(w).Encode(results)
	}
}

func (h *HttpEndpoint) IntervalForNumberHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	from, err := strconv.ParseUint(vars["from"], 10, 64)
	to, err := strconv.ParseUint(vars["to"], 10, 64)
	limit, err := strconv.ParseInt(vars["limit"], 10, 32)

	results, err := h.backend.RangesBetween(from, to, int(limit))

	if err != nil {
		json.NewEncoder(w).Encode(err)
	} else {
		json.NewEncoder(w).Encode(results)
	}
}
