package balance

import (
	"errors"
	"fmt"
	"hash/crc32"
	"math/rand"
	"sync"
)

type BalanceType int64

func (b BalanceType) Name() string {
	switch b {
	case 0:
		return "least job prefered"
	case 1:
		return "random select"
	case 2:
		return "round robin"
	case 3:
		return "consistant hash"
	case 4:
		return "shuffle"
	default:
		return "least job prefered"
	}
}

const (
	DEFAULTBALANCE        BalanceType = BalanceType(0)
	RADOMBALANCE          BalanceType = BalanceType(1)
	ROUNDROBINBALANCE     BalanceType = BalanceType(2)
	CONSISTANTHASHBALANCE BalanceType = BalanceType(3)
	SHUFFLEBALANCE        BalanceType = BalanceType(4)
)

type Balancer interface {
	DoBalance([]string) (string, error)
}

type RandomBalance struct{}

var DefaultRandomBalancer = RandomBalance{}

func (b *RandomBalance) DoBalance(nodeList []string) (string, error) {
	lens := len(nodeList)
	if lens == 0 {
		return "", errors.New("No node found!")
	}
	index := rand.Intn(lens)
	return nodeList[index], nil
}

type RoundRobinBalance struct {
	curIndex int
	sync.Mutex
}

var DefaultRoundRobinBalancer = RoundRobinBalance{
	curIndex: 0,
}

func (b *RoundRobinBalance) DoBalance(nodeList []string) (string, error) {
	lens := len(nodeList)
	if lens == 0 {
		return "", errors.New("No node found!")
	}
	if b.curIndex >= lens {
		b.curIndex = 0
	}
	nodeUUID := nodeList[b.curIndex]
	b.Mutex.Lock()
	b.curIndex = (b.curIndex + 1) % lens
	b.Mutex.Unlock()
	return nodeUUID, nil
}

type ConsistantHashBalance struct{}

var DefaultConsistantHashBalancer = ConsistantHashBalance{}

func (b *ConsistantHashBalance) DoBalance(nodeList []string) (string, error) {
	lens := len(nodeList)
	if lens == 0 {
		return "", errors.New("No node found!")
	}
	var defKey string = fmt.Sprintf("%d", rand.Int())
	crcTable := crc32.MakeTable(crc32.IEEE)
	hashVal := crc32.Checksum([]byte(defKey), crcTable)
	index := int(hashVal) % lens
	return nodeList[index], nil
}

type ShuffleBalance struct{}

var DefaultSuffleBalancer = ShuffleBalance{}

func (b *ShuffleBalance) DoBalance(nodeList []string) (string, error) {
	lens := len(nodeList)
	if lens == 0 {
		return "", errors.New("No node found!")
	}

	//shuffle
	for i := 0; i < lens/2; i++ {
		a := rand.Intn(lens)
		b := rand.Intn(lens)
		nodeList[a], nodeList[b] = nodeList[b], nodeList[a]
	}
	return nodeList[0], nil
}
