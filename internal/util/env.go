// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"os"
)

const (
	builderVariablePrefix           = "FABRIC_K8S_BUILDER_"
	ChaincodeNamespaceVariable      = builderVariablePrefix + "NAMESPACE"
	ChaincodeNodeRoleVariable       = builderVariablePrefix + "NODE_ROLE"
	ObjectNamePrefixVariable        = builderVariablePrefix + "OBJECT_NAME_PREFIX"
	ChaincodeServiceAccountVariable = builderVariablePrefix + "SERVICE_ACCOUNT"
	ChaincodeStartTimeoutVariable   = builderVariablePrefix + "START_TIMEOUT"
	ChaincodeHostAliasesVariable    = builderVariablePrefix + "HOST_ALIASES"
	ChaincodeAnnotationsVariable    = builderVariablePrefix + "ANNOTATIONS"
	ChaincodeEnvVarsVariable        = builderVariablePrefix + "CHAINCODE_ENV_VARS"
	ChaincodeImagePullSecretsVariable  = builderVariablePrefix + "IMAGE_PULL_SECRETS"
	ChaincodeLabelsVariable            = builderVariablePrefix + "LABELS"
	ChaincodeCPURequestVariable        = builderVariablePrefix + "CHAINCODE_CPU_REQUEST"
	ChaincodeMemoryRequestVariable     = builderVariablePrefix + "CHAINCODE_MEMORY_REQUEST"
	ChaincodeCPULimitVariable          = builderVariablePrefix + "CHAINCODE_CPU_LIMIT"
	ChaincodeMemoryLimitVariable       = builderVariablePrefix + "CHAINCODE_MEMORY_LIMIT"
	DebugVariable                      = builderVariablePrefix + "DEBUG"
	KubeconfigPathVariable          = "KUBECONFIG_PATH"
	PeerIDVariable                  = "CORE_PEER_ID"
)

func GetOptionalEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}

func GetRequiredEnv(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	}

	return "", fmt.Errorf("environment variable not set: %s", key)
}
