package turn

import (
	"context"
	"net"
	"sync"

	"github.com/gravitl/netclient/ncutils"
	"github.com/gravitl/netclient/nmproxy/config"
	"github.com/gravitl/netclient/nmproxy/models"
	peerpkg "github.com/gravitl/netclient/nmproxy/peer"
	wireguard "github.com/gravitl/netclient/nmproxy/wg"
	"github.com/gravitl/netmaker/logger"
	nm_models "github.com/gravitl/netmaker/models"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var PeerSignalCh = make(chan nm_models.Signal, 100)

func WatchPeerSignals(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case signal := <-PeerSignalCh:
			// recieved new signal from peer, check if turn endpoint is different
			peerTurnEndpoint, err := net.ResolveUDPAddr("udp", signal.TurnRelayEndpoint)
			if err != nil {
				continue
			}
			t, ok := config.GetCfg().GetPeerTurnCfg(signal.Server, signal.FromHostPubKey)
			if ok {
				if signal.TurnRelayEndpoint != "" {
					// reset
					if conn, ok := config.GetCfg().GetPeer(signal.FromHostPubKey); ok && conn.Config.UsingTurn {
						if conn.Config.PeerEndpoint.String() != peerTurnEndpoint.String() ||
							t.PeerTurnAddr != signal.TurnRelayEndpoint {
							config.GetCfg().UpdatePeerTurnAddr(signal.Server, signal.FromHostPubKey, signal.TurnRelayEndpoint)
							conn.Config.PeerEndpoint = peerTurnEndpoint
							config.GetCfg().UpdatePeer(&conn)
							conn.ResetConn()
						}

					} else {
						// new connection
						config.GetCfg().SetPeerTurnCfg(signal.Server, signal.FromHostPubKey, models.TurnPeerCfg{
							Server:       signal.Server,
							PeerConf:     t.PeerConf,
							PeerTurnAddr: signal.TurnRelayEndpoint,
						})

						peer, err := wireguard.GetPeer(ncutils.GetInterfaceName(), signal.FromHostPubKey)
						if err == nil {
							peerpkg.AddNew(t.Server, wgtypes.PeerConfig{
								PublicKey:                   peer.PublicKey,
								PresharedKey:                &peer.PresharedKey,
								Endpoint:                    peer.Endpoint,
								PersistentKeepaliveInterval: &peer.PersistentKeepaliveInterval,
								AllowedIPs:                  peer.AllowedIPs,
							}, t.PeerConf, false, peerTurnEndpoint, true)
						}

					}

				}
				// signal back to peer
				// signal peer with the host relay addr for the peer
				if signal.Reply {
					continue
				}
				if hostTurnCfg, ok := config.GetCfg().GetTurnCfg(signal.Server); ok && hostTurnCfg.TurnConn != nil {
					err := SignalPeer(signal.Server, nm_models.Signal{
						Server:            signal.Server,
						FromHostPubKey:    signal.ToHostPubKey,
						TurnRelayEndpoint: hostTurnCfg.TurnConn.LocalAddr().String(),
						ToHostPubKey:      signal.FromHostPubKey,
						Reply:             true,
					})
					if err != nil {
						logger.Log(0, "---> failed to signal peer: ", err.Error())
						continue
					}
				}

			}
		}
	}
}

func ShouldUseTurn(natType string) bool {
	return true
	// if behind  DOUBLE or ASYM Nat type, allocate turn address for the host
	if natType == nm_models.NAT_Types.Asymmetric || natType == nm_models.NAT_Types.Double {
		return true
	}
	return false
}
