//go:build !solution

package speller

var ones = []string{
	"zero", "one", "two", "three", "four",
	"five", "six", "seven", "eight", "nine",
}
var teens = []string{
	"ten", "eleven", "twelve", "thirteen", "fourteen",
	"fifteen", "sixteen", "seventeen", "eighteen", "nineteen",
}
var tens = []string{
	"", "", "twenty", "thirty", "forty",
	"fifty", "sixty", "seventy", "eighty", "ninety",
}

func Spell(n int64) string {
	if n == 0 {
		return ones[0]
	}
	if n < 0 {
		return "minus " + Spell(-n)
	}
	billions := n / 1_000_000_000
	millions := (n % 1_000_000_000) / 1_000_000
	thousands := (n % 1_000_000) / 1_000
	rest := n % 1_000

	result := ""
	if billions > 0 {
		result += Spell(billions) + " billion"
	}
	if millions > 0 {
		if result != "" {
			result += " "
		}
		result += Spell(millions) + " million"
	}
	if thousands > 0 {
		if result != "" {
			result += " "
		}
		result += Spell(thousands) + " thousand"
	}
	if rest > 0 {
		if result != "" {
			result += " "
		}
		result += spellRest(rest)
	}
	return result
}

func spellRest(n int64) string {
	hundreds := n / 100
	rest := n % 100
	
	result := ""
	if hundreds > 0 {
		result += ones[hundreds] + " hundred"
	}
	if rest >= 20 {
		if result != "" {
			result += " "
		}
		result += tens[rest/10]
		if rest%10 > 0 {
			result += "-" + ones[rest%10]
		}
	}
	if rest >= 10 && rest < 20 {
		if result != "" {
			result += " "
		}
		result += teens[rest-10]
	}
	if rest > 0 && rest < 10 {
		if result != "" {
			result += " "
		}
		result += ones[rest]
	}
	return result
}
