package db

import (
	"strconv"
	"strings"

	"github.com/cego/mysql-admin/internal/model"
)

func parseInnoDBStatus(status string) map[int64]*model.TransactionInfo {
	lines := strings.Split(status, "\n")

	startIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "LIST OF TRANSACTIONS FOR EACH SESSION:") {
			startIdx = i
			break
		}
	}

	if startIdx == -1 {
		return nil
	}

	transactions := make(map[int64]*model.TransactionInfo)
	var current *model.TransactionInfo

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "--------") {
			break
		}

		if strings.HasPrefix(line, "---TRANSACTION") {
			activeTime := int64(-1)
			idx := strings.Index(line, ", ACTIVE")
			if idx >= 0 {
				rest := strings.TrimSpace(line[idx+8:])
				fields := strings.Fields(rest)
				if len(fields) > 0 {
					parsed, err := strconv.ParseInt(fields[0], 10, 64)
					if err == nil {
						activeTime = parsed
					}
				}
			}
			current = &model.TransactionInfo{
				ActiveTime: activeTime,
				Info:       []string{},
			}
		}

		if strings.HasPrefix(line, "MariaDB thread id") || strings.HasPrefix(line, "MySQL thread id") {
			parts := strings.Fields(line)
			if len(parts) >= 4 && current != nil {
				threadID, err := strconv.ParseInt(strings.TrimRight(parts[3], ","), 10, 64)
				if err == nil {
					transactions[threadID] = current
				}
			}
		}

		if current != nil {
			current.Info = append(current.Info, line)
		}
	}

	return transactions
}
