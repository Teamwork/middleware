// Package auditMiddleware adds audit logs.
package auditMiddleware // import "github.com/teamwork/middleware/auditMiddleware"

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/teamwork/utils/httputilx"
	"github.com/teamwork/utils/httputilx/header"
	"github.com/teamwork/utils/sliceutil"
)

// Auditor is a function which will do something with a new audit.
type Auditor interface {
	// AddAudit does something with the audit log (e.g. store in database, log
	// somewhere, whatever you want).
	AddAudit(*http.Request, *Audit)

	// UserID gets the user identifier from the request.
	UserID(*http.Request) int64

	// InstallationID gets installation information from the request; for
	// example a company id, organisation id, etc.
	InstallationID(*http.Request) int64

	// Log errors.
	Log(*http.Request, error)
}

// Options for the middleware.
type Options struct {
	Auditor        Auditor
	Methods        []string
	IgnorePaths    []string
	FilteredFields []string

	// MaxSize sets the maximum size to read in bytes for the request body and
	// form params.
	MaxSize int64

	SkipRequestHeaders bool
	SkipQueryParams    bool
	SkipformParams     bool
	SkipRequestBody    bool
}

// Audit entry.
type Audit struct {
	ID             int64  `db:"id"`
	UserID         int64  `db:"users_id"`
	InstallationID int64  `db:"installations_id"`
	Host           string `db:"host"`   // e.g. "example.com"
	Path           string `db:"path"`   // e.g. "/hello"
	Method         string `db:"method"` // e.g. "GET", "POST"
	CreatedAt      Time   `db:"createdAt"`
	RequestHeaders Header `db:"requestHeaders"`
	QueryParams    Values `db:"queryParams"`
	FormParams     Values `db:"formParams"`
	RequestBody    []byte `db:"requestBody"`
}

// Time embeds time.Time and adds methods for the sql.Scanner interface.
type Time struct {
	time.Time
}

// Value implements the SQL Value function to determine what to store in the DB.
func (t Time) Value() (driver.Value, error) {
	return t.Time, nil
}

// Scan converts the data returned from the DB.
func (t *Time) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch tt := value.(type) {
	case time.Time:
		*t = Time{value.(time.Time)}
	case []uint8:
		return nil
	default:
		return fmt.Errorf("unexpected type %T", tt)
	}

	return nil
}

// Header embeds http.Header and adds methods for the sql.Scanner interface.
type Header struct {
	http.Header
}

// Value implements the SQL Value function to determine what to store in the DB.
func (h Header) Value() (driver.Value, error) {
	return json.Marshal(h.Header)
}

// Scan converts the data returned from the DB.
func (h *Header) Scan(val interface{}) error {
	if val == nil {
		return nil
	}

	return json.Unmarshal(val.([]byte), h)
}

// Values embeds url.Values and adds methods for the sql.Scanner interface.
type Values struct {
	url.Values
}

// Value implements the SQL Value function to determine what to store in the DB.
func (v Values) Value() (driver.Value, error) {
	return json.Marshal(v.Values)
}

// Scan converts the data returned from the DB.
func (v *Values) Scan(val interface{}) error {
	if val == nil {
		return nil
	}

	return json.Unmarshal(val.([]byte), v)
}

// Do audit HTTP requests.
func Do(opts Options) func(http.Handler) http.Handler {
	for i := range opts.Methods {
		opts.Methods[i] = strings.ToUpper(opts.Methods[i])
	}
	if opts.Auditor == nil {
		panic("auditMiddleware: Auditor is nil")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Middleware(opts, r)
			next.ServeHTTP(w, r)
		})
	}
}

// Middleware is the actual implementation.
func Middleware(opts Options, r *http.Request) {
	if len(opts.Methods) > 0 && !sliceutil.InStringSlice(opts.Methods, r.Method) {
		return
	}

	if sliceutil.InStringSlice(opts.IgnorePaths, r.URL.Path) {
		return
	}

	a := &Audit{
		InstallationID: opts.Auditor.InstallationID(r),
		UserID:         opts.Auditor.UserID(r),
		Host:           r.Host,
		Path:           r.URL.Path,
		Method:         r.Method,
		CreatedAt:      Time{time.Now()},
	}

	if !opts.SkipRequestHeaders {
		a.RequestHeaders = Header{r.Header}
	}
	if !opts.SkipQueryParams {
		a.QueryParams = Values{r.URL.Query()}
	}
	if !opts.SkipformParams {
		if opts.MaxSize < 0 {
			a.FormParams = Values{r.Form}
		} else {
			a.FormParams = Values{url.Values{}}
			for k, v := range r.Form {
				a.FormParams.Values[k] = make([]string, len(v))
				for i := range v {
					if int64(len(v[i])) > opts.MaxSize {
						a.FormParams.Values[k][i] = v[i][:opts.MaxSize]
					} else {
						a.FormParams.Values[k][i] = v[i]
					}
				}
			}
		}
	}

	if !opts.SkipRequestBody {
		b, err := httputilx.DumpBody(r, opts.MaxSize)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			opts.Auditor.Log(r, errors.Wrap(err, "could not read body"))
		}
		a.RequestBody = b
	}

	if len(opts.FilteredFields) > 0 {
		a.filterFields(r, opts)
	}

	opts.Auditor.AddAudit(r, a)
}

func (a *Audit) filterFields(r *http.Request, opts Options) {
	body := make(map[string]interface{})
	bodyIsJSON := isBodyJSON(r)

	if bodyIsJSON && len(a.RequestBody) > 0 {
		err := json.Unmarshal(a.RequestBody, &body)
		if err != nil {
			opts.Auditor.Log(r, errors.Wrap(err, "could not parse body"))
		}
	}

	for _, field := range opts.FilteredFields {
		if a.RequestHeaders.Get(field) != "" {
			a.RequestHeaders.Set(field, "[FILTERED]")
		}

		if a.FormParams.Get(field) != "" {
			a.FormParams.Set(field, "[FILTERED]")
		}

		if a.QueryParams.Get(field) != "" {
			a.QueryParams.Set(field, "[FILTERED]")
		}

		if _, ok := body[field]; ok {
			body[field] = "[FILTERED]"
		}
	}

	if bodyIsJSON && len(body) > 0 {
		b, err := json.Marshal(body)
		if err != nil {
			opts.Auditor.Log(r, errors.Wrap(err, "could not marshal body"))
		}

		a.RequestBody = b
	}
}

func isBodyJSON(r *http.Request) bool {
	specs := header.ParseAccept(r.Header, "Accept")
	sort.Slice(specs, func(i, j int) bool { return specs[i].Q > specs[j].Q })
	for _, spec := range specs {
		ct, _, _ := mime.ParseMediaType(spec.Value)
		if ct == "application/json" {
			return true
		}
	}

	return false
}
