// CouchDB resource
//
// This is the low-level wrapper functions of HTTP methods
//
// Used for communicating with CouchDB Server
package couchdb

import (
  "bytes"
  "crypto/tls"
  "encoding/json"
  "errors"
  // "fmt"
  "io"
  "io/ioutil"
  "net/http"
  "net/url"
  "strings"
)

var (
  httpClient *http.Client

  ErrNotModified error = errors.New("status 304 - not modified")
  ErrBadRequest error = errors.New("status 400 - bad request")
  ErrUnauthorized error = errors.New("status 401 - unauthorized")
  ErrForbidden error = errors.New("status 403 - forbidden")
  ErrNotFound error = errors.New("status 404 - not found")
  ErrResourceNotAllowed error = errors.New("status 405 - resource not allowed")
  ErrNotAcceptable error = errors.New("status 406 - not acceptable")
  ErrConflict error = errors.New("status 409 - conflict")
  ErrPreconditionFailed error = errors.New("status 412 - precondition failed")
  ErrBadContentType error = errors.New("status 415 - bad content type")
  ErrRequestRangeNotSatisfiable error = errors.New("status 416 - requested range not satisfiable")
  ErrExpectationFailed error = errors.New("status 417 - expectation failed")
  ErrInternalServerError error = errors.New("status 500 - internal server error")
)

func init() {
  httpClient = http.DefaultClient
}

// Resource handles all requests to CouchDB
type Resource struct {
  header http.Header
  base *url.URL
}

