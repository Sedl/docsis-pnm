package modem

import "github.com/sedl/docsis-pnm/internal/constants"

func (p *Poller) FindOfdmDownstreamIdx() (int, error) {
    ifaces, err := p.getInterfaceTypes()
    if err != nil {
        return 0, err
    }
    for idx, itype := range ifaces {
        if itype == constants.IntfTypeOfdmDownstream {
            return idx, nil
        }
    }
    return 0, nil
}