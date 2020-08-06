# Getting Started

This is a step-by-step guide on how to get started with your own raftified Cosmos validator. The main purpose of this guide is to point out the code changes needed to run Raftify and where to make them.

## Step 0: Preparation

1. Pick a Cosmos SDK version of your choice.
2. Check out the [`go.mod` file in the Cosmos SDK](https://github.com/cosmos/cosmos-sdk/blob/master/go.mod) to find out which Tendermint version your chosen Cosmos SDK version uses.
    For example:

    ```text
    github.com/tendermint/tendermint v0.33.7
    ```

## Step 1: Tendermint

Raftify needs to hook into Tendermint to directly intercept the signing process such that inside of the raftify cluster only one single node is allowed to talk to the rest of the blockchain network at all times. This ultimately serves as double-signing prevention.

1. Go to [Tendermint GitHub](https://github.com/tendermint/tendermint).
2. Switch to your desired branch/tag and clone it.
3. Open the cloned repository in a code editor
4. Open the `go.mod` file in the repository's root directory and add Raftify as a dependency:

    ```go
    require (
        github.com/BlockscapeLab/raftify v0.2.0
    )
    ```

5. Navigate to `privval/file.go` and add the following code directly under the imports:

    ```go
    var rpv *raftify.Node

    func Shutdown() {
        rpv.Shutdown()
    }
    ```

6. Modify the `LoadOrGenFilePV` function as follows:

    ```go
    func LoadOrGenFilePV(keyFilePath, stateFilePath string) *FilePV {
        // Add this...
        homeDir := filepath.Dir(keyFilePath)
        logger := log.New(os.Stderr, "", 0)

        var err error
        if rpv, err = raftify.InitNode(logger, homeDir); err != nil {
            logger.Printf("[ERR] raftify: %v\n", err.Error())
        }
        // ...until here

        var pv *FilePV
        if tmos.FileExists(keyFilePath) {
            pv = LoadFilePV(keyFilePath, stateFilePath)
        } else {
            pv = GenFilePV(keyFilePath, stateFilePath)
            pv.Save()
        }
        return pv
    }
    ```

7. Modify the `SignVote` function as follows:

    ```go
    func (pv *FilePV) SignVote(chainID string, vote *tmproto.Vote) error {
        // Add this...
        if rpv.GetState() != raftify.Leader {
            return fmt.Errorf("%v is not a leader", rpv.GetID())
        }
        // ...until here

        if err := pv.signVote(chainID, vote); err != nil {
            return fmt.Errorf("error signing vote: %v", err)
        }
        return nil
    }
    ```

8. Modify the `SignProposal` function as follows:

    ```go
    func (pv *FilePV) SignProposal(chainID string, proposal *tmproto.Proposal) error {
        // Add this...
        if rpv.GetState() != raftify.Leader {
            return fmt.Errorf("%v is not a leader", rpv.GetID())
        }
        // ...until here

        if err := pv.signProposal(chainID, proposal); err != nil {
            return fmt.Errorf("error signing proposal: %v", err)
        }
        return nil
    }
    ```

9. Save all changes and proceed with the next step.

## Step 2: Cosmos SDK

The Cosmos SDK only needs to make use of the modified Tendermint privval implementation we created in step 1.

1. Go to [Cosmos SDK Github](https://github.com/cosmos/cosmos-sdk)
2. Switch to your desired branch/tag and clone it
3. Open the cloned repository in a code editor and navigate to `server/start.go`
4. Scroll down to the anonymous function `TrapSignal` and modify the code as follows:

    ```go
    if tmNode.IsRunning() {
        _ = tmNode.Stop()
        pvm.Shutdown() // Add this
    }
    ```

5. In the repository's root directory, open the `go.mod` file and replace Tendermint's privval implementation with your own local one. To do this, add...

    ```go
    replace "github.com/tendermint/tendermint/privval" => "/<your-local-path-to>/tendermint/privval"
    ```

    to the very bottom of the file.

6. Build the Cosmos SDK.

## Step 3: Raftify

Finally, go to `~/.gaiad/config/` and create a `raftify.json` config file. Use the [README](https://github.com/BlockscapeLab/raftify/blob/master/README.md) for the configuration reference.
