package zitadel_authentication

import (
	"errors"
	"fmt"
	"net/http"
)

type CookieHandler interface {
	SetCookie(name, value, cookiePath string, seconds int) error
	GetCookie(name string) (string, error)
	DelCookie(name, cookiePath string) error
}

type HttpCookie struct {
	r *http.Request
	w http.ResponseWriter
}

func NewHttpCookie(r *http.Request, w http.ResponseWriter) CookieHandler {
	return &HttpCookie{
		r: r,
		w: w,
	}
}

func (h *HttpCookie) SetCookie(name, value, path string, seconds int) error {
	if path == "" {
		return fmt.Errorf("path is require to set cookie")
	}
	// isSecure := h.r.TLS != nil || h.r.Header.Get("X-Forwarded-Proto") == "https"
	http.SetCookie(h.w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   "",
		MaxAge:   seconds,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
	return nil
}

func (h *HttpCookie) DelCookie(name string, cookiePath string) error {
	// isSecure := h.r.TLS != nil || h.r.Header.Get("X-Forwarded-Proto") == "https"
	http.SetCookie(h.w, &http.Cookie{
		Name:     name,
		Path:     cookiePath,
		Domain:   "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
	return nil
}

func (h *HttpCookie) GetCookie(name string) (string, error) {
	cookie, err := h.r.Cookie(name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", nil
		}
	}
	return cookie.Value, nil
}
