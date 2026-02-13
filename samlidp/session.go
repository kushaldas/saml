package samlidp

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/zenazn/goji/web"

	"github.com/crewjam/saml"
)

var sessionMaxAge = time.Hour

const loginFormTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sign In &mdash; Kushal's Identity Provider for Teaching</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Lora:wght@400;600&amp;family=Source+Sans+3:wght@400;500;600&amp;display=swap" rel="stylesheet">
    <style>
        *, *::before, *::after { margin: 0; padding: 0; box-sizing: border-box; }
        :root {
            --navy: #002b5c; --navy-light: #003d82; --blue: #1d6fa5;
            --sky: #e8f1f8; --white: #ffffff; --text: #1c2536;
            --muted: #5a6678; --light: #8892a4; --border: #dde3ea;
            --err-bg: #fef2f2; --err-text: #b91c1c; --err-border: #fca5a5;
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
            width: 100%; max-width: 400px;
            border: 1px solid var(--border); border-radius: 8px;
            box-shadow: 0 1px 3px rgba(0,43,92,.05), 0 6px 16px rgba(0,43,92,.03);
            padding: 2.25rem 2rem;
        }
        .card-heading {
            font-family: 'Lora', Georgia, serif; font-weight: 600;
            font-size: 1.25rem; color: var(--navy); margin-bottom: 0.35rem;
        }
        .card-sub {
            font-size: 0.9rem; color: var(--muted);
            margin-bottom: 1.75rem; line-height: 1.5;
        }
        .toast {
            background: var(--err-bg); border: 1px solid var(--err-border);
            color: var(--err-text); padding: 0.65rem 0.9rem;
            border-radius: 6px; font-size: 0.875rem;
            margin-bottom: 1.25rem; line-height: 1.4;
        }
        .field { margin-bottom: 1.15rem; }
        .field label {
            display: block; font-size: 0.825rem; font-weight: 500;
            color: var(--text); margin-bottom: 0.35rem;
        }
        .field input {
            width: 100%; padding: 0.6rem 0.8rem;
            font-family: inherit; font-size: 0.925rem;
            color: var(--text); background: var(--white);
            border: 1px solid var(--border); border-radius: 6px;
            transition: border-color 0.15s, box-shadow 0.15s;
        }
        .field input:focus {
            outline: none; border-color: var(--blue);
            box-shadow: 0 0 0 3px rgba(29,111,165,.1);
        }
        .btn-submit {
            width: 100%; padding: 0.7rem;
            font-family: inherit; font-size: 0.95rem; font-weight: 600;
            color: var(--white); background: var(--navy);
            border: none; border-radius: 6px;
            cursor: pointer; transition: background 0.15s; margin-top: 0.25rem;
        }
        .btn-submit:hover { background: var(--navy-light); }
        .btn-submit:active { background: #001f42; }
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
            <div class="card-heading">Sign in</div>
            <div class="card-sub">Enter your credentials to access the service.</div>
            {{if .Toast}}<div class="toast">{{.Toast}}</div>{{end}}
            <form method="post" action="{{.URL}}">
                <div class="field">
                    <label for="user">Username</label>
                    <input type="text" id="user" name="user" autocomplete="username" autofocus required>
                </div>
                <div class="field">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" autocomplete="current-password" required>
                </div>
                <input type="hidden" name="SAMLRequest" value="{{.SAMLRequest}}">
                <input type="hidden" name="RelayState" value="{{.RelayState}}">
                <button type="submit" class="btn-submit">Sign in</button>
            </form>
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

// GetSession returns the *Session for this request.
//
// If the remote user has specified a username and password in the request
// then it is validated against the user database. If valid it sets a
// cookie and returns the newly created session object.
//
// If the remote user has specified invalid credentials then a login form
// is returned with an English-language toast telling the user their
// password was invalid.
//
// If a session cookie already exists and represents a valid session,
// then the session is returned
//
// If neither credentials nor a valid session cookie exist, this function
// sends a login form and returns nil.
func (s *Server) GetSession(w http.ResponseWriter, r *http.Request, req *saml.IdpAuthnRequest) *saml.Session {
	// if we received login credentials then maybe we can create a session
	if r.Method == "POST" && r.PostForm.Get("user") != "" {
		user := User{}
		if err := s.Store.Get(fmt.Sprintf("/users/%s", r.PostForm.Get("user")), &user); err != nil {
			s.sendLoginForm(w, r, req, "Invalid username or password")
			return nil
		}

		if err := bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(r.PostForm.Get("password"))); err != nil {
			s.sendLoginForm(w, r, req, "Invalid username or password")
			return nil
		}
		session := &saml.Session{
			ID:         base64.StdEncoding.EncodeToString(randomBytes(32)),
			NameID:     user.Email,
			CreateTime: saml.TimeNow(),
			ExpireTime: saml.TimeNow().Add(sessionMaxAge),
			Index:      hex.EncodeToString(randomBytes(32)),
			UserName:   user.Name,
			// nolint:gocritic // Groups should be a slice here.
			Groups:                user.Groups[:],
			UserEmail:             user.Email,
			UserCommonName:        user.CommonName,
			UserSurname:           user.Surname,
			UserGivenName:         user.GivenName,
			UserScopedAffiliation: user.ScopedAffiliation,
		}
		if err := s.Store.Put(fmt.Sprintf("/sessions/%s", session.ID), &session); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    session.ID,
			MaxAge:   int(sessionMaxAge.Seconds()),
			HttpOnly: true,
			Secure:   r.URL.Scheme == "https",
			Path:     "/",
		})
		return session
	}

	if sessionCookie, err := r.Cookie("session"); err == nil {
		session := &saml.Session{}
		if err := s.Store.Get(fmt.Sprintf("/sessions/%s", sessionCookie.Value), session); err != nil {
			if err == ErrNotFound {
				s.sendLoginForm(w, r, req, "")
				return nil
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil
		}

		if saml.TimeNow().After(session.ExpireTime) {
			s.sendLoginForm(w, r, req, "")
			return nil
		}
		return session
	}

	s.sendLoginForm(w, r, req, "")
	return nil
}

