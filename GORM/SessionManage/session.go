package sessionMgt

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

var sessionStore = sessions.NewCookieStore([]byte("your-secret-key"))

func LoginSession(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		fmt.Fprint(w, "Username is Required")
		return
	}
	//retrieve a session associated with a given HTTP request (r)
	session, err := sessionStore.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//storing the data into the sessionStore
	session.Values["username"] = username
	session.Values["authentication"] = true

	errr := session.Save(r, w)
	if errr != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, "Login Successfully")
}

func Protected(w http.ResponseWriter, r *http.Request) {
	session, err := sessionStore.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if auth, ok := session.Values["authentication"].(bool); !ok || !auth {
		http.Error(w, "Unauthenticated", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, "Protected")
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, err := sessionStore.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["authentication"] = false
	session.Options.MaxAge = -1 //for clearing the session
	errr := session.Save(r, w)
	if errr != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, "Logout Successfully")
}
