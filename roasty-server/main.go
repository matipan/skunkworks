package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"

	"dagger.io/dagger"
)

var (
	agentModule = flag.String("agent-mod", "github.com/matipan/skunkworks/roasty", "module to serve as agent")
	port        = flag.Int("port", 8080, "port to serve the agent on")

	mu   sync.RWMutex
	done bool
)

func main() {
	flag.Parse()

	ctx := context.TODO()
	dag, err := dagger.Connect(ctx,
		dagger.WithRunnerHost(os.Getenv("_EXPERIMENTAL_DAGGER_RUNNER_HOST")),
		dagger.WithVerbosity(10),
		dagger.WithLogOutput(os.Stdout),
	)
	if err != nil {
		log.Fatalf("failed to connect to dagger engine: %v", err)
	}

	go func() {
		mu.Lock()
		defer func() {
			slog.Info("done loading module")
			done = true
			mu.Unlock()
		}()

		// serve the module before starting the backend
		err = dag.ModuleSource(*agentModule).
			AsModule().
			Serve(ctx)
		if err != nil {
			log.Fatalf("failed to serve module: %v", err)
		}
	}()

	stravaAccessToken := dag.SetSecret("stravaAccessToken", os.Getenv("STRAVA_ACCESS_TOKEN"))
	webhookUrl := dag.SetSecret("webhookUrl", os.Getenv("WEBHOOK_URL"))

	slog.Info("module loaded")

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/strava", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		if !done {
			mu.RUnlock()
			return
		}
		mu.RUnlock()

		activityID := r.URL.Query().Get("activity-id")
		aid, err := strconv.Atoi(activityID)
		if err != nil {
			http.Error(w, "invalid activity id", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)

		go func() {
			resp := &dagger.Response{}
			if err := dag.Do(ctx, &dagger.Request{
				Query: `query RoastGroup($stravaAccessToken: SecretID!, $webhookUrl: SecretID!, $activityId: Int!) {
	roasty { 
		roastGroup(stravaAccessToken: $stravaAccessToken, webhookUrl: $webhookUrl, activityId: $activityId)
	}
}`,
				Variables: map[string]interface{}{
					"stravaAccessToken": stravaAccessToken,
					"webhookUrl":        webhookUrl,
					"activityId":        aid,
				},
			}, resp); err != nil {
				log.Fatalf("failed to execute query: %v", err)
			}

			slog.Info("model evaluated")
		}()

	})

	slog.Info("starting API", "port", fmt.Sprintf(":%d", *port))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
