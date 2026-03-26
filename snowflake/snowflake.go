package snowflake

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/sony/sonyflake"
	"github.com/spf13/viper"
)

var (
	sf *sonyflake.Sonyflake
)

func init() {
	st := sonyflake.Settings{
		MachineID: getMachineID,
	}
	sf = sonyflake.NewSonyflake(st)
}

// getMachineIDFromIP 基于当前Pod IP地址生成MachineID
// 遍历所有网络接口，找到第一个非loopback的IP地址
// 对于IPv4使用第3、4字节，对于IPv6使用最后2字节
func getMachineIDFromIP() (uint16, error) {
	// 获取本机所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return 0, err
	}

	for _, iface := range interfaces {
		// 跳过loopback和down的接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 跳过loopback地址
			if ip == nil || ip.IsLoopback() {
				continue
			}

			// 使用第一个找到的非loopback IP
			if ip.To4() != nil {
				// IPv4: 使用第3、4字节
				ipBytes := ip.To4()
				return binary.BigEndian.Uint16(ipBytes[2:4]), nil
			} else {
				// IPv6: 使用最后2字节
				ipBytes := ip.To16()
				return binary.BigEndian.Uint16(ipBytes[14:16]), nil
			}
		}
	}

	return 0, fmt.Errorf("no suitable IP address found")
}

// getMachineIDFromHostname 从hostname生成MachineID
// 使用hostname的MD5哈希值的前2字节作为MachineID
func getMachineIDFromHostname() (uint16, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return 0, err
	}

	// 使用hostname的MD5哈希
	hash := md5.Sum([]byte(hostname))
	return binary.BigEndian.Uint16(hash[:2]), nil
}

// getMachineIDFromProcess 从进程信息生成MachineID（最后回退方案）
// 结合进程ID和当前时间纳秒，使用MD5哈希值的前2字节作为MachineID
func getMachineIDFromProcess() (uint16, error) {
	// 结合进程ID和当前时间纳秒
	pid := os.Getpid()
	nano := time.Now().UnixNano()

	// 简单组合生成一个相对唯一的ID
	combined := fmt.Sprintf("%d-%d", pid, nano)
	hash := md5.Sum([]byte(combined))

	return binary.BigEndian.Uint16(hash[:2]), nil
}

// getMachineID 获取机器ID
// 按优先级顺序尝试以下方案：
// 1. 从配置文件读取
// 2. 从IP地址生成
// 3. 从hostname生成
// 4. 从进程信息生成（最后回退）
func getMachineID() (uint16, error) {
	// 优先使用配置文件中的设置（向后兼容）
	if viper.IsSet("snowflake.machineId") {
		machineID := viper.GetUint16("snowflake.machineId")
		return machineID, nil
	}

	// 方案1: 尝试从IP地址生成
	if machineID, err := getMachineIDFromIP(); err == nil {
		return machineID, nil
	}

	// 方案2: 从hostname生成（Pod名称）
	if machineID, err := getMachineIDFromHostname(); err == nil {
		return machineID, nil
	}

	// 方案3: 从进程信息生成（最后的回退方案）
	return getMachineIDFromProcess()
}

// GenId 生成唯一ID
// Deprecated: 此方法已废弃，请使用 GenerateId() 替代
func GenId() int64 {
	id, err := sf.NextID()
	if err != nil {
		return 0
	}
	return int64(id)
}

// GenerateId 生成唯一ID
// 使用Sonyflake算法生成全局唯一的64位ID
// 返回生成的ID和可能的错误
func GenerateId() (uint64, error) {
	return sf.NextID()
}
