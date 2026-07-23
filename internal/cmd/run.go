// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/hyperledger-labs/fabric-builder-k8s/internal/builder"
	"github.com/hyperledger-labs/fabric-builder-k8s/internal/log"
	"github.com/hyperledger-labs/fabric-builder-k8s/internal/util"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation"
)

//nolint:nonamedreturns // using the ok bool convention to indicate errors
func getPeerID(logger *log.CmdLogger) (peerID string, ok bool) {
	peerID, err := util.GetRequiredEnv(util.PeerIDVariable)
	if err != nil {
		logger.Printf("Expected %s environment variable\n", util.PeerIDVariable)

		return peerID, false
	}

	logger.Debugf("%s=%s", util.PeerIDVariable, peerID)

	return peerID, true
}

func getKubeconfigPath(logger *log.CmdLogger) string {
	kubeconfigPath := util.GetOptionalEnv(util.KubeconfigPathVariable, "")
	logger.Debugf("%s=%s", util.KubeconfigPathVariable, kubeconfigPath)

	return kubeconfigPath
}

func getKubeNamespace(logger *log.CmdLogger) string {
	kubeNamespace := util.GetOptionalEnv(util.ChaincodeNamespaceVariable, "")
	logger.Debugf("%s=%s", util.ChaincodeNamespaceVariable, kubeNamespace)

	if kubeNamespace == "" {
		var err error

		kubeNamespace, err = util.GetKubeNamespace()
		if err != nil {
			logger.Debugf("Error getting namespace: %+v\n", util.DefaultNamespace, err)
			kubeNamespace = util.DefaultNamespace
		}

		logger.Debugf("Using default namespace: %s\n", util.DefaultNamespace)
	}

	return kubeNamespace
}

//nolint:nonamedreturns // using the ok bool convention to indicate errors
func getKubeNodeRole(logger *log.CmdLogger) (kubeNodeRole string, ok bool) {
	kubeNodeRole = util.GetOptionalEnv(util.ChaincodeNodeRoleVariable, "")
	logger.Debugf("%s=%s", util.ChaincodeNodeRoleVariable, kubeNodeRole)

	// TODO: are valid taint values the same?!
	if msgs := validation.IsValidLabelValue(kubeNodeRole); len(msgs) > 0 {
		logger.Printf("The %s environment variable must be a valid Kubernetes label value: %s", util.ChaincodeNodeRoleVariable, msgs[0])

		return kubeNodeRole, false
	}

	return kubeNodeRole, true
}

func getKubeServiceAccount(logger *log.CmdLogger) string {
	kubeServiceAccount := util.GetOptionalEnv(util.ChaincodeServiceAccountVariable, util.DefaultServiceAccountName)
	logger.Debugf("%s=%s", util.ChaincodeServiceAccountVariable, kubeServiceAccount)

	return kubeServiceAccount
}

//nolint:nonamedreturns // using the ok bool convention to indicate errors
func getKubeNamePrefix(logger *log.CmdLogger) (kubeNamePrefix string, ok bool) {
	const maximumKubeNamePrefixLength = 30

	kubeNamePrefix = util.GetOptionalEnv(util.ObjectNamePrefixVariable, util.DefaultObjectNamePrefix)
	logger.Debugf("%s=%s", util.ObjectNamePrefixVariable, kubeNamePrefix)

	if len(kubeNamePrefix) > maximumKubeNamePrefixLength {
		logger.Printf("The %s environment variable must be a maximum of 30 characters", util.ObjectNamePrefixVariable)

		return kubeNamePrefix, false
	}

	if msgs := apivalidation.NameIsDNS1035Label(kubeNamePrefix, true); len(msgs) > 0 {
		logger.Printf("The %s environment variable must be a valid DNS-1035 label: %s", util.ObjectNamePrefixVariable, msgs[0])

		return kubeNamePrefix, false
	}

	return kubeNamePrefix, true
}

//nolint:nonamedreturns // using the ok bool convention to indicate errors
func getKubeHostAliases(logger *log.CmdLogger) (hostAliases []apiv1.HostAlias, ok bool) {
	raw := util.GetOptionalEnv(util.ChaincodeHostAliasesVariable, "")
	logger.Debugf("%s=%s", util.ChaincodeHostAliasesVariable, raw)

	if raw == "" {
		return nil, true
	}

	if err := json.Unmarshal([]byte(raw), &hostAliases); err != nil {
		logger.Printf(
			`The %s environment variable must be a valid JSON array, e.g. [{"ip":"1.2.3.4","hostnames":["foo.com"]}]: %v`,
			util.ChaincodeHostAliasesVariable, err,
		)

		return nil, false
	}

	return hostAliases, true
}

