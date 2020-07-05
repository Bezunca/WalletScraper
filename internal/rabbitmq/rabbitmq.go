package rabbitmq

import (
	"crypto/tls"
	"log"
	"os"
	"time"

	"WalletScraper/internal/config"

	"github.com/streadway/amqp"
)

type Session struct {
	configs               *config.RabbitMQConfig
	tlsConfig             *tls.Config
	errorLog              *log.Logger
	connection            *amqp.Connection
	channel               *amqp.Channel
	done                  chan bool
	notifyConnectionClose chan *amqp.Error
	notifyChannelClose    chan *amqp.Error
	notifyConfirm         chan amqp.Confirmation
	isReady               bool
}

var globalSession *Session = nil

func (session *Session) IsReady() bool{
	return session.isReady
}

// New creates a new consumer state instance, and automatically
// attempts to connect to the server.
func New(
	configs *config.RabbitMQConfig,
	tlsConfig *tls.Config,
) (*Session, error) {

	globalSession := &Session{
		errorLog:  log.New(os.Stderr, "RabbitMQ ERROR - ", log.LUTC|log.Ldate|log.Lmsgprefix|log.Ltime),
		configs:   configs,
		tlsConfig: tlsConfig,
		done:      make(chan bool),
	}

	go globalSession.handleReconnect(configs.FormatRabbitMQURL())

	return globalSession, nil
}

// Get returns the current open session with RabbitMQ
func Get() *Session {
	if globalSession == nil {
		panic("RabbitMQ session must be initialized before used!")
	}
	return globalSession
}

// handleReconnect will wait for a connection error on
// notifyConnClose, and then continuously attempt to reconnect.
func (session *Session) handleReconnect(addr string) {
	for {
		if session.connection == nil || session.connection.IsClosed() {
			session.isReady = false
			log.Println("Attempting to connect")

			conn, err := session.connect(addr)
			if err != nil {
				session.errorLog.Printf("Failed to connect: Reason: %v. Retrying... ", err)

				select {
				case <-session.done:
					return
				case <-time.After(session.configs.ReconnectDelay):
				}
				continue
			}

			if done := session.handleReInit(conn); done {
				break
			}
		}
	}
}

// handleReconnect will wait for a channel error
// and then continuously attempt to re-initialize both channels
func (session *Session) handleReInit(conn *amqp.Connection) bool {
	for {
		session.isReady = false

		err := session.channelInit(conn)

		if err != nil {
			session.errorLog.Println("Failed to initialize channel. Retrying...")

			select {
			case <-session.done:
				return true
			case <-time.After(session.configs.ReconnectDelay):
			}
			continue
		}

		select {
		case <-session.done:
			return true
		case err := <-session.notifyConnectionClose:
			log.Printf("Connection closed (%s). Reconnecting...", err)
			return false
		case err := <-session.notifyChannelClose:
			log.Printf("Channel closed (%s). Re-running channelInit...", err)
		}
	}
}

// channelInit will initialize channel & declare queue
func (session *Session) channelInit(conn *amqp.Connection) error {
	ch, err := conn.Channel()

	if err != nil {
		return err
	}

	err = ch.Confirm(false)

	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		session.configs.InputQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		session.configs.OutputQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	session.channel = ch
	session.notifyChannelClose = make(chan *amqp.Error)
	session.notifyConfirm = make(chan amqp.Confirmation, 1)
	session.channel.NotifyClose(session.notifyChannelClose)
	session.channel.NotifyPublish(session.notifyConfirm)

	session.isReady = true
	log.Println("Setup!")

	return nil
}

// connect will create a new AMQP connection
func (session *Session) connect(rabbitMQURL string) (*amqp.Connection, error) {
	conn, err := amqp.DialTLS(rabbitMQURL, session.tlsConfig)
	if err != nil {
		return nil, err
	}

	session.connection = conn
	session.notifyConnectionClose = make(chan *amqp.Error)
	session.connection.NotifyClose(session.notifyConnectionClose)
	session.isReady = true

	log.Println("Connected on RabbitMQ!")
	return conn, nil
}

// Close will cleanly shutdown the channel and connection.
func (session *Session) Close() error {
	if !session.isReady {
		return &AlreadyClosedError{}
	}

	err := session.connection.Close()
	if err != nil {
		return err
	}
	close(session.done)
	session.isReady = false
	return nil
}

// Push will push data onto the queue, and wait for a confirm.
// If no confirms are received until within the resendTimeout,
// it continuously re-sends messages until a confirm is received.
// This will block until the server sends a confirm. Errors are
// only returned if the push action itself fails
func (session *Session) Push(data []byte) error {
	if !session.isReady {
		return &NotConnectedError{}
	}
	for {
		err := session.channel.Publish(
			"",                          // Exchange
			session.configs.OutputQueue, // Routing key
			true,                        // Mandatory
			false,                       // Immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        data,
			},
		)
		if err != nil {
			session.errorLog.Println("Push failed. Retrying...")
			select {
			case <-session.done:
				return &ShutdownError{}
			case <-time.After(session.configs.ResendDelay):
			}
			continue
		}
		select {
		case confirm := <-session.notifyConfirm:
			if confirm.Ack {
				log.Println("Push confirmed!")
				return nil
			}
		case <-time.After(session.configs.ResendDelay):
		}
		session.errorLog.Println("Push didn't confirm. Retrying...")
	}
}

// Stream will continuously put queue items on the channel.
// It is required to call delivery.Ack when it has been
// successfully processed, or delivery.Nack when it fails.
// Ignoring this will cause data to build up on the server.
func (session *Session) Stream() (<-chan amqp.Delivery, error) {
	if !session.isReady {
		return nil, &NotConnectedError{}
	}
	return session.channel.Consume(
		session.configs.InputQueue,
		"",    // Consumer
		false, // Auto-Ack
		false, // Exclusive
		false, // No-local
		false, // No-Wait
		nil,   // Args
	)
}
