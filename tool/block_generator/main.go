package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dappley/go-dappley/config"
	configpb "github.com/dappley/go-dappley/config/pb"
	"github.com/dappley/go-dappley/consensus"
	"github.com/dappley/go-dappley/storage"
	tool "github.com/dappley/go-dappley/tool/block_generator/src"
)

const (
	genesisAddr           = "121yKAXeG4cw6uaGCBYjWk9yTWmMkhcoDD"
	genesisFilePath       = "conf/genesis.conf"
	defaultPassword       = "password"
	defaultTimeBetweenBlk = 3
)

func main() {
	// var filePath string
	numberBuffer := flag.Int("number", 1, "an int")

	flag.Parse()
	genesisConf := &configpb.DynastyConfig{}
	config.LoadConfig(genesisFilePath, genesisConf)
	fmt.Println("load genesis file")
	number := *numberBuffer
	files := make([]tool.FileInfo, number)
	maxProducers := (int)(genesisConf.GetMaxProducers())
	dynasty := consensus.NewDynastyWithConfigProducers(genesisConf.GetProducers(), maxProducers)
	keys := tool.LoadPrivateKey()
	reader := bufio.NewReader(os.Stdin)
	for i := 0; i < number; i++ {
		//enter filename
		fmt.Printf("Enter file name for blockchain%d: \n", i+1)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		text = "db/" + text
		db := storage.OpenDatabase(text)
		defer db.Close()
		files[i].Db = db
		//enter blockchain height
		fmt.Printf("Enter max height for blockchain%d: \n", i+1)
		height, _ := reader.ReadString('\n')
		height = strings.TrimSuffix(height, "\n")
		iheight, _ := strconv.Atoi(height)
		files[i].Height = iheight
		//enter height of blockchain have different with other block (0 means no different)
		fmt.Printf("Enter a different starting height for blockchain%d(0 for no different): \n", i+1)
		different, _ := reader.ReadString('\n')
		different = strings.TrimSuffix(different, "\n")
		idifferent, _ := strconv.Atoi(different)
		if iheight <= idifferent || idifferent < 1 {
			files[i].DifferentFrom = iheight
		} else {

			files[i].DifferentFrom = idifferent
		}

	}

	fmt.Printf("Enter number of normal transactions per block: \n")
	numOfTx, _ := reader.ReadString('\n')
	numOfTx = strings.TrimSuffix(numOfTx, "\n")
	numTx, _ := strconv.Atoi(numOfTx)

	fmt.Printf("Enter number of smart contract transactions per block: \n")
	numOfScTx, _ := reader.ReadString('\n')
	numOfScTx = strings.TrimSuffix(numOfScTx, "\n")
	numScTx, _ := strconv.Atoi(numOfScTx)

	config := tool.GeneralConfigs{numTx, numScTx}
	tool.GenerateNewBlockChain(files, dynasty, keys, config)
}
