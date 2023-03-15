package terraform_extension

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

type RefreshFunc func(info *Info) (result interface{}, state string, err error)

func RefreshAction(pending, target []string, timeout, delay time.Duration, refresh RefreshFunc) HandlerFunc {
	return func(info *Info) Then {
		stateConf := &resource.StateChangeConf{
			Pending: pending,
			Target:  target,
			Refresh: func() (result interface{}, state string, err error) {
				return refresh(info)
			},
			Timeout:    timeout,
			Delay:      delay,
			MinTimeout: 3 * time.Second,
		}
		_, err := stateConf.WaitForState()
		if err != nil {
			info.AppendError(err)
			return ThenHalt
		}
		return ThenContinue
	}
}
