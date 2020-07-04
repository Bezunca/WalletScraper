package main

import (
	"WalletScraper/internal/config"
	"WalletScraper/internal/process"
	"WalletScraper/internal/rabbitmq"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"

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

	process.Run(rabbitMQ, mongoClient)
}
