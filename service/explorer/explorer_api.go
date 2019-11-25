package explorer

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (ex *Explorer) APILogin() func(c echo.Context) error {
	return func(c echo.Context) error {
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		req := map[string]string{}
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}

		UserID, has := req["user_id"]
		if !has {
			return ErrInvalidRequest
		}
		Password, has := req["password"]
		if !has {
			return ErrInvalidRequest
		}

		if UserID == "admin" && Password == "testpw" {
			sess, _ := session.Get("session", c)
			sess.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   86400 * 7,
				HttpOnly: true,
			}
			sess.Values = map[interface{}]interface{}{}
			sess.Values["login"] = UserID
			sess.Save(c.Request(), c.Response())

			return c.NoContent(http.StatusOK)
		}
		return ErrInvalidAuthorization
	}
}

func (ex *Explorer) APIDataSearch() func(c echo.Context) error {
	return func(c echo.Context) error {
		args := map[string]interface{}{}
		args["summary"] = map[string]interface{}{
			"height":   ex.Height(),
			"tx_count": ex.TransactionCount(),
		}
		return c.JSON(http.StatusOK, args)
	}
}
