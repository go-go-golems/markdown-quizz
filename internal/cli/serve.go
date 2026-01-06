package cli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-go-golems/XXX/internal/db"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	pkgerrors "github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type ServerSettings struct {
	Host       string `glazed.parameter:"host"`
	Port       int    `glazed.parameter:"port"`
	SQLitePath string `glazed.parameter:"sqlite-path"`
	StaticDir  string `glazed.parameter:"static-dir"`
	LogLevel   string `glazed.parameter:"log-level"`
}

type ServeCommand struct {
	*cmds.CommandDefinition
}

func NewServeCommand() (*ServeCommand, error) {
	serverSection, err := schema.NewSection(
		"server",
		"Server",
		schema.WithDescription("HTTP server configuration"),
		schema.WithFields(
			fields.New("host", fields.TypeString,
				fields.WithHelp("Server listen host"),
				fields.WithDefault("127.0.0.1"),
			),
			fields.New("port", fields.TypeInteger,
				fields.WithHelp("Server listen port"),
				fields.WithDefault(8080),
			),
			fields.New("sqlite-path", fields.TypeString,
				fields.WithHelp("Path to sqlite database file"),
				fields.WithDefault("markdown-quizz.sqlite"),
			),
			fields.New("static-dir", fields.TypeString,
				fields.WithHelp("If set, serve static assets from this directory"),
				fields.WithDefault(""),
			),
			fields.New("log-level", fields.TypeChoice,
				fields.WithHelp("Log level"),
				fields.WithChoices("debug", "info", "warn", "error"),
				fields.WithDefault("info"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(
		schema.WithSections(serverSection),
	)

	desc := cmds.NewCommandDefinition(
		"serve",
		cmds.WithShort("Run the markdown-quizz HTTP server"),
		cmds.WithLong("Starts the markdown-quizz HTTP server (API + optional static assets)."),
		cmds.WithSchema(s),
	)

	return &ServeCommand{CommandDefinition: desc}, nil
}

var _ cmds.BareCommand = &ServeCommand{}

func (c *ServeCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	serverSettings := &ServerSettings{}
	if err := values.DecodeSectionInto(parsedLayers, "server", serverSettings); err != nil {
		return pkgerrors.Wrap(err, "failed to decode server settings")
	}

	runCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	sqliteDB, err := db.OpenSQLite(runCtx, db.SQLiteOptions{
		Path: serverSettings.SQLitePath,
	})
	if err != nil {
		return pkgerrors.Wrap(err, "open sqlite")
	}
	defer func() {
		_ = sqliteDB.Close()
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	trpcHandler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf("tRPC adapter not implemented yet (%s %s)", r.Method, r.URL.Path), http.StatusNotImplemented)
	}
	mux.HandleFunc("/api/trpc", trpcHandler)
	mux.HandleFunc("/api/trpc/", trpcHandler)

	addr := fmt.Sprintf("%s:%d", serverSettings.Host, serverSettings.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	group, groupCtx := errgroup.WithContext(runCtx)

	group.Go(func() error {
		err := srv.ListenAndServe()
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return pkgerrors.Wrap(err, "http server failed")
	})

	group.Go(func() error {
		<-groupCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	})

	if err := group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
