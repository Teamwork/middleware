package middleware // import "github.com/teamwork/middleware"

import "net/http"

// AdministratorLockdown ensures that the admin can only be accessed locally or
// on digitalcrew accounts
func AdministratorLockdown(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "digitalcrew.teamwork.com" ||
			r.Host == "digitalcreweu.eu.teamwork.com" ||
			r.Host == "sunbeam.teamwork.dev" {
			f(w, r)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
		return
	}
}
