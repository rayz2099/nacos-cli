package output

import (
	"fmt"
	"strings"
)

func NormalizeConfigGetError(err error, dataID string, group string, namespace string) error {
	if err == nil {
		return nil
	}

	message := err.Error()
	reason, ok := knownConfigFallbackReason(message)
	if ok {
		return fmt.Errorf("get config failed, dataId=%s, group=%s, namespace=%s, reason=%s", dataID, group, namespace, reason)
	}

	return err
}

func knownConfigFallbackReason(message string) (string, bool) {
	if strings.Contains(message, "read config from both server and cache fail") {
		return "server unavailable and local cache missing", true
	}
	if strings.Contains(message, "read encryptedDataKey from server and cache fail") {
		return "server unavailable and local encrypted key missing", true
	}
	if strings.Contains(message, "get config from remote nacos server fail") {
		return "server unavailable and local snapshot disabled", true
	}
	return "", false
}
