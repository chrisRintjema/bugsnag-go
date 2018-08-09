package bugsnag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	uuid "github.com/satori/go.uuid"
)

type session struct {
	startedAt time.Time
	id        uuid.UUID
}

type notifierPayload struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Version string `json:"version"`
}

type appPayload struct {
	Type         string `json:"type"`
	ReleaseStage string `json:"releaseStage"`
	Version      string `json:"version"`
}

type devicePayload struct {
	OsName   string `json:"osName"`
	Hostname string `json:"hostname"`
}

type sessionCountsPayload struct {
	StartedAt       string `json:"startedAt"`
	SessionsStarted int    `json:"sessionsStarted"`
}

type sessionPayload struct {
	Notifier      notifierPayload      `json:"notifier"`
	App           appPayload           `json:"app"`
	Device        devicePayload        `json:"device"`
	SessionCounts sessionCountsPayload `json:"sessionCounts"`
}

func deliverSessions(sessions []session, config Configuration) error {
	sp := makeSessionPayload(sessions, config)
	buf, err := json.Marshal(sp)
	if err != nil {
		return fmt.Errorf("bugsnag/session.deliverSession unable to marshal json: %v", err)
	}
	client := http.Client{Transport: config.Transport}
	req, err := http.NewRequest("POST", config.Endpoints.Sessions, bytes.NewBuffer(buf))
	if err != nil {
		return fmt.Errorf("bugsnag/session.deliverSession unable to create request: %v", err)
	}
	for k, v := range bugsnagPrefixedHeaders(config.APIKey) {
		req.Header.Add(k, v)
	}
	_, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("bugsnag/session.deliverSession unable to deliver session: %v", err)
	}
	return nil
}

func makeSessionPayload(sessions []session, config Configuration) sessionPayload {
	releaseStage := config.ReleaseStage
	if releaseStage == "" {
		releaseStage = "production"
	}
	hostname := config.Hostname
	if hostname == "" {
		hostname, _ = os.Hostname() //Ignore the hostname if this call errors
	}

	return sessionPayload{
		Notifier: notifierPayload{
			Name:    "Bugsnag Go",
			URL:     "https://github.com/bugsnag/bugsnag-go",
			Version: VERSION,
		},
		App: appPayload{
			Type:         config.AppType,
			Version:      config.AppVersion,
			ReleaseStage: releaseStage,
		},
		Device: devicePayload{
			OsName:   runtime.GOOS,
			Hostname: hostname,
		},
		SessionCounts: sessionCountsPayload{
			StartedAt:       sessions[0].startedAt.UTC().Format(time.RFC3339),
			SessionsStarted: len(sessions),
		},
	}
}
