package kivik

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"
)

func TestAttachmentBytes(t *testing.T) {
	tests := []struct {
		name     string
		att      *Attachment
		expected string
		err      string
	}{
		{
			name:     "read success",
			att:      NewAttachment("test.txt", "text/plain", ioutil.NopCloser(strings.NewReader("test content"))),
			expected: "test content",
		},
		{
			name: "buffered read",
			att: func() *Attachment {
				att := NewAttachment("test.txt", "text/plain", ioutil.NopCloser(strings.NewReader("test content")))
				_, _ = att.Bytes()
				return att
			}(),
			expected: "test content",
		},
		{
			name: "read error",
			att:  NewAttachment("test.txt", "text/plain", errReader("read error")),
			err:  "read error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.att.Bytes()
			testy.Error(t, test.err, err)
			if d := diff.Text(test.expected, string(result)); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestAttachmentRead(t *testing.T) {
	tests := []struct {
		name     string
		input    Attachment
		expected string
		status   int
		err      string
	}{
		{
			name:   "nil reader",
			input:  Attachment{},
			status: StatusUnknownError,
			err:    "kivik: attachment content not read",
		},
		{
			name:     "reader set",
			input:    Attachment{ReadCloser: ioutil.NopCloser(strings.NewReader("foo"))},
			expected: "foo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.input.Close() // nolint: errcheck
			result, err := ioutil.ReadAll(test.input)
			testy.StatusError(t, test.err, test.status, err)
			if d := diff.Text(test.expected, string(result)); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestAttachmentMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		att      *Attachment
		expected string
		err      string
	}{
		{
			name: "foo.txt",
			att: &Attachment{
				ReadCloser:  ioutil.NopCloser(strings.NewReader("test attachment\n")),
				Filename:    "foo.txt",
				ContentType: "text/plain",
			},
			expected: `{
				"content_type": "text/plain",
				"data": "dGVzdCBhdHRhY2htZW50Cg=="
			}`,
		},
		{
			name: "read error",
			att: &Attachment{
				ReadCloser:  ioutil.NopCloser(&errorReader{}),
				Filename:    "foo.txt",
				ContentType: "text/plain",
			},
			err: "json: error calling MarshalJSON for type *kivik.Attachment: errorReader",
		},
	}
	for _, test := range tests {
		result, err := json.Marshal(test.att)
		testy.Error(t, test.err, err)
		if d := diff.JSON([]byte(test.expected), result); d != nil {
			t.Error(d)
		}
	}
}
