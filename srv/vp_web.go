package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/raft"
)

const (
	MaxUploadSizeMb    = 8 << 20
	HContentType       = "Content-Type"
	VApplicationJson   = "application/json"
	VMultiPartFormData = "multipart/form-data"
	VTextPlain         = "text/plain"
	VTextHtml          = "text/html"
	VTextJavascript    = "text/javascript"
	VtextCss           = "text/css"
	PKey               = "key"
	PValue             = "value"
	PPeerId            = "peerId"
	PPrefix            = "prefix"
	POffset            = "offset"
	PPageSize          = "pageSize"
	Get                = "GET"
	Post               = "POST"
)

const (
	RKvList  = "/kv/list"
	RKvGet   = "/kv/get"
	RKvSet   = "/kv/set"
	RKvDel   = "/kv/del"
	RRfJoin  = "/raft/join"
	RRfLeave = "/raft/leave"
	RRfStat  = "/raft/status"
)

type Peer struct {
	Raft string
	Http string
}

type VpResponse struct {
	Data  interface{}
	Error string
}

type WebHandler struct {
	s *Server
}

func NewWebHandler(s *Server) *WebHandler {
	return &WebHandler{s: s}
}

/* ==================================================================================
                            Utility functions
================================================================================== */

func onError(w http.ResponseWriter, err error, statuscode int) {
	log.Error("Web handler error", "cause", err)
	res := VpResponse{Data: nil, Error: err.Error()}
	j, err := json.Marshal(&res)
	if err == nil {
		w.Header().Set(HContentType, VApplicationJson)
		w.WriteHeader(statuscode)
		w.Write([]byte(j))
	} else {
		w.Header().Set(HContentType, VTextPlain)
		w.WriteHeader(statuscode)
		w.Write([]byte(err.Error()))
	}
}

func onSuccess(w http.ResponseWriter, res *VpResponse, statusCode int) {
	j, err := json.Marshal(&res)
	if err != nil {
		onError(w, err, http.StatusInternalServerError)
	} else {
		w.Header().Set(HContentType, VApplicationJson)
		w.WriteHeader(statusCode)
		w.Write([]byte(j))
	}
}

func readTo(r io.Reader) ([]byte, error) {
	if body, err := ioutil.ReadAll(r); err != nil {
		return nil, err
	} else {
		return body, nil
	}
}

func bodyOf(w http.ResponseWriter, req *http.Request) ([]byte, error) {
	ct := req.Header.Get(HContentType)
	if strings.Contains(ct, VMultiPartFormData) {
		req.ParseMultipartForm(MaxUploadSizeMb)
		file, _, err := req.FormFile(PValue)
		if err != nil {
			return nil, err
		}
		return readTo(file)
	} else {
		req.Body = http.MaxBytesReader(w, req.Body, MaxUploadSizeMb)
		return readTo(req.Body)
	}
}

func onLeaderResponse(w http.ResponseWriter, res *http.Response, err error) {
	if err != nil {
		onError(w, err, http.StatusInternalServerError)
	} else {
		defer res.Body.Close()
		w.WriteHeader(res.StatusCode)
		io.Copy(w, res.Body)
	}
}

func (h *WebHandler) forwardToLeader(w http.ResponseWriter, req *http.Request) {
	var url string
	configFuture := h.s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		onError(w, err, http.StatusInternalServerError)
	} else {
		for _, peer := range configFuture.Configuration().Servers {
			if peer.Address == h.s.raft.Leader() {
				peerRaft, httpPort := parsePeer(string(peer.ID))
				url = fmt.Sprintf("http://%s:%s", strings.Split(peerRaft, ":")[0], httpPort) // TODO this may need to be customized
			}
		}
		if len(url) == 0 {
			onError(w, errors.New("leader not found"), http.StatusBadGateway)
		}
		url = url + req.RequestURI
		log.Info("forwarding request", "from", h.s.peerId, "to", url)
		switch req.Method {
		case Get:
			res, err := http.Get(url)
			onLeaderResponse(w, res, err)
		case Post:
			res, err := http.Post(url, req.Header.Get(HContentType), req.Body)
			onLeaderResponse(w, res, err)
		}
	}
}

/* ==================================================================================
                            Request methods
================================================================================== */

func (h *WebHandler) KeysRequest(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	prefix := req.Form.Get(PPrefix)
	pageSize := req.Form.Get(PPageSize)
	offset := req.Form.Get(POffset)
	if len(offset) == 0 {
		offset = prefix
	}
	var ps int
	if i, err := strconv.Atoi(pageSize); err != nil {
		onError(w, err, http.StatusBadRequest)
		return
	} else if i < 0 {
		ps = 0
	} else {
		ps = i
	}
	if keys, err := h.s.store.KeysOf([]byte(prefix), []byte(offset), uint16(ps)); err != nil {
		onError(w, err, http.StatusInternalServerError)
	} else {
		onSuccess(w, &VpResponse{Data: keys}, http.StatusOK)
	}
}

func (h *WebHandler) GetRequest(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	key := req.Form.Get(PKey)
	data, err := h.s.store.GetData([]byte(key))
	if len(data) == 0 {
		onError(w, fmt.Errorf("key not found: [%s]", key), http.StatusNotFound)
	} else if err != nil {
		onError(w, err, http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func (h *WebHandler) SetRequest(w http.ResponseWriter, req *http.Request) {
	if h.s.raft.State() != raft.Leader {
		h.forwardToLeader(w, req)
	} else {
		req.ParseForm()
		key := req.Form.Get(PKey)
		switch req.Method {
		case Get:
			value := req.Form.Get(PValue)
			if err := h.s.RaftSet(key, []byte(value)); err != nil {
				onError(w, err, http.StatusInternalServerError)
			} else {
				onSuccess(w, &VpResponse{Data: key}, http.StatusOK)
			}
		case Post:
			if value, err := bodyOf(w, req); err != nil {
				onError(w, err, http.StatusBadRequest)
			} else {
				h.s.RaftSet(key, value)
				onSuccess(w, &VpResponse{Data: key}, http.StatusOK)
			}
		}
	}
}

func (h *WebHandler) DeleteRequest(w http.ResponseWriter, req *http.Request) {
	if h.s.raft.State() != raft.Leader {
		h.forwardToLeader(w, req)
	} else {
		req.ParseForm()
		key := req.Form.Get(PKey)
		if err := h.s.RaftDelete(key); err != nil {
			onError(w, err, http.StatusInternalServerError)
		} else {
			onSuccess(w, &VpResponse{Data: key}, http.StatusOK)
		}
	}
}

func (h *WebHandler) RaftStatusRequest(w http.ResponseWriter, r *http.Request) {
	onSuccess(w, &VpResponse{Data: h.s.raft.Stats()}, http.StatusOK)
}

func (h *WebHandler) RaftJoinRequest(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	if err := h.s.RaftJoin(req.FormValue(PPeerId)); err != nil {
		onError(w, err, http.StatusInternalServerError)
	} else {
		onSuccess(w, &VpResponse{Data: h.s.raft.Stats()}, http.StatusCreated)
	}
}

func (h *WebHandler) RaftLeaveRequest(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	if err := h.s.RaftLeave(req.FormValue(PPeerId)); err != nil {
		onError(w, err, http.StatusInternalServerError)
	} else {
		onSuccess(w, &VpResponse{Data: h.s.raft.Stats()}, http.StatusGone)
	}
}
