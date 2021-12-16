package main

import (
	_ "embed"
	"net/http"
	"strings"
)

const (
	RUi       = "/ui"
	RIndexJs  = "/ui/vephar.js"
	RIndexCss = "/ui/vephar.css"
	RFavIcon  = "/favicon.ico"
)

const VRootHtml = `
<!DOCTYPE html>
<html>
	<head>
		<base href="/" />
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<link rel="stylesheet" href="/ui/vephar.css" />
	</head>
	<body class="dark">
		<div id="root" />
		<script src="/ui/vephar.js"></script>
		<noscript>
			<!-- Happiness = Reality minus Expectations -->
			<!-- :P -->
		</noscript>
	</body>
</html>
`

//go:embed dist/vephar.js
var vpJs []byte

//go:embed dist/vephar.css
var vpCss []byte

//go:embed dist/favicon.ico
var vpIco []byte

func writeFile(w http.ResponseWriter, data []byte, contentType string) {
	w.Header().Set(HContentType, contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func writeUiRoot(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set(HContentType, VTextHtml)
	w.Write([]byte(VRootHtml))
}

func ResourceHandler(w http.ResponseWriter, r *http.Request) {
	if log.IsDebug() {
		log.Debug(r.URL.Path)
	}
	switch r.URL.Path {
	case RUi:
		writeUiRoot(w)
	case RIndexJs:
		writeFile(w, vpJs, VTextJavascript) // TODO switch between dev/build modes
	case RIndexCss:
		writeFile(w, vpCss, VtextCss)
	case RFavIcon:
		writeFile(w, vpIco, "")
	}
}

func UiHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, RUi) {
		writeUiRoot(w)
	} else {
		w.Header().Set("Location", RUi)
		w.WriteHeader(301)
	}
}
