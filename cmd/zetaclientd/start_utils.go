package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/zeta-chain/node/zetaclient/config"
	"google.golang.org/grpc"
)

func waitForZetaCore(configData *config.Config, logger zerolog.Logger) {
	// wait until zetacore is up
	logger.Debug().Msg("Waiting for ZetaCore to open 9090 port...")
	for {
		_, err := grpc.Dial(
			fmt.Sprintf("%s:9090", configData.ZetaCoreURL),
			grpc.WithInsecure(),
		)
		if err != nil {
			logger.Warn().Err(err).Msg("grpc dial fail")
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
}

func validatePeer(seedPeer string) error {
	parsedPeer := strings.Split(seedPeer, "/")

	if len(parsedPeer) < 7 {
		return errors.New("seed peer missing IP or ID")
	}

	seedIP := parsedPeer[2]
	seedID := parsedPeer[6]

	if net.ParseIP(seedIP) == nil {
		return errors.New("invalid seed IP address")
	}

	if len(seedID) == 0 {
		return errors.New("seed id is empty")
	}

	return nil
}
