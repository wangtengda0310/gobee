package batch

import (
	"cmp"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/wangtengda0310/gobee/lvan/internal/execute"
	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
)

func WithSort(workdir string) error {
	entries, err := os.ReadDir(workdir)
	if err != nil {
		return err
	}
	var sorted []os.DirEntry
	for _, d := range entries {
		name := d.Name()
		// if name starts with number
		if canSort(name) {
			sorted = append(sorted, d)
		} else {
			if d.IsDir() {
				err := WithSort(filepath.Join(workdir, name))
				if err != nil {
					return err
				}
			} else {
				err := File(filepath.Join(workdir, name))
				if err != nil {
					return err
				}
			}
		}
	}

	slices.SortFunc(sorted, func(a, b os.DirEntry) int {
		return cmp.Compare(a.Name(), b.Name())
	})

	for _, entry := range sorted {
		name := entry.Name()
		if entry.IsDir() {
			err := WithSort(filepath.Join(workdir, name))
			if err != nil {
				return err
			}
		} else {
			err := File(filepath.Join(workdir, name))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func canSort(name string) bool {
	return len(name) > 0 && name[0] >= '0' && name[0] <= '9'
}

func File(file string) error {
	logger.Debug("执行 %s", file)
	file, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	command := exec.Command(file)

	workdir := filepath.Join(execute.TasksDir, "cron", filepath.Base(filepath.Dir(file)))

	err = os.MkdirAll(workdir, os.ModePerm)
	if err != nil {
		return err
	}
	err, stdout, stderr := execute.Cmd(command, workdir, os.Environ())
	if err != nil {
		return err
	}

	log := func(s string) {
		logger.Info("%s", s)
	}
	execute.CatchStdout(stdout, nil, log)

	execute.CatchStderr(stderr, nil, log)

	logger.Info("等待命令完成")
	return command.Wait()
}
