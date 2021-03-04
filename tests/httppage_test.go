package tests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/luis-quan/cellnet"
	"github.com/luis-quan/cellnet/peer"
	httppeer "github.com/luis-quan/cellnet/peer/http"
	"github.com/luis-quan/cellnet/proc"
)

const pageAddress = "127.0.0.1:10087"

func TestPrintPage(t *testing.T) {

	p := peer.NewGenericPeer("http.Acceptor", "httpserver", pageAddress, nil)

	proc.BindProcessorHandler(p, "http", func(raw cellnet.Event) {

		switch {
		case raw.Session().(httppeer.RequestMatcher).Match("GET", "/"):

			raw.Session().Send(&httppeer.HTMLRespond{
				StatusCode:    http.StatusOK,
				PageTemplate:  "index",
				TemplateModel: "world",
			})
		}

	})

	p.Start()

	validPage(t, fmt.Sprintf("http://%s", pageAddress), "<h1>Hello world</h1>")

	p.Stop()

}

func validPage(t *testing.T, url, expectAck string) {
	c := &http.Client{
		Timeout: time.Second * 5,
	}
	resp, err := c.Get(url)
	if err != nil {
		t.Log("http req failed", err)
		t.FailNow()
	}

	defer resp.Body.Close()
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Log("http response failed", err)
		t.FailNow()
	}

	body := string(bodyData)

	if body != expectAck {
		t.Log("unexpect result", err, body)
		t.FailNow()
	}
}
