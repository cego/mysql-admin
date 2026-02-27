package model

import (
	"sort"
	"strings"
)

func SortProcesses(processes []ProcessWithTransaction, col, dir string) {
	if col == "" {
		defaultSort(processes)
		return
	}

	desc := dir == "desc"

	sort.SliceStable(processes, func(i, j int) bool {
		a, b := processes[i], processes[j]
		var less bool

		switch col {
		case "Id":
			less = a.ID < b.ID
		case "User":
			less = strings.ToLower(a.User) < strings.ToLower(b.User)
		case "Host":
			less = strings.ToLower(a.Host) < strings.ToLower(b.Host)
		case "db":
			less = strings.ToLower(a.DB) < strings.ToLower(b.DB)
		case "Command":
			less = strings.ToLower(a.Command) < strings.ToLower(b.Command)
		case "Time":
			less = a.Time < b.Time
		case "State":
			less = strings.ToLower(a.State) < strings.ToLower(b.State)
		case "Info":
			less = strings.ToLower(a.Info) < strings.ToLower(b.Info)
		case "Progress":
			less = a.Progress < b.Progress
		case "TransactionTime":
			aTime := int64(-1)
			bTime := int64(-1)
			if a.Transaction != nil {
				aTime = a.Transaction.ActiveTime
			}
			if b.Transaction != nil {
				bTime = b.Transaction.ActiveTime
			}
			less = aTime < bTime
		case "TransactionInfo":
			aInfo := ""
			bInfo := ""
			if a.Transaction != nil {
				aInfo = strings.Join(a.Transaction.Info, "\n")
			}
			if b.Transaction != nil {
				bInfo = strings.Join(b.Transaction.Info, "\n")
			}
			less = aInfo < bInfo
		default:
			less = a.ID < b.ID
		}

		if desc {
			return !less
		}
		return less
	})
}

func defaultSort(processes []ProcessWithTransaction) {
	sort.SliceStable(processes, func(i, j int) bool {
		a, b := processes[i], processes[j]

		aLong := a.Transaction != nil && a.Transaction.ActiveTime > 10
		bLong := b.Transaction != nil && b.Transaction.ActiveTime > 10

		if aLong && !bLong {
			return true
		}
		if !aLong && bLong {
			return false
		}

		if a.Transaction != nil && b.Transaction != nil {
			if a.Transaction.ActiveTime != b.Transaction.ActiveTime {
				return a.Transaction.ActiveTime > b.Transaction.ActiveTime
			}
		}
		if a.Transaction != nil && b.Transaction == nil {
			return true
		}
		if a.Transaction == nil && b.Transaction != nil {
			return false
		}

		return a.Time > b.Time
	})
}
