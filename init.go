package zap_human

import (
	"go.uber.org/zap"
	"log"
)

func init() {
	if err := zap.RegisterEncoder("human", NewHumanEncoder); err != nil {
		log.Printf("WARNING: Failed registering 'human' zap encoder: %s", err)
	}
}
