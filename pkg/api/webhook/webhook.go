package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/parnurzeal/gorequest"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/sirupsen/logrus"
)

var (
	ErrConnectionFailed = errors.New("connection failed")
	ErrForbidden        = errors.New("not allowed")
	ErrUnknown          = errors.New("unknown error")
)

type Webhook interface {
	Connect(m map[string]string) (*IncomingConnectionWebhookResponse, error)
}

func NewClient(opts models.WebhookOptions) Webhook {
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = &http.Client{}
	retryClient.RetryMax = 3
	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if _, ok := err.(net.Error); ok {
			return true, nil
		}

		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	gorequest.DisableTransportSwap = true

	httpClient := gorequest.New()
	httpClient.Client = retryClient.StandardClient()

	w := &webhookClient{
		host: opts.URL,
		http: httpClient,
	}

	if w.logger != nil {
		retryClient.Logger = &client.LeveledLogger{w.logger}
	}

	return w
}

type webhookClient struct {
	host   string
	http   *gorequest.SuperAgent
	logger *logrus.Logger
}

func (w *webhookClient) Connect(m map[string]string) (*IncomingConnectionWebhookResponse, error) {
	payload := &IncomingConnectionWebhookRequest{
		Username:  m["username"],
		Hostname:  m["name"],
		Namespace: m["domain"],
		SourceIP:  m["ip_address"],
	}
	secret := "secret"
	uuid := uuid.Generate()
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(fmt.Sprintf("%v", payload))); err != nil {
		return nil, err
	}
	signature := mac.Sum(nil)

	var res *IncomingConnectionWebhookResponse
	resp, _, errs := w.http.Post(w.host).Set(WebhookIDHeader, uuid).Set(WebhookEventHeader, WebhookIncomingConnectionEvent).Set(WebhookSignatureHeader, hex.EncodeToString(signature)).Send(payload).EndStruct(&res)

	if len(errs) > 0 {
		return nil, ErrConnectionFailed
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}

	if resp.StatusCode == http.StatusOK {
		return res, nil
	}

	return nil, ErrUnknown
}
