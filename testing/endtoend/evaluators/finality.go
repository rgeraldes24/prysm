package evaluators

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/rpc/eth/beacon"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	eth "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/testing/endtoend/params"
	"github.com/prysmaticlabs/prysm/v4/testing/endtoend/policies"
	"github.com/prysmaticlabs/prysm/v4/testing/endtoend/types"
	"github.com/prysmaticlabs/prysm/v4/time/slots"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// FinalizationOccurs is an evaluator to make sure finalization is performing as it should.
// Requires to be run after at least 4 epochs have passed.
var FinalizationOccurs = func(epoch primitives.Epoch) types.Evaluator {
	return types.Evaluator{
		Name:       "finalizes_at_epoch_%d",
		Policy:     policies.AfterNthEpoch(epoch),
		Evaluation: finalizationOccurs,
	}
}

// NewFinalizedCheckpointOccurs is an evaluator to make sure that the chain keeps reaching finality.
// Requires to be run after at least 4 epochs have passed.
var NewFinalizedCheckpointOccurs = func(epoch primitives.Epoch) types.Evaluator {
	return types.Evaluator{
		Name:       "new_finalized_checkpoint_at_epoch_%d",
		Policy:     policies.AfterNthEpoch(epoch),
		Evaluation: newFinalizedCheckpointOccurs,
	}
}

func doMiddlewareJSONGetRequest(template string, requestPath string, beaconNodeIdx int, dst interface{}, bnType ...string) error {
	var port int
	if len(bnType) > 0 {
		switch bnType[0] {
		case "lighthouse":
			port = params.TestParams.Ports.LighthouseBeaconNodeHTTPPort
		default:
			port = params.TestParams.Ports.PrysmBeaconNodeGatewayPort
		}
	} else {
		port = params.TestParams.Ports.PrysmBeaconNodeGatewayPort
	}

	basePath := fmt.Sprintf(template, port+beaconNodeIdx)
	httpResp, err := http.Get(
		basePath + requestPath,
	)
	if err != nil {
		return err
	}
	if httpResp.StatusCode != http.StatusOK {
		var body interface{}
		if err := json.NewDecoder(httpResp.Body).Decode(&body); err != nil {
			return err
		}
		return fmt.Errorf("request failed with response code: %d with response body %s", httpResp.StatusCode, body)
	}
	return json.NewDecoder(httpResp.Body).Decode(&dst)
}

func newFinalizedCheckpointOccurs(_ *types.EvaluationContext, conns ...*grpc.ClientConn) error {
	conn := conns[0]
	client := eth.NewBeaconChainClient(conn)
	chainHead, err := client.GetChainHead(context.Background(), &emptypb.Empty{})
	if err != nil {
		return errors.Wrap(err, "failed to get chain head")
	}

	beaconNodeIdx := 0
	genesisResp := &beacon.GetGenesisResponse{}
	err = doMiddlewareJSONGetRequest(
		v1MiddlewarePathTemplate,
		"/beacon/genesis",
		beaconNodeIdx,
		genesisResp,
	)
	if err != nil {
		return errors.Wrap(err, "error getting genesis data")
	}
	genesisTime, err := strconv.ParseInt(genesisResp.Data.GenesisTime, 10, 64)
	if err != nil {
		return errors.Wrap(err, "could not parse genesis time")
	}

	currentEpoch := slots.EpochsSinceGenesis(time.Unix(genesisTime, 0))
	finalizedEpoch := chainHead.FinalizedEpoch

	expectedFinalizedEpoch := currentEpoch - 2
	if expectedFinalizedEpoch != finalizedEpoch {
		return fmt.Errorf(
			"expected finalized epoch to be %d, received: %d",
			expectedFinalizedEpoch,
			finalizedEpoch,
		)
	}

	return nil
}

func finalizationOccurs(_ *types.EvaluationContext, conns ...*grpc.ClientConn) error {
	conn := conns[0]
	client := eth.NewBeaconChainClient(conn)
	chainHead, err := client.GetChainHead(context.Background(), &emptypb.Empty{})
	if err != nil {
		return errors.Wrap(err, "failed to get chain head")
	}
	currentEpoch := chainHead.HeadEpoch
	finalizedEpoch := chainHead.FinalizedEpoch

	expectedFinalizedEpoch := currentEpoch - 2
	if expectedFinalizedEpoch != finalizedEpoch {
		return fmt.Errorf(
			"expected finalized epoch to be %d, received: %d",
			expectedFinalizedEpoch,
			finalizedEpoch,
		)
	}
	previousJustifiedEpoch := chainHead.PreviousJustifiedEpoch
	currentJustifiedEpoch := chainHead.JustifiedEpoch
	if previousJustifiedEpoch+1 != currentJustifiedEpoch {
		return fmt.Errorf(
			"there should be no gaps between current and previous justified epochs, received current %d and previous %d",
			currentJustifiedEpoch,
			previousJustifiedEpoch,
		)
	}
	if currentJustifiedEpoch+1 != currentEpoch {
		return fmt.Errorf(
			"there should be no gaps between current epoch and current justified epoch, received current %d and justified %d",
			currentEpoch,
			currentJustifiedEpoch,
		)
	}
	return nil
}
