package main

import (
	"WalletScraper/internal/config"
	"WalletScraper/internal/process"
	"WalletScraper/internal/rabbitmq"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"time"

	"github.com/Bezunca/mongo_connection"
)

func main() {
	// Loading configs
	configs := config.New()

	caChainBytes, err := ioutil.ReadFile(configs.CAFile)
	if err != nil {
		log.Fatal(err)
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caChainBytes)
	if !ok {
		log.Fatal("unable to parse CA Chain file")
	}

	tlsConfig := &tls.Config{
		RootCAs: roots,
	}

	rabbitMQ, err := rabbitmq.New(&configs.RabbitMQ, tlsConfig)
	if err != nil {
		log.Fatal(err)
	}

	mongoClient, err := mongo_connection.New(&configs.MongoDB, tlsConfig)
	if err != nil {
		log.Fatal(err)
	}

	i := 0
	for ;;i++{
		<-time.After(1 * time.Second)
		if rabbitMQ.IsReady(){
			break
		}
		if i == 10{
			log.Fatal("Cannot Connect to RMQ")
		}
	}
	process.Run(rabbitMQ, mongoClient)
}
