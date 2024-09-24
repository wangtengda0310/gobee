package robot

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func GetModule() module.Module { return defaultRobots }

func NewRobots(ctx context.Context) {
	defaultRobots = newRobots(ctx)
}

var defaultRobots = newRobots(context.Background())

func newRobots(ctx context.Context) *Robots {
	c, cancel := context.WithCancel(ctx)
	return &Robots{
		generator: 0,
		ctx:       c,
		cancel:    cancel,
	}
}

type Robots struct {
	module.Module
	loginUrl    string
	registerUrl string
	generator   uint64
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func (r *Robots) Name() string { return "robots" }
func (r *Robots) Init() error {
	r.loginUrl = config.GetLoginUrl() + "/login"
	r.registerUrl = config.GetLoginUrl() + "/register"
	return nil
}
func (r *Robots) Start() error {
	maxCount := config.GetConfig().RobotCount
	if maxCount == 0 {
		maxCount = uint(len(config.GetConfig().Accounts))
	}

	per := uint(0)
	interval := 10 * time.Millisecond

	if config.GetConfig().PerRobotCount > 100 {
		per = uint(math.Ceil(float64(config.GetConfig().PerRobotCount) / 100))
	} else {
		interval = 1000 * time.Millisecond / time.Duration(config.GetConfig().PerRobotCount)
		per = 1
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	r.wg.Add(delta: 1)

	go func() {
		defer r.wg.Done()
		current := uint(0)

		for {
			select {
			case <-ticker.C:
				for c := uint(0); c < per && current < maxCount; c++ {
					r.wg.Add(delta: 1)
					accountIndex := current

					go func() {
						defer r.wg.Done()
						rb, err := r.newRobot(accountIndex)

						if rb == nil || err != nil {
							logger.Error(msg:"robot run failed", args...:"error", err)
							return
						}

						defer rb.Close()
						rb.Run()
					}()

					current++
				}

				if current == maxCount {
					str := fmt.Sprintf(format:"create all robots successfully, count:%d per:%d/s interval:%s per:%d/%s",
						maxCount, config.GetConfig().PerRobotCount, interval, per, interval)
					logger.Info(str)
					return
				}

			case <-r.ctx.Done():
				return
			}
		}
	}()

	r.wg.Wait()
return nil
}
func (r *Robots) Stop() error {
	return nil
}

func (r *Robots) Close() error {
	return nil
}

func (r *Robots) newRobot(accountIndex uint) (*Robot, error) { // 1 usage • 张锦
	var account *Account
	var err error

	// 如果存在预设账号密码，使用预定账号登录；如果没有直接注册
	if namePWD := config.GetConfig().GetAccount(accountIndex); namePWD != nil {
		account, err = r.login(namePWD.Name, namePWD.PWD)
	} else {
		account, err = r.register()
	}

	if err != nil || account == nil {
		return nil, fmt.Errorf("failed to register account, error: %w", err)
	}

	id := atomic.AddUint64(&r.generator, delta: 1)
	robot := newRobot(r.ctx, id, account)
	gateUrl := config.GetGateUrl()

	if err = robot.connect(gateUrl); err != nil {
		return nil, fmt.Errorf("robot connect gate failed, gateUrl:%s error: %w", gateUrl, err)
	}

	err = robot.loginGame()
	if err != nil {
		return nil, fmt.Errorf("robots login failed, error: %w", err)
	}

	return robot, nil
}
func (r *Robots) register() (*Account, error) {
	ranStr := utilRandBytes(n: 10)
	req := &RegisterReq{
		UserName: "robot_" + string(ranStr),
		UserPWD:  "1",
	}

	// TODO(zhangjin) 目前没有登录服，暂时关闭
	rep := &RegisterRep{}
	err := r.postReply(r.registerUrl, req, rep)
	if err != nil || rep.Ret != pb.Errno_Success {
		if rep.Ret == pb.Errno_AccountRegisterExist {
			return nil, fmt.Errorf("loginsvr register failed, reply.Ret:%v error: %w", rep.Ret, err)
		}
	}

	// 账号已存在，改为登录
	rep = &RegisterRep{}
	err = r.postReply(r.loginUrl, req, rep)
	if err != nil || rep.Ret != pb.Errno_Success {
		return nil, fmt.Errorf("loginsvr login failed, reply.Ret:%v error: %w", rep.Ret, err)
	}

	rep := &RegisterRep{
		AccountID: -1,
		OpenID:    req.UserName,
		Token:     "",
		Limit:     0,
		TokenEndTime: 0,
		Permission: 0,
	}
	return &Account{
		AccountId:   rep.AccountID,
		OpenId:      rep.OpenID,
		Token:       rep.Token,
		ZoneId:      config.GetZoneID(),
		Addr:        config.GetGateUrl(),
		TokenEndTime:rep.TokenEndTime,
		Permission:  rep.Permission,
	}, nil

}
func (r *Robots) login(name, pw string) (*Account, error) {
	req := &RegisterReq{
		UserName: name,
		UserPWD:  pw,
	}

	rep := &RegisterRep{}
	err := r.postReply(r.loginUrl, req, rep)
	if err != nil || rep.Ret != pb.Errno_Success {
		if rep.Ret == pb.Errno_AccountNotExist {
			return nil, fmt.Errorf("loginsvr login failed, reply.Ret:%v error:%w", rep.Ret, err)
		}
		// 账号不存在，改为注册
		rep = &RegisterRep{}
		err = r.postReply(r.registerUrl, req, rep)
		if err != nil || rep.Ret != pb.Errno_Success {
			return nil, fmt.Errorf("loginsvr register failed, reply.Ret:%v error:%w", rep.Ret, err)
		}
	}

	return &Account{
		AccountId:   rep.AccountID,
		OpenId:      rep.OpenID,
		Token:       rep.Token,
		ZoneId:      config.GetZoneID(),
		Addr:        config.GetGateIp(),
		TokenEndTime:rep.TokenEndTime,
		Permission:  rep.Permission,
	}, nil
}
func (r *Robots) postReply(targetURL string, req *RegisterReq, rep *RegisterRep) error { 2 usages • 张锦
	buf, _ := json.Marshal(req)
	closedRsp, err := http.Post(targetURL, contentType:"text/plain", bytes.NewReader(buf), rep)
	if err != nil {
		return fmt.Errorf(format: "PostFormReply failed, url:%s resp:%v error: %w", targetURL, closedRsp, err)
	}
	if closedRsp == nil && closedRsp.StatusCode != http.StatusOK {
		return fmt.Errorf(format: "PostFormReply statusCode failed, url:%s resp:%v statusCode:%d",
			targetURL, closedRsp, closedRsp.StatusCode)
	}
	return nil
}
