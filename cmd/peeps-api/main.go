package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"expvar"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/mattlaver/peeps/internal/platform/auth"
	"github.com/mattlaver/peeps/internal/platform/db"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattlaver/peeps/cmd/peeps-api/handlers"

	"github.com/mattlaver/peeps/internal/platform/flag"
)

var build = "develop"

func main() {
	log := log.New(os.Stdout, "CONTACTS : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)


	var cfg struct {
		Web struct {
			APIHost         string        `default:"0.0.0.0:3000" envconfig:"API_HOST"`
			DebugHost       string        `default:"0.0.0.0:4000" envconfig:"DEBUG_HOST"`
			ReadTimeout     time.Duration `default:"5s" envconfig:"READ_TIMEOUT"`
			WriteTimeout    time.Duration `default:"5s" envconfig:"WRITE_TIMEOUT"`
			ShutdownTimeout time.Duration `default:"5s" envconfig:"SHUTDOWN_TIMEOUT"`
		}
		DB struct {
			DialTimeout time.Duration `default:"5s" envconfig:"DIAL_TIMEOUT"`
			Host        string        `default:"localhost:27017/gotraining" envconfig:"HOST"`
		}
		Trace struct {
			Host         string        `default:"http://tracer:3002/v1/publish" envconfig:"HOST"`
			BatchSize    int           `default:"1000" envconfig:"BATCH_SIZE"`
			SendInterval time.Duration `default:"15s" envconfig:"SEND_INTERVAL"`
			SendTimeout  time.Duration `default:"500ms" envconfig:"SEND_TIMEOUT"`
		}
		Auth struct {
			KeyID          string `default:"1" envconfig:"KEY_ID"`
			PrivateKeyFile string `default:"private.pem" envconfig:"PRIVATE_KEY_FILE"`
			Algorithm      string `default:"RS256" envconfig:"ALGORITHM"`
		}
	}

	if err := envconfig.Process("SALES", &cfg); err != nil {
		log.Fatalf("main : Parsing Config : %v", err)
	}

	if err := flag.Process(&cfg); err != nil {
		if err != flag.ErrHelp {
			log.Fatalf("main : Parsing Command Line : %v", err)
		}
		return // We displayed help.
	}

	expvar.NewString("build").Set(build)
	log.Printf("main : Started : Application Initializing version %q", build)
	defer log.Println("main : Completed")


	cfgJSON, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		log.Fatalf("main : Marshalling Config to JSON : %v", err)
	}

	log.Printf("main : Config : %v\n", string(cfgJSON))

	// =========================================================================
	// Find auth keys

	keyContents, err := ioutil.ReadFile(cfg.Auth.PrivateKeyFile)
	if err != nil {
		log.Fatalf("main : Reading auth private key : %v", err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyContents)
	if err != nil {
		log.Fatalf("main : Parsing auth private key : %v", err)
	}

	publicKeyLookup := auth.NewSingleKeyFunc(cfg.Auth.KeyID, key.Public().(*rsa.PublicKey))

	authenticator, err := auth.NewAuthenticator(key, cfg.Auth.KeyID, cfg.Auth.Algorithm, publicKeyLookup)
	if err != nil {
		log.Fatalf("main : Constructing authenticator : %v", err)
	}


	// =========================================================================
	// Start Mongo

	log.Println("main : Started : Initialize Mongo")
	masterDB, err := db.New(cfg.DB.Host, cfg.DB.DialTimeout)
	if err != nil {
		log.Fatalf("main : Register DB : %v", err)
	}
	defer masterDB.Close()


	// =========================================================================
	// Start API Service

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:           cfg.Web.APIHost,
		Handler:        handlers.API(shutdown, log, masterDB, authenticator),
		ReadTimeout:    cfg.Web.ReadTimeout,
		WriteTimeout:   cfg.Web.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("main : API Listening %s", cfg.Web.APIHost)
		serverErrors <- api.ListenAndServe()
	}()



	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("main : Error starting server: %v", err)

	case sig := <-shutdown:
		log.Printf("main : %v : Start shutdown..", sig)

		// Create context for Shutdown call.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		err := api.Shutdown(ctx)
		if err != nil {
			log.Printf("main : Graceful shutdown did not complete in %v : %v", cfg.Web.ShutdownTimeout, err)
			err = api.Close()
		}

		// Log the status of this shutdown.
		switch {
		case sig == syscall.SIGSTOP:
			log.Fatal("main : Integrity issue caused shutdown")
		case err != nil:
			log.Fatalf("main : Could not stop server gracefully : %v", err)
		}
	}



	fmt.Println("hello world")
}