//nolint:nonamedreturns // using the ok bool convention to indicate errors
func getKubeCustomAnnotations(logger *log.CmdLogger) (annotations map[string]string, ok bool) {
	raw := util.GetOptionalEnv(util.ChaincodeAnnotationsVariable, "")
	logger.Debugf("%s=%s", util.ChaincodeAnnotationsVariable, raw)

	if raw == "" {
		return nil, true
	}

	if err := json.Unmarshal([]byte(raw), &annotations); err != nil {
		logger.Printf(
			`The %s environment variable must be a valid JSON object, e.g. {"key":"value"}: %v`,
			util.ChaincodeAnnotationsVariable, err,
		)

		return nil, false
	}

	return annotations, true
}

//nolint:nonamedreturns // using the ok bool convention to indicate errors
func getKubeCustomLabels(logger *log.CmdLogger) (labels map[string]string, ok bool) {
	raw := util.GetOptionalEnv(util.ChaincodeLabelsVariable, "")
	logger.Debugf("%s=%s", util.ChaincodeLabelsVariable, raw)

	if raw == "" {
		return nil, true
	}

	if err := json.Unmarshal([]byte(raw), &labels); err != nil {
		logger.Printf(
			`The %s environment variable must be a valid JSON object, e.g. {"key":"value"}: %v`,
			util.ChaincodeLabelsVariable, err,
		)

		return nil, false
	}

	return labels, true
}

//nolint:nonamedreturns // using the ok bool convention to indicate errors
func getChaincodeEnvVars(logger *log.CmdLogger) (envVars []apiv1.EnvVar, ok bool) {
	raw := util.GetOptionalEnv(util.ChaincodeEnvVarsVariable, "")
	logger.Debugf("%s=%s", util.ChaincodeEnvVarsVariable, raw)

	if raw == "" {
		return nil, true
	}

	var parsed map[string]string
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		logger.Printf(
			`The %s environment variable must be a valid JSON object, e.g. {"KEY":"VALUE"}: %v`,
			util.ChaincodeEnvVarsVariable, err,
		)

		return nil, false
	}

	for k, v := range parsed {
		envVars = append(envVars, apiv1.EnvVar{Name: k, Value: v})
	}

	return envVars, true
}

//nolint:nonamedreturns // using the ok bool convention to indicate errors
func getKubeImagePullSecrets(logger *log.CmdLogger) (imagePullSecrets []apiv1.LocalObjectReference, ok bool) {
	raw := util.GetOptionalEnv(util.ChaincodeImagePullSecretsVariable, "")
	logger.Debugf("%s=%s", util.ChaincodeImagePullSecretsVariable, raw)

	if raw == "" {
		return nil, true
	}

	var secretNames []string
	if err := json.Unmarshal([]byte(raw), &secretNames); err != nil {
		logger.Printf(
			`The %s environment variable must be a valid JSON array of secret names, e.g. ["mysecret","anothersecret"]: %v`,
			util.ChaincodeImagePullSecretsVariable, err,
		)

		return nil, false
	}

	for _, name := range secretNames {
		imagePullSecrets = append(imagePullSecrets, apiv1.LocalObjectReference{Name: name})
	}

	return imagePullSecrets, true
}

//nolint:nonamedreturns // using the ok bool convention to indicate errors
func getChaincodeResources(logger *log.CmdLogger) (resources apiv1.ResourceRequirements, ok bool) {
	cpuReq := util.GetOptionalEnv(util.ChaincodeCPURequestVariable, "100m")
	memReq := util.GetOptionalEnv(util.ChaincodeMemoryRequestVariable, "128Mi")
	cpuLim := util.GetOptionalEnv(util.ChaincodeCPULimitVariable, "500m")
	memLim := util.GetOptionalEnv(util.ChaincodeMemoryLimitVariable, "2Gi")

	logger.Debugf("%s=%s", util.ChaincodeCPURequestVariable, cpuReq)
	logger.Debugf("%s=%s", util.ChaincodeMemoryRequestVariable, memReq)
	logger.Debugf("%s=%s", util.ChaincodeCPULimitVariable, cpuLim)
	logger.Debugf("%s=%s", util.ChaincodeMemoryLimitVariable, memLim)

	cpuReqQ, err := resource.ParseQuantity(cpuReq)
	if err != nil {
		logger.Printf("The %s environment variable must be a valid Kubernetes quantity, e.g. 100m: %v", util.ChaincodeCPURequestVariable, err)
		return apiv1.ResourceRequirements{}, false
	}

	memReqQ, err := resource.ParseQuantity(memReq)
	if err != nil {
		logger.Printf("The %s environment variable must be a valid Kubernetes quantity, e.g. 128Mi: %v", util.ChaincodeMemoryRequestVariable, err)
		return apiv1.ResourceRequirements{}, false
	}

	cpuLimQ, err := resource.ParseQuantity(cpuLim)
	if err != nil {
		logger.Printf("The %s environment variable must be a valid Kubernetes quantity, e.g. 500m: %v", util.ChaincodeCPULimitVariable, err)
		return apiv1.ResourceRequirements{}, false
	}

	memLimQ, err := resource.ParseQuantity(memLim)
	if err != nil {
		logger.Printf("The %s environment variable must be a valid Kubernetes quantity, e.g. 2Gi: %v", util.ChaincodeMemoryLimitVariable, err)
		return apiv1.ResourceRequirements{}, false
	}

	return apiv1.ResourceRequirements{
		Requests: apiv1.ResourceList{
			apiv1.ResourceCPU:    cpuReqQ,
			apiv1.ResourceMemory: memReqQ,
		},
		Limits: apiv1.ResourceList{
			apiv1.ResourceCPU:    cpuLimQ,
			apiv1.ResourceMemory: memLimQ,
		},
	}, true
}