// sendLoginForm produces a form which requests a username and password and directs the user
// back to the IDP authorize URL to restart the SAML login flow, this time establishing a
// session based on the credentials that were provided.
func (s *Server) sendLoginForm(w http.ResponseWriter, _ *http.Request, req *saml.IdpAuthnRequest, toast string) {
	tmpl := template.Must(template.New("saml-post-form").Parse(loginFormTemplate))
	data := struct {
		Toast       string
		URL         string
		SAMLRequest string
		RelayState  string
	}{
		Toast:       toast,
		URL:         req.IDP.SSOURL.String(),
		SAMLRequest: base64.StdEncoding.EncodeToString(req.RequestBuffer),
		RelayState:  req.RelayState,
	}

	if err := tmpl.Execute(w, data); err != nil {
		panic(err)
	}
}

// HandleLogin handles the `POST /login` and `GET /login` forms. If credentials are present
// in the request body, then they are validated. For valid credentials, the response is a
// 200 OK and the JSON session object. For invalid credentials, the HTML login prompt form
// is sent.
func (s *Server) HandleLogin(_ web.C, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	session := s.GetSession(w, r, &saml.IdpAuthnRequest{IDP: &s.IDP})
	if session == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(session); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// HandleListSessions handles the `GET /sessions/` request and responds with a JSON formatted list
// of session names.
func (s *Server) HandleListSessions(_ web.C, w http.ResponseWriter, _ *http.Request) {
	sessions, err := s.Store.List("/sessions/")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(struct {
		Sessions []string `json:"sessions"`
	}{Sessions: sessions})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// HandleGetSession handles the `GET /sessions/:id` request and responds with the session
// object in JSON format.
func (s *Server) HandleGetSession(c web.C, w http.ResponseWriter, _ *http.Request) {
	session := saml.Session{}
	err := s.Store.Get(fmt.Sprintf("/sessions/%s", c.URLParams["id"]), &session)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(session); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// HandleDeleteSession handles the `DELETE /sessions/:id` request. It invalidates the
// specified session.
func (s *Server) HandleDeleteSession(c web.C, w http.ResponseWriter, _ *http.Request) {
	err := s.Store.Delete(fmt.Sprintf("/sessions/%s", c.URLParams["id"]))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
