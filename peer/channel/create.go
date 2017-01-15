/*
Copyright IBM Corp. 2017 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

                 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package channel

import (
	"io/ioutil"
	"time"

	"github.com/hyperledger/fabric/common/configtx"
	configtxtest "github.com/hyperledger/fabric/common/configtx/test"
	"github.com/hyperledger/fabric/orderer/common/bootstrap/provisional"
	cb "github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/utils"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
)

func createCmd(cf *ChannelCmdFactory) *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a chain.",
		Long:  `Create a chain.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return create(cmd, args, cf)
		},
	}

	return createCmd
}

func sendCreateChainTransaction(cf *ChannelCmdFactory) error {
	//TODO this is a temporary hack until `orderer.template` is supplied from the CLI
	oTemplate := configtxtest.GetOrdererTemplate()

	mspcfg := configtx.NewSimpleTemplate(utils.EncodeMSPUnsigned(chainID))

	chCrtTemp := configtx.NewCompositeTemplate(oTemplate, mspcfg)

	chCrtEnv, err := configtx.MakeChainCreationTransaction(provisional.AcceptAllPolicyKey, chainID, chCrtTemp)

	if err != nil {
		return err
	}

	err = cf.BroadcastClient.Send(chCrtEnv)

	return err
}

func executeCreate(cf *ChannelCmdFactory) error {
	defer cf.BroadcastClient.Close()

	var err error

	if err = sendCreateChainTransaction(cf); err != nil {
		return err
	}

	time.Sleep(2 * time.Second)

	var block *cb.Block
	if block, err = cf.DeliverClient.getBlock(); err != nil {
		return err
	}

	b, err := proto.Marshal(block)
	if err != nil {
		return err
	}

	file := chainID + ".block"
	if err = ioutil.WriteFile(file, b, 0644); err != nil {
		return err
	}

	return nil
}

func create(cmd *cobra.Command, args []string, cf *ChannelCmdFactory) error {
	var err error
	if cf == nil {
		cf, err = InitCmdFactory(false)
		if err != nil {
			return err
		}
	}
	return executeCreate(cf)
}
