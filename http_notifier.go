package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"text/template"

	"github.com/spf13/viper"
)

var _ Notifier = &HTTPNotifier{}

const defaultBodyTemplate = `{
  "event": "{{.Event}}",
  "configName": "{{.ConfigName}}"
}`

// HTTPNotifier is a Notifier that supports making HTTP requests using GET, POST, and PUT methods.
type HTTPNotifier struct {
	config *HTTPConfig
	client http.Client
}

// NotificationURL is the type used to map from notification types to URLS.
type NotificationURL map[NotificationType]string

// HTTPConfig is the configuration used by the HTTPNotifier when sending notifications.
type HTTPConfig struct {
	Method              string
	URLS                NotificationURL
	Username            string
	Password            string
	Body                string
	AcceptInsecureCerts bool
	bodyTemplate        *template.Template
	notificationType    NotificationType
}

type notificationData struct {
	Event      string
	ConfigName string
}

// NewHTTPNotifierConfig tries to find an HTTP configuration for the notificationType. If one is not found, it falls back
// to the standard http configuration
func NewHTTPNotifierConfig(v *viper.Viper, notificationType NotificationType) (*HTTPConfig, error) {
	cfg := new(HTTPConfig)
	cfg.notificationType = notificationType
	err := v.UnmarshalKey("http", cfg)
	return cfg, err
}

// NewHTTPNotifier creates a Notifier that supports making HTTP requests using GET, POST, and PUT methods. The HTTP body
// in the configuration struct cfg can contain template variables that will be evaluated when the notifications are sent.
// The data passed to the template variables are named Event and ConfigName. Event represents the type of notification,
// ConfigName represents the name of the current configuration.
func NewHTTPNotifier(cfg *HTTPConfig) (*HTTPNotifier, error) {
	// Validate supported methods
	if cfg.Method != http.MethodPost && cfg.Method != http.MethodGet && cfg.Method != http.MethodPut {
		return nil, fmt.Errorf("http method %s not supported. Supported methods are %s, %s, %s",
			cfg.Method, http.MethodPost, http.MethodPut, http.MethodGet)
	}
	// Validate url has event
	urlValue, ok := cfg.URLS[cfg.notificationType]
	if !ok {
		return nil, fmt.Errorf("URL missing for event type %s", cfg.notificationType)
	}
	if _, err := url.ParseRequestURI(urlValue); err != nil {
		return nil, fmt.Errorf("URL %s for %s is invalid: %w", urlValue, cfg.notificationType, err)
	}

	bodyTmplContent := cfg.Body
	if bodyTmplContent == "" {
		bodyTmplContent = defaultBodyTemplate
	}
	bodyTmpl, err := template.New("bodyTemplate").Parse(bodyTmplContent)
	if err != nil {
		return nil, fmt.Errorf("could not parse body template: %w", err)
	}
	cfg.bodyTemplate = bodyTmpl

	return &HTTPNotifier{
		config: cfg,
		client: http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: cfg.AcceptInsecureCerts,
				},
			},
		},
	}, nil
}

// NotifyOfStart makes a HTTP request to the configured URL using the configured Method. If the request is of method type
// POST or PUT, the Body template is expanded and used in the request. The Event and ConfigName are made available for
// template expansion.
//
// This method emits an Event with value "onStart"
func (h HTTPNotifier) NotifyOfStart() error {
	return httpNotify(h.client, h.config)
}

// NotifyOfSkip makes a HTTP request to the configured URL using the configured Method. If the request is of method type
// POST or PUT, the Body template is expanded and used in the request. The Event and ConfigName are made available for
// template expansion.
//
// This method emits an Event with value "onSkip"
func (h HTTPNotifier) NotifyOfSkip() error {
	return httpNotify(h.client, h.config)
}

// NotifyOfSuccess makes a HTTP request to the configured URL using the configured Method. If the request is of method type
// POST or PUT, the Body template is expanded and used in the request. The Event and ConfigName are made available for
// template expansion.
//
// This method emits an Event with value "onSuccess"
func (h HTTPNotifier) NotifyOfSuccess() error {
	return httpNotify(h.client, h.config)
}

// NotifyOfFailure makes a HTTP request to the configured URL using the configured Method. If the request is of method type
// POST or PUT, the Body template is expanded and used in the request. The Event and ConfigName are made available for
// template expansion.
//
// This method emits an Event with value "onFailure"
func (h HTTPNotifier) NotifyOfFailure() error {
	return httpNotify(h.client, h.config)
}

func httpNotify(client http.Client, cfg *HTTPConfig) error {
	data := notificationData{
		Event:      string(cfg.notificationType),
		ConfigName: cmdConfig,
	}
	urlValue := cfg.URLS[cfg.notificationType]

	body := new(bytes.Buffer)
	if cfg.Method == http.MethodPost || cfg.Method == http.MethodPut {
		if err := cfg.bodyTemplate.Execute(body, data); err != nil {
			return err
		}
	}
	request, err := http.NewRequest(cfg.Method, urlValue, body)
	if err != nil {
		return err
	}
	if cfg.Username != "" {
		request.SetBasicAuth(cfg.Username, cfg.Password)
	}
	var response *http.Response
	if response, err = client.Do(request); err != nil {
		logError(nil, fmt.Sprint("Http notification error:", err))
		return err
	}
	msg := fmt.Sprintf("Http notification sent Status='%d' Event='%s' Method='%s' URL='%s' WithAuth='%t'",
		response.StatusCode, cfg.notificationType, cfg.Method, urlValue, cfg.Username != "")
	if response.StatusCode >= http.StatusBadRequest {
		logError(nil, msg)
		return fmt.Errorf("http notification returned %d", response.StatusCode)
	}
	logMessage(nil, msg)
	return nil
}
