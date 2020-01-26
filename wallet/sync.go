package wallet

import (
	"fmt"
	"sync"

	"github.com/raedahgroup/dcrlibwallet"
	"github.com/raedahgroup/godcr-gio/event"
)

const syncID = "godcr"

// Sync is the main wallet sync loop
func (wal *Wallet) Sync(wg *sync.WaitGroup) {
	defer wg.Done()

	err := wal.loadWallets(wal.root, wal.net)
	if err != nil {
		wal.Send <- err
		return
	}

	//fmt.Println("Sending loaded event")
	wal.Send <- event.WalletResponse{
		Resp: event.LoadedWalletsResp,
		Results: &event.ArgumentQueue{
			Queue: []interface{}{int(wal.multi.LoadedWalletsCount())},
		},
	}

	defer wal.multi.Shutdown()

	err = wal.multi.AddSyncProgressListener(&progressListener{
		Send: wal.Send,
	}, syncID)
	if err != nil {
		wal.Send <- err
		return
	}
	for {
		e := <-wal.Receive
		if cmd, ok := e.(event.WalletCmd); ok {
			switch cmd.Cmd {
			case event.StartSyncCmd:
				if !wal.multi.IsSyncing() {
					go func(c chan<- event.Event) {
						err := wal.multi.SpvSync()
						if err != nil {
							c <- err
						}
					}(wal.Send)
				}
			case event.ShutdownCmd:
				return
			default:
				if fun, ok := cmdMap[cmd.Cmd]; ok {
					err := fun(wal, cmd.Arguments)
					if err != nil {
						wal.Send <- err
					}
				}
			}
		} else {
			fmt.Printf("Not a wallet command %+v\n", e)
		}
	}
}

type progressListener struct {
	Send chan<- event.Event
}

func (listener *progressListener) Debug(info *dcrlibwallet.DebugInfo) {

}

func (listener *progressListener) OnSyncStarted() {
	listener.Send <- event.Sync{Event: event.SyncStart}
}

func (listener *progressListener) OnPeerConnectedOrDisconnected(numberOfConnectedPeers int32) {

}

func (listener *progressListener) OnHeadersFetchProgress(headersFetchProgress *dcrlibwallet.HeadersFetchProgressReport) {

}

func (listener *progressListener) OnAddressDiscoveryProgress(addressDiscoveryProgress *dcrlibwallet.AddressDiscoveryProgressReport) {

}

func (listener *progressListener) OnHeadersRescanProgress(headersRescanProgress *dcrlibwallet.HeadersRescanProgressReport) {

}

func (listener *progressListener) OnSyncCompleted() {
	listener.Send <- event.Sync{Event: event.SyncEnd}
}

func (listener *progressListener) OnSyncCanceled(willRestart bool) {

}

func (listener *progressListener) OnSyncEndedWithError(err error) {

}