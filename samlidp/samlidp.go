// Package samlidp a rudimentary SAML identity provider suitable for
// testing or as a starting point for a more complex service.
package samlidp

import (
	"crypto"
	"crypto/x509"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"text/template"

	"github.com/zenazn/goji/web"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/logger"
)

const indexPageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Kushal's Identity Provider for Teaching</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Lora:wght@400;600&amp;family=Source+Sans+3:wght@400;500;600&amp;display=swap" rel="stylesheet">
    <style>
        *, *::before, *::after { margin: 0; padding: 0; box-sizing: border-box; }
        :root {
            --navy: #002b5c; --navy-light: #003d82; --blue: #1d6fa5;
            --sky: #e8f1f8; --white: #ffffff; --text: #1c2536;
            --muted: #5a6678; --light: #8892a4; --border: #dde3ea;
        }
        body {
            font-family: 'Source Sans 3', system-ui, sans-serif;
            color: var(--text); background: var(--white);
            min-height: 100vh; display: flex; flex-direction: column;
        }
        .topbar { height: 4px; background: linear-gradient(90deg, var(--navy), var(--blue)); }
        .header { padding: 2.5rem 1rem 1rem; text-align: center; }
        .header h1 {
            font-family: 'Lora', Georgia, serif; font-weight: 600;
            font-size: 1.4rem; color: var(--navy); letter-spacing: -0.02em;
        }
        .main {
            flex: 1; display: flex; align-items: flex-start;
            justify-content: center; padding: 1rem 1.5rem 3rem;
        }
        .card {
            width: 100%; max-width: 480px;
            border: 1px solid var(--border); border-radius: 8px;
            box-shadow: 0 1px 3px rgba(0,43,92,.05), 0 6px 16px rgba(0,43,92,.03);
            padding: 2.25rem 2rem;
        }
        .card-heading {
            font-family: 'Lora', Georgia, serif; font-weight: 600;
            font-size: 1.25rem; color: var(--navy); margin-bottom: 0.5rem;
        }
        .card-text {
            font-size: 0.92rem; color: var(--muted);
            line-height: 1.6; margin-bottom: 1.5rem;
        }
        .nav-links { display: flex; flex-direction: column; gap: 0.5rem; margin-bottom: 0.25rem; }
        .nav-link {
            display: flex; align-items: center; gap: 0.6rem;
            padding: 0.7rem 0.85rem; border: 1px solid var(--border);
            border-radius: 6px; text-decoration: none;
            color: var(--navy); font-weight: 500; font-size: 0.9rem;
            transition: background 0.15s, border-color 0.15s;
        }
        .nav-link:hover { background: var(--sky); border-color: var(--blue); }
        .nav-link svg { flex-shrink: 0; color: var(--blue); }
        .nav-link-desc { font-size: 0.8rem; color: var(--muted); font-weight: 400; }
        .sep { border: none; border-top: 1px solid var(--border); margin: 1.5rem 0; }
        .resource {
            display: flex; align-items: center; gap: 0.65rem;
            padding: 0.75rem 0.85rem; background: var(--sky);
            border-radius: 6px; text-decoration: none; transition: background 0.15s;
        }
        .resource:hover { background: #dae7f2; }
        .resource svg { flex-shrink: 0; color: var(--navy); }
        .resource-text { line-height: 1.35; }
        .resource-title { font-size: 0.85rem; font-weight: 600; color: var(--navy); }
        .resource-desc { font-size: 0.78rem; color: var(--muted); }
        .footer { text-align: center; padding: 1.25rem; font-size: 0.78rem; color: var(--light); }
    </style>
</head>
<body>
    <div class="topbar"></div>
    <div class="header"><h1>Kushal's Identity Provider for Teaching</h1></div>
    <div class="main">
        <div class="card">
            <div class="card-heading">Welcome</div>
            <p class="card-text">
                This is a SAML 2.0 Identity Provider operated by Kushal
                for testing and development purposes.
            </p>
            <div class="nav-links">
                <a href="{{.MetadataURL}}" class="nav-link">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
                    <span>
                        SAML Metadata<br>
                        <span class="nav-link-desc">View the IdP metadata XML document</span>
                    </span>
                </a>
            </div>
            <hr class="sep">
            <a href="https://kushaldas.in/learningsaml/" class="resource" target="_blank" rel="noopener">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
                <span class="resource-text">
                    <span class="resource-title">Learning SAML</span><br>
                    <span class="resource-desc">An introduction to SAML for developers</span>
                </span>
            </a>
        </div>
    </div>
    <div class="footer"></div>
</body>
</html>`

// Options represent the parameters to New() for creating a new IDP server
type Options struct {
	URL           url.URL
	Key           crypto.PrivateKey
	Signer        crypto.Signer
	Logger        logger.Interface
	Certificate   *x509.Certificate
	Store         Store
	Intermediates []*x509.Certificate
}

// Server represents an IDP server. The server provides the following URLs:
//
//	/metadata     - the SAML metadata
//	/sso          - the SAML endpoint to initiate an authentication flow
//	/login        - prompt for a username and password if no session established
//	/login/:shortcut - kick off an IDP-initiated authentication flow
//	/services     - RESTful interface to Service objects
//	/users        - RESTful interface to User objects
//	/sessions     - RESTful interface to Session objects
//	/shortcuts    - RESTful interface to Shortcut objects
type Server struct {
	http.Handler
	idpConfigMu      sync.RWMutex // protects calls into the IDP
	logger           logger.Interface
	serviceProviders map[string]*saml.EntityDescriptor
	IDP              saml.IdentityProvider // the underlying IDP
	Store            Store                 // the data store
}

// New returns a new Server
func New(opts Options) (*Server, error) {
	metadataURL := opts.URL
	metadataURL.Path += "/metadata"
	ssoURL := opts.URL
	ssoURL.Path += "/sso"
	logr := opts.Logger
	if logr == nil {
		logr = logger.DefaultLogger
	}

	s := &Server{
		serviceProviders: map[string]*saml.EntityDescriptor{},
		IDP: saml.IdentityProvider{
			Key:         opts.Key,
			Signer:      opts.Signer,
			Logger:      logr,
			Certificate: opts.Certificate,
			MetadataURL: metadataURL,
			SSOURL:      ssoURL,
		},
		logger: logr,
		Store:  opts.Store,
	}

	s.IDP.SessionProvider = s
	s.IDP.ServiceProviderProvider = s

	if err := s.initializeServices(); err != nil {
		return nil, err
	}
	s.InitializeHTTP()
	return s, nil
}

// InitializeHTTP sets up the HTTP handler for the server. (This function
// is called automatically for you by New, but you may need to call it
// yourself if you don't create the object using New.)
func (s *Server) InitializeHTTP() {
	mux := web.New()
	s.Handler = mux

	mux.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		tmpl := template.Must(template.New("index").Parse(indexPageTemplate))
		data := struct {
			MetadataURL string
		}{
			MetadataURL: s.IDP.MetadataURL.String(),
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})

	mux.Get("/metadata", func(w http.ResponseWriter, r *http.Request) {
		s.idpConfigMu.RLock()
		defer s.idpConfigMu.RUnlock()
		s.IDP.ServeMetadata(w, r)
	})
	mux.Handle("/sso", func(w http.ResponseWriter, r *http.Request) {
		s.idpConfigMu.RLock()
		defer s.idpConfigMu.RUnlock()
		s.IDP.ServeSSO(w, r)
	})

	mux.Handle("/login", s.HandleLogin)
	mux.Handle("/login/:shortcut", s.HandleIDPInitiated)
	mux.Handle("/login/:shortcut/*", s.HandleIDPInitiated)

	mux.Get("/services/", s.HandleListServices)
	mux.Get("/services/:id", s.HandleGetService)
	mux.Put("/services/:id", s.HandlePutService)
	mux.Post("/services/:id", s.HandlePutService)
	mux.Delete("/services/:id", s.HandleDeleteService)

	mux.Get("/users/", s.HandleListUsers)
	mux.Get("/users/:id", s.HandleGetUser)
	mux.Put("/users/:id", s.HandlePutUser)
	mux.Delete("/users/:id", s.HandleDeleteUser)

	sessionPath := regexp.MustCompile("/sessions/(?P<id>.*)")
	mux.Get("/sessions/", s.HandleListSessions)
	mux.Get(sessionPath, s.HandleGetSession)
	mux.Delete(sessionPath, s.HandleDeleteSession)

	mux.Get("/shortcuts/", s.HandleListShortcuts)
	mux.Get("/shortcuts/:id", s.HandleGetShortcut)
	mux.Put("/shortcuts/:id", s.HandlePutShortcut)
	mux.Delete("/shortcuts/:id", s.HandleDeleteShortcut)
}
