package kcp

import (
	kcp "github.com/xtaci/kcp-go"
	"github.com/yinqiwen/gsnova/common/logger"
	"github.com/yinqiwen/gsnova/common/mux"
	"github.com/yinqiwen/gsnova/remote"
	"github.com/yinqiwen/pmux"
)

func StartKCPProxyServer(addr string) error {
	config := &remote.ServerConf.KCP
	block, _ := kcp.NewNoneBlockCrypt(nil)
	lis, err := kcp.ListenWithOptions(addr, block, config.DataShard, config.ParityShard)
	if nil != err {
		logger.Error("[ERROR]Failed to listen KCP address:%s with reason:%v", addr, err)
		return err
	}

	if err := lis.SetDSCP(config.DSCP); err != nil {
		logger.Debug("SetDSCP:%v", err)
	}
	if err := lis.SetReadBuffer(config.SockBuf); err != nil {
		logger.Debug("SetReadBuffer:%v", err)
	}
	if err := lis.SetWriteBuffer(config.SockBuf); err != nil {
		logger.Debug("SetWriteBuffer:%v", err)
	}
	logger.Info("Listen on KCP address:%s", addr)
	servKCP(lis)
	return nil
}

func servKCP(lp *kcp.Listener) {
	for {
		conn, err := lp.AcceptKCP()
		if nil != err {
			continue
		}
		config := &remote.ServerConf.KCP
		conn.SetStreamMode(true)
		conn.SetWriteDelay(true)
		conn.SetNoDelay(config.NoDelay, config.Interval, config.Resend, config.NoCongestion)
		conn.SetMtu(config.MTU)
		conn.SetWindowSize(config.SndWnd, config.RcvWnd)
		conn.SetACKNoDelay(config.AckNodelay)
		session, err := pmux.Server(conn, remote.InitialPMuxConfig())
		if nil != err {
			logger.Error("[ERROR]Failed to create mux session for tcp server with reason:%v", err)
			continue
		}
		muxSession := &mux.ProxyMuxSession{Session: session}
		go remote.ServProxyMuxSession(muxSession)
	}
	//ws.WriteMessage(websocket.CloseMessage, []byte{})
}
