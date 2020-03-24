// idgenerator
package tools

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

var (
	version               uint64 // 0~15
	shardId               uint64 // 0~4095
	sequence              uint64 // 0~255
	lastSequenceStartTime uint64
	mutex                 sync.Mutex
)

const (
	versionMask   = 0xF // 4bit
	versionOffset = 60
	timeMask      = 0x7FFFFFFFFF // 39 bit
	timeOffset    = 21
	shardIdMask   = 0x1FFF // 13 bit 支持4096个设备
	shardIdOffset = 8
	sequenceMask  = 0xFF          // 8 bit 每毫秒最多产生512个id
	epoch         = 1388506077506 // 时间起点 time够用17年，愿与公司同在！ 起始时间2014-01-01 00:07:57
)

func init() {
	version = 0
	shardId = 0
	sequence = 0
	lastSequenceStartTime = 0
}

// 设置分区id
func SetShardId(id int) error {
	if shardId > shardIdMask {
		return errors.New("shardId out of range" + strconv.Itoa(id))
	}
	shardId = uint64(id) << shardIdOffset
	return nil
}

// 设置id版本
func SetVersion(v int) error {
	if v > versionMask {
		return errors.New("version out of range" + strconv.Itoa(v))
	}
	version = uint64(v) << versionOffset
	return nil
}

// 获取id |version:4bit|timestamp:39bit|shardId:13bit|squence:8bit
func GenId() uint64 {
	mutex.Lock()
	nowMilli := genNowMillisecond()
	if nowMilli != lastSequenceStartTime {
		sequence = 0
		lastSequenceStartTime = nowMilli
	}

	if sequence == (sequenceMask + 1) {
		for {
			if nowMilli == lastSequenceStartTime {
				time.Sleep(2e5)
				//fmt.Println("sleep 6e5")
				nowMilli = genNowMillisecond()
				continue
			}
			lastSequenceStartTime = nowMilli
			sequence = 0
			break
		}
	}
	id := version | nowMilli | shardId | sequence
	sequence++
	mutex.Unlock()
	return id
}

func GenIdInt64() int64 {
	return int64(GenId())
}

// 获取id的版本信息
func GetVersion(id uint64) int {
	return int((id & (versionMask << versionOffset)) >> versionOffset)
}

// 获取id生成时间Time
func GetTime(id uint64) time.Time {
	t := int64((id&(timeMask<<timeOffset))>>timeOffset + epoch)
	return time.Unix(t/1000, (t%1000)*1000000)
}

// 获取id的UinxNano时间戳
func GetTimeUnixNano(id uint64) int64 {
	return int64(((id&(timeMask<<timeOffset))>>timeOffset + epoch) * 1000000)
}

// 获取id的UinxMill时间戳
func GetTimeUnixMill(id int64) int64 {
	return int64((id&(timeMask<<timeOffset))>>timeOffset + epoch)
}

// 获取id的分区id(组件唯一标识)
func GetShardId(id uint64) int {
	return int((id & (shardIdMask << shardIdOffset)) >> shardIdOffset)
}

// 获取id当前的序列号
func GetSequence(id uint64) uint64 {
	return id & sequenceMask
}

func genNowMillisecond() uint64 {
	return uint64(((time.Now().UnixNano()/1000000 - epoch) & timeMask) << timeOffset)
}
