package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const LND_API = "https://api.linode.com/v4/"

type Linode struct {
	Token  string
	Domain string

	c *http.Client
}

type LinodePagination struct {
	Page    int
	Pages   int
	Results int
}

type LinodeDomainResponse struct {
	LinodePagination
	Data []struct {
		ID     int
		Type   string
		Domain string
	}
}

type LinodeDomainRecordResponse struct {
	LinodePagination
	Data []struct {
		ID     int
		Type   string
		Name   string
		Target string
	}
}

func (l Linode) AuthValid() (bool, error) {
	r, err := l.fetch("GET", "profile", "", nil)
	if err != nil {
		return false, err
	}
	log.Debugf("authentication: %s\n", r.Status)
	if r.StatusCode != 200 {
		log.Infof("Authorization invalid, got %s\n", r.Status)
		return false, nil
	}
	return true, nil
}

func (l Linode) SetRecord(domain string, ip string) error {
	log.Infof("Reading current record...")

	//is it primary domain, or subdomain
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return errors.New("invalid domain: no .")
	}
	host := fmt.Sprintf("%s.%s", parts[len(parts)-2], parts[len(parts)-1])

	subdomain := ""
	if len(parts) > 2 {
		subdomain = strings.Join(parts[:len(parts)-2], ".")
	}

	r, err := l.fetch("GET", "domains", fmt.Sprintf(`"domain": "%s"`, host), nil)
	if err != nil {
		return err
	}

	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	log.Debugf("%s\n", body)

	var resp LinodeDomainResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return err
	}
	if resp.Results < 1 || len(resp.Data) < 1 {
		return fmt.Errorf("Domain not Found: %s\n", host)
	}

	d := resp.Data[0]

	r, err = l.fetch("GET", fmt.Sprintf("domains/%d/records", d.ID), "", nil)
	if err != nil {
		return err
	}

	body, _ = ioutil.ReadAll(r.Body)
	r.Body.Close()
	log.Debugf("%s\n", body)

	var subResp LinodeDomainRecordResponse
	err = json.Unmarshal(body, &subResp)
	if err != nil {
		return err
	}

	//todo: check other pages of data if applic

	// todo: IPv6, AAAA record

	for _, rec := range subResp.Data {
		if rec.Type == "A" && rec.Name == subdomain {
			// update
			log.Infof("Updating record %d to have %s point to %s\n", rec.ID, subdomain, ip)
			r, err := l.fetch(
				"PUT",
				fmt.Sprintf("domains/%d/records/%d", d.ID, rec.ID), "",
				strings.NewReader(fmt.Sprintf(`{"target": "%s"}`, ip)),
			)
			if err != nil {
				return err
			}
			body, _ := ioutil.ReadAll(r.Body)
			r.Body.Close()
			log.Debugf("%s\n", body)
			return nil
		}
	}

	// didn't find one, need to create the record
	log.Infof("Creating A record for %s, pointing to %s\n", subdomain, ip)
	r, err = l.fetch(
		"POST",
		fmt.Sprintf("domains/%d/records", d.ID), "",
		strings.NewReader(
			fmt.Sprintf(`{"type":"A", "name":"%s", "target":"%s"}`, subdomain, ip),
		),
	)
	body, _ = ioutil.ReadAll(r.Body)
	r.Body.Close()
	log.Debugf("%s\n", body)
	return err
}

func (l Linode) fetch(method, path string, filter string, body io.Reader) (*http.Response, error) {
	if l.c == nil {
		l.c = &http.Client{
			Timeout: time.Second & 10,
		}
	}

	api, err := url.Parse(LND_API)
	if err != nil {
		log.Errorf("Error parsing default Linode api, %s: %v\n", LND_API, err)
		return nil, err
	}

	p, err := url.Parse(path)
	if err != nil {
		log.Errorf("error parsing path, %s: %v\n", path, err)
		return nil, err
	}
	u := api.ResolveReference(p)
	log.Debugf("Performing %s request to %s\n", method, u.String())
	req, err := http.NewRequest(method, u.String(), body)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", l.Token))
	if filter != "" {
		req.Header.Add("X-Filter", fmt.Sprintf("{%s}", filter))
	}
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	return l.c.Do(req)
}
