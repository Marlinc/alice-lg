package http

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/alice-lg/alice-lg/client"
	"github.com/alice-lg/alice-lg/pkg/config"
)

// Web Client
// Handle assets and client app preprarations

// Prepare client HTML:
// Set paths and add version to assets.
func webPrepareClientHTML(html string) string {
	status, _ := NewAppStatus()

	// Replace paths and tags
	rewriter := strings.NewReplacer(
		// Paths
		"js/", "/static/js/",
		"css/", "/static/css/",

		// Tags
		"APP_VERSION", status.Version,
	)
	html = rewriter.Replace(html)
	return html
}

// Register assets handler and index handler
// at /static and /
func webRegisterAssets(cfg *config.Config, router *httprouter.Router) error {
	log.Println("Preparing and installing assets")

	// Prepare client html: Rewrite paths
	indexHTMLData, err := client.Assets.ReadFile("build/index.html")
	if err != nil {
		return err
	}
	indexHTML := string(indexHTMLData) // TODO: migrate to []byte

	theme := NewTheme(cfg.UI.Theme)
	err = theme.RegisterThemeAssets(router)
	if err != nil {
		log.Println("Warning:", err)
	}

	// Update paths
	indexHTML = webPrepareClientHTML(indexHTML)

	// Register static assets
	router.Handler("GET", "/static/*path", client.AssetsHTTPHandler("/static"))

	// Rewrite paths
	// Serve index html as root...
	router.GET("/",
		func(res http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
			// Include theme, we need to update the
			// hashes on reload, so we can check if the theme has
			// changed without restarting the app
			themedHTML := theme.PrepareClientHTML(indexHTML)
			io.WriteString(res, themedHTML)
		})

	// ...and all alice related paths aswell
	alicePaths := []string{
		"/routeservers/*path",
		"/search/*path",
	}
	for _, path := range alicePaths {
		// respond with app html
		router.GET(path,
			func(res http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
				// ditto here
				themedHTML := theme.PrepareClientHTML(indexHTML)
				io.WriteString(res, themedHTML)
			})
	}

	// ...install a catch all for /alice for graceful backwards compatibility
	router.GET("/alice/*path",
		func(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
			http.Redirect(res, req, "/", 301)
		})

	return nil
}
