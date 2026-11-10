//go:build windows

package executor

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// jobs maps a process ID to the Job Object that holds it.
var jobs sync.Map

// PrepareCommand has nothing to set on Windows. The Job Object groups the process.
func PrepareCommand(command *exec.Cmd) {}

// attachProcess puts the process in a Job Object that kills its children when the job closes.
func attachProcess(process *os.Process) error {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return fmt.Errorf("create job object: %w", err)
	}

	var limits windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&limits)),
		uint32(unsafe.Sizeof(limits)),
	); err != nil {
		_ = windows.CloseHandle(job)
		return fmt.Errorf("set job limits: %w", err)
	}

	processHandle, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(process.Pid))
	if err != nil {
		_ = windows.CloseHandle(job)
		return fmt.Errorf("open process: %w", err)
	}
	defer windows.CloseHandle(processHandle)

	if err := windows.AssignProcessToJobObject(job, processHandle); err != nil {
		_ = windows.CloseHandle(job)
		return fmt.Errorf("assign process to job: %w", err)
	}

	jobs.Store(process.Pid, job)
	return nil
}

// ReleaseProcess closes the process's Job Object. Closing it kills any child still running.
func ReleaseProcess(process *os.Process) {
	if process == nil {
		return
	}
	if job, ok := jobs.LoadAndDelete(process.Pid); ok {
		_ = windows.CloseHandle(job.(windows.Handle))
	}
}

// StopProcessGroup terminates the process's Job Object, which ends the process and its children.
func StopProcessGroup(process *os.Process) error {
	if process == nil {
		return nil
	}
	job, ok := jobs.Load(process.Pid)
	if !ok {
		return process.Kill()
	}
	return windows.TerminateJobObject(job.(windows.Handle), 1)
}
