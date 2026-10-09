package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"net.daoke/orp-backend/internal/config"
	"net.daoke/orp-backend/internal/httpapi"
	"net.daoke/orp-backend/internal/store"
)

func main() {
	configuration, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	database, err := store.Open(configuration.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	if err := store.Migrate(database); err != nil {
		log.Fatal(err)
	}
	if err := store.SeedAdmin(database); err != nil {
		log.Fatal(err)
	}
	if err := store.ImportLegacyCenters(database); err != nil {
		log.Fatal(err)
	}
	if err := store.SeedSettingGroups(database); err != nil {
		log.Fatal(err)
	}
	if err := httpapi.RecoverPendingPublishes(database); err != nil {
		log.Fatal(err)
	}
	httpapi.StartNodeSampler(context.Background(), database)
	httpapi.StartTLSExpiryAlertScheduler(context.Background(), database)
	redisClient := redis.NewClient(&redis.Options{
		Addr: configuration.RedisAddress, Password: configuration.RedisPassword, DB: configuration.RedisDB,
		DialTimeout: 2 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 2 * time.Second,
	})
	redisCtx, cancelRedis := context.WithTimeout(context.Background(), 2*time.Second)
	redisErr := redisClient.Ping(redisCtx).Err()
	cancelRedis()
	if redisErr != nil {
		log.Printf("Redis unavailable; dashboard will use direct snapshots: %v", redisErr)
		_ = redisClient.Close()
		redisClient = nil
	} else {
		defer redisClient.Close()
	}
	httpapi.StartKafkaLogConsumer(context.Background(), database, httpapi.KafkaLogConfig{
		Brokers: configuration.KafkaBrokers,
		Topic:   configuration.KafkaTopic,
		GroupID: configuration.KafkaConsumerID,
	})
	server := &http.Server{Addr: configuration.HTTPAddress, Handler: httpapi.NewWithRedis(database, redisClient)}
	log.Printf("openresty-plus Go control plane listening on %s", configuration.HTTPAddress)
	if configuration.AgentHTTPSAddress != "" {
		caPEM, err := os.ReadFile(configuration.AgentClientCAFile)
		if err != nil {
			log.Fatalf("read Agent client CA: %v", err)
		}
		clientCAs := x509.NewCertPool()
		if !clientCAs.AppendCertsFromPEM(caPEM) {
			log.Fatal("Agent client CA contains no certificates")
		}
		agentServer := &http.Server{Addr: configuration.AgentHTTPSAddress, Handler: httpapi.NewAgentMux(database), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: clientCAs}}
		go func() {
			log.Printf("openresty-plus Agent mTLS listener on %s", configuration.AgentHTTPSAddress)
			log.Fatal(agentServer.ListenAndServeTLS(configuration.TLSCertFile, configuration.TLSKeyFile))
		}()
	}
	log.Fatal(server.ListenAndServe())
}
