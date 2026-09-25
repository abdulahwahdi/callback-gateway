// Package api holds the service's API contracts: the JSON schemas used to
// validate request payloads, and the OpenAPI (Swagger) documentation served
// at /openapi.yaml with an interactive UI at /docs. Everything is embedded
// into the binary so no extra files need to ship alongside it.
package api

import (
	"embed"
	"net/http"

	"github.com/golangid/candi/codebase/interfaces"
)

// JSONSchema schema sources
//
//go:embed all:jsonschema
var JSONSchema embed.FS

//go:embed openapi.yaml
var openAPISpec []byte

const uiPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>webhook-middleware API docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>body { margin: 0; }</style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: '/openapi.yaml',
        dom_id: '#swagger-ui',
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
        layout: 'BaseLayout',
      });
    };
  </script>
</body>
</html>`

// MountDocs registers the /docs (Swagger UI) and /openapi.yaml (raw spec) routes.
func MountDocs(root interfaces.RESTRouter) {
	root.GET("/openapi.yaml", func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/yaml")
		rw.Write(openAPISpec)
	})
	root.GET("/docs", func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.Write([]byte(uiPage))
	})
}
