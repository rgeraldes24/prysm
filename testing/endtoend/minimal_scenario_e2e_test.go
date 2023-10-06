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

func TestEndToEnd_ScenarioRun_AllNodesGoOffline(t *testing.T) {
	runner := e2eMinimal(t, version.Phase0, types.WithEpochs(10))
	params.TestParams.BeaconNodeCount = 4

	evals := scenarioEvals()
	evals = append(evals, ev.NewFinalizedCheckpointOccurs(3))
	runner.config.BeaconFlags = append(runner.config.BeaconFlags, "--min-sync-peers=1", "--enable-crash-recovery")
	runner.config.Evaluators = evals
	runner.config.EvalInterceptor = runner.allNodesOffline

	runner.scenarioRunner()
}

func TestEndToEnd_ScenarioRun_SingleNode_NodeGoesOffline(t *testing.T) {
	runner := e2eMinimal(t, version.Phase0, types.WithEpochs(10))
	params.TestParams.BeaconNodeCount = 1

	evals := scenarioEvals()
	evals = append(evals, ev.NewFinalizedCheckpointOccurs(3))
	runner.config.BeaconFlags = append(runner.config.BeaconFlags, "--min-sync-peers=0", "--startup-unfinalized")
	runner.config.Evaluators = evals
	runner.config.EvalInterceptor = runner.allNodesOffline

	runner.scenarioRunner()
}
