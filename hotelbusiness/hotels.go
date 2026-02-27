//go:build !solution

package hotelbusiness

import "slices"

type Guest struct {
	CheckInDate  int
	CheckOutDate int
}

type Load struct {
	StartDate  int
	GuestCount int
}


func ComputeLoad(guests []Guest) []Load {
	if len(guests) == 0 {
		return []Load{}
	}
	checkIns := make([]int, len(guests))
	checkOuts := make([]int, len(guests))
	for idx, guest := range guests {
		checkIns[idx] = guest.CheckInDate
		checkOuts[idx] = guest.CheckOutDate
	}
	slices.Sort(checkIns)
	slices.Sort(checkOuts)
	
	result := make([]Load, 0, len(guests) * 2)
	checkInIdx, checkOutIdx := 0, 0
	var guestCount int

	for checkInIdx < len(guests) || checkOutIdx < len(guests) {
		var nextDay int
		switch {
		case checkOutIdx >= len(guests):
			nextDay = checkIns[checkInIdx]
		case checkInIdx >= len(guests):
			nextDay = checkOuts[checkOutIdx]
		case checkIns[checkInIdx] < checkOuts[checkOutIdx]:
			nextDay = checkIns[checkInIdx]
		default:
			nextDay = checkOuts[checkOutIdx]
		}

		for checkOutIdx < len(guests) && checkOuts[checkOutIdx] == nextDay {
			guestCount--
			checkOutIdx++
		}
		for checkInIdx < len(guests) && checkIns[checkInIdx] == nextDay {
			guestCount++
			checkInIdx++
		}

		if len(result) == 0 || result[len(result)-1].GuestCount != guestCount {
			load := Load{StartDate: nextDay, GuestCount: guestCount}
			result = append(result, load)
		}
	}
	return result
}
