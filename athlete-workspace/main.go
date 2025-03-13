package main

import (
	"context"
	"dagger/athlete-workspace/internal/dagger"
	"strconv"
)

type AthleteWorkspace struct {
	// +private
	StravaToken *dagger.Secret
	// +private
	DiscordWebhookURL *dagger.Secret
}

func New(stravaAccessToken, discordWebhookUrl *dagger.Secret) *AthleteWorkspace {
	return &AthleteWorkspace{
		StravaToken:       stravaAccessToken,
		DiscordWebhookURL: discordWebhookUrl,
	}
}

func (m *AthleteWorkspace) NotifyDiscord(ctx context.Context, message string) (string, error) {
	return dag.Notify().Discord(ctx, m.DiscordWebhookURL, message)
}

func (m *AthleteWorkspace) GetActivity(ctx context.Context, activityID string) (string, error) {
	aid, err := strconv.Atoi(activityID)
	if err != nil {
		return "", err
	}
	return dag.Strava(m.StravaToken).GetActivity(ctx, aid)
}

func (m *AthleteWorkspace) GetClubActivities(ctx context.Context, clubID string) (string, error) {
	cid, err := strconv.Atoi(clubID)
	if err != nil {
		return "", err
	}
	return dag.Strava(m.StravaToken).GetClubActivities(ctx, cid)
}

func (m *AthleteWorkspace) ListAthleteActivities(ctx context.Context) (string, error) {
	return dag.Strava(m.StravaToken).ListAthleteActivities(ctx)
}
