package sqlformatter

func lineColFromIndex(input string, index int) (int, int) {
	line := 1
	col := 1
	for i, r := range input {
		if i >= index {
			break
		}
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}
