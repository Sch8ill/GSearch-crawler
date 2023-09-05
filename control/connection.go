package control

import (
	"github.com/sch8ill/gscrawler/types"
)

type ControllerConnection struct {
	channels ChannelBundle
}

func NewConnection(channelBundle ChannelBundle) *ControllerConnection {
	return &ControllerConnection{channels: channelBundle}
}

// returns a requested job from the controller 
func (cc *ControllerConnection) GetJob() Job {
	cc.channels.JobRequestChannel <- true
	return <-cc.channels.JobChannel
}

// sends a site the controller
func (cc *ControllerConnection) SubmitResult(site types.Site) {
	cc.channels.ResultChannel <- site
}
