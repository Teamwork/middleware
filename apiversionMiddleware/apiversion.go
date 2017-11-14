package apiversion // import "github.com/teamwork/middleware/apiversion"

import (
	"fmt"
	"github.com/spf13/viper"
	"net/http"
)

// InitServerHeaders adds region, git commit & beta/prod to the server response
func APIVersion(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// output server header (example format US BETA 6ba849de1881e0d859a45714462ccb6d5ee9f015)
		// SERVICE_TARGET is PROD or BETA (set in travis env vars)
		apiVersion := fmt.Sprintf("%s %s %s", viper.GetString("AWS_REGION"),
			viper.GetString("SERVICE_TARGET"), viper.GetString("VERSION"))
		w.Header().Set("Api-Version", apiVersion)
		f(w, r)
	}
}
