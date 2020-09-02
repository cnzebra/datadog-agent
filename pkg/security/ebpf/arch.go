package ebpf

import (
	"bytes"
	"errors"
	"io/ioutil"
	"regexp"

	"golang.org/x/sys/unix"
)

const defaultSymFile = "/proc/kallsyms"

var RuntimeArch string

// Returns the qualified syscall named by going through '/proc/kallsyms' on the
// system on which its executed. It allows BPF programs that may have been compiled
// for older syscall functions to run on newer kernels
func GetSyscallFnName(name string) (string, error) {
	// Get kernel symbols
	syms, err := ioutil.ReadFile(defaultSymFile)
	if err != nil {
		return "", err
	}
	return getSyscallFnNameWithKallsyms(name, string(syms))
}

func getSyscallFnNameWithKallsyms(name string, kallsymsContent string) (string, error) {
	// We should search for new syscall function like "__x64__sys_open"
	// Note the start of word boundary. Should return exactly one string
	regexStr := `(\b__` + RuntimeArch + `_[Ss]y[sS]_` + name + `\b)`
	fnRegex := regexp.MustCompile(regexStr)

	match := fnRegex.FindAllString(kallsymsContent, -1)

	// If nothing found, search for old syscall function to be sure
	if len(match) == 0 {
		newRegexStr := `(\b[Ss]y[sS]_` + name + `\b)`
		fnRegex = regexp.MustCompile(newRegexStr)
		newMatch := fnRegex.FindAllString(kallsymsContent, -1)

		// If we get something like 'sys_open' or 'SyS_open', return
		// either (they have same addr) else, just return original string
		if len(newMatch) >= 1 {
			return newMatch[0], nil
		} else {
			return "", errors.New("could not find a valid syscall name")
		}
	}

	return match[0], nil
}

func init() {
	var uname unix.Utsname
	if err := unix.Uname(&uname); err != nil {
		panic(err)
	}

	switch string(uname.Machine[:bytes.IndexByte(uname.Machine[:], 0)]) {
	case "x86_64":
		RuntimeArch = "x64"
	default:
		RuntimeArch = "ia32"
	}
}
