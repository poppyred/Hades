package event

import (
	"hades-ebpf/user/decoder"
	"hades-ebpf/user/helper"
	"strconv"

	manager "github.com/ehids/ebpfmanager"
)

var _ decoder.Event = (*SocketBind)(nil)

type SocketBind struct {
	decoder.BasicEvent `json:"-"`
	Family             int16  `json:"family"`
	LocalPort          string `json:"local_port"`
	LocalAddr          string `json:"local_addr"`
	Protocol           uint16 `json:"protocol"`
	Exe                string `json:"-"`
}

func (SocketBind) ID() uint32 {
	return 1024
}

func (SocketBind) Name() string {
	return "socket_bind"
}

func (s *SocketBind) GetExe() string {
	return s.Exe
}

func (s *SocketBind) DecodeEvent(decoder *decoder.EbpfDecoder) (err error) {
	var (
		index  uint8
		family int16
	)
	if err = decoder.DecodeUint8(&index); err != nil {
		return
	}
	if err = decoder.DecodeInt16(&family); err != nil {
		return
	}
	s.Family = family
	switch s.Family {
	case 2:
		var _port uint16
		if decoder.DecodeUint16BigEndian(&_port); err != nil {
			return
		}
		s.LocalPort = strconv.FormatUint(uint64(_port), 10)
		var _addr uint32
		if decoder.DecodeUint32BigEndian(&_addr); err != nil {
			return
		}
		s.LocalAddr = helper.PrintUint32IP(_addr)
		decoder.ReadByteSliceFromBuff(8)
	case 10:
		var _port uint16
		if decoder.DecodeUint16BigEndian(&_port); err != nil {
			return
		}
		s.LocalPort = strconv.FormatUint(uint64(_port), 10)
		var _flowinfo uint32
		if decoder.DecodeUint32BigEndian(&_flowinfo); err != nil {
			return
		}
		var _addr []byte
		_addr, err = decoder.ReadByteSliceFromBuff(16)
		if err != nil {
			return
		}
		s.LocalAddr = helper.Print16BytesSliceIP(_addr)
		// reuse
		decoder.DecodeUint32BigEndian(&_flowinfo)
	}
	if s.Exe, err = decoder.DecodeString(); err != nil {
		return
	}
	if err = decoder.DecodeUint8(&index); err != nil {
		return
	}
	err = decoder.DecodeUint16(&s.Protocol)
	return
}

func (SocketBind) GetProbes() []*manager.Probe {
	return []*manager.Probe{
		{
			UID:              "KprobeSecuritySocketBind",
			Section:          "kprobe/security_socket_bind",
			EbpfFuncName:     "kprobe_security_socket_bind",
			AttachToFuncName: "security_socket_bind",
		},
	}
}

func init() {
	decoder.RegistEvent(&SocketBind{})
}
