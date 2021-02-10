package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHttpNotifier(t *testing.T) {

	tests := []struct {
		name    string
		cfg     *HTTPConfig
		wantErr bool
	}{
		{
			name: "valid GET configuration",
			cfg: &HTTPConfig{
				Method:           http.MethodGet,
				notificationType: OnStart,
				URLS: NotificationURL{
					OnStart: "http://localhost/start",
				},
			},
		},
		{
			name: "valid POST configuration",
			cfg: &HTTPConfig{
				Method:           http.MethodPost,
				notificationType: OnStart,
				URLS: NotificationURL{
					OnStart: "http://localhost/start",
				},
			},
		},
		{
			name: "valid PUT configuration",
			cfg: &HTTPConfig{
				Method:           http.MethodPut,
				notificationType: OnStart,
				URLS: NotificationURL{
					OnStart: "http://localhost/start",
				},
			},
		},
		{
			name:    "invalid unsupported method",
			wantErr: true,
			cfg: &HTTPConfig{
				Method:           http.MethodTrace,
				notificationType: OnStart,
				URLS: NotificationURL{
					OnStart: "http://localhost/start",
				},
			},
		},
		{
			name:    "empty url",
			wantErr: true,
			cfg: &HTTPConfig{
				Method:           http.MethodPost,
				notificationType: OnStart,
			},
		},
		{
			name:    "bad url",
			wantErr: true,
			cfg: &HTTPConfig{
				Method:           http.MethodPost,
				notificationType: OnStart,
				URLS: NotificationURL{
					OnStart: "http//localhost/start",
				},
			},
		},
		{
			name:    "missing url",
			wantErr: true,
			cfg: &HTTPConfig{
				Method:           http.MethodPost,
				notificationType: OnStart,
				URLS: NotificationURL{
					OnFailure: "http://localhost/start",
				},
			},
		},
		{
			name:    "bad body template",
			wantErr: true,
			cfg: &HTTPConfig{
				Method:           http.MethodPost,
				notificationType: OnStart,
				URLS: NotificationURL{
					OnStart: "http://localhost/start",
				},
				Body: "{ \"Event\" : \"{{.Event}\" }",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHTTPNotifier(tt.cfg)
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Nil(t, got)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, got)
		})
	}
}

type httpFuncTester func(notifier *HTTPNotifier) error

func Test_httpNotifier_NotifyOfStart(t *testing.T) {
	cmdConfig = "notify-start-config"
	notifyTest(t, "onStart", func(n *HTTPNotifier) error {
		return n.NotifyOfStart()
	})
}

func Test_httpNotifier_NotifyOfSkip(t *testing.T) {
	cmdConfig = "notify-skip-config"
	notifyTest(t, "onSkip", func(n *HTTPNotifier) error {
		return n.NotifyOfSkip()
	})
}

func Test_httpNotifier_NotifyOfSuccess(t *testing.T) {
	cmdConfig = "notify-success-config"
	notifyTest(t, "onSuccess", func(n *HTTPNotifier) error {
		return n.NotifyOfSuccess()
	})
}

func Test_httpNotifier_NotifyOfFailure(t *testing.T) {
	cmdConfig = "notify-failure-config"
	notifyTest(t, "onFailure", func(n *HTTPNotifier) error {
		return n.NotifyOfFailure()
	})
}

func notifyTest(t *testing.T, event NotificationType, notify httpFuncTester) {
	tests := []struct {
		name     string
		wantErr  bool
		config   *HTTPConfig
		wantBody string
	}{
		{
			name: "successful unauthenticated notification",
			config: &HTTPConfig{
				Method:              "GET",
				AcceptInsecureCerts: true,
				notificationType:    event,
			},
		},
		{
			name: "successful authenticated notification",
			config: &HTTPConfig{
				Method:              "GET",
				AcceptInsecureCerts: true,
				Username:            "someuser",
				Password:            "somepassword",
				notificationType:    event,
			},
		},
		{
			name: "successful POST notification",
			config: &HTTPConfig{
				Method:              "POST",
				AcceptInsecureCerts: true,
				notificationType:    event,
			},
			wantBody: fmt.Sprintf("{\n  \"event\": \"%s\",\n  \"configName\": \"%s\"\n}", event, cmdConfig),
		},
		{
			name: "successful PUT notification",
			config: &HTTPConfig{
				Method:              "PUT",
				AcceptInsecureCerts: true,
				notificationType:    event,
			},
			wantBody: fmt.Sprintf("{\n  \"event\": \"%s\",\n  \"configName\": \"%s\"\n}", event, cmdConfig),
		},
		{
			name: "errored notification",
			config: &HTTPConfig{
				Method:              "GET",
				AcceptInsecureCerts: true,
				notificationType:    event,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.wantErr {
					http.Error(w, "error", http.StatusInternalServerError)
					return
				}
				if tt.config.Username != "" {
					auth, password, ok := r.BasicAuth()
					assert.True(t, ok, "basic auth enabled")
					assert.Equal(t, tt.config.Username, auth)
					assert.Equal(t, tt.config.Password, password)
				}
				assert.Equal(t, tt.config.Method, r.Method)
				// Test Body
				body, _ := ioutil.ReadAll(r.Body)
				assert.Equal(t, tt.wantBody, string(body))
			}))
			defer ts.Close()
			tt.config.URLS = NotificationURL{event: ts.URL}
			h, err := NewHTTPNotifier(tt.config)
			require.NoError(t, err)

			err = notify(h)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewHTTPNotifierConfig(t *testing.T) {
	tests := []struct {
		name             string
		notificationType NotificationType
		want             *HTTPConfig
		wantErr          bool
		fileName         string
	}{
		{
			name:             "full valid configuration - default config",
			fileName:         "test/assets/globalConfigs/fullValidConfig.yml",
			notificationType: OnSuccess,
			want: &HTTPConfig{
				Method:              "POST",
				Username:            "basicusername",
				Password:            "basicpassword",
				Body:                "{\n  \"event\": \"{{.Event}}\",\n  \"configName\": \"{{.ConfigName}}\"\n}\n",
				AcceptInsecureCerts: true,
				notificationType:    OnSuccess,
				URLS: NotificationURL{
					OnSuccess: "http://localhost/some_guid",
					OnStart:   "http://localhost/some_guid/start",
					OnFailure: "http://localhost/some_guid/fail",
					OnSkipped: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quietFlag = true
			runningUnitTests = true
			defer func() {
				quietFlag = false
				runningUnitTests = false
			}()
			err := loadGlobalConfig(".", "test/assets/globalConfigs/fullValidConfig.yml")
			require.NoError(t, err)

			got, err := NewHTTPNotifierConfig(viper.GetViper(), tt.notificationType)
			if tt.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
