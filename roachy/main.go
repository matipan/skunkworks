package main

import (
	"context"
	"dagger/roachy/internal/dagger"
	"strconv"
)

type Roachy struct{}

func (m *Roachy) RoastGroup(ctx context.Context, stravaAccessToken *dagger.Secret, webhookUrl *dagger.Secret, activityID int) (string, error) {
	athleteWorkspace := dag.AthleteWorkspace(stravaAccessToken, webhookUrl)
	return dag.Llm().
		WithAthleteWorkspace(athleteWorkspace).
		WithPromptVar("activity", strconv.Itoa(activityID)).
		WithPromptFile(dag.CurrentModule().Source().File("group-prompt.txt")).
		LastReply(ctx)
}