//nolint:nonamedreturns // using the ok bool convention to indicate errors
func getChaincodeStartTimeout(logger *log.CmdLogger) (chaincodeStartTimeoutDuration time.Duration, ok bool) {
	chaincodeStartTimeout := util.GetOptionalEnv(util.ChaincodeStartTimeoutVariable, util.DefaultStartTimeout)
	logger.Debugf("%s=%s", util.ChaincodeStartTimeoutVariable, chaincodeStartTimeout)

	chaincodeStartTimeoutDuration, err := time.ParseDuration(chaincodeStartTimeout)
	if err != nil {
		logger.Printf("The %s environment variable must be a valid Go duration string, e.g. 3m40s: %v", util.ChaincodeStartTimeoutVariable, err)

		return 0 * time.Minute, false
	}

	return chaincodeStartTimeoutDuration, true
}

func Run() {
	const (
		expectedArgsLength      = 3
		buildOutputDirectoryArg = 1
		runMetadataDirectoryArg = 2
	)

	debug := util.GetOptionalEnv(util.DebugVariable, "false")
	ctx := log.NewCmdContext(context.Background(), debug == "true")
	logger := log.New(ctx)

	if len(os.Args) != expectedArgsLength {
		logger.Println("Expected BUILD_OUTPUT_DIR and RUN_METADATA_DIR arguments")

		os.Exit(1)
	}

	buildOutputDirectory := os.Args[buildOutputDirectoryArg]
	runMetadataDirectory := os.Args[runMetadataDirectoryArg]

	logger.Debugf("Build output directory: %s", buildOutputDirectory)
	logger.Debugf("Run metadata directory: %s", runMetadataDirectory)

	//nolint:varnamelen // using the ok bool convention to indicate errors
	var ok bool

	peerID, ok := getPeerID(logger)
	if !ok {
		os.Exit(1)
	}

	kubeconfigPath := getKubeconfigPath(logger)
	kubeNamespace := getKubeNamespace(logger)

	kubeNodeRole, ok := getKubeNodeRole(logger)
	if !ok {
		os.Exit(1)
	}

	kubeServiceAccount := getKubeServiceAccount(logger)

	kubeNamePrefix, ok := getKubeNamePrefix(logger)
	if !ok {
		os.Exit(1)
	}

	chaincodeStartTimeout, ok := getChaincodeStartTimeout(logger)
	if !ok {
		os.Exit(1)
	}

	kubeHostAliases, ok := getKubeHostAliases(logger)
	if !ok {
		os.Exit(1)
	}

	kubeCustomAnnotations, ok := getKubeCustomAnnotations(logger)
	if !ok {
		os.Exit(1)
	}

	kubeCustomLabels, ok := getKubeCustomLabels(logger)
	if !ok {
		os.Exit(1)
	}

	chaincodeEnvVars, ok := getChaincodeEnvVars(logger)
	if !ok {
		os.Exit(1)
	}

	kubeImagePullSecrets, ok := getKubeImagePullSecrets(logger)
	if !ok {
		os.Exit(1)
	}

	chaincodeResources, ok := getChaincodeResources(logger)
	if !ok {
		os.Exit(1)
	}

	run := &builder.Run{
		BuildOutputDirectory:  buildOutputDirectory,
		RunMetadataDirectory:  runMetadataDirectory,
		PeerID:                peerID,
		KubeconfigPath:        kubeconfigPath,
		KubeNamespace:         kubeNamespace,
		KubeNodeRole:          kubeNodeRole,
		KubeServiceAccount:    kubeServiceAccount,
		KubeNamePrefix:        kubeNamePrefix,
		ChaincodeStartTimeout: chaincodeStartTimeout,
		KubeHostAliases:       kubeHostAliases,
		KubeCustomAnnotations: kubeCustomAnnotations,
		KubeCustomLabels:      kubeCustomLabels,
		ChaincodeEnvVars:      chaincodeEnvVars,
		KubeImagePullSecrets:  kubeImagePullSecrets,
		ChaincodeResources:    chaincodeResources,
	}

	if err := run.Run(ctx); err != nil {
		logger.Printf("Error running chaincode: %+v", err)

		os.Exit(1)
	}

	os.Exit(0)
}
