package qbittorrent_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/download"
	_ "github.com/woliveiras/bookaneer/internal/download/qbittorrent"
)

func TestClientTest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		giveStatus int
		giveBody   string
		wantErr    error
	}{
		{name: "qBittorrent 5.2 success", giveStatus: http.StatusNoContent},
		{name: "legacy success", giveStatus: http.StatusOK, giveBody: "Ok."},
		{name: "legacy success with whitespace", giveStatus: http.StatusOK, giveBody: "Ok.\n"},
		{name: "legacy invalid credentials", giveStatus: http.StatusOK, giveBody: "Fails.", wantErr: download.ErrAuthFailed},
		{name: "invalid credentials", giveStatus: http.StatusUnauthorized, wantErr: download.ErrAuthFailed},
		{name: "banned address", giveStatus: http.StatusForbidden, wantErr: download.ErrAuthFailed},
		{name: "empty legacy response", giveStatus: http.StatusOK, wantErr: download.ErrAuthFailed},
		{name: "unexpected response containing Ok", giveStatus: http.StatusOK, giveBody: "<html>Ok</html>", wantErr: download.ErrAuthFailed},
		{name: "unexpected success status", giveStatus: http.StatusAccepted, giveBody: "Ok.", wantErr: download.ErrAuthFailed},
		{name: "server error", giveStatus: http.StatusInternalServerError, giveBody: "Ok.", wantErr: download.ErrAuthFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/v2/auth/login", r.URL.Path)
				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
				r.Body = http.MaxBytesReader(w, r.Body, 1024)
				if !assert.NoError(t, r.ParseForm()) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				assert.Equal(t, "test-user", r.Form.Get("username"))
				assert.Equal(t, "test+password&=", r.Form.Get("password"))
				w.WriteHeader(tt.giveStatus)
				if tt.giveBody != "" {
					_, err := io.WriteString(w, tt.giveBody)
					assert.NoError(t, err)
				}
			}))
			t.Cleanup(server.Close)

			err := newTestClient(t, server).Test(t.Context())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClientGetQueuePreservesSession(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		giveStatus     int
		giveBody       string
		giveCookieName string
	}{
		{name: "qBittorrent 5.2", giveStatus: http.StatusNoContent, giveCookieName: "QBT_SID_8081"},
		{name: "legacy", giveStatus: http.StatusOK, giveBody: "Ok.", giveCookieName: "SID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v2/auth/login":
					http.SetCookie(w, &http.Cookie{Name: tt.giveCookieName, Value: "test-session", Path: "/", HttpOnly: true})
					w.WriteHeader(tt.giveStatus)
					if tt.giveBody != "" {
						_, err := io.WriteString(w, tt.giveBody)
						assert.NoError(t, err)
					}
				case "/api/v2/torrents/info":
					cookie, err := r.Cookie(tt.giveCookieName)
					if !assert.NoError(t, err) || !assert.Equal(t, "test-session", cookie.Value) {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					assert.Equal(t, http.MethodGet, r.Method)
					w.Header().Set("Content-Type", "application/json")
					_, err = io.WriteString(w, `[{"hash":"test-hash","name":"Test book","state":"downloading","progress":0.5}]`)
					assert.NoError(t, err)
				default:
					t.Errorf("unexpected request: %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			t.Cleanup(server.Close)

			queue, err := newTestClient(t, server).GetQueue(t.Context())
			require.NoError(t, err)
			require.Len(t, queue, 1)
			assert.Equal(t, "test-hash", queue[0].ID)
			assert.Equal(t, float64(50), queue[0].Progress)
		})
	}
}

func TestClientTestTruncatedResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "10")
		_, err := io.WriteString(w, "Ok.")
		assert.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	err := newTestClient(t, server).Test(t.Context())
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

func newTestClient(t *testing.T, server *httptest.Server) download.Client {
	t.Helper()

	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	port, err := strconv.Atoi(u.Port())
	require.NoError(t, err)

	client, err := download.NewClient(download.ClientConfig{
		Type:     download.ClientTypeQBittorrent,
		Host:     u.Hostname(),
		Port:     port,
		Username: "test-user",
		Password: "test+password&=",
	})
	require.NoError(t, err)
	return client
}
