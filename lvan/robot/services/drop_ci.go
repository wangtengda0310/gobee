package main

func init() {
	if err := robot.RegisterService(name: "drop", &dropService{}); err != nil {
		logger.Fatal(msg: "register service failed", args...: "error", err)
	}
}

type dropService struct{}

func (s *dropService) Run(rbt *robot.Robot) {
	req := &client.DropReq{DropId: 101700001, Count: 100}
	if ack, err := rbt.Call(req); err != nil {
		logger.Warn(msg: "use pet failed", args...: "error", err)
	} else {
		if !ack.CheckErrCode(pb.Errno_Success) {
			logger.Warn(msg: "use pet not success", args...: "reply", ack)
		}
	}

	reqPool := &client.DropPoolReq{DropId: 6000001, Count: 100}
	if ack, err := rbt.Call(reqPool); err != nil {
		logger.Warn(msg: "use pet failed", args...: "error", err)
	} else {
		if !ack.CheckErrCode(pb.Errno_Success) {
			logger.Warn(msg: "use pet not success", args...: "reply", ack)
		}
	}

	logger.Info(msg: "pet ci finish", args...: "robot", rbt.Uid)
}
