package storage

import (
	"bufio"
	"os"
	"strings"
)

func GetStorageMounts(root string) ([]string, error) {
	file, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var mounts []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 5 {
			continue
		}

		mountPoint := fields[4]

		if strings.HasPrefix(mountPoint, root+"/") {
			mounts = append(mounts, mountPoint)
		}
	}

	return mounts, nil
}
