package pgdbsyncer

import "github.com/sedl/docsis-pnm/internal/types"

func (m *PgDbSyncer) InsertCmtsModemUpstream(usrec *types.UpstreamModemCMTS) {
    m.copyModemUpstream.Insert(usrec.ModemId, usrec.UpstreamId, usrec.PollTime, usrec.PowerRx, usrec.SNR,
        usrec.Microrefl, usrec.Unerroreds, usrec.Correcteds, usrec.Erroreds)
}