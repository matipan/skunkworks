package main

import (
	"context"
	"dagger/strava/internal/dagger"
	"io"
	"net/http"
	"strconv"
)

type Strava struct {
	// +private
	Token *dagger.Secret
}

func New(stravaAccessToken *dagger.Secret) *Strava {
	return &Strava{
		Token: stravaAccessToken,
	}
}

func (m *Strava) GetActivity(ctx context.Context, activityID int) (string, error) {
	return m.makeRequest(ctx, http.MethodGet, "https://www.strava.com/api/v3/activities/"+strconv.Itoa(activityID)+"?include_all_efforts=true")
}

func (m *Strava) GetClubActivities(ctx context.Context, clubID int) (string, error) {
	return m.makeRequest(ctx, http.MethodGet, "https://www.strava.com/api/v3/clubs/"+strconv.Itoa(clubID)+"/activities?per_page=30")
}

func (m *Strava) ListAthleteActivities(ctx context.Context) (string, error) {
	return m.makeRequest(ctx, http.MethodGet, "https://www.strava.com/api/v3/athlete/activities")
}

func (m *Strava) makeRequest(ctx context.Context, method, url string) (string, error) {
	r, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return "", err
	}
	token, err := m.Token.Plaintext(ctx)
	if err != nil {
		return "", err
	}

	r.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
