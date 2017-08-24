package middleware

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/pkg/errors"

	"github.com/teamwork/desk/dlmkeys"
	"github.com/teamwork/desk/helper"
	"github.com/teamwork/desk/models"
	"github.com/teamwork/desk/modules/cache"
	"github.com/teamwork/desk/modules/cache/dlm"
	"github.com/teamwork/desk/modules/context"
	"github.com/teamwork/desk/modules/database"
	"github.com/teamwork/desk/modules/httperr"
	"github.com/teamwork/utils/timeutil"
)

func DeskAdminOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			session, err := helper.GetSession(c)
			if err != nil {
				return c.JSON(http.StatusNotFound, err.Error())
			}

			if !session.IsDeskAdmin {
				return models.ErrAccessDenied
			}

			return next(c)
		}
	}
}

// NoAuth create a session were the user is the Robot
func NoAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			session := &models.Session{
				UserSession: models.UserSession{
					ID:               models.RobotUserID,
					Installation:     &models.Installation{},
					DeskInstallation: &models.DeskInstallation{},
				},
			}
			helper.SetSession(c, session)

			return next(c)
		}
	}
}

func InstallationBasedAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request().(*standard.Request).Request
			installation, err := models.GetInstallationFromDomain(&context.Context{}, req.Host)
			if err != nil {
				return httperr.NewWithCode(http.StatusUnauthorized, "Couldn't find installation.")
			}

			session := &models.Session{
				UserSession: models.UserSession{
					Installation: &installation,
				},
				DB: database.DB.GetDeskShard(installation.Shard),
			}

			session.DeskInstallation, err = models.GetDeskInstallationByID(&context.Context{}, models.GetDeskInstallationArgs{
				ID:      installation.ID,
				Session: session,
			})
			if err != nil {
				return err
			}

			// All good, allow continuation based on installation
			helper.SetSession(c, session)

			return next(c)
		}
	}
}

// CreateInstallationIfNotExists creates a Desk installation (fully set up with
// admins from project) if it does not already exist.  Should only be used on
// routes where an account might not yet be set up.
func CreateInstallationIfNotExists() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request().(*standard.Request).Request
			if _, err := models.GetInstallationFromDomainAndCreate(req.Host); err != nil {
				if errors.Cause(err) == sql.ErrNoRows {
					return httperr.NewWithCode(http.StatusNotFound, "Couldn't find installation")
				}

				return err
			}

			return next(c)
		}
	}
}

func CheckAuthorisation() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// If we're attempting to authenticate, we must stop execution
			// of this function at this point regardless of success or not
			if c.Request().Header().Get("Authorization") != "" {
				if valid, e := checkBasicAuthentication(c); !valid {
					return e
				}
				return next(c)
			}

			// Validate the installation's existence (must always be done on this middleware)
			req := c.Request().(*standard.Request).Request
			installation, err := models.GetInstallationFromDomain(&context.Context{}, req.Host)
			if err != nil {
				return err
			}

			// Next, we should check if this shard is in maintenance
			if database.IsShardInMaintenance(installation.Shard) {
				return models.ErrDownForMaintenance
			}

			// Get global session and check if it exists
			gSession, err := helper.GetGlobalAuthentication(c)
			if err != nil || gSession == nil {
				if c.Request().Header().Get("X-Requested-With") == "XMLHttpRequest" {
					return models.ErrNotAuthorized
				}

				// Called directly; try to auth with HTTP Auth
				if valid, e := checkBasicAuthentication(c); !valid {
					return errors.Wrap(e, "basic auth failed")
				}
				return next(c)
			}

			// Get session from local cache or database as required
			session, err := getSessionObject(c, installation, gSession)
			if err != nil {
				return httperr.WithCode(err, http.StatusNotFound)
			}

			// All is good, the user is valid
			helper.SetSession(c, session)

			// Convenient place to put this logic
			if session.Installation.SSLEnabled {
				c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000")
			}

			return next(c)
		}
	}
}

func getSessionObject(c echo.Context, installation models.Installation, gSession *cache.GlobalSession) (*models.Session, error) {
	// Check if user session is set up for this application or if the user object
	// requires an update
	session, err := models.LoadSessionFromRedis(c.Request().(*standard.Request).Request)
	installationDLM := dlm.Latest([]dlm.Key{dlmkeys.Key{
		InstallationID: installation.ID,
		ItemType:       "installation",
		ItemID:         installation.ID,
	}})

	// NOTE: Temporarily converting installationDLM to milliseconds as it's
	// currently in nanoseconds Should be removed when topper's refactoring is
	// done.
	if err != nil || gSession.UserID != session.ID || gSession.Timestamp > session.DateLastAuthenticated ||
		installationDLM > session.DateLastAuthenticated || session.Version != models.SessionVersion {
		// All other responses we should login the user to this application as
		// at this point they are globally authenticated
		tmpSession, e := models.LoginByUserID(&installation, gSession.UserID, false, &context.Context{})

		// This can happen if the user hasn't been set up in desk
		if e != nil {
			return nil, errors.Wrap(e, "user not set up in desk")
		}

		session = *tmpSession
		if e := session.Save(c.Request().(*standard.Request).Request); e != nil {
			return nil, errors.Wrap(e, "failed to save session")
		}
	}

	session.DB = database.DB.GetDeskShard(installation.Shard)
	session.ProjectsShard = database.DB.GetPMShard(installation.Shard)

	session.Installation = &installation
	session.DeskInstallation, err = models.GetDeskInstallationByID(&context.Context{}, models.GetDeskInstallationArgs{
		ID:      installation.ID,
		Session: &session,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get installation")
	}

	return &session, nil
}
func checkBasicAuthentication(c echo.Context) (bool, error) {
	// BASIC AUTH
	// If a new header has been sent up, we must perform authenication
	// Need to make sure they're not already logged in!
	host := c.Request().Host()
	auth := c.Request().Header().Get("Authorization")

	if strings.Contains(host, ":") {
		host = host[:strings.LastIndex(host, ":")]
	}

	session, installation, err := models.AuthenticateUser(auth, host)
	if err != nil {
		// Only send the WWW-Authenticate header if this is not the TeamworkDesk app
		if twDeskVer := c.Request().Header().Get("twDeskVer"); twDeskVer == "" {
			c.Response().Header().Set("WWW-Authenticate", `Basic realm="Teamwork Desk"`)
		}

		return false, err
	}

	if e := session.Save(c.Request().(*standard.Request).Request); e != nil {
		return false, e
	}

	// Just set it temporarily for the rest of this request
	session.Installation = installation
	session.DeskInstallation, err = models.GetDeskInstallationByID(&context.Context{}, models.GetDeskInstallationArgs{
		ID:      installation.ID,
		Session: session,
	})

	if err != nil {
		return false, err
	}

	helper.SetSession(c, session)

	session.DB = database.DB.GetDeskShard(session.Installation.Shard)

	_, err = cache.AddGlobalSession(&cache.GlobalSession{
		UserID:         session.ID,
		InstallationID: session.Installation.ID,
		Timestamp:      timeutil.UnixMilli(),
		RememberMe:     false,
	})

	if err != nil {
		panic(err)
	}

	return true, nil
}
