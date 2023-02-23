package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	STATE = "state"
	CODE  = "code"
	ERROR = "error"
)

type Response struct {
	Err  error
	Code string
}

func callbackServer(ctx context.Context, state, addr string) (string, error) {
	done := make(chan interface{})
	response := make(chan Response)

	u, err := url.Parse(addr)
	if err != nil {
		return "", fmt.Errorf("error parsing addr %s: %w", addr, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerFunc(state, response))
	srv := &http.Server{
		Handler:           mux,
		Addr:              fmt.Sprintf(":%s", u.Port()),
		ReadHeaderTimeout: time.Minute,
	}

	go func() {
		if e := srv.ListenAndServe(); e != http.ErrServerClosed {
			response <- Response{Err: e}
		}
	}()

	go func() {
		select {
		case <-done:
			if e := srv.Shutdown(ctx); errors.Is(e, http.ErrServerClosed) {
				response <- Response{Err: e}
			}
			return
		case <-ctx.Done():
			e := fmt.Errorf("timed out waiting for callback: %w", ctx.Err())
			response <- Response{Err: e}
		}
	}()

	res := <-response

	if res.Err != nil {
		return "", fmt.Errorf("error in response: %w", res.Err)
	}

	return res.Code, nil
}

func handlerFunc(state string, response chan Response) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qs := r.URL.Query()
		var res Response

		if subtle.ConstantTimeCompare([]byte(state), []byte(qs.Get(STATE))) == 0 {
			res.Err = fmt.Errorf("invalid state")
			fmt.Fprint(w, "invalid state")
		} else if errString := qs.Get(ERROR); errString != "" {
			errDesc := qs.Get("error_description")
			res.Err = fmt.Errorf("%s: %s", errString, errDesc)
			fmt.Fprint(w, errDesc)
		} else if code := qs.Get(CODE); code != "" {
			res.Code = code
			fmt.Fprint(w, "Code retrieved. You can close this window.")
		} else {
			res.Err = fmt.Errorf("no error or code returned")
			fmt.Fprint(w, "no error or code returned")
		}

		response <- res
	}
}
