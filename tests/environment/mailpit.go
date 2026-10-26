package environment

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Email is the HTML of a message Mailpit caught.
type Email struct {
	html *html.Node
}

// Labelled returns the trimmed text of the first element in the email whose aria-label is label,
// and false when no element carries that label.
func (e Email) Labelled(label string) (string, bool) {
	for node := range e.html.Descendants() {
		if node.Type == html.ElementNode && attribute(node, "aria-label") == label {
			return strings.TrimSpace(textOf(node)), true
		}
	}

	return "", false
}

// Text returns the text of the email's body, with each run of whitespace collapsed to one space.
func (e Email) Text() string {
	for node := range e.html.Descendants() {
		if node.Type == html.ElementNode && node.DataAtom == atom.Body {
			return strings.Join(strings.Fields(textOf(node)), " ")
		}
	}

	return ""
}

// Image returns the source of the first image in the email whose alt text is alt, and false when
// no image carries it.
func (e Email) Image(alt string) (string, bool) {
	for node := range e.html.Descendants() {
		if node.Type == html.ElementNode && node.DataAtom == atom.Img && attribute(node, "alt") == alt {
			return attribute(node, "src"), true
		}
	}

	return "", false
}

type mailpitAddress struct {
	Address string `json:"Address"`
}

type mailpitSummary struct {
	ID string           `json:"ID"`
	To []mailpitAddress `json:"To"`
}

type mailpitSearch struct {
	Messages []mailpitSummary `json:"messages"`
}

type mailpitMessage struct {
	HTML string `json:"HTML"`
}

// AwaitEmail waits until Mailpit holds an email to the address to whose HTML carries a link to path,
// and returns the newest such email, in the order Mailpit lists them. It fails t when the stack runs
// no Mailpit, which only a cloud stack does, and when no such email can be read within 30 seconds,
// whether none arrived or Mailpit kept answering with an error or HTML that does not parse.
func (dc *DockerCompose) AwaitEmail(t *testing.T, to, path string) Email {
	t.Helper()

	return dc.AwaitEmails(t, to, path, 1)[0]
}

// AwaitEmails waits until Mailpit holds at least count emails to the address to whose HTML carries
// a link to path, and returns every such email, newest first. It fails t as [DockerCompose.AwaitEmail]
// does, when fewer than count can be read within 30 seconds.
func (dc *DockerCompose) AwaitEmails(t *testing.T, to, path string, count int) []Email {
	t.Helper()

	base := dc.stack.MailpitURL()
	require.NotEmpty(t, base, "only a cloud stack catches email")

	client := resty.New().SetBaseURL(base)

	var found []Email

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		emails, err := matchingEmails(t.Context(), client, to, path)
		if assert.NoError(tt, err) && assert.GreaterOrEqual(tt, len(emails), count, "too few emails to %s link to %s", to, path) {
			found = emails
		}
	}, 30*time.Second, 250*time.Millisecond)

	return found
}

// Emails returns every email Mailpit holds to the address to whose HTML carries a link to path,
// newest first, without waiting for one to arrive. It fails t when the stack runs no Mailpit, or
// when Mailpit answers with an error or HTML that does not parse.
func (dc *DockerCompose) Emails(t *testing.T, to, path string) []Email {
	t.Helper()

	base := dc.stack.MailpitURL()
	require.NotEmpty(t, base, "only a cloud stack catches email")

	emails, err := matchingEmails(t.Context(), resty.New().SetBaseURL(base), to, path)
	require.NoError(t, err)

	return emails
}

func matchingEmails(ctx context.Context, client *resty.Client, to, path string) ([]Email, error) {
	search := new(mailpitSearch)

	resp, err := client.R().SetContext(ctx).
		SetQueryParam("query", fmt.Sprintf("to:%q", to)).
		SetResult(search).
		Get("/api/v1/search")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("searching Mailpit: %s", resp.Status())
	}

	var emails []Email

	for _, summary := range search.Messages {
		if !summary.sentTo(to) {
			continue
		}

		message := new(mailpitMessage)

		resp, err := client.R().SetContext(ctx).
			SetResult(message).
			Get("/api/v1/message/" + url.PathEscape(summary.ID))
		if err != nil {
			return nil, err
		}

		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("reading Mailpit message %s: %s", summary.ID, resp.Status())
		}

		doc, err := html.Parse(strings.NewReader(message.HTML))
		if err != nil {
			return nil, fmt.Errorf("parsing the HTML of Mailpit message %s: %w", summary.ID, err)
		}

		if linksTo(doc, path) {
			emails = append(emails, Email{html: doc})
		}
	}

	return emails, nil
}

func (s mailpitSummary) sentTo(address string) bool {
	return slices.ContainsFunc(s.To, func(to mailpitAddress) bool {
		return strings.EqualFold(to.Address, address)
	})
}

func linksTo(doc *html.Node, path string) bool {
	for node := range doc.Descendants() {
		if node.Type != html.ElementNode || node.DataAtom != atom.A {
			continue
		}

		if link, err := url.Parse(attribute(node, "href")); err == nil && link.Path == path {
			return true
		}
	}

	return false
}

func attribute(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}

	return ""
}

func textOf(node *html.Node) string {
	var text strings.Builder

	for child := range node.Descendants() {
		if child.Type == html.TextNode {
			text.WriteString(child.Data)
		}
	}

	return text.String()
}
