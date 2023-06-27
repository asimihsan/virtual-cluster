package main

import (
	"fmt"
	"github.com/asimihsan/virtual-cluster/internal/utils"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/Shopify/sarama"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "http-service-with-kafka",
		Usage: "HTTP service with Kafka",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "http-port",
				Usage:   "HTTP server port",
				EnvVars: []string{"PORT"},
				Value:   "1323",
			},
			&cli.StringFlag{
				Name:    "kafka-broker",
				Usage:   "Kafka broker address",
				EnvVars: []string{"KAFKA_BROKER"},
				Value:   "localhost:9095",
			},
		},
		Action: func(c *cli.Context) error {
			// Initialize Echo server
			e := echo.New()

			e.Use(middleware.Logger())
			e.Use(middleware.Recover())
			e.HideBanner = true
			e.HidePort = true

			// Initialize Kafka producer
			config := sarama.NewConfig()
			config.Producer.Return.Successes = true
			producer, err := sarama.NewSyncProducer([]string{c.String("kafka-broker")}, config)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to create Kafka producer")
			}
			defer func() {
				if err := producer.Close(); err != nil {
					log.Error().Err(err).Msg("Failed to close Kafka producer")
				}
			}()

			// Initialize atomic counter
			var counter int64 = 0

			// Define POST endpoint
			e.POST("/kafka", func(c echo.Context) error {
				log.Info().Msg("Received request at /kafka")

				// Increment counter
				value := atomic.AddInt64(&counter, 1)
				log.Info().Int64("value", value).Msg("Incremented counter")

				// Send message to Kafka
				message := fmt.Sprintf("Message %d", value)
				_, _, err := producer.SendMessage(&sarama.ProducerMessage{
					Topic: "my-topic",
					Value: sarama.StringEncoder(message),
				})
				if err != nil {
					log.Error().Err(err).Msg("Failed to send message to Kafka")
					return c.String(http.StatusInternalServerError, "Failed to send message to Kafka")
				}
				log.Info().Str("message", message).Msg("Sent message to Kafka")
				return c.String(http.StatusOK, message)
			})

			// Define other endpoints
			e.GET("/", func(c echo.Context) error {
				log.Info().Msg("Received request at /")
				return c.String(http.StatusOK, "Hello, World!")
			})

			e.GET("/ping", func(c echo.Context) error {
				log.Info().Msg("Received request at /health")
				return c.String(http.StatusOK, "healthy")
			})

			go func() {
				kw := utils.NewKafkaWaiter(c.String("kafka-broker"))
				if err := kw.Wait(); err != nil {
					log.Warn().Err(err).Msg("Failed to wait for Kafka")
				}
			}()

			// Start server
			port := c.String("http-port")
			if _, err := strconv.Atoi(port); err != nil {
				log.Fatal().Err(err).Msg("Invalid HTTP server port")
			}
			log.Info().Str("port", port).Msg("Starting HTTP server")
			return e.Start(":" + port)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("Failed to start app")
	}
}
