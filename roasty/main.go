package main

import (
	"context"
	"dagger/roasty/internal/dagger"
	"strconv"
)

type Roasty struct{}

func (m *Roasty) RoastGroup(ctx context.Context, stravaAccessToken *dagger.Secret, webhookUrl *dagger.Secret, activityID int) (string, error) {
	athleteWorkspace := dag.AthleteWorkspace(stravaAccessToken, webhookUrl)
	return dag.Llm().
		WithAthleteWorkspace(athleteWorkspace).
		WithPromptVar("activity", strconv.Itoa(activityID)).
		WithPromptFile(dag.CurrentModule().Source().File("prompts/roast-group.txt")).
		LastReply(ctx)
}
