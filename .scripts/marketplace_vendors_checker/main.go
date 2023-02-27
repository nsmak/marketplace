package main

import (
  "errors"
  "flag"
  "fmt"
  "net/http"
  "os"
  "path/filepath"
  "strings"

  "gopkg.in/yaml.v3"
)

const vendorsFile = ".vendors/vendors.yml"

var errLinterFailed = errors.New("linter failed")

type Vendor struct {
  ID      string    `yaml:"id"`
  IconURL yaml.Node `yaml:"icon_url"`
  Website yaml.Node `yaml:"website"`
}

func main() {
  marketplacePath := flag.String("p", "./", "path to marketplace directory")
  repoName := flag.String("r", "", "repo name")
  branchName := flag.String("b", "", "branch name")
  flag.Parse()

  if err := run(strings.Join([]string{*repoName, *branchName}, "/"), *marketplacePath, flag.Args()); err != nil {
    if !errors.Is(err, errLinterFailed) {
      fmt.Printf("::error:: failed run marketplace vendors validation %s\n", err)
    }
    os.Exit(1)
  }
}

func run(repoName, marketplacePath string, changedFilesPaths []string) error {
  if !vendorsFileChanged(changedFilesPaths) {
    return nil
  }

  vendors, err := parseVendors(filepath.Join(marketplacePath, vendorsFile))
  if err != nil {
    return fmt.Errorf("parse vendors: %w", err)
  }

  err = validateVendors(vendors, repoName)
  if err != nil {
    return fmt.Errorf("validate vendors: %w", err)
  }

  return nil
}

func vendorsFileChanged(changedFiles []string) bool {
  for _, file := range changedFiles {
    if file == vendorsFile {
      return true
    }
  }
  return false
}

func parseVendors(filePath string) ([]Vendor, error) {
  f, err := os.Open(filePath)
  if err != nil {
    return nil, fmt.Errorf("open file: %w", err)
  }
  defer f.Close()

  var vendors []Vendor
  err = yaml.NewDecoder(f).Decode(&vendors)
  if err != nil {
    return nil, fmt.Errorf("parse yaml: %w", err)
  }

  return vendors, nil
}

func validateVendors(vendors []Vendor, repoName string) error {
  resOk := true
  idsCount := make(map[string]int, len(vendors))

  for _, v := range vendors {
    ok, err := checkResourceExistsAtURL(v.Website.Value)
    if err != nil {
      return fmt.Errorf("check vendor website: %w", err)
    }

    if !ok {
      resOk = false
      logWarning(
        vendorsFile, v.Website.Line, v.Website.Column, fmt.Sprintf("website %s is not available", v.Website.Value),
      )
    }

    iconURL := strings.Replace(
      v.IconURL.Value,
      "Enapter/marketplace/main",
      repoName,
      1,
    )
    ok, err = checkResourceExistsAtURL(iconURL)
    if err != nil {
      return fmt.Errorf("check vendor icon: %w", err)
    }

    if !ok {
      resOk = false
      logWarning(
        vendorsFile, v.IconURL.Line, v.IconURL.Column, fmt.Sprintf("icon %s not found", iconURL),
      )
    }

    idsCount[v.ID]++
  }

  for id, count := range idsCount {
    if count > 1 {
      resOk = false
      logWarning(
        vendorsFile, 1, 1, fmt.Sprintf("vendor id %s is not unique (found %d things)", id, count),
      )
    }
  }

  if !resOk {
    return errLinterFailed
  }

  return nil
}

func checkResourceExistsAtURL(url string) (bool, error) {
  r, err := http.Get(url)
  if err != nil {
    return false, fmt.Errorf("check blueprint icon: %w", err)
  }
  defer r.Body.Close()

  return r.StatusCode == http.StatusOK, nil
}

func logWarning(filePath string, line, column int, msg string) {
  fmt.Printf("::warning file=%s,line=%d,col=%d::%s\n", filePath, line, column, msg)
}
