package systemsvc

import (
	"context"
	"strconv"
	"strings"
)

// readCPUTimes returns total and idle jiffies from /proc/stat.
func (s *Service) readCPUTimes() (float64, float64, bool) {
	return parseCPUTimes(readText(s.path("/proc/stat")))
}

// parseCPUTimes sums the aggregate "cpu" line of /proc/stat. Field 4 is idle time.
func parseCPUTimes(content string) (total, idle float64, ok bool) {
	firstLine, _, _ := strings.Cut(content, "\n")
	if !strings.HasPrefix(firstLine, "cpu ") {
		return 0, 0, false
	}
	values := parseNumbers(strings.Fields(firstLine)[1:])
	for _, value := range values {
		total += value
	}
	if len(values) > 3 {
		idle = values[3]
	}
	return total, idle, true
}

func parseNumbers(fields []string) []float64 {
	values := make([]float64, 0, len(fields))
	for _, field := range fields {
		value, err := strconv.ParseFloat(field, 64)
		if err != nil {
			value = 0
		}
		values = append(values, value)
	}
	return values
}

func (s *Service) readMemory() memInfo {
	return parseMemInfo(readText(s.path("/proc/meminfo")))
}

func parseMemInfo(content string) memInfo {
	var mem memInfo
	for _, line := range strings.Split(content, "\n") {
		switch {
		case strings.HasPrefix(line, "MemTotal:"):
			mem.totalKB = digitsValue(line)
		case strings.HasPrefix(line, "MemAvailable:"):
			mem.availableKB = digitsValue(line)
		}
		if mem.totalKB > 0 && mem.availableKB > 0 {
			break
		}
	}
	return mem
}

// digitsValue keeps only the digits of a line such as "MemTotal:   16318464 kB".
func digitsValue(line string) int64 {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, line)
	value, _ := strconv.ParseInt(digits, 10, 64)
	return value
}

func (s *Service) readOSName() string {
	for _, line := range strings.Split(readText(s.path("/etc/os-release")), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return cleanOSName(line)
		}
	}
	return "Unknown"
}

func cleanOSName(line string) string {
	name := strings.TrimPrefix(line, "PRETTY_NAME=")
	return strings.TrimSpace(strings.ReplaceAll(name, `"`, ""))
}

func (s *Service) readKernel(ctx context.Context) string {
	out, err := s.runner.Output(ctx, "uname", "-r")
	if err != nil {
		return "Unknown"
	}
	return strings.TrimSpace(string(out))
}

// readCPUName returns the model from lscpu. A missing lscpu gives "", not "Unknown", as before.
func (s *Service) readCPUName(ctx context.Context) string {
	out, _ := s.runner.Output(ctx, "lscpu")
	return parseCPUModel(string(out))
}

// parseCPUModel takes the second colon field of each "Model name" line, like `grep | cut -d: -f2`.
func parseCPUModel(output string) string {
	var models []string
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Model name") {
			models = append(models, colonField(line))
		}
	}
	return strings.TrimSpace(strings.Join(models, "\n"))
}

// colonField returns the text between the first and second colon, or the whole line when there is no colon.
func colonField(line string) string {
	fields := strings.Split(line, ":")
	if len(fields) < 2 {
		return line
	}
	return fields[1]
}

func (s *Service) readDriver(ctx context.Context) string {
	// Errors are ignored: the shell pipeline only ever reported awk's status.
	out, _ := s.runner.Output(ctx, "vulkaninfo", "--summary")
	if version := parseDriverVersion(string(out)); version != "" {
		return version
	}
	return "Unknown"
}

// parseDriverVersion returns the third field of the first driverVersion line.
func parseDriverVersion(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "driverVersion") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return ""
		}
		return fields[2]
	}
	return ""
}
