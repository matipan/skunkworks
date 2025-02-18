package main

import (
	"context"
	"dagger/athlete-workspace/internal/dagger"
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

func (m *AthleteWorkspace) GetActivity(ctx context.Context, activityID int) (string, error) {
	return dag.Strava(m.StravaToken).GetActivity(ctx, activityID)
}

func (m *AthleteWorkspace) GetClubActivities(ctx context.Context, clubID int) (string, error) {
	return dag.Strava(m.StravaToken).GetClubActivities(ctx, clubID)
}

func (m *AthleteWorkspace) ListAthleteActivities(ctx context.Context) (string, error) {
	return dag.Strava(m.StravaToken).ListAthleteActivities(ctx)
}
