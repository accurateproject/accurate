package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/kr/pty"
)

func InitDataDb(cfg *config.Config) error {
	ratingDb, err := ConfigureRatingStorage(*cfg.TariffPlanDb.Host, *cfg.TariffPlanDb.Port, *cfg.TariffPlanDb.Name, *cfg.TariffPlanDb.User, *cfg.TariffPlanDb.Password, cfg.Cache, *cfg.DataDb.LoadHistorySize)
	if err != nil {
		return err
	}
	accountDb, err := ConfigureAccountingStorage(*cfg.DataDb.Host, *cfg.DataDb.Port, *cfg.DataDb.Name,
		*cfg.DataDb.User, *cfg.DataDb.Password, cfg.Cache, *cfg.DataDb.LoadHistorySize)
	if err != nil {
		return err
	}
	for _, db := range []Storage{ratingDb, accountDb} {
		if err := db.Flush(); err != nil {
			return err
		}
	}
	ratingDb.PreloadRatingCache()
	CheckVersion(accountDb) // Write version before starting
	return nil
}

func InitCdrDb(cfg *config.Config) error {
	cdrDb, err := ConfigureCdrStorage(*cfg.CdrDb.Host, *cfg.CdrDb.Port, *cfg.CdrDb.Name, *cfg.CdrDb.User, *cfg.CdrDb.Password,
		*cfg.CdrDb.MaxOpenConns, *cfg.CdrDb.MaxIdleConns, cfg.CdrDb.CdrsIndexes)
	if err != nil {
		return err
	}
	if err := cdrDb.Flush(); err != nil {
		return err
	}
	return nil
}

// Return reference towards the command started so we can stop it if necessary
func StartEngine(cfgPath string, waitEngine int) (*exec.Cmd, error) {
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		return nil, err
	}
	engine := exec.Command(enginePath, "-config_dir", cfgPath)
	if err := engine.Start(); err != nil {
		return nil, err
	}
	time.Sleep(time.Duration(waitEngine) * time.Millisecond) // Give time to rater to fire up
	return engine, nil
}

func KillEngine(waitEngine int) error {
	if err := exec.Command("pkill", "cgr-engine").Run(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitEngine) * time.Millisecond)
	return nil
}

func StopStartEngine(cfgPath string, waitEngine int) (*exec.Cmd, error) {
	KillEngine(waitEngine)
	return StartEngine(cfgPath, waitEngine)
}

type PjsuaAccount struct {
	Id, Username, Password, Realm, Registrar string
}

// Returns file reference where we can write to control pjsua in terminal
func StartPjsuaListener(acnts []*PjsuaAccount, localPort, waitDur time.Duration) (*os.File, error) {
	cmdArgs := []string{fmt.Sprintf("--local-port=%d", localPort), "--null-audio", "--auto-answer=200", "--max-calls=32", "--app-log-level=0"}
	for idx, acnt := range acnts {
		if idx != 0 {
			cmdArgs = append(cmdArgs, "--next-account")
		}
		cmdArgs = append(cmdArgs, "--id="+acnt.Id, "--registrar="+acnt.Registrar, "--username="+acnt.Username, "--password="+acnt.Password, "--realm="+acnt.Realm)
	}
	pjsuaPath, err := exec.LookPath("pjsua")
	if err != nil {
		return nil, err
	}
	pjsua := exec.Command(pjsuaPath, cmdArgs...)
	fPty, err := pty.Start(pjsua)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	io.Copy(os.Stdout, buf) // Free the content since otherwise pjsua will not start
	time.Sleep(waitDur)     // Give time to rater to fire up
	return fPty, nil
}

func PjsuaCallUri(acnt *PjsuaAccount, dstUri, outboundUri string, callDur time.Duration, localPort int) error {
	cmdArgs := []string{"--null-audio", "--app-log-level=0", fmt.Sprintf("--local-port=%d", localPort), fmt.Sprintf("--duration=%d", int(callDur.Seconds())),
		"--outbound=" + outboundUri, "--id=" + acnt.Id, "--username=" + acnt.Username, "--password=" + acnt.Password, "--realm=" + acnt.Realm, dstUri}

	pjsuaPath, err := exec.LookPath("pjsua")
	if err != nil {
		return err
	}
	pjsua := exec.Command(pjsuaPath, cmdArgs...)
	fPty, err := pty.Start(pjsua)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	io.Copy(os.Stdout, buf)
	go func() {
		time.Sleep(callDur + (time.Duration(2) * time.Second))
		fPty.Write([]byte("q\n")) // Destroy the listener
	}()
	return nil
}

func KillProcName(procName string, waitMs int) error {
	if err := exec.Command("pkill", procName).Run(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitMs) * time.Millisecond)
	return nil
}

func CallScript(scriptPath string, subcommand string, waitMs int) error {
	if err := exec.Command(scriptPath, subcommand).Run(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitMs) * time.Millisecond) // Give time to rater to fire up
	return nil
}
