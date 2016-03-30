package rest

import (
	"encoding/json"
	"enum-dns/enum"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type HttpEndpoint struct {
	backend enum.Backend
	handler http.Handler
}

const PUT_LIMIT = 1048576
const RETURN_LIMIT = 100

func CreateHttpHandlerFor(b *enum.Backend, ui http.Handler) http.Handler {

	r := mux.NewRouter().StrictSlash(true)

	h := HttpEndpoint{
		backend: *b,
	}

	numRe := "[1-9][0-9]{0,14}"

	api := r.PathPrefix("/api/").Subrouter()
	api.Path("/interval/{from:"+numRe+"}:{to:"+numRe+"}").Methods("PUT", "GET").HandlerFunc(h.GetAndEditHandler)
	api.Path("/interval").Methods("GET").HandlerFunc(h.SearchHandler)

	r.PathPrefix("/ui").Handler(http.StripPrefix("/ui", ui))

	h.handler = r

	return &h
}

func (h *HttpEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		log.Printf(
			"%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	}()

	h.handler.ServeHTTP(w, r)
}

// Parse and return the limit and order variables from request.
func Pagination(vars url.Values) (after, before uint64, limit int64, err error) {
	if l := vars.Get("limit"); l != "" {
		limit, err = strconv.ParseInt(l, 10, 32)
	}
	if a := vars.Get("after"); a != "" {
		after, err = strconv.ParseUint(a, 10, 64)
	}
	if b := vars.Get("before"); b != "" {
		before, err = strconv.ParseUint(b, 10, 64)
	}
	return
}

// Parse prefix and compute corresponding from and to.
func Prefix(vars url.Values) (from, to uint64, err error) {

	if p := vars.Get("prefix"); p == "" {
		return 0, 0, nil
	}

	prefix, err := strconv.ParseUint(vars.Get("prefix"), 10, 64)
	if err != nil {
		return 0, 0, err
	}

	from, err = enum.PrefixToE164(prefix)
	if err != nil {
		return 0, 0, err
	}

	to, _ = enum.PrefixToE164(prefix + 1)

	return
}

// Extract and validate from and to variables.
func FromAndTo(vars url.Values) (from, to uint64, err error) {

	hasFrom := vars.Get("from") != ""
	hasTo := vars.Get("to") != ""

	if !hasFrom && !hasTo {
		return 0, 0, nil
	}

	if hasFrom != hasTo {
		return 0, 0, errors.New("missing from or to.")
	}

	from, err = strconv.ParseUint(vars.Get("from"), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("imposible to parse %s", vars.Get("from"))
	}

	to, err = strconv.ParseUint(vars.Get("to"), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("imposible to parse %s", vars.Get("to"))
	}

	return

}

func WriteError(w http.ResponseWriter, err error, status int) bool {
	if err == nil {
		return false
	}
	w.WriteHeader(status)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	return true
}

func (h *HttpEndpoint) GetAndEditHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	from, err := strconv.ParseUint(vars["from"], 10, 64)
	to, err := strconv.ParseUint(vars["to"], 10, 64)
	if WriteError(w, err, http.StatusBadRequest) {
		return
	}

	var insert enum.NumberRange
	if err := json.NewDecoder(io.LimitReader(r.Body, PUT_LIMIT)).Decode(&insert); err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	results, err := h.backend.RangesBetween(from, to, 2)
	if WriteError(w, err, http.StatusInternalServerError) {
		return
	}

	if len(results) != 1 {
		WriteError(w, nil, http.StatusNotFound)
		return
	}

	if r.Method == "PUT" {
		_, err = h.backend.PushRange(insert)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusCreated)
	} else {
		json.NewEncoder(w).Encode(results[0])
	}
}

func (h *HttpEndpoint) SearchHandler(w http.ResponseWriter, r *http.Request) {

	vars := r.URL.Query()

	hasFrom := vars.Get("from") != ""
	hasTo := vars.Get("to") != ""
	hasPrefix := vars.Get("prefix") != ""
	if hasPrefix && (hasFrom || hasTo) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errors.New("cannot use prefix with from or to").Error()))
		return
	}

	var from, to uint64
	var err error
	if hasPrefix {
		from, to, err = Prefix(vars)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
	} else {
		from, to, err = FromAndTo(vars)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
	}

	after, before, limit, err := Pagination(vars)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if after != 0 {
		if !(from <= after && after < to) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(errors.New("after value outside from and to").Error()))
			return
		}
		from = after
	}
	if before != 0 {
		if !(from < before && before <= to) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(errors.New("before value outside from and to").Error()))
			return
		}
		to = before
	}

	if limit == 0 {
		limit = RETURN_LIMIT
	}

	if from > to {
		to, from = from, to
		limit = -limit
	}

	results, err := h.backend.RangesBetween(from, to, int(limit))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		//json.NewEncoder(w).Encode(err)
		log.Panic(err.Error())
	} else {
		json.NewEncoder(w).Encode(results)
	}
}
