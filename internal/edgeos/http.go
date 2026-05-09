package edgeos

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// download creates http requests to download data.
func download(s *source) *source {
	var (
		body []byte
		err  error
		resp *http.Response
		req  *http.Request
	)

	timeout := s.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	baseCtx := s.HTTPCtx
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	ctx, cancel := context.WithTimeout(baseCtx, timeout)
	defer cancel()

	if req, err = http.NewRequestWithContext(ctx, s.Method, s.url, nil); err != nil {
		s.Log.Warning(fmt.Sprintf("Unable to form request for %s", s.url))
		s.r, s.err = bytes.NewReader([]byte{}), err
		return s
	}

	s.Log.Info(fmt.Sprintf("Downloading %s source %s", s.area(), s.name))

	req.Header.Set("User-Agent", agent)

	client := defaultHTTPClient
	if s.HTTP != nil {
		client = s.HTTP
	}

	if resp, err = client.Do(req); err != nil {
		s.Log.Warning(fmt.Sprintf("Unable to get response for %s", s.url))
		s.r, s.err = bytes.NewReader([]byte{}), err
		return s
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.Log.Warning(cerr.Error())
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("unexpected HTTP status %s for %s", resp.Status, s.url)
		s.Log.Warning(err.Error())
		s.r, s.err = bytes.NewReader([]byte{}), err
		return s
	}

	limited := io.LimitReader(resp.Body, MaxDownloadBytes)
	body, err = io.ReadAll(limited)
	if err != nil {
		s.Log.Warning(err.Error())
		s.r, s.err = bytes.NewReader([]byte{}), err
		return s
	}

	if len(body) < 1 {
		str := fmt.Sprintf("No data returned for %s", s.url)
		s.Log.Warning(str)
		s.r, s.err = bytes.NewReader([]byte{}), fmt.Errorf("%s", str)
		return s
	}

	s.r, s.err = bytes.NewBuffer(body), nil
	return s
}
