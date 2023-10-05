package endtoend

import (
	"testing"

	"github.com/prysmaticlabs/prysm/v4/runtime/version"
	ev "github.com/prysmaticlabs/prysm/v4/testing/endtoend/evaluators"
	"github.com/prysmaticlabs/prysm/v4/testing/endtoend/params"
	"github.com/prysmaticlabs/prysm/v4/testing/endtoend/types"
)

func TestEndToEnd_MultiScenarioRun(t *testing.T) {
	runner := e2eMinimal(t, version.Phase0, types.WithEpochs(24))

	runner.config.Evaluators = scenarioEvals()
	runner.config.EvalInterceptor = runner.multiScenario
	runner.scenarioRunner()
}

func TestEndToEnd_MinimalConfig_Web3Signer(t *testing.T) {
	e2eMinimal(t, version.Phase0, types.WithRemoteSigner()).run()
}

func TestEndToEnd_MinimalConfig_ValidatorRESTApi(t *testing.T) {
	e2eMinimal(t, version.Phase0, types.WithCheckpointSync(), types.WithValidatorRESTApi()).run()
}

func TestEndToEnd_ScenarioRun_EEOffline(t *testing.T) {
	t.Skip("TODO(#10242) Prysm is current unable to handle an offline e2e")
	runner := e2eMinimal(t, version.Phase0)

	runner.config.Evaluators = scenarioEvals()
	runner.config.EvalInterceptor = runner.eeOffline
	runner.scenarioRunner()
}

func TestEndToEnd_ScenarioRun_AllNodesOffline(t *testing.T) {
	runner := e2eMinimal(t, version.Phase0, types.WithEpochs(15))
	runner.config.Evaluators = scenarioEvals()
	runner.config.EvalInterceptor = runner.allNodesOffline
	runner.scenarioRunner()
}

func TestEndToEnd_ScenarioRun_AllNodesOffline_SingleNode(t *testing.T) {
	runner := e2eMinimal(t, version.Phase0, types.WithEpochs(10))

	params.TestParams.BeaconNodeCount = 1

	runner.config.BeaconFlags = append(runner.config.BeaconFlags, "--min-sync-peers=0")
	runner.config.Evaluators = []types.Evaluator{
		ev.PeersConnect,
		ev.HealthzCheck,
		ev.MetricsCheck,
		ev.ValidatorsParticipatingAtEpoch(2),
		ev.FinalizationOccurs(3),
		ev.VerifyBlockGraffiti,
		ev.ProposeVoluntaryExit,
		ev.ValidatorsHaveExited,
		ev.ColdStateCheckpoint,
		ev.AltairForkTransition,
		ev.BellatrixForkTransition,
		ev.CapellaForkTransition,
		// ev.DenebForkTransition, // TODO(12750): Enable this when geth main branch's engine API support.
		ev.APIMiddlewareVerifyIntegrity,
		ev.APIGatewayV1Alpha1VerifyIntegrity,
		ev.FinishedSyncing,
		ev.AllNodesHaveSameHead,
		ev.ValidatorSyncParticipation,
		ev.NewFinalizedCheckpointOccurs(3),
	}
	runner.config.EvalInterceptor = runner.allNodesOffline

	runner.scenarioRunner()
}
