package couchdb

import (
  "net/url"
  "os"
  "strings"
)

const (
  DEFAULT_BASE_URL = "http://localhost:5984"
)

// getDefaultCouchDBURL returns the default CouchDB server url
func getDefaultCouchDBURL() string {
  var couchdbUrlEnviron string
  for _, couchdbUrlEnviron = range os.Environ() {
    if strings.HasPrefix(couchdbUrlEnviron, "COUCHDB_URL") {
      break
    }
  }
  if len(couchdbUrlEnviron) == 0 {
    couchdbUrlEnviron = DEFAULT_BASE_URL
  } else {
    couchdbUrlEnviron = strings.Split(couchdbUrlEnviron, "=")[1]
  }
  return couchdbUrlEnviron
}

// Database represents a CouchDB database instance
type Database struct {
  resource *Resource
}

// NewDatabase returns a CouchDB database instance
func NewDatabase(urlStr string) *Database {
  var dbUrlStr string
  if !strings.HasPrefix(urlStr, "http") {
    base, err := url.Parse(getDefaultCouchDBURL())
    if err != nil {
      return nil
    }
    dbUrl, err := base.Parse(urlStr)
    if err != nil {
      return nil
    }
    dbUrlStr = dbUrl.String()
  } else {
    dbUrlStr = urlStr
  }
  res, err := NewResource(dbUrlStr, nil)

  if err != nil {
    return nil
  }

  return &Database{
    resource: res,
  }
}
