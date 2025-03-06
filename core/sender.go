package core

func SendDfStat(configContext *ConfigContext, dfsctx *DeliverStatCtx, dfstat *TTDINode) {
	dfstat.Name = configContext.ArgosId
	dfstat.FinishStatistic("", "", "", configContext.Log, true, dfsctx)
	configContext.Log.Printf("Sending dataflow statistic to kernel: %s\n", dfstat.Name)
	dfstatClone := *dfstat
	go func(dsc *TTDINode) {
		if configContext != nil && *configContext.DfsChan != nil {
			*configContext.DfsChan <- dsc
		}
	}(&dfstatClone)
}
