package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func TestCallback(t *testing.T) {
	callback := CallBackMgr{HTTPSCallback: "http://127.0.0.1:62323/callback", HTTPSCallbackAuth: "Bearer authkey"}

	e := echo.New()
	e.POST("/callback", func(c echo.Context) error {
		var data CallbackModel
		err := c.Bind(&data)
		if err != nil {
			t.Fail()
			return c.JSON(http.StatusInternalServerError, "")
		}

		if data.Action != "action" || data.IP != "ip" || data.ID != "id" {
			t.Fail()
		}

		return c.JSON(http.StatusOK, "")
	})
	e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Validator: func(key string, _ echo.Context) (bool, error) {
			if key != "authkey" {
				t.Fail()
				return false, nil
			}
			return key == "authkey", nil
		},
	}))
	go e.Start(":62323")
	time.Sleep(3 * time.Second)

	err := callback.callback("id", "action", "ip")

	if err != nil {
		t.Fail()
	}
}
