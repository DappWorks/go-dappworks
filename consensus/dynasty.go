// Copyright (C) 2018 go-dappley authors
//
// This file is part of the go-dappley library.
//
// the go-dappley library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-dappley library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-dappley library.  If not, see <http://www.gnu.org/licenses/>.
//

package consensus

import (
	"github.com/dappley/go-dappley/core"
	"bytes"
)

type Dynasty struct{
	producers      []string
	maxProducers   int
	timeBetweenBlk int
	dynastyTime    int
}

const (
	defaultMaxProducers   = 3
	defaultTimeBetweenBlk = 15
	defaultDynastyTime    = defaultMaxProducers * defaultTimeBetweenBlk
)

func NewDynasty() *Dynasty{
	return &Dynasty{
		producers:      []string{},
		maxProducers:   defaultMaxProducers,
		timeBetweenBlk: defaultTimeBetweenBlk,
		dynastyTime:    defaultDynastyTime,
	}
}

func NewDynastyWithProducers(producers []string) *Dynasty{
	validProducers := []string{}
	for _, producer := range producers {
		if IsProducerAddressValid(producer){
			validProducers = append(validProducers, producer)
		}
	}
	return &Dynasty{
		producers:      validProducers,
		maxProducers:   len(validProducers),
		timeBetweenBlk: defaultTimeBetweenBlk,
		dynastyTime:    len(validProducers)*defaultTimeBetweenBlk,
	}
}

func (dynasty *Dynasty) SetMaxProducers(maxProducers int){
	if maxProducers >=0 {
		dynasty.maxProducers = maxProducers
		dynasty.dynastyTime = maxProducers * dynasty.timeBetweenBlk
	}
	if maxProducers < len(dynasty.producers){
		dynasty.producers = dynasty.producers[:maxProducers]
	}
}

func (dynasty *Dynasty) SetTimeBetweenBlk(timeBetweenBlk int){
	if timeBetweenBlk > 0 {
		dynasty.timeBetweenBlk = timeBetweenBlk
		dynasty.dynastyTime = dynasty.maxProducers * timeBetweenBlk
	}
}

func (dynasty *Dynasty) AddProducer(producer string){
	if IsProducerAddressValid(producer) && len(dynasty.producers) < dynasty.maxProducers{
		dynasty.producers = append(dynasty.producers, producer)
	}
}

func (dynasty *Dynasty) AddMultipleProducers(producers []string){
	for _, producer := range producers {
		dynasty.AddProducer(producer)
	}
}

func (dynasty *Dynasty) IsMyTurn(producer string, now int64) bool{
	index := dynasty.GetProducerIndex(producer)
	return dynasty.isMyTurnByIndex(index, now)
}

func (dynasty *Dynasty) isMyTurnByIndex(producerIndex int, now int64) bool{
	if producerIndex < 0 {
		return false
	}
	dynastyTimeElapsed := int(now % int64(dynasty.dynastyTime))

	return dynastyTimeElapsed == producerIndex * dynasty.timeBetweenBlk
}

func (dynasty *Dynasty) ProducerAtATime(time int64) string{
	if time < 0{
		return ""
	}
	dynastyTimeElapsed := int(time % int64(dynasty.dynastyTime))
	index := dynastyTimeElapsed/dynasty.timeBetweenBlk
	return dynasty.producers[index]
}

//find the index of the producer. If not found, return -1
func (dynasty *Dynasty) GetProducerIndex(producer string) int{
	for i,m := range dynasty.producers {
		if producer == m {
			return i
		}
	}
	return -1
}

func (dynasty *Dynasty) ValidateProducer(block *core.Block) bool{

	if block == nil {
		return false
	}

	producer := dynasty.ProducerAtATime(block.GetTimestamp())
	producerHash := core.HashAddress([]byte(producer))

	cbtx := block.GetCoinbaseTransaction()
	if cbtx==nil {
		return false
	}

	if len(cbtx.Vout) == 0{
		return false
	}

	return bytes.Compare(producerHash, cbtx.Vout[0].PubKeyHash)==0
}

func IsProducerAddressValid(producer string) bool{
	addr := core.Address{producer}
	return addr.ValidateAddress()
}