// NewResource returns a newly-created Resource instance
func NewResource(urlStr string, header http.Header) (*Resource, error) {
  u, err := url.Parse(urlStr)
  if err != nil {
    return nil, err
  }

  if strings.HasPrefix(urlStr, "https") {
    tr := &http.Transport{
      TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    httpClient = &http.Client{
      Transport: tr,
    }
  }

  var h http.Header
  if header != nil {
    h = header
  }

  return &Resource{
    header: h,
    base: u,
  }, nil
}

func combine(basePath, resPath, separator string) string {
  if resPath == "" {
    return basePath
  }
  return strings.Join([]string{basePath, resPath}, separator)
}

func (r *Resource)NewResourceWithURL(resStr string) (*Resource, error) {
  u, err := r.base.Parse(combine(r.base.Path, resStr, "/"))
  if err != nil {
    return nil, err
  }

  return &Resource{
    header: r.header,
    base: u,
  }, nil
}

// Head is a wrapper around http.Head
func (r *Resource)Head(path string, header http.Header, params url.Values) (http.Header, []byte, error) {
  u, err := r.base.Parse(combine(r.base.Path, path, "/"))
  if err != nil {
    return nil, nil, err
  }
  return request(http.MethodHead, u, header, nil, params)
}

// Get is a wrapper around http.Get
func (r *Resource)Get(path string, header http.Header, params url.Values) (http.Header, []byte, error) {
  u, err := r.base.Parse(combine(r.base.Path, path, "/"))
  if err != nil {
    return nil, nil, err
  }
  return request(http.MethodGet, u, header, nil, params)
}

// Post is a wrapper around http.Post
func (r *Resource)Post(path string, header http.Header, body []byte, params url.Values) (http.Header, []byte, error) {
  u, err := r.base.Parse(combine(r.base.Path, path, "/"))
  if err != nil {
    return nil, nil, err
  }
  return request(http.MethodPost, u, header, bytes.NewReader(body), params)
}

// Delete is a wrapper around http.Delete
func (r *Resource)Delete(path string, header http.Header, params url.Values) (http.Header, []byte, error) {
  u, err := r.base.Parse(combine(r.base.Path, path, "/"))
  if err != nil {
    return nil, nil, err
  }
  return request(http.MethodDelete, u, header, nil, params)
}

// Put is a wrapper around http.Put
func (r *Resource)Put(path string, header http.Header, body []byte, params url.Values) (http.Header, []byte, error) {
  u, err := r.base.Parse(combine(r.base.Path, path, "/"))
  if err != nil {
    return nil, nil, err
  }
  return request(http.MethodPut, u, header, bytes.NewReader(body), params)
}

// GetJSON issues a GET to the specified URL, with data returned as json
func (r *Resource)GetJSON(path string, header http.Header, params url.Values) (http.Header, *json.RawMessage, error) {
  u, err := r.base.Parse(combine(r.base.Path, path, "/"))
  if err != nil {
    return nil, nil, err
  }
  return requestJSON(http.MethodGet, u, header, nil, params)
}

// PostJSON issues a POST to the specified URL, with data returned as json
func (r *Resource)PostJSON(path string, header http.Header, body map[string]interface{}, params url.Values) (http.Header, *json.RawMessage, error) {
  u, err := r.base.Parse(combine(r.base.Path, path, "/"))
  if err != nil {
    return nil, nil, err
  }

  jsonBody, err := json.Marshal(body)
  if err != nil {
    return nil, nil, err
  }

  return requestJSON(http.MethodPost, u, header, bytes.NewReader(jsonBody), params)
}

// DeleteJSON issues a DELETE to the specified URL, with data returned as json
func (r *Resource)DeleteJSON(path string, header http.Header, params url.Values) (http.Header, *json.RawMessage, error) {
  u, err := r.base.Parse(combine(r.base.Path, path, "/"))
  if err != nil {
    return nil, nil, err
  }

  return requestJSON(http.MethodDelete, u, header, nil, params)
}

// PutJSON issues a PUT to the specified URL, with data returned as json
func (r *Resource)PutJSON(path string, header http.Header, body map[string]interface{}, params url.Values) (http.Header, *json.RawMessage, error) {
  u, err := r.base.Parse(combine(r.base.Path, path, "/"))
  if err != nil {
    return nil, nil, err
  }

  jsonBody, err := json.Marshal(body)
  if err != nil {
    return nil, nil, err
  }

  return requestJSON(http.MethodPut, u, header, bytes.NewReader(jsonBody), params)
}

// helper function to make real request
func requestJSON(method string, u *url.URL, header http.Header, body io.Reader, params url.Values) (http.Header, *json.RawMessage, error) {
  err, header, data := request(method, u, header, body, params)
  if err != nil {
    return nil, nil, err
  }

  var jsonData json.RawMessage
  err := json.Unmarshal(data, &jsonData)
  if err != nil {
    return nil, nil, err
  }
  return header, &jsonData, err
}

// helper function to make real request
func request(method string, u *url.URL, header *http.Header, body io.Reader, params url.Values) (http.Header, []byte, error) {
  method = strings.ToUpper(method)

  u.RawQuery = params.Encode()
  var username, password string
  if u.User != nil {
    username = u.User.Username()
    password, _ = u.User.Password()
  }
  req, err := http.NewRequest(method, u.String(), body)
  if err != nil {
    return nil, nil, err
  }

  if len(username) > 0 && len(password) > 0 {
    req.SetBasicAuth(username, password)
  }

  // Accept and Content-type are highly recommended for CouchDB
  setDefault(&req.Header, "Accept", "application/json")
  setDefault(&req.Header, "Content-Type", "application/json")
  updateHeader(&req.Header, header)

  rsp, err := httpClient.Do(req)
  if err != nil {
    return nil, nil, err
  }
  defer rsp.Body.Close()
  data, err := ioutil.ReadAll(rsp.Body)
  if err != nil {
    return nil, nil ,err
  }

  return rsp.Header, data, nil
}

// setDefault sets the default value if key not existe in header
func setDefault(header *http.Header, key, value string) {
  if header.Get(key) == "" {
    header.Set(key, value)
  }
}

// updateHeader updates existing header with new values
func updateHeader(header *http.Header, extra *http.Header) {
  if header != nil && extra != nil {
    for k, _ := range *extra {
      header.Set(k, extra.Get(k))
    }
  }
}
