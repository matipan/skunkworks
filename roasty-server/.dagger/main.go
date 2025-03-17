package main

import (
	"context"
	"dagger/dev/internal/dagger"
	"fmt"
	"path"
	"time"

	"github.com/pelletier/go-toml"
)

const (
	DaggerVersion = "0.17.0-llm.8"
	GoVersion     = "1.23.3"
	FlyVersion    = "0.3.85"
)

type Dev struct {
	Source *dagger.Directory
}

func New(
	// +defaultPath="."
	source *dagger.Directory,
) *Dev {
	return &Dev{Source: source}
}

func (m *Dev) Bin() *dagger.File {
	baseImage := dag.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).From("golang:1")
	return dag.Go(dagger.GoOpts{Base: baseImage}).
		FromVersion(GoVersion).
		Build(m.Source, dagger.GoBuildOpts{
			Static: true,
		}).
		File("roasty-server")
}

func (m *Dev) App() *dagger.Container {
	return dag.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).
		From(fmt.Sprintf("registry.dagger.io/engine:v%s", DaggerVersion)).
		WithEnvVariable("DAGGER_VERSION", DaggerVersion).
		WithEnvVariable("_EXPERIMENTAL_DAGGER_RUNNER_HOST", "unix:///var/run/buildkit/buildkitd.sock").
		WithFile("/usr/local/bin/roasty", m.Bin()).
		WithFile("/usr/local/bin/goreman", m.GoremanBin()).
		WithWorkdir("/app").
		WithFile("/app/Procfile", m.Source.File("Procfile")).
		WithEnvVariable("PORT", "8080").
		WithEnvVariable("MIGRATE", "true").
		WithExposedPort(8080).
		WithoutDefaultArgs().
		WithEntrypoint([]string{"goreman", "--set-ports=false", "start"})
}

func (m *Dev) GoremanBin() *dagger.File {
	baseGolangImageAMD64 := dag.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).From("golang:1")
	return dag.Go(dagger.GoOpts{Base: baseGolangImageAMD64}).
		FromVersion(GoVersion).
		Base().
		WithEnvVariable("CGO_ENABLED", "0").
		WithEnvVariable("GOBIN", "/bin").
		WithExec([]string{"go", "install", "github.com/mattn/goreman@v0.3.15"}).
		File("/bin/goreman")
}

func (m *Dev) PublishAndDeploy(ctx context.Context, token *dagger.Secret) error {
	return publishAndDeploy(ctx, token, m.Source, m.App(), "--vm-size", "shared-cpu-4x", "--vm-memory", "8192")
}

func publishAndDeploy(ctx context.Context, token *dagger.Secret, src *dagger.Directory, ctr *dagger.Container, deployArgs ...string) error {
	tomlContent, err := src.File("fly.toml").Contents(ctx)
	if err != nil {
		return err
	}

	tomlTree, err := toml.Load(tomlContent)
	if err != nil {
		return err
	}

	appName := tomlTree.Get("app").(string)

	registry := "registry.fly.io"
	pinnedRef, err := ctr.
		WithRegistryAuth(registry, "x", token).
		Publish(ctx, path.Join(registry, appName))
	if err != nil {
		return err
	}

	tomlTree.SetPath([]string{"build", "image"}, pinnedRef)

	src = src.WithNewFile("fly.toml", tomlTree.String())

	return flyDeploy(ctx, src, token, deployArgs...)
}

func fly(token *dagger.Secret, args ...string) *dagger.Container {
	return flyIn(token, dag.Directory(), args...)
}

func flyDeploy(ctx context.Context, src *dagger.Directory, token *dagger.Secret, args ...string) error {
	_, err := flyIn(token, src, append([]string{"deploy"}, args...)...).Sync(ctx)
	return err
}

func flyIn(token *dagger.Secret, dir *dagger.Directory, args ...string) *dagger.Container {
	return dag.Container().
		From("flyio/flyctl:v"+FlyVersion).
		WithSecretVariable("FLY_API_TOKEN", token).
		WithMountedDirectory("/app", dir).
		WithWorkdir("/app").
		WithEnvVariable("LOG_LEVEL", "debug").
		WithEnvVariable("BUST", time.Now().String()).
		WithExec(args, dagger.ContainerWithExecOpts{UseEntrypoint: true})
}
