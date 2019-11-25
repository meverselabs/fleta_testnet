package handler

import (
	"fmt"
	"net/http"
	"time"

	mobiledetect "github.com/Shaked/gomobiledetect"
	"github.com/labstack/echo/v4"
)

type PageThema struct {
	TopLeft    string
	TopRight   string
	TopThema   string
	BottomMenu string
}

func FillArgs(c echo.Context, m map[string]interface{}) error {
	mErr, has := m["errmsg"]
	if !has {
		m["errmsg"] = ""
	}
	errmsg := c.QueryParam("errmsg")
	if errmsg != "" {
		if err, ok := mErr.(string); ok {
			m["errmsg"] = err + ":" + errmsg
		} else {
			m["errmsg"] = errmsg
		}
	}

	ck, _ := c.Cookie("lang")
	lang := "ko"
	if ck != nil {
		lang = string(ck.Value)
	}
	l := Lang.Store[lang]
	t := time.Now()
	if lang == "ko" {
		l["now"] = fmt.Sprintf("%d-%02d-%02d, %02d:%02d:%02d KST 기준", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	} else {
		l["now"] = fmt.Sprintf("As of %2d-%02d-%0d, %02d:%02d:%02d KST", t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), t.Second())
	}

	m["langType"] = lang
	m["lang"] = l

	detect := mobiledetect.NewMobileDetect(c.Request(), nil)
	if detect.IsMobile() {
		m["mobile"] = "mobile"
	}

	return nil
}

func StaticHandler(page string, t PageThema) func(c echo.Context) error {
	return func(c echo.Context) error {
		args := map[string]interface{}{
			"name":      page,
			"pageThema": t,
		}
		if err := FillArgs(c, args); err != nil {
			return err
		}
		return c.Render(http.StatusOK, page+".html", args)
	}
}
