package main

import (
	"QuotesSite/internal/models"
	"context"
	"crypto/tls"
	"flag"
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

var sessionManager *scs.SessionManager

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	quotes         *models.QuoteModel
	users          *models.UserModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// DATABASE CONNECTION
	DBS := "postgres://postgres:1234567@localhost:5432/quotes"
	pool, error := pgxpool.Connect(context.Background(), DBS)
	if error != nil {
		errorLog.Fatalf("Unable to connection to database: %v\n", error)
	}
	defer pool.Close()
	infoLog.Printf("Connected")

	templateCache, errr := newTemplateCache()
	if errr != nil {
		errorLog.Fatal(errr)
	}

	formDecoder := form.NewDecoder()

	sessionManager = scs.New()
	sessionManager.Store = pgxstore.New(pool)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		quotes:         &models.QuoteModel{DB: pool},
		users:          &models.UserModel{DB: pool},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	infoLog.Printf("Starting server on %s", *addr)
	err := srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}
